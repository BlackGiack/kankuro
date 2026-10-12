package session

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/BlackGiack/kankuro/internal/config"
	"github.com/BlackGiack/kankuro/internal/driver"
)

// ErrPersistFailed marks a "started, but its record could not be saved"
// outcome: the agent process is live and the in-memory session is fully
// usable, but the on-disk copy did not update, so it will not survive a
// restart. Callers should use [errors.Is] to tell this apart from a launch
// that never started a process at all, and can still attach/use the session
// while surfacing this as a (non-blocking) warning.
var ErrPersistFailed = errors.New("session state could not be saved")

// ErrManagerClosed is returned by Launch, Relaunch and SwitchAgent once
// Shutdown has started: the manager no longer starts processes it could not
// stop or watch.
var ErrManagerClosed = errors.New("session manager is shut down")

// syncDriverFiles clears temp directories (if needed) and syncs skill/persona/MCP files.
func (m *Manager) syncDriverFiles(
	card config.IDCard,
	skills []config.Skill,
	personas []config.Persona,
	mcps []config.MCPServer,
	workingDir string,
	clearTmp bool,
	appendAgentsMD bool,
) (driver.AgentDriver, error) {
	d := m.drivers.Get(card.Driver)

	if clearTmp {
		// TODO: if other names become relevant, implement a more elegant way of handling this
		agentDirs := []string{".agents", ".claude"}
		if d.GetAgentFolder() != agentDirs[0] && d.GetAgentFolder() != agentDirs[1] {
			agentDirs = append(agentDirs, d.GetAgentFolder())
		}
		for _, agentDir := range agentDirs {
			// Clear existing .agents directories in tmp workspace
			skillsDir := filepath.Join(workingDir, agentDir, "skills")
			if err := os.RemoveAll(skillsDir); err != nil {
				return nil, fmt.Errorf("clear tmp skills dir: %w", err)
			}
			personasDir := filepath.Join(workingDir, agentDir, "agents")
			if err := os.RemoveAll(personasDir); err != nil {
				return nil, fmt.Errorf("clear tmp personas dir: %w", err)
			}
		}
	}

	if err := d.SetAgentsMD(card, personas, mcps, workingDir, appendAgentsMD); err != nil {
		return nil, fmt.Errorf("set agents.md: %w", err)
	}
	if err := d.SetSkills(skills, workingDir); err != nil {
		return nil, fmt.Errorf("set skills: %w", err)
	}
	if err := d.SetPersonas(personas, workingDir); err != nil {
		return nil, fmt.Errorf("set personas: %w", err)
	}
	if err := d.SelectMCPs(mcps, workingDir); err != nil {
		return nil, fmt.Errorf("select mcps: %w", err)
	}
	return d, nil
}

func (m *Manager) Launch(card config.IDCard, skills []config.Skill, personas []config.Persona, mcps []config.MCPServer) (string, error) {
	sid := uuid.NewString()
	// The closed check and the Add share one critical section, the same
	// ordering argument as startWatchLocked: Shutdown sets closed under m.mu
	// before it Waits, so no Add can race a Wait that started at zero.
	m.mu.Lock()
	if m.closed {
		m.mu.Unlock()
		return sid, ErrManagerClosed
	}
	m.activations.Add(1)
	m.mu.Unlock()
	defer m.activations.Done()
	return sid, m.startSession(sid, card, skills, personas, mcps, true, "")
}

// activationSnapshot holds the session fields Relaunch/SwitchAgent need while
// the lock is not held, copied under m.mu so they are never read racily.
type activationSnapshot struct {
	WorkingDir      string
	UseTmpWorkspace bool
}

