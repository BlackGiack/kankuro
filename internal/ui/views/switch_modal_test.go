package views_test

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"kankuro/internal/config"
	"kankuro/internal/ui/views"
)

func TestSwitchModal_New(t *testing.T) {
	m := views.NewSwitchModal()

	if m == nil {
		t.Fatal("NewSwitchModal returned nil")
	}

	if m.IsOpen() {
		t.Error("SwitchModal should not be open initially")
	}
}

func TestSwitchModal_SetCards(t *testing.T) {
	m := views.NewSwitchModal()

	cards := []config.IDCard{
		{ID: "c1", Name: "Card 1"},
		{ID: "c2", Name: "Card 2"},
	}

	m.SetCards(cards, nil, nil)

	if len(m.List().Items()) != 2 {
		t.Errorf("Expected 2 items, got %d", len(m.List().Items()))
	}
}

func TestSwitchModal_OpenClose(t *testing.T) {
	m := views.NewSwitchModal()

	m.Open()
	if !m.IsOpen() {
		t.Error("SwitchModal should be open after Open()")
	}

	m.Close()
	if m.IsOpen() {
		t.Error("SwitchModal should be closed after Close()")
	}
}

func TestSwitchModal_Confirmation(t *testing.T) {
	m := views.NewSwitchModal()

	cards := []config.IDCard{
		{ID: "c1", Name: "Card 1"},
	}
	m.SetCards(cards, nil, nil)
	m.Open()

	// Simulate Enter key
	keyMsg := tea.KeyMsg{Type: tea.KeyEnter}
	m.Update(keyMsg)

	if !m.Confirmed() {
		t.Error("SwitchModal should be confirmed after Enter")
	}

	if m.IsOpen() {
		t.Error("SwitchModal should be closed after confirmation")
	}
}

func TestSwitchModal_Cancellation(t *testing.T) {
	m := views.NewSwitchModal()

	cards := []config.IDCard{
		{ID: "c1", Name: "Card 1"},
	}
	m.SetCards(cards, nil, nil)
	m.Open()

	// Simulate Esc key
	keyMsg := tea.KeyMsg{Type: tea.KeyEsc}
	m.Update(keyMsg)

	if !m.Cancelled() {
		t.Error("SwitchModal should be cancelled after Esc")
	}

	if m.IsOpen() {
		t.Error("SwitchModal should be closed after cancellation")
	}
}

func TestSwitchModal_SelectedCardID(t *testing.T) {
	m := views.NewSwitchModal()

	cards := []config.IDCard{
		{ID: "c1", Name: "Card 1"},
		{ID: "c2", Name: "Card 2"},
	}
	m.SetCards(cards, nil, nil)
	m.Select(1) // Select second card

	selectedID := m.SelectedCardID()
	if selectedID != "c2" {
		t.Errorf("Expected selected card ID 'c2', got '%s'", selectedID)
	}
}
