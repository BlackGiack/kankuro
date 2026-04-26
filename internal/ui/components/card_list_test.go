package components_test

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"

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
	lipgloss.SetColorProfile(termenv.TrueColor)
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
	output := components.RenderCardItem(card, skills, personas, false)
	if output == "" {
		t.Error("RenderCardItem returned empty string")
	}

	// Render selected
	outputSel := components.RenderCardItem(card, skills, personas, true)
	if outputSel == "" {
		t.Error("RenderCardItem (selected) returned empty string")
	}

	// Selected should be different (different border color)
	if output == outputSel {
		t.Error("Selected and unselected should render differently")
	}
}
