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

// TestFingerprintMissingFileRecordsStableStamp pins review_dp_5_2.md finding
// 3: a missing registry file (EnsureFiles only runs at startup, so a
// hand-delete while the app is running is possible) must not fail the whole
// Fingerprint call. It should record a stable zero stamp instead - one that
// compares equal to itself across repeated calls, so the idle-tick
// optimization keeps working (rather than silently falling back to a full
// reload every tick for the rest of the session) - and differs the moment
// the file reappears.
func TestFingerprintMissingFileRecordsStableStamp(t *testing.T) {
	dir := t.TempDir()
	repo, err := config.NewFileRepository(dir)
	if err != nil {
		t.Fatalf("new repo: %v", err)
	}

	idCardsPath := filepath.Join(dir, config.IDCardsFilename)
	if rmErr := os.Remove(idCardsPath); rmErr != nil {
		t.Fatalf("remove id_cards.yaml: %v", rmErr)
	}

	stamps1, err := repo.Fingerprint()
	if err != nil {
		t.Fatalf("expected Fingerprint to succeed despite a missing file, got: %v", err)
	}
	stamp1, ok := stamps1[config.IDCardsFilename]
	if !ok {
		t.Fatal("expected a stamp entry for the missing file")
	}
	if stamp1 != (config.FileStamp{}) {
		t.Fatalf("expected a zero stamp for a missing file, got %+v", stamp1)
	}

	stamps2, err := repo.Fingerprint()
	if err != nil {
		t.Fatalf("expected a second Fingerprint call to also succeed, got: %v", err)
	}
	if stamps2[config.IDCardsFilename] != stamp1 {
		t.Fatalf("expected the missing-file stamp to stay stable across calls, got %+v then %+v",
			stamp1, stamps2[config.IDCardsFilename])
	}

	// Recreate the file: the stamp must now differ from the zero sentinel.
	if writeErr := os.WriteFile(idCardsPath, []byte("schema_version: 1\nid_cards: []\n"), 0o600); writeErr != nil {
		t.Fatalf("recreate id_cards.yaml: %v", writeErr)
	}
	stamps3, err := repo.Fingerprint()
	if err != nil {
		t.Fatalf("expected Fingerprint to succeed after the file reappeared, got: %v", err)
	}
	if stamps3[config.IDCardsFilename] == stamp1 {
		t.Fatal("expected the stamp to change once the missing file reappeared")
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
