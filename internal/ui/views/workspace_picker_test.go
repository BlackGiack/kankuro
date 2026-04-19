package views_test

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"kankuro/internal/config"
	"kankuro/internal/ui/commands"
	"kankuro/internal/ui/messages"
	"kankuro/internal/ui/views"
)

func TestWorkspacePicker_SearchAndFilter(t *testing.T) {
	e := views.NewCardEditor(views.CardEditorCallbacks{}, &commands.Commands{})
	workspaces := []config.Workspace{
		{ID: "ws-1", Name: "Alpha Project", Path: "/path/alpha"},
		{ID: "ws-2", Name: "Beta Project", Path: "/path/beta"},
		{ID: "ws-3", Name: "Gamma Project", Path: "/path/gamma"},
	}
	e.SetData(messages.DataLoadedMsg{Workspaces: workspaces})

	// Open picker (simulated by focusing workspace and pressing space)
	e.CycleFocus(3) // FocusWorkspace is 3
	e.Update(tea.KeyMsg{Type: tea.KeySpace, Runes: []rune{' '}})

	view := e.View()
	if !strings.Contains(view, "Alpha Project") || !strings.Contains(view, "Beta Project") {
		t.Fatal("Picker should show all workspaces initially")
	}

	// Type 'b' to filter
	e.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'b'}})
	view = e.View()
	if !strings.Contains(view, "Beta Project") {
		t.Error("View should contain Beta Project")
	}
	if strings.Contains(view, "Alpha Project") {
		t.Error("View should NOT contain Alpha Project after filtering for 'b'")
	}
	if !strings.Contains(view, "Search: b") {
		t.Error("View should show current search query")
	}

	// Backspace to clear 'b'
	e.Update(tea.KeyMsg{Type: tea.KeyBackspace})
	view = e.View()
	if !strings.Contains(view, "Alpha Project") || !strings.Contains(view, "Beta Project") {
		t.Error("Picker should show all workspaces after clearing filter with backspace")
	}
}

func TestWorkspacePicker_EscapeBehavior(t *testing.T) {
	e := views.NewCardEditor(views.CardEditorCallbacks{}, &commands.Commands{})
	e.SetData(messages.DataLoadedMsg{Workspaces: []config.Workspace{{ID: "1", Name: "Alpha"}}})

	// Open picker
	e.CycleFocus(3)
	e.Update(tea.KeyMsg{Type: tea.KeySpace, Runes: []rune{' '}})

	// Type 'a'
	e.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	if !strings.Contains(e.View(), "Search: a") {
		t.Fatal("Search should be active")
	}

	// First Esc should clear search
	e.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if strings.Contains(e.View(), "Search: a") {
		t.Error("Esc should have cleared search")
	}
	if !strings.Contains(e.View(), "Select Workspace") {
		t.Error("Picker should still be open after clearing search")
	}

	// Second Esc should close picker
	e.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if strings.Contains(e.View(), "Select Workspace") {
		t.Error("Second Esc should have closed picker")
	}
}

func TestWorkspacePicker_QNonQuit(t *testing.T) {
	e := views.NewCardEditor(views.CardEditorCallbacks{}, &commands.Commands{})
	e.SetData(messages.DataLoadedMsg{Workspaces: []config.Workspace{{ID: "1", Name: "Alpha"}}})

	// Open picker
	e.CycleFocus(3)
	e.Update(tea.KeyMsg{Type: tea.KeySpace, Runes: []rune{' '}})

	// Press 'q'
	_, cmd := e.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	if cmd != nil {
		t.Error("Pressing 'q' in picker should not return any command (like tea.Quit)")
	}
	if !strings.Contains(e.View(), "Select Workspace") {
		t.Error("Picker should still be open after pressing 'q'")
	}
}

func TestWorkspacePicker_DedicatedRendering(t *testing.T) {
	e := views.NewCardEditor(views.CardEditorCallbacks{}, &commands.Commands{})
	e.SetSize(100, 50)

	// Open picker
	e.CycleFocus(3)
	e.Update(tea.KeyMsg{Type: tea.KeySpace, Runes: []rune{' '}})

	view := e.View()
	// It should NOT contain the "New ID Card" title from the form
	if strings.Contains(view, "New ID Card") {
		t.Error("Picker view should be dedicated and not contain the background form")
	}
	if !strings.Contains(view, "Select Workspace") {
		t.Error("View should show the picker")
	}
}

