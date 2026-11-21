package ui

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/BlackGiack/kankuro/internal/api"
	"github.com/BlackGiack/kankuro/internal/api/contracts"
	"github.com/BlackGiack/kankuro/internal/config"
	"github.com/BlackGiack/kankuro/internal/configstartup"
	"github.com/BlackGiack/kankuro/internal/session"
	"github.com/BlackGiack/kankuro/internal/ui/messages"
	"github.com/BlackGiack/kankuro/internal/ui/views"
)

type focusConfiguration struct{ configurationPort }

func (focusConfiguration) DefaultSettings() contracts.Settings {
	return contracts.Settings{ShowPersonasInCardEditor: true, ShowMCPsInCardEditor: true}
}
func (focusConfiguration) SaveSettings(contracts.Settings) error { return nil }
func (focusConfiguration) SaveMCP(value contracts.MCPEditor) (contracts.MCPEditor, error) {
	return value, nil
}

// Focus-only tests inject inert capabilities and supply scoped results directly.
func newLaunchFocusCoordinator(t *testing.T, sessions ...contracts.SessionSummary) (*Coordinator, *views.Dashboard) {
	t.Helper()
	coordinator := NewCoordinator(Capabilities{
		Configuration: focusConfiguration{}, Center: centerTestPort{}, Actions: &fakeSessionPort{},
	}, nil)
	coordinator.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	coordinator.Update(messages.DashboardOpenRequestedMsg{
		DashboardID: "default", DashboardName: "Default", DashboardState: contracts.DashboardActive,
	})
	dashboard := coordinator.views[ViewDashboard].(*views.Dashboard)
	coordinator.launchCardID = "test-card"
	coordinator.views[ViewSessionLauncher].(*views.SessionLauncher).SetCard(contracts.CardSummary{ID: "test-card"})
	setFocusSummaries(dashboard, sessions...)
	return coordinator, dashboard
}

func setFocusSummaries(dashboard *views.Dashboard, sessions ...contracts.SessionSummary) {
	dashboard.SetInventory(
		contracts.ReadResult[contracts.DashboardInventory]{Attempted: true, Value: contracts.DashboardInventory{Sessions: sessions}},
		contracts.DisplayChoicesResult{},
	)
}

// Real repository/service fixture for tests that verify persistence and refresh.
func newLaunchFocusCoordinatorWithRepo(
	t *testing.T,
) (*Coordinator, *views.Dashboard, *config.FileRepository, *session.Manager) {
	t.Helper()
	root := t.TempDir()
	if err := configstartup.PrepareStartup(root); err != nil {
		t.Fatal(err)
	}
	repo, err := config.NewFileRepository(root)
	if err != nil {
		t.Fatal(err)
	}
	manager := newTestManager(t, repo.Dir(), repo)
	coordinator := NewCoordinator(testCapabilities(api.New(repo, manager)), manager.LoadErrors())
	coordinator.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	coordinator.Update(messages.DashboardOpenRequestedMsg{
		DashboardID: "default", DashboardName: "Default", DashboardState: contracts.DashboardActive,
	})
	dashboard := coordinator.views[ViewDashboard].(*views.Dashboard)
	coordinator.launchCardID = "test-card"
	coordinator.views[ViewSessionLauncher].(*views.SessionLauncher).SetCard(contracts.CardSummary{ID: "test-card"})
	setFocusedSessionSnapshots(dashboard, nil)
	return coordinator, dashboard, repo, manager
}

func launchCompleted(coordinator *Coordinator, sessionID string, err error) messages.SessionLaunchCompletedMsg {
	result := contracts.LifecycleResult{Effect: contracts.UsableActivation, SessionID: sessionID, ProcessActivationID: "process"}
	if err != nil && !errors.Is(err, session.ErrPersistFailed) {
		result = contracts.LifecycleResult{Effect: contracts.NoEffect, Diagnostic: err.Error(), DiagnosticCode: contracts.DiagnosticFailure}
	} else if err != nil {
		result.Diagnostic = err.Error()
		result.DiagnosticCode = contracts.DiagnosticPersistence
	}
	return messages.SessionLaunchCompletedMsg{
		CardID:      "test-card",
		Result:      result,
		DashboardID: coordinator.activeDashboard,
		Generation:  coordinator.viewGeneration,
	}
}

