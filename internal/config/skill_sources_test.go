package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/BlackGiack/kankuro/internal/config"
)

func TestSyncSkillsFromSources(t *testing.T) {
	repo, err := config.NewFileRepository(t.TempDir())
	if err != nil {
		t.Fatalf("new repo: %v", err)
	}

	sourceRoot := t.TempDir()
	skillA := filepath.Join(sourceRoot, "alpha")
	skillB := filepath.Join(sourceRoot, "beta")
	skillHidden := filepath.Join(sourceRoot, ".hidden")
	if errDir := os.MkdirAll(skillA, 0o755); errDir != nil {
		t.Fatalf("mkdir alpha: %v", errDir)
	}
	if errDir := os.MkdirAll(skillB, 0o755); errDir != nil {
		t.Fatalf("mkdir beta: %v", errDir)
	}
	if errDir := os.MkdirAll(skillHidden, 0o755); errDir != nil {
		t.Fatalf("mkdir hidden: %v", errDir)
	}

	if errAtomic := repo.WriteAtomic(filepath.Join(repo.Dir(), config.SkillsSourcesFilename), config.SkillsSourcesFile{
		SchemaVersion: config.SchemaVersion,
		Paths:         []string{sourceRoot},
	}); errAtomic != nil {
		t.Fatalf("write skills_sources: %v", errAtomic)
	}

	if errSync := repo.SyncSkillsFromSources(); errSync != nil {
		t.Fatalf("sync skills: %v", errSync)
	}

	skills, errList := repo.ListSkills()
	if errList != nil {
		t.Fatalf("list skills: %v", errList)
	}
	// Should have 3 skills: alpha, beta (from source) + mcp-cli (builtin)
	if len(skills) != 3 {
		t.Fatalf("expected 3 skills (alpha, beta, mcp-cli), got %d", len(skills))
	}

	expected := map[string]string{
		"alpha": config.SkillIDFromSource(skillA),
		"beta":  config.SkillIDFromSource(skillB),
	}
	for _, skill := range skills {
		if wantID, ok := expected[skill.Name]; ok {
			if skill.ID != wantID {
				t.Fatalf("unexpected id for %s: %s", skill.Name, skill.ID)
			}
			if skill.Source == "" {
				t.Fatalf("missing source for %s", skill.Name)
			}
		}
	}
}