// reserveForActivation checks, under m.mu, that the session can be
// (re)activated and marks it as activating. It takes the lock itself; the
// caller must call releaseActivation on every failure path (success clears
// the mark in activateSessionLocked).
//
// On success it has also registered the activation with m.activations (in the
// same critical section as the !closed check), and the caller must call
// m.activations.Done() when the activation is over, after any
// releaseActivation, so Shutdown does not return while it is still working.
func (m *Manager) reserveForActivation(id, notTerminatedMsg string) (activationSnapshot, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.closed {
		return activationSnapshot{}, ErrManagerClosed
	}
	s, ok := m.sessions[id]
	if !ok {
		return activationSnapshot{}, fmt.Errorf("session not found: %s", id)
	}
	if s.Status != StatusTerminated {
		return activationSnapshot{}, errors.New(notTerminatedMsg)
	}
	if s.activating {
		return activationSnapshot{}, errors.New("session is already being started")
	}
	s.activating = true
	m.activations.Add(1)
	return activationSnapshot{WorkingDir: s.WorkingDir, UseTmpWorkspace: s.UseTmpWorkspace}, nil
}

// releaseActivation clears the reservation taken by reserveForActivation after
// an activation that did not happen.
func (m *Manager) releaseActivation(id string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if s, ok := m.sessions[id]; ok {
		s.activating = false
	}
}

// Relaunch restarts a terminated session using its existing workspace (if tmp).
func (m *Manager) Relaunch(id string, card config.IDCard, skills []config.Skill, personas []config.Persona, mcps []config.MCPServer) error {
	snap, err := m.reserveForActivation(id, "session is already running")
	if err != nil {
		return err
	}
	defer m.activations.Done()
	err = m.startSession(id, card, skills, personas, mcps, false, snap.WorkingDir)
	// ErrPersistFailed means the activation did happen (and already cleared
	// the reservation); any other error means it did not.
	if err != nil && !errors.Is(err, ErrPersistFailed) {
		m.releaseActivation(id)
	}
	return err
}

// SwitchAgent restarts a terminated session using its existing workspace, but with a new agent card.
func (m *Manager) SwitchAgent(
	id string,
	card config.IDCard,
	skills []config.Skill,
	personas []config.Persona,
	mcps []config.MCPServer,
) error {
	snap, err := m.reserveForActivation(id, "session must be terminated before switching agents")
	if err != nil {
		return err
	}
	defer m.activations.Done() // registered first, so it runs after releaseActivation below
	activated := false
	defer func() {
		if !activated {
			m.releaseActivation(id)
		}
	}()

	finalWD := snap.WorkingDir
	if finalWD == "" {
		return errors.New("session workspace is empty")
	}
	if _, errStat := os.Stat(finalWD); errStat != nil {
		return fmt.Errorf("session workspace missing: %w", errStat)
	}
	if prepareErr := m.prepareSubWorkspaces(card, finalWD); prepareErr != nil {
		return prepareErr
	}

	// Clear tmp directories if using tmp workspace
	clearTmp := snap.UseTmpWorkspace
	// SwitchAgent should replace agents.md content, not append
	d, err := m.syncDriverFiles(card, skills, personas, mcps, finalWD, clearTmp, false)
	if err != nil {
		return err
	}

	args := d.SetModel(card.Model)
	h, errSpawn := d.Spawn(driver.AgentConfig{
		Driver:            card.Driver,
		Model:             card.Model,
		Args:              args,
		WorkingDir:        finalWD,
		SystemInstruction: card.SystemInstruction,
		Env:               card.Env,
	})
	if errSpawn != nil {
		return errSpawn
	}

	now := time.Now().UTC()
	pid := 0
	if h.Cmd != nil && h.Cmd.Process != nil {
		pid = h.Cmd.Process.Pid
	}

	m.mu.Lock()
	if m.closed {
		// Shutdown started during the unlocked spawn. The process is already
		// running and no watcher will ever own it: kill it here, outside the
		// lock (Kill can block on teardown). The deferred releaseActivation
		// runs after this return.
		m.mu.Unlock()
		_ = d.Kill()
		return ErrManagerClosed
	}
	defer m.mu.Unlock()
	s, ok := m.sessions[id]
	if !ok {
		return fmt.Errorf("session not found: %s", id)
	}

	m.activateSessionLocked(s, card, d, h, pid, now)
	activated = true
	s.PendingAttachHeader = fmt.Appendf(nil, "\r\n--- Switched agent to %s at %s ---\r\n\r\n",
		card.Name, now.Format(time.RFC3339))
	gen := s.generation

	go m.trackOutput(id, gen, h.Output)
	m.startWatchLocked(id, gen, d)

	// A persist failure here must not unwind the spawn: the agent process is
	// already running. Surface the error so the caller can report it, but
	// leave the in-memory session as-is. Wrapped in ErrPersistFailed so
	// callers can tell "did not start" apart from "started, but its record
	// did not save" and still offer the normal attach flow.
	if persistErr := m.persistSessionLockedErr(s); persistErr != nil {
		return fmt.Errorf("%w: %w", ErrPersistFailed, persistErr)
	}
	return nil
}

