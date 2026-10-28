//nolint:testpackage // Capacity tests seed the private journal snapshot without 10,000 repeated writes.
package session

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/BlackGiack/kankuro/internal/config"
	"github.com/BlackGiack/kankuro/internal/driver"
	"github.com/google/uuid"
)

type failAfterPromptDriver struct {
	driver.AgentDriver

	path       string
	backupPath string
	calls      int
}

type promptCountingDriver struct {
	driver.AgentDriver

	calls atomic.Int32
}

func (d *promptCountingDriver) SubmitPrompt(context.Context, string, string, string) (driver.SubmissionEvidence, error) {
	d.calls.Add(1)
	return driver.SubmissionEvidence{State: driver.SubmissionComplete}, nil
}

func (d *failAfterPromptDriver) SubmitPrompt(
	ctx context.Context,
	messageID, senderSessionID, text string,
) (driver.SubmissionEvidence, error) {
	d.calls++
	evidence, err := d.AgentDriver.SubmitPrompt(ctx, messageID, senderSessionID, text)
	if d.calls == 1 {
		d.backupPath = d.path + ".review-backup"
		if renameErr := os.Rename(d.path, d.backupPath); renameErr != nil {
			return evidence, errors.Join(err, renameErr)
		}
		if mkdirErr := os.Mkdir(d.path, 0o700); mkdirErr != nil {
			return evidence, errors.Join(err, mkdirErr)
		}
	}
	return evidence, err
}

func TestSubmitMessageStorageFailuresReleaseRuntimeOrderingWithoutReplay(t *testing.T) {
	manager, callerID, callerToken, recipientID, recipientActivationID := newReviewMessagingManager(t)
	caller, found := manager.GetSession(callerID)
	if !found {
		t.Fatal("caller session disappeared")
	}
	store := manager.Messages()
	var recipientDriver *failAfterPromptDriver
	manager.mu.Lock()
	recipient := manager.sessions[recipientID]
	recipientDriver = &failAfterPromptDriver{AgentDriver: recipient.Driver, path: store.path}
	recipient.Driver = recipientDriver
	manager.mu.Unlock()

	firstID := submitReviewMessage(t, manager, caller, callerToken, recipientID, recipientActivationID, "first prompt")
	first, err := store.GetMessage(firstID)
	if err != nil || first.Delivery != MessageUncertain || first.Warning == "" {
		t.Fatalf("first message after terminal persistence failure = %#v, %v", first, err)
	}

	secondID := submitReviewMessage(t, manager, caller, callerToken, recipientID, recipientActivationID, "second prompt")
	second, err := store.GetMessage(secondID)
	if err != nil || second.Delivery != MessageSubmitted {
		t.Fatalf("later message after storage recovered = %#v, %v", second, err)
	}
	if recipientDriver.calls != 2 {
		t.Fatalf("prompt submission calls = %d, want two distinct submissions", recipientDriver.calls)
	}

	thirdID := createReviewIntent(t, manager, caller, recipientID, recipientActivationID, "canceled prompt")
	thirdAdmission, err := manager.ReserveMessageAdmission(callerToken, recipientID, recipientActivationID)
	if err != nil {
		t.Fatal(err)
	}
	if err = manager.PublishMessageAdmission(thirdAdmission, func() error { return nil }); err != nil {
		t.Fatal(err)
	}
	previousPath := store.path
	store.path = filepath.Dir(previousPath)
	canceled, cancel := context.WithCancel(context.Background())
	cancel()
	_, canceledErr := manager.SubmitMessage(canceled, thirdAdmission, callerToken,
		thirdID, recipientID, recipientActivationID, "canceled prompt")
	store.path = previousPath
	if canceledErr == nil {
		t.Fatal("pre-write cancellation unexpectedly succeeded")
	}
	third, err := store.GetMessage(thirdID)
	if err != nil || third.Delivery != MessageNotSubmitted || third.Warning == "" {
		t.Fatalf("canceled message after state persistence failure = %#v, %v", third, err)
	}

	fourthID := submitReviewMessage(t, manager, caller, callerToken, recipientID, recipientActivationID, "fourth prompt")
	fourth, err := store.GetMessage(fourthID)
	if err != nil || fourth.Delivery != MessageSubmitted {
		t.Fatalf("later message after pre-write cancellation = %#v, %v", fourth, err)
	}
	if recipientDriver.calls != 3 {
		t.Fatalf("prompt submission calls after canceled message = %d, want three", recipientDriver.calls)
	}
}

