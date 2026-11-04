//nolint:testpackage // Tests exercise the API/session operation integration fixtures directly.
package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/BlackGiack/kankuro/internal/config"
	"github.com/BlackGiack/kankuro/internal/fsutil"
	"github.com/BlackGiack/kankuro/internal/session"
	"github.com/google/uuid"
)

type messagingHarness struct {
	service        *Service
	manager        *session.Manager
	repo           *testRepository
	workspace      string
	callerID       string
	callerToken    string
	recipientID    string
	recipientToken string
	callerCard     config.IDCard
	inspection     SessionInspection
}

const reviewWaiterSlots = 31

func newMessagingHarness(t *testing.T) messagingHarness {
	t.Helper()
	mcpCLIContent, err := os.ReadFile("../../kankuro_skills/mcp-cli/SKILL.md")
	if err != nil {
		t.Fatal(err)
	}
	config.SetBuiltinMCPCLISkillContent(mcpCLIContent)
	t.Cleanup(func() { config.SetBuiltinMCPCLISkillContent(nil) })
	service, repo, manager, workspace, workspaceID := newServiceFixture(t)
	manager.SetMCPEndpoint("http://127.0.0.1:12345/mcp")
	callerCard := saveCard(t, repo.FileRepository, config.IDCard{ID: "review-caller", Name: "review caller", Driver: "cat",
		WorkspaceID: workspaceID, DashboardControl: true, MCPIDs: []string{config.BuiltinMCPKankuro}})
	recipientCard := saveCard(t, repo.FileRepository, config.IDCard{ID: "review-recipient", Name: "review recipient", Driver: "cat",
		WorkspaceID: workspaceID, MCPIDs: []string{config.BuiltinMCPKankuro}})
	callerID, err := launchCardTest(service, callerCard)
	if err != nil {
		t.Fatal(err)
	}
	callerToken := readActivationToken(t, workspace)
	recipientID, err := launchCardTest(service, recipientCard)
	if err != nil {
		t.Fatal(err)
	}
	recipientToken := readActivationToken(t, workspace)
	runner := NewOperationRunner(repo.Dir(), service)
	service.SetOperationRunner(runner)
	t.Cleanup(runner.Shutdown)
	inspection, err := service.GetSession(callerToken, recipientID)
	if err != nil {
		t.Fatal(err)
	}
	return messagingHarness{service: service, manager: manager, repo: repo, workspace: workspace, callerID: callerID,
		callerToken: callerToken, recipientID: recipientID, recipientToken: recipientToken, callerCard: callerCard,
		inspection: inspection}
}

func readActivationToken(t *testing.T, workspace string) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(workspace, ".kankuro", "mcp-servers.json"))
	if err != nil {
		t.Fatal(err)
	}
	var configuration struct {
		Servers map[string]struct {
			Headers map[string]string `json:"headers"`
		} `json:"mcpServers"`
	}
	if err = json.Unmarshal(data, &configuration); err != nil {
		t.Fatal(err)
	}
	token := strings.TrimPrefix(configuration.Servers["kankuro"].Headers["Authorization"], "Bearer ")
	if token == "" {
		t.Fatal("activation token missing from deployed MCP configuration")
	}
	return token
}

func (h messagingHarness) send(t *testing.T) OperationResult {
	t.Helper()
	requestID := uuid.NewString()
	result, err := h.service.SendMessage(context.Background(), h.callerToken, requestID,
		h.recipientID, h.inspection.ActivationID, "read the project file")
	if err != nil || result.Result == nil || result.Result.MessageID == "" {
		t.Fatalf("send result = %#v, %v", result, err)
	}
	return result
}

func requireRejectedCode(t *testing.T, err error, want string) {
	t.Helper()
	var rejected *RejectedError
	if !errors.As(err, &rejected) || rejected.Code != want {
		t.Fatalf("error = %v, want rejection code %q", err, want)
	}
}

func TestSendMessageCannotRecreateHistoryAfterRecipientDeletion(t *testing.T) {
	h := newMessagingHarness(t)
	if err := h.manager.Kill(h.recipientID); err != nil {
		t.Fatal(err)
	}
	if err := removeSessionTest(h.service, h.recipientID); err != nil {
		t.Fatal(err)
	}
	requestID := uuid.NewString()
	_, err := h.service.SendMessage(context.Background(), h.callerToken, requestID,
		h.recipientID, h.inspection.ActivationID, "private instruction")
	requireRejectedCode(t, err, "not_found")
	caller, err := h.service.ResolveCaller(h.callerToken)
	if err != nil {
		t.Fatal(err)
	}
	if _, found, findErr := h.manager.Messages().FindSend(caller.SessionID, requestID); findErr != nil || found {
		t.Fatalf("deleted recipient send receipt found=%v err=%v", found, findErr)
	}
	messageID := uuid.NewSHA1(uuid.NameSpaceURL, []byte("kankuro-message:"+caller.SessionID+":"+requestID)).String()
	if _, getErr := h.manager.Messages().GetMessage(messageID); getErr == nil {
		t.Fatal("send recreated retrievable content after recipient deletion")
	}
}

