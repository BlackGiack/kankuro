package session_test

import (
	"bytes"
	"kankuro/internal/config"
	"testing"
	"time"
)

func TestSessionPendingLineCapped(t *testing.T) {
	repo, _ := config.NewFileRepository(t.TempDir())
	wd := t.TempDir()
	_ = repo.SaveWorkspaceSources([]config.WorkspaceSource{{Path: wd, IsRoot: false}})
	workspaces, _ := repo.ListWorkspaces()

	sm := newTestManager(t, repo.Dir(), repo)
	card := config.IDCard{
		ID:          "card-pending",
		Name:        "pending",
		Driver:      "cat",
		WorkspaceID: workspaces[0].ID,
	}

	sid, err := sm.Launch(card, nil, nil, nil)
	if err != nil {
		t.Fatalf("launch: %v", err)
	}
	defer sm.Kill(sid)

	handle, errAttach := sm.Attach(sid)
	if errAttach != nil {
		t.Fatalf("attach: %v", errAttach)
	}
	defer handle.Close()

	payload := bytes.Repeat([]byte("a"), 8*1024*2) // pendingLineLimit is 8KB
	if _, errWrite := handle.Input.Write(payload); errWrite != nil {
		t.Fatalf("write: %v", errWrite)
	}

	deadline := time.Now().Add(2 * time.Second)
	for {
		n, ok := sm.PendingLineLen(sid)
		if !ok {
			t.Fatalf("session not found")
		}
		if n > 0 || time.Now().After(deadline) {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}

	n, ok := sm.PendingLineLen(sid)
	if !ok {
		t.Fatalf("session not found")
	}
	if n > 8*1024 {
		t.Fatalf("expected pending line <= %d, got %d", 8*1024, n)
	}
}
