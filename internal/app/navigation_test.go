package app_test

import (
	"testing"

	tea "charm.land/bubbletea/v2"

	"kankuro/internal/app"
	"kankuro/internal/config"
	"kankuro/internal/session"
	"kankuro/internal/ui/messages"
)

func TestCoordinator_NavigationKeys(t *testing.T) {
	repo, _ := config.NewFileRepository(t.TempDir())
	sm := session.NewManager(repo.Dir(), repo)

	coord := app.NewCoordinator(repo, sm)
	coord.Init()

	// Test 'n' key - should navigate to CardEditor
	_, cmd := coord.Update(tea.KeyPressMsg{Code: 0, Text: "n"})
	// Execute the command to get the NavigateMsg
	if cmd != nil {
		msg := cmd()
		if navigateMsg, ok := msg.(messages.NavigateMsg); ok {
			coord.Update(navigateMsg)
		}
	}
	if coord.CurrentView() != app.ViewCardEditor {
		t.Errorf("Expected ViewCardEditor after 'n' key, got %v", coord.CurrentView())
	}

	// Navigate back to dashboard
	coord.Update(messages.NavigateMsg{View: messages.ViewDashboard})

	// Test 'm' key - should navigate to MCPForm
	_, cmd = coord.Update(tea.KeyPressMsg{Code: 0, Text: "m"})
	// Execute the command to get the NavigateMsg
	if cmd != nil {
		msg := cmd()
		if navigateMsg, ok := msg.(messages.NavigateMsg); ok {
			coord.Update(navigateMsg)
		}
	}
	if coord.CurrentView() != app.ViewMCPForm {
		t.Errorf("Expected ViewMCPForm after 'm' key, got %v", coord.CurrentView())
	}

	// Navigate back to dashboard
	coord.Update(messages.NavigateMsg{View: messages.ViewDashboard})

	// Test 'w' key - should navigate to SourceManagerWorkspaces
	_, cmd = coord.Update(tea.KeyPressMsg{Code: 0, Text: "w"})
	// Execute the command to get the NavigateMsg
	if cmd != nil {
		msg := cmd()
		if navigateMsg, ok := msg.(messages.NavigateMsg); ok {
			coord.Update(navigateMsg)
		}
	}
	if coord.CurrentView() != app.ViewSourceManagerWorkspaces {
		t.Errorf("Expected ViewSourceManagerWorkspaces after 'w' key, got %v", coord.CurrentView())
	}

	// Navigate back to dashboard
	coord.Update(messages.NavigateMsg{View: messages.ViewDashboard})

	// Test 'c' key - should navigate to Settings
	_, cmd = coord.Update(tea.KeyPressMsg{Code: 0, Text: "c"})
	if cmd != nil {
		msg := cmd()
		if navigateMsg, ok := msg.(messages.NavigateMsg); ok {
			coord.Update(navigateMsg)
		}
	}
	if coord.CurrentView() != app.ViewSettings {
		t.Errorf("Expected ViewSettings after 'c' key, got %v", coord.CurrentView())
	}
}
