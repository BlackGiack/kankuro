# Act 2 — Plan 4: Split the coordinator (with §3a folded in)

Covers `tech_debt_act_2.md` **§2** (the coordinator is the architectural hotspot) and
**§3a** (`navigateTo` clears the error banner unconditionally). Plan 4 of 4; see the
table in `impl_plan_1.md`. Do it last, and **finish it before Phase 1 of the roadmap
starts**.

This is a refactor. Apart from step 1 (§3a), **no behaviour changes**. Everything lives
in `internal/app`, which is where Act 1's parallel sessions collided, so **don't run
this plan in parallel with anything else that touches `internal/app`**.

**Rules for this session:** don't commit unless asked, but run all five gates after
every step below, not just at the end. Each step is one move, so if a gate goes red you
know which move did it. If the user agrees, checkpoint with a commit between steps.

---

## Current state

`coordinator.go` (815 lines) + `callbacks.go` (610) + `attach.go` (292) are all methods
on `*Coordinator`. The `Coordinator` struct holds every concern's state together:

| Concern | State | Functions |
|---|---|---|
| navigation / views | `views`, `currentView`, `prevView`, `width`, `height` | `navigateTo`, `resizeAll`, `updateCurrentView`, `currentViewView`, `broadcastData`, `broadcastPreview`, `convertViewType`, `initViews` |
| error banner | `err` | set in ~12 handlers, cleared in `navigateTo` and `ClearErrorMsg`, rendered in `View` |
| preview subscription | `previewSessionID`, `previewSubChan` | `handlePreviewRequest`, `waitForOutput`, `maybeSubscribeAutoSelectedPreview`, the `SessionOutputMsg` case, the unsubscribe inside `navigateTo` |
| data refresh | `configFingerprint` | `refreshData`, `refreshDataSeeded`, `refreshSessions`, `maybeRefreshConfig`, `loadFullData`, `loadConfigData`, `fingerprintsEqual`, `configFingerprintMsg`, `tick`/`tickMsg`, `handleConfigFingerprint`, `handleDataLoaded`, `handleSessionsRefreshed` |
| session actions | none (stateless; uses `repo`, `session`) | `launchCard`, `relaunchSession`, `switchAgentAsync`, `resolveCard`, `resolveCapabilities`, `cardBundle`, `toIDSet`, `deleteCard`, `launchIDE`, `handleAgentSwitched`, the attach/launch cases of `handleSessionMessages` |
| settings | `settings` | `SettingsSavedMsg` case |

**Invariant 1** (no `tea.Cmd` closure touches coordinator state) is currently enforced
only by a grep and by comments. The main structural goal of this split is to **make it
hold by construction for the code that builds closures**. The refresh and session-action
helpers become types that *have no pointer to the coordinator or its views*, so a
closure they build has nothing to reach. The compiler then checks what the grep used to.

### The test surface that must not move

External tests (package `app_test`) use only: `NewCoordinator`, `Update`, `View`,
`Init`, `CurrentView`, `PrevView`, `Views`, `NavigateTo`, `Width`, `Height`,
`ResizeAll`, `RefreshData`, `Err`, `DashboardCallbacks`, `SourceManagerCallbacks`,
`AgentSwitchCallbacks`, and `app.New` / `App`. **Keep every one of these with the same
signature and behaviour.** The external tests then pass unchanged after each step, and
that is the evidence the refactor preserved behaviour.

`refresh_cadence_internal_test.go` (package `app`) reaches into internals:
`c.configFingerprint`, `c.maybeRefreshConfig`, `c.refreshDataSeeded`,
`c.refreshSessions`, `c.resolveCard`, `c.resolveCapabilities`, `c.views`. It will need
mechanical renames as things move. That's expected. Change **only** names in it, never
assertions, and say so in the step that does it.

---

## Step 1 — §3a: the banner clears when the user acts, not on navigation

Do this **first**, before any code moves, so the one behaviour change in this plan is
reviewed by itself.

