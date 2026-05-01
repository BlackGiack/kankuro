# Workspace List Fix Plan

## Goal
Restore the workspace picker behavior so it matches the old dedicated workspace view flow while fitting the refactored coordinator/view architecture.

## Tasks
- [ ] Task 1: Replace the current overlay-style workspace picker flow in `internal/ui/views/card_editor.go` and `internal/ui/views/workspace_picker_modal.go` with a dedicated full-screen workspace-picker view or a dedicated view-mode inside the card editor that renders alone, not underneath the editor. Verify: opening workspace selection from the card editor shows only the workspace picker on screen.
- [ ] Task 2: Port the old searchable workspace interaction from `old_ui/ui/app_workspaces.go` into the new picker model: maintain an explicit search query, handle printable input and backspace, rebuild the filtered workspace list on every keypress, and keep cursor selection valid when the filter changes. Verify: typing filters the workspace list immediately and backspace restores removed matches.
- [ ] Task 3: Stop `q` from quitting while the workspace picker is open by intercepting it before the underlying Bubble list default handler runs, and scope `ctrl+c` as the only quit path there. Verify: pressing `q` inside workspace selection does nothing to the app process and leaves the picker open.
- [ ] Task 4: Make `esc` context-sensitive in the workspace picker: if a search query is active, clear the search first; otherwise return to the card editor without mutating the selected workspace. Verify: `esc` clears filter on first press and only exits the picker when the search is already empty.
- [ ] Task 5: Add explicit in-view help for workspace picker controls, including `Esc=back`, search behavior, and selection confirmation. Verify: the dedicated workspace screen visibly documents `Esc`, selection, and search usage.
- [ ] Task 6: Preserve edit-form state around workspace picker round-trips in `internal/ui/views/card_editor.go` and the coordinator callbacks, so returning from workspace selection restores the exact editor state and focus. Verify: open picker from a partially edited card, search/select/cancel, and return to the editor with previous field values intact.
- [ ] Task 7: Add tests for the workspace picker flow in `internal/ui/views/card_editor_test.go` or a dedicated picker test file: search input, backspace filtering, `q` non-quit behavior, `esc` search-clear behavior, dedicated rendering, and return-to-editor behavior. Verify: tests fail on the current implementation and pass after the fix.

## Done When
- [ ] Workspace selection is searchable again.
- [ ] The workspace picker no longer quits on `q`.
- [ ] The workspace picker is rendered as its own screen instead of an underlay/overlay beneath the editor.
- [ ] `Esc` behavior is documented and works as back/cancel without surprising the user.

## Notes
- Use `old_ui/ui/app_workspaces.go` as the behavioral reference for search and escape semantics.
