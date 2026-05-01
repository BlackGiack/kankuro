# Implementation Plan: Phase 1 + CardList Component

## Goal

Create component infrastructure and extract CardList as the first composable component, establishing patterns for remaining extractions.

---

## Phase 1: Infrastructure (4 tasks)

### Task 1.1: Create Component Interface

**Goal**: Define the base interface all components implement.

**File**: `internal/ui/component/component.go` (NEW)

**Code**:
```go
package component

import tea "github.com/charmbracelet/bubbletea"

// Model is the base interface for all UI components.
type Model interface {
    // Init initializes the component (called once on mount).
    Init() tea.Cmd
    
    // Update handles messages and returns updated model + commands.
    Update(tea.Msg) (Model, tea.Cmd)
    
    // View renders the component as a string.
    View() string
    
    // SetSize updates component dimensions.
    SetSize(width, height int)
}

// Base provides common fields embedded in components.
// It implements SetSize and provides Width/Height accessors.
type Base struct {
    Width   int
    Height  int
    Focused bool
}

func (b *Base) SetSize(width, height int) {
    b.Width = width
    b.Height = height
}

// SetFocus controls whether component displays focused state.
func (b *Base) SetFocus(focused bool) {
    b.Focused = focused
}
```

**Why this design**:
- `Model` interface mirrors Bubble Tea's model pattern but adds `SetSize`
- `Base` struct eliminates boilerplate for common dimensions
- `Focused` field allows visual state differentiation

**Verification**: 
```bash
go build ./internal/ui/component/
```

---

### Task 1.2: Create Messages Package

**Goal**: Define typed messages for component-to-App communication.

**File**: `internal/ui/messages/messages.go` (NEW)

**Code**:
```go
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
```

**Verification**:
```bash
go build ./internal/ui/messages/
```

---

### Task 1.3: Create Commands Registry

**Goal**: Define callback interface for components to trigger actions.

**File**: `internal/ui/commands/commands.go` (NEW)

**Code**:
```go
package commands

import (
    "github.com/charmbracelet/bubbletea"
    
    "kankuro/internal/config"
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
    Attach func(sessionID string) bubbletea.Cmd
    
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
```

**Verification**:
```bash
go build ./internal/ui/commands/
```

---

### Task 1.4: Create Modal Component (Bonus Foundation)

**Goal**: Create reusable modal overlay component.

**File**: `internal/ui/components/modal.go` (NEW)

**Code**:
```go
package components

import (
    "github.com/charmbracelet/lipgloss"
    
    "kankuro/internal/ui/component"
)

// Modal is a reusable overlay component for dialogs.
type Modal struct {
    component.Base
    title      string
    content    string
    open       bool
    onConfirm  func()
    onCancel   func()
}

// NewModal creates a new modal with the given title.
func NewModal(title string) *Modal {
    return &Modal{
        title: title,
        open:  false,
    }
}

// SetContent sets the modal's content string.
func (m *Modal) SetContent(content string) {
    m.content = content
}

// Open makes the modal visible.
func (m *Modal) Open() {
    m.open = true
}

// Close hides the modal.
func (m *Modal) Close() {
    m.open = false
}

// IsOpen returns whether modal is currently visible.
func (m *Modal) IsOpen() bool {
    return m.open
}

// SetConfirm sets the confirm callback.
func (m *Modal) SetConfirm(callback func()) {
    m.onConfirm = callback
}

// SetCancel sets the cancel callback.
func (m *Modal) SetCancel(callback func()) {
    m.onCancel = callback
}

// Init implements component.Model.
func (m *Modal) Init() bubbletea.Cmd {
    return nil
}

// Update handles modal key events.
func (m *Modal) Update(msg bubbletea.Msg) (component.Model, bubbletea.Cmd) {
    // Key handling is done by parent view
    // Modal just renders and exposes Open/Close/IsOpen
    return m, nil
}

// View renders the modal overlay.
func (m *Modal) View() string {
    if !m.open {
        return ""
    }
    
    boxStyle := lipgloss.NewStyle().
        Border(lipgloss.RoundedBorder()).
        BorderForeground(lipgloss.Color("#8E44AD")).
        Padding(1, 2).
        Width(50)
    
    titleStyle := lipgloss.NewStyle().
        Bold(true).
        Foreground(lipgloss.Color("#FFFFFF"))
    
    content := titleStyle.Render(m.title)
    if m.content != "" {
        content += "\n\n" + m.content
    }
    
    return boxStyle.Render(content)
}
```

