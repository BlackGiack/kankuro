# Feature 5 implementation plan — separate configuration management

Status: planning only, 2026-11-16. No implementation, tests, commit or feature
approval is claimed by this document. Implementation requires a separate
instruction. Stop after implementation for the assigned planning/review agent's
independent final review in `feature_5_review.md`; do not advance to feature 6
without PASS.

## Scope and prerequisites

Implement all seven preserved tasks and four acceptance criteria in
`feature_5.md`, subject to `roadmap.md` and `project.md`. Features 1–4 are ordered
prerequisites. `feature_4_review.md` explicitly records PASS and readiness for
feature 5; recheck prerequisite records and the actual checkout before coding.
Existing untracked feature/review/workflow files are unrelated work to preserve.
Never commit without an explicit request.

The result is a complete configuration vertical slice: source listing/editing/
deletion/synchronization, MCP management, settings, registry reads and change
detection cross independent application contracts. Both `internal/app` and
`internal/ui` are frontend. Application policy and authoritative validation stay
in `internal/api` and configuration storage; UI owns text parsing, navigation,
focus, visibility, formatting and local feedback. Main owns composition/lifetime.
No storage migration, module, service, MCP wire change or visual redesign.
Preview remains feature 6; attachment, deletion/composition exceptions and
retirement of `internal/app` remain feature 7. No registry bridge survives step 5.

## Revalidated checkout evidence

| Surface | Current behavior and planned destination |
|---|---|
| `internal/app/callbacks.go` | `loadSources` synchronously lists sources and discards errors. Save/delete use indexes and stale edits can append. Workspace commands and refresh closures read `c.repo`. Replace with captured asynchronous configuration commands. Remove `sourceUpdateMsg` and frontend `SyncFailedError` interpretation. |
| `internal/ui/views/source_manager.go`, `internal/ui/components/source_list.go`, `source_item.go` | Reload occurs through Init/SetData/callbacks; workspace items expose `config.WorkspaceSource`. Replace with asynchronous intents, application identities and frontend rows/drafts. |
| `internal/app/refresh.go`, `scoped_refresh.go` | Global config reads use repository models alongside already focused display/editor choices. Existing before/after `RegistryChangeToken` certification protects the baseline. Finish migration without weakening partial-read or stale-result handling. |
| `internal/api/inventory_reads.go`, `contracts/inventory.go`, `contracts/card_editor.go` | `ReadResult[T]`, `Availability`, focused choices and opaque comparable change token already exist. Extend/reuse rather than introduce competing registry APIs. File stamps already belong behind `ReadRegistryChangeToken`. |
| `internal/app/coordinator.go`, `handlers.go`, `navigation.go` | Settings startup read/default fallback and settings/MCP result routing depend on backend models. Move I/O to commands and preserve request/dashboard/generation checks. |
| `internal/app/legacy_actions.go` | Both `deleteMCP` and `deleteMCPFromManager` call the repository. Feature text points at an older location in `session_actions.go`; migrate the actual paths and their callers. |
| `internal/ui/views/mcp_manager.go`, `mcp_form.go`, `settings.go`, `internal/ui/messages/messages.go` | MCP list rows carry entire server objects and check built-in identity; editor/settings/messages expose config types. Introduce narrow rows and separate editor/settings contracts. |
| `internal/config/{skill_sources,persona_sources,puppet_sources,workspaces}.go` | `Update*Sources` owns read/modify/write and returns `SyncFailedError` after saved-list/failed-derived-sync. Validate source identities inside these callbacks. |
| `internal/config/mcps.go`, `repository.go`, `types.go` | MCP storage enforces built-in shadow/edit/delete and referenced-delete protection; settings read overlays persisted fields onto defaults. Reuse these owners and behavior. |
| `docs/frontend-boundary.md`, `internal/architecture/*allowlist.json` | Exact import/type exceptions name owning features. Remove feature 5 edges after migration; retain only justified later-feature edges. |

