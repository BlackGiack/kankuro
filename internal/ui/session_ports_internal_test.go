package ui

import (
	"context"
	"io"
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/BlackGiack/kankuro/internal/api/contracts"
	"github.com/BlackGiack/kankuro/internal/ui/messages"
	"github.com/BlackGiack/kankuro/internal/ui/views"
)

type fakeAttachment struct{ closed *bool }

func (*fakeAttachment) Initial() contracts.AttachmentInitial { return contracts.AttachmentInitial{} }
func (*fakeAttachment) Next(context.Context) ([]byte, error) { return nil, io.EOF }
func (*fakeAttachment) Write(p []byte) (int, error)          { return len(p), nil }
func (*fakeAttachment) Resize(int, int) error                { return nil }
func (f *fakeAttachment) Close() error                       { *f.closed = true; return nil }

type fakeSessionPort struct {
	launch      contracts.LaunchSessionRequest
	launchCalls int
	relaunch    contracts.RelaunchSessionRequest
	open        contracts.OpenSessionWorkspaceRequest
}

func (f *fakeSessionPort) LaunchSession(_ context.Context, r contracts.LaunchSessionRequest) contracts.LifecycleResult {
	f.launch = r
	f.launchCalls++
	return contracts.LifecycleResult{Effect: contracts.UsableActivation, SessionID: "s", ProcessActivationID: "p"}
}
func (f *fakeSessionPort) RelaunchDashboardSession(_ context.Context, r contracts.RelaunchSessionRequest) contracts.LifecycleResult {
	f.relaunch = r
	return contracts.LifecycleResult{
		Effect:         contracts.UncertainEffect,
		SessionID:      r.SessionID,
		Diagnostic:     "termination unconfirmed",
		DiagnosticCode: contracts.DiagnosticUncertain,
	}
}
func (*fakeSessionPort) SwitchSession(context.Context, contracts.SwitchSessionRequest) contracts.LifecycleResult {
	return contracts.LifecycleResult{Effect: contracts.NoEffect, Diagnostic: "failed", DiagnosticCode: contracts.DiagnosticFailure}
}

func (*fakeSessionPort) ReadConversationPreparationResult(
	context.Context,
	contracts.ConversationPreparationRequest,
) contracts.ConversationPreparation {
	return contracts.ConversationPreparation{}
}
func (*fakeSessionPort) RenameDashboardSession(context.Context, contracts.RenameSessionRequest) contracts.RenameSessionResult {
	return contracts.RenameSessionResult{}
}
func (*fakeSessionPort) StopDashboardSession(context.Context, contracts.ScopedSessionRequest) contracts.StopSessionResult {
	return contracts.StopSessionResult{}
}
func (*fakeSessionPort) RemoveDashboardSession(context.Context, contracts.ScopedSessionRequest) contracts.RemoveSessionResult {
	return contracts.RemoveSessionResult{}
}

func (f *fakeSessionPort) OpenSessionWorkspace(
	_ context.Context,
	r contracts.OpenSessionWorkspaceRequest,
) contracts.OpenSessionWorkspaceResult {
	f.open = r
	return contracts.OpenSessionWorkspaceResult{Started: true}
}

func TestSessionCommandsForwardContractValues(t *testing.T) {
	port := &fakeSessionPort{}
	actions := sessionActions{port: port}
	launch := actions.launch("d", 7, "c", 9)().(messages.SessionLaunchCompletedMsg)
	if port.launch.DashboardID != "d" || port.launch.CardID != "c" || launch.Request != 9 || launch.Result.ProcessActivationID != "p" {
		t.Fatalf("launch: %+v %+v", port.launch, launch)
	}
	choice := contracts.ConversationChoice{Kind: contracts.ConversationResume, SelectionID: "opaque"}
	activation := actions.activate("d", 7, 10, "s", "c", true, choice)().(messages.SessionActivationCompletedMsg)
	if port.relaunch.CardID != "c" || port.relaunch.Choice != choice || activation.Result.Effect != contracts.UncertainEffect {
		t.Fatalf("activation: %+v %+v", port.relaunch, activation)
	}
	opened := actions.launchWorkspace("d", 7, 11, "s", contracts.WorkspaceTerminal)().(messages.WorkspaceOpenedMsg)
	if port.open.SessionID != "s" || port.open.Kind != contracts.WorkspaceTerminal || !opened.Result.Started {
		t.Fatalf("workspace: %+v %+v", port.open, opened)
	}
}

func TestStaleAttachReadyIsClosed(t *testing.T) {
	c, _ := newLaunchFocusCoordinator(t)
	c.lifecycleRequest = 2
	closed := false
	handle := &fakeAttachment{closed: &closed}
	_, cmd := c.Update(
		messages.ScopedAttachReadyMsg{
			DashboardID: c.activeDashboard,
			Generation:  c.viewGeneration,
			RequestKind: "lifecycle",
			Request:     1,
			SessionID:   "s",
			Handle:      handle,
		},
	)
	if !closed || cmd != nil {
		t.Fatalf("stale handle: closed=%t cmd=%v", closed, cmd != nil)
	}
}

