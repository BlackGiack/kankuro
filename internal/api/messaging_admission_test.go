//nolint:testpackage // The queue-full probe verifies the runner's private admission boundary.
package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/BlackGiack/kankuro/internal/session"
	"github.com/google/uuid"
)

func TestFullOperationQueueDoesNotStartMessageAdmissionTransaction(t *testing.T) {
	runner := &OperationRunner{
		dir: t.TempDir(), records: make(map[operationKey]*Operation), byID: make(map[string]operationKey),
		running: make(map[string]context.CancelFunc), queue: make(chan operationTask, operationQueueCapacity),
	}
	for range cap(runner.queue) {
		runner.queue <- operationTask{}
	}
	called := false
	_, _, err := runner.accept("sender", "dashboard", operationKey{caller: "sender", request: uuid.NewString()},
		uuid.NewString(), messageOperation, "recipient", []byte(`{}`), "fingerprint", "token", nil, nil,
		func(func() error) error { called = true; return nil })
	if err == nil || called {
		t.Fatalf("full queue admission err=%v, transaction callback called=%v", err, called)
	}
	var rejected *RejectedError
	if !errors.As(err, &rejected) || rejected.Code != "capacity_exhausted" {
		t.Fatalf("full queue error = %#v, want capacity_exhausted", err)
	}
}

func TestFullQueuePreservesAcceptedHistoryAtRetentionCapacity(t *testing.T) {
	configDir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(configDir, "messaging"), 0o700); err != nil {
		t.Fatal(err)
	}
	messages := make([]session.MessageRecord, 10000)
	for i := range messages {
		id := fmt.Sprintf("message-%05d", i)
		messages[i] = session.MessageRecord{ID: id, DashboardID: "dashboard", SenderSessionID: "sender",
			RecipientSessionID: "recipient", RecipientActivationID: "activation", Text: "accepted",
			Delivery: session.MessageSubmitted, Sequence: uint64(i + 1)}
	}
	messages[0].Replies = []session.MessageReply{
		{ID: "reply-oldest", MessageID: messages[0].ID, Kind: "completed", Text: "kept", Sequence: 1},
	}
	seed := struct {
		Messages             []session.MessageRecord `json:"messages"`
		Sends                []map[string]string     `json:"send_receipts"`
		Sequence             uint64                  `json:"sequence"`
		DashboardGenerations map[string]uint64       `json:"dashboard_generations"`
	}{
		Messages: messages,
		Sends: []map[string]string{{"sender_session_id": "sender", "request_id": "old-request",
			"fingerprint": "old-fingerprint", "message_id": messages[0].ID}},
		Sequence: 10000, DashboardGenerations: map[string]uint64{"dashboard": 7},
	}
	encoded, err := json.Marshal(seed)
	if err != nil {
		t.Fatal(err)
	}
	historyPath := filepath.Join(configDir, "messaging", "history.json")
	if err = os.WriteFile(historyPath, encoded, 0o600); err != nil {
		t.Fatal(err)
	}
	store, err := session.NewMessageStore(configDir)
	if err != nil {
		t.Fatal(err)
	}
	page, err := store.GetReplies("dashboard", "", "", 1)
	if err != nil || page.NextCursor == "" {
		t.Fatalf("seed reply cursor = %#v, %v", page, err)
	}
	runner := &OperationRunner{
		dir: t.TempDir(), records: make(map[operationKey]*Operation), byID: make(map[string]operationKey),
		running: make(map[string]context.CancelFunc), queue: make(chan operationTask, operationQueueCapacity),
	}
	for range cap(runner.queue) {
		runner.queue <- operationTask{}
	}
	called := false
	_, _, err = runner.accept("sender", "dashboard", operationKey{caller: "sender", request: uuid.NewString()},
		uuid.NewString(), messageOperation, "recipient", []byte(`{}`), "new-fingerprint", "token", nil, nil,
		func(persist func() error) error {
			called = true
			_, createErr := store.CreateIntentWithAdmission(session.MessageRecord{ID: "rejected", DashboardID: "dashboard",
				SenderSessionID: "sender", RecipientSessionID: "recipient", RecipientActivationID: "activation", Text: "new"},
				"new-request", "new-fingerprint", persist)
			return createErr
		})
	if err == nil || called {
		t.Fatalf("full queue admission err=%v, message transaction called=%v", err, called)
	}
	if _, err = store.GetMessage(messages[0].ID); err != nil {
		t.Fatalf("oldest accepted message was evicted by rejected request: %v", err)
	}
	if _, found, findErr := store.FindSend("sender", "old-request"); findErr != nil || !found {
		t.Fatalf("accepted send receipt found=%v err=%v", found, findErr)
	}
	resumed, err := store.GetReplies("dashboard", "", page.NextCursor, 10)
	if err != nil || resumed.Gap {
		t.Fatalf("retained cursor state after queue rejection = %#v, %v", resumed, err)
	}
}
