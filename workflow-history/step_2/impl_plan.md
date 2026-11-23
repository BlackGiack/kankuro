# Feature 2 implementation plan — focused dashboard inventories

Status: planning only. No implementation, tests of an implementation, review
approval or commit is implied. Sources: `project.md`, `feature_2.md`, `roadmap.md`,
`AGENTS.md`, the current code and `docs/frontend-boundary.md`.
`feature_1_review.md` records final-review **PASS**; revalidate that prerequisite
and the checkout before implementation. Implementation still requires instruction.

## Outcome and limits

Move the complete dashboard/card/session inventory read-to-render path onto
independent application summaries and frontend presentation rows. Both
`internal/app` and `internal/ui` are frontend code. New read consumers depend on
consumer-owned interfaces returning `internal/api/contracts` values; only backend
adapters inspect config, session and driver types. `main.go` supplies the concrete
service alongside the still-required legacy dependencies and retains lifetime
ownership. Keep the application usable after every integrated slice.

Preserve dashboard scope, shared-card ownership, deletion recovery, existing
layout/keys, search, sorting, selection, artwork and refresh cadence. No storage
migration, MCP wire changes, new module, center replacement, lifecycle redesign,
card-write redesign or package relocation is part of this step. Features 3–7
retain the exceptions identified below. Availability is advice from the current
read; every effect continues to revalidate through existing application operations.

## Revalidated starting points

| Surface | Current behavior and migration consequence |
|---|---|
| `internal/api/membership.go` | `ListDashboards` overlays deletion manifests, includes orphaned recoverable deletions, checks tombstone conflicts and sorts by name/ID. `ReadDashboardData` validates inventory, scopes shared cards and owned sessions, adds deleting-card rows and derives `CardDeleted`. Reuse these checks rather than reading repositories directly from a new adapter. |
| `internal/api/service.go` | Existing repository capability lacks puppet/fingerprint methods; workspace reads already have a focused backend lister. Add narrowly required backend read capabilities and update fixtures explicitly; unsupported capabilities must return errors rather than successful empty data. Existing MCP discovery/inspection types are broader and backend-coupled; do not reuse them as frontend contracts. |
| `internal/app/refresh.go`, `scoped_refresh.go`, `handlers.go` | Global registry reads and fingerprints are direct repository calls. Idle ticks read sessions and conditionally registries; mutations force refresh. Nil fields retain data; successful empty slices clear it. Stale scope/generation/session epochs are checked before applying results. |
| `internal/app/dashboard_list.go`, `navigation.go`, `callbacks.go` | List commands, open messages, edit and launch callbacks carry backend records. Dashboard-list requests have a separate request sequence. Card edit currently hands the complete card to the editor; preserve hidden fields when replacing the list callback with an identity. |
| `internal/ui/messages/messages.go` | Broad `DataLoadedMsg`, `SessionsRefreshedMsg`, `DashboardListLoadedMsg` and `DashboardOpenRequestedMsg` are shared with consumers outside lists. Split their responsibilities deliberately; do not mechanically replace every consumer with a list row. |
| `internal/ui/views/dashboard.go`, `dashboard_list.go` | Store backend inventories, gate keys on backend status/deletion fields, and expose selected records. Dashboard additionally uses those values to feed the legacy preview. |
| `internal/ui/components/card_list.go` | Joins capability IDs to names and puppet identity to artwork in the renderer. Pills follow selected ID order, omit unresolved IDs, and clip to available width. |
| `internal/ui/components/session_list.go` | Joins cards, searches the current card's conversation bindings, shows registration durability/error, PID and saved workspace, sorts newest first, filters name/card ID and preserves selection by ID. |
| `internal/architecture/allowlist.json`, `type_exposure_allowlist.json` | Feature 1 checks exact imports, declarations and transitive exposure. Remove retired exceptions and reconcile ownership with `docs/frontend-boundary.md`; some direct fingerprint exceptions say feature 5 while the inventory/transitive records assign this read seam to feature 2. |

## Contract and ownership decisions

Use focused reads over the existing service, with proposed names
`ReadDashboardInventory`, `ReadDashboardOverview(dashboardID)`,
`ReadDisplayChoices` and `ReadRegistryChangeToken`. Define cohesive frontend
interfaces near command adapters (for example `internal/app/read_ports.go`);
their signatures reference contracts only. Do not create a universal backend
interface or rename the application package. Backend policy helpers can retain
private config/session types. Migrate all callers of the old human read API and
remove its exported broad result; retain backend inventory helpers where deletion
and membership validation still need them. Do not change MCP discovery contracts.

