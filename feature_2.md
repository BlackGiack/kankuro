# Feature 2 — Separate dashboard, card and session reads

Status: planned; not implemented or reviewed.

Ordered prerequisite: `feature_1.md` must receive final-review PASS before starting; also honor the source dependencies below.

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

Lists receive broad storage/process objects and reconstruct policy/display joins.
Replace dashboard/card/session inventories and their complete read-to-render paths
with focused contracts and presentation rows. Card writes, lifecycle effects,
configuration management and preview removal remain owned by features 3–6.
Prerequisite: approved feature 1. This precedes the center replacement (6).

## Checkout evidence and affected callers/tests

- Checked human reads are `api.Service.ReadDashboardData` and `ListDashboards` in
  `internal/api/membership.go`, including deletion manifests and tombstones.
- `internal/app/scoped_refresh.go` combines scoped reads with registry data and
  `repo.Fingerprint`; `refresh.go` preserves nil on per-registry read failures.
- `internal/ui/messages/messages.go`, `views/dashboard.go`, `views/dashboard_list.go`,
  `components/card_list.go` and `components/session_list.go` carry broad models.
- Revalidate `internal/app/scoped_refresh_error_internal_test.go`,
  `refresh_error_test.go`, `refresh_cadence_internal_test.go`,
  `multi_dashboard_internal_test.go`, `internal/ui/components/card_list_test.go`,
  `session_list_test.go`, and both list layout suites.

## Contract and migration decisions

Return independent summaries with semantic status and action availability/reasons;
unknown statuses render safely with unavailable actions unless explicitly enabled.
Include currently displayed names, PID, saved workspace identity, registration,
active conversation and artwork identity, plus deleted-card/deletion-recovery facts.
Keep all card-local binding maps and process instructions out of list rows.
The application resolves the active conversation; UI mapping resolves names and
appearance using focused capability choices, then passes complete rows to renderers.
Preserve sorting/filtering and identity selection, including after row replacement.

Use explicit per-read success/error results instead of relying on nil versus empty
once contracts change. A successful empty slice clears; an error retains previous
values independently for each registry. Global choices needed now must already
come through an application capability; their source-edit operations finish in 5.
Use an opaque comparable registry change token; failed fingerprint reads cannot
advance its baseline. Poll cadence remains UI-owned. Feature 5 completes coverage.
Publish snapshots with owned copies, never pointers/maps into manager state.

Migrate API reads, mappings, messages, setters, callbacks and callers together,
then delete old read paths and renderer joins. Keep legacy preview-specific
payloads and subscription hooks only where still required, with exact exceptions
owned by feature 6; never send broad new rows merely to satisfy preview. Card-edit,
action/configuration and attachment exceptions keep their named owners (3/4/5/7).

## Preserved roadmap requirements

The following workstream is carried forward without dropping tasks or acceptance criteria.

Dependencies: step 1.

Tasks:

1. Expose dashboard-scoped read contracts over the existing checked read path.
   Preserve dashboard deletion recovery and deleted-card information.
2. Introduce dashboard/card/session presentation rows. Move registry joins and
   active conversation selection out of list renderers.
3. Provide application-derived action availability. Preserve execution-time checks.
4. Convert the relevant messages, setters, callbacks and selections to the new
   types; actions should generally carry identities rather than whole rows.
5. Keep request generations, session epochs and dashboard identities on asynchronous
   UI envelopes. These are frontend stale-result controls, not backend policy.
6. Preserve refresh failure semantics: failed reads must retain prior data, while
   successful empty results must clear it. If replacing nil-slice conventions,
   make the new distinction explicit and test it.
7. Move global registry reads needed by these screens behind application reads,
   or introduce their focused contracts for completion in step 5. Hide filesystem
   fingerprints behind opaque application change tokens; keep polling cadence in
   the UI and fingerprint implementation in the backend.

Acceptance:

- Dashboard and list rendering consume focused models without config/session
  imports, excluding only legacy preview plumbing scheduled for removal in step 6.
- Existing names, PID/registration/conversation details, sorting, filtering,
  selection preservation, artwork and deleted-card behavior remain available.
- Refreshes cannot update another dashboard/view or erase data after failed reads.

Review focus: accidental loss of fields currently used outside rendering, mutable
map/slice sharing, inconsistent action facts and unintended scope widening.

## Focused validation and final-review evidence

Run focused regressions while changing the slice, then these checks at its integration gate (adapt moved package paths after feature 7). Commands below are planned, not claimed results. Confirm test selection actually runs relevant cases; add behavioral regressions where coverage is missing.

```bash
go test ./internal/api ./internal/app ./internal/ui/components ./internal/ui/views
go test -race ./internal/app -run 'TestCommandClosuresDoNotRaceCoordinator|Dashboard|Refresh'
git diff --check
```

Require evidence for cross-dashboard/view in-flight refresh, failed versus empty per-registry reads, field/search/sort preservation and copy isolation.

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