func TestConcurrentAdmissionReleaseAndLifecycleInvalidationUseManagerLock(t *testing.T) {
	sender := &Session{ID: "sender", activeSubmissions: 1, submissionsDone: make(chan struct{})}
	recipient := &Session{ID: "recipient", activeSubmissions: 1, submissionsDone: make(chan struct{})}
	manager := &Manager{
		sessions:          map[string]*Session{sender.ID: sender, recipient.ID: recipient},
		messageAdmissions: make(map[*MessageAdmission]struct{}),
	}
	admissionContext, cancel := context.WithCancel(context.Background())
	defer cancel()
	admission := &MessageAdmission{manager: manager, sender: sender, recipient: recipient, ctx: admissionContext, cancel: cancel}
	manager.messageAdmissions[admission] = struct{}{}

	manager.mu.Lock()
	releaseStarted := make(chan struct{})
	releaseDone := make(chan struct{})
	go func() {
		close(releaseStarted)
		admission.Release()
		close(releaseDone)
	}()
	<-releaseStarted
	manager.cancelPendingMessageAdmissionsLocked(recipient)
	manager.mu.Unlock()

	select {
	case <-releaseDone:
	case <-time.After(time.Second):
		t.Fatal("release and lifecycle invalidation deadlocked")
	}
	if sender.activeSubmissions != 0 || recipient.activeSubmissions != 0 {
		t.Fatalf("reservation counts after concurrent release: sender=%d recipient=%d",
			sender.activeSubmissions, recipient.activeSubmissions)
	}
	responsive := make(chan struct{})
	go func() {
		manager.mu.RLock()
		_ = manager.sessions[sender.ID]
		manager.mu.RUnlock()
		close(responsive)
	}()
	select {
	case <-responsive:
	case <-time.After(time.Second):
		t.Fatal("manager operations remained blocked after reservation cleanup")
	}
}

