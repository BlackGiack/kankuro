package ui

import (
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"

	"kankuro/internal/config"
	"kankuro/internal/session"
)

type viewMode int

type dashboardPane int

const (
	viewDashboard viewMode = iota
	viewCardForm
	viewPuppetSelector
	viewMCPForm
	viewDeleteConfirm
	viewSwitchAgent
	viewAgentsEditor
	viewAgentsPathInput
	viewSkillSources
	viewSkillSourceForm
	viewPersonaSources
	viewPersonaSourceForm
	viewWorkspaces
	viewWorkspaceSourceForm
	viewWorkspacePicker
	viewNaming
	viewIDE
	viewIDEForm
)

func puppetStyles() []string {
	return []string{"default", "kankuro", "cat", "strawhat", "darth", "totoro"}
}

const (
	paneCards dashboardPane = iota
	paneSessions
)

const (
	defaultSkillListHeight = 10
	minSkillListHeight     = 3
	skillListHeightOffset  = 20

	cardInputCount      = 5
	mcpInputCount       = 7
	workspaceInputCount = 2

	cardListMinWidth  = 20
	cardListMinHeight = 10
	cardsListPadding  = 34
	cardsListHeight   = 4

	sessionListWidth     = 32
	sessionsListHeight   = 4
	switchListMinWidth   = 40
	switchListPadding    = 10
	switchListHeightDiff = 6

	inputWidthDefault = 60
	editorWidth       = 80
	editorHeight      = 20

	truncateLength = 28

	keyEsc       = "esc"
	keyEnter     = "enter"
	keyTab       = "tab"
	keyCtrlC     = "ctrl+c"
	keyBackspace = "backspace"
	keyUp        = "up"
	keyDown      = "down"

	faceNeutral = "neutral"

	sessionsDelegateHeight = 4
	cardDelegateHeight     = 8

	pathTruncateLimit = 25
	pathTruncateKeep  = 22

	contentSummaryLimit = 30
	contentSummaryKeep  = 27

	// Focus indices for non-text inputs in the card form.
	cardWorkspaceIndex   = 3
	cardFaceStyleIndex   = 4
	cardPuppetIndex      = 5
	cardUseTmpIndex      = 6
	cardAgentsMDIndex    = 7
	cardSkillListIndex   = 8
	cardPersonaListIndex = 9
	cardMCPListIndex     = 10
	skillWindowBuffer    = 2

	// default puppet style.
	defaultPuppet = "kankuro"

	skillListHeightDivisor = 3

	maxPillsShown = 4
	pillTextLimit = 10
	headerPadding = 2

	infoBoxWidth  = 30
	infoBoxHeight = 4
	faceBoxWidth  = 11
	faceBoxHeight = 5
	cardItemWidth = 42
)

type App struct {
	repo    config.Repository
	session *session.Manager

	mode viewMode

	cards      []config.IDCard
	skills     []config.Skill
	personas   []config.Persona
	mcps       []config.MCPServer
	sessions   []session.Snapshot
	workspaces []config.Workspace

	cardsList    list.Model
	sessionsList list.Model
	switchList   list.Model

	// View-specific state
	cardForm    CardFormState
	sessionView SessionViewState
	naming      NamingState

	// Source management
	skillSources   SourceFormState
	personaSources SourceFormState
	workspaceForm  WorkspaceFormState // Form state for workspace management

	// MCP form (remains separate as it's simple)
	mcpInputs []textinput.Model

	// IDE config
	ideConfig config.IDEConfig
	ideInput  textinput.Model

	help help.Model

	width  int
	height int
	err    error
}