func TestLaunchFocusSnapshotBeforeCompletion(t *testing.T) {
	coordinator, dashboard := newLaunchFocusCoordinator(t,
		contracts.SessionSummary{ID: "existing", Name: "Existing"},
		contracts.SessionSummary{ID: "new", Name: "New"},
	)
	coordinator.Update(launchCompleted(coordinator, "new", nil))
	assertFocusedSession(t, dashboard, "new")
}

func TestLaunchFocusRejectsRefreshStartedBeforeLaunchAndName(t *testing.T) {
	oldSession := contracts.SessionSummary{ID: "old", Name: "Agent", StartedAt: time.Unix(1, 0)}
	newSession := contracts.SessionSummary{ID: "new", Name: "Agent", StartedAt: time.Unix(2, 0)}
	coordinator, dashboard := newLaunchFocusCoordinator(t, oldSession)
	preLaunchEpoch := coordinator.sessionEpoch
	coordinator.Update(messages.ScopedNavigateMsg{
		View: messages.ViewSessionLauncher, DashboardID: coordinator.activeDashboard, Generation: coordinator.viewGeneration,
	})
	coordinator.Update(launchCompleted(coordinator, newSession.ID, nil))
	setFocusSummaries(dashboard, oldSession, newSession)
	assertFocusedSession(t, dashboard, newSession.ID)

	coordinator.Update(messages.ScopedSessionNamedMsg{
		DashboardID: coordinator.activeDashboard, Generation: coordinator.viewGeneration,
		SessionID: newSession.ID, Name: "My task", Result: contracts.RenameSessionResult{Applied: true, Persisted: true},
	})
	stale := messages.SessionsRefreshedMsg{
		DashboardID: coordinator.activeDashboard, Generation: coordinator.viewGeneration,
		SessionEpoch: preLaunchEpoch,
		Inventory: contracts.ReadResult[contracts.DashboardInventory]{
			Attempted: true, Value: contracts.DashboardInventory{Sessions: []contracts.SessionSummary{oldSession}},
		},
	}
	coordinator.Update(stale)
	assertFocusedSession(t, dashboard, newSession.ID)

	newSession.Name = "My task"
	current := messages.SessionsRefreshedMsg{
		DashboardID: coordinator.activeDashboard, Generation: coordinator.viewGeneration,
		SessionEpoch: coordinator.sessionEpoch,
		Inventory: contracts.ReadResult[contracts.DashboardInventory]{
			Attempted: true, Value: contracts.DashboardInventory{Sessions: []contracts.SessionSummary{oldSession, newSession}},
		},
	}
	coordinator.Update(current)
	assertFocusedSession(t, dashboard, newSession.ID)
}

func TestLaunchFocusFailureAndEmptyIDDoNotMoveFocus(t *testing.T) {
	for _, test := range []struct {
		name      string
		sessionID string
		err       error
	}{
		{name: "fatal error", sessionID: "new", err: errors.New("launch failed")},
		{name: "empty ID"},
	} {
		t.Run(test.name, func(t *testing.T) {
			coordinator, dashboard := newLaunchFocusCoordinator(t, contracts.SessionSummary{ID: "existing", Name: "Existing"})
			coordinator.Update(launchCompleted(coordinator, test.sessionID, test.err))
			selected, ok := dashboard.SessionList().Selected()
			if !ok || selected.ID != "existing" || dashboard.FocusedPane() != views.PaneCards {
				t.Fatalf("failed launch moved focus: selected=%+v ok=%v pane=%v", selected, ok, dashboard.FocusedPane())
			}
		})
	}
}

