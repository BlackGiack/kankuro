package session_test

import (
	"bytes"
	"kankuro/internal/config"
	"kankuro/internal/session"
	"testing"
	"time"
)

func TestSessionPendingLineCapped(t *testing.T) {
	sm := session.NewManager(t.TempDir())
	card := config.IDCard{
		ID:         "card-pending",
		Name:       "pending",
		Driver:     "cat",
		WorkingDir: t.TempDir(),
	}

	sid, err := sm.Launch(card, nil)
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
		s, ok := sm.GetSession(sid)
		if !ok {
			t.Fatalf("session not found")
		}
		if len(s.PendingLine) > 0 || time.Now().After(deadline) {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}

	s, ok := sm.GetSession(sid)
	if !ok {
		t.Fatalf("session not found")
	}
	if len(s.PendingLine) > 8*1024 {
		t.Fatalf("expected pending line <= %d, got %d", 8*1024, len(s.PendingLine))
	}
}
