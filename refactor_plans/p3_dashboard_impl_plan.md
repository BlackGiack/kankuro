# Implementation Plan: Phase 3 Task 3.1 - Dashboard Component

## Goal

Create Dashboard composite view that composes CardList, SessionList, SessionPreview, and modals into a cohesive three-pane layout, consolidating 4 view modes (viewDashboard, viewDeleteConfirm, viewSwitchAgent) into a single component.

---

## Analysis: What to Extract

### View Mode Consolidation

| Current View Mode | New Location |
|-------------------|--------------|
| `viewDashboard` | `Dashboard.base` |
| `viewDeleteConfirm` | `Dashboard.modal.deleteConfirm` |
| `viewSwitchAgent` | `Dashboard.modal.switchAgent` |

### From Existing Code

| File | Lines | Content | Destination |
|------|-------|---------|-------------|
| `app_dashboard.go` | 14-87 | `updateDashboard()` key handling | `views/dashboard.go` |
| `app_dashboard.go` | 89-133 | `syncPreviewSelection()` | `views/dashboard.go` |
| `app_dashboard.go` | 135-201 | Search handling | `views/dashboard.go` |
| `app_dashboard.go` | 203-239 | `handleDashboardEdit()` | Callbacks |
| `app_dashboard.go` | 241-263 | `handleDashboardEnter()` | Callbacks |
| `app_dashboard.go` | 272-301 | `handleDashboardDelete()` | Modal handling |
| `app_dashboard.go` | 303-324 | `handleDashboardSwitch()` | Modal handling |
| `app_dashboard.go` | 326-333 | `toggleDashboardPane()` | `views/dashboard.go` |
| `app_view.go` | 58-159 | `renderDashboard()` | `views/dashboard.go` |
| `app_state.go` | 54-71 | `SessionViewState` | `views/dashboard.go` (internal) |
| `view_preview.go` | 44-121 | `renderSessionPreview()` | `components/session_preview.go` (Phase 2.3) |
| `styles.go` | 6-55 | Theme colors | `views/styles.go` (shared) |

### Dashboard-Specific Constants

```go
// Layout proportions
headerPadding = 2

// Pane focus
paneCards
paneSessions
```

---

## Architecture Overview

### Component Hierarchy

```
Dashboard (CompositeView)
├── cardList      *components.CardList      // Phase 2.1
├── sessionList   *components.SessionList   // Phase 2.2
├── preview       *components.SessionPreview // Phase 2.3
├── modals
│   ├── deleteModal *components.Modal
│   └── switchModal *components.CardPickerModal (NEW)
└── state
    ├── focusedPane   dashboardPane
    ├── searchState   SearchState
    └── pendingAction PendingAction
```

### Message Flow

```
User Input → Dashboard.Update()
    │
    ├─→ If modal open → Modal.Update()
    │       └─→ Return ModalResult (Confirm/Cancel)
    │
    ├─→ Else dispatch to focused pane
    │       ├─→ paneCards    → cardList.Update()
    │       └─→ paneSessions → sessionList.Update()
    │
    └─→ Handle emitted messages
            ├─→ CardSelectedMsg    → LaunchSession
            ├─→ SessionSelectedMsg → SyncPreview
            └─→ NavigateMsg        → Change view
```

---

## Step 3.1.1: Create Dashboard View File

**File**: `internal/ui/views/dashboard.go` (NEW)

