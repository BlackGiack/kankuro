package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSyncSkillsFromSources(t *testing.T) {
	repo, err := NewFileRepository(t.TempDir())
	if err != nil {
		t.Fatalf("new repo: %v", err)
	}

	sourceRoot := t.TempDir()
	skillA := filepath.Join(sourceRoot, "alpha")
	skillB := filepath.Join(sourceRoot, "beta")
	skillHidden := filepath.Join(sourceRoot, ".hidden")
	if err := os.MkdirAll(skillA, 0o755); err != nil {
		t.Fatalf("mkdir alpha: %v", err)
	}
	if err := os.MkdirAll(skillB, 0o755); err != nil {
		t.Fatalf("mkdir beta: %v", err)
	}
	if err := os.MkdirAll(skillHidden, 0o755); err != nil {
		t.Fatalf("mkdir hidden: %v", err)
	}

	if err := repo.writeAtomic(filepath.Join(repo.Dir(), skillsSourcesFilename), skillsSourcesFile{
		SchemaVersion: SchemaVersion,
		Paths:         []string{sourceRoot},
	}); err != nil {
		t.Fatalf("write skills_sources: %v", err)
	}

	if err := repo.SyncSkillsFromSources(); err != nil {
		t.Fatalf("sync skills: %v", err)
	}

	skills, err := repo.ListSkills()
	if err != nil {
		t.Fatalf("list skills: %v", err)
	}
	if len(skills) != 2 {
		t.Fatalf("expected 2 skills, got %d", len(skills))
	}

	expected := map[string]string{
		"alpha": skillIDFromSource(skillA),
		"beta":  skillIDFromSource(skillB),
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
