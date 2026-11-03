package app_test

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/BlackGiack/kankuro/internal/api"
	"github.com/BlackGiack/kankuro/internal/app"
	"github.com/BlackGiack/kankuro/internal/ui/messages"
)

func TestCoordinator_New(t *testing.T) {
	repo, _ := newCurrentRepository(t.TempDir())
	sm := newTestManager(t, repo.Dir(), repo)

	coord := newScopedCoordinator(repo, sm)

	if coord == nil {
		t.Fatal("NewCoordinator returned nil")
	}

	if coord.CurrentView() != app.ViewDashboardList {
		t.Errorf("Expected initial dashboard list, got %v", coord.CurrentView())
	}

	if len(coord.Views()) != 12 {
		t.Errorf("Expected 12 views, got %d", len(coord.Views()))
	}
}

func TestCoordinator_Navigate(t *testing.T) {
	repo, _ := newCurrentRepository(t.TempDir())
	sm := newTestManager(t, repo.Dir(), repo)

	coord := newScopedCoordinator(repo, sm)
	coord.Init()

	openDefaultDashboard(coord)
	coord.Update(messages.ScopedNavigateMsg{View: messages.ViewCardEditor, DashboardID: "default", Generation: 1})

	if coord.CurrentView() != app.ViewCardEditor {
		t.Errorf("Expected ViewCardEditor, got %v", coord.CurrentView())
	}
	if coord.PrevView() != app.ViewDashboard {
		t.Errorf("Expected prevView ViewDashboard, got %v", coord.PrevView())
	}
}

func TestCoordinator_ResizeAll(t *testing.T) {
	repo, _ := newCurrentRepository(t.TempDir())
	sm := newTestManager(t, repo.Dir(), repo)

	coord := newScopedCoordinator(repo, sm)
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

func TestApp_CurrentConstructor(t *testing.T) {
	repo, _ := newCurrentRepository(t.TempDir())
	sm := newTestManager(t, repo.Dir(), repo)

	app, err := app.New(repo, sm, api.New(repo, sm))
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

// Cold-start initialization must disable Kitty keyboard reporting before the first update.
func TestInitDisablesKittyProtocolAtStartup(t *testing.T) {
	repo, _ := newCurrentRepository(t.TempDir())
	sm := newTestManager(t, repo.Dir(), repo)

	coord := newScopedCoordinator(repo, sm)
	cmd := coord.Init()
	if cmd == nil {
		t.Fatal("expected Coordinator.Init to return a command")
	}

	r, w, errPipe := os.Pipe()
	if errPipe != nil {
		t.Fatalf("pipe: %v", errPipe)
	}
	origStdout := os.Stdout
	os.Stdout = w //nolint:reassign // capturing stdout for the duration of this test only

	// Executing the returned Cmd tree (recursing into tea.BatchMsg, exactly
	// as Bubble Tea's runtime would) is what actually runs the leaf
	// disableKittyProtocolCmd inside Dashboard.Init - collectMsgs is used
	// purely for its recursive-execution behavior here, not its return
	// value. collectMsgs processes the batch in order and the dashboard's
	// Init is first, so the write we care about happens well before it
	// reaches tick() (last in the batch), which blocks for a real 2s
	// interval - run it in the background and don't wait for that, so this
	// test doesn't pay for it.
	done := make(chan struct{})
	go func() {
		_ = collectMsgs(t, cmd)
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(200 * time.Millisecond):
	}

	_ = w.Close()
	os.Stdout = origStdout //nolint:reassign // restoring stdout after capture above

	var buf bytes.Buffer
	if _, err := io.Copy(&buf, r); err != nil {
		t.Fatalf("read captured stdout: %v", err)
	}

	if !strings.Contains(buf.String(), "\x1b[=0u") {
		t.Fatalf("expected Coordinator.Init to disable the Kitty keyboard protocol at startup, got: %q", buf.String())
	}
}
