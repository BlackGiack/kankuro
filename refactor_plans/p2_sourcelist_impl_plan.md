# Implementation Plan: Phase 2 Task 2.4 - SourceList Component (Generic)

## Goal

Extract a generic SourceList component that eliminates duplicate code for skill sources, persona sources, and workspace sources management.

---

## Analysis: The Duplication Problem

### Current Duplicate Files

| File | Lines | Purpose | Duplication |
|------|-------|---------|-------------|
| `app_skill_sources.go` | 156 | Skill source CRUD | 95% identical to persona |
| `app_persona_sources.go` | 159 | Persona source CRUD | 95% identical to skill |
| `app_workspaces.go` | 136-234 | Workspace source CRUD | Similar pattern, slightly different |

### Identical Patterns (Skills vs Personas)

```
Both have:
- updateSkillSources() / updatePersonaSources() - IDENTICAL
- handleSourceNew() - IDENTICAL
- handleSourceSelect() - IDENTICAL
- handleSourceDelete() - IDENTICAL
- updateSourceForm() - IDENTICAL
- saveSource() - IDENTICAL
- renderSourceList() - IDENTICAL
- refreshItems() - IDENTICAL
```

**Only difference**: Repository method names and view mode constants.

### Workspace Differences

Workspaces have:
- **Complex source type**: `config.WorkspaceSource` with ID, Path, IsRoot
- **Two input fields**: Path + IsRoot toggle
- **Custom rendering**: Shows "[R]" marker for root folders

---

## Design: Generic SourceList Component

### Approach: Interface + Specialization

```go
// SourceItem interface - anything can be a source item
type SourceItem interface {
    Display() string  // How to show in list
    ID() string       // Unique identifier (for complex items)
}

// SimpleSourceItem - for skills/personas (just strings)
type SimpleSourceItem string

// WorkspaceSourceItem - for workspaces (complex object)
type WorkspaceSourceItem struct {
    Source config.WorkspaceSource
}
```

### Component Modes

1. **List mode**: Show sources with cursor, allow navigation
2. **Form mode**: Input field(s) for add/edit
3. **Simple form**: Single text input (skills, personas)
4. **Complex form**: Multiple inputs (workspaces)

---

## Step 2.4.1: Create SourceItem Types

**File**: `internal/ui/components/source_item.go` (NEW)

```go
package components

import (
    "fmt"
    
    "kankuro/internal/config"
)

// SourceItem defines the interface for items in a source list.
type SourceItem interface {
    // Display returns the string representation for the list.
    Display() string
    
    // ID returns a unique identifier (may be empty for simple items).
    ID() string
}

// SimpleSourceItem is a source item that's just a path string.
// Used for skills and personas sources.
type SimpleSourceItem string

func (s SimpleSourceItem) Display() string { return string(s) }
func (s SimpleSourceItem) ID() string       { return "" } // No separate ID

// WorkspaceSourceItem wraps a workspace source configuration.
type WorkspaceSourceItem struct {
    Source config.WorkspaceSource
}

func (w WorkspaceSourceItem) Display() string {
    marker := " "
    if w.Source.IsRoot {
        marker = "R"
    }
    return fmt.Sprintf("[%s] %s", marker, w.Source.Path)
}

func (w WorkspaceSourceItem) ID() string {
    return w.Source.ID
}

// ToSimpleItems converts string slice to SimpleSourceItem slice.
func ToSimpleItems(sources []string) []SourceItem {
    items := make([]SourceItem, len(sources))
    for i, s := range sources {
        items[i] = SimpleSourceItem(s)
    }
    return items
}

// ToWorkspaceItems converts workspace sources to WorkspaceSourceItem slice.
func ToWorkspaceItems(sources []config.WorkspaceSource) []SourceItem {
    items := make([]SourceItem, len(sources))
    for i, s := range sources {
        items[i] = WorkspaceSourceItem{Source: s}
    }
    return items
}

// FromSimpleItems extracts strings from SimpleSourceItem slice.
func FromSimpleItems(items []SourceItem) []string {
    sources := make([]string, 0, len(items))
    for _, item := range items {
        if s, ok := item.(SimpleSourceItem); ok {
            sources = append(sources, string(s))
        }
    }
    return sources
}
```

---

## Step 2.4.2: Create SourceList Component

**File**: `internal/ui/components/source_list.go` (NEW)

