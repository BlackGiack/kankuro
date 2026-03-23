package session_test

import (
	"bytes"
	"kankuro/internal/config"
	"kankuro/internal/session"
	"testing"
	"time"
)

func TestSessionAttach(t *testing.T) {
	sm := session.NewManager(t.TempDir())
	card := config.IDCard{
		ID:         "card-1",
		Name:       "echo",
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

	if _, errWrite := handle.Input.Write([]byte("hello\n")); errWrite != nil {
		t.Fatalf("write: %v", errWrite)
	}

	select {
	case chunk := <-handle.Output:
		if !bytes.Contains(chunk, []byte("hello")) {
			t.Fatalf("expected echoed output, got %q", string(chunk))
		}
	case <-time.After(2 * time.Second):
		t.Fatalf("timed out waiting for attach output")
	}
}

func TestSessionAttachReplaysScrollbackAndTracksPromptWithoutNewline(t *testing.T) {
	sm := session.NewManager(t.TempDir())
	card := config.IDCard{
		ID:         "card-2",
		Name:       "prompt",
		Driver:     "cat",
		WorkingDir: t.TempDir(),
	}

	sid, err := sm.Launch(card, nil, nil)
	if err != nil {
		t.Fatalf("launch: %v", err)
	}
	defer sm.Kill(sid)

	time.Sleep(150 * time.Millisecond)

	handle, errAttach := sm.Attach(sid)
	if errAttach != nil {
		t.Fatalf("attach: %v", errAttach)
	}
	defer handle.Close()

	if _, errWrite := handle.Input.Write([]byte("Ready?")); errWrite != nil {
		t.Fatalf("write prompt: %v", errWrite)
	}
	time.Sleep(250 * time.Millisecond)

	reattach, errReattach := sm.Attach(sid)
	if errReattach != nil {
		t.Fatalf("reattach: %v", errReattach)
	}
	defer reattach.Close()

	if !bytes.Contains(reattach.Scrollback, []byte("Ready?")) {
		t.Fatalf("expected scrollback replay, got %q", string(reattach.Scrollback))
	}

	waiting := false
	for _, s := range sm.List() {
		if s.ID == sid && s.Status == session.StatusWaiting {
			waiting = true
		}
	}
	if !waiting {
		t.Fatalf("expected prompt-like partial output to mark session waiting")
	}
}
