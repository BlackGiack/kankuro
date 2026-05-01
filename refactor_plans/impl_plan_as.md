# Agent Switch Fix Plan

## Goal
Restore the old dedicated agent-switch flow so switching agents on an existing session behaves like its own focused screen, with correct filtering and escape semantics.

## Tasks
- [ ] Task 1: Rework the switch-agent flow in `internal/ui/views/dashboard.go`, `internal/ui/views/switch_modal.go`, and `internal/ui/coordinator.go` so agent switching is presented as a dedicated view instead of an overlay rendered on top of the dashboard. Verify: pressing `s` on a terminated session opens a full-screen switch selector with no dashboard visible underneath.
- [ ] Task 2: Port the old dedicated switch rendering and control surface from `old_ui/ui/app_switch.go` and `old_ui/ui/view_switch.go` into the new flow, preserving the session-context header and dedicated action help. Verify: the switch screen shows the target session context and visible `Enter=switch Esc=cancel` guidance.
- [ ] Task 3: Add explicit switch-selector search state instead of relying only on the embedded list defaults, so the selected filter text is owned by the switch view and can be cleared independently from closing the screen. Verify: typing filters cards immediately and the current filter is visible to the user.
- [ ] Task 4: Make `esc` context-sensitive in the switch selector: if a search/filter is active, clear the filter first; only close the selector when the filter is already empty. Verify: `esc` removes the active filter without leaving the switch screen, and a second `esc` cancels the switch flow.
- [ ] Task 5: Preserve correct switch execution semantics in `internal/ui/callbacks.go`: selecting a card must still use the current restored switch-agent path, refresh sessions immediately, and reattach or reactivate the switched session exactly as intended by the restored flow. Verify: switch on a terminated session restarts it with the chosen card and lands the user in the expected attached/active state.
- [ ] Task 6: Ensure switch view key handling blocks dashboard-global bindings while active, especially quit-adjacent keys and pane navigation keys that should not leak through. Verify: while in switch view, dashboard controls do not fire underneath it.
- [ ] Task 7: Add focused tests for switch search, `esc`-clears-filter-before-close behavior, dedicated rendering, and successful switch execution path. Verify: current KO cases are covered by automated tests.

## Done When
- [ ] Agent switching is a dedicated screen, not an overlay under the dashboard.
- [ ] Filtering inside the switch selector works and can be cleared without closing the selector.
- [ ] The switch flow still performs the restored session-switch behavior end to end.

## Notes
- Use `old_ui/ui/app_switch.go` and `old_ui/ui/view_switch.go` as the behavior reference for dedicated rendering and cancel/confirm flow.
- Keep the current refactored coordinator structure; the fix should be a proper routed view, not a dashboard-specific hack.
