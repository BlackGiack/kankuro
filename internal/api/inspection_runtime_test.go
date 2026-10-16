package api //nolint:testpackage // Exercises the shared unexported inspection conversion.

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/BlackGiack/kankuro/internal/session"
)

func TestInspectionConversionPreservesActivationSnapshot(t *testing.T) {
	snapshot := session.Snapshot{ID: "session", CardID: "card", Name: "name", Status: session.StatusRunning,
		LifecycleActivity: "idle", Phase: "starting", ProcessStarted: true, MessageDeliveryAvailable: false,
		ActivationID: "activation", KankuroMCPEnabled: true, Registration: session.RegistrationPending}
	result := inspectionFromSnapshot(snapshot)
	if result.ActivationID != snapshot.ActivationID || !result.KankuroMCPEnabled || result.Phase != "starting" ||
		!result.ProcessStarted || result.MessageDeliveryAvailable || result.Conversations == nil || result.SubWorkspaces == nil {
		t.Fatalf("inspection = %+v", result)
	}
	data, err := json.Marshal(OperationEffects{SessionID: snapshot.ID, Inspection: &result})
	if err != nil {
		t.Fatal(err)
	}
	snapshot.Phase, snapshot.MessageDeliveryAvailable, snapshot.ActivationID = "active", true, "replacement"
	var historical OperationEffects
	if decodeErr := json.Unmarshal(data, &historical); decodeErr != nil {
		t.Fatal(decodeErr)
	}
	if !reflect.DeepEqual(historical.Inspection, &result) {
		t.Fatalf("historical outcome changed: %+v", historical.Inspection)
	}
	var old SessionInspection
	if decodeErr := json.Unmarshal([]byte(`{"session_id":"legacy","kankuro_mcp_enabled":true}`), &old); decodeErr != nil {
		t.Fatal(decodeErr)
	}
	if old.ProcessStarted || old.MessageDeliveryAvailable || old.Phase != "" {
		t.Fatalf("legacy data fabricated runtime state: %+v", old)
	}
}
