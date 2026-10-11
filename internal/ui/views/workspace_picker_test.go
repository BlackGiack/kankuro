package views_test

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"

	"kankuro/internal/config"
	"kankuro/internal/ui/messages"
	"kankuro/internal/ui/views"
)

func TestWorkspacePicker_SearchAndFilter(t *testing.T) {
	e := views.NewCardEditor(views.CardEditorCallbacks{})
	workspaces := []config.Workspace{
		{ID: "ws-1", Name: "Alpha Project", Path: "/path/alpha"},
		{ID: "ws-2", Name: "Beta Project", Path: "/path/beta"},
		{ID: "ws-3", Name: "Gamma Project", Path: "/path/gamma"},
	}
	e.SetData(messages.DataLoadedMsg{Workspaces: workspaces})

	// Open picker (simulated by focusing workspace and pressing space)
	e.CycleFocus(3) // FocusWorkspace is 3
	e.Update(tea.KeyPressMsg{Code: tea.KeySpace, Text: " "})

	view := e.View()
	if !strings.Contains(view, "Alpha Project") || !strings.Contains(view, "Beta Project") {
		t.Fatal("Picker should show all workspaces initially")
	}

	// Type 'b' to filter
	e.Update(tea.KeyPressMsg{Code: 0, Text: "b"})
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
	e.Update(tea.KeyPressMsg{Code: tea.KeyBackspace})
	view = e.View()
	if !strings.Contains(view, "Alpha Project") || !strings.Contains(view, "Beta Project") {
		t.Error("Picker should show all workspaces after clearing filter with backspace")
	}
}

func TestWorkspacePickerSpaceSearchesNamesWithSpaces(t *testing.T) {
	e := views.NewCardEditor(views.CardEditorCallbacks{})
	e.SetData(messages.DataLoadedMsg{Workspaces: []config.Workspace{
		{ID: "spaced", Name: "Alpha Project", Path: "/work/with space"},
		{ID: "joined", Name: "AlphaProject", Path: "/work/joined"},
	}})
	e.CycleFocus(views.FocusWorkspace)
	e.Update(tea.KeyPressMsg{Code: tea.KeySpace, Text: " "})
	for _, key := range []tea.KeyPressMsg{{Text: "Alpha"}, {Code: tea.KeySpace, Text: " "}, {Text: "Project"}} {
		e.Update(key)
	}
	view := e.View()
	hasExpectedQuery := strings.Contains(view, "Search: Alpha Project")
	hasSpacedWorkspace := strings.Contains(view, "Alpha Project")
	hasJoinedWorkspace := strings.Contains(view, "AlphaProject")
	if !hasExpectedQuery || !hasSpacedWorkspace || hasJoinedWorkspace {
		t.Fatalf("space should remain part of a single-picker search:\n%s", view)
	}
	e.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if got := e.WorkspaceID(); got != "spaced" {
		t.Fatalf("selected workspace = %q, want spaced", got)
	}
}

