package configstartup //nolint:testpackage // Exercise current-format validation with internal marker constants.

import (
	"os"
	"path/filepath"
	"testing"
)

func writeLayoutFixture(t *testing.T, dir, name, content string) {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
}

func TestValidateLayoutRejectsInvalidRecordsWithoutRewriting(t *testing.T) {
	for _, tc := range []struct{ name, path, content string }{
		{"session membership", "sessions/s.json", `{"schema_version":2,"id":"s","card_id":"card","dashboard_id":"unknown"}`},
		{"operation membership", "operations/op.json", `{"operation_id":"op","dashboard_id":"unknown"}`},
		{"message membership", "messaging/history.json", `{"messages":[{"dashboard_id":"unknown"}]}`},
		{"cursor membership", "messaging/history.json", `{"dashboard_generations":{"unknown":1}}`},
		{"card schema", "id_cards.yaml", "schema_version: 1\nid_cards: []\n"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			if err := PrepareStartup(dir); err != nil {
				t.Fatal(err)
			}
			writeLayoutFixture(t, dir, "id_cards.yaml", "schema_version: 2\nid_cards:\n  - {id: card, name: Card, dashboard_id: default}\n")
			writeLayoutFixture(t, dir, "sessions/s.json", `{"schema_version":2,"id":"s","card_id":"card","dashboard_id":"default"}`)
			if err := ValidateLayout(dir); err != nil {
				t.Fatal(err)
			}
			writeLayoutFixture(t, dir, tc.path, tc.content)
			if err := ValidateLayout(dir); err == nil {
				t.Fatal("invalid record accepted")
			}
			raw, err := os.ReadFile(filepath.Join(dir, tc.path))
			if err != nil || string(raw) != tc.content {
				t.Fatalf("invalid record was changed: %q, %v", raw, err)
			}
		})
	}
}

func TestValidateLayoutRequiresCompleteConfiguration(t *testing.T) {
	dir := t.TempDir()
	if err := PrepareStartup(dir); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(filepath.Join(dir, "dashboards.yaml")); err != nil {
		t.Fatal(err)
	}
	if err := PrepareStartup(dir); err == nil {
		t.Fatal("incomplete configuration accepted")
	}
	if _, err := os.Stat(filepath.Join(dir, "dashboards.yaml")); !os.IsNotExist(err) {
		t.Fatalf("startup rewrote missing configuration: %v", err)
	}
}
