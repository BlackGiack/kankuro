# Act 2 — Plan 2.2: Clean shutdown — no stray agents, no stray test files

A follow-up to plan 2, done **before plan 3**. It covers two problems with the same root
(the session manager has no real way to stop what it started):

- **Part A (steps 1–5):** tests leave files in the user's temp directories.
- **Part B (steps 6–9):** quitting the app does not kill running agents.

It touches `internal/session`, `internal/driver/driver.go`, `cmd/kankuro/main.go`, and
the test files of `internal/session` and `internal/app`. Step 1 is shared: Part B builds
on the `Shutdown` that step 1 rewrites, so do the parts in order.

**Rules for this session:** don't commit unless asked. Keep all five gates clean. Make
sure each new test or guard can fail.

---

# Part A — tests must not leave anything in the user's temp directories

## What is wrong today

Measured after plan 2 landed (commit `01d299e`):

- **`KANKURO_TMP_DIR` is now clean.** The `TestMain` that session 2 added to
  `internal/app` and `internal/session` points it at a throwaway directory, and a full run
  left no workspace copies in the developer's real `KANKURO_TMP_DIR`. Keep that.
- **`/tmp` is not clean.** Every `go test ./...` leaves about 30 directories named
  `/tmp/Test<Name><digits>/001/sessions/<id>.json`, one per test that launches a session
  (≈60 for a `go test` + `go test -race` pair). This predates Act 2: the commit before
  Act 2 (`e09469c`) leaks 15 per run of the `session` package alone.

**Cause.** `Manager.watch` (`internal/session/watch.go`) is a goroutine per activation
that only returns when a tick sees the process terminated, or when its generation goes
stale. Nothing stops it from outside. After a test's `Kill`, the watcher's next tick (up
to `watchInterval` = 2 s later) sees the dead process, sets `Terminated` again and calls
`persistSessionLocked`. If the test has already finished, `t.TempDir()` has been removed,
and `saveSession`'s `os.MkdirAll` recreates `<tempdir>/001/sessions/` just to write the
record. The testing package's cleanup has already run, so nobody deletes it.

`trackOutput` (`io.go`) also outlives the test, but it writes only memory (scrollback,
subscribers) and never touches the disk. It doesn't need the same treatment. Write that
down in its comment, so nobody "fixes" it later by giving it the same treatment for
symmetry.

The fix has two halves. The **product** half: a `Manager` gets a real way to stop its
goroutines. The **test** half: every test manager is stopped before its temp directory is
removed, and a package-level guard catches any test that forgets.

---

## Step 1 — `Manager` can stop its watchers

`internal/session/manager.go`, `watch.go`, `lifecycle.go`.

Add to `Manager`:

```go
// stop is closed by Shutdown to make every watch goroutine return without a
// final tick; watchers counts them so Shutdown can wait. closed (under mu)
// refuses new activations once Shutdown has started, so no watcher can be
// added after the Wait begins.
stop     chan struct{}
stopOnce sync.Once
watchers sync.WaitGroup
closed   bool
```

Initialise `stop` in `NewManager`.

**Starting a watcher.** Replace both `go m.watch(id, gen, d)` call sites (`startSession`
and `SwitchAgent`, both under `m.mu`) with one helper:

```go
// startWatchLocked must be called with m.mu held, after checking !m.closed.
func (m *Manager) startWatchLocked(id string, gen uint64, d driver.AgentDriver) {
	m.watchers.Add(1)
	go func() {
		defer m.watchers.Done()
		m.watch(id, gen, d)
	}()
}
```

`Add` happens under `m.mu` after the `!closed` check, and `Shutdown` sets `closed` under
`m.mu` before it calls `Wait`. That ordering is what makes the `WaitGroup` use correct (no
`Add` racing a `Wait` that started at zero). Say so in the comment.

**`watch`** waits on both channels:

```go
for {
	select {
	case <-m.stop:
		return
	case <-ticker.C:
	}
	...
}
```

**Refusing activations after shutdown.** Add `var ErrManagerClosed = errors.New("session manager is shut down")`.

- `reserveForActivation` returns it when `m.closed`. That covers `Relaunch` and
  `SwitchAgent`.
- `Launch` / `startSession(isNew)` checks `m.closed` under `m.mu` before doing any work.
- Both activation paths check again under the final `m.mu.Lock()`, just before
  `activateSessionLocked`, because `Shutdown` may have started during the unlocked
  spawn. If it has: unlock, `d.Kill()` the process that was just spawned, release the
  reservation (`Relaunch` / `SwitchAgent`), and return `ErrManagerClosed`. Never leave a
  spawned process without a watcher.

