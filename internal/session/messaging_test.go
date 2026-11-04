package session_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/BlackGiack/kankuro/internal/session"
	"github.com/google/uuid"
)

func TestLegacyReportsDoNotCreateNotifications(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, "messaging")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Fatal(err)
	}
	data, err := json.Marshal(
		map[string]any{
			"messages": []session.MessageRecord{
				{ID: "old", DashboardID: "dash", Replies: []session.MessageReply{{ID: "reply", Kind: "completed", Sequence: 1}}},
			},
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(filepath.Join(dir, "history.json"), data, 0o600); err != nil {
		t.Fatal(err)
	}
	store, err := session.NewMessageStore(root)
	if err != nil {
		t.Fatal(err)
	}
	if pending := store.Notifications(); len(pending) != 0 {
		t.Fatalf("legacy report generated notifications: %#v", pending)
	}
	page, err := store.GetReplies("dash", "old", "", 0)
	if err != nil || len(page.Items) != 1 || page.Items[0].Notification != nil {
		t.Fatalf("legacy report = %#v, %v", page, err)
	}
}

//nolint:gocognit // The table covers distinct durable outcomes and their retry semantics.
func TestReportNotificationInitialStateIsDurableAndIdempotent(t *testing.T) {
	for _, tc := range []struct {
		kind  string
		state session.NotificationState
	}{
		{"completed", "pending"}, {"blocked", "pending"}, {"question", "pending"}, {"progress", "not_required"},
	} {
		t.Run(tc.kind, func(t *testing.T) {
			root := t.TempDir()
			store, err := session.NewMessageStore(root)
			if err != nil {
				t.Fatal(err)
			}
			message := session.MessageRecord{
				ID:                    uuid.NewString(),
				DashboardID:           "dash",
				SenderSessionID:       "sender",
				SenderActivationID:    "sa",
				RecipientSessionID:    "worker",
				RecipientActivationID: "wa",
				Text:                  "task",
			}
			if _, err = store.CreateIntent(message, uuid.NewString(), session.SendFingerprint("worker", "wa", "task")); err != nil {
				t.Fatal(err)
			}
			request := uuid.NewString()
			reply, err := store.RecordReply("worker", "wa", request, message.ID, tc.kind, "report")
			if err != nil || reply.Notification == nil || reply.Notification.State != tc.state {
				t.Fatalf("reply = %#v, %v", reply, err)
			}
			reopened, err := session.NewMessageStore(root)
			if err != nil {
				t.Fatal(err)
			}
			retry, err := reopened.RecordReply("worker", "wa", request, message.ID, tc.kind, "report")
			if err != nil || retry.ID != reply.ID || retry.Notification == nil || retry.Notification.State != tc.state {
				t.Fatalf("retry = %#v, %v", retry, err)
			}
			page, err := reopened.GetReplies("dash", message.ID, "", 0)
			if err != nil || len(page.Items) != 1 || page.Items[0].Notification == nil || page.Items[0].Notification.State != tc.state {
				t.Fatalf("page = %#v, %v", page, err)
			}
		})
	}
}

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

func TestProgressRepliesPersistBeforeCompletion(t *testing.T) {
	root := t.TempDir()
	store, err := session.NewMessageStore(root)
	if err != nil {
		t.Fatal(err)
	}
	message := session.MessageRecord{
		ID:                    uuid.NewString(),
		DashboardID:           "dash",
		SenderSessionID:       "sender",
		RecipientSessionID:    "recipient",
		RecipientActivationID: "activation",
		Text:                  "work",
	}
	if _, err = store.CreateIntent(message, uuid.NewString(), session.SendFingerprint("recipient", "activation", "work")); err != nil {
		t.Fatal(err)
	}
	request := uuid.NewString()
	first, err := store.RecordReply("recipient", "activation", request, message.ID, "progress", "tests running")
	if err != nil {
		t.Fatal(err)
	}
	retry, err := store.RecordReply("recipient", "activation", request, message.ID, "progress", "tests running")
	if err != nil || first.ID != retry.ID {
		t.Fatalf("retry: %+v %v", retry, err)
	}
	if _, err = store.RecordReply("recipient", "activation", request, message.ID, "completed", "tests running"); err == nil {
		t.Fatal("changed kind accepted")
	}
	if _, err = store.RecordReply("recipient", "activation", uuid.NewString(), message.ID, "completed", "done"); err != nil {
		t.Fatal(err)
	}
	store, err = session.NewMessageStore(root)
	if err != nil {
		t.Fatal(err)
	}
	page, err := store.GetReplies("dash", message.ID, "", 0)
	if err != nil || len(page.Items) != 2 || page.Items[0].Kind != "progress" || page.Items[1].Kind != "completed" {
		t.Fatalf("reports: %+v %v", page, err)
	}
}

func TestRestartMarksAdmittedNotificationUncertain(t *testing.T) {
	root := t.TempDir()
	store, err := session.NewMessageStore(root)
	if err != nil {
		t.Fatal(err)
	}
	original := session.MessageRecord{
		ID:                    "task",
		DashboardID:           "dash",
		SenderSessionID:       "controller",
		SenderActivationID:    "ca",
		RecipientSessionID:    "worker",
		RecipientActivationID: "wa",
		Text:                  "task",
	}
	if _, err = store.CreateIntent(original, uuid.NewString(), "task-fingerprint"); err != nil {
		t.Fatal(err)
	}
	reply, err := store.RecordReply("worker", "wa", uuid.NewString(), original.ID, "completed", "done")
	if err != nil {
		t.Fatal(err)
	}
	request := uuid.NewString()
	if err = store.UpdateNotification(
		original.ID,
		reply.ID,
		0,
		session.NotificationPending,
		func(n *session.ReplyNotification) { n.RequestID = request },
	); err != nil {
		t.Fatal(err)
	}
	intent := session.MessageRecord{
		ID:                    "notification",
		Kind:                  "report_notification",
		OriginalMessageID:     original.ID,
		ReplyID:               reply.ID,
		ReportKind:            reply.Kind,
		DashboardID:           "dash",
		SenderSessionID:       "worker",
		SenderActivationID:    "wa",
		RecipientSessionID:    "controller",
		RecipientActivationID: "ca",
		Text:                  "report notification",
	}
	if _, err = store.CreateIntentWithAdmission(intent, request, "fingerprint", func() error { return nil }); err != nil {
		t.Fatal(err)
	}
	reopened, err := session.NewMessageStore(root)
	if err != nil {
		t.Fatal(err)
	}
	if err = reopened.Recover(map[string]*session.Session{"worker": {ID: "worker"}, "controller": {ID: "controller"}}); err != nil {
		t.Fatal(err)
	}
	loaded, err := reopened.GetMessage(intent.ID)
	if err != nil || loaded.Delivery != session.MessageUncertain {
		t.Fatalf("recovered notification = %#v, %v", loaded, err)
	}
	pending := reopened.Notifications()
	if len(pending) != 1 || pending[0].Reply.Notification.State != session.NotificationAccepted ||
		pending[0].Reply.Notification.RequestID != request {
		t.Fatalf("recovered obligation = %#v", pending)
	}
}
