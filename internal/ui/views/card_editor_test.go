package views_test

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/BlackGiack/kankuro/internal/config"
	"github.com/BlackGiack/kankuro/internal/ui/components"
	"github.com/BlackGiack/kankuro/internal/ui/messages"
	"github.com/BlackGiack/kankuro/internal/ui/views"
)

func TestCardEditor_New(t *testing.T) {
	callbacks := views.CardEditorCallbacks{
		OnSave: func(config.IDCard) tea.Cmd {
			return nil
		},
	}

	e := views.NewCardEditor(callbacks)

	if e == nil {
		t.Fatal("NewCardEditor returned nil")
	}

	if e.FocusIndex() != 0 {
		t.Error("CardEditor should start with focus on first field")
	}
}

func TestCardEditor_SetCard(t *testing.T) {
	e := views.NewCardEditor(views.CardEditorCallbacks{})

	card := config.IDCard{
		ID:          "test-1",
		Name:        "Test Card",
		Driver:      "claude",
		Model:       "claude-3-opus",
		WorkspaceID: "ws-1",
		Puppet:      "cat",
		SkillIDs:    []string{"s1", "s2"},
	}

	e.SetCard(card)

	if e.EditingID() != "test-1" {
		t.Errorf("Expected EditingID() 'test-1', got '%s'", e.EditingID())
	}
	if e.Inputs()[0].Value() != "Test Card" {
		t.Errorf("Expected name 'Test Card', got '%s'", e.Inputs()[0].Value())
	}
	if e.Puppet() != "cat" {
		t.Errorf("Expected puppet 'cat', got '%s'", e.Puppet())
	}
}

func TestCardEditor_Reset(t *testing.T) {
	e := views.NewCardEditor(views.CardEditorCallbacks{})

	e.SetCard(config.IDCard{
		ID:     "test-1",
		Name:   "Test Card",
		Puppet: "cat",
	})

	e.Reset()

	if e.EditingID() != "" {
		t.Error("EditingID() should be empty after reset")
	}
	if e.Inputs()[0].Value() != "" {
		t.Error("Name input should be empty after reset")
	}
	if e.Puppet() != "kankuro" {
		t.Errorf("Puppet() should reset to default, got '%s'", e.Puppet())
	}
}

func TestCardEditor_Card(t *testing.T) {
	e := views.NewCardEditor(views.CardEditorCallbacks{})

	e.Inputs()[0].SetValue("My Card")
	e.Inputs()[1].SetValue("claude")
	e.Inputs()[2].SetValue("claude-3-opus")
	e.SetWorkspaceID("ws-1")
	e.SetPuppet("darth")

	// Set up skills data
	skills := []config.Skill{
		{ID: "s1", Name: "Skill 1"},
	}
	e.SetData(messages.DataLoadedMsg{
		Skills: skills,
	})
	e.SkillList().SetSelected([]string{"s1"})

	card := e.Card()

	if card.Name != "My Card" {
		t.Errorf("Expected name 'My Card', got '%s'", card.Name)
	}
	if card.Driver != "claude" {
		t.Errorf("Expected driver 'claude', got '%s'", card.Driver)
	}
	if card.Puppet != "darth" {
		t.Errorf("Expected Puppet() 'darth', got '%s'", card.Puppet)
	}
	if len(card.SkillIDs) != 1 {
		t.Errorf("Expected 1 skill, got %d", len(card.SkillIDs))
	}
}

