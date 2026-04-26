package views_test

import (
	"errors"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"kankuro/internal/config"
	"kankuro/internal/session"
	"kankuro/internal/ui/commands"
	"kankuro/internal/ui/messages"
	"kankuro/internal/ui/views"
)

var errTest = errors.New("test")

func TestDashboard_New(t *testing.T) {
	callbacks := views.DashboardCallbacks{
		OnEditCard: func(_ config.IDCard) tea.Cmd {
			return func() tea.Msg { return nil }
		},
		OnLaunchCard: func(_ config.IDCard) tea.Cmd {
			return func() tea.Msg { return nil }
		},
	}

	d := views.NewDashboard(callbacks, &commands.Commands{})

	if d == nil {
		t.Fatal("NewDashboard returned nil")
	}

	if d.FocusedPane() != views.PaneCards {
		t.Error("Dashboard should start with cards pane focused")
	}
}

func TestDashboard_SetData(t *testing.T) {
	d := views.NewDashboard(views.DashboardCallbacks{}, &commands.Commands{})

	cards := []config.IDCard{
		{ID: "c1", Name: "Card 1"},
	}
	sessions := []session.Snapshot{
		{ID: "s1", Name: "Session 1"},
	}

	d.SetData(messages.DataLoadedMsg{
		Cards:    cards,
		Sessions: sessions,
	})

	if len(d.Cards()) != 1 {
		t.Errorf("Expected 1 card, got %d", len(d.Cards()))
	}
	if len(d.Sessions()) != 1 {
		t.Errorf("Expected 1 session, got %d", len(d.Sessions()))
	}
}

func TestDashboard_TogglePane(t *testing.T) {
	d := views.NewDashboard(views.DashboardCallbacks{}, &commands.Commands{})
	d.SetData(messages.DataLoadedMsg{
		Sessions: []session.Snapshot{{ID: "s1"}},
	})

	if d.FocusedPane() != views.PaneCards {
		t.Error("Should start on cards pane")
	}

	d.TogglePane()
	if d.FocusedPane() != views.PaneSessions {
		t.Error("Should switch to sessions pane")
	}

	d.TogglePane()
	if d.FocusedPane() != views.PaneCards {
		t.Error("Should switch back to cards pane")
	}
}

func TestDashboard_Search(t *testing.T) {
	d := views.NewDashboard(views.DashboardCallbacks{}, &commands.Commands{})

	// Initialize the dashboard
	d.Init()

	// Activate search
	keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}}
	d.Update(keyMsg)

	if !d.Search().IsActive {
		t.Error("Search should be active after pressing '/'")
	}

	// Type search query
	testChars := []rune{'t', 'e', 's', 't'}
	for _, r := range testChars {
		charMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}}
		d.Update(charMsg)
	}

	if d.Search().CardsQuery != "test" {
		t.Errorf("Expected query 'test', got '%s'", d.Search().CardsQuery)
	}

	// Exit search
	keyMsg = tea.KeyMsg{Type: tea.KeyEsc}
	d.Update(keyMsg)

	if d.Search().IsActive {
		t.Error("Search should be inactive after escape")
	}
}

func TestDashboard_DeleteModal(t *testing.T) {
	var deletedSessionID string

	callbacks := views.DashboardCallbacks{
		OnRemove: func(sessionID string) (tea.Cmd, error) {
			deletedSessionID = sessionID
			return func() tea.Msg { return nil }, errTest
		},
	}

	d := views.NewDashboard(callbacks, &commands.Commands{})
	d.SetData(messages.DataLoadedMsg{
		Sessions: []session.Snapshot{
			{ID: "s1", Name: "Test Session", UseTmpWorkspace: true, Status: session.StatusTerminated},
		},
	})
	// We need to be able to set focused pane for test
	// I'll add a setter or just use toggle if it starts on cards
	if d.FocusedPane() == views.PaneCards {
		d.TogglePane()
	}

	// Initialize and select the first session
	d.Init()
	d.SessionList().SetSelectedByID("s1")

	// Trigger delete
	keyMsg := tea.KeyMsg{Type: tea.KeyBackspace}
	d.Update(keyMsg)

	if !d.DeleteModal().IsOpen() {
		t.Error("Delete modal should be open")
	}

	// Confirm delete
	keyMsg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}}
	d.Update(keyMsg)

	if d.DeleteModal().IsOpen() {
		t.Error("Delete modal should be closed after confirm")
	}
	if deletedSessionID != "s1" {
		t.Errorf("Expected deleted session 's1', got '%s'", deletedSessionID)
	}
}

