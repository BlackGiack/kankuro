package session_test

import (
	"kankuro/internal/config"
	"kankuro/internal/session"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

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

	sid, err := sm.Launch(card, nil, nil, nil)
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
	if errSwitch := sm.SwitchAgent(sid, card, nil, nil, nil); errSwitch != nil {
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
	sid2, errLaunch := sm.Launch(card, nil, nil, nil)
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

func TestSessionClaudeAgentsMD(t *testing.T) {
	// Setup dummy claude binary
	tmpPath := t.TempDir()
	claudePath := filepath.Join(tmpPath, "claude")
	if err := os.WriteFile(claudePath, []byte("#!/bin/sh\ncat"), 0o755); err != nil {
		t.Fatalf("failed to create dummy claude: %v", err)
	}
	oldPath := os.Getenv("PATH")
	t.Setenv("PATH", tmpPath+string(os.PathListSeparator)+oldPath)

	sm := session.NewManager(t.TempDir())
	wd := t.TempDir()

	card := config.IDCard{
		ID:              "card-claude",
		Name:            "claude-test",
		Driver:          "claude", // Use Claude driver
		WorkingDir:      wd,
		UseTmpWorkspace: true,
		AgentsMD:        "# Claude Content",
	}

	sid, err := sm.Launch(card, nil, nil, nil)
	if err != nil {
		t.Fatalf("launch: %v", err)
	}
	time.Sleep(100 * time.Millisecond)

	s, ok := sm.GetSession(sid)
	if !ok {
		t.Fatalf("session not found")
	}
	finalWD := s.WorkingDir

	// Verify CLAUDE.md exists
	content, errRead := os.ReadFile(filepath.Join(finalWD, "CLAUDE.md"))
	if errRead != nil {
		t.Fatalf("read CLAUDE.md: %v", errRead)
	}
	if string(content) != "# Claude Content" {
		t.Fatalf("expected %q, got %q", "# Claude Content", string(content))
	}

	// Verify AGENTS.md does NOT exist
	if _, errStat := os.Stat(filepath.Join(finalWD, "AGENTS.md")); !os.IsNotExist(errStat) {
		t.Fatalf("expected AGENTS.md not to exist")
	}

	if errKill := sm.Kill(sid); errKill != nil {
		t.Fatalf("kill: %v", errKill)
	}
}

func TestSessionMCPs(t *testing.T) {
	sm := session.NewManager(t.TempDir())
	wd := t.TempDir()

	mcps := []config.MCPServer{
		{Name: "mcp1", Command: "node", Args: []string{"server.js"}},
		{Name: "mcp2", URL: "http://remote"},
	}

	card := config.IDCard{
		ID:              "card-mcp",
		Name:            "mcp-test",
		Driver:          "cat",
		WorkingDir:      wd,
		UseTmpWorkspace: true,
	}

	sid, err := sm.Launch(card, nil, nil, mcps)
	if err != nil {
		t.Fatalf("launch: %v", err)
	}
	time.Sleep(100 * time.Millisecond)

	s, ok := sm.GetSession(sid)
	if !ok {
		t.Fatalf("session not found")
	}
	finalWD := s.WorkingDir

	// Verify .kankuro/mcp-servers.json
	configPath := filepath.Join(finalWD, ".kankuro", "mcp-servers.json")
	if _, err = os.Stat(configPath); err != nil {
		t.Fatalf("mcp-servers.json missing: %v", err)
	}

	// Verify AGENTS.md
	agentsPath := filepath.Join(finalWD, "AGENTS.md")
	content, errRead := os.ReadFile(agentsPath)
	if errRead != nil {
		t.Fatalf("read AGENTS.md: %v", errRead)
	}
	expectedLine := `Available mcp: ["mcp1","mcp2"]`
	if !strings.Contains(string(content), expectedLine) {
		t.Fatalf("expected AGENTS.md to contain %q, got %q", expectedLine, string(content))
	}

	if errKill := sm.Kill(sid); errKill != nil {
		t.Fatalf("kill: %v", errKill)
	}
}
