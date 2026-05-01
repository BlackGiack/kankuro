# Pills And Puppet Fix Plan

## Goal
Restore missing card pills and the remaining puppet-selector UX issues without disturbing the now-stable card editor flow.

## Tasks
- [ ] Task 1: Fix pills rendering in `internal/ui/components/card_list.go` by making the card delegate render against the live `skills` and `personas` dataset instead of the stale `nil` delegate state created in `NewCardList()`. Update the delegate whenever `SetCards()` runs or rebuild the renderer so pills resolve names from current data. Verify: skills/personas pills reappear on ID cards as soon as cards with capabilities are loaded.
- [ ] Task 2: Confirm MCP-related pills behavior and decide whether MCPs should have dedicated pills on cards or only continue appearing indirectly through restored capability rendering; implement the intended behavior consistently in the same card renderer. Verify: cards show the expected set of capability pills for skills/personas/MCPs after reload and after editing a card.
- [ ] Task 3: Compare current card rendering with `old_ui/ui/app_render.go` and ensure puppet rendering still uses the intended fallback/default style when a card has no puppet set, so cards never show a blank or wrong puppet. Verify: cards with empty puppet config still render the default puppet correctly.
- [ ] Task 4: Add explicit `Esc=back` help to the puppet selector in `internal/ui/views/puppet_selector_modal.go`, using the old guidance from `old_ui/ui/view_puppet_selector.go` as the behavioral reference. Verify: the selector visibly documents cancel/back behavior.
- [ ] Task 5: Review the puppet selector presentation in the card editor and ensure its control help is complete for left/right navigation, confirm, and cancel. Verify: a user can understand selector controls without guessing.
- [ ] Task 6: Add tests for pill rendering and puppet selector help text in `internal/ui/components/card_list_test.go` and `internal/ui/views/card_editor_test.go` or a dedicated puppet-selector test file. Verify: pills presence and `Esc=back` guidance are covered automatically.

## Done When
- [ ] Capability pills are visible again on ID cards.
- [ ] Card puppets render with the intended data/fallback behavior.
- [ ] The puppet selector clearly documents `Esc=back`.

## Notes
- The pills bug is likely a data-flow regression in the card-list delegate, not a styling bug.
- Use `old_ui/ui/app_render.go` and `old_ui/ui/view_puppet_selector.go` as the visual/behavioral reference points.
