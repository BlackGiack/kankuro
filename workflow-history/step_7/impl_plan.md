# Step 7 implementation plan — close resource boundaries and retire internal/app

Status: **planned only; implementation and final review pending.**
Planning date: 2026-11-17.

## Authority, prerequisites and scope

This plan implements `feature_7.md` and roadmap step 7 after approved steps 1–6.
`feature_6_review.md` records final-review PASS and explicitly hands off to feature 7;
no feature-6 blocker remains. This task authorizes writing this plan only. Do not
implement, advance to feature 8, or commit without the corresponding instruction.
Never infer current commit permission from the historical feature-6 review.

Preserve all existing workflow files, including `workflow-history/step_1` through
`step_6`, feature specifications/reviews/evidence and roadmap. Preserve existing
uncommitted Appedia baselines, VERIFICATION.md, test evidence and maintenance
records. Do not regenerate Appedia as part of this planning task. Before eventual
implementation, capture working-tree status and revalidate instructions and paths.

Keep each implementation slice buildable and usable. No storage migration, MCP
wire change, new module/service, visual redesign or new product feature. Preserve
feature-6 report attribution, scoped cache continuity, selection cancellation,
structured read failures and backend Git process-group cleanup. Feature 8 owns
final architecture certification, not unfinished feature-7 production boundaries.

## Checkout findings and decisions

The current checkout confirms the following remaining work:

| Current surface | Required change |
|---|---|
| `internal/session/manager.go:AttachHandle`, `session/attach.go` | Keep process/subscription/input ownership in backend; hide PTY and concrete handle behind an application resource. |
| `api/membership.go:AttachInDashboard` and `AttachExpectedProcessInDashboard` | Return independent attachment contract deliberately; exact Go return signatures must match consumer ports. Preserve admission and expected-process checks. |
| `app/attach.go`, `ui/messages/messages.go:ScopedAttachReadyMsg` | Replace handle fields; move terminal execution to `ui/terminal`, backend resize behind resource. |
| `app/dashboard_list.go`, `callbacks.go`, coordinator deletion cases | Stop reading config manifests, full cards and session reference impacts; use summary/confirmation/progress contracts. Preserve operation results currently discarded by commands. |
| `api/card_confirmation.go`, `dashboard_confirmation.go`, `dashboard_deletion.go` | Retain authoritative token, admission and durable cleanup logic; project independent human-facing results. |
| `app/coordinator.go`, `app/app.go`, `legacy_actions.go` | Replace concrete service/manager construction and stored dependencies with cohesive consumer capabilities; delete legacy adapter. |
| `app/refresh.go:sessionLoadErrors` | Replace manager LoadErrors access with startup diagnostics supplied by main, preserving nonfatal hydration warnings. |
| `main.go` | Construct backend and frontend, inject capabilities, route runtime error through frontend entry point; retain lifetime ordering. |
| Both `internal/architecture/*allowlist.json` files | Remove every migrated production exception, including transitive signature/method exposure. |

`app.New` currently takes `(sm, service)`, and `NewCoordinator` takes the same two
arguments. Some project onboarding text still describes an older constructor;
actual declarations determine the move and current documentation must be corrected.
The coordinator's manager field appears unused outside construction, while the
refresher uses the manager for startup load errors; verify all callers before removal.

Recommendation: introduce focused attachment/deletion contracts and adapt existing
backend owners, then move the separated frontend mechanically. Moving files first
would retain the known dependencies and obscure policy review. Retaining raw handles
or a forwarding app package violates acceptance. A general backend facade/event bus
adds needless surface. The chosen approach costs explicit mappings and lifecycle
tests but preserves existing application policy and allows slice-by-slice validation.

## Contract design and ownership

### Attachment resource

Add `internal/api/contracts/attachment.go` with an application-owned resource and
initial-output value. Proposed signatures, to implement consistently in adapter,
consumer port and fixtures:

```go
type AttachmentInitial struct {
    Scrollback []byte
    Header     []byte
}

type Attachment interface {
    Initial() AttachmentInitial
    Next(context.Context) ([]byte, error)
    Write([]byte) (int, error)
    Resize(rows, columns int) error
    Close() error
}
```

`Next` is the read-only live output/completion stream: return queued chunks before
terminal EOF/error, then a stable terminal result. One reader owns consumption;
`Next` must unblock on context cancellation, resource close or backend completion.
Initial data and each returned chunk are detached/ownership-transferred; backend
never mutates them after publication. Define repeated `Initial` reads as detached
copies. Do not expose a file, PTY, driver, manager or session handle, even indirectly.

