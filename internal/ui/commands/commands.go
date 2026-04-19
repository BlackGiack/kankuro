package commands

import (
	tea "github.com/charmbracelet/bubbletea"

	"kankuro/internal/config"
	"kankuro/internal/session"
)

// CardCommands defines callbacks for card-related operations.
type CardCommands struct {
	// Create saves a new card and returns it with ID populated.
	Create func(card config.IDCard) (config.IDCard, error)

	// Update saves changes to an existing card.
	Update func(card config.IDCard) (config.IDCard, error)

	// Delete removes a card by ID.
	Delete func(cardID string) error

	// UpsertLocal updates the in-memory cards list without persistence.
	// Used after successful Create/Update to sync local state.
	UpsertLocal func(card config.IDCard)
}

// SessionCommands defines callbacks for session operations.
type SessionCommands struct {
	// Launch starts a new session.
	Launch func(card config.IDCard, skills []config.Skill, personas []config.Persona, mcps []config.MCPServer) (string, error)

	// Attach connects to running session.
	Attach func(sessionID string) tea.Cmd

	// Kill terminates a running session.
	Kill func(sessionID string) error

	// Remove deletes a terminated session record.
	Remove func(sessionID string) error

	// SetName updates session display name.
	SetName func(sessionID string, name string) error
}

// DataCommands defines callbacks for data loading.
type DataCommands struct {
	// RefreshCards reloads cards from repository.
	RefreshCards func() []config.IDCard

	// RefreshSessions reloads sessions from manager.
	RefreshSessions func() []session.Snapshot

	// RefreshSkills reloads skills from repository.
	RefreshSkills func() []config.Skill

	// RefreshPersonas reloads personas from repository.
	RefreshPersonas func() []config.Persona
}

// Commands aggregates all command callbacks.
type Commands struct {
	Cards    CardCommands
	Sessions SessionCommands
	Data     DataCommands
}
