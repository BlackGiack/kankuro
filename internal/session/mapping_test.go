package session_test

import (
	"reflect"
	"testing"
	"time"

	"kankuro/internal/config"
	"kankuro/internal/session"
)

// TestSnapshotCoversPersistedFields is a divergence guard: it fails when a
// field is added to Data without updating the mapping functions next to
// Session/Snapshot/Data (session.snapshot(), session.data(), Data.restore()).
// Add a field to Data, run this test, and watch it fail; that failure is the
// whole point of the exercise.
func TestSnapshotCoversPersistedFields(t *testing.T) {
	// fieldsExcludedFromSnapshot lists Data fields that deliberately have no
	// counterpart on Snapshot, with the reason why. Keep this list explicit
	// and short: anything not listed here must exist on both Session and
	// Snapshot.
	fieldsExcludedFromSnapshot := map[string]string{}

	dataType := reflect.TypeFor[session.Data]()
	sessionType := reflect.TypeFor[session.Session]()
	snapshotType := reflect.TypeFor[session.Snapshot]()

	for field := range dataType.Fields() {
		if _, excluded := fieldsExcludedFromSnapshot[field.Name]; excluded {
			continue
		}

		if _, ok := sessionType.FieldByName(field.Name); !ok {
			t.Errorf("Data.%s has no matching field on Session; update Data.restore()", field.Name)
		}
		if _, ok := snapshotType.FieldByName(field.Name); !ok {
			t.Errorf("Data.%s has no matching field on Snapshot; update Session.snapshot()", field.Name)
		}
	}
}

// TestSessionRoundTripSurvivesPersistence builds a Session, persists it via
// saveSession (through the public Manager API), reloads it via a fresh
// Manager, and asserts every persisted field survives and Status is forced
// to Terminated.
func TestSessionRoundTripSurvivesPersistence(t *testing.T) {
	dir := t.TempDir()
	repo, _ := config.NewFileRepository(dir)
	wd := t.TempDir()
	_ = repo.SaveWorkspaceSources([]config.WorkspaceSource{{Path: wd, IsRoot: false}})
	workspaces, _ := repo.ListWorkspaces()

	sm := session.NewManager(repo.Dir(), repo)
	card := config.IDCard{
		ID:          "card-roundtrip",
		Name:        "roundtrip",
		Driver:      "cat",
		WorkspaceID: workspaces[0].ID,
	}

	sid, err := sm.Launch(card, nil, nil, nil)
	if err != nil {
		t.Fatalf("launch: %v", err)
	}
	time.Sleep(100 * time.Millisecond)

	before, ok := sm.GetSession(sid)
	if !ok {
		t.Fatalf("session not found before kill")
	}

	if errKill := sm.Kill(sid); errKill != nil {
		t.Fatalf("kill: %v", errKill)
	}
	time.Sleep(100 * time.Millisecond)

	// Reload through a fresh Manager to exercise Data.restore().
	sm2 := session.NewManager(repo.Dir(), repo)
	if errLoad := sm2.LoadErrors(); errLoad != nil {
		t.Fatalf("unexpected load error: %v", errLoad)
	}

	after, ok := sm2.GetSession(sid)
	if !ok {
		t.Fatalf("session not found after reload")
	}

	if after.Status != session.StatusTerminated {
		t.Fatalf("expected reloaded session to be Terminated, got %s", after.Status)
	}
	if after.ID != before.ID {
		t.Fatalf("ID mismatch: got %s want %s", after.ID, before.ID)
	}
	if after.CardID != before.CardID {
		t.Fatalf("CardID mismatch: got %s want %s", after.CardID, before.CardID)
	}
	if after.Name != before.Name {
		t.Fatalf("Name mismatch: got %s want %s", after.Name, before.Name)
	}
	if after.WorkingDir != before.WorkingDir {
		t.Fatalf("WorkingDir mismatch: got %s want %s", after.WorkingDir, before.WorkingDir)
	}
	if after.OriginalDir != before.OriginalDir {
		t.Fatalf("OriginalDir mismatch: got %s want %s", after.OriginalDir, before.OriginalDir)
	}
	if after.UseTmpWorkspace != before.UseTmpWorkspace {
		t.Fatalf("UseTmpWorkspace mismatch: got %v want %v", after.UseTmpWorkspace, before.UseTmpWorkspace)
	}
	if !after.StartedAt.Equal(before.StartedAt) {
		t.Fatalf("StartedAt mismatch: got %v want %v", after.StartedAt, before.StartedAt)
	}
	if !after.LastWordAt.Equal(before.LastWordAt) {
		t.Fatalf("LastWordAt mismatch: got %v want %v", after.LastWordAt, before.LastWordAt)
	}
}
