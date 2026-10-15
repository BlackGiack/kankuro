package app

// TestCommandClosuresDoNotRaceCoordinator is a runtime backstop for invariant
// 1 (no tea.Cmd closure touches coordinator-goroutine state: c.views,
// c.preview, c.currentView, c.err, c.settings), in addition to the structural
// rule (refresher and sessionActions hold no *Coordinator) and the greps in
// project.md §6. It only means something under `go test -race`; without the
// race detector it merely checks that the flows still complete.
//
// How it works: for each message that returns commands, Update runs on the
// test goroutine (as it does in Bubble Tea), every returned command is then
// executed on its own goroutine (as Bubble Tea does), and meanwhile the test
// goroutine keeps *writing* the state those closures must not touch - through
// Update, which assigns into c.views, c.err, c.currentView and c.settings,
// not just reads. A closure that reads any of them is then an unsynchronised
// access the detector reports. The hammering has to write: two concurrent
// reads of a map are not a race.
//
// Limit: receiving a closure's result from the channel orders everything
// after it behind the closure, so a violation is only reported if a hammer
// write lands between the closure's read and that receive. runCommands
// therefore hammers for a whole slice before draining. That catches reads
// made early in a closure, including fast ones; a closure that reads state
// late, after other work or a delay, can still slip through. This is a
// backstop, not a proof.
//
// Can-fail check (redo it if you change the harness): the fast case is the
// one that matters, because it is what regressed before. In callbacks.go,
// wrap the command launchIDE returns in a closure that reads settings:
//
//	inner := c.actions.launchIDE(sessionID, c.settings.IDECommand)
//	return func() tea.Msg {
//		if c.settings.IDECommand == "planted" { return nil }
//		return inner()
//	}
//
// and this test must fail under -race (the "open IDE" case). Use the value
// in a condition: a bare `_ = c.settings.IDECommand` is a dead load the
// compiler can drop, which leaves the race detector nothing to see.

import (
	"errors"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/BlackGiack/kankuro/internal/config"
	"github.com/BlackGiack/kankuro/internal/ui/messages"
)

// raceNoopMsg is a message no handler knows, so Update routes it to the
// current view and then assigns the result back into c.views - the map write
// a leaked closure read would race with.
type raceNoopMsg struct{}

type raceFixture struct {
	coord     *Coordinator
	sessionID string
}

func newRaceFixture(t *testing.T, killSession bool) raceFixture {
	t.Helper()
	repo, err := config.NewFileRepository(t.TempDir())
	if err != nil {
		t.Fatalf("new repo: %v", err)
	}
	sm := newTestManager(t, repo.Dir(), repo)

	if saveErr := repo.SaveWorkspaceSources([]config.WorkspaceSource{{Path: t.TempDir()}}); saveErr != nil {
		t.Fatalf("save workspace sources: %v", saveErr)
	}
	workspaces, err := repo.ListWorkspaces()
	if err != nil || len(workspaces) == 0 {
		t.Fatalf("list workspaces: %v (%d)", err, len(workspaces))
	}

	var first config.IDCard
	for _, name := range []string{"race-a", "race-b"} {
		created, createErr := repo.CreateIDCard(config.IDCard{Name: name, Driver: "cat", WorkspaceID: workspaces[0].ID})
		if createErr != nil {
			t.Fatalf("create card %s: %v", name, createErr)
		}
		if first.ID == "" {
			first = created
		}
	}
	card := first

	sessionID, err := sm.Launch(card, nil, nil, nil)
	if err != nil {
		t.Fatalf("launch: %v", err)
	}
	if killSession {
		if killErr := sm.Kill(sessionID); killErr != nil {
			t.Fatalf("kill: %v", killErr)
		}
	}

	c := NewCoordinator(repo, sm)
	c.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	c.Update(c.refreshDataSeeded()())
	return raceFixture{coord: c, sessionID: sessionID}
}

