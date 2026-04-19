package views

import (
	"testing"

	bubbletea "github.com/charmbracelet/bubbletea"

	"kankuro/internal/config"
)

func TestSwitchModal_New(t *testing.T) {
	m := NewSwitchModal()

	if m == nil {
		t.Fatal("NewSwitchModal returned nil")
	}

	if m.IsOpen() {
		t.Error("SwitchModal should not be open initially")
	}
}

func TestSwitchModal_SetCards(t *testing.T) {
	m := NewSwitchModal()

	cards := []config.IDCard{
		{ID: "c1", Name: "Card 1"},
		{ID: "c2", Name: "Card 2"},
	}

	m.SetCards(cards, nil, nil)

	if len(m.list.Items()) != 2 {
		t.Errorf("Expected 2 items, got %d", len(m.list.Items()))
	}
}

func TestSwitchModal_OpenClose(t *testing.T) {
	m := NewSwitchModal()

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
	m := NewSwitchModal()

	cards := []config.IDCard{
		{ID: "c1", Name: "Card 1"},
	}
	m.SetCards(cards, nil, nil)
	m.Open()

	// Simulate Enter key
	keyMsg := bubbletea.KeyMsg{Type: bubbletea.KeyEnter}
	m.Update(keyMsg)

	if !m.Confirmed() {
		t.Error("SwitchModal should be confirmed after Enter")
	}

	if m.IsOpen() {
		t.Error("SwitchModal should be closed after confirmation")
	}
}

func TestSwitchModal_Cancellation(t *testing.T) {
	m := NewSwitchModal()

	cards := []config.IDCard{
		{ID: "c1", Name: "Card 1"},
	}
	m.SetCards(cards, nil, nil)
	m.Open()

	// Simulate Esc key
	keyMsg := bubbletea.KeyMsg{Type: bubbletea.KeyEsc}
	m.Update(keyMsg)

	if !m.Cancelled() {
		t.Error("SwitchModal should be cancelled after Esc")
	}

	if m.IsOpen() {
		t.Error("SwitchModal should be closed after cancellation")
	}
}

func TestSwitchModal_SelectedCardID(t *testing.T) {
	m := NewSwitchModal()

	cards := []config.IDCard{
		{ID: "c1", Name: "Card 1"},
		{ID: "c2", Name: "Card 2"},
	}
	m.SetCards(cards, nil, nil)
	m.list.Select(1) // Select second card

	selectedID := m.SelectedCardID()
	if selectedID != "c2" {
		t.Errorf("Expected selected card ID 'c2', got '%s'", selectedID)
	}
}
