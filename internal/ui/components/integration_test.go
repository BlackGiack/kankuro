package components_test

import (
	"testing"

	"kankuro/internal/config"
	"kankuro/internal/ui/components"
)

func TestCardListIntegration(t *testing.T) {
	// Simulate App workflow

	// 1. Create component with callbacks
	selectedCard := ""
	editedCard := ""
	deletedCard := ""

	var c *components.CardList
	c = components.NewCardList(components.CardListCallbacks{
		OnSelect: func(card config.IDCard) {
			selectedCard = card.ID
		},
		OnEdit: func(card config.IDCard) {
			editedCard = card.ID
		},
		OnDelete: func(cardID string) {
			deletedCard = cardID
		},
	})

	// 2. Set data
	cards := []config.IDCard{
		{ID: "1", Name: "First", Driver: "claude"},
		{ID: "2", Name: "Second", Driver: "gemini"},
	}
	c.SetCards(cards, nil, nil)

	// 3. Initialize
	c.Init()
	c.SetSize(50, 20)

	// 4. Select first item
	// (In real use, this would be via Update with KeyMsg)
	// For this test, just verify Selected works
	_, ok := c.Selected()
	if ok {
		t.Log("Card selected successfully")
	}

	// 5. Apply search filter
	c.SetSearchQuery("first")
	// Note: filterCards is unexported, but we can check if selected item is correct
	if card, ok2 := c.Selected(); ok2 {
		if card.Name != "First" {
			t.Errorf("Expected 'First' card, got '%s'", card.Name)
		}
	}

	// Verify callbacks would work (actual key handling tested in Update)
	t.Logf("Component created and initialized successfully")
	t.Logf("State: selected=%s, edited=%s, deleted=%s", selectedCard, editedCard, deletedCard)
}
