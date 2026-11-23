# Feature 4 implementation plan — separate session interactions and outcomes

Status: planning only, 2026-11-14. No implementation or commit is authorized by
this document. Final review of feature 4 remains required before advancing.

## Scope and prerequisite evidence

Read `AGENTS.md`, `project.md`, `roadmap.md`, `feature_4.md`,
`feature_3_review.md`, and the current boundary documentation and relevant code.
`feature_3_review.md` records explicit PASS and readiness for feature 4; recheck
that approval and the checkout before implementation. Features 1–3 are ordered
prerequisites. Preserve their read, availability, card revision and save guarantees.

This slice migrates launch, relaunch, switch, rename, stop, remove, conversation
selection, and IDE/terminal workspace commands. Preserve current interaction and
MCP behavior. No storage migration, new module, remote service, MCP wire change,
center redesign, or attachment-resource redesign is included. Never commit
without an explicit request. Preserve all unrelated tracked and untracked files.

## Current checkout and migration targets

| Current location | Evidence and intended disposition |
|---|---|
| `internal/api/contracts/{inventory,card_editor}.go` | Independent facts and editor contracts already exist. Reuse identities and `Availability`; do not expand session rows into lifecycle snapshots. |
| `internal/app/read_ports.go` | Consumer-owned inventory/editor interfaces provide the convention for new session action capabilities. |
| `internal/app/session_actions.go` | Holds concrete service/repository/manager; interprets `session.ErrPersistFailed`, checks dashboard separately, scans `Manager.List`, and calls workspace command execution. Replace session workflows with contract-based commands; isolate remaining MCP deletion helpers for feature 5. |
| `internal/app/callbacks.go` | Stop/remove/rename already defer effects to commands; relaunch/switch synchronously inspect manager bindings and derive destination drivers. Replace those reads and choices with focused application reads. Stop passing command settings from UI into workspace effects. |
| `internal/app/coordinator.go`, `handlers.go` | Launch/switch completion branches interpret persistence errors. Conversation opening/refresh inspects manager state, card rows and bindings. Replace with contract results and scoped asynchronous refreshes. |
| `internal/ui/views/conversation_choice.go` | Exposes bindings/choices and compares driver strings. Consume options and availability; preserve sorting, filtering, full-ID detail scrolling, new choice, cancellation and selection identity. |
| `internal/ui/views/session_launcher.go` | Stores `config.IDCard`, exposes it through `Card`, and interprets persistence errors. Use a focused launch presentation model and structured result. Preserve launch/name-entry/attach flow and warning visibility. |
| `internal/ui/views/agent_switch.go`, `internal/ui/messages/messages.go` | Remove driver-as-policy arguments, backend choice types, raw lifecycle error interpretation and legacy whole-card launch payloads. Keep operation identities and frontend navigation intent. |
| `internal/api/membership.go` | Reuse scoped policy and card/activation publication admission for effects. Retire the legacy `ReadDashboardCardDetail` frontend adapter once its launch callers disappear. |
| `internal/session/lifecycle.go`, `conversation.go` | Own choice validation, activation reservation, publication and persistence evidence. Keep locking and workspace/deployment semantics here. |
| `internal/session/workspace_launcher.go`, `ide_launcher.go` | Reuse argv parsing, `$PATH` replacement, actual working directory, detached process reaping and configured-command feedback. |
| `internal/api/operations.go`, MCP adapter/callers | Audit legacy outcome handling and retain existing MCP wire results through backend adaptation. Do not route human actions through agent credentials or create a new durable operation system. |

Existing production exceptions are not permission to retain migrated coupling.
Some conversation callback/coordinator type-exposure exceptions currently carry
owner 5; remove these in this feature despite that label.

## Contract and ownership decisions

Prefer focused methods on the existing service with independent types in
`internal/api/contracts/session_actions.go` and `conversation.go`. Put cohesive
consumer interfaces beside current coordination (for example,
`internal/app/session_ports.go`), moving them only with feature 7. Helpers receive
the narrow interface they consume; views receive presentation data and callbacks.
Main wires concrete service implementations. Backend code must not import UI.

