package app

import (
	"bytes"
	"context"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/BlackGiack/kankuro/internal/api"
	"github.com/BlackGiack/kankuro/internal/api/contracts"
	"github.com/BlackGiack/kankuro/internal/config"
	"github.com/BlackGiack/kankuro/internal/configstartup"
	"github.com/BlackGiack/kankuro/internal/session"
	"github.com/BlackGiack/kankuro/internal/ui/messages"
	"github.com/BlackGiack/kankuro/internal/ui/views"
)

func TestRefreshReconcilesMissedPreviewOutput(t *testing.T) {
	for _, refreshKind := range []string{"sessions poll", "attach detach full refresh"} {
		t.Run(refreshKind, func(t *testing.T) { assertMissedPreviewReconciles(t, refreshKind) })
	}
}

func assertMissedPreviewReconciles(t *testing.T, refreshKind string) {
	t.Helper()
	coordinator, manager, sessionID := newPreviewRecoveryFixture(t)
	dashboard := coordinator.views[ViewDashboard].(*views.Dashboard)
	assertPreviewSubscription(t, coordinator, sessionID)
	marker := []byte("INDEPENDENT_PREVIEW_RECOVERY_MARKER")
	writeRetainedOutput(t, manager, sessionID, marker)
	discardPreviewSubscriberOutput(coordinator)
	if bytes.Contains(dashboard.Preview().Scrollback(), marker) {
		t.Fatal("test setup failed: preview received the deliberately discarded marker")
	}
	applyPreviewRecoveryRefresh(t, coordinator, refreshKind, sessionID)
	if dashboard.PreviewSessionID() != sessionID {
		t.Fatalf("refresh changed preview selection: %q", dashboard.PreviewSessionID())
	}
	if !bytes.Contains(dashboard.Preview().Scrollback(), marker) {
		t.Fatalf("retained scrollback did not repair lost output: %q", dashboard.Preview().Scrollback())
	}
}

func assertPreviewSubscription(t *testing.T, coordinator *Coordinator, sessionID string) {
	t.Helper()
	if dashboard := coordinator.views[ViewDashboard].(*views.Dashboard); dashboard.PreviewSessionID() != sessionID ||
		coordinator.preview.current() != sessionID {
		t.Fatalf("session was not selected and subscribed: dashboard=%q coordinator=%q",
			dashboard.PreviewSessionID(), coordinator.preview.current())
	}
}

func writeRetainedOutput(t *testing.T, manager *session.Manager, sessionID string, marker []byte) {
	t.Helper()
	handle, err := manager.AttachChecked(sessionID, func([]session.Snapshot) error { return nil })
	if err != nil {
		t.Fatalf("attach: %v", err)
	}
	if _, err = handle.Input.Write(marker); err != nil {
		handle.Close()
		t.Fatalf("write marker: %v", err)
	}
	handle.Close()
	waitForScrollback(t, manager, sessionID, marker)
}

func applyPreviewRecoveryRefresh(t *testing.T, coordinator *Coordinator, refreshKind, sessionID string) {
	t.Helper()
	if refreshKind == "sessions poll" {
		coordinator.Update(coordinator.refreshSessions()())
		return
	}
	_, cmd := coordinator.Update(messages.AttachFinishedMsg{
		SessionID: sessionID, DashboardID: "default", Generation: coordinator.viewGeneration,
	})
	if cmd == nil {
		t.Fatal("attach detach did not schedule its full refresh")
	}
	coordinator.Update(cmd())
}

func newPreviewRecoveryFixture(t *testing.T) (*Coordinator, *session.Manager, string) {
	t.Helper()
	root := t.TempDir()
	if err := configstartup.PrepareStartup(root); err != nil {
		t.Fatal(err)
	}
	repo, err := config.NewFileRepository(root)
	if err != nil {
		t.Fatal(err)
	}
	workspace := t.TempDir()
	if err = repo.SaveWorkspaceSources([]config.WorkspaceSource{{Path: workspace}}); err != nil {
		t.Fatal(err)
	}
	workspaces, err := repo.ListWorkspaces()
	if err != nil || len(workspaces) != 1 {
		t.Fatalf("workspaces: %v, %v", workspaces, err)
	}
	card, err := repo.CreateIDCard(config.IDCard{
		DashboardID: "default", Name: "Preview recovery", Driver: "cat", WorkspaceID: workspaces[0].ID,
	})
	if err != nil {
		t.Fatal(err)
	}
	manager := newTestManager(t, repo.Dir(), repo)
	sessionID, err := api.New(repo, manager).LaunchInDashboard(context.Background(), "default", card.ID)
	if err != nil {
		t.Fatal(err)
	}
	coordinator := NewCoordinator(manager, api.New(repo, manager))
	coordinator.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	coordinator.Update(messages.DashboardOpenRequestedMsg{
		DashboardID: "default", DashboardName: "Default", DashboardState: contracts.DashboardActive,
		Generation: coordinator.viewGeneration,
	})
	coordinator.Update(coordinator.refreshDataSeeded()())
	return coordinator, manager, sessionID
}

func waitForScrollback(t *testing.T, manager *session.Manager, sessionID string, marker []byte) {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		if bytes.Contains(manager.GetScrollback(sessionID), marker) {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("manager scrollback did not receive marker; scrollback=%q", manager.GetScrollback(sessionID))
}

func discardPreviewSubscriberOutput(coordinator *Coordinator) {
	for coordinator.preview.ch != nil {
		select {
		case <-coordinator.preview.ch:
		default:
			return
		}
	}
}
