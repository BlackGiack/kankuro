package session

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"maps"
	"os"
	"path/filepath"
	"regexp"
	"sync"
	"time"

	"github.com/BlackGiack/kankuro/internal/config"
	"github.com/BlackGiack/kankuro/internal/driver"
)

type Status = driver.SessionStatus

const (
	StatusRunning    = driver.StatusRunning
	StatusWaiting    = driver.StatusWaiting
	StatusTerminated = driver.StatusTerminated

	ScrollbackLimit  = 256 * 1024
	pendingLineLimit = 8 * 1024
	outputChanSize   = 128
	watchInterval    = 2 * time.Second
)

type Session struct {
	ID                            string
	CardID                        string
	Name                          string
	PID                           int
	WorkingDir                    string
	OriginalDir                   string
	UseTmpWorkspace               bool
	SubWorkspaces                 []SubWorkspaceDescriptor
	SubWorkspaceMetadataAvailable bool
	KankuroMCPEnabled             bool
	DashboardControl              bool
	// InstructionOwnership tracks managed content independently per instruction file.
	InstructionOwnership    map[string]bool
	InstructionsManaged     bool
	StartedAt               time.Time
	LastWordAt              time.Time
	LastOutput              string
	Status                  Status
	LifecycleActivity       string
	Pinged                  bool
	ExitCode                int
	Conversations           map[string][]ConversationBinding
	Registration            RegistrationState
	RegistrationPersistence RegistrationPersistence
	RegistrationAt          time.Time
	RegistrationErr         string
	ActiveConversationID    string
	ActivationID            string
	ActivationToken         string
	ExpectedConversationID  string
	Driver                  driver.AgentDriver
	// PTY is the raw pseudo-terminal file, used only for ioctl-based resize
	// (Attach hands it out as AttachHandle.PTY). Keystrokes must NOT be
	// written directly to this fd - it races Kill, which closes and nils the
	// driver's own copy under lock. Use Driver.SendInputBytes for writes; see
	// debt_impl_plan_4.md step 4 and session/attach.go's driverInputWriter.
	PTY         *os.File
	Output      <-chan []byte
	Done        <-chan struct{}
	Scrollback  []byte
	PendingLine []byte
	Subscribers map[chan []byte]struct{}
	// PendingAttachHeader is a one-shot banner (e.g. "Relaunched at ...")
	// shown to the next attacher and then cleared, so it doesn't re-appear on
	// subsequent reattaches via scrollback replay.
	PendingAttachHeader []byte
	// generation is bumped on every activation (Launch/Relaunch/SwitchAgent).
	// It lets goroutines started by a previous activation detect that they
	// are stale and stop touching the session.
	generation uint64
	// activating is set, under m.mu, by Relaunch/SwitchAgent between passing
	// the "must be terminated" guard and activateSessionLocked (or failure).
	// It is what makes that guard exclusive: the unlocked middle of an
	// activation (workspace sync, Spawn) is long, and Status stays
	// Terminated throughout it.
	activating        bool
	stopping          bool
	deleting          bool
	activeSubmissions int
	submissionsDone   chan struct{}
}

type AttachHandle struct {
	Input      io.Writer
	PTY        *os.File
	Output     <-chan []byte
	Scrollback []byte
	// Header is a one-shot banner (e.g. a relaunch / agent-switch marker) to
	// emit before the scrollback. It is consumed by the manager when the
	// handle is created and will not appear on subsequent attaches.
	Header []byte
	Done   <-chan struct{}
	Close  func()
}

var (
	ansiRE = regexp.MustCompile(`\x1b\[[0-9;]*[A-Za-z]`)
	wordRE = regexp.MustCompile(`[a-zA-Z0-9]`)
)

type Snapshot struct {
	ID                            string
	CardID                        string
	Name                          string
	PID                           int
	WorkingDir                    string
	OriginalDir                   string
	UseTmpWorkspace               bool
	SubWorkspaces                 []SubWorkspaceDescriptor
	SubWorkspaceMetadataAvailable bool
	KankuroMCPEnabled             bool
	DashboardControl              bool
	InstructionOwnership          map[string]bool
	InstructionsManaged           bool
	StartedAt                     time.Time
	LastWordAt                    time.Time
	LastOutput                    string
	Status                        Status
	LifecycleActivity             string
	Pinged                        bool
	ExitCode                      int
	Conversations                 map[string][]ConversationBinding
	Registration                  RegistrationState
	RegistrationPersistence       RegistrationPersistence
	RegistrationErr               string
	ActiveConversationID          string
	ActivationID                  string
}

