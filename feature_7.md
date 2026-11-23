# Feature 7 — Close resource boundaries and retire internal/app

Status: planned; not implemented or reviewed.

Ordered prerequisite: `feature_6.md` must receive final-review PASS before starting; also honor the source dependencies below.

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

Raw attachment handles/PTYs and remaining deletion/backend dependencies prevent
frontend independence. Close resource/confirmation contracts and move frontend
coordination under internal/ui, deleting internal/app. Prerequisites: approved
features 1–6; the center must already have replaced and removed preview plumbing.
Feature 8 verifies the completed architecture without carrying legacy bridges.

## Checkout evidence and affected callers/tests

`internal/app/attach.go` uses session.AttachHandle, creack/pty and raw-mode/input
pumps; `session/manager.go:AttachHandle` and `session/attach.go:AttachChecked` own
initial output, subscription and serialized writer. `api/membership.go:AttachInDashboard`
returns the concrete handle. `ui/messages/messages.go:ScopedAttachReadyMsg` carries it.
Deletion adapters span `app/dashboard_list.go`, `callbacks.go`, `ui/views/dashboard.go`,
`api/card_confirmation.go`, `dashboard_confirmation.go`, `dashboard_deletion.go`
and config deletion manifests. `main.go`, `app/app.go`, `coordinator.go`, `navigation.go`,
`handlers.go` and command files are the composition/move callers.
Revalidate `internal/session/attach_test.go`, `io_test.go`,
`internal/app/coordinator_test.go`, `navigation_test.go`, `multi_dashboard_internal_test.go`,
`command_race_internal_test.go`, `internal/api/dashboard_deletion_internal_test.go`,
`card_cleanup_internal_test.go`, and `internal/ui/views/dashboard_card_deletion_test.go`.

## Contract and migration decisions

Use an application-owned attachment interface with initial output/header,
read-only output/completion stream, serialized input writer, Resize(rows, columns)
and idempotent Close. No PTY file, driver, manager or session concrete handle may
appear in public signatures. Change/adapt the service return type deliberately:
Go does not covariantly implement an interface return using *session.AttachHandle.
Backend adapter retains the exact subscribed activation and copies/transfers owned
initial/output bytes. Writes continue through the existing driver's input gate.

Close releases only that attachment's subscription/waiters; it never kills the
session or closes shared process resources. Define stale/replaced activation as a
closed/unusable resource with meaningful error; EOF completes output; context
cancellation unblocks pending waits. Preserve trailing output/completion ordering.
Idempotent close must work after partial initialization and alongside completion.
Unused or obsolete asynchronous attach results are closed immediately even on
navigation, missing sessions or rejected generations. Resize backend calls the
PTY mechanism and surfaces errors; input/output/resize failures restore terminal
mode and clean up without automatic effect replay. Frontend ui/terminal owns
raw mode, detach Ctrl+K, terminal size observation and keyboard protocol restoration.

Deletion contracts expose human summary, opaque revision/scope-bound confirmation
token, foreign-reference confirmation requirements and recoverable operation
outcome/progress. Never transport manifests or infer success from missing rows.
Execution revalidates scope/revision, persists intent, closes admission, drains,
checkpoints and retains retry state under existing backend owners. Missing/stale
confirmation cannot authorize deletion; UI interaction never replaces service checks.

Implement and migrate attachment and deletion callers first. Audit every remaining
frontend import and exported signature; remove unused repository/manager fields
and legacy APIs, then move frontend files/tests into internal/ui with acyclic child
packages. Keep application policy in api/session. Update main constructor to
consumer capabilities and expose a neutral frontend runtime-error entry point.
Preserve config lock/init, runner/session/MCP shutdown ordering and initial error
handling. Update references and remove internal/app entirely, no forwarding package.
Remove all migrated allowlist entries now; only verification work remains in 8.

## Preserved roadmap requirements

The following workstream is carried forward without dropping tasks or acceptance criteria.

Dependencies: steps 2–6.

Tasks:

1. Introduce and implement the application attachment contract from section 4.
   Replace concrete handles in frontend messages and terminal code.
2. Move PTY resize mechanics behind the backend contract; keep human terminal raw
   mode, keyboard protocol restoration and detach handling in `ui/terminal`.
3. Finish deletion/confirmation boundaries for cards and dashboards. UI receives
   summaries, tokens and operation outcomes rather than persistence manifests.
   Preserve revision-bound confirmation, foreign-reference confirmation and retry.
4. Audit all remaining frontend imports, signatures and operations.
   Remove legacy bridges and unused repository/manager fields and constructors.
5. Move the now-separated coordinator, navigation, commands and handlers from
   `internal/app` into `internal/ui`. Keep root/child dependencies acyclic.
6. Update `main` to inject application capabilities and forward runtime failures
   through a frontend entry point. Preserve initialization and shutdown ordering.
7. Move/update tests and documentation references; remove the obsolete package.
   Do not leave a permanent forwarding `internal/app` package.

Acceptance:

- `internal/app` is retired, and its frontend behavior lives under `internal/ui`.
- Production frontend constructor dependencies are application capabilities only.
- No raw backend PTY/manager/driver crosses the application boundary.
- Attach/detach, terminal restoration, resize, output completion and stale-result
  cleanup work without resource leaks or terminating the session on detach.
- Card/dashboard deletion remains recoverable, scoped and confirmation-bound.
- `main` retains composition and lifetime ownership; no backend imports the UI.

Review focus: input serialization, attachment ownership, goroutine/subscription
leaks, terminal restoration on errors and accidental business logic moved into UI.

## Focused validation and final-review evidence

Run focused regressions while changing the slice, then these checks at its integration gate (adapt moved package paths after feature 7). Commands below are planned, not claimed results. Confirm test selection actually runs relevant cases; add behavioral regressions where coverage is missing.

```bash
go build ./...
go test ./internal/ui/... ./internal/api ./internal/session ./internal/mcp
go test -race ./internal/ui/... ./internal/api ./internal/session
git diff --check
```

Require evidence for initial/live attach output, serialized input, resize/detach/EOF/error restoration, stale handle disposal, deletion revision/foreign-confirmation/retry and shutdown cleanup.

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
