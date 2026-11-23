# Feature 1 implementation plan

Status: planning complete; implementation and review have not started. This plan
does not authorize implementation, commits, or advancement to feature 2.

## Objective and scope

Implement `feature_1.md`: establish the frontend/application boundary, document
current behavior and migration exceptions, and enforce the boundary through a
normal Go architecture test. Preserve runtime behavior. Follow `roadmap.md` and
the invariants in `project.md`; implement features 1–8 serially with review gates.

This slice adds architectural documentation, package ownership documentation,
dependency enforcement and checker tests. It does not migrate screens, change
MCP schemas or storage, replace the preview, move the coordinator, introduce a
remote transport, or redesign the application. Preserve the existing untracked
requirements, AGENTS.md and .kankuro files. Never commit without an explicit request.

## Verified starting points

- `main.go:105` calls `app.New(repo, sm, sharedAPI)`. `internal/app/app.go` then
  calls `NewCoordinator`; the feature's direct-call example is shorthand.
- `internal/app/coordinator.go` retains concrete API, repository and manager
  dependencies. `refresh.go`, `scoped_refresh.go`, `callbacks.go`,
  `session_actions.go`, `dashboard_list.go`, `attach.go` and `preview.go` are
  migration entry points. `handlers.go` consumes the resulting UI messages.
- `api.Service.ReadDashboardData(string) (api.DashboardData, error)` already checks
  dashboard scope. Its result exposes `[]config.IDCard` and `[]session.Snapshot`.
  Preserve its policy boundary when feature 2 replaces its read representation.
- `ListDashboards() ([]config.Dashboard, error)`, card create/update methods taking
  `config.IDCard`, `AttachInDashboard(...) (*session.AttachHandle, error)` and
  relaunch/switch methods taking `session.ConversationChoice` leak implementation
  models to current consumers. `api/messaging.go` aliases session message types;
  those backend/MCP aliases are not frontend contract candidates.
- `ui/messages/messages.go` carries dashboard/card/config records, snapshots,
  conversation choices and an attachment handle. `DataLoadedMsg` and
  `SessionsRefreshedMsg` mix reads with legacy preview scrollback.
- List components, card/config forms, conversation picker and dashboard consume
  backend types. `ui/uitest/shape.go` is a non-`_test.go` helper with a config
  import: it requires explicit test-support classification, not silent exclusion.
- `internal/architecture`, `internal/api/contracts` and `internal/ui/models` do
  not currently exist. The module targets Go 1.27.3. No .github workflow directory
  was found; enforcement must run through `go test`, without assuming CI exists.

These observations establish scope, not a complete machine-readable allowlist.
Produce and review the exact inventory before accepting the enforcement baseline.

## Implementation sequence after authorization

### 1. Capture the baseline and exact migration inventory

Add a focused boundary document such as `docs/frontend-boundary.md`. Record each
production import violation as repository-relative file plus full import path.
Record each exposed backend type at its declaring symbol and field, parameter,
result, receiver or embedded member. Include operational calls, callback signatures,
UI message fields, and inferred/local dependencies that imports alone cannot
describe. Record ownership at symbol granularity where one file spans features.

Use `rg` for discovery and Go AST/type analysis for the authoritative inventory.
Scan both `internal/app` and `internal/ui`, including build-tagged and platform
files; inspect backend reverse dependencies and the transitive graph of shared
helpers such as `internal/ascii`. Record actual imports rather than assuming a
helper is pure. Inventory `_test.go` integration dependencies separately.

Assign every existing violation a removal owner:

| Owner | Existing surfaces to inventory |
|---|---|
| 2: reads and list models | `scoped_refresh.go`, dashboard inventory in `dashboard_list.go`, dashboard/list setters, `card_list.go`, `session_list.go`, renderer inputs, dashboard/card/session message fields |
| 3: card editor | `card_editor.go`, workspace/puppet selectors, editor callbacks and save messages; full-card expected/updated signatures in `api/membership.go` |
| 4: lifecycle and workspace actions | `session_actions.go`, launcher, agent switch, conversation picker, lifecycle messages, API conversation parameters and frontend persistence-error inspection |
| 5: configuration | `refresh.go` fingerprints and global registries, source/MCP/settings callbacks and views, source components, config errors and registry fields in messages |
| 6: legacy preview | `preview.go`, `session_preview.go`, preview portions of dashboard/coordinator/handlers/scoped refresh/messages and preview-specific tests |
| 7: attachment, deletion and composition | `attach.go`, attachment result messages, remaining deletion adapters, concrete `app.New`/`NewCoordinator` dependencies and coordinator fields |
| 8: final audit | Verify zero exceptions, removed paths and unused checks; do not postpone earlier owners' removal work to this feature |

A file/import exception shared by several surfaces lists every relevant symbol
and owner, and disappears when its last owned surface is removed. No package-wide,
directory-wide, wildcard or newly generated acceptance of violations is allowed.

### 2. Document workflow behavior that migration must preserve

Create a baseline matrix in the boundary document, linking implementation and
existing tests. For each workflow record visible behavior, policy owner, async
identity, failure/partial-effect semantics and mutable/resource ownership.