// ConversationBinding identifies one native CLI conversation within a card
// and Kankuro session. Names are display metadata, never identity.
type ConversationBinding struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Driver string `json:"driver"`
}

type RegistrationState string

const (
	RegistrationUnavailable RegistrationState = "unavailable"
	RegistrationPending     RegistrationState = "pending"
	RegistrationUnconfirmed RegistrationState = "unconfirmed"
	RegistrationRegistered  RegistrationState = "registered"
	RegistrationFailed      RegistrationState = "failed"
)

type RegistrationPersistence string

// ConversationChoice is the explicit activation decision shown by the UI.
// New starts without a native ID; ID resumes one binding for the destination card.
type ConversationChoice struct {
	New bool
	ID  string
}

const (
	RegistrationNotAttempted RegistrationPersistence = "not_attempted"
	RegistrationSaved        RegistrationPersistence = "saved"
	RegistrationUnsaved      RegistrationPersistence = "unsaved"
)

// WorkspaceReader provides the minimal interface the session manager needs
// to resolve workspace paths when launching sessions.
type WorkspaceReader interface {
	ListWorkspaces() ([]config.Workspace, error)
}

type Manager struct {
	mu               sync.RWMutex
	pendingCardRefs  map[string]int
	subWorkspaceGate chan struct{}
	sessions         map[string]*Session
	launching        map[string]struct{}
	configDir        string
	repo             WorkspaceReader
	workspaceCopier  WorkspaceCopier
	drivers          driver.Registry
	loadErr          error
	logger           *slog.Logger
	messages         *MessageStore

	// stop is closed by Shutdown to make every watch goroutine return without
	// a final tick; watchers counts them so Shutdown can wait. closed (under
	// mu) refuses new activations once Shutdown has started, so no watcher can
	// be added after the Wait begins.
	//
	// Only watch is tied to these: it is the one per-activation goroutine that
	// writes to disk. trackOutput touches only memory (scrollback and
	// subscribers) and ends when the driver closes its output channel, so it
	// needs no such handle - do not add one for symmetry.
	//
	// activations counts Launch/Relaunch/SwitchAgent calls in progress, from
	// the moment they pass the !closed check (under mu, with the Add in the
	// same critical section) until they return. Shutdown waits for them: an
	// activation is mostly unlocked work (workspace copy, Spawn), and one that
	// notices closed only afterwards must still get to kill its own process
	// and remove its tmp copy before the app exits.
	stop              chan struct{}
	shutdownStart     chan struct{}
	startupWait       func(context.Context, <-chan struct{}, <-chan struct{}, <-chan []byte) ([]byte, error)
	stopOnce          sync.Once
	watchers          sync.WaitGroup
	activations       sync.WaitGroup
	stops             sync.WaitGroup
	shutdownDone      chan struct{}
	shutdownCount     int
	closed            bool
	authorities       map[string]*activationAuthority
	messageAdmissions map[*MessageAdmission]struct{}
	mcpEndpoint       string
}

// Option configures the Manager.
type Option func(*Manager)

// WorkspaceCopier prepares a temporary main-workspace copy. It receives the
// activation context so cancellation can interrupt preparation.
type WorkspaceCopier func(ctx context.Context, source, destination string) error

// WithWorkspaceCopier injects the temporary-workspace copy operation.
func WithWorkspaceCopier(copier WorkspaceCopier) Option {
	return func(manager *Manager) { manager.workspaceCopier = copier }
}

// WithStartupWait overrides the post-spawn wait, primarily for deterministic tests.
func WithStartupWait(wait func(context.Context, <-chan struct{}, <-chan struct{}, <-chan []byte) ([]byte, error)) Option {
	return func(m *Manager) { m.startupWait = wait }
}

// WithStartupDelay sets the startup pause; zero is useful for tests.
func WithStartupDelay(delay time.Duration) Option {
	return WithStartupWait(func(ctx context.Context, shutdown, done <-chan struct{}, output <-chan []byte) ([]byte, error) {
		return waitForStartupDuration(ctx, shutdown, done, output, delay)
	})
}

