package driver_test

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/BlackGiack/kankuro/internal/config"
	"github.com/BlackGiack/kankuro/internal/driver"
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

func TestSubmitPromptValidatesAndSubmitsMarkedLine(t *testing.T) {
	d, _, output := spawnScript(t, "cat", nil)
	evidence, err := d.SubmitPrompt(context.Background(), "message-1", "sender-1", "inspect README")
	if err != nil {
		t.Fatal(err)
	}
	if evidence.State != driver.SubmissionComplete || evidence.BytesWritten == 0 {
		t.Fatalf("submission evidence = %#v", evidence)
	}
	waitForMatch(t, output, "marked prompt", func(value string) bool {
		return strings.Contains(value, "message_id=message-1 from_session=sender-1") && strings.Contains(value, "inspect README")
	})
	for _, invalid := range []string{"", "two\nlines", "escape\x1b[31m", "control\x00byte"} {
		if _, submitErr := d.SubmitPrompt(context.Background(), "message-2", "sender-1", invalid); submitErr == nil {
			t.Errorf("SubmitPrompt(%q) succeeded", invalid)
		}
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

	if _, err := d.SetAgentsMD(card, nil, nil, wd, driver.InstructionReplace); err != nil {
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

func TestManagedConversationInstructionsAreIdempotentInAppendMode(t *testing.T) {
	d := driver.NewGenericCLIDriver()
	wd := t.TempDir()
	if err := os.WriteFile(filepath.Join(wd, "AGENTS.md"), []byte("workspace instructions"), 0o600); err != nil {
		t.Fatal(err)
	}
	card := config.IDCard{AgentsMD: "card instructions"}
	for range 2 {
		if _, err := d.SetAgentsMD(card, nil, config.BuiltinMCPs(), wd, driver.InstructionAppend); err != nil {
			t.Fatal(err)
		}
	}
	content, err := os.ReadFile(filepath.Join(wd, "AGENTS.md"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Count(string(content), "KANKURO CONVERSATION INSTRUCTIONS START") != 0 {
		t.Fatalf("managed instruction block duplicated: %q", content)
	}
	if !strings.Contains(string(content), "workspace instructions") || !strings.Contains(string(content), "card instructions") {
		t.Fatalf("append behavior lost authored content: %q", content)
	}
}

func TestManagedConversationInstructionsExistWithoutCardInstructions(t *testing.T) {
	wd := t.TempDir()
	if _, err := driver.NewGenericCLIDriver().
		SetAgentsMD(config.IDCard{}, nil, config.BuiltinMCPs(), wd, driver.InstructionReplace); err != nil {
		t.Fatal(err)
	}
	content, err := os.ReadFile(filepath.Join(wd, "AGENTS.md"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(content), "call kankuro register_conversation") {
		t.Fatalf("empty card instruction did not receive managed conversation block: %q", content)
	}
}

func TestManagedConversationInstructionsAreRemovedWhenKankuroIsUnselected(t *testing.T) {
	wd := t.TempDir()
	d := driver.NewGenericCLIDriver()
	card := config.IDCard{AgentsMD: "card-authored instructions"}
	if _, err := d.SetAgentsMD(card, nil, config.BuiltinMCPs(), wd, driver.InstructionReplace); err != nil {
		t.Fatal(err)
	}
	if _, err := d.SetAgentsMD(config.IDCard{}, nil, nil, wd, driver.InstructionAppend); err != nil {
		t.Fatal(err)
	}
	content, err := os.ReadFile(filepath.Join(wd, "AGENTS.md"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(content), "KANKURO CONVERSATION INSTRUCTIONS START") {
		t.Fatalf("unselected Kankuro instructions remain: %q", content)
	}
	if !strings.Contains(string(content), "card-authored instructions") {
		t.Fatalf("removing the managed block lost authored content: %q", content)
	}
}

func TestEmptyReplacementClearsPreviousCardInstructions(t *testing.T) {
	wd := t.TempDir()
	d := driver.NewGenericCLIDriver()
	cardA := config.IDCard{AgentsMD: "ROLE_A_PRIVATE_INSTRUCTIONS"}
	if _, err := d.SetAgentsMD(cardA, nil, config.BuiltinMCPs(), wd, driver.InstructionReplace); err != nil {
		t.Fatal(err)
	}
	if _, err := d.SetAgentsMD(config.IDCard{}, nil, nil, wd, driver.InstructionReplace); err != nil {
		t.Fatal(err)
	}
	content, err := os.ReadFile(filepath.Join(wd, "AGENTS.md"))
	if err != nil {
		t.Fatal(err)
	}
	if len(content) != 0 {
		t.Fatalf("empty replacement retained previous card instructions: %q", content)
	}
}

func TestSetSessionArguments(t *testing.T) {
	for _, test := range []struct {
		name   string
		driver driver.AgentDriver
		want   []string
	}{
		{name: "generic", driver: driver.NewGenericCLIDriver(), want: []string{"--session", "native id; not shell syntax"}},
		{name: "claude", driver: driver.NewClaudeCodeDriver(), want: []string{"--resume", "native id; not shell syntax"}},
	} {
		t.Run(test.name, func(t *testing.T) {
			if got := test.driver.SetSession(""); len(got) != 0 {
				t.Fatalf("new-conversation args = %#v", got)
			}
			got := test.driver.SetSession("native id; not shell syntax")
			if !slices.Equal(got, test.want) {
				t.Fatalf("resume args = %#v, want %#v", got, test.want)
			}
		})
	}
}

func TestClaudeSetSessionArgumentsFromBuiltinRegistry(t *testing.T) {
	d := driver.NewBuiltinRegistry().Get("claude")
	if _, ok := d.(*driver.ClaudeCodeDriver); !ok {
		t.Fatalf("built-in Claude driver type = %T", d)
	}
	for _, id := range []string{"6cbf35c6-7e9d-4cb5-aefc-0ad55096ed8d", "opaque id; retained exactly"} {
		if got, want := d.SetSession(id), []string{"--resume", id}; !slices.Equal(got, want) {
			t.Errorf("Claude args for %q = %#v, want %#v", id, got, want)
		}
	}
}

func TestCodexSetSessionArguments(t *testing.T) {
	d := driver.NewCodexCLIDriver()
	if got := d.SetSession(""); len(got) != 0 {
		t.Fatalf("new-conversation args = %#v", got)
	}
	want := []string{"resume", "native-session-id"}
	if got := d.SetSession("native-session-id"); !slices.Equal(got, want) {
		t.Fatalf("Codex resume args = %#v, want %#v", got, want)
	}
}

func TestClaudeCodeDriver(t *testing.T) {
	d := driver.NewClaudeCodeDriver()
	wd := t.TempDir()

	card := config.IDCard{
		AgentsMD: "# Claude instruction",
	}

	if _, err := d.SetAgentsMD(card, nil, nil, wd, driver.InstructionReplace); err != nil {
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
	r := driver.NewBuiltinRegistry()
	d := r.Get("claude")
	if _, ok := d.(*driver.ClaudeCodeDriver); !ok {
		t.Fatalf("expected ClaudeCodeDriver")
	}
}

func TestSetSkillsBuiltinSkill(t *testing.T) {
	fixture := []byte("---\nname: mcp-cli\n---\n# MCP-CLI\ntest fixture\n")
	setBuiltinSkillFixture(t, fixture)
	d := driver.NewGenericCLIDriver()
	wd := t.TempDir()

	skills := []config.Skill{
		{
			ID:      config.BuiltinSkillMCPCLI,
			Name:    "mcp-cli",
			Builtin: true,
		},
	}

	if err := d.SetSkills(skills, wd); err != nil {
		t.Fatalf("failed to set builtin skills: %v", err)
	}

	// Verify the skill directory was created
	skillPath := filepath.Join(wd, ".agents", "skills", "mcp-cli")
	info, err := os.Stat(skillPath)
	if err != nil {
		t.Fatalf("expected mcp-cli skill dir to exist: %v", err)
	}
	if !info.IsDir() {
		t.Fatalf("expected mcp-cli to be a directory")
	}

	// Verify SKILL.md was written with embedded content
	skillFile := filepath.Join(skillPath, "SKILL.md")
	content, err := os.ReadFile(skillFile)
	if err != nil {
		t.Fatalf("failed to read SKILL.md: %v", err)
	}
	if !bytes.Equal(content, fixture) {
		t.Fatalf("SKILL.md = %q, want injected fixture %q", content, fixture)
	}
}

func TestSetSkillsMixedBuiltinAndRegular(t *testing.T) {
	setBuiltinSkillFixture(t, []byte("---\nname: mcp-cli\n---\n# MCP-CLI\n"))
	d := driver.NewGenericCLIDriver()
	wd := t.TempDir()

	// Create a regular skill source
	skillSource := filepath.Join(t.TempDir(), "custom-skill")
	if err := os.MkdirAll(skillSource, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(skillSource, "SKILL.md"), []byte("# Custom Skill"), 0o644); err != nil {
		t.Fatalf("write skill: %v", err)
	}

	skills := []config.Skill{
		{
			ID:      config.BuiltinSkillMCPCLI,
			Name:    "mcp-cli",
			Builtin: true,
		},
		{
			ID:     "custom-id",
			Name:   "custom-skill",
			Source: skillSource,
		},
	}

	if err := d.SetSkills(skills, wd); err != nil {
		t.Fatalf("failed to set skills: %v", err)
	}

	// Verify both skills were created
	for _, name := range []string{"mcp-cli", "custom-skill"} {
		skillPath := filepath.Join(wd, ".agents", "skills", name)
		info, errStat := os.Stat(skillPath)
		if errStat != nil {
			t.Fatalf("expected %s skill dir to exist: %v", name, errStat)
		}
		if !info.IsDir() {
			t.Fatalf("expected %s to be a directory", name)
		}
		skillFile := filepath.Join(skillPath, "SKILL.md")
		if _, errFile := os.Stat(skillFile); errFile != nil {
			t.Fatalf("expected SKILL.md in %s: %v", name, errFile)
		}
	}
}

func TestSetSkillsBuiltinSkillDeduplication(t *testing.T) {
	setBuiltinSkillFixture(t, []byte("---\nname: mcp-cli\n---\n# MCP-CLI\n"))
	d := driver.NewGenericCLIDriver()
	wd := t.TempDir()

	skills := []config.Skill{
		{
			ID:      config.BuiltinSkillMCPCLI,
			Name:    "mcp-cli",
			Builtin: true,
		},
	}

	// First call should create the skill
	if err := d.SetSkills(skills, wd); err != nil {
		t.Fatalf("first SetSkills failed: %v", err)
	}

	// Second call should skip (directory already exists)
	if err := d.SetSkills(skills, wd); err != nil {
		t.Fatalf("second SetSkills failed (should skip existing): %v", err)
	}

	skillPath := filepath.Join(wd, ".agents", "skills", "mcp-cli")
	info, err := os.Stat(skillPath)
	if err != nil {
		t.Fatalf("expected mcp-cli skill dir to exist: %v", err)
	}
	if !info.IsDir() {
		t.Fatalf("expected mcp-cli to be a directory")
	}
}

func setBuiltinSkillFixture(t *testing.T, content []byte) {
	t.Helper()
	previous, hadPrevious := config.GetBuiltinSkillContent("mcp-cli")
	config.SetBuiltinMCPCLISkillContent(content)
	t.Cleanup(func() {
		if hadPrevious {
			config.SetBuiltinMCPCLISkillContent(previous)
		} else {
			config.SetBuiltinMCPCLISkillContent(nil)
		}
	})
}

func TestSetSkillsRejectsNonBuiltinWithEmptySource(t *testing.T) {
	d := driver.NewGenericCLIDriver()
	wd := t.TempDir()

	skills := []config.Skill{
		{
			ID:   "some-skill",
			Name: "some-skill",
			// Source is empty but skill is not marked as builtin
		},
	}

	err := d.SetSkills(skills, wd)
	if err == nil {
		t.Fatalf("expected error for non-builtin skill with empty source")
	}
	if !strings.Contains(err.Error(), "no source directory") {
		t.Fatalf("expected 'no source directory' error, got: %v", err)
	}
}

func TestLegacyConversationBlockIsRemovedWithoutChangingSurroundingContent(t *testing.T) {
	wd := t.TempDir()
	path := filepath.Join(wd, "AGENTS.md")
	before := "before\n<!-- KANKURO CONVERSATION INSTRUCTIONS START -->\nold\n<!-- KANKURO CONVERSATION INSTRUCTIONS END -->\nafter\n"
	if err := os.WriteFile(path, []byte(before), 0600); err != nil {
		t.Fatal(err)
	}
	d := driver.NewGenericCLIDriver()
	for range 2 {
		if _, err := d.SetAgentsMD(config.IDCard{}, nil, nil, wd, driver.InstructionPreserveWhenEmpty); err != nil {
			t.Fatal(err)
		}
		got, err := os.ReadFile(path)
		if err != nil || string(got) != "before\n\nafter\n" {
			t.Fatalf("unexpected content %q: %v", got, err)
		}
	}
}
