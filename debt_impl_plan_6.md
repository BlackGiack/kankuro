# Debt Plan 6 — Card editor: field preservation & index model

> Covers **#1** (the editor silently deletes `Env` and `SystemInstruction` on
> save) and the card-editor Low items: fragile parallel index systems, `Card()`
> accessors that mutate state, and `renderAgentsMDRow`'s literal-`\n` split.
>
> Independent of every other plan. Nothing else in the set edits
> `internal/ui/views/card_editor.go`.

## Why these ship together

#1's *correct* fix is structural — stop rebuilding the card from form fields — and
the index fragility is the same disease in the focus dimension. Both are "the
editor's model of a card is a hand-maintained parallel copy". One session, one
mental model.

## Context budget

`internal/ui/views/card_editor.go` (1007) read in full, plus
`card_editor_test.go` (310), `card_editor_setsize_test.go` (134),
`internal/config/types.go:11-30` and `internal/config/id_cards.go:53-88`.
About 1600 lines, one substantial refactor. One session, comfortably — but read the
whole editor before editing: the parallel-index problem means a change in one place
has non-obvious partners.

---

## Step 1 — Never lose a field the form doesn't render (#1)

**Problem.** `Card()` (`card_editor.go:284-301`) constructs a **fresh**
`config.IDCard{}` from the 5 text inputs + the lists + the AGENTS.md state.
`SetCard()` (`:202-229`) likewise loads only those. `IDCard` has 18 fields; two of
them — `SystemInstruction` and `Env` — are read by nobody in the editor and written
by nobody on save. `UpdateIDCard` (`internal/config/id_cards.go:53`) then replaces
the **whole record** (`UpsertInSlice` at `:76`).

Consequence: open any card that has `system_instruction:` or `env:` in
`id_cards.yaml`, press Enter, and both are gone — the agent silently changes
behavior (a lost `SystemInstruction`) or loses its credentials/proxy config (a lost
`Env`). `CreatedAt`/`UpdatedAt` escape this only because `UpdateIDCard` explicitly
re-copies them (`:65`).

Note `Env` is not hypothetical even though the env-injection feature was cancelled:
it already flows to `Spawn` (`session/lifecycle.go:202`, `:123`) and is already
applied to the child (`driver/driver.go:236`), so a hand-authored `env:` block works
today — until the editor deletes it. Plan 1 step 5 fixes the other half of that
path.

**Do this — overlay, don't rebuild.**

1. Store the loaded card: add `loaded config.IDCard` to the `CardEditor` struct
   (`:57-91`), assigned in `SetCard()` (`e.loaded = card`) and zeroed in `Reset()`
   (`:232`ff — check what else `Reset` must clear, it is the "new card" path).
2. Rewrite `Card()` to start from `e.loaded` and assign only what the form owns:
   ```go
   func (e *CardEditor) Card() config.IDCard {
       card := e.loaded              // carries ID, SystemInstruction, Env,
       card.Name = ...               // CreatedAt, and anything added later
       card.Driver = ...
       // ... the 13 fields the form actually edits
       return card
   }
   ```
   Keep `ID` coming from `e.editingID` (it is already tracked separately and
   `Reset` clears it) — but assert they agree, or drop `editingID` in favor of
   `e.loaded.ID`. Prefer dropping it: two sources of truth for the identity of the
   record being edited is the same bug one level down.
3. **Do not** add UI inputs for `SystemInstruction` or `Env` in this plan. The
   round-trip makes them safe; exposing them is a feature (and `SystemInstruction`
   is multi-line, so it needs the modal treatment `AgentsMDEditorModal` already
   demonstrates). Note it as a follow-up.

**Test — this is the important one.** Add a `reflect`-based guard to
`card_editor_test.go`:
```go
// Fails when a new IDCard field is added that the editor neither edits nor preserves.
func TestCardEditorPreservesUnmanagedFields(t *testing.T)
```
Build an `IDCard` with **every** field set to a distinctive non-zero value
(`reflect` over the struct to fill it, or write it out explicitly and assert the
field count so the test fails when a field is added), `SetCard` it, mutate one
input, call `Card()`, and assert every field except the mutated one is unchanged.
That test is what makes finding #1 un-reintroducible; the overlay alone is not.

## Step 2 — One table for focus, inputs, and order (Low)

**Problem.** Five parallel structures must be kept in sync by hand
(`card_editor.go:20-45`, `:102-106`, `:577-600`, `:942-961`):

- the `Focus*` iota (`:20-32`) — 11 entries;
- `FocusCount = 11` (`:35`), **hard-coded**;
- `faceStyleIndex = 3` / `useTmpIndex = 4` (`:39-40`), duplicating what
  `mapFocusToInput` (`:589-600`) already encodes;
