# Feature 6 implementation plan — dashboard center replacement

Status: proposed implementation plan; planning only. No implementation, commit,
test execution claim or feature approval is conveyed by this document.

## Scope and prerequisites

Read `AGENTS.md`, `project.md`, `roadmap.md`, `feature_6.md` and
`feature-dbcenter-rework.md`. Feature 5's `feature_5_review.md` records final-review
PASS and readiness for feature 6. Revalidate the checkout and prerequisite review
before implementation. Preserve unrelated changes, including untracked roadmap
and workflow artifacts. Never commit without an explicit request.

Replace only the dashboard center: retained reports authored by the selected
session, otherwise Git changes for its saved actual workspace, otherwise a detailed
directory listing for a confirmed non-repository. Remove obsolete preview code.
Preserve attachment, layout proportions, session selection and lifecycle behavior.
Finish this slice before feature 7 retires `internal/app`.

No storage migration, MCP wire/authorization change, new module/service/event bus,
rendered-content persistence, report pager or changed retention is included.

## Checkout findings and design choice

- `internal/ui/views/dashboard.go` owns `SessionPreview`, selection reconciliation,
  sizing and rendering. `components/session_list.go` emits `OnPreview` intents.
- `internal/app/preview.go` owns output subscriptions. Preview state also enters
  coordinator handlers/navigation and scoped inventory/config refreshes through
  `retainedPreview` and preview-specific message fields.
- `internal/session/messaging.go` holds reports inside retained parent messages.
  `MessageReply.SenderSessionID` identifies the author; parent sender and recipient
  fields are different concepts. Report identity, activation, time and sequence
  already exist. Notification records must not become duplicate report rows.
- `MessageStore.Changed()` is a close-and-replace channel, also signaled by
  lifecycle/unrelated changes. It is not a persistent stream to range over.
- Existing `GetReplies` returns ascending sequence pages across a dashboard or
  message, capped at 100 and a byte budget. It is an MCP history contract, not the
  desired newest-first author query. Existing history bounds include 10,000 parent
  messages, 64 MiB retained history, 128 reports per message and 16 KiB per report.
- `api/membership.go` owns persisted dashboard/session checks; stopped sessions
  remain eligible for human reads. `Manager.Messages()` supplies the shared store.
- `refresher.sessions` also supplies startup `LoadErrors`; removing preview reads
  does not justify deleting that remaining feature-7 dependency blindly.

Use a focused application read plus a cancellable subscription over the existing
store. Add bounded workspace inspection behind the same application boundary.
Keeping the preview would miss the product requirement; polling reports on every
existing two-second inventory tick would miss the event-driven requirement and
couple independent refreshes. A new event bus or persisted panel cache adds no
needed capability. The accepted cost is a small subscription adapter and explicit
frontend state machine, with deterministic race/failure tests.

## Contracts and ownership

Add independent types in `internal/api/contracts/center.go` and a consumer-owned
`centerPort` in `internal/app/center_ports.go`. Proposed method names:

```go
ReadSessionReports(context.Context, contracts.SessionReportsRequest) (contracts.SessionReportsResult, error)
SubscribeSessionReports(context.Context, contracts.ScopedSessionRequest) (contracts.ReportSubscription, error)
ReadSessionWorkspace(context.Context, contracts.ScopedSessionRequest) (contracts.SessionWorkspaceResult, error)
```

Requests carry dashboard/session identity; report requests also carry a recent-count
limit. Reports carry reply/message identity, author session/activation attribution,
kind, text, `time.Time` and sequence. Results carry a change revision, effective
limit/maximum and omission metadata. Omission describes retained content excluded
from this read, not a claim about history evicted earlier. Keep the MCP page cap
separate from this read; clamp to existing retention-derived bounds and a documented
response budget. Never return an empty success merely because a limit was invalid.

`ReportSubscription` exposes a coalesced signal and idempotent release, with a
defined terminal error/closure outcome. Context cancellation must unblock waiters.
Only standard-library and contract types cross this boundary: no store records,
manager pointers, `tea.Msg`, PTYs, formatted times or terminal-width strings.
Clone published slices and nested data; ownership remains detached from storage.

