package messages

import (
	"github.com/BlackGiack/kankuro/internal/config"
	"github.com/BlackGiack/kankuro/internal/session"
)

// ViewType represents the current view mode.
type ViewType int

const (
	ViewDashboard ViewType = iota
	ViewCardEditor
	ViewSourceManagerSkills
	ViewSourceManagerPersonas
	ViewSourceManagerWorkspaces
	ViewSourceManagerPuppets
	ViewMCPForm
	ViewSettings
	ViewSessionLauncher
	ViewAgentSwitch
)

// ModalType represents modal overlays.
type ModalType int

const (
	ModalNone ModalType = iota
	ModalDeleteConfirm
	ModalSwitchAgent
	ModalPuppetSelector
	ModalWorkspacePicker
	ModalAgentsMDEditor
)

// ===== User Action Messages (Components → App) =====

// SessionPreviewRequestedMsg signals that a specific session should be previewed immediately.
type SessionPreviewRequestedMsg struct {
	SessionID string
}

// SessionOutputMsg signals that new output has been received for a session.
type SessionOutputMsg struct {
	SessionID string
	Output    []byte
}

// DataRefreshRequestedMsg asks the coordinator to run refreshData(). It exists
// so that callbacks running inside an already-async tea.Cmd closure (e.g. a
// save/delete handler) can hand the actual refresh back to Update, which runs
// on the main Bubble Tea goroutine. Building and immediately invoking
// refreshData() inside such a closure would touch coordinator-goroutine state
// (c.views, c.preview, ...) from the wrong goroutine.
type DataRefreshRequestedMsg struct{}

// PuppetsSyncedMsg signals that a background puppet-source sync has finished.
// Like DataRefreshRequestedMsg, it defers the actual refreshData() call to
// Update instead of running it inside the async closure that performed the
// sync.
type PuppetsSyncedMsg struct{}

// SessionAttachRequestedMsg signals user wants to attach to session.
// Warning carries a non-fatal problem the caller should still surface (e.g.
// session.ErrPersistFailed): the session started/relaunched successfully and
// should still be attached to, but something about the operation - usually
// that its state could not be saved - deserves the user's attention.
type SessionAttachRequestedMsg struct {
	SessionID string
	Warning   error
}

// SessionLaunchCompletedMsg signals that a session launch has completed.
type SessionLaunchCompletedMsg struct {
	SessionID string
	Err       error
}

// AgentSwitchRequestedMsg signals user wants to switch agent for a session.
type AgentSwitchRequestedMsg struct {
	SessionID   string
	SessionName string
}

// MCPDeleteRequestedMsg signals user wants to delete an MCP from the card editor.
type MCPDeleteRequestedMsg struct {
	MCPID string
}

// MCPDeletedMsg signals that an MCP has been deleted.
type MCPDeletedMsg struct {
	MCPID string
}

// AttachFinishedMsg signals that an attach session has finished.
type AttachFinishedMsg struct {
	SessionID string
	Err       error
}

// IDELaunchedMsg signals that a requested IDE launch has completed (or
// failed). Spawning the editor process happens inside a tea.Cmd, off the UI
// goroutine - see sessionActions.launchIDE.
type IDELaunchedMsg struct {
	Err error
}

// TerminalLaunchedMsg reports completion of a terminal launch request.
type TerminalLaunchedMsg struct {
	Err error
}

// AgentSwitchedMsg signals that an agent-switch request has completed (or
// failed). Err may wrap session.ErrPersistFailed: the agent process did
// restart, but its on-disk record failed to save - callers should still
// attach, but surface the warning. Syncing driver files and spawning the new
// PTY process happens inside a tea.Cmd, off the UI goroutine - see
// sessionActions.switchAgent.
type AgentSwitchedMsg struct {
	SessionID string
	Err       error
}

// NavigateMsg signals view change request.
type NavigateMsg struct {
	View ViewType
}

// NavigateBackMsg signals the user wants to return to the previous view.
type NavigateBackMsg struct{}

// CardSavedMsg signals that an ID card has been saved.
type CardSavedMsg struct {
	Card config.IDCard
	Err  error
}

// MCPSavedMsg signals that an MCP server has been saved.
type MCPSavedMsg struct {
	MCP config.MCPServer
	Err error
}

// SettingsSavedMsg signals that application settings have been saved.
type SettingsSavedMsg struct {
	Settings config.AppSettings
	Err      error
}

// ErrorMsg signals an error occurred.
type ErrorMsg struct {
	Err error
}

// ClearErrorMsg signals to clear the current error.
type ClearErrorMsg struct{}

// DataLoadedMsg contains a full data snapshot from the repository. It always
// represents a complete refresh: fields left nil mean "that registry could
// not be read this round, keep whatever the receiver already has" (see Err),
// never "this update doesn't concern that field". Partial, preview-only
// updates use PreviewDataMsg instead - see debt_impl_plan_4.md step 2b for why
// mixing the two in one message type caused silent data loss.
type DataLoadedMsg struct {
	Cards      []config.IDCard
	Sessions   []session.Snapshot
	Skills     []config.Skill
	Personas   []config.Persona
	MCPs       []config.MCPServer
	Workspaces []config.Workspace
	Puppets    []config.Puppet

	// PreviewSessionID/PreviewScrollback carry the current preview's
	// scrollback alongside a full refresh. The coordinator forwards these to
	// PreviewReceiver.SetPreviewData rather than handing them to
	// DataReceiver.SetData directly.
	PreviewSessionID  string
	PreviewScrollback []byte

	// Err reports any error hit while reading one or more registries. When
	// set, the corresponding field(s) above are left nil rather than
	// overwritten with an empty slice, so a transient read failure cannot
	// blank data a receiver already has.
	Err error
}

// DataReceiver is an interface for components that can receive a full data
// snapshot.
type DataReceiver interface {
	SetData(msg DataLoadedMsg)
}

// SessionsRefreshedMsg carries just the (in-memory) session list plus the
// current preview session ID/scrollback - the idle tick's sessions-only
// refresh. It is deliberately a distinct type from DataLoadedMsg, not a
// DataLoadedMsg with only Sessions set: broadcasting the latter to every
// DataReceiver used to reach config-registry views (e.g. the source
// managers), whose SetData unconditionally reloads their source files from
// disk - see debt_impl_plan_5.md review finding 1. This message is instead
// routed only to the dashboard, the sole consumer of Sessions.
type SessionsRefreshedMsg struct {
	Sessions          []session.Snapshot
	PreviewSessionID  string
	PreviewScrollback []byte
}

// PreviewDataMsg carries an update to the currently previewed session's
// scrollback only. It is a distinct type from DataLoadedMsg so a receiver can
// never mistake a partial preview update for a full snapshot.
type PreviewDataMsg struct {
	SessionID     string
	Scrollback    []byte
	IsIncremental bool
}

// PreviewReceiver is implemented by views that render a live session
// preview and need incremental scrollback updates.
type PreviewReceiver interface {
	SetPreviewData(msg PreviewDataMsg)
}
