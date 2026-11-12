package app_test

import (
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/BlackGiack/kankuro/internal/app"
	"github.com/BlackGiack/kankuro/internal/ui/messages"
)

//nolint:gocognit // Verify navigation across all source and settings views.
func TestCoordinator_NavigationKeys(t *testing.T) {
	repo, _ := newCurrentRepository(t.TempDir())
	sm := newTestManager(t, repo.Dir(), repo)

	coord := newScopedCoordinator(repo, sm)
	coord.Init()
	coord.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	openDefaultDashboard(coord)

	// Test 'n' key - should navigate to CardEditor
	_, cmd := coord.Update(tea.KeyPressMsg{Code: 0, Text: "n"})
	// Execute the captured-context navigation command
	if cmd != nil {
		msg := cmd()
		if navigateMsg, ok := msg.(messages.ScopedNavigateMsg); ok {
			coord.Update(navigateMsg)
		}
	}
	if coord.CurrentView() != app.ViewCardEditor {
		t.Errorf("Expected ViewCardEditor after 'n' key, got %v", coord.CurrentView())
	}

	// Navigate back to dashboard
	coord.Update(messages.ScopedNavigateMsg{View: messages.ViewDashboard, DashboardID: "default", Generation: 1})

	// Test 'm' key - should navigate to MCPManager
	_, cmd = coord.Update(tea.KeyPressMsg{Code: 0, Text: "m"})
	// Execute the captured-context navigation command
	if cmd != nil {
		msg := cmd()
		if navigateMsg, ok := msg.(messages.ScopedNavigateMsg); ok {
			coord.Update(navigateMsg)
		}
	}
	if coord.CurrentView() != app.ViewMCPManager {
		t.Errorf("Expected ViewMCPManager after 'm' key, got %v", coord.CurrentView())
	}

	// Navigate back to dashboard
	coord.Update(messages.ScopedNavigateMsg{View: messages.ViewDashboard, DashboardID: "default", Generation: 1})

	// Test 'w' key - should navigate to SourceManagerWorkspaces
	_, cmd = coord.Update(tea.KeyPressMsg{Code: 0, Text: "w"})
	// Execute the captured-context navigation command
	if cmd != nil {
		msg := cmd()
		if navigateMsg, ok := msg.(messages.ScopedNavigateMsg); ok {
			coord.Update(navigateMsg)
		}
	}
	if coord.CurrentView() != app.ViewSourceManagerWorkspaces {
		t.Errorf("Expected ViewSourceManagerWorkspaces after 'w' key, got %v", coord.CurrentView())
	}

	// Navigate back to dashboard
	coord.Update(messages.ScopedNavigateMsg{View: messages.ViewDashboard, DashboardID: "default", Generation: 1})

	// Test 'c' key - should navigate to Settings
	_, cmd = coord.Update(tea.KeyPressMsg{Code: 0, Text: "c"})
	if cmd != nil {
		msg := cmd()
		if navigateMsg, ok := msg.(messages.ScopedNavigateMsg); ok {
			coord.Update(navigateMsg)
		}
	}
	if coord.CurrentView() != app.ViewSettings {
		t.Errorf("Expected ViewSettings after 'c' key, got %v", coord.CurrentView())
	}
	_, cmd = coord.Update(tea.KeyPressMsg{Text: "p", Code: 'p'})
	if cmd == nil {
		t.Fatal("settings did not open puppet sources")
	}
	coord.Update(cmd())
	if coord.CurrentView() != app.ViewSourceManagerPuppets {
		t.Fatal("puppet sources did not open")
	}
	_, cmd = coord.Update(tea.KeyPressMsg{Code: tea.KeyEsc})
	if cmd == nil {
		t.Fatal("source manager did not return to settings")
	}
	coord.Update(cmd())
	if coord.CurrentView() != app.ViewSettings {
		t.Fatal("source manager returned to the wrong view")
	}
	_, cmd = coord.Update(tea.KeyPressMsg{Code: tea.KeyEsc})
	if cmd == nil {
		t.Fatal("settings did not return to dashboard")
	}
	coord.Update(cmd())
	if coord.CurrentView() != app.ViewDashboard {
		t.Fatal("settings returned to the wrong view")
	}
	_, cmd = coord.Update(tea.KeyPressMsg{Code: tea.KeyEsc})
	if cmd == nil {
		t.Fatal("dashboard did not return to list")
	}
	coord.Update(cmd())
	if coord.CurrentView() != app.ViewDashboardList {
		t.Fatal("dashboard list did not open")
	}
}

func TestMCPManagerNestedNavigation(t *testing.T) {
	repo, _ := newCurrentRepository(t.TempDir())
	sm := newTestManager(t, repo.Dir(), repo)
	coord := newScopedCoordinator(repo, sm)
	coord.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	openDefaultDashboard(coord)
	for range 2 {
		_, cmd := coord.Update(tea.KeyPressMsg{Text: "m"})
		coord.Update(cmd())
		_, cmd = coord.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
		coord.Update(cmd())
		if coord.CurrentView() != app.ViewMCPForm {
			t.Fatal("new MCP did not open form")
		}
		_, cmd = coord.Update(tea.KeyPressMsg{Code: tea.KeyEsc})
		coord.Update(cmd())
		if coord.CurrentView() != app.ViewMCPManager {
			t.Fatal("form did not return to list")
		}
		_, cmd = coord.Update(tea.KeyPressMsg{Code: tea.KeyEsc})
		coord.Update(cmd())
		if coord.CurrentView() != app.ViewDashboard {
			t.Fatal("list did not return to dashboard")
		}
	}
}