```go
package views

import (
    "fmt"
    "strings"

    bubbletea "github.com/charmbracelet/bubbletea"
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
    Type       string // "delete-session", "delete-card", "switch"
    SessionID  string
    CardID     string
}

// DashboardCallbacks defines callbacks for dashboard actions.
type DashboardCallbacks struct {
    OnEditCard    func(card config.IDCard)
    OnLaunchCard  func(card config.IDCard)
    OnAttach      func(sessionID string) bubbletea.Cmd
    OnKill        func(sessionID string) error
    OnRemove      func(sessionID string) error
    OnSwitch      func(sessionID string, newCardID string) error
    OnDeleteCard  func(cardID string) error
    OnOpenMCP     func()
    OnOpenWorkspaces func()
    OnOpenIDE     func()
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
    
    callbacks   DashboardCallbacks
    cmds        *commands.Commands
    
    // Data (passed from App)
    cards      []config.IDCard
    sessions   []session.Snapshot
    skills     []config.Skill
    personas   []config.Persona
    mcps       []config.MCPServer
    
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
        OnDelete: func(cardID string) {
            // Parent handles via message
        },
    })
    
    sessionList := components.NewSessionList(components.SessionListCallbacks{
        OnSelect: func(s session.Snapshot) {},
        OnAttach: func(sessionID string) {
            // Parent handles via message
        },
        OnKill:   func(sessionID string) {},
        OnDelete: func(sessionID string) {},
        OnSwitch: func(sessionID string) {},
        OnPreview: func(sessionID string) {},
    })
    
    preview := components.NewSessionPreview()
    
    deleteModal := components.NewModal("Confirm Delete")
    switchModal := NewSwitchModal()
    
    return &Dashboard{
        cardList:     cardList,
        sessionList:  sessionList,
        preview:      preview,
        deleteModal:  deleteModal,
        switchModal:  switchModal,
        focusedPane:  PaneCards,
        callbacks:    callbacks,
        cmds:         cmds,
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
        d.preview.SetSession(sessions[0], findCardByID(cards, sessions[0].CardID))
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
                d.callbacks.OnRemove(sess.ID)
            }
            return nil, true
        }
    }
    
    if d.focusedPane == PaneCards {
        if card, ok := d.cardList.Selected(); ok {
            d.pending = PendingAction{
                Type:  "delete-card",
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
                    d.callbacks.OnRemove(pending.SessionID)
                }
            case "delete-card":
                if d.callbacks.OnDeleteCard != nil {
                    d.callbacks.OnDeleteCard(pending.CardID)
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
    d.switchModal = model.(*SwitchModal)
    
    if d.switchModal.Confirmed() {
        cardID := d.switchModal.SelectedCardID()
        d.switchModal.Close()
        
        if d.callbacks.OnSwitch != nil {
            d.callbacks.OnSwitch(d.pending.SessionID, cardID)
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
        d.sessionList = model.(*components.SessionList)
        cmd = c
        
        // Sync preview
        if sess, ok := d.sessionList.Selected(); ok {
            d.preview.SetSession(sess, findCardByID(d.cards, sess.CardID))
        }
    } else {
        model, c := d.cardList.Update(msg)
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
            Render("Search: " + d.search.CardsQuery)
    case d.search.CardsQuery != "":
        title += "\n" + lipgloss.NewStyle().Foreground(gray).
            Render("Filter: " + d.search.CardsQuery)
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
            Render("Search: " + d.search.SessionsQuery)
    case d.search.SessionsQuery != "":
        title += "\n" + lipgloss.NewStyle().Foreground(gray).
            Render("Filter: " + d.search.SessionsQuery)
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

// Theme colors (should be shared)
var (
    primaryColor     = lipgloss.Color("#8E44AD")
    highlightColor   = lipgloss.Color("#16A085")
    gray             = lipgloss.Color("240")
    white            = lipgloss.Color("#FFFFFF")
    normalBorderColor  = gray
    focusedBorderColor = primaryColor
    titleStyle = lipgloss.NewStyle().
        Bold(true).
        Foreground(primaryColor)
)
```

---

## Step 3.1.2: Create SwitchModal Component

**File**: `internal/ui/views/switch_modal.go` (NEW)

