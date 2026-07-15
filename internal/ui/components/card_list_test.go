package components_test

import (
	"strings"
	"testing"
	"unicode/utf8"

	tea "charm.land/bubbletea/v2"

	"kankuro/internal/config"
	"kankuro/internal/ui/components"
)

func TestCardList_New(t *testing.T) {
	callbacks := components.CardListCallbacks{
		OnSelect: func(config.IDCard) tea.Cmd {
			return nil
		},
		OnEdit: func(config.IDCard) tea.Cmd {
			return nil
		},
		OnDelete: func(string) {},
	}

	c := components.NewCardList(callbacks)
	if c == nil {
		t.Fatal("NewCardList returned nil")
	}

	// Check initial state
	if c.Focused {
		t.Error("CardList should not be focused initially")
	}
}

func TestCardList_SetCards(t *testing.T) {
	callbacks := components.CardListCallbacks{}
	c := components.NewCardList(callbacks)

	cards := []config.IDCard{
		{ID: "1", Name: "Card 1"},
		{ID: "2", Name: "Card 2"},
	}
	skills := []config.Skill{}
	personas := []config.Persona{}

	c.SetCards(cards, skills, personas)

	// Verify internal state updated
	if len(c.Cards()) != 2 {
		t.Errorf("Expected 2 cards, got %d", len(c.Cards()))
	}
}

func TestCardList_SetSearchQuery(t *testing.T) {
	callbacks := components.CardListCallbacks{}
	c := components.NewCardList(callbacks)

	cards := []config.IDCard{
		{ID: "1", Name: "Alpha"},
		{ID: "2", Name: "Beta"},
		{ID: "3", Name: "Gamma"},
	}

	c.SetCards(cards, nil, nil)

	// No filter
	filtered := c.FilteredCards()
	if len(filtered) != 3 {
		t.Errorf("Expected 3 cards without filter, got %d", len(filtered))
	}

	// With filter
	c.SetSearchQuery("al")
	filtered = c.FilteredCards()
	if len(filtered) != 1 {
		t.Errorf("Expected 1 card matching 'al', got %d", len(filtered))
	}
	if len(filtered) > 0 && filtered[0].ID != "1" {
		t.Errorf("Expected card ID '1', got '%s'", filtered[0].ID)
	}
}

func TestCardList_Selected(t *testing.T) {
	callbacks := components.CardListCallbacks{}
	c := components.NewCardList(callbacks)

	// No selection
	_, ok := c.Selected()
	if ok {
		t.Error("Selected should return false when no selection")
	}

	// Set cards and select first
	cards := []config.IDCard{
		{ID: "1", Name: "Test"},
	}
	c.SetCards(cards, nil, nil)
	c.Select(0)

	card, ok := c.Selected()
	if !ok {
		t.Error("Selected should return true when item selected")
	}
	if card.ID != "1" {
		t.Errorf("Expected card ID '1', got '%s'", card.ID)
	}
}

func TestRenderCardItem(t *testing.T) {
	card := config.IDCard{
		ID:     "1",
		Name:   "Test Card",
		Driver: "claude",
		Model:  "claude-3",
		Puppet: "kankuro",
	}

	skills := []config.Skill{
		{ID: "s1", Name: "Skill1"},
	}
	personas := []config.Persona{
		{ID: "p1", Name: "Persona1"},
	}

	// Render unselected
	output := components.RenderCardItem(card, skills, personas, nil, false, 50)
	if output == "" {
		t.Error("RenderCardItem returned empty string")
	}

	// Render selected
	outputSel := components.RenderCardItem(card, skills, personas, nil, true, 50)
	if outputSel == "" {
		t.Error("RenderCardItem (selected) returned empty string")
	}

	// Selected should be different (different border color)
	if output == outputSel {
		t.Error("Selected and unselected should render differently")
	}
}

func TestRenderCardItemWithPills(t *testing.T) {
	// Test with skills and personas
	card := config.IDCard{
		ID:         "1",
		Name:       "Test Card",
		Driver:     "claude",
		Model:      "claude-3",
		Puppet:     "kankuro",
		SkillIDs:   []string{"s1", "s2"},
		PersonaIDs: []string{"p1"},
	}

	skills := []config.Skill{
		{ID: "s1", Name: "Skill1"},
		{ID: "s2", Name: "Skill2"},
	}
	personas := []config.Persona{
		{ID: "p1", Name: "Persona1"},
	}

	output := components.RenderCardItem(card, skills, personas, nil, false, 50)

	// Should contain skill and persona pills
	if !strings.Contains(output, "Skill1") {
		t.Error("Expected skill pill 'Skill1' in output")
	}
	if !strings.Contains(output, "Skill2") {
		t.Error("Expected skill pill 'Skill2' in output")
	}
	if !strings.Contains(output, "Persona1") {
		t.Error("Expected persona pill 'Persona1' in output")
	}
}

func TestRenderCardItemWithEmptyPuppet(t *testing.T) {
	// Test with empty puppet - should use default
	card := config.IDCard{
		ID:     "1",
		Name:   "Test Card",
		Driver: "claude",
		Model:  "claude-3",
		Puppet: "", // Empty puppet
	}

	output := components.RenderCardItem(card, nil, nil, nil, false, 50)

	// Should not be empty - should render default puppet
	if output == "" {
		t.Error("RenderCardItem with empty puppet should still render")
	}
}

func TestRenderCardItemNoCapabilities(t *testing.T) {
	card := config.IDCard{
		ID:     "1",
		Name:   "Test Card",
		Driver: "claude",
		Model:  "claude-3",
		Puppet: "kankuro",
	}

	// Empty skills and personas
	output := components.RenderCardItem(card, nil, nil, nil, false, 50)

	// Should show "no capabilities" message
	if !strings.Contains(output, "no capabilities") {
		t.Error("Expected 'no capabilities' message when no skills/personas")
	}
}

func TestRenderCardItem_MultibyteName(t *testing.T) {
	card := config.IDCard{
		ID:     "1",
		Name:   "日本語のとても長いカード名前テスト",
		Driver: "claude",
		Model:  "claude-3",
		Puppet: "kankuro",
	}

	output := components.RenderCardItem(card, nil, nil, nil, false, 30)

	if !utf8.ValidString(output) {
		t.Fatalf("output is not valid UTF-8: %q", output)
	}
	if strings.ContainsRune(output, '\uFFFD') {
		t.Errorf("output contains a replacement rune, name was cut mid-rune: %q", output)
	}
}
