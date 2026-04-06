package config_test

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"kankuro/internal/config"
)

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
