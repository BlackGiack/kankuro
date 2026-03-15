package session

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"maps"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sync"
	"time"

	"github.com/google/uuid"

	"kankuro/internal/config"
	"kankuro/internal/driver"
)

type Status = driver.SessionStatus

const (
	StatusRunning    = driver.StatusRunning
	StatusWaiting    = driver.StatusWaiting
	StatusTerminated = driver.StatusTerminated

	scrollbackLimit  = 256 * 1024
	pendingLineLimit = 8 * 1024
	outputChanSize   = 128
	watchInterval    = 2 * time.Second
)

type Session struct {
	ID              string
	CardID          string
	Name            string
	PID             int
	WorkingDir      string
	OriginalDir     string
	UseTmpWorkspace bool
	StartedAt       time.Time
	LastWordAt      time.Time
	LastOutput      string
	Status          Status
	Pinged          bool
	ExitCode        int
	Driver          driver.AgentDriver
	Input           io.Writer
	Output          <-chan []byte
	Done            <-chan struct{}
	Scrollback      []byte
	PendingLine     []byte
	Subscribers     map[chan []byte]struct{}
}

type AttachHandle struct {
	Input      io.Writer
	PTY        *os.File
	Output     <-chan []byte
	Scrollback []byte
	Done       <-chan struct{}
	Close      func()
}

var (
	ansiRE = regexp.MustCompile(`\x1b\[[0-9;]*[A-Za-z]`)
	wordRE = regexp.MustCompile(`[a-zA-Z0-9]`)
)

type Snapshot struct {
	ID              string
	CardID          string
	Name            string
	PID             int
	WorkingDir      string
	UseTmpWorkspace bool
	StartedAt       time.Time
	LastWordAt      time.Time
	Status          Status
	Pinged          bool
}

type Manager struct {
	mu        sync.RWMutex
	sessions  map[string]*Session
	configDir string
}

// NewManager creates a session registry and loads persisted sessions.
func NewManager(configDir string) *Manager {
	m := &Manager{
		sessions:  make(map[string]*Session),
		configDir: configDir,
	}

	sessionsDir := filepath.Join(configDir, "sessions")
	loaded, _ := loadSessions(sessionsDir) // ignore error, just start empty

	for _, d := range loaded {
		m.sessions[d.ID] = &Session{
			ID:              d.ID,
			CardID:          d.CardID,
			Name:            d.Name,
			PID:             d.PID,
			WorkingDir:      d.WorkingDir,
			OriginalDir:     d.OriginalDir,
			UseTmpWorkspace: d.UseTmpWorkspace,
			StartedAt:       d.StartedAt,
			LastWordAt:      d.LastWordAt,
			Status:          StatusTerminated, // All loaded sessions are terminated
			Subscribers:     make(map[chan []byte]struct{}),
		}
	}

	return m
}

// Sessions returns a thread-safe snapshot of all sessions.
func (m *Manager) Sessions() map[string]*Session {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make(map[string]*Session, len(m.sessions))
	maps.Copy(out, m.sessions)
	return out
}

// GetSession returns a session by ID.
func (m *Manager) GetSession(id string) (*Session, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	s, ok := m.sessions[id]
	return s, ok
}

// SetLastWordAt is used for testing word silence detection.
func (m *Manager) SetLastWordAt(id string, t time.Time) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if s, ok := m.sessions[id]; ok {
		s.LastWordAt = t
	}
}

// Launch starts a new session process from an ID card.
func (m *Manager) Launch(card config.IDCard, skills []config.Skill) (string, error) {
	sid := uuid.NewString()
	return sid, m.startSession(sid, card, skills, true)
}

// Relaunch restarts a terminated session using its existing workspace (if tmp).
func (m *Manager) Relaunch(id string, card config.IDCard, skills []config.Skill) error {
	m.mu.RLock()
	s, ok := m.sessions[id]
	m.mu.RUnlock()
	if !ok {
		return fmt.Errorf("session not found: %s", id)
	}
	if s.Status != StatusTerminated {
		return errors.New("session is already running")
	}
	return m.startSession(id, card, skills, false)
}

