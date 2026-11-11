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
	"github.com/BlackGiack/kankuro/internal/config"
	"github.com/BlackGiack/kankuro/internal/session"
	"github.com/BlackGiack/kankuro/internal/ui/messages"
	"github.com/BlackGiack/kankuro/internal/ui/views"
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

func TestCoordinator_DashboardRenameFailureStaysInDialogAndIgnoresStaleResult(t *testing.T) {
	repo, _ := newCurrentRepository(t.TempDir())
	sm := newTestManager(t, repo.Dir(), repo)
	coord := newScopedCoordinator(repo, sm)
	openDefaultDashboard(coord)
	dashboard, ok := coord.Views()[app.ViewDashboard].(*views.Dashboard)
	if !ok {
		t.Fatal("dashboard view unavailable")
	}
	dashboard.SetData(messages.DataLoadedMsg{Sessions: []session.Snapshot{{ID: "missing-session", Name: "Session"}}})
	dashboard.TogglePane()
	dashboard.SessionList().SetSelectedByID("missing-session")
	dashboard.Update(tea.KeyPressMsg{Text: "e"})
	_, cmd := dashboard.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("rename did not start")
	}
	result, ok := cmd().(messages.DashboardSessionRenamedMsg)
	if !ok {
		t.Fatal("rename command returned unexpected message")
	}
	stale := result
	stale.Generation++
	coord.Update(stale)
	if !strings.Contains(dashboard.View(), "Saving") {
		t.Fatal("stale rename result modified the pending dialog")
	}
	coord.Update(result)
	if coord.CurrentView() != app.ViewDashboard {
		t.Fatalf("rename failure navigated to view %v", coord.CurrentView())
	}
	if !strings.Contains(dashboard.View(), "Save failed") {
		t.Fatal("rename failure was not retained in the dialog")
	}
}

func newRunningRenameSession(t *testing.T) (*config.FileRepository, *session.Manager, string) {
	t.Helper()
	repo, err := newCurrentRepository(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	sm := newTestManager(t, repo.Dir(), repo)
	if err = repo.SaveWorkspaceSources([]config.WorkspaceSource{{Path: t.TempDir()}}); err != nil {
		t.Fatal(err)
	}
	workspaces, err := repo.ListWorkspaces()
	if err != nil || len(workspaces) != 1 {
		t.Fatalf("workspace discovery: %v, %v", workspaces, err)
	}
	card, err := createCurrentCard(repo, config.IDCard{
		Name: "Rename fixture", Driver: "cat", WorkspaceID: workspaces[0].ID,
	})
	if err != nil {
		t.Fatal(err)
	}
	sessionID, err := launchCurrentSession(repo, sm, card)
	if err != nil {
		t.Fatal(err)
	}
	return repo, sm, sessionID
}

func TestCoordinator_DashboardRenameRefreshesWithoutAttachAndSurvivesRestart(t *testing.T) {
	repo, sm, sessionID := newRunningRenameSession(t)
	before, ok := sm.GetSession(sessionID)
	if !ok || before.Status == session.StatusTerminated || before.PID <= 0 || !before.ProcessStarted {
		t.Fatalf("expected a live activation before rename: %+v", before)
	}
	coord := newOpenedCoordinator(repo, sm)
	coord.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	coord.Update(coord.RefreshData()())
	dashboard := coord.Views()[app.ViewDashboard].(*views.Dashboard)
	coord.Update(tea.KeyPressMsg{Code: tea.KeyTab})
	dashboard.SessionList().SetSelectedByID(sessionID)
	coord.Update(tea.KeyPressMsg{Text: "e"})
	if !strings.Contains(dashboard.View(), "Rename Session") {
		t.Fatal("session edit did not open the rename dialog")
	}
	// Replace the existing name through the same text input events as the UI.
	coord.Update(tea.KeyPressMsg{Code: 'a', Mod: tea.ModCtrl})
	coord.Update(tea.KeyPressMsg{Code: 'k', Mod: tea.ModCtrl})
	const newName = "Renamed 日本語 session"
	coord.Update(tea.KeyPressMsg{Text: newName})
	_, submit := coord.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if submit == nil {
		t.Fatal("rename did not return a save command")
	}
	result, ok := submit().(messages.DashboardSessionRenamedMsg)
	if !ok || result.Err != nil || result.SessionID != sessionID {
		t.Fatalf("rename result: %+v", result)
	}
	_, refresh := coord.Update(result)
	if strings.Contains(dashboard.View(), "Rename Session") {
		t.Fatal("successful rename did not close the dialog")
	}
	// A rename must return only the data refresh, never the launch naming
	// flow's attachment or navigation commands. Do not execute preview waits.
	updates := collectMsgs(t, refresh)
	for _, update := range updates {
		if attach, isAttach := update.(messages.ScopedAttachReadyMsg); isAttach && attach.Handle != nil {
			attach.Handle.Close()
			t.Fatal("rename requested terminal attachment")
		}
	}
	if len(updates) != 1 {
		t.Fatalf("rename returned %d commands, want only a data refresh", len(updates))
	}
	coord.Update(updates[0])
	if coord.CurrentView() != app.ViewDashboard || coord.Err() != nil {
		t.Fatalf("rename changed view or reported error: view=%v error=%v", coord.CurrentView(), coord.Err())
	}
	selected, ok := dashboard.SessionList().Selected()
	if !ok || selected.ID != sessionID || selected.Name != newName {
		t.Fatalf("dashboard did not refresh the selected name: %+v", selected)
	}
	after, ok := sm.GetSession(sessionID)
	if !ok || after.PID != before.PID || after.ActivationID != before.ActivationID || after.Status == session.StatusTerminated {
		t.Fatalf("rename replaced or stopped the activation: before=%+v after=%+v", before, after)
	}
	if err := sm.Kill(sessionID); err != nil {
		t.Fatalf("stop renamed session: %v", err)
	}
	sm.Shutdown()
	checkRenamedSessionAfterRestart(t, repo.Dir(), sessionID, newName)
}

func checkRenamedSessionAfterRestart(t *testing.T, dir, sessionID, name string) {
	t.Helper()
	restoredRepo, err := newCurrentRepository(dir)
	if err != nil {
		t.Fatal(err)
	}
	restoredManager := newTestManager(t, restoredRepo.Dir(), restoredRepo)
	if err = restoredManager.LoadErrors(); err != nil {
		t.Fatal(err)
	}
	restoredCoord := newOpenedCoordinator(restoredRepo, restoredManager)
	restoredCoord.Update(restoredCoord.RefreshData()())
	restoredDashboard := restoredCoord.Views()[app.ViewDashboard].(*views.Dashboard)
	if got := restoredDashboard.Sessions(); len(got) != 1 || got[0].ID != sessionID || got[0].Name != name {
		t.Fatalf("rename did not survive restart: %+v", got)
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