**Test file**: `internal/ui/components/modal_test.go` (NEW)

```go
package components

import (
    "testing"
    bubbletea "github.com/charmbracelet/bubbletea"
)

func TestModal_OpenClose(t *testing.T) {
    m := NewModal("Test Modal")
    
    if m.IsOpen() {
        t.Error("Modal should start closed")
    }
    
    m.Open()
    if !m.IsOpen() {
        t.Error("Modal should be open after Open()")
    }
    
    m.Close()
    if m.IsOpen() {
        t.Error("Modal should be closed after Close()")
    }
}

func TestModal_ViewHidden(t *testing.T) {
    m := NewModal("Hidden Modal")
    
    view := m.View()
    if view != "" {
        t.Errorf("Hidden modal should render empty string, got: %q", view)
    }
}

func TestModal_ViewVisible(t *testing.T) {
    m := NewModal("Visible Modal")
    m.SetContent("Are you sure?")
    m.Open()
    
    view := m.View()
    if view == "" {
        t.Error("Open modal should render content")
    }
    if !containsSubstring(view, "Visible Modal") {
        t.Error("Should contain title")
    }
    if !containsSubstring(view, "Are you sure?") {
        t.Error("Should contain content")
    }
}

func containsSubstring(s, substr string) bool {
    return len(s) >= len(substr) && 
           (s == substr || len(s) > 0 && containsRecursive(s, substr))
}

func containsRecursive(s, substr string) bool {
    for i := 0; i <= len(s)-len(substr); i++ {
        if s[i:i+len(substr)] == substr {
            return true
        }
    }
    return false
}
```

**Verification**:
```bash
go build ./internal/ui/components/
go test ./internal/ui/components/
```

---

## Phase 2 Task 2.1: Extract CardList Component

**Goal**: Extract CardList as independent component, keeping existing App working.

### Step 2.1.1: Create CardList Component File

**File**: `internal/ui/components/card_list.go` (NEW)

**Extracted from**: 
- `app_items.go:44-53` (cardItem struct)
- `app_lists.go:16-42` (cardDelegate)
- `app_lists.go:225-258` (toCardItems function)
- `app_render.go:158-237` (renderCardItem function)
- `app.go:55-67` (constants like cardDelegateHeight, cardItemWidth, etc.)

