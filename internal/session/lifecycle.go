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
		// Clear existing .agents directories in tmp workspace
		skillsDir := filepath.Join(workingDir, ".agents", "skills")
		if err := os.RemoveAll(skillsDir); err != nil {
			return nil, fmt.Errorf("clear tmp skills dir: %w", err)
		}
		personasDir := filepath.Join(workingDir, ".agents", "agents")
		if err := os.RemoveAll(personasDir); err != nil {
			return nil, fmt.Errorf("clear tmp personas dir: %w", err)
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
	s.Scrollback = fmt.Appendf(s.Scrollback, "\n\n--- Switched agent to %s at %s ---\n\n",
		card.Name, now.Format(time.RFC3339))

	go m.trackOutput(id, h.Output)
	go m.watch(id, d)

	m.persistSessionLocked(s)
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
	} else {
		s.Scrollback = fmt.Appendf(s.Scrollback, "\n\n--- Relaunched at %s ---\n\n",
			now.Format(time.RFC3339))
	}

	go m.trackOutput(sid, h.Output)
	go m.watch(sid, d)

	m.persistSessionLocked(s)
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
	s.Input = h.PTY
	s.Output = h.Output
	s.Done = h.Done
}

func (m *Manager) persistSessionLocked(s *Session) {
	d := &Data{
		ID:              s.ID,
		CardID:          s.CardID,
		Name:            s.Name,
		PID:             s.PID,
		WorkingDir:      s.WorkingDir,
		OriginalDir:     s.OriginalDir,
		UseTmpWorkspace: s.UseTmpWorkspace,
		StartedAt:       s.StartedAt,
		LastWordAt:      s.LastWordAt,
		Status:          s.Status,
	}
	sessionsDir := filepath.Join(m.configDir, "sessions")
	_ = saveSession(sessionsDir, d)
}

// Kill terminates one managed session.
func (m *Manager) Kill(id string) error {
	m.mu.RLock()
	s, ok := m.sessions[id]
	m.mu.RUnlock()
	if !ok {
		return fmt.Errorf("session not found: %s", id)
	}
	if s.Driver != nil {
		if err := s.Driver.Kill(); err != nil {
			return err
		}
	}
	m.mu.Lock()
	s.Status = StatusTerminated
	s.Pinged = false
	m.persistSessionLocked(s)
	m.mu.Unlock()
	return nil
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
	_ = deleteSession(sessionsDir, id)

	if s.UseTmpWorkspace {
		tmpRoot := os.Getenv("KANKURO_TMP_DIR")
		if tmpRoot == "" {
			tmpRoot = filepath.Join(m.configDir, "tmp")
		}
		sessionRoot := filepath.Join(tmpRoot, id)
		// G703: Path traversal via taint analysis.
		// sessionRoot is derived from managed ID.
		_ = os.RemoveAll(sessionRoot) //nolint:gosec // sessionRoot is derived from managed ID
	}

	delete(m.sessions, id)
	return nil
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
