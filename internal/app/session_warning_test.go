package app_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/BlackGiack/kankuro/internal/api/contracts"
	"github.com/BlackGiack/kankuro/internal/app"
	"github.com/BlackGiack/kankuro/internal/config"
	"github.com/BlackGiack/kankuro/internal/session"
	"github.com/BlackGiack/kankuro/internal/ui/messages"
	"github.com/BlackGiack/kankuro/internal/ui/views"
)

// collectMsgs executes cmd, recursing into tea.BatchMsg, and returns every
// leaf message it produces. It does not chain into commands returned by
// Update itself: the flows under test only need the messages a callback's
// own Cmd produces, not whatever those in turn trigger (e.g. an actual PTY
// attach, which shows up here as an opaque internal execMsg we don't need to
// unpack further).
func collectMsgs(t *testing.T, cmd tea.Cmd) []tea.Msg {
	t.Helper()
	if cmd == nil {
		return nil
	}
	msg := cmd()
	if msg == nil {
		return nil
	}
	if batch, ok := msg.(tea.BatchMsg); ok {
		var out []tea.Msg
		for _, c := range batch {
			out = append(out, collectMsgs(t, c)...)
		}
		return out
	}
	return []tea.Msg{msg}
}

// drive executes the command batch and delivers each leaf message once. It does not
// execute commands returned by Update, which may block on a PTY or subscription.
func drive(t *testing.T, coord *app.Coordinator, cmd tea.Cmd) {
	t.Helper()
	for _, msg := range collectMsgs(t, cmd) {
		coord.Update(msg)
	}
}

// executePreparedActivation drives the preparation read and then its activation.
func executePreparedActivation(t *testing.T, coord *app.Coordinator, cmd tea.Cmd) tea.Msg {
	t.Helper()
	if cmd == nil {
		t.Fatal("missing preparation command")
	}
	_, next := coord.Update(cmd())
	if next == nil {
		t.Fatalf("preparation did not start activation: %v", coord.Err())
	}
	return next()
}

// Persistence warnings must reach the error banner on both relaunch and switch.
func TestPersistFailureWarningReachesViewOnBothPaths(t *testing.T) {
	t.Run("relaunch", func(t *testing.T) {
		repo, sm, sessionsDir := newPersistFailureFixture(t)
		wd := t.TempDir()
		_ = repo.SaveWorkspaceSources([]config.WorkspaceSource{{Path: wd, IsRoot: false}})
		workspaces, _ := repo.ListWorkspaces()

		card, errCreate := createCurrentCard(repo, config.IDCard{
			Name:        "relaunch-warning",
			Driver:      "cat",
			WorkspaceID: workspaces[0].ID,
		})
		if errCreate != nil {
			t.Fatalf("create card: %v", errCreate)
		}

		sessionID, errLaunch := launchCurrentSession(repo, sm, card)
		if errLaunch != nil {
			t.Fatalf("launch: %v", errLaunch)
		}
		if errKill := sm.Kill(sessionID); errKill != nil {
			t.Fatalf("kill: %v", errKill)
		}
		time.Sleep(50 * time.Millisecond)

		restore := makeSessionsDirReadOnly(t, sessionsDir)
		defer restore()

		coord := newOpenedCoordinator(repo, sm)
		coord.Init()

		cmd := coord.DashboardCallbacks().OnRelaunch(sessionID)
		msgs := []tea.Msg{executePreparedActivation(t, coord, cmd)}
		defer func() {
			restore()
			_ = sm.Kill(sessionID)
		}()

		assertWarningVisible(t, coord, msgs, "relaunch")
	})

	t.Run("switch", func(t *testing.T) {
		repo, sm, sessionsDir := newPersistFailureFixture(t)
		wd := t.TempDir()
		_ = repo.SaveWorkspaceSources([]config.WorkspaceSource{{Path: wd, IsRoot: false}})
		workspaces, _ := repo.ListWorkspaces()

		card, errCreate := createCurrentCard(repo, config.IDCard{
			Name:        "switch-warning-a",
			Driver:      "cat",
			WorkspaceID: workspaces[0].ID,
		})
		if errCreate != nil {
			t.Fatalf("create card: %v", errCreate)
		}
		card2, errCreate2 := createCurrentCard(repo, config.IDCard{
			Name:        "switch-warning-b",
			Driver:      "cat",
			WorkspaceID: workspaces[0].ID,
		})
		if errCreate2 != nil {
			t.Fatalf("create card2: %v", errCreate2)
		}

		sessionID, errLaunch := launchCurrentSession(repo, sm, card)
		if errLaunch != nil {
			t.Fatalf("launch: %v", errLaunch)
		}
		if errKill := sm.Kill(sessionID); errKill != nil {
			t.Fatalf("kill: %v", errKill)
		}
		time.Sleep(50 * time.Millisecond)

		restore := makeSessionsDirReadOnly(t, sessionsDir)
		defer restore()

		coord := newOpenedCoordinator(repo, sm)
		coord.Init()
		coord.Update(messages.AgentSwitchRequestedMsg{SessionID: sessionID, DashboardID: "default", Generation: 1})

		cmd := coord.AgentSwitchCallbacks().OnSwitch(sessionID, card2.ID)
		msgs := []tea.Msg{executePreparedActivation(t, coord, cmd)}
		defer func() {
			restore()
			_ = sm.Kill(sessionID)
		}()

		assertWarningVisible(t, coord, msgs, "switch")
	})
}