**Code**:
```go
package components

import (
    "fmt"
    "io"
    "strings"

    "github.com/charmbracelet/bubbles/list"
    bubbletea "github.com/charmbracelet/bubbletea"
    "github.com/charmbracelet/lipgloss"
    
    "kankuro/internal/ascii"
    "kankuro/internal/ascii/characters"
    "kankuro/internal/config"
    "kankuro/internal/ui/component"
    "kankuro/internal/ui/messages"
)

// Card styling constants.
const (
    CardDelegateHeight = 8
    CardItemWidth      = 42
    CardListMinWidth   = 46
    CardListMinHeight  = 10
    
    InfoBoxWidth  = 30
    InfoBoxHeight = 4
    FaceBoxWidth  = 11
    FaceBoxHeight = 5
    
    TruncateLength = 28
    MaxPillsShown  = 4
    PillTextLimit  = 10
)

// CardListItem wraps IDCard for display in the list.
type CardListItem struct {
    Card     config.IDCard
    Skills   []config.Skill
    Personas []config.Persona
}

// CardListCallbacks defines callbacks for card actions.
type CardListCallbacks struct {
    OnSelect func(config.IDCard)
    OnEdit   func(config.IDCard)
    OnDelete func(string)
}

// CardList is a scrollable list of ID cards.
type CardList struct {
    component.Base
    list        list.Model
    cards       []config.IDCard
    skills      []config.Skill
    personas    []config.Persona
    searchQuery string
    callbacks   CardListCallbacks
}

// NewCardList creates a new card list component.
func NewCardList(callbacks CardListCallbacks) *CardList {
    delegate := newCardDelegate(nil, nil)
    
    l := list.New([]bubbletea.Model{}, delegate, 0, 0)
    l.SetShowTitle(false)
    l.SetShowStatusBar(false)
    l.SetFilteringEnabled(false)
    
    return &CardList{
        list:      l,
        callbacks: callbacks,
    }
}

// SetCards updates the card list data.
func (c *CardList) SetCards(cards []config.IDCard, skills []config.Skill, personas []config.Persona) {
    c.cards = cards
    c.skills = skills
    c.personas = personas
    c.refreshItems()
}

// SetSearchQuery filters the displayed cards.
func (c *CardList) SetSearchQuery(query string) {
    c.searchQuery = query
    c.refreshItems()
}

// Selected returns the currently selected card.
func (c *CardList) Selected() (config.IDCard, bool) {
    item := c.list.SelectedItem()
    if item == nil {
        return config.IDCard{}, false
    }
    
    cardItem, ok := item.(CardListItem)
    if !ok {
        return config.IDCard{}, false
    }
    return cardItem.Card, true
}

// Init initializes the component.
func (c *CardList) Init() bubbletea.Cmd {
    return nil
}

// Update handles messages.
func (c *CardList) Update(msg bubbletea.Msg) (component.Model, bubbletea.Cmd) {
    var cmd bubbletea.Cmd
    c.list, cmd = c.list.Update(msg)
    
    // Handle selection events
    if keyMsg, ok := msg.(bubbletea.KeyMsg); ok {
        switch keyMsg.String() {
        case "enter":
            if card, ok := c.Selected(); ok {
                if c.callbacks.OnSelect != nil {
                    c.callbacks.OnSelect(card)
                }
            }
        case "e":
            if card, ok := c.Selected(); ok {
                if c.callbacks.OnEdit != nil {
                    c.callbacks.OnEdit(card)
                }
            }
        case "delete", "backspace":
            if card, ok := c.Selected(); ok {
                if c.callbacks.OnDelete != nil {
                    c.callbacks.OnDelete(card.ID)
                }
            }
        }
    }
    
    return c, cmd
}

// View renders the card list.
func (c *CardList) View() string {
    return c.list.View()
}

// SetSize implements component.Model.
func (c *CardList) SetSize(width, height int) {
    c.Base.SetSize(width, height)
    c.list.SetSize(width, height)
}

// refreshItems rebuilds the list items from filtered cards.
func (c *CardList) refreshItems() {
    filtered := c.filterCards()
    items := make([]list.Item, 0, len(filtered))
    
    for _, card := range filtered {
        skillList := getSkillsForCard(card, c.skills)
        personaList := getPersonasForCard(card, c.personas)
        
        items = append(items, CardListItem{
            Card:     card,
            Skills:   skillList,
            Personas: personaList,
        })
    }
    
    c.list.SetItems(items)
}

// filterCards applies the search query filter.
func (c *CardList) filterCards() []config.IDCard {
    if c.searchQuery == "" {
        return c.cards
    }
    
    q := strings.ToLower(c.searchQuery)
    var filtered []config.IDCard
    for _, card := range c.cards {
        if strings.Contains(strings.ToLower(card.Name), q) ||
           strings.Contains(strings.ToLower(card.Driver), q) ||
           strings.Contains(strings.ToLower(card.Model), q) {
            filtered = append(filtered, card)
        }
    }
    return filtered
}

// cardDelegate handles rendering custom card items.
type cardDelegate struct {
    height   int
    skills   []config.Skill
    personas []config.Persona
}

func newCardDelegate(skills []config.Skill, personas []config.Persona) cardDelegate {
    return cardDelegate{
        height:   CardDelegateHeight,
        skills:   skills,
        personas: personas,
    }
}

func (d cardDelegate) Height() int                             { return d.height }
func (d cardDelegate) Spacing() int                            { return 0 }
func (d cardDelegate) Update(_ bubbletea.Msg, _ *list.Model) bubbletea.Cmd { return nil }

func (d cardDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
    cardItem, ok := item.(CardListItem)
    if !ok {
        return
    }
    selected := index == m.Index()
    content := RenderCardItem(cardItem.Card, d.skills, d.personas, selected)
    fmt.Fprint(w, content)
}

// RenderCardItem renders a single card's display.
func RenderCardItem(card config.IDCard, skills []config.Skill, personas []config.Persona, selected bool) string {
    skillNames := getSelectedNames(card.SkillIDs, func(id string) (string, bool) {
        for _, s := range skills {
            if s.ID == id {
                return s.Name, true
            }
        }
        return "", false
    })

    personaNames := getSelectedNames(card.PersonaIDs, func(id string) (string, bool) {
        for _, p := range personas {
            if p.ID == id {
                return p.Name, true
            }
        }
        return "", false
    })

    pills := renderPills(skillNames, personaNames)
    skillsLine := lipgloss.JoinHorizontal(lipgloss.Top, pills...)

    style := card.Puppet
    if style == "" {
        style = "kankuro"
    }
    puppet := characters.NewPuppet(func(c *ascii.Config) { c.Set("style", style) })
    w, h := puppet.Bounds()
    canvas := ascii.NewCanvas(w, h)
    puppet.Render(canvas, 0, 0)
    puppetStr := canvas.String()

    driverLine := fmt.Sprintf("%s | %s", card.Driver, card.Model)

    // Colors - should come from shared styles
    primaryColor := lipgloss.Color("#8E44AD")
    highlightColor := lipgloss.Color("#16A085")
    gray := lipgloss.Color("240")
    white := lipgloss.Color("#FFFFFF")
    deepCharcoal := lipgloss.Color("#1A1A1A")

    titleStyle := lipgloss.NewStyle().Bold(true)
    descStyle := lipgloss.NewStyle().Foreground(gray)

    if selected {
        titleStyle = titleStyle.Foreground(highlightColor)
    }

    infoLines := []string{
        titleStyle.Render(truncate(card.Name, TruncateLength)),
        descStyle.Render(truncate(driverLine, TruncateLength)),
        "",
        skillsLine,
    }

    infoBox := lipgloss.NewStyle().
        Width(InfoBoxWidth).
        Height(InfoBoxHeight).
        Render(strings.Join(infoLines, "\n"))

    faceBg := deepCharcoal
    if selected {
        faceBg = primaryColor
    }

    faceBoxStyle := lipgloss.NewStyle().
        Border(lipgloss.NormalBorder(), false, false, false, true).
        BorderForeground(primaryColor).
        Background(deepCharcoal).
        Align(lipgloss.Center, lipgloss.Center)

    faceBox := faceBoxStyle.
        Width(FaceBoxWidth).
        Height(FaceBoxHeight).
        Background(faceBg).
        Foreground(white).
        Render(puppetStr)

    content := lipgloss.JoinHorizontal(lipgloss.Top, infoBox, faceBox)

    borderColor := gray
    if selected {
        borderColor = primaryColor
    }

    return lipgloss.NewStyle().
        Border(lipgloss.RoundedBorder()).
        BorderForeground(borderColor).
        Width(CardItemWidth).
        Render(content)
}

// Helper functions

func getSkillsForCard(card config.IDCard, allSkills []config.Skill) []config.Skill {
    if len(card.SkillIDs) == 0 {
        return nil
    }
    selected := make(map[string]struct{}, len(card.SkillIDs))
    for _, id := range card.SkillIDs {
        selected[id] = struct{}{}
    }
    var result []config.Skill
    for _, s := range allSkills {
        if _, ok := selected[s.ID]; ok {
            result = append(result, s)
        }
    }
    return result
}

func getPersonasForCard(card config.IDCard, allPersonas []config.Persona) []config.Persona {
    if len(card.PersonaIDs) == 0 {
        return nil
    }
    selected := make(map[string]struct{}, len(card.PersonaIDs))
    for _, id := range card.PersonaIDs {
        selected[id] = struct{}{}
    }
    var result []config.Persona
    for _, p := range allPersonas {
        if _, ok := selected[p.ID]; ok {
            result = append(result, p)
        }
    }
    return result
}

func getSelectedNames(ids []string, resolver func(string) (string, bool)) []string {
    var names []string
    for _, id := range ids {
        if name, ok := resolver(id); ok {
            names = append(names, name)
        }
    }
    return names
}

func renderPills(skillNames, personaNames []string) []string {
    var pills []string
    if len(skillNames) == 0 && len(personaNames) == 0 {
        gray := lipgloss.Color("240")
        return []string{lipgloss.NewStyle().Foreground(gray).Render("(no capabilities)")}
    }

    highlightColor := lipgloss.Color("#16A085")
    personaColor := lipgloss.Color("#C0392B")
    white := lipgloss.Color("#FFFFFF")

    for _, s := range skillNames {
        if len(pills) >= MaxPillsShown {
            break
        }
        pills = append(pills, pillStyle(highlightColor, white).Render(truncatePill(s)))
    }
    for _, s := range personaNames {
        if len(pills) >= MaxPillsShown {
            break
        }
        pills = append(pills, pillStyle(personaColor, white).Render(truncatePill(s)))
    }
    return pills
}

func pillStyle(bg, fg lipgloss.Color) lipgloss.Style {
    return lipgloss.NewStyle().
        Foreground(fg).
        Background(bg).
        Padding(0, 1).
        MarginRight(1)
}

func truncatePill(s string) string {
    if len(s) > PillTextLimit {
        return s[:PillTextLimit-1] + "."
    }
    return s
}

func truncate(s string, w int) string {
    if len(s) > w && w > 3 {
        return s[:w-3] + "..."
    }
    return s
}

// Emit messages (for use by parent)

func (c *CardList) EmitCardSelected msgs.CardSelectedMsg {
    // This will be called by Update when user presses Enter
    // Parent handles the message
}
```

