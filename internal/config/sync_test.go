package config_test

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/BlackGiack/kankuro/internal/config"
)

// TestUpdateSkillSourcesReportsSyncFailureButStillSaves is the R2 case: an
// UpdateSkillSources call whose new source cannot be scanned must still
// durably persist the source list (the user's edit is not lost) while
// telling the caller, via *config.SyncFailedError, that skills.yaml is now
// stale.
func TestUpdateSkillSourcesReportsSyncFailureButStillSaves(t *testing.T) {
	repo, err := config.NewFileRepository(t.TempDir())
	if err != nil {
		t.Fatalf("new repo: %v", err)
	}
	badPath := filepath.Join(t.TempDir(), "nonexistent")

	updateErr := repo.UpdateSkillSources(func(sources []string) ([]string, error) {
		return append(sources, badPath), nil
	})
	if updateErr == nil {
		t.Fatal("expected an error for an unreadable source")
	}
	var syncErr *config.SyncFailedError
	if !errors.As(updateErr, &syncErr) {
		t.Fatalf("expected *config.SyncFailedError, got %T: %v", updateErr, updateErr)
	}

	sources, err := repo.ListSkillSources()
	if err != nil {
		t.Fatalf("list sources: %v", err)
	}
	if len(sources) != 1 || sources[0] != badPath {
		t.Fatalf("expected the source list to still contain %q despite the sync failure, got %v", badPath, sources)
	}
}

func TestScanEntriesRejectsEmptyPath(t *testing.T) {
	err := config.ScanEntries([]string{""}, func(os.DirEntry) bool { return true }, func(_, _ string, _ os.DirEntry) error {
		return nil
	})
	if err == nil {
		t.Fatal("expected error for empty path")
	}
	if !strings.Contains(err.Error(), `""`) {
		t.Fatalf("expected error to mention source path, got: %v", err)
	}
}

func TestScanEntriesRejectsInvalidPath(t *testing.T) {
	badPath := filepath.Join(t.TempDir(), "nonexistent")
	err := config.ScanEntries([]string{badPath}, func(os.DirEntry) bool { return true }, func(_, _ string, _ os.DirEntry) error {
		return nil
	})
	if err == nil {
		t.Fatal("expected error for nonexistent path")
	}
	if !strings.Contains(err.Error(), badPath) {
		t.Fatalf("expected error to mention source path %q, got: %v", badPath, err)
	}
}

func TestScanEntriesPropagatesCallbackError(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "file"), nil, 0o644); err != nil {
		t.Fatal(err)
	}
	cbErr := errors.New("callback failed")
	err := config.ScanEntries([]string{dir}, func(os.DirEntry) bool { return true }, func(_, _ string, _ os.DirEntry) error {
		return cbErr
	})
	if !errors.Is(err, cbErr) {
		t.Fatalf("expected callback error, got: %v", err)
	}
}

func TestScanEntriesProcessesEntries(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "sub"), 0o755); err != nil {
		t.Fatal(err)
	}

	var names []string
	err := config.ScanEntries([]string{dir},
		func(e os.DirEntry) bool { return e.IsDir() },
		func(_, _ string, entry os.DirEntry) error {
			names = append(names, entry.Name())
			return nil
		},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(names) != 1 || names[0] != "sub" {
		t.Fatalf("expected [sub], got %v", names)
	}
}

func TestScanEntriesWithSkipSkipsInvalidSources(t *testing.T) {
	validDir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(validDir, "item"), 0o755); err != nil {
		t.Fatal(err)
	}
	badPath := filepath.Join(t.TempDir(), "nonexistent")

	var names []string
	err := config.ScanEntriesWithSkip([]string{badPath, validDir},
		func(e os.DirEntry) bool { return e.IsDir() },
		func(_, _ string, entry os.DirEntry) error {
			names = append(names, entry.Name())
			return nil
		},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(names) != 1 || names[0] != "item" {
		t.Fatalf("expected [item], got %v", names)
	}
}

func TestScanEntriesWithSkipPropagatesCallbackError(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "file"), nil, 0o644); err != nil {
		t.Fatal(err)
	}
	cbErr := errors.New("callback failed")
	err := config.ScanEntriesWithSkip([]string{dir},
		func(os.DirEntry) bool { return true },
		func(_, _ string, _ os.DirEntry) error {
			return cbErr
		},
	)
	if !errors.Is(err, cbErr) {
		t.Fatalf("expected callback error, got: %v", err)
	}
}

func TestSyncSkillsFromSourcesReportsInvalidSource(t *testing.T) {
	repo, err := config.NewFileRepository(t.TempDir())
	if err != nil {
		t.Fatalf("new repo: %v", err)
	}
	badPath := filepath.Join(t.TempDir(), "nonexistent")
	if errWrite := repo.SaveSkillSources([]string{badPath}); errWrite != nil {
		t.Fatalf("save sources: %v", errWrite)
	}

	err = repo.SyncSkillsFromSources()
	if err == nil {
		t.Fatal("expected error for invalid source")
	}
	if !strings.Contains(err.Error(), badPath) {
		t.Fatalf("expected error to mention source path %q, got: %v", badPath, err)
	}
}

// TestExpandPathRejectsUndefinedEnvVar guards against [os.ExpandEnv]'s silent
// collapse-to-empty-string behavior: "$UNSET_VAR/skills" must error, not
// resolve to "/skills" and scan the filesystem root.
func TestExpandPathRejectsUndefinedEnvVar(t *testing.T) {
	const undefinedVar = "KANKURO_TEST_DEFINITELY_UNSET_VAR"
	if _, ok := os.LookupEnv(undefinedVar); ok {
		t.Fatalf("test invariant broken: %s is set in the environment", undefinedVar)
	}

	repo, err := config.NewFileRepository(t.TempDir())
	if err != nil {
		t.Fatalf("new repo: %v", err)
	}
	badPath := "$" + undefinedVar + "/skills"
	if errWrite := repo.SaveSkillSources([]string{badPath}); errWrite != nil {
		t.Fatalf("save sources: %v", errWrite)
	}

	err = repo.SyncSkillsFromSources()
	if err == nil {
		t.Fatal("expected error for source with an undefined env var")
	}
	if !strings.Contains(err.Error(), undefinedVar) {
		t.Fatalf("expected error to mention the undefined variable %q, got: %v", undefinedVar, err)
	}
}

func TestSyncPersonasFromSourcesReportsInvalidSource(t *testing.T) {
	repo, err := config.NewFileRepository(t.TempDir())
	if err != nil {
		t.Fatalf("new repo: %v", err)
	}
	badPath := filepath.Join(t.TempDir(), "nonexistent")
	if errWrite := repo.SavePersonaSources([]string{badPath}); errWrite != nil {
		t.Fatalf("save sources: %v", errWrite)
	}

	err = repo.SyncPersonasFromSources()
	if err == nil {
		t.Fatal("expected error for invalid source")
	}
	if !strings.Contains(err.Error(), badPath) {
		t.Fatalf("expected error to mention source path %q, got: %v", badPath, err)
	}
}