Keeping current backend-shaped signatures would leave UI policy coupled to
sentinels and bindings. Moving helpers alone into the coordinator has the same
problem. A new generalized operation bus adds unnecessary lifetime and replay
semantics. Focused contracts over existing checked policy are the smallest useful
change. The cost is explicit backend mapping and migration of all callers; retain
legacy backend-only signatures only where actual MCP/domain consumers need them,
with a shared implementation rather than two independent policy paths.

### Lifecycle outcomes

Use one discriminant with three variants, validated at publication and consumption.
Do not combine independently meaningful success/started/uncertain booleans. The
proposed names below may follow local naming conventions without changing rules.

| Variant | Required evidence | UI behavior |
|---|---|---|
| `NoEffect` | Fatal diagnostic or cancellation before activation; no usable activation payload. Preparatory filesystem effects/cleanup warnings must still be described if present. | Show failure/cancellation; do not attach or rename as a successful launch. No automatic effect retry. |
| `UsableActivation` | Session and activation identity captured from the effect, optional metadata warning. | Preserve current launch/name flow or relaunch/switch attachment decision; visibly surface warning without disabling the usable result. |
| `UncertainEffect` | Known session/activation identity when available, bounded effect evidence and diagnostic; never invent an identity. | Show uncertainty, refresh observable state, do not automatically attach or retry. |

Use variant-specific payloads or a validated tagged struct with documented allowed
fields. Unknown tags, missing usable identity, contradictory payloads, and zero
results fail closed with a UI error. Error details published to the frontend are
contract diagnostics (semantic code/message), not backend error objects requiring
`errors.Is`. Keep one result as the authority for effect classification rather
than splitting it across an outcome and contradictory error return.

Map `ErrPersistFailed` once in backend application mapping for these UI workflows.
Audit all launch/relaunch/switch exits and cancellation points. A post-effect
cancellation must retain known effect evidence, even when the caller context is
done; never overwrite a usable/uncertain result with a pre-effect cancellation.
Do not infer success from a nonempty ID, warning wording, status heuristics, or a
later unpinned manager lookup. Capture effect identity at publication; if current
return values cannot supply it, minimally enrich private manager effect evidence
and adapt legacy backend callers. A refresh cannot reconstruct what happened to
an earlier activation. The concrete manager changes below are required groundwork,
not a reason to stop the migration. Unpublished spawn with inconclusive teardown
is an existing path requiring uncertain-result handling.

Launch/relaunch/switch requests carry dashboard/card/session identities as needed,
context, and explicit choice. Rename/stop/remove use scoped requests and focused
mutation results distinguishing applied changes, diagnostics and any known partial
effect. Audit rename persistence and removal cleanup behavior before mapping them;
errors must not hide changes that already occurred. Do not generalize activation
outcomes to unrelated mutations solely for code reuse.

### Required first slice: publication-time evidence in the manager

The executor's reported gap is confirmed by source inspection. It is solvable
inside feature 4 without a product decision. Implement the following in order
before converting UI result handling. Names here specify a concrete design;
adjust spelling to local conventions without weakening the semantics.

1. **Introduce backend-only result types.** Add `internal/session/activation_result.go`
   with a value `ActivationResult` containing a single effect tag (not published,
   published, or uncertain), session identity, optional runtime process identity,
   observed PID if available, and the cause/warning error. A valid published result
   requires both identities. Keep preparation/rollback errors joined in the
   result's error, not lost through unnamed early returns. These are backend
   evidence types, never aliases in contracts. Public UI diagnostics are mapped
   by API policy later. Document that published means usable at publication,
   not guaranteed alive at return or CLI-ready. Preallocated session ID is a
   request target and never proves a process existed.
2. **Allocate an identity for every published process.** Existing
   `Session.ActivationID` is an MCP authority identity: `prepareActivationMCPConfig`
   allocates it only for selected Kankuro MCP, and `publishActivationLocked` clears
   it when token is empty. Do not require it for ordinary launches or change its
   authorization meaning. Add an unpersisted `Session.ProcessActivationID` and
   assign a fresh UUID in `activateSessionLocked` alongside the generation bump
   for every activation, including MCP-disabled ones. Never put it in SessionData,
   storage JSON or MCP wire structs. It is process correlation, not a credential.
   Reuse that exact value in evidence and any expected-process guard; do not mint
   a different UI identity later. Keep the existing MCP ID and generation logic.
