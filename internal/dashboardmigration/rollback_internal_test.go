package dashboardmigration

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/BlackGiack/kankuro/internal/layout"
)

func TestRollbackInterruptedRestoresSources(t *testing.T) {
	dir := t.TempDir()
	original := "schema_version: 1\nid_cards: []\n"
	writeFixture(t, dir, "id_cards.yaml", original)
	if err := os.Chmod(filepath.Join(dir, "id_cards.yaml"), 0o640); err != nil {
		t.Fatal(err)
	}
	manifest, err := StageLegacy(dir)
	if err != nil {
		t.Fatal(err)
	}
	stage := filepath.Join(dir, stageDirectory)
	for _, item := range manifest.Files {
		output, readErr := os.ReadFile(filepath.Join(stage, "output", item.Path))
		if readErr != nil {
			t.Fatal(readErr)
		}
		writeFixture(t, dir, item.Path, string(output))
	}
	if errAt25 := RollbackStaged(dir); errAt25 != nil {
		t.Fatal(errAt25)
	}
	raw, err := os.ReadFile(filepath.Join(dir, "id_cards.yaml"))
	if err != nil || string(raw) != original {
		t.Fatalf("restored cards: %q, %v", raw, err)
	}
	info, err := os.Stat(filepath.Join(dir, "id_cards.yaml"))
	if err != nil || info.Mode().Perm() != 0o640 {
		t.Fatalf("restored mode: %v, %v", info, err)
	}
	if _, errAt32 := os.Stat(filepath.Join(dir, dashboardFileName)); !os.IsNotExist(errAt32) {
		t.Fatalf("dashboard remains: %v", errAt32)
	}
	if _, errAt35 := os.Stat(stage); !os.IsNotExist(errAt35) {
		t.Fatalf("stage remains: %v", errAt35)
	}
}

func TestRollbackRejectsChangedSourceAndUsedMigration(t *testing.T) {
	dir := t.TempDir()
	writeFixture(t, dir, "id_cards.yaml", "schema_version: 1\nid_cards: []\n")
	if _, err := StageLegacy(dir); err != nil {
		t.Fatal(err)
	}
	writeFixture(t, dir, "id_cards.yaml", "changed")
	if err := RollbackStaged(dir); err == nil {
		t.Fatal("changed source accepted")
	}
	writeFixture(t, dir, layoutUsedMarker, "used\n")
	if err := RollbackStaged(dir); err == nil {
		t.Fatal("used layout rollback accepted")
	}
}

func TestRollbackCompletedBeforeFirstUse(t *testing.T) {
	dir := t.TempDir()
	original := "schema_version: 1\nid_cards: []\n"
	writeFixture(t, dir, "id_cards.yaml", original)
	if _, err := StageLegacy(dir); err != nil {
		t.Fatal(err)
	}
	if err := PublishStaged(dir); err != nil {
		t.Fatal(err)
	}
	if err := RollbackStaged(dir); err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(filepath.Join(dir, "id_cards.yaml"))
	if err != nil || string(raw) != original {
		t.Fatalf("restored cards: %q, %v", raw, err)
	}
	if _, errAt80 := os.Stat(filepath.Join(dir, layoutMarker)); !os.IsNotExist(errAt80) {
		t.Fatalf("marker remains: %v", errAt80)
	}
}

func TestMarkLayoutUsedPreventsRollback(t *testing.T) {
	dir := t.TempDir()
	writeFixture(t, dir, "id_cards.yaml", "schema_version: 1\nid_cards: []\n")
	if _, err := StageLegacy(dir); err != nil {
		t.Fatal(err)
	}
	if err := PublishStaged(dir); err != nil {
		t.Fatal(err)
	}
	if err := layout.MarkLayoutUsed(dir); err != nil {
		t.Fatal(err)
	}
	if err := RollbackStaged(dir); err == nil {
		t.Fatal("used layout rollback accepted")
	}
}

func TestRollbackDiscardsUnpublishedIncompleteStage(t *testing.T) {
	dir := t.TempDir()
	stage := filepath.Join(dir, stageDirectory)
	if err := os.Mkdir(stage, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(stage, stageSentinel), []byte("kankuro-dashboard-migration-v1\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	writeFixture(t, stage, "output/partial", "data")
	if err := Rollback(dir); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(stage); !os.IsNotExist(err) {
		t.Fatalf("stage remains: %v", err)
	}
}

func TestRollbackFenceBlocksRuntimeBeforeRollbackCheckpoint(t *testing.T) {
	dir := t.TempDir()
	writeFixture(t, dir, "id_cards.yaml", "schema_version: 1\nid_cards: []\n")
	if _, err := StageLegacy(dir); err != nil {
		t.Fatal(err)
	}
	if err := PublishStaged(dir); err != nil {
		t.Fatal(err)
	}
	writeFixture(t, dir, layoutMarker, "rollback\n")
	if err := layout.PrepareStartup(dir); err == nil {
		t.Fatal("runtime accepted rollback fence")
	}
	if err := RollbackStaged(dir); err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(filepath.Join(dir, "id_cards.yaml"))
	if err != nil || string(raw) != "schema_version: 1\nid_cards: []\n" {
		t.Fatalf("rollback restoration: %q %v", raw, err)
	}
}
