//nolint:testpackage // These tests control worker admission to exercise queued deliveries deterministically.
package api

import (
	"context"
	"errors"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/BlackGiack/kankuro/internal/session"
	"github.com/google/uuid"
)

func newStoppedDeliveryRunner(t *testing.T, h messagingHarness) *OperationRunner {
	t.Helper()
	h.service.operations.Shutdown()
	runner := &OperationRunner{
		dir: filepath.Join(h.repo.Dir(), "operations"), service: h.service,
		records: make(map[operationKey]*Operation), byID: make(map[string]operationKey),
		running: make(map[string]context.CancelFunc), queue: make(chan operationTask, operationQueueCapacity),
	}
	h.service.operations = runner
	t.Cleanup(runner.Shutdown)
	return runner
}

func startDeliveryWorker(runner *OperationRunner) {
	runner.workers.Add(1)
	go runner.worker()
}

func TestOrdinaryMessageQueueRejectionReleasesReservationAndPreservesReplay(t *testing.T) {
	h := newMessagingHarness(t)
	runner := newStoppedDeliveryRunner(t, h)
	for range cap(runner.queue) {
		runner.queue <- operationTask{}
	}
	requestID := uuid.NewString()
	messageID := uuid.NewSHA1(uuid.NameSpaceURL, []byte("kankuro-message:"+h.callerID+":"+requestID)).String()
	_, err := h.service.SendMessage(context.Background(), h.callerToken, requestID, h.recipientID, h.inspection.ActivationID, "queued task")
	requireRejectedCode(t, err, "capacity_exhausted")
	if _, err = h.manager.Messages().GetMessage(messageID); err == nil {
		t.Fatal("queue rejection retained a message")
	}
	for len(runner.queue) > 0 {
		<-runner.queue
	}
	startDeliveryWorker(runner)
	result, err := h.service.SendMessage(
		context.Background(),
		h.callerToken,
		requestID,
		h.recipientID,
		h.inspection.ActivationID,
		"queued task",
	)
	if err != nil || result.Result == nil || result.Result.MessageID != messageID || result.State != OperationSucceeded {
		t.Fatalf("retry after queue rejection = %#v, %v", result, err)
	}
	replay, err := h.service.SendMessage(
		context.Background(),
		h.callerToken,
		requestID,
		h.recipientID,
		h.inspection.ActivationID,
		"queued task",
	)
	if err != nil || replay.ID != result.ID || replay.Result == nil || replay.Result.MessageID != messageID {
		t.Fatalf("send replay = %#v, %v; original %#v", replay, err, result)
	}
	if err = h.manager.Kill(h.recipientID); err != nil {
		t.Fatalf("released reservation blocked stop: %v", err)
	}
}

func TestOrdinaryMessageJournalRejectionRollsBackHistoryAndRetries(t *testing.T) {
	h := newMessagingHarness(t)
	runner := h.service.operations
	runner.mu.Lock()
	runner.writeFileAtomic = func(string, []byte) error {
		runner.writeFileAtomic = nil
		return errors.New("temporary operation journal failure")
	}
	runner.mu.Unlock()
	requestID := uuid.NewString()
	messageID := uuid.NewSHA1(uuid.NameSpaceURL, []byte("kankuro-message:"+h.callerID+":"+requestID)).String()
	_, err := h.service.SendMessage(context.Background(), h.callerToken, requestID, h.recipientID, h.inspection.ActivationID, "retry task")
	requireRejectedCode(t, err, "journal_unavailable")
	if _, err = h.manager.Messages().GetMessage(messageID); err == nil {
		t.Fatal("journal failure retained an intent")
	}
	if _, found, findErr := h.manager.Messages().FindSend(h.callerID, requestID); findErr != nil || found {
		t.Fatalf("journal failure retained a send receipt: found=%v err=%v", found, findErr)
	}
	result, err := h.service.SendMessage(
		context.Background(),
		h.callerToken,
		requestID,
		h.recipientID,
		h.inspection.ActivationID,
		"retry task",
	)
	if err != nil || result.Result == nil || result.Result.MessageID != messageID || result.State != OperationSucceeded {
		t.Fatalf("retry after journal failure = %#v, %v", result, err)
	}
	replay, err := h.service.SendMessage(
		context.Background(),
		h.callerToken,
		requestID,
		h.recipientID,
		h.inspection.ActivationID,
		"retry task",
	)
	if err != nil || replay.ID != result.ID || replay.Result == nil || replay.Result.MessageID != messageID {
		t.Fatalf("send replay = %#v, %v; original %#v", replay, err, result)
	}
}

