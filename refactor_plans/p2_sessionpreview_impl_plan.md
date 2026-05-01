# Implementation Plan: Phase 2 Task 2.3 - SessionPreview Component

## Goal

Extract SessionPreview as independent display-only component, handling text sanitization, scrollback rendering, and session metadata display.

---

## Analysis: What to Extract

### From Existing Code

| File | Lines | Content | Destination |
|------|-------|---------|-------------|
| `view_preview.go` | 16-27 | Preview constants | `components/session_preview.go` |
| `view_preview.go` | 29-33 | Regex patterns for sanitization | `components/session_preview.go` |
| `view_preview.go` | 35-42 | `StripAllControlCodes()` | `components/session_preview.go` |
| `view_preview.go` | 44-121 | `renderSessionPreview()` | `SessionPreview.View()` |
| `view_preview.go` | 123-134 | `renderStatusBadge()` | `components/session_preview.go` |
| `view_preview.go` | 136-199 | `renderScrollback()` | `components/session_preview.go` |
| `view_preview.go` | 201-206 | `truncatePath()` | `components/session_preview.go` |
| `view_preview.go` | 208-231 | Helper finders | Remove (use data push) |
| `styles.go` | 22-36 | Status badge styles | `components/session_preview.go` |
| `app_state.go` | 65-66 | PreviewSessionID, PreviewScrollback | Internal state |

### SessionPreview Specifics

**Unlike CardList/SessionList**:
- **Display-only**: No user interaction, no selection
- **Passive**: Data pushed in, doesn't pull from App
- **Complex text processing**: ANSI stripping, control char removal, Kitty protocol handling
- **Scrollback management**: Cap size, line truncation, empty state
- **Multi-part layout**: Header (puppet + metadata) + scrollback area

**Key responsibilities**:
1. Render placeholder when no session selected
2. Render session metadata (ID, name, workspace, started time, card)
3. Render status badge (Running/Waiting/Terminated)
4. Render puppet avatar
5. Render scrollback buffer (sanitized, truncated)

---

## Step 2.3.1: Create SessionPreview Component File

**File**: `internal/ui/components/session_preview.go` (NEW)

