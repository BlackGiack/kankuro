# Feature 8 — Final architectural and behavioral verification

Status: planned; not implemented or reviewed.

Ordered prerequisite: `feature_7.md` must receive final-review PASS before starting; also honor the source dependencies below.

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

A buildable migration can still leak types, change policy or leak resources.
Prove the complete separation and behavioral criteria with actual code and runtime
evidence. Prerequisites: approved features 1–7; feature 6's center and preview
removal must already be complete. No new functionality is included in this audit.

## Checkout evidence and affected callers/tests

The pre-migration entry points are `main.go`, `internal/app`, `internal/ui`,
`internal/api`, `internal/config`, `internal/session`, `internal/driver` and
`internal/mcp`. By this step internal/app must be absent and its orchestration
and command-race tests relocated. Use the feature 1 dependency inventory and all
feature review records as a starting index, then independently inspect the final
checkout and production/build-tag/platform files. Review project.md, README.md,
kankuro_specs.md and appedia references for changes required by final ownership.

## Audit decisions and removal scope

Delete migration exceptions/shims, not the enforcement that revealed them. Check
neutral contracts recursively (aliases, embedding, resource methods, exported
signature returns and mutable ownership), and prove forbidden imports/type leaks
still fail. Real-backend integration tests must remain explicitly identifiable;
frontend unit tests use small cohesive capability fakes. Reject moved broad
snapshots or convenience backend fields as proof of separation.

Check scope/authorization/admission and MCP behavior at backend seams, not solely
through frontend tests. Compare all feature acceptance criteria, the failure policy
in feature 6 and the final definition below with observable evidence. No implicit
approval for changes to MCP wire contracts, history retention, storage format or
other product behavior is created by passing tests. Unrelated issues are recorded
without expanding the migration; unmet required criteria remain blockers.

Update project.md to describe implemented ownership, constructor, package map and
startup/shutdown behavior. Keep task status, limitations and test logs in review
artifacts. Produce a final report mapping every final criterion to code and checks,
including actual command output/result, manual scenarios, unresolved issues and
scope of verification. Only the designated final reviewer can mark this complete.

## Preserved roadmap requirements

The following workstream is carried forward without dropping tasks or acceptance criteria.

Dependencies: steps 1–7.

Tasks:

1. Remove migration allowlists and compatibility shims. Check production imports,
   exported contracts, aliases and platform-specific/build-tagged files.
2. Ensure unit tests for frontend orchestration use small fakes implementing
   application capabilities. Keep deliberate real-backend integration tests
   separately identifiable; their imports must not mask production dependencies.
3. Review public contracts for backend fields added merely for convenience,
   presentation concerns in backend results and shared mutable data ownership.
4. Verify behavior across dashboard scope, refresh failure, stale results, card
   preservation/concurrency, lifecycle warnings, conversations, deletion and attach.
5. Run final appropriate build/static/test/race checks and isolated terminal checks.
6. Update `project.md` and relevant architecture/onboarding references to describe
   the final implementation. Keep task logs and review evidence out of project.md.
7. Produce a final review report with checks run, results, known limitations and
   explicit evidence for each completion criterion.

Acceptance:

- Every final criterion in section 7 is satisfied with evidence.
- No legacy exception is silently carried into a claim of full separation.
- Any remaining unrelated issue is documented without expanding this migration.

## Required validation strategy


Validate each vertical slice when it changes; do not postpone all testing until
step 8. Use the current repository's required checks and skills for implementation.
Typical commands are:

```bash
go build ./...
go vet ./...
go test ./...
go test -race ./...
golangci-lint run
git diff --check
```

Start with focused tests. Lifecycle, attachment, shared-state or command changes
need relevant race coverage, including the existing coordinator command race
backstop after it is moved. Broaden checks at integration points and final review;
avoid repeatedly running unchanged expensive suites without a reason.

Tests should verify behavior and contracts, not merely mirror mapping assignments.
Priority scenarios:

- Another dashboard/view is selected while a read/save/action is in flight.
- A stale attachment result is discarded and its subscription is released.
- A registry read fails versus successfully returns no items.
- A card save preserves unexposed fields and hidden selections; explicit clear
  works; a concurrent backend edit rejects the stale save.
- Shared-card ownership/unsharing constraints and deletion admission remain intact.
- A process starts but metadata persistence fails; the UI still offers usable flow.
- A conversation is replaced or becomes incompatible between selection and action.
- Source ordering changes between form load and mutation.
- Source configuration saves while derived synchronization fails.
- Built-in MCP edit/delete is attempted through both UI and application entry points.
- Attach resizes, detaches, handles process exit/error, and restores terminal state.
- The center panel shows reports authored across session activations, switches from
  files to reports on receipt, refreshes workspace data every 30 seconds without
  reports, clips content to available space, preserves it through stop/relaunch
  and exposes read errors alongside labelled fallback.
- Application shutdown still drains operations, sessions and MCP in the right order.

Use fake application capabilities for frontend behavior and backend tests for
policy/ownership/persistence. Add integration coverage where the seam itself is
material, especially attachment and post-effect outcomes. Preserve MCP wire and
authorization coverage when shared service implementation changes.

Manual terminal runs must isolate configuration and temporary workspaces as
specified by `project.md`. Native-agent runs additionally need isolated user
configuration. Native model-driven E2E runs are not automatically authorized by
this roadmap: follow `tests/e2e/AGENTS.md` and the documented provider/model and
total token-budget approval requirements before running them.

No commits unless explicitly requested. Preserve unrelated working-tree changes.

## Final definition of done — required evidence matrix


1. All frontend production code, including the former `internal/app`, lives within
   the frontend boundary and depends on independent application contracts.
2. Frontend code imports no configuration/session/driver/MCP implementation or
   concrete application service. No public contract leaks those types indirectly.
3. Backend production code imports no UI package or presentation framework.
4. `internal/app` and temporary migration bridges are removed.
5. Views/components consume focused presentation data and emit identities/intents;
   they do not round-trip full storage objects or interpret backend error sentinels.
6. Application logic owns validation, authorization, compatibility, mutations and
   action availability; operations revalidate rather than trusting stale UI facts.
7. Card preservation and concurrent-edit safety are at least as strong as before.
8. Attachment hides backend resources while preserving serialization and cleanup.
9. Existing product behavior and failure guarantees are preserved except for the
   explicitly included preview replacement. The new center meets step 6 criteria;
   any further intentional behavior change is separately identified and approved.
10. Automated dependency enforcement and meaningful behavioral coverage protect the
    boundary. Old preview code and exceptions are removed without removing shared
    attachment functionality.
11. Main remains the concrete composition root, and documentation matches the final
    ownership and startup/shutdown design.

## Focused validation and final-review evidence

Run focused regressions while changing the slice, then these checks at its integration gate (adapt moved package paths after feature 7). Commands below are planned, not claimed results. Confirm test selection actually runs relevant cases; add behavioral regressions where coverage is missing.

```bash
go build ./...
go vet ./...
go test ./...
go test -race ./...
golangci-lint run
git diff --check
```

Require evidence for every final definition-of-done item with repository-wide dependency checks, backend policy evidence and isolated terminal scenarios.

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
