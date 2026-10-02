package session_test

import (
	"encoding/json"
	"errors"
	"kankuro/internal/config"
	"kankuro/internal/driver"
	"kankuro/internal/session"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

var errStubKill = errors.New("stub kill error")

// stubKillDriver is a minimal driver.AgentDriver whose Kill always fails, so
// tests can exercise manager.Kill's error-handling path deterministically
// instead of racing the real process's exit. GetStatus reflects whether Kill
// has run, and Kill is safe to call more than once, so the stub behaves
// consistently if a future test relaunches or defers a cleanup kill.
type stubKillDriver struct {
	output     chan []byte
	done       chan struct{}
	terminated atomic.Bool
}

func newStubKillDriver() *stubKillDriver {
	return &stubKillDriver{}
}

func (d *stubKillDriver) Spawn(_ driver.AgentConfig) (*driver.SessionHandle, error) {
	d.output = make(chan []byte)
	d.done = make(chan struct{})
	return &driver.SessionHandle{Output: d.output, Done: d.done}, nil
}

func (d *stubKillDriver) Kill() error {
	if d.terminated.CompareAndSwap(false, true) {
		close(d.output)
		close(d.done)
	}
	return errStubKill
}

func (d *stubKillDriver) SetAgentsMD(_ config.IDCard, _ []config.Persona, _ []config.MCPServer, _ string, _ bool) error {
	return nil
}
func (d *stubKillDriver) SetSkills(_ []config.Skill, _ ...string) error      { return nil }
func (d *stubKillDriver) SetPersonas(_ []config.Persona, _ ...string) error  { return nil }
func (d *stubKillDriver) SelectMCPs(_ []config.MCPServer, _ ...string) error { return nil }
func (d *stubKillDriver) SetModel(_ string) []string                         { return nil }
func (d *stubKillDriver) SendInput(_ string) error                           { return nil }
func (d *stubKillDriver) SendInputBytes(_ []byte) (int, error)               { return 0, nil }
func (d *stubKillDriver) GetAgentFolder() string                             { return ".agents" }

func (d *stubKillDriver) GetStatus() driver.SessionStatus {
	if d.terminated.Load() {
		return driver.StatusTerminated
	}
	return driver.StatusRunning
}

func TestSessionLaunchAndKill(t *testing.T) {
	repo, _ := config.NewFileRepository(t.TempDir())
	wd := t.TempDir()
	_ = repo.SaveWorkspaceSources([]config.WorkspaceSource{{Path: wd, IsRoot: false}})
	workspaces, _ := repo.ListWorkspaces()

	sm := newTestManager(t, repo.Dir(), repo)
	card := config.IDCard{
		ID:          "card-1",
		Name:        "echo",
		Driver:      "cat",
		WorkspaceID: workspaces[0].ID,
	}

	sid, err := sm.Launch(card, nil, nil, nil)
	if err != nil {
		t.Fatalf("launch: %v", err)
	}
	time.Sleep(250 * time.Millisecond)

	found := false
	for _, s := range sm.List() {
		if s.ID == sid && (strings.EqualFold(string(s.Status), "running") || s.Status == session.StatusWaiting) {
			found = true
		}
	}
	if !found {
		t.Fatalf("session not found in list")
	}

	if errKill := sm.Kill(sid); errKill != nil {
		t.Fatalf("kill: %v", errKill)
	}
	time.Sleep(100 * time.Millisecond)

	for _, s := range sm.List() {
		if s.ID == sid && s.Status != session.StatusTerminated {
			t.Fatalf("expected terminated, got %s", s.Status)
		}
	}
}

// TestKillIsIdempotent checks the pre-existing early-return path in
// GenericCLIDriver.Kill: by the time a second Kill runs, watchProcess has
// already nilled d.cmd, so Kill returns nil at the "no process" guard. This
// does not exercise the new [os.ErrProcessDone] mapping added in step 3 (see
// TestKillPersistsTerminatedDespiteDriverError for that); it pins the
// idempotency behavior a caller can observe either way.
func TestKillIsIdempotent(t *testing.T) {
	repo, _ := config.NewFileRepository(t.TempDir())
	wd := t.TempDir()
	_ = repo.SaveWorkspaceSources([]config.WorkspaceSource{{Path: wd, IsRoot: false}})
	workspaces, _ := repo.ListWorkspaces()

	sm := newTestManager(t, repo.Dir(), repo)
	card := config.IDCard{
		ID:          "card-kill-twice",
		Name:        "kill-twice",
		Driver:      "cat",
		WorkspaceID: workspaces[0].ID,
	}

	sid, err := sm.Launch(card, nil, nil, nil)
	if err != nil {
		t.Fatalf("launch: %v", err)
	}
	time.Sleep(100 * time.Millisecond)

	if errKill := sm.Kill(sid); errKill != nil {
		t.Fatalf("first kill: %v", errKill)
	}
	time.Sleep(100 * time.Millisecond)

	// The process is already gone; killing again must still succeed and
	// leave the session Terminated, not surface a "process already finished"
	// error.
	if errKill := sm.Kill(sid); errKill != nil {
		t.Fatalf("second kill should be benign, got: %v", errKill)
	}

	s, ok := sm.GetSession(sid)
	if !ok {
		t.Fatalf("session not found")
	}
	if s.Status != session.StatusTerminated {
		t.Fatalf("expected terminated, got %s", s.Status)
	}
}

// TestKillPersistsTerminatedDespiteDriverError pins the actual step 3 fix:
// manager.Kill must mark the session Terminated and persist it even when the
// driver's Kill returns an error, instead of bailing out before the state
// update. Uses a stub driver (via WithDriverRegistry) whose Kill always
// fails, so the outcome does not depend on the real process having already
// exited.
func TestKillPersistsTerminatedDespiteDriverError(t *testing.T) {
	repo, _ := config.NewFileRepository(t.TempDir())
	wd := t.TempDir()
	_ = repo.SaveWorkspaceSources([]config.WorkspaceSource{{Path: wd, IsRoot: false}})
	workspaces, _ := repo.ListWorkspaces()

	stub := newStubKillDriver()
	reg := driver.NewRegistry()
	reg.Register("stub-kill-fail", func() driver.AgentDriver { return stub })

	sm := newTestManager(t, repo.Dir(), repo, session.WithDriverRegistry(reg))
	card := config.IDCard{
		ID:          "card-kill-fail",
		Name:        "kill-fail",
		Driver:      "stub-kill-fail",
		WorkspaceID: workspaces[0].ID,
	}

	sid, err := sm.Launch(card, nil, nil, nil)
	if err != nil {
		t.Fatalf("launch: %v", err)
	}

	errKill := sm.Kill(sid)
	if !errors.Is(errKill, errStubKill) {
		t.Fatalf("expected kill to surface driver error %v, got %v", errStubKill, errKill)
	}

	s, ok := sm.GetSession(sid)
	if !ok {
		t.Fatalf("session not found")
	}
	if s.Status != session.StatusTerminated {
		t.Fatalf("expected Terminated despite driver error, got %s", s.Status)
	}

	// Confirm the state made it to disk too, not just the in-memory Session.
	// Read the persisted file directly: NewManager's hydrate loop hard-codes
	// every loaded session's Status to Terminated (a process cannot survive a
	// restart), so re-loading through the manager would pass even if Kill's
	// persist were deleted.
	raw, errRead := os.ReadFile(filepath.Join(repo.Dir(), "sessions", sid+".json"))
	if errRead != nil {
		t.Fatalf("session not persisted: %v", errRead)
	}
	var persisted struct {
		Status string `json:"status"`
	}
	if errJSON := json.Unmarshal(raw, &persisted); errJSON != nil {
		t.Fatalf("persisted session is not valid JSON: %v", errJSON)
	}
	if persisted.Status != string(session.StatusTerminated) {
		t.Fatalf("expected persisted status terminated, got %q", persisted.Status)
	}
}

// TestLaunchWrapsErrPersistFailedButSessionStaysUsable pins the Finding-1
// fix: a persist failure after a successful spawn must be reported as
// ErrPersistFailed (distinguishable via [errors.Is] from "never started"),
// while the process keeps running and the session stays fully usable.
func TestLaunchWrapsErrPersistFailedButSessionStaysUsable(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("running as root: file permissions are not enforced")
	}

	repo, _ := config.NewFileRepository(t.TempDir())
	wd := t.TempDir()
	_ = repo.SaveWorkspaceSources([]config.WorkspaceSource{{Path: wd, IsRoot: false}})
	workspaces, _ := repo.ListWorkspaces()

	sessionsDir := filepath.Join(repo.Dir(), "sessions")
	if errMkdir := os.MkdirAll(sessionsDir, 0o700); errMkdir != nil {
		t.Fatalf("mkdir sessions dir: %v", errMkdir)
	}
	if errChmod := os.Chmod(sessionsDir, 0o500); errChmod != nil {
		t.Fatalf("chmod: %v", errChmod)
	}
	defer os.Chmod(sessionsDir, 0o700)

	sm := newTestManager(t, repo.Dir(), repo)
	card := config.IDCard{
		ID:          "card-persist-fail",
		Name:        "persist-fail",
		Driver:      "cat",
		WorkspaceID: workspaces[0].ID,
	}

	sid, err := sm.Launch(card, nil, nil, nil)
	if !errors.Is(err, session.ErrPersistFailed) {
		t.Fatalf("expected ErrPersistFailed, got: %v", err)
	}
	defer func() {
		_ = os.Chmod(sessionsDir, 0o700)
		_ = sm.Kill(sid)
	}()
	time.Sleep(100 * time.Millisecond)

	s, ok := sm.GetSession(sid)
	if !ok {
		t.Fatalf("expected session to exist despite the persist failure")
	}
	if s.Status == session.StatusTerminated {
		t.Fatalf("expected the process to still be running, got %s", s.Status)
	}

	handle, errAttach := sm.Attach(sid)
	if errAttach != nil {
		t.Fatalf("expected attach to succeed despite the persist failure: %v", errAttach)
	}
	handle.Close()
}