Workspace results distinguish Git status, directory listing, confirmed non-Git,
inspection failure, partial/truncated data and warnings. Git entries carry index
and worktree status, repository-relative current/original paths and applicable
conflict/submodule facts. Directory entries carry name, type/mode, optional link
count and owner/group, size, modification time and optional symlink target.
Represent unavailable platform metadata explicitly. Return the saved workspace
and repository-root facts needed to label paths accurately.

Service implementations in proposed `internal/api/center.go` and
`center_workspace.go` own scope and saved workspace resolution. Reuse authoritative
membership checks; reject foreign/deleted sessions or inactive dashboards and
surface inventory read failure. Recheck scope across asynchronous setup/reads so
deletion cannot publish a newly authorized resource from stale inventory. Do not
require a live activation or Dashboard Control for a local human panel read, and
do not manufacture agent credentials. Validate workspace identity rather than
resolving a card's current configured path. `main` supplies the capability through
existing composition; views receive presentation models and semantic intents.

## Ordered implementation slices

### 1. Authored report projection and race-free subscription

Add a store-owned recent-authored-report query in `internal/session`, filtering
both parent dashboard membership and reply author session across all activations.
Scan retained history without cloning all reports; retain only bounded newest
candidates, then sort by CreatedAt descending, Sequence descending and reply ID
ascending for deterministic residual ties. Return each reply once, including
progress reports; notification state does not gate display. Do not use MCP cursor
pagination internally to fetch all history. Honor cancellation during long scans
and keep lock work bounded by the already bounded store.

Implement the human read mapping and subscription adapter. Register/capture the
store signal before the first read; pair reads with an internal change revision
or equivalent snapshot/check protocol under store synchronization. A change during
the read schedules another read. After every signal, capture the replacement
channel before refreshing; a closed old channel must not cause a busy loop.
The adapter can coalesce signals into a single buffered notice because reads
retrieve the latest retained result. Global store signals may cause a harmless
refresh; they must never inject another author's report.

Define synchronous setup completion before initial reading, cancellation and
release ownership, store shutdown/terminal closure recovery, and scope removal.
Do not consume MCP long-poll authority or change its waiter behavior. Never hold
manager/store/repository locks while waiting on channels, doing external I/O or
rendering. Tests must force a report before/during/after the first read and during
subscription rearming, not depend on sleeps.

### 2. Bounded workspace inspection

Resolve and inspect the saved actual workspace, including temporary copies. A
missing/inaccessible workspace returns an error without recreation or substitution.
Use cancellable argv execution, a bounded stdout reader/parser and bounded stderr,
with an initial 10-second command deadline (adjust only with evidence). Kill and
reap subprocesses on cancellation or limit exhaustion. Use no shell and no `ls`.

Probe Git explicitly and parse NUL-delimited porcelain status. Run status from
the resolved repository root so a saved subdirectory still covers the entire
containing repository; include staged, unstaged and untracked changes, renames
and both rename paths. Use `--untracked-files=all`, a documented porcelain version
and explicit repository-relative labels. Parse bounded complete records; truncated
or malformed output is never a clean working tree. Treat paths containing spaces,
newlines or leading dashes as data. Preserve paths unsanitized until presentation.

Do not classify every Git exit failure as non-Git or rely only on localized stderr
text. Use repository discovery evidence (including worktree `.git` files/linked
worktrees and ancestor access errors), plus a controlled Git probe. Recognized
repository metadata with a failed probe/status is an inspection error. Missing Git,
permissions, timeouts, corruption and unsupported repository contexts remain errors.
Test the probe decision explicitly; show a labelled directory fallback for errors.

For directory mode, read immediate entries in bounded batches using filesystem
APIs, including hidden files and synthetic `.` / `..` metadata rows. Use lstat and
readlink; never recurse through symlinks. Collect at most 1,000 entries including
synthetic rows plus enough evidence to indicate more; sort the bounded result
deterministically and expose truncation without enumerating an unbounded directory.
Document that this is a bounded sample when truncated. Use platform helpers for
optional owner/group/link metadata. Per-entry failures produce explicit partial
data/warnings; a failed directory read does not silently become an empty listing.

### 3. Center state, commands and rendering

Add `components/session_center.go`, presentation data in `internal/ui/models`,
center messages and coordinator command adapters (`internal/app/center.go`).
Replace the dashboard center component while retaining existing pane geometry,
minimum terminal behavior and navigation ownership. No filesystem/process/storage
imports in frontend production code. Render timestamps, permissions and safe text
only at the UI boundary; sanitize ANSI/control sequences, paths and symlink targets.