func TestReservedSendCannotPublishAfterRecipientDeletion(t *testing.T) {
	h := newMessagingHarness(t)
	admission, err := h.manager.ReserveMessageAdmission(h.callerToken, h.recipientID, h.inspection.ActivationID)
	if err != nil {
		t.Fatal(err)
	}
	defer admission.Release()
	if err = h.manager.Kill(h.recipientID); err != nil {
		t.Fatal(err)
	}
	if err = removeSessionTest(h.service, h.recipientID); err != nil {
		t.Fatal(err)
	}
	caller, err := h.service.ResolveCaller(h.callerToken)
	if err != nil {
		t.Fatal(err)
	}
	requestID, messageID := uuid.NewString(), uuid.NewString()
	fingerprint := session.SendFingerprint(h.recipientID, h.inspection.ActivationID, "private instruction")
	intent := session.MessageRecord{ID: messageID, DashboardID: caller.DashboardID, SenderSessionID: caller.SessionID,
		SenderActivationID: caller.ActivationID, RecipientSessionID: h.recipientID,
		RecipientActivationID: h.inspection.ActivationID, Text: "private instruction"}
	called := false
	err = h.manager.PublishMessageAdmission(admission, func() error {
		called = true
		_, createErr := h.manager.Messages().CreateIntent(intent, requestID, fingerprint)
		return createErr
	})
	var admissionErr *session.MessagingError
	if !errors.As(err, &admissionErr) || admissionErr.Code != "lifecycle_conflict" {
		t.Fatalf("publication error = %v, want lifecycle_conflict", err)
	}
	if called {
		t.Fatal("intent transaction ran after recipient deletion")
	}
	if _, getErr := h.manager.Messages().GetMessage(messageID); getErr == nil {
		t.Fatal("deleted recipient's message content was published")
	}
}

//nolint:gocognit // The four-worker admission/stop/shutdown interleaving is the regression under test.
func TestQueuedMessageReservationsCannotDeadlockStopWorkersOrShutdown(t *testing.T) {
	h := newMessagingHarness(t)
	runner := h.service.operations
	card, err := h.service.loadCard(h.inspection.CardID)
	if err != nil {
		t.Fatal(err)
	}
	targets := []string{h.recipientID}
	for len(targets) < operationWorkers {
		id, launchErr := launchCardTest(h.service, card)
		if launchErr != nil {
			t.Fatal(launchErr)
		}
		targets = append(targets, id)
	}
	admissions := make([]*session.MessageAdmission, 0, len(targets))
	activations := make([]string, 0, len(targets))
	defer func() {
		for _, admission := range admissions {
			admission.Release()
		}
	}()
	for _, target := range targets {
		snapshot, found := h.manager.GetSession(target)
		if !found {
			t.Fatalf("target %s disappeared", target)
		}
		admission, reserveErr := h.manager.ReserveMessageAdmission(h.callerToken, target, snapshot.ActivationID)
		if reserveErr != nil {
			t.Fatal(reserveErr)
		}
		admissions = append(admissions, admission)
		activations = append(activations, snapshot.ActivationID)
	}
	messageIDs := make([]string, len(targets))
	requestIDs := make([]string, len(targets))
	argumentsByTarget := make([]sendMessageArguments, len(targets))
	for i, target := range targets {
		messageIDs[i], requestIDs[i] = fmt.Sprintf("queued-review-message-%d", i), uuid.NewString()
		fingerprint := session.SendFingerprint(target, activations[i], "task")
		argumentsByTarget[i] = sendMessageArguments{target, activations[i], fingerprint}
		intent := session.MessageRecord{ID: messageIDs[i], DashboardID: DefaultDashboardID,
			SenderSessionID: h.callerID, RecipientSessionID: target, RecipientActivationID: activations[i], Text: "task"}
		if publishErr := h.manager.PublishMessageAdmission(admissions[i], func() error {
			_, createErr := h.manager.Messages().CreateIntent(intent, requestIDs[i], fingerprint)
			return createErr
		}); publishErr != nil {
			t.Fatal(publishErr)
		}
	}

	stoppedContext := context.Background()
	stopResults := make(chan error, len(targets))
	var stopWait sync.WaitGroup
	for _, target := range targets {
		stopWait.Add(1)
		go func(sessionID string) {
			defer stopWait.Done()
			_, stopErr := h.service.StopSession(stoppedContext, h.callerToken, uuid.NewString(), sessionID)
			stopResults <- stopErr
		}(target)
	}
	for i, target := range targets {
		admission, activation := admissions[i], activations[i]
		messageID := messageIDs[i]
		operation, submitErr := runner.SubmitWithAdmission(stoppedContext, h.callerToken, requestIDs[i],
			messageOperation, target, argumentsByTarget[i], nil,
			func(operationContext context.Context, _ Operation) (json.RawMessage, error) {
				_, sendErr := h.manager.SubmitMessage(operationContext, admission, h.callerToken,
					messageID, target, activation, "task")
				return nil, sendErr
			},
			func(Operation) {
				_ = h.manager.Messages().UpdateDelivery(messageID, session.MessageNotSubmitted,
					"operation was canceled before prompt submission began")
				admission.Release()
			},
		)
		if submitErr != nil || operation.ID == "" {
			t.Fatalf("send admission = %#v, %v", operation, submitErr)
		}
		if _, cancelErr := h.service.CancelOperation(h.callerToken, operation.ID); cancelErr != nil {
			t.Fatal(cancelErr)
		}
	}
	stopWait.Wait()
	close(stopResults)
	for stopErr := range stopResults {
		if stopErr != nil {
			t.Fatalf("stop operation failed: %v", stopErr)
		}
	}

	shutdownDone := make(chan struct{})
	go func() {
		runner.Shutdown()
		close(shutdownDone)
	}()
	select {
	case <-shutdownDone:
	case <-time.After(5 * time.Second):
		t.Fatal("operation shutdown deadlocked behind queued message reservations")
	}
	for _, target := range targets {
		snapshot, found := h.manager.GetSession(target)
		if !found || snapshot.Status != session.StatusTerminated {
			t.Fatalf("target process %s was not stopped during cleanup: %#v", target, snapshot)
		}
	}
}

