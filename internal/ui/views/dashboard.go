package views

import (
	"errors"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"kankuro/internal/config"
	"kankuro/internal/session"
	"kankuro/internal/ui/commands"
	"kankuro/internal/ui/component"
	"kankuro/internal/ui/components"
	"kankuro/internal/ui/messages"
)

// Pane identifiers.
type Pane int

// UI constants.
const (
	cardsListPadding = 2
)

const (
	PaneCards Pane = iota
	PaneSessions
)

// SearchState holds search/filter state.
type SearchState struct {
	CardsQuery    string
	SessionsQuery string
	IsActive      bool
}

// PendingAction holds pending modal actions.
type PendingAction struct {
	Type      string // "delete-session", "delete-card", "switch"
	SessionID string
	CardID    string
}

// DashboardCallbacks defines callbacks for dashboard actions.
type DashboardCallbacks struct {
	OnEditCard func(card config.IDCard) tea.Cmd

	OnLaunchCard func(card config.IDCard) tea.Cmd

	OnAttach   func(sessionID string) tea.Cmd
	OnRelaunch func(sessionID string) tea.Cmd
	OnKill     func(sessionID string) (tea.Cmd, error)
	OnRemove   func(sessionID string) (tea.Cmd, error)
	OnSwitch   func(sessionID string, newCardID string) (tea.Cmd, error)

	OnDeleteCard     func(cardID string) (tea.Cmd, error)
	OnOpenMCP        func() tea.Cmd
	OnOpenWorkspaces func() tea.Cmd
	OnOpenIDE        func() tea.Cmd
	OnOpenCardEditor func() tea.Cmd
	OnLaunchIDE      func(sessionID string) error
}

// Dashboard is the main composite view showing cards, preview, and sessions.
type Dashboard struct {
	component.Base

	cardList    *components.CardList
	sessionList *components.SessionList
	preview     *components.SessionPreview

	deleteModal *components.Modal
	switchModal *SwitchModal

	focusedPane Pane
	search      SearchState
	pending     PendingAction

	callbacks DashboardCallbacks
	cmds      *commands.Commands

	// Data (passed from App)
	cards    []config.IDCard
	sessions []session.Snapshot
	skills   []config.Skill
	personas []config.Persona
	mcps     []config.MCPServer

	// Layout (computed from size)
	cardsWidth    int
	sessionsWidth int
	previewWidth  int
}

// NewDashboard creates a new dashboard view.
func NewDashboard(callbacks DashboardCallbacks, cmds *commands.Commands) *Dashboard {
	preview := components.NewSessionPreview()

	deleteModal := components.NewModal("Confirm Delete")
	switchModal := NewSwitchModal()

	d := &Dashboard{
		cardList:    nil,
		sessionList: nil,
		preview:     preview,
		deleteModal: deleteModal,
		switchModal: switchModal,
		focusedPane: PaneCards,
		callbacks:   callbacks,
		cmds:        cmds,
	}

	d.cardList = components.NewCardList(components.CardListCallbacks{
		OnSelect: d.callbacks.OnLaunchCard,
		OnEdit:   d.callbacks.OnEditCard,
		OnDelete: func(_ string) {
			// Parent handles via message
		},
	})

	d.sessionList = components.NewSessionList(components.SessionListCallbacks{
		OnSelect: func(_ session.Snapshot) tea.Cmd {
			return nil
		},
		OnAttach: func(_ string) {
			// Parent handles via message
		},
		OnKill:    func(_ string) {},
		OnDelete:  func(_ string) {},
		OnSwitch:  func(_ string) {},
		OnPreview: func(sessionID string) tea.Cmd {
			return func() tea.Msg {
				return messages.SessionPreviewRequestedMsg{SessionID: sessionID}
			}
		},
	})

	return d
}

