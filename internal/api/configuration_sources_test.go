package api_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/BlackGiack/kankuro/internal/api"
	"github.com/BlackGiack/kankuro/internal/api/contracts"
	"github.com/BlackGiack/kankuro/internal/config"
)

func TestSourceIdentityRejectsReorderAndAddressesDuplicates(t *testing.T) {
	repo, err := config.NewFileRepository(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	a, b := t.TempDir(), t.TempDir()
	if saveErr := repo.SaveSkillSources([]string{a, a, b}); saveErr != nil {
		t.Fatal(saveErr)
	}
	service := api.New(repo, nil)
	rows, err := service.ReadSources(contracts.SourceSkills)
	if err != nil {
		t.Fatal(err)
	}
	if rows[0].Identity == rows[1].Identity {
		t.Fatal("duplicate entries have identical identities")
	}
	result := service.MutateSource(
		context.Background(),
		contracts.SourceMutation{Kind: contracts.SourceSkills, Identity: rows[1].Identity, Delete: true},
	)
	if !result.Saved || result.Err != nil {
		t.Fatalf("delete duplicate: %+v", result)
	}
	paths, err := repo.ListSkillSources()
	if err != nil || len(paths) != 2 || paths[0] != a || paths[1] != b {
		t.Fatalf("paths: %v, %v", paths, err)
	}
	result = service.MutateSource(
		context.Background(),
		contracts.SourceMutation{Kind: contracts.SourceSkills, Identity: rows[2].Identity, Delete: true},
	)
	if result.Saved || !result.Reload || result.Err == nil {
		t.Fatalf("stale identity: %+v", result)
	}
}

func TestSourceMutationReportsSavedSyncFailure(t *testing.T) {
	repo, err := config.NewFileRepository(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	service := api.New(repo, nil)
	result := service.MutateSource(
		context.Background(),
		contracts.SourceMutation{Kind: contracts.SourceSkills, Path: "/nonexistent/kankuro-source"},
	)
	if !result.Saved || !result.Refresh || result.Warning == nil || result.Err != nil {
		t.Fatalf("result: %+v", result)
	}
}

func TestSourceIdentityRejectsReorderAndWrongKind(t *testing.T) {
	repo, err := config.NewFileRepository(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	a, b := t.TempDir(), t.TempDir()
	if err = repo.SaveSkillSources([]string{a, b}); err != nil {
		t.Fatal(err)
	}
	if err = repo.SavePersonaSources([]string{a, b}); err != nil {
		t.Fatal(err)
	}
	service := api.New(repo, nil)
	rows, err := service.ReadSources(contracts.SourceSkills)
	if err != nil {
		t.Fatal(err)
	}
	if err = repo.SaveSkillSources([]string{b, a}); err != nil {
		t.Fatal(err)
	}
	for _, input := range []contracts.SourceMutation{{Kind: contracts.SourceSkills, Identity: rows[0].Identity, Path: b}, {Kind: contracts.SourcePersonas, Identity: rows[0].Identity, Delete: true}} {
		result := service.MutateSource(context.Background(), input)
		if result.Saved || !result.Reload || result.Err == nil {
			t.Fatalf("expected conflict: %+v", result)
		}
	}
	paths, err := repo.ListSkillSources()
	if err != nil || len(paths) != 2 || paths[0] != b || paths[1] != a {
		t.Fatalf("paths: %v %v", paths, err)
	}
}

func TestWorkspaceSourceIdentitySurvivesReorderAndRejectsEdit(t *testing.T) {
	repo, err := config.NewFileRepository(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	a, b := t.TempDir(), t.TempDir()
	if err = repo.SaveWorkspaceSources([]config.WorkspaceSource{{Path: a}, {Path: b}}); err != nil {
		t.Fatal(err)
	}
	service := api.New(repo, nil)
	rows, err := service.ReadSources(contracts.SourceWorkspaces)
	if err != nil {
		t.Fatal(err)
	}
	if err = repo.UpdateWorkspaceSources(func(sources []config.WorkspaceSource) ([]config.WorkspaceSource, error) {
		sources[0], sources[1] = sources[1], sources[0]
		return sources, nil
	}); err != nil {
		t.Fatal(err)
	}
	result := service.MutateSource(context.Background(),
		contracts.SourceMutation{Kind: contracts.SourceWorkspaces, Identity: rows[0].Identity, Path: a, IsRoot: true},
	)
	if !result.Saved {
		t.Fatalf("reordered edit: %+v", result)
	}
	result = service.MutateSource(
		context.Background(),
		contracts.SourceMutation{Kind: contracts.SourceWorkspaces, Identity: rows[0].Identity, Delete: true},
	)
	if result.Saved || !result.Reload || result.Err == nil {
		t.Fatalf("stale delete: %+v", result)
	}
	sources, err := repo.ListWorkspaceSources()
	if err != nil || len(sources) != 2 || sources[1].Path != a || !sources[1].IsRoot {
		t.Fatalf("sources: %v %v", sources, err)
	}
}

func TestConcurrentSourceCreatesPreserveEverySuccess(t *testing.T) {
	repo, err := config.NewFileRepository(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	service := api.New(repo, nil)
	const count = 8
	results := make(chan contracts.SourceMutationResult, count)
	for range count {
		path := t.TempDir()
		go func() {
			results <- service.MutateSource(context.Background(), contracts.SourceMutation{Kind: contracts.SourceSkills, Path: path})
		}()
	}
	for range count {
		if result := <-results; !result.Saved || result.Err != nil {
			t.Fatalf("create: %+v", result)
		}
	}
	paths, err := repo.ListSkillSources()
	if err != nil || len(paths) != count {
		t.Fatalf("paths: %v %v", paths, err)
	}
}

func TestCancelledSourceMutationDoesNotSave(t *testing.T) {
	repo, err := config.NewFileRepository(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	service := api.New(repo, nil)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	result := service.MutateSource(ctx, contracts.SourceMutation{Kind: contracts.SourceSkills, Path: t.TempDir()})
	if result.Saved || !errors.Is(result.Err, context.Canceled) {
		t.Fatalf("result: %+v", result)
	}
	paths, err := repo.ListSkillSources()
	if err != nil || len(paths) != 0 {
		t.Fatalf("paths: %v %v", paths, err)
	}
}

func TestSourceSyncFailureAfterSaveForEveryKind(t *testing.T) {
	tests := []struct {
		kind     contracts.SourceKind
		registry string
	}{
		{contracts.SourceSkills, config.SkillsFilename},
		{contracts.SourcePersonas, config.PersonasFilename},
		{contracts.SourcePuppets, config.PuppetsFilename},
		{contracts.SourceWorkspaces, config.WorkspacesFilename},
	}
	for _, tt := range tests {
		t.Run(string(tt.kind), func(t *testing.T) {
			root := t.TempDir()
			repo, err := config.NewFileRepository(root)
			if err != nil {
				t.Fatal(err)
			}
			registry := filepath.Join(root, tt.registry)
			if removeErr := os.Remove(registry); removeErr != nil {
				t.Fatal(removeErr)
			}
			if mkdirErr := os.Mkdir(registry, 0o700); mkdirErr != nil {
				t.Fatal(mkdirErr)
			}
			service := api.New(repo, nil)
			path := t.TempDir()
			result := service.MutateSource(context.Background(), contracts.SourceMutation{Kind: tt.kind, Path: path})
			if !result.Saved || !result.Refresh || result.Warning == nil || result.Err != nil {
				t.Fatalf("result: %+v", result)
			}
			rows, readErr := service.ReadSources(tt.kind)
			if readErr != nil || len(rows) != 1 || rows[0].Path != path {
				t.Fatalf("saved rows: %v %v", rows, readErr)
			}
		})
	}
}

type sourceWriteFailure struct {
	*config.FileRepository

	failure error
}

func (r sourceWriteFailure) UpdateSkillSources(func([]string) ([]string, error)) error {
	return r.failure
}
func (r sourceWriteFailure) UpdatePersonaSources(func([]string) ([]string, error)) error {
	return r.failure
}
func (r sourceWriteFailure) UpdatePuppetSources(func([]string) ([]string, error)) error {
	return r.failure
}
func (r sourceWriteFailure) UpdateWorkspaceSources(func([]config.WorkspaceSource) ([]config.WorkspaceSource, error)) error {
	return r.failure
}

func TestSourceWriteFailureNeverClaimsSaved(t *testing.T) {
	for _, kind := range []contracts.SourceKind{contracts.SourceSkills, contracts.SourcePersonas, contracts.SourcePuppets, contracts.SourceWorkspaces} {
		t.Run(string(kind), func(t *testing.T) {
			repo, err := config.NewFileRepository(t.TempDir())
			if err != nil {
				t.Fatal(err)
			}
			failure := errors.New("source write failed")
			service := api.New(sourceWriteFailure{FileRepository: repo, failure: failure}, nil)
			result := service.MutateSource(context.Background(), contracts.SourceMutation{Kind: kind, Path: t.TempDir()})
			if result.Saved || result.Refresh || !errors.Is(result.Err, failure) {
				t.Fatalf("result: %+v", result)
			}
			rows, readErr := service.ReadSources(kind)
			if readErr != nil || len(rows) != 0 {
				t.Fatalf("rows: %v %v", rows, readErr)
			}
		})
	}
}
