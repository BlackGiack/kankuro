fix-delete-modal-esc
Goal
Fix Esc key not canceling the session deletion confirmation modal in the dashboard.
Tasks
- [ ] Task 1: Add an Esc-cancel test case in dashboard_test.go (alongside the existing TestDashboard_DeleteModal test) that sends tea.KeyMsg{Type: tea.KeyEsc} after opening the modal and asserts the modal closes and pending action is cleared → Verify: go test ./internal/ui/views/ -run TestDashboard_DeleteModal -v
- [ ] Task 2: If the test passes, the code is correct and only the test coverage was missing — skip to verification. If the test fails, investigate at dashboard.go:637-654 why case keyEsc doesn't match at runtime (possible causes: keyMsg.String() returning unexpected value, or key constant mismatch)
- [ ] Task 3: If needed, fix the key matching in updateDeleteModal (e.g., compare keyMsg.Type instead of keyMsg.String(), or add explicit tea.KeyEsc type check)
Done When
- [ ] Pressing Esc on the delete confirmation modal closes it without deleting
- [ ] Existing tests pass (go test ./internal/ui/...)
- [ ] New test covers both confirm (y/Enter) and cancel (n/Esc) paths
Notes
- The keyEsc constant is "esc" at views/constants.go:6, matching tea.KeyEsc.String() return value
- Root cause could be a runtime key encoding issue (terminal-level) not visible in static analysis — the test will confirm
- The modal content at dashboard.go:584 says "Delete session and its temporary workspace?" — consider adding key hint text like "(Enter=Confirm / Esc=Cancel)" as a UX improvement once the functional bug is fixed