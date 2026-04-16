package ui //nolint:testpackage // Access unexported App helpers for UI testing.

import (
	"kankuro/internal/config"
	"kankuro/internal/session"
	"testing"
)

func setupPreviewTest(t *testing.T) (*App, config.Repository, *session.Manager) {
	repo, _ := config.NewFileRepository(t.TempDir())
	wd := t.TempDir()
	repo.SaveWorkspaceSources([]config.WorkspaceSource{{Path: wd, IsRoot: false}})
	_ = repo.SyncWorkspacesFromSources()

	sm := session.NewManager(repo.Dir(), repo)
	app, _ := New(repo, sm)
	app.width = 100
	app.height = 100
	app.resize()
	return app, repo, sm
}

func TestPreviewInitialization(t *testing.T) {
	app, repo, sm := setupPreviewTest(t)

	workspaces, _ := repo.ListWorkspaces()
	// Create a dummy session
	card := config.IDCard{ID: "c1", Name: "Card 1", Driver: "echo", Model: "m1", WorkspaceID: workspaces[0].ID}
	_, _ = repo.CreateIDCard(card)
	id, err := sm.Launch(card, nil, nil, nil)
	if err != nil {
		t.Fatalf("launch failed: %v", err)
	}
	defer func() { _ = sm.Kill(id) }()

	// Refresh to pick up session
	app.sessions = sm.List()
	app.refreshSessionList()

	if app.sessionView.PreviewSessionID != id {
		t.Errorf("expected preview ID %s, got %s", id, app.sessionView.PreviewSessionID)
	}
}

func TestPreviewPersistsOnPaneSwitch(t *testing.T) {
	app, repo, sm := setupPreviewTest(t)
	workspaces, _ := repo.ListWorkspaces()

	card := config.IDCard{ID: "c1", Name: "Card 1", Driver: "echo", Model: "m1", WorkspaceID: workspaces[0].ID}
	_, _ = repo.CreateIDCard(card)
	id, _ := sm.Launch(card, nil, nil, nil)
	defer func() { _ = sm.Kill(id) }()

	app.sessions = sm.List()
	app.refreshSessionList()

	if app.sessionView.PreviewSessionID != id {
		t.Fatalf("setup: expected preview ID %s, got %s", id, app.sessionView.PreviewSessionID)
	}

	// Switch to cards pane
	app.sessionView.FocusedPane = paneCards
	app.syncPreviewSelection()

	if app.sessionView.PreviewSessionID != id {
		t.Errorf("preview should persist when switching to cards pane, got %s", app.sessionView.PreviewSessionID)
	}

	// Switch back to sessions pane
	app.sessionView.FocusedPane = paneSessions
	app.syncPreviewSelection()

	if app.sessionView.PreviewSessionID != id {
		t.Errorf("preview should remain %s in sessions pane, got %s", id, app.sessionView.PreviewSessionID)
	}
}

func TestPreviewFilteredList(t *testing.T) {
	app, repo, sm := setupPreviewTest(t)
	workspaces, _ := repo.ListWorkspaces()

	c1 := config.IDCard{ID: "c1", Name: "Alpha", Driver: "echo", WorkspaceID: workspaces[0].ID}
	c2 := config.IDCard{ID: "c2", Name: "Beta", Driver: "echo", WorkspaceID: workspaces[0].ID}
	_, _ = repo.CreateIDCard(c1)
	_, _ = repo.CreateIDCard(c2)

	id1, _ := sm.Launch(c1, nil, nil, nil)
	id2, _ := sm.Launch(c2, nil, nil, nil)
	defer func() { _ = sm.Kill(id1) }()
	defer func() { _ = sm.Kill(id2) }()

	app.sessions = sm.List()
	app.refreshSessionList()

	// Filter for Beta
	app.sessionView.SessionsSearchQuery = "beta"
	app.refreshSessionList()
	app.sessionView.FocusedPane = paneSessions
	app.syncPreviewSelection()

	if app.sessionView.PreviewSessionID != id2 {
		t.Errorf("expected preview ID %s (Beta), got %s", id2, app.sessionView.PreviewSessionID)
	}
}

