package session

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/BlackGiack/kankuro/internal/config"
	"github.com/google/uuid"
)

func TestRemoveResidualUsesSavedTemporaryIdentity(t *testing.T) {
	dir := t.TempDir()
	repo, err := config.NewFileRepository(dir)
	if err != nil {
		t.Fatal(err)
	}
	manager := NewManager(dir, repo)
	t.Cleanup(func() { manager.Shutdown() })
	id := uuid.NewString()
	root := filepath.Join(t.TempDir(), id)
	working := filepath.Join(root, "workspace")
	if errAt24 := os.MkdirAll(working, 0o700); errAt24 != nil {
		t.Fatal(errAt24)
	}
	if errAt27 := os.MkdirAll(filepath.Join(dir, "sessions"), 0o700); errAt27 != nil {
		t.Fatal(errAt27)
	}
	if errAt30 := os.WriteFile(filepath.Join(dir, "sessions", id+".json"), []byte(`{"id":"orphan"}`), 0o600); errAt30 != nil {
		t.Fatal(errAt30)
	}
	if _, errAt33 := manager.RemoveResidual(context.Background(), id, working, filepath.Dir(root), true); errAt33 == nil {
		t.Fatal("mismatched cleanup root accepted")
	}
	result, err := manager.RemoveResidual(context.Background(), id, working, root, true)
	if err != nil || !result.SessionRecordDeleted || !result.TemporaryWorkspaceRemoved {
		t.Fatalf("cleanup: %+v, %v", result, err)
	}
	if _, errAt40 := os.Stat(root); !os.IsNotExist(errAt40) {
		t.Fatalf("temporary workspace remains: %v", errAt40)
	}
	if _, errAt43 := manager.RemoveResidual(context.Background(), id, working, root, true); errAt43 != nil {
		t.Fatalf("retry: %v", errAt43)
	}
}
