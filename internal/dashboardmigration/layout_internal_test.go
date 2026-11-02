package dashboardmigration

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/BlackGiack/kankuro/internal/layout"
)

func TestValidateLayoutRejectsUnknownMembershipWithoutRewriting(t *testing.T) {
	dir := t.TempDir()
	writeFixture(t, dir, "id_cards.yaml", "schema_version: 1\nid_cards:\n  - {id: card, name: Card}\n")
	writeFixture(t, dir, "sessions/s.json", `{"id":"s","card_id":"card"}`)
	if _, err := StageLegacy(dir); err != nil {
		t.Fatal(err)
	}
	if err := PublishStaged(dir); err != nil {
		t.Fatal(err)
	}
	if err := layout.ValidateLayout(dir); err != nil {
		t.Fatal(err)
	}
	bad := `{"schema_version":2,"id":"s","card_id":"card","dashboard_id":"unknown"}`
	writeFixture(t, dir, "sessions/s.json", bad)
	if err := layout.ValidateLayout(dir); err == nil {
		t.Fatal("unknown dashboard membership accepted")
	}
	raw, err := os.ReadFile(filepath.Join(dir, "sessions", "s.json"))
	if err != nil || string(raw) != bad {
		t.Fatalf("invalid record was changed: %q, %v", raw, err)
	}
}

func TestValidateLayoutRejectsIncompleteMigration(t *testing.T) {
	dir := t.TempDir()
	if _, err := StageLegacy(dir); err != nil {
		t.Fatal(err)
	}
	if err := layout.ValidateLayout(dir); err == nil {
		t.Fatal("unpublished migration accepted")
	}
}

func TestPrepareStartupInitializesFreshAndBlocksLegacy(t *testing.T) {
	fresh := t.TempDir()
	lock, err := layout.AcquireLock(fresh)
	if err != nil {
		t.Fatal(err)
	}
	defer lock.Close()
	if errAt50 := layout.PrepareStartup(fresh); errAt50 != nil {
		t.Fatal(errAt50)
	}
	if errAt53 := layout.ValidateLayout(fresh); errAt53 != nil {
		t.Fatal(errAt53)
	}
	if errAt56 := Rollback(fresh); errAt56 == nil {
		t.Fatal("used fresh layout rolled back")
	}
	legacy := t.TempDir()
	writeFixture(t, legacy, "id_cards.yaml", "schema_version: 1\nid_cards: []\n")
	legacyLock, err := layout.AcquireLock(legacy)
	if err != nil {
		t.Fatal(err)
	}
	defer legacyLock.Close()
	if errAt66 := layout.PrepareStartup(legacy); errAt66 == nil {
		t.Fatal("legacy layout started without migration")
	}
	if _, errAt69 := os.Stat(filepath.Join(legacy, layoutMarker)); !os.IsNotExist(errAt69) {
		t.Fatalf("legacy marker published: %v", errAt69)
	}
}