**Change.**
- In `Coordinator.Update`, before `handleSpecialMessages`: if `msg` is a
  `tea.KeyPressMsg`, clear `c.err`. The banner then disappears at the user's next
  keystroke, which is the "clear when the user acts" option §3a recommends.
- Remove `c.err = nil` from `navigateTo`.
- Keep `ClearErrorMsg` as is. Nothing sends it today (only the coordinator's handler
  references it), so it's effectively an API for later. Don't delete it in a refactor
  plan.

**Why this ordering is safe.** Every error reaches the coordinator as a *later* message
(`ErrorMsg`, `CardSavedMsg{Err}`, …), never inside the same `Update` as the keypress
that caused it. Clearing on the keypress and then setting on the result therefore shows
the new error, not the old one. Every user-initiated navigation starts with a keypress,
so a stale banner still never follows the user to a new screen. Programmatic navigations
(after a save) follow a keypress too.

**Simplify what the old trap forced.** Remove the "Set after navigating: navigateTo
unconditionally clears c.err" ordering in the `SessionAttachRequestedMsg` case, and the
explanatory half of `handleAgentSwitched`'s comment about the batched-warning race.
Leave the code paths in place: routing the warning through
`SessionAttachRequestedMsg.Warning` is still correct, just no longer load-bearing. Update
the comments so they stay true (`project.md` §9).

**Tests** (new `internal/app/error_banner_test.go`, package `app_test`, driven through
`Update`):
1. **`TestErrorSurvivesNavigationInEitherOrder`:** `Update(ErrorMsg)` then
   `Update(NavigateMsg)`, and the reverse. In both cases `coord.Err() != nil` and
   `View()` contains the error text. *Fails on the old code* in the
   error-then-navigate order. Confirm it does before changing anything.
2. **`TestKeyPressClearsBanner`:** set an error, send any `tea.KeyPressMsg`, and
   `Err() == nil`.
3. **`TestKeyPressThenErrorShowsNewError`:** a keypress, then an `ErrorMsg`, and the
   banner shows it.
4. Re-run `session_warning_test.go`: the agent-switch persist-warning path must still
   end with the warning visible.

Update `project.md` §6 invariant 4 (rewrite it: "the banner is cleared by the next
keypress or `ClearErrorMsg`; navigation leaves it alone") and §11 (remove the
"`navigateTo` clears the error banner" bullet).

## Step 2 — `refresher`: data refresh with no coordinator pointer

New file `internal/app/refresh.go`:

```go
// refresher builds the commands that read config and session state off the
// UI goroutine. It deliberately holds no reference to the Coordinator or its
// views: every closure it returns can only reach repo, sessions and the
// values passed in as arguments, so invariant 1 (no tea.Cmd closure touches
// coordinator state) holds here by construction.
type refresher struct {
	repo     config.Repository
	sessions *session.Manager
}

func (r refresher) full(previewID string) tea.Cmd                    // was refreshData
func (r refresher) fullSeeded(previewID string) tea.Cmd              // was refreshDataSeeded
func (r refresher) sessionsOnly(previewID string) tea.Cmd            // was refreshSessions
func (r refresher) configIfChanged(prev map[string]config.FileStamp) tea.Cmd // was maybeRefreshConfig
```

Move `loadFullData`, `loadConfigData`, `fingerprintsEqual`, `configFingerprintMsg`,
`tick` and `tickMsg` with it, along with their doc comments.

The coordinator keeps the fingerprint *baseline* (`configFingerprint`), because it's
coordinator-goroutine state, and keeps thin wrappers that snapshot first and then
delegate:

```go
func (c *Coordinator) refreshDataSeeded() tea.Cmd {
	return c.refresh.fullSeeded(c.currentPreviewSessionID())
}
```

That wrapper is the seam where invariant 1 matters: "snapshot on this goroutine, hand
values to the refresher". Put a short comment there pointing at invariant 1.