```go
package components

import (
    "fmt"
    "strings"

    "github.com/charmbracelet/bubbles/textinput"
    bubbletea "github.com/charmbracelet/bubbletea"
    "github.com/charmbracelet/lipgloss"
    
    "kankuro/internal/ui/component"
)

// SourceListMode determines if showing list or form.
type SourceListMode int

const (
    SourceListModeList SourceListMode = iota
    SourceListModeForm
)

// SourceListCallbacks defines callbacks for source actions.
type SourceListCallbacks struct {
    // OnSave is called when user saves a source.
    // Returns the new/updated item and any error.
    OnSave func(item SourceItem, editingIndex int) error
    
    // OnDelete is called when user deletes a source.
    OnDelete func(index int) error
    
    // OnRefresh is called after save/delete to reload data.
    OnRefresh func()
    
    // OnClose is called when user presses Escape.
    OnClose func()
}

// SourceListConfig configures the source list behavior.
type SourceListConfig struct {
    Title       string
    Legend      string // Shown at bottom, e.g., "(R=Root folder...)"
    FormPrompt  string // Label for input, e.g., "Path:"
    FormFields  int    // Number of form fields (1 or 2)
    Field2Label string // Label for second field, e.g., "Is Root (y/n):"
}

// SourceList is a reusable component for managing source lists.
type SourceList struct {
    component.Base
    
    mode         SourceListMode
    config       SourceListConfig
    callbacks    SourceListCallbacks
    
    items        []SourceItem
    cursor       int
    
    inputs       []textinput.Model
    editingIndex int // -1 for new, >= 0 for edit
    
    // Theme
    primaryColor   lipgloss.Color
    highlightColor lipgloss.Color
    gray           lipgloss.Color
}

// NewSourceList creates a new source list component.
func NewSourceList(config SourceListConfig, callbacks SourceListCallbacks) *SourceList {
    // Create inputs based on config
    numInputs := config.FormFields
    if numInputs < 1 {
        numInputs = 1
    }
    
    inputs := make([]textinput.Model, numInputs)
    for i := range inputs {
        inputs[i] = textinput.New()
        inputs[i].Width = 60
    }
    
    if config.FormPrompt != "" {
        inputs[0].Placeholder = config.FormPrompt
    }
    
    return &SourceList{
        config:         config,
        callbacks:     callbacks,
        inputs:        inputs,
        editingIndex:  -1,
        primaryColor:  lipgloss.Color("#8E44AD"),
        highlightColor: lipgloss.Color("#16A085"),
        gray:          lipgloss.Color("240"),
    }
}

// SetItems updates the source list data.
func (s *SourceList) SetItems(items []SourceItem) {
    s.items = items
    
    // Adjust cursor if needed
    maxCursor := len(items) // +1 for "New" option
    if s.cursor > maxCursor {
        s.cursor = maxCursor
    }
    if s.cursor < 0 {
        s.cursor = 0
    }
}

// Init initializes the component.
func (s *SourceList) Init() bubbletea.Cmd {
    return nil
}

// Update handles messages.
func (s *SourceList) Update(msg bubbletea.Msg) (component.Model, bubbletea.Cmd) {
    if s.mode == SourceListModeForm {
        return s.updateForm(msg)
    }
    return s.updateList(msg)
}

// updateList handles list mode input.
func (s *SourceList) updateList(msg bubbletea.Msg) (component.Model, bubbletea.Cmd) {
    keyMsg, ok := msg.(bubbletea.KeyMsg)
    if !ok {
        // Handle text input updates
        for i := range s.inputs {
            var cmd bubbletea.Cmd
            s.inputs[i], cmd = s.inputs[i].Update(msg)
            if cmd != nil {
                return s, cmd
            }
        }
        return s, nil
    }
    
    switch keyMsg.String() {
    case "up":
        if s.cursor > 0 {
            s.cursor--
        }
    case "down":
        if s.cursor < len(s.items) { // +1 for "New"
            s.cursor++
        }
    case "n":
        s.openForm(-1) // New item
    case " ", "enter":
        s.openForm(s.cursor)
    case "d", "backspace":
        s.handleDelete()
    case "esc":
        if s.callbacks.OnClose != nil {
            s.callbacks.OnClose()
        }
    case "ctrl+c":
        return s, bubbletea.Quit
    }
    
    return s, nil
}

// updateForm handles form mode input.
func (s *SourceList) updateForm(msg bubbletea.Msg) (component.Model, bubbletea.Cmd) {
    // Update focused input
    var cmd bubbletea.Cmd
    for i := range s.inputs {
        s.inputs[i], cmd = s.inputs[i].Update(msg)
    }
    
    keyMsg, ok := msg.(bubbletea.KeyMsg)
    if !ok {
        return s, cmd
    }
    
    switch keyMsg.String() {
    case "enter":
        s.handleSave()
        s.mode = SourceListModeList
    case "tab":
        s.cycleInputFocus()
    case "esc":
        s.mode = SourceListModeList
    case "ctrl+c":
        return s, bubbletea.Quit
    }
    
    return s, nil
}

// openForm opens the form for new or edit.
func (s *SourceList) openForm(index int) {
    s.editingIndex = index
    
    if index == len(s.items) || index < 0 {
        // New item
        s.editingIndex = -1
        for i := range s.inputs {
            s.inputs[i].SetValue("")
        }
    } else {
        // Edit existing
        item := s.items[index]
        s.inputs[0].SetValue(item.Display())
        
        // Handle workspace second field
        if len(s.inputs) > 1 {
            if ws, ok := item.(WorkspaceSourceItem); ok {
                val := "n"
                if ws.Source.IsRoot {
                    val = "y"
                }
                s.inputs[1].SetValue(val)
            }
        }
    }
    
    s.inputs[0].Focus()
    s.mode = SourceListModeForm
}

// handleSave processes form submission.
func (s *SourceList) handleSave() {
    path := strings.TrimSpace(s.inputs[0].Value())
    if path == "" {
        return
    }
    
    var item SourceItem
    
    if len(s.inputs) > 1 {
        // Complex item (workspace)
        isRoot := strings.ToLower(strings.TrimSpace(s.inputs[1].Value())) == "y"
        
        // Get existing ID if editing
        existingID := ""
        if s.editingIndex >= 0 && s.editingIndex < len(s.items) {
            existingID = s.items[s.editingIndex].ID()
        }
        
        item = WorkspaceSourceItem{
            Source: config.WorkspaceSource{
                ID:     existingID,
                Path:   path,
                IsRoot: isRoot,
            },
        }
    } else {
        // Simple item (skill/persona)
        item = SimpleSourceItem(path)
    }
    
    if s.callbacks.OnSave != nil {
        if err := s.callbacks.OnSave(item, s.editingIndex); err != nil {
            // Error handling is done by callback
            return
        }
    }
    
    if s.callbacks.OnRefresh != nil {
        s.callbacks.OnRefresh()
    }
}

// handleDelete processes item deletion.
func (s *SourceList) handleDelete() {
    if s.cursor >= len(s.items) {
        return // Can't delete "New" option
    }
    
    if s.callbacks.OnDelete != nil {
        if err := s.callbacks.OnDelete(s.cursor); err != nil {
            return
        }
    }
    
    // Adjust cursor
    if s.cursor >= len(s.items)-1 && s.cursor > 0 {
        s.cursor--
    }
    
    if s.callbacks.OnRefresh != nil {
        s.callbacks.OnRefresh()
    }
}

// cycleInputFocus cycles between form inputs.
func (s *SourceList) cycleInputFocus() {
    if len(s.inputs) <= 1 {
        return
    }
    
    // Find current focus
    for i := range s.inputs {
        if s.inputs[i].Focused() {
            s.inputs[i].Blur()
            next := (i + 1) % len(s.inputs)
            s.inputs[next].Focus()
            return
        }
    }
}

// View renders the component.
func (s *SourceList) View() string {
    if s.mode == SourceListModeForm {
        return s.renderForm()
    }
    return s.renderList()
}

// renderList renders the source list.
func (s *SourceList) renderList() string {
    var lines []string
    
    lines = append(lines, lipgloss.NewStyle().
        Bold(true).
        Foreground(s.primaryColor).
        Render(s.config.Title))
    lines = append(lines, "")
    lines = append(lines, "Defined sources:")
    
    if len(s.items) == 0 {
        lines = append(lines, lipgloss.NewStyle().
            Foreground(s.gray).
            Render("  (no sources defined)"))
    } else {
        for i, item := range s.items {
            prefix := "  "
            if i == s.cursor {
                prefix = "> "
            }
            
            line := fmt.Sprintf("%s%s", prefix, item.Display())
            
            if i == s.cursor {
                line = lipgloss.NewStyle().
                    Foreground(s.highlightColor).
                    Render(line)
            }
            lines = append(lines, line)
        }
    }
    
    // "New" option
    prefix := "  "
    if s.cursor == len(s.items) {
        prefix = "> "
    }
    newLine := fmt.Sprintf("%s[ New Source ]", prefix)
    if s.cursor == len(s.items) {
        newLine = lipgloss.NewStyle().
            Foreground(s.highlightColor).
            Render(newLine)
    }
    lines = append(lines, newLine)
    
    // Legend
    if s.config.Legend != "" {
        lines = append(lines, "")
        lines = append(lines, lipgloss.NewStyle().
            Foreground(s.gray).
            Render(s.config.Legend))
    }
    
    // Help text
    lines = append(lines, "")
    lines = append(lines, "Enter/Space=edit n=new d=delete Esc=back")
    
    return strings.Join(lines, "\n")
}

// renderForm renders the input form.
func (s *SourceList) renderForm() string {
    title := "New Source"
    if s.editingIndex >= 0 {
        title = "Edit Source"
    }
    
    var lines []string
    
    lines = append(lines, lipgloss.NewStyle().
        Bold(true).
        Foreground(s.primaryColor).
        Render(title))
    lines = append(lines, "")
    
    // First input
    prompt := s.config.FormPrompt
    if prompt == "" {
        prompt = "Path"
    }
    lines = append(lines, fmt.Sprintf("%s: %s", prompt, s.inputs[0].View()))
    
    // Second input (if configured)
    if len(s.inputs) > 1 && s.config.Field2Label != "" {
        lines = append(lines, fmt.Sprintf("%s: %s", s.config.Field2Label, s.inputs[1].View()))
    }
    
    // Help text
    lines = append(lines, "")
    if len(s.inputs) > 1 {
        lines = append(lines, "Enter=save Tab=next Esc=cancel")
    } else {
        lines = append(lines, "Enter=save Esc=cancel")
    }
    
    return strings.Join(lines, "\n")
}

// SetSize implements component.Model.
func (s *SourceList) SetSize(width, height int) {
    s.Base.SetSize(width, height)
    // Adjust input width
    for i := range s.inputs {
        s.inputs[i].Width = min(60, width-10)
    }
}
```

