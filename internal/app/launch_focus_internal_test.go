package app

import (
	"errors"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/BlackGiack/kankuro/internal/api"
	"github.com/BlackGiack/kankuro/internal/config"
	"github.com/BlackGiack/kankuro/internal/configstartup"
	"github.com/BlackGiack/kankuro/internal/session"
	"github.com/BlackGiack/kankuro/internal/ui/messages"
	"github.com/BlackGiack/kankuro/internal/ui/views"
)

func newLaunchFocusCoordinator(t *testing.T, sessions ...session.Snapshot) (*Coordinator, *views.Dashboard) {
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
	coordinator := NewCoordinator(repo, manager, api.New(repo, manager))
	coordinator.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	coordinator.Update(messages.DashboardOpenRequestedMsg{
		Dashboard: config.Dashboard{ID: "default", Name: "Default", State: config.DashboardActive},
	})
	dashboard := coordinator.views[ViewDashboard].(*views.Dashboard)
	dashboard.SetData(messages.DataLoadedMsg{Sessions: sessions})
	return coordinator, dashboard
}

func launchCompleted(coordinator *Coordinator, sessionID string, err error) messages.SessionLaunchCompletedMsg {
	return messages.SessionLaunchCompletedMsg{
		SessionID: sessionID, Err: err, DashboardID: coordinator.activeDashboard, Generation: coordinator.viewGeneration,
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
	dashboard.SetSessions([]session.Snapshot{oldSession, newSession})
	assertFocusedSession(t, dashboard, newSession.ID)

	coordinator.Update(messages.ScopedSessionNamedMsg{
		DashboardID: coordinator.activeDashboard, Generation: coordinator.viewGeneration,
		SessionID: newSession.ID, Name: "My task",
	})
	stale := messages.SessionsRefreshedMsg{
		DashboardID: coordinator.activeDashboard, Generation: coordinator.viewGeneration,
		SessionEpoch: preLaunchEpoch, Sessions: []session.Snapshot{oldSession},
	}
	coordinator.Update(stale)
	assertFocusedSession(t, dashboard, newSession.ID)

	newSession.Name = "My task"
	current := messages.SessionsRefreshedMsg{
		DashboardID: coordinator.activeDashboard, Generation: coordinator.viewGeneration,
		SessionEpoch: coordinator.sessionEpoch, Sessions: []session.Snapshot{oldSession, newSession},
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
	if !errors.Is(launcher.LaunchErr(), session.ErrPersistFailed) {
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
				SessionID: "new", DashboardID: test.dashboardID, Generation: test.generation(coordinator),
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
		Dashboard: config.Dashboard{ID: "default", Name: "Default", State: config.DashboardActive},
	})
	coordinator.Update(staleLaunch)
	dashboard.SetSessions([]session.Snapshot{
		{ID: "existing", Name: "Existing"},
		{ID: "new", Name: "New"},
	})
	selected, ok := dashboard.SessionList().Selected()
	if !ok || selected.ID != "existing" || dashboard.FocusedPane() != views.PaneCards {
		t.Fatalf("pending focus crossed dashboard exit: selected=%+v ok=%v pane=%v", selected, ok, dashboard.FocusedPane())
	}
}

func assertFocusedSession(t *testing.T, dashboard *views.Dashboard, sessionID string) {
	t.Helper()
	selected, ok := dashboard.SessionList().Selected()
	if !ok || selected.ID != sessionID || dashboard.FocusedPane() != views.PaneSessions || dashboard.PreviewSessionID() != sessionID {
		t.Fatalf("session focus mismatch: selected=%+v ok=%v pane=%v preview=%q", selected, ok,
			dashboard.FocusedPane(), dashboard.PreviewSessionID())
	}
}
