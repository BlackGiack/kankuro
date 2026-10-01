package app_test

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"

	"kankuro/internal/app"
	"kankuro/internal/config"
	"kankuro/internal/session"
	"kankuro/internal/ui/messages"
	"kankuro/internal/ui/views"
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

// drive executes cmd (recursing into tea.BatchMsg via collectMsgs) and feeds
// every resulting leaf message through coord.Update exactly once. It does
// NOT recurse into whatever Cmd those Update calls in turn return - like
// collectMsgs, on purpose: some of those (e.g. an actual PTY attach, or a
// preview subscription's blocking channel read) are exactly the kind of
// thing a unit test must not execute. This is enough to observe a callback's
// deferred side effect (debt_impl_plan_4.md step 3B moved several of these
// from synchronous calls into a returned tea.Cmd) without turning the test
// into a full Bubble Tea runtime.
func drive(t *testing.T, coord *app.Coordinator, cmd tea.Cmd) {
	t.Helper()
	for _, msg := range collectMsgs(t, cmd) {
		coord.Update(msg)
	}
}

// permute calls fn once for every permutation of msgs (Heap's algorithm).
// tea.Batch makes no delivery-order guarantee, so a fix that only happens to
// work for one arrival order isn't actually fixed - this is what makes the
// test below catch Finding 4 instead of getting lucky on a single order.
func permute(msgs []tea.Msg, fn func([]tea.Msg)) {
	n := len(msgs)
	cur := make([]tea.Msg, n)
	copy(cur, msgs)
	var generate func(k int)
	generate = func(k int) {
		if k == 1 {
			fn(cur)
			return
		}
		for i := range k {
			generate(k - 1)
			if k%2 == 0 {
				cur[i], cur[k-1] = cur[k-1], cur[i]
			} else {
				cur[0], cur[k-1] = cur[k-1], cur[0]
			}
		}
	}
	generate(n)
}

// TestPersistFailureWarningReachesViewOnBothPaths pins findings 4 and 5 from
// review_dp_2_2.md: a session.ErrPersistFailed outcome must reach the
// coordinator's error banner (via SessionAttachRequestedMsg.Warning) on
// *both* the relaunch path and the agent-switch path, not just one of them.
// Before the Finding-4 fix, the switch path batched a separate ErrorMsg
// beside a NavigateMsg, which unconditionally clears c.err - so whether the
// warning survived depended on tea.Batch's unspecified delivery order.
func TestPersistFailureWarningReachesViewOnBothPaths(t *testing.T) {
	t.Run("relaunch", func(t *testing.T) {
		repo, sm, sessionsDir := newPersistFailureFixture(t)
		wd := t.TempDir()
		_ = repo.SaveWorkspaceSources([]config.WorkspaceSource{{Path: wd, IsRoot: false}})
		workspaces, _ := repo.ListWorkspaces()

		card, errCreate := repo.CreateIDCard(config.IDCard{
			Name:        "relaunch-warning",
			Driver:      "cat",
			WorkspaceID: workspaces[0].ID,
		})
		if errCreate != nil {
			t.Fatalf("create card: %v", errCreate)
		}

		sessionID, errLaunch := sm.Launch(card, nil, nil, nil)
		if errLaunch != nil {
			t.Fatalf("launch: %v", errLaunch)
		}
		if errKill := sm.Kill(sessionID); errKill != nil {
			t.Fatalf("kill: %v", errKill)
		}
		time.Sleep(50 * time.Millisecond)

		restore := makeSessionsDirReadOnly(t, sessionsDir)
		defer restore()

		coord := app.NewCoordinator(repo, sm)
		coord.Init()

		cmd := coord.DashboardCallbacks().OnRelaunch(sessionID)
		msgs := collectMsgs(t, cmd)
		defer func() {
			restore()
			_ = sm.Kill(sessionID)
		}()

		assertWarningSurvivesEveryOrder(t, repo, sm, msgs, "relaunch")
	})

	t.Run("switch", func(t *testing.T) {
		repo, sm, sessionsDir := newPersistFailureFixture(t)
		wd := t.TempDir()
		_ = repo.SaveWorkspaceSources([]config.WorkspaceSource{{Path: wd, IsRoot: false}})
		workspaces, _ := repo.ListWorkspaces()

		card, errCreate := repo.CreateIDCard(config.IDCard{
			Name:        "switch-warning-a",
			Driver:      "cat",
			WorkspaceID: workspaces[0].ID,
		})
		if errCreate != nil {
			t.Fatalf("create card: %v", errCreate)
		}
		card2, errCreate2 := repo.CreateIDCard(config.IDCard{
			Name:        "switch-warning-b",
			Driver:      "cat",
			WorkspaceID: workspaces[0].ID,
		})
		if errCreate2 != nil {
			t.Fatalf("create card2: %v", errCreate2)
		}

		sessionID, errLaunch := sm.Launch(card, nil, nil, nil)
		if errLaunch != nil {
			t.Fatalf("launch: %v", errLaunch)
		}
		if errKill := sm.Kill(sessionID); errKill != nil {
			t.Fatalf("kill: %v", errKill)
		}
		time.Sleep(50 * time.Millisecond)

		restore := makeSessionsDirReadOnly(t, sessionsDir)
		defer restore()

		coord := app.NewCoordinator(repo, sm)
		coord.Init()

		cmd := coord.AgentSwitchCallbacks().OnSwitch(sessionID, card2.ID)
		msgs := collectMsgs(t, cmd)
		defer func() {
			restore()
			_ = sm.Kill(sessionID)
		}()

		assertWarningSurvivesEveryOrder(t, repo, sm, msgs, "switch")
	})
}