## Contracts and decisions

Use independent types in `internal/api/contracts`, backend mapping/operations in
focused `internal/api` files, consumer-owned cohesive interfaces beside current
frontend coordination (for example `configuration_ports.go`), and rows/drafts in
`internal/ui/models`. Service satisfies interfaces implicitly and never imports
the frontend. Illustrative names below may follow existing naming conventions;
the ownership and semantics are required.

- Registry capabilities return per-inventory `ReadResult` values for required
  skills/personas/workspaces/puppets/MCP choices or rows. Reuse existing editor and
  display choices where sufficient. Inspect every `DataLoadedMsg` consumer before
  deleting fields; provide only the fields each surviving consumer needs. General
  MCP rows contain identity/name and action availability/reasons, with display
  facts needed for the current built-in label, not command/args/env/headers or
  credentials. Puppet artwork and workspace choice details remain available to
  their actual consumers through detached focused data.
- Source capability groups list/editor, explicit create, edit, delete and sync.
  `SourceKind` is an application semantic value, not a view enum. Source rows
  carry editable path/root facts and opaque mutation identity. UI rows carry
  display text and local selection; an index may locate a UI row but is never
  authoritative in a mutation request.
- Path-only sources: produce a versioned opaque token covering kind, the full
  ordered expected list (including duplicates), selected position and value. A
  canonical content digest bound to position/value is sufficient; avoid ambiguous
  concatenation. Validate against freshly loaded data inside `Update*Sources`.
  Reject malformed, cross-kind, missing or stale tokens with reload guidance,
  before returning modified data. Reordering/editing/deleting invalidates the
  token; do not find the first equal path or silently append. Identical duplicate
  entries remain addressable by position within that exact expected list. This
  is expected-content comparison, not a persisted generation/ABA guarantee.
- Workspace sources: use existing stable source ID plus a revision of the full
  expected source content. Resolve by ID and compare revision inside
  `UpdateWorkspaceSources`; a missing ID is a conflict. Reordering alone does not
  retarget an edit. Create explicitly appends with repository-generated identity;
  edit preserves identity. Keep existing assignment and atomic write behavior.
- Source mutations return structured saved state, refresh requirement, optional
  sync warning and failure diagnostics. Valid normal combinations: saved +
  refresh + no warning; saved + refresh + sync warning; not saved + no refresh +
  failure. A stale conflict may additionally request a source reload while still
  reporting not saved; distinguish that from a persisted-change refresh. Map
  `SyncFailedError` exactly once in the backend. Explicit sync has its own outcome
  (no source-save claim), and refreshes potentially affected inventories even
  when sync reports failure, preserving current behavior.
- MCP capability groups list, new/editor load, save and delete. Editor data alone
  includes command, args, working directory, URL, env and headers needed for the
  form. Requests carry identity plus explicit editable values; preserve metadata
  currently retained by edit. New form data is blank and never inherits a previous
  editor. Built-in/read-only decisions and reasons are backend facts. Execution
  rechecks built-in identity/name collision and referenced deletion using existing
  lock-owning repository operations; a stale availability result grants nothing.
  Do not introduce an unrelated MCP concurrency or authorization redesign.
- Settings capability returns all four current fields and application defaults.
  Preserve absent-field defaults and explicit false/empty values. Startup may
  obtain defaults through an application-owned defaults contract/bootstrap value
  without I/O, then asynchronously read persisted settings; report read errors
  while retaining defaults or the last valid value. Do not replace an edited
  draft with a late load. Save updates applied frontend settings only on success.
- Reuse `RegistryChangeToken` as an invalidation hint, never edit revision or
  authorization. Keep before/after certification around relevant reads; only
  successful accepted reads advance the baseline. A failed inventory retains its
  prior valid value; successful empty clears it. A partial read may update its
  successful inventories but must not certify the failed set as current.