func TestLaunchFocusPersistenceWarningStillSelectsSession(t *testing.T) {
	coordinator, dashboard := newLaunchFocusCoordinator(t,
		contracts.SessionSummary{ID: "existing", Name: "Existing"},
		contracts.SessionSummary{ID: "new", Name: "New"},
	)
	warning := errors.Join(session.ErrPersistFailed, errors.New("disk full"))
	coordinator.Update(launchCompleted(coordinator, "new", warning))
	assertFocusedSession(t, dashboard, "new")
	launcher := coordinator.views[ViewSessionLauncher].(*views.SessionLauncher)
	if launcher.LaunchErr() == nil || launcher.LaunchErr().Error() != warning.Error() {
		t.Fatalf("launcher did not retain persistence warning: %v", launcher.LaunchErr())
	}
}

func TestLaunchFocusIgnoresStaleResults(t *testing.T) {
	for _, test := range []struct {
		name        string
		dashboardID string
		generation  func(*Coordinator) uint64
	}{
		{name: "dashboard", dashboardID: "other", generation: func(c *Coordinator) uint64 { return c.viewGeneration }},
		{name: "generation", dashboardID: "default", generation: func(c *Coordinator) uint64 { return c.viewGeneration - 1 }},
	} {
		t.Run(test.name, func(t *testing.T) {
			coordinator, dashboard := newLaunchFocusCoordinator(t, contracts.SessionSummary{ID: "existing", Name: "Existing"})
			coordinator.Update(messages.SessionLaunchCompletedMsg{
				Result: contracts.LifecycleResult{
					Effect:              contracts.UsableActivation,
					SessionID:           "new",
					ProcessActivationID: "process",
				},
				DashboardID: test.dashboardID,
				Generation:  test.generation(coordinator),
			})
			selected, ok := dashboard.SessionList().Selected()
			if !ok || selected.ID != "existing" || dashboard.FocusedPane() != views.PaneCards {
				t.Fatalf("stale launch moved focus: selected=%+v ok=%v pane=%v", selected, ok, dashboard.FocusedPane())
			}
		})
	}
}

func TestLaunchFocusPendingRequestIsCancelledOnDashboardExit(t *testing.T) {
	coordinator, dashboard := newLaunchFocusCoordinator(t, contracts.SessionSummary{ID: "existing", Name: "Existing"})
	staleLaunch := launchCompleted(coordinator, "new", nil)
	coordinator.Update(staleLaunch)
	coordinator.Update(messages.DashboardListRequestedMsg{
		DashboardID: coordinator.activeDashboard, Generation: coordinator.viewGeneration,
	})
	coordinator.Update(messages.DashboardOpenRequestedMsg{
		DashboardID: "default", DashboardName: "Default", DashboardState: contracts.DashboardActive,
	})
	coordinator.Update(staleLaunch)
	setFocusSummaries(dashboard, []contracts.SessionSummary{
		{ID: "existing", Name: "Existing"},
		{ID: "new", Name: "New"},
	}...)
	selected, ok := dashboard.SessionList().Selected()
	if !ok || selected.ID != "existing" || dashboard.FocusedPane() != views.PaneCards {
		t.Fatalf("pending focus crossed dashboard exit: selected=%+v ok=%v pane=%v", selected, ok, dashboard.FocusedPane())
	}
}

func setFocusedSessionSnapshots(dashboard *views.Dashboard, snapshots []session.Snapshot) {
	dashboard.SetInventory(
		contracts.ReadResult[contracts.DashboardInventory]{Attempted: true, Value: sessionInventory(snapshots...)},
		contracts.DisplayChoicesResult{},
	)
}

func sessionInventory(snapshots ...session.Snapshot) contracts.DashboardInventory {
	summaries := make([]contracts.SessionSummary, 0, len(snapshots))
	for _, snapshot := range snapshots {
		summaries = append(summaries, contracts.SessionSummary{
			ID: snapshot.ID, DashboardID: snapshot.DashboardID, CardID: snapshot.CardID, Name: snapshot.Name,
			Status: string(snapshot.Status), StartedAt: snapshot.StartedAt,
		})
	}
	return contracts.DashboardInventory{Sessions: summaries}
}