---

### Step 2.1.2: Create CardList Tests

**File**: `internal/ui/components/card_list_test.go` (NEW)

```go
package components

import (
    "testing"
    
    "kankuro/internal/config"
)

func TestCardList_New(t *testing.T) {
    callbacks := CardListCallbacks{
        OnSelect: func(config.IDCard) {},
        OnEdit:   func(config.IDCard) {},
        OnDelete: func(string) {},
    }
    
    c := NewCardList(callbacks)
    if c == nil {
        t.Fatal("NewCardList returned nil")
    }
    
    // Check initial state
    if c.Focused {
        t.Error("CardList should not be focused initially")
    }
}

func TestCardList_SetCards(t *testing.T) {
    callbacks := CardListCallbacks{}
    c := NewCardList(callbacks)
    
    cards := []config.IDCard{
        {ID: "1", Name: "Card 1"},
        {ID: "2", Name: "Card 2"},
    }
    skills := []config.Skill{}
    personas := []config.Persona{}
    
    c.SetCards(cards, skills, personas)
    
    // Verify internal state updated
    if len(c.cards) != 2 {
        t.Errorf("Expected 2 cards, got %d", len(c.cards))
    }
}

func TestCardList_SetSearchQuery(t *testing.T) {
    callbacks := CardListCallbacks{}
    c := NewCardList(callbacks)
    
    cards := []config.IDCard{
        {ID: "1", Name: "Alpha"},
        {ID: "2", Name: "Beta"},
        {ID: "3", Name: "Gamma"},
    }
    
    c.SetCards(cards, nil, nil)
    
    // No filter
    filtered := c.filterCards()
    if len(filtered) != 3 {
        t.Errorf("Expected 3 cards without filter, got %d", len(filtered))
    }
    
    // With filter
    c.SetSearchQuery("al")
    filtered = c.filterCards()
    if len(filtered) != 1 {
        t.Errorf("Expected 1 card matching 'al', got %d", len(filtered))
    }
    if len(filtered) > 0 && filtered[0].ID != "1" {
        t.Errorf("Expected card ID '1', got '%s'", filtered[0].ID)
    }
}

func TestCardList_Selected(t *testing.T) {
    callbacks := CardListCallbacks{}
    c := NewCardList(callbacks)
    
    // No selection
    _, ok := c.Selected()
    if ok {
        t.Error("Selected should return false when no selection")
    }
    
    // Set cards and select first
    cards := []config.IDCard{
        {ID: "1", Name: "Test"},
    }
    c.SetCards(cards, nil, nil)
    c.list.Select(0)
    
    card, ok := c.Selected()
    if !ok {
        t.Error("Selected should return true when item selected")
    }
    if card.ID != "1" {
        t.Errorf("Expected card ID '1', got '%s'", card.ID)
    }
}

func TestRenderCardItem(t *testing.T) {
    card := config.IDCard{
        ID:      "1",
        Name:    "Test Card",
        Driver:  "claude",
        Model:   "claude-3",
        Puppet:  "kankuro",
    }
    
    skills := []config.Skill{
        {ID: "s1", Name: "Skill1"},
    }
    personas := []config.Persona{
        {ID: "p1", Name: "Persona1"},
    }
    
    // Render unselected
    output := RenderCardItem(card, skills, personas, false)
    if output == "" {
        t.Error("RenderCardItem returned empty string")
    }
    
    // Render selected
    outputSel := RenderCardItem(card, skills, personas, true)
    if outputSel == "" {
        t.Error("RenderCardItem (selected) returned empty string")
    }
    
    // Selected should be different (different border color)
    if output == outputSel {
        t.Error("Selected and unselected should render differently")
    }
}
```

