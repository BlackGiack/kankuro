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
		"schema_version: 1\n" +
			"id_cards:\n" +
			"  - id: card-1\n" +
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
