package ui

import (
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"

	"kankuro/internal/config"
)

// CardFormState holds all state for the ID card form view.
type CardFormState struct {
	Inputs              []textinput.Model
	FocusIndex          int
	Puppet              string
	PuppetSelectorIndex int
	PuppetSelectorPrev  string
	EditingID           string

	Editor     textarea.Model
	Content    string
	Path       string
	PreferPath bool
	PathInput  textinput.Model

	SelectedWorkspaceID string
	SelectedSkillIDs    map[string]bool
	SelectedPersonaIDs  map[string]bool
	SelectedMCPIDs      map[string]bool

	SkillCursor     int
	PersonaCursor   int
	MCPCursor       int
	WorkspaceCursor int

	SkillSearchQuery     string
	PersonaSearchQuery   string
	MCPSearchQuery       string
	WorkspaceSearchQuery string

	MCPFormPrevMode viewMode
}

func NewCardFormState() CardFormState {
	return CardFormState{
		Inputs:             make([]textinput.Model, cardInputCount),
		SelectedSkillIDs:   make(map[string]bool),
		SelectedPersonaIDs: make(map[string]bool),
		SelectedMCPIDs:     make(map[string]bool),
		Puppet:             defaultPuppet,
	}
}

// SessionViewState holds session-related UI state.
type SessionViewState struct {
	ActiveID        string
	PendingDeleteID string
	SwitchID        string
	FocusedPane     dashboardPane
}

// NamingState holds the session naming prompt state.
type NamingState struct {
	SessionID  string
	Card       config.IDCard
	Input      textinput.Model
	LaunchDone bool
	LaunchErr  error
	Submitted  bool
}

func NewNamingState() NamingState {
	ni := textinput.New()
	ni.Placeholder = "Session name (Enter to confirm)"
	ni.Width = inputWidthDefault
	return NamingState{Input: ni}
}

// SourceFormState holds generic source form state.
type SourceFormState struct {
	Sources    []string
	Cursor     int
	Input      textinput.Model
	EditingIdx int
}

func NewSourceFormState() SourceFormState {
	ti := textinput.New()
	ti.Width = inputWidthDefault
	return SourceFormState{
		EditingIdx: -1,
		Input:      ti,
	}
}

// WorkspaceFormState holds workspace source form state.
type WorkspaceFormState struct {
	Sources     []config.WorkspaceSource
	Cursor      int
	Inputs      []textinput.Model
	EditingID   string
	SearchQuery string
}

func NewWorkspaceFormState() WorkspaceFormState {
	inputs := make([]textinput.Model, workspaceInputCount)
	inputs[0] = textinput.New()
	inputs[0].Placeholder = "Path to workspace or root directory..."
	inputs[0].Width = inputWidthDefault
	inputs[1] = textinput.New()
	inputs[1].Placeholder = "Is root directory (y/n)"
	inputs[1].Width = inputWidthDefault
	return WorkspaceFormState{Inputs: inputs}
}