// SetData updates all child components with fresh data.
func (d *Dashboard) SetData(msg messages.DataLoadedMsg) {
	if msg.Cards != nil {
		d.cards = msg.Cards
		d.cardList.SetCards(d.cards, d.skills, d.personas)
	}
	if msg.Sessions != nil {
		d.sessions = msg.Sessions
		d.sessionList.SetSessions(d.sessions, d.cards)
	}
	if msg.Skills != nil {
		d.skills = msg.Skills
		if msg.Cards != nil {
			d.cardList.SetCards(d.cards, d.skills, d.personas)
		}
	}
	if msg.Personas != nil {
		d.personas = msg.Personas
		if msg.Cards != nil {
			d.cardList.SetCards(d.cards, d.skills, d.personas)
		}
	}
	if msg.MCPs != nil {
		d.mcps = msg.MCPs
	}

	d.syncPreviewSelection()

	// Update scrollback if it matches current preview
	if msg.PreviewSessionID != "" && msg.PreviewSessionID == d.preview.SessionID() {
		if msg.IsIncremental {
			d.preview.AppendScrollback(msg.PreviewScrollback)
		} else {
			d.preview.SetScrollback(msg.PreviewScrollback)
		}
	}
}

func (d *Dashboard) PreviewSessionID() string {
	return d.preview.SessionID()
}

func (d *Dashboard) syncPreviewSelection() {
	// If there are no sessions at all, clear preview.
	if len(d.sessions) == 0 {
		d.preview.Clear()
		return
	}

	// Check if current preview session still exists
	currentStillExists := false
	var currentSess *session.Snapshot
	if d.preview.SessionID() != "" {
		for i := range d.sessions {
			if d.sessions[i].ID == d.preview.SessionID() {
				currentStillExists = true
				currentSess = &d.sessions[i]
				break
			}
		}
	}

	// If we are NOT in sessions pane AND current preview still exists, just refresh it.
	if d.focusedPane != PaneSessions && currentStillExists {
		d.preview.SetSession(currentSess, findCardByID(d.cards, currentSess.CardID))
		return
	}

	// Otherwise, try to follow selection.
	selected, ok := d.sessionList.Selected()
	if !ok {
		// If nothing selected but we have sessions, pick first
		if len(d.sessions) > 0 {
			selected = d.sessions[0]
		} else {
			d.preview.Clear()
			return
		}
	}

	d.preview.SetSession(&selected, findCardByID(d.cards, selected.CardID))
}

// FocusedPane returns the currently focused pane.
func (d *Dashboard) FocusedPane() Pane {
	return d.focusedPane
}

// Cards returns the underlying cards.
func (d *Dashboard) Cards() []config.IDCard {
	return d.cards
}

// Sessions returns the underlying sessions.
func (d *Dashboard) Sessions() []session.Snapshot {
	return d.sessions
}

// TogglePane switches between cards and sessions.
func (d *Dashboard) TogglePane() {
	d.togglePane()
}

// Search returns the current search state.
func (d *Dashboard) Search() SearchState {
	return d.search
}

// DeleteModal returns the delete confirmation modal.
func (d *Dashboard) DeleteModal() *components.Modal {
	return d.deleteModal
}

// SessionList returns the session list component.
func (d *Dashboard) SessionList() *components.SessionList {
	return d.sessionList
}

// SwitchModal returns the switch agent modal.
func (d *Dashboard) SwitchModal() *SwitchModal {
	return d.switchModal
}

// Preview returns the session preview component.
func (d *Dashboard) Preview() *components.SessionPreview {
	return d.preview
}

// SyncPreviewSelection synchronizes preview with selection.
func (d *Dashboard) SyncPreviewSelection() {
	d.syncPreviewSelection()
}

// CardsWidth returns the width of the cards pane.
func (d *Dashboard) CardsWidth() int {
	return d.cardsWidth
}

// PreviewWidth returns the width of the preview pane.
func (d *Dashboard) PreviewWidth() int {
	return d.previewWidth
}

