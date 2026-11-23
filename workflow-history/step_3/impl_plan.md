# Feature 3 implementation plan — separate card editing and save semantics

## Status and execution gate

Planning artifact only. No implementation, commit, feature approval or application
test result is implied. Scope comes from `feature_3.md` and `roadmap.md`; project
invariants come from `AGENTS.md` and `project.md`.

Feature 1 has a recorded PASS. **Feature 2's review prerequisite is satisfied:**
the orchestrator explicitly relayed independent Reviewer approval in Kankuro
reply `830f1ae8-0f2d-4887-9c58-c8d0e7154760` for message
`e486ec3c-53b0-56f2-99ed-3e5c01c4ca93`, followed by commit
`5403b2cf8fbe0a7044d37c448ea6337b6eb5ab78`. Git confirmed that commit is the
current HEAD. `feature_2_review.md` now records the follow-up approval in place
of its outdated pending status. Direct report retrieval was attempted but denied
because this activation lacks Dashboard Control; approval provenance is the
orchestrator's explicit relay, not an independently read report body.
Explicit implementation authorization is still required before executing this
plan. Do not fix feature 2 within feature 3.

Feature 3 must finish and receive its own final-review PASS before dependent
configuration editing (5). Center replacement (6) consumes identities without
changing saves. Keep intermediate slices buildable and the existing editor usable.
No storage migration, extra module, MCP wire change, center/preview change or
unrelated redesign is included. Preserve existing unrelated working-tree files.

## Checkout findings

- `internal/ui/views/card_editor.go` holds `loaded config.IDCard`, exposes
  `SetCard`/`Card`, and overlays fields to preserve `SystemInstruction`, `Env`,
  identity and timestamps. It also consumes concrete registry/settings types.
  `SetSettings` captures explicit deselection before a section becomes hidden.
- `internal/app/callbacks.go` retains the full expected card in
  `Coordinator.editingCard`; save callbacks call `CreateCardInDashboard` or
  `UpdateCardInDashboard`. `CardSavedMsg` transports another full card although
  the successful handler only refreshes and navigates.
- `internal/app/dashboard_list.go:dashboardCardCommand` and
  `messages.DashboardCardLoadedMsg` share full-card loading for edit and launch.
  The edit branch must split away; launch remains a feature-4 migration surface.
- `internal/api/membership.go:UpdateCardInDashboard` rechecks ownership and compares
  complete expected/current cards using `reflect.DeepEqual` inside
  `WithCardReferencesAndSnapshot`, then checks foreign references before writing.
  `cardOwnedBy` and membership inventory own deletion/availability checks.
- `internal/config/id_cards.go` owns write validation, identity/timestamps and
  immutable dashboard membership. Its repository mutex is acquired after manager
  admission. `validation.go` checks current registry references and can reject a
  save if a reference is missing; retaining an unresolved draft ID must not silently
  remove it or bypass that validation.
- Features 1–2 already provide independent `api/contracts`, UI inventory models,
  per-read success/failure results and scoped request ordering. Architecture
  exceptions live in `internal/architecture/allowlist.json`,
  `type_exposure_allowlist.json` and `docs/frontend-boundary.md`.

Recheck these locations and all callers after the prerequisite review; this plan
does not certify the later implementation checkout.

## Contract and ownership design

Use one cohesive consumer-owned card-editor capability beside current app
coordination, implemented implicitly by `api.Service`. Move it with coordination
in feature 7. Views receive data/drafts and emit save intents, never services.
Introduce `internal/api/contracts/card_editor.go` and
`internal/ui/models/card_editor.go`; names below describe the intended API.

