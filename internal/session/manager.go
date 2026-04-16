package session

import (
	"fmt"
	"io"
	"maps"
	"os"
	"path/filepath"
	"regexp"
	"sync"
	"time"

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

// WorkspaceReader provides the minimal interface the session manager needs
// to resolve workspace paths when launching sessions.
type WorkspaceReader interface {
	ListWorkspaces() ([]config.Workspace, error)
}

type Manager struct {
	mu        sync.RWMutex
	sessions  map[string]*Session
	configDir string
	repo      WorkspaceReader
	drivers   driver.Registry
}

// Option configures the Manager.
type Option func(*Manager)

// WithDriverRegistry sets a custom driver registry.
func WithDriverRegistry(reg driver.Registry) Option {
	return func(m *Manager) {
		m.drivers = reg
	}
}

// NewManager creates a session registry and loads persisted sessions.
func NewManager(configDir string, repo WorkspaceReader, opts ...Option) *Manager {
	m := &Manager{
		sessions:  make(map[string]*Session),
		configDir: configDir,
		repo:      repo,
		drivers:   driver.NewBuiltinRegistry(), // default builtin registry
	}
	for _, opt := range opts {
		opt(m)
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

// SetName updates the name of an existing session and persists it.
func (m *Manager) SetName(id, name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	s, ok := m.sessions[id]
	if !ok {
		return fmt.Errorf("session not found: %s", id)
	}
	s.Name = name
	m.persistSessionLocked(s)
	return nil
}

// GetScrollback returns a copy of the cached scrollback for a session.
func (m *Manager) GetScrollback(id string) []byte {
	m.mu.RLock()
	defer m.mu.RUnlock()
	s, ok := m.sessions[id]
	if !ok {
		return nil
	}
	out := make([]byte, len(s.Scrollback))
	copy(out, s.Scrollback)
	return out
}

// SubscribeOutput registers a channel to receive live output from a session.
func (m *Manager) SubscribeOutput(id string) (chan []byte, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	s, ok := m.sessions[id]
	if !ok {
		return nil, fmt.Errorf("session not found: %s", id)
	}
	ch := make(chan []byte, outputChanSize)
	s.Subscribers[ch] = struct{}{}
	return ch, nil
}

// UnsubscribeOutput removes a subscriber channel from a session.
func (m *Manager) UnsubscribeOutput(id string, ch chan []byte) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if s, ok := m.sessions[id]; ok {
		delete(s.Subscribers, ch)
	}
}
