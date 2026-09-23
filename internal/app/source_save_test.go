package app_test

import (
	"os"
	"path/filepath"
	"testing"

	"kankuro/internal/app"
	"kankuro/internal/config"
	"kankuro/internal/session"
	"kankuro/internal/ui/messages"
	"kankuro/internal/ui/views"
)

// TestSaveSkillSourceAbortsOnMalformedRegistry verifies #5: a save that
// cannot read the current source list must not durably overwrite it with a
// truncated one-element list. Before the fix, saveStringSource discarded the
// list-read error and happily wrote `[newPath]` over whatever was on disk.
func TestSaveSkillSourceAbortsOnMalformedRegistry(t *testing.T) {
	repo, err := config.NewFileRepository(t.TempDir())
	if err != nil {
		t.Fatalf("new repo: %v", err)
	}
	sm := session.NewManager(repo.Dir(), repo)
	coord := app.NewCoordinator(repo, sm)
	coord.Init()

	sourcesPath := filepath.Join(repo.Dir(), config.SkillsSourcesFilename)
	original, err := os.ReadFile(sourcesPath)
	if err != nil {
		t.Fatalf("read original sources file: %v", err)
	}

	malformed := []byte("this: is: not: valid: yaml: [")
	if writeErr := os.WriteFile(sourcesPath, malformed, 0o600); writeErr != nil {
		t.Fatalf("corrupt sources file: %v", writeErr)
	}

	callbacks := coord.SourceManagerCallbacks(views.SourceTypeSkills)
	cmd, err := callbacks.SaveSimpleSource("/some/new/skill/source", -1)
	if err != nil {
		t.Fatalf("SaveSimpleSource returned error building cmd: %v", err)
	}
	if cmd == nil {
		t.Fatal("expected a non-nil command")
	}

	msg := cmd()
	if _, ok := msg.(messages.ErrorMsg); !ok {
		t.Fatalf("expected ErrorMsg, got %#v", msg)
	}

	after, err := os.ReadFile(sourcesPath)
	if err != nil {
		t.Fatalf("read sources file after failed save: %v", err)
	}
	if string(after) != string(malformed) {
		t.Fatalf("save must not touch the file on read failure: got %q, want unchanged %q", after, malformed)
	}
	if string(after) == string(original) {
		t.Fatal("test setup invariant broken: malformed content equals original")
	}
}

// TestDeleteSourceAbortsOnMalformedRegistry is TestSaveSkillSourceAbortsOnMalformedRegistry's
// counterpart for the delete path (a separate code site: deleteSource /
// UpdateSkillSources called with a delete-by-index mutator instead of an
// append/edit one).
func TestDeleteSourceAbortsOnMalformedRegistry(t *testing.T) {
	repo, err := config.NewFileRepository(t.TempDir())
	if err != nil {
		t.Fatalf("new repo: %v", err)
	}
	sm := session.NewManager(repo.Dir(), repo)
	coord := app.NewCoordinator(repo, sm)
	coord.Init()

	if saveErr := repo.SaveSkillSources([]string{"/one", "/two"}); saveErr != nil {
		t.Fatalf("seed sources: %v", saveErr)
	}

	sourcesPath := filepath.Join(repo.Dir(), config.SkillsSourcesFilename)
	malformed := []byte("this: is: not: valid: yaml: [")
	if writeErr := os.WriteFile(sourcesPath, malformed, 0o600); writeErr != nil {
		t.Fatalf("corrupt sources file: %v", writeErr)
	}

	dashboardless := coord.SourceManagerCallbacks(views.SourceTypeSkills)
	cmd, err := dashboardless.DeleteSource(0)
	if err != nil {
		t.Fatalf("DeleteSource returned error building cmd: %v", err)
	}
	if cmd == nil {
		t.Fatal("expected a non-nil command")
	}

	msg := cmd()
	if _, ok := msg.(messages.ErrorMsg); !ok {
		t.Fatalf("expected ErrorMsg, got %#v", msg)
	}

	after, err := os.ReadFile(sourcesPath)
	if err != nil {
		t.Fatalf("read sources file after failed delete: %v", err)
	}
	if string(after) != string(malformed) {
		t.Fatalf("delete must not touch the file on read failure: got %q, want unchanged %q", after, malformed)
	}
}

