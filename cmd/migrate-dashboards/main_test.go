package main

import (
	"bytes"
	"encoding/json"

	"os"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"

	"github.com/BlackGiack/kankuro/internal/layout"
)

func TestOfflineMigrationDryRunApplyAndRollback(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "id_cards.yaml")
	original := "schema_version: 1\nid_cards:\n  - {id: card, name: Card}\n"
	if err := os.WriteFile(path, []byte(original), 0o600); err != nil {
		t.Fatal(err)
	}
	var output bytes.Buffer
	if err := run([]string{"-config-dir", dir}, &output); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(output.String(), "Dry run: 1 cards") {
		t.Fatalf("dry run: %s", output.String())
	}
	raw, err := os.ReadFile(path)
	if err != nil || string(raw) != original {
		t.Fatalf("dry run changed source: %q, %v", raw, err)
	}
	output.Reset()
	if errAt32 := run([]string{"-config-dir", dir, "-apply"}, &output); errAt32 != nil {
		t.Fatal(errAt32)
	}
	if !strings.Contains(output.String(), "Backup:") {
		t.Fatalf("backup not reported: %s", output.String())
	}
	if errAt38 := layout.ValidateLayout(dir); errAt38 != nil {
		t.Fatal(errAt38)
	}
	if errAt41 := run([]string{"-config-dir", dir, "-apply"}, &output); errAt41 != nil {
		t.Fatalf("repeated apply: %v", errAt41)
	}
	if errAt44 := run([]string{"-config-dir", dir, "-rollback"}, &output); errAt44 != nil {
		t.Fatal(errAt44)
	}
	raw, err = os.ReadFile(path)
	if err != nil || string(raw) != original {
		t.Fatalf("rollback: %q, %v", raw, err)
	}
}

func TestOfflineMigrationRefusesConcurrentWriter(t *testing.T) {
	dir := t.TempDir()
	lock, err := layout.AcquireLock(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer lock.Close()
	if errAt60 := run([]string{"-config-dir", dir}, &bytes.Buffer{}); errAt60 == nil {
		t.Fatal("migration ran while writer lock held")
	}
}

//nolint:gocognit,gocyclo,cyclop,govet // Verify the complete offline conversion and rollback transaction.
func TestOfflineMigrationLegacyCardDefaults(t *testing.T) {
	cases := []struct{ name, cards, session string }{
		{"missing version", "id_cards: []\nextension: keep\n", ""},
		{"zero version", "schema_version: 0\nid_cards: []\nextension: keep\n", ""},
		{"missing cards", "schema_version: 1\nextension: keep\n", ""},
		{"null cards", "schema_version: 1\nid_cards: null\nextension: keep\n", ""},
		{
			"existing session and binding",
			"id_cards:\n  - id: card\n    name: Card\n    extension: {nested: keep}\nextension: keep\n",
			`{"id":"session","card_id":"card","working_dir":"/workspace/copy","original_dir":"/workspace/original","conversations":{"card":[{"id":"native","driver":"cat"}]},"extension":{"keep":true}}`,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			cardsPath := filepath.Join(dir, "id_cards.yaml")
			if err := os.WriteFile(cardsPath, []byte(tc.cards), 0o640); err != nil {
				t.Fatal(err)
			}
			sessionPath := filepath.Join(dir, "sessions", "session.json")
			if tc.session != "" {
				if err := os.MkdirAll(filepath.Dir(sessionPath), 0o700); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(sessionPath, []byte(tc.session), 0o600); err != nil {
					t.Fatal(err)
				}
			}
			var output bytes.Buffer
			if err := run([]string{"-config-dir", dir}, &output); err != nil {
				t.Fatalf("dry run: %v", err)
			}
			original, err := os.ReadFile(cardsPath)
			if err != nil || string(original) != tc.cards {
				t.Fatalf("dry run modified source: %q %v", original, err)
			}
			output.Reset()
			if err := run([]string{"-config-dir", dir, "-apply"}, &output); err != nil {
				t.Fatalf("apply: %v", err)
			}
			if err := layout.ValidateLayout(dir); err != nil {
				t.Fatalf("current layout: %v", err)
			}
			backup, err := os.ReadFile(filepath.Join(dir, ".dashboard-migration", "backup", "id_cards.yaml"))
			if err != nil || string(backup) != tc.cards {
				t.Fatalf("card backup: %q %v", backup, err)
			}
			raw, err := os.ReadFile(cardsPath)
			if err != nil {
				t.Fatal(err)
			}
			var cards struct {
				Version   int              `yaml:"schema_version"`
				Cards     []map[string]any `yaml:"id_cards"`
				Extension string           `yaml:"extension"`
			}
			if err := yaml.Unmarshal(raw, &cards); err != nil {
				t.Fatal(err)
			}
			if cards.Version != 2 || cards.Cards == nil || cards.Extension != "keep" {
				t.Fatalf("cards: %s", raw)
			}
			//nolint:nestif // Verify the full preserved session and binding fixture.
			if tc.session != "" {
				if len(cards.Cards) != 1 || cards.Cards[0]["id"] != "card" || cards.Cards[0]["dashboard_id"] != "default" ||
					cards.Cards[0]["shared"] != false ||
					cards.Cards[0]["extension"].(map[string]any)["nested"] != "keep" {
					t.Fatalf("card changed: %s", raw)
				}
				sessionBackup, err := os.ReadFile(filepath.Join(dir, ".dashboard-migration", "backup", "sessions", "session.json"))
				if err != nil || string(sessionBackup) != tc.session {
					t.Fatalf("session backup: %q %v", sessionBackup, err)
				}
				sessionRaw, err := os.ReadFile(sessionPath)
				if err != nil {
					t.Fatal(err)
				}
				var record map[string]json.RawMessage
				if err := json.Unmarshal(sessionRaw, &record); err != nil {
					t.Fatal(err)
				}
				var before map[string]json.RawMessage
				if err := json.Unmarshal([]byte(tc.session), &before); err != nil {
					t.Fatal(err)
				}
				for key, value := range before {
					var expected, actual bytes.Buffer
					if err := json.Compact(&expected, value); err != nil {
						t.Fatal(err)
					}
					if err := json.Compact(&actual, record[key]); err != nil {
						t.Fatal(err)
					}
					if !bytes.Equal(expected.Bytes(), actual.Bytes()) {
						t.Fatalf("session field %s changed: %s vs %s", key, value, record[key])
					}
				}
				if string(record["dashboard_id"]) != `"default"` || string(record["schema_version"]) != "2" {
					t.Fatalf("membership: %s", sessionRaw)
				}
			} else if len(cards.Cards) != 0 {
				t.Fatalf("unexpected cards: %s", raw)
			}
			if err := run([]string{"-config-dir", dir, "-rollback"}, &output); err != nil {
				t.Fatalf("rollback: %v", err)
			}
			restored, err := os.ReadFile(cardsPath)
			if err != nil || string(restored) != tc.cards {
				t.Fatalf("restored cards: %q %v", restored, err)
			}
			if tc.session != "" {
				restored, err := os.ReadFile(sessionPath)
				if err != nil || string(restored) != tc.session {
					t.Fatalf("restored session: %q %v", restored, err)
				}
			}
		})
	}
}
