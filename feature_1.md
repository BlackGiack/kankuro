# Feature 1 — Establish contracts, boundaries and enforcement

Status: planned; not implemented or reviewed.

Ordered prerequisite: None; start here.

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

Existing frontend dependencies cross persistence, process and policy ownership.
Establish an enforceable migration boundary and the smallest contracts needed for
feature 2. Keep runtime behavior intact. This feature does not convert all screens.

## Checkout evidence and affected callers/tests

- `main.go` constructs `app.NewCoordinator(repo, manager, service)`.
- `internal/app/coordinator.go` retains repo/manager/service and preview state;
  `refresh.go`, `scoped_refresh.go`, `callbacks.go` and `session_actions.go` use them.
- `internal/ui/messages/messages.go` imports config/session; `api/service.go` and
  `api/messaging.go` expose implementation models/aliases. Those aliases are not
  suitable frontend contracts.
- Inspect `internal/app/command_race_internal_test.go`,
  `internal/app/multi_dashboard_internal_test.go`, and `internal/api/membership.go`.
  Record affected production files and exported signatures, not only package pairs.

## Planning decisions and migration mechanics

Use `internal/api/contracts`, `internal/ui/models`, and root consumer interfaces
in `internal/ui/ports.go` at the final layout; temporary ports beside the existing
coordinator are allowed until feature 7. Group capabilities by dashboard reads,
card editing, session actions, configuration management, center inspection and
attachment/deletion where coherent. Helpers accept narrower subsets; views get
presentation data and intents. Reject both a universal Backend and one interface
per method. The service implements implicitly with exact Go signatures; main
injects it without backend importing UI. Add contracts only as callers migrate.

Use a normal Go architectural test under an architecture-test package (e.g.
`internal/architecture`) so `go test ./...` runs enforcement. Inspect all production
Go files, including inactive platform/build-tag files. Check direct imports and
recursively inspect exported contract types/method signatures, aliases and embeds.
Tests may deliberately integrate real backends only under explicit test-file
classification; no production allowlist may be hidden by test exemptions.
Maintain exact file/symbol/import exceptions with owner features, reject unused
exceptions and additions, and prove forbidden import and indirect type leakage
fail the checker using controlled fixtures. A directory move alone is not proof.

Feature removal ownership: reads/list models -> 2; card editor/full-record saves
-> 3; lifecycle/conversation/workspace actions -> 4; registries, source/MCP/settings,
fingerprints and config errors -> 5; preview component/messages/output wiring -> 6;
attachment/deletion/concrete constructor/coordinator -> 7; final zero-exception
audit and unused checks -> 8. Inventory exact instances before approving feature 1.
No new preview abstraction is permitted.

Decisions specified in later features refine the agreed direction: canonical
full-record revisions and presence-aware edits (3), outcome variants (4), source
expected-state tokens and opaque change tokens (5), bounded center reads (6),
resource lifetime (7). Approve their semantics before implementing dependent APIs.
Mapping costs are accepted to isolate ownership. Revisit if the design duplicates
policy, needs new persistence infrastructure or grows a universal snapshot.

## Sequence and coverage map

Implement/review `feature_1.md` through `feature_8.md` in numeric order. The roadmap
permits some parallel work, but this execution plan deliberately uses serial
review gates. Original minimum dependencies remain below. Feature 6 replaces the
preview before feature 7 moves the coordinator. All steps remain complete slices.

| Source requirements | Implementing plan |
|---|---|
| Roadmap step 1; sections 2–4 boundary and ownership conventions | feature_1.md |
| Step 2; section 4 session/read models | feature_2.md |
| Step 3; section 4 card preservation and revisions | feature_3.md |
| Step 4; section 4 action/outcome semantics | feature_4.md |
| Step 5; section 9 source identity/change tokens | feature_5.md |
| Step 6 and all of feature-dbcenter-rework.md | feature_6.md |
| Step 7; section 4 attachment resource ownership | feature_7.md |
| Step 8, section 6 validation and section 7 final criteria | feature_8.md |

