package driver_test

import (
	"encoding/json"
	"kankuro/internal/config"
	"kankuro/internal/driver"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestIsPromptLike(t *testing.T) {
	now := time.Now()
	// Test immediate prompt detection
	if !driver.IsPromptLike("Continue?", now, now) {
		t.Fatalf("expected prompt-like line")
	}
	// Test idle word-silence detection (waiting after 15s)
	if !driver.IsPromptLike("done", now.Add(-20*time.Second), now) {
		t.Fatalf("old output SHOULD be prompt-like (waiting) due to word silence")
	}
	// Test that it's NOT prompt-like if recent words were seen and no prompt char
	if driver.IsPromptLike("processing", now.Add(-5*time.Second), now) {
		t.Fatalf("recent output without prompt char should not be prompt-like")
	}
}

func TestSetModel(t *testing.T) {
	d := driver.NewGenericCLIDriver()
	args := d.SetModel("opus")
	if len(args) != 2 || args[0] != "--model" || args[1] != "opus" {
		t.Fatalf("unexpected SetModel result: %v", args)
	}
}

func TestSelectMCPs(t *testing.T) {
	d := driver.NewGenericCLIDriver()
	wd := t.TempDir()

	mcps := []config.MCPServer{
		{
			Name:    "local-mcp",
			Command: "node",
			Args:    []string{"server.js"},
			Env:     map[string]string{"API_KEY": "123"},
			CWD:     "/some/path",
		},
		{
			Name:    "remote-mcp",
			URL:     "http://remote",
			Headers: map[string]string{"Auth": "Bearer token"},
		},
	}

	if err := d.SelectMCPs(mcps, wd); err != nil {
		t.Fatalf("failed to select mcps: %v", err)
	}

	configPath := filepath.Join(wd, ".kankuro", "mcp-servers.json")
	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("failed to read mcp config: %v", err)
	}

	// Verify it's valid JSON and contains what we expect
	var cfg struct {
		MCPServers map[string]struct {
			Command string            `json:"command"`
			Args    []string          `json:"args"`
			Env     map[string]string `json:"env"`
			CWD     string            `json:"cwd"`
			URL     string            `json:"url"`
			Headers map[string]string `json:"headers"`
		} `json:"mcpServers"`
	}
	if err = json.Unmarshal(data, &cfg); err != nil {
		t.Fatalf("failed to unmarshal mcp config: %v", err)
	}

	if len(cfg.MCPServers) != 2 {
		t.Fatalf("expected 2 mcp servers, got %d", len(cfg.MCPServers))
	}

	if local, ok := cfg.MCPServers["local-mcp"]; !ok {
		t.Fatalf("local-mcp missing")
	} else if local.Command != "node" || local.Args[0] != "server.js" || local.Env["API_KEY"] != "123" || local.CWD != "/some/path" {
		t.Fatalf("local-mcp fields mismatch: %+v", local)
	}

	if remote, ok := cfg.MCPServers["remote-mcp"]; !ok {
		t.Fatalf("remote-mcp missing")
	} else if remote.URL != "http://remote" || remote.Headers["Auth"] != "Bearer token" {
		t.Fatalf("remote-mcp fields mismatch: %+v", remote)
	}
}

func TestSetAgentsMD(t *testing.T) {
	d := driver.NewGenericCLIDriver()
	wd := t.TempDir()

	card := config.IDCard{
		AgentsMD: "# Inline instruction",
	}

	if err := d.SetAgentsMD(card, nil, nil, wd, false); err != nil {
		t.Fatalf("failed to set agents md: %v", err)
	}

	content, err := os.ReadFile(filepath.Join(wd, "AGENTS.md"))
	if err != nil {
		t.Fatalf("failed to read AGENTS.md: %v", err)
	}
	if string(content) != "# Inline instruction" {
		t.Fatalf("expected %q, got %q", "# Inline instruction", string(content))
	}
}

func TestClaudeCodeDriver(t *testing.T) {
	d := driver.NewClaudeCodeDriver()
	wd := t.TempDir()

	card := config.IDCard{
		AgentsMD: "# Claude instruction",
	}

	if err := d.SetAgentsMD(card, nil, nil, wd, false); err != nil {
		t.Fatalf("failed to set agents md: %v", err)
	}

	// Should create CLAUDE.md
	content, err := os.ReadFile(filepath.Join(wd, "CLAUDE.md"))
	if err != nil {
		t.Fatalf("failed to read CLAUDE.md: %v", err)
	}
	if string(content) != "# Claude instruction" {
		t.Fatalf("expected %q, got %q", "# Claude instruction", string(content))
	}

	// Should NOT create AGENTS.md
	if _, errStat := os.Stat(filepath.Join(wd, "AGENTS.md")); !os.IsNotExist(errStat) {
		t.Fatalf("expected AGENTS.md not to exist")
	}
}

func TestNewCLIDriverClaude(t *testing.T) {
	d := driver.NewCLIDriver("claude")
	if _, ok := d.(*driver.ClaudeCodeDriver); !ok {
		t.Fatalf("expected ClaudeCodeDriver")
	}
}