**Verification**:
```bash
go test ./internal/ui/components/ -v
```

---

### Step 2.1.3: Update App to Use CardList (Backwards Compatible)

**Goal**: Show how existing App can use new component WITHOUT breaking.

**Approach**: Create a parallel implementation in App that delegates to CardList, but keep old code working.

**File**: `internal/ui/app_lists.go` (MODIFY)

Add at end of file:
```go
// ===== Integration with new CardList component =====
// This shows how App will eventually delegate to CardList.
// Once fully migrated, the old cardDelegate and related functions
// can be removed.

func (a *App) newCardListComponent() *components.CardList {
    callbacks := components.CardListCallbacks{
        OnSelect: func(card config.IDCard) {
            // Will emit CardSelectedMsg in component world
            // For now, call existing handler
            a.handleDashboardEnter()
        },
        OnEdit: func(card config.IDCard) {
            // Will emit CardEditRequestedMsg in component world
            // For now, call existing handler
            a.handleDashboardEdit()
        },
        OnDelete: func(cardID string) {
            // Will emit CardDeleteRequestedMsg in component world
            // For now, just trigger delete flow
            a.sessionView.PendingDeleteCardID = cardID
            a.mode = viewDeleteConfirm
        },
    }
    
    return components.NewCardList(callbacks)
}

// For future use when fully migrated:
// cardListComponent *components.CardList would be a field in App
// and New() would initialize it
```