func assertFocusedSession(t *testing.T, dashboard *views.Dashboard, sessionID string) {
	t.Helper()
	selected, ok := dashboard.SessionList().Selected()
	if !ok || selected.ID != sessionID || dashboard.FocusedPane() != views.PaneSessions || dashboard.CenterSessionID() != sessionID {
		t.Fatalf("session focus mismatch: selected=%+v ok=%v pane=%v preview=%q", selected, ok,
			dashboard.FocusedPane(), dashboard.CenterSessionID())
	}
}

func TestLaunchTransitionCannotNavigateAfterLeavingOrNewerLaunch(t *testing.T) {
	coordinator, _ := newLaunchFocusCoordinator(t)
	coordinator.actions.port = &fakeSessionPort{}
	coordinator.launchRequest, coordinator.launchCardID = 1, "old-card"
	oldCard := contracts.CardSummary{ID: "old-card", Name: "Old"}
	_, oldCmd := coordinator.Update(messages.DashboardCardLoadedMsg{
		DashboardID: coordinator.activeDashboard, Generation: coordinator.viewGeneration,
		Request: 1, CardID: oldCard.ID, Card: oldCard, Action: messages.DashboardCardLaunch,
	})
	if coordinator.currentView != ViewSessionLauncher || oldCmd == nil {
		t.Fatal("checked card load did not enter launcher")
	}
	oldMessages := collectMsgs(oldCmd)
	if len(oldMessages) != 1 {
		t.Fatalf("old launch messages: %+v", oldMessages)
	}
	if _, ok := oldMessages[0].(messages.SessionLaunchCompletedMsg); !ok {
		t.Fatalf("launch queued a navigation: %T", oldMessages[0])
	}
	coordinator.Update(
		messages.ScopedNavigateMsg{
			View:        messages.ViewSettings,
			DashboardID: coordinator.activeDashboard,
			Generation:  coordinator.viewGeneration,
		},
	)
	coordinator.Update(oldMessages[0])
	if coordinator.currentView != ViewSettings {
		t.Fatalf("late launch changed settings view to %v", coordinator.currentView)
	}

	coordinator.Update(
		messages.ScopedNavigateMsg{
			View:        messages.ViewDashboard,
			DashboardID: coordinator.activeDashboard,
			Generation:  coordinator.viewGeneration,
		},
	)
	coordinator.launchRequest++
	coordinator.launchCardID = "new-card"
	newCard := contracts.CardSummary{ID: "new-card", Name: "New"}
	_, newCmd := coordinator.Update(messages.DashboardCardLoadedMsg{
		DashboardID: coordinator.activeDashboard, Generation: coordinator.viewGeneration,
		Request: coordinator.launchRequest, CardID: newCard.ID, Card: newCard, Action: messages.DashboardCardLaunch,
	})
	if coordinator.currentView != ViewSessionLauncher || newCmd == nil {
		t.Fatal("newer launch did not enter launcher")
	}
	coordinator.Update(oldMessages[0])
	launcher := coordinator.views[ViewSessionLauncher].(*views.SessionLauncher)
	if coordinator.currentView != ViewSessionLauncher || launcher.Card().ID != newCard.ID || launcher.LaunchDone() {
		t.Fatal("old launch changed the newer launcher")
	}
	newMessages := collectMsgs(newCmd)
	if len(newMessages) != 1 {
		t.Fatalf("new launch messages: %+v", newMessages)
	}
	coordinator.Update(newMessages[0])
	if !launcher.LaunchDone() {
		t.Fatal("new launch completion was not accepted")
	}
}

func startWorkspaceLaunchDraft(t *testing.T, c *Coordinator, cardID string) {
	t.Helper()
	c.launchRequest++
	c.launchCardID = cardID
	card := contracts.CardSummary{ID: cardID, Name: cardID}
	preparation := contracts.LaunchPreparation{Card: card, WorkspaceID: "main"}
	c.Update(messages.DashboardCardLoadedMsg{
		DashboardID: c.activeDashboard, Generation: c.viewGeneration, Request: c.launchRequest,
		CardID: cardID, Card: card, Preparation: &preparation, Action: messages.DashboardCardLaunch,
	})
	if c.currentView != ViewSessionLauncher {
		t.Fatal("workspace draft did not open")
	}
}

