# Debt Plan 7 — Shared UI components: correctness, encapsulation, dead code

> Covers **#16** (the dead `commands`/`wiring.go` layer) and the remaining Low
> items: comma-unsafe **and unordered** key/value handling, byte-length truncation
> of multibyte glyphs + ASCII-only input filters, `CapabilityList.Selected`
> encapsulation, dead code (including the 4 stale `//nolint:gosec` directives from
> the lint baseline), and duplicated filter/modal boilerplate.
>
> Independent of Plans 1-5. **Run after Plan 6** — both touch
> `internal/ui/views/card_editor.go` (Plan 6 restructures it; this plan changes how
> it talks to `CapabilityList`), and Plan 6's refactor is the bigger of the two.

## Why this is last

Nothing here can corrupt user data or crash the app. It is the cleanup pass: after
it, the codebase stops teaching the wrong patterns to the next feature. Sequence it
last so it rebases onto the other six rather than the reverse.

## Context budget

`internal/ui/views/mcp_form.go` (241), `internal/ui/components/{capability_list.go
(281), card_list.go (429), session_preview.go (713), modal.go (113)}`,
`internal/ui/views/{agent_switch.go (216), workspace_picker_modal.go (266),
puppet_selector_modal.go (159), constants.go (39)}`, `internal/ui/component/
component.go (36)`, plus `internal/app/wiring.go` (139) and
`internal/ui/commands/commands.go` (37) for step 4, and their tests. That is a lot
of files — **the steps are fully independent, so drop step 5 (dedup) if context
runs short**, and consider doing step 4 in its own session (its ~60-call-site
mechanical edit is cheap in thought but not in tokens). Each step is separately
committable.

---

## Step 1 — Key/value round-tripping: sort it, then make it comma-safe

**Problem (a) — ordering.** `formatKeyValueMap` (`views/mcp_form.go:235-241`)
iterates a Go map and joins. Go **randomizes** map iteration order deliberately, so
the MCP env field renders in a different order on every call — the text visibly
reshuffles itself as the user edits the form, and no test on the output can be
stable.

**Problem (b) — escaping.** `parseKeyValueList` (`:225`) and `parseCommaList`
(`:214`) split on `,` with no escaping. Any env value, header, or arg containing a
comma is silently corrupted — and comma-bearing values are ordinary
(`Accept: a,b`, a JSON arg, a `PATH`-like list). `parseKeyValueList` also silently
**drops** any pair without `=` (`strings.Cut`'s `ok` is used only to gate the
insert), so a typo vanishes instead of erroring.

**Do this.**

1. Sort before joining — one line (`slices.Sorted(maps.Keys(m))`), fixes (a)
   immediately and unblocks a stable test.
2. For (b), choose a scheme and write the choice down in a comment. Options, in
   order of preference:
   - **Newline-delimited input** (one `KEY=value` per line) using the multi-line
     modal pattern that `AgentsMDEditorModal` (`views/agentsmd_editor_modal.go`)
     already establishes. Newlines cannot appear in an env value, so this needs no
     escaping at all — it is the only option that is actually correct, and the
     component already exists.
   - Backslash escaping (`\,` → `,`), if the single-line `textinput` must stay.
     More code, and users will not discover the rule.