| Type/operation | Contents and responsibility |
|---|---|
| `CardEditorData` / `LoadCardEditor(dashboardID, cardID)` | Selected identity, owner fact, edit availability/reason, opaque revision and editable values only. Permit visible shared consumer cards to load read-only; inaccessible/deleting cards fail closed. |
| New-editor read | Explicit application defaults, create availability and editor choices; empty identity/revision. May share a focused options read with existing-editor load. |
| `CardEditorChoices` | Independent workspace, skill, persona, MCP and puppet option data needed by existing labels/details/modals, with independent attempted/value/error results. No stored records or command/header/environment secrets. |
| `CardEditorDraft` | UI-owned editable values, selected IDs (including unresolved IDs), edit intent, optional-section visibility and local interaction state. Identity/revision are separate read metadata. |
| `CreateCardInput` | Target dashboard plus explicitly supported initial editable values. No caller-provided card ID, owner, timestamps, environment or system instruction. |
| `UpdateCardInput` | Dashboard/card identities, expected revision and `CardChanges`. No whole stored record. |
| `CardChanges` | An explicit `Change[T]{Present bool; Value T}` per editable field. Absent preserves current value; present false, empty string and empty list intentionally update it. |
| Save result | Saved identity, refreshed revision and structured status/reason (including conflict requiring reload); no full storage card. Preserve meaningful persistence errors. |

Editable fields are Name, Driver, Model, WorkspaceID, SubWorkspaceIDs, FaceStyle,
Puppet, UseTmpWorkspace, DashboardControl, Shared, AgentsMD, AgentsMDPath,
AgentsMDPreferPath, SkillIDs, PersonaIDs and MCPIDs. Keep existing text trimming
and local validation behavior. Application validation remains authoritative.
IDs, membership and timestamps cannot be mutated through changes.

Define create defaults explicitly from existing behavior: empty text/reference
values, false toggles/sharing, default `kankuro` puppet, and empty selections.
Application/storage allocate identity and UTC timestamps. Centralize defaults in
the application read/create path rather than leaving competing UI defaults.

Every publication detaches mutable values: editor reads and choices, UI mapping,
draft setters/getters, outgoing inputs, command capture, changes applied to the
current record and returned results. Clone slices/maps and nested artwork buffers
where applicable; preserve nil/empty distinctions rather than normalizing blindly.
Contracts contain no backend aliases, embedded structs or framework types.

### Revision representation

Use a version-prefixed SHA-256 digest of a deterministic encoding of the complete
stored card, computed only by the application. Do not use timestamps, form-only
fields, ordinary YAML serialization or `omitempty` JSON as the revision source.

Encode persisted fields with field names/type framing, sorted map keys, ordered
slice elements, explicit nil versus empty markers, and stable timestamp values.
Exclude only explicitly transient storage-excluded fields such as DeletionState;
deletion state is revalidated independently during admission. Derive field
coverage from the storage type so newly added uneditable stored fields participate
without updating a UI mapper. Unsupported new kinds must fail closed, never be
silently skipped. Document the versioned encoding and test its field coverage.
Hash the loaded record before forming the editable projection and recompute from
fresh current storage at save. Empty, malformed, unknown-version and mismatched
revisions reject updates with reload guidance. The revision is concurrency
identity, never proof of permission. No persisted revision column is needed.

### Atomic application save

Keep the existing manager admission -> repository lock order. Within
`WithCardReferencesAndSnapshot(cardID, ...)`: load/validate fresh membership,
require active owner dashboard and available nondeleting card, compare the
revision, clone current storage, apply only present changes, check the resulting
sharing state against foreign current sessions, saved bindings and pending
activations, then perform the repository's validated atomic write. Return the
revision of the actual saved record including its new UpdatedAt. Create uses
the same admission boundary with active-dashboard revalidation and builds a new
storage card from defaults plus create input.

Do not split revision checking and writing into independent service operations,
move admission into the UI, reenter repository locking methods while holding its
mutex, or acquire manager admission from inside a repository lock. Audit all
in-process card writers when replacing the old methods: concurrent supported
mutations must share admission. Arbitrary external file editors are not given a
new filesystem compare-and-swap guarantee by this feature. Retain current storage
failure semantics and never automatically replay a failed/uncertain save.

### Selection preservation and hidden sections

