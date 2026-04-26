package ui_test

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"kankuro/internal/config"
	"kankuro/internal/session"
	"kankuro/internal/ui"
	"kankuro/internal/ui/messages"
)

func TestCoordinator_NavigationKeys(t *testing.T) {
	repo, _ := config.NewFileRepository(t.TempDir())
	sm := session.NewManager(repo.Dir(), repo)

	coord := ui.NewCoordinator(repo, sm)
	coord.Init()

	// Test 'n' key - should navigate to CardEditor
	_, cmd := coord.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	// Execute the command to get the NavigateMsg
	if cmd != nil {
		msg := cmd()
		if navigateMsg, ok := msg.(messages.NavigateMsg); ok {
			coord.Update(navigateMsg)
		}
	}
	if coord.CurrentView() != ui.ViewCardEditor {
		t.Errorf("Expected ViewCardEditor after 'n' key, got %v", coord.CurrentView())
	}

	// Navigate back to dashboard
	coord.Update(messages.NavigateMsg{View: messages.ViewDashboard})

	// Test 'm' key - should navigate to MCPForm
	_, cmd = coord.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'m'}})
	// Execute the command to get the NavigateMsg
	if cmd != nil {
		msg := cmd()
		if navigateMsg, ok := msg.(messages.NavigateMsg); ok {
			coord.Update(navigateMsg)
		}
	}
	if coord.CurrentView() != ui.ViewMCPForm {
		t.Errorf("Expected ViewMCPForm after 'm' key, got %v", coord.CurrentView())
	}

	// Navigate back to dashboard
	coord.Update(messages.NavigateMsg{View: messages.ViewDashboard})

	// Test 'w' key - should navigate to SourceManagerWorkspaces
	_, cmd = coord.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'w'}})
	// Execute the command to get the NavigateMsg
	if cmd != nil {
		msg := cmd()
		if navigateMsg, ok := msg.(messages.NavigateMsg); ok {
			coord.Update(navigateMsg)
		}
	}
	if coord.CurrentView() != ui.ViewSourceManagerWorkspaces {
		t.Errorf("Expected ViewSourceManagerWorkspaces after 'w' key, got %v", coord.CurrentView())
	}

	// Navigate back to dashboard
	coord.Update(messages.NavigateMsg{View: messages.ViewDashboard})

	// Test 'i' key - should navigate to IDEConfig
	_, cmd = coord.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'i'}})
	// Execute the command to get the NavigateMsg
	if cmd != nil {
		msg := cmd()
		if navigateMsg, ok := msg.(messages.NavigateMsg); ok {
			coord.Update(navigateMsg)
		}
	}
	if coord.CurrentView() != ui.ViewIDEConfig {
		t.Errorf("Expected ViewIDEConfig after 'i' key, got %v", coord.CurrentView())
	}
}
