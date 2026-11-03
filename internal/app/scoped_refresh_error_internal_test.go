package app

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/BlackGiack/kankuro/internal/api"
	"github.com/BlackGiack/kankuro/internal/config"
	"github.com/BlackGiack/kankuro/internal/configstartup"
	"github.com/BlackGiack/kankuro/internal/ui/messages"
	"github.com/BlackGiack/kankuro/internal/ui/views"
)

//nolint:gocognit // Exercise all refresh producers through the same failure and recovery scenario.
func TestScopedRefreshPreservesDataOnInventoryFailure(t *testing.T) {
	for _, kind := range []string{"full", "sessions", "config"} {
		t.Run(kind, func(t *testing.T) {
			root := t.TempDir()
			if err := configstartup.PrepareStartup(root); err != nil {
				t.Fatal(err)
			}
			repo, err := config.NewFileRepository(root)
			if err != nil {
				t.Fatal(err)
			}
			if err = os.WriteFile(
				filepath.Join(repo.Dir(), config.DashboardsFilename),
				[]byte("schema_version: 1\ndashboards:\n  - {id: a, name: A, state: active}\n"),
				0600,
			); err != nil {
				t.Fatal(err)
			}
			if err = repo.SaveWorkspaceSources([]config.WorkspaceSource{{Path: t.TempDir()}}); err != nil {
				t.Fatal(err)
			}
			workspaces, err := repo.ListWorkspaces()
			if err != nil || len(workspaces) != 1 {
				t.Fatalf("workspaces: %v %v", workspaces, err)
			}
			card, err := repo.CreateIDCard(config.IDCard{Name: "card", Driver: "cat", WorkspaceID: workspaces[0].ID, DashboardID: "a"})
			if err != nil {
				t.Fatal(err)
			}
			manager := newTestManager(t, repo.Dir(), repo)
			service := api.New(repo, manager)
			id, err := service.LaunchInDashboard(context.Background(), "a", card.ID)
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
			c.Update(c.refreshData()())
			dashboard := c.views[ViewDashboard].(*views.Dashboard)
			if len(dashboard.Cards()) != 1 || len(dashboard.Sessions()) != 1 {
				t.Fatalf("initial cards/sessions: %v %v", dashboard.Cards(), dashboard.Sessions())
			}
			path := filepath.Join(repo.Dir(), config.IDCardsFilename)
			raw, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			if err = os.WriteFile(path, []byte("invalid: ["), 0600); err != nil {
				t.Fatal(err)
			}
			var cmd tea.Cmd
			switch kind {
			case "full":
				cmd = c.refreshData()
			case "sessions":
				cmd = c.refreshSessions()
			default:
				cmd = c.maybeRefreshConfig()
			}
			c.Update(cmd())
			if c.err == nil || len(dashboard.Cards()) != 1 || dashboard.Cards()[0].ID != card.ID || len(dashboard.Sessions()) != 1 ||
				dashboard.Sessions()[0].ID != id {
				t.Fatalf("failure cleared data: error=%v cards=%v sessions=%v", c.err, dashboard.Cards(), dashboard.Sessions())
			}
			if _, err = service.LaunchInDashboard(context.Background(), "a", card.ID); err == nil {
				t.Fatal("corrupt inventory allowed launch")
			}
			if err = os.WriteFile(path, raw, 0600); err != nil {
				t.Fatal(err)
			}
			c.Update(c.refreshData()())
			if len(dashboard.Cards()) != 1 || len(dashboard.Sessions()) != 1 {
				t.Fatal("data did not recover")
			}
		})
	}
}

func TestScopedRefreshReturnsExplicitEmptyInventory(t *testing.T) {
	root := t.TempDir()
	if err := configstartup.PrepareStartup(root); err != nil {
		t.Fatal(err)
	}
	repo, err := config.NewFileRepository(root)
	if err != nil {
		t.Fatal(err)
	}
	manager := newTestManager(t, repo.Dir(), repo)
	service := api.New(repo, manager)
	r := refresher{repo: repo, sessions: manager}
	full := unwrapRefresh(r.scopedFull(service, "default", 1, "")()).(messages.DataLoadedMsg)
	if full.Err != nil || full.Cards == nil || full.Sessions == nil || len(full.Cards) != 0 || len(full.Sessions) != 0 {
		t.Fatalf("empty full refresh: %+v", full)
	}
	tick := r.scopedSessions(service, "default", 1, "")().(messages.SessionsRefreshedMsg)
	if tick.Err != nil || tick.Cards == nil || tick.Sessions == nil || len(tick.Cards) != 0 || len(tick.Sessions) != 0 {
		t.Fatalf("empty session refresh: %+v", tick)
	}
	changed := r.scopedConfigIfChanged(service, "default", 1, nil)().(configFingerprintMsg)
	if changed.data == nil || changed.data.Err != nil || changed.data.Cards == nil || len(changed.data.Cards) != 0 {
		t.Fatalf("empty config refresh: %+v", changed)
	}
}