// TestCardEditorPreservesUnmanagedFields fails when a new IDCard field is
// added that the editor neither edits nor preserves (finding #1: Card() used
// to rebuild config.IDCard{} from scratch, silently dropping
// SystemInstruction and Env on every save).
func TestCardEditorPreservesUnmanagedFields(t *testing.T) {
	original := config.IDCard{
		ID:                 "id-1",
		Name:               "orig-name",
		Driver:             "orig-driver",
		Model:              "orig-model",
		WorkspaceID:        "ws-1",
		SubWorkspaceIDs:    []string{"ws-2"},
		UseTmpWorkspace:    true,
		DashboardControl:   true,
		SystemInstruction:  "be a test",
		SkillIDs:           []string{"s1"},
		PersonaIDs:         []string{"p1"},
		MCPIDs:             []string{"m1"},
		Env:                map[string]string{"FOO": "bar"},
		FaceStyle:          "orig-face",
		Puppet:             "orig-puppet",
		AgentsMD:           "orig agents md",
		AgentsMDPath:       "/orig/path",
		AgentsMDPreferPath: true,
		CreatedAt:          time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
		UpdatedAt:          time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC),
	}

	// Every field must be non-zero, or this test would vacuously pass for a
	// field mistakenly left at its zero value.
	v := reflect.ValueOf(original)
	typ := v.Type()
	for i := range typ.NumField() {
		if v.Field(i).IsZero() {
			t.Fatalf("test setup bug: field %s is zero-valued", typ.Field(i).Name)
		}
	}

	e := views.NewCardEditor(views.CardEditorCallbacks{})
	// CapabilityList.SelectedIDs() only returns IDs present in its current
	// items (see finding #17), so the lists must know about s1/p1/m1 before
	// SetCard's selection can round-trip through Card().
	e.SetData(messages.DataLoadedMsg{
		Skills:   []config.Skill{{ID: "s1", Name: "Skill 1"}},
		Personas: []config.Persona{{ID: "p1", Name: "Persona 1"}},
		MCPs:     []config.MCPServer{{ID: "m1", Name: "MCP 1"}},
	})
	e.SetCard(original)

	// Mutate exactly one field the form manages.
	e.Inputs()[0].SetValue("changed-name")

	got := e.Card()

	want := original
	want.Name = "changed-name"

	if !reflect.DeepEqual(got, want) {
		t.Errorf("Card() after editing only Name changed other fields too, or "+
			"dropped one it should have preserved:\ngot:  %+v\nwant: %+v", got, want)
	}
}

func TestDashboardControlCardEditorToggleRoundTrip(t *testing.T) {
	editor := views.NewCardEditor(views.CardEditorCallbacks{})
	card := config.IDCard{ID: "card", Name: "control", DashboardControl: true}
	editor.SetCard(card)
	for editor.FocusIndex() != views.FocusDashboardControl {
		editor.CycleFocus(1)
	}
	editor.Update(tea.KeyPressMsg{Code: tea.KeySpace, Text: " "})
	if editor.Card().DashboardControl {
		t.Fatal("Space did not turn Dashboard Control off")
	}
	editor.Update(tea.KeyPressMsg{Code: tea.KeySpace, Text: " "})
	if !editor.Card().DashboardControl {
		t.Fatal("Dashboard Control did not round-trip through the card")
	}
}

func TestSelectedOnlySkillCursorNormalizesAfterRegistryRefresh(t *testing.T) {
	e := views.NewCardEditor(views.CardEditorCallbacks{})
	e.SetData(messages.DataLoadedMsg{Skills: []config.Skill{
		{ID: "a", Name: "A"}, {ID: "b", Name: "B"}, {ID: "c", Name: "C"},
	}})
	e.SkillList().SetSelected([]string{"a", "b"})
	e.SkillList().SetSelectedOnly(true)
	e.SkillList().SetCursor(1)
	e.SetData(messages.DataLoadedMsg{Skills: []config.Skill{{ID: "a", Name: "A"}, {ID: "c", Name: "C"}}})
	item, ok := e.SkillList().SelectedItem()
	if !ok || item.GetID() != "a" {
		t.Fatalf("selected-only cursor item = %v, %v; want A", item, ok)
	}
	e.CycleFocus(views.FocusSkills)
	e.Update(tea.KeyPressMsg{Code: tea.KeySpace, Text: " "})
	if e.SkillList().IsSelected("a") {
		t.Fatal("Space did not toggle the normalized visible skill")
	}
}