```go
package views

import (
    bubbletea "github.com/charmbracelet/bubbletea"
    "github.com/charmbracelet/bubbles/list"
    "github.com/charmbracelet/lipgloss"
    
    "kankuro/internal/config"
    "kankuro/internal/ui/component"
    "kankuro/internal/ui/components"
)

// SwitchModal allows selecting a new agent for a terminated session.
type SwitchModal struct {
    component.Base
    list      list.Model
    cards     []config.IDCard
    confirmed bool
    cancelled bool
}

// NewSwitchModal creates a new switch modal.
func NewSwitchModal() *SwitchModal {
    delegate := list.NewDefaultDelegate()
    delegate.SetHeight(4)
    
    l := list.New([]list.Item{}, delegate, 40, 10)
    l.SetShowTitle(true)
    l.Title = "Select New Agent"
    l.SetShowStatusBar(false)
    
    return &SwitchModal{
        list: l,
    }
}

// SetCards populates the card selection list.
func (m *SwitchModal) SetCards(cards []config.IDCard, skills []config.Skill, personas []config.Persona) {
    m.cards = cards
    items := make([]list.Item, len(cards))
    for i, card := range cards {
        items[i] = switchCardItem{
            card:     card,
            skills:   skills,
            personas: personas,
        }
    }
    m.list.SetItems(items)
}

// Open shows the modal.
func (m *SwitchModal) Open() {
    m.confirmed = false
    m.cancelled = false
}

// Close hides the modal.
func (m *SwitchModal) Close() {
    m.confirmed = false
    m.cancelled = false
}

// IsOpen returns whether modal is active.
func (m *SwitchModal) IsOpen() bool {
    return !m.cancelled && (m.confirmed || len(m.cards) > 0)
}

// Confirmed returns true if user confirmed selection.
func (m *SwitchModal) Confirmed() bool {
    return m.confirmed
}

// Cancelled returns true if user cancelled.
func (m *SwitchModal) Cancelled() bool {
    return m.cancelled
}

// SelectedCardID returns the selected card ID.
func (m *SwitchModal) SelectedCardID() string {
    if item := m.list.SelectedItem(); item != nil {
        if si, ok := item.(switchCardItem); ok {
            return si.card.ID
        }
    }
    return ""
}

// Init initializes the modal.
func (m *SwitchModal) Init() bubbletea.Cmd {
    return nil
}

// Update handles modal input.
func (m *SwitchModal) Update(msg bubbletea.Msg) (component.Model, bubbletea.Cmd) {
    if keyMsg, ok := msg.(bubbletea.KeyMsg); ok {
        switch keyMsg.String() {
        case "enter":
            m.confirmed = true
            return m, nil
        case "esc":
            m.cancelled = true
            return m, nil
        }
    }
    
    var cmd bubbletea.Cmd
    m.list, cmd = m.list.Update(msg)
    return m, cmd
}

// View renders the modal.
func (m *SwitchModal) View() string {
    boxStyle := lipgloss.NewStyle().
        Border(lipgloss.RoundedBorder()).
        BorderForeground(lipgloss.Color("#8E44AD")).
        Padding(1, 2).
        Width(50)
    
    return boxStyle.Render(m.list.View())
}

// switchCardItem wraps a card for display in switch modal.
type switchCardItem struct {
    card     config.IDCard
    skills   []config.Skill
    personas []config.Persona
}

func (i switchCardItem) Title() string       { return i.card.Name }
func (i switchCardItem) Description() string { return i.card.Driver + " | " + i.card.Model }
func (i switchCardItem) FilterValue() string { return i.card.Name }
```

---

## Step 3.1.3: Create Dashboard Tests

**File**: `internal/ui/views/dashboard_test.go` (NEW)

