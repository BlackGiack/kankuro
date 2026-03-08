package session

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"kankuro/internal/config"
)

func TestSessionLaunchAttachKill(t *testing.T) {
	sm := NewManager()
	card := config.IDCard{
		ID:         "card-1",
		Name:       "echo",
		Driver:     "cat",
		WorkingDir: t.TempDir(),
	}

	sid, err := sm.Launch(card)
	if err != nil {
		t.Fatalf("launch: %v", err)
	}

	handle, err := sm.Attach(sid)
	if err != nil {
		t.Fatalf("attach: %v", err)
	}
	defer handle.Close()
	if _, err := handle.Input.Write([]byte("hello\n")); err != nil {
		t.Fatalf("write: %v", err)
	}
	select {
	case chunk := <-handle.Output:
		if !bytes.Contains(chunk, []byte("hello")) {
			t.Fatalf("expected echoed output, got %q", string(chunk))
		}
	case <-time.After(2 * time.Second):
		t.Fatalf("timed out waiting for attach output")
	}

	time.Sleep(250 * time.Millisecond)
	found := false
	for _, s := range sm.List() {
		if s.ID == sid && (strings.EqualFold(string(s.Status), "running") || s.Status == StatusWaiting) {
			found = true
		}
	}
	if !found {
		t.Fatalf("session not found in list")
	}

	if err := sm.Kill(sid); err != nil {
		t.Fatalf("kill: %v", err)
	}
	time.Sleep(100 * time.Millisecond)

	for _, s := range sm.List() {
		if s.ID == sid && s.Status != StatusTerminated {
			t.Fatalf("expected terminated, got %s", s.Status)
		}
	}
}

func TestSessionAttachReplaysScrollbackAndTracksPromptWithoutNewline(t *testing.T) {
	sm := NewManager()
	card := config.IDCard{
		ID:         "card-2",
		Name:       "prompt",
		Driver:     "cat",
		WorkingDir: t.TempDir(),
	}

	sid, err := sm.Launch(card)
	if err != nil {
		t.Fatalf("launch: %v", err)
	}
	time.Sleep(150 * time.Millisecond)

	handle, err := sm.Attach(sid)
	if err != nil {
		t.Fatalf("attach: %v", err)
	}
	defer handle.Close()

	if _, err := handle.Input.Write([]byte("Ready?")); err != nil {
		t.Fatalf("write prompt: %v", err)
	}
	time.Sleep(250 * time.Millisecond)

	reattach, err := sm.Attach(sid)
	if err != nil {
		t.Fatalf("reattach: %v", err)
	}
	defer reattach.Close()

	if !bytes.Contains(reattach.Scrollback, []byte("Ready?")) {
		t.Fatalf("expected scrollback replay, got %q", string(reattach.Scrollback))
	}

	waiting := false
	for _, s := range sm.List() {
		if s.ID == sid && s.Status == StatusWaiting {
			waiting = true
		}
	}
	if !waiting {
		t.Fatalf("expected prompt-like partial output to mark session waiting")
	}

	if err := sm.Kill(sid); err != nil {
		t.Fatalf("kill: %v", err)
	}
}

func TestRemoveOnlyDeletesTerminatedSessions(t *testing.T) {
	sm := NewManager()
	card := config.IDCard{
		ID:         "card-3",
		Name:       "remove",
		Driver:     "cat",
		WorkingDir: t.TempDir(),
	}

	sid, err := sm.Launch(card)
	if err != nil {
		t.Fatalf("launch: %v", err)
	}

	if err := sm.Remove(sid); err == nil {
		t.Fatalf("expected remove to fail for running session")
	}

	if err := sm.Kill(sid); err != nil {
		t.Fatalf("kill: %v", err)
	}
	time.Sleep(100 * time.Millisecond)

	if err := sm.Remove(sid); err != nil {
		t.Fatalf("remove: %v", err)
	}

	for _, s := range sm.List() {
		if s.ID == sid {
			t.Fatalf("expected session to be removed")
		}
	}
}