3. Surface malformed pairs instead of dropping them: have the parser return
   `(map[string]string, error)` and let the form show the error it already has a
   channel for (`mcp_form.go`'s save path returns `tea.Cmd`/`ErrorMsg`).
4. `parseCommaList` (args/headers) has the same flaw. If step 2 moves env to
   newline-delimited, do the same for args, where commas are just as likely.

**Test.** `mcp_form_test.go` (98 lines, already exists): round-trip a map whose
values contain commas and `=`; assert `format(parse(s)) == s` for the canonical
form; assert output is sorted; assert a malformed entry produces an error.

## Step 2 — Measure and slice strings by runes, not bytes

**Problem.** Four functions do `len(s)` (byte length) as if it were display width
and then slice at a byte offset:
`components/card_list.go:417` (`truncatePill`), `:424` (`truncate`),
`components/session_preview.go:670` (`TruncatePath`), `:678` (`truncateString`).
Both mis-measure any non-ASCII string (a CJK card name, an accented workspace path,
an emoji in a puppet face) and can slice **mid-rune**, emitting replacement
characters and breaking the surrounding lipgloss layout math. The correct helper
already exists in the same package: `truncateToVisibleWidth`
(`session_preview.go:688`), built on `runewidth`.

**Do this.** Route all four through `truncateToVisibleWidth`. Note
`TruncatePath` truncates from the **left** (keeps the tail, prepends "…") — it needs
a right-anchored variant, so add `truncateToVisibleWidthLeft` next to the existing
helper rather than bending the call site. Then check for other `len(` -as-width uses
in the render paths while you are in these files.

**Related — ASCII-only input filters.** `card_editor.go:655`,
`components/capability_list.go:173`, `views/dashboard.go:486` filter typed runes
with `ch >= 33 && ch <= 126`, silently discarding every non-ASCII character — so a
user cannot type `é` or `日` into a search box or a card name. Replace with a
`unicode.IsPrint(ch) && !unicode.IsSpace(ch)`-style test (keep whatever
space-exclusion each site currently relies on — check each one, they differ) so
printable non-ASCII gets through while control characters stay out.

**Test.** `components/card_list_test.go` and
`components/session_preview_test.go` already exist: add cases with CJK and
combining-mark strings asserting the result is valid UTF-8 (`utf8.ValidString`) and
its `runewidth.StringWidth` is within the limit. For the filters, assert a typed
`é` reaches the search query.

## Step 3 — Close `CapabilityList`'s encapsulation hole

**Problem.** `CapabilityList.Selected` (`components/capability_list.go:43`) is an
exported `map[string]bool` that the card editor mutates directly:
`e.mcpList.Selected[...]` / `delete(e.mcpList.Selected, ...)`
(`card_editor.go:313`, `:734`), bypassing the component's own `SetSelected` /
`SelectedIDs` API. `IsSelectedList()` (`:65`) exists only to serve
`card_editor_test.go:205,211`. Any invariant the component wants to hold (cursor
consistency, selection limits) is unenforceable.

**Do this.** Rename the field to `selected` (unexported) and add the small API the
callers actually need: `Select(id string)`, `Deselect(id string)`,
`IsSelected(id string) bool`, `Toggle(id string)`. Update the two card-editor call
sites and migrate the two test call sites off `IsSelectedList()` to `IsSelected`,
then delete `IsSelectedList`. `SetSelected`/`SelectedIDs` stay as they are.

Search for other direct `.Selected` accesses before renaming — the compiler will
find them all, which is the point of doing it this way.

## Step 4 — Delete the dead `commands` / `wiring.go` layer (#16)

**Problem.** `WireCommands` (`internal/app/wiring.go:9-18`) builds a
`commands.Commands` struct of ~20 closures — `LaunchSession`, `AttachSession`,
`SaveCard`, `GetCard`, `SaveSkillSource`, `DeleteWorkspaceSource`, `OpenIDE`,
`RefreshData`, … — and `NewCoordinator` stores it (`coordinator.go:53`) and passes
it as the second argument to **all nine** view constructors
(`coordinator.go:129-150`). Every view stores it in a `cmds *commands.Commands`
field (`dashboard.go:81`, `card_editor.go:92`, `settings.go:53`,
`source_manager.go:37`, `mcp_form.go:37`, `session_launcher.go:30`) and **not one
of them ever reads it**. No closure in the struct is ever invoked. The only
reference anywhere is `coordinator_test.go:79` asserting one field is non-nil.

It is an abandoned first draft of the callback layer that `callbacks.go` now
implements — 176 lines duplicating live logic, including its own copy of the #5
list-destroying bug eight times over and its own buggy `GetCard`. It also drags
`&commands.Commands{}` through ~60 test call sites.

**Do this.** Delete `internal/ui/commands/` and `internal/app/wiring.go` entirely;
drop the `cmds` field and the constructor parameter from the nine views; drop the
`cmds` field from `Coordinator` (`coordinator.go:40,53`); delete
`TestCoordinator_WireCommands` (`coordinator_test.go:69-81`). Then fix the ~60 test
call sites — mechanical, e.g.
`views.NewDashboard(cb, &commands.Commands{})` → `views.NewDashboard(cb)` — and
drop the now-unused `commands` import from each test file. The compiler enumerates
every site for you; `go build ./... && go test ./...` is the completeness check.

Before deleting, do one grep for `cmds\.` outside `wiring.go`/`commands.go` to
confirm nothing started using it since this plan was written.

Commit this step **on its own** — it is a large, purely mechanical diff and mixing
it with the semantic changes in the other steps makes both unreviewable.

## Step 4b — Delete the smaller dead code

Each of these is verified unused; remove them and let the compiler confirm:

- `Modal.onConfirm` / `onCancel` + `SetConfirm`/`SetCancel`
  (`components/modal.go:63-71`) — the setters are never invoked and the fields are
  never called. Either wire them in step 5's modal base or delete them; do not
  leave them.
- `Base.SetFocus` (`component/component.go:34`) — unused.
- `SetSessionByID` (`components/session_preview.go:202`) — called only from its own
  test (`session_preview_test.go:367`). Delete both, unless the preview genuinely
  needs an ID-only setter, in which case give it a caller.
- `prevStyle` (`views/puppet_selector_modal.go:21`) — written at `:63`, never read.
- Unused constants in `views/constants.go`.
- `keyCtrlE` (`views/constants.go:38`) is unused while `card_editor.go:388`
  hard-codes `"ctrl+e"`; `keyShiftTab` (`:37`) is used in `settings.go:152` while
  `dashboard.go:336` and `card_editor.go:349` hard-code `"shift+tab"`. Use the
  constants at all four sites — they exist precisely so these do not drift.
- The 4 stale `//nolint:gosec` directives the linter already flags:
  `session/lifecycle.go:331`, `session/workspace.go:32,37,43`. They are the
  `nolintlint` half of the lint baseline; removing them takes the baseline from 5
  issues to 1 (the `revive` `driver`-package-name warning, which is a deliberate
  naming choice — leave it, or suppress it in `.golangci.yml` with a comment).

## Step 5 — Deduplicate filters and modals (drop this step if context is short)

**Problem.** `filterCards` is implemented three times with the same body
(`components/card_list.go:208`, `views/agent_switch.go:85`,
`views/workspace_picker_modal.go:232`), and four modals each re-implement
`Open`/`Close`/`IsOpen`/`SetSize`/overlay rendering. `puppet_selector_modal.go:13`
carries a `// TODO use modal component` admitting it.

**Do this.** Extract one generic filter into `internal/ui/components`:
```go
func Filter[T any](items []T, query string, label func(T) string) []T
```
and have all three call it. For the modals, give `components.Modal` (which already
has open/close/size state) an embeddable form and embed it in the four
hand-rolled modals, deleting their duplicated state. Keep each modal's *content*
rendering as-is — only the chrome is shared.

Do this **last** and commit it separately: it is the highest-churn, lowest-risk
change in the set, and it is the one most likely to conflict if any other plan is
still in flight.

---

## Verification

```
go build ./...
go test ./...
go test ./internal/ui/...
golangci-lint run ./...    # expect 1 issue (revive/driver) instead of the baseline 5
```

Manual smoke (`go run ./cmd/kankuro`):
1. Create an MCP (`m`) with an env value containing a comma and a header containing
   a comma; save, reopen the form — both must come back intact and in a stable
   order (step 1).
2. Name a card with CJK characters or an emoji and check the dashboard grid and the
   session preview render without mojibake or broken borders (step 2).
3. Type a non-ASCII character into a capability search box — it must appear
   (step 2).
4. Toggle skills/MCPs in the card editor and save — selection must round-trip
   (step 3).

## Out of scope

- `internal/config/types.go`'s standing `// TODO: Split this file` — cosmetic,
  touches every package's imports, and is not worth a plan. Leave it.
- Any change to what the UI *shows* (new fields, new keys, new views). This plan is
  strictly behavior-preserving except where the current behavior is the bug
  (comma corruption, dropped non-ASCII input, mid-rune slicing).
