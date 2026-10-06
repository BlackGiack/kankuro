# Act 2 — Plan 2: Rendering-shape tests

Covers `tech_debt_act_2.md` **§5**: nothing checks what the views look like.
Plan 2 of 4; see the table in `impl_plan_1.md`. Run it after plan 1 (no shared files)
and **before plan 3**, because plan 3 refactors the modals and these tests guard that
refactor.

**Rules for this session:** don't commit unless asked. Keep all five gates clean. Every
assertion added here must be seen to fail at least once, either against a bug it finds
or against a fix deliberately reverted. A shape test that has never been red proves
nothing. That was Act 1's recurring failure.

---

## Scope and expected outcome

This plan writes tests, and it will probably find bugs. Fixing what it finds is in
scope, with a limit: **if the sweeps turn up more than about three distinct layout bugs,
stop after writing the tests.** Mark the failing cases `t.Skip("bug: …")` with a short
description, list them at the bottom of this file under "Found, not fixed", and plan the
fixes as a follow-up. Without that limit, "write tests" turns into an open-ended layout
rewrite.

Order, by the risk ranking in §5: session list → session preview → dashboard → card
editor → modals. Each step is self-contained. If the session runs long, stopping after
any step leaves the work in a usable state.

---

## Step 0 — decide what each view promises

The card test can assert "every line has the same width" because `RenderCardItem` pads
to a fixed box. Not every component does. `SessionList.View()` is a bare
`bubbles/list` `View()`, which does **not** pad lines to the width. Asserting exact width
there would fail by design, not because of a bug.

So for each target, first write down its contract in a comment at the top of its test.
There are two kinds:

- **Exact:** every line is exactly `W` columns wide and there are exactly `H` lines.
  This applies to anything that draws a border or is placed into a fixed box: the
  dashboard panes, the full dashboard, `lipgloss.Place` overlays, and the card editor if
  it fills its size.
- **Bounded:** no line is wider than `W` and there are at most `H` lines. This applies
  to unpadded components (`SessionList`, probably `SessionPreview`'s inner body), which
  their parent pads.

Find out which kind each one is by reading its `View` / `SetSize`, then confirm by
rendering it once. If you can't tell what a view was *meant* to do, choose the stricter
contract the parent relies on, and note that you chose it.

Also check the height contract the coordinator sets: views are sized to
`(width, height - heightAdjustment)` with `heightAdjustment = 4` (`coordinator.go`), and
the dashboard's `SetSize` comment says its full render is `height + 4` lines tall
(header, body of `height + 1`, footer). Whatever the real number turns out to be, pin it
down in the dashboard test as a named constant with that comment as its justification.

## Step 1 — shared assertion helpers

The card helpers live in package `components_test`. The views tests are in package
`views` / `views_test`, so the helpers can't be shared from there. Create
`internal/ui/uitest/shape.go`: a normal, non-`_test` package that only tests import.

```go
// AssertExactRect fails unless out is exactly h lines of exactly w display columns.
func AssertExactRect(t testing.TB, out string, w, h int, label string)
// AssertWithin fails if any line is wider than w or there are more than h lines.
func AssertWithin(t testing.TB, out string, w, h int, label string)
// Region returns the rectangle [x, x+w) × [y, y+h) of the ANSI-stripped render,
// for region checks ("this cell must be blank", "this row must not contain X").
func Region(out string, x, y, w, h int) []string
```

Measure with `ansi.StringWidth`, never `len` (`project.md` invariant 9). Make failure
messages print `ansi.Strip(out)`, as `assertIntactCard` does, so a failure shows the
picture.

Also add the adversarial-content fixtures that every sweep below shares:

```go
var AdversarialNames = []string{
	"", "a", strings.Repeat("x", 200), "漢字のセッション名", "🚀 rocket 🚀",
	"name\twith\ttabs", "trailing space   ", "\x1b[31mred\x1b[0m",
}
```

plus a `UserPuppet(w, h int) config.Puppet` builder for a puppet that is wider/taller
than its box.

Then **rewrite `assertIntactCard` on top of `AssertExactRect`**, and confirm the card
sweep still passes, and still fails when you revert `renderPills`' fitting. This proves
the helper is equivalent before you rely on it elsewhere.

Don't move `collectMsgs` or anything in `internal/app` here. That belongs to plan 4.

## Step 2 — `components/session_list.go`

New file `components/session_list_layout_test.go`.

- Sweep: widths from the dashboard's minimum sessions-pane width to 80 (find the
  minimum: `width / dashboardSessionsWidthRatio` at the smallest supported terminal),
  heights 5–30.
- Content: 0, 1 and 20 sessions; names from `AdversarialNames`; every `Status`; cards
  with each built-in puppet face, a user puppet, and `Card == nil` (session whose card
  was deleted); a session with `Pinged` set (the highlighted row).
