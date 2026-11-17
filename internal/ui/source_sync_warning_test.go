package ui_test

import (
	"context"
	"path/filepath"
	"strings"
	"testing"

	"github.com/BlackGiack/kankuro/internal/api/contracts"
	"github.com/BlackGiack/kankuro/internal/ui/views"
)

func TestSaveSimpleSourceSurfacesSyncFailure(t *testing.T) {
	repo, err := newCurrentRepository(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	callbacks := newScopedCoordinator(repo, newTestManager(t, repo.Dir(), repo)).SourceManagerCallbacks(views.SourceTypeSkills)
	path := filepath.Join(t.TempDir(), "nonexistent")
	result := callbacks.Mutate(context.Background(), contracts.SourceMutation{Kind: contracts.SourceSkills, Path: path})
	if !result.Saved || !result.Refresh || result.Warning == nil || !strings.Contains(result.Warning.Error(), "sync failed") {
		t.Fatalf("result: %+v", result)
	}
	sources, err := repo.ListSkillSources()
	if err != nil || len(sources) != 1 || sources[0] != path {
		t.Fatalf("sources: %v %v", sources, err)
	}
}
