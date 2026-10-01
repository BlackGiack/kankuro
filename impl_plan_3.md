# Act 2 — Plan 3: MCP form indices and modal boilerplate

Covers `tech_debt_act_2.md` **§3c** (`mcp_form.go` indexes its inputs by literal
position) and **§4** (three modals duplicate `Open`/`Close`/`IsOpen`). Plan 3 of 4; see
the table in `impl_plan_1.md`.

**Why §4 is here and not left "opportunistic".** §4 says to do the modal clean-up when
one of those modals is next touched. Plan 2 just touched all three: it gave them layout
tests. That makes now the cheapest time, because the tests that would catch a rendering
regression from the refactor already exist. **Run plan 2 first.**

Both halves live in `internal/ui/views` (plus `components/modal.go`), so they share a
session without conflicting with anything else.

**Rules for this session:** don't commit unless asked. Keep all five gates clean. Make
sure each new test can fail.

---

## Part A — §3c: named indices for the MCP form

### Current state (`internal/ui/views/mcp_form.go`)

`mcpInputCount = 7`. The inputs are addressed as `inputs[0]` … `inputs[6]` in four
independent places that must agree:

1. `NewMCPForm`: placeholders (`0` Name, `1` Command, `2` Args, `3` Working directory,
   `4` URL, `5` Environment, `6` Headers).
2. `SetMCP` (`:95-101`): model → inputs.
3. `View` (`:159-165`): the label table.
4. `handleSubmit` (`:193-213`): inputs → model.

`Reset` and `cycleFocus` loop over the slice and use `inputs[0]` for the initial focus.
`MCPInputCount` is exported, and `mcp_form_test.go` (package `views_test`) checks
`len(f.Inputs())` against it.

### Step A1 — one field table, named constants

Copy the card editor's pattern (`nameInputIdx` … plus `focusTargetTable()`), simplified:
every MCP field is a text input, so there is no `notAnInput` case.

```go
const (
	mcpNameIdx = iota
	mcpCommandIdx
	mcpArgsIdx
	mcpCWDIdx
	mcpURLIdx
	mcpEnvIdx
	mcpHeadersIdx

	mcpInputCount // keep last
)

type mcpField struct {
	idx         int
	label       string // shown above the input in View
	placeholder string
}

// mcpFieldTable is the single source for the form's field order, labels and
// placeholders. NewMCPForm and View both read it; SetMCP and handleSubmit
// address inputs by the named constants above.
func mcpFieldTable() []mcpField { ... }
```

- `NewMCPForm` builds `inputs` by iterating the table (`inputs[f.idx] = textinput.New()`,
  placeholder, `SetWidth(defaultInputWidth)`), then `inputs[mcpNameIdx].Focus()`.
- `View` iterates the table for labels: `&f.inputs[field.idx]`.
- `SetMCP` and `handleSubmit` use the constants: `f.inputs[mcpCommandIdx].SetValue(mcp.Command)`
  and so on.
- `Reset` and `cycleFocus` focus `mcpNameIdx` instead of `0`.
- `MCPInputCount = mcpInputCount` stays exported. The `mcpInputCount` sentinel
  replaces the literal `7`.

Afterwards, `grep -n 'inputs\[[0-9]' internal/ui/views/mcp_form.go` must be empty. That
grep is how you check the job is done.

### Step A2 — tests

In a new internal test file, `views/mcp_form_internal_test.go` (package `views`):

1. **`TestMCPFieldTableMatchesConstants`:** `len(mcpFieldTable()) == mcpInputCount`, and
   `table[i].idx == i` for every `i` (no gaps, no duplicates, in order). This is the
   drift test §3c asks for, the same shape as `TestInputIndexConstantsMatchTargets`.

In `views/mcp_form_test.go` (package `views_test`):

2. **`TestMCPFormRoundTripsEveryField`:** this is the test that actually catches a
   swapped index; the drift test cannot. Build an `MCPServer` with a **distinct** value
   in every field (`Name: "n"`, `Command: "cmd"`, `Args: []string{"a1", "a,2"}`,
   `CWD: "/cwd"`, `URL: "http://u"`, `Env: {"E": "1"}`, `Headers: {"H": "2"}`). Call
   `SetMCP`, submit through `Update` with Enter, capture the `OnSave` argument, and
   assert it equals the input field by field. Preserve the ID.
   *Can-fail check:* swap `mcpCommandIdx` and `mcpURLIdx` in `SetMCP` only. The test
   must go red. Also run this test **before** A1 against the old literal code: it should
   pass there too, which proves it tests the behaviour rather than the refactor.
3. **`TestMCPFormLabelsMatchPlaceholders`** (optional, cheap): render `View()` and
   assert that each label appears before the corresponding input's placeholder. This
   catches a table row whose label and input index got out of step.

Update `project.md` §11: remove the "`mcp_form.go` addresses its inputs as
`f.inputs[0]` … `f.inputs[6]`" bullet.

---

## Part B — §4: an embeddable modal state

### Current state

