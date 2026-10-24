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
	callerID, err := service.Launch(callerCard)
	if err != nil {
		t.Fatal(err)
	}
	callerToken := readActivationToken(t, workspace)
	recipientID, err := service.Launch(recipientCard)
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
	if err := h.service.RemoveSession(h.recipientID); err != nil {
		t.Fatal(err)
	}
	requestID := uuid.NewString()
	_, err := h.service.SendMessage(context.Background(), h.callerToken, requestID,
		h.recipientID, h.inspection.ActivationID, "private instruction")
	requireRejectedCode(t, err, "lifecycle_conflict")
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
	if err = h.service.RemoveSession(h.recipientID); err != nil {
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
		id, launchErr := h.service.Launch(card)
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
	if err := h.service.RemoveSession(h.recipientID); err != nil {
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
