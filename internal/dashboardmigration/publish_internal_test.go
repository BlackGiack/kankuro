package dashboardmigration

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/BlackGiack/kankuro/internal/layout"
)

//nolint:gocognit // Exercise every publication boundary with the same fixture.
func TestPublishResumesAfterReplacementBeforeManifestWrite(t *testing.T) {
	for boundary := range 5 {
		t.Run(string(rune('0'+boundary)), func(t *testing.T) {
			dir := t.TempDir()
			writeFixture(t, dir, "id_cards.yaml", "schema_version: 1\nid_cards:\n  - {id: a}\n")
			writeFixture(t, dir, "sessions/s.json", `{"id":"s","card_id":"a"}`)
			manifest, err := StageLegacy(dir)
			if err != nil {
				t.Fatal(err)
			}
			stage := filepath.Join(dir, stageDirectory)
			for index := 0; index <= boundary; index++ {
				item := &manifest.Files[index]
				raw, errAt22 := os.ReadFile(filepath.Join(stage, "output", item.Path))
				if errAt22 != nil {
					t.Fatal(errAt22)
				}
				writeFixture(t, dir, item.Path, string(raw))
				if index < boundary {
					item.Published = true
				}
			}
			if errAt31 := writeManifest(stage, manifest); errAt31 != nil {
				t.Fatal(errAt31)
			}
			if errAt34 := PublishStaged(dir); errAt34 != nil {
				t.Fatal(errAt34)
			}
			if errAt37 := PublishStaged(dir); errAt37 != nil {
				t.Fatalf("repeat publish: %v", errAt37)
			}
			marker, err := os.ReadFile(filepath.Join(dir, layoutMarker))
			if err != nil || string(marker) != layout.CurrentMarker {
				t.Fatalf("marker: %q, %v", marker, err)
			}
			complete, err := readManifest(stage)
			if err != nil || !complete.Complete {
				t.Fatalf("manifest: %+v, %v", complete, err)
			}
		})
	}
}

func TestPublishRefusesChangedSourceBeforeEffects(t *testing.T) {
	dir := t.TempDir()
	writeFixture(t, dir, "id_cards.yaml", "schema_version: 1\nid_cards: []\n")
	if _, err := StageLegacy(dir); err != nil {
		t.Fatal(err)
	}
	writeFixture(t, dir, "id_cards.yaml", "modified")
	if err := PublishStaged(dir); err == nil {
		t.Fatal("changed source accepted")
	}
	if _, err := os.Stat(filepath.Join(dir, "dashboards.yaml")); !os.IsNotExist(err) {
		t.Fatalf("published before validation: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, layoutMarker)); !os.IsNotExist(err) {
		t.Fatalf("marker published: %v", err)
	}
}

func TestPublishRejectsUnexpectedMarkerBeforeEffects(t *testing.T) {
	dir := t.TempDir()
	writeFixture(t, dir, "id_cards.yaml", "schema_version: 1\nid_cards: []\n")
	if _, err := StageLegacy(dir); err != nil {
		t.Fatal(err)
	}
	writeFixture(t, dir, layoutMarker, "foreign\n")
	if err := PublishStaged(dir); err == nil {
		t.Fatal("unexpected marker accepted")
	}
	if _, err := os.Stat(filepath.Join(dir, dashboardFileName)); !os.IsNotExist(err) {
		t.Fatalf("published before validation: %v", err)
	}
}

func TestPublishRejectsChangedHistoryBeforeEffects(t *testing.T) {
	dir := t.TempDir()
	writeFixture(t, dir, "id_cards.yaml", "schema_version: 1\nid_cards: []\n")
	writeFixture(t, dir, "messaging/history.json", `{"messages":[]}`)
	if _, err := StageLegacy(dir); err != nil {
		t.Fatal(err)
	}
	writeFixture(t, dir, "messaging/history.json", `{"messages":[],"sequence":1}`)
	if err := PublishStaged(dir); err == nil {
		t.Fatal("changed history accepted")
	}
	if _, err := os.Stat(filepath.Join(dir, dashboardFileName)); !os.IsNotExist(err) {
		t.Fatalf("published before validation: %v", err)
	}
}

func TestPublishedLayoutWorksAfterArchivingMigrationStage(t *testing.T) {
	dir := t.TempDir()
	writeFixture(t, dir, "id_cards.yaml", "schema_version: 1\nid_cards: []\n")
	if _, err := StageLegacy(dir); err != nil {
		t.Fatal(err)
	}
	if err := PublishStaged(dir); err != nil {
		t.Fatal(err)
	}
	if err := os.Rename(filepath.Join(dir, stageDirectory), filepath.Join(t.TempDir(), "archived-stage")); err != nil {
		t.Fatal(err)
	}
	if err := layout.PrepareStartup(dir); err != nil {
		t.Fatalf("archived backup blocked runtime startup: %v", err)
	}
}

func TestCompletedCheckpointWithoutFenceDoesNotStart(t *testing.T) {
	dir := t.TempDir()
	writeFixture(t, dir, "id_cards.yaml", "schema_version: 1\nid_cards: []\n")
	manifest, err := StageLegacy(dir)
	if err != nil {
		t.Fatal(err)
	}
	stage := filepath.Join(dir, stageDirectory)
	for i := range manifest.Files {
		item := &manifest.Files[i]
		raw, readErr := os.ReadFile(filepath.Join(stage, "output", item.Path))
		if readErr != nil {
			t.Fatal(readErr)
		}
		writeFixture(t, dir, item.Path, string(raw))
		item.Published = true
	}
	manifest.Complete = true
	if err = writeManifest(stage, manifest); err != nil {
		t.Fatal(err)
	}
	if err = layout.PrepareStartup(dir); err == nil {
		t.Fatal("unfenced migration started")
	}
	if err = PublishStaged(dir); err != nil {
		t.Fatal(err)
	}
	if err = layout.PrepareStartup(dir); err != nil {
		t.Fatal(err)
	}
}

func TestApplyUpgradesOldUnfencedCompletedMigration(t *testing.T) {
	dir := t.TempDir()
	writeFixture(t, dir, "id_cards.yaml", "schema_version: 1\nid_cards: []\n")
	if _, err := StageLegacy(dir); err != nil {
		t.Fatal(err)
	}
	if err := PublishStaged(dir); err != nil {
		t.Fatal(err)
	}
	writeFixture(t, dir, layoutMarker, "2\n")
	if err := layout.PrepareStartup(dir); err == nil {
		t.Fatal("old unfenced layout unexpectedly started")
	}
	if _, _, err := Apply(dir); err != nil {
		t.Fatal(err)
	}
	if err := layout.PrepareStartup(dir); err != nil {
		t.Fatal(err)
	}
}