- Assert: the contract from step 0. **Region check:** each item is a fixed number of
  rows (`Title` + `Description`; read the delegate's height). Assert that the rendered
  item count × item height + chrome never exceeds `h`, and that no row of one item
  bleeds into the next. In practice: the row that must hold the status badge actually
  does, for every item.
- Also sweep with a search query set (`SetSearchQuery`), since the filter line changes
  the chrome height.

## Step 3 — `components/session_preview.go`

`session_preview_overflow_test.go` already covers width/height with wide runes, but only
as `≤` checks at three sizes. Extend it into a sweep in a new
`session_preview_layout_test.go` (leave the old test alone):

- Sweep widths 20–160 and heights 5–50, in steps that land on odd and even values.
- Content: wide runes and emoji split across the clipping edge, a single
  10 000-column line, `\r` overwrites, cursor-movement CSI, OSC 8 hyperlinks, an
  unterminated escape at the end of the buffer, 5 000 lines of scrollback, empty
  scrollback (`renderNoOutput`), and no session at all (`renderPlaceholder`).
- Assert the contract from step 0.
- **Region check / backstop lesson.** `Dashboard.renderPreviewPane` ends with a
  defensive `components.ClampToSize`. That clamp is exactly the kind of width backstop
  §5 warns can hide a real bug. So assert on `SessionPreview.View()` **by itself**,
  without the dashboard's clamp, and verify it by temporarily removing the preview's own
  `clampToSize` call. The test must go red even though the dashboard would still look
  fine.
- Assert that the title/header block keeps its rows. If the preview has a fixed title
  height (`previewTitleBlock` in the dashboard), the first N lines must stay the title
  whatever the content.

## Step 4 — `views/dashboard.go`

New file `views/dashboard_layout_test.go`.

- Build a `Dashboard`, call `SetSize` with the sizes the coordinator would pass
  (terminal 60×20 to 220×60, minus `heightAdjustment`), and feed it `SetData` with cards
  (use the card test's `cardWithSkills`-style fixtures; copy them, since they're in
  another test package, or move them into `uitest`), sessions with adversarial names, and
  preview data with adversarial content.
- Assert:
  1. **The whole render** is an exact rectangle of the promised size (step 0).
  2. **The three panes stay in their columns.** Using `uitest.Region`, the column where
     each pane's right border should be holds a border rune on every body row. A pane
     that grows by one column moves its neighbour's border, and this check catches that
     even if the total width happens to survive.
  3. Every focus state (`PaneCards`, preview, sessions) and the search-active state:
     the focused border changes colour, and the search line adds a row.
- Sweep the **narrow end hard**. `SetSize` forces `cardsWidth ≥ 46` by taking columns
  from the preview. Below roughly 46 + 2 × (minimum useful pane) columns that can drive
  `previewWidth` to zero or negative. Find the smallest width at which the dashboard
  should still render and test across that boundary. If it's undefined, decide (render a
  "terminal too small" message, say) and record it as a found bug rather than silently
  testing around it.
- **Delete-modal state.** Open the delete modal (`Backspace` on a terminated session)
  and check the overlay render. `renderWithOverlay` is `lipgloss.Place(d.Width,
  d.Height, …)`, so it renders `d.Height` lines, while the base dashboard (step 0)
  renders a different height. Check whether the screen jumps when the modal opens. If so,
  that's a finding. Also check widths below 40: `components.Modal.SetSize` sets
  `modalWidth = max(40, width*3/5)`, so the modal is wider than the screen there and
  `Place` cannot pad it down.

## Step 5 — `views/card_editor.go`

New file `views/card_editor_layout_test.go`. `card_editor_setsize_test.go` exists. Read
it first and extend its approach rather than duplicating it.

- Sweep sizes × the four combinations of `ShowPersonasInCardEditor` /
  `ShowMCPsInCardEditor` (these add and remove whole sections, which is the main risk).
- Content: 0 and 40 skills/personas/MCPs with adversarial names, a loaded card with long
  name/model/driver values, every focus target focused in turn (the focused field may
  render differently).
- Assert the contract, and that the help line is the last line for every combination.
  A section that grows pushes it off-screen or wraps it.

## Step 6 — the three modals

Note that `card_editor.go` renders modals as a replacement, not a composite:
`workspacePicker.View()` is returned bare, and the other two go through
`renderWithOverlay` = `lipgloss.Place(e.Width, e.Height, …)`. So "corrupts the view
underneath" in §5 really means "is larger than the screen, so `Place` passes it through
unpadded and unclipped".

For each of `WorkspacePickerModal`, `PuppetSelectorModal`, `AgentsMDEditorModal`, test
through the card editor (open it via its key, then call `e.View()`), and also directly:

