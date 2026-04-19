package views

import (
	"testing"

	bubbletea "github.com/charmbracelet/bubbletea"
	"kankuro/internal/config"
	"kankuro/internal/ui/commands"
)

func TestCardEditor_New(t *testing.T) {
	callbacks := CardEditorCallbacks{
		OnSave: func(config.IDCard) (config.IDCard, error) {
			return config.IDCard{}, nil
		},
	}

	e := NewCardEditor(callbacks, &commands.Commands{})

	if e == nil {
		t.Fatal("NewCardEditor returned nil")
	}

	if e.focusIndex != 0 {
		t.Error("CardEditor should start with focus on first field")
	}
}

func TestCardEditor_SetCard(t *testing.T) {
	e := NewCardEditor(CardEditorCallbacks{}, &commands.Commands{})

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

	if e.editingID != "test-1" {
		t.Errorf("Expected editingID 'test-1', got '%s'", e.editingID)
	}
	if e.inputs[0].Value() != "Test Card" {
		t.Errorf("Expected name 'Test Card', got '%s'", e.inputs[0].Value())
	}
	if e.puppet != "cat" {
		t.Errorf("Expected puppet 'cat', got '%s'", e.puppet)
	}
}

func TestCardEditor_Reset(t *testing.T) {
	e := NewCardEditor(CardEditorCallbacks{}, &commands.Commands{})

	e.SetCard(config.IDCard{
		ID:     "test-1",
		Name:   "Test Card",
		Puppet: "cat",
	})

	e.Reset()

	if e.editingID != "" {
		t.Error("editingID should be empty after reset")
	}
	if e.inputs[0].Value() != "" {
		t.Error("Name input should be empty after reset")
	}
	if e.puppet != "kankuro" {
		t.Errorf("Puppet should reset to default, got '%s'", e.puppet)
	}
}

func TestCardEditor_Card(t *testing.T) {
	e := NewCardEditor(CardEditorCallbacks{}, &commands.Commands{})

	e.inputs[0].SetValue("My Card")
	e.inputs[1].SetValue("claude")
	e.inputs[2].SetValue("claude-3-opus")
	e.workspaceID = "ws-1"
	e.puppet = "darth"

	// Set up skills data
	skills := []config.Skill{
		{ID: "s1", Name: "Skill 1"},
	}
	e.SetData(nil, skills, nil, nil)
	e.skillList.SetSelected([]string{"s1"})

	card := e.Card()

	if card.Name != "My Card" {
		t.Errorf("Expected name 'My Card', got '%s'", card.Name)
	}
	if card.Driver != "claude" {
		t.Errorf("Expected driver 'claude', got '%s'", card.Driver)
	}
	if card.Puppet != "darth" {
		t.Errorf("Expected puppet 'darth', got '%s'", card.Puppet)
	}
	if len(card.SkillIDs) != 1 {
		t.Errorf("Expected 1 skill, got %d", len(card.SkillIDs))
	}
}

func TestCardEditor_CycleFocus(t *testing.T) {
	e := NewCardEditor(CardEditorCallbacks{}, &commands.Commands{})

	if e.focusIndex != 0 {
		t.Error("Should start on first field")
	}

	e.cycleFocus(1)
	if e.focusIndex != 1 {
		t.Errorf("Expected focus 1, got %d", e.focusIndex)
	}

	// Cycle to end
	for i := 0; i < FocusCount-1; i++ {
		e.cycleFocus(1)
	}
	if e.focusIndex != 0 {
		t.Errorf("Should wrap to 0, got %d", e.focusIndex)
	}

	// Cycle backwards
	e.cycleFocus(-1)
	if e.focusIndex != FocusCount-1 {
		t.Errorf("Should wrap to last, got %d", e.focusIndex)
	}
}

