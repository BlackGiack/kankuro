package session

import (
	"testing"

	"github.com/google/uuid"
)

func TestDeleteDashboardHistoryPreservesOtherScopeAndReceipts(t *testing.T) {
	store, err := NewMessageStore(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	request := uuid.NewString()
	first := MessageRecord{
		ID:                    uuid.NewString(),
		DashboardID:           "a",
		SenderSessionID:       "sender-a",
		RecipientSessionID:    "recipient-a",
		RecipientActivationID: "activation-a",
		Text:                  "first",
	}
	second := MessageRecord{
		ID:                    uuid.NewString(),
		DashboardID:           "b",
		SenderSessionID:       "sender-b",
		RecipientSessionID:    "recipient-b",
		RecipientActivationID: "activation-b",
		Text:                  "second",
	}
	if _, errAt31 := store.CreateIntent(
		first,
		request,
		SendFingerprint(first.RecipientSessionID, first.RecipientActivationID, first.Text),
	); errAt31 != nil {
		t.Fatal(errAt31)
	}
	if _, errAt38 := store.CreateIntent(
		second,
		uuid.NewString(),
		SendFingerprint(second.RecipientSessionID, second.RecipientActivationID, second.Text),
	); errAt38 != nil {
		t.Fatal(errAt38)
	}
	changed := store.Changed()
	if errAt46 := store.DeleteDashboard("a"); errAt46 != nil {
		t.Fatal(errAt46)
	}
	select {
	case <-changed:
	default:
		t.Fatal("waiters were not notified")
	}
	if _, errAt54 := store.GetMessage(first.ID); errAt54 == nil {
		t.Fatal("deleted dashboard history survived")
	}
	if _, errAt57 := store.GetMessage(second.ID); errAt57 != nil {
		t.Fatalf("foreign history removed: %v", errAt57)
	}
	if _, found, errAt60 := store.FindSend("sender-a", request); errAt60 != nil || !found {
		t.Fatalf("send receipt lost: %v, %v", found, errAt60)
	}
}
