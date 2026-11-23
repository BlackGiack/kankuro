# Feature 6 — Replace the dashboard center and remove the old preview

Status: planned; not implemented or reviewed.

Ordered prerequisite: `feature_5.md` must receive final-review PASS before starting; also honor the source dependencies below.

## Execution rules and review gate

This is a specification, not evidence of implementation or approval. Planning only
is authorized now. Do not implement or commit until instructed. Never commit
without an explicit request. Preserve unrelated working-tree changes.

Read `AGENTS.md`, `project.md` and `roadmap.md` before implementation and revalidate
these checkout paths. The roadmap's scope, dependency rules, behavioral guarantees
and non-goals apply. Keep each completed slice buildable and usable. No storage
migration, extra module, remote service, MCP wire change or unrelated redesign is
included. The only included product change is the center replacement in feature 6.

The frontend includes both `internal/app` during migration and `internal/ui`.
Frontend production code may depend on independent `internal/api/contracts` and
its own child packages, not concrete api/config/configstartup/session/driver/mcp.
Contracts must not leak backend structs through aliases, embedding, signatures or
resources, and must not depend on UI/framework types. Backend production code
must not import UI or terminal presentation frameworks. `main` owns concrete
composition and lifetime. Pure shared render helpers are permitted only after
checking their dependency graph.

Application operations own policy, authoritative validation and revalidation;
UI owns interaction, navigation, formatting and local feedback. Availability in a
read result never authorizes later mutation. Do not perform blocking work in view
Update/callbacks or access coordinator state from asynchronous command closures.
Capture inputs first; preserve dashboard/session identities, request generations
and epochs on async envelopes. Failed reads retain prior valid data and surface
errors; confirmed empty reads clear it. Define copy/immutable ownership for all
published slices/maps/buffers. Never automatically replay uncertain effects.

The assigned planning/review agent is the final reviewer for every feature. After
implementation, stop at this feature's review gate before advancing. The reviewer
must inspect the actual code, all callers and removals, backend policy tests and
actual command results, not just the implementation author's summary. Record
PASS or CHANGES REQUIRED, changed files, acceptance evidence, checks/results,
remaining exceptions and limitations in a separate review artifact such as
`feature_N_review.md` (replace N with this feature number). Do not put task logs
in `project.md`. Unrun checks remain unverified; failed criteria block advancement.
No implementation feature or review is approved by creation of these plans.

## Problem, result and scope

Replace the middle live session terminal preview with session-authored reports,
then workspace information when there are no reports. This is the sole included
product change. Its source is `feature-dbcenter-rework.md`; all five bullets and
the clarified roadmap behavior/failure policy below are binding.
Prerequisites: approved features 1–5 in this ordered plan (original minimum 1–2,
coordinating registry contracts with 5). Finish before retiring internal/app in 7.

## Checkout evidence and affected callers/tests

- `internal/ui/views/dashboard.go` owns SessionPreview, selection sync, pane sizing
  and rendering; `components/session_list.go` emits preview selection intents.
- `internal/ui/components/session_preview.go` and its test/layout/overflow/integration
  suites implement the old display.
- `internal/app/preview.go` subscribes to manager output; `handlers.go`,
  `coordinator.go`, `scoped_refresh.go`, `refresh.go` and `ui/messages/messages.go`
  carry output/scrollback, PreviewDataMsg/PreviewReceiver and preview refresh state.
- `internal/session/messaging.go` has MessageReply author/session/activation,
  CreatedAt/Sequence, bounded durable history and Changed signal; replies carry
  SenderSessionID (the reporting author), distinct from original message sender.
  `api/messaging.go` has MCP-authorized history reads, not a human panel API.
- Revalidate `internal/app/preview_broadcast_test.go`, `preview_integration_test.go`,
  `command_race_internal_test.go`, `internal/ui/views/dashboard_layout_test.go`,
  `internal/session/messaging_watch_test.go`, `messaging_test.go` and
  `internal/api/messaging_review_test.go`. Preserve `session/attach.go`, `io.go`,
  `driver` input serialization and their tests; output has shared attach callers.

## Contract and migration decisions

Add a human application read scoped to persisted dashboard/session membership,
including stopped sessions; do not call agent MCP tools with synthetic credentials.
Project retained authored replies by SenderSessionID across activations and include
reply identity, kind, text, time.Time and sequence. Sort CreatedAt descending then
Sequence descending, then stable reply ID. Avoid duplicate notification records.
No activation replacement may clear the stable session's report display.

Expose cancellable change subscriptions through application contracts, no tea.Msg.
Subscribe/register the change signal before reading, with version/check logic so
a change during the initial read forces another read. A buffered/coalesced signal
is sufficient because refresh reads retained records. Define idempotent release,
closed-stream recovery and cancellation; do not hold store/manager locks while
waiting or rendering. Reuse the store change mechanism and durable history;
no new event bus or rendered cache persistence.