// WithDriverRegistry sets a custom driver registry.
func WithDriverRegistry(reg driver.Registry) Option {
	return func(m *Manager) {
		m.drivers = reg
	}
}

// WithLogger sets the logger used for background failures that have no
// caller to report them to (e.g. a persist failure from the watch ticker).
// Defaults to a discarding logger, so callers that don't care about this
// need not wire anything up.
func WithLogger(logger *slog.Logger) Option {
	return func(m *Manager) {
		if logger != nil {
			m.logger = logger
		}
	}
}

// NewManager creates a session registry and loads persisted sessions.
func NewManager(configDir string, repo WorkspaceReader, opts ...Option) *Manager {
	m := &Manager{
		sessions:          make(map[string]*Session),
		pendingCardRefs:   make(map[string]int),
		subWorkspaceGate:  make(chan struct{}, 1),
		launching:         make(map[string]struct{}),
		configDir:         configDir,
		repo:              repo,
		drivers:           driver.NewBuiltinRegistry(), // default builtin registry
		logger:            slog.New(slog.DiscardHandler),
		stop:              make(chan struct{}),
		shutdownStart:     make(chan struct{}),
		startupWait:       waitForStartup,
		authorities:       make(map[string]*activationAuthority),
		messageAdmissions: make(map[*MessageAdmission]struct{}),
	}
	for _, opt := range opts {
		opt(m)
	}

	sessionsDir := filepath.Join(configDir, "sessions")
	loaded, loadErr := loadSessions(sessionsDir)
	m.loadErr = loadErr

	for _, d := range loaded {
		m.sessions[d.ID] = d.restore()
	}
	messageStore, messageErr := NewMessageStore(configDir)
	if messageErr != nil {
		m.loadErr = errors.Join(m.loadErr, messageErr)
	} else {
		m.messages = messageStore
		if recoverErr := messageStore.Recover(m.sessions); recoverErr != nil {
			m.loadErr = errors.Join(m.loadErr, recoverErr)
		}
	}

	return m
}

// Messages returns the manager-owned durable messaging store shared by all frontends.
func (m *Manager) Messages() *MessageStore { return m.messages }

// SetMCPEndpoint sets the loopback endpoint before any activation is started.
func (m *Manager) SetMCPEndpoint(endpoint string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.mcpEndpoint = endpoint
}

// LoadErrors returns any error encountered while hydrating persisted
// sessions at startup (corrupt or unreadable session files are skipped, not
// fatal). Callers should surface this once at startup.
//
// m.loadErr is intentionally read without m.mu: it is written exactly once,
// in NewManager, before the Manager is returned to (and can escape to) any
// other goroutine - so there is a happens-before edge and no lock is needed.
// It must stay write-once; do not add another writer without adding
// synchronization.
func (m *Manager) LoadErrors() error {
	return m.loadErr
}

// GetSession returns a safe value snapshot of a session by ID.
func (m *Manager) GetSession(id string) (Snapshot, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	s, ok := m.sessions[id]
	if !ok {
		return Snapshot{}, false
	}
	return s.snapshot(), true
}

// PendingLineLen reports the buffered partial-line size. Used by tests to
// assert the pendingLineLimit cap.
func (m *Manager) PendingLineLen(id string) (int, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	s, ok := m.sessions[id]
	if !ok {
		return 0, false
	}
	return len(s.PendingLine), true
}

// BroadcastOutput delivers raw output to a session's subscribers. It exists
// for tests that simulate live process output without a real driver.
func (m *Manager) BroadcastOutput(id string, data []byte) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	s, ok := m.sessions[id]
	if !ok {
		return false
	}
	broadcast(data, s.Subscribers)
	return true
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
		out = append(out, s.snapshot())
	}
	return out
}

