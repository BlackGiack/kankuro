package views

import (
	"errors"
	"fmt"
	"strings"
	"unicode"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/BlackGiack/kankuro/internal/config"
	"github.com/BlackGiack/kankuro/internal/session"
	"github.com/BlackGiack/kankuro/internal/ui/component"
	"github.com/BlackGiack/kankuro/internal/ui/components"
	"github.com/BlackGiack/kankuro/internal/ui/messages"
	"github.com/BlackGiack/kankuro/internal/ui/styles"
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
	Code      string // "delete-session", "delete-card", "switch"
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

	OnDeleteCard     func(cardID string) (tea.Cmd, error)
	OnOpenMCP        func() tea.Cmd
	OnOpenWorkspaces func() tea.Cmd
	OnOpenSettings   func() tea.Cmd
	OnOpenCardEditor func() tea.Cmd
	OnLaunchIDE      func(sessionID string) tea.Cmd
	OnLaunchTerminal func(sessionID string) tea.Cmd
}

// Dashboard is the main composite view showing cards, preview, and sessions.
type Dashboard struct {
	component.Base

	cardList    *components.CardList
	sessionList *components.SessionList
	preview     *components.SessionPreview

	deleteModal *components.Modal

	focusedPane Pane
	search      SearchState
	pending     PendingAction

	callbacks DashboardCallbacks

	// Data (passed from App)
	cards    []config.IDCard
	sessions []session.Snapshot
	skills   []config.Skill
	personas []config.Persona
	mcps     []config.MCPServer
	puppets  []config.Puppet

	// Layout (computed from size)
	cardsWidth    int
	sessionsWidth int
	previewWidth  int
	contentHeight int
}

// NewDashboard creates a new dashboard view.
func NewDashboard(callbacks DashboardCallbacks) *Dashboard {
	preview := components.NewSessionPreview()

	deleteModal := components.NewModal("Confirm Delete")

	d := &Dashboard{
		cardList:    nil,
		sessionList: nil,
		preview:     preview,
		deleteModal: deleteModal,
		focusedPane: PaneCards,
		callbacks:   callbacks,
	}

	d.puppets = []config.Puppet{}
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
		OnKill:   func(_ string) {},
		OnDelete: func(_ string) {},
		OnSwitch: func(_ string) {},
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
	if msg.Puppets != nil {
		d.puppets = msg.Puppets
		d.cardList.SetPuppets(d.puppets)
		d.preview.SetPuppets(d.puppets)
	}

	d.syncPreviewSelection()

	// Update scrollback if it matches current preview. This full-snapshot
	// path and SetPreviewData (below) both write to d.preview, but never
	// concurrently: both run on the main Bubble Tea goroutine.
	if msg.PreviewSessionID != "" && msg.PreviewSessionID == d.preview.SessionID() {
		d.preview.SetScrollback(msg.PreviewScrollback)
	}
}

// SetSessions applies just the session list - used by the coordinator for
// the idle tick's sessions-only refresh (messages.SessionsRefreshedMsg),
// which is routed directly here rather than through SetData/broadcastData so
// it never reaches an unrelated DataReceiver - see
// debt_impl_plan_5.md review finding 1.
func (d *Dashboard) SetSessions(sessions []session.Snapshot) {
	if sessions == nil {
		return
	}
	d.sessions = sessions
	d.sessionList.SetSessions(d.sessions, d.cards)
	d.syncPreviewSelection()
}

// SetPreviewData implements messages.PreviewReceiver, handling partial,
// preview-only updates that must never be confused with a full SetData
// snapshot (see messages.PreviewDataMsg).
func (d *Dashboard) SetPreviewData(msg messages.PreviewDataMsg) {
	if msg.SessionID == "" || msg.SessionID != d.preview.SessionID() {
		return
	}
	if msg.IsIncremental {
		d.preview.AppendScrollback(msg.Scrollback)
	} else {
		d.preview.SetScrollback(msg.Scrollback)
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
		disableKittyProtocolCmd,
		d.cardList.Init(),
		d.sessionList.Init(),
		d.preview.Init(),
	)
}

