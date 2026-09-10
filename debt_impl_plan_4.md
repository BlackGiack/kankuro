# Debt Plan 4 — Coordinator concurrency & UI-thread discipline

> Covers **#4** (concurrent map access can panic), **#12** (`waitForOutput`
> goroutine leak), **#10** (`View()` side effect + blocking work in `Update`),
> **#14** (dead `SendInput` vs. unlocked PTY write), the `internal/app` half of
> **#8** (swallowed errors), and the Low item "synchronous `DeleteMCP` inside
> `Update`".
>
> **Best run after Plan 1** so that `go test -race ./internal/app/...` is a
> trustworthy gate (it passes today, but Plan 1's session fixes remove the noise
> from the shared-pointer race). Otherwise independent.

## Why these ship together

Every item is the same invariant: **Bubble Tea's `Update`/`View` run on one
goroutine; `tea.Cmd` closures do not.** #4 and #12 are violations in the Cmd
direction (async code touching coordinator state), #10 and the `DeleteMCP` item are
violations in the other (blocking/side-effecting work on the UI goroutine). Fixing
them piecemeal means re-deriving the same rule four times.

## Context budget

`internal/app/coordinator.go` (~440), `callbacks.go` (~540, but only the dashboard
and MCP callback blocks matter), `attach.go` (~280), plus
`internal/ui/views/dashboard.go` around `:400-430` and `:780-800`, and
`internal/ui/messages/messages.go` (134) to add message types. ~1200 lines of
reading with several structural edits. This is the densest of the seven plans —
don't add scope to it.

---

## Step 1 — Stop the async closure from touching `c.views` (#4)

**Problem.** `refreshData` (`coordinator.go:349-379`) returns a `tea.Cmd` — a
closure Bubble Tea runs on **its own goroutine** — and inside it:
```go
if dashboard, ok := c.views[ViewDashboard].(*views.Dashboard); ok {
    previewID = dashboard.PreviewSessionID()
```
It reads the `c.views` map and calls into a view's method. Meanwhile the main
`Update` loop writes that same map at `coordinator.go:176`
(`c.views[c.currentView] = updated`). Concurrent map read/write is a **hard Go
runtime throw** ("concurrent map read and map write") that no `recover` in Bubble
Tea will save — the app dies. And the 2s tick (`:189`) re-arms this closure
forever, so the window is permanently open. `PreviewSessionID()` also reads
dashboard fields the main goroutine mutates.

**Do this.** Snapshot on the main goroutine, capture by value:

```go
func (c *Coordinator) refreshData() tea.Cmd {
    previewID := c.currentPreviewSessionID()   // runs NOW, on the Update goroutine
    return func() tea.Msg {
        ...
        var scrollback []byte
        if previewID != "" { scrollback = c.session.GetScrollback(previewID) }
        ...
    }
}
```
`refreshData` is called from `Init` (`:155`), from the tick handler (`:189`), from
`syncPuppetsAsync`'s closure (`:165`), and from ~10 callbacks. **Audit every call
site**: the `syncPuppetsAsync` one (`coordinator.go:162-167`) is itself already
inside a `tea.Cmd`, so the "runs now" part would run on the wrong goroutine there.
Fix that site by having `syncPuppetsAsync` return a plain
`messages.PuppetsSyncedMsg` and letting `Update` fan out to `refreshData()` in
response — that keeps the invariant honest instead of special-casing it.

The callbacks in `callbacks.go` that end with `return c.refreshData()()` (double
call — building the Cmd and immediately invoking it, e.g. `:44`, `:53`, `:274`,
`:369`, `:415`) are also inside `tea.Cmd` closures and hit the same problem. The
cleanest fix for all of them at once: have those closures return a
`messages.DataRefreshRequestedMsg{}` and handle that in `handleSpecialMessages` by
returning `c.refreshData()`. That single message type removes the whole class.

Note `c.session` (the manager) **is** safe to call from the closure — it is
mutex-protected. Only `c.views`, `c.previewSessionID`, `c.previewSubChan`,
`c.currentView` and `c.err` are coordinator-goroutine state.