func TestCtrlXClearsOnlySubWorkspaceSelection(t *testing.T) {
	original := config.IDCard{
		ID: "clear-test", Name: "Card", Driver: "claude", Model: "model",
		WorkspaceID: "main", SubWorkspaceIDs: []string{"sub", "removed"},
		UseTmpWorkspace: true, SkillIDs: []string{"skill"}, PersonaIDs: []string{"persona"},
		MCPIDs: []string{"mcp"}, Puppet: "cat", FaceStyle: "face", AgentsMD: "instruction",
	}
	var saves, cancels int
	e := views.NewCardEditor(views.CardEditorCallbacks{
		OnSave:   func(config.IDCard) tea.Cmd { saves++; return nil },
		OnCancel: func() tea.Cmd { cancels++; return nil },
	})
	e.SetData(messages.DataLoadedMsg{
		Workspaces: []config.Workspace{{ID: "main", Name: "Main"}, {ID: "sub", Name: "Sub"}},
		Skills:     []config.Skill{{ID: "skill", Name: "Skill"}},
		Personas:   []config.Persona{{ID: "persona", Name: "Persona"}},
		MCPs:       []config.MCPServer{{ID: "mcp", Name: "MCP"}},
	})
	e.SetCard(original)
	e.CycleFocus(views.FocusSubWorkspaces)
	focus := e.FocusIndex()
	ctrlX := tea.KeyPressMsg{Code: 'x', Mod: tea.ModCtrl}
	e.Update(ctrlX)
	if got := e.Card(); len(got.SubWorkspaceIDs) != 0 {
		t.Fatalf("Ctrl+X left sub workspace IDs: %v", got.SubWorkspaceIDs)
	}
	if e.FocusIndex() != focus {
		t.Fatalf("Ctrl+X changed focus from %d to %d", focus, e.FocusIndex())
	}
	want := original
	want.SubWorkspaceIDs = nil
	if got := e.Card(); !reflect.DeepEqual(got, want) {
		t.Fatalf("Ctrl+X changed unrelated card fields:\ngot:  %+v\nwant: %+v", got, want)
	}
	if saves != 0 || cancels != 0 {
		t.Fatalf("Ctrl+X invoked callbacks: saves=%d cancels=%d", saves, cancels)
	}
	if view := e.View(); !strings.Contains(view, "Sub Workspaces: none") || !strings.Contains(view, "ctrl+x=clear") {
		t.Fatalf("cleared focused field is not rendered with its shortcut:\n%s", view)
	}
	e.Update(ctrlX) // clearing an empty selection is harmless
	if saves != 0 || cancels != 0 {
		t.Fatalf("empty Ctrl+X invoked callbacks: saves=%d cancels=%d", saves, cancels)
	}

	e.SetCard(original)
	e.CycleFocus(views.FocusSubWorkspaces)
	e.Update(tea.KeyPressMsg{Code: 'x', Text: "x"})
	if got := e.Card().SubWorkspaceIDs; !reflect.DeepEqual(got, original.SubWorkspaceIDs) {
		t.Fatalf("plain x cleared sub workspaces: %v", got)
	}
}

func TestCtrlXOutsideSubWorkspaceFieldPreservesSelection(t *testing.T) {
	for _, focus := range []int{
		views.FocusName, views.FocusDriver, views.FocusModel, views.FocusWorkspace,
		views.FocusFaceStyle, views.FocusPuppet, views.FocusUseTmp, views.FocusAgentsMD,
		views.FocusSkills, views.FocusPersonas, views.FocusMCPs,
	} {
		t.Run(fmt.Sprintf("focus-%d", focus), func(t *testing.T) {
			e := views.NewCardEditor(views.CardEditorCallbacks{})
			e.SetData(messages.DataLoadedMsg{Skills: []config.Skill{{ID: "skill", Name: "Skill"}}})
			e.SetCard(config.IDCard{SubWorkspaceIDs: []string{"sub", "unresolved"}, SkillIDs: []string{"skill"}})
			e.CycleFocus(focus)
			e.Update(tea.KeyPressMsg{Code: 'x', Mod: tea.ModCtrl})
			if got := e.Card().SubWorkspaceIDs; !reflect.DeepEqual(got, []string{"sub", "unresolved"}) {
				t.Fatalf("Ctrl+X at focus %d changed sub workspace IDs: %v", focus, got)
			}
		})
	}
}

