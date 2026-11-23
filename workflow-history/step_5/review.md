# Step 5 review

Decision: **CHANGES REQUIRED** — 2026-11-16. Do not advance to step 6 yet.

Reviewed the working-tree implementation against `feature_5.md`, `impl_plan.md`
and the roadmap boundary/behavior rules. `review.md` was created because the two
issues below were reproduced. No implementation fixes or commits were made.

## Required changes

### 1. [P2] Correlate settings save completion to the originating view visit/draft

Location: `internal/app/coordinator.go:496` (especially the navigation at line
518), with capture in `internal/app/callbacks.go:290`.

Settings save results check dashboard identity and dashboard generation, but
neither changes when navigating between screens within that dashboard. The
request counter advances on save, not on navigation or starting a new draft.
Consequently, save settings, open Puppet Sources while the save is pending, and
then deliver its successful result: the coordinator immediately navigates away
from Puppet Sources to Dashboard. A delayed error likewise appears on the new
screen. Leaving and reopening Settings before completion can also let the old
result close a newer edit.

This violates the feature's stale-result rule and the plan's explicit requirement
that obsolete configuration completions must not navigate or overwrite a newer
screen/draft. Global applied settings can still reflect a completed successful
save; that does not authorize navigation or local feedback in a different visit.

Required: track the originating settings visit/draft identity separately from
global save completion. Apply global persisted facts and release the matching
pending operation, but gate navigation/local feedback on the still-current visit
and draft. Add deterministic coordinator tests for success/error after leaving,
leaving/reopening, and starting a newer draft while completion is held.

Verification: a temporary `TestReviewSettingsSaveDoesNotNavigateFromNewView`
used the existing `newLaunchFocusCoordinator` fixture, entered Settings through
the coordinator, submitted with `Update(Enter)`, held the returned command,
navigated to Puppet Sources, then delivered the saved result through `Update`.
It failed with `old settings save navigated to 0` (Dashboard). The temporary test
was removed after verification.

### 2. [P2] Do not silently accept edits and discard submissions during an MCP save

Location: `internal/app/callbacks.go:317` and
`internal/app/coordinator.go:462`; related interaction in
`internal/ui/views/mcp_form.go:140`.

The new `mcpSaving` guard silently returns nil for another submission, while the
form remains fully editable (its `loading` guard covers editor loading only).
Submit a valid MCP, change the name while its save is pending, and press Enter
again: no second save is scheduled and there is no pending-state feedback. When
the first save succeeds, its unchanged request identity still matches, so the
coordinator navigates away and discards the newer draft. Only the first values
reach persistence. The settings save guard has the same silent-repeat behavior
and should be covered when correcting pending form interaction.

Required: make pending-save interaction explicit. Either prevent editing and
submitting while that operation is pending, with appropriate feedback and recovery
on failure, or retain/version newer drafts so an earlier completion cannot close
them. Do not queue/replay uncertain effects automatically. Preserve existing
navigation correlation and failed-save retry behavior. Add a held-completion test
that exercises actual form/coordinator interaction, rather than only callback
invocation or synthetic saved messages.

Verification: temporary `TestReviewMCPSaveDoesNotDiscardNewDraft` initialized a
valid form, submitted through coordinator `Update(Enter)`, held the command,
changed the name input and submitted again. The second command was nil; delivering
the first result failed the assertion that the newer form remain open:
`first save closed newer unsaved draft; view=0`. The temporary test was removed.

## Acceptance evidence and remaining scope

- Tasks 1/6: configuration operations now cross independent contracts; direct
  frontend repository fields/calls and configuration model imports were removed.
  Inspected production callers, both legacy MCP deletion paths, message payloads,
  constructor/main wiring, refresh/token certification and architecture removals.
  Remaining concrete composition/deletion/attachment and preview dependencies
  retain their feature 6/7 owners; test-support cleanup remains feature 8.
- Tasks 2/3: MCP list rows omit connection maps/arguments; editor loads are separate
  and clone mutable values. Built-in availability/reasons come from the service;
  repository execution still protects built-in edit/delete/shadow and referenced
  deletion. Form metadata and blank-new behavior remain covered.
- Tasks 4/5: path tokens bind kind/full ordered content/position; workspace tokens
  bind stable identity and content. Validation runs within repository
  `Update*Sources` callbacks. Inspected duplicate/reorder/cross-kind/workspace
  revision/concurrent-create and fault tests. Saved-plus-sync-warning outcomes
  retain refresh, and pre-save failure does not claim persistence. Source UI
  commands capture identity, retain stale drafts and distinguish failed/empty
  reads. No storage migration or MCP wire change was introduced.
