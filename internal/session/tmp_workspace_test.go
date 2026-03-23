package session_test

import (
	"kankuro/internal/config"
	"kankuro/internal/session"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestTmpWorkspaceLifecycle(t *testing.T) {
	// Setup
	tmpBase := t.TempDir()
	configDir := filepath.Join(tmpBase, "config")
	sourceDir := filepath.Join(tmpBase, "source")
	if err := os.MkdirAll(sourceDir, 0o755); err != nil {
		t.Fatal(err)
	}

	// Create a file in source
	originalContent := []byte("original")
	if err := os.WriteFile(filepath.Join(sourceDir, "data.txt"), originalContent, 0o644); err != nil {
		t.Fatal(err)
	}

	// Set env var for tmp dir location
	t.Setenv("KANKURO_TMP_DIR", filepath.Join(tmpBase, "sessions_tmp"))

	sm := session.NewManager(configDir)

	card := config.IDCard{
		ID:              "card-tmp",
		Name:            "tmp-test",
		Driver:          "cat", // simple driver
		WorkingDir:      sourceDir,
		UseTmpWorkspace: true,
	}

	// 1. Launch
	sid, err := sm.Launch(card, nil, nil)
	if err != nil {
		t.Fatalf("launch: %v", err)
	}

	// Wait for start
	time.Sleep(100 * time.Millisecond)

	s, ok := sm.GetSession(sid)
	if !ok {
		t.Fatalf("session not found")
	}
	if s.WorkingDir == sourceDir {
		t.Fatalf("expected working dir to be different from source, got %s", s.WorkingDir)
	}

	// Check file copy
	copiedFile := filepath.Join(s.WorkingDir, "data.txt")
	content, errRead := os.ReadFile(copiedFile)
	if errRead != nil {
		t.Fatalf("read copied file: %v", errRead)
	}
	if string(content) != "original" {
		t.Fatalf("expected content 'original', got '%s'", string(content))
	}

	// 2. Modify file in session dir
	newContent := []byte("modified")
	if errWrite := os.WriteFile(copiedFile, newContent, 0o644); errWrite != nil {
		t.Fatal(errWrite)
	}

	// 3. Kill session
	if errKill := sm.Kill(sid); errKill != nil {
		t.Fatal(errKill)
	}

	// 4. Verify original unchanged
	origContent, errOrig := os.ReadFile(filepath.Join(sourceDir, "data.txt"))
	if errOrig != nil {
		t.Fatal(errOrig)
	}
	if string(origContent) != "original" {
		t.Fatalf("original file was modified!")
	}

	// 5. Relaunch
	if errRelaunch := sm.Relaunch(sid, card, nil, nil); errRelaunch != nil {
		t.Fatalf("relaunch: %v", errRelaunch)
	}

	time.Sleep(100 * time.Millisecond)
	s2, ok2 := sm.GetSession(sid)
	if !ok2 {
		t.Fatalf("session not found after relaunch")
	}

	// Verify same working dir
	if s2.WorkingDir != s.WorkingDir {
		t.Fatalf("expected same working dir on relaunch, got %s vs %s", s2.WorkingDir, s.WorkingDir)
	}

	// Verify modification persists
	content2, errRead2 := os.ReadFile(filepath.Join(s2.WorkingDir, "data.txt"))
	if errRead2 != nil {
		t.Fatal(errRead2)
	}
	if string(content2) != "modified" {
		t.Fatalf("expected content 'modified' after relaunch, got '%s'", string(content2))
	}

	// 6. Persistence check
	// Create new manager (simulate restart)
	sm2 := session.NewManager(configDir)
	restoredSession, okRestored := sm2.GetSession(sid)
	if !okRestored {
		t.Fatalf("session not restored from disk")
	}
	if restoredSession.Status != session.StatusTerminated {
		t.Fatalf("expected restored status Terminated, got %s", restoredSession.Status)
	}
	if restoredSession.WorkingDir != s.WorkingDir {
		t.Fatalf("restored working dir mismatch")
	}
}

func TestSwitchAgentReplacesSkillsInTmpWorkspace(t *testing.T) {
	tmpBase := t.TempDir()
	configDir := filepath.Join(tmpBase, "config")
	sourceDir := filepath.Join(tmpBase, "source")
	if err := os.MkdirAll(sourceDir, 0o755); err != nil {
		t.Fatal(err)
	}

	skill1Dir := filepath.Join(tmpBase, "skill-one")
	skill2Dir := filepath.Join(tmpBase, "skill-two")
	if err := os.MkdirAll(skill1Dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(skill2Dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(skill1Dir, "SKILL.md"), []byte("one"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(skill2Dir, "SKILL.md"), []byte("two"), 0o644); err != nil {
		t.Fatal(err)
	}

	t.Setenv("KANKURO_TMP_DIR", filepath.Join(tmpBase, "sessions_tmp"))

	sm := session.NewManager(configDir)

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
	sid, err := sm.Launch(card1, skills1, nil)
	if err != nil {
		t.Fatalf("launch: %v", err)
	}
	time.Sleep(100 * time.Millisecond)

	s, ok := sm.GetSession(sid)
	if !ok {
		t.Fatalf("session not found")
	}
	wd := s.WorkingDir
	skill1Path := filepath.Join(wd, ".agents", "skills", "alpha-skill")
	if _, errStat := os.Stat(skill1Path); errStat != nil {
		t.Fatalf("expected skill1 present: %v", errStat)
	}

	if errKill := sm.Kill(sid); errKill != nil {
		t.Fatalf("kill: %v", errKill)
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
	if errSwitch := sm.SwitchAgent(sid, card2, skills2, nil); errSwitch != nil {
		t.Fatalf("switch: %v", errSwitch)
	}
	time.Sleep(100 * time.Millisecond)

	if _, errStat := os.Stat(skill1Path); errStat == nil || !os.IsNotExist(errStat) {
		t.Fatalf("expected skill1 removed, got %v", errStat)
	}
	skill2Path := filepath.Join(wd, ".agents", "skills", "beta-skill")
	if _, errStat2 := os.Stat(skill2Path); errStat2 != nil {
		t.Fatalf("expected skill2 present: %v", errStat2)
	}

	if errKillAfter := sm.Kill(sid); errKillAfter != nil {
		t.Fatalf("kill after switch: %v", errKillAfter)
	}
}