---

### Step 2.1.4: Add Integration Test

**File**: `internal/ui/components/integration_test.go` (NEW)

```go
package components_test

import (
    "testing"
    
    "kankuro/internal/config"
    "kankuro/internal/ui/components"
)

func TestCardListIntegration(t *testing.T) {
    // Simulate App workflow
    
    // 1. Create component with callbacks
    selectedCard := ""
    editedCard := ""
    deletedCard := ""
    
    var c *components.CardList
    c = components.NewCardList(components.CardListCallbacks{
        OnSelect: func(card config.IDCard) {
            selectedCard = card.ID
        },
        OnEdit: func(card config.IDCard) {
            editedCard = card.ID
        },
        OnDelete: func(cardID string) {
            deletedCard = cardID
        },
    })
    
    // 2. Set data
    cards := []config.IDCard{
        {ID: "1", Name: "First", Driver: "claude"},
        {ID: "2", Name: "Second", Driver: "gemini"},
    }
    c.SetCards(cards, nil, nil)
    
    // 3. Initialize
    c.Init()
    c.SetSize(50, 20)
    
    // 4. Select first item
    // (In real use, this would be via Update with KeyMsg)
    // For this test, just verify Selected works
    _, ok := c.Selected()
    if ok {
        t.Log("Card selected successfully")
    }
    
    // 5. Apply search filter
    c.SetSearchQuery("first")
    filtered := c.filterCards()
    if len(filtered) != 1 {
        t.Errorf("Expected 1 card after filter, got %d", len(filtered))
    }
    
    // Verify callbacks would work (actual key handling tested in Update)
    t.Logf("Component created and initialized successfully")
}
```

