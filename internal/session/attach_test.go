package session_test

import (
	"kankuro/internal/config"
	"kankuro/internal/session"
	"testing"
	"time"
)

func TestSessionAttachReplay(t *testing.T) {
	repo, _ := config.NewFileRepository(t.TempDir())
	wd := t.TempDir()
	_ = repo.SaveWorkspaceSources([]config.WorkspaceSource{{Path: wd, IsRoot: false}})
	workspaces, _ := repo.ListWorkspaces()

	sm := session.NewManager(repo.Dir(), repo)
	card := config.IDCard{
		ID:          "card-attach",
		Name:        "attach",
		Driver:      "cat",
		WorkspaceID: workspaces[0].ID,
	}

	sid, err := sm.Launch(card, nil, nil, nil)
	if err != nil {
		t.Fatalf("launch: %v", err)
	}
	defer sm.Kill(sid)

	handle1, errAttach1 := sm.Attach(sid)
	if errAttach1 != nil {
		t.Fatalf("attach 1: %v", errAttach1)
	}
	if _, errWrite := handle1.Input.Write([]byte("hello")); errWrite != nil {
		t.Fatalf("write: %v", errWrite)
	}
	time.Sleep(100 * time.Millisecond)
	handle1.Close()

	handle2, errAttach2 := sm.Attach(sid)
	if errAttach2 != nil {
		t.Fatalf("attach 2: %v", errAttach2)
	}
	defer handle2.Close()

	if string(handle2.Scrollback) != "hello" {
		t.Fatalf("expected scrollback 'hello', got %q", handle2.Scrollback)
	}
}

func TestSessionOutputBroadcasting(t *testing.T) {
	repo, _ := config.NewFileRepository(t.TempDir())
	wd := t.TempDir()
	_ = repo.SaveWorkspaceSources([]config.WorkspaceSource{{Path: wd, IsRoot: false}})
	workspaces, _ := repo.ListWorkspaces()

	sm := session.NewManager(repo.Dir(), repo)
	card := config.IDCard{
		ID:          "card-broadcast",
		Name:        "broadcast",
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

	if _, errWrite := handle.Input.Write([]byte("world")); errWrite != nil {
		t.Fatalf("write: %v", errWrite)
	}

	select {
	case out := <-handle.Output:
		if string(out) != "world" {
			t.Fatalf("expected 'world', got %q", out)
		}
	case <-time.After(1 * time.Second):
		t.Fatal("timed out waiting for output")
	}
}
