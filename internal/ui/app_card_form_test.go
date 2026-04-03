package ui //nolint:testpackage // Regression tests need unexported UI helpers.

import (
	"kankuro/internal/config"
	"kankuro/internal/session"
	"testing"
)

func TestHandleDashboardEditMapsWorkspaceFormInputs(t *testing.T) {
	repo, err := config.NewFileRepository(t.TempDir())
	if err != nil {
		t.Fatalf("new repo: %v", err)
	}
	wd := t.TempDir()
	if saveErr := repo.SaveWorkspaceSources([]config.WorkspaceSource{{Path: wd, IsRoot: false}}); saveErr != nil {
		t.Fatalf("save workspace source: %v", saveErr)
	}
	workspaces, err := repo.ListWorkspaces()
	if err != nil {
		t.Fatalf("list workspaces: %v", err)
	}
	card, err := repo.CreateIDCard(config.IDCard{
		Name:            "Claude",
		Driver:          "claude",
		Model:           "opus",
		WorkspaceID:     workspaces[0].ID,
		FaceStyle:       "focused",
		Puppet:          "cat",
		UseTmpWorkspace: true,
	})
	if err != nil {
		t.Fatalf("create card: %v", err)
	}

	app, err := New(repo, session.NewManager(repo.Dir(), repo))
	if err != nil {
		t.Fatalf("new app: %v", err)
	}
	app.cards = []config.IDCard{card}
	app.refreshCardList()
	app.focusedPane = paneCards

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("handleDashboardEdit panicked: %v", r)
		}
	}()

	if _, _ = app.handleDashboardEdit(); app.mode != viewCardForm {
		t.Fatalf("expected card form mode, got %v", app.mode)
	}
	if got := app.cardInputs[3].Value(); got != "focused" {
		t.Fatalf("expected face style input to be populated, got %q", got)
	}
	if got := app.cardInputs[4].Value(); got != "y" {
		t.Fatalf("expected tmp workspace input to be populated, got %q", got)
	}
	if app.selectedWorkspaceID != workspaces[0].ID {
		t.Fatalf("expected workspace selection %q, got %q", workspaces[0].ID, app.selectedWorkspaceID)
	}
}

func TestResetCardFormClearsWorkspaceSelection(t *testing.T) {
	app := &App{}
	app.initInputs()
	app.selectedWorkspaceID = "workspace-1"
	app.workspaceCursor = 2
	app.workspaceSearchQuery = "demo"

	app.resetCardForm()

	if app.selectedWorkspaceID != "" {
		t.Fatalf("expected workspace selection to be cleared, got %q", app.selectedWorkspaceID)
	}
	if app.workspaceCursor != 0 {
		t.Fatalf("expected workspace cursor reset, got %d", app.workspaceCursor)
	}
	if app.workspaceSearchQuery != "" {
		t.Fatalf("expected workspace search query reset, got %q", app.workspaceSearchQuery)
	}
}
