package session_test

import (
	"os"
	"path/filepath"
	"testing"

	"kankuro/internal/config"
	"kankuro/internal/session"
)

// TestLoadSessionsSkipsCorruptButReportsError exercises loadSessions
// indirectly through NewManager/LoadErrors: a directory with one valid
// session file and one truncated/corrupt one must still hydrate the valid
// session, while surfacing a non-nil error describing the skipped file.
func TestLoadSessionsSkipsCorruptButReportsError(t *testing.T) {
	dir := t.TempDir()
	repo, _ := config.NewFileRepository(dir)
	wd := t.TempDir()
	_ = repo.SaveWorkspaceSources([]config.WorkspaceSource{{Path: wd, IsRoot: false}})
	workspaces, _ := repo.ListWorkspaces()

	sm := session.NewManager(repo.Dir(), repo)
	card := config.IDCard{
		ID:          "card-corrupt",
		Name:        "corrupt",
		Driver:      "cat",
		WorkspaceID: workspaces[0].ID,
	}
	sid, err := sm.Launch(card, nil, nil, nil)
	if err != nil {
		t.Fatalf("launch: %v", err)
	}
	if errKill := sm.Kill(sid); errKill != nil {
		t.Fatalf("kill: %v", errKill)
	}

	sessionsDir := filepath.Join(repo.Dir(), "sessions")
	corruptPath := filepath.Join(sessionsDir, "corrupt-session.json")
	if errWrite := os.WriteFile(corruptPath, []byte("{not valid json"), 0o600); errWrite != nil {
		t.Fatalf("write corrupt file: %v", errWrite)
	}

	sm2 := session.NewManager(repo.Dir(), repo)

	loadErr := sm2.LoadErrors()
	if loadErr == nil {
		t.Fatalf("expected a non-nil load error for the corrupt session file")
	}

	found := false
	for _, s := range sm2.List() {
		if s.ID == sid {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected the valid session to still be loaded")
	}
}