// hammer writes coordinator-goroutine state through the public message path.
func hammer(c *Coordinator) {
	c.Update(raceNoopMsg{})
	c.Update(messages.ErrorMsg{Err: errors.New("hammer")})
	c.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	c.Update(messages.SettingsSavedMsg{Settings: config.DefaultAppSettings()})
	c.Update(messages.NavigateMsg{View: messages.ViewSettings})
	_ = c.View()
	c.Update(messages.NavigateMsg{View: messages.ViewDashboard})
	_ = c.View()
}

// spawnCommands executes cmd (unwrapping tea.BatchMsg) with every leaf on its
// own goroutine, sending each leaf's message to results.
func spawnCommands(cmd tea.Cmd, results chan<- tea.Msg) {
	if cmd == nil {
		return
	}
	go func() {
		msg := cmd()
		if batch, ok := msg.(tea.BatchMsg); ok {
			for _, sub := range batch {
				spawnCommands(sub, results)
			}
			return
		}
		if msg != nil {
			results <- msg
		}
	}()
}

// drainResults moves whatever is currently queued on results into got
// without blocking, and reports whether any of it satisfied want.
func drainResults(results <-chan tea.Msg, got []tea.Msg, want func(tea.Msg) bool) ([]tea.Msg, bool) {
	matched := false
	for {
		select {
		case m := <-results:
			got = append(got, m)
			matched = matched || (want != nil && want(m))
		default:
			return got, matched
		}
	}
}

// runCommands executes cmd concurrently (see spawnCommands) while hammering
// the coordinator on this goroutine, until want has seen a result (or, with a
// nil want, for minHammer). It returns every leaf message that arrived.
func runCommands(t *testing.T, c *Coordinator, cmd tea.Cmd, want func(tea.Msg) bool) []tea.Msg {
	t.Helper()
	const (
		minHammer   = 100 * time.Millisecond
		hammerSlice = 300 * time.Millisecond
		maxWait     = 3 * time.Second
	)
	results := make(chan tea.Msg, 64)
	spawnCommands(cmd, results)

	var got []tea.Msg
	seen := want == nil
	start := time.Now()
	for time.Since(start) < maxWait && (time.Since(start) < minHammer || !seen) {
		// Hammer for a whole slice before looking at results: receiving a
		// result orders everything after it behind the closure, so a write
		// must land between the closure's read and that receive. Draining
		// after every hammer() would miss any closure that returns fast.
		for slice := time.Now(); time.Since(slice) < hammerSlice; {
			hammer(c)
		}
		var matched bool
		got, matched = drainResults(results, got, want)
		seen = seen || matched
	}
	if !seen {
		t.Fatalf("expected command to produce the awaited message within %v; got %d messages: %v", maxWait, len(got), got)
	}
	return got
}