A proposed `api/attachment.go` adapts the subscribed activation. Capture its exact
process identity, subscription and driver; never resolve a newly relaunched session
by stable ID for subsequent I/O. Backend resize validates dimensions, checks resource
liveness and invokes PTY sizing. Stale/replaced activations reject input and resize
with meaningful errors. Writes use the existing driver's serialized input gate;
short writes and errors retain evidence and are never automatically replayed.

`Close` is idempotent, safe with concurrent completion and partial setup, and releases
only owned subscription/waiters. It cannot kill the session or close shared PTY/driver
resources. Do not wait holding manager/store/repository locks. Prefer a direct
context-aware channel adapter; add a pump only if required for ordering, and give it
explicit cancel/join ownership. Keep bounded existing buffering semantics and record
any overflow behavior; no unbounded queue. A process-done signal alone cannot discard
queued trailing output. Audit watcher/close ordering in `session/io.go`, `watch.go`
and lifecycle teardown before relying on channel closure as completion.

Service attach capabilities accept context plus dashboard/session identity and the
optional expected process identity via two explicit operations. Cancellation after
subscription but before publication must close the resource. Unused resource-plus-error
results must also close. Preserve local human scope rules; do not require Dashboard
Control or synthesize MCP credentials.

### Deletion and dashboard operations

Add `contracts/deletion.go` with independent prepare, execution and status values.
Use typed IDs/facts appropriate to existing conventions, without backend aliases:

- Preparation: dashboard/card identity, names, owned-session/card summaries,
  current/saved/pending references and foreign-dashboard facts, opaque confirmation
  token and `RequiresForeignConfirmation`.
- Execution/status: target identity, semantic outcome (no intent, pending/retryable,
  completed, or uncertain), human-relevant checkpoint progress and diagnostics.
  Represent unavailable status distinctly from confirmed absence; failed reads keep
  valid prior data. Never infer completion from disappearance in an inventory.
- Progress records expose relevant stopped/removed/deleted/cleaned facts and errors,
  not raw manifests, schema fields, storage revisions or cleanup filesystem paths.
  Clone nested reference/progress slices. UI formats/sanitizes summary lines.

Add service-owned projections in a focused deletion adapter and consumer-owned
`deletionPort`/dashboard mutation capability. Prefer focused human-facing methods
wrapping existing confirmation/deletion operations so internal/backend consumers
can retain legitimate persistence APIs. Audit all callers before changing or removing
an existing public method; keep no unused frontend compatibility bridge. Dashboard
create/update methods must also return contract-only values or error-only outcomes.

Backend preparation/confirmation retains existing two-minute token expiry, impact
and full revision digest, owner/scope binding, foreign confirmation and admission
synchronization. Confirm must revalidate immediately before durable intent. Retry
resumes an existing authorized intent, never creates one. Preserve intent persistence,
admission closure, pending activation cancellation/drain, per-effect checkpoints,
cleanup after row removal and restart recovery. Cancellation or failure after intent
returns durable pending progress plus diagnostic. If durability is uncertain, expose
that uncertainty without retrying an effect automatically. UI acknowledgement cannot
replace these checks. Do not add a new durable journal or alter MCP responses.

## Ordered buildable slices

### 1. Backend attachment boundary

Implement the contract and backend adapter, deliberately update scoped service
signatures and all callers. Convert `ScopedAttachReadyMsg` and the attachment command
adapter to the independent resource. Add a narrow consumer `attachmentPort`; remove
attachment use of `legacy.service`. Keep admission validation and expected activation
identity atomic with subscription/header capture. Verify initial replay plus live
output, serialized input, backend resize and close behavior before terminal movement.

### 2. Human terminal ownership and stale-result disposal

Move raw mode, stdin polling, Ctrl+K/Kitty detach parsing, stdout interaction, terminal
size observation and keyboard protocol push/pop/restoration to `internal/ui/terminal`.
The root coordinator retains Tea command/message routing; terminal code consumes only
the contract and injected human I/O. Read human dimensions using terminal APIs and
pass rows/columns to backend Resize; remove frontend creack/pty use.

Use deferred cleanup immediately after acquiring each resource. Join owned output
workers before returning; do not restore terminal state while a pump can still write.
Propagate header/output/input/resize errors with terminal restoration and no effect
replay. Cover setup failure before raw mode and after raw mode, stdin EOF, backend
EOF, cancellation and detach. Preserve header/newline behavior and terminal sizing.