3. **Return evidence through the single implementation path.** Change private
   `startSession` and `switchAgentWithChoiceContext` to named `ActivationResult`
   returns; propagate it through private launch/relaunch helpers. Add detailed
   checked manager entry points for launch/relaunch/switch returning this result.
   Keep existing exported methods as thin projections for real legacy callers:
   launch returns its historical session ID/error, relaunch/switch return the
   result error. Both surfaces call the same detailed implementation exactly once.
   In `internal/api/membership.go`, factor each existing scoped operation into
   a shared detailed backend helper preserving reservation, card resolution,
   publication check, name handling and deferred cancel/release. New UI methods
   map its evidence; existing human/MCP methods project legacy results. Do not
   duplicate admission policy or reconstruct evidence using `GetSession`.
4. **Capture at both publication sites before persistence.** In `startSession`
   (launch and relaunch) and `switchAgentWithChoiceContext`, retain the existing
   order: context/closed/scope/process-exit checks; `activateSessionLocked`;
   deployment/registration and workspace/header updates; publication flags;
   output tracking and watcher registration. Copy session and process identities
   and PID into the local published result under the same `m.mu` critical section,
   before `persistSessionLockedErr`. On record failure retain that result and
   attach the existing `ErrPersistFailed` wrapping for legacy projection. Success
   returns the same captured identities. Do not call lock-taking accessors from
   publication callbacks, expose `*Session`, or add application callbacks under
   the lock. Post-return mutation/termination cannot change the copied evidence.
5. **Preserve rollback and lifecycle accounting.** Keep `m.activations.Add/Done`,
   `m.launching` cleanup, publication checks, authority revocation and watcher
   ownership in their current order. Relaunch must release its activation
   reservation for every unpublished result, including uncertain cleanup, but
   never release a published one merely because its error is non-nil. Switch's
   `activated`/`published` guards and startSession's `committedInstructions` guard
   remain internal rollback controls; do not replace them with `err == nil`.
   Update deferred instruction restore and trust cleanup to join the named
   result's error while preserving its effect tag and identities. Never roll back
   a published process or delete its workspace on metadata failure/cancellation.
6. **Record spawn and teardown evidence before joining errors.** Both paths can
   spawn successfully and then reject publication due to context cancellation,
   shutdown, failed publication check, or closed `h.Done`. Today they call
   `teardownUnpublished`, which returns only `d.Kill()`'s error. Add a detailed
   private teardown helper that retains the current output-drain ownership and
   reports termination confirmation separately from cleanup warnings. Reuse the
   driver's observed terminated status, as `KillDetailed` does, and inspect
   `h.Done` only within its actual documented guarantees. A successful Spawn
   with confirmed teardown has no usable activation; report NoEffect with
   explicit spawned-then-terminated evidence and cancellation/cleanup diagnostics.
   If teardown cannot confirm termination, return UncertainEffect with the known
   target/PID and spawn/teardown evidence, never NoEffect or UsableActivation.
   No published process identity is required for this unpublished uncertainty.
   Keep cleanup failures visible and do not infer teardown success from a joined
   error string or erase uncertainty because temporary cleanup also failed.
   If a spawned process may still use its workspace, do not remove that temporary
   directory as if teardown succeeded. Retain the known path in backend evidence
   and actionable diagnostics; preserve revocation and best-effort process cleanup.
   This does not authorize a new persistent recovery queue or automatic relaunch.
   Audit Spawn's error contract: ordinary failed Spawn is pre-publication failure;
   if a driver can return a live handle with an error, retain and tear down that
   handle explicitly instead of dropping it. No driver API redesign is needed
   unless that audit demonstrates a real unsupported case.
7. **Cancellation has no result override.** Cancellation before spawn remains a
   NoEffect diagnostic; cancellation after spawn follows step 6; after publication
   it cannot erase the captured effect. The current code has no cancellation
   check after publication, so do not add one that turns success into failure.
   Preserve trust/instruction cleanup errors even for pre-spawn cancellation.
   Uncertain unpublished cleanup must not be automatically retried or disguised
   as a published activation. Preserve legacy error wrapping and current MCP
   projections; do not add lifecycle fields to MCP schemas/journals.