// Init initializes the dashboard.
func (d *Dashboard) Init() tea.Cmd {
	return tea.Batch(
		d.cardList.Init(),
		d.sessionList.Init(),
		d.preview.Init(),
	)
}

// Update handles messages.
func (d *Dashboard) Update(msg tea.Msg) (component.Model, tea.Cmd) {
	// Handle modals first
	if d.deleteModal.IsOpen() {
		return d.updateDeleteModal(msg)
	}
	if d.switchModal.IsOpen() {
		return d.updateSwitchModal(msg)
	}

	// Handle key events
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		if cmd, handled := d.handleKey(keyMsg); handled {
			return d, cmd
		}
	}

	// Route to focused pane
	return d.updateFocusedPane(msg)
}

// handleKey processes global dashboard key events.
func (d *Dashboard) handleKey(keyMsg tea.KeyMsg) (tea.Cmd, bool) {
	// Handle search input
	if handled := d.handleSearch(keyMsg); handled {
		return nil, true
	}

	switch keyMsg.String() {
	case "q", keyCtrlC:
		return tea.Quit, true
	case keyEsc:
		return nil, true
	case "tab", "shift+tab", "left", "right":
		d.togglePane()
		return nil, true
	case "n":
		return d.handleNewCard()
	case "e":
		return d.handleEdit()
	case "m":
		return d.handleOpenMCP()
	case "w":
		return d.handleOpenWorkspaces()
	case "i":
		return d.handleOpenIDE()
	case "o":
		return d.handleLaunchIDE()
	case keyEnter:
		return d.handleEnter()
	case "a":
		return d.handleAttach()
	case "x":
		return d.handleKill()
	case "backspace", "delete":
		return d.handleDelete()
	case "s":
		return d.handleSwitch()
	}

	return nil, false
}

func (d *Dashboard) handleNewCard() (tea.Cmd, bool) {
	if d.callbacks.OnOpenCardEditor != nil {
		return d.callbacks.OnOpenCardEditor(), true
	}
	return nil, true
}

func (d *Dashboard) handleEdit() (tea.Cmd, bool) {
	if d.focusedPane == PaneCards {
		if card, ok := d.cardList.Selected(); ok {
			if d.callbacks.OnEditCard != nil {
				return d.callbacks.OnEditCard(card), true
			}
		}
	}
	return nil, true
}

func (d *Dashboard) handleOpenMCP() (tea.Cmd, bool) {
	if d.callbacks.OnOpenMCP != nil {
		return d.callbacks.OnOpenMCP(), true
	}
	return nil, true
}

func (d *Dashboard) handleOpenWorkspaces() (tea.Cmd, bool) {
	if d.callbacks.OnOpenWorkspaces != nil {
		return d.callbacks.OnOpenWorkspaces(), true
	}
	return nil, true
}

func (d *Dashboard) handleOpenIDE() (tea.Cmd, bool) {
	if d.callbacks.OnOpenIDE != nil {
		return d.callbacks.OnOpenIDE(), true
	}
	return nil, true
}
func (d *Dashboard) handleLaunchIDE() (tea.Cmd, bool) {
	if d.focusedPane != PaneSessions {
		return nil, true
	}

	sess, ok := d.sessionList.Selected()
	if !ok {
		return nil, true
	}

	if d.callbacks.OnLaunchIDE == nil {
		return nil, true
	}

	err := d.callbacks.OnLaunchIDE(sess.ID)
	if err != nil {
		// Return error message
		return func() tea.Msg {
			return messages.ErrorMsg{Err: err}
		}, true
	}
	return nil, true
}

// handleSearch processes search input.
func (d *Dashboard) handleSearch(keyMsg tea.KeyMsg) bool {
	key := keyMsg.String()

	if key == "/" && !d.search.IsActive {
		d.search.IsActive = true
		return true
	}

	if !d.search.IsActive {
		return false
	}

	switch key {
	case keyEsc:
		d.cancelSearch()
		return true
	case keyEnter:
		d.search.IsActive = false
		d.syncPreviewSelection()
		return true
	case "backspace":
		d.handleSearchBackspace()
		return true
	}

	if len(keyMsg.Runes) > 0 {
		d.handleSearchRune(keyMsg.Runes[0])
		return true
	}

	return false
}