func TestPreviewRecovery(t *testing.T) {
	app, repo, sm := setupPreviewTest(t)
	workspaces, _ := repo.ListWorkspaces()

	c1 := config.IDCard{ID: "c1", Name: "C1", Driver: "echo", WorkspaceID: workspaces[0].ID}
	c2 := config.IDCard{ID: "c2", Name: "C2", Driver: "echo", WorkspaceID: workspaces[0].ID}
	_, _ = repo.CreateIDCard(c1)
	_, _ = repo.CreateIDCard(c2)

	id1, _ := sm.Launch(c1, nil, nil, nil)
	id2, _ := sm.Launch(c2, nil, nil, nil)
	defer func() { _ = sm.Kill(id1) }()
	defer func() { _ = sm.Kill(id2) }()

	app.sessions = sm.List()
	app.refreshSessionList()

	// Select id1
	app.sessionsList.Select(0)
	app.syncPreviewSelection()

	// Remove id1
	_ = sm.Kill(id1)
	_ = sm.Remove(id1)
	app.sessions = sm.List()
	app.refreshSessionList() // calls syncPreviewSelection

	if app.sessionView.PreviewSessionID != id2 {
		t.Errorf("expected preview to rebind to %s, got %s", id2, app.sessionView.PreviewSessionID)
	}

	// Remove id2
	_ = sm.Kill(id2)
	_ = sm.Remove(id2)
	app.sessions = sm.List()
	app.refreshSessionList()

	if app.sessionView.PreviewSessionID != "" {
		t.Errorf("expected preview to clear, got %s", app.sessionView.PreviewSessionID)
	}
}

func TestOrphanedSessionPreview(t *testing.T) {
	app, _, _ := setupPreviewTest(t)

	// Create a session without a card in repo/app
	snap := session.Snapshot{
		ID:     "orphaned-1",
		CardID: "non-existent",
		Name:   "Orphaned Session",
		Status: session.StatusTerminated,
	}
	// We need to trick the app into having this session
	app.sessions = []session.Snapshot{snap}
	app.sessionView.PreviewSessionID = snap.ID

	// This should not panic
	output := renderSessionPreview(app, 100)

	if output == "" {
		t.Fatal("expected non-empty output for orphaned session")
	}
}

func TestPreviewScrollbackUpdates(t *testing.T) {
	app, repo, sm := setupPreviewTest(t)
	workspaces, _ := repo.ListWorkspaces()

	card := config.IDCard{ID: "c1", Name: "C1", Driver: "echo", WorkspaceID: workspaces[0].ID}
	_, _ = repo.CreateIDCard(card)
	id, err := sm.Launch(card, nil, nil, nil)
	if err != nil {
		t.Fatalf("launch failed: %v", err)
	}
	defer func() { _ = sm.Kill(id) }()

	app.sessions = sm.List()
	app.refreshSessionList()
	app.sessionView.FocusedPane = paneSessions

	if app.sessionView.PreviewSessionID != id {
		t.Fatalf("expected preview ID %s, got %s", id, app.sessionView.PreviewSessionID)
	}

	app.sessionView.PreviewScrollback = []byte("old")

	// We can't easily push to a session's scrollback from outside without a driver
	// but we can mock the scrollback by just verifying GetScrollback is called.
	// Since we are in the same package (ui), we could technically access sm if it had a way to mock.

	// Let's use the fact that GetScrollback returns what's in the manager.
	// If we call syncPreviewSelection, it should refresh from the manager.
	// If the manager has nothing, it will be nil.

	app.sessionView.PreviewScrollback = []byte("should-be-overwritten")
	app.syncPreviewSelection()

	// Manager should have empty scrollback for a new session without input
	if string(app.sessionView.PreviewScrollback) == "should-be-overwritten" {
		t.Errorf("expected scrollback to be refreshed (overwritten)")
	}
}
