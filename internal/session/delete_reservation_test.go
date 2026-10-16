package session //nolint:testpackage // The deletion reservation is intentionally internal state.

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestSetNameRejectsDeletionReservation(t *testing.T) {
	configDir := t.TempDir()
	item := &Session{ID: "session-id", Name: "before", Subscribers: make(map[chan []byte]struct{})}
	manager := &Manager{configDir: configDir, sessions: map[string]*Session{item.ID: item}}
	if err := manager.persistSessionLockedErr(item); err != nil {
		t.Fatalf("persist initial session: %v", err)
	}
	item.deleting = true // Represents RemoveDetailed after unlink, before final map removal.
	if err := manager.SetName(item.ID, "during-delete"); !errors.Is(err, ErrLifecycleConflict) {
		t.Fatalf("rename during deletion = %v, want lifecycle conflict", err)
	}
	data, err := os.ReadFile(filepath.Join(configDir, "sessions", item.ID+".json"))
	if err != nil {
		t.Fatal(err)
	}
	var persisted struct {
		Name string `json:"name"`
	}
	if decodeErr := json.Unmarshal(data, &persisted); decodeErr != nil {
		t.Fatal(decodeErr)
	}
	if persisted.Name != "before" {
		t.Fatalf("rename recreated/changed session record: name=%q", persisted.Name)
	}
}
