package app

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/BlackGiack/kankuro/internal/api"
	"github.com/BlackGiack/kankuro/internal/config"
	"github.com/BlackGiack/kankuro/internal/layout"
	"github.com/BlackGiack/kankuro/internal/ui/messages"
	"github.com/BlackGiack/kankuro/internal/ui/views"
)

//nolint:govet // Setup errors are checked at the point of use.
func TestMultiDashboardDiscardsStaleResultsAndPreservesViews(t *testing.T) {
	root := t.TempDir()
	if err := layout.PrepareStartup(root); err != nil {
		t.Fatal(err)
	}
	repo, err := config.NewFileRepository(root)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(
		filepath.Join(repo.Dir(), config.DashboardsFilename),
		[]byte("schema_version: 1\ndashboards:\n  - {id: a, name: A, state: active}\n  - {id: b, name: B, state: active}\n"),
		0o600,
	); err != nil {
		t.Fatal(err)
	}
	if err := repo.SaveWorkspaceSources([]config.WorkspaceSource{{Path: t.TempDir()}}); err != nil {
		t.Fatal(err)
	}
	workspaces, err := repo.ListWorkspaces()
	if err != nil || len(workspaces) != 1 {
		t.Fatalf("workspaces: %v, %v", workspaces, err)
	}
	cardA, err := repo.CreateIDCard(config.IDCard{Name: "A card", Driver: "cat", WorkspaceID: workspaces[0].ID, DashboardID: "a"})
	if err != nil {
		t.Fatal(err)
	}
	cardB, err := repo.CreateIDCard(config.IDCard{Name: "B card", Driver: "cat", WorkspaceID: workspaces[0].ID, DashboardID: "b"})
	if err != nil {
		t.Fatal(err)
	}
	manager := newTestManager(t, repo.Dir(), repo)
	service := api.New(repo, manager)
	idA, err := service.LaunchInDashboard(context.Background(), "a", cardA.ID)
	if err != nil {
		t.Fatal(err)
	}
	idB, err := service.LaunchInDashboard(context.Background(), "b", cardB.ID)
	if err != nil {
		t.Fatal(err)
	}
	c := NewCoordinator(repo, manager, service)
	c.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	c.Update(
		messages.DashboardOpenRequestedMsg{
			Dashboard:  config.Dashboard{ID: "a", Name: "A", State: config.DashboardActive},
			Generation: c.viewGeneration,
		},
	)
	viewA := c.views[ViewDashboard]
	c.Update(messages.SessionPreviewRequestedMsg{SessionID: idB})
	if c.preview.current() == idB {
		t.Fatal("foreign session preview was subscribed")
	}
	stale := c.refreshData()
	stalePuppetNavigation := c.settingsCallbacks().OnOpenPuppetSources()
	oldGeneration := c.viewGeneration
	c.Update(messages.DashboardListRequestedMsg{DashboardID: c.activeDashboard, Generation: c.viewGeneration})
	c.Update(
		messages.DashboardOpenRequestedMsg{
			Dashboard:  config.Dashboard{ID: "b", Name: "B", State: config.DashboardActive},
			Generation: c.viewGeneration,
		},
	)
	current := c.refreshData()
	c.Update(current())
	c.Update(stale())
	c.Update(stalePuppetNavigation())
	if c.currentView != ViewDashboard {
		t.Fatal("stale settings navigation crossed dashboard context")
	}
	dashboard := c.views[ViewDashboard].(*views.Dashboard)
	if len(dashboard.Cards()) != 1 || dashboard.Cards()[0].ID != cardB.ID || len(dashboard.Sessions()) != 1 ||
		dashboard.Sessions()[0].DashboardID != "b" {
		t.Fatalf("stale refresh crossed scope: cards=%+v sessions=%+v", dashboard.Cards(), dashboard.Sessions())
	}
	c.Update(messages.SessionAttachRequestedMsg{SessionID: idA, DashboardID: "a", Generation: oldGeneration})
	c.Update(messages.ScopedNavigateMsg{View: messages.ViewSessionLauncher, DashboardID: "a", Generation: oldGeneration})
	if c.currentView != ViewDashboard || c.activeDashboard != "b" {
		t.Fatal("stale attach changed dashboard")
	}
	c.Update(messages.DashboardListRequestedMsg{DashboardID: c.activeDashboard, Generation: c.viewGeneration})
	c.Update(
		messages.DashboardOpenRequestedMsg{
			Dashboard:  config.Dashboard{ID: "a", Name: "A", State: config.DashboardActive},
			Generation: c.viewGeneration,
		},
	)
	if c.views[ViewDashboard] != viewA {
		t.Fatal("dashboard context was not restored")
	}
}
