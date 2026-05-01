# Implementation Plan: Phase 2 Task 2.2 - SessionList Component

## Goal

Extract SessionList as independent component following CardList pattern, establishing consistency for remaining extractions.

---

## Analysis: What to Extract

### From Existing Code

| File | Lines | Content | Destination |
|------|-------|---------|-------------|
| `app_items.go` | 10-42 | `sessionListItem` struct + methods | `components/session_list.go` |
| `app_new.go` | 74-88 | Delegate styling setup | `components/session_list.go` |
| `app_lists.go` | 89-143 | `refreshSessionList()` logic | `components/session_list.go` |
| `app_lists.go` | 161-172 | `selectedSession()` | `components/session_list.go` |
| `app.go` | 80 | `sessionsDelegateHeight = 4` | `components/session_list.go` |
| `app.go` | 59 | `sessionsListHeight = 7` | Keep in App (layout constant) |

### SessionList Specifics

**Unlike CardList**:
- Uses **default delegate** with custom styling (not custom delegate)
- Renders **face style** from card reference (requires cards data)
- Has **Waiting status** special highlighting (hot pink)
- Triggers **preview sync** on selection change
- Shows **ID or working directory** based on tmp workspace flag

**Key constants**:
```go
sessionsDelegateHeight = 4
```

**Face rendering**:
```go
// Already extracted in CardList
renderAgentCard(config.IDCard{FaceStyle: faceStyle})
```

---

## Step 2.2.1: Create SessionList Component File

**File**: `internal/ui/components/session_list.go` (NEW)