8. **Prove evidence before migrating frontend consumers.** Extend session tests
   using existing gate drivers and persistence failure fixtures. For each of
   launch/relaunch/switch prove nonempty process identity with MCP off, correct
   identity with MCP on, fresh identity on replacement, and the same published
   evidence on metadata failure. Gate cancellation/shutdown/publication rejection
   after Spawn and cover both confirmed and inconclusive teardown. Verify
   rollback errors survive. Capture a result, terminate/replace the session, and
   prove the returned identity remains unchanged. Assert no duplicate spawn,
   watcher, authority or activation-registration leaks. Keep legacy helper tests
   (`scoped_helpers_internal_test.go`, `scoped_helpers_external_test.go`,
   `dashboard_launch_test.go`) and operation/MCP warning tests passing.

Then map backend tags to the three contract variants once in API policy. Published
evidence maps to usable with optional metadata warning; unpublished uncertainty
maps to uncertain; other unpublished failures map to no usable activation with
their known effects/diagnostics. Never inspect current manager state to fabricate
the historical result. `KillDetailed` and `RemoveDetailed` already expose partial
effect facts for stop/remove; reuse them behind scoped policy rather than treating
all non-nil errors as no effect. Rename needs the separate persistence audit noted
above. These result types must not leak into UI or application contracts.

For attach-after-success, carry the process identity through existing scoped
messages and launcher naming/attachment intent. Add a narrow expected-process
check at `AttachChecked`'s subscription boundary (and corresponding scoped service
adapter), comparing under `m.mu` before consuming header/subscribing. Preserve
ordinary user attachment to the currently selected session. A replaced process
rejects stale automatic attachment. This only adds lifecycle correlation; keep
the current AttachHandle/raw terminal resources and their feature 7 exception.
Do not implement a new attachment resource contract in this slice. Close unused
stale ready handles as today. Test replacement between outcome and attach.

### Conversation reads and execution

Expose a scoped read for relaunch/switch preparation containing target identities,
action availability/reason, and detached conversation options. Each option has an
opaque selection identity, native ID/name/driver display facts, compatibility and
reason. Driver text is display-only. Return whether saved options exist so the UI
can preserve the existing prompt-only-when-needed interaction without a binding map.
Represent choice explicitly as new or resume(option identity); reject empty,
unknown or mixed variants. No public session binding aliases or maps.

Resolve current and destination card/driver in application policy. Ensure option
identity distinguishes equal native IDs across drivers and binds the session/card
context. The token is an identifier, not authorization. Validate it against the
current binding and destination compatibility under existing activation admission,
including changed/removed bindings and card changes between read and effect.
Preserve the existing matching-driver-after-same-ID regression. If rejecting
replacement of the same logical binding requires expected binding facts, carry a
deterministic opaque expected identity checked under admission, without persistent
schema additions. Do not reject harmless display-name refreshes unnecessarily.

UI sorting/searching/detail formatting remains local. Preserve selected identity
on successful refresh; missing/incompatible selection disables submission and
surfaces the application reason. Failed reads retain valid displayed options and
show an error rather than pretending they are empty; block submission while the
context is invalid. A confirmed empty read clears saved options. Opening,
refreshing and submitting all run application work inside commands, never Update
or callbacks. Execution always revalidates, including the fast explicit-new path.

### Workspace commands

Provide `OpenSessionWorkspace(ctx, request)` with dashboard/session identity and
command kind (IDE or terminal). The application loads configured behavior and the
manager's saved actual workspace at the checked effect boundary. UI does not
submit a command string, current-directory guess, or card workspace path.

Use a manager-owned checked workspace operation or short reservation to protect
session identity and workspace cleanup while validating scope and starting the
command. A relocated `CheckSessionInDashboard` followed by unlocked `List` and
execution is insufficient. Recheck active dashboard/session state and serialize
with deletion; keep repository locks short, release them before entering manager
admission, and never hold a broad lock through external process lifetime. Reserve
only through start as necessary; the externally opened IDE/terminal outlives the
request. Document cancellation-before-start versus already-started evidence.

Preserve original/temporary saved paths, `$PATH` substitution, quoting, argv
execution, missing IDE/terminal configuration messages, and child reaping. Missing
or inaccessible directories stay errors and are never created. No shell
interpolation. Validate unknown command kinds before effects. Application settings
reads for this operation do not require migrating the settings editor (feature 5).

### Asynchronous results and ownership