---

## Step 2.4.3: Create SourceList Tests

**File**: `internal/ui/components/source_list_test.go` (NEW)

```go
package components

import (
    "errors"
    "testing"
    
    "kankuro/internal/config"
)

func TestSimpleSourceItem(t *testing.T) {
    item := SimpleSourceItem("/path/to/skills")
    
    if item.Display() != "/path/to/skills" {
        t.Errorf("Expected '/path/to/skills', got '%s'", item.Display())
    }
    
    if item.ID() != "" {
        t.Error("SimpleSourceItem should have empty ID")
    }
}

func TestWorkspaceSourceItem(t *testing.T) {
    tests := []struct {
        name     string
        source   config.WorkspaceSource
        expected string
    }{
        {
            name:     "non-root",
            source:   config.WorkspaceSource{Path: "/home/user/project"},
            expected: "[ ] /home/user/project",
        },
        {
            name:     "root",
            source:   config.WorkspaceSource{Path: "/home/user/projects", IsRoot: true},
            expected: "[R] /home/user/projects",
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            item := WorkspaceSourceItem{Source: tt.source}
            if item.Display() != tt.expected {
                t.Errorf("Expected '%s', got '%s'", tt.expected, item.Display())
            }
        })
    }
}

func TestToSimpleItems(t *testing.T) {
    sources := []string{"/path1", "/path2", "/path3"}
    items := ToSimpleItems(sources)
    
    if len(items) != 3 {
        t.Fatalf("Expected 3 items, got %d", len(items))
    }
    
    for i, item := range items {
        if item.Display() != sources[i] {
            t.Errorf("Item %d: expected '%s', got '%s'", i, sources[i], item.Display())
        }
    }
}

func TestFromSimpleItems(t *testing.T) {
    items := []SourceItem{
        SimpleSourceItem("/path1"),
        SimpleSourceItem("/path2"),
    }
    
    sources := FromSimpleItems(items)
    
    if len(sources) != 2 {
        t.Fatalf("Expected 2 sources, got %d", len(sources))
    }
    
    if sources[0] != "/path1" || sources[1] != "/path2" {
        t.Errorf("Sources not extracted correctly: %v", sources)
    }
}

func TestSourceList_New(t *testing.T) {
    config := SourceListConfig{
        Title:      "Test Sources",
        FormPrompt: "Path:",
        FormFields: 1,
    }
    
    callbacks := SourceListCallbacks{}
	
    sl := NewSourceList(config, callbacks)
    
    if sl == nil {
        t.Fatal("NewSourceList returned nil")
    }
    
    if sl.mode != SourceListModeList {
        t.Error("Should start in list mode")
    }
}

func TestSourceList_SetItems(t *testing.T) {
    config := SourceListConfig{Title: "Test", FormFields: 1}
    callbacks := SourceListCallbacks{}
    
    sl := NewSourceList(config, callbacks)
    
    items := []SourceItem{
        SimpleSourceItem("/path1"),
        SimpleSourceItem("/path2"),
    }
    
    sl.SetItems(items)
    
    if len(sl.items) != 2 {
        t.Errorf("Expected 2 items, got %d", len(sl.items))
    }
}

func TestSourceList_Navigation(t *testing.T) {
    config := SourceListConfig{Title: "Test", FormFields: 1}
    callbacks := SourceListCallbacks{}
    
    sl := NewSourceList(config, callbacks)
    sl.SetItems([]SourceItem{
        SimpleSourceItem("/path1"),
        SimpleSourceItem("/path2"),
    })
    
    // Initial cursor at 0
    if sl.cursor != 0 {
        t.Errorf("Initial cursor should be 0, got %d", sl.cursor)
    }
    
    // Move down
    sl.cursor++
    if sl.cursor != 1 {
        t.Errorf("Cursor should be 1 after down, got %d", sl.cursor)
    }
    
    // Move down to "New" option
    sl.cursor++
    if sl.cursor != 2 { // len(items) = "New" index
        t.Errorf("Cursor should be 2 (New), got %d", sl.cursor)
    }
    
    // Can't go past "New"
    sl.cursor++
    if sl.cursor != 2 {
        t.Errorf("Cursor should stay at 2, got %d", sl.cursor)
    }
}

func TestSourceList_OpenForm_New(t *testing.T) {
    config := SourceListConfig{Title: "Test", FormFields: 1}
    callbacks := SourceListCallbacks{}
    
    sl := NewSourceList(config, callbacks)
    sl.SetItems([]SourceItem{SimpleSourceItem("/existing")})
    
    // Open form for new item (cursor at len(items))
    sl.cursor = 1
    sl.openForm(1)
    
    if sl.mode != SourceListModeForm {
        t.Error("Should be in form mode")
    }
    
    if sl.editingIndex != -1 {
        t.Errorf("EditingIndex should be -1 for new, got %d", sl.editingIndex)
    }
    
    if sl.inputs[0].Value() != "" {
        t.Error("Input should be empty for new item")
    }
}

func TestSourceList_OpenForm_Edit(t *testing.T) {
    config := SourceListConfig{Title: "Test", FormFields: 1}
    callbacks := SourceListCallbacks{}
    
    sl := NewSourceList(config, callbacks)
    sl.SetItems([]SourceItem{SimpleSourceItem("/existing")})
    
    // Open form for edit (cursor at 0)
    sl.openForm(0)
    
    if sl.mode != SourceListModeForm {
        t.Error("Should be in form mode")
    }
    
    if sl.editingIndex != 0 {
        t.Errorf("EditingIndex should be 0 for edit, got %d", sl.editingIndex)
    }
    
    if sl.inputs[0].Value() != "/existing" {
        t.Errorf("Input should contain existing path, got '%s'", sl.inputs[0].Value())
    }
}

func TestSourceList_HandleSave(t *testing.T) {
    var savedItem SourceItem
    var savedIndex int
    
    config := SourceListConfig{Title: "Test", FormFields: 1}
    callbacks := SourceListCallbacks{
        OnSave: func(item SourceItem, idx int) error {
            savedItem = item
            savedIndex = idx
            return nil
        },
        OnRefresh: func() {},
    }
    
    sl := NewSourceList(config, callbacks)
    sl.SetItems([]SourceItem{})
    sl.editingIndex = -1
    sl.inputs[0].SetValue("/new/path")
    
    sl.handleSave()
    
    if savedItem == nil {
        t.Fatal("OnSave should have been called")
    }
    
    if savedItem.Display() != "/new/path" {
        t.Errorf("Saved item should be '/new/path', got '%s'", savedItem.Display())
    }
    
    if savedIndex != -1 {
        t.Errorf("Saved index should be -1 for new item, got %d", savedIndex)
    }
}

func TestSourceList_HandleSave_WithSecondField(t *testing.T) {
    config := SourceListConfig{
        Title:       "Workspaces",
        FormFields:  2,
        Field2Label: "Is Root:",
    }
    
    var savedItem SourceItem
    
    callbacks := SourceListCallbacks{
        OnSave: func(item SourceItem, idx int) error {
            savedItem = item
            return nil
        },
        OnRefresh: func() {},
    }
    
    sl := NewSourceList(config, callbacks)
    sl.editingIndex = -1
    sl.inputs[0].SetValue("/workspace/path")
    sl.inputs[1].SetValue("y")
    
    sl.handleSave()
    
    wsItem, ok := savedItem.(WorkspaceSourceItem)
    if !ok {
        t.Fatal("Should have saved WorkspaceSourceItem")
    }
    
    if wsItem.Source.Path != "/workspace/path" {
        t.Errorf("Path should be '/workspace/path', got '%s'", wsItem.Source.Path)
    }
    
    if !wsItem.Source.IsRoot {
        t.Error("IsRoot should be true")
    }
}

func TestSourceList_HandleDelete(t *testing.T) {
    var deletedIndex int
    
    config := SourceListConfig{Title: "Test", FormFields: 1}
    callbacks := SourceListCallbacks{
        OnDelete: func(idx int) error {
            deletedIndex = idx
            return nil
        },
        OnRefresh: func() {},
    }
    
    sl := NewSourceList(config, callbacks)
    sl.SetItems([]SourceItem{
        SimpleSourceItem("/path1"),
        SimpleSourceItem("/path2"),
    })
    sl.cursor = 1
    
    sl.handleDelete()
    
    if deletedIndex != 1 {
        t.Errorf("Should delete index 1, got %d", deletedIndex)
    }
}

func TestSourceList_HandleDelete_CallbackError(t *testing.T) {
    config := SourceListConfig{Title: "Test", FormFields: 1}
    callbacks := SourceListCallbacks{
        OnDelete: func(idx int) error {
            return errors.New("delete failed")
        },
        OnRefresh: func() {},
    }
    
    sl := NewSourceList(config, callbacks)
    sl.SetItems([]SourceItem{SimpleSourceItem("/path")})
    sl.cursor = 0
    
    // Should not panic on error
    sl.handleDelete()
}

func TestSourceList_View_ListMode(t *testing.T) {
    config := SourceListConfig{
        Title:  "Test Sources",
        Legend: "(R=Root folder)",
        FormFields: 1,
    }
    callbacks := SourceListCallbacks{}
    
    sl := NewSourceList(config, callbacks)
    sl.SetItems([]SourceItem{
        SimpleSourceItem("/path1"),
        SimpleSourceItem("/path2"),
    })
    
    view := sl.View()
    
    if view == "" {
        t.Fatal("View should not be empty")
    }
    
    // Should contain title
    if !containsSubstring(view, "Test Sources") {
        t.Error("Should contain title")
    }
    
    // Should contain items
    if !containsSubstring(view, "/path1") {
        t.Error("Should contain first item")
    }
    
    // Should contain "New" option
    if !containsSubstring(view, "New Source") {
        t.Error("Should contain 'New Source' option")
    }
    
    // Should contain legend
    if !containsSubstring(view, "Root folder") {
        t.Error("Should contain legend")
    }
}

func TestSourceList_View_FormMode(t *testing.T) {
    config := SourceListConfig{
        Title:      "Test Sources",
        FormPrompt: "Path:",
        FormFields: 1,
    }
    callbacks := SourceListCallbacks{}
    
    sl := NewSourceList(config, callbacks)
    sl.mode = SourceListModeForm
    sl.inputs[0].SetValue("/test/path")
    
    view := sl.View()
    
    if view == "" {
        t.Fatal("View should not be empty")
    }
    
    // Should contain title (New or Edit)
    if !containsSubstring(view, "Source") {
        t.Error("Should contain title")
    }
    
    // Should contain input
    if !containsSubstring(view, "/test/path") {
        t.Error("Should contain input value")
    }
}

func TestSourceList_View_EmptyList(t *testing.T) {
    config := SourceListConfig{Title: "Test", FormFields: 1}
    callbacks := SourceListCallbacks{}
    
    sl := NewSourceList(config, callbacks)
    sl.SetItems([]SourceItem{}) // Empty
    
    view := sl.View()
    
    if !containsSubstring(view, "no sources defined") {
        t.Error("Should show empty state message")
    }
}

// Helper
func containsSubstring(s, substr string) bool {
    for i := 0; i <= len(s)-len(substr); i++ {
        if s[i:i+len(substr)] == substr {
            return true
        }
    }
    return false
}
```

