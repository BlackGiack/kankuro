package ui_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/BlackGiack/kankuro/internal/api/contracts"
	"github.com/BlackGiack/kankuro/internal/config"
	"github.com/BlackGiack/kankuro/internal/ui/views"
)

//nolint:gocognit // Covers each source family and mutation against malformed storage.
func TestSourceMutationAbortsOnMalformedRegistry(t *testing.T) {
	for _, tt := range []struct {
		name, filename string
		kind           contracts.SourceKind
		view           views.SourceType
		delete         bool
	}{
		{"skill create", config.SkillsSourcesFilename, contracts.SourceSkills, views.SourceTypeSkills, false},
		{"persona create", config.PersonaSourcesFilename, contracts.SourcePersonas, views.SourceTypePersonas, false},
		{"puppet create", config.PuppetSourcesFilename, contracts.SourcePuppets, views.SourceTypePuppets, false},
		{"skill delete", config.SkillsSourcesFilename, contracts.SourceSkills, views.SourceTypeSkills, true},
		{"workspace create", config.WorkspaceSourcesFilename, contracts.SourceWorkspaces, views.SourceTypeWorkspaces, false},
		{"workspace delete", config.WorkspaceSourcesFilename, contracts.SourceWorkspaces, views.SourceTypeWorkspaces, true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			repo, err := newCurrentRepository(t.TempDir())
			if err != nil {
				t.Fatal(err)
			}
			callbacks := newScopedCoordinator(repo, newTestManager(t, repo.Dir(), repo)).SourceManagerCallbacks(tt.view)
			var identity string
			if tt.delete {
				result := callbacks.Mutate(context.Background(), contracts.SourceMutation{Kind: tt.kind, Path: t.TempDir()})
				if !result.Saved {
					t.Fatalf("seed: %+v", result)
				}
				rows, loadErr := callbacks.Load(tt.kind)
				if loadErr != nil || len(rows) == 0 {
					t.Fatalf("load: %v, %v", rows, loadErr)
				}
				identity = rows[0].Identity
			}
			path := filepath.Join(repo.Dir(), tt.filename)
			malformed := []byte("this: is: not: valid: yaml: [")
			if writeErr := os.WriteFile(path, malformed, 0o600); writeErr != nil {
				t.Fatal(writeErr)
			}
			result := callbacks.Mutate(
				context.Background(),
				contracts.SourceMutation{Kind: tt.kind, Identity: identity, Delete: tt.delete, Path: t.TempDir()},
			)
			if result.Saved || result.Err == nil {
				t.Fatalf("result: %+v", result)
			}
			after, err := os.ReadFile(path)
			if err != nil || string(after) != string(malformed) {
				t.Fatalf("file changed: %q %v", after, err)
			}
		})
	}
}

func TestWorkspaceSourceIDAssignedAndPreservedOnEdit(t *testing.T) {
	repo, err := newCurrentRepository(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	callbacks := newScopedCoordinator(repo, newTestManager(t, repo.Dir(), repo)).SourceManagerCallbacks(views.SourceTypeWorkspaces)
	if result := callbacks.Mutate(
		context.Background(),
		contracts.SourceMutation{Kind: contracts.SourceWorkspaces, Path: t.TempDir()},
	); !result.Saved {
		t.Fatalf("create: %+v", result)
	}
	sources, err := repo.ListWorkspaceSources()
	if err != nil || len(sources) != 1 || sources[0].ID == "" {
		t.Fatalf("sources: %v %v", sources, err)
	}
	id := sources[0].ID
	rows, err := callbacks.Load(contracts.SourceWorkspaces)
	if err != nil {
		t.Fatal(err)
	}
	result := callbacks.Mutate(context.Background(),
		contracts.SourceMutation{Kind: contracts.SourceWorkspaces, Identity: rows[0].Identity, Path: rows[0].Path, IsRoot: true},
	)
	if !result.Saved {
		t.Fatalf("edit: %+v", result)
	}
	sources, err = repo.ListWorkspaceSources()
	if err != nil || len(sources) != 1 || sources[0].ID != id || !sources[0].IsRoot {
		t.Fatalf("sources: %v %v", sources, err)
	}
}