Capture port/context/request values on the coordinator goroutine. Every new
envelope carries dashboard, view generation, target session/card, operation/request
identity, and relevant launcher/picker epoch. Rapid same-view actions must not
allow an earlier completion to overwrite a later request. Stale completions must
not navigate, attach, rename a different target or replace newer options. Refresh
authoritative inventory as appropriate without replaying effects. Cancel obsolete
reads and prevent overlapping picker refreshes; cancellation is not rollback.

Application publishes detached slices and value-only diagnostics; no shared maps,
bindings or buffers. UI takes ownership or copies before mutating sorting/filter
state. Document this at contract/port boundaries and verify mutation isolation.
Raw attachment resources retain their current ownership until feature 7; this
slice must not create an attachment resource for a stale/fatal/uncertain result.

## Implementation sequence

Each integration point must build and remain usable. Add contracts with their
first backend consumer, migrate the complete workflow, then delete its old path.

1. **Outcome boundary and application operations.** Complete the eight ordered
   manager-evidence steps above, including the existing-wrapper compatibility
   tests, before migrating UI effect interpretation. Define outcomes/requests and backend mapping, then expose
   focused session mutation capabilities. Reuse membership/admission checks and
   shared policy. Add backend regressions for effect classification, scope,
   warning durability and legacy MCP equivalence before UI migration.
2. **Launch/relaunch/switch and simple mutations.** Migrate command helpers,
   callback wiring, messages, coordinator completion handlers, launcher and switch
   views to contracts and consumer ports. Replace whole-card launcher state with
   identity plus needed display fields. Remove `ReadDashboardCardDetail` and its
   obsolete refresh/message paths after checking all callers. Preserve optional
   naming, inline rename refresh, warning visibility, confirmation and navigation.
   Add the expected-process attachment guard described above before automatic
   attachment consumes the new lifecycle outcome.
3. **Conversation vertical slice.** Introduce application option reads and
   selection validation, then migrate opening/refresh/submission and picker
   presentation together. Remove direct manager reads, binding-map transport and
   driver comparisons. Add epoch/request filtering and live invalidation tests.
   Lifecycle methods must continue using existing admission/publication checks.
4. **Workspace vertical slice.** Add scoped configured IDE/terminal operation and
   manager effect reservation/check; migrate callbacks/commands and remove
   frontend workspace scanning and sentinel feedback. Verify concurrent deletion,
   saved path behavior, settings changes and child cleanup.
5. **Remove obsolete paths and tighten enforcement.** Remove migrated concrete
   imports, session/driver types and sentinels from workflow messages/views/helpers.
   Separate configuration-only helpers where needed so session adapters depend
   only on their ports. Audit every production/test caller and dead API. Update
   exact import and transitive exposure allowlists and `docs/frontend-boundary.md`;
   do not regenerate a broader allowlist or add wrappers hiding old leakage.
   Update current project documentation only where behavior/entry points changed.
6. **Integration evidence and review handoff.** Run the checks below, record real
   outputs and limitations separately, and stop for the assigned planning/review
   agent's independent `feature_4_review.md` decision. Do not advance or commit.

## Acceptance and regression matrix

| Requirement | Planned evidence |
|---|---|
| Independent requests/results and no backend lifecycle policy in frontend | Architecture checker plus inspection of every migrated signature/caller; contract-only fakes for UI lifecycle tests; no session/driver sentinels or binding models in these workflows. |
| Usable process survives persistence warning | Extend `internal/app/session_warning_test.go`, launcher tests and API effect tests for launch/relaunch/switch; verify real usable identity, visible warning, exactly one effect and correct attachment/name flow. |
| Fatal, cancelled, uncertain and invalid outcomes | Backend pre/post-effect cancellation and mapping tests; UI tests verify no unsafe attachment/replay, unknown tag and missing identity rejection; refresh cannot overwrite outcome evidence. |
| Conversation authority and freshness | Extend app/view picker tests and `internal/session/conversation_choice_internal_test.go`: same native ID across drivers, removed/replaced binding, changed destination driver/card, session reactivated, foreign scope and stale request/epoch. Assert rejected cases spawn nothing. |
| Workspace/deployment semantics | Lifecycle regressions for original/temporary identity, missing saved directory, relaunch saved deployment versus switch destination configuration; retain Dashboard Control/MCP deployment behavior. |
| Scoped rename/stop/remove | `internal/api/membership_internal_test.go`, lifecycle/deletion tests and `inline_rename_refresh_internal_test.go`: wrong dashboard, deleted dashboard, concurrent lifecycle/deletion, persistence/cleanup partial effects, captured identity and refresh behavior. |
| Checked IDE/terminal effects | Migrate `internal/app/terminal_launch_test.go` toward API policy tests with helper processes: actual temporary/original cwd, arguments/quotes/placeholders, empty/missing command, missing directory, unknown kind, foreign scope and deletion racing start. UI tests assert kind/identity forwarding and visible failures. |
| Async/data ownership | `TestCommandClosuresDoNotRaceCoordinator`, scoped stale completion tests for all mutations, same-view overlapping picker reads, late launcher naming, failed-read retention/confirmed-empty clearing, copied option slices. Drive UI through Update and follow returned commands. |
| Existing flow/render behavior | `internal/ui/views/session_launcher_test.go`, conversation/switch tests and coordinator navigation tests preserve naming, selection, filtering, full-ID detail, warning layout and view ownership. |
| MCP compatibility | Existing operation/MCP suites plus targeted equivalence checks if backend effect reporting changes; preserve historical results, warnings, activation pinning and uncertain-effect handling. |