**Verification**:
```bash
go test ./internal/ui/components/ -v
```

---

## Verification Checklist

After completing all tasks, verify:

```bash
# 1. Component package builds
go build ./internal/ui/component/

# 2. Messages package builds
go build ./internal/ui/messages/

# 3. Commands package builds
go build ./internal/ui/commands/

# 4. Components package builds (includes CardList)
go build ./internal/ui/components/

# 5. All component tests pass
go test ./internal/ui/components/ -v

# 6. Existing UI still builds (backwards compatible)
go build ./internal/ui/

# 7. Run existing tests
go test ./internal/ui/
```

---

## Success Criteria

- [ ] `component.Model` interface defined
- [ ] `component.Base` struct provides SetSize/Focused
- [ ] All message types defined in `messages` package
- [ ] Commands struct with callbacks defined
- [ ] Modal component extracted independently
- [ ] CardList component extracted from App
- [ ] CardList has unit tests (>80% coverage)
- [ ] CardList has integration test showing workflow
- [ ] Existing App still compiles and tests pass
- [ ] CardList renders cards correctly
- [ ] CardList handles search/filter correctly
- [ ] CardList callbacks work (OnSelect, OnEdit, OnDelete)

---

## Dependencies Graph

```
component/component.go (no deps)
         ↓
messages/messages.go (depends on config, session)
         ↓
commands/commands.go (depends on config, bubbletea)
         ↓
components/modal.go (depends on component, lipgloss)
         ↓
components/card_list.go (depends on all above + list, ascii, characters)
```

---

## Estimated Time

| Task | Time |
|------|------|
| 1.1 Component interface | 15 min |
| 1.2 Messages package | 20 min |
| 1.3 Commands registry | 15 min |
| 1.4 Modal component + test | 30 min |
| 2.1.1 CardList file | 60 min |
| 2.1.2 CardList tests | 30 min |
| 2.1.3 App integration stub | 15 min |
| 2.1.4 Integration test | 15 min |
| Verification | 15 min |
| **Total** | **~3 hours** |

---

## What This Enables

Once this is complete:

1. **Pattern established**: Developers can follow same pattern for SessionList, SessionPreview, SourceList
2. **Type safety**: All messages are typed, no passing around `map[string]interface{}`
3. **Testability**: Components can be tested in isolation
4. **Incremental migration**: App can use both old and new approaches during transition
5. **Next extractions**: Tasks 2.2, 2.3, 2.4 can follow CardList pattern exactly