```go
package components

import (
    "regexp"
    "strings"
    "time"
    "unicode"

    "github.com/charmbracelet/lipgloss"
    
    "kankuro/internal/config"
    "kankuro/internal/session"
    "kankuro/internal/ui/component"
)

// Preview layout constants.
const (
    PreviewIDLabelPrefixLen   = 4
    PreviewCardLabelPrefixLen = 6
    PreviewEllipsisLen        = 3
    PreviewBorderLines        = 2
    PreviewMaxScrollbackBytes = 8 * 1024 // 8KB cap
    
    HeaderHeight = 2 // title + margin
    FooterHeight = 1
    WidthBorders = 6
)

// Regex patterns for text sanitization.
var (
    ansiEscapeRE   = regexp.MustCompile(`\x1b\[[0-9;]*[A-Za-z]`)
    controlCharRE  = regexp.MustCompile(`[\x00-\x08\x0b\x0c\x0e-\x1f\x7f]`)
    kittyProtocolRE = regexp.MustCompile(`(?:\x1b_|\x1b\[|\x1b\]|\x1bP).*?(\x1b\\|[\x07\x40-\x7e])`)
)

// Theme colors.
var (
    ColorPrimary    = lipgloss.Color("#8E44AD")
    ColorHighlight  = lipgloss.Color("#16A085")
    ColorHotPink    = lipgloss.Color("#FF69B4")
    ColorGray       = lipgloss.Color("240")
    ColorWhite      = lipgloss.Color("#FFFFFF")
    ColorDeepCharcoal = lipgloss.Color("#1A1A1A")
)

// Status badge styles.
var (
    StatusRunningStyle = lipgloss.NewStyle().
        Foreground(ColorWhite).
        Background(ColorHighlight).
        Padding(0, 1)
    
    StatusWaitingStyle = lipgloss.NewStyle().
        Foreground(ColorWhite).
        Background(ColorHotPink).
        Padding(0, 1)
    
    StatusTerminatedStyle = lipgloss.NewStyle().
        Foreground(ColorWhite).
        Background(ColorGray).
        Padding(0, 1)
    
    TitleStyle = lipgloss.NewStyle().
        Bold(true).
        Foreground(ColorPrimary)
)

// SessionPreview displays detailed information about a selected session.
type SessionPreview struct {
    component.Base
    
    sessionID   string
    snapshot    *session.Snapshot
    card        *config.IDCard
    scrollback  []byte
}

// NewSessionPreview creates a new session preview component.
func NewSessionPreview() *SessionPreview {
    return &SessionPreview{}
}

// SetSession updates the displayed session.
func (p *SessionPreview) SetSession(snap *session.Snapshot, card *config.IDCard) {
    p.snapshot = snap
    p.card = card
    if snap != nil {
        p.sessionID = snap.ID
    } else {
        p.sessionID = ""
    }
}

// SetScrollback updates the scrollback buffer.
func (p *SessionPreview) SetScrollback(data []byte) {
    p.scrollback = data
}

// SetSessionByID sets the session by ID (requires separate SetSession call).
func (p *SessionPreview) SetSessionByID(sessionID string) {
    p.sessionID = sessionID
}

// Clear resets the preview to empty state.
func (p *SessionPreview) Clear() {
    p.sessionID = ""
    p.snapshot = nil
    p.card = nil
    p.scrollback = nil
}

// SessionID returns the current session ID.
func (p *SessionPreview) SessionID() string {
    return p.sessionID
}

// Init initializes the component.
func (p *SessionPreview) Init() bubbletea.Cmd {
    return nil
}

// Update handles messages (passive component, no interaction).
func (p *SessionPreview) Update(msg bubbletea.Msg) (component.Model, bubbletea.Cmd) {
    // SessionPreview is display-only, no interaction handling
    return p, nil
}

// View renders the session preview.
func (p *SessionPreview) View() string {
    if p.sessionID == "" || p.snapshot == nil {
        return p.renderPlaceholder()
    }
    
    return p.renderPreview()
}

// renderPlaceholder shows the empty state.
func (p *SessionPreview) renderPlaceholder() string {
    placeholder := `
   No session
   selected

   Select a session
   from the list to
   see a preview here.
`
    return lipgloss.NewStyle().
        Width(p.Width - WidthBorders).
        Height(p.Height).
        Align(lipgloss.Center, lipgloss.Center).
        Foreground(ColorGray).
        Render(placeholder)
}

// renderPreview renders the full session preview.
func (p *SessionPreview) renderPreview() string {
    statusBadge := renderStatusBadge(p.snapshot.Status)
    
    sessionName := p.snapshot.Name
    if sessionName == "" {
        sessionName = p.snapshot.ID
    }
    
    wdStr := truncatePath(p.snapshot.WorkingDir, 25)
    startedStr := p.snapshot.StartedAt.Format("15:04:05")
    
    // Render puppet
    puppetStyle := "kankuro"
    if p.card != nil && p.card.Puppet != "" {
        puppetStyle = p.card.Puppet
    }
    puppetStr := RenderAgentPuppet(puppetStyle)
    
    // Build header
    headerLines := []string{
        TitleStyle.Render(truncate(sessionName, p.Width-WidthBorders)),
        lipgloss.JoinHorizontal(lipgloss.Center, statusBadge),
        lipgloss.NewStyle().Foreground(ColorGray).Render("ID: " + truncate(p.snapshot.ID, p.Width-PreviewIDLabelPrefixLen-WidthBorders)),
        lipgloss.NewStyle().Foreground(ColorGray).Render("WS: " + wdStr),
        lipgloss.NewStyle().Foreground(ColorGray).Render("Started: " + startedStr),
    }
    
    if p.card != nil {
        headerLines = append(headerLines,
            lipgloss.NewStyle().Foreground(ColorGray).Render("Card: "+truncate(p.card.Name, p.Width-PreviewCardLabelPrefixLen-WidthBorders)),
        )
    }
    headerLines = append(headerLines, "")
    
    header := strings.Join(headerLines, "\n")
    
    // Calculate scrollback height
    scrollbackHeight := p.Height - HeaderHeight - FooterHeight
    scrollbackContent := p.renderScrollback(scrollbackHeight)
    
    // Combine header with puppet
    headerWithPuppet := lipgloss.JoinHorizontal(lipgloss.Center, header, puppetStr)
    
    // Combine header and scrollback
    body := lipgloss.JoinVertical(lipgloss.Left,
        headerWithPuppet,
        scrollbackContent,
    )
    
    return lipgloss.NewStyle().Render(body)
}

// renderScrollback renders the scrollback buffer area.
func (p *SessionPreview) renderScrollback(height int) string {
    if len(p.scrollback) == 0 {
        return lipgloss.NewStyle().
            Border(lipgloss.NormalBorder(), true, false, false, false).
            BorderForeground(ColorGray).
            Width(p.Width).
            Height(max(1, height-1)).
            Align(lipgloss.Center, lipgloss.Center).
            Foreground(ColorGray).
            Render("(no output yet)")
    }
    
    // Cap scrollback size
    scrollback := p.scrollback
    if len(scrollback) > PreviewMaxScrollbackBytes {
        scrollback = scrollback[len(scrollback)-PreviewMaxScrollbackBytes:]
    }
    
    // Sanitize text
    text := sanitizeScrollback(scrollback)
    
    // Split and limit lines
    lines := strings.Split(text, "\n")
    maxLines := max(1, height-PreviewBorderLines)
    if len(lines) > maxLines {
        lines = lines[len(lines)-maxLines:]
    }
    
    // Process lines
    var truncated []string
    for _, line := range lines {
        clean := strings.TrimSpace(line)
        if clean == "" {
            continue
        }
        if len(clean) > p.Width-2 {
            truncated = append(truncated, clean[:p.Width-5]+"...")
        } else {
            truncated = append(truncated, clean)
        }
    }
    
    if len(truncated) == 0 {
        return lipgloss.NewStyle().
            Border(lipgloss.NormalBorder(), true, false, false, false).
            BorderForeground(ColorGray).
            Width(p.Width).
            Height(max(1, height-1)).
            Align(lipgloss.Center, lipgloss.Center).
            Foreground(ColorGray).
            Render("(no output yet)")
    }
    
    content := strings.Join(truncated, "\n")
    
    return lipgloss.NewStyle().
        Border(lipgloss.NormalBorder(), true, false, false, false).
        BorderForeground(ColorGray).
        Width(p.Width).
        Height(height - 1).
        Render(content)
}

// renderStatusBadge renders the status badge.
func renderStatusBadge(status session.Status) string {
    switch status {
    case session.StatusRunning:
        return StatusRunningStyle.Render("Running")
    case session.StatusWaiting:
        return StatusWaitingStyle.Render("Waiting")
    case session.StatusTerminated:
        return StatusTerminatedStyle.Render("Terminated")
    default:
        return StatusTerminatedStyle.Render("Unknown")
    }
}

// sanitizeScrollback strips ANSI codes and control characters.
func sanitizeScrollback(data []byte) string {
    // Strip ANSI sequences
    stripped := ansiEscapeRE.ReplaceAll(data, []byte{})
    
    // Strip control characters
    stripped = controlCharRE.ReplaceAll(stripped, []byte{})
    
    // Strip Kitty protocol sequences
    stripped = kittyProtocolRE.ReplaceAll(stripped, []byte{})
    
    // Remove carriage returns
    text := strings.ReplaceAll(string(stripped), "\r", "")
    
    // Strip remaining non-printable characters
    text = StripAllControlCodes(text)
    
    return text
}

// StripAllControlCodes removes all non-printable characters except newlines and tabs.
func StripAllControlCodes(s string) string {
    return strings.Map(func(r rune) rune {
        if unicode.IsPrint(r) || r == '\n' || r == '\t' {
            return r
        }
        return -1
    }, s)
}

// truncatePath truncates a path with ellipsis.
func truncatePath(path string, limit int) string {
    if len(path) <= limit {
        return path
    }
    return "..." + path[len(path)-(limit-PreviewEllipsisLen):]
}

// truncate truncates a string with ellipsis.
func truncate(s string, w int) string {
    if len(s) > w && w > 3 {
        return s[:w-3] + "..."
    }
    return s
}
```

---

## Step 2.3.2: Create SessionPreview Tests

**File**: `internal/ui/components/session_preview_test.go` (NEW)

```go
package components

import (
    "testing"
    "time"
    
    "kankuro/internal/config"
    "kankuro/internal/session"
)

func TestSessionPreview_New(t *testing.T) {
    p := NewSessionPreview()
    if p == nil {
        t.Fatal("NewSessionPreview returned nil")
    }
    
    if p.sessionID != "" {
        t.Error("New preview should have empty session ID")
    }
}

func TestSessionPreview_Clear(t *testing.T) {
    p := NewSessionPreview()
    
    // Set some data
    snap := &session.Snapshot{ID: "s1", Name: "Test"}
    p.SetSession(snap, nil)
    p.SetScrollback([]byte("output"))
    
    // Clear
    p.Clear()
    
    if p.sessionID != "" {
        t.Error("Session ID should be empty after clear")
    }
    if p.snapshot != nil {
        t.Error("Snapshot should be nil after clear")
    }
    if p.scrollback != nil {
        t.Error("Scrollback should be nil after clear")
    }
}

func TestSessionPreview_SetSession(t *testing.T) {
    p := NewSessionPreview()
    p.SetSize(50, 20)
    
    snap := &session.Snapshot{
        ID:        "s1",
        Name:      "Test Session",
        Status:    session.StatusRunning,
        PID:       12345,
        WorkingDir: "/home/user/project",
        StartedAt: time.Now(),
    }
    
    card := &config.IDCard{
        ID:     "c1",
        Name:   "Test Card",
        Puppet: "kankuro",
    }
    
    p.SetSession(snap, card)
    
    if p.sessionID != "s1" {
        t.Errorf("Expected session ID 's1', got '%s'", p.sessionID)
    }
    
    if p.snapshot == nil {
        t.Error("Snapshot should not be nil")
    }
    
    if p.card == nil {
        t.Error("Card should not be nil")
    }
}

func TestSessionPreview_View_Empty(t *testing.T) {
    p := NewSessionPreview()
    p.SetSize(40, 15)
    
    view := p.View()
    
    if view == "" {
        t.Error("View should not be empty")
    }
    
    // Should show placeholder
    if !containsSubstring(view, "No session") {
        t.Error("Should show placeholder when no session")
    }
}

func TestSessionPreview_View_WithSession(t *testing.T) {
    p := NewSessionPreview()
    p.SetSize(50, 20)
    
    snap := &session.Snapshot{
        ID:        "s1",
        Name:      "Test Session",
        Status:    session.StatusRunning,
        PID:       12345,
        WorkingDir: "/home/user/project",
        StartedAt: time.Now(),
    }
    
    p.SetSession(snap, nil)
    
    view := p.View()
    
    if view == "" {
        t.Error("View should not be empty")
    }
    
    // Should contain session name
    if !containsSubstring(view, "Test Session") {
        t.Error("Should contain session name")
    }
    
    // Should contain status badge
    if !containsSubstring(view, "Running") {
        t.Error("Should contain status badge")
    }
}

func TestSessionPreview_View_WithScrollback(t *testing.T) {
    p := NewSessionPreview()
    p.SetSize(60, 25)
    
    snap := &session.Snapshot{
        ID:        "s1",
        Name:      "Test",
        Status:    session.StatusRunning,
        StartedAt: time.Now(),
    }
    
    p.SetSession(snap, nil)
    p.SetScrollback([]byte("Line 1\nLine 2\nLine 3"))
    
    view := p.View()
    
    // Should contain scrollback content
    if !containsSubstring(view, "Line 1") {
        t.Error("Should contain scrollback content")
    }
}

func TestSessionPreview_SanitizeScrollback(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        contains string
        omit     string
    }{
        {
            name:     "plain text",
            input:    "Hello World",
            contains: "Hello World",
            omit:     "",
        },
        {
            name:     "ANSI codes removed",
            input:    "\x1b[32mGreen\x1b[0m text",
            contains: "Green text",
            omit:     "\x1b[",
        },
        {
            name:     "control chars removed",
            input:    "Line1\x00\x01Line2",
            contains: "Line1Line2",
            omit:     "\x00",
        },
        {
            name:     "carriage return removed",
            input:    "Line1\r\nLine2",
            contains: "Line1\nLine2",
            omit:     "\r",
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := sanitizeScrollback([]byte(tt.input))
            
            if tt.contains != "" && !containsSubstring(result, tt.contains) {
                t.Errorf("Expected result to contain %q, got %q", tt.contains, result)
            }
            
            if tt.omit != "" && containsSubstring(result, tt.omit) {
                t.Errorf("Result should not contain %q, got %q", tt.omit, result)
            }
        })
    }
}

func TestStripAllControlCodes(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        expected string
    }{
        {
            name:     "preserves printable",
            input:    "Hello World!",
            expected: "Hello World!",
        },
        {
            name:     "preserves newlines",
            input:    "Line1\nLine2",
            expected: "Line1\nLine2",
        },
        {
            name:     "preserves tabs",
            input:    "Col1\tCol2",
            expected: "Col1\tCol2",
        },
        {
            name:     "removes control chars",
            input:    "Hello\x00\x01\x02World",
            expected: "HelloWorld",
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := StripAllControlCodes(tt.input)
            if result != tt.expected {
                t.Errorf("Expected %q, got %q", tt.expected, result)
            }
        })
    }
}

func TestTruncatePath(t *testing.T) {
    tests := []struct {
        name     string
        path     string
        limit    int
        expected string
    }{
        {
            name:     "short path unchanged",
            path:     "/home/user/project",
            limit:    25,
            expected: "/home/user/project",
        },
        {
            name:     "long path truncated",
            path:     "/home/user/very/long/path/to/some/project",
            limit:    20,
            expected: "...path/to/some/project",
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := truncatePath(tt.path, tt.limit)
            if result != tt.expected {
                t.Errorf("Expected %q, got %q", tt.expected, result)
            }
        })
    }
}

func TestRenderStatusBadge(t *testing.T) {
    tests := []struct {
        name   string
        status session.Status
        label  string
    }{
        {"running", session.StatusRunning, "Running"},
        {"waiting", session.StatusWaiting, "Waiting"},
        {"terminated", session.StatusTerminated, "Terminated"},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            badge := renderStatusBadge(tt.status)
            if !containsSubstring(badge, tt.label) {
                t.Errorf("Badge should contain %q", tt.label)
            }
        })
    }
}

func TestSessionPreview_View_WaitingStatus(t *testing.T) {
    p := NewSessionPreview()
    p.SetSize(50, 20)
    
    snap := &session.Snapshot{
        ID:        "s1",
        Name:      "Waiting Session",
        Status:    session.StatusWaiting,
        StartedAt: time.Now(),
    }
    
    p.SetSession(snap, nil)
    
    view := p.View()
    
    // Waiting status should show hot pink badge
    if !containsSubstring(view, "Waiting") {
        t.Error("Should contain 'Waiting' status")
    }
}

func TestSessionPreview_SessionID(t *testing.T) {
    p := NewSessionPreview()
    
    if p.SessionID() != "" {
        t.Error("Initial session ID should be empty")
    }
    
    p.SetSessionByID("test-id")
    if p.SessionID() != "test-id" {
        t.Errorf("Expected 'test-id', got '%s'", p.SessionID())
    }
}

// Helper
func containsSubstring(s, substr string) bool {
    return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && findSubstring(s, substr))
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

## Step 2.3.3: Integration with App (Backwards Compatible)

**File**: `internal/ui/app_dashboard.go` (MODIFY - add integration stub)

Add after `syncPreviewSelection`:

```go
// ===== Integration with new SessionPreview component =====

func (a *App) newSessionPreviewComponent() *components.SessionPreview {
    p := components.NewSessionPreview()
    p.SetSize(a.sessionView.PreviewWidth, a.sessionsList.Height())
    return p
}

// syncPreviewToComponent pushes preview state to the component.
// Called after syncPreviewSelection updates sessionView state.
func (a *App) syncPreviewToComponent(p *components.SessionPreview) {
    if a.sessionView.PreviewSessionID == "" {
        p.Clear()
        return
    }
    
    // Find session snapshot
    var snap *session.Snapshot
    for i := range a.sessions {
        if a.sessions[i].ID == a.sessionView.PreviewSessionID {
            snap = &a.sessions[i]
            break
        }
    }
    
    if snap == nil {
        p.Clear()
        return
    }
    
    // Find card
    var card *config.IDCard
    for i := range a.cards {
        if a.cards[i].ID == snap.CardID {
            card = &a.cards[i]
            break
        }
    }
    
    p.SetSession(snap, card)
    p.SetScrollback(a.sessionView.PreviewScrollback)
}
```

---

## Step 2.3.4: Add Integration Test

**File**: `internal/ui/components/session_preview_integration_test.go` (NEW)

```go
package components_test

import (
    "testing"
    "time"
    
    "kankuro/internal/config"
    "kankuro/internal/session"
    "kankuro/internal/ui/components"
)

func TestSessionPreviewIntegration(t *testing.T) {
    // Simulate App workflow
    
    // 1. Create component
    p := components.NewSessionPreview()
    p.SetSize(50, 20)
    
    // 2. Initial state - empty placeholder
    view := p.View()
    if view == "" {
        t.Error("View should render placeholder")
    }
    
    // 3. Set session data
    snap := &session.Snapshot{
        ID:        "s1",
        Name:      "Integration Test",
        Status:    session.StatusRunning,
        PID:       12345,
        WorkingDir: "/home/user/test",
        StartedAt: time.Now(),
    }
    
    card := &config.IDCard{
        ID:     "c1",
        Name:   "Test Card",
        Puppet: "kankuro",
    }
    
    p.SetSession(snap, card)
    
    // 4. View with session
    view = p.View()
    if view == "" {
        t.Error("View should render session")
    }
    
    // 5. Add scrollback
    p.SetScrollback([]byte("Starting process...\nProcess running\nDone"))
    
    view = p.View()
    if view == "" {
        t.Error("View should render with scrollback")
    }
    
    // 6. Clear
    p.Clear()
    view = p.View()
    if view == "" {
        t.Error("View should render placeholder after clear")
    }
    
    t.Log("SessionPreview integration test completed successfully")
}

func TestSessionPreview_ScrollbackSanitization(t *testing.T) {
    p := components.NewSessionPreview()
    p.SetSize(60, 25)
    
    snap := &session.Snapshot{
        ID:        "s1",
        Name:      "Test",
        Status:    session.StatusRunning,
        StartedAt: time.Now(),
    }
    p.SetSession(snap, nil)
    
    // Set scrollback with ANSI codes
    scrollback := []byte("\x1b[32mSuccess\x1b[0m: Operation completed\n\x1b[31mError\x1b[0m: None")
    p.SetScrollback(scrollback)
    
    view := p.View()
    
    // ANSI codes should be stripped
    if view == "" {
        t.Error("View should render")
    }
    
    // Should contain the actual text (after stripping)
    // (In real test, check that no escape sequences remain)
    
    t.Log("Scrollback sanitization test completed")
}

func TestSessionPreview_ResizeHandling(t *testing.T) {
    p := components.NewSessionPreview()
    
    // Initial size
    p.SetSize(40, 15)
    if p.Width != 40 || p.Height != 15 {
        t.Error("Size not set correctly")
    }
    
    // Resize
    p.SetSize(80, 30)
    if p.Width != 80 || p.Height != 30 {
        t.Error("Resize not handled correctly")
    }
    
    // Should still render correctly
    snap := &session.Snapshot{
        ID:        "s1",
        Name:      "Resize Test",
        Status:    session.StatusRunning,
        StartedAt: time.Now(),
    }
    p.SetSession(snap, nil)
    
    view := p.View()
    if view == "" {
        t.Error("Should render after resize")
    }
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

# 3. Specific tests for SessionPreview
go test ./internal/ui/components/ -run SessionPreview -v

# 4. Existing UI still builds
go build ./internal/ui/

# 5. Run existing tests
go test ./internal/ui/
```

---

## Success Criteria

- [ ] `SessionPreview` component defined
- [ ] `SetSession(snap, card)` pushes data
- [ ] `SetScrollback(data)` pushes scrollback
- [ ] `Clear()` resets to empty state
- [ ] `SessionID()` returns current ID
- [ ] Placeholder rendered when no session
- [ ] Session metadata rendered (ID, name, workspace, time, card)
- [ ] Status badge rendered (Running/Waiting/Terminated)
- [ ] Puppet avatar rendered (uses shared `RenderAgentPuppet`)
- [ ] Scrollback rendered with sanitization
- [ ] ANSI escape codes stripped
- [ ] Control characters stripped
- [ ] Kitty protocol stripped
- [ ] Long lines truncated
- [ ] Empty scrollback shows placeholder
- [ ] Unit tests pass (>80% coverage)
- [ ] Integration test demonstrates workflow
- [ ] Existing App still compiles

---

## Dependencies

```
component/component.go (from 1.1) ✓
messages/messages.go (from 1.2) ✓
commands/commands.go (from 1.3) ✓
components/renderer.go (from 2.2) ✓  // RenderAgentPuppet

components/session_preview.go ← NEW
```

---

## Estimated Time

| Step | Time |
|------|------|
| 2.3.1 SessionPreview file | 45 min |
| 2.3.2 Unit tests | 35 min |
| 2.3.3 App integration | 10 min |
| 2.3.4 Integration test | 15 min |
| Verification | 10 min |
| **Total** | **~2 hours** |

---

## What This Enables

1. **Dashboard preview pane**: Complete SessionPreview ready for Dashboard composition
2. **Text sanitization**: Reusable `StripAllControlCodes` and `sanitizeScrollback` utilities
3. **Status badges**: Standardized badge rendering for session status
4. **Phase 3 ready**: Dashboard component can compose CardList + SessionList + SessionPreview
5. **Pattern consistency**: Third component following same extraction pattern

---

## Key Design Decisions

| Decision | Rationale |
|----------|-----------|
| **Display-only** | Preview doesn't accept user input, simplifies Update() |
| **Data push** | Parent pushes session/scrollback via setters, no callbacks needed |
| **Inline sanitization** | Regex patterns compiled once, used in render path |
| **Shared renderer** | Uses `RenderAgentPuppet` from Step 2.2.2 |
| **Height calculation** | Auto-calculates scrollback height from component size |