**`Shutdown`** becomes:

```go
// Shutdown terminates all managed sessions and stops their watchers, and
// returns only once no watcher can write to disk any more. It is idempotent.
func (m *Manager) Shutdown() {
	m.mu.Lock()
	m.closed = true
	ids := ...
	m.mu.Unlock()
	for _, id := range ids {
		_ = m.Kill(id)
	}
	m.stopOnce.Do(func() { close(m.stop) })
	m.watchers.Wait()
}
```

Kill comes first, so each record is persisted as `Terminated` by `Kill` itself. The
watchers then exit on `stop` without their redundant final persist.

One consequence: at shutdown, the watcher no longer records `ExitCode` after a `Kill`.
That's acceptable, because a killed process's exit code is just the signal. It's the only
behavioural difference, so mention it in the comment.

`main.go` calls `Shutdown` only on an error exit today. Part B (step 8) makes every exit
path call it.

### Tests for step 1 (`internal/session/lifecycle_test.go`, reuse the `gateHarness`)

1. **`TestShutdownStopsWatchersBeforeTheyPersist`:** launch, `Shutdown()`, then delete the
   whole `sessions/` directory and wait `watchInterval + 500ms`. The directory must not
   reappear.
   *Fails without step 1:* the watcher's tick recreates it.
   This is the one place a sleep is the point of the test. Say so in its comment.
2. **`TestActivationAfterShutdownIsRefused`:** after `Shutdown`, `Launch`, `Relaunch` and
   `SwitchAgent` all return `ErrManagerClosed` (`errors.Is`), and no driver `Spawn` is
   called (the harness's spawn counter).
3. **`TestShutdownDuringSpawnKillsTheNewProcess`:** gate a relaunch's `Spawn`. While it
   blocks, call `Shutdown`. It can return before the gate opens, because the relaunch has
   not added a watcher yet. Release the gate. The relaunch must return
   `ErrManagerClosed`, the *new* gate driver's `Kill` must have been called (no orphaned
   process), and the session must still be `Terminated`.
4. **`TestShutdownIsIdempotent`:** two calls, no panic (`close` of a closed channel), and
   the second returns at once.

All must pass under `go test -race -count=20 ./internal/session/`.

## Step 2 — every test manager is shut down before its temp dir goes

There are about 70 `session.NewManager(` calls in test files across
`internal/session` (package `session_test`) and `internal/app` (packages `app_test`
and, in `refresh_cadence_internal_test.go`, `app`).

Add a helper next to each package's `TestMain` (`main_test.go`):

```go
// newTestManager is session.NewManager plus a Cleanup that shuts it down.
// t.Cleanup runs last-in-first-out, so as long as the config dir's
// t.TempDir() was created *before* this call - always the case when the
// manager is built from it - Shutdown runs before that dir is removed, and
// no watcher can write into it afterwards.
func newTestManager(t testing.TB, configDir string, repo session.WorkspaceReader, opts ...session.Option) *session.Manager
```

`refresh_cadence_internal_test.go` is package `app` and cannot see the `app_test` copy.
Give it its own unexported copy, in an `internal`-package test file
(`helpers_internal_test.go`), rather than moving the whole file between packages.

Then replace every test call of `session.NewManager(` with `newTestManager(t, …)`, and
`newGateHarness` / `newPersistFailureFixture` with the same inside them. Afterwards:

```bash
grep -rn "session.NewManager(" --include=*_test.go internal   # only the helper definitions
```

Things to watch while doing it:

- **Two managers on one directory** (`storage_test.go`'s `sm2`, the reload tests): both
  get the helper. `Shutdown` on a manager with only restored sessions just re-persists
  them as `Terminated`, before the cleanup removes the directory. That's harmless.
- **A test that already calls `sm.Kill` / `sm.Shutdown` in a `defer`**: keep it. Shutdown
  is idempotent, and the `defer` still decides when the processes die within the test.
- **The persist-failure tests make `sessions/` read-only.** Their restore (`chmod 0o700`)
  is a `defer`, which runs before `t.Cleanup`. `Shutdown`'s persists then succeed. If one
  of them makes the directory read-only from a `t.Cleanup` instead, `Shutdown` would run
  first and its persist would fail silently (logged). Check that each one restores
  permissions in a `defer`.
- **Subtests:** a manager must be created in the same (sub)test as the `t.TempDir()` it
  writes into, after it. Grep for managers built in a parent test from a subtest's dir
  (unlikely, but it would make `Shutdown` run after the dir is gone).

## Step 3 — a guard so a forgotten helper call is caught, and never pollutes `/tmp`

Extend each package's `runWithTmpDir` (`internal/app/main_test.go`,
`internal/session/main_test.go`):