- `inputCount = 5` (`:102`) with a positional `placeholders` slice (`:113-119`) and
  bare `e.inputs[0..4]` indexing scattered through `SetCard`/`Card`;
- `visibleFocusOrder()` (`:942-961`), a hand-listed slice that must contain every
  visible `Focus*` in display order.

`FocusCount` is only used for wrap-around; if someone adds a focus target and
forgets the constant, the last field becomes unreachable — no test catches it.

**Do this.** Drive everything from one table:
```go
type focusTarget struct {
    focus     int           // Focus* constant
    inputIdx  int           // index into inputs, or -1
    placeholder string      // "" for non-inputs
    visible   func(config.AppSettings) bool  // nil == always
}
var focusTargets = []focusTarget{...}
```
Then: `FocusCount = len(focusTargets)` (or keep the iota and add
`focusCount = iota` sentinel — the idiomatic Go trick: a final unexported
`focusCount` entry in the same `const` block), build `inputs` from the table's
placeholders, derive `mapFocusToInput` by lookup, and derive `visibleFocusOrder()`
by filtering on `visible`. Delete `faceStyleIndex`/`useTmpIndex`/`inputCount`.

Keep the exported `Focus*` and `FocusCount` names — `card_editor_test.go` and
`card_editor_setsize_test.go` reference them, and gratuitously renaming exported
identifiers is churn this plan does not need.

**Test.** Assert `FocusCount == len(focusTargets)`; assert
`visibleFocusOrder()` with personas+MCPs enabled equals the full table order and
with both disabled excludes exactly those two; assert tabbing forward from the last
visible target wraps to the first (the bug `FocusCount` drift would cause).

## Step 3 — Reads must not write (Low)

**Problem.** `Card()` calls `e.selectedPersonaIDs()` / `e.selectedMCPIDs()`
(`:298-299`), and those accessors **write back** into `e.personaIDs` / `e.mcpIDs`
(`:967-981`) as a side effect of being read. The write exists to preserve the IDs
of a section that Settings has hidden (`ShowPersonasInCardEditor` /
`ShowMCPsInCardEditor`), which is legitimate state-keeping in the wrong place — a
`View()`-time or test-time call to `Card()` silently mutates the editor.

**Do this.** Sync explicitly where selection actually changes — in the `Space`
toggle handler and in `SetCard` — so `e.personaIDs`/`e.mcpIDs` are always current;
then make the accessors pure reads. After step 1's overlay, the hidden-section case
gets simpler anyway: if a section is hidden, don't assign that field at all and let
`e.loaded`'s value carry through, which is exactly the desired semantics and removes
the need for the shadow slices. Prefer that.

## Step 4 — `renderAgentsMDRow` splits on the wrong thing (Low)

**Problem.** `card_editor.go:854`:
```go
lines := strings.Split(e.agentsMD.Content, `\n`)
```
A **backtick** raw string — that is a literal backslash-n, not a newline. So
`lines` is always a 1-element slice for real multi-line content, `len(lines) > 1`
never fires, and the "…" multi-line indicator never appears. Change the delimiter to
`"\n"`.

Careful: the branch below it also does byte slicing
(`contentSummary[:summaryLimit-ellipsisLen]`, `summaryLimit = 27`) which will
panic-free but mangle multibyte content — that class is Plan 7's
(unicode truncation). Fix the delimiter here; if the fix makes the truncation path
newly reachable for real multi-line content, route it through the
`runewidth`-based helper now rather than leaving a known mid-rune slice behind, and
note the crossover in the commit message.

**Test.** `card_editor_test.go`: content `"line one\nline two"` produces a summary
containing "line one" and an ellipsis, and does not contain "line two".

---

## Verification

```
go build ./...
go test ./...
go test ./internal/ui/...
golangci-lint run ./...
```

Manual smoke (`go run ./cmd/kankuro`) — step 1 is the whole point, so verify it by
hand:
1. Hand-add to `~/.config/kankuro/id_cards.yaml` on an existing card:
   ```yaml
   system_instruction: "you are a test"
   env:
     FOO: bar
   ```
2. Start KANKURO, `e` on that card, change its name, `Enter`.
3. `cat` the YAML: both `system_instruction` and `env` must still be there, with
   the new name. Before this plan, both are gone.
4. Tab through the editor with personas/MCPs disabled in Settings (`c`) and then
   enabled — every field must be reachable and the wrap-around must work (step 2).

## Out of scope

- Adding form inputs for `SystemInstruction` or `Env` (a feature, not debt).
- `CapabilityList.Selected` being a public map the editor mutates directly
  (`:313`, `:734`) → **Plan 7**, which owns `internal/ui/components`. Step 3 here
  touches only the editor's own shadow state; leave the component's API alone so the
  two plans don't collide.
- The unused `cmds *commands.Commands` field on this view → **Plan 7** (it is one
  entry in a five-view sweep).
