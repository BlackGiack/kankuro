package session_test

import (
	"testing"

	"github.com/BlackGiack/kankuro/internal/session"
	"github.com/google/uuid"
)

func TestMessageStorePersistsAndDeduplicatesReplies(t *testing.T) {
	root := t.TempDir()
	store, err := session.NewMessageStore(root)
	if err != nil {
		t.Fatal(err)
	}
	message := session.MessageRecord{ID: uuid.NewString(), DashboardID: "dash", SenderSessionID: "sender", SenderActivationID: "sa",
		RecipientSessionID: "recipient", RecipientActivationID: "ra", Text: "do the task"}
	fingerprint := session.SendFingerprint("recipient", "ra", message.Text)
	if _, err = store.CreateIntent(message, uuid.NewString(), fingerprint); err != nil {
		t.Fatal(err)
	}
	requestID := uuid.NewString()
	reply, err := store.RecordReply("recipient", "ra", requestID, message.ID, "completed", "finished")
	if err != nil {
		t.Fatal(err)
	}
	retry, err := store.RecordReply("recipient", "ra", requestID, message.ID, "completed", "finished")
	if err != nil || retry.ID != reply.ID {
		t.Fatalf("retry = %#v, %v; want same reply %q", retry, err, reply.ID)
	}
	if _, err = store.RecordReply("recipient", "ra", requestID, message.ID, "question", "changed"); err == nil {
		t.Fatal("changed retry was accepted")
	}

	reopened, err := session.NewMessageStore(root)
	if err != nil {
		t.Fatal(err)
	}
	loaded, err := reopened.GetMessage(message.ID)
	if err != nil || len(loaded.Replies) != 1 || loaded.Replies[0].ID != reply.ID {
		t.Fatalf("loaded message = %#v, %v", loaded, err)
	}
	page, err := reopened.GetReplies("dash", message.ID, "", 0)
	if err != nil || len(page.Items) != 1 || page.Items[0].ID != reply.ID {
		t.Fatalf("reply page = %#v, %v", page, err)
	}
	if _, err = reopened.GetMessage("missing"); err == nil {
		t.Fatal("missing message returned without an error")
	}
}

func TestMessageStorePinsReportIdentityAndDeletesRecipientHistory(t *testing.T) {
	store, err := session.NewMessageStore(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	message := session.MessageRecord{ID: uuid.NewString(), DashboardID: "dash", SenderSessionID: "sender",
		RecipientSessionID: "recipient", RecipientActivationID: "activation", Text: "work"}
	if _, err = store.CreateIntent(message, uuid.NewString(), session.SendFingerprint("recipient", "activation", "work")); err != nil {
		t.Fatal(err)
	}
	if _, err = store.RecordReply("recipient", "replacement", uuid.NewString(), message.ID, "completed", "wrong activation"); err == nil {
		t.Fatal("replacement activation report was accepted")
	}
	if err = store.DeleteRecipient("recipient"); err != nil {
		t.Fatal(err)
	}
	if _, err = store.GetMessage(message.ID); err == nil {
		t.Fatal("deleted recipient history remains available")
	}
	if _, err = store.RecordReply("recipient", "activation", uuid.NewString(), message.ID, "completed", "late"); err == nil {
		t.Fatal("late report recreated deleted history")
	}
}

func TestMessageStoreRecoversInterruptedSubmissionAndCursorGap(t *testing.T) {
	root := t.TempDir()
	store, err := session.NewMessageStore(root)
	if err != nil {
		t.Fatal(err)
	}
	first := session.MessageRecord{ID: uuid.NewString(), DashboardID: "dash", SenderSessionID: "sender",
		RecipientSessionID: "recipient-a", RecipientActivationID: "activation-a", Text: "first"}
	if _, err = store.CreateIntent(
		first,
		uuid.NewString(),
		session.SendFingerprint("recipient-a", "activation-a", first.Text),
	); err != nil {
		t.Fatal(err)
	}
	if err = store.UpdateDelivery(first.ID, session.MessageSubmitting, ""); err != nil {
		t.Fatal(err)
	}
	reopened, err := session.NewMessageStore(root)
	if err != nil {
		t.Fatal(err)
	}
	if err = reopened.Recover(map[string]*session.Session{"recipient-a": {ID: "recipient-a"}}); err != nil {
		t.Fatal(err)
	}
	recovered, err := reopened.GetMessage(first.ID)
	if err != nil || recovered.Delivery != session.MessageUncertain {
		t.Fatalf("recovered submission = %#v, %v", recovered, err)
	}
	requestID := uuid.NewString()
	if _, err = reopened.RecordReply("recipient-a", "activation-a", requestID, first.ID, "completed", "done"); err != nil {
		t.Fatal(err)
	}
	page, err := reopened.GetReplies("dash", "", "", 50)
	if err != nil || len(page.Items) != 1 {
		t.Fatalf("initial reply page = %#v, %v", page, err)
	}

	second := session.MessageRecord{ID: uuid.NewString(), DashboardID: "dash", SenderSessionID: "sender",
		RecipientSessionID: "recipient-b", RecipientActivationID: "activation-b", Text: "second"}
	if _, err = reopened.CreateIntent(
		second,
		uuid.NewString(),
		session.SendFingerprint("recipient-b", "activation-b", second.Text),
	); err != nil {
		t.Fatal(err)
	}
	if _, err = reopened.RecordReply("recipient-b", "activation-b", uuid.NewString(), second.ID, "question", "need detail"); err != nil {
		t.Fatal(err)
	}
	oldCursor := page.NextCursor
	if err = reopened.DeleteRecipient("recipient-a"); err != nil {
		t.Fatal(err)
	}
	resumed, err := reopened.GetReplies("dash", "", oldCursor, 50)
	if err != nil || !resumed.Gap || len(resumed.Items) != 1 || resumed.Items[0].MessageID != second.ID {
		t.Fatalf("gap recovery page = %#v, %v", resumed, err)
	}
}
