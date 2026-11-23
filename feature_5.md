# Feature 5 — Separate configuration management

Status: planned; not implemented or reviewed.

Ordered prerequisite: `feature_4.md` must receive final-review PASS before starting; also honor the source dependencies below.

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

Source/settings/MCP callbacks access repositories directly and identify entries by
mutable list indexes. Introduce application configuration capabilities end to end.
Prerequisites: approved features 1–4 in this ordered plan (original minimum 1–3).
Complete shared registry/change contracts before center feature 6 and retirement 7.

## Checkout evidence and affected callers/tests

`internal/app/callbacks.go:sourceManagerCallbacks/loadSources/saveSimpleSource/
deleteSource/sourceUpdateMsg` use repo methods, indexes and SyncFailedError;
`refresh.go`/`scoped_refresh.go` use fingerprints and registries;
`session_actions.go` directly deletes MCP entries. `views/mcp_manager.go` checks
built-in identity; `views/mcp_form.go`, `settings.go`, `source_manager.go` and
`components/source_item.go` expose config shapes.
Revalidate `internal/app/source_save_test.go`, `source_sync_warning_test.go`,
`settings_error_test.go`, `mcp_manager_review_test.go`, `mcp_manager_gaps_test.go`,
`internal/ui/views/mcp_form_internal_test.go`, `source_manager_test.go`,
`internal/config/repository_concurrency_test.go` and `builtin_mcp_test.go`.

## Contract and migration decisions

Use focused registry choices/rows, source editor data/changes, separate MCP editor
inputs and settings contracts. Credentials, full command maps and arguments belong
only to the necessary editable/read capability, never general list rows. Keep
comma/backslash text syntax, map/argument escaping and form parsing in UI;
authoritative configuration validation and built-in protection stay backend-owned.
Rows receive edit/delete availability and reasons, not built-in-ID policy checks.

For workspace sources use existing stable source ID plus an expected-content
revision. For path-only source lists return an opaque token bound to source kind,
full expected ordered list and selected position/value; validate it under the
repository Update*Sources read-modify-write operation before changing anything.
A reorder/edit/delete invalidates the stale token and asks for reload. This handles
duplicate values without new persisted identifiers or a schema migration. Create
is explicit append; a stale edit must never silently become append. No UI index
alone is authoritative.

Represent source mutation results as saved/not-saved plus refresh-required and
optional synchronization warning, defining valid combinations: sync failure after
save still requests refresh and reports saved state; pre-save failure must not
claim persistence. Map SyncFailedError once behind the application boundary.
Use per-inventory success/error and owned copies. Opaque comparable change tokens
hide file stamps; only successful reads advance their baseline. Keep polling and
form focus/visibility in UI, preserve settings defaults and backend source atomicity.

Move load/save/delete/sync/settings/MCP work into captured asynchronous commands,
including currently synchronous source listing. Replace messages and all callers,
then remove direct repository/manager config fields, file stamps and config errors
from these paths. Feature 6 alone owns preview leftovers; feature 7 owns remaining
attachment/deletion/composition exceptions. No registry bridge remains after 5.

## Preserved roadmap requirements

The following workstream is carried forward without dropping tasks or acceptance criteria.

Dependencies: steps 1–3; step 4 can proceed independently where contracts permit.

Tasks:

1. Move source listing, save, delete, refresh/sync, MCP configuration and settings
   operations behind focused application capabilities.
2. Introduce registry choice/row models and separate MCP editor inputs. Do not pass
   command credentials/maps into list rows merely because the stored object has them.
3. Replace built-in-ID policy checks in views with application-provided edit/delete
   availability and reasons. Backend enforcement remains mandatory.
4. Define source edit identity/concurrency semantics. Current callbacks use list
   indexes; do not let an outdated index edit/delete a different source after a
   refresh. Use stable identities or validated expected state as appropriate,
   without inventing a storage migration solely for this purpose.
5. Preserve repository read-modify-write atomicity and saved-source-but-sync-failed
   outcomes. Return structured saved/refresh/warning facts instead of requiring
   the UI to identify `config.SyncFailedError`.
6. Finish migration of registry reads/change detection and remove direct repository
   operations from callbacks and command helpers.
7. Preserve MCP argument/map parsing and escaping behavior. Text-field syntax is
   frontend concern; configuration validation is application/backend concern.

Acceptance:

- No frontend repository, file stamp or config error dependency remains for these
  operations.
- Built-in MCP protections, settings defaults and source synchronization behavior
  are preserved.
- Stale source edits cannot silently affect a different entry.
- Failed reads remain distinct from empty inventories, and partial success causes
  the right refresh and warning.

Review focus: error propagation, synchronous I/O inside frontend callbacks,
credential-bearing data in unnecessary models and concurrent source updates.

## Focused validation and final-review evidence

Run focused regressions while changing the slice, then these checks at its integration gate (adapt moved package paths after feature 7). Commands below are planned, not claimed results. Confirm test selection actually runs relevant cases; add behavioral regressions where coverage is missing.

```bash
go test ./internal/config ./internal/api ./internal/app ./internal/ui/views
go test -race ./internal/config ./internal/api ./internal/app -run 'Source|MCP|Settings|Refresh|CommandClosures'
git diff --check
```

Require evidence for stale reordered source identity, duplicate sources, saved-but-sync-failed refresh/warning, built-in MCP protection through UI/service and escaped input parsing.

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