func TestCommandClosuresDoNotRaceCoordinator(t *testing.T) {
	key := func(s string) tea.KeyPressMsg { return tea.KeyPressMsg{Code: 0, Text: s} }
	isMsg := func(match func(tea.Msg) bool) func(tea.Msg) bool { return match }

	tests := []struct {
		name       string
		killedSess bool
		// run drives Update on the test goroutine and returns the command(s)
		// whose closures will then be executed concurrently.
		run  func(t *testing.T, f raceFixture) tea.Cmd
		want func(tea.Msg) bool
	}{
		{
			name: "idle tick",
			run: func(_ *testing.T, f raceFixture) tea.Cmd {
				_, cmd := f.coord.Update(tickMsg(time.Now()))
				return cmd
			},
			want: isMsg(func(m tea.Msg) bool { _, ok := m.(messages.SessionsRefreshedMsg); return ok }),
		},
		{
			name: "data refresh requested",
			run: func(_ *testing.T, f raceFixture) tea.Cmd {
				_, cmd := f.coord.Update(messages.DataRefreshRequestedMsg{})
				return cmd
			},
			want: isMsg(func(m tea.Msg) bool { _, ok := m.(configFingerprintMsg); return ok }),
		},
		{
			name: "launch card (Enter on the cards pane)",
			run: func(_ *testing.T, f raceFixture) tea.Cmd {
				_, cmd := f.coord.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
				return cmd
			},
			want: isMsg(func(m tea.Msg) bool { _, ok := m.(messages.SessionLaunchCompletedMsg); return ok }),
		},
		{
			name: "preview request",
			run: func(t *testing.T, f raceFixture) tea.Cmd {
				_, cmd := f.coord.Update(messages.SessionPreviewRequestedMsg{SessionID: f.sessionID})
				if cmd == nil {
					t.Fatal("expected a live-output wait command: the preview should have subscribed")
				}
				return cmd
			},
		},
		{
			name: "kill the selected session",
			run: func(_ *testing.T, f raceFixture) tea.Cmd {
				f.coord.Update(tea.KeyPressMsg{Code: tea.KeyTab})
				_, cmd := f.coord.Update(key("x"))
				return cmd
			},
			want: isMsg(func(m tea.Msg) bool { _, ok := m.(messages.DataRefreshRequestedMsg); return ok }),
		},
		{
			name: "open IDE for the selected session",
			run: func(_ *testing.T, f raceFixture) tea.Cmd {
				f.coord.Update(tea.KeyPressMsg{Code: tea.KeyTab})
				_, cmd := f.coord.Update(key("o"))
				return cmd
			},
			want: isMsg(func(m tea.Msg) bool { _, ok := m.(messages.IDELaunchedMsg); return ok }),
		},
		{
			name: "open terminal for the selected session",
			run: func(_ *testing.T, f raceFixture) tea.Cmd {
				f.coord.Update(messages.SettingsSavedMsg{Settings: config.AppSettings{TerminalCommand: "/bin/true"}})
				f.coord.Update(tea.KeyPressMsg{Code: tea.KeyTab})
				_, cmd := f.coord.Update(key("t"))
				return cmd
			},
			want: isMsg(func(m tea.Msg) bool { _, ok := m.(messages.TerminalLaunchedMsg); return ok }),
		},
		{
			name:       "relaunch a terminated session (Enter on the sessions pane)",
			killedSess: true,
			run: func(_ *testing.T, f raceFixture) tea.Cmd {
				f.coord.Update(tea.KeyPressMsg{Code: tea.KeyTab})
				_, cmd := f.coord.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
				return cmd
			},
			want: isMsg(func(m tea.Msg) bool { _, ok := m.(messages.SessionAttachRequestedMsg); return ok }),
		},
		{
			name:       "switch agent on a terminated session",
			killedSess: true,
			run: func(t *testing.T, f raceFixture) tea.Cmd {
				f.coord.Update(messages.AgentSwitchRequestedMsg{SessionID: f.sessionID, SessionName: "race"})
				if f.coord.CurrentView() != ViewAgentSwitch {
					t.Fatalf("expected the agent-switch view, got %v", f.coord.CurrentView())
				}
				// Refresh after navigating, or the switch view's list is
				// empty and OnSwitch never fires. Twice: see
				// TestPersistFailureWarningReachesViewOnBothPaths.
				f.coord.Update(f.coord.RefreshData()())
				f.coord.Update(f.coord.RefreshData()())
				_, cmd := f.coord.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
				return cmd
			},
			want: isMsg(func(m tea.Msg) bool { _, ok := m.(messages.AgentSwitchedMsg); return ok }),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newRaceFixture(t, tt.killedSess)
			cmd := tt.run(t, f)
			if cmd == nil {
				t.Fatal("expected Update to return a command")
			}

			// Round 1: the commands Update returned, run concurrently.
			results := runCommands(t, f.coord, cmd, tt.want)

			// Round 2: feed the results back through Update the way Bubble
			// Tea would, and run whatever those return concurrently too.
			for _, msg := range results {
				_, next := f.coord.Update(msg)
				if next != nil {
					runCommands(t, f.coord, next, nil)
				}
			}
		})
	}
}