**Test.** `internal/app`: drive the coordinator with a tick and a view-swapping
`NavigateMsg` in a tight loop from the test goroutine while running the returned
Cmds concurrently, under `-race`. Even without a detector hit, assert structurally
that `refreshData`'s closure no longer references `c.views` (a small
`go/ast`-free check is overkill — a code comment plus the behavioral test is
enough; the real guard is the review).

## Step 2 — `waitForOutput` must never block on a nil channel (#12)

**Problem.** `navigateTo` (`coordinator.go:325-331`) unsubscribes and sets
`c.previewSubChan = nil` but **leaves `c.previewSessionID` unchanged**. A
`SessionOutputMsg` already in Bubble Tea's queue then still matches
`m.SessionID == c.previewSessionID` (`:264`) and re-arms
`c.waitForOutput(m.SessionID, c.previewSubChan)` — with a nil channel.
`<-ch` on a nil channel blocks **forever**: one permanently leaked goroutine per
navigation-with-queued-output. Repeated across a session, this is an unbounded
leak.

**Do this.** Belt and braces, both are one line:
1. Guard the producer: in the `SessionOutputMsg` branch (`:263-273`), require
   `c.previewSubChan != nil` in the condition.
2. Guard the primitive: `waitForOutput` (`:416-427`) returns `nil` immediately if
   `ch == nil`.
3. Clear `c.previewSessionID = ""` in `navigateTo` alongside the unsubscribe, so
   the stale-match condition cannot arise at all.

**Test.** `internal/app`: subscribe a preview, navigate away, then feed a
`SessionOutputMsg` for the old session and assert the returned `tea.Cmd` is nil.
(`internal/app/preview_integration_test.go` already has the scaffolding for
preview subscriptions — extend it.)

## Step 3 — Get side effects and blocking work off the UI goroutine (#10)

**Problem A.** `views/dashboard.go:785` — `View()` starts with
`fmt.Print("\x1b[=0u")`. `View()` is a pure-render function Bubble Tea may call
arbitrarily often; printing directly to stdout from it fights the renderer's own
output, and it makes every dashboard render test do I/O. The `//nolint:forbidigo`
comment acknowledges the smell.

**Do this.** The sequence disables the Kitty keyboard protocol when control returns
to KANKURO. That is a *transition*, not a render: emit it once from
`Dashboard.Init()` (which already exists, `dashboard.go` after `Card()`), and again
on the attach→detach transition where it is actually needed
(`internal/app/attach.go`'s restore path, which is already doing raw terminal
work). Prefer the framework's own facility if available in `charm.land/bubbletea/v2`
(a `tea.Cmd` that writes a sequence) over a bare `fmt.Print`; check what the v2 API
offers before hand-rolling. Then drop the `//nolint:forbidigo`.

**Problem B.** Two dashboard callbacks run blocking work synchronously inside
`Update`:
- `OnLaunchIDE` (`dashboard.go:407-429` → `callbacks.go:171`ff) **spawns an
  editor process** and returns `error`.
- `OnSwitch` (`callbacks.go:157-166`) calls `c.session.SwitchAgent`, which syncs
  driver files to disk **and spawns a PTY process**, then returns.

Both freeze the entire TUI for the duration — visibly, on a cold filesystem or a
slow editor.

**Do this.** Change both callback signatures to return `tea.Cmd` only (drop the
`error` return) and move the work into the closure, reporting completion via a new
message (`messages.IDELaunchedMsg{Err error}`,
`messages.AgentSwitchedMsg{SessionID string, Err error}`) handled in
`handleSpecialMessages`. The dashboard's `handleLaunchIDE`/`handleSwitch` then just
return the Cmd. `OnKill`/`OnRemove` (`callbacks.go:38-56`) already demonstrate the
pattern — follow it exactly.

**Problem C (Low).** `coordinator.go:229` calls `c.repo.DeleteMCP(m.MCPID)`
directly in the `MCPDeleteRequestedMsg` branch — synchronous disk I/O inside
`Update`, inconsistent with every other mutation. Wrap it in the `tea.Cmd` that
already follows it (the branch already returns a Cmd emitting `MCPDeletedMsg`;
move the delete inside and carry the error).

## Step 4 — Resolve the input-path duplication (#14)

