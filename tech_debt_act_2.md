# KANKURO — Technical Debt, Act 2

Follow-up to `tech_debt_founding.md`. All seven plans from that document are
landed on `feat/envvar`: the silent-data-loss paths (#1, #5, #7, #17), the
crash/race findings (#3, #4, #6) and the swallowed-error findings are closed,
`go test -race ./internal/...` is clean, and `golangci-lint run` exits 0.

What is below is what remains after that work — items explicitly left out of
scope by one of the plans, sharp edges noticed during the reviews, and one
testing gap exposed by real use. None of it loses data silently or crashes the
app today; it is ordinary maintenance debt, ranked by how much it will cost to
leave it.

Line references are as of the commit this was written against; symbol names
are the stable reference.

---

## Context for a fresh session

Act 1 was reviewed round by round in one long session. This section carries
over what that session learned, so it does not have to be rediscovered.

### Working rules

- **Never commit unless explicitly asked** (`AGENTS.md`).
- **When running the app by hand**, point `KANKURO_TMP_DIR` and
  `KANKURO_CONFIG_DIR` at a throwaway directory (`kankuro_specs.md` §5.1).
  The real config at `~/.config/kankuro/` is fine to *read* when reproducing a
  user-reported bug — resolving the reported card's skill IDs to names there is
  how the §5 bug was reproduced exactly — but never write to it.
- **The gates** are `go build ./...`, `go vet ./...`, `go test ./...`,
  `go test -race ./internal/...` and `golangci-lint run`. All five are clean;
  lint exits 0 (the `driver` package-name warning is excluded in
  `.golangci.yml` on purpose, with a comment). Keep all five clean.
- **Work one item at a time, or check merges first.** Act 1 ran plans in
  parallel sessions and paid for it: two of the three branch conflicts landed
  in `internal/app`. If items do run in parallel, `git merge-tree` the branches
  before merging, and prefer cherry-picking a single commit onto the other
  branch over stash/pop — a stash has no merge base, and file deletions make
  stash pops go wrong.

### Invariants established in Act 1 — do not break these

They matter most for §2 (the coordinator split), where they are easiest to lose.

1. **No `tea.Cmd` closure touches coordinator state.** Bubble Tea runs a
   command's closure on its own goroutine, so a closure must never read or
   write `c.views`, `c.previewSessionID`, `c.previewSubChan`, `c.currentView`
   or `c.err`. `refreshData()` snapshots what it needs *before* returning its
   closure. Code already inside a closure asks for a refresh with
   `messages.DataRefreshRequestedMsg{}` (or `refreshRequestedCmd()`), never by
   building `refreshData()` there. Check with
   `grep -n "c.refreshData()" internal/app/callbacks.go` → must be empty. The
   test suite does **not** catch a violation; the grep does.
2. **Never navigate from inside a view's `Update`.** `Coordinator.Update` writes
   the view returned by `updateCurrentView` back into
   `c.views[c.currentView]` *after* the view's `Update` returns. If a callback
   running inside that `Update` calls `navigateTo`, `c.currentView` has already
   changed and the old view is stored under the new view's key — Act 1 did
   exactly this once and replaced the dashboard with the agent-switch view.
   Navigate from a message handler in `handleSpecialMessages` /
   `handleSessionMessages`, which return early and never write `c.views`.
3. **`DataLoadedMsg` is always a full snapshot.** A `nil` field means "that
   registry could not be read this round, keep what you have" — never "this
   update does not concern that field". Partial updates have their own types:
   `PreviewDataMsg` for preview scrollback, `SessionsRefreshedMsg` for the
   sessions-only tick. Every `SetData` must guard each field with
   `if msg.X != nil`. This matters because `CapabilityList.SelectedIDs()` only
   returns IDs whose items are currently loaded: wipe the items and a save
   silently writes the card with no skills. That was finding #17.
4. **The error banner is cleared by `navigateTo`** (§3a). Until that changes,
   never batch an `ErrorMsg` beside a navigation; set `c.err` in the handler
   *after* navigating.

### How to test `internal/app` without fooling yourself

The coordinator is tested by driving it: `app.NewCoordinator(repo, sm)`, then
`coord.Update(msg)`, then inspect `coord.View()`, `coord.CurrentView()` and
`coord.Views()`. Traps met in Act 1:

- **Go through `Update`, not the callback.** Calling
  `coord.AgentSwitchCallbacks().OnSwitch(...)` directly skips the write-back in
  `Coordinator.Update` — which is exactly where invariant 2 breaks — so a test
  written that way cannot see that bug.
- **Assert that the operation actually ran** before asserting its effects.
  The agent-switch view needs a refresh *after* navigating to it or its list is
  empty, `OnSwitch` never fires, and the test passes for the wrong reason.
- **`collectMsgs`** unwraps `tea.BatchMsg` only; commands built with
  `tea.Sequence` come back opaque. There are two copies with that same
  behaviour — `session_warning_test.go` (package `app_test`) and
  `refresh_cadence_internal_test.go` (package `app`) — so change both if you
  change one.
- **Read the real files, not a filtered diff.** Grepping only a diff's `+` lines
  joins unrelated hunks together and invents bugs that are not there — three
  false alarms in one Act 1 review came from that.

For §5, `card_render_layout_test.go` already has reusable helpers:
`assertIntactCard`, `assertSinglePillRow`, `builtinPuppets` and
`cardWithSkills`.

### History, if you need it

The Act 1 reasoning is in the repository root: `tech_debt_founding.md` (all
17 findings with severity and status), `debt_impl_plan_1.md` … `_7.md` (what
each plan set out to do and why), and `review_dp_*.md` (what each review round
found, including the traps above).

---

## 1. Session records are not written atomically [medium]

`internal/session/storage.go:75-84` (`saveSession`)

Plan 3 made every config write atomic (`WriteAtomic`: unique temp file in the
same directory, `fsync`, rename, cleanup on error). Session records never got
the same treatment — `saveSession` still ends in a plain
`os.WriteFile(path, data, 0o600)`. A crash or power loss mid-write can leave a
truncated `<id>.json`.

Since Plan 2 that no longer disappears silently — `loadSessions` reports the
corrupt file through `LoadErrors()` and the coordinator surfaces it at startup —
but the session record itself is lost: its name, working directory and, for a
temp-workspace session, the only pointer to the workspace copy on disk.

**Exposure.** Session JSON is written on launch, relaunch, agent switch, rename,
kill and watcher-detected termination (`manager.go` `SetName`, `lifecycle.go`
`startSession` / `SwitchAgent` / `Kill`, `watch.go`). It is *not* written on the
Running↔Waiting flips, so the write rate is modest — the risk is a crash landing
during one of those writes, not a constant stream of them.

**Fix.** Reuse the Plan 3 recipe rather than inventing a second one. The
simplest route is to factor the temp-file/fsync/rename core of
`config.WriteAtomic` into a byte-level helper both packages can call (it is YAML
in `config`, JSON here — only the marshalling differs). Test the same way Plan 3
did: no `*.tmp` residue after a write, correct file mode, and concurrent writers
never producing a spliced file. Plan 3's review deliberately excluded this
because session writes needed their own thought about fsync cost; at the
measured write rate above, a per-write fsync is not a concern.

## 2. The coordinator is the architectural hotspot [structural]

**Status: done in plan 4.** `coordinator.go` is down to ~350 lines of routing; the rest is
`navigation.go`, `refresh.go`, `preview.go`, `session_actions.go` and `handlers.go`.
`refresher` and `sessionActions` hold no `*Coordinator`, and `TestCommandClosuresDoNotRaceCoordinator`
checks invariant 1 under `-race`. The text below is the original analysis.

`internal/app/coordinator.go` (~815 lines, 44 functions), with
`internal/app/callbacks.go` (~610) as its other half

The three largest files are `card_editor.go` (~1066), `dashboard.go` (~1018)
and `coordinator.go` (~815), but only the coordinator is a structural problem.
The two views are large because they render a lot; the coordinator is large
because it is where every cross-cutting concern meets — navigation, the error
banner, preview subscriptions, config refresh and fingerprinting, session
messages, attach, agent switching, and the callback wiring for every view.

The evidence is in how Act 1 went: nearly every plan edited it, and two of the
three merge conflicts between parallel plans landed in `internal/app`. The next
roadmap phase in `kankuro_specs.md` — inter-session communication and linking
agents in a chain — will land squarely in the same place.

**Recommendation.** Split it *before* Phase 1 starts, not during. The seams are
already visible in the handler names:

- **preview** — `handlePreviewRequest`, `broadcastPreview`, `waitForOutput`,
  `maybeSubscribeAutoSelectedPreview`, the `previewSessionID`/`previewSubChan`
  state;
- **data refresh** — `refreshData`, `refreshSessions`, `maybeRefreshConfig`,
  the fingerprint baseline, `handleDataLoaded`, `handleSessionsRefreshed`;
- **navigation and the error banner** — `navigateTo`, `c.err` (see 3a);
- **session actions** — attach, relaunch, switch, and their messages.

Each can become a small type owned by the coordinator, which then only routes
messages. Keep the invariants Act 1 established while doing it — in particular
that no `tea.Cmd` closure ever touches coordinator-goroutine state
(`c.views`, `c.previewSessionID`, `c.currentView`, `c.err`). A split is exactly
when that invariant gets lost, so it is worth a comment at each seam.

## 3. Sharp edges noticed during the reviews [low, each]

Three things that were seen, understood, and deliberately left alone because no
plan owned them. Each is small; each has either already caused a bug or is the
same shape as one that did.

### 3a. `navigateTo` clears the error banner unconditionally

**Status: done in plan 4.** Navigation leaves the banner alone; `Coordinator.Update` clears it
on the next keypress (`ClearErrorMsg` still works). `error_banner_test.go` covers it.

`internal/app/coordinator.go:495-509`

`navigateTo` ends with `c.err = nil`. That is a reasonable default — a stale
error should not follow you to a new view — but it means any error or warning
that is batched alongside a navigation is lost whenever the navigation happens
to be processed second. That already happened once: during Plan 2 the
persist-failure warning on the agent-switch path was dropped this way, and the
first attempt at fixing it went on to overwrite the dashboard view.

**Fix.** Make the clear explicit rather than a side effect: clear the banner
when the user *acts* (a keypress, a `ClearErrorMsg`), or give navigation an
option to preserve it. At minimum, a comment on `navigateTo` warning callers
never to batch an `ErrorMsg` beside a `NavigateMsg`.

### 3b. `Manager.Kill` reads `s.Driver` outside the lock

`internal/session/lifecycle.go:318-330`

`Kill` looks the session up under `RLock`, releases it, and then reads
`s.Driver` to call `Kill()` on it. `s.Driver` is written under `m.mu` in
`activateSessionLocked`, so a concurrent relaunch of the same session would race
this read. `Relaunch` (`:77`) and `SwitchAgent` (`:91`) have the same shape: they
check `s.Status` after releasing the lock, so two concurrent relaunches could
both pass the "must be terminated" guard.

**Why it is only latent.** The TUI issues these one keypress at a time, and
`go test -race` passes because no test drives them concurrently. Plan 1
introduced a per-activation generation counter; it makes the fix cheap — capture
the driver and generation under the lock, then act on the captured values.

### 3c. `mcp_form.go` indexes its inputs by literal position

**Status: done in plan 3.** Named `mcp*Idx` constants plus `mcpFieldTable()`; drift
test `TestMCPFieldTableMatchesConstants`, round-trip test `TestMCPFormRoundTripsEveryField`.

`internal/ui/views/mcp_form.go` (~23 `f.inputs[N]` accesses, e.g. `SetMCP` at `:95-101`)

The same disease Plan 6 cured in the card editor: seven text inputs addressed as
`f.inputs[0]` … `f.inputs[6]` throughout `SetMCP`, submit and focus handling.
Swapping two indices silently writes one field into another — the command into
the URL, say — and neither the compiler nor the current tests would notice.

**Fix.** Named index constants, as the card editor now has
(`nameInputIdx` etc.), plus a drift test tying the constant count to the input
slice. The card editor's `TestInputIndexConstantsMatchTargets` is the template.

## 4. Duplicated modal boilerplate [low]

**Status: done in plan 3.** `components.ModalState` carries open/confirmed/cancelled;
`components.Modal` and the three modals embed it, and the `TODO use modal component` is gone.

`internal/ui/views/workspace_picker_modal.go:118`,
`internal/ui/views/agentsmd_editor_modal.go:69`,
`internal/ui/views/puppet_selector_modal.go:78`

Carried over from Act 1, where it is recorded as deferred out of Plan 7 step 5
(the plan allowed dropping it; the filter half of that item was done). Three
modals still each implement `Open`/`Close`/`IsOpen`, and
`puppet_selector_modal.go:13` still carries `// TODO use modal component`.
Give `components.Modal` an embeddable form and embed it in the three modals,
keeping each modal's own content rendering. Purely cosmetic — do it when one of
those modals is next touched, not on its own.

---

## 5. Nothing checks what the views actually look like [testing gap]

### What real use found

After Act 1, the first real use of the app found a card whose cat puppet was
split across the card, part of it rendered under the driver name. The cause was
a 1–2 column overflow: each skill pill ends in two columns of whitespace, and
when lipgloss word-wrapped the pill row, that whitespace hung past the info box
edge whenever the edge landed inside it. The row came out too wide, the card's
outer border wrapped it, and the face box spilled onto new lines. It only
happened for particular skill sets at particular terminal widths, which is why
it looked so arbitrary.

The instructive part is that **the bug predates all of Act 1** — it reproduces
identically on `6f3ca0d` — and it was invisible to roughly 11k lines of tests,
because no test asserted anything about the *shape* of rendered output. Tests
checked that the right text was present; none checked that the result was still
a rectangle.

The card is now fixed and guarded
(`internal/ui/components/card_render_layout_test.go`): pills are fitted to one
row by measured width with a `+N` summary, the info and face boxes are capped
with `MaxWidth`, puppet art is clipped rather than wrapped, and a sweep renders
eight skill sets and an oversized puppet at every width from `CardListMinWidth`
to 160. The other views now have the same kind of sweep; see Status below.

### Status

Written (`impl_plan_2.md`): shape tests for the session list, the session
preview, the dashboard, the card editor and the three modals, on a shared
`internal/ui/uitest` package; `assertIntactCard` now uses it. The sweeps found 12
layout bugs. One is fixed: the sessions list's built-in help line was wider than
its pane on 80- and 120-column terminals and made the pane grow two rows, so the
list no longer shows it (`SetShowHelp(false)`). The other 11 are skipped tests that
fail against the current code (three-bug limit); the list is in `impl_plan_2.md`,
"Found, not fixed". The ones to look at first: the card editor is taller than a
24-row terminal with default settings, the AGENTS.md modal wraps once any text row
fills its text area, and the footer is a fixed 150 columns.

### What to test

The same sweep, applied to every view that lays content out in fixed columns.
In rough order of risk:

| View / component | Why it is exposed |
| --- | --- |
| `components/session_list.go` | Rows combine a puppet face, a name, a status and highlighting in a fixed width — the same ingredients as the card. |
| `components/session_preview.go` | Renders arbitrary agent output (ANSI, wide runes, long lines) into a fixed pane; truncation was rewritten in Plan 7. |
| `views/dashboard.go` | Composes three panes side by side; one pane overflowing by a column pushes its neighbours, the same mechanism as the card. |
| `views/card_editor.go` | Many rows, optional sections that appear and disappear with settings, capability lists. |
| the three modals | Overlaid on the dashboard; an overflowing modal corrupts the view underneath. |

### What to assert

For each view, render it across a range of widths (and heights, where the view
takes a height) with realistic and adversarial content, and assert:

1. **Every line has the same display width** (`ansi.StringWidth`), equal to the
   width the view was given. A ragged edge is the signature of an overflow that
   something downstream wrapped.
2. **The line count is exactly what the layout promises** — not "at least". An
   extra line means content wrapped instead of being fitted or truncated.
3. **Content stays in its region** where a view has fixed regions: the
   card test's check that the info box's fifth row is blank is the pattern —
   it is what catches pills spilling onto a second row even when the outer
   rectangle survives.

Adversarial content matters more than the width range. The card sweep uses
names whose pill boundaries land at different columns, a CJK name, an empty
capability list and a puppet wider than its box; the equivalents here are long
session names, wide runes, ANSI-heavy preview output, and user puppets.

### Lessons worth carrying into these tests

Two things the card fix taught that apply to all of them:

- **A width backstop can hide the real bug.** `MaxWidth` alone kept the card a
  clean rectangle while the pills quietly went back to spilling onto a second
  row. Test the intended layout (region checks), not only the outer shape.
- **Check that each test can fail.** Every guard in the card sweep was verified
  by removing the fix and watching the test fail — fitting removed: 134 failing
  renders; fitting and backstop removed: 26. Act 1's recurring failure mode was
  tests that passed with the bug present; revert-and-rerun is the habit that
  caught every one of them.

---

## Found during Act 2 — fixed by plan 2.2

Two problems with one root (the session manager had no way to stop what it
started), found while checking what the test suite leaves behind:

- **Tests littered `/tmp`.** Each `Manager.watch` goroutine only returned on a
  dead process or a stale generation; after a test's `Kill` its next tick
  re-persisted the record, and when the test's `t.TempDir()` was already gone,
  `saveSession`'s `MkdirAll` recreated `<tempdir>/001/sessions/`. About 30
  directories per `go test ./...`. Fixed: `Manager.Shutdown` closes a stop
  channel and waits for the watchers, activations after it return
  `ErrManagerClosed`, every test manager is built with `newTestManager` (which
  shuts it down), and `internal/testenv` fails a package that leaves anything
  in its private temp root.
- **Quitting the app did not kill the agents.** `q` returned from `run()`
  without `Shutdown`; agents died only by accident, from the SIGHUP of the
  closing PTY, and anything ignoring SIGHUP (or a child it started) survived,
  orphaned. `Kill` was a bare SIGKILL to the direct child. Fixed: `Kill` sends
  SIGTERM to the process group, SIGKILL after a 2 s grace if anything in the
  group (not just the leader) is still alive; `Shutdown` kills in parallel and
  waits for launches already in flight; `main.go` shuts down after `p.Run()` on every path and turns
  SIGHUP into a quit. Known limit: descendants that start their own session or
  process group still survive.

Possible follow-up, not done: a "N sessions are running, quit anyway?"
confirmation on the dashboard's `q`.

---

## Suggested order

1. **Session atomic writes (§1)** — small, contained, and it closes the last
   durability gap using a recipe that already exists.
2. **Rendering-shape tests (§5)**, starting with the session list and the
   preview — cheap to write, and they catch the class of bug users actually see.
3. **The three sharp edges (§3)** — each is an afternoon; fold 3a into 4 if the
   split happens first.
4. **Split the coordinator (§2)** before starting Phase 1 of the roadmap.
5. **Modal boilerplate (§4)** opportunistically.
