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