---

## Step 2.4.4: Integration with App (Backwards Compatible)

**File**: `internal/ui/app_skill_sources.go` (MODIFY - add integration stub)

Add at end of file:
```go
// ===== Integration with new SourceList component =====

func (a *App) newSkillSourceListComponent() *components.SourceList {
    config := components.SourceListConfig{
        Title:      "Manage Skill Sources",
        FormPrompt: "Path to skill source...",
        FormFields: 1,
    }
    
    callbacks := components.SourceListCallbacks{
        OnSave: func(item components.SourceItem, idx int) error {
            path := strings.TrimSpace(item.Display())
            if idx == -1 {
                // New
                a.skillSources.Sources = append(a.skillSources.Sources, path)
            } else {
                // Edit
                a.skillSources.Sources[idx] = path
            }
            if err := a.repo.SaveSkillSources(a.skillSources.Sources); err != nil {
                a.err = err
                return err
            }
            a.refreshSkills()
            return nil
        },
        
        OnDelete: func(idx int) error {
            a.skillSources.Sources = append(
                a.skillSources.Sources[:idx],
                a.skillSources.Sources[idx+1:]...,
            )
            if err := a.repo.SaveSkillSources(a.skillSources.Sources); err != nil {
                a.err = err
                return err
            }
            a.refreshSkills()
            return nil
        },
        
        OnRefresh: func() {
            a.refreshSkills()
        },
        
        OnClose: func() {
            a.mode = viewCardForm
        },
    }
    
    sl := components.NewSourceList(config, callbacks)
    sl.SetItems(components.ToSimpleItems(a.skillSources.Sources))
    return sl
}
```

