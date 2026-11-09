package session_test

import (
	"testing"
	"time"

	"github.com/BlackGiack/kankuro/internal/config"
	"github.com/BlackGiack/kankuro/internal/session"
)

// TestSessionRoundTripSurvivesPersistence builds a Session, persists it via
// saveSession (through the public Manager API), reloads it via a fresh
// Manager, and asserts every persisted field survives and Status is forced
// to Terminated.
func TestSessionRoundTripSurvivesPersistence(t *testing.T) {
	dir := t.TempDir()
	repo, _ := config.NewFileRepository(dir)
	wd := t.TempDir()
	_ = repo.SaveWorkspaceSources([]config.WorkspaceSource{{Path: wd, IsRoot: false}})
	workspaces, _ := repo.ListWorkspaces()

	sm := newTestManager(t, repo.Dir(), repo)
	card := config.IDCard{
		ID:          "card-roundtrip",
		Name:        "roundtrip",
		Driver:      "cat",
		WorkspaceID: workspaces[0].ID,
	}

	sid, err := testLaunch(sm, card, nil, nil, nil)
	if err != nil {
		t.Fatalf("launch: %v", err)
	}
	time.Sleep(100 * time.Millisecond)

	before, ok := sm.GetSession(sid)
	if !ok {
		t.Fatalf("session not found before kill")
	}

	if errKill := sm.Kill(sid); errKill != nil {
		t.Fatalf("kill: %v", errKill)
	}
	time.Sleep(100 * time.Millisecond)

	// Reload through a fresh Manager to exercise Data.restore().
	sm2 := newTestManager(t, repo.Dir(), repo)
	if errLoad := sm2.LoadErrors(); errLoad != nil {
		t.Fatalf("unexpected load error: %v", errLoad)
	}

	after, ok := sm2.GetSession(sid)
	if !ok {
		t.Fatalf("session not found after reload")
	}

	if after.Status != session.StatusTerminated {
		t.Fatalf("expected reloaded session to be Terminated, got %s", after.Status)
	}
	if after.ID != before.ID {
		t.Fatalf("ID mismatch: got %s want %s", after.ID, before.ID)
	}
	if after.CardID != before.CardID {
		t.Fatalf("CardID mismatch: got %s want %s", after.CardID, before.CardID)
	}
	if after.Name != before.Name {
		t.Fatalf("Name mismatch: got %s want %s", after.Name, before.Name)
	}
	if after.WorkingDir != before.WorkingDir {
		t.Fatalf("WorkingDir mismatch: got %s want %s", after.WorkingDir, before.WorkingDir)
	}
	if after.OriginalDir != before.OriginalDir {
		t.Fatalf("OriginalDir mismatch: got %s want %s", after.OriginalDir, before.OriginalDir)
	}
	if after.UseTmpWorkspace != before.UseTmpWorkspace {
		t.Fatalf("UseTmpWorkspace mismatch: got %v want %v", after.UseTmpWorkspace, before.UseTmpWorkspace)
	}
	if !after.StartedAt.Equal(before.StartedAt) {
		t.Fatalf("StartedAt mismatch: got %v want %v", after.StartedAt, before.StartedAt)
	}
	if !after.LastWordAt.Equal(before.LastWordAt) {
		t.Fatalf("LastWordAt mismatch: got %v want %v", after.LastWordAt, before.LastWordAt)
	}
}