func TestSessionSwitchAgent(t *testing.T) {
	repo, _ := config.NewFileRepository(t.TempDir())
	wd := t.TempDir()
	_ = repo.SaveWorkspaceSources([]config.WorkspaceSource{{Path: wd, IsRoot: false}})
	workspaces, _ := repo.ListWorkspaces()

	sm := newTestManager(t, repo.Dir(), repo)
	card1 := config.IDCard{
		ID:          "card-1",
		Name:        "alpha",
		Driver:      "cat",
		WorkspaceID: workspaces[0].ID,
	}

	sid, err := sm.Launch(card1, nil, nil, nil)
	if err != nil {
		t.Fatalf("launch: %v", err)
	}
	time.Sleep(100 * time.Millisecond)

	sess, ok := sm.GetSession(sid)
	if !ok {
		t.Fatalf("session not found")
	}
	origWD := sess.WorkingDir
	if errKill := sm.Kill(sid); errKill != nil {
		t.Fatalf("kill: %v", errKill)
	}

	card2 := config.IDCard{
		ID:          "card-2",
		Name:        "beta",
		Driver:      "cat",
		WorkspaceID: workspaces[0].ID,
	}
	if errSwitch := sm.SwitchAgent(sid, card2, nil, nil, nil); errSwitch != nil {
		t.Fatalf("switch: %v", errSwitch)
	}
	time.Sleep(100 * time.Millisecond)

	s, ok := sm.GetSession(sid)
	if !ok {
		t.Fatalf("session not found after switch")
	}
	if s.CardID != card2.ID {
		t.Fatalf("expected card id %s, got %s", card2.ID, s.CardID)
	}
	if s.Name != card1.Name {
		t.Fatalf("expected name %s, got %s", card1.Name, s.Name)
	}
	if s.WorkingDir != origWD {
		t.Fatalf("expected working dir %s, got %s", origWD, s.WorkingDir)
	}
	if s.Status == session.StatusTerminated {
		t.Fatalf("expected running after switch, got %s", s.Status)
	}
	if errKillAfter := sm.Kill(sid); errKillAfter != nil {
		t.Fatalf("kill after switch: %v", errKillAfter)
	}
}

