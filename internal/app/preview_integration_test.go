package app_test

import (
	"bytes"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/BlackGiack/kankuro/internal/app"
	"github.com/BlackGiack/kankuro/internal/config"
	"github.com/BlackGiack/kankuro/internal/ui/messages"
	"github.com/BlackGiack/kankuro/internal/ui/views"
)

func TestPreviewImmediateUpdate(t *testing.T) {
	repo, _ := config.NewFileRepository(t.TempDir())
	sm := newTestManager(t, repo.Dir(), repo)

	wd := t.TempDir()
	repo.SaveWorkspaceSources([]config.WorkspaceSource{{Path: wd, IsRoot: false}})
	workspaces, _ := repo.ListWorkspaces()

	card, _ := repo.CreateIDCard(config.IDCard{
		Name:        "Test Card",
		Driver:      "echo",
		Model:       "test",
		WorkspaceID: workspaces[0].ID,
	})

	// Launch session
	sessionID, _ := sm.Launch(card, nil, nil, nil)

	coord := app.NewCoordinator(repo, sm)
	coord.Update(tea.WindowSizeMsg{Width: 100, Height: 40})

	// Execute refreshData from Init
	_ = coord.Init()
	// Init returns a batch, but we know refreshData is there.
	// For simplicity in test, let's just call RefreshData directly.
	msg := coord.RefreshData()()
	coord.Update(msg)

	dashboard := coord.Views()[app.ViewDashboard].(*views.Dashboard)

	// Trigger preview request (simulate selection change)
	msgReq := messages.SessionPreviewRequestedMsg{SessionID: sessionID}
	coord.Update(msgReq)

	// Check if preview is updated immediately
	if dashboard.PreviewSessionID() != sessionID {
		t.Errorf("Expected preview to be %s immediately, got %s", sessionID, dashboard.PreviewSessionID())
	}
}

func TestPreviewAutoSelection(t *testing.T) {
	repo, _ := config.NewFileRepository(t.TempDir())
	sm := newTestManager(t, repo.Dir(), repo)

	wd := t.TempDir()
	repo.SaveWorkspaceSources([]config.WorkspaceSource{{Path: wd, IsRoot: false}})
	workspaces, _ := repo.ListWorkspaces()

	card, _ := repo.CreateIDCard(config.IDCard{
		Name:        "Test Card",
		Driver:      "echo",
		Model:       "test",
		WorkspaceID: workspaces[0].ID,
	})

	// Launch session
	sessionID, _ := sm.Launch(card, nil, nil, nil)

	coord := app.NewCoordinator(repo, sm)
	coord.Update(tea.WindowSizeMsg{Width: 100, Height: 40})

	// Initial data load: refreshData will have PreviewSessionID="" because dashboard hasn't seen the session yet
	msg := coord.RefreshData()()
	dm := msg.(messages.DataLoadedMsg)
	if dm.PreviewSessionID != "" {
		t.Fatal("Expected empty PreviewSessionID in initial refreshData")
	}

	// Update coordinator: this should trigger Dashboard.SetData, which auto-selects sessionID,
	// and then Coordinator should realize it and return a subscription command.
	_, cmd := coord.Update(dm)
	if cmd == nil {
		t.Fatal("Expected subscription command from Coordinator.Update(DataLoadedMsg)")
	}

	dashboard := coord.Views()[app.ViewDashboard].(*views.Dashboard)
	if dashboard.PreviewSessionID() != sessionID {
		t.Errorf("Expected dashboard to auto-select %s, got %s", sessionID, dashboard.PreviewSessionID())
	}
}