// resolveWorkspacePath returns the override when given, else the path of the
// card's registered workspace.
func (m *Manager) resolveWorkspacePath(card config.IDCard, override string) (string, error) {
	if override != "" {
		return override, nil
	}
	workspaces, err := m.repo.ListWorkspaces()
	if err != nil {
		return "", fmt.Errorf("list workspaces: %w", err)
	}
	for _, w := range workspaces {
		if w.ID == card.WorkspaceID && w.Path != "" {
			return w.Path, nil
		}
	}
	return "", fmt.Errorf("workspace not found: %s", card.WorkspaceID)
}

// startSession runs one activation. Its callers hold an m.activations
// registration (Launch, or reserveForActivation for Relaunch), which is also
// what guarantees Shutdown had not started when they began.
func (m *Manager) startSession(
	sid string,
	card config.IDCard,
	skills []config.Skill,
	personas []config.Persona,
	mcps []config.MCPServer,
	isNew bool,
	workspaceOverride string,
) error {
	wd, err := m.resolveWorkspacePath(card, workspaceOverride)
	if err != nil {
		return err
	}

	finalWD, errPrep := m.prepareWorkspace(sid, card, wd, isNew)
	if errPrep != nil {
		return m.cleanupNewTemporaryWorkspace(sid, card, isNew, errPrep)
	}
	if prepareErr := m.prepareSubWorkspaces(card, finalWD); prepareErr != nil {
		return m.cleanupNewTemporaryWorkspace(sid, card, isNew, prepareErr)
	}

	// Sync driver files (clear generated content for tmp workspaces on every launch)
	clearTmp := card.UseTmpWorkspace
	// startSession should append to agents.md if it's a new session and using tmp workspace
	appendAgentsMD := card.UseTmpWorkspace && isNew
	d, err := m.syncDriverFiles(card, skills, personas, mcps, finalWD, clearTmp, appendAgentsMD)
	if err != nil {
		return m.cleanupNewTemporaryWorkspace(sid, card, isNew, err)
	}

	args := d.SetModel(card.Model)
	h, errSpawn := d.Spawn(driver.AgentConfig{
		Driver:            card.Driver,
		Model:             card.Model,
		Args:              args,
		WorkingDir:        finalWD,
		SystemInstruction: card.SystemInstruction,
		Env:               card.Env,
	})
	if errSpawn != nil {
		return m.cleanupNewTemporaryWorkspace(sid, card, isNew, errSpawn)
	}

	now := time.Now().UTC()
	pid := 0
	if h.Cmd != nil && h.Cmd.Process != nil {
		pid = h.Cmd.Process.Pid
	}

	m.mu.Lock()
	if m.closed {
		// Shutdown started during the unlocked spawn; see SwitchAgent. A new
		// session has no record, so nothing would ever point at its tmp copy.
		m.mu.Unlock()
		return m.cleanupNewTemporaryWorkspace(sid, card, isNew, errors.Join(ErrManagerClosed, d.Kill()))
	}
	defer m.mu.Unlock()

	s, exists := m.sessions[sid]
	if !exists {
		s = &Session{
			ID:          sid,
			Subscribers: make(map[chan []byte]struct{}),
		}
		m.sessions[sid] = s
	}

	m.activateSessionLocked(s, card, d, h, pid, now)
	s.WorkingDir = finalWD
	s.OriginalDir = wd
	s.UseTmpWorkspace = card.UseTmpWorkspace
	if isNew {
		s.Scrollback = nil
		s.PendingAttachHeader = nil
	} else {
		// Show the relaunch marker only to the next attacher (not stored in
		// scrollback so it doesn't replay on every subsequent reattach).
		s.PendingAttachHeader = fmt.Appendf(nil, "\r\n--- Relaunched at %s ---\r\n\r\n",
			now.Format(time.RFC3339))
	}
	gen := s.generation

	go m.trackOutput(sid, gen, h.Output)
	m.startWatchLocked(sid, gen, d)

	// A persist failure here must not unwind the spawn: the agent process is
	// already running. Surface the error so the caller can report it, but
	// leave the in-memory session as-is. Wrapped in ErrPersistFailed so
	// callers can tell "did not start" apart from "started, but its record
	// did not save" and still offer the normal attach flow.
	if persistErr := m.persistSessionLockedErr(s); persistErr != nil {
		return fmt.Errorf("%w: %w", ErrPersistFailed, persistErr)
	}
	return nil
}