// TestWorkspacePicker_ReopenAfterSearch tests the exact crash scenario:
// open -> search -> select -> close -> reopen -> render.
func TestWorkspacePicker_ReopenAfterSearch(t *testing.T) {
	e := views.NewCardEditor(views.CardEditorCallbacks{}, &commands.Commands{})
	e.SetSize(100, 50)

	// Setup workspaces
	workspaces := []config.Workspace{
		{ID: "ws-1", Name: "Alpha Project", Path: "/path/alpha"},
		{ID: "ws-2", Name: "Beta Project", Path: "/path/beta"},
		{ID: "ws-3", Name: "Gamma Project", Path: "/path/gamma"},
		{ID: "ws-4", Name: "Delta Project", Path: "/path/delta"},
		{ID: "ws-5", Name: "Epsilon Project", Path: "/path/epsilon"},
	}
	e.SetData(messages.DataLoadedMsg{Workspaces: workspaces})

	// First open
	e.CycleFocus(3) // FocusWorkspace
	e.Update(tea.KeyMsg{Type: tea.KeySpace, Runes: []rune{' '}})

	// Search for 'a' to filter results
	e.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	view := e.View()
	if !strings.Contains(view, "Search: a") {
		t.Error("Search should be active")
	}

	// Select first filtered item
	e.Update(tea.KeyMsg{Type: tea.KeyEnter})

	// Verify selection was saved
	if e.WorkspaceID() == "" {
		t.Error("Workspace ID should be set after selection")
	}

	// Reopen picker - this is where the crash would occur
	e.Update(tea.KeyMsg{Type: tea.KeySpace, Runes: []rune{' '}})

	// This should not panic - if it does, the test will fail
	view = e.View()
	if !strings.Contains(view, "Select Workspace") {
		t.Error("Picker should be open after reopening")
	}

	// Verify search state was reset
	if strings.Contains(view, "Search: a") {
		t.Error("Search query should be reset after reopening")
	}

	// Verify selected workspace is highlighted
	if !strings.Contains(view, e.WorkspaceSummary()) {
		t.Error("Selected workspace should be visible in picker")
	}
}

// TestWorkspacePicker_SizeRestoration tests that picker dimensions are preserved across reopens.
func TestWorkspacePicker_SizeRestoration(t *testing.T) {
	e := views.NewCardEditor(views.CardEditorCallbacks{}, &commands.Commands{})
	e.SetSize(50, 20) // Non-default size

	workspaces := []config.Workspace{
		{ID: "ws-1", Name: "Alpha Project", Path: "/path/alpha"},
		{ID: "ws-2", Name: "Beta Project", Path: "/path/beta"},
	}
	e.SetData(messages.DataLoadedMsg{Workspaces: workspaces})

	// Open picker
	e.CycleFocus(3)
	e.Update(tea.KeyMsg{Type: tea.KeySpace, Runes: []rune{' '}})

	// Get the view from the first open
	view1 := e.View()

	// Close and reopen
	e.Update(tea.KeyMsg{Type: tea.KeyEsc})
	e.Update(tea.KeyMsg{Type: tea.KeySpace, Runes: []rune{' '}})

	// Get the view from the second open
	view2 := e.View()

	// Both views should have similar structure (same dimensions)
	// This verifies that the list dimensions were preserved
	if len(view1) == 0 || len(view2) == 0 {
		t.Error("Both views should have content")
	}

	// The views should be structurally similar (same border, similar content)
	lines1 := strings.Count(view1, "\n")
	lines2 := strings.Count(view2, "\n")
	if lines1 != lines2 {
		t.Errorf("View structure should be preserved: first open had %d lines, second open had %d lines", lines1, lines2)
	}
}

// TestWorkspacePicker_SmallTerminal tests that small terminal sizes don't cause panics.
func TestWorkspacePicker_SmallTerminal(t *testing.T) {
	e := views.NewCardEditor(views.CardEditorCallbacks{}, &commands.Commands{})
	e.SetSize(20, 10) // Very small terminal

	workspaces := []config.Workspace{
		{ID: "ws-1", Name: "Alpha", Path: "/path/alpha"},
		{ID: "ws-2", Name: "Beta", Path: "/path/beta"},
	}
	e.SetData(messages.DataLoadedMsg{Workspaces: workspaces})

	// Open picker on small terminal
	e.CycleFocus(3)
	e.Update(tea.KeyMsg{Type: tea.KeySpace, Runes: []rune{' '}})

	// This should not panic even with tiny dimensions
	view := e.View()
	// The picker might not show the full title on very small terminals, but it shouldn't crash
	if !strings.Contains(view, "Workspace") && !strings.Contains(view, "Alpha") {
		t.Error("Picker should work on small terminals and show some content")
	}
}