func TestInvalidatedSubmissionDoesNotReadReplacementDriverOutsideManagerLock(t *testing.T) {
	store, err := NewMessageStore(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	sender := &Session{ID: "sender", ActivationID: "sender-activation", DashboardControl: true, Status: StatusRunning}
	recipient := &Session{ID: "recipient", ActivationID: "recipient-old", Status: StatusRunning}
	firstDriver, replacementDriver := &promptCountingDriver{}, &promptCountingDriver{}
	recipient.Driver = firstDriver
	manager := &Manager{sessions: map[string]*Session{sender.ID: sender, recipient.ID: recipient}, messages: store,
		messageAdmissions: make(map[*MessageAdmission]struct{})}
	const attempts = 64
	admissions := make([]*MessageAdmission, attempts)
	for i := range attempts {
		id := fmt.Sprintf("stale-driver-%d", i)
		admissionContext, admissionCancel := context.WithCancel(context.Background())
		activationID := fmt.Sprintf("recipient-old-%d", i)
		admissions[i] = &MessageAdmission{manager: manager, sender: sender, recipient: recipient,
			senderID: sender.ID, senderActivationID: sender.ActivationID, recipientID: recipient.ID,
			recipientActivationID: activationID, token: "stale-token", ctx: admissionContext, cancel: admissionCancel,
			invalidated: true, published: true}
		if _, err = store.CreateIntent(MessageRecord{ID: id, SenderSessionID: sender.ID,
			RecipientSessionID: recipient.ID, RecipientActivationID: activationID,
			Text: "stale prompt"}, uuid.NewString(), "fingerprint"); err != nil {
			t.Fatal(err)
		}
	}
	start := make(chan struct{})
	var updates sync.WaitGroup
	updates.Go(func() {
		<-start
		for range attempts * 4 {
			manager.mu.Lock()
			recipient.Driver = firstDriver
			manager.mu.Unlock()
			manager.mu.Lock()
			recipient.Driver = replacementDriver
			manager.mu.Unlock()
		}
	})
	results := make(chan error, attempts)
	for i := range attempts {
		go func(index int) {
			<-start
			admission := admissions[index]
			_, submitErr := manager.SubmitMessage(context.Background(), admission, "stale-token",
				fmt.Sprintf("stale-driver-%d", index), recipient.ID, admission.recipientActivationID, "stale prompt")
			results <- submitErr
		}(i)
	}
	close(start)
	for range attempts {
		if err = <-results; err == nil {
			t.Fatal("invalidated admission unexpectedly submitted")
		}
	}
	updates.Wait()
	if firstDriver.calls.Load() != 0 || replacementDriver.calls.Load() != 0 {
		t.Fatalf("stale submission reached driver: original=%d replacement=%d", firstDriver.calls.Load(), replacementDriver.calls.Load())
	}
}

func newReviewMessagingManager(t *testing.T) (*Manager, string, string, string, string) {
	t.Helper()
	repository, err := config.NewFileRepository(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	workspace := t.TempDir()
	if err = repository.SaveWorkspaceSources([]config.WorkspaceSource{{Path: workspace}}); err != nil {
		t.Fatal(err)
	}
	workspaces, err := repository.ListWorkspaces()
	if err != nil || len(workspaces) != 1 {
		t.Fatalf("list workspaces: %v", err)
	}
	manager := NewManager(repository.Dir(), repository, WithStartupDelay(0))
	manager.SetMCPEndpoint("http://127.0.0.1:12345/mcp")
	t.Cleanup(func() { manager.Shutdown() })
	callerCard, err := repository.CreateIDCard(config.IDCard{ID: "review-controller", Name: "controller", Driver: "cat",
		WorkspaceID: workspaces[0].ID, DashboardControl: true, MCPIDs: []string{config.BuiltinMCPKankuro}})
	if err != nil {
		t.Fatal(err)
	}
	recipientCard, err := repository.CreateIDCard(config.IDCard{ID: "review-message-target", Name: "recipient", Driver: "cat",
		WorkspaceID: workspaces[0].ID, MCPIDs: []string{config.BuiltinMCPKankuro}})
	if err != nil {
		t.Fatal(err)
	}
	callerID, err := manager.Launch(callerCard, nil, nil, config.BuiltinMCPs())
	if err != nil {
		t.Fatal(err)
	}
	callerToken := readReviewActivationToken(t, filepath.Join(workspace, ".kankuro", "mcp-servers.json"))
	recipientID, err := manager.Launch(recipientCard, nil, nil, config.BuiltinMCPs())
	if err != nil {
		t.Fatal(err)
	}
	recipient, found := manager.GetSession(recipientID)
	if !found {
		t.Fatal("recipient session was not registered")
	}
	return manager, callerID, callerToken, recipientID, recipient.ActivationID
}

func readReviewActivationToken(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
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
		t.Fatal("controller activation token was not generated")
	}
	return token
}

func createReviewIntent(
	t *testing.T,
	manager *Manager,
	caller Snapshot,
	recipientID, recipientActivationID, text string,
) string {
	t.Helper()
	messageID, requestID := uuid.NewString(), uuid.NewString()
	record := MessageRecord{ID: messageID, DashboardID: "default", SenderSessionID: caller.ID,
		SenderActivationID: caller.ActivationID, RecipientSessionID: recipientID,
		RecipientActivationID: recipientActivationID, Text: text}
	if _, err := manager.Messages().CreateIntent(record, requestID, SendFingerprint(recipientID, recipientActivationID, text)); err != nil {
		t.Fatal(err)
	}
	return messageID
}

func submitReviewMessage(
	t *testing.T,
	manager *Manager,
	caller Snapshot,
	token, recipientID, recipientActivationID, text string,
) string {
	t.Helper()
	admission, err := manager.ReserveMessageAdmission(token, recipientID, recipientActivationID)
	if err != nil {
		t.Fatal(err)
	}
	messageID, requestID := uuid.NewString(), uuid.NewString()
	record := MessageRecord{ID: messageID, DashboardID: "default", SenderSessionID: caller.ID,
		SenderActivationID: caller.ActivationID, RecipientSessionID: recipientID,
		RecipientActivationID: recipientActivationID, Text: text}
	if err = manager.PublishMessageAdmission(admission, func() error {
		_, createErr := manager.Messages().CreateIntent(record, requestID, SendFingerprint(recipientID, recipientActivationID, text))
		return createErr
	}); err != nil {
		t.Fatal(err)
	}
	_, err = manager.SubmitMessage(context.Background(), admission, token, messageID, recipientID, recipientActivationID, text)
	manager.mu.RLock()
	recipientDriver, _ := manager.sessions[recipientID].Driver.(*failAfterPromptDriver)
	manager.mu.RUnlock()
	if recipientDriver != nil && recipientDriver.backupPath != "" {
		if restoreErr := os.Remove(recipientDriver.path); restoreErr != nil {
			t.Fatal(restoreErr)
		}
		if restoreErr := os.Rename(recipientDriver.backupPath, recipientDriver.path); restoreErr != nil {
			t.Fatal(restoreErr)
		}
		recipientDriver.backupPath = ""
	}
	if err != nil && !strings.Contains(err.Error(), "message delivery state could not be saved") {
		t.Fatalf("submit message %s: %v", messageID, err)
	}
	return messageID
}

//nolint:gocognit // The failure matrix covers both terminal outcomes and later FIFO progress.
func TestTerminalPersistenceFailureReleasesInMemorySendOrder(t *testing.T) {
	for _, terminalState := range []MessageDelivery{MessageSubmitted, MessageNotSubmitted} {
		t.Run(string(terminalState), func(t *testing.T) {
			store, err := NewMessageStore(t.TempDir())
			if err != nil {
				t.Fatal(err)
			}
			first := MessageRecord{ID: uuid.NewString(), DashboardID: "dash", SenderSessionID: "sender",
				RecipientSessionID: "recipient", RecipientActivationID: "activation", Text: "first"}
			if _, err = store.CreateIntent(first, uuid.NewString(), SendFingerprint("recipient", "activation", first.Text)); err != nil {
				t.Fatal(err)
			}
			if err = store.UpdateDelivery(first.ID, MessageSubmitting, ""); err != nil {
				t.Fatal(err)
			}
			path := store.path
			store.path = filepath.Dir(path)
			if err = store.UpdateDelivery(first.ID, terminalState, "terminal transition failed to persist"); err == nil {
				t.Fatal("expected terminal state persistence failure")
			}
			store.path = path
			got, err := store.GetMessage(first.ID)
			want := terminalState
			if terminalState == MessageSubmitted {
				want = MessageUncertain
			}
			if err != nil || got.Delivery != want || got.Warning == "" {
				t.Fatalf("in-memory terminal result = %#v, %v; want %q and warning", got, err, want)
			}

			second := MessageRecord{ID: uuid.NewString(), DashboardID: "dash", SenderSessionID: "sender",
				RecipientSessionID: "recipient", RecipientActivationID: "activation", Text: "second"}
			if _, err = store.CreateIntent(second, uuid.NewString(), SendFingerprint("recipient", "activation", second.Text)); err != nil {
				t.Fatal(err)
			}
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()
			if err = store.WaitTurn(ctx, second.ID); err != nil {
				t.Fatalf("later send remained blocked by unsaved terminal state: %v", err)
			}
		})
	}
}

func TestRecordReplyEnforcesCapacityEvictsWholeMessageAndAdvancesCursor(t *testing.T) {
	store, targetID := seedMessageHistoryNearCapacity(t)
	store.data.Sends = append(store.data.Sends, sendReceipt{SenderSessionID: "sender", RequestID: "old-send",
		Fingerprint: "old-fingerprint", MessageID: "m-000"})
	store.data.Reports = append(store.data.Reports, reportReceipt{
		RecipientSessionID:    "recipient",
		RecipientActivationID: "activation",
		MessageID:             "m-000",
		RequestID:             "old-report",
		Fingerprint:           "old-report-fingerprint",
		ReplyID:               "old-reply",
	})
	page, err := store.GetReplies("dash", "", "", 1)
	if err != nil || page.NextCursor == "" {
		t.Fatalf("initial cursor page = %#v, %v", page, err)
	}

	if _, err = store.RecordReply(
		"recipient",
		"activation",
		uuid.NewString(),
		targetID,
		"completed",
		strings.Repeat("r", messageReportLimitBytes),
	); err != nil {
		t.Fatal(err)
	}
	if historyBytes(t, store.data) > messageHistoryLimitBytes {
		t.Fatalf("reply exceeded per-dashboard cap: bytes=%d limit=%d", historyBytes(t, store.data), messageHistoryLimitBytes)
	}
	if _, err = store.GetMessage("m-000"); err == nil {
		t.Fatal("capacity eviction did not remove the oldest whole message")
	}
	if len(store.data.Reports) != 1 || store.data.Reports[0].MessageID != targetID {
		t.Fatalf("evicted report receipt was retained: %#v", store.data.Reports)
	}
	if _, found, findErr := store.FindSend("sender", "old-send"); findErr != nil || !found {
		t.Fatalf("compact send receipt was not preserved: found=%v err=%v", found, findErr)
	}
	resumed, err := store.GetReplies("dash", "", page.NextCursor, 10)
	if err != nil || !resumed.Gap {
		t.Fatalf("cursor after report eviction = %#v, %v; want gap", resumed, err)
	}
}

func TestRecordReplyPersistenceFailureRestoresEvictionAndCursorState(t *testing.T) {
	store, targetID := seedMessageHistoryNearCapacity(t)
	if err := store.saveLocked(); err != nil {
		t.Fatal(err)
	}
	before, err := json.Marshal(store.data)
	if err != nil {
		t.Fatal(err)
	}
	store.path = filepath.Dir(store.path) // Replacing a directory with a file must fail.
	_, reportErr := store.RecordReply(
		"recipient",
		"activation",
		uuid.NewString(),
		targetID,
		"completed",
		strings.Repeat("r", messageReportLimitBytes),
	)
	if reportErr == nil {
		t.Fatal("expected report persistence failure")
	}
	var classified *MessagingError
	if !errors.As(reportErr, &classified) || classified.Code != "storage_unavailable" {
		t.Fatalf("report persistence error = %v, want storage_unavailable", reportErr)
	}
	after, err := json.Marshal(store.data)
	if err != nil {
		t.Fatal(err)
	}
	if string(after) != string(before) {
		t.Fatal("failed report persistence did not restore messages, replies, receipts, sequence, and cursor generation")
	}
}

func TestFailedOperationAdmissionRestoresRetainedHistoryAndGeneration(t *testing.T) {
	store, err := NewMessageStore(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	store.data.DashboardGenerations["dash"] = 17
	for i := range messageHistoryLimitCount {
		id := fmt.Sprintf("old-%05d", i)
		reply := MessageReply{ID: "reply-" + id, MessageID: id, Text: "retained", Kind: "completed", Sequence: uint64(i + 1)}
		store.data.Messages = append(store.data.Messages, MessageRecord{ID: id, DashboardID: "dash", SenderSessionID: "sender",
			RecipientSessionID: "recipient", RecipientActivationID: "activation", Text: "task", Delivery: MessageSubmitted,
			Sequence: uint64(i + 1), Replies: []MessageReply{reply}})
	}
	store.data.Sequence = messageHistoryLimitCount
	store.data.Sends = append(
		store.data.Sends,
		sendReceipt{SenderSessionID: "sender", RequestID: "old-request", Fingerprint: "old-fingerprint", MessageID: "old-00000"},
	)
	if err = store.saveLocked(); err != nil {
		t.Fatal(err)
	}
	before, err := json.Marshal(store.data)
	if err != nil {
		t.Fatal(err)
	}
	_, err = store.CreateIntentWithAdmission(MessageRecord{ID: "unaccepted", DashboardID: "dash", SenderSessionID: "sender",
		RecipientSessionID: "recipient", RecipientActivationID: "activation", Text: "new task"}, "new-request", "new-fingerprint",
		func() error { return errors.New("injected operation journal failure") })
	if err == nil {
		t.Fatal("expected operation admission failure")
	}
	var classified *MessagingError
	if !errors.As(err, &classified) || classified.Code != "journal_unavailable" {
		t.Fatalf("admission error = %v, want journal_unavailable", err)
	}
	after, err := json.Marshal(store.data)
	if err != nil {
		t.Fatal(err)
	}
	if string(after) != string(before) {
		t.Fatal("rejected admission changed accepted content, replies, receipts, sequence, or cursor generation")
	}
	if _, err = store.GetMessage("old-00000"); err != nil {
		t.Fatalf("oldest accepted message was evicted by rejected send: %v", err)
	}
	if _, found, findErr := store.FindSend("sender", "new-request"); findErr != nil || found {
		t.Fatalf("rejected send receipt exists=%v, err=%v", found, findErr)
	}
}

func seedMessageHistoryNearCapacity(t *testing.T) (*MessageStore, string) {
	t.Helper()
	store, err := NewMessageStore(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	store.data.DashboardGenerations["dash"] = 4
	sequence := uint64(0)
	addMessage := func(id string, reports int) MessageRecord {
		message := MessageRecord{ID: id, DashboardID: "dash", SenderSessionID: "sender", RecipientSessionID: "recipient",
			RecipientActivationID: "activation", Text: "task", Delivery: MessageSubmitted, CreatedAt: time.Now().UTC()}
		for i := range reports {
			sequence++
			message.Replies = append(message.Replies, MessageReply{
				ID:                 fmt.Sprintf("r-%s-%03d", id, i),
				MessageID:          id,
				SenderSessionID:    "recipient",
				SenderActivationID: "activation",
				Kind:               "completed",
				Text:               strings.Repeat("x", messageReportLimitBytes),
				Sequence:           sequence,
			})
		}
		message.Sequence = sequence
		return message
	}
	for i := range 31 {
		store.data.Messages = append(store.data.Messages, addMessage(fmt.Sprintf("m-%03d", i), messageReportsPerMessage))
	}
	target := addMessage("m-target", 0)
	baseBytes := historyBytes(t, messageStoreData{Messages: append(slicesClone(store.data.Messages), target)})
	sampleReply := MessageReply{ID: "r-target-sample", MessageID: target.ID, SenderSessionID: "recipient",
		SenderActivationID: "activation", Kind: "completed", Text: strings.Repeat("x", messageReportLimitBytes), Sequence: sequence + 1}
	sampleBytes, err := json.Marshal(sampleReply)
	if err != nil {
		t.Fatal(err)
	}
	reportCount := (messageHistoryLimitBytes - 1024 - baseBytes) / len(sampleBytes)
	reportCount = min(reportCount, messageReportsPerMessage-1)
	for i := range reportCount {
		sequence++
		target.Replies = append(target.Replies, MessageReply{ID: fmt.Sprintf("r-target-%03d", i), MessageID: target.ID,
			SenderSessionID: "recipient", SenderActivationID: "activation", Kind: "completed",
			Text: strings.Repeat("x", messageReportLimitBytes), Sequence: sequence})
	}
	target.Sequence = sequence
	store.data.Messages = append(store.data.Messages, target)
	store.data.Sequence = sequence
	if got := historyBytes(t, store.data); got <= messageHistoryLimitBytes-messageReportLimitBytes {
		t.Fatalf("test fixture is not close enough to capacity: %d", got)
	}
	return store, target.ID
}

func historyBytes(t *testing.T, data messageStoreData) int {
	t.Helper()
	total := 0
	for _, message := range data.Messages {
		encoded, err := json.Marshal(message)
		if err != nil {
			t.Fatal(err)
		}
		total += len(encoded)
	}
	return total
}

func slicesClone[T any](items []T) []T { return append([]T(nil), items...) }
