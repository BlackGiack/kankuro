# Workspace Picker Crash Fix Plan

## Goal
Fix the workspace picker panic that happens after searching/selecting/saving a workspace and then reopening the picker, especially on small terminal sizes. The known crash is `runtime error: slice bounds out of range [14:2]` and the stack goes through `internal/ui/views/workspace_picker_modal.go:153` into `bubbles/list/list.go`.

## Important Root Cause
The crash is not in our string concatenation. It is almost certainly caused by bad internal `bubbles/list` state when the picker is reopened:

- the picker keeps stale search/filter/pagination/cursor state from the previous open
- the visible item count after filtering is smaller than the old page/cursor state expects
- on a small terminal, `bubbles/list` has fewer rows per page, so invalid page slice bounds are easier to hit
- when `m.list.View()` runs, `bubbles/list` tries to slice the visible items with stale paginator bounds and panics

This means the fix must normalize the picker state every time the picker is opened, resized, filtered, confirmed, or cancelled.

## Tasks
- [ ] Task 1: Audit and document the current picker lifecycle in `internal/ui/views/workspace_picker_modal.go` and `internal/ui/views/card_editor.go`.
  Verify:
  - confirm exactly when `SetSelectedID`, `Open`, `Update`, `SelectedID`, and `Close` are called
  - confirm the crash flow matches: open -> search -> select/save -> reopen -> render

- [ ] Task 2: Add a dedicated picker reset path in `WorkspacePickerModal`.
  Implementation details:
  - add a private helper such as `resetTransientState()` or `prepareForOpen()`
  - it must clear all transient state that should not survive between openings:
    - `searchQuery = ""`
    - any stale cursor/selection that points into filtered results
    - any stale paginator/page state indirectly stored in `m.list`
  - do not clear persistent state such as `workspaces` or `selectedID`
  - call this helper from `Open()` before the picker becomes visible
  Verify:
  - reopening the picker always starts from a clean unfiltered list
  - reopening after a previous search does not reuse the old search string

- [ ] Task 3: Rebuild the embedded `list.Model` safely on open, not only the items slice.
  Implementation details:
  - stale state likely lives inside `list.Model`, not just in our fields
  - either:
    - fully recreate the `list.Model` in a helper like `newListModel()` and assign it on open, or
    - explicitly normalize page/cursor/selection after every refresh
  - if recreating the list model:
    - reuse the same delegate configuration as `NewWorkspacePickerModal`
    - reapply title/status/filter flags
    - reapply current size with `SetSize`
    - repopulate items
  - this is the safest option because it clears hidden paginator state in `bubbles/list`
  Verify:
  - after reopening, `m.list.View()` does not panic even after previous filtering

- [ ] Task 4: Fix selected workspace restoration so it uses visible list items, not the raw workspace slice index.
  Implementation details:
  - current `SetSelectedID()` loops over `m.workspaces` and calls `m.list.Select(i)`
  - that index can be wrong whenever the list items are filtered or rebuilt differently
  - change the logic to:
    1. ensure current visible items are already refreshed
    2. iterate over `m.list.Items()`
    3. find the `workspaceItem` whose `workspace.ID` matches the selected ID
    4. call `m.list.Select(visibleIndex)`
  - if the selected workspace is not visible, select index `0` if there are visible items
  - if there are no visible items, do not select anything invalid
  Verify:
  - reopening highlights the saved workspace correctly
  - no selection code uses raw `m.workspaces` index against filtered `m.list.Items()`

- [ ] Task 5: Clamp picker and list dimensions to safe minimums in `WorkspacePickerModal.SetSize()`.
  Implementation details:
  - current code computes:
    - list width = `width - widthPadding*2 - 4`
    - list height = `height - heightPadding*2 - 8`
  - on small terminals these can become too small or non-positive
  - clamp values before passing them into `m.list.SetSize(...)`
  - minimum safe values:
    - picker width >= 1
    - picker height >= 1
    - list width >= 1
    - list height >= 1
  - if the picker needs more room for help/header than available, still keep the inner list height at least `1`
  Verify:
  - tiny terminal sizes do not produce negative or zero slice math in `bubbles/list`
  - calling `View()` on a tiny picker no longer panics

- [ ] Task 6: Normalize selection after every search mutation.
  Implementation details:
  - after typing a rune, backspace, or clearing search with `Esc`:
    - call `refreshItems()`
    - then ensure selected index is valid for the new filtered list
  - add a helper such as `normalizeSelection()`
  - logic for `normalizeSelection()`:
    - get `items := m.list.Items()`
    - if `len(items) == 0`, do nothing or reset to safe empty state
    - if current selected index is out of range, select `0`
  - if using full list recreation, still run this after repopulating
  Verify:
  - filtering from many results to few results never leaves the list on an invalid selection/page
  - clearing search from `Esc` also leaves a valid selection

- [ ] Task 7: Normalize state on confirm and cancel before closing.
  Implementation details:
  - when `Enter` confirms:
    - compute selected workspace safely first
    - then reset transient search/list state before the next open
  - when `Esc` cancels with empty search:
    - mark cancelled
    - also reset transient search/list state before the next open
  - avoid leaving the picker closed with stale filtered list internals
  Verify:
  - confirm and cancel both prepare the picker for a safe future reopen

- [ ] Task 8: Add a regression test for the exact reported crash flow in `internal/ui/views/workspace_picker_modal_test.go` or similar.
  Implementation details:
  - create the picker
  - set a small size that mimics the problematic environment
  - load enough workspaces to require pagination
  - open picker
  - type a search that narrows the list significantly
  - confirm/select a workspace
  - close
  - reopen
  - call `View()`
  - the test must fail on panic if the bug still exists
  Verify:
  - the exact reopen-after-search flow is covered automatically

- [ ] Task 9: Add a second regression test for selected workspace restoration under filtering.
  Implementation details:
  - set a `selectedID`
  - load multiple workspaces
  - apply a search
  - reopen picker
  - ensure selection restoration uses visible items safely
  - this is important because raw-index selection is a likely contributor to stale list state
  Verify:
  - selected workspace highlight works correctly with filtered and unfiltered item sets

- [ ] Task 10: Run targeted and full UI tests.
  Verify:
  - `go test ./internal/ui/views/...`
  - `go test ./internal/ui/...`
  - specifically confirm the new crash regression tests pass on both normal and very small picker sizes

## Done When
- [ ] Reopening the workspace picker after a filtered selection no longer panics.
- [ ] Small terminal sizes no longer produce invalid inner list dimensions.
- [ ] Search state does not leak across picker reopenings.
- [ ] Selected workspace restoration is based on visible list items, not raw workspace slice indices.
- [ ] There is an automated regression test for the exact reported crash flow.

## Notes
- Prefer recreating or fully normalizing `list.Model` on open. This is safer than trying to surgically fix hidden paginator state.
- The likely bug is hidden state inside `bubbles/list`, not just our own `searchQuery` field.
- Be careful not to “fix” this only for large screens. The target environment is a small screen where pagination is tight.