Keep selection identity in the draft independently of currently displayed options.
A failed registry read retains the last valid options and surfaces its error;
a confirmed empty read clears options but does not deselect stored IDs. Incomplete
successful reads also preserve unresolved selected IDs. A visible section emits
the full intended selected-ID set, including unresolved IDs not explicitly removed;
an explicit clear emits present empty. Do the same for sub-workspaces and scalar
references such as workspace/puppet; never silently select a substitute on load.

An untouched hidden persona/MCP section emits no change. Preserve prior explicit
edits made while visible, including deselect-all, across a subsequent hide: retain
their intent in the draft and submit it as an intentional edit, not as a value
derived from hidden options. This preserves the existing
`TestCardEditor_HiddenSectionsCaptureDeselection` guarantee. Visibility alone must
never create an edit. Retain dirty selections when moving through source/MCP
screens; reset/new clears all draft intent, identity, revision and modal state.

If authoritative validation rejects a retained missing reference, keep the draft
and explain the failure. Do not weaken registry validation or report success by
dropping that reference. Availability and UI read-only controls are feedback;
saves always revalidate policy.

## Ordered implementation slices

1. **Contracts, revision and backend operations.** Add the independent contracts,
   revision encoder, focused editor/default/options reads and create/update save
   operations. Reuse existing membership/admission/validation helpers. Add backend
   policy, preservation and revision regressions first. Keep the old wired editor
   usable only while this migration is in progress; no permanent duplicate API.
2. **Draft and presentation inputs.** Add UI draft/choice/visibility models and
   mapping, preserving existing labels and behavior. Refactor CardEditor and its
   workspace/puppet/capability modal inputs to independent presentation data.
   Replace `SetCard`/`Card` and storage-overlay state with editor-data loading and
   explicit draft intent. Adapt helpers/getters/tests that expose storage records.
   Use focused editor settings facts rather than `config.AppSettings`. Add only
   the editor-facing registry adapter needed here; broader settings/source/MCP
   configuration read/write workflows remain feature 5.
3. **Wire the entire asynchronous editor slice.** Split edit loading from the
   legacy launch detail command/message. Add the consumer-owned editor interface
   and commands using captured service, dashboard/card identity, view generation,
   editor epoch and monotonic request ID. Replace `editingCard`, full-card callback
   arguments and `CardSavedMsg.Card` with draft inputs and save results. Opening
   new/edit, cancel, explicit reload and successful save advance/invalidate the
   editor epoch. Apply only results matching captured identity and latest request.
   Failed loads/saves retain valid draft data and surface errors. Conflicts keep
   edits visible and provide explicit reload; do not retry or overwrite silently.
   Prevent duplicate in-flight saves, especially duplicate creation. A late
   successful save may invalidate inventory, but cannot navigate or replace a
   newer editor. Execute all blocking reads/writes in commands, never Update.
4. **Remove replaced paths and reconcile boundaries.** Migrate API tests/callers
   of old full-record create/update methods, remove their obsolete comparison
   transport and imports, and remove edit use of ReadDashboardCardDetail. Keep
   launch-only detail until feature 4. Search all callers rather than relying on
   listed files. Remove obsolete editor overlay regressions only after equivalent
   backend preservation and UI-intent regressions exist. Reconcile exact direct
   and transitive exceptions and `docs/frontend-boundary.md`; remove every retired
   editor edge, without broadening allowances or moving payloads into aliases.
   Some coarse historical owner labels cover unrelated configuration callbacks:
   document and precisely retain their feature-5 ownership rather than rewriting
   those workflows here. No full-record editor exception remains.
5. **Integrate and stop at review.** Update current architecture/project invariant
   documentation once behavior actually changes, perform the checks below, and
   hand the actual diff and results to the assigned planning/review agent. Do not
   advance features or commit. Record its independent verdict separately in
   `feature_3_review.md`.

## Acceptance and evidence map