- Published slices/maps/artwork buffers are owned detached copies. Backend results
  do not alias repository state; form drafts clone mutable result values before
  editing; command inputs snapshot mutable drafts before scheduling. Document
  this at contract boundaries and test both directions of mutation isolation.

Keeping callbacks with neutral wrappers would leave I/O, policy and stale-index
risks in the frontend. Persisted IDs for every path source would solve identity
but violate the no-migration constraint. Focused service operations plus validated
expected content provide the required boundary using existing repository locks.
The accepted cost is conservative reload conflicts for path-list changes. Revisit
only if stable identity across such changes becomes a product requirement.

## Implementation sequence

Each slice must compile and keep the application usable; migrate all callers of
a changed signature together. Do not leave a permanent compatibility API.

1. **Contracts and backend configuration operations.** Add focused contracts and
   service methods, reuse choice/token types, and define outcome validation and
   ownership. Implement source token/revision validation under `Update*Sources`,
   centralize sync warning mapping, and adapt MCP/settings operations. Keep
   repository lock ownership intact; do not call locking methods from a callback
   already running under that lock. Add backend behavioral tests before wiring UI.
2. **Source manager end to end.** Replace synchronous LoadSources/Reload with
   load intents returning commands and explicit results. Init, navigation,
   SetData/reload and post-mutation refresh schedule commands rather than read
   files. Preserve path/root forms, list order and selection behavior. Capture
   edit identity when editing starts, not from the latest cursor at submit time.
   Failed loads retain data/draft and show errors; empty success clears rows.
   Stale edit/delete requests keep user input, report conflict and offer reload.
   Saved-but-sync-failed reloads saved sources/registries and shows the warning.
3. **MCP and settings end to end.** Replace storage objects in rows, forms,
   callbacks/messages and coordinator settings. Load MCP editor data asynchronously
   by identity; the manager does not transport full server configuration. Migrate
   both legacy deletion adapters and keep confirmation admission/visit correlation,
   retry behavior and pending-state cleanup. Keep comma/backslash args/map parsing,
   escaping, field order and local validation in the form. Move settings startup,
   load/save and puppet sync to captured application commands; expose sync errors
   currently discarded by `syncPuppets`.
4. **Registry reads and cleanup.** Replace `refresher.repo` and config payloads in
   `DataLoadedMsg` with focused results, keeping existing dashboard reads, editor
   choice retention and refresh cadence. Remove redundant raw registries after
   checking all consumers, including preview. Adapt preview inputs minimally if
   needed without a new preview architecture. Remove direct config fields from
   migrated helpers/coordinator, legacy repository deletion and frontend config
   errors/file stamps. Constructor signature changes are limited to what removes
   configuration access; retained composition-only parameters must be explicitly
   feature-7-owned and must not store a usable frontend repository bridge.
5. **Enforcement and review handoff.** Remove obsolete feature-5 import/type
   exceptions and update `docs/frontend-boundary.md` to actual remaining edges.
   Inspect all callers and removals, including platform/tagged code. Do not widen
   allowlists to make new leakage pass. Record actual checks and limitations in
   feature evidence, then stop for independent final review.

Every async envelope retains applicable dashboard ID, view generation, inventory
request ID and session epoch, plus source kind/visit and operation request identity
for source screens and existing MCP manager/form identity. Global persistence may
finish after navigation; obsolete results must not navigate, overwrite drafts,
clear newer pending state or apply feedback to another screen. Arrange a fresh
global read when needed without applying stale screen results. Construct commands
from captured values/ports only, never `c` or mutable view state. Cancellation
before mutation produces no save; cancellation after an effect must preserve
known outcome evidence. Do not replay uncertain saves/sync/input effects or add
an automatic effect retry mechanism.

## Acceptance and evidence map

