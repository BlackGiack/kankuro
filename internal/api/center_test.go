//nolint:testpackage // Tests exercise internal storage and inspection boundaries.
package api

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/BlackGiack/kankuro/internal/api/contracts"
	"github.com/BlackGiack/kankuro/internal/config"
	"github.com/BlackGiack/kankuro/internal/session"
	"github.com/google/uuid"
)

func TestCenterReportsSubscriptionAndScope(t *testing.T) {
	h := newMessagingHarness(t)
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	scope := contracts.ScopedSessionRequest{DashboardID: DefaultDashboardID, SessionID: h.recipientID}
	sub, err := h.service.SubscribeSessionReports(ctx, scope)
	if err != nil {
		t.Fatal(err)
	}
	defer sub.Release()
	req := contracts.SessionReportsRequest{ScopedSessionRequest: scope, Limit: 10}
	initial, err := h.service.ReadSessionReports(ctx, req)
	if err != nil || len(initial.Reports) != 0 {
		t.Fatalf("initial = %#v, %v", initial, err)
	}
	store := h.manager.Messages()
	message := session.MessageRecord{
		ID:                    uuid.NewString(),
		DashboardID:           DefaultDashboardID,
		SenderSessionID:       h.callerID,
		SenderActivationID:    "sender",
		RecipientSessionID:    h.recipientID,
		RecipientActivationID: "worker",
		Text:                  "task",
	}
	if _, err = store.CreateIntent(message, uuid.NewString(), session.SendFingerprint(h.recipientID, "worker", "task")); err != nil {
		t.Fatal(err)
	}
	if err = store.UpdateDelivery(message.ID, session.MessageSubmitted, ""); err != nil {
		t.Fatal(err)
	}
	select {
	case <-sub.Changes():
		t.Fatal("prompt and delivery updates woke the report subscription")
	case <-time.After(40 * time.Millisecond):
	}
	if _, err = store.RecordReply(h.recipientID, "worker", uuid.NewString(), message.ID, "progress", "working"); err != nil {
		t.Fatal(err)
	}
	select {
	case <-sub.Changes():
	case <-time.After(5 * time.Second):
		t.Fatal("missing change")
	}
	got, err := h.service.ReadSessionReports(ctx, req)
	if err != nil || len(got.Reports) != 1 || got.Reports[0].Text != "working" {
		t.Fatalf("reports = %#v, %v", got, err)
	}
	req.DashboardID = "foreign"
	if _, err = h.service.ReadSessionReports(ctx, req); err == nil {
		t.Fatal("foreign dashboard accepted")
	}
	sub.Release()
	sub.Release()
}

func TestCenterReportsChangeBeforeInitialReadAndStop(t *testing.T) {
	h := newMessagingHarness(t)
	scope := contracts.ScopedSessionRequest{DashboardID: DefaultDashboardID, SessionID: h.recipientID}
	store := h.manager.Messages()
	message := session.MessageRecord{
		ID:                    uuid.NewString(),
		DashboardID:           DefaultDashboardID,
		SenderSessionID:       h.callerID,
		SenderActivationID:    "sender",
		RecipientSessionID:    h.recipientID,
		RecipientActivationID: "worker",
		Text:                  "task",
	}
	if _, err := store.CreateIntent(message, uuid.NewString(), session.SendFingerprint(h.recipientID, "worker", "task")); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	sub, err := h.service.SubscribeSessionReports(ctx, scope)
	if err != nil {
		t.Fatal(err)
	}
	defer sub.Release()
	if _, err = store.RecordReply(h.recipientID, "worker", uuid.NewString(), message.ID, "completed", "done"); err != nil {
		t.Fatal(err)
	}
	got, err := h.service.ReadSessionReports(ctx, contracts.SessionReportsRequest{ScopedSessionRequest: scope, Limit: 1})
	if err != nil || len(got.Reports) != 1 || got.Reports[0].Text != "done" {
		t.Fatalf("initial read %+v: %v", got, err)
	}
	select {
	case <-sub.Changes():
	case <-time.After(5 * time.Second):
		t.Fatal("first report signal lost")
	}
	if result := h.service.StopDashboardSession(t.Context(), scope); result.Diagnostic != "" || !result.ProcessTerminated {
		t.Fatalf("stop: %+v", result)
	}
	got, err = h.service.ReadSessionReports(ctx, contracts.SessionReportsRequest{ScopedSessionRequest: scope, Limit: 1})
	if err != nil || len(got.Reports) != 1 {
		t.Fatalf("stopped session reports %+v: %v", got, err)
	}
	cancel()
	select {
	case _, ok := <-sub.Changes():
		if ok {
			select {
			case <-sub.Changes():
			case <-time.After(5 * time.Second):
				t.Fatal("cancel did not close")
			}
		}
	case <-time.After(5 * time.Second):
		t.Fatal("cancel did not close")
	}
}