func TestCtrlXPickerSaveAndCancelFlow(t *testing.T) {
	workspaces := []config.Workspace{
		{ID: "main", Name: "Main"}, {ID: "sub-a", Name: "Sub A"}, {ID: "sub-b", Name: "Sub B"},
	}
	original := config.IDCard{ID: "picker-clear", Name: "Card", WorkspaceID: "main", SubWorkspaceIDs: []string{"sub-a", "missing"}}
	var saved config.IDCard
	saves, cancels := 0, 0
	e := views.NewCardEditor(views.CardEditorCallbacks{
		OnSave:   func(card config.IDCard) tea.Cmd { saved = card; saves++; return nil },
		OnCancel: func() tea.Cmd { cancels++; return nil },
	})
	e.SetData(messages.DataLoadedMsg{Workspaces: workspaces})
	e.SetCard(original)
	e.CycleFocus(views.FocusSubWorkspaces)
	e.Update(tea.KeyPressMsg{Code: 'x', Mod: tea.ModCtrl})
	e.Update(tea.KeyPressMsg{Code: tea.KeySpace, Text: " "})
	if view := e.View(); strings.Contains(view, "[x]") {
		t.Fatalf("reopened picker retained cleared selections:\n%s", view)
	}
	e.Update(tea.KeyPressMsg{Code: tea.KeyEnter}) // confirm empty selection
	if got := e.Card().SubWorkspaceIDs; len(got) != 0 {
		t.Fatalf("empty picker confirmation restored selection: %v", got)
	}

	// A provisional picker selection is unaffected by Ctrl+X while the modal owns input.
	e.Update(tea.KeyPressMsg{Code: tea.KeySpace, Text: " "})
	e.Update(tea.KeyPressMsg{Code: tea.KeySpace, Text: " "}) // check Sub A
	e.Update(tea.KeyPressMsg{Code: 'x', Mod: tea.ModCtrl})
	if got := e.Card().SubWorkspaceIDs; len(got) != 0 {
		t.Fatalf("picker Ctrl+X changed the underlying form: %v", got)
	}
	if view := e.View(); !strings.Contains(view, "[x] Sub A") {
		t.Fatalf("picker Ctrl+X changed provisional selection:\n%s", view)
	}
	e.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if got := e.Card().SubWorkspaceIDs; !reflect.DeepEqual(got, []string{"sub-a"}) {
		t.Fatalf("ordinary picker confirmation selected %v, want [sub-a]", got)
	}

	e.Update(tea.KeyPressMsg{Code: 'x', Mod: tea.ModCtrl})
	if saves != 0 || cancels != 0 {
		t.Fatalf("clear invoked callbacks before save: saves=%d cancels=%d", saves, cancels)
	}
	e.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if saves != 1 || len(saved.SubWorkspaceIDs) != 0 {
		t.Fatalf("Enter save = %d, IDs=%v; want one save with no sub workspaces", saves, saved.SubWorkspaceIDs)
	}

	cancelEditor := views.NewCardEditor(views.CardEditorCallbacks{
		OnSave:   func(config.IDCard) tea.Cmd { t.Error("cancel path called OnSave"); return nil },
		OnCancel: func() tea.Cmd { cancels++; return nil },
	})
	cancelEditor.SetData(messages.DataLoadedMsg{Workspaces: workspaces})
	cancelEditor.SetCard(original)
	cancelEditor.CycleFocus(views.FocusSubWorkspaces)
	cancelEditor.Update(tea.KeyPressMsg{Code: 'x', Mod: tea.ModCtrl})
	cancelEditor.Update(tea.KeyPressMsg{Code: tea.KeyEsc})
	if cancels != 1 {
		t.Fatalf("Esc cancel count = %d, want 1", cancels)
	}
	cancelEditor.SetCard(original)
	if got := cancelEditor.Card().SubWorkspaceIDs; !reflect.DeepEqual(got, original.SubWorkspaceIDs) {
		t.Fatalf("reloading original card after cancel got %v, want %v", got, original.SubWorkspaceIDs)
	}
}

func TestCardEditor_CycleFocus(t *testing.T) {
	e := views.NewCardEditor(views.CardEditorCallbacks{})

	if e.FocusIndex() != 0 {
		t.Error("Should start on first field")
	}

	e.CycleFocus(1)
	if e.FocusIndex() != 1 {
		t.Errorf("Expected focus 1, got %d", e.FocusIndex())
	}

	// Cycle to end
	for range views.FocusCount - 1 {
		e.CycleFocus(1)
	}
	if e.FocusIndex() != 0 {
		t.Errorf("Should wrap to 0, got %d", e.FocusIndex())
	}

	// Cycle backwards
	e.CycleFocus(-1)
	if e.FocusIndex() != views.FocusCount-1 {
		t.Errorf("Should wrap to last, got %d", e.FocusIndex())
	}
}