func TestInvalidActivationOutcomeDoesNotAttach(t *testing.T) {
	c, _ := newLaunchFocusCoordinator(t)
	c.lifecycleRequest = 1
	_, cmd := c.Update(
		messages.SessionActivationCompletedMsg{
			DashboardID: c.activeDashboard,
			Generation:  c.viewGeneration,
			Request:     1,
			SessionID:   "s",
			Result:      contracts.LifecycleResult{Effect: contracts.UsableActivation, SessionID: "s"},
		},
	)
	if c.err == nil {
		t.Fatal("invalid activation was accepted")
	}
	if c.currentView != ViewDashboard {
		t.Fatal("invalid outcome navigated")
	}
	_ = cmd // refresh only; no attachment is created from invalid evidence
}

func TestConversationReadCancellationAndOverlap(t *testing.T) {
	c, _ := newLaunchFocusCoordinator(t)
	entered := make(chan struct{})
	port := &blockingConversationPort{fakeSessionPort: &fakeSessionPort{}, entered: entered}
	c.actions.port = port
	cmd := c.prepareConversation("s", "", messages.ViewDashboard)
	completed := make(chan tea.Msg, 1)
	go func() { completed <- cmd() }()
	<-entered
	c.navigateTo(ViewSettings)
	msg := <-completed
	c.Update(msg)
	if c.currentView != ViewSettings {
		t.Fatal("cancelled preparation navigated")
	}

	c.navigateTo(ViewConversationChoice)
	picker := c.views[ViewConversationChoice].(*views.ConversationChoiceView)
	picker.SetContext("s", "c", messages.ViewDashboard, nil)
	port.entered = make(chan struct{})
	first := c.refreshConversationChoiceContext()
	if first == nil {
		t.Fatal("missing refresh")
	}
	go func() { completed <- first() }()
	<-port.entered
	if duplicate := c.refreshConversationChoiceContext(); duplicate != nil {
		t.Fatal("overlapping refresh started")
	}
	c.navigateTo(ViewDashboard)
	c.Update(<-completed)
}

type blockingConversationPort struct {
	*fakeSessionPort

	entered chan struct{}
}

func (p *blockingConversationPort) ReadConversationPreparationResult(
	ctx context.Context,
	request contracts.ConversationPreparationRequest,
) contracts.ConversationPreparation {
	close(p.entered)
	<-ctx.Done()
	return contracts.ConversationPreparation{
		DashboardID:    request.DashboardID,
		SessionID:      request.SessionID,
		Diagnostic:     ctx.Err().Error(),
		DiagnosticCode: contracts.DiagnosticCancelled,
	}
}

func TestStaleMutationResultsDoNotReplaceCurrentFeedback(t *testing.T) {
	c, _ := newLaunchFocusCoordinator(t)
	c.mutationRequest = 2
	c.workspaceRequest = 2
	oldStop := messages.SessionStoppedMsg{
		DashboardID: c.activeDashboard,
		Generation:  c.viewGeneration,
		Request:     1,
		SessionID:   "s",
		Result:      contracts.StopSessionResult{Diagnostic: "old stop", DiagnosticCode: contracts.DiagnosticFailure},
	}
	oldWorkspace := messages.WorkspaceOpenedMsg{
		DashboardID: c.activeDashboard,
		Generation:  c.viewGeneration,
		Request:     1,
		SessionID:   "s",
		Result:      contracts.OpenSessionWorkspaceResult{Diagnostic: "old workspace", DiagnosticCode: contracts.DiagnosticFailure},
	}
	c.Update(oldStop)
	c.Update(oldWorkspace)
	if c.err != nil {
		t.Fatalf("stale result changed feedback: %v", c.err)
	}
}

func TestMismatchedConversationReadCannotStartActivation(t *testing.T) {
	c, _ := newLaunchFocusCoordinator(t)
	c.conversationEpoch, c.conversationReadRequest = 1, 1
	msg := messages.ConversationPreparedMsg{
		DashboardID: c.activeDashboard,
		Generation:  c.viewGeneration,
		Epoch:       1,
		Request:     1,
		Initial:     true,
		SessionID:   "expected",
		ReturnView:  messages.ViewDashboard,
		Preparation: contracts.ConversationPreparation{
			DashboardID:  c.activeDashboard,
			SessionID:    "different",
			CardID:       "card",
			Availability: contracts.Availability{Enabled: true},
		},
	}
	_, cmd := c.Update(msg)
	if cmd != nil || c.err == nil || c.lifecycleExecuting {
		t.Fatalf("mismatched read accepted: command=%v error=%v", cmd != nil, c.err)
	}
}

func TestUncertainActivationInvalidatesOlderInventory(t *testing.T) {
	c, _ := newLaunchFocusCoordinator(t)
	c.lifecycleRequest = 1
	before := c.sessionEpoch
	result := contracts.LifecycleResult{
		Effect:         contracts.UncertainEffect,
		SessionID:      "s",
		Diagnostic:     "termination unknown",
		DiagnosticCode: contracts.DiagnosticUncertain,
	}
	_, cmd := c.Update(
		messages.SessionActivationCompletedMsg{
			DashboardID: c.activeDashboard,
			Generation:  c.viewGeneration,
			Request:     1,
			SessionID:   "s",
			Result:      result,
		},
	)
	if c.sessionEpoch != before+1 || c.err == nil || cmd == nil {
		t.Fatalf("uncertain effect not refreshed: epoch=%d error=%v", c.sessionEpoch, c.err)
	}
}
