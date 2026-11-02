package session

import (
	"context"
	"testing"
	"time"
)

func TestDashboardAdmissionCancellationIsScoped(t *testing.T) {
	m := &Manager{sessions: map[string]*Session{}}
	ctxA, cancelA := context.WithCancel(context.Background())
	defer cancelA()
	ctxB, cancelB := context.WithCancel(context.Background())
	defer cancelB()
	ctxOther, cancelOther := context.WithCancel(context.Background())
	defer cancelOther()
	check := func([]Snapshot) (bool, error) { return true, nil }
	releaseA, err := m.ReserveCardReferenceForDashboardWithCancel("foreign-card", "a", "session-a", cancelA, check)
	if err != nil {
		t.Fatal(err)
	}
	defer releaseA()
	releaseB, err := m.ReserveCardReferenceForDashboardWithCancel("owned-card", "b", "session-b", cancelB, check)
	if err != nil {
		t.Fatal(err)
	}
	defer releaseB()
	releaseOther, err := m.ReserveCardReferenceForDashboardWithCancel("foreign-card", "b", "other", cancelOther, check)
	if err != nil {
		t.Fatal(err)
	}
	defer releaseOther()
	m.CancelDashboardAdmissions("a", map[string]bool{"owned-card": true})
	if ctxA.Err() == nil || ctxB.Err() == nil || ctxOther.Err() != nil {
		t.Fatal("dashboard cancellation crossed its intended scope")
	}
	releaseA()
	releaseB()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if errAt41 := m.WaitDashboardAdmissions(ctx, "a", map[string]bool{"owned-card": true}); errAt41 != nil {
		t.Fatal(errAt41)
	}
	if ctxOther.Err() != nil {
		t.Fatal("unrelated activation was cancelled")
	}
}
