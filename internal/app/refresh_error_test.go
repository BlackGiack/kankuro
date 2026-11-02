package app_test

import (
	"os"
	"path/filepath"
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/BlackGiack/kankuro/internal/api"
	"github.com/BlackGiack/kankuro/internal/app"
	"github.com/BlackGiack/kankuro/internal/config"
	"github.com/BlackGiack/kankuro/internal/configstartup"
	"github.com/BlackGiack/kankuro/internal/ui/messages"
	"github.com/BlackGiack/kankuro/internal/ui/views"
)

// A failed registry read must surface an error while preserving previously loaded cards.
func TestRefreshDataSurfacesErrorAndKeepsPreviousCards(t *testing.T) {
	root := t.TempDir()
	lock, errLock := configstartup.AcquireLock(root)
	if errLock != nil {
		t.Fatal(errLock)
	}
	t.Cleanup(func() {
		if err := lock.Close(); err != nil {
			t.Error(err)
		}
	})
	if err := configstartup.PrepareStartup(root); err != nil {
		t.Fatal(err)
	}
	repo, errRepo := config.NewFileRepository(root)
	if errRepo != nil {
		t.Fatalf("new repo: %v", errRepo)
	}
	sm := newTestManager(t, repo.Dir(), repo)

	wd := t.TempDir()
	_ = repo.SaveWorkspaceSources([]config.WorkspaceSource{{Path: wd, IsRoot: false}})
	workspaces, _ := repo.ListWorkspaces()

	card, errCreate := repo.CreateIDCard(config.IDCard{
		Name:        "Good Card",
		Driver:      "cat",
		WorkspaceID: workspaces[0].ID,
		DashboardID: "default",
	})
	if errCreate != nil {
		t.Fatalf("create card: %v", errCreate)
	}

	coord := app.NewCoordinator(repo, sm, api.New(repo, sm))
	coord.Init()
	coord.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	coord.Update(
		messages.DashboardOpenRequestedMsg{Dashboard: config.Dashboard{ID: "default", Name: "Default", State: config.DashboardActive}},
	)

	// First refresh: the card is loaded cleanly.
	coord.Update(coord.RefreshData()())
	dashboard, ok := coord.Views()[app.ViewDashboard].(*views.Dashboard)
	if !ok {
		t.Fatal("expected *views.Dashboard for ViewDashboard")
	}
	if got := dashboard.Cards(); len(got) != 1 || got[0].ID != card.ID {
		t.Fatalf("expected the dashboard to have loaded the card, got: %v", got)
	}
	if coord.Err() != nil {
		t.Fatalf("expected no error after a clean refresh, got: %v", coord.Err())
	}

	// Corrupt id_cards.yaml so the next ListIDCards() call fails.
	idCardsPath := filepath.Join(repo.Dir(), config.IDCardsFilename)
	if err := os.WriteFile(idCardsPath, []byte("not: valid: yaml: at: all: :::"), 0o600); err != nil {
		t.Fatalf("corrupt id_cards.yaml: %v", err)
	}

	msg, ok := coord.RefreshData()().(messages.DataLoadedMsg)
	if !ok {
		t.Fatal("expected refreshData to produce a DataLoadedMsg")
	}
	if msg.Err == nil {
		t.Fatal("expected refreshData to surface the corrupt id_cards.yaml read error")
	}
	if msg.Cards != nil {
		t.Fatalf("expected Cards to be nil on a read error (not an emptied slice), got: %v", msg.Cards)
	}

	coord.Update(msg)

	if coord.Err() == nil {
		t.Fatal("expected coord.Err() to carry the refresh error")
	}
	if got := dashboard.Cards(); len(got) != 1 || got[0].ID != card.ID {
		t.Fatalf("expected the dashboard to keep its previously loaded card after a failed refresh, got: %v", got)
	}
}
