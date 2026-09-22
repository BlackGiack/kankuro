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

	sm := session.NewManager(repo.Dir(), repo)
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

	sm := session.NewManager(repo.Dir(), repo)
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

	sm := session.NewManager(repo.Dir(), repo, session.WithDriverRegistry(reg))
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

func TestSessionSwitchAgent(t *testing.T) {
	repo, _ := config.NewFileRepository(t.TempDir())
	wd := t.TempDir()
	_ = repo.SaveWorkspaceSources([]config.WorkspaceSource{{Path: wd, IsRoot: false}})
	workspaces, _ := repo.ListWorkspaces()

	sm := session.NewManager(repo.Dir(), repo)
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

	sm := session.NewManager(repo.Dir(), repo)
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

func TestRemoveOnlyDeletesTerminatedSessions(t *testing.T) {
	repo, _ := config.NewFileRepository(t.TempDir())
	wd := t.TempDir()
	_ = repo.SaveWorkspaceSources([]config.WorkspaceSource{{Path: wd, IsRoot: false}})
	workspaces, _ := repo.ListWorkspaces()

	sm := session.NewManager(repo.Dir(), repo)
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
