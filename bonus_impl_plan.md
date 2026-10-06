# Act 2 — Bonus plan: things that move when the content changes

Two layout problems found while testing the app by hand after plan 5. Neither breaks the
shape tests from plans 2 and 5: every render still fits its box. The trouble is that
elements **move** when the content changes (paging, scrolling, searching), and no test
checks for that. One session, all in `internal/ui`.

**Rules for this session:** don't commit unless asked. All five gates
(`go build ./...`, `go vet ./...`, `go test ./...`, `go test -race ./internal/...`,
`golangci-lint run`) must be clean at the end. Every new test must be seen to fail on the
current code before the fix.

---

## The rule this plan adds

> **A component that is given a size keeps its elements in place.** Where an element sits
> (its row, and its column from the box edge) depends on the size the component was
> given, never on what it happens to show: which page, how far it scrolled, what the
> search matched.

Add this to `project.md` §6 as an extension of invariant 9 ("Rendered output keeps its
shape"). Shape tests check *that* a render fits; the tests below check *where things are*,
across content changes.

---

## Part A — workspace picker: list and legend shift sideways from page to page

`internal/ui/views/workspace_picker_modal.go`, `View`

### What happens (reproduced)

Picker at 120 wide, 40 workspaces with short paths on the first pages and long ones in
the middle, paged with →. Each column is the gap between the left border and that
element:

| page | title | first item | page dots | legend |
|---|---|---|---|---|
| 0–1 | 7 | 4 | 6 | 2 |
| 2 | 5 | 2 | 4 | 2 |
| 3 → 7 | 5 | 2 | 4 | 4, 5, 7, 8, 10 |

The list block moves on pages whose paths are short, and the legend moves on almost every
page.

### Cause

`View` builds the box content with
`lipgloss.JoinVertical(lipgloss.Center, m.list.View(), searchDisplay, "\n", help)`.

1. The Bubbles list pads its lines only to the **widest line on the current page**, not to
   the width it was given in `SetSize`. So the list block's width depends on the longest
   name or path being shown.
2. `JoinVertical(Center)` centres **every line** against the widest line of all the
   pieces:
   - When the list is wider than the legend (long paths), the legend is centred against
     the list's width, which changes from page to page. So the legend moves.
   - When the list is narrower than the legend (short paths), the list itself gets
     centred, so it moves right.

### Also: the legend moves down one row when a search starts

With the picker at 120×30, the legend is on row 25 with no search and row 26 while
searching (the box stays 30 rows). `searchDisplay` is `""` until a query is typed, and
then becomes `"\nSearch: …"`, which inserts rows between the list and the legend.

### Fix

1. **Give the list its full assigned width:**
   `lipgloss.NewStyle().Width(listWidth).Render(m.list.View())`, using the `listWidth`
   that `SetSize` already computes. Store it on the modal rather than recomputing it.
2. **Join left-aligned:** `lipgloss.JoinVertical(lipgloss.Left, …)`. Nothing should be
   positioned relative to its neighbours' content.
3. **The legend** keeps a fixed position against the box, not the content. Either
   left-align it with the list, or, if it should stay centred, centre it in the box's
   **fixed inner width**:
   `lipgloss.NewStyle().Width(m.modalWidth - pickerBoxChrome).Align(lipgloss.Center).Render(help)`.
   Choose one and say so in a comment. Note that the legend's text changes while
   searching ("Esc=back" becomes "Esc=clear search"). With the centred variant its start
   column then changes once. That's acceptable, but if a fully fixed legend is wanted,
   left-align it.
4. **Always reserve the search row:** render the search line even when there is no query,
   as a blank line of the same height, so the legend's row doesn't depend on whether a
   search is active. Account for it in `workspacePickerListHeightPadding` if the row count
   changes, and keep `TestWorkspacePickerModalIsAnExactBox` passing unchanged.

### Tests (`internal/ui/views/modals_layout_test.go`)

- **`TestWorkspacePickerKeepsColumnsAcrossPages`:** use the data set above (short paths
  on the first and last pages, long ones in the middle, enough for at least 6 pages). For
  each page, record the left offset (columns from the border) of the title line, the
  first item, the pagination line and the legend. Each must be the **same on every page**.
  The table above is the current failure, so confirm it fails before fixing.
- **`TestWorkspacePickerLegendRowIsStable`:** the legend is on the same row with no query,
  after typing one character, and after typing enough to filter everything out.
- Both run at a couple of sizes (80×24 and 120×30 terminals, through `CardEditor.View`,
  and on the modal directly).

---

## Part B — capability lists in the card editor change height while you scroll or search

`internal/ui/components/capability_list.go` (`View`, `SetSize`, `calculateWindow`);
`internal/ui/views/card_editor.go` (`SetSize`)

### What happens (reproduced)

A list of 20 items given a height of 7 (`SetSize(80, 7)`) renders:

| state | rows rendered |
|---|---|
| at the top | 7 |
| scrolled to the middle (a "..." above **and** below) | **8** |
| at the bottom | 7 |
| search matching 1 item | **2** |
| search matching nothing ("(no matching items)") | **2** |

In the card editor these lists are stacked (skills, personas, MCPs, then the help line),
so every height change moves everything below it: the next list's header, its items and
the legend. Scrolling into the middle of a list pushes the rest down a row. Typing a
search pulls everything up by however many items disappeared.

### Cause

`View` writes the header, then a `"  ..."` line **only if** something is cut off above,
then at most `maxVisible` items, then a `"  ..."` line **only if** something is cut off
below. Nothing pads the result to the given height. `SetSize` sets
`maxVisible = height - 2`, which budgets for one "..." row; when both appear, the list is
one row over. Plan 5 worked around that in `CardEditor.SetSize` with a slack row
("a capability list showing both of its '...' rows is one line taller than the height it
is given"). That keeps the editor within its box, but the list still changes height.

### Fix: a sized list renders exactly its height, in a fixed layout

When the list has a height (`c.Height > 0`), `View` returns **exactly `c.Height` lines**,
and the row each element sits on doesn't depend on the scroll position or the search:

- **Row 0:** the header (title, source hint, search query), as today.
- **The body** has `slots = c.Height - 1` rows. Decide the body layout from the
  **unfiltered** item count (`len(c.items)`), not the filtered one, so a search never
  changes the layout:
  - **The list fits** (`len(items) <= slots`): items start on row 1. Show the filtered
    items and pad with blank lines to `slots`.
  - **The list overflows** (`len(items) > slots`) and `slots >= 3`: the first and last
    body rows are reserved for the indicators. They show `"  ..."` when something is cut
    off on that side, and are blank otherwise. The rows between hold `slots - 2` items,
    padded with blank lines when the filtered set is smaller.
  - **Very small** (`slots < 3`, overflowing): no indicator rows. Show `slots` items. It
    still renders exactly `c.Height` lines.
  - **"(no matching items)"** takes the first item row; the rest is blank padding.
- **`calculateWindow`** works on the number of item rows in use (`slots - 2` or
  `slots`), not on a `maxVisible` computed in `SetSize`. `maxVisible` can go if
  `CapabilityListConfig.MaxVisible` only seeds it for an unsized list. Check its callers,
  and keep it for the unsized case if anything relies on that.
- **An unsized list** (`c.Height == 0`) keeps today's behaviour, so components that are
  built before `SetSize` runs still render.
- Keep the existing rules: labels clipped with `fit`, the focus prefix, the highlight. Pad
  with plain empty lines, not spaces. Width padding isn't part of this fix.

Then, in **`CardEditor.SetSize`**, remove the slack row (`headerHeight := 1` and its
comment) now that a list never renders taller than it is given. Check whether
`minCapListHeight = 2` still makes sense: with height 2 a list shows its header and one
item, and no indicators. Raise it to 3 (header plus one item row plus …) only if a test
shows a problem. The card editor's scrolling (`fieldBlocks` / `listBlocks` /
`syncScroll`) needs no change. Its blocks simply stop changing size.

### Tests

**`internal/ui/components/capability_list_test.go`** (or a new
`capability_list_layout_test.go`):

1. **`TestCapabilityListRendersExactlyItsHeight`:** heights 2 to 12, item counts 0, 1,
   `slots`, `slots + 1` and 30. For each, walk the cursor from top to bottom and also set
   the search queries `""`, one that matches 1, one that matches some, and one that matches
   none. Every render has exactly `Height` lines.
   *Fails today:* 8 rows in the middle of a 7-row list, and 2 rows for a search with 1 or
   0 matches.
2. **`TestCapabilityListItemsKeepTheirRows`:** for a fixed item set, the header is always
   on row 0, and the first item row (1 when the list fits, 2 when it overflows) doesn't
   change when the cursor moves or the query changes. When the list overflows, the "..."
   rows are always the first and last body rows, blank or not.

**`internal/ui/views/card_editor_layout_test.go`**:

3. **`TestCardEditorListsKeepTheirPlace`:** an 80×30 terminal (sized to
   `30 - views.CoordinatorRows`), all three sections on, 20 skills, 20 personas and 20
   MCPs. Record the rows of the persona list's header, the MCP list's header and the help
   line. Then:
   - move the skills cursor to the middle, then to the bottom;
   - type a search in the skills list that matches 1 item, then one that matches nothing;
   - Tab to the personas list and repeat.

   Those rows must not change, apart from what the editor's own scroll accounts for: pick
   a terminal height where the form fits without scrolling, so the scroll doesn't play a
   part.
   *Fails today:* the middle position pushes the persona header down a row, and the
   1-match search pulls it up.

Existing tests: the plan 5 card-editor tests (`FitsSmallTerminal`,
`FocusedFieldAlwaysVisible`, `ScrollIndicators`, the sweeps) must still pass. If one
encoded the old variable height, fix the expectation and say why in the session summary.

---

## Part C — close out

- `project.md` §6, invariant 9: add the rule from the top of this plan, and point at the
  two new kinds of tests (column offsets across pages, rows across scroll and search) as
  the pattern for any future list-like component.
- `project.md` §7 testing notes: one line saying that shape tests check *fit* and these
  check *position*, and that a new list component needs both.
- `tech_debt_act_2.md`: a short "found after Act 2 by hand" entry listing the two
  problems and that this plan fixed them.
- Manual check with the isolated config: open the workspace picker with paths of mixed
  lengths and page through it. In the card editor, scroll a long skill list and type a
  search. Nothing should move except the content.

## Done when

- [ ] The picker's title, items, pagination and legend keep the same distance from the border on every page; the legend keeps its row whether or not a search is active.
- [ ] A sized `CapabilityList` renders exactly `Height` lines in every state, and its rows don't move with the cursor or the search.
- [ ] The slack row in `CardEditor.SetSize` is gone.
- [ ] The five new tests were seen failing on the current code and pass after the fix; the existing layout tests still pass.
- [ ] `project.md` and `tech_debt_act_2.md` updated; manual check done.
- [ ] All five gates clean.