func TestPreviewStreaming(t *testing.T) {
	repo, _ := config.NewFileRepository(t.TempDir())
	sm := newTestManager(t, repo.Dir(), repo)

	wd := t.TempDir()
	repo.SaveWorkspaceSources([]config.WorkspaceSource{{Path: wd, IsRoot: false}})
	workspaces, _ := repo.ListWorkspaces()

	card, _ := repo.CreateIDCard(config.IDCard{
		Name:        "Test Card",
		Driver:      "echo",
		Model:       "test",
		WorkspaceID: workspaces[0].ID,
	})

	// Launch session
	sessionID, _ := sm.Launch(card, nil, nil, nil)

	coord := app.NewCoordinator(repo, sm)
	coord.Update(tea.WindowSizeMsg{Width: 100, Height: 40})

	// Initial data load
	msg := coord.RefreshData()()
	coord.Update(msg)

	// Select session for preview
	coord.Update(messages.SessionPreviewRequestedMsg{SessionID: sessionID})

	dashboard := coord.Views()[app.ViewDashboard].(*views.Dashboard)

	// Simulate output from session
	testOutput := []byte("hello world")

	// Re-trigger preview to get the cmd
	var cmd tea.Cmd
	_, cmd = coord.Update(messages.SessionPreviewRequestedMsg{SessionID: sessionID})
	if cmd == nil {
		t.Fatal("Expected cmd from PreviewRequestedMsg")
	}

	// Manually send output to session subscribers
	sm.BroadcastOutput(sessionID, testOutput)

	// Execute the captured command (it should read from the channel)
	outputMsg := cmd()
	if outputMsg == nil {
		t.Fatal("Expected SessionOutputMsg from cmd")
	}

	// Update coordinator with output message
	coord.Update(outputMsg)

	// Verify preview scrollback contains the output
	preview := dashboard.Preview()
	if !bytes.Contains(preview.Scrollback(), testOutput) {
		t.Errorf("Expected scrollback to contain '%s', got '%s'", string(testOutput), string(preview.Scrollback()))
	}
}

func TestPreviewStreamingKitty(t *testing.T) {
	repo, _ := config.NewFileRepository(t.TempDir())
	sm := newTestManager(t, repo.Dir(), repo)

	wd := t.TempDir()
	repo.SaveWorkspaceSources([]config.WorkspaceSource{{Path: wd, IsRoot: false}})
	workspaces, _ := repo.ListWorkspaces()

	card, _ := repo.CreateIDCard(config.IDCard{
		Name:        "Test Card",
		Driver:      "echo",
		Model:       "test",
		WorkspaceID: workspaces[0].ID,
	})

	// Launch session
	sessionID, _ := sm.Launch(card, nil, nil, nil)

	coord := app.NewCoordinator(repo, sm)
	coord.Update(tea.WindowSizeMsg{Width: 100, Height: 40})

	// Initial data load
	msg := coord.RefreshData()()
	coord.Update(msg)

	// Select session for preview
	coord.Update(messages.SessionPreviewRequestedMsg{SessionID: sessionID})

	dashboard := coord.Views()[app.ViewDashboard].(*views.Dashboard)

	// Simulate Kitty output from session
	kittyKeyboard := []byte("\x1b[97u") // 'a'
	kittyGraphics := []byte("\x1b_Ga=d;PAYLOAD\x07")

	testOutput := make([]byte, 0, len(kittyKeyboard)+len(kittyGraphics))
	testOutput = append(testOutput, kittyKeyboard...)
	testOutput = append(testOutput, kittyGraphics...)

	// Re-trigger preview to get the cmd
	var cmd tea.Cmd
	_, cmd = coord.Update(messages.SessionPreviewRequestedMsg{SessionID: sessionID})

	// Send output to session subscribers
	sm.BroadcastOutput(sessionID, testOutput)

	// Execute cmd to get SessionOutputMsg
	outputMsg := cmd()
	coord.Update(outputMsg)

	// Verify RAW scrollback in preview component (before sanitization for view)
	preview := dashboard.Preview()
	if !bytes.Contains(preview.Scrollback(), testOutput) {
		t.Errorf("Expected RAW scrollback to contain %q, got %q", testOutput, preview.Scrollback())
	}

	// Kitty sequences are now silently swallowed: no translation to 'a' from
	// keyboard sequences, no "[image]" placeholder from graphics. The only
	// requirement is that nothing leaks into the preview, neither the raw
	// escape payload nor old placeholder strings.
	view := preview.View()
	if strings.Contains(view, "[image]") {
		t.Errorf("View should no longer contain [image] placeholder, got:\n%s", view)
	}
	if strings.Contains(view, "PAYLOAD") || strings.Contains(view, "\x1b_") || strings.Contains(view, "\x07") {
		t.Errorf("View should not contain raw Kitty payload, got:\n%s", view)
	}
}
