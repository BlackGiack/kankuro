package config_test

import (
	"bytes"
	"kankuro/internal/config"
	"os"
	"path/filepath"
	"slices"
	"testing"
)

func TestListSkillsIncludesBuiltin(t *testing.T) {
	repo, err := config.NewFileRepository(t.TempDir())
	if err != nil {
		t.Fatalf("new repo: %v", err)
	}

	skills, err := repo.ListSkills()
	if err != nil {
		t.Fatalf("list skills: %v", err)
	}

	found := slices.ContainsFunc(skills, func(s config.Skill) bool {
		return s.Name == "mcp-cli" && s.ID == config.BuiltinSkillMCPCLI
	})
	if !found {
		t.Fatalf("expected mcp-cli built-in skill in list, got: %v", skills)
	}
}

func TestListSkillsMergesBuiltinsWithExisting(t *testing.T) {
	repo, err := config.NewFileRepository(t.TempDir())
	if err != nil {
		t.Fatalf("new repo: %v", err)
	}

	// Add a custom skill
	_, err = repo.UpsertSkill(config.Skill{
		ID:     "custom-skill-id",
		Name:   "custom-skill",
		Source: "/tmp/custom",
	})
	if err != nil {
		t.Fatalf("upsert skill: %v", err)
	}

	skills, err := repo.ListSkills()
	if err != nil {
		t.Fatalf("list skills: %v", err)
	}

	if len(skills) != 2 {
		t.Fatalf("expected 2 skills (custom + builtin), got %d", len(skills))
	}

	names := make(map[string]struct{})
	for _, s := range skills {
		names[s.Name] = struct{}{}
	}
	if _, ok := names["custom-skill"]; !ok {
		t.Fatal("missing custom skill")
	}
	if _, ok := names["mcp-cli"]; !ok {
		t.Fatal("missing mcp-cli builtin")
	}
}

func TestSyncSkillsFromSourcesPreservesBuiltins(t *testing.T) {
	repo, err := config.NewFileRepository(t.TempDir())
	if err != nil {
		t.Fatalf("new repo: %v", err)
	}

	// Create a skill source
	sourceRoot := t.TempDir()
	skillA := filepath.Join(sourceRoot, "alpha")
	if errDir := os.MkdirAll(skillA, 0o755); errDir != nil {
		t.Fatalf("mkdir alpha: %v", errDir)
	}

	// Configure source
	if errAtomic := repo.WriteAtomic(filepath.Join(repo.Dir(), config.SkillsSourcesFilename), config.SkillsSourcesFile{
		SchemaVersion: config.SchemaVersion,
		Paths:         []string{sourceRoot},
	}); errAtomic != nil {
		t.Fatalf("write skills_sources: %v", errAtomic)
	}

	// Sync from sources
	if errSync := repo.SyncSkillsFromSources(); errSync != nil {
		t.Fatalf("sync skills: %v", errSync)
	}

	skills, errList := repo.ListSkills()
	if errList != nil {
		t.Fatalf("list skills: %v", errList)
	}

	// Should have 2 skills: alpha (from source) + mcp-cli (builtin)
	if len(skills) != 2 {
		t.Fatalf("expected 2 skills (alpha + mcp-cli), got %d", len(skills))
	}

	names := make(map[string]struct{})
	for _, s := range skills {
		names[s.Name] = struct{}{}
	}
	if _, ok := names["alpha"]; !ok {
		t.Fatal("missing alpha skill from source")
	}
	if _, ok := names["mcp-cli"]; !ok {
		t.Fatal("missing mcp-cli builtin after sync")
	}
}

func TestValidateCardAcceptsBuiltinSkillID(t *testing.T) {
	repo, err := config.NewFileRepository(t.TempDir())
	if err != nil {
		t.Fatalf("new repo: %v", err)
	}

	// Setup workspace
	wd := t.TempDir()
	_ = repo.SaveWorkspaceSources([]config.WorkspaceSource{{Path: wd, IsRoot: false}})
	workspaces, _ := repo.ListWorkspaces()
	wsID := workspaces[0].ID

	// Create card with builtin skill ID
	card := config.IDCard{
		Name:        "TestCard",
		Driver:      "cat",
		WorkspaceID: wsID,
		SkillIDs:    []string{config.BuiltinSkillMCPCLI},
	}

	err = repo.ValidateCard(card)
	if err != nil {
		t.Fatalf("expected validation to accept builtin skill ID, got: %v", err)
	}
}

func TestBuiltinSkillsReturnsMCPCLI(t *testing.T) {
	builtins := config.BuiltinSkills()
	if len(builtins) != 1 {
		t.Fatalf("expected 1 builtin skill, got %d", len(builtins))
	}
	if builtins[0].Name != "mcp-cli" {
		t.Fatalf("expected mcp-cli, got %s", builtins[0].Name)
	}
	if !builtins[0].Builtin {
		t.Fatal("expected Builtin flag to be true")
	}
}

func TestIsBuiltinSkillIdentifiesBuiltin(t *testing.T) {
	if !config.IsBuiltinSkill(config.BuiltinSkillMCPCLI) {
		t.Fatal("expected BuiltinSkillMCPCLI to be recognized as builtin")
	}
	if config.IsBuiltinSkill("some-random-id") {
		t.Fatal("expected random ID to not be recognized as builtin")
	}
}

func TestGetBuiltinSkillReturnsCorrectEntry(t *testing.T) {
	skill, ok := config.GetBuiltinSkill(config.BuiltinSkillMCPCLI)
	if !ok {
		t.Fatal("expected to find builtin skill")
	}
	if skill.Name != "mcp-cli" {
		t.Fatalf("expected mcp-cli, got %s", skill.Name)
	}
}

func TestGetBuiltinSkillContentReturnsEmbeddedContent(t *testing.T) {
	content, ok := config.GetBuiltinSkillContent("mcp-cli")
	if !ok {
		t.Fatal("expected to find builtin skill content")
	}
	if len(content) == 0 {
		t.Fatal("expected non-empty content")
	}
	if !bytes.Contains(content, []byte("name: mcp-cli")) {
		t.Fatal("content should contain skill metadata")
	}
	if !bytes.Contains(content, []byte("# MCP-CLI")) {
		t.Fatal("content should contain # MCP-CLI header")
	}
}

func TestMergeBuiltinSkillsDoesNotDuplicate(t *testing.T) {
	skills := []config.Skill{
		{ID: config.BuiltinSkillMCPCLI, Name: "mcp-cli", Builtin: true},
	}
	merged := config.MergeBuiltinSkills(skills)
	count := 0
	for _, s := range merged {
		if s.Name == "mcp-cli" {
			count++
		}
	}
	if count != 1 {
		t.Fatalf("expected 1 mcp-cli after merge, got %d", count)
	}
}
