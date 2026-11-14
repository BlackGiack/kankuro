package views_test

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/BlackGiack/kankuro/internal/config"
	"github.com/BlackGiack/kankuro/internal/ui/components"
	"github.com/BlackGiack/kankuro/internal/ui/messages"
	"github.com/BlackGiack/kankuro/internal/ui/models"
	"github.com/BlackGiack/kankuro/internal/ui/views"
)

func TestCardEditor_New(t *testing.T) {
	callbacks := views.CardEditorCallbacks{
		OnSave: func(models.CardEditorDraft) tea.Cmd {
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

	loadEditor(e, card)

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

	loadEditor(e, config.IDCard{
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
	setEditorData(e, messages.DataLoadedMsg{
		Skills: skills,
	})
	e.SkillList().SetSelected([]string{"s1"})

	card := editorCard(e)

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

func TestDashboardControlCardEditorToggleRoundTrip(t *testing.T) {
	editor := views.NewCardEditor(views.CardEditorCallbacks{})
	card := config.IDCard{ID: "card", Name: "control", DashboardControl: true}
	loadEditor(editor, card)
	for editor.FocusIndex() != views.FocusDashboardControl {
		editor.CycleFocus(1)
	}
	editor.Update(tea.KeyPressMsg{Code: tea.KeySpace, Text: " "})
	if editorCard(editor).DashboardControl {
		t.Fatal("Space did not turn Dashboard Control off")
	}
	editor.Update(tea.KeyPressMsg{Code: tea.KeySpace, Text: " "})
	if !editorCard(editor).DashboardControl {
		t.Fatal("Dashboard Control did not round-trip through the card")
	}
}

func TestSelectedOnlySkillCursorNormalizesAfterRegistryRefresh(t *testing.T) {
	e := views.NewCardEditor(views.CardEditorCallbacks{})
	setEditorData(e, messages.DataLoadedMsg{Skills: []config.Skill{
		{ID: "a", Name: "A"}, {ID: "b", Name: "B"}, {ID: "c", Name: "C"},
	}})
	e.SkillList().SetSelected([]string{"a", "b"})
	e.SkillList().SetSelectedOnly(true)
	e.SkillList().SetCursor(1)
	setEditorData(e, messages.DataLoadedMsg{Skills: []config.Skill{{ID: "a", Name: "A"}, {ID: "c", Name: "C"}}})
	item, ok := e.SkillList().SelectedItem()
	if !ok || item.GetID() != "a" {
		t.Fatalf("selected-only cursor item = %v, %v; want A", item, ok)
	}
	for e.FocusIndex() != views.FocusSkills {
		e.CycleFocus(1)
	}
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
		OnSave:   func(models.CardEditorDraft) tea.Cmd { saves++; return nil },
		OnCancel: func() tea.Cmd { cancels++; return nil },
	})
	setEditorData(e, messages.DataLoadedMsg{
		Workspaces: []config.Workspace{{ID: "main", Name: "Main"}, {ID: "sub", Name: "Sub"}},
		Skills:     []config.Skill{{ID: "skill", Name: "Skill"}},
		Personas:   []config.Persona{{ID: "persona", Name: "Persona"}},
		MCPs:       []config.MCPServer{{ID: "mcp", Name: "MCP"}},
	})
	loadEditor(e, original)
	e.CycleFocus(views.FocusSubWorkspaces)
	focus := e.FocusIndex()
	ctrlX := tea.KeyPressMsg{Code: 'x', Mod: tea.ModCtrl}
	e.Update(ctrlX)
	if got := editorCard(e); len(got.SubWorkspaceIDs) != 0 {
		t.Fatalf("Ctrl+X left sub workspace IDs: %v", got.SubWorkspaceIDs)
	}
	if e.FocusIndex() != focus {
		t.Fatalf("Ctrl+X changed focus from %d to %d", focus, e.FocusIndex())
	}
	want := original
	want.SubWorkspaceIDs = []string{}
	if got := editorCard(e); !reflect.DeepEqual(got, want) {
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

	loadEditor(e, original)
	e.CycleFocus(views.FocusSubWorkspaces)
	e.Update(tea.KeyPressMsg{Code: 'x', Text: "x"})
	if got := editorCard(e).SubWorkspaceIDs; !reflect.DeepEqual(got, original.SubWorkspaceIDs) {
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
			setEditorData(e, messages.DataLoadedMsg{Skills: []config.Skill{{ID: "skill", Name: "Skill"}}})
			loadEditor(e, config.IDCard{SubWorkspaceIDs: []string{"sub", "unresolved"}, SkillIDs: []string{"skill"}})
			e.CycleFocus(focus)
			e.Update(tea.KeyPressMsg{Code: 'x', Mod: tea.ModCtrl})
			if got := editorCard(e).SubWorkspaceIDs; !reflect.DeepEqual(got, []string{"sub", "unresolved"}) {
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
		OnSave:   func(draft models.CardEditorDraft) tea.Cmd { saved = valuesCard(draft.Values); saves++; return nil },
		OnCancel: func() tea.Cmd { cancels++; return nil },
	})
	setEditorData(e, messages.DataLoadedMsg{Workspaces: workspaces})
	loadEditor(e, original)
	e.CycleFocus(views.FocusSubWorkspaces)
	e.Update(tea.KeyPressMsg{Code: 'x', Mod: tea.ModCtrl})
	e.Update(tea.KeyPressMsg{Code: tea.KeySpace, Text: " "})
	if view := e.View(); strings.Contains(view, "[x]") {
		t.Fatalf("reopened picker retained cleared selections:\n%s", view)
	}
	e.Update(tea.KeyPressMsg{Code: tea.KeyEnter}) // confirm empty selection
	if got := editorCard(e).SubWorkspaceIDs; len(got) != 0 {
		t.Fatalf("empty picker confirmation restored selection: %v", got)
	}

	// A provisional picker selection is unaffected by Ctrl+X while the modal owns input.
	e.Update(tea.KeyPressMsg{Code: tea.KeySpace, Text: " "})
	e.Update(tea.KeyPressMsg{Code: tea.KeySpace, Text: " "}) // check Sub A
	e.Update(tea.KeyPressMsg{Code: 'x', Mod: tea.ModCtrl})
	if got := editorCard(e).SubWorkspaceIDs; len(got) != 0 {
		t.Fatalf("picker Ctrl+X changed the underlying form: %v", got)
	}
	if view := e.View(); !strings.Contains(view, "[x] Sub A") {
		t.Fatalf("picker Ctrl+X changed provisional selection:\n%s", view)
	}
	e.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if got := editorCard(e).SubWorkspaceIDs; !reflect.DeepEqual(got, []string{"sub-a"}) {
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
		OnSave:   func(models.CardEditorDraft) tea.Cmd { t.Error("cancel path called OnSave"); return nil },
		OnCancel: func() tea.Cmd { cancels++; return nil },
	})
	setEditorData(cancelEditor, messages.DataLoadedMsg{Workspaces: workspaces})
	loadEditor(cancelEditor, original)
	cancelEditor.CycleFocus(views.FocusSubWorkspaces)
	cancelEditor.Update(tea.KeyPressMsg{Code: 'x', Mod: tea.ModCtrl})
	cancelEditor.Update(tea.KeyPressMsg{Code: tea.KeyEsc})
	if cancels != 1 {
		t.Fatalf("Esc cancel count = %d, want 1", cancels)
	}
	loadEditor(cancelEditor, original)
	if got := editorCard(cancelEditor).SubWorkspaceIDs; !reflect.DeepEqual(got, original.SubWorkspaceIDs) {
		t.Fatalf("reloading original card after cancel got %v, want %v", got, original.SubWorkspaceIDs)
	}
}

func TestCardEditor_CycleFocus(t *testing.T) {
	e := views.NewCardEditor(views.CardEditorCallbacks{})
	e.SetDashboardContext("dashboard", false)

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

func TestCardEditorSharedAndConsumerReadOnly(t *testing.T) {
	card := config.IDCard{ID: "card", Name: "Shared", DashboardID: "owner", Shared: true}
	owner := views.NewCardEditor(views.CardEditorCallbacks{})
	owner.SetSize(80, 21)
	loadEditor(owner, card)
	owner.SetDashboardContext("owner", false)
	for i := 0; owner.FocusIndex() != views.FocusShared && i < views.FocusCount; i++ {
		owner.CycleFocus(1)
	}
	if owner.FocusIndex() != views.FocusShared {
		t.Fatal("shared control is not reachable")
	}
	owner.Update(tea.KeyPressMsg{Code: tea.KeySpace})
	if editorCard(owner).Shared {
		t.Fatal("shared toggle did not change")
	}
	var saves int
	consumer := views.NewCardEditor(views.CardEditorCallbacks{OnSave: func(models.CardEditorDraft) tea.Cmd { saves++; return nil }})
	consumer.SetSize(80, 21)
	loadEditor(consumer, card)
	consumer.SetDashboardContext("consumer", true)
	consumer.Update(tea.KeyPressMsg{Text: "changed", Code: 'c'})
	consumer.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if saves != 0 || editorCard(consumer).Name != card.Name || !editorCard(consumer).Shared {
		t.Fatal("consumer changed a read-only card")
	}
	if !strings.Contains(consumer.View(), "read-only") || !strings.Contains(consumer.View(), "owner") {
		t.Fatal("read-only owner details missing")
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

	setEditorData(e, messages.DataLoadedMsg{
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
	setEditorData(e, messages.DataLoadedMsg{
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
		testCapability{ID: "s1", Name: "Skill 1"},
		testCapability{ID: "s2", Name: "Skill 2"},
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
		testCapability{ID: "s1", Name: "Alpha"},
		testCapability{ID: "s2", Name: "Beta"},
		testCapability{ID: "s3", Name: "Gamma"},
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
	setEditorSettings(e, config.AppSettings{
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
	loadEditor(e, config.IDCard{
		ID:         "card-1",
		PersonaIDs: []string{"p1"},
		MCPIDs:     []string{"m1"},
	})
	setEditorSettings(e, config.AppSettings{
		ShowPersonasInCardEditor: false,
		ShowMCPsInCardEditor:     false,
	})

	card := editorCard(e)
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
	setEditorData(e, messages.DataLoadedMsg{
		Personas: []config.Persona{{ID: "p1", Name: "Persona 1"}},
		MCPs:     []config.MCPServer{{ID: "m1", Name: "MCP 1"}},
	})
	loadEditor(e, config.IDCard{
		ID:         "card-1",
		PersonaIDs: []string{"p1"},
		MCPIDs:     []string{"m1"},
	})

	e.PersonaList().SetSelected(nil)
	e.MCPList().SetSelected(nil)

	setEditorSettings(e, config.AppSettings{
		ShowPersonasInCardEditor: false,
		ShowMCPsInCardEditor:     false,
	})

	card := editorCard(e)
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
	loadEditor(e, config.IDCard{ID: "c1", AgentsMD: "line one\nline two"})
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

func TestTemporaryWorkspaceSpaceAndSave(t *testing.T) {
	for _, want := range []bool{false, true} {
		t.Run(fmt.Sprintf("save-%t", want), func(t *testing.T) {
			original := config.IDCard{
				ID:          "card",
				Name:        "Temporary test",
				Driver:      "codex",
				WorkspaceID: "main",
				Puppet:      "kankuro",
				AgentsMD:    "instructions",
			}
			var saved config.IDCard
			saves := 0
			editor := views.NewCardEditor(views.CardEditorCallbacks{
				OnSave: func(draft models.CardEditorDraft) tea.Cmd { saved = valuesCard(draft.Values); saves++; return nil },
			})
			setEditorData(editor, messages.DataLoadedMsg{Workspaces: []config.Workspace{{ID: "main", Name: "Main"}}})
			loadEditor(editor, original)
			original = editorCard(editor)
			editor.CycleFocus(views.FocusUseTmp)
			for _, expected := range []bool{true, false} {
				editor.Update(tea.KeyPressMsg{Code: tea.KeySpace, Text: " "})
				if got := editorCard(editor).UseTmpWorkspace; got != expected {
					t.Fatalf("Space: got %t, want %t", got, expected)
				}
			}
			if want {
				editor.Update(tea.KeyPressMsg{Code: tea.KeySpace, Text: " "})
			}
			editor.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
			original.UseTmpWorkspace = want
			original.ID = "" // identity is read metadata, not part of the draft
			if saves != 1 || !reflect.DeepEqual(saved, original) {
				t.Fatalf("saves=%d saved=%+v, want %+v", saves, saved, original)
			}
		})
	}
}
