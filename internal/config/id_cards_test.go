package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/BlackGiack/kankuro/internal/config"
)

func TestLoadIDCardsFileIgnoresLegacyWorkingDir(t *testing.T) {
	repo, err := config.NewFileRepository(t.TempDir())
	if err != nil {
		t.Fatalf("new repo: %v", err)
	}

	raw := []byte(
		"schema_version: 2\n" +
			"id_cards:\n" +
			"  - id: card-1\n" +
			"    dashboard_id: default\n" +
			"    name: Claude\n" +
			"    driver: claude\n" +
			"    workspace_id: \"\"\n" +
			"    working_dir: /tmp/legacy\n",
	)
	path := filepath.Join(repo.Dir(), config.IDCardsFilename)
	if writeErr := os.WriteFile(path, raw, 0o600); writeErr != nil {
		t.Fatalf("write id cards yaml: %v", writeErr)
	}

	file, err := repo.LoadIDCardsFile()
	if err != nil {
		t.Fatalf("load id cards: %v", err)
	}
	if len(file.IDCards) != 1 {
		t.Fatalf("expected 1 id card, got %d", len(file.IDCards))
	}
	if file.IDCards[0].WorkspaceID != "" {
		t.Fatalf("expected legacy working_dir to be ignored, got workspace_id %q", file.IDCards[0].WorkspaceID)
	}
}

//nolint:govet // Keep each invalid-format assertion local.
func TestCardInventoryRequiresCurrentSchemaAndMembership(t *testing.T) {
	repo, err := config.NewFileRepository(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(repo.Dir(), config.IDCardsFilename)
	for _, raw := range []string{
		"id_cards: []\n",
		"schema_version: 1\nid_cards: []\n",
		"schema_version: 2\nid_cards:\n  - id: card\n    name: Card\n",
	} {
		if err := os.WriteFile(path, []byte(raw), 0o600); err != nil {
			t.Fatal(err)
		}
		if _, err := repo.LoadIDCardsFile(); err == nil {
			t.Fatalf("accepted invalid inventory: %s", raw)
		}
	}
	if err := repo.WriteIDCardsFile(config.IDCardsFile{SchemaVersion: 1}); err == nil {
		t.Fatal("wrote legacy schema")
	}
	if err := repo.WriteIDCardsFile(
		config.IDCardsFile{SchemaVersion: config.CardSchemaVersion, IDCards: []config.IDCard{{ID: "card"}}},
	); err == nil {
		t.Fatal("wrote card without membership")
	}
}