```go
package components

import (
    "fmt"
    "sort"
    "strings"

    "github.com/charmbracelet/bubbles/list"
    bubbletea "github.com/charmbracelet/bubbletea"
    "github.com/charmbracelet/lipgloss"
    
    "kankuro/internal/config"
    "kankuro/internal/session"
    "kankuro/internal/ui/component"
    "kankuro/internal/ui/messages"
)

// Session styling constants.
const (
    SessionDelegateHeight = 4
)

// SessionListItem wraps session.Snapshot for list display.
type SessionListItem struct {
    Session session.Snapshot
    Card    *config.IDCard // Reference to card for face style
}

// SessionListCallbacks defines callbacks for session actions.
type SessionListCallbacks struct {
    OnSelect   func(session.Snapshot)
    OnAttach   func(sessionID string)
    OnKill     func(sessionID string)
    OnDelete   func(sessionID string)
    OnSwitch   func(sessionID string)
    OnPreview  func(sessionID string) // Called when selection changes
}

// SessionList is a scrollable list of active sessions.
type SessionList struct {
    component.Base
    list        list.Model
    sessions    []session.Snapshot
    cards       []config.IDCard
    searchQuery string
    callbacks   SessionListCallbacks
    
    // Theme colors (should come from shared styles)
    primaryColor   lipgloss.Color
    highlightColor lipgloss.Color
    gray           lipgloss.Color
}

// NewSessionList creates a new session list component.
func NewSessionList(callbacks SessionListCallbacks) *SessionList {
    // Create default delegate with custom styling
    delegate := list.NewDefaultDelegate()
    delegate.SetHeight(SessionDelegateHeight)
    
    // Apply custom styles
    delegate.Styles.NormalTitle = delegate.Styles.NormalTitle.Foreground(lipgloss.Color("#8E44AD"))
    delegate.Styles.NormalDesc = delegate.Styles.NormalDesc.Foreground(lipgloss.Color("240"))
    
    delegate.Styles.SelectedTitle = delegate.Styles.SelectedTitle.
        Foreground(lipgloss.Color("#16A085")).
        BorderForeground(lipgloss.Color("#8E44AD")).
        UnsetBackground()
    delegate.Styles.SelectedDesc = delegate.Styles.SelectedDesc.
        Foreground(lipgloss.Color("#16A085")).
        BorderForeground(lipgloss.Color("#8E44AD")).
        UnsetBackground()
    
    delegate.Styles.DimmedTitle = delegate.Styles.DimmedTitle.Foreground(lipgloss.Color("240"))
    delegate.Styles.DimmedDesc = delegate.Styles.DimmedDesc.Foreground(lipgloss.Color("240"))
    
    l := list.New([]list.Item{}, delegate, 0, 0)
    l.SetShowTitle(false)
    l.SetShowStatusBar(false)
    l.SetFilteringEnabled(false)
    
    return &SessionList{
        list:            l,
        callbacks:       callbacks,
        primaryColor:    lipgloss.Color("#8E44AD"),
        highlightColor:  lipgloss.Color("#16A085"),
        gray:            lipgloss.Color("240"),
    }
}

// SetSessions updates the session list data.
func (s *SessionList) SetSessions(sessions []session.Snapshot, cards []config.IDCard) {
    s.sessions = sessions
    s.cards = cards
    s.refreshItems()
}

// SetSearchQuery filters the displayed sessions.
func (s *SessionList) SetSearchQuery(query string) {
    s.searchQuery = query
    s.refreshItems()
}

// Selected returns the currently selected session.
func (s *SessionList) Selected() (session.Snapshot, bool) {
    item := s.list.SelectedItem()
    if item == nil {
        return session.Snapshot{}, false
    }
    
    sessionItem, ok := item.(SessionListItem)
    if !ok {
        return session.Snapshot{}, false
    }
    return sessionItem.Session, true
}

// SetSelectedByID selects a session by ID.
func (s *SessionList) SetSelectedByID(sessionID string) {
    for i, item := range s.list.Items() {
        if si, ok := item.(SessionListItem); ok {
            if si.Session.ID == sessionID {
                s.list.Select(i)
                return
            }
        }
    }
}

// Init initializes the component.
func (s *SessionList) Init() bubbletea.Cmd {
    return nil
}

// Update handles messages.
func (s *SessionList) Update(msg bubbletea.Msg) (component.Model, bubbletea.Cmd) {
    var cmd bubbletea.Cmd
    
    // Track previous selection for preview sync
    prevIndex := s.list.Index()
    
    s.list, cmd = s.list.Update(msg)
    
    // Handle selection change
    if s.list.Index() != prevIndex {
        if sel, ok := s.Selected(); ok {
            if s.callbacks.OnPreview != nil {
                s.callbacks.OnPreview(sel.ID)
            }
        }
    }
    
    // Handle key events
    if keyMsg, ok := msg.(bubbletea.KeyMsg); ok {
        switch keyMsg.String() {
        case "enter":
            if session, ok := s.Selected(); ok {
                if session.Status == session.StatusTerminated {
                    // Relaunch/attach
                    if s.callbacks.OnAttach != nil {
                        s.callbacks.OnAttach(session.ID)
                    }
                } else {
                    // Attach to running
                    if s.callbacks.OnAttach != nil {
                        s.callbacks.OnAttach(session.ID)
                    }
                }
            }
        case "x":
            if _, ok := s.Selected(); ok {
                if s.callbacks.OnKill != nil {
                    s.callbacks.OnKill(s.list.Index())
                }
            }
        case "delete", "backspace":
            if _, ok := s.Selected(); ok {
                if s.callbacks.OnDelete != nil {
                    s.callbacks.OnDelete(s.list.Index())
                }
            }
        case "s":
            if _, ok := s.Selected(); ok {
                if s.callbacks.OnSwitch != nil {
                    s.callbacks.OnSwitch(s.list.Index())
                }
            }
        }
    }
    
    return s, cmd
}

// View renders the session list.
func (s *SessionList) View() string {
    return s.list.View()
}

// SetSize implements component.Model.
func (s *SessionList) SetSize(width, height int) {
    s.Base.SetSize(width, height)
    s.list.SetSize(width, height)
}

// refreshItems rebuilds the list items from sessions.
func (s *SessionList) refreshItems() {
    // Sort sessions by StartedAt (newest first)
    sorted := make([]session.Snapshot, len(s.sessions))
    copy(sorted, s.sessions)
    sort.Slice(sorted, func(i, j int) bool {
        return sorted[i].StartedAt.After(sorted[j].StartedAt)
    })
    
    // Apply search filter
    filtered := s.filterSessions(sorted)
    
    // Convert to list items
    items := make([]list.Item, 0, len(filtered))
    for _, sess := range filtered {
        // Find card for face style
        var card *config.IDCard
        var cardName string
        for i := range s.cards {
            if s.cards[i].ID == sess.CardID {
                card = &s.cards[i]
                cardName = card.Name
                break
            }
        }
        if cardName == "" {
            cardName = sess.CardID
        }
        
        // Determine face style (neutral if waiting)
        faceStyle := "neutral"
        if sess.Status != session.StatusWaiting && card != nil && card.FaceStyle != "" {
            faceStyle = card.FaceStyle
        }
        
        items = append(items, SessionListItem{
            Session: sess,
            Card:    card,
        })
    }
    
    s.list.SetItems(items)
}

// filterSessions applies the search query filter.
func (s *SessionList) filterSessions(sessions []session.Snapshot) []session.Snapshot {
    if s.searchQuery == "" {
        return sessions
    }
    
    q := strings.ToLower(s.searchQuery)
    var filtered []session.Snapshot
    for _, sess := range sessions {
        if strings.Contains(strings.ToLower(sess.Name), q) ||
           strings.Contains(strings.ToLower(sess.CardID), q) {
            filtered = append(filtered, sess)
        }
    }
    return filtered
}

// ===== SessionListItem implements list.Item =====

func (i SessionListItem) Title() string {
    if i.Session.Status == session.StatusWaiting {
        return lipgloss.NewStyle().
            Foreground(lipgloss.Color("#FF69B4")).
            Bold(true).
            Render(i.Session.Name)
    }
    return i.Session.Name
}

func (i SessionListItem) Description() string {
    // Find card name for display
    cardName := i.Session.CardID
    if i.Card != nil {
        cardName = i.Card.Name
    }
    
    // Render face
    faceStyle := "neutral"
    if i.Session.Status != session.StatusWaiting && i.Card != nil && i.Card.FaceStyle != "" {
        faceStyle = i.Card.FaceStyle
    }
    face := RenderAgentFace(config.IDCard{FaceStyle: faceStyle})
    
    cardLine := cardName + " " + face
    descLine := fmt.Sprintf("%s pid=%d", i.Session.Status, i.Session.PID)
    
    // Session detail (ID or working dir)
    var detailLine string
    if i.Session.UseTmpWorkspace {
        detailLine = "ID: " + i.Session.ID
    } else {
        wsPath := i.Session.WorkingDir
        if len(wsPath) > 25 {
            wsPath = "..." + wsPath[len(wsPath)-22:]
        }
        detailLine = "WS: " + wsPath
    }
    
    // Apply waiting status styling
    if i.Session.Status == session.StatusWaiting {
        cardLine = lipgloss.NewStyle().
            Foreground(lipgloss.Color("#FF69B4")).
            Bold(true).
            Render(cardLine)
        descLine = lipgloss.NewStyle().
            Foreground(lipgloss.Color("#16A085")).
            Render(descLine)
    }
    
    return cardLine + "\n" + descLine + "\n" + detailLine
}

func (i SessionListItem) FilterValue() string {
    return i.Session.Name
}

// RenderAgentFace renders the ASCII face for a session.
// Uses the already-extracted function pattern from CardList.
func RenderAgentFace(card config.IDCard) string {
    // Import from ascii package - same pattern as CardList
    // This can be shared via a common renderer in the future
    style := card.FaceStyle
    if style == "" {
        style = "neutral"
    }
    
    // Use ascii package directly
    // (Import path: kankuro/internal/ascii and kankuro/internal/ascii/characters)
    // This mirrors renderAgentCard from app_render.go:292-301
    return "" // Will call characters.NewFace with style
}
```