Capture dashboard/session, view generation, attach request identity, session epoch
where applicable, and expected process before creating commands. Give ordinary attach
requests an ordering identity as well as launch/lifecycle attaches. Close results on
navigation, replacement/removal, rejected generations, superseded requests, error,
shutdown or delivery after frontend close. An old A→B→A result cannot win by ID alone.
Define ownership from request through queued result to running terminal command;
application Close must cancel outstanding requests and dispose late resources even
if Bubble Tea is no longer consuming messages. No coordinator access in closures.

### 3. Deletion boundaries and async interaction

Implement deletion/dashboard mutation contracts and service projections. Migrate
`dashboard_list.go`, card command helpers, callbacks, messages and view consumers.
Replace `dashboardDeletionSummary`/`dashboardDeletionStatus` backend arguments with
contract-to-presentation mapping. Carry execution progress/outcome through result
messages instead of discarding it. Preserve existing confirmations and retry/status
interaction without visual redesign.

Capture target dashboard/card, view visit/generation and distinct request IDs for
prepare, confirm, status and retry. Reject stale results and repeated confirm input;
expired/changed impact requires a new preflight. Background commands own the existing
bounded deletion contexts; navigation may invalidate presentation but cannot imply
rollback or erase a persisted intent. Shutdown cancellation must finish owned workers
before backend teardown. Keep blocking calls out of Update and callbacks.

### 4. Finish constructor and production dependency cleanup

Define root consumer interfaces for attachment, dashboard/deletion and startup/error
inputs alongside existing read/editor/session/configuration/center ports. A constructor
may accept a capability bundle of cohesive interfaces; avoid one enormous interface.
Main supplies the existing service implicitly; backend must not import UI to declare
conformance. Use contract/stdlib types in every reachable method/result.

Remove `legacy_actions.go`, concrete manager fields and unused constructor parameters.
Read `sm.LoadErrors()` in main and supply the nonfatal initial diagnostic through the
frontend options/runtime error mechanism; show it after startup so an initialization
keypress cannot erase it. Retain operation-journal stderr diagnostics. Audit helpers,
exported methods, aliases, embedded fields, callbacks and test-only exports for leaks.
Check shared artwork/helper graphs before retaining them as dependencies.

### 5. Move frontend into internal/ui and integrate main

Move App, Coordinator, navigation, handlers, refresh, center, callbacks and command
files into root `internal/ui`; keep models/messages/views/components/terminal as
children. Rename package clauses and migrate internal/external tests, export fixtures
and import references. Children never import root UI. Keep public behavior and Tea
model methods intact. Remove `internal/app` completely with no forwarding package.

Update main to call `ui.New` with consumer capabilities and startup diagnostics.
Expose a neutral `ui.RuntimeError(error)` message factory or equivalent frontend entry
point; main forwards MCP failure without constructing internal messages. Main retains
config lock/direct initialization, repository/driver/manager/service/runner/MCP
construction, endpoint publication and failure-watch ownership. Preserve startup
failure cleanup, frontend Close after Run, failure-watcher cancellation/join, then
runner shutdown → session shutdown → MCP shutdown → config-lock release. Do not move
policy, runner lifetime or process shutdown into views or terminal detach.

### 6. Remove exceptions and update current references

Remove every feature-7/migrated production edge from both architecture allowlists,
including transitive App/Coordinator/resource signatures. Do not rename exceptions
to new UI paths. The existing feature-8 `ui/uitest/shape.go` test-support exception is
separate; retain only if still exclusively test support, document it explicitly, and
add no production exception. Inspect checker behavior with the new root UI layout.

Update `docs/frontend-boundary.md`, `project.md`, README/spec/current test instructions
and active source references for new ownership/paths. Do not rewrite historical
workflow files or Appedia evidence. Record evidence in `feature_7_evidence.md` and
final independent review in `feature_7_review.md`; any future Appedia refresh must
preserve existing local evidence and use its own authorized workflow.

## Acceptance and verification mapping

All checks below are planned, not claimed passing results. Prefer controlled channels,
fake resources/terminal operations and existing backend fixtures; no native model is
needed. Clean up subprocess groups, managers and subscriptions before temp directories.

