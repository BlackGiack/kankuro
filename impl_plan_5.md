# Act 2 — Plan 5: Fix the layout bugs plan 2 found

Fixes the 11 bugs listed in `impl_plan_2.md` under "Found, not fixed" (#2–#12; #1 was
fixed in plan 2).

**When to run it.**
- **After plan 3:** Part D rewrites the rendering of the three modals whose state
  plumbing plan 3 refactors.
- **Not in parallel with plan 4:** step B3 adds a guard to `internal/app/coordinator.go`,
  the file plan 4 splits. Run 5 and 4 one after the other, in either order. If plan 4
  lands first, step B3 goes into its routing (`Update` / `View`) instead.

**Size.** This plan is bigger than the others. If one session isn't enough, split after
Part B. Parts A–B (text, dashboard) and Parts C–D (card editor, modals) are independent.

**Rules for this session:** don't commit unless asked. Keep all five gates clean.

## How every fix in this plan is proven

Each bug already has a failing test, skipped with
`t.Skip("bug: … 'Found, not fixed' #N")`. Before this plan was written, all 14 skipped
tests were removed from their skips in a scratch copy and **all 14 failed**. So for each
fix:

1. Delete the `t.Skip` line. Run the test and see it fail on the unfixed code.
2. Fix the code at the source of the overflow, then see the test pass.
3. **Don't fix by clamping the output after the fact** (`ClampToSize`, `MaxWidth`,
   `ansi.Truncate` on a whole rendered view). §5 of `tech_debt_act_2.md` explains why: a
   backstop keeps the rectangle while the real layout breaks underneath. Truncating one
   *field* to the width it was given is a fix; truncating the whole view is a backstop.
4. Where a sweep excludes a known cause (`uitest.NamesWithoutTabs()`, the card editor's
   `minH`, the dashboard's 80-column floor), remove the exclusion once that cause is
   fixed, so the general sweep covers it from then on.

Measure everything by display width (`ansi.StringWidth` / `lipgloss.Width`), never
`len` (`project.md` invariant 9).

---

## Part A — text helpers and truncation by rune (#2, #4)

### A1 — two exported helpers in `components`

`session_preview.go` already has correct rune-aware truncation:
`truncateToVisibleWidth` (keeps the tail, exported as `TruncatePath`) and
`truncateToVisibleWidthLeft` (keeps the head, unexported as `truncateString`). Move both
into a new `components/text.go` with their tests, and export:

- `TruncateTail(s, w)`: keeps the end. Keep `TruncatePath` as an alias, or switch its
  callers over.
- `TruncateHead(s, w)`: keeps the start.
- `DisplayText(s string) string`: for **user-provided** text shown on one line (session
  and card names, workspace names, puppet names, capability labels, search queries). It
  replaces `\t`, `\n` and `\r` with a space, and drops other C0 control bytes and ANSI
  escape sequences. Use `ansi.Strip`; this is user text, so it shouldn't carry styling.

Test `text.go` with a table: CJK and emoji at every cut position (the result is valid
UTF-8 and `≤ w` wide), `w` smaller than the ellipsis, and tabs and escapes in
`DisplayText`.

### A2 — #2: byte slicing

- `SessionListItem.Description` (`session_list.go`): replace the `wsPath[len(wsPath)-…]`
  slice with `TruncateTail(wsPath, maxWorkspaceLen)`.
- `CardEditor.renderAgentsMDRow` (`card_editor.go`): use `TruncateHead` for the content
  summary and `TruncateTail` for the path, both with `summaryLimit`.

Unskip `TestSessionListDescriptionIsValidUTF8` and
`TestCardEditorAgentsMDSummaryIsValidUTF8`.

Then grep for any other spot that truncates by bytes:
`grep -rnE '\[:[a-zA-Z]*(Limit|Len|Width|width)|\[len\([a-zA-Z.]+\)-' internal/ui --include=*.go | grep -v _test`.
Fix each display-path hit the same way. `puppet_selector_modal.go`'s `maxLineWidthStatic`
measures with `len`; that one is handled in D2.

### A3 — #4: tabs in names

Apply `DisplayText` where user text enters a one-line slot. The session list's `Title`
(name) and `Description` (card name) are the ones the test covers. Also apply it in the
card list name (`card_list.go`), the dashboard search line (B1), and the card editor's
workspace, puppet and capability lines (C1). Sanitise at display time, not when saving:
the stored name stays what the user typed.

Unskip `TestSessionListNameWithTabsFitsWidth`, then replace `uitest.NamesWithoutTabs()`
with `uitest.AdversarialNames` in every sweep that used it (`session_list_layout_test.go`,
`dashboard_layout_test.go`, `card_editor_layout_test.go`, `modals_layout_test.go`). If one
of those sweeps goes red, the tab is entering through a slot you haven't sanitised yet;
fix that slot, not the sweep. Delete `NamesWithoutTabs` from `uitest` once nothing uses it.

---

## Part B — dashboard (#7, #3, #5, #6)

### B1 — #7: the search/filter line wraps the pane title

`renderCardsPane` and `renderSessionsPane` build `"Search: "+query` / `"Filter: "+query`
untruncated. Fit it to the pane's inner width (pane width − `paneBorderRing` − the title
style's horizontal padding of 2) with `TruncateTail` on the **query**, keeping the label.
The end of the query is where the user is typing, so it has to stay visible. Pass the
query through `DisplayText` first.

Factor the two near-identical `switch` blocks into one `renderPaneTitle(title, query,
active, width)` while you're there. Unskip `TestDashboardLongSearchQueryFits`. The plan 2
comment in `TestDashboard…Search…` about using a 4-character query "for that reason" can
then go: use a long query.

### B2 — #3: the footer is a fixed 150 columns

`renderFooter` joins four fixed groups. Make it fit:

1. **Context-sensitive:** show `[Nav]`, the focused pane's group (`[Card]` when the cards
   pane is focused, `[Session]` when sessions is, neither for the preview) and
   `[Global]`. That roughly halves it, and it's more useful anyway.