func (m *Manager) cleanupNewTemporaryWorkspace(sid string, card config.IDCard, isNew bool, cause error) error {
	if !isNew || !card.UseTmpWorkspace {
		return cause
	}
	return errors.Join(cause, os.RemoveAll(filepath.Join(m.tmpRoot(), sid)))
}

func (m *Manager) activateSessionLocked(
	s *Session,
	card config.IDCard,
	d driver.AgentDriver,
	h *driver.SessionHandle,
	pid int,
	now time.Time,
) {
	s.generation++
	s.activating = false
	s.CardID = card.ID
	if s.Name == "" {
		s.Name = card.Name
	}
	s.PID = pid
	s.StartedAt = now
	s.LastWordAt = now
	s.Status = StatusRunning
	s.Pinged = false
	s.Driver = d
	s.PTY = h.PTY
	s.Output = h.Output
	s.Done = h.Done
}

// persistSessionLockedErr writes s to disk and returns any error. Callers
// must hold m.mu.
func (m *Manager) persistSessionLockedErr(s *Session) error {
	sessionsDir := filepath.Join(m.configDir, "sessions")
	return saveSession(sessionsDir, s.data())
}

// persistSessionLocked is the background-safe variant: failures cannot
// propagate to a caller (e.g. the watch ticker), so they are logged instead
// of discarded. See debt_impl_plan_2.md step 3 for the logging decision:
// callers must never log unconditionally to stderr (it corrupts the
// rendered TUI frame), so this goes through the Manager's injected logger
// (WithLogger), which main.go points at the debug log file, or discards
// entirely when --debug is not set.
func (m *Manager) persistSessionLocked(s *Session) {
	if err := m.persistSessionLockedErr(s); err != nil {
		m.logger.Warn("session persist failed", "session_id", s.ID, "error", err)
	}
}