// disableKittyProtocolCmd disables the Kitty keyboard protocol when control
// returns to the Kankuro dashboard (e.g. on startup, or navigating back from
// another view). This used to run unconditionally from View() - a pure
// render function Bubble Tea may call arbitrarily often - which fought the
// renderer's own output on every frame. Emitting it once as a tea.Cmd from
// Init() (the transition point) keeps the same effect without the render-time
// side effect. See also the attach->detach transition in app/attach.go, which
// needs the same sequence for the same reason.
func disableKittyProtocolCmd() tea.Msg {
	fmt.Print("\x1b[=0u") //nolint:forbidigo // disable kitty protocol on dashboard transition, not on every render
	return nil
}

// Update handles messages.
func (d *Dashboard) Update(msg tea.Msg) (component.Model, tea.Cmd) {
	// Handle modals first
	if d.deleteModal.IsOpen() {
		return d.updateDeleteModal(msg)
	}

	// Handle key events
	if keyMsg, ok := msg.(tea.KeyPressMsg); ok {
		if cmd, handled := d.handleKey(keyMsg); handled {
			return d, cmd
		}
	}

	// Route to focused pane
	return d.updateFocusedPane(msg)
}

// handleKey processes global dashboard key events.
func (d *Dashboard) handleKey(keyMsg tea.KeyPressMsg) (tea.Cmd, bool) {
	// Handle search input
	if handled := d.handleSearch(keyMsg); handled {
		return nil, true
	}

	if keyMsg.String() == "ctrl+c" {
		return tea.Quit, true
	}

	switch keyMsg.Code {
	case tea.KeyEsc:
		return nil, true
	case tea.KeyTab:
		d.togglePane()
		return nil, true
	case tea.KeyEnter:
		return d.handleEnter()
	case tea.KeyBackspace, tea.KeyDelete:
		return d.handleDelete()
	}

	if keyMsg.String() == keyShiftTab || keyMsg.Code == tea.KeyLeft || keyMsg.Code == tea.KeyRight {
		d.togglePane()
		return nil, true
	}

	if len(keyMsg.Text) > 0 {
		switch keyMsg.String() {
		case "q":
			return tea.Quit, true
		case "n":
			return d.handleNewCard()
		case "e":
			return d.handleEdit()
		case "m":
			return d.handleOpenMCP()
		case "w":
			return d.handleOpenWorkspaces()
		case "c":
			return d.handleOpenSettings()
		case "o":
			return d.handleLaunchSession(d.callbacks.OnLaunchIDE)
		case "t":
			return d.handleLaunchSession(d.callbacks.OnLaunchTerminal)
		case "a":
			return d.handleAttach()
		case "x":
			return d.handleKill()
		case "s":
			return d.handleSwitch()
		}
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

func (d *Dashboard) handleOpenSettings() (tea.Cmd, bool) {
	if d.callbacks.OnOpenSettings != nil {
		return d.callbacks.OnOpenSettings(), true
	}
	return nil, true
}
func (d *Dashboard) handleLaunchSession(launch func(sessionID string) tea.Cmd) (tea.Cmd, bool) {
	if d.focusedPane != PaneSessions {
		return nil, true
	}

	sess, ok := d.sessionList.Selected()
	if !ok {
		return nil, true
	}

	if launch == nil {
		return nil, true
	}

	return launch(sess.ID), true
}

// handleSearch processes search input.
func (d *Dashboard) handleSearch(keyMsg tea.KeyPressMsg) bool {
	if keyMsg.String() == "/" && !d.search.IsActive {
		d.search.IsActive = true
		return true
	}

	if !d.search.IsActive {
		return false
	}

	switch keyMsg.Code {
	case tea.KeyEsc:
		d.cancelSearch()
		return true
	case tea.KeyEnter:
		d.search.IsActive = false
		d.syncPreviewSelection()
		return true
	case tea.KeyBackspace:
		d.handleSearchBackspace()
		return true
	default:
		if len(keyMsg.Text) > 0 {
			for _, ch := range keyMsg.Text {
				d.handleSearchRune(ch)
			}
			return true
		}
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
	if unicode.IsPrint(ch) && !unicode.IsSpace(ch) {
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
			Code:      "delete-session",
			SessionID: sess.ID,
		}
		d.deleteModal.SetContent("Delete session and its temporary workspace?")
		d.deleteModal.SetLegend("Enter/y: Confirm • Esc/n: Cancel")
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
		Code:   "delete-card",
		CardID: card.ID,
	}
	d.deleteModal.SetContent("Delete this ID card?")
	d.deleteModal.SetLegend("Enter/y: Confirm • Esc/n: Cancel")
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
			return func() tea.Msg {
				return messages.AgentSwitchRequestedMsg{
					SessionID:   sess.ID,
					SessionName: sess.Name,
				}
			}, true
		}
	}

	return nil, false
}