- Task 7/settings: MCP parsing/escaping stays frontend-owned and existing round
  trips pass. Backend settings defaults and explicit false/empty values survive.
  Late settings reads retain an active edit and failed reads retain applied data.
  Pending save/navigation behavior still needs the fixes above.
- The remaining acceptance gap is asynchronous form/result safety, not a failing
  architecture check. Passing existing tests does not cover the reproduced paths.

## Independent verification

All of these passed on the reviewed implementation (temporary probes removed
for the final full run):

```text
go test ./internal/config ./internal/api ./internal/app ./internal/ui/views ./internal/ui/components ./internal/api/contracts ./internal/architecture -count=1
go test -race ./internal/config ./internal/api ./internal/app -run 'Source|MCP|Settings|Refresh|CommandClosures' -count=1
go test ./... -count=1
go build ./...
go vet ./...
golangci-lint run --new-from-rev=HEAD ./...
git diff --check
```

Diff-filtered lint reported zero issues. Full unrestricted lint was not rerun;
the implementation evidence reports existing architecture-checker debt, which
this review does not claim resolved. An earlier full-suite invocation included
the temporary regression probes and failed those two assertions as expected;
the subsequent full-suite invocation after their removal passed.

Intermittent-test investigation:

```text
go test ./internal/app -run '^(TestPreviewAutoSelection|TestSessionLauncherBehavior)$' -count=50
go test ./internal/api -run '^TestReplyWaitEndsOnRecipientStopAndCallerRevocation$' -count=30
```

Both passed (50 repetitions of each app test; 30 of the reply-wait test and its
subtests). The focused feature checks and final full suite also passed. The
reported intermittency was not reproduced; these runs neither prove its absence
nor establish a root cause. Preview and launcher tests use short-lived `echo`
processes, and the reply-wait test exercises shutdown timing; no new messaging or
session production implementation is part of this diff. Do not attribute those
earlier failures to feature 5 without further evidence.

No independent manual terminal or native-model E2E run was performed in this
review. The author's isolated terminal evidence was read but is not represented
as an independently repeated check. No external/environment blocker was found;
the two required changes block approval. Re-review their fixes and regressions
before recording PASS. No commit is authorized by this review.

## Reviewed changed files

The inventory below includes implementation/test changes and the boundary
documentation/enforcement updates; task plans, workflow history and activation
configuration are excluded.

```text
docs/frontend-boundary.md
internal/api/configuration.go
internal/api/configuration_sources.go
internal/api/configuration_sources_test.go
internal/api/configuration_test.go
internal/api/contracts/configuration.go
internal/api/contracts/configuration_sources.go
internal/app/app.go
internal/app/callbacks.go
internal/app/command_race_internal_test.go
internal/app/configuration_ports.go
internal/app/conversation_choice_internal_test.go
internal/app/coordinator.go
internal/app/coordinator_test.go
internal/app/current_fixture_external_test.go
internal/app/handlers.go
internal/app/inline_rename_refresh_internal_test.go
internal/app/integration_test.go
internal/app/launch_focus_internal_test.go
internal/app/legacy_actions.go
internal/app/mcp_manager_gaps_test.go
internal/app/mcp_manager_review_test.go
internal/app/multi_dashboard_internal_test.go
internal/app/navigation.go
internal/app/preview_broadcast_test.go
internal/app/preview_reconciliation_internal_test.go
internal/app/refresh.go
internal/app/refresh_cadence_internal_test.go
internal/app/refresh_error_test.go
internal/app/scoped_refresh_error_internal_test.go
internal/app/settings_async_internal_test.go
internal/app/source_save_test.go
internal/app/source_sync_warning_test.go
internal/app/terminal_launch_test.go
internal/architecture/allowlist.json
internal/architecture/type_exposure_allowlist.json
internal/ui/components/source_item.go
internal/ui/components/source_list.go
internal/ui/components/source_list_integration_test.go
internal/ui/components/source_list_test.go
internal/ui/messages/messages.go
internal/ui/views/card_editor_layout_test.go
internal/ui/views/card_editor_test.go
internal/ui/views/editor_helpers_external_test.go
internal/ui/views/mcp_form.go
internal/ui/views/mcp_form_test.go
internal/ui/views/mcp_manager.go
internal/ui/views/mcp_manager_test.go
internal/ui/views/modals_layout_test.go
internal/ui/views/settings.go
internal/ui/views/settings_test.go
internal/ui/views/source_manager.go
internal/ui/views/source_manager_test.go
internal/ui/views/workspace_picker_test.go
main.go
```