func TestReplacementActivationCannotReplayPreviousReport(t *testing.T) {
	h := newMessagingHarness(t)
	result := h.send(t)
	reportRequestID := uuid.NewString()
	if _, err := h.service.ReportReply(h.recipientToken, reportRequestID, result.Result.MessageID, "completed", "done"); err != nil {
		t.Fatal(err)
	}
	if err := h.manager.Kill(h.recipientID); err != nil {
		t.Fatal(err)
	}
	relaunched, err := h.service.RelaunchSession(context.Background(), h.callerToken, uuid.NewString(), h.recipientID, nil)
	if err != nil || relaunched.State != OperationSucceeded {
		t.Fatalf("recipient relaunch = %#v, %v", relaunched, err)
	}
	newToken := readActivationToken(t, h.workspace)
	newInspection, err := h.service.GetSession(h.callerToken, h.recipientID)
	if err != nil || newInspection.ActivationID == h.inspection.ActivationID {
		t.Fatalf("replacement inspection = %#v, %v", newInspection, err)
	}
	_, err = h.service.ReportReply(newToken, reportRequestID, result.Result.MessageID, "completed", "done")
	requireRejectedCode(t, err, "stale_activation")
}

//nolint:gocognit // One deterministic waiter harness checks every lifecycle wake reason.
func TestReplyWaitEndsOnRecipientStopAndCallerRevocation(t *testing.T) {
	for _, test := range []struct {
		name       string
		stopCaller bool
		shutdown   bool
		wantCode   string
	}{
		{name: "recipient stop", wantCode: "recipient_unavailable"},
		{name: "caller revocation", stopCaller: true, wantCode: "forbidden"},
		{name: "manager shutdown", shutdown: true, wantCode: "forbidden"},
	} {
		t.Run(test.name, func(t *testing.T) {
			h := newMessagingHarness(t)
			result := h.send(t)
			store := h.manager.Messages()
			for range reviewWaiterSlots {
				if !store.AcquireWait() {
					t.Fatal("failed to reserve setup waiter slots")
				}
			}
			defer func() {
				for range reviewWaiterSlots {
					store.ReleaseWait()
				}
			}()
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			done := make(chan error, 1)
			go func() {
				_, waitErr := h.service.GetReplies(ctx, h.callerToken, result.Result.MessageID, "", 50, 4*time.Second)
				done <- waitErr
			}()
			deadline := time.Now().Add(time.Second)
			for {
				if !store.AcquireWait() {
					break
				}
				store.ReleaseWait()
				if time.Now().After(deadline) {
					cancel()
					t.Fatal("reply waiter did not acquire its bounded slot")
				}
				time.Sleep(time.Millisecond)
			}
			if test.shutdown {
				h.manager.Shutdown()
			} else {
				stopID := h.recipientID
				if test.stopCaller {
					stopID = h.callerID
				}
				if err := h.manager.Kill(stopID); err != nil {
					t.Fatal(err)
				}
			}
			select {
			case waitErr := <-done:
				requireRejectedCode(t, waitErr, test.wantCode)
			case <-time.After(time.Second):
				cancel()
				t.Fatal("reply wait remained blocked after lifecycle change")
			}
			if !store.AcquireWait() {
				t.Fatal("reply wait did not release its waiter slot")
			}
			store.ReleaseWait()
		})
	}
}