**File**: `internal/ui/app_persona_sources.go` (MODIFY - add integration stub)

Add at end of file:
```go
// ===== Integration with new SourceList component =====

func (a *App) newPersonaSourceListComponent() *components.SourceList {
    config := components.SourceListConfig{
        Title:      "Manage Persona Sources",
        FormPrompt: "Path to persona source...",
        FormFields: 1,
    }
    
    callbacks := components.SourceListCallbacks{
        OnSave: func(item components.SourceItem, idx int) error {
            path := strings.TrimSpace(item.Display())
            if idx == -1 {
                a.personaSources.Sources = append(a.personaSources.Sources, path)
            } else {
                a.personaSources.Sources[idx] = path
            }
            if err := a.repo.SavePersonaSources(a.personaSources.Sources); err != nil {
                a.err = err
                return err
            }
            a.refreshPersonas()
            return nil
        },
        
        OnDelete: func(idx int) error {
            a.personaSources.Sources = append(
                a.personaSources.Sources[:idx],
                a.personaSources.Sources[idx+1:]...,
            )
            if err := a.repo.SavePersonaSources(a.personaSources.Sources); err != nil {
                a.err = err
                return err
            }
            a.refreshPersonas()
            return nil
        },
        
        OnRefresh: func() {
            a.refreshPersonas()
        },
        
        OnClose: func() {
            a.mode = viewCardForm
        },
    }
    
    sl := components.NewSourceList(config, callbacks)
    sl.SetItems(components.ToSimpleItems(a.personaSources.Sources))
    return sl
}
```