| Requirement | Required evidence and likely test home |
|---|---|
| Save carries edits and revision only | Contract inspection plus editor callback/command tests; search messages/coordinator for full-card editor payloads; architecture checks. |
| Preserve unseen/current/future fields | API updates changing one visible field preserve SystemInstruction, Env, CreatedAt, ID and owner; reflectively check every unedited storage field, replacing UI overlay coverage from `card_editor_test.go`. |
| Complete concurrency detection | Same revision for map insertion-order variations; different revisions for each stored-field mutation, Env/SystemInstruction changes, booleans, list order and meaningful nil/empty values; stale update cannot write. Two saves from one revision produce exactly one success. |
| Presence semantics | API table tests for absent and present empty/false on every editable field family; UI tests for clear-all, hidden untouched fields and visible-edit-then-hide; no reliance on nonempty list length. |
| Missing registry choices | Failed, incomplete and empty choice reads retain loaded IDs; choices recover without resetting draft; explicit user removal works. Extend `preview_broadcast_test.go` editor regressions and view tests. |
| Authority and atomic admission | Adapt `membership_internal_test.go`, `card_cleanup_internal_test.go` and `delete_card_guard_test.go`: foreign owner denial, referenced unsharing rejection for sessions/bindings/pending starts, tombstones/deletion, inactive dashboard, lost authority since load and concurrent deletion/save. Assert no write on rejection. |
| Correct create/new/reset | Application allocates identity/timestamps and rejects invalid dashboard; UI reset after existing/read-only card clears revision/ID/dirty values/modals; repeated save input cannot submit twice while pending. |
| Async and mutable ownership | Coordinator.Update tests with gated fake editor ports for reversed loads/saves, cancel/reopen/same-card epoch changes and dashboard changes; mutate published inputs/outputs after capture; race check command closures. |
| Failure behavior | Failed read retains valid editor and reports error, confirmed empty choices clear options only, validation/write failures retain edits, conflict requires reload, cancellation/uncertain writes never trigger automatic replay. |
| UI continuity | Adapt `card_editor_internal_test.go`, `card_editor_test.go`, `card_editor_setsize_test.go`, layout tests, integration/navigation tests and source-return paths; preserve supported 80x24 layout and keys. |

Use existing service/lifecycle fixtures and test cleanup. Synchronize concurrency
tests with explicit gates, not sleeps. Tests must demonstrate policy/data loss or
stale-result failures, not merely mirror type mapping.

## Planned validation and final-review gate

The following commands are **planned and unrun** for this documentation task.
Run focused tests as slices land, then at the integration gate:

```sh
go test ./internal/api ./internal/ui/views ./internal/app -count=1
go test ./internal/config ./internal/ui/models ./internal/ui/components -count=1
go test ./internal/architecture/... -count=1
go test -race ./internal/api ./internal/app -run 'Card|CommandClosures' -count=1
go build ./...
go vet ./...
go test ./...
golangci-lint run
git diff --check
```

Verify regex-selected tests actually include the new revision/save/admission
regressions; run additional explicitly named cases if needed. Record actual
commands/results, not inherited feature-2 results. Feature 2 documents existing
full-lint debt in checker tests; recheck the current diagnostics and distinguish
that debt from new findings rather than claiming full lint passes.

Manually exercise create, owner edit, foreign read-only, hidden/cleared selections,
conflict/reload and source-screen return with isolated config/tmp directories.
No model-driven E2E is needed to plan or approve this slice; any such run requires
the separate provider/model and total-token-budget authorization documented in
`tests/e2e/AGENTS.md`.

The assigned final reviewer must inspect actual code, every migrated/remaining
caller, removed paths, contract dependency graphs, atomic backend policy tests,
ownership behavior and recorded command results. Record PASS or CHANGES REQUIRED,
changed files, acceptance evidence, remaining exact exceptions and limitations in
`feature_3_review.md`. Unrun checks remain unverified; failed criteria block
advancement. The feature-2 prerequisite is satisfied by the relayed independent
approval documented above; direct report retrieval remains an authorization
limitation. Planning is complete, and implementation awaits explicit instruction.