2. If that still doesn't fit `d.Width`, drop whole groups from the right, keeping
   `[Global]`'s `q=Quit` for last. As a last resort, `TruncateHead` the remaining
   string. That's field-level truncation of one line, which is allowed.

Unskip `TestDashboardFooterFitsWidth`, and extend it to assert the footer shows the
focused pane's keys for each focus state.

### B3 — #5: below 80 columns there is no defined behaviour

`SetSize` forces `cardsWidth ≥ 46` by taking columns from the preview, which goes
negative below about 62 columns, and from 62 to 79 the render grows a row. Plan 2's tests
assumed 80×24 as the minimum. **Make that the documented minimum:**

- Add `const MinTerminalWidth, MinTerminalHeight = 80, 24` in `views/constants.go`.
- Put the guard in the **coordinator**, so it covers every view and not only the
  dashboard (the card editor has the same problem): when `c.width < MinTerminalWidth ||
  c.height < MinTerminalHeight`, `View()` returns a centred
  `"Terminal too small: need 80×24, have W×H"` placed in exactly `c.width × c.height`
  (`lipgloss.Place`). `Update` then drops every `tea.KeyPressMsg` except quit (`q`,
  `ctrl+c`), so a blind keypress can't launch or delete anything. Non-key messages
  (ticks, data, sessions) keep flowing, so the screen is current the moment the terminal
  grows.
- Leave `Dashboard.SetSize` alone for sizes below the minimum, but make it never produce
  a negative width (`max(0, …)`). It's still called at those sizes.
- Unskip `TestDashboardBelowMinimumWidth`. It should now test the coordinator's
  too-small screen rather than the dashboard: exact `w × h` at widths 20–79 and heights
  5–23, plus a keypress (`n`, `Enter`) that doesn't change `CurrentView()`.
- Document the minimum: `project.md` §4 (below the dashboard keys) and the UI section of
  `kankuro_specs.md`.

The alternative is a degraded layout (hide the preview pane below 80 columns, keeping the
cards and sessions panes usable). It's more work and adds a second layout to test.
Mention it to the user as a possible follow-up; don't build it here.

### B4 — #6: opening the delete modal shrinks the screen by four rows

`renderWithOverlay` places the modal in `d.Width × d.Height`, but `renderBase` renders
`d.Height + 4` rows (header with margin, bordered body of `contentHeight` =
`d.Height + 1`, footer). Add
`func (d *Dashboard) renderedHeight() int`, derived from the same terms `SetSize` and
`renderBase` use. Name the terms as constants instead of the comment's arithmetic. Use it
in `renderWithOverlay`.

