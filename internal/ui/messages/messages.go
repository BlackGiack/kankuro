package messages

import (
	"kankuro/internal/config"
	"kankuro/internal/session"
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

// DataLoadedMsg contains all loaded data from the repository.
type DataLoadedMsg struct {
	Cards      []config.IDCard
	Sessions   []session.Snapshot
	Skills     []config.Skill
	Personas   []config.Persona
	MCPs       []config.MCPServer
	Workspaces []config.Workspace
	Puppets    []config.Puppet

	PreviewSessionID  string
	PreviewScrollback []byte
	IsIncremental     bool
}

// DataReceiver is an interface for components that can receive data updates.
type DataReceiver interface {
	SetData(msg DataLoadedMsg)
}
