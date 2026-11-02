package configstartup //nolint:testpackage // Test current-layout validation without exporting internal paths.

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
	partial := t.TempDir()
	if err = os.WriteFile(filepath.Join(partial, "id_cards.yaml"), []byte("partial"), 0600); err != nil {
		t.Fatal(err)
	}
	if err = PrepareStartup(partial); err == nil || !strings.Contains(err.Error(), "no current layout marker") {
		t.Fatalf("partial root accepted: %v", err)
	}
}

func TestUnsupportedLayoutMarkerIsRejected(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, layoutMarker), []byte("2\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := PrepareStartup(dir); err == nil {
		t.Fatal("unsupported layout marker accepted")
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

func TestCurrentLayoutCanRestart(t *testing.T) {
	dir := t.TempDir()
	if err := PrepareStartup(dir); err != nil {
		t.Fatal(err)
	}
	before, err := os.ReadFile(filepath.Join(dir, layoutMarker))
	if err != nil {
		t.Fatal(err)
	}
	if err = PrepareStartup(dir); err != nil {
		t.Fatal(err)
	}
	after, err := os.ReadFile(filepath.Join(dir, layoutMarker))
	if err != nil || string(after) != string(before) {
		t.Fatalf("layout marker changed: %q, %v", after, err)
	}
}
