package views

import (
	bubbletea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"kankuro/internal/config"
	"kankuro/internal/session"
	"kankuro/internal/ui/commands"
	"kankuro/internal/ui/component"
	"kankuro/internal/ui/components"
)

// Pane identifiers.
type Pane int

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
	OnEditCard       func(card config.IDCard)
	OnLaunchCard     func(card config.IDCard)
	OnAttach         func(sessionID string) bubbletea.Cmd
	OnKill           func(sessionID string) error
	OnRemove         func(sessionID string) error
	OnSwitch         func(sessionID string, newCardID string) error
	OnDeleteCard     func(cardID string) error
	OnOpenMCP        func()
	OnOpenWorkspaces func()
	OnOpenIDE        func()
	OnOpenCardEditor func()
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
	cardList := components.NewCardList(components.CardListCallbacks{
		OnSelect: callbacks.OnLaunchCard,
		OnEdit:   callbacks.OnEditCard,
		OnDelete: func(_ string) {
			// Parent handles via message
		},
	})

	sessionList := components.NewSessionList(components.SessionListCallbacks{
		OnSelect: func(_ session.Snapshot) {},
		OnAttach: func(_ string) {
			// Parent handles via message
		},
		OnKill:    func(_ string) {},
		OnDelete:  func(_ string) {},
		OnSwitch:  func(_ string) {},
		OnPreview: func(_ string) {},
	})

	preview := components.NewSessionPreview()

	deleteModal := components.NewModal("Confirm Delete")
	switchModal := NewSwitchModal()

	return &Dashboard{
		cardList:    cardList,
		sessionList: sessionList,
		preview:     preview,
		deleteModal: deleteModal,
		switchModal: switchModal,
		focusedPane: PaneCards,
		callbacks:   callbacks,
		cmds:        cmds,
	}
}

// SetData updates all child components with fresh data.
func (d *Dashboard) SetData(
	cards []config.IDCard,
	sessions []session.Snapshot,
	skills []config.Skill,
	personas []config.Persona,
	mcps []config.MCPServer,
) {
	d.cards = cards
	d.sessions = sessions
	d.skills = skills
	d.personas = personas
	d.mcps = mcps

	d.cardList.SetCards(cards, skills, personas)
	d.sessionList.SetSessions(sessions, cards)

	// Sync preview to first session if we have sessions
	if len(sessions) > 0 && d.preview.SessionID() == "" {
		sess := sessions[0]
		d.preview.SetSession(&sess, findCardByID(cards, sess.CardID))
	}
}

// Init initializes the dashboard.
func (d *Dashboard) Init() bubbletea.Cmd {
	return bubbletea.Batch(
		d.cardList.Init(),
		d.sessionList.Init(),
		d.preview.Init(),
	)
}

// Update handles messages.
func (d *Dashboard) Update(msg bubbletea.Msg) (component.Model, bubbletea.Cmd) {
	// Handle modals first
	if d.deleteModal.IsOpen() {
		return d.updateDeleteModal(msg)
	}
	if d.switchModal.IsOpen() {
		return d.updateSwitchModal(msg)
	}

	// Handle key events
	if keyMsg, ok := msg.(bubbletea.KeyMsg); ok {
		if cmd, handled := d.handleKey(keyMsg); handled {
			return d, cmd
		}
	}

	// Route to focused pane
	return d.updateFocusedPane(msg)
}