Proposed independent values in `internal/api/contracts`:

| Value | Required information |
|---|---|
| Dashboard summary | ID, name, description, semantic active/deleting/unknown state, recoverable-deletion facts, availability for open/edit/delete/inspect/retry. No deletion manifest. |
| Card summary | ID, owner dashboard ID, shared flag, name, driver/model labels, skill/persona IDs, puppet and face identities, semantic deletion state and availability for launch/edit/delete/retry. Include only further fields proven necessary by migrated callers. |
| Session summary | ID, dashboard/card IDs, name, semantic status, start time, PID, saved actual/original workspace identity and temporary-workspace flag, deleted/unavailable-card facts, registration state/durability/error, optional resolved active-conversation ID/name, face identity where needed, action availability. No binding map, instructions, credentials, process handle or output buffer. |
| Display choices | Focused skill/persona ID/name choices and puppet identity plus owned artwork data. Resolve names/artwork in UI mapping; do not return config registries or MCP credentials just for display. |
| Availability | Explicit enabled flag and semantic reason per relevant action. Zero/unknown values disable actions. UI formats reasons into existing feedback. |
| Registry change token | Opaque comparable string value; zero means no valid baseline. Its backend representation encodes the existing fingerprint deterministically, without exposing filenames, stamps or filesystem types in signatures. It is an invalidation hint, not a revision or authorization token. |

Application availability covers dashboard open/edit/delete/retry/inspect; card
launch/edit/delete/retry; session attach/relaunch/stop/remove/switch/rename and
workspace commands where those controls are displayed. Derive from the checked
inventory and existing operation preconditions, preserving owner-only mutation,
shared-card inspection where currently supported, deleted-card restrictions and
recovery actions. Do not run filesystem probes or acquire effect reservations to
produce availability. Factor shared pure policy predicates where useful; do not
make the frontend reconstruct them. Read facts cannot replace admission checks,
conversation compatibility checks or revision-bound deletion confirmation.

Add `DashboardRow`, `CardRow`, `SessionRow` in `internal/ui/models` and pure mapping
in the frontend. Rows contain complete render inputs: resolved pill names, card
label/deleted-card fallback, selected appearance/artwork, registration and active
conversation details, workspace identity, semantic style/action facts and search/
sort keys. Renderers still own terminal formatting, escaping, clipping, truncation
and size-dependent layout. Contracts contain no ANSI strings or framework types.
Check the dependency graph of artwork helpers before sharing them; replace
config-typed artwork lookup inputs with focused presentation data when required.

Copy nested slices/artwork buffers at publication boundaries; no row or contract
aliases repository or manager collections. The backend resolves the active binding
only within the session's current card and publishes an optional value, never a
map. Frontend caches own their values; published values are immutable after handoff.
If a setter/getter permits mutation, copy its nested data rather than relying on
an undocumented convention. Rows must never be converted back into storage cards.

## Refresh semantics and validity

Use explicit read outcomes: each attempted read has success or error, and partial
frontend envelopes separately indicate whether a read was attempted. This can use
a small `ReadResult[T]` plus an optional envelope field; avoid a zero value that
accidentally means successful empty. On success, even a nil/empty collection clears
that collection. On error, retain the previous valid collection and surface the
error. A sessions-only update leaves unattempted registries untouched. Skills,
personas, puppets and legacy registries retain independent outcomes, so one failure
does not discard other successes. Scoped inventory validation remains all-or-error;
never treat invalid membership as a partially scoped successful result.

Maintain frontend-owned caches of successful summaries and choices. Rebuild rows
from these caches whenever either input changes, including capability renames or
artwork changes without card changes. A failed choice read retains its prior names/
artwork while showing the error. Confirmed empty choices remove resolved labels
using current fallback behavior. Avoid transient row publication between applying
the successful inputs of one refresh. Failed scoped reads retain existing rows;
successful empty sessions clear list and preview selection as they do today.

Keep dashboard ID, view generation and session epoch on every scoped asynchronous
result, and request sequencing for dashboard-list reads. Capture capabilities,
identities, baseline and sequence values before constructing commands. Reject
obsolete results before updating caches, rows, errors, preview or token baseline.
Preserve the rename/action epoch protection against older full and sessions-only
snapshots. Guard overlapping config refreshes with a frontend request sequence so
an older changed result cannot overwrite a newer baseline. Unchanged results have
no baseline update and cannot clobber a forced mutation refresh.