---

## Step 2.2.2: Add Shared Face Renderer

Since both CardList and SessionList need to render agent faces, extract to shared location.

**File**: `internal/ui/components/renderer.go` (NEW)

```go
package components

import (
    "kankuro/internal/ascii"
    "kankuro/internal/ascii/characters"
    "kankuro/internal/config"
)

// RenderAgentFace renders ASCII face for a session/card.
func RenderAgentFace(faceStyle string) string {
    if faceStyle == "" {
        faceStyle = "neutral"
    }
    
    face := characters.NewFace(ascii.WithStyle(faceStyle))
    w, _ := face.Bounds()
    canvas := ascii.NewCanvas(w, 1)
    face.Render(canvas, 0, 0)
    return canvas.String()
}

// RenderAgentPuppet renders ASCII puppet for a card.
func RenderAgentPuppet(puppetStyle string) string {
    if puppetStyle == "" {
        puppetStyle = "kankuro"
    }
    
    puppet := characters.NewPuppet(func(c *ascii.Config) { c.Set("style", puppetStyle) })
    w, h := puppet.Bounds()
    canvas := ascii.NewCanvas(w, h)
    puppet.Render(canvas, 0, 0)
    return canvas.String()
}
```

**Update CardList to use shared renderer**:

In `card_list.go`, replace puppet rendering with:
```go
puppetStr := RenderAgentPuppet(card.Puppet)
```

---