// handleKey processes global dashboard key events.
func (d *Dashboard) handleKey(keyMsg bubbletea.KeyMsg) (bubbletea.Cmd, bool) {
	// Handle search input
	if cmd, handled := d.handleSearch(keyMsg); handled {
		return cmd, true
	}

	switch keyMsg.String() {
	case "q", "ctrl+c":
		return bubbletea.Quit, true

	case "esc":
		return nil, true

	case "tab", "shift+tab", "left", "right":
		d.togglePane()
		return nil, true

	case "n":
		if d.callbacks.OnOpenCardEditor != nil {
			d.callbacks.OnOpenCardEditor()
		}
		return nil, true

	case "e":
		if d.focusedPane == PaneCards {
			if card, ok := d.cardList.Selected(); ok {
				if d.callbacks.OnEditCard != nil {
					d.callbacks.OnEditCard(card)
				}
			}
		}
		return nil, true

	case "m":
		if d.callbacks.OnOpenMCP != nil {
			d.callbacks.OnOpenMCP()
		}
		return nil, true

	case "w":
		if d.callbacks.OnOpenWorkspaces != nil {
			d.callbacks.OnOpenWorkspaces()
		}
		return nil, true

	case "i":
		if d.callbacks.OnOpenIDE != nil {
			d.callbacks.OnOpenIDE()
		}
		return nil, true

	case "enter":
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

// handleSearch processes search input.
func (d *Dashboard) handleSearch(keyMsg bubbletea.KeyMsg) (bubbletea.Cmd, bool) {
	key := keyMsg.String()

	switch key {
	case "esc":
		if d.search.IsActive {
			d.search.CardsQuery = ""
			d.search.SessionsQuery = ""
			d.search.IsActive = false
			d.cardList.SetSearchQuery("")
			d.sessionList.SetSearchQuery("")
			return nil, true
		}

	case "/":
		if !d.search.IsActive {
			d.search.IsActive = true
			return nil, true
		}

	case "enter":
		if d.search.IsActive {
			d.search.IsActive = false
			return nil, true
		}

	case "backspace":
		if d.search.IsActive {
			if d.focusedPane == PaneCards && len(d.search.CardsQuery) > 0 {
				d.search.CardsQuery = d.search.CardsQuery[:len(d.search.CardsQuery)-1]
				d.cardList.SetSearchQuery(d.search.CardsQuery)
			} else if d.focusedPane == PaneSessions && len(d.search.SessionsQuery) > 0 {
				d.search.SessionsQuery = d.search.SessionsQuery[:len(d.search.SessionsQuery)-1]
				d.sessionList.SetSearchQuery(d.search.SessionsQuery)
			}
			return nil, true
		}
	}

	// Handle printable characters
	if len(keyMsg.Runes) > 0 && d.search.IsActive {
		ch := keyMsg.Runes[0]
		if ch >= 33 && ch <= 126 {
			if d.focusedPane == PaneCards {
				d.search.CardsQuery += string(ch)
				d.cardList.SetSearchQuery(d.search.CardsQuery)
			} else {
				d.search.SessionsQuery += string(ch)
				d.sessionList.SetSearchQuery(d.search.SessionsQuery)
			}
			return nil, true
		}
	}

	return nil, false
}

// handleEnter processes enter key based on focused pane.
func (d *Dashboard) handleEnter() (bubbletea.Cmd, bool) {
	if d.focusedPane == PaneSessions {
		if sess, ok := d.sessionList.Selected(); ok {
			if d.callbacks.OnAttach != nil {
				return d.callbacks.OnAttach(sess.ID), true
			}
		}
		return nil, true
	}

	if card, ok := d.cardList.Selected(); ok {
		if d.callbacks.OnLaunchCard != nil {
			d.callbacks.OnLaunchCard(card)
		}
	}
	return nil, true
}

// handleAttach attaches to selected session.
func (d *Dashboard) handleAttach() (bubbletea.Cmd, bool) {
	if d.focusedPane != PaneSessions {
		return nil, false
	}
	return d.handleEnter()
}

// handleKill kills the selected session.
func (d *Dashboard) handleKill() (bubbletea.Cmd, bool) {
	if d.focusedPane != PaneSessions {
		return nil, false
	}

	if sess, ok := d.sessionList.Selected(); ok {
		if d.callbacks.OnKill != nil {
			if err := d.callbacks.OnKill(sess.ID); err != nil {
				// Emit error message
			}
		}
	}
	return nil, true
}

// handleDelete shows delete confirmation.
func (d *Dashboard) handleDelete() (bubbletea.Cmd, bool) {
	if d.focusedPane == PaneSessions {
		if sess, ok := d.sessionList.Selected(); ok {
			if sess.UseTmpWorkspace {
				d.pending = PendingAction{
					Type:      "delete-session",
					SessionID: sess.ID,
				}
				d.deleteModal.SetContent("Delete session and its temporary workspace?")
				d.deleteModal.Open()
				return nil, true
			}
			if d.callbacks.OnRemove != nil {
				if err := d.callbacks.OnRemove(sess.ID); err != nil {
					// Error is handled by callback implementation
				}
			}
			return nil, true
		}
	}

	if d.focusedPane == PaneCards {
		if card, ok := d.cardList.Selected(); ok {
			d.pending = PendingAction{
				Type:   "delete-card",
				CardID: card.ID,
			}
			d.deleteModal.SetContent("Delete this ID card?")
			d.deleteModal.Open()
			return nil, true
		}
	}

	return nil, false
}

// handleSwitch shows switch modal for terminated sessions.
func (d *Dashboard) handleSwitch() (bubbletea.Cmd, bool) {
	if d.focusedPane != PaneSessions {
		return nil, false
	}

	if sess, ok := d.sessionList.Selected(); ok {
		if sess.Status != session.StatusTerminated {
			return nil, false
		}

		d.pending = PendingAction{
			Type:      "switch",
			SessionID: sess.ID,
		}
		d.switchModal.SetCards(d.cards, d.skills, d.personas)
		d.switchModal.Open()
		return nil, true
	}

	return nil, false
}

// updateDeleteModal handles delete confirmation modal.
func (d *Dashboard) updateDeleteModal(msg bubbletea.Msg) (component.Model, bubbletea.Cmd) {
	if keyMsg, ok := msg.(bubbletea.KeyMsg); ok {
		switch keyMsg.String() {
		case "enter", "y":
			d.deleteModal.Close()
			pending := d.pending
			d.pending = PendingAction{}

			switch pending.Type {
			case "delete-session":
				if d.callbacks.OnRemove != nil {
					if err := d.callbacks.OnRemove(pending.SessionID); err != nil {
						// Error is handled by callback implementation
					}
				}
			case "delete-card":
				if d.callbacks.OnDeleteCard != nil {
					if err := d.callbacks.OnDeleteCard(pending.CardID); err != nil {
						// Error is handled by callback implementation
					}
				}
			}
			return d, nil

		case "esc", "n":
			d.deleteModal.Close()
			d.pending = PendingAction{}
			return d, nil
		}
	}

	return d, nil
}

// updateSwitchModal handles agent switch modal.
func (d *Dashboard) updateSwitchModal(msg bubbletea.Msg) (component.Model, bubbletea.Cmd) {
	model, cmd := d.switchModal.Update(msg)
	//nolint:errcheck // Type assertion is safe - we know the type
	d.switchModal = model.(*SwitchModal)

	if d.switchModal.Confirmed() {
		cardID := d.switchModal.SelectedCardID()
		d.switchModal.Close()

		if d.callbacks.OnSwitch != nil {
			if err := d.callbacks.OnSwitch(d.pending.SessionID, cardID); err != nil {
				// Error is handled by callback implementation
			}
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
func (d *Dashboard) updateFocusedPane(msg bubbletea.Msg) (component.Model, bubbletea.Cmd) {
	var cmd bubbletea.Cmd

	if d.focusedPane == PaneSessions {
		model, c := d.sessionList.Update(msg)
		//nolint:errcheck // Type assertion is safe - we know the type
		d.sessionList = model.(*components.SessionList)
		cmd = c

		// Sync preview
		if sess, ok := d.sessionList.Selected(); ok {
			sessCopy := sess
			d.preview.SetSession(&sessCopy, findCardByID(d.cards, sess.CardID))
		}
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
}

// SetSize implements component.Model.
func (d *Dashboard) SetSize(width, height int) {
	d.Base.SetSize(width, height)

	// Calculate layout proportions (matches existing dashboard layout)
	// Cards: ~25%, Preview: ~35%, Sessions: ~40%
	d.cardsWidth = width / 4
	d.previewWidth = width / 3
	d.sessionsWidth = width - d.cardsWidth - d.previewWidth

	// Account for header (2 lines) + footer (1 line)
	contentHeight := height - 3

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
		Padding(0, 2).
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
		lipgloss.NewStyle().Foreground(gray).Render(" [Session] Enter=Attach x=Kill Del=Remove "),
		lipgloss.NewStyle().Foreground(gray).Render(" [Global] m=MCP w=Workspaces i=IDE q=Quit"),
	)
}

// renderWithOverlay renders dashboard with modal overlay.
func (d *Dashboard) renderWithOverlay(modalContent string) string {
	base := d.renderBase()

	// Center modal over content
	overlayStyle := lipgloss.NewStyle().
		Padding(1, 2)

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
		Height(10).
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
