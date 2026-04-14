package ui //nolint:testpackage // Access unexported App helpers.

import (
	"kankuro/internal/config"
	"kankuro/internal/session"
	"testing"
)

func TestHandleDashboardDeleteSetsPendingCardID(t *testing.T) {
	repo, _ := config.NewFileRepository(t.TempDir())
	sm := session.NewManager(repo.Dir(), repo)
	app, _ := New(repo, sm)

	card := config.IDCard{ID: "card-1", Name: "Test Card", Driver: "claude", Model: "opus"}
	app.cards = []config.IDCard{card}
	app.refreshCardList()
	app.sessionView.FocusedPane = paneCards
	app.cardsList.Select(0)

	app.handleDashboardDelete()

	if app.mode != viewDeleteConfirm {
		t.Errorf("expected mode viewDeleteConfirm, got %v", app.mode)
	}
	if app.sessionView.PendingDeleteCardID != card.ID {
		t.Errorf("expected pending delete card ID %s, got %s", card.ID, app.sessionView.PendingDeleteCardID)
	}
}

func TestHandleDeleteConfirmYesRemovesCard(t *testing.T) {
	repo, _ := config.NewFileRepository(t.TempDir())
	sm := session.NewManager(repo.Dir(), repo)

	wd := t.TempDir()
	repo.SaveWorkspaceSources([]config.WorkspaceSource{{Path: wd, IsRoot: false}})
	workspaces, _ := repo.ListWorkspaces()

	card, err := repo.CreateIDCard(config.IDCard{
		Name:        "To Delete",
		Driver:      "claude",
		Model:       "opus",
		WorkspaceID: workspaces[0].ID,
	})
	if err != nil {
		t.Fatalf("create card failed: %v", err)
	}
	app, _ := New(repo, sm)
	app.cards = []config.IDCard{card}
	app.refreshCardList()

	app.sessionView.PendingDeleteCardID = card.ID
	app.mode = viewDeleteConfirm

	app.handleDeleteConfirmYes()

	if app.err != nil {
		t.Errorf("expected no error, got %v", app.err)
	}

	if app.mode != viewDashboard {
		t.Errorf("expected mode viewDashboard, got %v", app.mode)
	}
	if app.sessionView.PendingDeleteCardID != "" {
		t.Error("expected PendingDeleteCardID to be cleared")
	}

	// Verify it's gone from repo
	cards, _ := repo.ListIDCards()
	if len(cards) != 0 {
		t.Errorf("expected 0 cards in repo, got %d", len(cards))
	}
	// Verify it's gone from app.cards
	if len(app.cards) != 0 {
		t.Errorf("expected 0 cards in app, got %d", len(app.cards))
	}
}

func TestDashboardSearchFiltersCards(t *testing.T) {
	repo, _ := config.NewFileRepository(t.TempDir())
	sm := session.NewManager(repo.Dir(), repo)
	app, _ := New(repo, sm)

	app.cards = []config.IDCard{
		{ID: "1", Name: "Alpha"},
		{ID: "2", Name: "Beta"},
		{ID: "3", Name: "Gamma"},
	}
	app.sessionView.FocusedPane = paneCards
	app.sessionView.CardsSearchQuery = "al"
	app.refreshCardList()

	items := app.cardsList.Items()
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
	if items[0].(cardItem).card.Name != "Alpha" {
		t.Errorf("expected Alpha, got %s", items[0].(cardItem).card.Name)
	}
}