func (d *Dashboard) cancelSearch() {
	d.search.CardsQuery = ""
	d.search.SessionsQuery = ""
	d.search.IsActive = false
	d.cardList.SetSearchQuery("")
	d.sessionList.SetSearchQuery("")
	d.syncPreviewSelection()
}

func (d *Dashboard) handleSearchBackspace() {
	if d.focusedPane == PaneCards && len(d.search.CardsQuery) > 0 {
		d.search.CardsQuery = d.search.CardsQuery[:len(d.search.CardsQuery)-1]
		d.cardList.SetSearchQuery(d.search.CardsQuery)
	} else if d.focusedPane == PaneSessions && len(d.search.SessionsQuery) > 0 {
		d.search.SessionsQuery = d.search.SessionsQuery[:len(d.search.SessionsQuery)-1]
		d.sessionList.SetSearchQuery(d.search.SessionsQuery)
		d.syncPreviewSelection()
	}
}

func (d *Dashboard) handleSearchRune(ch rune) {
	if ch >= 33 && ch <= 126 {
		if d.focusedPane == PaneCards {
			d.search.CardsQuery += string(ch)
			d.cardList.SetSearchQuery(d.search.CardsQuery)
		} else {
			d.search.SessionsQuery += string(ch)
			d.sessionList.SetSearchQuery(d.search.SessionsQuery)
			d.syncPreviewSelection()
		}
	}
}

// handleEnter processes enter key based on focused pane.
func (d *Dashboard) handleEnter() (tea.Cmd, bool) {
	if d.focusedPane == PaneSessions {
		return d.handleSessionEnter()
	}

	if card, ok := d.cardList.Selected(); ok {
		if d.callbacks.OnLaunchCard != nil {
			return d.callbacks.OnLaunchCard(card), true
		}
	}
	return nil, true
}

func (d *Dashboard) handleSessionEnter() (tea.Cmd, bool) {
	sess, ok := d.sessionList.Selected()
	if !ok {
		return nil, true
	}

	if sess.Status == session.StatusTerminated {
		if d.callbacks.OnRelaunch != nil {
			return d.callbacks.OnRelaunch(sess.ID), true
		}
		return nil, true
	}

	if d.callbacks.OnAttach != nil {
		return d.callbacks.OnAttach(sess.ID), true
	}

	return nil, true
}

// handleAttach attaches to selected session.
func (d *Dashboard) handleAttach() (tea.Cmd, bool) {
	if d.focusedPane != PaneSessions {
		return nil, false
	}

	sess, ok := d.sessionList.Selected()
	if !ok || sess.Status == session.StatusTerminated {
		return nil, true
	}

	if d.callbacks.OnAttach != nil {
		return d.callbacks.OnAttach(sess.ID), true
	}

	return nil, true
}

// handleKill kills the selected session.
func (d *Dashboard) handleKill() (tea.Cmd, bool) {
	if d.focusedPane != PaneSessions {
		return nil, false
	}

	sess, ok := d.sessionList.Selected()
	if !ok {
		return nil, true
	}

	if sess.Status != session.StatusRunning && sess.Status != session.StatusWaiting {
		return nil, true
	}

	if d.callbacks.OnKill == nil {
		return nil, true
	}

	cmd, err := d.callbacks.OnKill(sess.ID)
	if err != nil {
		return func() tea.Msg {
			return messages.ErrorMsg{Err: err}
		}, true
	}
	return cmd, true
}

// handleDelete shows delete confirmation.
func (d *Dashboard) handleDelete() (tea.Cmd, bool) {
	if d.focusedPane == PaneSessions {
		return d.handleDeleteSession()
	}

	if d.focusedPane == PaneCards {
		return d.handleDeleteCard()
	}

	return nil, false
}

