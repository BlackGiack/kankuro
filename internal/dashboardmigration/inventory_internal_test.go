package dashboardmigration

import (
	"os"
	"path/filepath"
	"testing"
)

func writeFixture(t *testing.T, dir, name, content string) {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
}

func TestInspectLegacy(t *testing.T) {
	dir := t.TempDir()
	writeFixture(t, dir, "id_cards.yaml", "schema_version: 1\nid_cards:\n  - id: card-a\n    custom: preserved\n")
	writeFixture(t, dir, "sessions/a.json", `{"id":"a","card_id":"card-a","conversations":{"card-a":[{"id":"x"}]}}`)
	writeFixture(t, dir, "operations/a.json", `{"operation_id":"op-a","dashboard_id":"default"}`)
	writeFixture(t, dir, "messaging/history.json",
		`{"messages":[{"message_id":"m","dashboard_id":"default"}],"dashboard_generations":{"default":1}}`)
	got, err := InspectLegacy(dir)
	if err != nil {
		t.Fatal(err)
	}
	if got != (Inventory{Cards: 1, Sessions: 1, Operations: 1, Messages: 1}) {
		t.Fatalf("inventory: %+v", got)
	}
	raw, err := os.ReadFile(filepath.Join(dir, "id_cards.yaml"))
	if err != nil || string(raw) != "schema_version: 1\nid_cards:\n  - id: card-a\n    custom: preserved\n" {
		t.Fatalf("inspection changed input: %q, %v", raw, err)
	}
}

func TestInspectLegacyRejectsAmbiguousData(t *testing.T) {
	tests := []struct{ name, path, content string }{
		{"mixed layout", "dashboards.yaml", ""},
		{"unknown schema", "id_cards.yaml", "schema_version: 2\nid_cards: []\n"},
		{"duplicate cards", "id_cards.yaml", "schema_version: 1\nid_cards:\n  - {id: a}\n  - {id: a}\n"},
		{"mixed card", "id_cards.yaml", "schema_version: 1\nid_cards:\n  - {id: a, dashboard_id: default}\n"},
		{"mixed session", "sessions/a.json", `{"id":"a","dashboard_id":"default"}`},
		{"foreign operation", "operations/a.json", `{"operation_id":"a","dashboard_id":"other"}`},
		{"invalid history", "messaging/history.json", `{"messages":[{"dashboard_id":"other"}]}`},
		{"duplicate json key", "sessions/a.json", `{"id":"a","extension":{"key":1,"key":2}}`},
		{"multiple yaml documents", "id_cards.yaml", "schema_version: 1\nid_cards: []\n---\nignored: true\n"},
		{"missing card reference", "sessions/a.json", `{"id":"a","card_id":"missing"}`},
		{"session filename mismatch", "sessions/wrong.json", `{"id":"a","card_id":"missing"}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			writeFixture(t, dir, tt.path, tt.content)
			if _, err := InspectLegacy(dir); err == nil {
				t.Fatal("expected error")
			}
		})
	}
}
