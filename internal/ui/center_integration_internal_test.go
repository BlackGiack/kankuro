package ui

import (
	"context"
	"testing"

	"github.com/BlackGiack/kankuro/internal/api"
	"github.com/BlackGiack/kankuro/internal/api/contracts"
	"github.com/BlackGiack/kankuro/internal/config"
	"github.com/BlackGiack/kankuro/internal/session"
	"github.com/google/uuid"
)

// TestCenterReportSurvivesBackendReconstruction covers the real history, service,
// and coordinator seam. It does not start the terminal UI.
func TestCenterReportSurvivesBackendReconstruction(t *testing.T) {
	_, _, repo, manager := newLaunchFocusCoordinatorWithRepo(t)
	workspace := t.TempDir()
	if err := repo.SaveWorkspaceSources([]config.WorkspaceSource{{Path: workspace}}); err != nil {
		t.Fatal(err)
	}
	workspaces, err := repo.ListWorkspaces()
	if err != nil || len(workspaces) != 1 {
		t.Fatalf("workspaces: %v, %v", workspaces, err)
	}
	card, err := repo.CreateIDCard(config.IDCard{DashboardID: "default", Name: "worker", Driver: "cat", WorkspaceID: workspaces[0].ID})
	if err != nil {
		t.Fatal(err)
	}
	service := api.New(repo, manager)
	id, err := service.LaunchInDashboard(context.Background(), "default", card.ID)
	if err != nil {
		t.Fatal(err)
	}
	message := session.MessageRecord{
		ID:                    uuid.NewString(),
		DashboardID:           "default",
		SenderSessionID:       "sender",
		SenderActivationID:    "sender",
		RecipientSessionID:    id,
		RecipientActivationID: "worker",
		Text:                  "task",
	}
	store := manager.Messages()
	if _, err = store.CreateIntent(message, uuid.NewString(), session.SendFingerprint(id, "worker", "task")); err != nil {
		t.Fatal(err)
	}
	if _, err = store.RecordReply(id, "worker", uuid.NewString(), message.ID, "progress", "persisted center report"); err != nil {
		t.Fatal(err)
	}
	verify := func(port centerPort) {
		t.Helper()
		c := centerTestCoordinator(t)
		c.center.port = port
		c.center.dashboardID = "default"
		c.center.sessionID = id
		c.center.data.SessionID = id
		c.center.data.Workspace = &contracts.SessionWorkspaceResult{WorkingDir: workspace}
		cmd := c.centerRead()
		if cmd == nil {
			t.Fatal("no center read")
		}
		c.handleCenter(cmd())
		if len(c.center.data.Reports) != 1 || c.center.data.Reports[0].Text != "persisted center report" || c.center.data.Workspace != nil {
			t.Fatalf("center = %+v", c.center.data)
		}
	}
	verify(service)
	manager.Shutdown()
	restarted := newTestManager(t, repo.Dir(), repo)
	verify(api.New(repo, restarted))
}