// TestWorkspacePicker_SelectionRestoration tests that selecting a workspace actually saves the new selection
// and doesn't restore the old workspace due to premature state reset.
func TestWorkspacePicker_SelectionRestoration(t *testing.T) {
	e := views.NewCardEditor(views.CardEditorCallbacks{}, &commands.Commands{})
	e.SetSize(100, 50)

	workspaces := []config.Workspace{
		{ID: "ws-1", Name: "Alpha Project", Path: "/path/alpha"},
		{ID: "ws-2", Name: "Beta Project", Path: "/path/beta"},
		{ID: "ws-3", Name: "Gamma Project", Path: "/path/gamma"},
	}
	e.SetData(messages.DataLoadedMsg{Workspaces: workspaces})

	// Start with Alpha selected
	e.SetWorkspaceID("ws-1")
	if e.WorkspaceID() != "ws-1" {
		t.Error("Should start with Alpha selected")
	}

	// Open picker
	e.CycleFocus(3)
	e.Update(tea.KeyMsg{Type: tea.KeySpace, Runes: []rune{' '}})

	// Navigate to Gamma and select it
	e.Update(tea.KeyMsg{Type: tea.KeyDown})  // Move to Beta
	e.Update(tea.KeyMsg{Type: tea.KeyDown})  // Move to Gamma
	e.Update(tea.KeyMsg{Type: tea.KeyEnter}) // Select Gamma

	// Verify that Gamma was actually selected, not the original Alpha
	if e.WorkspaceID() != "ws-3" {
		t.Errorf("Expected Gamma (ws-3) to be selected, but got %s", e.WorkspaceID())
	}

	// Reopen picker and verify Gamma is still selected
	e.Update(tea.KeyMsg{Type: tea.KeySpace, Runes: []rune{' '}})
	view := e.View()
	if !strings.Contains(view, "Gamma Project") {
		t.Error("Gamma should still be selected after reopening")
	}
}

// TestWorkspacePicker_Filtering tests that workspace filtering works correctly.
func TestWorkspacePicker_Filtering(t *testing.T) {
	e := views.NewCardEditor(views.CardEditorCallbacks{}, &commands.Commands{})
	e.SetSize(100, 50)

	workspaces := []config.Workspace{
		{ID: "ws-1", Name: "Alpha Project", Path: "/path/alpha"},
		{ID: "ws-2", Name: "Beta Project", Path: "/path/beta"},
		{ID: "ws-3", Name: "Gamma Project", Path: "/path/gamma"},
	}
	e.SetData(messages.DataLoadedMsg{Workspaces: workspaces})

	// Open picker
	e.CycleFocus(3)
	e.Update(tea.KeyMsg{Type: tea.KeySpace, Runes: []rune{' '}})

	// All workspaces should be visible initially
	view := e.View()
	if !strings.Contains(view, "Alpha Project") || !strings.Contains(view, "Beta Project") || !strings.Contains(view, "Gamma Project") {
		t.Error("All workspaces should be visible initially")
	}

	// Search for 'lph' to filter to Alpha only (unique part of Alpha)
	e.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'l'}})
	e.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'p'}})
	e.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'h'}})

	view = e.View()
	if !strings.Contains(view, "Search: lph") {
		t.Error("Search query should be active")
	}
	if !strings.Contains(view, "Alpha Project") {
		t.Error("Filtered results should show Alpha")
	}
	// Beta and Gamma should be filtered out - they don't contain 'lph'
	if strings.Contains(view, "Beta Project") {
		t.Error("Beta should be filtered out")
	}
	if strings.Contains(view, "Gamma Project") {
		t.Error("Gamma should be filtered out")
	}

	// Clear search
	e.Update(tea.KeyMsg{Type: tea.KeyEsc})
	view = e.View()

	// All workspaces should be visible again
	if !strings.Contains(view, "Alpha Project") || !strings.Contains(view, "Beta Project") || !strings.Contains(view, "Gamma Project") {
		t.Error("All workspaces should be visible after clearing search")
	}
}
