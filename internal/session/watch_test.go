package session_test

import (
	"kankuro/internal/config"
	"kankuro/internal/session"
	"testing"
	"time"
)

func TestSessionIdleWordDetection(t *testing.T) {
	sm := session.NewManager(t.TempDir())
	card := config.IDCard{
		ID:         "card-idle",
		Name:       "idle",
		Driver:     "cat",
		WorkingDir: t.TempDir(),
	}

	sid, err := sm.Launch(card, nil, nil)
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