## Step 2.2.3: Create SessionList Tests

**File**: `internal/ui/components/session_list_test.go` (NEW)

```go
package components

import (
    "sort"
    "testing"
    "time"
    
    "kankuro/internal/config"
    "kankuro/internal/session"
)

func TestSessionList_New(t *testing.T) {
    callbacks := SessionListCallbacks{
        OnSelect: func(session.Snapshot) {},
        OnAttach: func(string) {},
        OnKill: func(string) {},
        OnDelete: func(string) {},
    }
    
    sl := NewSessionList(callbacks)
    if sl == nil {
        t.Fatal("NewSessionList returned nil")
    }
    
    if sl.Focused {
        t.Error("SessionList should not be focused initially")
    }
}

func TestSessionList_SetSessions(t *testing.T) {
    callbacks := SessionListCallbacks{}
    sl := NewSessionList(callbacks)
    
    sessions := []session.Snapshot{
        {ID: "s1", Name: "Session 1", Status: session.StatusRunning},
        {ID: "s2", Name: "Session 2", Status: session.StatusTerminated},
    }
    cards := []config.IDCard{
        {ID: "c1", Name: "Card 1"},
    }
    
    sl.SetSessions(sessions, cards)
    
    // Verify internal state updated
    if len(sl.sessions) != 2 {
        t.Errorf("Expected 2 sessions, got %d", len(sl.sessions))
    }
    
    // Verify sorting (newest first)
    sorted := make([]session.Snapshot, len(sessions))
    copy(sorted, sessions)
    sort.Slice(sorted, func(i, j int) bool {
        return sorted[i].StartedAt.After(sorted[j].StartedAt)
    })
}

func TestSessionList_SetSearchQuery(t *testing.T) {
    callbacks := SessionListCallbacks{}
    sl := NewSessionList(callbacks)
    
    sessions := []session.Snapshot{
        {ID: "s1", Name: "Alpha Session"},
        {ID: "s2", Name: "Beta Session"},
        {ID: "s3", Name: "Gamma Session"},
    }
    
    sl.SetSessions(sessions, nil)
    
    // No filter
    filtered := sl.filterSessions(sessions)
    if len(filtered) != 3 {
        t.Errorf("Expected 3 sessions without filter, got %d", len(filtered))
    }
    
    // With filter
    sl.SetSearchQuery("beta")
    filtered = sl.filterSessions(sessions)
    if len(filtered) != 1 {
        t.Errorf("Expected 1 session matching 'beta', got %d", len(filtered))
    }
    if len(filtered) > 0 && filtered[0].ID != "s2" {
        t.Errorf("Expected session ID 's2', got '%s'", filtered[0].ID)
    }
}

func TestSessionList_Selected(t *testing.T) {
    callbacks := SessionListCallbacks{}
    sl := NewSessionList(callbacks)
    
    // No selection
    _, ok := sl.Selected()
    if ok {
        t.Error("Selected should return false when no selection")
    }
    
    // Set sessions and select first
    sessions := []session.Snapshot{
        {ID: "s1", Name: "Test Session", StartedAt: time.Now()},
    }
    sl.SetSessions(sessions, nil)
    sl.list.Select(0)
    
    sess, ok := sl.Selected()
    if !ok {
        t.Error("Selected should return true when item selected")
    }
    if sess.ID != "s1" {
        t.Errorf("Expected session ID 's1', got '%s'", sess.ID)
    }
}

func TestSessionListItem_Title(t *testing.T) {
    tests := []struct {
        name     string
        session  session.Snapshot
        expected string
    }{
        {
            name:     "running session",
            session:  session.Snapshot{ID: "s1", Name: "Test", Status: session.StatusRunning},
            expected: "Test",
        },
        {
            name: "waiting session",
            session: session.Snapshot{
                ID:     "s2",
                Name:   "Waiting",
                Status: session.StatusWaiting,
            },
            // Expected will contain pink color codes
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            item := SessionListItem{Session: tt.session}
            title := item.Title()
            
            if tt.session.Status != session.StatusWaiting {
                if title != tt.expected {
                    t.Errorf("Expected title %q, got %q", tt.expected, title)
                }
            } else {
                // Waiting status should contain color codes
                if title == tt.expected {
                    t.Error("Waiting session title should include color styling")
                }
            }
        })
    }
}

func TestSessionListItem_Description(t *testing.T) {
    session := session.Snapshot{
        ID:           "s1",
        Name:         "Test Session",
        CardID:       "c1",
        Status:       session.StatusRunning,
        PID:          12345,
        WorkingDir:   "/home/user/project",
        UseTmpWorkspace: false,
    }
    
    card := &config.IDCard{
        ID:        "c1",
        Name:      "Test Card",
        FaceStyle: "happy",
    }
    
    item := SessionListItem{
        Session: session,
        Card:    card,
    }
    
    desc := item.Description()
    
    if desc == "" {
        t.Error("Description should not be empty")
    }
    
    // Should contain card name
    // Should contain "pid="
    // Should contain "WS:" (not using tmp)
}

func TestSessionList_SetSelectedByID(t *testing.T) {
    callbacks := SessionListCallbacks{}
    sl := NewSessionList(callbacks)
    
    sessions := []session.Snapshot{
        {ID: "s1", Name: "First"},
        {ID: "s2", Name: "Second"},
        {ID: "s3", Name: "Third"},
    }
    
    sl.SetSessions(sessions, nil)
    
    // Select second session
    sl.SetSelectedByID("s2")
    
    sess, ok := sl.Selected()
    if !ok {
        t.Fatal("Should have selection after SetSelectedByID")
    }
    if sess.ID != "s2" {
        t.Errorf("Expected session ID 's2', got '%s'", sess.ID)
    }
    
    // Try to select non-existent ID
    sl.SetSelectedByID("nonexistent")
    // Should still be on s2
    sess, _ = sl.Selected()
    if sess.ID != "s2" {
        t.Error("Should not change selection for non-existent ID")
    }
}

func TestRenderAgentFace(t *testing.T) {
    // Test that renderer works
    face := RenderAgentFace("neutral")
    if face == "" {
        t.Error("RenderAgentFace should return non-empty string")
    }
}

func TestRenderAgentPuppet(t *testing.T) {
    // Test that renderer works
    puppet := RenderAgentPuppet("kankuro")
    if puppet == "" {
        t.Error("RenderAgentPuppet should return non-empty string")
    }
}
```

