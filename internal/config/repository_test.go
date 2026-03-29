package config_test

import (
	"kankuro/internal/config"
	"os"
	"path/filepath"
	"testing"
)

func TestCreateAndListIDCard(t *testing.T) {
	repo, err := config.NewFileRepository(t.TempDir())
	if err != nil {
		t.Fatalf("new repo: %v", err)
	}

	skill, err := repo.UpsertSkill(config.Skill{Name: "analysis"})
	if err != nil {
		t.Fatalf("upsert skill: %v", err)
	}
	mcp, err := repo.UpsertMCP(config.MCPServer{Name: "local", Command: "mcpd"})
	if err != nil {
		t.Fatalf("upsert mcp: %v", err)
	}
	persona, err := repo.UpsertPersona(config.Persona{Name: "senior-go", Source: "/tmp/go.md"})
	if err != nil {
		t.Fatalf("upsert persona: %v", err)
	}

	_, err = repo.CreateIDCard(config.IDCard{
		Name:       "Claude",
		Driver:     "cat",
		Model:      "opus",
		WorkingDir: t.TempDir(),
		SkillIDs:   []string{skill.ID},
		MCPIDs:     []string{mcp.ID},
		PersonaIDs: []string{persona.ID},
		AgentsMD:   "# Custom AgentsMD\nHello world",
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
	if cards[0].AgentsMD != "# Custom AgentsMD\nHello world" {
		t.Fatalf("unexpected AgentsMD: %s", cards[0].AgentsMD)
	}
}

func TestCreateIDCardRejectsUnknownRefs(t *testing.T) {
	repo, err := config.NewFileRepository(t.TempDir())
	if err != nil {
		t.Fatalf("new repo: %v", err)
	}

	_, err = repo.CreateIDCard(config.IDCard{
		Name:       "Bad",
		Driver:     "cat",
		WorkingDir: t.TempDir(),
		SkillIDs:   []string{"missing"},
	})
	if err == nil {
		t.Fatalf("expected validation error")
	}
}

func TestUpsertMCPValidation(t *testing.T) {
	repo, _ := config.NewFileRepository(t.TempDir())

	// Rejects missing name
	if _, err := repo.UpsertMCP(config.MCPServer{Command: "foo"}); err == nil {
		t.Fatalf("expected error for missing name")
	}

	// Rejects both command and url empty
	if _, err := repo.UpsertMCP(config.MCPServer{Name: "bad"}); err == nil {
		t.Fatalf("expected error for both command and url empty")
	}

	// Accepts local server
	if _, err := repo.UpsertMCP(config.MCPServer{Name: "local", Command: "foo"}); err != nil {
		t.Fatalf("failed to upsert local mcp: %v", err)
	}

	// Accepts remote server
	if _, err := repo.UpsertMCP(config.MCPServer{Name: "remote", URL: "http://foo"}); err != nil {
		t.Fatalf("failed to upsert remote mcp: %v", err)
	}
}

func TestRepositoryBootstrapsFiles(t *testing.T) {
	dir := t.TempDir()
	repo, err := config.NewFileRepository(dir)
	if err != nil {
		t.Fatalf("new repo: %v", err)
	}
	paths := []string{
		filepath.Join(repo.Dir(), "id_cards.yaml"),
		filepath.Join(repo.Dir(), "skills.yaml"),
		filepath.Join(repo.Dir(), "skills_sources.yaml"),
		filepath.Join(repo.Dir(), "personas.yaml"),
		filepath.Join(repo.Dir(), "persona_sources.yaml"),
		filepath.Join(repo.Dir(), "mcps.yaml"),
	}
	for _, p := range paths {
		if _, errStat := os.Stat(p); errStat != nil {
			t.Fatalf("stat %s: %v", p, errStat)
		}
	}
}