`handleConfigFingerprint`, `handleDataLoaded` and `handleSessionsRefreshed` stay on the
coordinator for now (they touch views and the preview). Step 4 moves their preview half.

Update `refresh_cadence_internal_test.go`: rename calls only.

## Step 3 — `sessionActions`: launch / relaunch / switch with no coordinator pointer

New file `internal/app/session_actions.go`, built the same way:

```go
// sessionActions runs blocking session work (capability resolution, spawn)
// inside tea.Cmd closures. Like refresher, it has no Coordinator pointer.
type sessionActions struct {
	repo     config.Repository
	sessions *session.Manager
}
```

Move into it: `cardBundle`, `resolveCard`, `resolveCapabilities`, `toIDSet`,
`switchAgentAsync` (→ `switchAgent`), the closure body of `relaunchSession`
(→ `relaunch(id) tea.Cmd`), the launch closure inside `launchCard`
(→ `launch(card) tea.Cmd`), `deleteCard`, and the closure body of `launchIDE` if it fits
the same shape.

What **stays** in `callbacks.go`: the part of `launchCard` that touches
`c.views[ViewSessionLauncher]` (`Reset`, `SetCard`). That is coordinator-goroutine work
and runs before the closure. Keep the existing comment on `agentSwitchCallbacks`
explaining why `OnSwitch` returns a command rather than switching synchronously, and
reword it to point at `sessionActions`.

After this step, `grep -n "c\.views\|c\.err\|c\.currentView\|c\.preview" internal/app/refresh.go internal/app/session_actions.go`
must be empty. It will be, because those files have no `c`.

## Step 4 — `previewSub`: the preview subscription

New file `internal/app/preview.go`:

```go
// previewSub owns the dashboard preview's live-output subscription. It is
// coordinator-goroutine state: only call its methods from Update/handlers,
// never from inside a tea.Cmd closure. waitForOutput is the one method that
// returns a closure, and that closure captures only the channel and ID.
type previewSub struct {
	sessions  *session.Manager
	sessionID string
	ch        chan []byte
}

func (p *previewSub) subscribe(id string) (scrollback []byte, cmd tea.Cmd) // was handlePreviewRequest's state half
func (p *previewSub) unsubscribe()                                         // was the navigateTo block
func (p *previewSub) accepts(m messages.SessionOutputMsg) bool             // the SessionOutputMsg guard
func (p *previewSub) next() tea.Cmd                                         // waitForOutput on the current ch
```

