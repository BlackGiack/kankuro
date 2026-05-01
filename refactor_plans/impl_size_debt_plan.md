# Sizing Debt Plan
## Goal
Replace fixed size constants with relative/percentage-based calculations across the internal/ui package so the TUI adapts to any terminal size. No functional changes — just make the math dynamic.
## Tasks
- [ ] Task 1: Make card editor input widths relative (views/card_editor.go)
  - Replace inputWidth = 60 with a computed value: max(40, parentWidth*3/5)
  - All 5 textinput.Model instances use inputWidth — they all get the fix
  - Verify: go build ./internal/ui/... succeeds; run card_editor_test.go
- [ ] Task 2: Make AgentsMD editor dimensions relative (views/agentsmd_editor_modal.go)
  - editorWidth = 60 → max(40, parentWidth*8/10) in SetSize()
  - editorHeight = 10 → max(10, parentHeight/3) in SetSize()
  - pathInput.Width = 50 → computed relative to parent width
  - modalWidth = 70 (in View()) → computed relative to parent width
  - Verify: go build ./internal/ui/... succeeds
- [ ] Task 3: Make IDE config input width fully relative (views/ide_config.go)
  - inputMaxWidth = 60 is currently min(60, width-10) — change to min(width*9/10, 80)
  - Verify: go build ./internal/ui/... succeeds; run ide_config_test.go
- [ ] Task 4: Make card item layout dimensions relative (components/card_list.go)
  - CardItemWidth = 42 → computed as max(36, listWidth-4) in RenderCardItem (the render function receives a width param)
  - InfoBoxWidth = 30 → computed as cardItemWidth - faceBoxWidth - borderPadding
  - Verify: go build ./internal/ui/... succeeds; run card_list_test.go
- [ ] Task 5: Make truncation limits relative to container (components/card_list.go, components/session_preview.go)
  - TruncateLength = 28 → computed as a fraction of CardItemWidth: max(20, cardItemWidth - 14)
  - previewTruncateLimit = 25 → computed as a fraction of preview width in SetSize: max(15, width/3)
  - Verify: go build ./internal/ui/... succeeds; run session_preview_test.go
- [ ] Task 6: Make modal/dialog dimensions relative (views/constants.go, components/constants.go)
  - workspacePickerWidth = 50 → max(40, width*2/3)
  - workspacePickerHeight = 30 → max(15, height*2/3)
  - switchModalWidth = 40 → max(35, width/2)
  - switchModalHeight = 10 → max(8, height/3)
  - modalWidth = 50 (components/constants.go) → max(40, width*3/5)
  - These are used in SetSize() calls — pass parent dimensions through
  - Verify: go build ./internal/ui/... succeeds; run all view tests
- [ ] Task 7: Run full test suite — verify nothing broke
  - go test ./internal/ui/... — all tests pass
## Done When
- [ ] All fixed size constants replaced with relative calculations
- [ ] go build ./internal/ui/... compiles
- [ ] go test ./internal/ui/... passes
- [ ] No regressions in UI rendering at different terminal sizes
## Notes
- The dashboard column layout (views/dashboard.go lines 712-739) already uses relative sizing correctly — do not touch it
- The face box dimensions (FaceBoxWidth = 11, FaceBoxHeight = 5) are fine as fixed values (ASCII art constraint)
- The delegate heights (CardDelegateHeight = 8, sessionsDelegateHeight = 4) are fine as fixed values (row consistency)