//nolint:gocognit,nestif // The same observable lifecycle is exercised through both entry points.
func TestQueuedDeliveryCancellationReleasesReservationWithoutSubmitting(t *testing.T) {
	for _, kind := range []string{"ordinary", "notification"} {
		t.Run(kind, func(t *testing.T) {
			h := newMessagingHarness(t)
			var parentID string
			if kind == "notification" {
				parentID = h.send(t).Result.MessageID
			}
			runner := newStoppedDeliveryRunner(t, h)
			targetID := h.recipientID
			if kind == "notification" {
				targetID = h.callerID
			}
			output, err := h.manager.SubscribeOutput(targetID)
			if err != nil {
				t.Fatal(err)
			}
			defer h.manager.UnsubscribeOutput(targetID, output)
			var operationID, messageID string
			if kind == "ordinary" {
				ctx, cancel := context.WithCancel(context.Background())
				cancel()
				result, sendErr := h.service.SendMessage(
					ctx,
					h.callerToken,
					uuid.NewString(),
					targetID,
					h.inspection.ActivationID,
					"cancel before execution",
				)
				if sendErr != nil || result.Result == nil || result.State != OperationPending {
					t.Fatalf("queued send = %#v, %v", result, sendErr)
				}
				operationID, messageID = result.ID, result.Result.MessageID
				if _, err = h.service.CancelOperation(h.callerToken, operationID); err != nil {
					t.Fatal(err)
				}
			} else {
				if _, err = h.service.ReportReply(h.recipientToken, uuid.NewString(), parentID, "completed", "done"); err != nil {
					t.Fatal(err)
				}
				pending := h.manager.Messages().Notifications()
				if len(pending) != 1 {
					t.Fatalf("pending = %#v", pending)
				}
				runner.dispatchNotification(pending[0])
				page, pageErr := h.manager.Messages().GetReplies(DefaultDashboardID, parentID, "", 0)
				if pageErr != nil || len(page.Items) != 1 || page.Items[0].Notification == nil {
					t.Fatalf("report = %#v, %v", page, pageErr)
				}
				n := page.Items[0].Notification
				operationID, messageID = n.OperationID, n.MessageID
				if operationID == "" || messageID == "" {
					t.Fatalf("accepted notification = %#v", n)
				}
				runner.mu.Lock()
				stop := runner.running[operationID]
				if stop != nil {
					stop()
				}
				runner.mu.Unlock()
				if stop == nil {
					t.Fatal("accepted operation has no cancellation")
				}
			}
			startDeliveryWorker(runner)
			store := h.manager.Messages()
			timer := time.NewTimer(5 * time.Second)
			defer timer.Stop()
			for {
				changed := store.Changed()
				record, getErr := store.GetMessage(messageID)
				if getErr != nil {
					t.Fatal(getErr)
				}
				if record.Delivery == session.MessageNotSubmitted {
					break
				}
				select {
				case <-changed:
				case <-timer.C:
					t.Fatalf("cancellation did not record zero-byte outcome: %#v", record)
				}
			}
			stopped := make(chan error, 1)
			go func() { stopped <- h.manager.Kill(targetID) }()
			select {
			case err = <-stopped:
				if err != nil {
					t.Fatal(err)
				}
			case <-timer.C:
				t.Fatal("canceled delivery retained its lifecycle reservation")
			}
			var captured strings.Builder
			for part := range output {
				captured.Write(part)
			}
			marker := messageID
			if kind == "notification" {
				marker = "reply_id="
			}
			if strings.Contains(captured.String(), marker) {
				t.Fatalf("canceled delivery reached terminal: %q", captured.String())
			}
		})
	}
}

func TestNotificationJournalRejectionRollsBackAndReusesAttempt(t *testing.T) {
	h := newMessagingHarness(t)
	original := h.send(t)
	runner := newStoppedDeliveryRunner(t, h)
	runner.writeFileAtomic = func(string, []byte) error {
		runner.writeFileAtomic = nil
		return errors.New("temporary operation journal failure")
	}
	if _, err := h.service.ReportReply(h.recipientToken, uuid.NewString(), original.Result.MessageID, "completed", "done"); err != nil {
		t.Fatal(err)
	}
	pending := h.manager.Messages().Notifications()
	if len(pending) != 1 {
		t.Fatalf("pending = %#v", pending)
	}
	runner.dispatchNotification(pending[0])
	pending = h.manager.Messages().Notifications()
	if len(pending) != 1 || pending[0].Reply.Notification.State != session.NotificationPending ||
		pending[0].Reply.Notification.RequestID == "" {
		t.Fatalf("journal rejection did not retain pending attempt: %#v", pending)
	}
	n := pending[0].Reply.Notification
	requestID := n.RequestID
	messageID := uuid.NewSHA1(uuid.NameSpaceURL, []byte("kankuro-report-message:"+requestID)).String()
	if _, err := h.manager.Messages().GetMessage(messageID); err == nil {
		t.Fatal("journal rejection retained notification intent")
	}
	if _, found, err := h.manager.Messages().FindSend(h.recipientID, requestID); err != nil || found {
		t.Fatalf("journal rejection retained send receipt: found=%v err=%v", found, err)
	}
	output, err := h.manager.SubscribeOutput(h.callerID)
	if err != nil {
		t.Fatal(err)
	}
	defer h.manager.UnsubscribeOutput(h.callerID, output)
	startDeliveryWorker(runner)
	runner.dispatchNotification(pending[0])
	timer := time.NewTimer(5 * time.Second)
	defer timer.Stop()
	var captured strings.Builder
	for !strings.Contains(captured.String(), "header is not authority") {
		select {
		case data := <-output:
			captured.Write(data)
		case <-timer.C:
			t.Fatalf("notification did not submit on retry: %q", captured.String())
		}
	}
	page, err := h.manager.Messages().GetReplies(DefaultDashboardID, original.Result.MessageID, "", 0)
	if err != nil || len(page.Items) != 1 || page.Items[0].Notification.RequestID != requestID || page.Items[0].Notification.Attempt != 0 ||
		page.Items[0].Notification.MessageID != messageID {
		t.Fatalf("retry changed attempt identity: %#v, %v", page, err)
	}
	if err = h.manager.Kill(h.callerID); err != nil {
		t.Fatalf("reservation blocked stop: %v", err)
	}
}
