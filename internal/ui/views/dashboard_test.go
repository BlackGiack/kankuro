package views

import (
	"testing"

	bubbletea "github.com/charmbracelet/bubbletea"

	"kankuro/internal/config"
	"kankuro/internal/session"
	"kankuro/internal/ui/commands"
)

func TestDashboard_New(t *testing.T) {
	callbacks := DashboardCallbacks{
		OnEditCard:   func(_ config.IDCard) {},
		OnLaunchCard: func(_ config.IDCard) {},
	}

	d := NewDashboard(callbacks, &commands.Commands{})

	if d == nil {
		t.Fatal("NewDashboard returned nil")
	}

	if d.focusedPane != PaneCards {
		t.Error("Dashboard should start with cards pane focused")
	}
}

func TestDashboard_SetData(t *testing.T) {
	d := NewDashboard(DashboardCallbacks{}, &commands.Commands{})

	cards := []config.IDCard{
		{ID: "c1", Name: "Card 1"},
	}
	sessions := []session.Snapshot{
		{ID: "s1", Name: "Session 1"},
	}

	d.SetData(cards, sessions, nil, nil, nil)

	if len(d.cards) != 1 {
		t.Errorf("Expected 1 card, got %d", len(d.cards))
	}
	if len(d.sessions) != 1 {
		t.Errorf("Expected 1 session, got %d", len(d.sessions))
	}
}

func TestDashboard_TogglePane(t *testing.T) {
	d := NewDashboard(DashboardCallbacks{}, &commands.Commands{})
	d.SetData(nil, []session.Snapshot{{ID: "s1"}}, nil, nil, nil)

	if d.focusedPane != PaneCards {
		t.Error("Should start on cards pane")
	}

	d.togglePane()
	if d.focusedPane != PaneSessions {
		t.Error("Should switch to sessions pane")
	}

	d.togglePane()
	if d.focusedPane != PaneCards {
		t.Error("Should switch back to cards pane")
	}
}

func TestDashboard_Search(t *testing.T) {
	d := NewDashboard(DashboardCallbacks{}, &commands.Commands{})

	// Initialize the dashboard
	d.Init()

	// Activate search
	keyMsg := bubbletea.KeyMsg{Type: bubbletea.KeyRunes, Runes: []rune{'/'}}
	d.Update(keyMsg)

	if !d.search.IsActive {
		t.Error("Search should be active after pressing '/'")
	}

	// Type search query
	testChars := []rune{'t', 'e', 's', 't'}
	for _, r := range testChars {
		charMsg := bubbletea.KeyMsg{Type: bubbletea.KeyRunes, Runes: []rune{r}}
		d.Update(charMsg)
	}

	if d.search.CardsQuery != "test" {
		t.Errorf("Expected query 'test', got '%s'", d.search.CardsQuery)
	}

	// Exit search
	keyMsg = bubbletea.KeyMsg{Type: bubbletea.KeyEsc}
	d.Update(keyMsg)

	if d.search.IsActive {
		t.Error("Search should be inactive after escape")
	}
}

func TestDashboard_DeleteModal(t *testing.T) {
	var deletedSessionID string

	callbacks := DashboardCallbacks{
		OnRemove: func(sessionID string) error {
			deletedSessionID = sessionID
			return nil
		},
	}

	d := NewDashboard(callbacks, &commands.Commands{})
	d.SetData(nil, []session.Snapshot{
		{ID: "s1", Name: "Test Session", UseTmpWorkspace: true},
	}, nil, nil, nil)
	d.focusedPane = PaneSessions

	// Initialize and select the first session
	d.Init()
	d.sessionList.SetSelectedByID("s1")

	// Trigger delete
	keyMsg := bubbletea.KeyMsg{Type: bubbletea.KeyBackspace}
	d.Update(keyMsg)

	if !d.deleteModal.IsOpen() {
		t.Error("Delete modal should be open")
	}

	// Confirm delete
	keyMsg = bubbletea.KeyMsg{Type: bubbletea.KeyRunes, Runes: []rune{'y'}}
	d.Update(keyMsg)

	if d.deleteModal.IsOpen() {
		t.Error("Delete modal should be closed after confirm")
	}
	if deletedSessionID != "s1" {
		t.Errorf("Expected deleted session 's1', got '%s'", deletedSessionID)
	}
}

func TestDashboard_SwitchModal(t *testing.T) {
	var switchedSessionID string

	callbacks := DashboardCallbacks{
		OnSwitch: func(sessionID, _ string) error {
			switchedSessionID = sessionID
			return nil
		},
	}

	d := NewDashboard(callbacks, &commands.Commands{})
	d.SetData(
		[]config.IDCard{{ID: "c1", Name: "Card 1"}},
		[]session.Snapshot{{ID: "s1", Name: "Test Session", Status: session.StatusTerminated}},
		nil, nil, nil,
	)
	d.focusedPane = PaneSessions

	// Initialize and select the first session
	d.Init()
	d.sessionList.SetSelectedByID("s1")

	// Trigger switch
	keyMsg := bubbletea.KeyMsg{Type: bubbletea.KeyRunes, Runes: []rune{'s'}}
	d.Update(keyMsg)

	if !d.switchModal.IsOpen() {
		t.Error("Switch modal should be open")
	}

	// Check if the switch modal has items
	if len(d.switchModal.list.Items()) == 0 {
		t.Error("Switch modal should have items")
	}

	// Initialize the switch modal list
	d.switchModal.Init()

	// Select the first item in the switch modal
	d.switchModal.list.Select(0)

	// Confirm selection
	keyMsg = bubbletea.KeyMsg{Type: bubbletea.KeyEnter}
	model, _ := d.Update(keyMsg)
	d = model.(*Dashboard)

	// Check if the modal is confirmed
	if !d.switchModal.Confirmed() {
		t.Error("Switch modal should be confirmed")
	}

	if d.switchModal.IsOpen() {
		t.Error("Switch modal should be closed after confirm")
	}
	if switchedSessionID != "s1" {
		t.Errorf("Expected session 's1', got '%s'", switchedSessionID)
	}
}

func TestDashboard_SetSize(t *testing.T) {
	d := NewDashboard(DashboardCallbacks{}, &commands.Commands{})
	d.SetSize(120, 40)

	if d.Width != 120 {
		t.Errorf("Expected width 120, got %d", d.Width)
	}
	if d.Height != 40 {
		t.Errorf("Expected height 40, got %d", d.Height)
	}

	// Verify proportions
	expectedCardsWidth := 120 / 4
	if d.cardsWidth != expectedCardsWidth {
		t.Errorf("Expected cards width %d, got %d", expectedCardsWidth, d.cardsWidth)
	}
}

func TestDashboard_View(t *testing.T) {
	d := NewDashboard(DashboardCallbacks{}, &commands.Commands{})
	d.SetData(
		[]config.IDCard{{ID: "c1", Name: "Test Card", Driver: "claude", Model: "claude-3"}},
		[]session.Snapshot{{ID: "s1", Name: "Test Session"}},
		nil, nil, nil,
	)
	d.SetSize(120, 40)

	view := d.View()

	if view == "" {
		t.Error("View should not be empty")
	}
	if !containsString(view, "K A N K U R O") {
		t.Error("View should contain header")
	}
	if !containsString(view, "ID Cards") {
		t.Error("View should contain cards pane title")
	}
	if !containsString(view, "Active Sessions") {
		t.Error("View should contain sessions pane title")
	}
}

func containsString(s, substr string) bool {
	return len(s) >= len(substr) &&
		(s == substr || findSubstring(s, substr))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
