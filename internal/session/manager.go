package session

import (
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

type Manager struct {
	mu        sync.RWMutex
	sessions  map[string]*Session
	configDir string
	repo      config.Repository
}

// NewManager creates a session registry and loads persisted sessions.
func NewManager(configDir string, repo config.Repository) *Manager {
	m := &Manager{
		sessions:  make(map[string]*Session),
		configDir: configDir,
		repo:      repo,
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
