package session

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"

	"kankuro/internal/config"
	"kankuro/internal/driver"
)

// ErrPersistFailed marks a "started, but its record could not be saved"
// outcome: the agent process is live and the in-memory session is fully
// usable, but the on-disk copy did not update, so it will not survive a
// restart. Callers should use [errors.Is] to tell this apart from a launch
// that never started a process at all, and can still attach/use the session
// while surfacing this as a (non-blocking) warning.
var ErrPersistFailed = errors.New("session state could not be saved")

// Launch starts a new session process from an ID card.
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
	return sid, m.startSession(sid, card, skills, personas, mcps, true, "")
}

// Relaunch restarts a terminated session using its existing workspace (if tmp).
func (m *Manager) Relaunch(id string, card config.IDCard, skills []config.Skill, personas []config.Persona, mcps []config.MCPServer) error {
	m.mu.RLock()
	s, ok := m.sessions[id]
	m.mu.RUnlock()
	if !ok {
		return fmt.Errorf("session not found: %s", id)
	}
	if s.Status != StatusTerminated {
		return errors.New("session is already running")
	}
	return m.startSession(id, card, skills, personas, mcps, false, s.WorkingDir)
}

// SwitchAgent restarts a terminated session using its existing workspace, but with a new agent card.
func (m *Manager) SwitchAgent(
	id string,
	card config.IDCard,
	skills []config.Skill,
	personas []config.Persona,
	mcps []config.MCPServer,
) error {
	m.mu.RLock()
	s, ok := m.sessions[id]
	m.mu.RUnlock()
	if !ok {
		return fmt.Errorf("session not found: %s", id)
	}
	if s.Status != StatusTerminated {
		return errors.New("session must be terminated before switching agents")
	}

	finalWD := s.WorkingDir
	if finalWD == "" {
		return errors.New("session workspace is empty")
	}
	if _, errStat := os.Stat(finalWD); errStat != nil {
		return fmt.Errorf("session workspace missing: %w", errStat)
	}

	// Clear tmp directories if using tmp workspace
	clearTmp := s.UseTmpWorkspace
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
	defer m.mu.Unlock()
	s, ok = m.sessions[id]
	if !ok {
		return fmt.Errorf("session not found: %s", id)
	}

	m.activateSessionLocked(s, card, d, h, pid, now)
	s.PendingAttachHeader = fmt.Appendf(nil, "\r\n--- Switched agent to %s at %s ---\r\n\r\n",
		card.Name, now.Format(time.RFC3339))
	gen := s.generation

	go m.trackOutput(id, gen, h.Output)
	go m.watch(id, gen, d)

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

func (m *Manager) startSession(
	sid string,
	card config.IDCard,
	skills []config.Skill,
	personas []config.Persona,
	mcps []config.MCPServer,
	isNew bool,
	workspaceOverride string,
) error {
	var wd string
	if workspaceOverride != "" {
		wd = workspaceOverride
	} else {
		workspaces, err := m.repo.ListWorkspaces()
		if err != nil {
			return fmt.Errorf("list workspaces: %w", err)
		}
		for _, w := range workspaces {
			if w.ID == card.WorkspaceID {
				wd = w.Path
				break
			}
		}
		if wd == "" {
			return fmt.Errorf("workspace not found: %s", card.WorkspaceID)
		}
	}

	finalWD, errPrep := m.prepareWorkspace(sid, card, wd, isNew)
	if errPrep != nil {
		return errPrep
	}

	// Sync driver files (clear generated content for tmp workspaces on every launch)
	clearTmp := card.UseTmpWorkspace
	// startSession should append to agents.md if it's a new session and using tmp workspace
	appendAgentsMD := card.UseTmpWorkspace && isNew
	d, err := m.syncDriverFiles(card, skills, personas, mcps, finalWD, clearTmp, appendAgentsMD)
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
	go m.watch(sid, gen, d)

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

func (m *Manager) activateSessionLocked(
	s *Session,
	card config.IDCard,
	d driver.AgentDriver,
	h *driver.SessionHandle,
	pid int,
	now time.Time,
) {
	s.generation++
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
	m.mu.RUnlock()
	if !ok {
		return fmt.Errorf("session not found: %s", id)
	}
	var killErr error
	if s.Driver != nil {
		killErr = s.Driver.Kill()
	}
	m.mu.Lock()
	s.Status = StatusTerminated
	s.Pinged = false
	persistErr := m.persistSessionLockedErr(s)
	m.mu.Unlock()
	return errors.Join(killErr, persistErr)
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
		tmpRoot := os.Getenv("KANKURO_TMP_DIR")
		if tmpRoot == "" {
			tmpRoot = filepath.Join(m.configDir, "tmp")
		}
		sessionRoot := filepath.Join(tmpRoot, id)
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

// Shutdown terminates all managed sessions, used when Kankuro exits.
func (m *Manager) Shutdown() {
	m.mu.RLock()
	ids := make([]string, 0, len(m.sessions))
	for id := range m.sessions {
		ids = append(ids, id)
	}
	m.mu.RUnlock()
	for _, id := range ids {
		_ = m.Kill(id)
	}
}
