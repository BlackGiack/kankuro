# Preview Fix Plan

## Goal
Remove preview lag and restore live-feeling preview updates so session selection and scrollback behave as immediately as they did in the old UI.

## Tasks
- [ ] Task 1: Change preview update flow in `internal/ui/views/dashboard.go`, `internal/ui/components/session_list.go`, `internal/ui/coordinator.go`, and `internal/ui/messages/messages.go` so session selection emits an immediate preview-update message instead of waiting for the periodic `refreshData()` tick. Verify: moving the session cursor updates the preview body instantly in the same interaction cycle.
- [ ] Task 2: Split preview metadata updates from preview scrollback refreshes: selection changes should immediately set the selected session snapshot/card in the dashboard view, while scrollback should be fetched right away from `session.Manager.GetScrollback()` for the newly selected session. Verify: preview header and scrollback both change as soon as the selected session changes, with no visible 2-second lag.
- [ ] Task 3: Restore streaming-like preview refresh by subscribing the dashboard preview to session output updates instead of only repopulating scrollback on the coordinator tick. Use `session.SubscribeOutput`/`UnsubscribeOutput` or an equivalent coordinator-owned live refresh path. Verify: while a session emits output, the preview updates incrementally without waiting for batch refreshes.
- [ ] Task 4: Keep the already-restored stable-selection semantics from `f2788fb`: preview should stay on the chosen session when switching away from the sessions pane, recover when the previewed session disappears, and clear when no sessions remain. Verify: the existing persistence/recovery scenarios still pass after the live-refresh changes.
- [ ] Task 5: Ensure preview updates are scoped to the actively previewed session only, with proper cleanup when the selected preview session changes or when leaving the dashboard. Verify: changing preview target does not continue streaming stale output from the previously previewed session.
- [ ] Task 6: Reconcile the preview tests with the new live path by adding integration coverage for immediate selection updates and streaming refresh behavior in `internal/ui/views/dashboard_test.go` and `internal/ui/integration_test.go`. Verify: tests explicitly catch selection lag and non-streaming regressions.

## Done When
- [ ] Preview changes instantly when the highlighted session changes.
- [ ] Preview output feels live again instead of arriving only in batches.
- [ ] Existing persistence/recovery preview behavior remains intact.

## Notes
- Use `old_ui/ui/view_preview.go` and `old_ui/ui/app_preview_test.go` as the semantic reference, then extend that behavior to the current componentized architecture.
- The current KO is caused by event timing, not only rendering; fix the coordinator/view message path, not just the preview component.