1. Create one private root with `os.MkdirTemp("", "kankuro-<pkg>-test-")`, and inside it
   `t/` and `ws/`.
2. Set **`TMPDIR=<root>/t`**. `t.TempDir()` uses `os.TempDir()`, which reads `$TMPDIR` on
   every call on Unix, so every test temp dir now lives under the private root. Set
   `KANKURO_TMP_DIR=<root>/ws`, which replaces the separate directory it uses today.
3. After `m.Run()`, wait `3 * time.Second`. That's longer than `watchInterval` (2 s).
   `session.watchInterval` is unexported, so name it in a comment. The wait gives a stray
   watcher the chance to write. Then list `<root>/t`. If it isn't empty, print the leaked
   paths to stderr and return exit code 1, failing the package.
4. `os.RemoveAll(root)` in every case.

The wait costs 3 s per package, about 6 s per full run. That's the price of catching a
forgotten `newTestManager` deterministically; write the reasoning in the comment.

Even when the guard trips, nothing reaches the real `/tmp`: the leak lands in the private
root, which is removed.

**Can-fail check (mandatory):** make `newTestManager` skip its `t.Cleanup`. `go test
./internal/session/` must fail with the guard's message listing the leaked `sessions/`
directories. Then restore it. Also run the guard once with step 1 reverted, but the
helper in place, so `Shutdown` doesn't stop watchers: it must fail too.

## Step 4 — cover the default temp-workspace location again

The `TestMain` sets `KANKURO_TMP_DIR` for whole packages, so no test exercises the
default `<config>/tmp` location any more (`Remove` and the workspace creation code both
fall back to it when the variable is empty).

Add **`TestTmpWorkspaceDefaultsUnderConfigDir`** (`internal/session/tmp_workspace_test.go`):
`t.Setenv("KANKURO_TMP_DIR", "")`, launch a `UseTmpWorkspace` card, and assert that
`WorkingDir` is `<configDir>/tmp/<sid>`. Then kill it, call `Remove`, and assert that
directory is gone. `prepareWorkspace` (`workspace.go:23`) and `Remove` (`lifecycle.go`)
each repeat the same `os.Getenv` + `<config>/tmp` fallback. Fold them into one
`m.tmpRoot()` helper, so the test covers both through a single path.

## Step 5 — docs and the one-off cleanup

- `project.md` §7 "Testing notes": add that session-launching tests must use
  `newTestManager`, that `TestMain` redirects `TMPDIR` and `KANKURO_TMP_DIR` into a private
  root and fails the package if anything is left in it, and that this costs about 3 s per
  package.
- `project.md` §10, "Change session lifecycle…": add "goroutines that can write to disk
  must be stoppable by `Shutdown`".
- Part B's docs are in step 9.
- **Don't delete anything in the user's real temp directories from code or tests.** In
  the session summary, tell the user what earlier runs left behind, so they can clean it
  themselves:
  - `/tmp/Test*` directories containing only `001/sessions/*.json`;
  - UUID-named workspace copies in their `KANKURO_TMP_DIR` dated before `01d299e`.

---

# Part B — quitting the app kills the running agents

## What is wrong today

Verified in this session, not assumed:

- **`q` and `ctrl+c` on the dashboard** (`dashboard.go:358`, `:381`) return `tea.Quit`.
  `p.Run()` returns `nil`, and `run()` in `cmd/kankuro/main.go` returns **without calling
  `sm.Shutdown()`**. It only calls it on the error path (`if _, err = p.Run(); err != nil`).
- **What then happens to the agents** was checked with a small program that spawns
  through `creack/pty` exactly as `GenericCLIDriver.Spawn` does, then exits without
  killing anything:

  | Agent | After kankuro exits |
  |---|---|
  | plain process (`exec sleep`) | dies: closing the PTY master hangs up the terminal, and the kernel sends SIGHUP |
  | process that ignores SIGHUP | **keeps running**, orphaned, no PTY |
  | child of the agent that ignores SIGHUP (a tool, an MCP server, a detached shell) | **keeps running**, orphaned |

  So most agents die by accident, from SIGHUP, with no chance to save state. Anything
  that ignores SIGHUP, or that the agent started, can survive indefinitely.
- **`Kill` has the same gap.** `GenericCLIDriver.Kill` calls `d.cmd.Process.Kill()`:
  SIGKILL to the *direct child only*. The agent gets no chance to shut down cleanly, and
  its descendants are only reached by the SIGHUP side effect. `pty.Start` runs the child
  with `Setsid`, so the child is the leader of its own session and process group
  (`pgid == pid`). That makes it possible to signal the whole group.