func TestReplyWaitReturnsWhenRecipientHistoryWasDeleted(t *testing.T) {
	h := newMessagingHarness(t)
	result := h.send(t)
	if err := h.manager.Kill(h.recipientID); err != nil {
		t.Fatal(err)
	}
	if err := removeSessionTest(h.service, h.recipientID); err != nil {
		t.Fatal(err)
	}
	start := time.Now()
	_, err := h.service.GetReplies(context.Background(), h.callerToken, result.Result.MessageID, "", 50, 4*time.Second)
	requireRejectedCode(t, err, "history_unavailable")
	if time.Since(start) > time.Second {
		t.Fatal("reply wait did not promptly observe deleted message history")
	}
}

func TestMessagingErrorsHaveStableInputAndRetryCodes(t *testing.T) {
	h := newMessagingHarness(t)
	if _, err := h.service.GetReplies(context.Background(), h.callerToken, "message", "%%%", 10, 0); err == nil {
		t.Fatal("malformed cursor accepted")
	} else {
		requireRejectedCode(t, err, "invalid_cursor")
	}
	if _, err := h.service.GetReplies(context.Background(), h.callerToken, "", "", 10, 11*time.Second); err == nil {
		t.Fatal("out-of-range wait accepted")
	} else {
		requireRejectedCode(t, err, "invalid_request")
	}
	result := h.send(t)
	requestID := uuid.NewString()
	if _, err := h.service.ReportReply(h.recipientToken, requestID, result.Result.MessageID, "completed", "done"); err != nil {
		t.Fatal(err)
	}
	_, err := h.service.ReportReply(h.recipientToken, requestID, result.Result.MessageID, "completed", "changed")
	requireRejectedCode(t, err, "request_id_conflict")
	if _, err = h.service.GetMessage(h.callerToken, "missing-message"); err == nil {
		t.Fatal("missing history returned without an error")
	} else {
		requireRejectedCode(t, err, "history_unavailable")
	}
	storageErr := messagingRejection(&session.MessagingError{Code: "storage_unavailable",
		Message: "report could not be saved; retry with the same request_id", Cause: errors.New("private filesystem path")})
	requireRejectedCode(t, storageErr, "storage_unavailable")
	if strings.Contains(storageErr.Error(), "private filesystem path") {
		t.Fatal("safe storage rejection exposed filesystem diagnostics")
	}
}

//nolint:govet // Error variables are scoped to individual test assertions.
func TestDeleteSessionReportsTrustCleanupWarningAndAllowsRetry(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	h := newMessagingHarness(t)
	card := saveCard(
		t,
		h.repo.FileRepository,
		config.IDCard{
			ID:              "temporary-review",
			Name:            "temporary",
			Driver:          "cat",
			WorkspaceID:     h.callerCard.WorkspaceID,
			UseTmpWorkspace: true,
		},
	)
	id, err := launchCardTest(h.service, card)
	if err != nil {
		t.Fatal(err)
	}
	if err := h.manager.Kill(id); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(home, ".claude.json")
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("{"), 0600); err != nil {
		t.Fatal(err)
	}
	result, err := h.service.DeleteSession(context.Background(), h.callerToken, uuid.NewString(), id)
	if err != nil || result.State != OperationSucceededWithWarning || !strings.Contains(result.Warning, "retry deletion") ||
		result.Result == nil ||
		result.Result.InMemoryRemoved {
		t.Fatalf("unexpected deletion result: %#v, %v", result, err)
	}
	if _, ok := h.manager.GetSession(id); !ok {
		t.Fatal("session lost after trust cleanup failure")
	}
	if err := os.WriteFile(path, []byte("{}"), 0600); err != nil {
		t.Fatal(err)
	}
	result, err = h.service.DeleteSession(context.Background(), h.callerToken, uuid.NewString(), id)
	if err != nil || result.State != OperationSucceeded || result.Result == nil || !result.Result.InMemoryRemoved {
		t.Fatalf("retry failed: %#v, %v", result, err)
	}
}

