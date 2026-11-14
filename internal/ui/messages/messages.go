package messages

import (
	"github.com/BlackGiack/kankuro/internal/api/contracts"
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
	ViewMCPManager
	ViewSettings
	ViewSessionLauncher
	ViewAgentSwitch
	ViewConversationChoice
	ViewDashboardList
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

// SessionSelectedMsg signals a change in the selected session.
type SessionSelectedMsg struct{ SessionID string }

// DataRefreshRequestedMsg asks the coordinator to run refreshData(). It exists
// so that callbacks running inside an already-async tea.Cmd closure (e.g. a
// save/delete handler) can hand the actual refresh back to Update, which runs
// on the main Bubble Tea goroutine. Building and immediately invoking
// refreshData() inside such a closure would touch coordinator-goroutine state
// (c.views, c.center, ...) from the wrong goroutine.
type DataRefreshRequestedMsg struct{}

// DashboardDeletionPreparedMsg carries a human-only deletion confirmation.
type DashboardDeletionPreparedMsg struct {
	Request, Visit              uint64
	DashboardID                 string
	Token                       string
	Summary                     []string
	RequiresForeignConfirmation bool
	Err                         error
}

// DashboardDeletionStatusMsg reports recoverable per-item cascade outcomes.
type DashboardDeletionStatusMsg struct {
	Request, Visit uint64
	DashboardID    string
	Summary        []string
	Err            error
}

// CardDeletionPreparedMsg carries a revision-bound card deletion impact.
type CardDeletionPreparedMsg struct {
	Request, Visit              uint64
	DashboardID                 string
	Generation                  uint64
	CardID                      string
	Token                       string
	Summary                     []string
	RequiresForeignConfirmation bool
	Err                         error
}

// DashboardOpenRequestedMsg asks the coordinator to open a captured identity.
type DashboardOpenRequestedMsg struct {
	DashboardID    string
	DashboardName  string
	DashboardState contracts.DashboardState
	Generation     uint64
}

type DashboardCardAction string

const (
	DashboardCardLaunch DashboardCardAction = "launch"
)

// DashboardCardLoadedMsg carries a full card only to the legacy action adapter
// that owns launching.
type DashboardCardLoadedMsg struct {
	Request     uint64
	DashboardID string
	Generation  uint64
	CardID      string
	Action      DashboardCardAction
	Card        contracts.CardSummary
	Err         error
}

// DashboardListRequestedMsg returns to the dashboard list without stopping sessions.
type DashboardListRequestedMsg struct {
	DashboardID string
	Generation  uint64
}

// DashboardListLoadedMsg carries a refreshed human dashboard inventory.
type DashboardListLoadedMsg struct {
	Dashboards []contracts.DashboardSummary
	Err        error
	RequestID  uint64
}

// DashboardMutationFinishedMsg reports a save or cascade completion.
type DashboardMutationFinishedMsg struct {
	Request, Visit uint64
	Progress       contracts.DeletionProgress
	Err            error
	Generation     uint64
}

// ScopedActionFinishedMsg carries a result bound to the originating view.
type ScopedActionFinishedMsg struct {
	Request, Visit uint64
	Progress       contracts.DeletionProgress
	DashboardID    string
	Generation     uint64
	Err            error
}

const (
	AttachRequestLaunch    = "launch"
	AttachRequestLifecycle = "lifecycle"
)

// ScopedAttachReadyMsg carries an authorized attachment subscription.
type ScopedAttachReadyMsg struct {
	Sequence    uint64
	Request     uint64
	RequestKind string
	DashboardID string
	Generation  uint64
	SessionID   string
	Handle      contracts.Attachment
	Err         error
}

// ScopedSessionNamedMsg reports a session rename before optional attachment.
type ScopedSessionNamedMsg struct {
	Request     uint64
	DashboardID string
	Generation  uint64
	Name        string
	SessionID   string
	Result      contracts.RenameSessionResult
}

// DashboardSessionRenamedMsg reports a rename initiated from the dashboard.
type DashboardSessionRenamedMsg struct {
	Request     uint64
	DashboardID string
	Generation  uint64
	SessionID   string
	Result      contracts.RenameSessionResult
}

type WorkspaceOpenedMsg struct {
	DashboardID, SessionID string
	Generation, Request    uint64
	Kind                   contracts.WorkspaceCommandKind
	Result                 contracts.OpenSessionWorkspaceResult
}

type SessionStoppedMsg struct {
	DashboardID, SessionID string
	Generation, Request    uint64
	Result                 contracts.StopSessionResult
}
type SessionRemovedMsg struct {
	DashboardID, SessionID string
	Generation, Request    uint64
	Result                 contracts.RemoveSessionResult
}

// SessionAttachRequestedMsg signals user wants to attach to session.
// Warning carries a non-fatal problem to surface during navigation.
type SessionAttachRequestedMsg struct {
	SessionID   string
	Warning     error
	DashboardID string
	Generation  uint64
}

// SessionLaunchCompletedMsg signals that a session launch has completed.
type SessionLaunchCompletedMsg struct {
	CardID      string
	Request     uint64
	Result      contracts.LifecycleResult
	DashboardID string
	Generation  uint64
}

// AgentSwitchRequestedMsg signals user wants to switch agent for a session.
type AgentSwitchRequestedMsg struct {
	SessionID   string
	SessionName string
	DashboardID string
	Generation  uint64
}

type ConversationPreparedMsg struct {
	SessionID, CardID          string
	DashboardID                string
	Generation, Epoch, Request uint64
	ReturnView                 ViewType
	Initial                    bool
	Preparation                contracts.ConversationPreparation
}
type SessionActivationCompletedMsg struct {
	DashboardID, SessionID, CardID string
	Generation, Request            uint64
	Result                         contracts.LifecycleResult
}

type MCPManagerScope struct {
	DashboardID string
	Generation  uint64
	Visit       uint64
}
type MCPManagerEditMsg struct {
	MCPID string
	Scope MCPManagerScope
}
type MCPManagerDeleteMsg struct {
	MCPID string
	Scope MCPManagerScope
}
type MCPManagerDeletedMsg struct {
	MCPID       string
	Err         error
	DashboardID string
	Generation  uint64
	Request     uint64
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
	Sequence    uint64
	SessionID   string
	Err         error
	DashboardID string
	Generation  uint64
}

// ScopedNavigateMsg is ignored after the originating dashboard view changes.
type ScopedNavigateMsg struct {
	View        ViewType
	DashboardID string
	Generation  uint64
}

// NavigateBackMsg signals the user wants to return to the previous view.
type NavigateBackMsg struct{}

// EditorCancelMsg ties cancellation to the editor that requested it.
type EditorCancelMsg struct {
	DashboardID                string
	Generation, Epoch, Request uint64
}

// EditorNavigateMsg ties source navigation to the editor that requested it.
type EditorNavigateMsg struct {
	View              ViewType
	DashboardID       string
	Generation, Epoch uint64
}

// CardEditorLoadedMsg carries a detached editor read and its request identity.
type CardEditorLoadedMsg struct {
	Choices                    contracts.CardEditorChoices
	DashboardID, CardID        string
	Generation, Epoch, Request uint64
	Data                       contracts.CardEditorData
	Err                        error
}

type CardSavedMsg struct {
	Result                     contracts.CardSaveResult
	Err                        error
	DashboardID, CardID        string
	Generation, Epoch, Request uint64
}

// MCPSavedMsg signals that an MCP server has been saved.
type MCPSavedMsg struct {
	Request     uint64
	MCP         contracts.MCPEditor
	Err         error
	DashboardID string
	Generation  uint64
}

type MCPEditorLoadedMsg struct {
	MCP                 contracts.MCPEditor
	Err                 error
	DashboardID         string
	Generation, Request uint64
}

type SettingsLoadedMsg struct {
	Settings contracts.Settings
	Err      error
	Request  uint64
}

// SettingsSavedMsg signals that application settings have been saved.
type SettingsSavedMsg struct {
	Visit       uint64
	Draft       uint64
	Request     uint64
	Settings    contracts.Settings
	Err         error
	DashboardID string
	Generation  uint64
}

// ErrorMsg signals an error occurred.
type ErrorMsg struct {
	Err error
}

// ClearErrorMsg signals to clear the current error.
type ClearErrorMsg struct{}

// MCPServerFailedMsg reports an unexpected failure of the local MCP endpoint.
type MCPServerFailedMsg struct{ Err error }

// DataLoadedMsg contains focused inventory reads. Failed reads preserve receiver data.
type DataLoadedMsg struct {
	DashboardID        string
	InventoryRequestID uint64
	RegistryErr        error
	Generation         uint64
	SessionEpoch       uint64
	MCPRows            []contracts.MCPRow
	Inventory          contracts.ReadResult[contracts.DashboardInventory]
	Choices            contracts.DisplayChoicesResult
	EditorChoices      contracts.CardEditorChoices

	// Err reports read failures; individual read results distinguish failure from empty.
	Err error
}

// DataReceiver is an interface for components that can receive a full data
// snapshot.
type DataReceiver interface {
	SetData(msg DataLoadedMsg)
}

// SessionsRefreshedMsg carries scoped cards and sessions with the
// originating dashboard and generation. Only their consumers receive this update.
type SessionsRefreshedMsg struct {
	DashboardID        string
	InventoryRequestID uint64
	Generation         uint64
	SessionEpoch       uint64
	Err                error
	Inventory          contracts.ReadResult[contracts.DashboardInventory]
	Choices            contracts.DisplayChoicesResult
	EditorChoices      contracts.CardEditorChoices
}
