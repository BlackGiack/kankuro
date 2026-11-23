# Feature 4 — Separate session interactions and outcomes

Status: planned; not implemented or reviewed.

Ordered prerequisite: `feature_3.md` must receive final-review PASS before starting; also honor the source dependencies below.

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

Frontend helpers interpret manager errors, conversation bindings and workspace
lookups. Move those semantics to checked application operations and return focused
effects/options. Prerequisites in this ordered plan: approved features 1–3 (the
original minimum is 1–2). Attachment resources remain in 7; center replacement is 6.

## Checkout evidence and affected callers/tests

- `internal/app/session_actions.go` interprets ErrPersistFailed and performs
  CheckSessionInDashboard followed by manager List/workspace command execution.
- `callbacks.go:requestRelaunch/requestSwitch` and `views/conversation_choice.go`,
  `session_launcher.go`, `agent_switch.go`, and `ui/messages/messages.go` expose
  lifecycle/binding details.
- `internal/api/membership.go` supplies Launch/Relaunch/Switch/Stop/Remove/Rename
  and checked membership; `internal/session/lifecycle.go`, `conversation.go` and
  `ide_launcher.go` own effects.
- Revalidate `internal/app/session_warning_test.go`, `conversation_choice_internal_test.go`,
  `terminal_launch_test.go`, `inline_rename_refresh_internal_test.go`,
  `internal/api/membership_internal_test.go`, `internal/session/conversation_choice_internal_test.go`,
  `lifecycle_test.go`, and `internal/ui/views/session_launcher_test.go`.

## Contract and migration decisions

Use one discriminated lifecycle result with documented valid variants: no effect
(fatal/cancelled before effect), usable activation (identity, optional metadata
warning), and uncertain effect (known evidence, warning, no auto-retry). A usable
variant must have identity; no-effect must not imply attachment; cancellation
following an effect must retain that evidence. Unexpected result/variant renders
an error, never enables unsafe actions. Avoid independent contradictory booleans.
Backend maps ErrPersistFailed and other implementation evidence once; UI chooses
navigation/attachment and warning display. Preserve current MCP results via its
adapter rather than leaking backend sentinels into the new UI API.

Conversation options contain identity, display facts, compatibility and reason.
A choice explicitly discriminates new versus resume(binding identity), with no
backend binding maps or UI comparison of drivers. Revalidate selected binding,
destination compatibility and activation under existing admission immediately
before the effect. Relaunch preserves saved workspace/deployed settings; switch
applies destination configuration and preserves workspace identity.

IDE/terminal application operations accept dashboard/session identity and command
kind; resolve saved actual workspace and configured behavior at the checked effect
boundary. Eliminate UI check-then-use and concrete manager lookup. Missing paths
remain errors, not newly created directories. Preserve original/temporary paths,
argument parsing and missing-command feedback. No shell interpolation is added.
Migrate helpers, callers, messages and views, then delete old policy/error paths.
Stop/remove/rename remain dashboard scoped. Configuration/MCP helper dependencies
belong to 5 and raw attach handling to 7; enumerate only those remaining exceptions.

## Preserved roadmap requirements

The following workstream is carried forward without dropping tasks or acceptance criteria.

Dependencies: steps 1–2; coordinate card identities with step 3.

Tasks:

1. Migrate launch, relaunch, switch, rename, stop and remove requests/results to
   application contracts and consumer interfaces.
2. Introduce structured lifecycle outcomes; remove frontend interpretation of
   manager-specific persistence errors from views, messages and command helpers.
3. Provide conversation options with compatibility/reasons from application logic.
   Represent new/resume selection explicitly without exposing session bindings.
4. Revalidate conversation identity and compatibility during execution under the
   existing admission rules; preserve workspace identity on relaunch/switch.
5. Put IDE/terminal workspace operations behind application methods. Eliminate
   the frontend's manager lookup and its separate check-then-use implementation;
   preserve scope, actual saved workspace path and configured command behavior.
6. Keep frontend navigation, confirmation interaction and attach-after-success
   decisions in the UI. No blocking work may run in a view update callback.

Acceptance:

- The frontend uses no session/driver types or error sentinels for these workflows.
- A usable process remains usable after a metadata warning, with a visible warning.
- Fatal failures do not trigger attachment; uncertain effects are not replayed.
- Conversation selection and dashboard-scoped actions remain revalidated.
- Original/temporary workspace identity, missing-directory behavior and switch
  versus relaunch semantics are preserved.

Review focus: preserve actual effect evidence, cancellation behavior and locking;
do not reclassify every warning as either complete success or complete failure.

## Focused validation and final-review evidence

Run focused regressions while changing the slice, then these checks at its integration gate (adapt moved package paths after feature 7). Commands below are planned, not claimed results. Confirm test selection actually runs relevant cases; add behavioral regressions where coverage is missing.

```bash
go test ./internal/api ./internal/app ./internal/session ./internal/ui/views
go test -race ./internal/api ./internal/app ./internal/session -run 'Launch|Switch|Conversation|CommandClosures|Workspace'
git diff --check
```

Require evidence for spawn plus persistence warning, fatal/uncertain/cancelled effects, conversation replacement after choice, scoped actual workspace operations and missing directories.

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