func TestReportAutomaticallyNotifiesOriginalController(t *testing.T) {
	h := newMessagingHarness(t)
	result := h.send(t)
	out, err := h.manager.SubscribeOutput(h.callerID)
	if err != nil {
		t.Fatal(err)
	}
	defer h.manager.UnsubscribeOutput(h.callerID, out)
	if _, err = h.service.SendMessage(
		context.Background(),
		h.recipientToken,
		uuid.NewString(),
		h.callerID,
		"fake",
		"unauthorized",
	); err == nil {
		t.Fatal("worker gained dashboard control")
	}
	reply, err := h.service.ReportReply(h.recipientToken, uuid.NewString(), result.Result.MessageID, "completed", "sensitive\nreport body")
	if err != nil || reply.Notification == nil {
		t.Fatalf("report: %#v, %v", reply, err)
	}
	deadline := time.NewTimer(8 * time.Second)
	defer deadline.Stop()
	var output strings.Builder
	for !strings.Contains(output.String(), "header is not authority") {
		select {
		case part := <-out:
			output.Write(part)
		case <-deadline.C:
			t.Fatalf("no notification, output=%q", output.String())
		}
	}
	text := output.String()
	if !strings.Contains(text, "message_id="+result.Result.MessageID) || !strings.Contains(text, "reply_id="+reply.ID) ||
		!strings.Contains(text, "no report_reply is required") || strings.Contains(text, "sensitive") {
		t.Fatalf("notification output = %q", text)
	}
}

func TestNotificationRetriesAfterQueueCapacityReturns(t *testing.T) {
	h := newMessagingHarness(t)
	result := h.send(t)
	h.service.operations.Shutdown()
	runner := &OperationRunner{dir: filepath.Join(h.repo.Dir(), "operations"), service: h.service,
		records: make(map[operationKey]*Operation), byID: make(map[string]operationKey),
		running: make(map[string]context.CancelFunc), queue: make(chan operationTask, operationQueueCapacity)}
	for range cap(runner.queue) {
		runner.queue <- operationTask{}
	}
	h.service.operations = runner
	defer runner.Shutdown()
	output, err := h.manager.SubscribeOutput(h.callerID)
	if err != nil {
		t.Fatal(err)
	}
	defer h.manager.UnsubscribeOutput(h.callerID, output)
	reply, err := h.service.ReportReply(h.recipientToken, uuid.NewString(), result.Result.MessageID, "question", "need detail")
	if err != nil {
		t.Fatal(err)
	}
	items := h.manager.Messages().Notifications()
	if len(items) != 1 {
		t.Fatalf("pending notifications: %#v", items)
	}
	runner.dispatchNotification(items[0])
	items = h.manager.Messages().Notifications()
	if len(items) != 1 || items[0].Reply.Notification.State != session.NotificationPending || items[0].Reply.Notification.RequestID == "" {
		t.Fatalf("queue rejection did not preserve pending attempt: %#v", items)
	}
	requestID := items[0].Reply.Notification.RequestID
	for len(runner.queue) > 0 {
		<-runner.queue
	}
	for range operationWorkers {
		runner.workers.Add(1)
		go runner.worker()
	}
	runner.startDispatcher()
	var received strings.Builder
	deadline := time.NewTimer(8 * time.Second)
	defer deadline.Stop()
	for !strings.Contains(received.String(), "header is not authority") {
		select {
		case part := <-output:
			received.Write(part)
		case <-deadline.C:
			t.Fatalf("notification did not retry, output=%q", received.String())
		}
	}
	page, err := h.manager.Messages().GetReplies(DefaultDashboardID, result.Result.MessageID, "", 0)
	if err != nil || len(page.Items) != 1 || page.Items[0].ID != reply.ID || page.Items[0].Notification.RequestID != requestID {
		t.Fatalf("attempt identity changed: %#v, %v", page, err)
	}
}