```go
package views

import (
    "testing"
    
    bubbletea "github.com/charmbracelet/bubbletea"
    
    "kankuro/internal/config"
    "kankuro/internal/session"
    "kankuro/internal/ui/commands"
)

func TestDashboard_New(t *testing.T) {
    callbacks := DashboardCallbacks{
        OnEditCard:   func(config.IDCard) {},
        OnLaunchCard: func(config.IDCard) {},
    }
    
    d := NewDashboard(callbacks, &commands.Commands{})
    
    if d == nil {
        t.Fatal("NewDashboard returned nil")
    }
    
    if d.focusedPane != PaneCards {
        t.Error("Dashboard should start with cards pane focused")
    }
}

func TestDashboard_SetData(t *testing.T) {
    d := NewDashboard(DashboardCallbacks{}, &commands.Commands{})
    
    cards := []config.IDCard{
        {ID: "c1", Name: "Card 1"},
    }
    sessions := []session.Snapshot{
        {ID: "s1", Name: "Session 1"},
    }
    
    d.SetData(cards, sessions, nil, nil, nil)
    
    if len(d.cards) != 1 {
        t.Errorf("Expected 1 card, got %d", len(d.cards))
    }
    if len(d.sessions) != 1 {
        t.Errorf("Expected 1 session, got %d", len(d.sessions))
    }
}

func TestDashboard_TogglePane(t *testing.T) {
    d := NewDashboard(DashboardCallbacks{}, &commands.Commands{})
    d.SetData(nil, []session.Snapshot{{ID: "s1"}}, nil, nil, nil)
    
    if d.focusedPane != PaneCards {
        t.Error("Should start on cards pane")
    }
    
    d.togglePane()
    if d.focusedPane != PaneSessions {
        t.Error("Should switch to sessions pane")
    }
    
    d.togglePane()
    if d.focusedPane != PaneCards {
        t.Error("Should switch back to cards pane")
    }
}

func TestDashboard_Search(t *testing.T) {
    d := NewDashboard(DashboardCallbacks{}, &commands.Commands{})
    
    // Activate search
    keyMsg := bubbletea.KeyMsg{Runes: []rune{'/'}}
    d.Update(keyMsg)
    
    if !d.search.IsActive {
        t.Error("Search should be active after pressing '/'")
    }
    
    // Type search query
    keyMsg = bubbletea.KeyMsg{Runes: []rune{'t', 'e', 's', 't'}}
    for _, r := range keyMsg.Runes {
        d.Update(bubbletea.KeyMsg{Runes: []rune{r}})
    }
    
    if d.search.CardsQuery != "test" {
        t.Errorf("Expected query 'test', got '%s'", d.search.CardsQuery)
    }
    
    // Exit search
    keyMsg = bubbletea.KeyMsg{Runes: []rune{}, Type: bubbletea.KeyEsc}
    d.Update(keyMsg)
    
    if d.search.IsActive {
        t.Error("Search should be inactive after escape")
    }
}

func TestDashboard_DeleteModal(t *testing.T) {
    var deletedSessionID string
    
    callbacks := DashboardCallbacks{
        OnRemove: func(sessionID string) error {
            deletedSessionID = sessionID
            return nil
        },
    }
    
    d := NewDashboard(callbacks, &commands.Commands{})
    d.SetData(nil, []session.Snapshot{
        {ID: "s1", UseTmpWorkspace: true},
    }, nil, nil, nil)
    d.focusedPane = PaneSessions
    
    // Trigger delete
    keyMsg := bubbletea.KeyMsg{Runes: []rune{}, Type: bubbletea.KeyBackspace}
    d.Update(keyMsg)
    
    if !d.deleteModal.IsOpen() {
        t.Error("Delete modal should be open")
    }
    
    // Confirm delete
    keyMsg = bubbletea.KeyMsg{Runes: []rune{'y'}}
    d.Update(keyMsg)
    
    if d.deleteModal.IsOpen() {
        t.Error("Delete modal should be closed after confirm")
    }
    if deletedSessionID != "s1" {
        t.Errorf("Expected deleted session 's1', got '%s'", deletedSessionID)
    }
}

func TestDashboard_SwitchModal(t *testing.T) {
    var switchedSessionID, switchedCardID string
    
    callbacks := DashboardCallbacks{
        OnSwitch: func(sessionID, newCardID string) error {
            switchedSessionID = sessionID
            switchedCardID = newCardID
            return nil
        },
    }
    
    d := NewDashboard(callbacks, &commands.Commands{})
    d.SetData(
        []config.IDCard{{ID: "c1", Name: "Card 1"}},
        []session.Snapshot{{ID: "s1", Status: session.StatusTerminated}},
        nil, nil, nil,
    )
    d.focusedPane = PaneSessions
    
    // Trigger switch
    keyMsg := bubbletea.KeyMsg{Runes: []rune{'s'}}
    d.Update(keyMsg)
    
    if !d.switchModal.IsOpen() {
        t.Error("Switch modal should be open")
    }
    
    // Confirm selection
    keyMsg = bubbletea.KeyMsg{Runes: []rune{}, Type: bubbletea.KeyEnter}
    d.Update(keyMsg)
    
    if d.switchModal.IsOpen() {
        t.Error("Switch modal should be closed after confirm")
    }
    if switchedSessionID != "s1" {
        t.Errorf("Expected session 's1', got '%s'", switchedSessionID)
    }
}

func TestDashboard_SetSize(t *testing.T) {
    d := NewDashboard(DashboardCallbacks{}, &commands.Commands{})
    d.SetSize(120, 40)
    
    if d.Width != 120 {
        t.Errorf("Expected width 120, got %d", d.Width)
    }
    if d.Height != 40 {
        t.Errorf("Expected height 40, got %d", d.Height)
    }
    
    // Verify proportions
    expectedCardsWidth := 120 / 4
    if d.cardsWidth != expectedCardsWidth {
        t.Errorf("Expected cards width %d, got %d", expectedCardsWidth, d.cardsWidth)
    }
}

func TestDashboard_View(t *testing.T) {
    d := NewDashboard(DashboardCallbacks{}, &commands.Commands{})
    d.SetData(
        []config.IDCard{{ID: "c1", Name: "Test Card", Driver: "claude", Model: "claude-3"}},
        []session.Snapshot{{ID: "s1", Name: "Test Session"}},
        nil, nil, nil,
    )
    d.SetSize(120, 40)
    
    view := d.View()
    
    if view == "" {
        t.Error("View should not be empty")
    }
    if !containsString(view, "K A N K U R O") {
        t.Error("View should contain header")
    }
    if !containsString(view, "ID Cards") {
        t.Error("View should contain cards pane title")
    }
    if !containsString(view, "Active Sessions") {
        t.Error("View should contain sessions pane title")
    }
}

func containsString(s, substr string) bool {
    return len(s) >= len(substr) && 
           (s == substr || findSubstring(s, substr))
}

func findSubstring(s, substr string) bool {
    for i := 0; i <= len(s)-len(substr); i++ {
        if s[i:i+len(substr)] == substr {
            return true
        }
    }
    return false
}
```

