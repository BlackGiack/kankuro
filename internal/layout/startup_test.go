package layout //nolint:testpackage // Test current-layout fences without exporting internal paths.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFreshInitializationAndPartialRoot(t *testing.T) {
	dir := t.TempDir()
	lock, err := AcquireLock(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer lock.Close()
	if err = PrepareStartup(dir); err != nil {
		t.Fatal(err)
	}
	if err = ValidateLayout(dir); err != nil {
		t.Fatal(err)
	}
	if _, err = os.Stat(filepath.Join(dir, ".dashboard-migration")); !os.IsNotExist(err) {
		t.Fatalf("fresh root contains migration stage: %v", err)
	}
	if _, err = os.Stat(filepath.Join(dir, UsedMarkerFilename)); err != nil {
		t.Fatal(err)
	}
	partial := t.TempDir()
	if err = os.WriteFile(filepath.Join(partial, "id_cards.yaml"), []byte("partial"), 0600); err != nil {
		t.Fatal(err)
	}
	if err = PrepareStartup(partial); err == nil || !strings.Contains(err.Error(), "no completed layout marker") {
		t.Fatalf("partial root accepted: %v", err)
	}
}

func TestLegacyMarkerWithoutFirstUseIsRejected(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, MarkerFilename), []byte("2\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := PrepareStartup(dir); err == nil {
		t.Fatal("unfenced legacy migration accepted")
	}
}

func TestExclusiveLock(t *testing.T) {
	dir := t.TempDir()
	first, err := AcquireLock(dir)
	if err != nil {
		t.Fatal(err)
	}
	if second, secondErr := AcquireLock(dir); secondErr == nil {
		if closeErr := second.Close(); closeErr != nil {
			t.Fatal(closeErr)
		}
		t.Fatal("second writer acquired lock")
	}
	if err = first.Close(); err != nil {
		t.Fatal(err)
	}
	second, err := AcquireLock(dir)
	if err != nil {
		t.Fatal(err)
	}
	if err = second.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestPreviouslyUsedVersionTwoLayoutStartsWithoutMigrationFiles(t *testing.T) {
	dir := t.TempDir()
	if err := PrepareStartup(dir); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, MarkerFilename), []byte("2\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := PrepareStartup(dir); err != nil {
		t.Fatalf("previously used layout refused: %v", err)
	}
}