func TestCardEditor_SetData(t *testing.T) {
	e := NewCardEditor(CardEditorCallbacks{}, &commands.Commands{})

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

	e.SetData(workspaces, skills, personas, mcps)

	if len(e.workspaces) != 1 {
		t.Errorf("Expected 1 workspace, got %d", len(e.workspaces))
	}
	if len(e.skillList.items) != 1 {
		t.Errorf("Expected 1 skill item, got %d", len(e.skillList.items))
	}
}

func TestCardEditor_WorkspaceSummary(t *testing.T) {
	e := NewCardEditor(CardEditorCallbacks{}, &commands.Commands{})
	e.SetData([]config.Workspace{
		{ID: "ws-1", Name: "My Workspace"},
	}, nil, nil, nil)

	// No selection
	if e.workspaceSummary() != "(none selected)" {
		t.Errorf("Expected '(none selected)', got '%s'", e.workspaceSummary())
	}

	// With selection
	e.workspaceID = "ws-1"
	if e.workspaceSummary() != "My Workspace" {
		t.Errorf("Expected 'My Workspace', got '%s'", e.workspaceSummary())
	}
}

func TestCapabilityList_Toggle(t *testing.T) {
	list := NewCapabilityList(CapabilityListConfig{Title: "Test"})
	list.SetItems([]CapabilityItem{
		skillItem{skill: config.Skill{ID: "s1", Name: "Skill 1"}},
		skillItem{skill: config.Skill{ID: "s2", Name: "Skill 2"}},
	})

	// Toggle first item
	list.cursor = 0
	list.ToggleSelected()

	if !list.selected["s1"] {
		t.Error("s1 should be selected after toggle")
	}

	// Toggle again
	list.ToggleSelected()
	if list.selected["s1"] {
		t.Error("s1 should be deselected after second toggle")
	}
}

func TestCapabilityList_Search(t *testing.T) {
	list := NewCapabilityList(CapabilityListConfig{Title: "Test"})
	list.SetItems([]CapabilityItem{
		skillItem{skill: config.Skill{ID: "s1", Name: "Alpha"}},
		skillItem{skill: config.Skill{ID: "s2", Name: "Beta"}},
		skillItem{skill: config.Skill{ID: "s3", Name: "Gamma"}},
	})

	filtered := list.filteredItems()
	if len(filtered) != 3 {
		t.Errorf("Expected 3 items without filter, got %d", len(filtered))
	}

	list.SetSearchQuery("be")
	filtered = list.filteredItems()
	if len(filtered) != 1 {
		t.Errorf("Expected 1 item matching 'be', got %d", len(filtered))
	}
	if filtered[0].GetID() != "s2" {
		t.Errorf("Expected 's2', got '%s'", filtered[0].GetID())
	}
}

func TestCardEditor_View(t *testing.T) {
	e := NewCardEditor(CardEditorCallbacks{}, &commands.Commands{})
	e.SetSize(80, 40)

	view := e.View()

	if view == "" {
		t.Error("View should not be empty")
	}
	if !containsString(view, "New ID Card") {
		t.Error("View should contain title")
	}
	if !containsString(view, "Name:") {
		t.Error("View should contain Name field")
	}
	if !containsString(view, "Skills") {
		t.Error("View should contain Skills section")
	}
}

func TestPuppetSelector_Navigation(t *testing.T) {
	m := NewPuppetSelectorModal()
	m.Open()

	if m.cursor != 0 {
		t.Error("Should start at cursor 0")
	}

	// Right arrow
	m.Update(bubbletea.KeyMsg{Runes: []rune{}, Type: bubbletea.KeyRight})
	if m.cursor != 1 {
		t.Errorf("Expected cursor 1, got %d", m.cursor)
	}

	// Wrap around
	for i := 0; i < len(puppetStyles)-1; i++ {
		m.Update(bubbletea.KeyMsg{Runes: []rune{}, Type: bubbletea.KeyRight})
	}
	if m.cursor != 0 {
		t.Errorf("Should wrap to 0, got %d", m.cursor)
	}
}