The application computes the opaque token from the existing sorted filename/stamp
set. Read/check fingerprints around the corresponding registry load; if either
check fails or the token changes during loading, return useful successful reads
but do not certify a new baseline. Surface fingerprint errors. Any failed required
registry read invalidates the baseline so the next scheduled poll retries even
when file metadata is unchanged. Never accept a token sampled after a read as
proof that the preceding data contains a concurrent write. Keep the existing
metadata-based fingerprint limitations; do not introduce content hashing or a new
persistent revision store. Polling remains on the existing two-second UI tick,
with immediate forced refresh after mutations and no error-driven busy loop.

## Ordered implementation slices

1. **Lock down behavior and callers.** Enumerate every use of the four migrated
   messages, list accessors and `RenderCardItem`, including launch/switch/editor,
   rename/focus and preview consumers. Record fields and availability expectations.
   Extend meaningful behavioral fixtures for gaps before removing those paths.
2. **Add focused backend reads.** Define neutral contracts, map checked inventories,
   resolve active conversations, derive availability, and deep-copy results. Add
   display-choice and change-token adapters over the repository. Test policy,
   scope, deleting/tombstoned identities, unknown statuses and ownership here.
   Keep existing consumers operational until their replacement is connected.
3. **Migrate dashboard inventory end to end.** Inject its read capability, convert
   list load/open messages, setters, form seeds and callbacks to focused values or
   identities, preserve request ordering and name/ID ordering, and remove the old
   frontend `config.Dashboard` read path. Dashboard create/update/delete effects
   continue through the existing checked operations.
4. **Migrate scoped refresh and mapping.** Inject scoped/display reads; add explicit
   outcomes, owned caches, pure mapping and scoped envelopes. Move fingerprint
   implementation behind the service and adopt opaque tokens. Update handlers and
   cadence/race fixtures together. Split legacy consumer delivery as described
   below; do not widen rows to preserve a consumer outside this feature.
5. **Migrate card/session presentation end to end.** Change list item types,
   setters/getters, delegates, filters, selected values and dashboard state to rows.
   Replace status-based action policy in dashboard handlers with availability;
   preserve confirmations and local interaction logic. Convert launch/edit/switch/
   selection callbacks to identities. Move capability/card joins and active-binding
   searches out of renderers, then delete their old helpers and signatures.
6. **Remove obsolete paths and verify integration.** Delete unused broad inventory
   messages/fields and reads after migrating callers; inspect backend callers before
   removing internal helpers. Update precise architecture exceptions and inventory
   documentation without adding a broad allowance. Run integration checks and stop
   for final review. Do not start feature 3 or 6 on the strength of this plan.

Each slice must compile with its callers. Additive contracts can precede their use,
but temporary unused/dual read APIs must be removed before this feature's gate.

## Temporary exceptions and handoffs

| Owner | Exact retained responsibility and boundary |
|---|---|
| Feature 3 | Card editor full-record load/save, expected-record concurrency and hidden-field preservation. List edit emits card ID; an async legacy edit loader rechecks dashboard context and loads the complete card for `CardEditor.SetCard`. Keep that record solely in the editor adapter, never reconstruct it from a summary or cache it in list rows. Shared-card read-only inspection must survive. Remove this loader when feature 3 supplies its focused edit contract. |
| Feature 4 | Launcher, agent-switch, conversation picker and lifecycle/workspace effects may retain their current detail contracts/error handling. Migrate any shared list component use to rows now; fetch legacy operation detail by captured IDs only inside named action adapters. Preserve explicit conversation choice and post-effect persistence warnings; no broad inventory payload solely for action convenience. |
| Feature 5 | Source editing, settings, MCP configuration and remaining editor/config registry payloads retain existing APIs. Display choices used by lists must already use application reads in feature 2. Legacy global read delivery may remain in a separate config/editor message, isolated from dashboard rows, until feature 5 completes it. The frontend fingerprint seam moves now, including its direct exceptions previously assigned to feature 5. |
| Feature 6 | `internal/app/preview.go`, `SessionPreview`, preview-only selection/detail payloads, scrollback and subscriptions stay functional. Separate legacy preview metadata from inventory rows and deliver it only to the preview adapter, with captured scope/selection guards. Do not keep a dashboard-wide snapshot/card map for preview. Use a scoped selected-session detail adapter if needed, only with fields the existing preview actually consumes; remove it with the preview. Keep shared attachment output machinery. |
| Feature 7 | Attachment handles, deletion preflight/manifest formatting and remaining concrete composition dependencies retain their existing owners. Do not expose manifests in new dashboard rows or change confirmation authority. `internal/app` remains until this step completes. |