Use a new semantic selected-session intent in session-list/dashboard reconciliation;
remove `OnPreview` rather than bridging it to the center. Carry dashboard ID,
stable session ID, view/selection generation, request ID and applicable existing
session epoch on every async envelope. Distinguish request ordering from stable
session identity: stop/relaunch can refresh facts without clearing report content.

The coordinator owns one selected-session context, active subscription, current
read identity, pending-refresh bit and timer generation. Commands capture values
and capabilities before creation and never access coordinator state later.
Subscribe before the immediate report read even in workspace mode. Coalesce
events during an in-flight read into one follow-up. Keep report reads responsive
while workspace I/O runs: the first report cancels/invalidates obsolete workspace
work and immediately takes precedence. Discard late workspace results even for
the same session if a newer report result has superseded their mode/request.

Use exactly one cancellable, non-overlapping 30-second refresh cycle for workspace
mode or degraded report/subscription recovery. Start the next cycle after the
current attempt completes; do not let the two-second inventory tick drive it.
Persistent errors and closed streams cannot trigger rapid retries. Recovery
resubscribes before reading, then healthy report mode stops periodic workspace
inspection. Relevant events/resize can request normal reads; failure loops remain
rate-limited. Keep warning/stale/fallback state within the panel so useful errors
remain visible independently of the coordinator's transient error banner.

On selection change, dashboard navigation, session removal and application exit,
cancel reads/timers/subscription and release resources exactly once. A late setup
result must release its subscription instead of installing it. Cancellation of old
selection work is not a new warning. Never reuse another session's cached content.
Stop leaves reports displayed; relaunch retains them; application restart rebuilds
from retained reports without saved rendered lines.

Derive the report query count from usable content rows after borders/header/error
space, at least one per possible report; cap it through the contract. Wrap by
display columns, render newest-first and clip to the rectangle, including partial
newest report content when necessary. No scrolling keys or history pager. Show an
omission indicator when space permits without displacing all report content.
Resize immediately refits cached structured data and reissues a larger bounded
read when more reports can fit. Bounded text processing follows the 16 KiB report
limit and requested count. Zero-sized panels render safely without repeated reads.

### 4. Preview removal and integration cleanup

Once the replacement works, remove `session_preview.go` and its obsolete component
tests; replace their useful layout/overflow assertions with center assertions.
Remove `internal/app/preview.go`, preview handlers/coordinator cases, subscription
state, navigation broadcasts, `retainedPreview`, refresh parameters/payload fields,
`PreviewReceiver`, `PreviewDataMsg`, `SessionPreviewRequestedMsg` and preview-only
`SessionOutputMsg` after a full caller audit. Rename dashboard preview sizing and
selection helpers to center terminology without compatibility aliases.

Audit every caller of `GetScrollback`, `SubscribeOutput`, `UnsubscribeOutput` and
`GetSession` before deleting any API. Preserve `internal/session/attach.go`, `io.go`,
attachment initial scrollback/live stream, input serialization, detach/release and
process lifetime. Retain legitimate startup/composition/attachment dependencies
owned by feature 7; do not hide them in a new preview exception.

Update preview-specific app tests (`preview_broadcast`, `preview_integration`,
`preview_reconciliation`), dashboard layout/selection tests and command-race tests.
Remove feature-6 edges from both architecture allowlists and update
`docs/frontend-boundary.md`. Audit `project.md`, `kankuro_specs.md`, appedia and
other current behavior docs for obsolete preview descriptions; update current
documentation/references without rewriting historical workflow evidence. Keep
implementation evidence in a separate feature artifact, not `project.md`.

## Required failure behavior

| Observation | Center behavior | Recovery |
|---|---|---|
| Report read succeeds with reports | Show reports; no workspace inspection | Event-driven subscription |
| Report read succeeds empty | Clear old report state and inspect saved workspace | Subscribe plus 30-second workspace cycle |
| Report read fails, same-session report cache exists | Keep reports marked stale with warning | 30-second degraded retry |
| Report read fails, no usable report cache | Inspect workspace, label report-unavailable fallback | 30-second degraded retry |
| Git succeeds with zero changes | Show `Working tree clean` | Workspace cycle |
| Confirmed non-Git | Detailed directory listing, no Git error | Workspace cycle |
| Git fails otherwise | Keep Git warning and attempt labelled directory fallback | Workspace cycle |
| Directory/workspace fails | Show error; keep suitable same-session prior content stale, if any | Applicable 30-second cycle |
| Subscription fails/closes unexpectedly | Preserve useful content, mark updates degraded | Bounded resubscribe/read recovery |

