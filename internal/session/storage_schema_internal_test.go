package session

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSessionStorageRequiresCurrentSchemaAndMembership(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "session.json")
	for _, raw := range []string{`{"id":"session","dashboard_id":"default"}`, `{"schema_version":1,"id":"session","dashboard_id":"default"}`, `{"schema_version":2,"id":"session"}`} {
		if err := os.WriteFile(path, []byte(raw), 0o600); err != nil {
			t.Fatal(err)
		}
		records, err := loadSessions(dir)
		if err == nil || len(records) != 0 {
			t.Fatalf("accepted invalid record %s: %v %v", raw, records, err)
		}
	}
	if err := saveSession(dir, &Data{ID: "session", SchemaVersion: 2}); err == nil {
		t.Fatal("wrote session without membership")
	}
}

func TestCurrentRecordDoesNotInferInstructionOwnership(t *testing.T) {
	item := (&Session{ID: "session", DashboardID: "default", InstructionsManaged: true}).data()
	if len(item.InstructionOwnership) != 0 {
		t.Fatalf("inferred instruction ownership: %v", item.InstructionOwnership)
	}
}