## Preserved roadmap requirements

The following workstream is carried forward without dropping tasks or acceptance criteria.

Dependencies: none.

Tasks:

1. Inventory frontend production dependencies, exported model leakage, operational
   calls and existing tests. Recheck the evidence in section 2 against the checkout.
2. Record current user-visible behavior and the relevant concurrency, scope and
   persistence guarantees for each workflow.
3. Establish `api/contracts`, the presentation model location and consumer-owned
   capability interface conventions. Define only contracts required by the first
   slices; do not speculate about remote transports.
4. Agree revision, partial-update, action availability, outcome and resource
   ownership semantics before dependent features implement them.
5. Add a dependency check that covers production imports and contract leakage.
   Initially allowlist exact existing violations, with owning removal steps. New
   violations must fail immediately. Account for build tags/platform files.
6. Document existing preview-only dependencies for removal in step 6. Do not
   create a new long-lived preview contract.

Acceptance:

- The boundary is precise enough that another agent can determine where new code
  belongs and which types it may expose.
- Contracts are independent types, not aliases or embedded backend structs.
- The application builds and existing behavior is unchanged.
- Dependency enforcement runs as a normal check; temporary exceptions are explicit.

Review focus: avoid premature public API proliferation and unused abstractions;
verify the check catches a representative forbidden import/type leak.

## Shared target ownership and contract specification


### Package responsibilities

Proposed structure, to be refined in feature planning:

```text
main.go                         composition and process lifetime
internal/
  api/
    contracts/                  application inputs, read results, outcomes,
                                resource contracts and shared semantic values
    ...                         application policy and backend adapters
  config/                       configuration storage and discovery
  session/                      process/activation/workspace ownership
  driver/                       CLI integrations and serialized process input
  ui/
    model.go                    root Bubble Tea model and coordination
    navigation.go               frontend navigation and view ownership
    commands.go                 application calls adapted to tea.Cmd
    ports.go                    consumer-owned capability interfaces
    mapping.go                  application results to presentation models
    models/                     rows, form drafts and presentation data
    messages/                   frontend events and asynchronous results
    views/                      screens and interaction state
    components/                 reusable rendering and interaction
    terminal/                   raw human terminal interaction during attachment
```

Do not create a package for every file or helper. `terminal` is justified by its
terminal I/O responsibility; command/mapping files can stay in the root UI package
unless independent package boundaries provide a demonstrated benefit.

`internal/api/contracts` is the default location for stable application-facing
types. Do not rename the whole `api` package as part of this work without a
specific need. Avoid creating two competing public application APIs.

### Permitted direction

```text
main -> ui and concrete backend construction
ui -> api/contracts
ui -> its own models/messages/views/components/terminal packages
api -> api/contracts, config, session and other backend dependencies
mcp -> api and application contracts as appropriate
```

The frontend production code must not import `config`, `configstartup`, `session`,
`driver`, `mcp`, or the concrete `internal/api` implementation package. Contracts
must not import any frontend package, Bubble Tea, Lip Gloss, repositories,
managers or drivers. Backend production code must not import frontend packages
or terminal presentation frameworks.

These rules apply to indirect model leakage too: no aliases to backend models,
embedded snapshots, concrete handles, or methods returning implementation-owned
types inside otherwise neutral contracts. Inspect exported fields and signatures,
not just interface declarations.

Pure rendering helpers such as artwork utilities may remain shared if they carry
no backend implementation dependency. Verify the actual dependency graph rather
than applying a blanket prohibition to every package outside `ui`.

The composition root is intentionally exempt from frontend import restrictions.
It must still construct and wire concrete implementations. It should no longer
inject a repository or manager directly into the frontend constructor. Expose a
small frontend notification entry point for runtime failures if needed, so main
does not construct implementation-specific UI messages to report MCP failure.