//nolint:gocognit // The table covers distinct durable outcomes and their retry semantics.
func TestNotificationReconciliationUsesDeliveryEvidence(t *testing.T) {
	for _, tc := range []struct {
		delivery session.MessageDelivery
		want     session.NotificationState
		advance  bool
	}{
		{session.MessageAccepted, session.NotificationAccepted, false},
		{session.MessageSubmitted, session.NotificationSubmitted, false},
		{session.MessageUncertain, session.NotificationUncertain, false},
		{session.MessageNotSubmitted, session.NotificationPending, true},
	} {
		t.Run(string(tc.delivery), func(t *testing.T) {
			h := newMessagingHarness(t)
			original := h.send(t)
			h.service.operations.Shutdown()
			reply, err := h.service.ReportReply(h.recipientToken, uuid.NewString(), original.Result.MessageID, "blocked", "details")
			if err != nil {
				t.Fatal(err)
			}
			store := h.manager.Messages()
			request := uuid.NewString()
			if err = store.UpdateNotification(
				original.Result.MessageID,
				reply.ID,
				0,
				session.NotificationPending,
				func(n *session.ReplyNotification) { n.RequestID = request },
			); err != nil {
				t.Fatal(err)
			}
			parent, err := store.GetMessage(original.Result.MessageID)
			if err != nil {
				t.Fatal(err)
			}
			notification := session.MessageRecord{
				ID:                    uuid.NewString(),
				Kind:                  "report_notification",
				OriginalMessageID:     parent.ID,
				ReplyID:               reply.ID,
				ReportKind:            reply.Kind,
				DashboardID:           parent.DashboardID,
				SenderSessionID:       reply.SenderSessionID,
				SenderActivationID:    reply.SenderActivationID,
				RecipientSessionID:    parent.SenderSessionID,
				RecipientActivationID: parent.SenderActivationID,
				Text:                  "report notification",
			}
			if _, err = store.CreateIntentWithAdmission(notification, request, "fingerprint", func() error { return nil }); err != nil {
				t.Fatal(err)
			}
			if err = store.UpdateDelivery(notification.ID, tc.delivery, ""); err != nil {
				t.Fatal(err)
			}
			items := store.Notifications()
			if len(items) != 1 {
				t.Fatalf("notifications = %#v", items)
			}
			h.service.operations.reconcileNotification(items[0])
			page, err := store.GetReplies(DefaultDashboardID, parent.ID, "", 0)
			if err != nil || len(page.Items) != 1 {
				t.Fatalf("page = %#v, %v", page, err)
			}
			got := page.Items[0].Notification
			if got.State != tc.want || (got.Attempt == 1) != tc.advance {
				t.Fatalf("notification = %#v", got)
			}
			// A stale callback cannot advance or downgrade a completed attempt.
			h.service.operations.reconcileNotification(items[0])
			page, err = store.GetReplies(DefaultDashboardID, parent.ID, "", 0)
			if err != nil || page.Items[0].Notification.State != tc.want || (page.Items[0].Notification.Attempt == 1) != tc.advance {
				t.Fatalf("stale reconciliation: %#v, %v", page, err)
			}
		})
	}
}

func TestSavedNotificationSurvivesReporterStopping(t *testing.T) {
	h := newMessagingHarness(t)
	original := h.send(t)
	h.service.operations.Shutdown()
	output, err := h.manager.SubscribeOutput(h.callerID)
	if err != nil {
		t.Fatal(err)
	}
	defer h.manager.UnsubscribeOutput(h.callerID, output)
	if _, err = h.service.ReportReply(h.recipientToken, uuid.NewString(), original.Result.MessageID, "completed", "done"); err != nil {
		t.Fatal(err)
	}
	if err = h.manager.Kill(h.recipientID); err != nil {
		t.Fatal(err)
	}
	runner := NewOperationRunner(h.repo.Dir(), h.service)
	h.service.SetOperationRunner(runner)
	defer runner.Shutdown()
	deadline := time.NewTimer(8 * time.Second)
	defer deadline.Stop()
	var received strings.Builder
	for !strings.Contains(received.String(), "header is not authority") {
		select {
		case part := <-output:
			received.Write(part)
		case <-deadline.C:
			t.Fatalf("saved report not delivered after reporter stopped: %q", received.String())
		}
	}
}

func TestStoppedOriginalActivationMakesNotificationUnavailable(t *testing.T) {
	h := newMessagingHarness(t)
	original := h.send(t)
	h.service.operations.Shutdown()
	if _, err := h.service.ReportReply(h.recipientToken, uuid.NewString(), original.Result.MessageID, "completed", "done"); err != nil {
		t.Fatal(err)
	}
	if err := h.manager.Kill(h.callerID); err != nil {
		t.Fatal(err)
	}
	if err := relaunchCardTest(h.service, h.callerID); err != nil {
		t.Fatal(err)
	}
	items := h.manager.Messages().Notifications()
	if len(items) != 1 {
		t.Fatalf("pending notifications = %#v", items)
	}
	h.service.operations.dispatchNotification(items[0])
	page, err := h.manager.Messages().GetReplies(DefaultDashboardID, original.Result.MessageID, "", 0)
	if err != nil || len(page.Items) != 1 || page.Items[0].Notification.State != session.NotificationUnavailable {
		t.Fatalf("notification after sender stop = %#v, %v", page, err)
	}
}