All three modals in `internal/ui/views` carry the same three fields
(`open`, `confirmed`, `cancelled`) and the same five methods:

| Modal | `Open` | `Close` | Sets `confirmed`/`cancelled` in `Update` |
|---|---|---|---|
| `WorkspacePickerModal` (`:118-134`) | also calls `resetTransientState()` | plain | `:154`, `:166` |
| `AgentsMDEditorModal` (`:69-73`) | plain one-liner | plain | `:95`, `:99`, `:115` |
| `PuppetSelectorModal` (`:78-82`) | plain one-liner | plain | `:102`, `:104`; carries `// TODO use modal component` at `:13` |

`components.Modal` has `open` plus `Open`/`Close`/`IsOpen`, but no confirm/cancel. It
also does its own rendering (title, content string, legend), which the three modals
don't want: each renders its own list or textarea. No code outside each modal's own file
touches those fields (checked with grep).

### Step B1 — `components.ModalState`

Add it to `internal/ui/components/modal.go`:

```go
// ModalState is the open/confirmed/cancelled bookkeeping shared by every
// modal. Embed it (by value) in a modal type to get Open, Close, IsOpen,
// Confirmed and Cancelled; call Confirm/Cancel from the modal's Update. A
// modal that needs extra work on open defines its own Open and calls
// ModalState.Open from it.
type ModalState struct {
	open, confirmed, cancelled bool
}

func (s *ModalState) Open()           { s.open = true; s.confirmed = false; s.cancelled = false }
func (s *ModalState) Close()          { s.open = false; s.confirmed = false; s.cancelled = false }
func (s *ModalState) IsOpen() bool    { return s.open }
func (s *ModalState) Confirmed() bool { return s.confirmed }
func (s *ModalState) Cancelled() bool { return s.cancelled }
func (s *ModalState) Confirm()        { s.confirmed = true }
func (s *ModalState) Cancel()         { s.cancelled = true }
```

Check one thing before writing `Confirm`/`Cancel`: do any of the existing `Update`
paths also set `open = false` when confirming? If they do, keep that behaviour exactly.
Either `Confirm` closes too, or the call site keeps its explicit `Close()`. Don't change
when a modal closes as a side effect of this refactor.

Make `components.Modal` embed `ModalState` too, replacing its own `open` field and its
`Open`/`Close`/`IsOpen`. Its title/content/legend rendering stays. The dashboard's
delete modal doesn't use confirm/cancel, which is harmless.

### Step B2 — embed it in the three modals

For each modal: delete the three fields and the five methods, embed
`components.ModalState`, replace `m.confirmed = true` / `m.cancelled = true` with
`m.Confirm()` / `m.Cancel()`, and replace reads of `m.open` with `m.IsOpen()`.

- `WorkspacePickerModal` keeps its own `Open()`:
  ```go
  func (m *WorkspacePickerModal) Open() {
  	m.ModalState.Open()
  	// Reset transient state to ensure clean state on reopen
  	m.resetTransientState()
  }
  ```
- `PuppetSelectorModal`: delete the `// TODO use modal component`.
- Each modal's `View`, `SetSize` and content rendering stay as they are. §4 explicitly
  keeps each modal's own rendering.

The card editor (`card_editor.go:761-800`, `:837-845`) only calls `IsOpen`, `Open`,
`Confirmed` and `Cancelled`, which are all still promoted, so it needs no change. If it
compiles, the method set is intact.

### Step B3 — tests

1. `components/modal_test.go`: **`TestModalStateLifecycle`:** `Open` clears a previous
   confirm/cancel, `Close` clears everything, and `Confirm`/`Cancel` set only their own
   flag. A table-driven test is enough.
2. The existing modal tests (`workspace_picker_test.go`, `puppet_selector_test.go`,
   card editor tests that open modals) must pass unchanged. They're the behavioural
   guard.
3. Plan 2's modal layout tests must pass unchanged. They're the rendering guard.
4. **Reopen regression:** for each of the three modals, confirm → `Close` → `Open` →
   assert `Confirmed() == false`. That's the behaviour the hand-written `Open`s provided,
   and the one most easily lost if someone overrides `Open` and forgets to call
   `ModalState.Open()`. *Can-fail check:* make `WorkspacePickerModal.Open` skip
   `m.ModalState.Open()` and confirm the test goes red.

Update `tech_debt_act_2.md` §3c and §4 with a one-line "done in plan 3" status.

---

## Done when

- [ ] `grep -n 'inputs\[[0-9]' internal/ui/views/mcp_form.go` is empty.
- [ ] Drift test and round-trip test pass; the round-trip test goes red with a swapped index.
- [ ] `components.ModalState` exists, and `components.Modal` and the three modals embed it.
- [ ] No modal declares its own `open`/`confirmed`/`cancelled` field; the `TODO use modal component` is gone.
- [ ] Existing modal tests and plan 2's layout tests pass unchanged; reopen regression test added and seen red.
- [ ] `project.md` §11 and `tech_debt_act_2.md` updated.
- [ ] All five gates clean.
