package session

import (
	"bytes"
	"fmt"
	"io"
	"os"
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
)

type Session struct {
	ID          string
	CardID      string
	Name        string
	PID         int
	WorkingDir  string
	StartedAt   time.Time
	LastWordAt  time.Time
	LastOutput  string
	Status      Status
	Pinged      bool
	ExitCode    int
	Driver      driver.AgentDriver
	Input       io.Writer
	Output      <-chan []byte
	Done        <-chan struct{}
	Scrollback  []byte
	PendingLine []byte
	Subscribers map[chan []byte]struct{}
}

type AttachHandle struct {
	Input      io.Writer
	PTY        *os.File
	Output     <-chan []byte
	Scrollback []byte
	Done       <-chan struct{}
	Close      func()
}

const (
	scrollbackLimit  = 256 * 1024
	pendingLineLimit = 8 * 1024
)

var (
	ansiRE = regexp.MustCompile(`\x1b\[[0-9;]*[A-Za-z]`)
	wordRE = regexp.MustCompile(`[a-zA-Z0-9]`)
)

type Snapshot struct {
	ID         string
	CardID     string
	Name       string
	PID        int
	WorkingDir string
	StartedAt  time.Time
	LastWordAt time.Time
	Status     Status
	Pinged     bool
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
	delete(m.sessions, id)
	return nil
}

type Manager struct {
	mu       sync.RWMutex
	sessions map[string]*Session
}

// NewManager creates an empty in-memory session registry.
func NewManager() *Manager {
	return &Manager{sessions: map[string]*Session{}}
}

// Launch starts a new session process from an ID card and begins status tracking.
func (m *Manager) Launch(card config.IDCard) (string, error) {
	wd, err := driver.SanitizeWorkingDir(card.WorkingDir)
	if err != nil {
		return "", err
	}
	d := driver.NewCLIDriver(card.Driver)
	h, err := d.Spawn(driver.AgentConfig{
		Driver:            card.Driver,
		Model:             card.Model,
		WorkingDir:        wd,
		SystemInstruction: card.SystemInstruction,
		Env:               card.Env,
	})
	if err != nil {
		return "", err
	}
	sid := uuid.NewString()
	now := time.Now().UTC()
	pid := 0
	if h.Cmd != nil && h.Cmd.Process != nil {
		pid = h.Cmd.Process.Pid
	}
	s := &Session{
		ID:          sid,
		CardID:      card.ID,
		Name:        card.Name,
		PID:         pid,
		WorkingDir:  wd,
		StartedAt:   now,
		LastWordAt:  now,
		Status:      StatusRunning,
		Driver:      d,
		Input:       h.PTY,
		Output:      h.Output,
		Done:        h.Done,
		Subscribers: make(map[chan []byte]struct{}),
	}
	m.mu.Lock()
	m.sessions[sid] = s
	m.mu.Unlock()

	go m.trackOutput(sid, h.Output)
	go m.watch(sid, d)

	return sid, nil
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
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	for range ticker.C {
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
	ch := make(chan []byte, 128)
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
	if err := s.Driver.Kill(); err != nil {
		return err
	}
	m.mu.Lock()
	s.Status = StatusTerminated
	s.Pinged = false
	m.mu.Unlock()
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
			ID:         s.ID,
			CardID:     s.CardID,
			Name:       s.Name,
			PID:        s.PID,
			WorkingDir: s.WorkingDir,
			StartedAt:  s.StartedAt,
			LastWordAt: s.LastWordAt,
			Status:     s.Status,
			Pinged:     s.Pinged,
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
	if _, ok := s.Subscribers[ch]; ok {
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