---

## Step 3.1.4: Integration with App (Backwards Compatible)

**File**: `internal/ui/app_dashboard.go` (MODIFY - add integration stub)

```go
// ===== Integration with new Dashboard view =====
// This shows how App will eventually delegate to Dashboard.
// Once fully migrated, the existing updateDashboard and renderDashboard
// methods can be removed.

func (a *App) newDashboardView() *views.Dashboard {
    callbacks := views.DashboardCallbacks{
        OnEditCard: func(card config.IDCard) {
            a.populateCardForm(card)
            a.mode = viewCardForm
        },
        OnLaunchCard: func(card config.IDCard) {
            a.naming.Card = card
            a.naming.Input.SetValue("")
            a.mode = viewNaming
        },
        OnAttach: func(sessionID string) tea.Cmd {
            a.sessionView.ActiveID = sessionID
            return a.attachSession(sessionID)
        },
        OnKill: func(sessionID string) error {
            return a.session.Kill(sessionID)
        },
        OnRemove: func(sessionID string) error {
            return a.session.Remove(sessionID)
        },
        OnSwitch: func(sessionID, newCardID string) error {
            // Find card and relaunch
            for _, card := range a.cards {
                if card.ID == newCardID {
                    skills := a.skillsForCard(card)
                    personas := a.personasForCard(card)
                    mcps := a.mcpsForCard(card)
                    return a.session.Relaunch(sessionID, card, skills, personas, mcps)
                }
            }
            return errors.New("card not found")
        },
        OnDeleteCard: func(cardID string) error {
            return a.repo.DeleteIDCard(cardID)
        },
        OnOpenMCP: func() {
            a.cardForm.MCPFormPrevMode = viewDashboard
            a.mode = viewMCPForm
        },
        OnOpenWorkspaces: func() {
            a.refreshWorkspaces()
            a.mode = viewWorkspaces
        },
        OnOpenIDE: func() {
            a.mode = viewIDE
        },
        OnOpenCardEditor: func() {
            a.resetCardForm()
            a.mode = viewCardForm
        },
    }
    
    return views.NewDashboard(callbacks, &a.cmds)
}

// For future use when fully migrated:
// dashboardView *views.Dashboard would be a field in App
// and New() would initialize it
```

---

## Verification Checklist

```bash
# 1. Build all packages
go build ./internal/ui/component/
go build ./internal/ui/messages/
go build ./internal/ui/commands/
go build ./internal/ui/components/
go build ./internal/ui/views/

# 2. Run all tests
go test ./internal/ui/views/ -v

# 3. Existing UI still builds
go build ./internal/ui/

# 4. Run existing tests
go test ./internal/ui/
```

---

## Success Criteria

- [ ] `Dashboard` struct defined in `internal/ui/views/dashboard.go`
- [ ] Composes CardList, SessionList, SessionPreview
- [ ] Pane switching (Tab key) works correctly
- [ ] Search/filter state managed correctly
- [ ] Delete confirmation modal works
- [ ] Switch agent modal works
- [ ] Layout proportions computed from size
- [ ] Header and footer rendered
- [ ] Empty sessions state rendered
- [ ] `SetData()` propagates to child components
- [ ] Unit tests pass (>80% coverage)
- [ ] Integration with App (backwards compatible)
- [ ] Existing App still compiles and tests pass

---

## Dependencies

```
component/component.go (from 1.1) ✓
messages/messages.go (from 1.2) ✓
commands/commands.go (from 1.3) ✓
components/modal.go (from 1.4) ✓
components/card_list.go (from 2.1) ✓
components/session_list.go (from 2.2) ✓
components/session_preview.go (from 2.3) ✓

views/dashboard.go ← NEW
views/switch_modal.go ← NEW
```

---

## Estimated Time

| Step | Time |
|------|------|
| 3.1.1 Dashboard view file | 60 min |
| 3.1.2 SwitchModal component | 30 min |
| 3.1.3 Unit tests | 45 min |
| 3.1.4 App integration | 20 min |
| Verification | 15 min |
| **Total** | **~2.5 hours** |

---

## What This Enables

1. **Main screen complete**: Dashboard is the primary user interface
2. **Modal pattern established**: Delete confirm and switch agent show overlay pattern
3. **Three-pane layout**: Responsive layout from size
4. **Phase 4 ready**: App coordinator can delegate to Dashboard
5. **Other views**: CardEditor (Task 3.2) can follow same composite view pattern