// TestSaveWorkspaceSourceAbortsOnMalformedRegistry covers the workspace
// branch of the read-error-abort guard: SaveWorkspaceSource is a separate
// code path from the string-source save above (it goes through
// UpdateWorkspaceSources, not UpdateSkillSources).
func TestSaveWorkspaceSourceAbortsOnMalformedRegistry(t *testing.T) {
	repo, err := config.NewFileRepository(t.TempDir())
	if err != nil {
		t.Fatalf("new repo: %v", err)
	}
	sm := session.NewManager(repo.Dir(), repo)
	coord := app.NewCoordinator(repo, sm)
	coord.Init()

	sourcesPath := filepath.Join(repo.Dir(), config.WorkspaceSourcesFilename)
	malformed := []byte("this: is: not: valid: yaml: [")
	if writeErr := os.WriteFile(sourcesPath, malformed, 0o600); writeErr != nil {
		t.Fatalf("corrupt sources file: %v", writeErr)
	}

	callbacks := coord.SourceManagerCallbacks(views.SourceTypeWorkspaces)
	cmd, err := callbacks.SaveWorkspaceSource(config.WorkspaceSource{Path: t.TempDir()}, -1)
	if err != nil {
		t.Fatalf("SaveWorkspaceSource returned error building cmd: %v", err)
	}
	if cmd == nil {
		t.Fatal("expected a non-nil command")
	}

	msg := cmd()
	if _, ok := msg.(messages.ErrorMsg); !ok {
		t.Fatalf("expected ErrorMsg, got %#v", msg)
	}

	after, err := os.ReadFile(sourcesPath)
	if err != nil {
		t.Fatalf("read sources file after failed save: %v", err)
	}
	if string(after) != string(malformed) {
		t.Fatalf("save must not touch the file on read failure: got %q, want unchanged %q", after, malformed)
	}
}

// TestDeleteWorkspaceSourceAbortsOnMalformedRegistry is the delete-path
// counterpart for workspaces.
func TestDeleteWorkspaceSourceAbortsOnMalformedRegistry(t *testing.T) {
	repo, err := config.NewFileRepository(t.TempDir())
	if err != nil {
		t.Fatalf("new repo: %v", err)
	}
	sm := session.NewManager(repo.Dir(), repo)
	coord := app.NewCoordinator(repo, sm)
	coord.Init()

	if saveErr := repo.SaveWorkspaceSources([]config.WorkspaceSource{{Path: t.TempDir()}}); saveErr != nil {
		t.Fatalf("seed workspace source: %v", saveErr)
	}

	sourcesPath := filepath.Join(repo.Dir(), config.WorkspaceSourcesFilename)
	malformed := []byte("this: is: not: valid: yaml: [")
	if writeErr := os.WriteFile(sourcesPath, malformed, 0o600); writeErr != nil {
		t.Fatalf("corrupt sources file: %v", writeErr)
	}

	callbacks := coord.SourceManagerCallbacks(views.SourceTypeWorkspaces)
	cmd, err := callbacks.DeleteSource(0)
	if err != nil {
		t.Fatalf("DeleteSource returned error building cmd: %v", err)
	}
	if cmd == nil {
		t.Fatal("expected a non-nil command")
	}

	msg := cmd()
	if _, ok := msg.(messages.ErrorMsg); !ok {
		t.Fatalf("expected ErrorMsg, got %#v", msg)
	}

	after, err := os.ReadFile(sourcesPath)
	if err != nil {
		t.Fatalf("read sources file after failed delete: %v", err)
	}
	if string(after) != string(malformed) {
		t.Fatalf("delete must not touch the file on read failure: got %q, want unchanged %q", after, malformed)
	}
}

// TestWorkspaceSourceIDAssignedAndPreservedOnEdit is the Step 4.2 test the
// review flagged as missing: a newly saved workspace source must get a
// non-empty, repository-assigned ID, and editing that source afterwards must
// not change its ID.
func TestWorkspaceSourceIDAssignedAndPreservedOnEdit(t *testing.T) {
	repo, err := config.NewFileRepository(t.TempDir())
	if err != nil {
		t.Fatalf("new repo: %v", err)
	}
	sm := session.NewManager(repo.Dir(), repo)
	coord := app.NewCoordinator(repo, sm)
	coord.Init()

	callbacks := coord.SourceManagerCallbacks(views.SourceTypeWorkspaces)

	wd := t.TempDir()
	cmd, err := callbacks.SaveWorkspaceSource(config.WorkspaceSource{Path: wd}, -1)
	if err != nil {
		t.Fatalf("SaveWorkspaceSource returned error building cmd: %v", err)
	}
	_ = cmd()

	sources, err := repo.ListWorkspaceSources()
	if err != nil {
		t.Fatalf("list workspace sources: %v", err)
	}
	if len(sources) != 1 {
		t.Fatalf("expected exactly one workspace source, got %d: %v", len(sources), sources)
	}
	if sources[0].ID == "" {
		t.Fatal("expected the repository to assign a non-empty ID")
	}
	originalID := sources[0].ID

	// Edit it (same index, different flag) and confirm the ID is preserved.
	edited := sources[0]
	edited.IsRoot = !edited.IsRoot
	cmd, err = callbacks.SaveWorkspaceSource(edited, 0)
	if err != nil {
		t.Fatalf("SaveWorkspaceSource (edit) returned error building cmd: %v", err)
	}
	_ = cmd()

	sources, err = repo.ListWorkspaceSources()
	if err != nil {
		t.Fatalf("list workspace sources after edit: %v", err)
	}
	if len(sources) != 1 {
		t.Fatalf("expected exactly one workspace source after edit, got %d: %v", len(sources), sources)
	}
	if sources[0].ID != originalID {
		t.Fatalf("expected ID to survive the edit: want %q, got %q", originalID, sources[0].ID)
	}
	if sources[0].IsRoot != edited.IsRoot {
		t.Fatalf("expected the edit itself to apply: want IsRoot=%v, got %v", edited.IsRoot, sources[0].IsRoot)
	}
}
