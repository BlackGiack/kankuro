package session

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"kankuro/internal/config"
)

func TestTmpWorkspaceLifecycle(t *testing.T) {
	// Setup
	tmpBase := t.TempDir()
	configDir := filepath.Join(tmpBase, "config")
	sourceDir := filepath.Join(tmpBase, "source")
	if err := os.MkdirAll(sourceDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create a file in source
	originalContent := []byte("original")
	if err := os.WriteFile(filepath.Join(sourceDir, "data.txt"), originalContent, 0644); err != nil {
		t.Fatal(err)
	}

	// Set env var for tmp dir location
	os.Setenv("KANKURO_TMP_DIR", filepath.Join(tmpBase, "sessions_tmp"))
	defer os.Unsetenv("KANKURO_TMP_DIR")

	sm := NewManager(configDir)

	card := config.IDCard{
		ID:              "card-tmp",
		Name:            "tmp-test",
		Driver:          "cat", // simple driver
		WorkingDir:      sourceDir,
		UseTmpWorkspace: true,
	}

	// 1. Launch
	sid, err := sm.Launch(card, nil)
	if err != nil {
		t.Fatalf("launch: %v", err)
	}

	// Wait for start
	time.Sleep(100 * time.Millisecond)

	s := sm.sessions[sid]
	if s.WorkingDir == sourceDir {
		t.Fatalf("expected working dir to be different from source, got %s", s.WorkingDir)
	}

	// Check file copy
	copiedFile := filepath.Join(s.WorkingDir, "data.txt")
	content, err := os.ReadFile(copiedFile)
	if err != nil {
		t.Fatalf("read copied file: %v", err)
	}
	if string(content) != "original" {
		t.Fatalf("expected content 'original', got '%s'", string(content))
	}

	// 2. Modify file in session dir
	newContent := []byte("modified")
	if err := os.WriteFile(copiedFile, newContent, 0644); err != nil {
		t.Fatal(err)
	}

	// 3. Kill session
	if err := sm.Kill(sid); err != nil {
		t.Fatal(err)
	}

	// 4. Verify original unchanged
	origContent, err := os.ReadFile(filepath.Join(sourceDir, "data.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if string(origContent) != "original" {
		t.Fatalf("original file was modified!")
	}

	// 5. Relaunch
	if err := sm.Relaunch(sid, card, nil); err != nil {
		t.Fatalf("relaunch: %v", err)
	}

	time.Sleep(100 * time.Millisecond)
	s2 := sm.sessions[sid]

	// Verify same working dir
	if s2.WorkingDir != s.WorkingDir {
		t.Fatalf("expected same working dir on relaunch, got %s vs %s", s2.WorkingDir, s.WorkingDir)
	}

	// Verify modification persists
	content2, err := os.ReadFile(filepath.Join(s2.WorkingDir, "data.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if string(content2) != "modified" {
		t.Fatalf("expected content 'modified' after relaunch, got '%s'", string(content2))
	}

	// 6. Persistence check
	// Create new manager (simulate restart)
	sm2 := NewManager(configDir)
	restoredSession, ok := sm2.sessions[sid]
	if !ok {
		t.Fatalf("session not restored from disk")
	}
	if restoredSession.Status != StatusTerminated {
		// Wait, Kill sets status to Terminated and persists.
		// Relaunch sets to Running.
		// But we didn't Kill after Relaunch.
		// The Manager.Launch/Relaunch persists initial state.
		// But status updates are persisted on termination.
		// If we just check disk, it might show Running if we didn't kill it cleanly?
		// But NewManager sets status to Terminated for loaded sessions.
		t.Fatalf("expected restored status Terminated (forced by NewManager), got %s", restoredSession.Status)
	}
	if restoredSession.WorkingDir != s.WorkingDir {
		t.Fatalf("restored working dir mismatch")
	}
}

func TestSwitchAgentReplacesSkillsInTmpWorkspace(t *testing.T) {
	tmpBase := t.TempDir()
	configDir := filepath.Join(tmpBase, "config")
	sourceDir := filepath.Join(tmpBase, "source")
	if err := os.MkdirAll(sourceDir, 0755); err != nil {
		t.Fatal(err)
	}

	skill1Dir := filepath.Join(tmpBase, "skill-one")
	skill2Dir := filepath.Join(tmpBase, "skill-two")
	if err := os.MkdirAll(skill1Dir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(skill2Dir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(skill1Dir, "SKILL.md"), []byte("one"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(skill2Dir, "SKILL.md"), []byte("two"), 0644); err != nil {
		t.Fatal(err)
	}

	os.Setenv("KANKURO_TMP_DIR", filepath.Join(tmpBase, "sessions_tmp"))
	defer os.Unsetenv("KANKURO_TMP_DIR")

	sm := NewManager(configDir)

	card1 := config.IDCard{
		ID:              "card-tmp-1",
		Name:            "tmp-one",
		Driver:          "cat",
		WorkingDir:      sourceDir,
		UseTmpWorkspace: true,
	}
	skills1 := []config.Skill{
		{ID: "s1", Name: "alpha-skill", Source: skill1Dir},
	}
	sid, err := sm.Launch(card1, skills1)
	if err != nil {
		t.Fatalf("launch: %v", err)
	}
	time.Sleep(100 * time.Millisecond)

	wd := sm.sessions[sid].WorkingDir
	skill1Path := filepath.Join(wd, ".agents", "skills", "alpha-skill")
	if _, err := os.Stat(skill1Path); err != nil {
		t.Fatalf("expected skill1 present: %v", err)
	}

	if err := sm.Kill(sid); err != nil {
		t.Fatalf("kill: %v", err)
	}

	card2 := config.IDCard{
		ID:         "card-tmp-2",
		Name:       "tmp-two",
		Driver:     "cat",
		WorkingDir: sourceDir,
	}
	skills2 := []config.Skill{
		{ID: "s2", Name: "beta-skill", Source: skill2Dir},
	}
	if err := sm.SwitchAgent(sid, card2, skills2); err != nil {
		t.Fatalf("switch: %v", err)
	}
	time.Sleep(100 * time.Millisecond)

	if _, err := os.Stat(skill1Path); err == nil || !os.IsNotExist(err) {
		t.Fatalf("expected skill1 removed, got %v", err)
	}
	skill2Path := filepath.Join(wd, ".agents", "skills", "beta-skill")
	if _, err := os.Stat(skill2Path); err != nil {
		t.Fatalf("expected skill2 present: %v", err)
	}

	if err := sm.Kill(sid); err != nil {
		t.Fatalf("kill after switch: %v", err)
	}
}
