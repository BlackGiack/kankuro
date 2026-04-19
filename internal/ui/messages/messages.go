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
	ViewSourceManager
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

// SkillsLoadedMsg signals that skills have been loaded/refreshed.
type SkillsLoadedMsg struct {
	Skills []config.Skill
}

// PersonasLoadedMsg signals that personas have been loaded/refreshed.
type PersonasLoadedMsg struct {
	Personas []config.Persona
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

// NavigateMsg signals view change request.
type NavigateMsg struct {
	View ViewType
}

// ShowModalMsg signals modal should be shown.
type ShowModalMsg struct {
	Modal ModalType
}

// CloseModalMsg signals modal should be closed.
type CloseModalMsg struct{}