func TestSubWorkspacePickerMultipleUnresolvedAndMainModes(t *testing.T) {
	workspaces := []config.Workspace{
		{ID: "main", Name: "Main", Path: "/main"},
		{ID: "sub-a", Name: "Alpha", Path: "/alpha"},
		{ID: "sub-b", Name: "Beta", Path: "/beta"},
	}
	var saved config.IDCard
	e := views.NewCardEditor(views.CardEditorCallbacks{OnSave: func(card config.IDCard) tea.Cmd { saved = card; return nil }})
	e.SetData(messages.DataLoadedMsg{Workspaces: workspaces})
	e.SetCard(config.IDCard{WorkspaceID: "main", SubWorkspaceIDs: []string{"removed"}})
	e.CycleFocus(views.FocusSubWorkspaces)
	e.Update(tea.KeyPressMsg{Code: tea.KeySpace, Text: " "})
	out := e.View()
	if !strings.Contains(out, "Select Sub Workspaces") || !strings.Contains(out, "[missing] removed") || strings.Contains(out, "Main") {
		t.Fatalf("multiple picker must show unresolved selection and exclude unselected main workspace:\n%s", out)
	}
	// Provisional changes across searches are discarded by cancellation.
	e.Update(tea.KeyPressMsg{Code: 0, Text: "removed"})
	e.Update(tea.KeyPressMsg{Code: tea.KeySpace, Text: " "})
	e.Update(tea.KeyPressMsg{Code: tea.KeyEsc}) // clear search first
	e.Update(tea.KeyPressMsg{Code: tea.KeyEsc}) // cancel without committing
	if got := e.Card().SubWorkspaceIDs; len(got) != 1 || got[0] != "removed" {
		t.Fatalf("cancel changed editor selection: %v", got)
	}
	// Reopen, clear unresolved explicitly, and commit two checked IDs.
	e.Update(tea.KeyPressMsg{Code: tea.KeySpace, Text: " "})
	e.Update(tea.KeyPressMsg{Code: 0, Text: "removed"})
	e.Update(tea.KeyPressMsg{Code: tea.KeySpace, Text: " "})
	e.Update(tea.KeyPressMsg{Code: tea.KeyEsc})
	e.Update(tea.KeyPressMsg{Code: 0, Text: "Alpha"})
	e.Update(tea.KeyPressMsg{Code: tea.KeySpace, Text: " "})
	e.Update(tea.KeyPressMsg{Code: tea.KeyEsc})
	e.Update(tea.KeyPressMsg{Code: 0, Text: "Beta"})
	e.Update(tea.KeyPressMsg{Code: tea.KeySpace, Text: " "})
	e.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if got := e.Card().SubWorkspaceIDs; len(got) != 2 || got[0] != "sub-a" || got[1] != "sub-b" {
		t.Fatalf("confirmed sub workspace IDs = %v, want [sub-a sub-b]", got)
	}
	// Opening the main picker must switch back to single-select mode.
	e.CycleFocus(-1)
	e.Update(tea.KeyPressMsg{Code: tea.KeySpace, Text: " "})
	mainPickerView := e.View()
	if !strings.Contains(mainPickerView, "Select Workspace") || strings.Contains(mainPickerView, "Select Sub Workspaces") {
		t.Fatalf("main picker retained multiple mode:\n%s", mainPickerView)
	}
	e.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	e.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if len(saved.SubWorkspaceIDs) != 2 || saved.SubWorkspaceIDs[0] != "sub-a" || saved.SubWorkspaceIDs[1] != "sub-b" {
		t.Fatalf("saved repaired card has sub workspace IDs %v", saved.SubWorkspaceIDs)
	}
}

func TestSubWorkspaceNewSelectionSurvivesRegistryRefresh(t *testing.T) {
	alpha := config.Workspace{ID: "a", Name: "Alpha", Path: "/alpha"}
	beta := config.Workspace{ID: "b", Name: "Beta", Path: "/beta"}
	e := views.NewCardEditor(views.CardEditorCallbacks{})
	e.SetData(messages.DataLoadedMsg{Workspaces: []config.Workspace{alpha, beta}})
	e.SetCard(config.IDCard{WorkspaceID: "main"})
	e.CycleFocus(views.FocusSubWorkspaces)

	openAndSelectAlpha := func() {
		t.Helper()
		e.Update(tea.KeyPressMsg{Code: tea.KeySpace, Text: " "})
		e.Update(tea.KeyPressMsg{Text: "Alpha"})
		e.Update(tea.KeyPressMsg{Code: tea.KeySpace, Text: " "})
	}

	openAndSelectAlpha()
	e.SetData(messages.DataLoadedMsg{Workspaces: []config.Workspace{beta}})
	if view := e.View(); !strings.Contains(view, "[missing] Alpha (a)") {
		t.Fatalf("newly checked workspace disappeared after removal:\n%s", view)
	}
	e.SetData(messages.DataLoadedMsg{Workspaces: []config.Workspace{alpha, beta}})
	e.Update(tea.KeyPressMsg{Code: tea.KeyEsc}) // clear search
	e.Update(tea.KeyPressMsg{Code: tea.KeyEsc}) // cancel
	if got := e.Card().SubWorkspaceIDs; len(got) != 0 {
		t.Fatalf("cancel committed provisional selection: %v", got)
	}

	// Reopen, remove Alpha again, and explicitly deselect its unresolved row.
	openAndSelectAlpha()
	e.SetData(messages.DataLoadedMsg{Workspaces: []config.Workspace{beta}})
	if view := e.View(); !strings.Contains(view, "[missing] Alpha (a)") {
		t.Fatalf("missing workspace was not available for deselection:\n%s", view)
	}
	e.Update(tea.KeyPressMsg{Code: tea.KeySpace, Text: " "})
	e.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if got := e.Card().SubWorkspaceIDs; len(got) != 0 {
		t.Fatalf("explicitly deselected ID remains: %v", got)
	}

	// A newly checked ID that is absent at confirmation remains selected.
	e.SetData(messages.DataLoadedMsg{Workspaces: []config.Workspace{alpha, beta}})
	openAndSelectAlpha()
	e.SetData(messages.DataLoadedMsg{Workspaces: []config.Workspace{beta}})
	e.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if got := e.Card().SubWorkspaceIDs; len(got) != 1 || got[0] != "a" {
		t.Fatalf("newly checked ID disappeared on confirmation: %v", got)
	}
}