// SwitchAgent restarts a terminated session using its existing workspace, but with a new agent card.
func (m *Manager) SwitchAgent(id string, card config.IDCard, skills []config.Skill) error {
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

	if s.UseTmpWorkspace {
		skillsDir := filepath.Join(finalWD, ".agents", "skills")
		if errRemove := os.RemoveAll(skillsDir); errRemove != nil {
			return fmt.Errorf("clear tmp skills dir: %w", errRemove)
		}
		if errSync := syncAgentsMD(finalWD, card, false); errSync != nil {
			return errSync
		}
	}

	d := driver.NewCLIDriver(card.Driver)
	if errSkills := d.SetSkills(skills, finalWD); errSkills != nil {
		return errSkills
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

	s.CardID = card.ID
	s.Name = card.Name
	s.PID = pid
	s.StartedAt = now
	s.LastWordAt = now
	s.Status = StatusRunning
	s.Pinged = false
	s.Driver = d
	s.Input = h.PTY
	s.Output = h.Output
	s.Done = h.Done
	s.Scrollback = fmt.Appendf(s.Scrollback, "\n\n--- Switched agent to %s at %s ---\n\n",
		card.Name, now.Format(time.RFC3339))

	go m.trackOutput(id, h.Output)
	go m.watch(id, d)

	m.persistSessionLocked(s)
	return nil
}

func syncAgentsMD(finalWD string, card config.IDCard, appendIfExist bool) error {
	agentsMD := card.AgentsMD
	if card.AgentsMDPreferPath && card.AgentsMDPath != "" {
		data, errRead := os.ReadFile(card.AgentsMDPath)
		if errRead != nil {
			return fmt.Errorf("read agents md path %s: %w", card.AgentsMDPath, errRead)
		}
		agentsMD = string(data)
	}

	if agentsMD == "" {
		return nil
	}

	path := filepath.Join(finalWD, "AGENTS.md")

	if !appendIfExist {
		// G703: Path traversal via taint analysis.
		// Path constructed using filepath.Join with trusted parts or verified paths.
		//nolint:gosec // Path is controlled by session logic
		if errWrite := os.WriteFile(path, []byte(agentsMD), 0o600); errWrite != nil {
			return fmt.Errorf("write AGENTS.md: %w", errWrite)
		}
		return nil
	}

	// G703: Path traversal via taint analysis.
	//nolint:gosec // Path is controlled by session logic
	f, errOpen := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
	if errOpen != nil {
		return fmt.Errorf("open AGENTS.md: %w", errOpen)
	}
	defer f.Close()

	stat, errStat := f.Stat()
	if errStat != nil {
		return fmt.Errorf("stat AGENTS.md: %w", errStat)
	}

	content := agentsMD
	if stat.Size() > 0 {
		content = "\n\n" + agentsMD
	}

	if _, errWrite := f.WriteString(content); errWrite != nil {
		return fmt.Errorf("write AGENTS.md: %w", errWrite)
	}
	return nil
}

func (m *Manager) prepareWorkspace(sid string, card config.IDCard, wd string, isNew bool) (string, error) {
	if !card.UseTmpWorkspace {
		return wd, nil
	}

	tmpRoot := os.Getenv("KANKURO_TMP_DIR")
	if tmpRoot == "" {
		tmpRoot = filepath.Join(m.configDir, "tmp")
	}
	sessionRoot := filepath.Join(tmpRoot, sid)
	targetDir := filepath.Join(sessionRoot, filepath.Base(wd))

	if isNew {
		// G703: Path traversal via taint analysis.
		if err := os.MkdirAll(sessionRoot, 0o700); err != nil { //nolint:gosec // sessionRoot is derived from managed ID
			return "", fmt.Errorf("create tmp session dir: %w", err)
		}
		// G702: Command injection via taint analysis.
		// Used for local copy with verified paths.
		cmd := exec.CommandContext(context.Background(), "cp", "-r", wd, targetDir) //nolint:gosec // cp is safe here
		if out, errCopy := cmd.CombinedOutput(); errCopy != nil {
			return "", fmt.Errorf("copy workspace: %s: %w", string(out), errCopy)
		}
	} else {
		// G703: Path traversal via taint analysis.
		if _, errStat := os.Stat(targetDir); errStat != nil { //nolint:gosec // targetDir is derived from managed ID
			return "", fmt.Errorf("tmp workspace missing: %w", errStat)
		}
	}

	skillsDir := filepath.Join(targetDir, ".agents", "skills")
	// G703: Path traversal via taint analysis.
	// This is intentional as the path is derived from managed IDs.
	if errRemove := os.RemoveAll(skillsDir); errRemove != nil { //nolint:gosec // skillsDir is derived from managed path
		return "", fmt.Errorf("clear tmp skills dir: %w", errRemove)
	}
	if errSync := syncAgentsMD(targetDir, card, isNew); errSync != nil {
		return "", errSync
	}
	return targetDir, nil
}

func (m *Manager) startSession(sid string, card config.IDCard, skills []config.Skill, isNew bool) error {
	wd, errWD := driver.SanitizeWorkingDir(card.WorkingDir)
	if errWD != nil {
		return errWD
	}

	finalWD, errPrep := m.prepareWorkspace(sid, card, wd, isNew)
	if errPrep != nil {
		return errPrep
	}

	d := driver.NewCLIDriver(card.Driver)
	if errSkills := d.SetSkills(skills, finalWD); errSkills != nil {
		return errSkills
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

	s.CardID = card.ID
	s.Name = card.Name
	s.PID = pid
	s.WorkingDir = finalWD
	s.OriginalDir = wd
	s.UseTmpWorkspace = card.UseTmpWorkspace
	s.StartedAt = now
	s.LastWordAt = now
	s.Status = StatusRunning
	s.Driver = d
	s.Input = h.PTY
	s.Output = h.Output
	s.Done = h.Done
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

func (m *Manager) trackOutput(sessionID string, output <-chan []byte) {
	for chunk := range output {
		m.mu.Lock()
		s, ok := m.sessions[sessionID]
		if !ok {
			m.mu.Unlock()
			return
		}
		now := time.Now().UTC()
		s.Scrollback = appendScrollback(s.Scrollback, chunk)
		broadcast(chunk, s.Subscribers)
		updateSessionOutputState(s, chunk, now)
		if s.Status != StatusTerminated {
			if s.Pinged {
				s.Status = StatusWaiting
			} else {
				s.Status = StatusRunning
			}
		}
		m.mu.Unlock()
	}
	m.mu.Lock()
	m.closeSubscribersLocked(sessionID)
	m.mu.Unlock()
}

func (m *Manager) watch(sessionID string, d driver.AgentDriver) {
	ticker := time.NewTicker(watchInterval)
	defer ticker.Stop()
	for {
		<-ticker.C
		status := d.GetStatus()
		m.mu.Lock()
		s, ok := m.sessions[sessionID]
		if !ok {
			m.mu.Unlock()
			return
		}
		if status == driver.StatusTerminated {
			s.Status = StatusTerminated
			s.Pinged = false
			m.persistSessionLocked(s)
			m.closeSubscribersLocked(sessionID)
			m.mu.Unlock()
			return
		}
		if driver.IsPromptLike(s.LastOutput, s.LastWordAt, time.Now().UTC()) {
			s.Status = StatusWaiting
			s.Pinged = true
		} else {
			s.Status = StatusRunning
			s.Pinged = false
		}
		m.mu.Unlock()
	}
}

// Attach returns an attach handle with scrollback replay and live output stream.
func (m *Manager) Attach(id string) (*AttachHandle, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	s, ok := m.sessions[id]
	if !ok {
		return nil, fmt.Errorf("session not found: %s", id)
	}
	if s.Status == StatusTerminated {
		return nil, fmt.Errorf("session terminated: %s", id)
	}
	ch := make(chan []byte, outputChanSize)
	s.Subscribers[ch] = struct{}{}
	ptyFile, _ := s.Input.(*os.File)
	return &AttachHandle{
		Input:      s.Input,
		PTY:        ptyFile,
		Output:     ch,
		Scrollback: bytes.Clone(s.Scrollback),
		Done:       s.Done,
		Close: func() {
			m.removeSubscriber(id, ch)
		},
	}, nil
}

// Detach validates that a session exists while leaving it running in background.
func (m *Manager) Detach(id string) error {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if _, ok := m.sessions[id]; !ok {
		return fmt.Errorf("session not found: %s", id)
	}
	return nil
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

// List returns lightweight snapshots for dashboard rendering.
func (m *Manager) List() []Snapshot {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]Snapshot, 0, len(m.sessions))
	for _, s := range m.sessions {
		out = append(out, Snapshot{
			ID:              s.ID,
			CardID:          s.CardID,
			Name:            s.Name,
			PID:             s.PID,
			WorkingDir:      s.WorkingDir,
			UseTmpWorkspace: s.UseTmpWorkspace,
			StartedAt:       s.StartedAt,
			LastWordAt:      s.LastWordAt,
			Status:          s.Status,
			Pinged:          s.Pinged,
		})
	}
	return out
}

func (m *Manager) removeSubscriber(sessionID string, ch chan []byte) {
	m.mu.Lock()
	defer m.mu.Unlock()
	s, ok := m.sessions[sessionID]
	if !ok {
		return
	}
	if _, okSub := s.Subscribers[ch]; okSub {
		delete(s.Subscribers, ch)
		close(ch)
	}
}

func (m *Manager) closeSubscribersLocked(sessionID string) {
	s, ok := m.sessions[sessionID]
	if !ok {
		return
	}
	for ch := range s.Subscribers {
		close(ch)
		delete(s.Subscribers, ch)
	}
}

func appendScrollback(existing, chunk []byte) []byte {
	if len(chunk) >= scrollbackLimit {
		return bytes.Clone(chunk[len(chunk)-scrollbackLimit:])
	}
	if len(existing)+len(chunk) <= scrollbackLimit {
		return append(existing, chunk...)
	}
	trim := len(existing) + len(chunk) - scrollbackLimit
	next := make([]byte, len(existing)-trim, scrollbackLimit)
	copy(next, existing[trim:])
	return append(next, chunk...)
}

func broadcast(chunk []byte, subscribers map[chan []byte]struct{}) {
	if len(subscribers) == 0 {
		return
	}
	for ch := range subscribers {
		data := bytes.Clone(chunk)
		select {
		case ch <- data:
		default:
		}
	}
}

func updateSessionOutputState(s *Session, chunk []byte, now time.Time) {
	s.PendingLine = append(s.PendingLine, chunk...)
	if len(s.PendingLine) > pendingLineLimit {
		s.PendingLine = append([]byte(nil), s.PendingLine[len(s.PendingLine)-pendingLineLimit:]...)
	}

	if wordRE.Match(chunk) {
		s.LastWordAt = now
	}

	for {
		idx := bytes.IndexByte(s.PendingLine, '\n')
		if idx == -1 {
			break
		}
		line := sanitizeLine(s.PendingLine[:idx])
		s.PendingLine = append([]byte(nil), s.PendingLine[idx+1:]...)
		if line == "" {
			continue
		}
		s.LastOutput = line
		s.Pinged = driver.IsPromptLike(line, s.LastWordAt, now)
	}

	sanitized := sanitizeLine(s.PendingLine)
	if sanitized != "" {
		s.LastOutput = sanitized
		s.Pinged = driver.IsPromptLike(sanitized, s.LastWordAt, now)
	}
}

func sanitizeLine(line []byte) string {
	clean := ansiRE.ReplaceAll(bytes.Clone(line), nil)
	clean = bytes.ReplaceAll(clean, []byte("\r"), nil)
	return string(bytes.TrimSpace(clean))
}