func TestSessionSwitchAgentPreservesCustomName(t *testing.T) {
	repo, _ := config.NewFileRepository(t.TempDir())
	wd := t.TempDir()
	_ = repo.SaveWorkspaceSources([]config.WorkspaceSource{{Path: wd, IsRoot: false}})
	workspaces, _ := repo.ListWorkspaces()

	sm := newTestManager(t, repo.Dir(), repo)
	card1 := config.IDCard{
		ID:          "card-1",
		Name:        "alpha",
		Driver:      "cat",
		WorkspaceID: workspaces[0].ID,
	}

	sid, err := sm.Launch(card1, nil, nil, nil)
	if err != nil {
		t.Fatalf("launch: %v", err)
	}
	time.Sleep(100 * time.Millisecond)

	customName := "My Project Session"
	if errSet := sm.SetName(sid, customName); errSet != nil {
		t.Fatalf("set name: %v", errSet)
	}

	if errKill := sm.Kill(sid); errKill != nil {
		t.Fatalf("kill: %v", errKill)
	}
	time.Sleep(100 * time.Millisecond)

	card2 := config.IDCard{
		ID:          "card-2",
		Name:        "beta",
		Driver:      "cat",
		WorkspaceID: workspaces[0].ID,
	}
	if errSwitch := sm.SwitchAgent(sid, card2, nil, nil, nil); errSwitch != nil {
		t.Fatalf("switch: %v", errSwitch)
	}
	time.Sleep(100 * time.Millisecond)

	s, ok := sm.GetSession(sid)
	if !ok {
		t.Fatalf("session not found after switch")
	}
	if s.Name != customName {
		t.Errorf("expected name %s, got %s", customName, s.Name)
	}

	_ = sm.Kill(sid)
}

