package app_test

import (
	"testing"

	"github.com/BlackGiack/kankuro/internal/app"
	"github.com/BlackGiack/kankuro/internal/config"
	"github.com/BlackGiack/kankuro/internal/ui/messages"
)

// TestOnDeleteCardRefusesWhenSessionReferencesCard guards against the
// DeleteIDCard finding: config.DeleteIDCard has no reference check (it
// can't see sessions), so the coordinator's OnDeleteCard callback must
// refuse to delete a card a persisted session still points at - otherwise
// that session's relaunch/switch breaks with a bare "card not found".
func TestOnDeleteCardRefusesWhenSessionReferencesCard(t *testing.T) {
	repo, err := config.NewFileRepository(t.TempDir())
	if err != nil {
		t.Fatalf("new repo: %v", err)
	}
	sm := newTestManager(t, repo.Dir(), repo)

	wd := t.TempDir()
	if saveErr := repo.SaveWorkspaceSources([]config.WorkspaceSource{{Path: wd}}); saveErr != nil {
		t.Fatalf("save workspace source: %v", saveErr)
	}
	workspaces, err := repo.ListWorkspaces()
	if err != nil || len(workspaces) == 0 {
		t.Fatalf("list workspaces: %v (%v)", err, workspaces)
	}

	card, err := repo.CreateIDCard(config.IDCard{
		Name:        "Referenced Card",
		Driver:      "echo",
		WorkspaceID: workspaces[0].ID,
	})
	if err != nil {
		t.Fatalf("create card: %v", err)
	}

	sessionID, err := sm.Launch(card, nil, nil, nil)
	if err != nil {
		t.Fatalf("launch: %v", err)
	}
	t.Cleanup(func() { _ = sm.Kill(sessionID) })

	coord := app.NewCoordinator(repo, sm)
	coord.Init()

	cmd, err := coord.DashboardCallbacks().OnDeleteCard(card.ID)
	if err != nil {
		t.Fatalf("OnDeleteCard returned error building cmd: %v", err)
	}
	msg := cmd()
	if _, ok := msg.(messages.ErrorMsg); !ok {
		t.Fatalf("expected ErrorMsg refusing the delete, got %#v", msg)
	}

	cards, err := repo.ListIDCards()
	if err != nil {
		t.Fatalf("list cards: %v", err)
	}
	found := false
	for _, c := range cards {
		if c.ID == card.ID {
			found = true
		}
	}
	if !found {
		t.Fatal("card was deleted despite being referenced by a live session")
	}
}

// TestOnDeleteCardAllowsWhenNoSessionReferencesCard is the control case: an
// unreferenced card must still delete normally.
func TestOnDeleteCardAllowsWhenNoSessionReferencesCard(t *testing.T) {
	repo, err := config.NewFileRepository(t.TempDir())
	if err != nil {
		t.Fatalf("new repo: %v", err)
	}
	sm := newTestManager(t, repo.Dir(), repo)

	wd := t.TempDir()
	if saveErr := repo.SaveWorkspaceSources([]config.WorkspaceSource{{Path: wd}}); saveErr != nil {
		t.Fatalf("save workspace source: %v", saveErr)
	}
	workspaces, err := repo.ListWorkspaces()
	if err != nil || len(workspaces) == 0 {
		t.Fatalf("list workspaces: %v (%v)", err, workspaces)
	}

	card, err := repo.CreateIDCard(config.IDCard{
		Name:        "Unreferenced Card",
		Driver:      "echo",
		WorkspaceID: workspaces[0].ID,
	})
	if err != nil {
		t.Fatalf("create card: %v", err)
	}

	coord := app.NewCoordinator(repo, sm)
	coord.Init()

	cmd, err := coord.DashboardCallbacks().OnDeleteCard(card.ID)
	if err != nil {
		t.Fatalf("OnDeleteCard returned error building cmd: %v", err)
	}
	_ = cmd()

	cards, err := repo.ListIDCards()
	if err != nil {
		t.Fatalf("list cards: %v", err)
	}
	for _, c := range cards {
		if c.ID == card.ID {
			t.Fatal("expected unreferenced card to be deleted")
		}
	}
}
