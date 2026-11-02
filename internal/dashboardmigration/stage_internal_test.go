package dashboardmigration

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/BlackGiack/kankuro/internal/layout"
)

func TestStageLegacyPreservesSourcesAndBackups(t *testing.T) {
	dir := t.TempDir()
	cards := "schema_version: 1\nid_cards:\n  - {id: a, name: A}\n"
	session := `{"id":"s","card_id":"a","working_dir":"/workspace"}`
	writeFixture(t, dir, "id_cards.yaml", cards)
	writeFixture(t, dir, "sessions/s.json", session)
	lock, err := layout.AcquireLock(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer lock.Close()
	manifest, err := StageLegacy(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(manifest.Files) != 5 {
		t.Fatalf("files: %+v", manifest.Files)
	}
	for _, file := range manifest.Files {
		output, errAt29 := os.ReadFile(filepath.Join(dir, stageDirectory, "output", file.Path))
		if errAt29 != nil || digest(output) != file.OutputSHA256 {
			t.Fatalf("staged output %s: %v", file.Path, errAt29)
		}
		if file.Existed {
			backup, errAt34 := os.ReadFile(filepath.Join(dir, stageDirectory, "backup", file.Path))
			if errAt34 != nil || digest(backup) != file.OriginalSHA256 {
				t.Fatalf("backup %s: %v", file.Path, errAt34)
			}
			original, errAt34 := os.ReadFile(filepath.Join(dir, file.Path))
			if errAt34 != nil || !bytes.Equal(backup, original) {
				t.Fatalf("source changed: %s: %v", file.Path, errAt34)
			}
		}
	}
	if _, errAt44 := os.Stat(filepath.Join(dir, "dashboards.yaml")); !os.IsNotExist(errAt44) {
		t.Fatalf("dashboard was published: %v", errAt44)
	}
	if _, errAt47 := StageLegacy(dir); errAt47 == nil {
		t.Fatal("existing stage was overwritten")
	}
}

func TestStageLegacyRejectsInvalidInputWithoutStaging(t *testing.T) {
	dir := t.TempDir()
	writeFixture(t, dir, "id_cards.yaml", "schema_version: 99\nid_cards: []\n")
	if _, err := StageLegacy(dir); err == nil {
		t.Fatal("invalid input accepted")
	}
	if _, err := os.Stat(filepath.Join(dir, stageDirectory)); !os.IsNotExist(err) {
		t.Fatalf("stage exists: %v", err)
	}
}