`previewSub` returns data and never broadcasts. The coordinator still calls
`broadcastPreview` with what `subscribe` returns, so `previewSub` needs no access to
views. Move the comment explaining why `navigateTo` clears `previewSessionID` together
with the unsubscribe (debt_impl_plan_4.md #12/#17) onto `unsubscribe`, and keep the call
in `navigateTo`.

`maybeSubscribeAutoSelectedPreview` stays on the coordinator (it reads the dashboard)
and calls `c.preview.subscribe`.

## Step 5 — navigation, views and the banner

New file `internal/app/navigation.go`: move `navigateTo`, `resizeAll`,
`updateCurrentView`, `currentViewView`, `broadcastData`, `broadcastPreview`,
`currentPreviewSessionID`, `convertViewType`, `heightAdjustment`, and `initViews`.
These stay as **methods on `*Coordinator`** in their own file: they are the
coordinator's core state, and wrapping them in another type would only add indirection.

Put invariant 2 (never navigate from inside a view's `Update`) as the doc comment on
`navigateTo` itself, with the agent-switch regression as the example. Right now it lives
only in the `SessionAttachRequestedMsg` case and in `callbacks.go`.

The banner: after step 1 it's a field, one clear and many sets. Leave it as `c.err`. A
type with one field and two methods isn't worth it, and §3a is already fixed.

## Step 6 — what is left in `coordinator.go`

`coordinator.go` should now hold: `ViewType` and its constants, the `Coordinator` struct
(now with `refresh refresher`, `actions sessionActions`, `preview previewSub` fields),
`NewCoordinator`, the exported accessors, `Init`, `Update`, `handleSpecialMessages`,
`handleSessionMessages`, the `handle*` functions that combine several concerns, and
`View`. Target: **under ~350 lines**, mostly routing.

Re-read `handleSpecialMessages` / `handleSessionMessages` once more. They should now be
a switch of one-to-three-line cases, each delegating. A case still doing real work
(for example `SettingsSavedMsg` reaching into the card editor) is fine to leave, but
each such case should be obvious at a glance.

Then reorder the struct fields by concern, each group with a one-line comment saying
which goroutine may touch it:

```go
type Coordinator struct {
	// --- coordinator-goroutine state: never read or written from a tea.Cmd closure ---
	currentView, prevView ViewType
	views                 map[ViewType]component.Model
	width, height         int
	err                   error
	settings              config.AppSettings
	preview               previewSub
	configFingerprint     map[string]config.FileStamp

	// --- safe from any goroutine ---
	repo    config.Repository
	session *session.Manager
	refresh refresher
	actions sessionActions
}
```

## Step 7 — a test that catches invariant 1 violations

Today a violation is found by grep. Add a race test that finds one by running the code:

**`TestCommandClosuresDoNotRaceCoordinator`** (`internal/app`, run under `-race`): build a
coordinator with a few cards and a running `cat` session. Then, for a list of messages
that each return commands (`tickMsg`, `DataRefreshRequestedMsg`, the dashboard's
launch key, `SessionPreviewRequestedMsg`, `AgentSwitchRequestedMsg`, …), call `Update` on
the test goroutine, **execute every returned command on its own goroutine** (unwrap
`tea.BatchMsg`, as `collectMsgs` does), and keep calling `Update(tea.WindowSizeMsg{})` and
`View()` on the test goroutine while they run. Any closure that reads `c.views` /
`c.previewSessionID` / `c.currentView` / `c.err` is then a data race the detector reports.

*Can-fail check (mandatory):* temporarily put `_ = c.views[ViewDashboard]` inside one
closure in `callbacks.go` and confirm `go test -race` fails. Without that check the test
proves nothing.

Keep the grep as well; it's cheap. Add both to `project.md` §6 invariant 1.

## Step 8 — docs

- `project.md` §3 repository map, the `internal/app` block: list `coordinator.go`
  (routing), `navigation.go`, `refresh.go`, `preview.go`, `session_actions.go`,
  `callbacks.go` (wiring), `attach.go`.
- `project.md` §6 invariant 1: "`refresher` and `sessionActions` hold no coordinator
  pointer, so their closures cannot violate this; the risk now lives only in
  `callbacks.go` and the few `waitForOutput`-style closures", plus the race test.
  Invariant 2: point at `navigateTo`'s doc comment.
- `project.md` §10: "Change refresh behaviour" → `refresh.go`; add a row "Add a
  session-level action (Phase 1 linking)" → `session_actions.go` + a handler in
  `coordinator.go`.
- `tech_debt_act_2.md` §2 and §3a: a one-line "done in plan 4" status each.

---

## Done when

- [ ] Step 1's banner tests pass, and test 1 failed on the old code.
- [ ] `refresh.go`, `session_actions.go`, `preview.go`, `navigation.go` exist; `coordinator.go` is under ~350 lines.
- [ ] `refresher` and `sessionActions` have no `*Coordinator` field; the grep in step 3 is empty.
- [ ] Every external `app_test` test passes **unchanged**; `refresh_cadence_internal_test.go` has changed only in names.
- [ ] The race test exists and was seen to fail with a planted violation.
- [ ] `grep -n "c.refreshData()" internal/app/callbacks.go` is still empty.
- [ ] `project.md` and `tech_debt_act_2.md` updated.
- [ ] All five gates clean after every step, not only at the end.
