package views_test

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"

	"kankuro/internal/config"
	"kankuro/internal/ui/commands"
	"kankuro/internal/ui/components"
	"kankuro/internal/ui/messages"
	"kankuro/internal/ui/views"
)

func TestCardEditor_New(t *testing.T) {
	callbacks := views.CardEditorCallbacks{
		OnSave: func(config.IDCard) tea.Cmd {
			return nil
		},
	}

	e := views.NewCardEditor(callbacks, &commands.Commands{})

	if e == nil {
		t.Fatal("NewCardEditor returned nil")
	}

	if e.FocusIndex() != 0 {
		t.Error("CardEditor should start with focus on first field")
	}
}

func TestCardEditor_SetCard(t *testing.T) {
	e := views.NewCardEditor(views.CardEditorCallbacks{}, &commands.Commands{})

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
	e := views.NewCardEditor(views.CardEditorCallbacks{}, &commands.Commands{})

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
	e := views.NewCardEditor(views.CardEditorCallbacks{}, &commands.Commands{})

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

func TestCardEditor_CycleFocus(t *testing.T) {
	e := views.NewCardEditor(views.CardEditorCallbacks{}, &commands.Commands{})

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
	e := views.NewCardEditor(views.CardEditorCallbacks{}, &commands.Commands{})

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
	e := views.NewCardEditor(views.CardEditorCallbacks{}, &commands.Commands{})
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

	if !list.IsSelectedList()["s1"] {
		t.Error("s1 should be selected after toggle")
	}

	// Toggle again
	list.ToggleSelected()
	if list.IsSelectedList()["s1"] {
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
	e := views.NewCardEditor(views.CardEditorCallbacks{}, &commands.Commands{})
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
	e := views.NewCardEditor(views.CardEditorCallbacks{}, &commands.Commands{})
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
	e := views.NewCardEditor(views.CardEditorCallbacks{}, &commands.Commands{})
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
