package config_test

import (
	"os"
	"path/filepath"
	"sync"
	"testing"

	"kankuro/internal/config"

	"gopkg.in/yaml.v3"
)

func TestWriteAtomicLeavesNoTempResidue(t *testing.T) {
	dir := t.TempDir()
	repo, err := config.NewFileRepository(dir)
	if err != nil {
		t.Fatalf("new repo: %v", err)
	}

	path := filepath.Join(dir, "probe.yaml")
	if writeErr := repo.WriteAtomic(path, map[string]string{"a": "b"}); writeErr != nil {
		t.Fatalf("write atomic: %v", writeErr)
	}

	matches, err := filepath.Glob(filepath.Join(dir, "*.tmp"))
	if err != nil {
		t.Fatalf("glob: %v", err)
	}
	if len(matches) != 0 {
		t.Fatalf("expected no leftover temp files, found %v", matches)
	}
}

func TestWriteAtomicMode(t *testing.T) {
	dir := t.TempDir()
	repo, err := config.NewFileRepository(dir)
	if err != nil {
		t.Fatalf("new repo: %v", err)
	}

	path := filepath.Join(dir, "probe.yaml")
	if writeErr := repo.WriteAtomic(path, map[string]string{"a": "b"}); writeErr != nil {
		t.Fatalf("write atomic: %v", writeErr)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat: %v", err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("expected mode 0600, got %o", info.Mode().Perm())
	}
}

func TestWriteAtomicConcurrentWritesNeverSplice(t *testing.T) {
	dir := t.TempDir()
	repo, err := config.NewFileRepository(dir)
	if err != nil {
		t.Fatalf("new repo: %v", err)
	}

	path := filepath.Join(dir, "probe.yaml")
	const n = 20
	var wg sync.WaitGroup
	errs := make(chan error, n)
	for i := range n {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			if writeErr := repo.WriteAtomic(path, map[string]int{"value": i}); writeErr != nil {
				errs <- writeErr
			}
		}(i)
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		t.Fatalf("concurrent write failed: %v", err)
	}

	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read final file: %v", err)
	}
	var out map[string]int
	if unmarshalErr := yaml.Unmarshal(raw, &out); unmarshalErr != nil {
		t.Fatalf("final file is not valid yaml (spliced write?): %v\ncontent:\n%s", unmarshalErr, raw)
	}

	matches, err := filepath.Glob(filepath.Join(dir, "*.tmp"))
	if err != nil {
		t.Fatalf("glob: %v", err)
	}
	if len(matches) != 0 {
		t.Fatalf("expected no leftover temp files, found %v", matches)
	}
}
