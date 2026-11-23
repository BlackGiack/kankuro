# Feature 3 — Separate card editing and save semantics

Status: planned; not implemented or reviewed.

Ordered prerequisite: `feature_2.md` must receive final-review PASS before starting; also honor the source dependencies below.

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

The editor currently overlays a loaded full `config.IDCard`. Simply dropping that
object loses unexposed data and concurrency detection. Replace the read/edit/save
vertical slice, moving preservation into authoritative atomic application saves.
Prerequisites: approved features 1–2; finish before configuration editing (5).
Center replacement (6) consumes the resulting identities without changing saves.

## Checkout evidence and affected callers/tests

`internal/ui/views/card_editor.go` implements SetCard/Card overlays;
`internal/app/callbacks.go` and coordinator/messages carry full cards.
`internal/api/membership.go:UpdateCardInDashboard` compares expected/current with
reflect.DeepEqual within manager card-reference admission before repository write.
Recheck `internal/config/id_cards.go`, `internal/session/card_references.go`,
`internal/api/membership_internal_test.go`, `card_cleanup_internal_test.go`,
`internal/ui/views/card_editor_internal_test.go`, `card_editor_test.go`,
`card_editor_setsize_test.go`, and `internal/app/delete_card_guard_test.go`.

## Contract and migration decisions

Define separate editor data, UI draft, create input, and update changes plus
opaque expected revision. Application-generated IDs, ownership and timestamps are
read facts or storage responsibilities, never round-tripped editable fields.
Use an application-owned digest of deterministic complete stored-card content,
including unexposed fields; sort map keys, retain meaningful nil/empty distinctions
and version the encoding. Timestamp-only revisions are insufficient. No persisted
revision column or schema migration. Verify current content and apply changes
inside the existing admission/locking operation, not in separate check/write calls.
Do not weaken concurrent-update detection or lock ordering while adapting it.

Every editable field uses explicit presence: absent = unchanged; present empty
string/list or false = an intentional edit. This applies to booleans and zero values
as well as optional sections. Hidden form sections omit edits; visible cleared
selections submit present empty lists. Keep loaded IDs whose choices are missing
because a registry failed or returned an incomplete set. Preserve SystemInstruction,
Env and all uneditable/current/future storage fields by applying changes to the
fresh current record. Clone mutable draft and change values at ownership boundaries.
Create has explicit defaults; new/reset must erase prior loaded identity/revision.

Authoritative saves verify dashboard ownership, revision, sharing/unsharing foreign
references and deletion admission before mutation. UI read-only/visibility state
is feedback, not permission. Translate conflicts into an actionable reload result.
Migrate all editor callbacks/messages/coordinator expected-card state, then remove
the UI overlay and full-card comparison transport. Keep no permanent legacy editor
API; adapt other backend callers only if required without changing MCP wire rules.
No exception remains for full-record editor payloads after this feature.

## Preserved roadmap requirements

The following workstream is carried forward without dropping tasks or acceptance criteria.

Dependencies: steps 1–2.

Tasks:

1. Introduce editor data, form drafts, explicit changes and opaque expected revision.
2. Move overlay/preservation into application save logic. Cover `SystemInstruction`,
   `Env`, immutable identity/ownership, timestamps and future uneditable fields.
3. Preserve hidden persona/MCP selections and explicit clearing. Preserve loaded
   references missing from a failed or incomplete registry read.
4. Keep shared cards read-only for consumers and enforce owner-only mutation in
   the service. Preserve restrictions on unsharing referenced cards.
5. Keep create and update semantics explicit. The application allocates identity
   and timestamps and validates the target dashboard.
6. Replace full-card payloads in editor callbacks, coordinator state and save
   messages. Remove obsolete full-card concurrency plumbing.

Acceptance:

- A save request contains intended edits and concurrency identity, not an entire
  storage record.
- Unexposed/hidden fields survive saves; explicit empty selections are saved.
- Concurrent changes, including unexposed-field changes, reject stale saves.
- Ownership, sharing and deletion/admission checks remain enforced atomically.
- Editor reset/new-card flows do not inherit a previous card's state.

Review focus: this is a data preservation and concurrency change, not just type
mapping. Require targeted behavioral tests before approving the feature.

## Focused validation and final-review evidence

Run focused regressions while changing the slice, then these checks at its integration gate (adapt moved package paths after feature 7). Commands below are planned, not claimed results. Confirm test selection actually runs relevant cases; add behavioral regressions where coverage is missing.

```bash
go test ./internal/api ./internal/ui/views ./internal/app
go test -race ./internal/api ./internal/app -run 'Card|CommandClosures'
git diff --check
```

Require evidence for hidden/unexposed-field preservation, explicit clear/false/empty edits, unseen-field concurrency conflicts, owner/unsharing/deletion admission and new-editor reset.

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