// TestAgentSwitchDoesNotClobberDashboardView pins Finding 6 from
// review_dp_2_3.md: switching an agent through the real event loop must not
// leave c.views[ViewDashboard] holding the AgentSwitchView. That regression
// lived entirely in Coordinator.Update's write-back (updateCurrentView's
// caller stores the returned model under c.currentView right after the view
// itself is updated), which callback-level tests like
// TestPersistFailureWarningReachesViewOnBothPaths cannot exercise, because
// they call the callback's returned Cmd directly and never go through
// Update. This test drives the switch through coord.Update using the same
// messages the real TUI would send.
func TestAgentSwitchDoesNotClobberDashboardView(t *testing.T) {
	repo, sm, _ := newPersistFailureFixture(t)
	wd := t.TempDir()
	_ = repo.SaveWorkspaceSources([]config.WorkspaceSource{{Path: wd, IsRoot: false}})
	workspaces, _ := repo.ListWorkspaces()

	card, errCreate := repo.CreateIDCard(config.IDCard{
		Name:        "clobber-a",
		Driver:      "cat",
		WorkspaceID: workspaces[0].ID,
	})
	if errCreate != nil {
		t.Fatalf("create card: %v", errCreate)
	}
	card2, errCreate2 := repo.CreateIDCard(config.IDCard{
		Name:        "clobber-b",
		Driver:      "cat",
		WorkspaceID: workspaces[0].ID,
	})
	if errCreate2 != nil {
		t.Fatalf("create card2: %v", errCreate2)
	}

	sessionID, errLaunch := sm.Launch(card, nil, nil, nil)
	if errLaunch != nil {
		t.Fatalf("launch: %v", errLaunch)
	}
	if errKill := sm.Kill(sessionID); errKill != nil {
		t.Fatalf("kill: %v", errKill)
	}
	time.Sleep(50 * time.Millisecond)
	defer func() { _ = sm.Kill(sessionID) }()

	coord := app.NewCoordinator(repo, sm)
	coord.Init()

	// Give every view a real size: AgentSwitchView's list needs a non-zero
	// size for SelectedItem() to return anything.
	coord.Update(tea.WindowSizeMsg{Width: 120, Height: 40})

	// Navigate to the switch view first, THEN refresh. Refreshing before
	// navigating (or not at all) leaves AgentSwitchView's card list empty,
	// SelectedItem() returns nil, OnSwitch never runs, and the test would
	// pass for the wrong reason.
	coord.Update(messages.AgentSwitchRequestedMsg{SessionID: sessionID, SessionName: card.Name})
	if coord.CurrentView() != app.ViewAgentSwitch {
		t.Fatalf("expected to be on ViewAgentSwitch, got %v", coord.CurrentView())
	}
	coord.Update(coord.RefreshData()())
	// Refresh twice: the dashboard auto-selects a preview session on first
	// data load, which makes the coordinator immediately broadcast a
	// second, cards-less DataLoadedMsg (handlePreviewRequest) to fetch its
	// scrollback - and every DataReceiver, including AgentSwitchView,
	// overwrites its card list unconditionally on every SetData call. That
	// wipes the list we just populated. This is a pre-existing quirk
	// unrelated to Finding 6/7 (broadcastData has no notion of a partial
	// update); the second refresh's preview id matches what's already
	// stored, so it does not re-trigger the wipe, and the list survives.
	coord.Update(coord.RefreshData()())

	// Cards are listed alphabetically by name (repo.ListIDCards), so
	// "clobber-a" (the session's current card) sits at index 0. Move down
	// once to select "clobber-b" - selecting the already-active card would
	// make SwitchAgent's CardID assertion below trivially true.
	coord.Update(tea.KeyPressMsg{Code: tea.KeyDown})

	// Real OnSwitch, through Update: this is what exercises the
	// updateCurrentView write-back that Finding 6 lived in. OnSwitch now
	// defers the actual switch into the returned Cmd (debt_impl_plan_4.md
	// step 3B), so drive() must run that Cmd - and everything it in turn
	// produces - for the switch to actually happen, the same way a real
	// Bubble Tea program's runtime would.
	_, cmd := coord.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	drive(t, coord, cmd)

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

	coord.Update(messages.NavigateMsg{View: messages.ViewDashboard})
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
	repo, errRepo := config.NewFileRepository(t.TempDir())
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

// assertWarningSurvivesEveryOrder replays msgs, in every possible order,
// against a fresh Coordinator each time, and requires the persist-failure
// warning to be visible after every single ordering. tea.Batch gives no
// delivery-order guarantee, so anything less than "every order" leaves the
// warning's visibility down to luck - which is exactly the bug in Finding 4.
func assertWarningSurvivesEveryOrder(t *testing.T, repo config.Repository, sm *session.Manager, msgs []tea.Msg, label string) {
	t.Helper()
	if len(msgs) == 0 {
		t.Fatalf("%s: expected at least one message from the callback's command", label)
	}

	tested := 0
	permute(msgs, func(order []tea.Msg) {
		tested++
		coord := app.NewCoordinator(repo, sm)
		coord.Init()
		for _, m := range order {
			coord.Update(m)
		}

		if coord.Err() == nil {
			t.Errorf("%s: order %v: expected coord.Err() to carry the persist-failure warning, got nil", label, order)
			return
		}
		if !errors.Is(coord.Err(), session.ErrPersistFailed) {
			t.Errorf("%s: order %v: expected coord.Err() to wrap session.ErrPersistFailed, got: %v", label, order, coord.Err())
			return
		}
		if !strings.Contains(coord.View(), "could not be saved") {
			t.Errorf("%s: order %v: expected the rendered view to contain the warning, got: %q", label, order, coord.View())
		}
	})
	if tested == 0 {
		t.Fatalf("%s: permute never invoked fn", label)
	}
}
