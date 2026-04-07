package config_test

import (
	"kankuro/internal/config"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestUpsertDomainTypes(t *testing.T) {
	repo, err := config.NewFileRepository(t.TempDir())
	if err != nil {
		t.Fatalf("new repo: %v", err)
	}

	// Test Skills
	s1, err := repo.UpsertSkill(config.Skill{Name: "skill1"})
	if err != nil || s1.ID == "" {
		t.Fatalf("upsert skill: %v", err)
	}
	s1.Description = "updated"
	s1Upd, err := repo.UpsertSkill(s1)
	if err != nil || s1Upd.ID != s1.ID || s1Upd.Description != "updated" {
		t.Fatalf("update skill: %v", err)
	}
	skills, _ := repo.ListSkills()
	if len(skills) != 1 {
		t.Fatalf("expected 1 skill, got %d", len(skills))
	}

	// Test Personas
	p1, err := repo.UpsertPersona(config.Persona{Name: "persona1"})
	if err != nil || p1.ID == "" {
		t.Fatalf("upsert persona: %v", err)
	}
	p1.Source = "updated"
	p1Upd, err := repo.UpsertPersona(p1)
	if err != nil || p1Upd.ID != p1.ID || p1Upd.Source != "updated" {
		t.Fatalf("update persona: %v", err)
	}
	personas, _ := repo.ListPersonas()
	if len(personas) != 1 {
		t.Fatalf("expected 1 persona, got %d", len(personas))
	}

	// Test MCPs
	m1, err := repo.UpsertMCP(config.MCPServer{Name: "mcp1", Command: "cmd1"})
	if err != nil || m1.ID == "" {
		t.Fatalf("upsert mcp: %v", err)
	}
	m1.Command = "updated"
	m1Upd, err := repo.UpsertMCP(m1)
	if err != nil || m1Upd.ID != m1.ID || m1Upd.Command != "updated" {
		t.Fatalf("update mcp: %v", err)
	}
	mcps, _ := repo.ListMCPs()
	if len(mcps) != 1 {
		t.Fatalf("expected 1 mcp, got %d", len(mcps))
	}
}

func TestUpdateIDCardRefactored(t *testing.T) {
	repo, err := config.NewFileRepository(t.TempDir())
	if err != nil {
		t.Fatalf("new repo: %v", err)
	}

	wd := t.TempDir()
	_ = repo.SaveWorkspaceSources([]config.WorkspaceSource{{Path: wd, IsRoot: false}})
	workspaces, _ := repo.ListWorkspaces()
	wsID := workspaces[0].ID

	card := config.IDCard{
		Name:        "Original",
		Driver:      "cat",
		WorkspaceID: wsID,
	}

	created, err := repo.CreateIDCard(card)
	if err != nil {
		t.Fatalf("create card: %v", err)
	}
	originalCreatedAt := created.CreatedAt
	originalUpdatedAt := created.UpdatedAt

	// Wait a bit to ensure UpdatedAt changes
	time.Sleep(10 * time.Millisecond)

	created.Name = "Updated"
	updated, err := repo.UpdateIDCard(created)
	if err != nil {
		t.Fatalf("update card: %v", err)
	}

	if updated.ID != created.ID {
		t.Fatalf("ID mismatch: %s != %s", updated.ID, created.ID)
	}
	if !updated.CreatedAt.Equal(originalCreatedAt) {
		t.Fatalf("CreatedAt changed: %v != %v", updated.CreatedAt, originalCreatedAt)
	}
	if !updated.UpdatedAt.After(originalUpdatedAt) {
		t.Fatalf("UpdatedAt did not increase: %v <= %v", updated.UpdatedAt, originalUpdatedAt)
	}

	// Test update unknown ID
	unknownCard := updated
	unknownCard.ID = "unknown"
	_, err = repo.UpdateIDCard(unknownCard)
	if err == nil {
		t.Fatal("expected error updating unknown card")
	}
}

func TestValidateCardRefactored(t *testing.T) {
	repo, err := config.NewFileRepository(t.TempDir())
	if err != nil {
		t.Fatalf("new repo: %v", err)
	}

	wd := t.TempDir()
	_ = repo.SaveWorkspaceSources([]config.WorkspaceSource{{Path: wd, IsRoot: false}})
	workspaces, _ := repo.ListWorkspaces()
	wsID := workspaces[0].ID

	t.Run("MissingFields", func(t *testing.T) {
		cases := []struct {
			card config.IDCard
			err  string
		}{
			{config.IDCard{}, "id card name is required"},
			{config.IDCard{Name: "N"}, "driver is required"},
			{config.IDCard{Name: "N", Driver: "D"}, "workspace is required"},
		}
		for _, tc := range cases {
			if vErr := repo.ValidateCard(tc.card); vErr == nil || vErr.Error() != tc.err {
				t.Errorf("expected error %q, got %v", tc.err, vErr)
			}
		}
	})

	t.Run("UnknownWorkspace", func(t *testing.T) {
		card := config.IDCard{Name: "N", Driver: "D", WorkspaceID: "missing"}
		if vErr := repo.ValidateCard(card); vErr == nil {
			t.Fatal("expected error for unknown workspace")
		}
	})

	t.Run("UnknownRefs", func(t *testing.T) {
		card := config.IDCard{Name: "N", Driver: "D", WorkspaceID: wsID}

		card.SkillIDs = []string{"missing-skill"}
		if vErr := repo.ValidateCard(card); vErr == nil {
			t.Fatal("expected error for unknown skill")
		}
		card.SkillIDs = nil

		card.PersonaIDs = []string{"missing-persona"}
		if vErr := repo.ValidateCard(card); vErr == nil {
			t.Fatal("expected error for unknown persona")
		}
		card.PersonaIDs = nil

		card.MCPIDs = []string{"missing-mcp"}
		if vErr := repo.ValidateCard(card); vErr == nil {
			t.Fatal("expected error for unknown mcp")
		}
		card.MCPIDs = nil
	})

	t.Run("ValidCard", func(t *testing.T) {
		skill, _ := repo.UpsertSkill(config.Skill{Name: "s"})
		persona, _ := repo.UpsertPersona(config.Persona{Name: "p"})
		mcp, _ := repo.UpsertMCP(config.MCPServer{Name: "m", Command: "c"})

		card := config.IDCard{
			Name:        "N",
			Driver:      "D",
			WorkspaceID: wsID,
			SkillIDs:    []string{skill.ID},
			PersonaIDs:  []string{persona.ID},
			MCPIDs:      []string{mcp.ID},
		}
		if vErr := repo.ValidateCard(card); vErr != nil {
			t.Fatalf("unexpected validation error: %v", vErr)
		}
	})
}

func TestValidateCardRegistryIntegrity(t *testing.T) {
	repo, err := config.NewFileRepository(t.TempDir())
	if err != nil {
		t.Fatalf("new repo: %v", err)
	}

	wd := t.TempDir()
	_ = repo.SaveWorkspaceSources([]config.WorkspaceSource{{Path: wd, IsRoot: false}})
	workspaces, _ := repo.ListWorkspaces()
	wsID := workspaces[0].ID

	card := config.IDCard{Name: "N", Driver: "D", WorkspaceID: wsID}

	corruptFile := func(filename string) {
		path := filepath.Join(repo.Dir(), filename)
		if wErr := os.WriteFile(path, []byte("invalid: yaml: :"), 0o600); wErr != nil {
			t.Fatalf("corrupt %s: %v", filename, wErr)
		}
	}

	// Corrupt Skills
	corruptFile(config.SkillsFilename)
	if vErr := repo.ValidateCard(card); vErr == nil {
		t.Fatal("expected error with corrupt skills.yaml")
	}
	_ = repo.EnsureFiles() // restore

	// Corrupt Personas
	corruptFile(config.PersonasFilename)
	if vErr := repo.ValidateCard(card); vErr == nil {
		t.Fatal("expected error with corrupt personas.yaml")
	}
	_ = repo.EnsureFiles() // restore

	// Corrupt MCPs
	corruptFile(config.MCPsFilename)
	if vErr := repo.ValidateCard(card); vErr == nil {
		t.Fatal("expected error with corrupt mcps.yaml")
	}
}