Use existing API/service and app/session fixtures; stop runners/managers before
temporary-directory teardown and clean helper process groups on every failure.
Fault injection must exercise real effect boundaries where feasible. No tests
merely mirroring the new DTO fields in place of behavioral evidence.

## Validation and review gate

Planning validation: confirm paths and source requirements, inspect plan coverage,
and run whitespace checks. No application tests are required or claimed for this
documentation-only task.

During implementation, run focused regressions for each changed workflow, then:

```bash
go test ./internal/api ./internal/app ./internal/session ./internal/ui/views
go test ./internal/architecture/... ./internal/mcp
go test -race ./internal/api ./internal/app ./internal/session -run 'Launch|Switch|Conversation|CommandClosures|Workspace'
go build ./...
go vet ./...
git diff --check
```

Use `-list`/verbose output to confirm the regex actually selects the intended
regressions, including relaunch, cancellation, rename, stop/remove and stale-result
cases; run those explicitly if unmatched. Apply the repository lint check and
record its actual scope/results and any existing debt rather than implying full
lint success. Broaden tests when shared backend/MCP changes justify it.

Manual verification uses isolated `KANKURO_CONFIG_DIR` and `KANKURO_TMP_DIR` as in
`project.md`; native-agent experiments additionally isolate HOME. Verify launch,
name/attach, warning flow, stopped-session new/resume/switch, rename/stop/remove,
IDE/terminal actions and navigation during outstanding work. Native model-driven
E2E requires provider/model and total token-budget authorization under
`tests/e2e/AGENTS.md`; this plan grants none. Record unrun manual/E2E checks.

The final reviewer must inspect actual code, callers/removals, locking and
cancellation paths, backend policy tests, contract graphs, reduced exceptions and
actual command results. Record PASS or CHANGES REQUIRED, changed files, criterion
coverage, remaining exceptions and limitations in `feature_4_review.md`. Any failed
criterion blocks advancement. Creating this plan is not implementation approval.

## Remaining exceptions, risks and open items

- Feature 5 owns settings/source/MCP configuration helpers and remaining repository
  access. Feature 6 owns old preview snapshots/output subscriptions. Feature 7
  owns concrete attach handles/raw terminal I/O, composition constructor cleanup,
  deletion resources and retirement of `internal/app`. Enumerate remaining exact
  edges with these owners; none may cover migrated conversation/effect policy.
- Biggest implementation risks: losing publication identity by reading later
  state, flattening post-effect cancellation into failure, reentering manager or
  repository locks, and allowing workspace deletion between check and start.
  Validate these with focused concurrency/fault-injection tests before widening
  the UI migration. Do not change established admission ordering to fit a DTO.
- No unresolved product question blocks planning or the publication-evidence
  groundwork. The ordered design above resolves the executor's missing manager
  evidence dependency within this feature; implementation still requires the
  orchestrator's separate authorization. Workspace reservation helper names
  remain implementation choices. If backend evidence cannot classify a path reliably,
  report it as uncertain and document it for review rather than guessing/retrying.
- Revisit the contract shape only if inspection reveals a required existing
  effect that cannot be represented faithfully; report a scope-changing question
  through Kankuro before changing roadmap guarantees.
