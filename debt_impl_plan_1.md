# Debt Plan 1 — Session ownership, relaunch safety & child exit status

> Covers findings **#6, #3, #15, the exit-code half of #9, and #2** from
> `tech_debt_founding.md`. **Run this plan first**: step 1 is what makes
> `go test -race ./internal/session/...` pass again, and plans 2 and 4 rely on
> `-race` as their verification gate.

## Why these ship together

They are all the same subsystem (`internal/session` + the driver's process
lifecycle) and they all touch `manager.go` / `lifecycle.go`. Splitting them would
mean two sessions editing the same 40 lines.

## Context budget

Read in full: `internal/session/manager.go` (240), `lifecycle.go` (349),
`watch.go` (38), `io.go` (97), `attach.go` (89). Read selectively:
`internal/driver/driver.go` (~250-300, 530-560). Plus the `internal/session`
test files (~900 lines total) for the call-site migration in step 1.
Roughly 2000 lines of reading, ~15 edited hunks. Comfortable for one session.

---

## Step 1 — Stop handing out shared `*Session` pointers (#6)

**Problem.** `GetSession` (`manager.go:156`) and `Sessions()` (`:147`) take
`RLock` only long enough to return the shared pointer; the caller then reads
fields that `trackOutput`/`watch` mutate concurrently. This is a *live, reproducing*
data race: `go test -race ./internal/session/...` fails today in
`TestSessionPendingLineCapped` (write `io.go:63` vs read `io_test.go:48`).

**Do this.**

1. Delete `Sessions()` entirely (`manager.go:147-153`) — it has no callers, in
   production or tests. Drop the now-unused `maps` import if nothing else uses it.
2. Change `GetSession` to return a value:
   `func (m *Manager) GetSession(id string) (Snapshot, bool)`, built under the
   lock via the same mapping used by `List()`. There are **no production callers**,
   so no behavior changes — this is purely making the test-facing API safe.
3. `Snapshot` (`manager.go:75-86`) is missing fields the tests read. Add
   `OriginalDir` and `LastOutput` to it (both are already tracked on `Session`;
   `OriginalDir` is even persisted, so its absence from `Snapshot` is a divergence
   noted in #9).
4. `internal/session/io_test.go:44,54` asserts on `len(s.PendingLine)`. Do **not**
   put a `PendingLine` byte-slice copy on `Snapshot` — it is an internal parsing
   buffer, not dashboard state. Instead add an explicit test affordance next to the
   existing one (`SetLastWordAt`, `manager.go:163`, already exists for exactly this
   purpose):
   ```go
   // PendingLineLen reports the buffered partial-line size. Used by tests to
   // assert the pendingLineLimit cap.
   func (m *Manager) PendingLineLen(id string) (int, bool)
   ```
   Take `m.mu.RLock()` and return `len(s.PendingLine)`.
5. Migrate the 16 call sites: `internal/session/{workspace_test.go:42,104,156,208,
   tmp_workspace_test.go:53,97,119,178, io_test.go:44,54, lifecycle_test.go:73,93,155}`
   and `internal/app/{preview_integration_test.go:115,177, integration_test.go:275}`.
   All of them only read `WorkingDir`, `Name`, `CardID`, `Status` or `PendingLine`,
   so each is a mechanical swap (`*Session` → `Snapshot`, plus the two
   `PendingLine` sites → `PendingLineLen`).

**Gate.** `go test -race ./internal/session/... ./internal/app/...` must pass
before moving on. Everything below depends on it.

## Step 2 — Generation guard so relaunch survives its predecessor (#3)

**Problem.** `watch` (`watch.go:9`) and `trackOutput` (`io.go:10`) are spawned per
launch (`lifecycle.go:240-241`, and again in `SwitchAgent` at `:142-143`) but keyed
**only by `sessionID`**, re-resolving the session on each iteration. Across a
Kill→Relaunch of the same ID — the ordinary `Enter`-on-terminated-session path, and
the path any future "apply new config" feature would use — the previous life's
goroutines can still fire and:

- `watch.go:22-27`: see the *old* driver report `StatusTerminated` and write
  `StatusTerminated` onto the freshly relaunched session, then
  `closeSubscribersLocked` — killing the new session's preview stream.
- `io.go:30-32`: observe the *old* output channel close and run
  `closeSubscribersLocked` against the new session.

Both write under `m.mu`, so `-race` will never report it; the user-visible symptom
is a session that flips to Terminated seconds after a successful relaunch.

**Do this.**

1. Add `generation uint64` to `Session` (`manager.go:30-55`) — unexported, mutated
   only under `m.mu`.
2. In `activateSessionLocked` (`lifecycle.go:247`), `s.generation++`. This is the
   single choke point for both `startSession` and `SwitchAgent`, so both paths get
   the bump for free.
3. Change both goroutine signatures to carry the generation they were started for:
   `go m.trackOutput(sid, gen, h.Output)` / `go m.watch(sid, gen, d)`. Capture
   `gen := s.generation` immediately after `activateSessionLocked` while still
   holding the lock (`lifecycle.go:240-241` and `:142-143`).
4. In both loops, after re-resolving `s` under the lock, add the ownership check
   right next to the existing `if !ok` guard:
   ```go
   if !ok || s.generation != gen {
       m.mu.Unlock()
       return
   }
   ```
   In `trackOutput` the same check must also guard the **post-loop**
   `closeSubscribersLocked` (`io.go:30-32`) — that is the half that closes the new
   session's subscribers when the old channel drains.
5. While in `watch.go`: it currently `return`s after setting `StatusTerminated`,
   which is correct; leave that.

**Test.** Add `internal/session/watch_test.go` (or extend it) with a
Kill→Relaunch cycle: launch with driver `cat`, `Kill`, `Relaunch`, then wait past
`watchInterval` (2s) and assert the status is still `StatusRunning` and a
subscriber channel obtained after the relaunch is still open. Without the guard
this test fails within ~2-4s; with it, it passes.

## Step 3 — `Kill` must not bail out on a benign kill error (#15)

**Problem.** `manager.Kill` (`lifecycle.go:295-305`) returns as soon as
`s.Driver.Kill()` errors — before setting `StatusTerminated` and persisting. When
the process has already exited, `driver.Kill` (`driver.go:286-300`) returns
`os: process already finished` from `Process.Kill()`, so a perfectly normal
"kill an agent that just finished" leaves the manager's status stale and surfaces a
spurious error to the user (`callbacks.go:38-46` renders it as `ErrorMsg`).
`Shutdown` (`lifecycle.go:340-348`) hits the same path for every session on exit.

**Do this.** In `driver.GenericCLIDriver.Kill`, treat an already-finished process as
success: check `errors.Is(err, os.ErrProcessDone)` and return `nil` (keep the
status/PTY/`closeDone` teardown, which already runs before the return). Fix it in
the driver rather than in the manager so `Shutdown` and any other caller benefit.
Then make `manager.Kill` set `StatusTerminated` and persist **even if** the driver
returns an error — record the error and return it *after* the state update, not
instead of it.

**Test.** `internal/session/lifecycle_test.go`: launch, wait for the process to
exit on its own (or kill twice), assert the second `Kill` returns `nil` and the
status is `StatusTerminated`.

## Step 4 — Capture the child's exit status (exit-code half of #9)

**Problem.** `watchProcess` (`driver.go:272-284`) does `_ = cmd.Wait()`, so
`Session.ExitCode` (`manager.go:43`) is never assigned — it is dead weight, and a
crashed agent is indistinguishable from a clean exit.

**Do this.**

1. In `watchProcess`, keep the `cmd.Wait()` error and store the exit code on the
   driver (new field, written under `d.mu` alongside `d.status = StatusTerminated`).
   Use `cmd.ProcessState.ExitCode()` when available (it is set by `Wait`), which
   returns `-1` for a signalled process.
2. Expose it: `func (d *GenericCLIDriver) ExitCode() int` under `RLock`. Add it to
   the `AgentDriver` interface only if you also update the test stub
   (`internal/driver/factory_test.go:32` region) — otherwise keep it off the
   interface and type-assert in `watch`, which is the only consumer.
3. In `watch.go`, when the driver reports `StatusTerminated`, set `s.ExitCode`
   before persisting.

**Do not** surface the exit code in the UI in this plan — persisting it and
rendering it are separate concerns; `Data`/`Snapshot` plumbing for it belongs to
Plan 2 (which owns the struct mapping). Just make it correct on the in-memory
`Session`.

## Step 5 — `mergeEnvWithPwd` must upsert, not append (#2)

**Problem.** `driver.go:537-543` does
`base = append(base, fmt.Sprintf("%s=%s", k, v))` onto `os.Environ()`. Duplicate
`KEY=` entries result, and `getenv(3)` returns the **first** match — the inherited
one. So a card's `Env` entry is silently ignored for any variable that already
exists in KANKURO's own environment (`HTTP_PROXY`, `ANTHROPIC_*`, `NO_COLOR`,
anything a shell profile exports). The correct helper, `upsertEnv`, sits at
`driver.go:545` and is already used for `PWD` on the next line.

