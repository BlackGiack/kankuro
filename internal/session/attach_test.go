package session_test

import (
	"sync"
	"testing"
	"time"

	"github.com/BlackGiack/kankuro/internal/config"
)

func TestSessionAttachReplay(t *testing.T) {
	repo, _ := config.NewFileRepository(t.TempDir())
	wd := t.TempDir()
	_ = repo.SaveWorkspaceSources([]config.WorkspaceSource{{Path: wd, IsRoot: false}})
	workspaces, _ := repo.ListWorkspaces()

	sm := newTestManager(t, repo.Dir(), repo)
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

// TestAttachInputSurvivesConcurrentKill pins finding #14 from
// tech_debt_founding.md: attach input must be routed through the driver's
// locked SendInputBytes, not written directly to the raw PTY file, so a
// keystroke racing Kill (which closes and nils the pty under the driver's
// own lock) cannot land on an already-closed fd. Run with -race: the fix is
// the RLock/Lock pairing between SendInputBytes and Kill in
// internal/driver/driver.go, not just the absence of a torn read/write.
func TestAttachInputSurvivesConcurrentKill(t *testing.T) {
	repo, _ := config.NewFileRepository(t.TempDir())
	wd := t.TempDir()
	_ = repo.SaveWorkspaceSources([]config.WorkspaceSource{{Path: wd, IsRoot: false}})
	workspaces, _ := repo.ListWorkspaces()

	sm := newTestManager(t, repo.Dir(), repo)
	card := config.IDCard{
		ID:          "card-attach-kill-race",
		Name:        "attach-kill-race",
		Driver:      "cat",
		WorkspaceID: workspaces[0].ID,
	}

	sid, err := sm.Launch(card, nil, nil, nil)
	if err != nil {
		t.Fatalf("launch: %v", err)
	}

	handle, errAttach := sm.Attach(sid)
	if errAttach != nil {
		t.Fatalf("attach: %v", errAttach)
	}
	defer handle.Close()

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		for range 50 {
			_, _ = handle.Input.Write([]byte("x"))
		}
	}()
	go func() {
		defer wg.Done()
		_ = sm.Kill(sid)
	}()
	wg.Wait()

	// A write landing after Kill has completed must fail cleanly, not
	// succeed into (or panic on) a closed fd.
	if _, errWrite := handle.Input.Write([]byte("y")); errWrite == nil {
		t.Fatal("expected a post-kill write to return an error")
	}
}

func TestSessionOutputBroadcasting(t *testing.T) {
	repo, _ := config.NewFileRepository(t.TempDir())
	wd := t.TempDir()
	_ = repo.SaveWorkspaceSources([]config.WorkspaceSource{{Path: wd, IsRoot: false}})
	workspaces, _ := repo.ListWorkspaces()

	sm := newTestManager(t, repo.Dir(), repo)
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