func TestWorkspacePicker_EscapeBehavior(t *testing.T) {
	e := views.NewCardEditor(views.CardEditorCallbacks{})
	e.SetData(messages.DataLoadedMsg{Workspaces: []config.Workspace{{ID: "1", Name: "Alpha"}}})

	// Open picker
	e.CycleFocus(3)
	e.Update(tea.KeyPressMsg{Code: tea.KeySpace, Text: " "})

	// Type 'a'
	e.Update(tea.KeyPressMsg{Code: 0, Text: "a"})
	if !strings.Contains(e.View(), "Search: a") {
		t.Fatal("Search should be active")
	}

	// First Esc should clear search
	e.Update(tea.KeyPressMsg{Code: tea.KeyEsc})
	if strings.Contains(e.View(), "Search: a") {
		t.Error("Esc should have cleared search")
	}
	if !strings.Contains(e.View(), "Select Workspace") {
		t.Error("Picker should still be open after clearing search")
	}

	// Second Esc should close picker
	e.Update(tea.KeyPressMsg{Code: tea.KeyEsc})
	if strings.Contains(e.View(), "Select Workspace") {
		t.Error("Second Esc should have closed picker")
	}
}

func TestWorkspacePicker_QNonQuit(t *testing.T) {
	e := views.NewCardEditor(views.CardEditorCallbacks{})
	e.SetData(messages.DataLoadedMsg{Workspaces: []config.Workspace{{ID: "1", Name: "Alpha"}}})

	// Open picker
	e.CycleFocus(3)
	e.Update(tea.KeyPressMsg{Code: tea.KeySpace, Text: " "})

	// Press 'q'
	_, cmd := e.Update(tea.KeyPressMsg{Code: 0, Text: "q"})
	if cmd != nil {
		t.Error("Pressing 'q' in picker should not return any command (like tea.Quit)")
	}
	if !strings.Contains(e.View(), "Select Workspace") {
		t.Error("Picker should still be open after pressing 'q'")
	}
}