func TestCardEditor_SetData(t *testing.T) {
	e := views.NewCardEditor(views.CardEditorCallbacks{})

	workspaces := []config.Workspace{
		{ID: "ws-1", Name: "Workspace 1"},
	}
	skills := []config.Skill{
		{ID: "s1", Name: "Skill 1"},
	}
	personas := []config.Persona{
		{ID: "p1", Name: "Persona 1"},
	}
	mcps := []config.MCPServer{
		{ID: "m1", Name: "MCP 1"},
	}

	e.SetData(messages.DataLoadedMsg{
		Workspaces: workspaces,
		Skills:     skills,
		Personas:   personas,
		MCPs:       mcps,
	})

	if len(e.Workspaces()) != 1 {
		t.Errorf("Expected 1 workspace, got %d", len(e.Workspaces()))
	}
	if len(e.SkillList().Items()) != 1 {
		t.Errorf("Expected 1 skill item, got %d", len(e.SkillList().Items()))
	}
}

func TestCardEditor_WorkspaceSummary(t *testing.T) {
	e := views.NewCardEditor(views.CardEditorCallbacks{})
	e.SetData(messages.DataLoadedMsg{
		Workspaces: []config.Workspace{
			{ID: "ws-1", Name: "My Workspace"},
		},
	})

	// No selection
	if e.WorkspaceSummary() != "(none selected)" {
		t.Errorf("Expected '(none selected)', got '%s'", e.WorkspaceSummary())
	}

	// With selection
	e.SetWorkspaceID("ws-1")
	if e.WorkspaceSummary() != "My Workspace" {
		t.Errorf("Expected 'My Workspace', got '%s'", e.WorkspaceSummary())
	}
}

func TestCapabilityList_Toggle(t *testing.T) {
	list := components.NewCapabilityList(components.CapabilityListConfig{Title: "Test"})
	list.SetItems([]components.CapabilityItem{
		views.SkillItem{Skill: config.Skill{ID: "s1", Name: "Skill 1"}},
		views.SkillItem{Skill: config.Skill{ID: "s2", Name: "Skill 2"}},
	})

	// Toggle first item
	list.SetCursor(0)
	list.ToggleSelected()

	if !list.IsSelected("s1") {
		t.Error("s1 should be selected after toggle")
	}

	// Toggle again
	list.ToggleSelected()
	if list.IsSelected("s1") {
		t.Error("s1 should be deselected after second toggle")
	}
}

func TestCapabilityList_Search(t *testing.T) {
	list := components.NewCapabilityList(components.CapabilityListConfig{Title: "Test"})
	list.SetItems([]components.CapabilityItem{
		views.SkillItem{Skill: config.Skill{ID: "s1", Name: "Alpha"}},
		views.SkillItem{Skill: config.Skill{ID: "s2", Name: "Beta"}},
		views.SkillItem{Skill: config.Skill{ID: "s3", Name: "Gamma"}},
	})

	filtered := list.FilteredItems()
	if len(filtered) != 3 {
		t.Errorf("Expected 3 items without filter, got %d", len(filtered))
	}

	list.SetSearchQuery("be")
	filtered = list.FilteredItems()
	if len(filtered) != 1 {
		t.Errorf("Expected 1 item matching 'be', got %d", len(filtered))
	}
	if filtered[0].GetID() != "s2" {
		t.Errorf("Expected 's2', got '%s'", filtered[0].GetID())
	}
}

func TestCardEditor_View(t *testing.T) {
	e := views.NewCardEditor(views.CardEditorCallbacks{})
	e.SetSize(80, 40)

	view := e.View()

	if view == "" {
		t.Error("View should not be empty")
	}
	if !strings.Contains(view, "New ID Card") {
		t.Error("View should contain title")
	}
	if !strings.Contains(view, "Name:") {
		t.Error("View should contain Name field")
	}
	if !strings.Contains(view, "Skills") {
		t.Error("View should contain Skills section")
	}
}