Warnings can coexist (report read plus Git plus listing); successful fallback must
not erase the reason for fallback. A scope rejection/session deletion ends the
selection rather than continuing to expose cached content for an unavailable
session. Confirmed empty results and failed reads are distinct states throughout.

## Verification and acceptance evidence

Implement behavioral tests with fake capabilities, controlled channels/clocks and
temporary real Git repositories/directories. No native model calls are required.

| Feature-6 task / acceptance | Required evidence |
|---|---|
| 1: human scope and authored history | Foreign dashboard rejection; inventory error; stopped session; author vs received/parent sender; all report kinds; old/new activations; restart reload; notifications not duplicated |
| 2: change subscription | Initial-read and rearm races; burst coalescing; unrelated changes; cancellation/idempotent release; scope removal; closed-stream recovery; no leaks/busy loops |
| 3: structured contracts | Detached data; `time.Time`; independent signatures/imports; raw paths preserved until UI sanitization |
| 4: workspace boundary | Saved temporary path vs card path; subdirectory/full repo; staged/unstaged/untracked/rename; linked worktree; clean vs non-Git vs error; missing Git/path; malformed output; cancellation and timeout |
| 5: bounded reads/display | Limit clamping and response bounds; 1,000-entry truncation; symlinks/dotfiles/metadata absence; large report text; wide/combining/control characters; tiny panel; grow/shrink resize fetching older retained rows; no scrolling |
| 6: failure policy | Every table row, combined warnings, stale same-session content, confirmed empty clearing and successful recovery |
| 7: lifecycle/orchestration | Immediate initial load/first-report takeover; fake-clock 30-second refresh and degraded retry; no overlap; slow workspace cannot block reports; selection A→B→A, dashboard switch, navigation, removal/exit and late resource cleanup; stop/relaunch continuity |
| 8: removal/preservation | No preview-exclusive production symbols/allowlist entries; revised docs/tests; attach initial/live output, input gate and detach regressions pass |

Run focused tests per slice, then record actual integration results (these commands
are planned, not already run):

```bash
go test ./internal/api ./internal/session ./internal/app ./internal/ui/components ./internal/ui/views
go test ./internal/architecture/...
go test -race ./internal/api ./internal/session ./internal/app -run 'Report|Message|Center|Attach|CommandClosures'
go build ./...
go vet ./...
git diff --check
```

Confirm filtered tests actually execute the new cases. Apply repository lint
policy and record any existing debt separately. Manually inspect isolated terminal
fixtures at 80×24 and larger sizes, including resize, first report, clean/non-Git
workspace, fallback warnings and selection switching. Isolate configuration and
temporary workspaces as in `project.md`; native-agent experiments additionally
isolate HOME and require the E2E authorization from `tests/e2e/AGENTS.md`.

## Review gate, risks and remaining questions

No blocking product question remains: use the binding failure policy, full-repository
Git semantics, immediate-child directory semantics and available-space report fit.
The command timeout/byte budgets and exact contract names are engineering defaults
to confirm with focused tests, not reasons to weaken semantics or change MCP.
Revisit bounds if measured cancellation/latency or common inventories show they
prevent useful content; preserve visible truncation and finite work.

Main risks are close-and-replace signal misuse, stale same-session workspace results
overwriting reports, conflating non-Git with Git errors, and deleting shared attach
paths. The targeted tests above are required mitigations. Keep each integrated
slice buildable and retain the existing preview until replacement cutover; do not
leave permanent dual plumbing. No data migration makes rollback a code change;
existing durable reports and attachment data remain authoritative.

After implementation, the assigned planning/review agent independently inspects
actual code, callers/removals, contracts, architecture exceptions and test output.
Record changed files, acceptance evidence, limitations and explicit PASS or
CHANGES REQUIRED in `feature_6_review.md`. Failed/unverified criteria block approval;
do not advance to feature 7 or commit without the required instruction.
