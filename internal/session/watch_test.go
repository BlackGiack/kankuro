package session_test

import (
	"kankuro/internal/config"
	"kankuro/internal/session"
	"testing"
	"time"
)

func TestSessionIdleWordDetection(t *testing.T) {
	repo, _ := config.NewFileRepository(t.TempDir())
	wd := t.TempDir()
	_ = repo.SaveWorkspaceSources([]config.WorkspaceSource{{Path: wd, IsRoot: false}})
	workspaces, _ := repo.ListWorkspaces()

	sm := session.NewManager(repo.Dir(), repo)
	card := config.IDCard{
		ID:          "card-idle",
		Name:        "idle",
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

	// Send non-word animation characters
	if _, errWrite := handle.Input.Write([]byte("...")); errWrite != nil {
		t.Fatalf("write: %v", errWrite)
	}
	time.Sleep(100 * time.Millisecond)

	// Should still be running
	running := false
	for _, s := range sm.List() {
		if s.ID == sid && s.Status == session.StatusRunning {
			running = true
		}
	}
	if !running {
		t.Fatalf("expected session to be running after non-word output")
	}

	// Manually advance the "last word" time to simulate 16s silence in the test
	sm.SetLastWordAt(sid, time.Now().Add(-16*time.Second))

	// Wait for the watcher ticker (2s)
	time.Sleep(2500 * time.Millisecond)

	waiting := false
	for _, s := range sm.List() {
		if s.ID == sid && s.Status == session.StatusWaiting {
			waiting = true
		}
	}
	if !waiting {
		t.Fatalf("expected session to become waiting after 15s of word silence")
	}
}

// TestRelaunchSurvivesStaleWatchers reproduces a Kill->Relaunch cycle and
// asserts that the previous activation's watch/trackOutput goroutines do not
// clobber the freshly relaunched session once the watch ticker fires again.
func TestRelaunchSurvivesStaleWatchers(t *testing.T) {
	repo, _ := config.NewFileRepository(t.TempDir())
	wd := t.TempDir()
	_ = repo.SaveWorkspaceSources([]config.WorkspaceSource{{Path: wd, IsRoot: false}})
	workspaces, _ := repo.ListWorkspaces()

	sm := session.NewManager(repo.Dir(), repo)
	card := config.IDCard{
		ID:          "card-relaunch-race",
		Name:        "relaunch-race",
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

	if errRelaunch := sm.Relaunch(sid, card, nil, nil, nil); errRelaunch != nil {
		t.Fatalf("relaunch: %v", errRelaunch)
	}
	defer sm.Kill(sid)

	// Attach BEFORE the stale watcher's tick. A subscriber channel closed by
	// the stale watcher's closeSubscribersLocked is irreversible, unlike
	// Status, which the new watcher repaints on its own next tick a moment
	// later - so attaching first is what makes this test actually catch the
	// bug instead of observing a self-healed state.
	handle, errAttach := sm.Attach(sid)
	if errAttach != nil {
		t.Fatalf("attach after relaunch: %v", errAttach)
	}
	defer handle.Close()

	// Wait past the watch interval so the stale (pre-relaunch) watch/
	// trackOutput goroutines would have had a chance to fire and clobber the
	// new session, if the generation guard were missing. Margin above the 2s
	// watchInterval to keep this reliable under CI load.
	time.Sleep(3 * time.Second)

	s, ok := sm.GetSession(sid)
	if !ok {
		t.Fatalf("session not found after relaunch")
	}
	if s.Status == session.StatusTerminated {
		t.Fatalf("expected session to still be running/waiting after relaunch, got %s", s.Status)
	}

	if !sm.BroadcastOutput(sid, []byte("ping")) {
		t.Fatalf("expected session to still accept broadcasts after relaunch")
	}
	select {
	case data, chOpen := <-handle.Output:
		if !chOpen {
			t.Fatalf("subscriber channel closed unexpectedly after relaunch")
		}
		if string(data) != "ping" {
			t.Fatalf("expected 'ping', got %q", data)
		}
	case <-time.After(time.Second):
		t.Fatalf("timed out waiting for broadcast on relaunched session")
	}
}