| Requirement | Planned implementation and behavioral verification |
|---|---|
| Tasks 1/6; no repository, stamp or config-error dependency | Configuration ports/service methods, asynchronous source/settings/MCP paths, focused refresh payloads; architecture tests and production-call/import search over both frontend trees. Inspect signatures, aliases, maps and nested types as well as imports. |
| Tasks 2/3; narrow rows and policy ownership | MCP rows lack sensitive editor fields; load only through editor capability; availability/reasons drive UI. Service tests attempt built-in edit/delete/shadow directly; UI tests exercise unavailable actions; referenced delete still fails after a stale allowed read. |
| Task 4; stale edits cannot target another entry | Service tests for reorder, edit, deletion, duplicates, wrong-kind token, invalid position and explicit create; workspace ID/revision conflicts and reorder safety. Concurrent source updates preserve all successful appends and reject stale edits atomically. |
| Task 5; source partial effects | Fault-inject read/write/sync errors for each source family: malformed input never saves, failed write never claims saved, post-save sync error returns saved/refresh/warning and UI reloads while retaining warning. Dedicated sync failure is visible. |
| Task 7; parsing preserved | Form round trips for commas, backslashes, equals in map values, args, empty maps/lists, env/headers and working directory. Keep existing parse/serialize rules and authoritative backend validation. |
| Settings defaults and failed reads | Missing fields use defaults, explicit false/empty survives; read failure retains defaults/prior values and reports error; failed save stays visible and does not apply settings; late load cannot erase current edits. |
| Inventory partial/empty and ownership | Per-inventory success/error, confirmed empty, failed baseline certification and later retry, owned-copy tests; card selections survive missing choices; stale navigation/load/save results cannot replace newer state. |
| Commands and usability | Delayed fake capabilities prove callbacks/Update do not block; `TestCommandClosuresDoNotRaceCoordinator` covers source/settings/MCP/sync command execution; coordinator tests drive Update and verify view write-back/navigation. |

Revalidate and adapt the existing tests in `internal/app/source_save_test.go`,
`source_sync_warning_test.go`, `settings_error_test.go`,
`mcp_manager_review_test.go`, `mcp_manager_gaps_test.go`,
`internal/ui/views/mcp_form_internal_test.go`, `source_manager_test.go`,
`internal/ui/components/source_list_test.go`, and config concurrency/built-in/
settings tests. Inspect existing MCP form tests for escape coverage and add gaps
rather than replacing assertions with implementation-shaped tests. Add focused
service tests for newly authoritative identity/outcome rules and coordinator
tests for async routing. Use existing isolated fixtures and cleanup ownership.

## Planned validation and final gate

Documentation-only planning needs content/reference checks, not application test
runs. The following commands are planned for implementation and are not results:

```bash
go test ./internal/config ./internal/api ./internal/app ./internal/ui/views ./internal/ui/components ./internal/api/contracts ./internal/architecture -count=1
go test -race ./internal/config ./internal/api ./internal/app -run 'Source|MCP|Settings|Refresh|CommandClosures' -count=1
go build ./...
go vet ./...
git diff --check
```

Confirm test selection actually executes each required case, including any
concurrency case whose name falls outside the race filter. Run applicable lint
under repository conventions and state its scope and existing exceptions. Do
manual source/MCP/settings terminal checks using isolated config/temp roots at
80×24 and a larger terminal. Native-agent runs also isolate user configuration;
model-driven E2E requires separate provider/model and total-budget authorization
under `tests/e2e/AGENTS.md`; this plan grants none.

Reviewer must inspect actual code, all migrated callers/removals, service policy
tests, architecture exceptions and actual command output. Record PASS or CHANGES
REQUIRED, changed files, requirement coverage, checks/results, remaining owned
exceptions and limitations in `feature_5_review.md`. Failed or unrun required
criteria prevent approval; the plan and old review evidence do not prove the new
checkout passes. No planning questions or blockers are identified. Implementation
remains unauthorized by this planning request.
