# Independent review — feature 3

Decision: **CHANGES REQUIRED. Not ready to advance to feature 4.**

Reviewed the step 3 working-tree changes against `feature_3.md`, inspecting the
application save/revision logic, repository persistence, editor draft and async
flows, pickers and migrated tests. Did not read `impl_plan.md` or other
implementation plans/review summaries. Created this file because changes are
needed. No implementation files changed and no commit was made.

## Required changes

1. **[P2] Preserve the currently highlighted workspace when options refresh.**
   `internal/ui/views/workspace_picker_modal.go:106` refreshes the list and then
   calls `restoreSingleSelection`, which restores `m.selectedID`. That field is
   updated by explicit Up/Down navigation, but not by filtering the list through
   search. Start with workspace A selected, type a query matching only workspace B,
   and B is correctly highlighted and returned by `SelectedID()`. If a background
   `SetChoices`/`SetWorkspaces` arrives before Enter, the restore uses the old A
   identity, marks it missing from the filtered results, and `SelectedID()` now
   returns A even though B is still displayed. Enter confirms A; the editor can
   therefore save a different workspace from the user's visible choice.

   An independent temporary-overlay test reproduced this with Alpha/Beta options,
   selected A, search text `Beta`, and an unchanged option refresh. Selection was B
   before refresh, then A after refresh and confirmation. This is introduced by
   the new restoration logic and violates the preservation of selection/intent
   across asynchronous option refreshes. Capture the actual current selection
   before replacing options, while preserving an already unresolved selection
   when appropriate. Ensure search and other supported list-navigation paths
   cannot leave the saved identity behind the visible choice. Add a regression
   covering search -> refresh -> Enter, alongside the arrow-navigation and
   missing-reference tests.

2. **[P2] Restore the workspace mutation in the migrated cleanup regression.**
   `internal/api/card_editor_legacy_helpers_test.go:21` builds changes only for
   Shared and Name. Its caller,
   `TestCardDeletionRetriesPendingWorkspaceCleanupAfterRestart` in
   `internal/api/card_cleanup_internal_test.go`, changes `updated.WorkspaceID`
   while a temporary-workspace copy is pending and relies on that edit to prove
   deletion retains the original activation's cleanup identity. The migrated
   helper silently ignores that edit, and the test never checks the returned or
   persisted workspace. It now passes without exercising the defining concurrent
   workspace-change scenario. This finding concerns the test adapter, not a claim
   that production `UpdateEditorCard` ignores WorkspaceID changes.

   An independent overlay test using the real service and two discovered
   workspaces confirmed that this helper returns the original workspace after
   receiving an updated card with the second workspace. Migrate the cleanup test
   directly to a presence-aware WorkspaceID change, or make the test adapter carry
   the intended edit. Assert the persisted workspace actually changed before
   deletion so the scenario cannot silently weaken again. Keep the admission,
   foreign-session and retry assertions intact.

## Verification

Passed independently:

- `go test ./internal/api ./internal/ui/views ./internal/app ./internal/config ./internal/architecture -count=1`
- `go test -race ./internal/api ./internal/app -run 'Card|CommandClosures' -count=1`
- `go build ./...`
- `go vet ./...`
- `git diff --check`
- `golangci-lint run --new-from-rev=HEAD --whole-files=false` — zero issues

Both additional expectation tests supplied through temporary `go test -overlay`
files failed with the specific results described above. All temporary test and
overlay files were removed; they were not added to the implementation tree.
The ordinary focused suite and race selection remain green without these probes.
Full lint was not rerun; this is not a claim that the previously recorded
architecture-checker lint debt is resolved.

The new contracts separate editable values, explicit changes and expected
revision from stored records. Application updates preserve unexposed fields and
use the existing card-admission boundary; revision tests cover stored fields,
map order, nil/empty distinctions and hidden-field conflicts. Returned revision
tests cover the YAML-normalized persisted representation. The editor no longer
round-trips a stored card for saves, and async save/load correlation and duplicate
save protection have focused coverage. These positives do not replace the two
missing scenarios above.

No native-model or manual terminal E2E run was performed. The independent probes
used isolated unit-test fixtures and no model provider.

## Readiness and blockers

- `ireview.md` **was created**.
- **Not ready to advance:** resolve findings 1–2, add meaningful regressions, and
  repeat final review.
- No user question or environment blocker. No commit was made.