- **Signals to kankuro itself** (Bubble Tea v2.0.7, `tea.go:646-676`): Bubble Tea catches
  SIGINT (`Run` returns `ErrInterrupted`, so today's error path does call `Shutdown`) and
  SIGTERM (it becomes a `QuitMsg`, so `Run` returns `nil` and nothing is killed).
  **Nothing catches SIGHUP**, which is what closing the terminal window sends. Go's
  default action terminates kankuro on the spot, with no cleanup at all.

Required behaviour (from the user): **on app exit, every running session is killed
before kankuro exits.**

## Step 6 — `Kill` terminates the agent's process group, gracefully

`internal/driver/driver.go`, `GenericCLIDriver.Kill` (`ClaudeCodeDriver` embeds it; it is
the only `Kill` implementation).

New behaviour:

1. Under `d.mu`, read `pid` and `done`. If there's no process, return `nil`, as today.
   Then **release the lock**. `watchProcess` needs `d.mu` to record the exit and close
   `done`, so waiting for `done` while holding it would deadlock.
2. Send `syscall.Kill(-pid, syscall.SIGTERM)` to the whole group. Give `ESRCH` the same
   treatment as `os.ErrProcessDone` today: it means the process is already gone, which
   counts as success.
3. Wait for `done`, up to `killGrace` (a field defaulting to `2 * time.Second`, so tests
   can shorten it; add an option or an unexported setter used from an `export_test.go`).
4. If it's still running, send `syscall.Kill(-pid, syscall.SIGKILL)` and wait for `done`
   again, briefly.
5. Re-take `d.mu` and do what `Kill` does today: set the status, close and nil the PTY,
   `closeDone()`. Keep it idempotent against `watchProcess` having done part of this
   already.

Write the *why* in comments at both points: SIGTERM first, so an agent like Claude Code
can save its state; the negative pid, so tools and servers the agent started die with it.

**Known limit (document it, don't fix it):** a descendant that calls `setsid()` or
creates its own process group (a daemonising tool, a job-control shell) is outside the
group and survives. Linux's `Pdeathsig` doesn't help either: it fires when the spawning
*OS thread* exits, which in Go can happen at any time unless that thread is locked. Don't
use it.

`syscall.Kill` with a negative pid is Unix-only. That's acceptable, since `pty.Start` is
the Unix path too. Keep the code buildable with `GOOS=darwin`, and check it with
`GOOS=darwin go build ./...`.

**Callers.** `Manager.Kill` already calls `d.Kill()` without holding `m.mu`, by design
(plan 1, step B3), so the up-to-2-second wait blocks nothing else. The dashboard's `x`
runs `Kill` inside a `tea.Cmd`, so the UI doesn't freeze either.

### Tests (`internal/driver`, real processes via `sh`)

1. **`TestKillTerminatesTheProcessGroup`:** spawn
   `sh -c '(trap "" HUP; exec sleep 1000) & echo $!; wait'`, read the grandchild's pid
   from the output, `Kill`, then poll `syscall.Kill(gpid, 0)` until it returns `ESRCH`
   (timeout 5 s). *Fails today:* this is exactly the survivor from the experiment above.
2. **`TestKillIsGraceful`:** spawn `sh -c 'trap "echo bye > $F; exit 0" TERM; while :; do sleep 0.1; done'`.
   After `Kill`, `$F` must exist. *Fails today:* SIGKILL gives the trap no chance.
3. **`TestKillEscalatesWhenTERMIsIgnored`:** with `killGrace` set to 200 ms, spawn
   `sh -c 'trap "" TERM; while :; do sleep 0.1; done'`. `Kill` must return within about
   1 s, and the process must be gone.
4. The existing `Kill` tests (idempotent, already-exited, concurrent attach input) must
   pass unchanged.

Clean up anything a failing test leaves running: `t.Cleanup` that SIGKILLs the recorded
pids. A failing kill test must not leave `sleep 1000` processes behind.

## Step 7 — `Shutdown` kills in parallel and reports how many

With a 2-second grace per agent, the sequential loop in `Shutdown` takes `N × 2 s`.
Kill every session concurrently (a `sync.WaitGroup` over `m.Kill(id)`), so the total is
about one grace period.

Change the signature to `func (m *Manager) Shutdown() int`, returning how many sessions
were **running or waiting** when it started. `main.go` uses that for one line of output
(step 8). `t.Cleanup` takes a `func()`, so step 2's helper must register
`t.Cleanup(func() { sm.Shutdown() })`, not `t.Cleanup(sm.Shutdown)`. If step 2 was written
with the method value, the compiler will point at it.

Test: **`TestShutdownKillsSessionsInParallel`:** two `gateDriver` sessions whose `Kill`
each sleeps 300 ms. `Shutdown` must return in under 500 ms, report `2`, and leave both
`Terminated` and persisted.

## Step 8 — every exit path calls `Shutdown`

`cmd/kankuro/main.go`:

- Call `Shutdown` after `p.Run()` returns, **whatever it returned**. Write it as plain
  sequential code, not a `defer`, so the order is visible: the TUI is gone → shut down →
  return the error, if any. This replaces the current error-only call.
- **SIGHUP:** before `p.Run()`, `signal.Notify(hup, syscall.SIGHUP)`, and on receipt call
  `p.Quit()`. `Run` then returns, and the shutdown above runs. Comment why: Bubble Tea
  catches SIGINT and SIGTERM but not SIGHUP, and SIGHUP's default action would kill
  kankuro before any cleanup. When the terminal has really gone away, `Run` may return a
  write error instead of `nil`. That's fine, because the shutdown runs either way.
- **Output:** `Shutdown` runs after the TUI has released the terminal, so one line to
  stderr is allowed there and only there:
  `kankuro: stopping N running session(s)…`, printed before the wait and only when
  `N > 0`, so a quit with nothing running stays silent. The "never write to stdout or
  stderr" rule (`project.md` §5) is about the TUI being live; say that in the comment so
  nobody copies the line into app code.
- **A second signal** while shutdown is waiting: Bubble Tea has stopped its own signal
  handling by then, so a second `ctrl+c` or SIGTERM gets Go's default action and exits at
  once. That's an acceptable "force quit", so leave it. Document it in the comment.

Don't add a "N sessions are running, quit anyway?" confirmation. The requirement is to
kill on exit. If the user wants a confirmation later, it belongs on the dashboard's `q`.
Mention it in the session summary as a possible follow-up.

`main.go` has no tests. Verify by hand with the isolated config (`project.md` §7):

1. Launch two sessions of a card whose driver is a script that ignores SIGHUP and starts
   a background child, for example `sh -c 'trap "" HUP; sleep 1000 & sleep 1001'` saved
   as an executable.
2. Quit with `q`. Check `pgrep -f "sleep 100[01]"` is empty and both session records say
   `"status": "terminated"`.
3. Repeat with `kill -TERM <kankuro pid>` from another terminal, and with closing the
   terminal window (SIGHUP).
4. Repeat 1–2 *before* the change and note that the processes survive. That's the
   "seen to fail" for this step.

## Step 9 — docs

- `project.md` §4, "A session's life": add "**Quit**: every running session is
  terminated: SIGTERM to its process group, then SIGKILL after 2 s."
- `project.md` §11: add the known limit, that descendants which start their own session
  or process group survive `Kill`.
- `kankuro_specs.md`: wherever quitting is described, state that quitting kills running
  sessions (records stay, and they can be relaunched next time).
- `tech_debt_act_2.md`: a short "found during Act 2" entry recording what was wrong and
  that plan 2.2 fixed it.

---

## Done when

- [ ] `Manager` has `stop`, `watchers` and `closed`; `Shutdown` kills, stops and waits; activations after shutdown return `ErrManagerClosed`.
- [ ] Step 1 tests pass, and test 1 fails with the `stop` case removed from `watch`.
- [ ] `grep -rn "session.NewManager(" --include=*_test.go internal` shows only the helper definitions.
- [ ] The `TestMain` guard exists in both packages and was seen to fail with a missing cleanup.
- [ ] Default-location test added.
- [ ] Two consecutive `go test ./...` + `go test -race ./internal/...` runs add **nothing** to `/tmp` or to the user's `KANKURO_TMP_DIR`. Check with a before/after `ls -A` diff of both, and paste the empty diff in the session summary.
- [ ] `GenericCLIDriver.Kill` sends SIGTERM to the process group, escalates to SIGKILL after the grace period, and waits without holding `d.mu`; step 6 tests pass, and tests 1–2 fail on the old `Kill`.
- [ ] `Shutdown` kills in parallel and returns the running count; `main.go` calls it on every exit path and handles SIGHUP.
- [ ] The manual check in step 8 was done before and after the change; after it, nothing survives `q`, SIGTERM or closing the terminal.
- [ ] `GOOS=darwin go build ./...` succeeds.
- [ ] `project.md`, `kankuro_specs.md` and `tech_debt_act_2.md` updated.
- [ ] All five gates clean.
