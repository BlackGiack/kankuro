package session

import (
	"encoding/json"
	"testing"
)

func TestDashboardMembershipSurvivesRecordRoundTrip(t *testing.T) {
	original := &Session{ID: "session", DashboardID: "dashboard-a", CardID: "shared-card"}
	raw, err := json.Marshal(original.data())
	if err != nil {
		t.Fatal(err)
	}
	var stored Data
	if errAt15 := json.Unmarshal(raw, &stored); errAt15 != nil {
		t.Fatal(errAt15)
	}
	if stored.SchemaVersion != 2 {
		t.Fatalf("schema version = %d", stored.SchemaVersion)
	}
	restored := stored.restore()
	if got := restored.snapshot().DashboardID; got != "dashboard-a" {
		t.Fatalf("dashboard membership = %q", got)
	}
}