// updateDeleteModal handles delete confirmation modal.
func (d *Dashboard) updateDeleteModal(msg tea.Msg) (component.Model, tea.Cmd) {
	keyMsg, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return d, nil
	}

	switch keyMsg.Code {
	case tea.KeyEnter:
		return d.confirmDelete()
	case tea.KeyEsc:
		d.deleteModal.Close()
		d.pending = PendingAction{}
		return d, nil
	default:
		if len(keyMsg.Text) > 0 {
			switch keyMsg.String() {
			case "y":
				return d.confirmDelete()
			case "n":
				d.deleteModal.Close()
				d.pending = PendingAction{}
				return d, nil
			}
		}
	}

	return d, nil
}

func (d *Dashboard) confirmDelete() (component.Model, tea.Cmd) {
	d.deleteModal.Close()
	pending := d.pending
	d.pending = PendingAction{}

	switch pending.Code {
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

// updateFocusedPane routes update to focused pane.
func (d *Dashboard) updateFocusedPane(msg tea.Msg) (component.Model, tea.Cmd) {
	var cmd tea.Cmd

	if d.focusedPane == PaneSessions {
		model, c := d.sessionList.Update(msg)
		d.sessionList, _ = model.(*components.SessionList)
		cmd = c

		d.syncPreviewSelection()
	} else {
		model, c := d.cardList.Update(msg)
		d.cardList, _ = model.(*components.CardList)
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

	const cardListMinWidth = 46

	d.sessionsWidth = width / dashboardSessionsWidthRatio
	d.previewWidth = (width / dashboardSessionsWidthRatio) * dashboardPreviewWidthRatio
	d.cardsWidth = width - d.sessionsWidth - d.previewWidth

	if d.cardsWidth < cardListMinWidth {
		diff := cardListMinWidth - d.cardsWidth
		d.cardsWidth += diff
		d.previewWidth -= diff
	}
	// Below the minimum terminal size the coordinator shows a "too small"
	// screen instead of this view, but it is still sized; never go negative.
	d.previewWidth = max(0, d.previewWidth)

	// The coordinator sizes the dashboard to (terminal rows - CoordinatorRows).
	// For the whole render (header + body + footer) to fill the terminal with no
	// gap, the body - the outer pane height - is every terminal row minus the
	// header and footer rows.
	paneHeight := height + CoordinatorRows - dashboardHeaderRows - dashboardFooterRows
	d.contentHeight = paneHeight

	// In border-box, lists must fit inside outer Height - borders (2) - title block (2) = paneHeight - 4.
	// Preview pane has no title block, so its component fits in outer Height - borders (2) = paneHeight - 2.
	d.cardList.SetSize(d.cardsWidth-paneBorderRing, paneHeight-paneBorderRing-previewTitleBlock)
	d.sessionList.SetSize(d.sessionsWidth-paneBorderRing, paneHeight-paneBorderRing-previewTitleBlock)
	d.preview.SetSize(
		d.previewWidth-paneBorderRing-2*paneContentGutter,
		paneHeight-paneBorderRing,
	)
	d.deleteModal.SetSize(width, height)
}

// View renders the dashboard.
func (d *Dashboard) View() string {
	// Handle modal overlays
	if d.deleteModal.IsOpen() {
		return d.renderWithOverlay(d.deleteModal.View())
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
		Foreground(styles.White).
		Background(styles.PrimaryColor).
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
		Height(d.contentHeight).
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(normalBorderColor)

	if d.focusedPane == PaneCards {
		style = style.BorderForeground(focusedBorderColor)
	}

	title := d.renderPaneTitle("ID Cards", d.search.CardsQuery,
		d.search.IsActive && d.focusedPane == PaneCards, d.cardsWidth)

	content := lipgloss.JoinVertical(lipgloss.Left,
		paneTitleStyle().Render(title),
		d.cardList.View(),
	)

	return style.Render(content)
}

// renderPreviewPane renders the center pane.
func (d *Dashboard) renderPreviewPane() string {
	// Note: MaxWidth/MaxHeight here would clip the right/bottom borders because
	// they cap the total rendered size *including* the border ring. The inner
	// SessionPreview already hard-clamps its own output to previewWidth ×
	// contentHeight via clampToSize, so the bordered pane below is guaranteed
	// to be exactly previewWidth+2 × contentHeight+2.
	style := lipgloss.NewStyle().
		Width(d.previewWidth).
		Height(d.contentHeight).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(normalBorderColor)

	// Final defensive clamp: even though SessionPreview clamps its own View()
	// and the bordered style above adds exactly 2 cells in each dimension,
	// hard-truncate the rendered pane to the expected outer size as a last
	// safety net. This is what prevents punctual overflow from a single
	// unforeseen content shape (malformed escapes, exotic graphemes, etc.)
	// from breaking the whole dashboard layout.
	// The preview was sized to (previewWidth, contentHeight + previewTitleBlock)
	// in SetSize, so its body is up to contentHeight + previewTitleBlock rows.
	// Adding the 2-cell border ring gives the expected bordered pane height.
	// Cutting this short would shave the bottom border off.
	content := lipgloss.NewStyle().Padding(0, paneContentGutter).Render(d.preview.View())
	rendered := style.Render(content)
	maxW := d.previewWidth
	maxH := d.contentHeight
	return components.ClampToSize(rendered, maxW, maxH)
}

// paneBorderRing is the number of rows/columns added by a 4-side lipgloss
// border around a pane (1 top + 1 bottom, 1 left + 1 right).
// previewTitleBlock is the number of rows the cards/sessions panes spend on
// their internal title + filter/spacer line before the list. The preview pane
// has no such title block, so we extend its inner height by the same amount
// to keep the three panes' bottom borders aligned.
const (
	paneBorderRing    = 2
	previewTitleBlock = 2
	paneContentGutter = 2

	// paneTitlePadding is the columns the title block's horizontal padding takes
	// from a pane's inner width.
	paneTitlePadding      = paneContentGutter + paneTitleRightPadding
	paneTitleRightPadding = 1
)

// Rows of the dashboard render, top to bottom: the header and its bottom
// margin, the panes, the footer. The dashboard fills all of the terminal's rows
// (CoordinatorRows, defined in constants.go, is what the coordinator holds back
// before sizing it).
const (
	dashboardHeaderRows = 2
	dashboardFooterRows = 1
)

// renderedHeight is how many rows renderBase produces: every terminal row.
func (d *Dashboard) renderedHeight() int {
	return dashboardHeaderRows + d.contentHeight + dashboardFooterRows
}

// renderPaneTitle renders a pane's two-line title block: the title, then the
// search or filter line (or a blank spacer). The query is user text; the
// label stays and the head of the query is clipped, because its end is where
// the user is typing. width is the pane's outer width.
func (d *Dashboard) renderPaneTitle(title, query string, searching bool, width int) string {
	const searchLabel, filterLabel = "Search: ", "Filter: "

	label := filterLabel
	switch {
	case searching:
		label = searchLabel
	case query == "":
		return styles.TitleStyle.Render(title) + "\n "
	}
	inner := max(0, width-paneBorderRing-paneTitlePadding)
	query = components.TruncateTail(components.DisplayText(query), inner-len(label))
	return styles.TitleStyle.Render(title) + "\n" +
		lipgloss.NewStyle().Foreground(styles.Gray).Render(label+query)
}

// renderSessionsPane renders the right pane.
func (d *Dashboard) renderSessionsPane() string {
	style := lipgloss.NewStyle().
		Width(d.sessionsWidth).
		Height(d.contentHeight).
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(normalBorderColor)

	if d.focusedPane == PaneSessions {
		style = style.BorderForeground(focusedBorderColor)
	}

	title := d.renderPaneTitle("Active Sessions", d.search.SessionsQuery,
		d.search.IsActive && d.focusedPane == PaneSessions, d.sessionsWidth)

	var content string
	if len(d.sessions) == 0 && !d.search.IsActive && d.search.SessionsQuery == "" {
		// Empty state
		content = renderEmptySessionsState(d.sessionsWidth, d.sessionList.Height)
	} else {
		content = d.sessionList.View()
	}

	pane := lipgloss.JoinVertical(lipgloss.Left,
		paneTitleStyle().Render(title),
		content,
	)

	return style.Render(pane)
}

func paneTitleStyle() lipgloss.Style {
	return lipgloss.NewStyle().Padding(0, paneTitleRightPadding, 0, paneContentGutter)
}

// Footer key groups. The footer shows the navigation group, the group of the
// focused pane and the global group, so it carries only keys that do
// something here.
const (
	footerNav             = "[Nav] Tab"
	footerCards           = "[Card] Enter=Launch n=New e=Edit"
	footerSessions        = "[Session] Enter=Attach/Relaunch o=IDE t=Terminal s=Switch x=Kill Del=Remove"
	footerSessionsCompact = "[Session] Enter=Attach/Relaunch o=IDE t=Terminal x=Kill"
	footerGlobal          = "[Global] m=MCP w=Workspaces c=Settings q=Quit"
	footerGlobalMinimal   = "[Global] q=Quit"
	footerGroupGap        = "  "
)

// footerText picks the richest key legend that fits width. It gives up, in
// order: the navigation group, the rest of the global group (q=Quit stays),
// the focused pane's group - and as a last resort clips the line.
func (d *Dashboard) footerText(width int) string {
	pane := footerCards
	if d.focusedPane == PaneSessions {
		pane = footerSessions
	}
	candidates := [][]string{
		{footerNav, pane, footerGlobal},
		{pane, footerGlobal},
		{pane, footerGlobalMinimal},
	}
	if d.focusedPane == PaneSessions {
		candidates = append(candidates,
			[]string{footerSessionsCompact, footerGlobalMinimal},
			[]string{footerSessionsCompact},
		)
	}
	candidates = append(candidates, []string{footerGlobalMinimal})
	if width <= 0 { // not sized yet: nothing to fit
		return strings.Join(candidates[0], footerGroupGap)
	}
	for _, groups := range candidates {
		if text := strings.Join(groups, footerGroupGap); lipgloss.Width(text) <= width {
			return text
		}
	}
	return components.TruncateHead(footerGlobalMinimal, width)
}

// renderFooter renders the keyboard shortcuts, fitted to the dashboard's width.
func (d *Dashboard) renderFooter() string {
	return lipgloss.NewStyle().Foreground(styles.Gray).Render(d.footerText(d.Width))
}

// renderWithOverlay renders dashboard with modal overlay.
func (d *Dashboard) renderWithOverlay(modalContent string) string {
	return lipgloss.Place(d.Width, d.renderedHeight(), lipgloss.Center, lipgloss.Center, modalContent)
}

// renderEmptySessionsState renders the empty state artwork.
func renderEmptySessionsState(width, height int) string {
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
		Height(height).
		Align(lipgloss.Center, lipgloss.Center).
		Foreground(styles.Gray).
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
	normalBorderColor  = styles.Gray         //nolint:gochecknoglobals // UI theme colors
	focusedBorderColor = styles.PrimaryColor //nolint:gochecknoglobals // UI theme colors
)