| Workflow | Guarantees and existing evidence to preserve |
|---|---|
| Inventory/refresh | Scoped reads; PID, names, workspace, registration and active conversation; sorting/filtering and identity selection; deleted-card display. Failed reads retain data, successful empty reads clear it. Inspect `scoped_refresh_error_internal_test.go`, `refresh_cadence_internal_test.go`, `multi_dashboard_internal_test.go`, `api/membership_internal_test.go`. |
| Commands/navigation | Capture inputs before `tea.Cmd`; preserve dashboard ID, generation, session epoch and request identity. No navigation inside a view Update or blocking repository work in callbacks. Inspect `command_race_internal_test.go`, `navigation_test.go` and coordinator tests. |
| Card edit/save | Full-card overlay preserves hidden/uneditable values today; stale edits and owner-only mutation remain checked under existing locking/admission. Inspect `card_editor.go`, editor tests and `api/membership.go` before replacing any payload. |
| Lifecycle/conversations | Application revalidates state, scope and driver compatibility. Preserve saved workspace identity, missing-directory failures and usable activations after metadata warnings. Inspect `session_warning_test.go`, `conversation_choice_internal_test.go` and API/session lifecycle tests. |
| Config/source/MCP/settings | Preserve atomic source updates, saved-but-sync-failed warnings, built-in MCP restrictions, escaping and defaults. Inspect `source_save_test.go`, `source_sync_warning_test.go`, MCP and settings error tests. |
| Preview/attachment | Keep preview working until 6; retain output subscriptions needed by attachment. Ctrl+K releases attachment without stopping the process; preserve serialized input, resize/EOF and stale-result cleanup. Inspect preview tests, `attach.go` and `session/attach.go`. |
| Deletion | Preserve confirmation tokens, scope, admission closure, durable intent, checkpointed retry and foreign-session preservation. Inspect API confirmation/deletion tests. |

Record discovered test gaps explicitly; a linked test filename is not proof that
every guarantee is covered or that it passed.

### 3. Establish ownership and contract conventions without unused APIs

Add package documentation at `internal/api/contracts/doc.go` and
`internal/ui/models/doc.go`, with detailed conventions in the boundary document.
This feature can establish package locations without shipping unconsumed DTOs,
empty interfaces or service adapters. Specify the first feature-2 read family
(dashboard inventory, scoped card/session summaries, action availability) in the
document, deriving its field checklist from actual non-preview consumers.
Implement concrete contract fields and mapping only with their migrating callers.

- Frontend production packages may use independent contracts and presentation
  packages. Forbid concrete `api`, `config`, `configstartup`, `session`, `driver`
  and `mcp`, including descendants except the explicit contracts subtree.
- Contracts contain application facts, identity, policy and outcomes. No backend
  aliases, embedded snapshots, repository/manager/PTY handles, UI/framework types,
  formatting, clipping or terminal styles. Backend code may depend on contracts,
  never frontend packages or terminal presentation frameworks.
- Presentation rows/drafts own display data, selection and local form state.
  Root coordination may import children; views/components cannot import the root
  coordinator. Shared types/events live in child packages to prevent cycles.
- Place cohesive consumer interfaces beside the current coordinator when first
  needed, moving them to `ui/ports.go` in 7. Service methods satisfy exact Go
  signatures implicitly; backend code must not import UI for conformance checks.
  Main remains the concrete composition root. Views receive models and intents.

Agree these semantics now; implement their APIs only in the owning slice:

| Concern | Decision |
|---|---|
| Revisions (3) | Opaque canonical full-record content revision, checked atomically against current storage; include uneditable fields, not just timestamps. No stored revision column or schema change. Revision never grants permission. |
| Updates (3/5) | Explicit presence distinguishes unchanged from clear, false, zero or empty string. Hidden sections remain unchanged; explicit empty selections clear. Preserve loaded missing references when discovery fails. |
| Availability (2/4/5) | Application supplies action facts/reasons; execution revalidates authority and state. Unknown statuses have a defined display fallback and enable no unspecified action. |
| Outcomes (4/5) | Use discriminated, validated variants for pre-effect failure, usable activation, usable activation with metadata warning, cancellation and uncertain effect where applicable. Preserve actual effect/identity evidence; never replay uncertainty. Configuration saved-with-sync-warning remains distinguishable. |
| Refresh (2/5) | Explicit read success/absence/failure and opaque change tokens; frontend owns polling cadence and stale-result envelopes. Partial registry failure retains previous valid values. |
| Source identity (5) | Validate expected source identity/state instead of trusting a stale list index; no storage migration. |
| Value ownership | Published results are detached from mutable backend state; immutable publication is documented. Clone before form mutation and command capture as needed; avoid redundant copies. |
| Resources (6/7) | Document cancellation, unsubscribe/release, stale results, EOF and idempotent close. Attachment exposes input/output/resize/completion without PTY ownership; release must not terminate the agent. Input retains driver serialization. |