func TestNotificationRetriesAfterJournalFailure(t *testing.T) {
	h := newMessagingHarness(t)
	original := h.send(t)
	output, err := h.manager.SubscribeOutput(h.callerID)
	if err != nil {
		t.Fatal(err)
	}
	defer h.manager.UnsubscribeOutput(h.callerID, output)
	failed := make(chan struct{}, 1)
	runner := h.service.operations
	runner.mu.Lock()
	runner.writeFileAtomic = func(path string, data []byte) error {
		select {
		case failed <- struct{}{}:
			runner.writeFileAtomic = nil
			return errors.New("temporary journal failure")
		default:
			return fsutil.WriteFileAtomic(path, data, operationFileMode)
		}
	}
	runner.mu.Unlock()
	if _, err = h.service.ReportReply(h.recipientToken, uuid.NewString(), original.Result.MessageID, "completed", "done"); err != nil {
		t.Fatal(err)
	}
	deadline := time.NewTimer(8 * time.Second)
	defer deadline.Stop()
	select {
	case <-failed:
	case <-deadline.C:
		t.Fatal("journal failure was not exercised")
	}
	var received strings.Builder
	for !strings.Contains(received.String(), "header is not authority") {
		select {
		case part := <-output:
			received.Write(part)
		case <-deadline.C:
			t.Fatalf("notification did not recover from journal failure: %q", received.String())
		}
	}
}

func TestDeletingReportHistoryCancelsAdmittedNotification(t *testing.T) {
	h := newMessagingHarness(t)
	original := h.send(t)
	h.service.operations.Shutdown()
	reply, err := h.service.ReportReply(h.recipientToken, uuid.NewString(), original.Result.MessageID, "question", "details")
	if err != nil {
		t.Fatal(err)
	}
	store := h.manager.Messages()
	request := uuid.NewString()
	if err = store.UpdateNotification(
		original.Result.MessageID,
		reply.ID,
		0,
		session.NotificationPending,
		func(n *session.ReplyNotification) { n.RequestID = request },
	); err != nil {
		t.Fatal(err)
	}
	parent, err := store.GetMessage(original.Result.MessageID)
	if err != nil {
		t.Fatal(err)
	}
	admission, err := h.manager.ReserveReportNotification(parent.ID, reply.ID)
	if err != nil {
		t.Fatal(err)
	}
	defer admission.Release()
	intent := session.MessageRecord{
		ID:                    uuid.NewString(),
		Kind:                  "report_notification",
		OriginalMessageID:     parent.ID,
		ReplyID:               reply.ID,
		ReportKind:            reply.Kind,
		DashboardID:           parent.DashboardID,
		SenderSessionID:       reply.SenderSessionID,
		SenderActivationID:    reply.SenderActivationID,
		RecipientSessionID:    parent.SenderSessionID,
		RecipientActivationID: parent.SenderActivationID,
		Text:                  "report notification",
	}
	if err = h.manager.PublishMessageAdmission(admission, func() error {
		_, createErr := store.CreateIntentWithAdmission(intent, request, "fingerprint", func() error { return nil })
		return createErr
	}); err != nil {
		t.Fatal(err)
	}
	if err = store.DeleteDashboard(parent.DashboardID); err != nil {
		t.Fatal(err)
	}
	_, err = h.manager.SubmitMessage(
		context.Background(),
		admission,
		"",
		intent.ID,
		intent.RecipientSessionID,
		intent.RecipientActivationID,
		intent.Text,
	)
	if err == nil {
		t.Fatal("notification submitted after parent history deletion")
	}
	if len(store.Notifications()) != 0 {
		t.Fatal("deleted history was recreated")
	}
}

func TestNotificationRetryBackoff(t *testing.T) {
	var previous time.Duration
	for _, seconds := range []int{1, 2, 4, 8, 16, 30, 30} {
		previous = nextNotificationDelay(previous)
		if previous != time.Duration(seconds)*time.Second {
			t.Fatalf("retry delay = %s, want %d seconds", previous, seconds)
		}
	}
}