Within the frontend, root coordination may import child packages. Views and
components must not import the root coordinator package. Put shared presentation
types and events in child packages to avoid cycles.

### Interfaces and implicit implementation

Define a small set of cohesive, consumer-owned interfaces where the frontend
calls application capabilities. Examples are dashboard reading, card editing,
session actions and configuration management. Avoid one enormous `Backend`
interface and avoid an interface for every trivial method.

The concrete service should satisfy these interfaces implicitly. `main` passes
the service into the UI constructor; the application service does not import the
frontend to declare conformance. Helpers can accept narrower interfaces than the
root constructor needs. Views and components normally receive models and
callbacks/intents, not services.

Illustrative signatures only; settle names and fields during planning:

```go
type cardEditor interface {
    LoadCardEditor(context.Context, string, string) (contracts.CardEditorData, error)
    SaveCard(context.Context, contracts.SaveCardRequest) (contracts.CardSaveResult, error)
}
```

Implicit interfaces do not remove data coupling. Return and argument types must
also use independent application contracts. Go requires exactly matching method
signatures: a method returning `*session.AttachHandle` does not implement a method
returning an attachment interface. Deliberately change or adapt that method.

Use packages in the existing Go module. Directory nesting does not enforce
architectural ownership; automated import checks must enforce the chosen rules.

## 4. Application contracts versus presentation models

Application contracts describe operations, facts, identity, revision and policy
outcomes. Presentation models describe the data and state useful to a screen.
Storage structs describe persistence. They are related but have different reasons
to change.

| Concern | Application contract | UI presentation model |
|---|---|---|
| Dashboard inventory | Dashboard identity, details, application lifecycle facts and available actions | Dashboard row, display status, help/action state |
| Session inventory | Scoped session summary, registration facts, active conversation summary and available actions | Session row with resolved card display data and presentation fields |
| Card inventory | Scoped card summary, capability references/details and ownership/action facts | Card row with resolved capability names, appearance/artwork reference and action state |
| Card editing | Editable fields, revision, edit permissions, available options; explicit changes on save | Form draft, selected options, focus, modal state, local validation feedback |
| Conversation selection | Options, identity, compatibility and reason; explicit new/resume request | Sorted/filtered choices, selection, detail scrolling and help text |
| MCP management | Summary/actions and separate editable configuration | List row and form draft; command arguments/maps can have text-entry representations |
| Settings | Small settings read/write contract | Settings form state; fields may initially closely resemble the contract |
| Launch/relaunch/switch | Structured effect, identity, warning and failure information | Navigation/attachment decision and displayed warning/error |
| Dashboard center | Recent session-authored reports or workspace Git/directory entries, read status and warnings | Report/file rows, timestamp formatting, available-space clipping, loading/stale/error indicators |
| Attachment | Authorized stream/input/resize/close capability | Terminal mode, detach key and terminal output interaction |

Do not pass terminal styles, width-dependent text, focus, keyboard shortcuts or
Bubble Tea values into application contracts. Keep escaping, display sanitization,
clipping, colors and rendering at the frontend boundary. Do not pretruncate data
in application results.

### Session models

The session manager's snapshot includes fields such as instruction ownership,
activation state, delivery readiness, workspace metadata and all card-local
conversation bindings. The session list should not receive this complete object.

Define a narrow summary sufficient for the existing non-preview screen. Preserve
the currently displayed PID, workspace identity, registration information and
active conversation. Resolve the relevant conversation rather than making a
component search the manager's map representation. Resolve card names/appearance
in mapping code rather than giving a row renderer the whole registry.

If another workflow needs more data, give it a focused read contract; do not grow
the session row back into a universal snapshot. Preserve existing searching and
sorting behavior, including identity-based selection and deleted-card display.

### Card editing and revision safety

The existing full-card overlay is necessary under the current contract. Replacing
it with a newly constructed card would lose data. Replace the contract before
removing the overlay.