func TestWorkspaceDraftCancelBlocksQueuedEnter(t *testing.T) {
	c, _ := newLaunchFocusCoordinator(t)
	port := &fakeSessionPort{}
	c.actions.port = port
	startWorkspaceLaunchDraft(t, c, "card")
	_, cancel := c.Update(tea.KeyPressMsg{Code: tea.KeyEsc})
	_, launch := c.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if launch != nil {
		for _, msg := range collectMsgs(launch) {
			c.Update(msg)
		}
	}
	if port.launchCalls != 0 {
		t.Fatal("Enter launched after cancellation")
	}
	if cancel == nil {
		t.Fatal("cancellation did not produce navigation")
	}
	c.Update(cancel())
	if c.currentView != ViewDashboard {
		t.Fatalf("cancel returned to %v", c.currentView)
	}
}

func TestWorkspaceSourcesNavigationIsScopedToDraft(t *testing.T) {
	c, _, _, _ := newLaunchFocusCoordinatorWithRepo(t)
	startWorkspaceLaunchDraft(t, c, "first")
	_, staleSource := c.Update(tea.KeyPressMsg{Code: 'e', Mod: tea.ModCtrl})
	_, cancel := c.Update(tea.KeyPressMsg{Code: tea.KeyEsc})
	c.Update(cancel())
	c.Update(staleSource())
	if c.currentView != ViewDashboard {
		t.Fatalf("stale source navigation opened %v", c.currentView)
	}
	startWorkspaceLaunchDraft(t, c, "second")
	c.Update(staleSource())
	if c.currentView != ViewSessionLauncher {
		t.Fatal("stale source navigation interrupted later draft")
	}
	_, validSource := c.Update(tea.KeyPressMsg{Code: 'e', Mod: tea.ModCtrl})
	c.Update(validSource())
	if c.currentView != ViewSourceManagerWorkspaces {
		t.Fatalf("source navigation opened %v", c.currentView)
	}
	c.Update(messages.NavigateBackMsg{})
	if c.currentView != ViewSessionLauncher {
		t.Fatalf("source return opened %v", c.currentView)
	}
	launcher := c.views[ViewSessionLauncher].(*views.SessionLauncher)
	if !launcher.WorkspaceDraftActive() || launcher.Card().ID != "second" {
		t.Fatal("source visit lost the draft")
	}
	_, cancel = c.Update(tea.KeyPressMsg{Code: tea.KeyEsc})
	c.Update(cancel())
	if c.currentView != ViewDashboard {
		t.Fatalf("cancel after source return opened %v", c.currentView)
	}
}

func TestWorkspaceSourcesNavigationRejectedAfterSubmission(t *testing.T) {
	c, _ := newLaunchFocusCoordinator(t)
	startWorkspaceLaunchDraft(t, c, "card")
	_, source := c.Update(tea.KeyPressMsg{Code: 'e', Mod: tea.ModCtrl})
	c.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	c.Update(source())
	if c.currentView != ViewSessionLauncher {
		t.Fatal("stale source navigation interrupted launch")
	}
}

func TestWorkspaceDraftCtrlCQuits(t *testing.T) {
	coordinator, _ := newLaunchFocusCoordinator(t)
	startWorkspaceLaunchDraft(t, coordinator, "card")
	_, cmd := coordinator.Update(tea.KeyPressMsg{Code: 'c', Mod: tea.ModCtrl})
	messages := collectMsgs(cmd)
	if len(messages) != 1 {
		t.Fatalf("Ctrl+C commands: %v", messages)
	}
	if _, ok := messages[0].(tea.QuitMsg); !ok {
		t.Fatalf("Ctrl+C returned %T, want tea.QuitMsg", messages[0])
	}
}

