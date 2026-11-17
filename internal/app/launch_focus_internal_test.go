package app

import (
	"errors"
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

func newLaunchFocusCoordinator(t *testing.T, sessions ...session.Snapshot) (*Coordinator, *views.Dashboard) {
	t.Helper()
	coordinator, dashboard, _ := newLaunchFocusCoordinatorWithRepo(t, sessions...)
	return coordinator, dashboard
}

func newLaunchFocusCoordinatorWithRepo(
	t *testing.T,
	sessions ...session.Snapshot,
) (*Coordinator, *views.Dashboard, *config.FileRepository) {
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
	coordinator := NewCoordinator(manager, api.New(repo, manager))
	coordinator.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	coordinator.Update(messages.DashboardOpenRequestedMsg{
		DashboardID: "default", DashboardName: "Default", DashboardState: contracts.DashboardActive,
	})
	dashboard := coordinator.views[ViewDashboard].(*views.Dashboard)
	coordinator.launchCardID = "test-card"
	coordinator.views[ViewSessionLauncher].(*views.SessionLauncher).SetCard(contracts.CardSummary{ID: "test-card"})
	setFocusedSessionSnapshots(dashboard, sessions)
	return coordinator, dashboard, repo
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
		session.Snapshot{ID: "existing", Name: "Existing"},
		session.Snapshot{ID: "new", Name: "New"},
	)
	coordinator.Update(launchCompleted(coordinator, "new", nil))
	assertFocusedSession(t, dashboard, "new")
}

func TestLaunchFocusRejectsRefreshStartedBeforeLaunchAndName(t *testing.T) {
	oldSession := session.Snapshot{ID: "old", Name: "Agent", StartedAt: time.Unix(1, 0)}
	newSession := session.Snapshot{ID: "new", Name: "Agent", StartedAt: time.Unix(2, 0)}
	coordinator, dashboard := newLaunchFocusCoordinator(t, oldSession)
	preLaunchEpoch := coordinator.sessionEpoch
	coordinator.Update(messages.ScopedNavigateMsg{
		View: messages.ViewSessionLauncher, DashboardID: coordinator.activeDashboard, Generation: coordinator.viewGeneration,
	})
	coordinator.Update(launchCompleted(coordinator, newSession.ID, nil))
	setFocusedSessionSnapshots(dashboard, []session.Snapshot{oldSession, newSession})
	assertFocusedSession(t, dashboard, newSession.ID)

	coordinator.Update(messages.ScopedSessionNamedMsg{
		DashboardID: coordinator.activeDashboard, Generation: coordinator.viewGeneration,
		SessionID: newSession.ID, Name: "My task", Result: contracts.RenameSessionResult{Applied: true, Persisted: true},
	})
	stale := messages.SessionsRefreshedMsg{
		DashboardID: coordinator.activeDashboard, Generation: coordinator.viewGeneration,
		SessionEpoch: preLaunchEpoch,
		Inventory:    contracts.ReadResult[contracts.DashboardInventory]{Attempted: true, Value: sessionInventory(oldSession)},
	}
	coordinator.Update(stale)
	assertFocusedSession(t, dashboard, newSession.ID)

	newSession.Name = "My task"
	current := messages.SessionsRefreshedMsg{
		DashboardID: coordinator.activeDashboard, Generation: coordinator.viewGeneration,
		SessionEpoch: coordinator.sessionEpoch,
		Inventory:    contracts.ReadResult[contracts.DashboardInventory]{Attempted: true, Value: sessionInventory(oldSession, newSession)},
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
			coordinator, dashboard := newLaunchFocusCoordinator(t, session.Snapshot{ID: "existing", Name: "Existing"})
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
		session.Snapshot{ID: "existing", Name: "Existing"},
		session.Snapshot{ID: "new", Name: "New"},
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
			coordinator, dashboard := newLaunchFocusCoordinator(t, session.Snapshot{ID: "existing", Name: "Existing"})
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
	coordinator, dashboard := newLaunchFocusCoordinator(t, session.Snapshot{ID: "existing", Name: "Existing"})
	staleLaunch := launchCompleted(coordinator, "new", nil)
	coordinator.Update(staleLaunch)
	coordinator.Update(messages.DashboardListRequestedMsg{
		DashboardID: coordinator.activeDashboard, Generation: coordinator.viewGeneration,
	})
	coordinator.Update(messages.DashboardOpenRequestedMsg{
		DashboardID: "default", DashboardName: "Default", DashboardState: contracts.DashboardActive,
	})
	coordinator.Update(staleLaunch)
	setFocusedSessionSnapshots(dashboard, []session.Snapshot{
		{ID: "existing", Name: "Existing"},
		{ID: "new", Name: "New"},
	})
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