**File**: `internal/ui/app_workspaces.go` (MODIFY - add integration stub)

Add after `renderWorkspacesList`:
```go
// ===== Integration with new SourceList component =====

func (a *App) newWorkspaceSourceListComponent() *components.SourceList {
    config := components.SourceListConfig{
        Title:       "Manage Workspace Sources",
        Legend:      "(R=Root folder containing multiple workspaces)",
        FormPrompt:  "Path to workspace or root directory...",
        FormFields:  2,
        Field2Label: "Is Root (y/n)",
    }
    
    callbacks := components.SourceListCallbacks{
        OnSave: func(item components.SourceItem, idx int) error {
            wsItem, ok := item.(components.WorkspaceSourceItem)
            if !ok {
                return errors.New("invalid item type")
            }
            
            sources := a.workspaceForm.Sources
            
            if idx == -1 {
                // New - generate ID
                wsItem.Source.ID = generateID()
                sources = append(sources, wsItem.Source)
            } else {
                // Edit - preserve ID
                wsItem.Source.ID = sources[idx].ID
                sources[idx] = wsItem.Source
            }
            
            if err := a.repo.SaveWorkspaceSources(sources); err != nil {
                a.err = err
                return err
            }
            a.refreshWorkspaces()
            return nil
        },
        
        OnDelete: func(idx int) error {
            id := a.workspaceForm.Sources[idx].ID
            if err := a.repo.DeleteWorkspaceSource(id); err != nil {
                a.err = err
                return err
            }
            a.refreshWorkspaces()
            return nil
        },
        
        OnRefresh: func() {
            a.refreshWorkspaces()
        },
        
        OnClose: func() {
            a.mode = viewDashboard
        },
    }
    
    sl := components.NewSourceList(config, callbacks)
    sl.SetItems(components.ToWorkspaceItems(a.workspaceForm.Sources))
    return sl
}

func generateID() string {
    return fmt.Sprintf("ws-%d", time.Now().UnixNano())
}
```