// Kill terminates one managed session. If the driver reports an error (e.g.
// the process had already exited), the session is still marked Terminated
// and persisted; the error is returned after the state update.
func (m *Manager) Kill(id string) error {
	m.mu.RLock()
	s, ok := m.sessions[id]
	var d driver.AgentDriver
	var gen uint64
	if ok {
		d, gen = s.Driver, s.generation
	}
	m.mu.RUnlock()
	if !ok {
		return fmt.Errorf("session not found: %s", id)
	}

	// Called with no lock held on purpose: it can block on process teardown,
	// and watch needs m.mu in the meantime.
	var killErr error
	if d != nil {
		killErr = d.Kill()
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	// Only mark terminated if this is still the same activation of the same
	// session. The lock was released above, so a Remove in that window must
	// not have its deleted record re-created by our persist, and a Relaunch
	// must not have its new, running activation marked dead.
	if cur, still := m.sessions[id]; !still || cur != s || s.generation != gen {
		return killErr
	}
	s.Status = StatusTerminated
	s.Pinged = false
	return errors.Join(killErr, m.persistSessionLockedErr(s))
}

// Remove deletes a terminated session from the registry.
func (m *Manager) Remove(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	s, ok := m.sessions[id]
	if !ok {
		return fmt.Errorf("session not found: %s", id)
	}
	if s.Status != StatusTerminated {
		return fmt.Errorf("session still running: %s", id)
	}
	// A relaunch/switch in its unlocked phase is still Terminated; removing
	// now would delete the tmp workspace it is syncing into.
	if s.activating {
		return fmt.Errorf("session is being started: %s", id)
	}

	sessionsDir := filepath.Join(m.configDir, "sessions")
	deleteErr := deleteSession(sessionsDir, id)
	// A file that is already gone is not a failure: the caller asked for it
	// not to exist, and it doesn't. Surfacing this as an error (as Plan 2
	// originally did) produced a red banner for a delete that did exactly
	// what was asked - the same shape as os.ErrProcessDone in Kill.
	if errors.Is(deleteErr, os.ErrNotExist) {
		deleteErr = nil
	}

	var rmErr error
	if s.UseTmpWorkspace {
		sessionRoot := filepath.Join(m.tmpRoot(), id)
		// G703: Path traversal via taint analysis.
		// sessionRoot is derived from managed ID.
		rmErr = os.RemoveAll(sessionRoot)
	}

	// Always drop the in-memory session so the UI does not wedge on a row it
	// cannot delete, even if the on-disk cleanup below failed.
	delete(m.sessions, id)

	var errs []error
	if deleteErr != nil {
		errs = append(errs, fmt.Errorf("delete session file: %w", deleteErr))
	}
	if rmErr != nil {
		errs = append(errs, fmt.Errorf("remove tmp workspace: %w", rmErr))
	}
	return errors.Join(errs...)
}

// RunningCount reports how many sessions are running or waiting, so a caller
// can announce a Shutdown before it blocks.
func (m *Manager) RunningCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	n := 0
	for _, s := range m.sessions {
		if s.Status != StatusTerminated {
			n++
		}
	}
	return n
}

// Shutdown terminates all managed sessions, waits for activations that were
// already in flight to finish (each notices the shutdown at its final check,
// kills the process it just spawned and removes a new session's tmp copy), and
// stops the watchers. It returns only once no process it could still be
// responsible for is running and no watcher can write to disk any more. It is
// idempotent. It returns how many sessions were running or waiting when it
// started.
//
// An activation that passed its final check before closed was set is already
// Running; Shutdown collects its live list in the same critical section that
// sets closed, so it kills that session like any other.
//
// Sessions are killed concurrently: each Kill may spend a grace period on a
// graceful stop, and doing them one after another would multiply it by the
// session count. Kill comes before the watchers are stopped, so each record is
// persisted as Terminated by Kill itself; the watchers then exit on m.stop
// without their redundant final persist. One consequence: a watcher no longer
// records ExitCode after a Shutdown kill - for a killed process that is just
// the signal, so nothing is lost.
//
// Sessions already Terminated are left alone: their record is current, and
// rewriting every historic record on each quit would be pure churn.
func (m *Manager) Shutdown() int {
	m.mu.Lock()
	m.closed = true
	var live []string
	for id, s := range m.sessions {
		if s.Status != StatusTerminated {
			live = append(live, id)
		}
	}
	m.mu.Unlock()

	var wg sync.WaitGroup
	for _, id := range live {
		wg.Go(func() { _ = m.Kill(id) })
	}
	wg.Wait()

	m.activations.Wait()

	m.stopOnce.Do(func() { close(m.stop) })
	m.watchers.Wait()
	return len(live)
}