| Preserved task / acceptance | Required evidence |
|---|---|
| 1: independent attachment | Exact service/port signatures; no backend concrete types; initial/live detached bytes, cancellation during setup and stream read, idempotent close and partial initialization. |
| 2: backend resize and human terminal | Correct row/column propagation; actual backend PTY resize; Ctrl+K and Kitty detach; raw mode/protocol restoration on every exit/error; partial input never replayed. |
| Resource lifecycle | Buffered trailing output before EOF; concurrent Close/Next/write/resize/kill/relaunch; stale activation cannot affect replacement; detach leaves process alive; no pump/subscription leaks. |
| Stale results | Navigate, missing session, dashboard switch, A→B→A, overlapping manual/launch/lifecycle attaches, error with resource, shutdown with late setup; each unused resource released. |
| 3: deletion contracts | Detached nested summaries/progress; no manifests exposed; wrong owner/dashboard, missing/expired/stale token, changed hidden fields/reference impact, missing foreign acknowledgement rejected by backend. |
| Deletion recovery | Cancellation before/after intent; persistence failure; admission drain; pending temp cleanup; removed session row; retry after restart; foreign sessions/workspaces/global registries survive; missing row is not success. |
| 4: capability constructors | Root frontend accepts only focused capabilities/neutral diagnostics; no concrete backend fields, legacy bridges, transitive aliases or exposed handles; command closures remain race safe. |
| 5/7: package retirement | `internal/app` directory absent, all active Go imports updated, moved tests execute, UI child graph acyclic, architecture checks enforce root UI and backend direction. |
| 6: composition/lifetime | Nonfatal session-load warning, initial errors, MCP start/runtime failure, frontend cleanup and runner/session/MCP shutdown order; detach never owns process lifetime. |
| Preserved earlier slices | Inventory failures retain data; empty reads clear; center report cache scope/continuity/cancellation; editor revision/hidden fields; lifecycle warning/uncertainty; registry/source behavior and minimum-size/navigation regressions. |

Start with existing `session/attach_test.go` replay/input/broadcast tests,
`session/io_test.go`, app coordinator/navigation/multi-dashboard/command-race tests,
API confirmation/deletion/card-cleanup tests and
`ui/views/dashboard_card_deletion_test.go`. Audit their assertions; add focused
behavioral cases for uncovered semantics above. Migrate tests without weakening
coverage, especially `TestCommandClosuresDoNotRaceCoordinator` and feature-6 center
regressions. Test backend policy without constructing the frontend.

During early slices run affected app/UI/API/session packages, plus architecture
checks after exception removal. After the move run:

```bash
go build ./...
go test ./internal/ui/... ./internal/api ./internal/session ./internal/mcp
go test -race ./internal/ui/... ./internal/api ./internal/session
go test ./internal/architecture/...
go test ./...
go vet ./...
git diff --check
```

Apply repository changed-code lint policy and record actual commands/results and
existing debt separately. Confirm selected tests execute rather than reporting no
matches. Full architecture certification remains step 8; feature-7 checks must
already reject all remaining production boundary violations.

Manually verify isolated model-free PTY fixtures at 80×24 and larger: attach,
initial/live/trailing output, resize, Ctrl+K detach, reattach, process EOF and terminal
restoration after injected I/O failure. Also inspect deletion confirmation/cancel,
foreign-impact acknowledgement, pending retry and stale navigation behavior. Use
isolated config/temp roots as documented in project.md. Native model-driven E2E
requires the provider/model and total token-budget authorization in
`tests/e2e/AGENTS.md`; this plan grants none. Non-Unix checks remain explicit limits
unless actually run. Documentation-only planning requires no application test run.

## Risks, open questions and review gate

No blocking product question remains. Proposed method/type names can follow local
conventions while preserving the stated semantics. Largest risks are process-done
racing queued output, attachment cleanup after frontend shutdown, stale activation
resize/input, discarded deletion outcomes and cycles introduced by moving tests or
helpers. The targeted evidence above is required, not optional future verification.
Revisit the stream adapter if existing channel shutdown cannot preserve completion
ordering without a bounded owned pump; do not expand into an event framework.

After implementation, the assigned planning/review agent independently inspects
actual code, all callers/removals, contracts, backend policy, exception ownership and
actual command results. Record changed files, acceptance evidence, limitations and
explicit **PASS** or **CHANGES REQUIRED** in `feature_7_review.md`. Unrun checks are
unverified and failed criteria block advancement. Stop at this review gate; neither
creating this plan nor implementing it approves feature 7 or authorizes a commit.