// TestSetNamePropagatesPersistFailure pins the step-2 fix: a save failure
// inside persistSessionLocked must propagate out of SetName rather than be
// silently discarded. The sessions dir is made unwritable to force the
// failure deterministically.
func TestSetNamePropagatesPersistFailure(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("running as root: file permissions are not enforced")
	}

	repo, _ := config.NewFileRepository(t.TempDir())
	wd := t.TempDir()
	_ = repo.SaveWorkspaceSources([]config.WorkspaceSource{{Path: wd, IsRoot: false}})
	workspaces, _ := repo.ListWorkspaces()

	sm := newTestManager(t, repo.Dir(), repo)
	card := config.IDCard{
		ID:          "card-setname-fail",
		Name:        "setname-fail",
		Driver:      "cat",
		WorkspaceID: workspaces[0].ID,
	}

	sid, err := sm.Launch(card, nil, nil, nil)
	if err != nil {
		t.Fatalf("launch: %v", err)
	}
	defer sm.Kill(sid)
	time.Sleep(100 * time.Millisecond)

	// Session records are written with temp-file + rename, which replaces the
	// target regardless of its own mode - so making the file read-only no
	// longer blocks a write. Strip write permission from the *directory*
	// instead: creating the temp file there is what fails.
	sessionsDir := filepath.Join(repo.Dir(), "sessions")
	if errChmod := os.Chmod(sessionsDir, 0o500); errChmod != nil {
		t.Fatalf("chmod: %v", errChmod)
	}
	defer os.Chmod(sessionsDir, 0o700)

	if errSet := sm.SetName(sid, "new-name"); errSet == nil {
		t.Fatalf("expected SetName to surface the persist failure")
	}
}