- **Workspace picker:** 0, 1 and 200 workspaces; long paths; CJK names; with a filter
  typed. Note that it is returned without `Place`, so its own `SetSize` must produce the
  exact size.
- **Puppet selector:** it has **no `SetSize`**, so its width comes from the art
  (`maxLineWidthStatic`). Test with a user puppet much wider than the terminal. This is
  the likeliest real overflow among the three.
- **AGENTS.md editor:** long lines, many lines, wide runes in the textarea.

Assert that the editor's `View()` while the modal is open has the editor's exact size.

## Step 7 — wire it up and record

- `project.md` §7 "Testing notes": replace the rendering-tests bullet with one that
  names `internal/ui/uitest`, the exact/bounded contracts, and the rule "test a
  component's own output, not only after its parent's clamp".
- `tech_debt_act_2.md` §5: add a short "Status" paragraph listing which views now have
  sweeps and what they found.
- Keep the whole `components` + `views` test run under about 10 s. If the sweeps blow
  past that, step the widths (every column near the boundaries, every 7th column in
  between) instead of trimming content cases. Content variety matters more than width
  density (§5).

---

## Done when

- [x] `internal/ui/uitest` exists; `assertIntactCard` uses it and the card sweep is still green, and still red with fitting reverted.
- [x] Layout tests exist for the session list, session preview, dashboard, card editor and the three modals, each with its contract written at the top.
- [x] Each new test file has been seen to fail at least once (bug found, or fix reverted). Write down which in the test's comment.
- [x] Bugs found are fixed, or skipped with a reason and listed below.
- [x] `project.md` §7 and `tech_debt_act_2.md` §5 updated.
- [x] All five gates clean.

## Found, not fixed

The sweeps found 12 distinct layout bugs, well past the three-bug limit. Review
asked for #1 to be fixed anyway - it is one line, it is the bug users hit first, and
while it was open it hid the whole 80-90 column band from the dashboard sweep - so
that one is **fixed**. The other 11 each had a test that failed and was marked
`t.Skip("bug: ...")` with the number below. **All 11 are now fixed by plan 5** (`impl_plan_5.md`):
the skips are removed and the sweeps cover every cause with no exclusions.
(While they were open, a sweep excluded exactly its known cause - a tab-free name set, a
minimum height - and said so in its header comment.) #5's test, `TestDashboardBelowMinimumWidth`,
became the coordinator's `TestCoordinatorTooSmallScreenIsExact` in `app/size_guard_test.go`.

Ordered by how likely a user is to hit it:

1. **FIXED - the sessions list's help line was wider than the list**
   (`TestSessionListFitsEveryWidth`). bubbles' help is rendered with 2 columns of
   left padding that `Help.Width` does not account for, and at some widths was not
   truncated at all (list widths 17-20, 26-28, 35-36, and below 17 when empty). The
   sessions pane word-wrapped it and grew two rows, on a stock 80- and 120-column
   terminal. `NewSessionList` now calls `SetShowHelp(false)`: the dashboard footer
   already lists the keys, and the list's help advertised `q quit`. The regression
   test sweeps every list width from 10 to 80; the dashboard sweep and search-state
   test now cover 80 columns with sessions present.
2. **FIXED (plan 5)** - `components/text.go` `TruncateTail` / `TruncateHead` measure by display width; `Description()` and the AGENTS.md row use them.
   _Was:_ **User text is truncated by byte** (`TestSessionListDescriptionIsValidUTF8`,
   `TestCardEditorAgentsMDSummaryIsValidUTF8`). `SessionListItem.Description` slices
   the working dir, and `renderAgentsMDRow` slices the AGENTS.md content and path, at a
   byte offset, so CJK or emoji text is cut in the middle of a rune and invalid
   UTF-8 reaches the terminal. `truncateToVisibleWidth` in `session_preview.go` already
   does it right.
3. **FIXED (plan 5)** - the footer shows `[Nav]`, the focused pane's group and `[Global]`, drops groups as the width shrinks, and clips only as a last resort (`footerText`).
   _Was:_ **The dashboard footer is a fixed 150 columns** (`TestDashboardFooterFitsWidth`),
   so it is cut off on any terminal narrower than that.
4. **FIXED (plan 5)** - `DisplayText` turns tabs and line breaks into a space and strips escapes before the text is fitted, in every one-line slot (lists, card, search lines, editor lines); the sweeps now use every adversarial name.
   _Was:_ **Tabs in names** (`TestSessionListNameWithTabsFitsWidth`): lipgloss expands a tab
   to four spaces after the list delegate has truncated the title, so the row ends up
   wider than the list.
