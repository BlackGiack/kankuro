package dashboardmigration

import (
	"encoding/json"
	"reflect"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestTransformPreservesRecordData(t *testing.T) {
	cards, err := TransformCards([]byte("schema_version: 1\nid_cards:\n  - id: a\n    custom: {nested: value}\nextra: keep\n"))
	if err != nil {
		t.Fatal(err)
	}
	var document struct {
		SchemaVersion int    `yaml:"schema_version"`
		Extra         string `yaml:"extra"`
	}
	var actual map[string]any
	if errAt26 := yaml.Unmarshal(cards, &actual); errAt26 != nil {
		t.Fatal(errAt26)
	}
	if errAt29 := yaml.Unmarshal(cards, &document); errAt29 != nil {
		t.Fatal(errAt29)
	}
	entry := actual["id_cards"].([]any)[0].(map[string]any)
	if document.SchemaVersion != 2 || document.Extra != "keep" || entry["dashboard_id"] != "default" || entry["shared"] != false ||
		entry["custom"].(map[string]any)["nested"] != "value" {
		t.Fatalf("unexpected cards: %s", cards)
	}
	session, err := TransformSession(
		[]byte(`{"id":"a","working_dir":"/tmp/path","conversations":{"card":[{"id":"native"}]},"extension":{"keep":true}}`),
	)
	if err != nil {
		t.Fatal(err)
	}
	var saved map[string]json.RawMessage
	if errAt44 := json.Unmarshal(session, &saved); errAt44 != nil {
		t.Fatal(errAt44)
	}
	var extension map[string]bool
	if errAt48 := json.Unmarshal(saved["extension"], &extension); errAt48 != nil {
		t.Fatal(errAt48)
	}
	if string(saved["dashboard_id"]) != `"default"` || !extension["keep"] {
		t.Fatalf("unexpected session: %s", session)
	}
}

func TestTransformsRejectExistingMembership(t *testing.T) {
	if _, err := TransformCards([]byte("schema_version: 1\nid_cards:\n  - {id: a, shared: true}\n")); err == nil {
		t.Fatal("accepted mixed card")
	}
	if _, err := TransformSession([]byte(`{"id":"a","dashboard_id":"default"}`)); err == nil {
		t.Fatal("accepted mixed session")
	}
	if _, err := TransformCards([]byte("schema_version: 1\nid_cards:\n  - &card {id: a}\n  - *card\n")); err == nil {
		t.Fatal("accepted YAML alias")
	}
}

func TestTransformSessionPreservesInstructionOwnership(t *testing.T) {
	cases := []struct {
		name, input string
		want        map[string]bool
	}{
		{"missing map", `{"id":"a","instructions_managed":true}`, map[string]bool{"AGENTS.md": true}},
		{"empty map", `{"id":"a","instructions_managed":true,"instruction_ownership":{}}`, map[string]bool{"AGENTS.md": true}},
		{
			"explicit map",
			`{"id":"a","instructions_managed":true,"instruction_ownership":{"CLAUDE.md":true}}`,
			map[string]bool{"CLAUDE.md": true},
		},
		{"unmanaged", `{"id":"a","instructions_managed":false}`, nil},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			raw, err := TransformSession([]byte(tc.input))
			if err != nil {
				t.Fatal(err)
			}
			var data struct {
				Ownership map[string]bool `json:"instruction_ownership"`
			}
			if err = json.Unmarshal(raw, &data); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(data.Ownership, tc.want) {
				t.Fatalf("ownership: %v want %v", data.Ownership, tc.want)
			}
		})
	}
}

func TestTransformCardsLegacyDefaults(t *testing.T) {
	for _, tc := range []struct{ name, raw string }{
		{"missing version", "id_cards: []\nextension: keep\n"},
		{"zero version", "schema_version: 0\nid_cards: []\nextension: keep\n"},
		{"missing cards", "schema_version: 1\nextension: keep\n"},
		{"null cards", "schema_version: 1\nid_cards: null\nextension: keep\n"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			output, err := TransformCards([]byte(tc.raw))
			if err != nil {
				t.Fatal(err)
			}
			var doc struct {
				Version   int    `yaml:"schema_version"`
				Cards     []any  `yaml:"id_cards"`
				Extension string `yaml:"extension"`
			}
			if decodeErr := yaml.Unmarshal(output, &doc); decodeErr != nil {
				t.Fatal(decodeErr)
			}
			if doc.Version != 2 || doc.Cards == nil || len(doc.Cards) != 0 || doc.Extension != "keep" {
				t.Fatalf("unexpected transformed document: %s", output)
			}
		})
	}
}

func TestTransformCardsRejectsMalformedDefaults(t *testing.T) {
	for _, raw := range []string{"schema_version: 3\nid_cards: []\n", "schema_version: '0'\nid_cards: []\n", "schema_version: null\nid_cards: []\n", "id_cards: {}\n", "id_cards: invalid\n", "id_cards: [{id: '  '}]\n", "id_cards: [{id: a}, {id: a}]\n", "id_cards: [{id: a, shared: false}]\n", "id_cards: []\nid_cards: []\n"} {
		if _, err := TransformCards([]byte(raw)); err == nil {
			t.Fatalf("accepted malformed cards: %s", raw)
		}
	}
}