func TestCardEditor_View_HidesOptionalSectionsFromSettings(t *testing.T) {
	e := views.NewCardEditor(views.CardEditorCallbacks{})
	e.SetSettings(config.AppSettings{
		ShowPersonasInCardEditor: false,
		ShowMCPsInCardEditor:     false,
	})
	e.SetSize(80, 40)

	view := e.View()
	if strings.Contains(view, "Personas") {
		t.Error("view should hide Personas section when disabled")
	}
	if strings.Contains(view, "MCPs") {
		t.Error("view should hide MCPs section when disabled")
	}
}

func TestCardEditor_HiddenSectionsPreserveSelections(t *testing.T) {
	e := views.NewCardEditor(views.CardEditorCallbacks{})
	e.SetCard(config.IDCard{
		ID:         "card-1",
		PersonaIDs: []string{"p1"},
		MCPIDs:     []string{"m1"},
	})
	e.SetSettings(config.AppSettings{
		ShowPersonasInCardEditor: false,
		ShowMCPsInCardEditor:     false,
	})

	card := e.Card()
	if len(card.PersonaIDs) != 1 || card.PersonaIDs[0] != "p1" {
		t.Fatalf("expected hidden persona selection to be preserved, got %+v", card.PersonaIDs)
	}
	if len(card.MCPIDs) != 1 || card.MCPIDs[0] != "m1" {
		t.Fatalf("expected hidden MCP selection to be preserved, got %+v", card.MCPIDs)
	}
}

// TestCardEditor_HiddenSectionsCaptureDeselection guards the SetSettings
// capture-on-hide guard: when the list's items are actually loaded, an
// explicit deselect-everything before hiding the section must stick (not be
// mistaken for "list not populated yet", which is what the guard exists to
// protect).
func TestCardEditor_HiddenSectionsCaptureDeselection(t *testing.T) {
	e := views.NewCardEditor(views.CardEditorCallbacks{})
	e.SetData(messages.DataLoadedMsg{
		Personas: []config.Persona{{ID: "p1", Name: "Persona 1"}},
		MCPs:     []config.MCPServer{{ID: "m1", Name: "MCP 1"}},
	})
	e.SetCard(config.IDCard{
		ID:         "card-1",
		PersonaIDs: []string{"p1"},
		MCPIDs:     []string{"m1"},
	})

	e.PersonaList().SetSelected(nil)
	e.MCPList().SetSelected(nil)

	e.SetSettings(config.AppSettings{
		ShowPersonasInCardEditor: false,
		ShowMCPsInCardEditor:     false,
	})

	card := e.Card()
	if len(card.PersonaIDs) != 0 {
		t.Fatalf("expected deselected persona to stay deselected after hiding, got %+v", card.PersonaIDs)
	}
	if len(card.MCPIDs) != 0 {
		t.Fatalf("expected deselected MCP to stay deselected after hiding, got %+v", card.MCPIDs)
	}
}

// TestCardEditor_AgentsMDRow_MultilineSummary guards the renderAgentsMDRow
// delimiter bug: it used to split on the literal two-character string
// `\n` (a backtick raw string) instead of a real newline, so the
// multi-line summary/ellipsis branch never fired.
func TestCardEditor_AgentsMDRow_MultilineSummary(t *testing.T) {
	e := views.NewCardEditor(views.CardEditorCallbacks{})
	e.SetCard(config.IDCard{ID: "c1", AgentsMD: "line one\nline two"})
	e.SetSize(80, 40)

	view := e.View()
	if !strings.Contains(view, "line one") {
		t.Error("view should contain the first line of AGENTS.md content")
	}
	if !strings.Contains(view, "line one...") {
		t.Error("multi-line AGENTS.md content should render with an ellipsis")
	}
	if strings.Contains(view, "line two") {
		t.Error("view should not contain the second line of AGENTS.md content")
	}
}

func TestPuppetSelector_Navigation(t *testing.T) {
	m := views.NewPuppetSelectorModal()
	m.Open()

	if m.Cursor() != 0 {
		t.Error("Should start at cursor 0")
	}

	// Right arrow
	m.Update(tea.KeyPressMsg{Text: "", Code: tea.KeyRight})
	if m.Cursor() != 1 {
		t.Errorf("Expected cursor 1, got %d", m.Cursor())
	}
}