**Problem.** `driver.SendInput` (`driver.go:493-502`) is on the `AgentDriver`
interface, is correctly `RLock`-protected, and has **zero callers** — only a test
stub implements it. The path input actually takes is the raw `*os.File` PTY handed
out by `Attach` (`session/attach.go:26`, `Input: s.Input`) and written directly at
`app/attach.go:249,258`, holding no lock. `Kill`/`watchProcess`
(`driver.go:286-300`, `:272-284`) close that same PTY and nil the field under
`d.mu` — so a keystroke landing during a kill can hit a closed FD and be silently
dropped (the `forwardInput` error is returned, but the byte is gone).

**Do this — pick one and make the code say which:**
- **Recommended:** route attach writes through the locked driver method. `Attach`
  would hand back a small `io.Writer` adapter that calls `driver.SendInput` (or a
  `[]byte` variant — `SendInput(string)` forcing a `string` conversion on every
  keystroke buffer is wasteful for a hot path, so consider
  `SendInputBytes([]byte) (int, error)` and have `SendInput` wrap it). `attach.go`
  keeps needing the real `*os.File` for `ioctl`-based resize (`PTY` field), so keep
  `PTY` for resize and use the adapter only for `Input`. Verify the resize path
  still works — that FD is nil-checked there already.
- **Alternative:** delete `SendInput` from the interface and the implementation,
  and document on `AttachHandle.Input` that it is a raw PTY whose lifetime the
  driver owns, with the race accepted.

Do not leave both paths in place with neither documented — that ambiguity is the
finding.

**Test.** `internal/session`: attach, then `Kill` concurrently with a write, under
`-race`; assert no panic and that a post-kill write returns an error rather than
succeeding into a closed FD.

## Step 5 — Surface `refreshData`'s discarded errors (#8, app half)

**Problem.** `refreshData` (`coordinator.go:351-357`) discards **seven** read
errors in a row (`cards, _ :=`, `skills, _ :=`, …). One corrupt YAML file therefore
renders as an *empty list*, which (a) looks like data loss to the user and (b) is
actively dangerous: with the list views now empty, the next save writes the empty
view back. Plan 3 step 3 closes the write half; this is the visibility half.

**Do this.** Join the errors (`errors.Join`) inside the closure and carry them on
`messages.DataLoadedMsg` as a new `Err error` field; in the handler, set `c.err`
when non-nil (the coordinator already renders `c.err` above the view,
`coordinator.go:302-305`). Crucially: when a read fails, **do not** broadcast the
empty slice for that registry — leave the previous value in place, so a transient
failure cannot blank the UI. Reuse the logging facility Plan 2 step 3 established
rather than inventing a second one.

**Test.** `internal/app`: corrupt one registry file, run a refresh, assert an error
reaches `c.err` and the previously loaded cards are still rendered.

---

## Verification

```
go build ./...
go test ./...
go test -race ./internal/app/... ./internal/ui/...
golangci-lint run ./...
```

Manual smoke (`go run ./cmd/kankuro`) — this plan's changes are mostly invisible to
tests, so actually run these:
1. Launch a session, attach, type, `Ctrl+K` to detach — keystrokes must land and
   the dashboard must repaint cleanly with no stray escape sequences (step 3A).
2. With a preview open, navigate to the card editor and back several times, then
   check the goroutine count is stable (`SIGQUIT` dump, or add a temporary
   `runtime.NumGoroutine()` to the footer while testing) (step 2).
3. Press `o` (launch IDE) and `s` (switch agent) — the TUI must stay responsive
   while the work happens (step 3B).
4. Leave the app idle for a few minutes — no panic (step 1). The #4 panic is
   probabilistic; the test suite under `-race` is the real gate.

## Out of scope

- Reducing the 2s tick's frequency or caching config reads → **Plan 5**. This plan
  makes the tick's refresh *safe*, not *cheap*, and deliberately leaves the cadence
  alone so the two plans don't collide in `refreshData`.
- Anything in `internal/ui/views/card_editor.go` → **Plan 6**.
- The `fmt.Print` cousins elsewhere in the UI, if any turn up — note them for
  Plan 7.