func (d *Dashboard) handleDeleteSession() (tea.Cmd, bool) {
	sess, ok := d.sessionList.Selected()
	if !ok {
		return nil, false
	}

	if sess.Status != session.StatusTerminated {
		return func() tea.Msg {
			return messages.ErrorMsg{Err: errors.New("only terminated sessions can be removed")}
		}, true
	}

	if sess.UseTmpWorkspace {
		d.pending = PendingAction{
			Type:      "delete-session",
			SessionID: sess.ID,
		}
		d.deleteModal.SetContent("Delete session and its temporary workspace?")
		d.deleteModal.Open()
		return nil, true
	}

	if d.callbacks.OnRemove == nil {
		return nil, true
	}

	cmd, err := d.callbacks.OnRemove(sess.ID)
	if err != nil {
		return func() tea.Msg {
			return messages.ErrorMsg{Err: err}
		}, true
	}
	return cmd, true
}

func (d *Dashboard) handleDeleteCard() (tea.Cmd, bool) {
	card, ok := d.cardList.Selected()
	if !ok {
		return nil, false
	}

	d.pending = PendingAction{
		Type:   "delete-card",
		CardID: card.ID,
	}
	d.deleteModal.SetContent("Delete this ID card?")
	d.deleteModal.Open()
	return nil, true
}

// handleSwitch shows switch modal for terminated sessions.
func (d *Dashboard) handleSwitch() (tea.Cmd, bool) {
	if d.focusedPane != PaneSessions {
		return nil, false
	}

	if sess, ok := d.sessionList.Selected(); ok {
		if sess.Status == session.StatusTerminated {
			d.pending = PendingAction{
				Type:      "switch",
				SessionID: sess.ID,
			}
			d.switchModal.SetCards(d.cards, d.skills, d.personas)
			d.switchModal.Open()
			return nil, true
		}
	}

	return nil, false
}

// updateDeleteModal handles delete confirmation modal.
func (d *Dashboard) updateDeleteModal(msg tea.Msg) (component.Model, tea.Cmd) {
	keyMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		return d, nil
	}

	switch keyMsg.String() {
	case keyEnter, "y":
		return d.confirmDelete()
	case keyEsc, "n":
		d.deleteModal.Close()
		d.pending = PendingAction{}
		return d, nil
	}

	return d, nil
}

func (d *Dashboard) confirmDelete() (component.Model, tea.Cmd) {
	d.deleteModal.Close()
	pending := d.pending
	d.pending = PendingAction{}

	switch pending.Type {
	case "delete-session":
		if d.callbacks.OnRemove != nil {
			cmd, err := d.callbacks.OnRemove(pending.SessionID)
			if err != nil {
				return d, func() tea.Msg { return messages.ErrorMsg{Err: err} }
			}
			d.syncPreviewSelection()
			return d, cmd
		}
	case "delete-card":
		if d.callbacks.OnDeleteCard != nil {
			cmd, err := d.callbacks.OnDeleteCard(pending.CardID)
			if err != nil {
				return d, func() tea.Msg { return messages.ErrorMsg{Err: err} }
			}
			return d, cmd
		}
	}
	return d, nil
}

// updateSwitchModal handles agent switch modal.
func (d *Dashboard) updateSwitchModal(msg tea.Msg) (component.Model, tea.Cmd) {
	model, cmd := d.switchModal.Update(msg)
	//nolint:errcheck // Type assertion is safe - we know the type
	d.switchModal = model.(*SwitchModal)

	if d.switchModal.Confirmed() {
		cardID := d.switchModal.SelectedCardID()
		d.switchModal.Close()

		if d.callbacks.OnSwitch != nil {
			switchCmd, switchErr := d.callbacks.OnSwitch(d.pending.SessionID, cardID)
			if switchErr != nil {
				return d, func() tea.Msg {
					return messages.ErrorMsg{Err: switchErr}
				}
			}
			d.pending = PendingAction{}
			return d, switchCmd
		}
		d.pending = PendingAction{}
	}

	if d.switchModal.Cancelled() {
		d.switchModal.Close()
		d.pending = PendingAction{}
	}

	return d, cmd
}