---

## Step 2.2.4: Integration with App (Backwards Compatible)

**File**: `internal/ui/app_lists.go` (MODIFY - add integration stub)

```go
// ===== Integration with new SessionList component =====
// This shows how App will eventually delegate to SessionList.

func (a *App) newSessionListComponent() *components.SessionList {
    callbacks := components.SessionListCallbacks{
        OnSelect: func(sess session.Snapshot) {
            // Triggers preview update
            a.sessionView.PreviewSessionID = sess.ID
            a.refreshPreviewScrollback()
        },
        OnAttach: func(sessionID string) {
            // Same as handleDashboardEnter for sessions
            a.sessionView.ActiveID = sessionID
            a.mode = viewDashboard
            // Would return tea.Cmd: a.attachSession(sessionID)
        },
        OnKill: func(sessionID string) {
            // Kill the session
            if err := a.session.Kill(sessionID); err != nil {
                a.err = err
            }
            a.refreshSessionList()
        },
        OnDelete: func(sessionID string) {
            // Trigger delete confirmation
            // Need to find session by ID first
            for _, s := range a.sessions {
                if s.ID == sessionID {
                    if s.UseTmpWorkspace {
                        a.sessionView.PendingDeleteID = sessionID
                        a.mode = viewDeleteConfirm
                    } else {
                        if err := a.session.Remove(sessionID); err != nil {
                            a.err = err
                        }
                        a.refreshSessionList()
                    }
                    break
                }
            }
        },
        OnSwitch: func(sessionID string) {
            // Trigger agent switch
            a.sessionView.SwitchID = sessionID
            a.switchList.SetItems(toCardItems(a.cards, a.skills, a.personas))
            a.switchList.Select(0)
            a.mode = viewSwitchAgent
        },
        OnPreview: func(sessionID string) {
            // Sync preview
            a.sessionView.PreviewSessionID = sessionID
            a.refreshPreviewScrollback()
        },
    }
    
    return components.NewSessionList(callbacks)
}

// Future: once fully migrated, App will have:
// sessionListComponent *components.SessionList
// And New() will initialize it
```

---

## Step 2.2.5: Add Integration Test

**File**: `internal/ui/components/session_list_integration_test.go` (NEW)

