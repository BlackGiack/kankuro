package ui

import (
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textarea"
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
	skillListHeightOffset  = 18

	cardInputCount = 6
	mcpInputCount  = 2

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

	cardInputTotalOffset = 3
	// Focus indices for non-text inputs in the card form.
	cardPuppetIndex      = 5
	cardUseTmpIndex      = 6
	cardAgentsMDIndex    = 7
	cardSkillListIndex   = 8
	cardPersonaListIndex = 9
	skillWindowBuffer    = 2

	// default puppet style.
	defaultPuppet = "kankuro"

	skillListHeightDivisor = 2

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

	cards    []config.IDCard
	skills   []config.Skill
	personas []config.Persona
	mcps     []config.MCPServer
	sessions []session.Snapshot

	cardsList    list.Model
	sessionsList list.Model
	switchList   list.Model

	cardInputs []textinput.Model
	mcpInputs  []textinput.Model

	agentsEditor       textarea.Model
	agentsMD           string
	agentsMDPath       string
	agentsMDPreferPath bool
	agentsPathInput    textinput.Model

	skillSources            []string
	skillSourceCursor       int
	skillSourceInput        textinput.Model
	editingSkillSourceIndex int

	personaSources            []string
	personaSourceCursor       int
	personaSourceInput        textinput.Model
	editingPersonaSourceIndex int

	cardFocusIndex     int
	skillCursor        int
	selectedSkillIDs   map[string]bool
	skillSearchQuery   string
	personaCursor      int
	selectedPersonaIDs map[string]bool
	personaSearchQuery string

	editingCardID   string
	activeSessionID string
	pendingDeleteID string
	switchSessionID string
	focusedPane     dashboardPane

	// Puppet selection state for the card form
	cardPuppet          string
	puppetSelectorIndex int
	puppetSelectorPrev  string

	help help.Model

	width  int
	height int
	err    error
}