// updateFocusedPane routes update to focused pane.
func (d *Dashboard) updateFocusedPane(msg tea.Msg) (component.Model, tea.Cmd) {
	var cmd tea.Cmd

	if d.focusedPane == PaneSessions {
		model, c := d.sessionList.Update(msg)
		//nolint:errcheck // Type assertion is safe - we know the type
		d.sessionList = model.(*components.SessionList)
		cmd = c

		d.syncPreviewSelection()
	} else {
		model, c := d.cardList.Update(msg)
		//nolint:errcheck // Type assertion is safe - we know the type
		d.cardList = model.(*components.CardList)
		cmd = c
	}

	return d, cmd
}

// togglePane switches between cards and sessions.
func (d *Dashboard) togglePane() {
	if d.focusedPane == PaneCards && len(d.sessions) > 0 {
		d.focusedPane = PaneSessions
	} else {
		d.focusedPane = PaneCards
	}
	d.syncPreviewSelection()
}

// SetSize implements component.Model.
func (d *Dashboard) SetSize(width, height int) {
	d.Base.SetSize(width, height)

	const (
		borderPadding    = 4
		cardListMinWidth = 46
		headerFooter     = 3
		borderAdjustment = 2
	)

	d.sessionsWidth = width / dashboardSessionsWidthRatio
	d.previewWidth = (width/dashboardSessionsWidthRatio)*dashboardPreviewWidthRatio - borderPadding
	totalSideWidths := d.sessionsWidth + d.previewWidth + borderPadding
	d.cardsWidth = width - totalSideWidths - borderAdjustment

	if d.cardsWidth < cardListMinWidth {
		diff := cardListMinWidth - d.cardsWidth
		d.cardsWidth += diff
		d.previewWidth -= diff
	}

	// Account for header (2 lines) + footer (1 line)
	contentHeight := height - headerFooter

	d.cardList.SetSize(d.cardsWidth, contentHeight)
	d.sessionList.SetSize(d.sessionsWidth, contentHeight)
	d.preview.SetSize(d.previewWidth, contentHeight)
}

// View renders the dashboard.
func (d *Dashboard) View() string {
	// Handle modal overlays
	if d.deleteModal.IsOpen() {
		return d.renderWithOverlay(d.deleteModal.View())
	}
	if d.switchModal.IsOpen() {
		return d.renderWithOverlay(d.switchModal.View())
	}

	return d.renderBase()
}

// renderBase renders the three-pane layout.
func (d *Dashboard) renderBase() string {
	// Build panes
	leftPane := d.renderCardsPane()
	centerPane := d.renderPreviewPane()
	rightPane := d.renderSessionsPane()

	// Join horizontally
	body := lipgloss.JoinHorizontal(lipgloss.Top,
		leftPane,
		centerPane,
		rightPane,
	)

	// Header
	header := lipgloss.NewStyle().
		Foreground(white).
		Background(primaryColor).
		Padding(0, cardsListPadding).
		MarginBottom(1).
		Bold(true).
		Render("K A N K U R O")

	// Footer
	footer := d.renderFooter()

	return header + "\n" + body + "\n" + footer
}

// renderCardsPane renders the left pane.
func (d *Dashboard) renderCardsPane() string {
	style := lipgloss.NewStyle().
		Width(d.cardsWidth).
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(normalBorderColor)

	if d.focusedPane == PaneCards {
		style = style.BorderForeground(focusedBorderColor)
	}

	title := titleStyle.Render("ID Cards")

	// Search indicator
	switch {
	case d.search.IsActive && d.focusedPane == PaneCards:
		title += "\n" + lipgloss.NewStyle().Foreground(gray).
			Render("Search: "+d.search.CardsQuery)
	case d.search.CardsQuery != "":
		title += "\n" + lipgloss.NewStyle().Foreground(gray).
			Render("Filter: "+d.search.CardsQuery)
	default:
		title += "\n "
	}

	content := lipgloss.JoinVertical(lipgloss.Left,
		lipgloss.NewStyle().Padding(0, 1).Render(title),
		d.cardList.View(),
	)

	return style.Render(content)
}

