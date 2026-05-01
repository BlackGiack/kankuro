# Lint Errors Fix Plan

## Goal
Resolve the current `golangci-lint run` failures without regressing the recent UI behavior fixes, especially around preview, workspace picker, and agent switch flows.

## Reported Errors
- `internal/ui/coordinator.go:155`: `funlen` on `Coordinator.Update`
- `internal/ui/components/card_list_test.go:149`: `goimports`
- `internal/ui/preview_integration_test.go:7`: `goimports`
- `internal/ui/views/dashboard.go:125`: `goimports`
- `internal/ui/views/puppet_selector_test.go:14`: `goimports`
- `internal/ui/views/workspace_picker_modal.go:197`: `govet` shadowing `helpStyle`
- `internal/ui/preview_integration_test.go:150`: `mirror` suggests `bytes.Contains`
- `internal/ui/views/agent_switch.go:198`: `mnd` on literal `10`
- `internal/ui/views/card_editor.go:32`: `unused` const `screenSplitDivisor`
- `internal/ui/views/agent_switch.go:167`: `wastedassign` on `searchDisplay`

## Tasks
- [ ] Task 1: Reformat all files failing `goimports`.
  Files:
  - `internal/ui/components/card_list_test.go`
  - `internal/ui/preview_integration_test.go`
  - `internal/ui/views/dashboard.go`
  - `internal/ui/views/puppet_selector_test.go`
  Implementation details:
  - run `goimports -w` on these exact files
  - do not mix behavioral edits into this task
  Verify:
  - rerunning `golangci-lint run` removes all `goimports` errors

- [ ] Task 2: Split `Coordinator.Update` into smaller helpers to satisfy `funlen` without changing coordinator behavior.
  File:
  - `internal/ui/coordinator.go`
  Implementation details:
  - keep `Update` as the public dispatcher, but move grouped branches into helper methods
  - recommended split:
    - `handleCoordinatorMessage(msg tea.Msg) (handled bool, model tea.Model, cmd tea.Cmd)`
    - `handleWindowSize(msg tea.WindowSizeMsg) tea.Cmd`
    - `handleDataLoaded(msg messages.DataLoadedMsg) tea.Cmd`
    - `handleNavigation(msg tea.Msg) tea.Cmd` or similarly named helpers
  - preserve the current ordering of special-case messages before delegating to `updateCurrentView`
  - do not change behavior for:
    - preview subscription requests
    - attach completion
    - session launch completion
    - agent switch request navigation
    - MCP update/delete refresh path
  Verify:
  - `Coordinator.Update` drops below the `funlen` threshold
  - existing UI/integration tests still pass

- [ ] Task 3: Fix the `helpStyle` shadowing warning in the workspace picker view.
  File:
  - `internal/ui/views/workspace_picker_modal.go`
  Implementation details:
  - there is already a package-level `helpStyle` in the views package
  - rename the local variable in `WorkspacePickerModal.View()` to something unambiguous like `pickerHelpStyle`
  - keep the rendering behavior exactly the same
  Verify:
  - `govet` shadow warning disappears

- [ ] Task 4: Replace the preview test string allocation with `bytes.Contains`.
  File:
  - `internal/ui/preview_integration_test.go`
  Implementation details:
  - replace `strings.Contains(string(preview.Scrollback()), string(testOutput))`
    with `bytes.Contains(preview.Scrollback(), testOutput)`
  - update imports accordingly:
    - add `bytes`
    - remove `strings` only if no longer used
  Verify:
  - `mirror` lint error disappears
  - test intent remains unchanged

- [ ] Task 5: Remove the magic number in agent switch sizing by introducing a named constant.
  File:
  - `internal/ui/views/agent_switch.go`
  Implementation details:
  - identify what the trailing `10` in `height-heightPadding*2-10` represents
  - replace it with a named constant near the other view constants, for example:
    - `agentSwitchReservedHeight`
    - `agentSwitchListHeightPadding`
  - keep the computed size exactly equivalent
  Verify:
  - `mnd` error disappears
  - agent-switch rendering tests still pass

- [ ] Task 6: Remove the unused `screenSplitDivisor` constant.
  File:
  - `internal/ui/views/card_editor.go`
  Implementation details:
  - if it is genuinely unused, delete it
  - do not replace it unless there is an actual live use nearby
  Verify:
  - `unused` error disappears

- [ ] Task 7: Remove the wasted initial assignment in agent switch view rendering.
  File:
  - `internal/ui/views/agent_switch.go`
  Implementation details:
  - current pattern initializes `searchDisplay := ""` and then overwrites it in both branches
  - rewrite to one of:
    - `var searchDisplay string` followed by branches
    - direct branch-local assignment without the wasted initial value
  - keep output identical
  Verify:
  - `wastedassign` error disappears

- [ ] Task 8: Run focused tests for the touched areas after the lint fixes.
  Commands:
  - `go test ./internal/ui/views -run WorkspacePicker`
  - `go test ./internal/ui/views -run AgentSwitch`
  - `go test ./internal/ui/...`
  Implementation details:
  - run these before the final lint pass so any behavioral regression is caught early
  Verify:
  - all targeted test commands pass

- [ ] Task 9: Run the final linter pass and confirm zero reported issues.
  Command:
  - `golangci-lint run`
  Verify:
  - no remaining issues from `funlen`, `goimports`, `govet`, `mirror`, `mnd`, `unused`, or `wastedassign`

## Done When
- [ ] `golangci-lint run` passes cleanly.
- [ ] All touched UI tests still pass.
- [ ] No recent workspace picker, preview, or agent-switch behavior is changed by the lint-only cleanup.

## Notes
- Keep this pass narrowly scoped to the currently reported lint issues. Do not opportunistically refactor unrelated files.
- The `Coordinator.Update` split is the only task that can accidentally change behavior. Implement that one carefully and run tests immediately after it.
