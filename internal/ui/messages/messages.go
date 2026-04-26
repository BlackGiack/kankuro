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
	ViewMCPForm
	ViewIDEConfig
	ViewSessionLauncher
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

// ===== Data Loaded Messages (App → Components) =====

// CardsLoadedMsg signals that cards have been loaded/refreshed.
type CardsLoadedMsg struct {
	Cards []config.IDCard
}

// SessionsLoadedMsg signals that sessions have been loaded/refreshed.
type SessionsLoadedMsg struct {
	Sessions []session.Snapshot
}

// SessionScrollbackLoadedMsg signals that a session's scrollback has been loaded.
type SessionScrollbackLoadedMsg struct {
	SessionID  string
	Scrollback []byte
}

// SkillsLoadedMsg signals that skills have been loaded/refreshed.
type SkillsLoadedMsg struct {
	Skills []config.Skill
}

// PersonasLoadedMsg signals that personas have been loaded/refreshed.
type PersonasLoadedMsg struct {
	Personas []config.Persona
}

// MCPsLoadedMsg signals that MCPs have been loaded/refreshed.
type MCPsLoadedMsg struct {
	MCPs []config.MCPServer
}

// WorkspacesLoadedMsg signals that workspaces have been loaded/refreshed.
type WorkspacesLoadedMsg struct {
	Workspaces []config.Workspace
}

// ===== User Action Messages (Components → App) =====

// CardSelectedMsg signals user selected a card in the list.
type CardSelectedMsg struct {
	Card config.IDCard
}

// CardEditRequestedMsg signals user wants to edit a card.
type CardEditRequestedMsg struct {
	Card config.IDCard
}

// CardDeleteRequestedMsg signals user wants to delete a card.
type CardDeleteRequestedMsg struct {
	CardID string
}

// CardCreatedMsg signals a new card was created.
type CardCreatedMsg struct {
	Card config.IDCard
}

// CardUpdatedMsg signals a card was updated.
type CardUpdatedMsg struct {
	Card config.IDCard
}

// SessionSelectedMsg signals user selected a session.
type SessionSelectedMsg struct {
	Session session.Snapshot
}

// SessionAttachRequestedMsg signals user wants to attach to session.
type SessionAttachRequestedMsg struct {
	SessionID string
}

// SessionLaunchCompletedMsg signals that a session launch has completed.
type SessionLaunchCompletedMsg struct {
	SessionID string
	Err       error
}

// MCPDeleteRequestedMsg signals user wants to delete an MCP from the card editor.
type MCPDeleteRequestedMsg struct {
	MCPID string
}

// MCPDeletedMsg signals that an MCP has been deleted.
type MCPDeletedMsg struct {
	MCPID string
}

// MCPUpdatedMsg signals that an MCP has been created or updated.
type MCPUpdatedMsg struct {
	MCP config.MCPServer
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

// ShowModalMsg signals modal should be shown.
type ShowModalMsg struct {
	Modal ModalType
}

// CloseModalMsg signals modal should be closed.
type CloseModalMsg struct{}

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

	PreviewSessionID  string
	PreviewScrollback []byte
}

// DataReceiver is an interface for components that can receive data updates.
type DataReceiver interface {
	SetData(msg DataLoadedMsg)
}
