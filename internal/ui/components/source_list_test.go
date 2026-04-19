package components

import (
	"errors"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"

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

	if sl.Mode != SourceListModeList {
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

	if len(sl.Items) != 2 {
		t.Errorf("Expected 2 items, got %d", len(sl.Items))
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
	if sl.Cursor != 0 {
		t.Errorf("Initial cursor should be 0, got %d", sl.Cursor)
	}

	// Move down
	sl.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	if sl.Cursor != 1 {
		t.Errorf("Cursor should be 1 after down, got %d", sl.Cursor)
	}

	// Move down to "New" option
	sl.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	if sl.Cursor != 2 { // len(items) = "New" index
		t.Errorf("Cursor should be 2 (New), got %d", sl.Cursor)
	}

	// Can't go past "New"
	sl.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	if sl.Cursor != 2 {
		t.Errorf("Cursor should stay at 2, got %d", sl.Cursor)
	}
}

func TestSourceList_OpenForm_New(t *testing.T) {
	config := SourceListConfig{Title: "Test", FormFields: 1}
	callbacks := SourceListCallbacks{}

	sl := NewSourceList(config, callbacks)
	sl.SetItems([]SourceItem{SimpleSourceItem("/existing")})

	// Open form for new item (cursor at len(items))
	sl.Cursor = 1
	sl.OpenForm(1)

	if sl.Mode != SourceListModeForm {
		t.Error("Should be in form mode")
	}

	if sl.EditingIndex != -1 {
		t.Errorf("EditingIndex should be -1 for new, got %d", sl.EditingIndex)
	}

	if sl.Inputs[0].Value() != "" {
		t.Error("Input should be empty for new item")
	}
}

func TestSourceList_OpenForm_Edit(t *testing.T) {
	config := SourceListConfig{Title: "Test", FormFields: 1}
	callbacks := SourceListCallbacks{}

	sl := NewSourceList(config, callbacks)
	sl.SetItems([]SourceItem{SimpleSourceItem("/existing")})

	// Open form for edit (cursor at 0)
	sl.OpenForm(0)

	if sl.Mode != SourceListModeForm {
		t.Error("Should be in form mode")
	}

	if sl.EditingIndex != 0 {
		t.Errorf("EditingIndex should be 0 for edit, got %d", sl.EditingIndex)
	}

	if sl.Inputs[0].Value() != "/existing" {
		t.Errorf("Input should contain existing path, got '%s'", sl.Inputs[0].Value())
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
	sl.EditingIndex = -1
	sl.Inputs[0].SetValue("/new/path")

	sl.HandleSave()

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
	sl.EditingIndex = -1
	sl.Inputs[0].SetValue("/workspace/path")
	sl.Inputs[1].SetValue("y")

	sl.HandleSave()

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
	sl.Cursor = 1

	sl.HandleDelete()

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
	sl.Cursor = 0

	// Should not panic on error
	sl.HandleDelete()
}

func TestSourceList_View_ListMode(t *testing.T) {
	lipgloss.SetColorProfile(termenv.TrueColor)
	config := SourceListConfig{
		Title:      "Test Sources",
		Legend:     "(R=Root folder)",
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
	if !hasText(view, "Test Sources") {
		t.Error("Should contain title")
	}

	// Should contain items
	if !hasText(view, "/path1") {
		t.Error("Should contain first item")
	}

	// Should contain "New" option
	if !hasText(view, "New Source") {
		t.Error("Should contain 'New Source' option")
	}

	// Should contain legend
	if !hasText(view, "Root folder") {
		t.Error("Should contain legend")
	}
}

func TestSourceList_View_FormMode(t *testing.T) {
	lipgloss.SetColorProfile(termenv.TrueColor)
	config := SourceListConfig{
		Title:      "Test Sources",
		FormPrompt: "Path:",
		FormFields: 1,
	}
	callbacks := SourceListCallbacks{}

	sl := NewSourceList(config, callbacks)
	sl.Mode = SourceListModeForm
	sl.Inputs[0].SetValue("/test/path")

	view := sl.View()

	if view == "" {
		t.Fatal("View should not be empty")
	}

	// Should contain title (New or Edit)
	if !hasText(view, "Source") {
		t.Error("Should contain title")
	}

	// Should contain input
	// Text input view might contain ANSI codes or other decoration,
	// but the value should be present.
	if !hasText(view, "/test/path") {
		t.Error("Should contain input value")
	}
}

func TestSourceList_View_EmptyList(t *testing.T) {
	lipgloss.SetColorProfile(termenv.TrueColor)
	config := SourceListConfig{Title: "Test", FormFields: 1}
	callbacks := SourceListCallbacks{}

	sl := NewSourceList(config, callbacks)
	sl.SetItems([]SourceItem{}) // Empty

	view := sl.View()

	if !hasText(view, "no sources defined") {
		t.Error("Should show empty state message")
	}
}

// Helper
func hasText(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