Backend reads accept a recent-count limit capped by existing retention; return
structured records and omission/truncation metadata. UI derives a count from its
available content rows (at least one row per report), increases/reissues on resize
within that cap, wraps text by display columns and clips newest-first. Records are
not terminal-preformatted or width-truncated. Bound pathological text processing
using existing report size bounds. No report pager, scrolling or new retention rule.
If retention supplies fewer records than space allows, show the available data.

Planning defaults for workspace inspection: Git status covers the entire containing
repository (including staged, unstaged, untracked and rename paths), matching git
status from the saved workspace. Use cancellable argv-based porcelain -z parsing;
label repository-relative paths. Classify non-repository explicitly; exit errors
alone do not prove non-Git. Detailed filesystem entries cover the saved workspace's
immediate children, including dotfiles and synthetic . / .. rows for ls -la parity;
show type/mode, link count where available, owner/group where available, size,
mtime and symlink target. Do not follow symlinks recursively. UI sanitizes paths
and formats time/permissions. Portable unavailable fields have explicit absence.
Use an initial maximum of 1,000 workspace entries and visible omitted/more state;
cap parser output and execution time, cancel on selection change. Exact limits
can be adjusted with measured evidence without changing semantics. No shell ls/git.

Frontend maintains same-session successful content plus read state, warning and
stale/fallback labels. Load immediately, subscribe even in workspace mode, and use
one non-overlapping 30-second timer for workspace mode or degraded recovery. On
report/subscription failures retry no faster than that cadence; successful recovery
returns to event-driven reports. Reports with a successful read take priority and
suppress unnecessary filesystem work. Apply the complete failure table below.

Selection identity, dashboard and generation accompany every command result. Cancel
read/timer/subscription on selection/navigation/session removal/exit, close stale
results, retain reports on stop/relaunch, reconstruct from retained history after
restart. Recompute fit on resize. Remove all preview-exclusive APIs/tests/docs and
allowlist entries only after auditing output consumers. Replace obsolete tests
with report/fallback/cleanup tests; keep attachment output/scrollback/input gates.
Any selected-session hook needed by the center must emit a new semantic selection
intent, with no misleading preview compatibility bridge.

## Preserved roadmap requirements

The following workstream is carried forward without dropping tasks or acceptance criteria.

Dependencies: steps 1–2; coordinate shared read contracts with step 5. Must finish
before step 7 retires `internal/app`.

Source requirement: `feature-dbcenter-rework.md`. The decisions below incorporate
the user's clarifications, including approval of the failure policy and the
available-space report display. The restart reconstruction strategy below avoids
introducing another persistence mechanism for content already in report history.

Confirmed behavior:

- Show reports authored by the selected session, not reports it received or all
  reports in the dashboard. Include report text, kind and the time it was recorded.
- Keep reports across relaunches of that session. Select by stable session identity,
  not only the current activation. Preserve activation attribution internally where
  needed; a different session must not inherit the history.
- Order reports by timestamp descending, with a stable identity/sequence tie-breaker.
- Show only report content that fits the center panel's available space. Older or
  overflowing content is omitted from this display; do not provide report paging
  or history scrolling. New reports take precedence and push older content out.
  Omission from the display does not delete the underlying message-store reports
  or change existing MCP history behavior and retention.
- On session stop, keep the current report display. Relaunch must not clear it
  merely because the activation changes. On Kankuro restart, reconstruct the
  fitted display from retained durable reports rather than saving rendered lines.
  Do not promise reports beyond the existing store's retention guarantees.
- Refresh reports when a report is received. When no reports exist, refresh the
  workspace view every 30 seconds. Load immediately on selection; subscribe to new
  reports even in workspace mode so the first report changes the panel promptly.
- If no reports exist, show changed files from the session's actual workspace,
  equivalent in purpose to `git status`. If it is not a Git repository, show a
  detailed listing equivalent in purpose to `ls -la`.

Tasks:

1. Add dashboard-scoped application reads for retained reports authored by the
   selected session across activations. Reuse the existing message store without
   changing MCP authorization or impersonating an agent to call MCP tools. Define
   the human read scope and validate persisted session/dashboard membership.
2. Expose report changes through an application-owned subscription/change signal,
   with explicit unsubscribe/cancellation semantics. Deliver no Bubble Tea values
   from backend code. Close the read/subscribe race so the first new report cannot
   be missed between the initial query and subscription. Coalescing changes is fine
   if a refresh retrieves all new retained records.
3. Introduce structured contracts for reports, Git entries and directory entries;
   do not return raw storage records or terminal-formatted command output. Keep
   timestamps as time values and format them in the frontend.
4. Keep Git/filesystem inspection behind the application boundary. Inspect the
   saved actual workspace, including temporary workspaces. Use safe argv-based Git
   execution and structured status parsing; use filesystem APIs for directory
   entries where practical. No shell interpolation or blocking I/O in UI Update.
