package session_test

import (
	"bytes"
	"kankuro/internal/config"
	"kankuro/internal/session"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestSessionLaunchAttachKill(t *testing.T) {
	sm := session.NewManager(t.TempDir())
	card := config.IDCard{
		ID:         "card-1",
		Name:       "echo",
		Driver:     "cat",
		WorkingDir: t.TempDir(),
	}

	sid, err := sm.Launch(card, nil)
	if err != nil {
		t.Fatalf("launch: %v", err)
	}

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

	time.Sleep(250 * time.Millisecond)
	found := false
	for _, s := range sm.List() {
		if s.ID == sid && (strings.EqualFold(string(s.Status), "running") || s.Status == session.StatusWaiting) {
			found = true
		}
	}
	if !found {
		t.Fatalf("session not found in list")
	}

	if errKill := sm.Kill(sid); errKill != nil {
		t.Fatalf("kill: %v", errKill)
	}
	time.Sleep(100 * time.Millisecond)

	for _, s := range sm.List() {
		if s.ID == sid && s.Status != session.StatusTerminated {
			t.Fatalf("expected terminated, got %s", s.Status)
		}
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

	sid, err := sm.Launch(card, nil)
	if err != nil {
		t.Fatalf("launch: %v", err)
	}
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

	if errKill := sm.Kill(sid); errKill != nil {
		t.Fatalf("kill: %v", errKill)
	}
}

func TestSessionSwitchAgent(t *testing.T) {
	sm := session.NewManager(t.TempDir())
	wd := t.TempDir()
	card1 := config.IDCard{
		ID:         "card-1",
		Name:       "alpha",
		Driver:     "cat",
		WorkingDir: wd,
	}

	sid, err := sm.Launch(card1, nil)
	if err != nil {
		t.Fatalf("launch: %v", err)
	}
	time.Sleep(100 * time.Millisecond)

	sess, ok := sm.GetSession(sid)
	if !ok {
		t.Fatalf("session not found")
	}
	origWD := sess.WorkingDir
	if errKill := sm.Kill(sid); errKill != nil {
		t.Fatalf("kill: %v", errKill)
	}

	card2 := config.IDCard{
		ID:         "card-2",
		Name:       "beta",
		Driver:     "cat",
		WorkingDir: wd,
	}
	if errSwitch := sm.SwitchAgent(sid, card2, nil); errSwitch != nil {
		t.Fatalf("switch: %v", errSwitch)
	}
	time.Sleep(100 * time.Millisecond)

	s, ok := sm.GetSession(sid)
	if !ok {
		t.Fatalf("session not found after switch")
	}
	if s.CardID != card2.ID {
		t.Fatalf("expected card id %s, got %s", card2.ID, s.CardID)
	}
	if s.Name != card2.Name {
		t.Fatalf("expected name %s, got %s", card2.Name, s.Name)
	}
	if s.WorkingDir != origWD {
		t.Fatalf("expected working dir %s, got %s", origWD, s.WorkingDir)
	}
	if s.Status == session.StatusTerminated {
		t.Fatalf("expected running after switch, got %s", s.Status)
	}
	if errKillAfter := sm.Kill(sid); errKillAfter != nil {
		t.Fatalf("kill after switch: %v", errKillAfter)
	}
}

func TestSessionIdleWordDetection(t *testing.T) {
	sm := session.NewManager(t.TempDir())
	card := config.IDCard{
		ID:         "card-idle",
		Name:       "idle",
		Driver:     "cat",
		WorkingDir: t.TempDir(),
	}

	sid, err := sm.Launch(card, nil)
	if err != nil {
		t.Fatalf("launch: %v", err)
	}

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

	if errKill := sm.Kill(sid); errKill != nil {
		t.Fatalf("kill: %v", errKill)
	}
}

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

	if errKill := sm.Kill(sid); errKill != nil {
		t.Fatalf("kill: %v", errKill)
	}
}

func TestRemoveOnlyDeletesTerminatedSessions(t *testing.T) {
	sm := session.NewManager(t.TempDir())
	card := config.IDCard{
		ID:         "card-3",
		Name:       "remove",
		Driver:     "cat",
		WorkingDir: t.TempDir(),
	}

	sid, err := sm.Launch(card, nil)
	if err != nil {
		t.Fatalf("launch: %v", err)
	}

	if errRemove := sm.Remove(sid); errRemove == nil {
		t.Fatalf("expected remove to fail for running session")
	}

	if errKill := sm.Kill(sid); errKill != nil {
		t.Fatalf("kill: %v", errKill)
	}
	time.Sleep(100 * time.Millisecond)

	if errRemove := sm.Remove(sid); errRemove != nil {
		t.Fatalf("remove: %v", errRemove)
	}

	for _, s := range sm.List() {
		if s.ID == sid {
			t.Fatalf("expected session to be removed")
		}
	}
}

func TestSessionAgentsMD(t *testing.T) {
	sm := session.NewManager(t.TempDir())
	wd := t.TempDir()

	// Pre-create AGENTS.md in original workspace
	originalContent := "# Original Content"
	if errWrite := os.WriteFile(filepath.Join(wd, "AGENTS.md"), []byte(originalContent), 0o600); errWrite != nil {
		t.Fatalf("write original agents.md: %v", errWrite)
	}

	card := config.IDCard{
		ID:              "card-agents-md",
		Name:            "agents-md",
		Driver:          "cat",
		WorkingDir:      wd,
		UseTmpWorkspace: true,
		AgentsMD:        "# First Content",
	}

	sid, err := sm.Launch(card, nil)
	if err != nil {
		t.Fatalf("launch: %v", err)
	}
	time.Sleep(100 * time.Millisecond)

	s, ok := sm.GetSession(sid)
	if !ok {
		t.Fatalf("session not found")
	}
	finalWD := s.WorkingDir
	content, errRead := os.ReadFile(filepath.Join(finalWD, "AGENTS.md"))
	if errRead != nil {
		t.Fatalf("read AGENTS.md: %v", errRead)
	}
	// Should be appended because it's a new session and it existed in original wd
	expectedFirst := "# Original Content\n\n# First Content"
	if string(content) != expectedFirst {
		t.Fatalf("expected %q, got %q", expectedFirst, string(content))
	}

	if errKill := sm.Kill(sid); errKill != nil {
		t.Fatalf("kill: %v", errKill)
	}
	time.Sleep(100 * time.Millisecond)

	// Switch agent with different content -> should REPLACE
	card.AgentsMD = "# Second Content"
	if errSwitch := sm.SwitchAgent(sid, card, nil); errSwitch != nil {
		t.Fatalf("switch agent: %v", errSwitch)
	}
	time.Sleep(100 * time.Millisecond)

	content, errRead = os.ReadFile(filepath.Join(finalWD, "AGENTS.md"))
	if errRead != nil {
		t.Fatalf("read AGENTS.md after switch: %v", errRead)
	}
	expectedSecond := "# Second Content"
	if string(content) != expectedSecond {
		t.Fatalf("expected %q, got %q", expectedSecond, string(content))
	}

	if errKillAfter := sm.Kill(sid); errKillAfter != nil {
		t.Fatalf("kill after switch: %v", errKillAfter)
	}

	// Test path option
	tmpFile := filepath.Join(t.TempDir(), "external_agents.md")
	expectedPathContent := "# Content from file"
	if errWrite := os.WriteFile(tmpFile, []byte(expectedPathContent), 0o600); errWrite != nil {
		t.Fatalf("write tmp file: %v", errWrite)
	}

	card.AgentsMDPreferPath = true
	card.AgentsMDPath = tmpFile
	card.WorkingDir = t.TempDir() // Fresh WD without AGENTS.md
	sid2, errLaunch := sm.Launch(card, nil)
	if errLaunch != nil {
		t.Fatalf("launch with path: %v", errLaunch)
	}
	time.Sleep(100 * time.Millisecond)

	s2, ok := sm.GetSession(sid2)
	if !ok {
		t.Fatalf("session 2 not found")
	}
	finalWD2 := s2.WorkingDir
	content, errRead = os.ReadFile(filepath.Join(finalWD2, "AGENTS.md"))
	if errRead != nil {
		t.Fatalf("read AGENTS.md from path: %v", errRead)
	}
	// Should NOT append here because we didn't pre-create AGENTS.md in a fresh WD for sid2
	if string(content) != expectedPathContent {
		t.Fatalf("expected %q, got %q", expectedPathContent, string(content))
	}

	if errKill2 := sm.Kill(sid2); errKill2 != nil {
		t.Fatalf("kill: %v", errKill2)
	}
}