func TestLaunchUsesDisplayedWorkspaceAfterCardEdit(t *testing.T) {
	coordinator, _, repo, manager := newLaunchFocusCoordinatorWithRepo(t)
	mainDir, changedDir := t.TempDir(), t.TempDir()
	if err := repo.SaveWorkspaceSources([]config.WorkspaceSource{{Path: mainDir}, {Path: changedDir}}); err != nil {
		t.Fatal(err)
	}
	workspaces, err := repo.ListWorkspaces()
	if err != nil {
		t.Fatal(err)
	}
	var mainID, changedID string
	for _, ws := range workspaces {
		if ws.Path == mainDir {
			mainID = ws.ID
		}
		if ws.Path == changedDir {
			changedID = ws.ID
		}
	}
	card, err := repo.CreateIDCard(config.IDCard{Name: "launch", Driver: "cat", WorkspaceID: mainID, DashboardID: "default"})
	if err != nil {
		t.Fatal(err)
	}
	load := coordinator.dashboardCallbacks().OnLaunchCard(card.ID)
	coordinator.Update(load())
	launcher := coordinator.views[ViewSessionLauncher].(*views.SessionLauncher)
	if !strings.Contains(launcher.View(), filepath.Base(mainDir)) {
		t.Fatalf("main workspace absent from form: %s", launcher.View())
	}
	card.WorkspaceID = changedID
	if _, err = repo.UpdateIDCard(card); err != nil {
		t.Fatal(err)
	}
	_, launch := coordinator.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	result, ok := launch().(messages.SessionLaunchCompletedMsg)
	if !ok || result.Result.Effect != contracts.UsableActivation {
		t.Fatalf("launch result: %+v", result)
	}
	snapshot, ok := manager.GetSession(result.Result.SessionID)
	if !ok || snapshot.WorkingDir != mainDir {
		t.Fatalf("launched workspace: %+v", snapshot)
	}
}

func TestLaunchPreparationWorkspaceErrorKeepsCachedChoices(t *testing.T) {
	coordinator, _ := newLaunchFocusCoordinator(t)
	launcher := coordinator.views[ViewSessionLauncher].(*views.SessionLauncher)
	launcher.SetWorkspaceChoices(contracts.CardEditorChoices{Workspaces: contracts.ReadResult[[]contracts.WorkspaceOption]{
		Attempted: true, Value: []contracts.WorkspaceOption{{ID: "main", Name: "Cached workspace", Path: "/cached"}},
	}})
	coordinator.launchRequest, coordinator.launchCardID = 1, "card"
	card := contracts.CardSummary{ID: "card", Name: "Card"}
	failure := errors.New("workspace registry unavailable")
	preparation := contracts.LaunchPreparation{
		Card:        card,
		WorkspaceID: "main",
		Workspaces:  contracts.ReadResult[[]contracts.WorkspaceOption]{Attempted: true, Err: failure},
	}
	stale := messages.DashboardCardLoadedMsg{DashboardID: coordinator.activeDashboard, Generation: coordinator.viewGeneration,
		Request: 0, CardID: card.ID, Card: card, Action: messages.DashboardCardLaunch, Preparation: &preparation}
	coordinator.Update(stale)
	if coordinator.err != nil || coordinator.currentView != ViewDashboard {
		t.Fatal("stale registry failure changed the current view")
	}
	stale.Request = 1
	coordinator.Update(stale)
	if !errors.Is(coordinator.err, failure) || !strings.Contains(coordinator.View(), failure.Error()) {
		t.Fatalf("missing registry error: %v", coordinator.err)
	}
	if !strings.Contains(launcher.View(), "Cached workspace") {
		t.Fatalf("cached workspace missing from draft: %s", launcher.View())
	}
	coordinator.Update(tea.KeyPressMsg{Code: tea.KeySpace})
	if !strings.Contains(launcher.View(), "Cached workspace") {
		t.Fatalf("cached workspace missing from picker: %s", launcher.View())
	}
}