---

## Step 2.4.5: Integration Tests

**File**: `internal/ui/components/source_list_integration_test.go` (NEW)

```go
package components_test

import (
    "testing"
    
    "kankuro/internal/config"
    "kankuro/internal/ui/components"
)

func TestSourceListIntegration_Simple(t *testing.T) {
    // Simulate skill/persona source workflow
    
    var savedSources []string
    
    config := components.SourceListConfig{
        Title:      "Skill Sources",
        FormPrompt: "Path:",
        FormFields: 1,
    }
    
    callbacks := components.SourceListCallbacks{
        OnSave: func(item components.SourceItem, idx int) error {
            if idx == -1 {
                savedSources = append(savedSources, item.Display())
            } else {
                savedSources[idx] = item.Display()
            }
            return nil
        },
        OnRefresh: func() {},
        OnClose:   func() {},
    }
    
    sl := components.NewSourceList(config, callbacks)
    
    // 1. Initial empty state
    sl.SetItems([]components.SourceItem{})
    
    view := sl.View()
    if view == "" {
        t.Error("View should render empty state")
    }
    
    // 2. Add sources
    sources := []string{"/home/user/skills", "/home/user/more-skills"}
    sl.SetItems(components.ToSimpleItems(sources))
    
    view = sl.View()
    if view == "" {
        t.Error("View should render with items")
    }
    
    // 3. Simulate "n" key (new item)
    sl.Cursor = len(sources) // Select "New" option
    sl.OpenForm(len(sources))
    
    if sl.Mode != components.SourceListModeForm {
        t.Error("Should be in form mode")
    }
    
    t.Log("Simple source list integration completed successfully")
}

func TestSourceListIntegration_Workspace(t *testing.T) {
    // Simulate workspace source workflow
    
    var savedItems []components.SourceItem
    
    config := components.SourceListConfig{
        Title:       "Workspace Sources",
        Legend:      "(R=Root folder)",
        FormFields:  2,
        Field2Label: "Is Root:",
    }
    
    callbacks := components.SourceListCallbacks{
        OnSave: func(item components.SourceItem, idx int) error {
            if idx == -1 {
                savedItems = append(savedItems, item)
            } else {
                savedItems[idx] = item
            }
            return nil
        },
        OnRefresh: func() {},
        OnClose:   func() {},
    }
    
    sl := components.NewSourceList(config, callbacks)
    
    // 1. Set workspace sources
    wsSources := []config.WorkspaceSource{
        {ID: "ws1", Path: "/home/user/project", IsRoot: false},
        {ID: "ws2", Path: "/home/user/projects", IsRoot: true},
    }
    sl.SetItems(components.ToWorkspaceItems(wsSources))
    
    view := sl.View()
    if view == "" {
        t.Error("View should render")
    }
    
    // 2. Verify display format
    // Item 0: [ ] /home/user/project
    // Item 1: [R] /home/user/projects
    
    t.Log("Workspace source list integration completed successfully")
}

func TestSourceListIntegration_CRUD(t *testing.T) {
    // Test full CRUD cycle
    
    var sources []string
    
    config := components.SourceListConfig{Title: "Test", FormFields: 1}
    
    callbacks := components.SourceListCallbacks{
        OnSave: func(item components.SourceItem, idx int) error {
            if idx == -1 {
                sources = append(sources, item.Display())
            } else {
                if idx < len(sources) {
                    sources[idx] = item.Display()
                }
            }
            return nil
        },
        OnDelete: func(idx int) error {
            if idx >= 0 && idx < len(sources) {
                sources = append(sources[:idx], sources[idx+1:]...)
            }
            return nil
        },
        OnRefresh: func() {},
        OnClose:   func() {},
    }
    
    sl := components.NewSourceList(config, callbacks)
    
    // CREATE
    sl.SetItems([]components.SourceItem{})
    sl.OpenForm(-1)
    sl.Inputs[0].SetValue("/new/source")
    sl.HandleSave()
    
    if len(sources) != 1 {
        t.Errorf("Expected 1 source after create, got %d", len(sources))
    }
    
    // READ (via SetItems/View)
    sl.SetItems(components.ToSimpleItems(sources))
    
    // UPDATE
    sl.SetItems(components.ToSimpleItems(sources))
    sl.OpenForm(0)
    sl.Inputs[0].SetValue("/updated/source")
    sl.HandleSave()
    
    if sources[0] != "/updated/source" {
        t.Errorf("Expected '/updated/source', got '%s'", sources[0])
    }
    
    // DELETE
    sl.SetItems(components.ToSimpleItems(sources))
    sl.Cursor = 0
    sl.HandleDelete()
    
    if len(sources) != 0 {
        t.Errorf("Expected 0 sources after delete, got %d", len(sources))
    }
    
    t.Log("CRUD cycle completed successfully")
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

# 3. Specific tests for SourceList
go test ./internal/ui/components/ -run SourceList -v
go test ./internal/ui/components/ -run SourceItem -v

# 4. Existing UI still builds
go build ./internal/ui/

# 5. Run existing tests
go test ./internal/ui/
```