// TestRemoveIsBenignForMissingSessionFile pins the Remove contract: if the
// on-disk session file is already gone, that is not a failure - the delete
// the caller asked for already holds - so Remove must return nil (mirroring
// Kill's treatment of [os.ErrProcessDone] as benign), and must still drop the
// session from the in-memory registry.
func TestRemoveIsBenignForMissingSessionFile(t *testing.T) {
	repo, _ := config.NewFileRepository(t.TempDir())
	wd := t.TempDir()
	_ = repo.SaveWorkspaceSources([]config.WorkspaceSource{{Path: wd, IsRoot: false}})
	workspaces, _ := repo.ListWorkspaces()

	sm := newTestManager(t, repo.Dir(), repo)
	card := config.IDCard{
		ID:          "card-remove-missing",
		Name:        "remove-missing",
		Driver:      "cat",
		WorkspaceID: workspaces[0].ID,
	}

	sid, err := sm.Launch(card, nil, nil, nil)
	if err != nil {
		t.Fatalf("launch: %v", err)
	}
	if errKill := sm.Kill(sid); errKill != nil {
		t.Fatalf("kill: %v", errKill)
	}
	time.Sleep(100 * time.Millisecond)

	sessionsDir := filepath.Join(repo.Dir(), "sessions")
	if errRemoveFile := os.Remove(filepath.Join(sessionsDir, sid+".json")); errRemoveFile != nil {
		t.Fatalf("pre-remove session file: %v", errRemoveFile)
	}

	if errRemove := sm.Remove(sid); errRemove != nil {
		t.Fatalf("expected Remove to be benign for an already-missing session file, got: %v", errRemove)
	}

	for _, s := range sm.List() {
		if s.ID == sid {
			t.Fatalf("expected session to be dropped from the registry")
		}
	}
}

func TestRemoveOnlyDeletesTerminatedSessions(t *testing.T) {
	repo, _ := config.NewFileRepository(t.TempDir())
	wd := t.TempDir()
	_ = repo.SaveWorkspaceSources([]config.WorkspaceSource{{Path: wd, IsRoot: false}})
	workspaces, _ := repo.ListWorkspaces()

	sm := newTestManager(t, repo.Dir(), repo)
	card := config.IDCard{
		ID:          "card-3",
		Name:        "remove",
		Driver:      "cat",
		WorkspaceID: workspaces[0].ID,
	}

	sid, err := sm.Launch(card, nil, nil, nil)
	if err != nil {
		t.Fatalf("launch: %v", err)
	}

	if errRemove := sm.Remove(sid); errRemove == nil {
		t.Fatalf("expected remove to fail for running session")
	}

	if errKill := sm.Kill(sid); errKill != nil {
		t.Fatalf("kill: %v", errKill)
	}
	time.Sleep(100 * time.Millisecond)

	if errRemove := sm.Remove(sid); errRemove != nil {
		t.Fatalf("remove: %v", errRemove)
	}

	for _, s := range sm.List() {
		if s.ID == sid {
			t.Fatalf("expected session to be removed")
		}
	}
}

// gateDriver is a fake driver whose Spawn and Kill can block on channels the
// test controls, so tests can hold a Relaunch/SwitchAgent/Kill in its
// unlocked phase and interleave other calls deterministically (no sleeps).
type gateDriver struct {
	spawns *atomic.Int32

	// spawnEntered (buffered) is signalled once Spawn is running; Spawn then
	// blocks until spawnGate is closed. A nil channel disables that step.
	spawnEntered chan struct{}
	spawnGate    chan struct{}
	spawnErr     error

	// killEntered/killGate do the same for Kill.
	killEntered chan struct{}
	killGate    chan struct{}
	// killDelay makes Kill take that long, like a graceful stop would.
	killDelay time.Duration

	mu         sync.Mutex
	output     chan []byte
	done       chan struct{}
	terminated atomic.Bool
}