Prefer this incremental approach over a complete upfront contract catalog or an
immediate frontend rewrite: both add unused abstraction or expand this slice's
behavioral risk. Accepted costs are temporary exact exceptions and explicit
mapping. Revisit if mapping duplicates policy, requires persistence infrastructure
or begins rebuilding a universal snapshot.

### 4. Implement normal-test boundary enforcement

Add `internal/architecture` with an ordinary `_test.go` entry point, checker
helpers, controlled fixtures under `testdata`, and an explicit reviewed allowlist.
Use standard-library parsing/type inspection where sufficient; do not add a new
module or invoke tests recursively from the checker.

1. Discover every production `.go` file directly from disk. Exclude `_test.go`,
   fixture trees and explicitly identified test-support packages. Do not use the
   host's active build file set as the sole inventory. Explain classification in
   diagnostics; ordinary non-test files under frontend paths are production by
   default. For `ui/uitest`, verify it is test-only with no production importers
   before granting the exact helper exclusion; otherwise inventory its violation.
2. Parse direct imports for all build variants. Enforce frontend/backend/contracts
   direction and the root-versus-child frontend rule. Traverse shared-helper
   dependencies so an apparently neutral helper cannot conceal backend coupling.
3. Inspect the complete type graph reachable from exported contract declarations:
   fields (including private representation), aliases, embeds, methods, interface
   signatures, function types, pointer/container elements and generic arguments
   and constraints. Follow local intermediate types with cycle detection; reject
   backend/framework references reached indirectly. Also inspect consumer port
   signatures and classify legacy frontend exported leakage against the baseline.
4. Resolve names/import aliases and shadowing correctly; do not rely on grep or
   exported identifier spelling. Use `go/types` for compatible build sets and an
   AST declaration resolver for inactive variants. Mutually exclusive declarations
   must not be combined into one invalid package. Unknown/unresolved type edges
   must fail with a diagnostic rather than silently passing; exercise variant
   handling with synthetic cross-platform/tag fixtures before selecting machinery.
5. Match exceptions by rule, file, import or declaration/member/type edge, reason
   and removal feature. Reject unmatched findings, duplicate/invalid entries and
   unused exceptions. Keep contracts free of exceptions. Never auto-regenerate
   the allowlist as part of the normal test; additions require explicit review.
6. Emit deterministic diagnostics showing path, symbol and dependency/type chain.
   Include every offending edge, so fixing one import cannot mask another leak.

Required checker fixtures: clean graph; forbidden direct import; intermediary
helper coupling; alias and embedded backend type; nested fields/container or
method return leakage; generics; framework leakage; reverse backend-to-UI import;
child-to-root UI dependency; inactive build-tag/platform violation; cycles without
infinite recursion; exact allowed legacy finding; unused exception; a new symbol
in an already excepted file; and production code misclassified as test support.
No fixture should require editing real production source during a test.

### 5. Record preview retirement and review the baseline

Document `previewSub` manager access (`GetScrollback`, `SubscribeOutput`,
`UnsubscribeOutput`), `SessionPreview.SetSession`/`Snapshot`/`Card`, status rendering,
`SessionPreviewRequestedMsg`, `SessionOutputMsg`, `PreviewDataMsg`/receiver, and
preview fields in refresh messages. Trace coordinator broadcast/handlers and
dashboard setters as part of the exact symbol inventory. Assign all preview-only
exceptions to feature 6. Distinguish helpers and output paths still used by
attachment; do not delete those just because the preview used them too. Add no
replacement preview interface or permanent legacy snapshot bridge.

## Validation and acceptance gate

For this planning task, validate document references and scope only. Application
tests are intentionally not run: no application implementation is authorized.

After implementation is authorized, run focused checker tests first, then:

```sh
go test ./internal/architecture/...
go test ./internal/app/... ./internal/ui/... ./internal/api/...
go test -race ./internal/app -run '^TestCommandClosuresDoNotRaceCoordinator$' -count=1
go build ./...
git diff --check
```

Confirm the selected tests actually run. Broaden race checks if implementation
changes concurrency or lifecycle code. Existing build/test failures must be
recorded with evidence; never report unrun checks as passing. Native model-driven
E2E requires the separate provider/model/token-budget authorization described by
`tests/e2e/AGENTS.md`; this plan grants none.

The assigned planning/review agent must inspect the actual diff and callers and
record PASS or CHANGES REQUIRED in `feature_1_review.md`, including:

- Inventory completeness: production files, exposed symbols, operational calls,
  test classifications, exact exceptions and removal ownership.
- All six feature-1 tasks mapped to artifacts; boundary and ownership conventions
  precise enough for a new implementer to place code and choose legal types.
- No backend aliases/embeds in contracts, no unused API expansion, and no runtime,
  storage or MCP behavior changes.
- Actual checker rejection evidence, including indirect leaks and inactive files;
  enforcement executes under normal `go test ./...`.
- Baseline regression/build/race results, accepted limitations and any unresolved
  behavior or coverage gaps. A failing criterion blocks advancement to feature 2.

No blocking product question was identified during planning. The implementation
must resolve checker type-resolution mechanics against controlled build-variant
fixtures and review the complete baseline before approval; these are engineering
tasks, not permission to weaken the stated rules.