```go
package components_test

import (
    "testing"
    "time"
    
    "kankuro/internal/config"
    "kankuro/internal/session"
    "kankuro/internal/ui/components"
)

func TestSessionListIntegration(t *testing.T) {
    // Simulate App workflow
    
    // 1. Create component with callbacks
    var selectedID string
    var previews []string
    
    sl := components.NewSessionList(components.SessionListCallbacks{
        OnSelect: func(s session.Snapshot) {
            selectedID = s.ID
        },
        OnPreview: func(sessionID string) {
            previews = append(previews, sessionID)
        },
    })
    
    // 2. Set data
    sessions := []session.Snapshot{
        {ID: "s1", Name: "Alpha", Status: session.StatusRunning, StartedAt: time.Now()},
        {ID: "s2", Name: "Beta", Status: session.StatusWaiting, StartedAt: time.Now().Add(-time.Hour)},
    }
    cards := []config.IDCard{
        {ID: "c1", Name: "Card1", FaceStyle: "happy"},
    }
    
    sl.SetSessions(sessions, cards)
    
    // 3. Initialize
    sl.Init()
    sl.SetSize(30, 10)
    
    // 4. Test selection
    sel, ok := sl.Selected()
    if !ok {
        t.Log("No initial selection (list empty or no items)")
    }
    
    // 5. Test SetSelectedByID
    sl.SetSelectedByID("s2")
    sel, ok = sl.Selected()
    if !ok {
        t.Error("Should have selection after SetSelectedByID")
    }
    if sel.ID != "s2" {
        t.Errorf("Expected session ID 's2', got '%s'", sel.ID)
    }
    
    // 6. Test search filter
    sl.SetSearchQuery("alpha")
    // In real use, this would be tested via Update with KeyMsg
    // For now, verify filter works
    filtered := sl.FilterSessions(sessions)
    if len(filtered) != 1 {
        t.Errorf("Expected 1 session after filter, got %d", len(filtered))
    }
    
    t.Logf("SessionList integration test completed successfully")
    t.Logf("Previews called: %d times", len(previews))
}
```

---

## Step 2.2.6: Export FilterSessions for Testing

Add to `session_list.go`:

```go
// FilterSessions exposes the filter logic for testing.
func (s *SessionList) FilterSessions(sessions []session.Snapshot) []session.Snapshot {
    return s.filterSessions(sessions)
}
```

---

## Verification Checklist

```bash
# 1. Build all packages
go build ./internal/ui/component/
go build ./internal/ui/messages/
go build ./internal/ui/commands/
go build ./internal/ui/components/

# 2. Run all tests
go test ./internal/ui/components/ -v

# 3. Existing UI still builds
go build ./internal/ui/

# 4. Run existing tests
go test ./internal/ui/
```

---

## Success Criteria

- [ ] `SessionList` component defined
- [ ] `SessionListItem` implements `list.Item`
- [ ] Theme colors defined in component (will refactor to shared later)
- [ ] `NewSessionList` with callbacks
- [ ] `SetSessions(sessions, cards)` pushes data
- [ ] `SetSearchQuery` filters
- [ ] `Selected()` returns session
- [ ] `SetSelectedByID` changes selection
- [ ] Preview sync on selection change (OnPreview callback)
- [ ] Waiting status special styling (hot pink)
- [ ] Working directory vs temp workspace display
- [ ] Shared `RenderAgentFace` and `RenderAgentPuppet`
- [ ] Unit tests pass (>80% coverage)
- [ ] Integration test demonstrates workflow
- [ ] Existing App still compiles

---

## Dependencies

```
component/component.go (from 1.1) ✓
messages/messages.go (from 1.2) ✓
commands/commands.go (from 1.3) ✓
components/card_list.go (from 2.1) ✓

components/session_list.go ← NEW
components/renderer.go ← NEW (shared face/puppet rendering)
```

---

## Estimated Time

| Step | Time |
|------|------|
| 2.2.1 SessionList file | 45 min |
| 2.2.2 Shared renderer | 15 min |
| 2.2.3 Unit tests | 30 min |
| 2.2.4 App integration | 15 min |
| 2.2.5 Integration test | 15 min |
| Verification | 10 min |
| **Total** | **~2 hours** |

---

## What This Enables

1. **Pattern consistency**: SessionList follows same extraction pattern as CardList
2. **Preview sync**: Dashboard preview updates work via OnPreview callback
3. **Status handling**: Waiting status special styling is encapsulated
4. **Next extractions**: SessionPreview (Task 2.3) and SourceList (Task 2.4) can follow same pattern
5. **Phase 3 readiness**: Dashboard component can compose CardList + SessionList + SessionPreview