func (d *gateDriver) Spawn(_ driver.AgentConfig) (*driver.SessionHandle, error) {
	d.spawns.Add(1)
	if d.spawnEntered != nil {
		d.spawnEntered <- struct{}{}
	}
	if d.spawnGate != nil {
		<-d.spawnGate
	}
	if d.spawnErr != nil {
		return nil, d.spawnErr
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	d.output = make(chan []byte)
	d.done = make(chan struct{})
	return &driver.SessionHandle{Output: d.output, Done: d.done}, nil
}

func (d *gateDriver) Kill() error {
	if d.killEntered != nil {
		d.killEntered <- struct{}{}
	}
	if d.killGate != nil {
		<-d.killGate
	}
	time.Sleep(d.killDelay)
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.terminated.CompareAndSwap(false, true) && d.output != nil {
		close(d.output)
		close(d.done)
	}
	return nil
}

func (d *gateDriver) SetAgentsMD(_ config.IDCard, _ []config.Persona, _ []config.MCPServer, _ string, _ bool) error {
	return nil
}
func (d *gateDriver) SetSkills(_ []config.Skill, _ ...string) error      { return nil }
func (d *gateDriver) SetPersonas(_ []config.Persona, _ ...string) error  { return nil }
func (d *gateDriver) SelectMCPs(_ []config.MCPServer, _ ...string) error { return nil }
func (d *gateDriver) SetModel(_ string) []string                         { return nil }
func (d *gateDriver) SendInput(_ string) error                           { return nil }
func (d *gateDriver) SendInputBytes(_ []byte) (int, error)               { return 0, nil }
func (d *gateDriver) GetAgentFolder() string                             { return ".agents" }

func (d *gateDriver) GetStatus() driver.SessionStatus {
	if d.terminated.Load() {
		return driver.StatusTerminated
	}
	return driver.StatusRunning
}

// gateHarness builds a Manager whose driver registry hands out gateDrivers.
// setup is called for the n-th driver constructed (0 = the one used by the
// initial Launch), letting a test gate individual activations.
type gateHarness struct {
	sm       *session.Manager
	repoDir  string
	card     config.IDCard
	spawns   *atomic.Int32
	driverMu sync.Mutex
	made     int
}

func newGateHarness(t *testing.T, useTmp bool, setup func(n int, d *gateDriver)) *gateHarness {
	t.Helper()
	root := t.TempDir()
	if useTmp {
		t.Setenv("KANKURO_TMP_DIR", filepath.Join(root, "tmp"))
	}
	repo, _ := config.NewFileRepository(filepath.Join(root, "config"))
	wd := filepath.Join(root, "ws")
	if err := os.MkdirAll(wd, 0o755); err != nil {
		t.Fatal(err)
	}
	_ = repo.SaveWorkspaceSources([]config.WorkspaceSource{{Path: wd, IsRoot: false}})
	workspaces, _ := repo.ListWorkspaces()

	h := &gateHarness{repoDir: repo.Dir(), spawns: &atomic.Int32{}}
	reg := driver.NewRegistry()
	reg.Register("gate", func() driver.AgentDriver {
		h.driverMu.Lock()
		n := h.made
		h.made++
		h.driverMu.Unlock()
		d := &gateDriver{spawns: h.spawns}
		if setup != nil {
			setup(n, d)
		}
		return d
	})
	h.sm = newTestManager(t, repo.Dir(), repo, session.WithDriverRegistry(reg))
	h.card = config.IDCard{
		ID:              "card-gate",
		Name:            "gate",
		Driver:          "gate",
		WorkspaceID:     workspaces[0].ID,
		UseTmpWorkspace: useTmp,
	}
	return h
}

// launchTerminated launches a session and leaves it Terminated.
func (h *gateHarness) launchTerminated(t *testing.T) string {
	t.Helper()
	sid, err := h.sm.Launch(h.card, nil, nil, nil)
	if err != nil {
		t.Fatalf("launch: %v", err)
	}
	if errKill := h.sm.Kill(sid); errKill != nil {
		t.Fatalf("kill: %v", errKill)
	}
	return sid
}

func waitSignal(t *testing.T, ch <-chan struct{}, what string) {
	t.Helper()
	select {
	case <-ch:
	case <-time.After(5 * time.Second):
		t.Fatalf("timed out waiting for %s", what)
	}
}

func waitErr(t *testing.T, ch <-chan error, what string) error {
	t.Helper()
	select {
	case err := <-ch:
		return err
	case <-time.After(5 * time.Second):
		t.Fatalf("timed out waiting for %s", what)
		return nil
	}
}

// gateSecondSpawn gates the second driver's Spawn (the first relaunch).
func gateSecondSpawn(entered chan struct{}, gate chan struct{}) func(int, *gateDriver) {
	return func(n int, d *gateDriver) {
		if n == 1 {
			d.spawnEntered, d.spawnGate = entered, gate
		}
	}
}

// TestConcurrentRelaunchStartsOnce: while one Relaunch is in its unlocked
// phase (status still Terminated), a second must be rejected, not spawn again.
func TestConcurrentRelaunchStartsOnce(t *testing.T) {
	entered, gate := make(chan struct{}, 1), make(chan struct{})
	h := newGateHarness(t, false, gateSecondSpawn(entered, gate))
	sid := h.launchTerminated(t)

	first := make(chan error, 1)
	go func() { first <- h.sm.Relaunch(sid, h.card, nil, nil, nil) }()
	waitSignal(t, entered, "first Relaunch to enter Spawn")

	if err := h.sm.Relaunch(sid, h.card, nil, nil, nil); err == nil {
		t.Fatal("second Relaunch during an in-flight Relaunch must be rejected")
	}

	close(gate)
	if err := waitErr(t, first, "first Relaunch"); err != nil {
		t.Fatalf("first Relaunch: %v", err)
	}
	if got := h.spawns.Load(); got != 2 { // the initial Launch + one Relaunch
		t.Fatalf("expected exactly 2 spawns (Launch + one Relaunch), got %d", got)
	}
	s, _ := h.sm.GetSession(sid)
	if s.Status != session.StatusRunning {
		t.Fatalf("expected Running, got %s", s.Status)
	}
	_ = h.sm.Kill(sid)
}

// TestSwitchAgentDuringRelaunchRejected: same guard, entered via SwitchAgent.
func TestSwitchAgentDuringRelaunchRejected(t *testing.T) {
	entered, gate := make(chan struct{}, 1), make(chan struct{})
	h := newGateHarness(t, false, gateSecondSpawn(entered, gate))
	sid := h.launchTerminated(t)

	first := make(chan error, 1)
	go func() { first <- h.sm.Relaunch(sid, h.card, nil, nil, nil) }()
	waitSignal(t, entered, "Relaunch to enter Spawn")

	if err := h.sm.SwitchAgent(sid, h.card, nil, nil, nil); err == nil {
		t.Fatal("SwitchAgent during an in-flight Relaunch must be rejected")
	}

	close(gate)
	if err := waitErr(t, first, "Relaunch"); err != nil {
		t.Fatalf("Relaunch: %v", err)
	}
	if got := h.spawns.Load(); got != 2 {
		t.Fatalf("expected exactly 2 spawns, got %d", got)
	}
	_ = h.sm.Kill(sid)
}

// TestFailedRelaunchReleasesReservation: a Relaunch that fails before the
// session activates must not leave it permanently "being started".
func TestFailedRelaunchReleasesReservation(t *testing.T) {
	errSpawn := errors.New("spawn failed")
	h := newGateHarness(t, false, func(n int, d *gateDriver) {
		if n == 1 {
			d.spawnErr = errSpawn
		}
	})
	sid := h.launchTerminated(t)

	if err := h.sm.Relaunch(sid, h.card, nil, nil, nil); !errors.Is(err, errSpawn) {
		t.Fatalf("expected spawn error, got %v", err)
	}
	if err := h.sm.Relaunch(sid, h.card, nil, nil, nil); err != nil {
		t.Fatalf("Relaunch after a failed one must succeed, got %v", err)
	}
	_ = h.sm.Kill(sid)
}

// TestFailedSwitchAgentReleasesReservation: same for SwitchAgent's early
// returns, here a Spawn failure.
func TestFailedSwitchAgentReleasesReservation(t *testing.T) {
	errSpawn := errors.New("spawn failed")
	h := newGateHarness(t, false, func(n int, d *gateDriver) {
		if n == 1 {
			d.spawnErr = errSpawn
		}
	})
	sid := h.launchTerminated(t)

	if err := h.sm.SwitchAgent(sid, h.card, nil, nil, nil); !errors.Is(err, errSpawn) {
		t.Fatalf("expected spawn error, got %v", err)
	}
	if err := h.sm.SwitchAgent(sid, h.card, nil, nil, nil); err != nil {
		t.Fatalf("SwitchAgent after a failed one must succeed, got %v", err)
	}
	_ = h.sm.Kill(sid)
}

// TestRemoveDuringRelaunchRejected: a relaunching session is still Terminated,
// but Remove must not delete the tmp workspace that is being synced into.
func TestRemoveDuringRelaunchRejected(t *testing.T) {
	entered, gate := make(chan struct{}, 1), make(chan struct{})
	h := newGateHarness(t, true, gateSecondSpawn(entered, gate))
	sid := h.launchTerminated(t)
	s, _ := h.sm.GetSession(sid)
	tmpWD := s.WorkingDir
	if _, err := os.Stat(tmpWD); err != nil {
		t.Fatalf("tmp workspace should exist: %v", err)
	}

	first := make(chan error, 1)
	go func() { first <- h.sm.Relaunch(sid, h.card, nil, nil, nil) }()
	waitSignal(t, entered, "Relaunch to enter Spawn")

	if err := h.sm.Remove(sid); err == nil {
		t.Fatal("Remove during an in-flight Relaunch must be rejected")
	}
	if _, err := os.Stat(tmpWD); err != nil {
		t.Fatalf("tmp workspace must survive the rejected Remove: %v", err)
	}
	if _, ok := h.sm.GetSession(sid); !ok {
		t.Fatal("session must still be registered")
	}

	close(gate)
	if err := waitErr(t, first, "Relaunch"); err != nil {
		t.Fatalf("Relaunch: %v", err)
	}
	_ = h.sm.Kill(sid)
}

// TestKillAfterRemoveDoesNotResurrectRecord: Kill released the lock around the
// (slow) driver Kill; if the session was removed meanwhile, Kill must not
// persist it and re-create the record Remove just deleted.
func TestKillAfterRemoveDoesNotResurrectRecord(t *testing.T) {
	entered, gate := make(chan struct{}, 1), make(chan struct{})
	h := newGateHarness(t, false, func(n int, d *gateDriver) {
		if n == 0 {
			d.killEntered, d.killGate = entered, gate
		}
	})
	sid, err := h.sm.Launch(h.card, nil, nil, nil)
	if err != nil {
		t.Fatalf("launch: %v", err)
	}

	killed := make(chan error, 1)
	go func() { killed <- h.sm.Kill(sid) }()
	waitSignal(t, entered, "Kill to enter the driver")

	// Simulate the process exiting on its own while Kill is still blocked.
	h.sm.SetStatusForTest(sid, session.StatusTerminated)
	if errRemove := h.sm.Remove(sid); errRemove != nil {
		t.Fatalf("remove: %v", errRemove)
	}

	close(gate)
	if errKill := waitErr(t, killed, "Kill"); errKill != nil {
		t.Fatalf("kill: %v", errKill)
	}

	record := filepath.Join(h.repoDir, "sessions", sid+".json")
	if _, errStat := os.Stat(record); !errors.Is(errStat, os.ErrNotExist) {
		t.Fatalf("Kill resurrected the removed session's record (stat err: %v)", errStat)
	}
}

// TestKillDuringRelaunchDoesNotTerminateNewActivation: a Kill that is still
// tearing down the old process must not mark the relaunched session dead.
func TestKillDuringRelaunchDoesNotTerminateNewActivation(t *testing.T) {
	entered, gate := make(chan struct{}, 1), make(chan struct{})
	h := newGateHarness(t, false, func(n int, d *gateDriver) {
		if n == 0 {
			d.killEntered, d.killGate = entered, gate
		}
	})
	sid, err := h.sm.Launch(h.card, nil, nil, nil)
	if err != nil {
		t.Fatalf("launch: %v", err)
	}

	killed := make(chan error, 1)
	go func() { killed <- h.sm.Kill(sid) }()
	waitSignal(t, entered, "Kill to enter the driver")

	// The old process exits, then the user relaunches before Kill returns.
	h.sm.SetStatusForTest(sid, session.StatusTerminated)
	if errRelaunch := h.sm.Relaunch(sid, h.card, nil, nil, nil); errRelaunch != nil {
		t.Fatalf("relaunch: %v", errRelaunch)
	}

	close(gate)
	if errKill := waitErr(t, killed, "Kill"); errKill != nil {
		t.Fatalf("kill: %v", errKill)
	}

	s, _ := h.sm.GetSession(sid)
	if s.Status == session.StatusTerminated {
		t.Fatal("the stale Kill marked the new activation Terminated")
	}
	_ = h.sm.Kill(sid)
}