func TestCenterWorkspaceUsesSavedSessionDirectory(t *testing.T) {
	h := newMessagingHarness(t)
	scope := contracts.ScopedSessionRequest{DashboardID: DefaultDashboardID, SessionID: h.recipientID}
	got, err := h.service.ReadSessionWorkspace(context.Background(), scope)
	if err != nil || got.WorkingDir != h.workspace {
		t.Fatalf("saved workspace = %+v, %v; want %q", got, err, h.workspace)
	}
	scope.DashboardID = "foreign"
	if _, err = h.service.ReadSessionWorkspace(context.Background(), scope); !errors.Is(err, contracts.ErrSessionScope) {
		t.Fatalf("foreign scope = %v", err)
	}
}

func TestCenterReportChangeDuringRearm(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	first := make(chan struct{})
	second := make(chan struct{})
	stable := make(chan struct{})
	entered := make(chan struct{})
	resume := make(chan struct{})
	rearmed := make(chan struct{})
	calls := 0
	changed := func() <-chan struct{} {
		calls++
		if calls == 1 {
			close(entered)
			<-resume
			return second
		}
		close(rearmed)
		return stable
	}
	sub := &reportSubscription{changes: make(chan struct{}, 1), cancel: cancel}
	go watchReportChanges(ctx, cancel, make(chan struct{}), first, nil, changed, nil, func() error { return nil }, sub)
	close(first)
	<-entered
	close(second)
	close(resume)
	<-rearmed
	select {
	case <-sub.Changes():
	case <-time.After(5 * time.Second):
		t.Fatal("rearm race lost change")
	}
	cancel()
	deadline := time.After(5 * time.Second)
	for {
		select {
		case _, ok := <-sub.Changes():
			if !ok {
				return
			}
		case <-deadline:
			t.Fatal("subscription did not close")
		}
	}
}

func TestCenterWorkspaceUsesTemporaryCopy(t *testing.T) {
	service, repo, manager, workspace, workspaceID := newServiceFixture(t)
	card := saveCard(
		t,
		repo.FileRepository,
		config.IDCard{ID: "center-tmp", Name: "center temporary", Driver: "cat", WorkspaceID: workspaceID, UseTmpWorkspace: true},
	)
	sessionID, err := launchCardTest(service, card)
	if err != nil {
		t.Fatal(err)
	}
	snapshot, ok := manager.GetSession(sessionID)
	if !ok || snapshot.WorkingDir == workspace {
		t.Fatalf("temporary workspace not created: %+v", snapshot)
	}
	got, err := service.ReadSessionWorkspace(
		t.Context(),
		contracts.ScopedSessionRequest{DashboardID: DefaultDashboardID, SessionID: sessionID},
	)
	if err != nil || got.WorkingDir != snapshot.WorkingDir {
		t.Fatalf("temporary workspace read=%+v err=%v", got, err)
	}
}

func TestCenterReportsInventoryFailureIsNotEmptyHistory(t *testing.T) {
	h := newMessagingHarness(t)
	if err := os.WriteFile(filepath.Join(h.repo.Dir(), config.IDCardsFilename), []byte("not: valid: yaml: :::"), 0600); err != nil {
		t.Fatal(err)
	}
	req := contracts.SessionReportsRequest{Limit: 1}
	req.DashboardID, req.SessionID = DefaultDashboardID, h.recipientID
	_, err := h.service.ReadSessionReports(t.Context(), req)
	if err == nil {
		t.Fatal("failed membership inventory became an empty success")
	}
}

func TestCenterReportScopeLossWithoutReportChanges(t *testing.T) {
	h := newMessagingHarness(t)
	scope := contracts.ScopedSessionRequest{DashboardID: DefaultDashboardID, SessionID: h.recipientID}
	sub, err := h.service.SubscribeSessionReports(t.Context(), scope)
	if err != nil {
		t.Fatal(err)
	}
	defer sub.Release()
	if result := h.service.StopDashboardSession(t.Context(), scope); result.Diagnostic != "" || !result.ProcessTerminated {
		t.Fatalf("stop: %+v", result)
	}
	if result := h.service.RemoveDashboardSession(t.Context(), scope); result.Diagnostic != "" || !result.InMemoryRemoved {
		t.Fatalf("remove: %+v", result)
	}
	select {
	case _, open := <-sub.Changes():
		if open || !errors.Is(sub.Err(), contracts.ErrSessionScope) {
			t.Fatalf("expected scope closure, open=%t err=%v", open, sub.Err())
		}
	case <-time.After(5 * time.Second):
		t.Fatal("scope loss was not detected")
	}
}