// snapshot builds a value Snapshot from a Session. Callers must hold m.mu
// (read or write) while calling this.
func (s *Session) snapshot() Snapshot {
	activity := "idle"
	switch {
	case s.activating:
		activity = "activating"
	case s.stopping:
		activity = "stopping"
	case s.deleting:
		activity = "deleting"
	}
	return Snapshot{
		ID:                            s.ID,
		CardID:                        s.CardID,
		Name:                          s.Name,
		PID:                           s.PID,
		WorkingDir:                    s.WorkingDir,
		OriginalDir:                   s.OriginalDir,
		UseTmpWorkspace:               s.UseTmpWorkspace,
		SubWorkspaces:                 append([]SubWorkspaceDescriptor(nil), s.SubWorkspaces...),
		SubWorkspaceMetadataAvailable: s.SubWorkspaceMetadataAvailable,
		KankuroMCPEnabled:             s.KankuroMCPEnabled,
		DashboardControl:              s.DashboardControl,
		InstructionsManaged:           s.InstructionsManaged,
		InstructionOwnership:          sessionInstructionOwnership(s),
		StartedAt:                     s.StartedAt,
		LastWordAt:                    s.LastWordAt,
		LastOutput:                    s.LastOutput,
		Status:                        s.Status,
		LifecycleActivity:             activity,
		Pinged:                        s.Pinged,
		ExitCode:                      s.ExitCode,
		Conversations:                 cloneConversations(s.Conversations),
		Registration:                  s.Registration,
		RegistrationPersistence:       s.RegistrationPersistence,
		RegistrationErr:               s.RegistrationErr,
		ActiveConversationID:          s.ExpectedConversationID,
		ActivationID:                  s.ActivationID,
	}
}

// data builds the persisted Data record from a Session. Callers must hold
// m.mu (read or write) while calling this.
func (s *Session) data() *Data {
	return &Data{
		ID:                            s.ID,
		CardID:                        s.CardID,
		Name:                          s.Name,
		PID:                           s.PID,
		WorkingDir:                    s.WorkingDir,
		OriginalDir:                   s.OriginalDir,
		UseTmpWorkspace:               s.UseTmpWorkspace,
		SubWorkspaces:                 append([]SubWorkspaceDescriptor(nil), s.SubWorkspaces...),
		SubWorkspaceMetadataAvailable: s.SubWorkspaceMetadataAvailable,
		KankuroMCPEnabled:             s.KankuroMCPEnabled,
		DashboardControl:              s.DashboardControl,
		InstructionsManaged:           s.InstructionsManaged,
		InstructionOwnership:          sessionInstructionOwnership(s),
		StartedAt:                     s.StartedAt,
		LastWordAt:                    s.LastWordAt,
		LastOutput:                    s.LastOutput,
		Status:                        s.Status,
		ExitCode:                      s.ExitCode,
		Conversations:                 cloneConversations(s.Conversations),
	}
}

func cloneConversations(in map[string][]ConversationBinding) map[string][]ConversationBinding {
	if in == nil {
		return nil
	}
	out := make(map[string][]ConversationBinding, len(in))
	for cardID, bindings := range in {
		out[cardID] = append([]ConversationBinding(nil), bindings...)
	}
	return out
}

func cloneInstructionOwnership(in map[string]bool) map[string]bool {
	if in == nil {
		return nil
	}
	out := make(map[string]bool, len(in))
	maps.Copy(out, in)
	return out
}

func sessionInstructionOwnership(s *Session) map[string]bool {
	ownership := cloneInstructionOwnership(s.InstructionOwnership)
	if ownership == nil {
		ownership = make(map[string]bool)
	}
	// Older session records only stored the generic driver's ownership flag.
	if s.InstructionsManaged && len(s.InstructionOwnership) == 0 {
		ownership["AGENTS.md"] = true
	}
	return ownership
}

func updateInstructionOwnership(previous map[string]bool, filename string, managed bool) map[string]bool {
	updated := cloneInstructionOwnership(previous)
	if updated == nil {
		updated = make(map[string]bool)
	}
	if managed {
		updated[filename] = true
	} else {
		delete(updated, filename)
	}
	return updated
}

// SetName updates the name of an existing session and persists it.
func (m *Manager) SetName(id, name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	s, ok := m.sessions[id]
	if !ok {
		return fmt.Errorf("session not found: %s", id)
	}
	if s.deleting || s.stopping || s.activating {
		return fmt.Errorf("%w: rename is not allowed during a lifecycle operation", ErrLifecycleConflict)
	}
	s.Name = name
	return m.persistSessionLockedErr(s)
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