// renderPreviewPane renders the center pane.
func (d *Dashboard) renderPreviewPane() string {
	style := lipgloss.NewStyle().
		Width(d.previewWidth).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(normalBorderColor)

	return style.Render(d.preview.View())
}

// renderSessionsPane renders the right pane.
func (d *Dashboard) renderSessionsPane() string {
	style := lipgloss.NewStyle().
		Width(d.sessionsWidth).
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(normalBorderColor)

	if d.focusedPane == PaneSessions {
		style = style.BorderForeground(focusedBorderColor)
	}

	title := titleStyle.Render("Active Sessions")

	// Search indicator
	switch {
	case d.search.IsActive && d.focusedPane == PaneSessions:
		title += "\n" + lipgloss.NewStyle().Foreground(gray).
			Render("Search: "+d.search.SessionsQuery)
	case d.search.SessionsQuery != "":
		title += "\n" + lipgloss.NewStyle().Foreground(gray).
			Render("Filter: "+d.search.SessionsQuery)
	default:
		title += "\n "
	}

	var content string
	if len(d.sessions) == 0 && !d.search.IsActive && d.search.SessionsQuery == "" {
		// Empty state
		content = renderEmptySessionsState(d.sessionsWidth)
	} else {
		content = d.sessionList.View()
	}

	pane := lipgloss.JoinVertical(lipgloss.Left,
		lipgloss.NewStyle().Padding(0, 1).Render(title),
		content,
	)

	return style.Render(pane)
}

// renderFooter renders the keyboard shortcuts.
func (d *Dashboard) renderFooter() string {
	return lipgloss.JoinHorizontal(lipgloss.Top,
		lipgloss.NewStyle().Foreground(gray).Render("[Nav] Tab "),
		lipgloss.NewStyle().Foreground(gray).Render(" [Card] Enter=Launch n=New e=Edit "),
		lipgloss.NewStyle().Foreground(gray).Render(" [Session] Enter=Attach/Relaunch s=Switch x=Kill Del=Remove "),
		lipgloss.NewStyle().Foreground(gray).Render(" [Global] m=MCP w=Workspaces i=IDE q=Quit"),
	)
}

// renderWithOverlay renders dashboard with modal overlay.
func (d *Dashboard) renderWithOverlay(modalContent string) string {
	base := d.renderBase()

	// Center modal over content
	overlayStyle := lipgloss.NewStyle().
		Padding(1, cardsListPadding)

	return base + "\n" + overlayStyle.Render(modalContent)
}

// renderEmptySessionsState renders the empty state artwork.
func renderEmptySessionsState(width int) string {
	crowASCII := `
    ,   ,
     / \_/ \
    ( o . o )
     \  ^  /
    |||||

   The puppets
   are resting
`
	return lipgloss.NewStyle().
		Width(width).
		Height(dashboardDeleteModalHeight).
		Align(lipgloss.Center, lipgloss.Center).
		Foreground(gray).
		Render(crowASCII)
}

// Helper functions

func findCardByID(cards []config.IDCard, id string) *config.IDCard {
	for i := range cards {
		if cards[i].ID == id {
			return &cards[i]
		}
	}
	return nil
}

// Dashboard-specific theme colors.
var (
	white              = lipgloss.Color("#FFFFFF") //nolint:gochecknoglobals // UI theme colors
	normalBorderColor  = gray                      //nolint:gochecknoglobals // UI theme colors
	focusedBorderColor = primaryColor              //nolint:gochecknoglobals // UI theme colors
)