func TestConcurrentNotificationDispatchAdmitsOneAttempt(t *testing.T) {
	h := newMessagingHarness(t)
	original := h.send(t)
	h.service.operations.Shutdown()
	if _, err := h.service.ReportReply(h.recipientToken, uuid.NewString(), original.Result.MessageID, "completed", "done"); err != nil {
		t.Fatal(err)
	}
	runner := NewOperationRunner(h.repo.Dir(), h.service)
	defer runner.Shutdown()
	items := h.manager.Messages().Notifications()
	if len(items) != 1 {
		t.Fatalf("notifications = %#v", items)
	}
	var group sync.WaitGroup
	for range 12 {
		group.Go(func() { runner.dispatchNotification(items[0]) })
	}
	group.Wait()
	runner.mu.Lock()
	count := 0
	for _, operation := range runner.records {
		if operation.Kind == messageOperation && operation.TargetSession == h.callerID {
			count++
		}
	}
	runner.mu.Unlock()
	if count != 1 {
		t.Fatalf("admitted %d notification attempts, want one", count)
	}
}

func TestDashboardDeletionManifestCancelsAdmittedNotification(t *testing.T) {
	h := newMessagingHarness(t)
	original := h.send(t)
	h.service.operations.Shutdown()
	reply, err := h.service.ReportReply(h.recipientToken, uuid.NewString(), original.Result.MessageID, "question", "details")
	if err != nil {
		t.Fatal(err)
	}
	store := h.manager.Messages()
	request := uuid.NewString()
	if err = store.UpdateNotification(
		original.Result.MessageID,
		reply.ID,
		0,
		session.NotificationPending,
		func(n *session.ReplyNotification) { n.RequestID = request },
	); err != nil {
		t.Fatal(err)
	}
	parent, err := store.GetMessage(original.Result.MessageID)
	if err != nil {
		t.Fatal(err)
	}
	admission, err := h.manager.ReserveReportNotification(parent.ID, reply.ID)
	if err != nil {
		t.Fatal(err)
	}
	defer admission.Release()
	intent := session.MessageRecord{
		ID:                    uuid.NewString(),
		Kind:                  "report_notification",
		OriginalMessageID:     parent.ID,
		ReplyID:               reply.ID,
		ReportKind:            reply.Kind,
		DashboardID:           parent.DashboardID,
		SenderSessionID:       reply.SenderSessionID,
		SenderActivationID:    reply.SenderActivationID,
		RecipientSessionID:    parent.SenderSessionID,
		RecipientActivationID: parent.SenderActivationID,
		Text:                  "report notification",
	}
	if err = h.manager.PublishMessageAdmission(admission, func() error {
		_, createErr := store.CreateIntentWithAdmission(intent, request, "fingerprint", func() error { return nil })
		return createErr
	}); err != nil {
		t.Fatal(err)
	}
	if err = h.repo.BeginDashboardDeletion(
		config.DashboardDeletion{SchemaVersion: 1, DashboardID: parent.DashboardID, Name: "default"},
	); err != nil {
		t.Fatal(err)
	}
	result, err := h.manager.SubmitMessage(
		context.Background(),
		admission,
		"",
		intent.ID,
		intent.RecipientSessionID,
		intent.RecipientActivationID,
		intent.Text,
	)
	if err == nil || result.Delivery != session.MessageNotSubmitted {
		t.Fatalf("notification after dashboard deletion: %#v, %v", result, err)
	}
}

func TestNotificationSchedulingPreservesRecipientOrderAndFairness(t *testing.T) {
	now := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	item := func(id, recipient string, state session.NotificationState) session.PendingNotification {
		return session.PendingNotification{Message: session.MessageRecord{SenderSessionID: recipient, SenderActivationID: "activation"},
			Reply: session.MessageReply{ID: id, Notification: &session.ReplyNotification{State: state}}}
	}
	items := []session.PendingNotification{
		item("first", "a", session.NotificationPending),
		item("second", "a", session.NotificationPending),
		item("third", "b", session.NotificationPending),
	}
	schedule := map[string]notificationSchedule{"first": {due: now.Add(time.Second)}}
	due := dueNotifications(items, schedule, now)
	if len(due) != 1 || due[0].Reply.ID != "third" {
		t.Fatalf("due while first backs off = %#v", due)
	}
	due = dueNotifications(items, schedule, now.Add(time.Second))
	if len(due) != 2 || due[0].Reply.ID != "first" || due[1].Reply.ID != "third" {
		t.Fatalf("due after backoff = %#v", due)
	}
	items[0].Reply.Notification.State = session.NotificationAccepted
	due = dueNotifications(items, schedule, now.Add(time.Second))
	if len(due) != 2 || due[0].Reply.ID != "first" || due[1].Reply.ID != "third" {
		t.Fatalf("accepted predecessor did not preserve recipient order: %#v", due)
	}
}
