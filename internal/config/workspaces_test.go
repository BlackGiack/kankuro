package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"kankuro/internal/config"
)

func TestSyncWorkspacesFromSources(t *testing.T) {
	repo, err := config.NewFileRepository(t.TempDir())
	if err != nil {
		t.Fatalf("new repo: %v", err)
	}

	root := t.TempDir()
	alpha := filepath.Join(root, "alpha")
	beta := filepath.Join(root, "beta")
	hidden := filepath.Join(root, ".hidden")
	if mkdirErr := os.MkdirAll(alpha, 0o755); mkdirErr != nil {
		t.Fatalf("mkdir alpha: %v", mkdirErr)
	}
	if mkdirErr := os.MkdirAll(beta, 0o755); mkdirErr != nil {
		t.Fatalf("mkdir beta: %v", mkdirErr)
	}
	if mkdirErr := os.MkdirAll(hidden, 0o755); mkdirErr != nil {
		t.Fatalf("mkdir hidden: %v", mkdirErr)
	}

	if saveErr := repo.SaveWorkspaceSources([]config.WorkspaceSource{{Path: root, IsRoot: true}}); saveErr != nil {
		t.Fatalf("save workspace sources: %v", saveErr)
	}

	workspaces, err := repo.ListWorkspaces()
	if err != nil {
		t.Fatalf("list workspaces: %v", err)
	}
	if len(workspaces) != 2 {
		t.Fatalf("expected 2 workspaces, got %d", len(workspaces))
	}
}

func TestSyncWorkspacesFromSourcesSkipsInvalidAndDedupesPaths(t *testing.T) {
	repo, err := config.NewFileRepository(t.TempDir())
	if err != nil {
		t.Fatalf("new repo: %v", err)
	}

	workspace := t.TempDir()
	if saveErr := repo.SaveWorkspaceSources([]config.WorkspaceSource{
		{Path: "/path/that/does/not/exist", IsRoot: false},
		{Path: workspace, IsRoot: false},
		{Path: workspace, IsRoot: false},
	}); saveErr != nil {
		t.Fatalf("save workspace sources: %v", saveErr)
	}

	workspaces, err := repo.ListWorkspaces()
	if err != nil {
		t.Fatalf("list workspaces: %v", err)
	}
	if len(workspaces) != 1 {
		t.Fatalf("expected 1 deduped workspace, got %d", len(workspaces))
	}
	if workspaces[0].Path != workspace {
		t.Fatalf("expected workspace path %q, got %q", workspace, workspaces[0].Path)
	}
}
