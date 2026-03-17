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
	viewMCPForm
	viewDeleteConfirm
	viewSwitchAgent
	viewAgentsEditor
	viewAgentsPathInput
	viewSkillSources
	viewSkillSourceForm
)

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

	faceNeutral = "neutral"

	sessionsDelegateHeight = 4
	cardDelegateHeight     = 7

	pathTruncateLimit = 25
	pathTruncateKeep  = 22

	contentSummaryLimit = 30
	contentSummaryKeep  = 27

	cardInputTotalOffset = 2
	skillWindowBuffer    = 2

	maxSkillsSummary = 3

	infoBoxWidth  = 30
	infoBoxHeight = 3
	faceBoxWidth  = 7
	faceBoxHeight = 3
	cardItemWidth = 40
)

type App struct {
	repo    config.Repository
	session *session.Manager

	mode viewMode

	cards    []config.IDCard
	skills   []config.Skill
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

	cardFocusIndex   int
	skillCursor      int
	selectedSkillIDs map[string]bool
	skillSearchQuery string

	editingCardID   string
	activeSessionID string
	pendingDeleteID string
	switchSessionID string
	focusedPane     dashboardPane

	help help.Model

	width  int
	height int
	err    error
}