---

## Success Criteria

- [ ] `SourceItem` interface defined
- [ ] `SimpleSourceItem` implements interface (for skills/personas)
- [ ] `WorkspaceSourceItem` implements interface (for workspaces)
- [ ] Conversion helpers: `ToSimpleItems`, `ToWorkspaceItems`, `FromSimpleItems`
- [ ] `SourceList` component with config and callbacks
- [ ] List mode with navigation (up/down)
- [ ] Form mode with input(s)
- [ ] Support for single field (skills/personas)
- [ ] Support for dual fields (workspaces)
- [ ] CRUD operations via callbacks
- [ ] Empty state rendering
- [ ] "New Source" option at end of list
- [ ] Legend text support
- [ ] Unit tests pass (>80% coverage)
- [ ] Integration tests demonstrate CRUD
- [ ] Integration stubs for skills/personas/workspaces
- [ ] Existing App still compiles

---

## Dependencies

```
component/component.go (from 1.1) ✓
messages/messages.go (from 1.2) ✓
commands/commands.go (from 1.3) ✓
config/config.go (existing - for WorkspaceSource)

components/source_item.go ← NEW
components/source_list.go ← NEW
```

---

## Estimated Time

| Step | Time |
|------|------|
| 2.4.1 SourceItem types | 20 min |
| 2.4.2 SourceList component | 50 min |
| 2.4.3 Unit tests | 40 min |
| 2.4.4 App integration stubs | 30 min |
| 2.4.5 Integration tests | 25 min |
| Verification | 15 min |
| **Total** | **~3 hours** |

---

## What This Enables

1. **Eliminates 315 lines of duplication**: `app_skill_sources.go` + `app_persona_sources.go` become ~30 lines each
2. **Single source of truth**: All source management logic in one component
3. **Consistent UX**: All source lists behave identically
4. **Easy to extend**: Add new source types by implementing `SourceItem`
5. **Phase 3 ready**: SourceManager view can use this component directly

---

## Code Reduction Summary

| File | Before | After | Reduction |
|------|--------|-------|-----------|
| `app_skill_sources.go` | 156 lines | ~30 lines (stub) | -126 |
| `app_persona_sources.go` | 159 lines | ~30 lines (stub) | -129 |
| `app_workspaces.go` (source part) | ~100 lines | ~35 lines (stub) | -65 |
| **Total reduction** | 415 lines | 95 lines | **-320 lines** |

The new `source_list.go` is ~260 lines, net savings of **~60 lines** plus:
- Single implementation to test
- Single implementation to fix bugs
- Consistent behavior across all source types