**Do this.** Loop `extra` through `upsertEnv`:
```go
func mergeEnvWithPwd(extra map[string]string, pwd string) []string {
    base := os.Environ()
    for k, v := range extra {
        base = upsertEnv(base, k, v)
    }
    return upsertEnv(base, "PWD", pwd)
}
```
Iterate `extra` in sorted key order so the produced slice is deterministic and the
test can assert on it.

**Note on why this matters with the env feature cancelled:** `IDCard.Env` is not
speculative — it exists (`config/types.go:22`), is validated, and already reaches
`Spawn` from both `startSession` (`lifecycle.go:202`) and `SwitchAgent` (`:123`).
A card with a hand-written `env:` block in `id_cards.yaml` works today, except for
this bug (and except that the editor deletes it — Plan 6, finding #1).

**Test.** `internal/driver`: table test on `mergeEnvWithPwd` — (a) a fresh key is
appended, (b) a key already present in `os.Environ()` (use `t.Setenv` to plant one)
appears **exactly once** with the override's value, (c) `PWD` still wins.

---

## Verification

```
go build ./...
go test ./...
go test -race ./internal/session/... ./internal/app/...   # must be clean
golangci-lint run ./...                                    # expect only the 5 baseline issues
```

Manual smoke (`go run ./cmd/kankuro`): launch a session, `x` to kill it, `Enter` to
relaunch, then leave the dashboard open for ~10s — the session must stay Running
and the preview must keep streaming (guards step 2). Kill an agent that already
exited (let a short-lived driver finish, then press `x`) — no error banner
(step 3).

## Out of scope

- The `Session`/`Snapshot`/`Data` triplication itself and error propagation from
  `persistSessionLocked` → **Plan 2**.
- `Relaunch`/`SwitchAgent` read `s.Status` after releasing `RLock`
  (`lifecycle.go:69-77`, `:88-96`) — a TOCTOU that lets two concurrent relaunches
  both pass the guard. Not currently reachable from the single-threaded TUI
  (one keypress at a time); note it, leave it. If step 2's generation counter makes
  a fix trivial, taking it is acceptable.
- Attach-path PTY writes and `driver.SendInput` (#14) → **Plan 4**.
