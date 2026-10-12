package session_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/BlackGiack/kankuro/internal/config"
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

	sm := newTestManager(t, repo.Dir(), repo)
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

	sm2 := newTestManager(t, repo.Dir(), repo)

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

// TestSaveSessionIsAtomic checks the observable result of an atomic write: the
// sessions directory holds only <id>.json (no temp residue) with mode 0600.
func TestSaveSessionIsAtomic(t *testing.T) {
	dir := t.TempDir()
	repo, _ := config.NewFileRepository(dir)
	wd := t.TempDir()
	_ = repo.SaveWorkspaceSources([]config.WorkspaceSource{{Path: wd, IsRoot: false}})
	workspaces, _ := repo.ListWorkspaces()

	sm := newTestManager(t, repo.Dir(), repo)
	card := config.IDCard{ID: "card-atomic", Name: "atomic", Driver: "cat", WorkspaceID: workspaces[0].ID}
	sid, err := sm.Launch(card, nil, nil, nil)
	if err != nil {
		t.Fatalf("launch: %v", err)
	}
	if errKill := sm.Kill(sid); errKill != nil {
		t.Fatalf("kill: %v", errKill)
	}

	sessionsDir := filepath.Join(repo.Dir(), "sessions")
	entries, err := os.ReadDir(sessionsDir)
	if err != nil {
		t.Fatalf("readdir: %v", err)
	}
	if len(entries) != 1 || entries[0].Name() != sid+".json" {
		names := make([]string, 0, len(entries))
		for _, e := range entries {
			names = append(names, e.Name())
		}
		t.Fatalf("expected only %s.json, found %v", sid, names)
	}
	info, err := os.Stat(filepath.Join(sessionsDir, sid+".json"))
	if err != nil {
		t.Fatalf("stat: %v", err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("expected mode 0600, got %o", info.Mode().Perm())
	}
}

// TestLoadSessionsIgnoresLeftoverTempFiles pins the temp-file naming: a
// ".<id>.json.<rand>.tmp" left behind by a crash must not match loadSessions'
// "*.json" glob, so it is neither parsed nor reported as a load error.
func TestLoadSessionsIgnoresLeftoverTempFiles(t *testing.T) {
	dir := t.TempDir()
	repo, _ := config.NewFileRepository(dir)
	wd := t.TempDir()
	_ = repo.SaveWorkspaceSources([]config.WorkspaceSource{{Path: wd, IsRoot: false}})
	workspaces, _ := repo.ListWorkspaces()

	sm := newTestManager(t, repo.Dir(), repo)
	card := config.IDCard{ID: "card-leftover", Name: "leftover", Driver: "cat", WorkspaceID: workspaces[0].ID}
	sid, err := sm.Launch(card, nil, nil, nil)
	if err != nil {
		t.Fatalf("launch: %v", err)
	}
	if errKill := sm.Kill(sid); errKill != nil {
		t.Fatalf("kill: %v", errKill)
	}

	leftover := filepath.Join(repo.Dir(), "sessions", "."+sid+".json.123.tmp")
	if errWrite := os.WriteFile(leftover, []byte("{garbage"), 0o600); errWrite != nil {
		t.Fatalf("write leftover: %v", errWrite)
	}

	sm2 := newTestManager(t, repo.Dir(), repo)
	if loadErr := sm2.LoadErrors(); loadErr != nil {
		t.Fatalf("leftover temp file must be ignored, got load error: %v", loadErr)
	}
	if _, ok := sm2.GetSession(sid); !ok {
		t.Fatal("expected the valid session to still load")
	}
}