// Switching through the event loop must preserve the dashboard model. Callback-only
// tests cannot exercise the coordinator's model write-back.
func TestAgentSwitchDoesNotClobberDashboardView(t *testing.T) {
	repo, sm, _ := newPersistFailureFixture(t)
	wd := t.TempDir()
	_ = repo.SaveWorkspaceSources([]config.WorkspaceSource{{Path: wd, IsRoot: false}})
	workspaces, _ := repo.ListWorkspaces()

	card, errCreate := createCurrentCard(repo, config.IDCard{
		Name:        "clobber-a",
		Driver:      "cat",
		WorkspaceID: workspaces[0].ID,
	})
	if errCreate != nil {
		t.Fatalf("create card: %v", errCreate)
	}
	card2, errCreate2 := createCurrentCard(repo, config.IDCard{
		Name:        "clobber-b",
		Driver:      "cat",
		WorkspaceID: workspaces[0].ID,
	})
	if errCreate2 != nil {
		t.Fatalf("create card2: %v", errCreate2)
	}

	sessionID, errLaunch := launchCurrentSession(repo, sm, card)
	if errLaunch != nil {
		t.Fatalf("launch: %v", errLaunch)
	}
	if errKill := sm.Kill(sessionID); errKill != nil {
		t.Fatalf("kill: %v", errKill)
	}
	time.Sleep(50 * time.Millisecond)
	defer func() { _ = sm.Kill(sessionID) }()

	coord := newOpenedCoordinator(repo, sm)
	coord.Init()

	// Give every view a real size: AgentSwitchView's list needs a non-zero
	// size for SelectedItem() to return anything.
	coord.Update(tea.WindowSizeMsg{Width: 120, Height: 40})

	// Navigate to the switch view first, THEN refresh. Refreshing before
	// navigating (or not at all) leaves AgentSwitchView's card list empty,
	// SelectedItem() returns nil, OnSwitch never runs, and the test would
	// pass for the wrong reason.
	coord.Update(messages.AgentSwitchRequestedMsg{SessionID: sessionID, SessionName: card.Name, DashboardID: "default", Generation: 1})
	if coord.CurrentView() != app.ViewAgentSwitch {
		t.Fatalf("expected to be on ViewAgentSwitch, got %v", coord.CurrentView())
	}
	coord.Update(coord.RefreshData()())

	// Cards are listed alphabetically by name (repo.ListIDCards), so
	// "clobber-a" (the session's current card) sits at index 0. Move down
	// once to select "clobber-b" - selecting the already-active card would
	// make SwitchAgent's CardID assertion below trivially true.
	coord.Update(tea.KeyPressMsg{Code: tea.KeyDown})

	// Run the actual switch command and its resulting messages through the event loop.
	_, cmd := coord.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	coord.Update(executePreparedActivation(t, coord, cmd))

	// The switch must have actually run before this test's result means
	// anything - otherwise a no-op Enter would pass for the wrong reason too.
	sess, ok := sm.GetSession(sessionID)
	if !ok {
		t.Fatalf("session not found after switch")
	}
	if sess.CardID != card2.ID {
		t.Fatalf("expected SwitchAgent to have run (CardID=%s), got CardID=%s", card2.ID, sess.CardID)
	}
	if sess.Status != session.StatusRunning && sess.Status != session.StatusWaiting {
		t.Fatalf("expected the switched session to be running, got status=%s", sess.Status)
	}

	coord.Update(messages.ScopedNavigateMsg{View: messages.ViewDashboard, DashboardID: "default", Generation: 1})
	if coord.CurrentView() != app.ViewDashboard {
		t.Fatalf("expected to be back on ViewDashboard, got %v", coord.CurrentView())
	}
	if strings.Contains(coord.View(), "Switch Agent") {
		t.Fatalf("dashboard slot was overwritten with the AgentSwitchView: %q", coord.View())
	}
	if _, isDashboard := coord.Views()[app.ViewDashboard].(*views.Dashboard); !isDashboard {
		t.Fatalf("c.views[ViewDashboard] is no longer a *views.Dashboard")
	}
}

func newPersistFailureFixture(t *testing.T) (config.Repository, *session.Manager, string) {
	t.Helper()
	repo, errRepo := newCurrentRepository(t.TempDir())
	if errRepo != nil {
		t.Fatalf("new repo: %v", errRepo)
	}
	sm := newTestManager(t, repo.Dir(), repo)
	sessionsDir := filepath.Join(repo.Dir(), "sessions")
	return repo, sm, sessionsDir
}

// makeSessionsDirReadOnly strips write permission from the sessions directory
// so the next session save fails deterministically. It has to be the
// directory, not the record: records are written with temp-file + rename,
// which replaces the target whatever its own mode is. Returns a restore func
// to put permissions back (idempotent, safe to call from a defer even after
// an earlier explicit call).
func makeSessionsDirReadOnly(t *testing.T, sessionsDir string) func() {
	t.Helper()
	if os.Geteuid() == 0 {
		t.Skip("running as root: file permissions are not enforced")
	}
	if err := os.Chmod(sessionsDir, 0o500); err != nil {
		t.Fatalf("chmod: %v", err)
	}
	return func() {
		_ = os.Chmod(sessionsDir, 0o700)
	}
}

func assertWarningVisible(t *testing.T, coord *app.Coordinator, msgs []tea.Msg, label string) {
	t.Helper()
	if len(msgs) != 1 {
		t.Fatalf("%s: expected one result, got %v", label, msgs)
	}
	result, ok := msgs[0].(messages.SessionActivationCompletedMsg)
	if !ok || result.Result.Effect != contracts.UsableActivation || result.Result.Diagnostic == "" {
		t.Fatalf("%s: activation result = %+v", label, msgs[0])
	}
	coord.Update(result)
	if coord.Err() == nil || !strings.Contains(coord.View(), "could not be saved") {
		t.Fatalf("%s: warning missing from view: %q", label, coord.View())
	}
}