5. **FIXED (plan 5)** - 80x24 is the documented minimum (`views.MinTerminalWidth/Height`); below it the coordinator shows a "Terminal too small" notice in exactly the terminal's size and drops keys except quit (`app/size_guard.go`). `Dashboard.SetSize` never goes negative.
   _Was:_ **No behaviour below 80 columns** (`TestDashboardBelowMinimumWidth`). `SetSize`
   forces the cards pane to 46 columns and takes the difference from the preview,
   whose width goes negative below about 62 columns; from 62 to 79 the render grows a
   row. There is no "terminal too small" screen. The tests treat 80x24 as the smallest
   supported terminal, which is an assumption of this plan, not a documented
   requirement.
6. **FIXED (plan 5)** - `renderWithOverlay` places the modal in `renderedHeight()` (header + panes + footer rows, named constants), not `d.Height`.
   _Was:_ **Opening the delete modal shrinks the screen by four rows**
   (`TestDashboardDeleteModalFillsTheScreen`): `renderWithOverlay` places the modal in
   `d.Height`, which is the terminal rows minus `heightAdjustment`.
7. **FIXED (plan 5)** - `renderPaneTitle` keeps the label and clips the head of the query (`TruncateTail`) to the pane's inner width.
   _Was:_ **A long search query wraps the pane title** (`TestDashboardLongSearchQueryFits`):
   the `Search: ...` / `Filter: ...` line is not truncated, so the dashboard grows by
   two to four rows. The plan expected the search line to add a row; in fact it replaces
   a blank spacer, and the test asserts the size does not change. It starts early on
   a narrow terminal: at 80 columns a 9-character query already wraps in the sessions
   pane (the dashboard search test uses a 4-character no-match query for that reason).
8. **FIXED (plan 5)** - the form scrolls: it is a list of focus-tagged blocks, `scrollOffset` keeps the focused block inside the window, `↑ more` / `↓ more` mark cut edges and the help line is pinned.
   _Was:_ **The card editor overflows on small terminals** (`TestCardEditorFitsSmallTerminal`).
   The capability lists are floored at `minCapListHeight`, so below 34 rows with the
   default settings (Personas and MCPs both on) the form is taller than its box and the
   help line is cut off. At 24 rows the editor is 30 lines tall in a 20-line box.
9. **FIXED (plan 5)** - workspace/puppet lines, capability labels and the search header are clipped to their width, and the text inputs' width is derived from the editor's.
   _Was:_ **Card editor lines are not clipped** (`TestCardEditorLongLinesAreClipped`): the
   Workspace and Puppet lines and the capability list labels are written out as is, so
   a long name widens the line past the editor.
10. **FIXED (plan 5)** - `CardEditor.View` routes the picker through `renderWithOverlay`; its box is capped at the space given.
   _Was:_ **The workspace picker is not placed** (`TestWorkspacePickerModalHasTheEditorsSize`):
    `CardEditor.View` returns it bare, at two thirds of the width, top-left, while the
    other two modals are centred in the full size.
11. **FIXED (plan 5)** - every puppet is drawn in a fixed cell (the largest built-in's size, user art clipped with `ClipBlock`) and the row scrolls horizontally with `‹` / `›`.
   _Was:_ **The puppet selector's width comes from the art**
    (`TestPuppetSelectorModalClipsUserPuppets`): all puppets sit side by side in one
    row and the modal has no `SetSize`, so one wide user puppet makes it wider than the
    terminal and `Place` cannot pad it down. The built-in puppets fit.
12. **FIXED (plan 5)** - sizes are derived from the inside out in `SetSize`, the `> ` marker is its own column, and the text area takes the rows left after the fixed ones.
   _Was:_ **The AGENTS.md editor modal** (`TestAgentsMDEditorModalWrapsLongLines`,
    `TestAgentsMDEditorModalHonoursItsWidth`, `TestAgentsMDEditorModalFitsSmallTerminal`):
    the text area is 80% of the screen plus its marker, wider than the box's inner
    width. The box comes out 6 columns wider than declared, and any text row that fills
    the area wraps, doubling the modal's height. Separately the modal is 22 rows tall at
    the 80x24 minimum.

Also noted, not a bug: the session list is bounded rather than exact, and the
dashboard header and footer are not padded to the width, so "the whole render is an
exact rectangle" in step 4 holds only for the body (rows 2 to h+2). The tests assert
that and pin the total to `h + 4` lines.

## Outcome

Landed (and the one-line fix for #1): `internal/ui/uitest` (`AssertExactRect`, `AssertWithin`, `Region`,
`AdversarialNames`, `NamesWithoutTabs`, `UserPuppet`), `assertIntactCard` rebuilt on it,
and layout tests `session_list_layout_test.go`, `session_preview_layout_test.go`,
`dashboard_layout_test.go`, `card_editor_layout_test.go`, `modals_layout_test.go`. Each
has its contract in a header comment and lists how it was seen to fail. `components`
and `views` take about 5-6 s each.