Separate editor reads, submitted changes and stored cards. On save the application
must load the current record, verify scope/ownership and expected revision, apply
only intended changes, and preserve other fields. The UI must not carry
timestamps, immutable ownership or uneditable fields merely to round-trip them.

Specify presence semantics for every optional update. A hidden section means
unchanged; a submitted empty list means clear. Ordinary zero values are not enough
to distinguish these cases. Apply the same care to booleans and fields whose empty
string is a valid edit. Keep form visibility distinct from authorization.

Replace the full original card used for `reflect.DeepEqual` with an opaque
revision mechanism owned and verified by the application. Choose a mechanism
that detects the same concurrent changes, including changes to fields not exposed
by the editor. A canonical content fingerprint is one possible implementation;
do not assume timestamps alone provide equivalent protection. A revision is not
authorization. Preserve the existing atomicity/locking guarantees and rejection
of stale edits. Avoid introducing a persistence schema change for revision data.

### Action availability and operation outcomes

The application derives action availability and meaningful reasons from its
policy. The frontend uses these facts to present or disable actions and chooses
keyboard/navigation behavior. Execution must revalidate current state: an action
being enabled in a read result does not authorize a later mutation.

For example, the application determines whether a session can be relaunched,
whether the selected dashboard may edit a shared card, and whether a saved
conversation is compatible with the destination driver. It returns availability
and reasons through application contracts. The UI decides how to present those
results: disable an action, display the reason, or adjust help text. It must not
independently reconstruct eligibility from lifecycle, ownership and deletion
fields. Otherwise a backend rule change could leave the frontend applying an
outdated rule. The same application policy should serve UI and MCP operations
where their rules overlap, while preserving their different caller authorities.

Rechecking at execution time is intentional and remains in the application:
state may have changed since the UI read availability. Do not treat a UI-supplied
availability flag as authoritative or remove backend validation because the
screen already disabled an action.

Focus, navigation, keyboard shortcuts, form visibility, selection and confirmation
interaction remain frontend rules. Local form feedback may help users correct
input, but authoritative validation remains in the application. Form visibility
does not grant permission, and a confirmation dialog does not replace the
application's confirmation-token and scope checks.

Avoid duplicating the lifecycle state machine in presentation models. A semantic
status for display and a coherent set of action facts are sufficient. Unknown
statuses must have defined presentation behavior; backend extensions should not
silently enable unsafe actions.

Expose structured lifecycle outcomes so the UI does not identify successful
process effects by inspecting `session.ErrPersistFailed`. Represent a usable
activation with a metadata warning separately from a failure before spawn. State
which combinations are valid; do not invent several independent flags allowing
contradictory outcomes. Preserve cancellation and post-effect evidence, including
uncertain outcomes, without retrying process/input effects automatically.

### Resource ownership and copy semantics

Contracts must define whether slices/maps/byte buffers are transferred, copied or
immutable after publication. Do not accidentally share mutable backend state
with UI commands or form drafts. Clone at the appropriate ownership boundary;
do not deep-copy indiscriminately at every layer.

An attachment contract must hide the PTY file and offer input, output, resize,
initial output/header, completion and release behavior as required by the current
implementation. Closing the frontend attachment releases its subscription; it
must not accidentally terminate the agent. Input continues through the driver's
serialization gate. Specify stale activation, EOF, cancellation, resize failure,
idempotent close and cleanup of unused/stale asynchronous results.

## Focused validation and final-review evidence

Run focused regressions while changing the slice, then these checks at its integration gate (adapt moved package paths after feature 7). Commands below are planned, not claimed results. Confirm test selection actually runs relevant cases; add behavioral regressions where coverage is missing.

```bash
go test ./internal/architecture/...
go build ./...
git diff --check
```

Require evidence for dependency checker rejection fixtures, exact allowlist ownership and unchanged baseline behavior.

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
