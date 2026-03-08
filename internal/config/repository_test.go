package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCreateAndListIDCard(t *testing.T) {
	repo, err := NewFileRepository(t.TempDir())
	if err != nil {
		t.Fatalf("new repo: %v", err)
	}

	skill, err := repo.UpsertSkill(Skill{Name: "analysis"})
	if err != nil {
		t.Fatalf("upsert skill: %v", err)
	}
	mcp, err := repo.UpsertMCP(MCPServer{Name: "local", Command: "mcpd"})
	if err != nil {
		t.Fatalf("upsert mcp: %v", err)
	}

	_, err = repo.CreateIDCard(IDCard{
		Name:       "Claude",
		Driver:     "cat",
		Model:      "opus",
		WorkingDir: t.TempDir(),
		SkillIDs:   []string{skill.ID},
		MCPIDs:     []string{mcp.ID},
	})
	if err != nil {
		t.Fatalf("create card: %v", err)
	}

	cards, err := repo.ListIDCards()
	if err != nil {
		t.Fatalf("list cards: %v", err)
	}
	if len(cards) != 1 {
		t.Fatalf("expected 1 card, got %d", len(cards))
	}
	if cards[0].Name != "Claude" {
		t.Fatalf("unexpected card name: %s", cards[0].Name)
	}
}

func TestCreateIDCardRejectsUnknownRefs(t *testing.T) {
	repo, err := NewFileRepository(t.TempDir())
	if err != nil {
		t.Fatalf("new repo: %v", err)
	}

	_, err = repo.CreateIDCard(IDCard{
		Name:       "Bad",
		Driver:     "cat",
		WorkingDir: t.TempDir(),
		SkillIDs:   []string{"missing"},
	})
	if err == nil {
		t.Fatalf("expected validation error")
	}
}

func TestRepositoryBootstrapsFiles(t *testing.T) {
	dir := t.TempDir()
	repo, err := NewFileRepository(dir)
	if err != nil {
		t.Fatalf("new repo: %v", err)
	}
	paths := []string{
		filepath.Join(repo.Dir(), "id_cards.yaml"),
		filepath.Join(repo.Dir(), "skills.yaml"),
		filepath.Join(repo.Dir(), "mcps.yaml"),
	}
	for _, p := range paths {
		if _, err := os.Stat(p); err != nil {
			t.Fatalf("stat %s: %v", p, err)
		}
	}
}