Also cap the modal's width in `components.Modal.SetSize`: `min(width, max(40,
width*3/5))`. That's moot above the 80-column minimum, but it keeps the component honest.

Unskip `TestDashboardDeleteModalFillsTheScreen`.

---

## Part C — card editor (#9, #8)

### C1 — #9: long lines are not clipped

- `renderWorkspaceLine` / `renderPuppetLine`: `TruncateHead(DisplayText(summary), e.Width
  - prefixWidth)`, where `prefixWidth` is the display width of `"> Workspace: "` /
  `"> Puppet: "`. Use the focused, longer prefix for both states, so the line doesn't
  change width when focus moves onto it.
- `CapabilityList.View` (`capability_list.go:245`): truncate each label to `c.Width −`
  the width of `prefix + "[x] "`.
- The text inputs: `calculatedWidth = max(40, width*3/5)` plus the `"Name: "` label can
  exceed the width. Make it `min(width − widestLabel, max(40, width*3/5))`.
- While in `renderFormAboveLists`: it still addresses inputs as `e.inputs[0]` …
  `e.inputs[4]`. That's the same disease as §3c, which plan 6 cured elsewhere in this
  file. Switch to the existing `nameInputIdx` … `useTmpInputIdx` constants. Afterwards,
  `grep -n 'inputs\[[0-9]' internal/ui/views/card_editor.go` must be empty.

Unskip `TestCardEditorLongLinesAreClipped`.

### C2 — #8: the form is taller than a small terminal

At 80×24 with default settings, the editor renders 30 lines in a 20-line box. The fixed
rows above the lists (seven field lines, AGENTS.md block, blanks: about 13), plus three
lists each floored at `minCapListHeight` with separators, plus help add up to more than 20.
Shrinking the lists further can't fix it. **The form has to scroll.**

Design:

- Render the form as an ordered list of **blocks**, each tagged with the focus target(s)
  it contains: one block per field line, the AGENTS.md block, one per capability list.
  The help line is not a block; it's pinned to the last row.
- Keep `e.scrollOffset` (in lines). After any focus change and in `SetSize`, adjust it
  minimally, so the focused block is fully inside the window
  `[scrollOffset, scrollOffset + e.Height − helpLineHeight)`.
- When content is cut above or below, replace the first/last visible row with a dim
  `↑ more` / `↓ more` line. Count those rows in the window arithmetic.
- Capability lists already window their own items around their cursor. Keep giving them
  the height `SetSize` computes, floored at `minCapListHeight`. Block-level scrolling
  only has to keep the list block itself on screen.
- Modals keep rendering over the full size; they don't scroll with the form.
- When everything fits, `scrollOffset` is 0 and the output is identical to today's. The
  existing sweep at large sizes guards that.

Tests:
- Unskip `TestCardEditorFitsSmallTerminal`.
- Add **`TestCardEditorFocusedFieldAlwaysVisible`**: at 80×20 and 80×24 (in terminal
  rows), for all four section settings, Tab through every visible focus target. After
  each Tab, `View()` must stay within `W × H` and contain the focused target's marker
  line (`"> "`, or the field's highlighted label).
  *Can-fail check:* freeze `scrollOffset` at 0 and confirm the test goes red at the lower
  targets.
- Remove the `minH` exclusion from the card editor sweep and sweep heights from the 24-row
  minimum (`MinTerminalHeight − heightAdjustment`).

---

## Part D — the three modals (#10, #11, #12)

Plan 3 has already moved their open/confirm/cancel state into `components.ModalState`.
This part changes only sizing and rendering.

### D1 — #10: the workspace picker isn't placed

`CardEditor.View` returns `e.workspacePicker.View()` bare, while the other two go through
`renderWithOverlay`. Route it through `renderWithOverlay` too. In
`WorkspacePickerModal.SetSize`, cap both dimensions at the space given:
`modalWidth = min(width, max(40, width*2/3))`, `modalHeight = min(height, max(15, …))`.
Check lipgloss v2's `Width`/`Height` semantics (the dashboard comments say border-box)
against the test rather than assuming them. Unskip
`TestWorkspacePickerModalHasTheEditorsSize`.

### D2 — #11: the puppet selector's width comes from the art

`PuppetSelectorModal` has no `SetSize` of its own. The card editor's
`e.puppetSelector.SetSize(width, height)` reaches `component.Base`, so `m.Width` /
`m.Height` are set, but `View` ignores them and joins every puppet side by side.

- **Cell size:** every puppet renders in a fixed cell. That's the largest built-in
  puppet's `Bounds()`, computed once. User puppets are clipped into it with the card
  list's `clipBlock` (`card_list.go:451`; export it as `components.ClipBlock` rather than
  copying it). Invariant 9: clip art, don't wrap it.
- **Horizontal window:** show as many cells as fit in the modal's inner width
  (`m.Width` − border − padding), keep the cursor's cell inside the window, and show `‹`
  / `›` at the edges when there's more.
- `maxLineWidthStatic` measures with `len`. Switch it to `ansi.StringWidth`. If
  `ClipBlock` replaces its only use, delete it.

Unskip `TestPuppetSelectorModalClipsUserPuppets`. Add a test that moves the cursor past
the visible window and asserts the selected puppet is still drawn (with its highlight
border) and the shape still holds.

### D3 — #12: the AGENTS.md editor overflows

There are three causes, one per skipped test:

- **6 columns too wide** (`…HonoursItsWidth`): the text area is `width*8/10` wide, inside
  a box that is `width*4/5` wide *including* border and padding, plus a 2-column `"> "`
  marker. Derive the sizes from the inside out, in `SetSize`, not in `View`:
  `modalWidth = min(width, max(50, width*4/5))`,
  `inner = modalWidth − 2 (border) − 2*hPadding`,
  `editorWidth = inner − markerWidth`.
- **Text rows wrap and double the height** (`…WrapsLongLines`): `"> " + m.editor.View()`
  prefixes only the first line of a multi-line view, and the result exceeds `inner`.
  Render the marker as its own column (`lipgloss.JoinHorizontal` of a 2-column
  marker block, `"> "` on the first row, or `"  "` when the field isn't focused, and the
  editor view), so both focus states have the same width. Do the same for the path input.
- **Taller than 24 rows** (`…FitsSmallTerminal`): `editorHeight = max(10, height/3)`
  ignores the fixed rows around it. Compute
  `chrome` = border (2) + `2*vPadding` + the fixed text rows (title, blank, `Content:`,
  blank, `Path:`, path input, blank, help). Then
  `editorHeight = max(3, min(max(10, height/3), height − chrome))`.

Unskip all three AGENTS.md tests.

---

## Part E — close out

- Run `grep -rn 't.Skip("bug' internal/ui`. It must be empty. Run
  `grep -rn NamesWithoutTabs internal`. It must be empty.
- `impl_plan_2.md` "Found, not fixed": mark each entry **FIXED (plan 5)**, with a one-line
  note on how.
- `tech_debt_act_2.md` §5 Status: all 12 found bugs fixed; the minimum terminal size is
  80×24.
- `project.md`: §4, the minimum terminal size and the context-sensitive footer; §7, the
  `components/text.go` helpers (`TruncateHead`, `TruncateTail`, `DisplayText`) as the way
  to put user text into a fixed slot; §11, any quirk that remains (for example the
  degraded-layout follow-up, if the user wants it).
- Run the app once by hand with the isolated config (`project.md` §7): an 80×24 terminal,
  a long search query, a card with a long workspace name, the AGENTS.md modal. Then
  shrink the terminal below 80 columns and grow it back. Tests can't see resize
  behaviour.

## Done when

- [ ] All 14 skipped tests are unskipped and pass; each was seen to fail before its fix.
- [ ] The sweep exclusions (`NamesWithoutTabs`, `minH`, the 80-column floor) are gone and the sweeps are green.
- [ ] No fix is a whole-view clamp; every fix truncates or sizes one field, block or cell.
- [ ] `TestCardEditorFocusedFieldAlwaysVisible` passes and goes red with scrolling disabled.
- [ ] `grep -n 'inputs\[[0-9]' internal/ui/views/card_editor.go` is empty.
- [ ] Docs updated (`project.md`, `kankuro_specs.md`, `tech_debt_act_2.md`, `impl_plan_2.md`).
- [ ] The manual check above has been done.
- [ ] All five gates clean.