5. Bound reads and rendering to the recent content useful for the panel. The UI
   owns display-width measurement, timestamp layout, wrapping and clipping to its
   current dimensions; the backend returns structured recent reports with a
   bounded query limit, not terminal-sized or preformatted strings. Fetch enough
   recent entries to fill the available space without loading the entire history.
   Recompute on resize, allowing retained older reports to fill newly available
   space. No report pagination, history scrolling or new history deletion policy.
   Keep large workspace inventories bounded too, with visible omission indication;
   no unbounded file listing should block the UI.
6. Implement failure handling using the policy below. Distinguish confirmed empty
   results from failed reads, and mark stale/fallback content explicitly.
7. Add the UI center component and commands. Keep selected session identity,
   dashboard scope and generation on asynchronous results. Cancel obsolete reads,
   timers and subscriptions when selection changes, navigation leaves the screen,
   the session disappears or the application exits. Avoid overlapping 30-second
   workspace refreshes. New reports update the fixed newest-first display rather
   than preserving a scroll position. Session stop must not clear report content.
8. Remove the old preview component, dead selection hooks, output subscription
   wiring, preview-specific payloads and obsolete tests/documentation. Inspect
   callers before deleting shared session output/scrollback functionality used by
   attachment. Replace preview tests with center-panel behavior tests.

Agreed failure policy:

- Successful report read with reports: display reports. Workspace failure must not
  hide reports, and there is no need to inspect the workspace in this mode.
- Report read failure: show a warning and keep the last successful report view for
  that session marked stale. If there is no usable cached report view, attempt the
  workspace view and label it as fallback because reports could not be loaded.
  Never report a failed history read as proof that there are no reports.
- Git workspace with no changes: display "Working tree clean"; do not switch to a
  directory listing merely because Git returned an empty successful result.
- Confirmed non-Git workspace: display its directory listing without an error.
- Git inspection fails for another reason: show the Git error and attempt the
  directory listing as explicitly labelled fallback. Do not classify permission,
  execution, timeout or repository errors as "not a Git repository".
- Directory listing fails, or the workspace is missing/inaccessible: show the
  error. Retain suitable prior content for that same session marked stale if
  available. Do not create the workspace or show unrelated files.
- Report subscription/read failures must have bounded retry/recovery behavior;
  use the 30-second fallback refresh while degraded, then return to event-driven
  updates when recovered. Avoid error-driven busy loops.

Acceptance:

- Report author/session selection, relaunch continuity, descending order and
  immediate event-driven updates match the confirmed behavior above.
- With no reports, Git/directory information loads immediately and refreshes every
  30 seconds; a received report promptly takes precedence.
- A clean Git workspace, non-repository workspace and read failure are distinct.
- Errors remain visible alongside useful fallback information where available.
- Only the newest content fitting the available panel area is shown, with no
  report paging or history scrolling. Resize, new reports, stop, relaunch and
  Kankuro restart follow the display/continuity rules above.
- Stale results cannot display another session's reports/files. Resource cleanup
  remains correct when switching selection during reads/events.
- No obsolete preview UI plumbing or migration exception remains. Attachment still
  receives its initial output and live stream and retains the input gate.
- Frontend code imports no Git/process/storage implementation to provide the panel.

Review focus: report attribution across activations, fitting/clipping and restart continuity,
read/subscribe races, scope, cancellation, large inventories, fallback labels and
shared attachment functionality accidentally removed during preview cleanup.

## Focused validation and final-review evidence

Run focused regressions while changing the slice, then these checks at its integration gate (adapt moved package paths after feature 7). Commands below are planned, not claimed results. Confirm test selection actually runs relevant cases; add behavioral regressions where coverage is missing.

```bash
go test ./internal/api ./internal/session ./internal/app ./internal/ui/components ./internal/ui/views
go test -race ./internal/api ./internal/session ./internal/app -run 'Report|Message|Center|Attach|CommandClosures'
git diff --check
```

Require evidence for authored-versus-received attribution, old/new activations, timestamp ties, first-report race, fake-clock 30-second refresh/recovery, full failure table, bounded huge inventories, resize clipping and stale-session cleanup.

- [ ] Every preserved task and acceptance criterion is mapped to actual code/tests or documented unresolved work.
- [ ] Reviewer inspected migrated callers, removed paths, contract signatures and remaining exception ownership.
- [ ] Focused behavioral checks and applicable race checks have recorded actual results; failures are resolved or block approval.
- [ ] Application policy still revalidates authority/state; frontend controls only interaction and presentation.
- [ ] Read failures, partial effects, cancellation, stale results and mutable ownership meet this feature's validity rules.
- [ ] No unrelated behavior change, dead path, permanent bridge or new undocumented dependency exception remains.
- [ ] Reviewer records explicit PASS or CHANGES REQUIRED before the next feature advances.

Do manual terminal verification with isolated configuration and temporary workspaces
as documented in project.md; native-agent experiments also isolate user configuration.
Native model-driven E2E needs the provider/model and total token-budget authorization
specified by tests/e2e/AGENTS.md; these plans do not grant it. Documentation-only
planning validation does not require application tests.