Enumerate actual surviving declarations/imports in `docs/frontend-boundary.md` and
the two exact allowlists. Line/column records may need mechanical adjustment after
edits; review every changed exposure rather than accepting a wholesale baseline
regeneration. A surviving feature-2 inventory exposure is unfinished work, not an
exception to reassign to a later feature. New bridges must have the narrow scope
and removal owner above and cannot establish a second permanent read API.

## Acceptance evidence and planned checks

| Feature task / acceptance | Required evidence |
|---|---|
| 1: checked scoped reads and recovery | API regressions in `membership_internal_test.go` plus focused read tests: foreign sessions excluded, shared cards retain ownership, deleting/orphan-manifest dashboards visible, tombstone conflicts fail, deleted-card sessions remain inspectable. |
| 2: focused rows and complete display | Mapping and component tests for names, driver/model, PID, registration pending/unsaved/error, current-card active conversation, actual/temp workspace, deleted card fallback, pill ordering/missing IDs, custom/default artwork and waiting style. No map-based renderer joins. |
| 3: application action facts | Backend matrix for active/deleting/unknown dashboards/cards, running/waiting/terminated/unknown sessions, shared ownership, deleted/unavailable card and recovery actions. Change state after a successful read; existing execution checks must still reject stale or unauthorized intent. |
| 4: messages/callers/selection | Drive coordinator `Update`; verify identity callbacks, read-only shared edit flow, row replacement/reordering and filtered selection preservation, rename pending focus and fallback when selected identity disappears. No row-to-record round trip. |
| 5: stale result controls | Extend `multi_dashboard_internal_test.go`, `inline_rename_refresh_internal_test.go` and command race coverage: dashboard switch, same-dashboard view change, list request reordering, epoch changes and overlapping forced/config reads cannot apply obsolete data or baseline. |
| 6: error versus empty | Extend `scoped_refresh_error_internal_test.go`, `refresh_error_test.go`: each registry fails independently, prior values survive, errors surface, successful empty clears, unattempted fields retain, recovery succeeds without a metadata change. |
| 7: application registry reads/token | Extend `refresh_cadence_internal_test.go`: unchanged poll skips reload, changes reload, mutations refresh immediately, no source reload on idle tick, full refresh seeds baseline, fingerprint failure/partial failure cannot advance it, write during load retries, unchanged/older responses cannot overwrite newer baseline. |
| Copy isolation and layout | Mutate source and published nested slices/artwork in ownership tests; later reads and unrelated caches stay unchanged. Adapt `card_list_test.go`, `session_list_test.go`, `card_render_layout_test.go`, `session_list_layout_test.go` and relevant dashboard layout tests. Assert positions as well as bounded size, Unicode/control text and oversized art. |
| Boundary and removal | Architecture checks plus caller/import/signature searches show independent contract graphs and list models, no feature-2 legacy read exposure, no UI/framework import in backend, and documented exact exceptions only. |

Planned implementation gate commands (not run by this planning task):

```bash
go test ./internal/api ./internal/app ./internal/ui/components ./internal/ui/views ./internal/ui/models
go test ./internal/architecture/... -count=1
go test -race ./internal/app -run 'TestCommandClosuresDoNotRaceCoordinator|Dashboard|Refresh'
go test -race ./internal/api
go build ./...
go vet ./...
go test ./...
git diff --check
```

Confirm the race regex actually includes new stale-result tests; run named tests
separately where necessary. Use ordinary isolated fixtures and fakes for read
capabilities, retaining real backend tests for scope/admission. Manual terminal
verification uses isolated config and temporary workspace roots: verify dashboard
recovery, both list selections, refresh after a mutation, artwork and preview
continuity. Native-agent experiments additionally isolate HOME; model-driven E2E
requires separate provider/model and total-budget authorization per project rules.

The final reviewer must inspect actual contracts, all callers/removals, backend
policy and ownership tests, remaining exception declarations and actual command
results. Record changed files, acceptance evidence, failures/unrun checks and
limitations in `feature_2_review.md` with **PASS** or **CHANGES REQUIRED**. Failed or
unverified required criteria block advancement. No questions currently block
planning; revalidate implementation details against the checkout when authorized.

## Trade-offs and planning validation

Keeping broad objects or merely moving their structs would preserve the coupling.
Focused summaries plus UI mapping are the smallest option meeting the agreed
boundary; their costs are explicit mapping, owned caches and temporary legacy
adapters. Avoid universal DTOs or a new event/store system. Revisit a contract if it
needs whole binding maps, process configuration or repeated frontend policy checks.

Planning validation: source paths and current call sites were inspected; feature 1's
recorded PASS was read, not rerun. Only this plan is authored by this task. Application
checks above remain planned; documentation validation does not establish runtime
acceptance or authorize implementation.