func TestDashboard_SwitchModal(t *testing.T) {
	var switchedSessionID string

	callbacks := views.DashboardCallbacks{
		OnSwitch: func(sessionID, _ string) (tea.Cmd, error) {
			switchedSessionID = sessionID
			return func() tea.Msg { return nil }, errTest
		},
	}

	d := views.NewDashboard(callbacks, &commands.Commands{})
	d.SetData(messages.DataLoadedMsg{
		Cards:    []config.IDCard{{ID: "c1", Name: "Card 1"}},
		Sessions: []session.Snapshot{{ID: "s1", Name: "Test Session", Status: session.StatusTerminated}},
	})
	if d.FocusedPane() == views.PaneCards {
		d.TogglePane()
	}

	// Initialize and select the first session
	d.Init()
	d.SessionList().SetSelectedByID("s1")

	// Trigger switch
	keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}}
	d.Update(keyMsg)

	if !d.SwitchModal().IsOpen() {
		t.Error("Switch modal should be open")
	}

	// Check if the switch modal has items
	if len(d.SwitchModal().List().Items()) == 0 {
		t.Error("Switch modal should have items")
	}

	// Initialize the switch modal list
	d.SwitchModal().Init()

	// Select the first item in the switch modal
	d.SwitchModal().Select(0)

	// Confirm selection
	keyMsg = tea.KeyMsg{Type: tea.KeyEnter}
	model, _ := d.Update(keyMsg)
	d = model.(*views.Dashboard)

	// Check if the modal is confirmed
	if !d.SwitchModal().Confirmed() {
		t.Error("Switch modal should be confirmed")
	}

	if d.SwitchModal().IsOpen() {
		t.Error("Switch modal should be closed after confirm")
	}
	if switchedSessionID != "s1" {
		t.Errorf("Expected session 's1', got '%s'", switchedSessionID)
	}
}

func TestDashboard_SetSize_Flexible(t *testing.T) {
	d := views.NewDashboard(views.DashboardCallbacks{}, &commands.Commands{})

	// Test wide terminal
	d.SetSize(200, 40)
	if d.CardsWidth() != 48 {
		t.Errorf("Expected cards width 48, got %d", d.CardsWidth())
	}

	// Test narrow terminal
	d.SetSize(80, 40)
	if d.CardsWidth() != 46 {
		t.Errorf("Expected cards width 46, got %d", d.CardsWidth())
	}
	if d.PreviewWidth() != 8 {
		t.Errorf("Expected preview width 8, got %d", d.PreviewWidth())
	}
}

func TestDashboard_PreviewPersistence(t *testing.T) {
	d := views.NewDashboard(views.DashboardCallbacks{}, &commands.Commands{})
	d.Init()

	s1 := session.Snapshot{ID: "s1", Name: "Session 1"}
	s2 := session.Snapshot{ID: "s2", Name: "Session 2"}
	cards := []config.IDCard{{ID: "c1", Name: "Card 1"}}

	// Load sessions
	d.SetData(messages.DataLoadedMsg{
		Sessions: []session.Snapshot{s1, s2},
		Cards:    cards,
	})

	// Focus sessions and select s2
	if d.FocusedPane() == views.PaneCards {
		d.TogglePane()
	}
	d.SessionList().SetSelectedByID("s2")
	d.SyncPreviewSelection()

	if d.Preview().SessionID() != "s2" {
		t.Errorf("Expected previewing s2, got %s", d.Preview().SessionID())
	}

	// Switch back to cards pane
	d.TogglePane()
	d.SyncPreviewSelection()

	// Should still preview s2 (stable preview)
	if d.Preview().SessionID() != "s2" {
		t.Errorf("Expected preview to stay on s2 when switching panes, got %s", d.Preview().SessionID())
	}

	// Remove s2
	d.SetData(messages.DataLoadedMsg{
		Sessions: []session.Snapshot{s1},
		Cards:    cards,
	})

	// Preview should recover to s1
	if d.Preview().SessionID() != "s1" {
		t.Errorf("Expected preview to recover to s1 after s2 removed, got %s", d.Preview().SessionID())
	}

	// Remove all sessions
	d.SetData(messages.DataLoadedMsg{
		Sessions: []session.Snapshot{},
		Cards:    cards,
	})

	// Preview should be cleared
	if d.Preview().SessionID() != "" {
		t.Error("Expected preview to be cleared when no sessions remain")
	}
}

func TestDashboard_MCPActions(t *testing.T) {
	var mcpOpened bool
	callbacks := views.DashboardCallbacks{
		OnOpenMCP: func() tea.Cmd {
			mcpOpened = true
			return nil
		},
	}
	d := views.NewDashboard(callbacks, &commands.Commands{})
	d.Init()

	// Press 'm' to open MCP
	keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'m'}}
	d.Update(keyMsg)

	if !mcpOpened {
		t.Error("Expected MCP form to open on 'm'")
	}
}

func TestDashboard_View(t *testing.T) {
	d := views.NewDashboard(views.DashboardCallbacks{}, &commands.Commands{})
	d.SetData(messages.DataLoadedMsg{
		Cards:    []config.IDCard{{ID: "c1", Name: "Test Card", Driver: "claude", Model: "claude-3"}},
		Sessions: []session.Snapshot{{ID: "s1", Name: "Test Session"}},
	})
	d.SetSize(120, 40)

	view := d.View()

	if view == "" {
		t.Error("View should not be empty")
	}
	if !strings.Contains(view, "K A N K U R O") {
		t.Error("View should contain header")
	}
	if !strings.Contains(view, "ID Cards") {
		t.Error("View should contain cards pane title")
	}
	if !strings.Contains(view, "Active Sessions") {
		t.Error("View should contain sessions pane title")
	}
}
