package app_test

import (
	"path/filepath"
	"strings"
	"testing"

	"kankuro/internal/app"
	"kankuro/internal/config"
	"kankuro/internal/session"
	"kankuro/internal/ui/messages"
	"kankuro/internal/ui/views"
)

// TestSaveSimpleSourceSurfacesSyncFailure is the R2 regression test: adding a
// skill source that cannot be scanned (e.g. it does not exist) must still
// persist the source and request a UI refresh to show it - the save itself
// succeeded - while also reporting an ErrorMsg so the user is not left
// thinking the source synced cleanly when skills.yaml is actually stale.
func TestSaveSimpleSourceSurfacesSyncFailure(t *testing.T) {
	repo, err := config.NewFileRepository(t.TempDir())
	if err != nil {
		t.Fatalf("new repo: %v", err)
	}
	sm := session.NewManager(repo.Dir(), repo)
	coord := app.NewCoordinator(repo, sm)
	coord.Init()

	badPath := filepath.Join(t.TempDir(), "nonexistent")
	callbacks := coord.SourceManagerCallbacks(views.SourceTypeSkills)
	cmd, err := callbacks.SaveSimpleSource(badPath, -1)
	if err != nil {
		t.Fatalf("SaveSimpleSource returned error building cmd: %v", err)
	}

	msgs := collectMsgs(t, cmd)

	var sawError, sawRefresh bool
	for _, msg := range msgs {
		switch m := msg.(type) {
		case messages.ErrorMsg:
			sawError = true
			if !strings.Contains(m.Err.Error(), "sync failed") {
				t.Fatalf("expected the error to mention the sync failure, got: %v", m.Err)
			}
		// After Plan 4, a callback running inside a tea.Cmd closure asks for
		// the refresh with DataRefreshRequestedMsg instead of building
		// refreshData() there (which would read c.views off the main
		// goroutine). Accept either so the assertion is about "a refresh was
		// requested", not about which mechanism delivers it.
		case messages.DataRefreshRequestedMsg:
			sawRefresh = true
		case messages.DataLoadedMsg:
			sawRefresh = true
		}
	}
	if !sawError {
		t.Fatalf("expected an ErrorMsg warning about the sync failure, got: %#v", msgs)
	}
	if !sawRefresh {
		t.Fatalf("expected a refresh to be requested even though sync failed, got: %#v", msgs)
	}

	sources, err := repo.ListSkillSources()
	if err != nil {
		t.Fatalf("list sources: %v", err)
	}
	if len(sources) != 1 || sources[0] != badPath {
		t.Fatalf("expected the source to be saved despite the sync failure, got %v", sources)
	}
}