func TestWorkspacePicker_DedicatedRendering(t *testing.T) {
	e := views.NewCardEditor(views.CardEditorCallbacks{})
	e.SetSize(100, 50)

	// Open picker
	e.CycleFocus(3)
	e.Update(tea.KeyPressMsg{Code: tea.KeySpace, Text: " "})

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
	e := views.NewCardEditor(views.CardEditorCallbacks{})
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
	e.Update(tea.KeyPressMsg{Code: tea.KeySpace, Text: " "})

	// Search for 'a' to filter results
	e.Update(tea.KeyPressMsg{Code: 0, Text: "a"})
	view := e.View()
	if !strings.Contains(view, "Search: a") {
		t.Error("Search should be active")
	}

	// Select first filtered item
	e.Update(tea.KeyPressMsg{Code: tea.KeyEnter})

	// Verify selection was saved
	if e.WorkspaceID() == "" {
		t.Error("Workspace ID should be set after selection")
	}

	// Reopen picker - this is where the crash would occur
	e.Update(tea.KeyPressMsg{Code: tea.KeySpace, Text: " "})

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
	e := views.NewCardEditor(views.CardEditorCallbacks{})
	e.SetSize(50, 20) // Non-default size

	workspaces := []config.Workspace{
		{ID: "ws-1", Name: "Alpha Project", Path: "/path/alpha"},
		{ID: "ws-2", Name: "Beta Project", Path: "/path/beta"},
	}
	e.SetData(messages.DataLoadedMsg{Workspaces: workspaces})

	// Open picker
	e.CycleFocus(3)
	e.Update(tea.KeyPressMsg{Code: tea.KeySpace, Text: " "})

	// Get the view from the first open
	view1 := e.View()

	// Close and reopen
	e.Update(tea.KeyPressMsg{Code: tea.KeyEsc})
	e.Update(tea.KeyPressMsg{Code: tea.KeySpace, Text: " "})

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
	e := views.NewCardEditor(views.CardEditorCallbacks{})
	e.SetSize(20, 10) // Very small terminal

	workspaces := []config.Workspace{
		{ID: "ws-1", Name: "Alpha", Path: "/path/alpha"},
		{ID: "ws-2", Name: "Beta", Path: "/path/beta"},
	}
	e.SetData(messages.DataLoadedMsg{Workspaces: workspaces})

	// Open picker on small terminal
	e.CycleFocus(3)
	e.Update(tea.KeyPressMsg{Code: tea.KeySpace, Text: " "})

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
	e := views.NewCardEditor(views.CardEditorCallbacks{})
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
	e.Update(tea.KeyPressMsg{Code: tea.KeySpace, Text: " "})

	// Navigate to Gamma and select it
	e.Update(tea.KeyPressMsg{Code: tea.KeyDown})  // Move to Beta
	e.Update(tea.KeyPressMsg{Code: tea.KeyDown})  // Move to Gamma
	e.Update(tea.KeyPressMsg{Code: tea.KeyEnter}) // Select Gamma

	// Verify that Gamma was actually selected, not the original Alpha
	if e.WorkspaceID() != "ws-3" {
		t.Errorf("Expected Gamma (ws-3) to be selected, but got %s", e.WorkspaceID())
	}

	// Reopen picker and verify Gamma is still selected
	e.Update(tea.KeyPressMsg{Code: tea.KeySpace, Text: " "})
	view := e.View()
	if !strings.Contains(view, "Gamma Project") {
		t.Error("Gamma should still be selected after reopening")
	}
}

// TestWorkspacePicker_Filtering tests that workspace filtering works correctly.
func TestWorkspacePicker_Filtering(t *testing.T) {
	e := views.NewCardEditor(views.CardEditorCallbacks{})
	e.SetSize(100, 50)

	workspaces := []config.Workspace{
		{ID: "ws-1", Name: "Alpha Project", Path: "/path/alpha"},
		{ID: "ws-2", Name: "Beta Project", Path: "/path/beta"},
		{ID: "ws-3", Name: "Gamma Project", Path: "/path/gamma"},
	}
	e.SetData(messages.DataLoadedMsg{Workspaces: workspaces})

	// Open picker
	e.CycleFocus(3)
	e.Update(tea.KeyPressMsg{Code: tea.KeySpace, Text: " "})

	// All workspaces should be visible initially
	view := e.View()
	if !strings.Contains(view, "Alpha Project") || !strings.Contains(view, "Beta Project") || !strings.Contains(view, "Gamma Project") {
		t.Error("All workspaces should be visible initially")
	}

	// Search for 'lph' to filter to Alpha only (unique part of Alpha)
	e.Update(tea.KeyPressMsg{Code: 0, Text: "l"})
	e.Update(tea.KeyPressMsg{Code: 0, Text: "p"})
	e.Update(tea.KeyPressMsg{Code: 0, Text: "h"})

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
	e.Update(tea.KeyPressMsg{Code: tea.KeyEsc})
	view = e.View()

	// All workspaces should be visible again
	if !strings.Contains(view, "Alpha Project") || !strings.Contains(view, "Beta Project") || !strings.Contains(view, "Gamma Project") {
		t.Error("All workspaces should be visible after clearing search")
	}
}
