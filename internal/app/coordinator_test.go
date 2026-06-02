package app_test

import (
	"testing"

	tea "charm.land/bubbletea/v2"

	"kankuro/internal/app"
	"kankuro/internal/config"
	"kankuro/internal/session"
)

func TestCoordinator_New(t *testing.T) {
	repo, _ := config.NewFileRepository(t.TempDir())
	sm := session.NewManager(repo.Dir(), repo)

	coord := app.NewCoordinator(repo, sm)

	if coord == nil {
		t.Fatal("NewCoordinator returned nil")
	}

	if coord.CurrentView() != app.ViewDashboard {
		t.Errorf("Expected initial view ViewDashboard, got %v", coord.CurrentView())
	}

	if len(coord.Views()) != 10 {
		t.Errorf("Expected 10 views, got %d", len(coord.Views()))
	}
}

func TestCoordinator_Navigate(t *testing.T) {
	repo, _ := config.NewFileRepository(t.TempDir())
	sm := session.NewManager(repo.Dir(), repo)

	coord := app.NewCoordinator(repo, sm)
	coord.Init()

	coord.NavigateTo(app.ViewCardEditor)

	if coord.CurrentView() != app.ViewCardEditor {
		t.Errorf("Expected ViewCardEditor, got %v", coord.CurrentView())
	}
	if coord.PrevView() != app.ViewDashboard {
		t.Errorf("Expected prevView ViewDashboard, got %v", coord.PrevView())
	}
}

func TestCoordinator_ResizeAll(t *testing.T) {
	repo, _ := config.NewFileRepository(t.TempDir())
	sm := session.NewManager(repo.Dir(), repo)

	coord := app.NewCoordinator(repo, sm)
	coord.Init()

	// Since we can't directly set private width/height, we use Update with WindowSizeMsg
	// or we could add setters. I added getters and ResizeAll, but not setters for width/height.
	// Actually, I can use tea.WindowSizeMsg to set them via Update.
	coord.Update(tea.WindowSizeMsg{Width: 120, Height: 40})

	if coord.Width() != 120 {
		t.Errorf("Coordinator width should be 120, got %d", coord.Width())
	}
	if coord.Height() != 40 {
		t.Errorf("Coordinator height should be 40, got %d", coord.Height())
	}
}

func TestCoordinator_WireCommands(t *testing.T) {
	repo, _ := config.NewFileRepository(t.TempDir())
	sm := session.NewManager(repo.Dir(), repo)

	cmds := app.WireCommands(repo, sm)

	if cmds == nil {
		t.Fatal("WireCommands returned nil")
	}

	if cmds.LaunchSession == nil {
		t.Error("LaunchSession should not be nil")
	}
	if cmds.SaveCard == nil {
		t.Error("SaveCard should not be nil")
	}
}

func TestApp_BackwardsCompatibility(t *testing.T) {
	repo, _ := config.NewFileRepository(t.TempDir())
	sm := session.NewManager(repo.Dir(), repo)

	app, err := app.New(repo, sm)
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}

	if app.Coordinator() == nil {
		t.Error("App.coordinator should not be nil")
	}

	cmd := app.Init()
	if cmd == nil {
		t.Error("App.Init should return a command")
	}

	app.SetSize(120, 40)
	if app.Coordinator().Width() != 120 {
		t.Error("SetSize should propagate to coordinator")
	}

	view := app.View()
	if view.Content == "" {
		t.Error("View should not be empty")
	}
}
