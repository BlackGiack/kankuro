# KANKURO — Technical Debt & Review Findings

Findings from a read-only review of `internal/session`, `internal/driver`,
`internal/config`, `internal/app`, and `internal/ui`. Prioritized by severity.

Confidence key: **[confirmed]** = verified by directly reading the code;
**[reported]** = surfaced by review, plausible, worth confirming (e.g. with
`go test -race`) before acting.

Findings are numbered in **discovery order, not severity order** — the plans below
reference these numbers, so new findings are appended rather than renumbered. Read
the section heading (Critical/High, Medium, Low) for severity.

> **Progress (2026-09-24).** All seven plans are implemented and landed on
> `feat/envvar` (plans 3 and 5 were rebased onto the plans that overtook them).
> `go build`, `go vet`, `go test ./...` and `go test -race ./internal/...` are all
> clean, and `golangci-lint run` now exits 0. The one item knowingly left behind is
> the modal-boilerplate half of Plan 7 step 5 — see the **Low** section.
>
> Earlier progress note (2026-09-22): Plans 1 and 2 are implemented and merged/staged:
> findings **#2, #3, #6, #15** and the exit-code half of **#9** are fixed (Plan 1),
> and **#9** (struct mapping) plus the `internal/session` half of **#8** are fixed
> (Plan 2). `go test -race ./internal/session/...` now passes — the Baseline
> section below records the pre-remediation state, so treat its `-race` line as
> historical. Line references inside findings touched by those two plans have
> shifted by a few lines; the symbol names are still accurate. Finding **#17** was
> discovered while testing Plan 2 and is not yet fixed.

> **Scope update (2026-09-10).** The env-var *injection feature* was cancelled;
> `session_envvar_plan.md` is not going to be implemented. The correctness fixes
> that plan had bundled ("Bundled correctness fixes" A–D, plus the `Env` half of
> the card-editor field-drop bug) are **real defects that stand on their own** and
> are now first-class findings here: #1 (card editor drops `Env`), #2
> (`mergeEnvWithPwd` append instead of upsert), #3 (relaunch clobbered by stale
> goroutines), #8 (swallowed session persist error), and the ordering half of the
> `formatKeyValueMap` bullet under **Low**. Nothing from that plan is dropped.
>
> Note the two that survive *without* any new UI: `IDCard.Env`
> (`internal/config/types.go:22`) already exists, is already validated, already
> flows into `driver.AgentConfig.Env` (`internal/session/lifecycle.go:202`) and is
> already applied to the child process (`internal/driver/driver.go:236`). A card
> with `env:` written by hand into `id_cards.yaml` is therefore live today — and
> both #1 and #2 silently break it.

## Baseline as of 2026-09-10

Measured before any remediation, so plans can tell regressions from pre-existing state:

- `go build ./...` — clean.
- `go test ./...` — **all packages pass**.
- `go test -race ./internal/session/...` — **FAILS** (`TestSessionPendingLineCapped`,
  data race `io.go:63` write vs `io_test.go:48` read). This is finding #6 and it
  blocks using `-race` as a verification gate for every other concurrency fix.
- `go test -race ./internal/app/...` — passes.
- A debug logger already exists: `cmd/kankuro/main.go` wires `tea.LogToFile` to
  `$XDG_CACHE_HOME/kankuro/kankuro.log` behind the `--debug` flag. Without that
  flag the standard `log` package writes to **stderr**, which corrupts the rendered
  TUI frame — so error *logging* must route through that file, never bare
  `log.Printf`. Relevant to #8.
- `golangci-lint run ./...` — **5 pre-existing issues** at the time this baseline
  was taken: 4 × `nolintlint` (unused `//nolint:gosec` directives at
  `session/lifecycle.go:331`, `session/workspace.go:32,37,43`) and 1 × `revive`
  (`driver` package name shadows a stdlib name, `driver/registry.go:1`).
  **Superseded as of Plan 7 — lint now exits 0:** the four stale directives were
  removed, and the `revive` collision is excluded in `.golangci.yml` with a
  comment saying the name is deliberate (`warn-unused: true` will flag the rule
  if it ever stops matching).

---

## Critical / High

### 1. Card editor drops `Env` and `SystemInstruction` on save — silent data loss [confirmed]
`internal/ui/views/card_editor.go:284` (`Card()`), `:202` (`SetCard()`)
`Card()` rebuilds `IDCard{}` from only the fields the form exposes. `SystemInstruction`
and `Env` are never loaded in `SetCard()` nor written in `Card()`.
`UpdateIDCard` (`internal/config/id_cards.go:53`) replaces the whole record, so
editing a card that has either field (e.g. hand-authored YAML) wipes it and
silently changes agent behavior — a dropped `Env` means the agent loses API keys /
proxy settings it was configured with. `CreatedAt`/`UpdatedAt` are correctly preserved.
**Fix:** make `Card()` overlay edits onto the card loaded in `SetCard()` so *no*
unmanaged field can ever be lost, rather than adding two more hand-copied fields.

### 2. `mergeEnvWithPwd` appends instead of upserting — card `Env` silently ignored [confirmed]
`internal/driver/driver.go:537-543`
Custom env is `append`-ed onto `os.Environ()`, producing duplicate `KEY=` entries.
`getenv(3)` returns the *first* match, i.e. the **inherited** value, so any card
`Env` key that collides with a variable already in the parent environment is
silently discarded. The `upsertEnv` helper that does this correctly already exists
two lines below and is already used for `PWD`.
**Fix:** route each `config.Env` entry through `upsertEnv`.

### 3. Relaunch is clobbered by the previous launch's goroutines [reported — verify with -race]
`internal/session/watch.go:9-38`, `internal/session/io.go:10-34`, `internal/session/lifecycle.go:240-247`
`watch` and `trackOutput` are spawned per launch but keyed **only by `sessionID`**,
and they look the session up fresh each iteration. Across a Kill→Relaunch of the
same ID, the *old* goroutines (holding the old driver and the old output channel)
can still wake up and (a) flip the freshly relaunched session to `Terminated`
(`watch.go:22-27`) and (b) close the new subscribers (`io.go:30-32`, on the old
channel's close). Both write through the mutex, so `-race` will not flag it — the
symptom is a session that shows as dead moments after a successful relaunch, and a
preview that goes silent. The window is exactly the relaunch path (`Kill` then
`Relaunch`), which is a normal user action (`Enter` on a terminated session).
**Fix:** give each activation a generation counter (field on `Session`, bumped in
`activateSessionLocked`), capture it in the goroutine closures, and make both
goroutines return when the generation no longer matches.

### 4. Concurrent map read/write can panic the app [reported — verify with -race]
`internal/app/coordinator.go:349-379` vs `:176`
`refreshData` returns a `tea.Cmd` that runs on a separate goroutine and reads
`c.views` (and calls `dashboard.PreviewSessionID()`), while the main `Update` loop
writes `c.views[c.currentView] = updated`. Concurrent map access is a hard runtime
panic, not a benign race, and the 2s tick keeps the window open continuously.
**Fix:** snapshot everything the closure needs on the main goroutine before
returning the `tea.Cmd`; do the preview read in `Update`, not in the async closure.

### 5. Ignored read error before a save destroys a whole source list [reported]
`internal/app/callbacks.go:357` (`saveStringSource`), `:405` (`deleteStringSource`),
`:262` (`SaveWorkspaceSource`); mirrored eight more times in
`internal/app/wiring.go:70-129` — but that whole file is dead code (see #16), so
only the `callbacks.go` sites are live.
Pattern `sources, _ := listFn()` then append/filter then save. A transient read
failure yields `nil`, and the subsequent atomic write durably persists a list
containing only the new entry — destroying every previously configured source.
**Fix:** propagate the list error and abort the save; never write a list derived
from a failed read.

### 6. Shared `*Session` pointers escape the lock — `-race` fails today [confirmed]
`internal/session/manager.go:147-161`
`GetSession`/`Sessions()` take `m.mu.RLock()` only to hand out the shared
`*Session`, whose fields the `trackOutput`/`watch` goroutines mutate under the
lock. Every read by the caller is then unsynchronized. `GetSession` has **no
production callers** (16 test call sites across `internal/session` and
`internal/app`), so today's exposure is test-only — but it makes
`go test -race ./internal/session/...` fail (see Baseline), which removes the race
detector as a gate for #3, #4 and any future concurrency work. `Sessions()` has no
callers at all.
**Fix:** return a value snapshot from `GetSession`, delete `Sessions()`, migrate
the test call sites.

### 17. A partial `DataLoadedMsg` broadcast wipes other views' data — silent card data loss [confirmed]
`internal/app/coordinator.go:415-421` (`broadcastData`), partial senders at `:284`
and `:434`; unguarded receivers at `internal/ui/views/agent_switch.go:65` and
`internal/ui/views/card_editor.go:252`

`broadcastData` sends one `DataLoadedMsg` to **every** `DataReceiver`, and two
senders deliberately fill in only the preview fields: `handlePreviewRequest`
(`:434`) and the incremental-output path (`:284`, `IsIncremental: true`). Receivers
that assign unconditionally therefore overwrite their real data with `nil`.

`Dashboard.SetData` (`dashboard.go:144-160`) already guards every field with
`if msg.X != nil` — so this was hit once and fixed *in one view only*.
`AgentSwitchView.SetData` and `CardEditor.SetData` still assign unconditionally.

The card editor makes this **silent data loss, not a cosmetic flash**:
`CapabilityList.SelectedIDs()` (`components/capability_list.go:76-84`) builds its
result by iterating the list's *items* and keeping those marked selected — so once
`SetData` has replaced the items with an empty slice, `SelectedIDs()` returns
nothing, and `CardEditor.Card()` saves the card with **no skills / personas / MCPs**.
Measured, driving a real coordinator:

```
skills on the card BEFORE:   [builtin:mcp-cli]
skills on the card AFTER:    []          <- after one incremental broadcast
=> saving now would write the card with NO skills
```

Reachability: opening the card editor does not clear `c.previewSessionID` (that is
finding #12), so a `SessionOutputMsg` already queued for the previously previewed
session still matches at `:263` and fires the `:284` broadcast while the editor is
open. The window closes at the next full refresh (≤2s, the tick), so the damage
needs a save inside that window — but an agent producing output makes the stray
message likely, and the loss is silent when it lands. Same root cause, milder
symptom, for the agent-switch view: its card list empties until the next tick.

**Fix:** the root cause is that `DataLoadedMsg` doubles as both a full snapshot and
a partial preview update with no way to tell them apart. Prefer fixing that — give
the preview updates their own message type (or have `broadcastData` only reach
receivers for the fields actually present) — rather than adding nil guards to each
of the five `SetData` implementations, which is what left the codebase in this
half-fixed state. If guards are used as a stopgap, mirror the dashboard's exactly
and add a test that a partial broadcast cannot change `CardEditor.Card()`.

---

## Medium

### 7. `WriteAtomic` is not crash- or concurrency-safe [reported]
`internal/config/storage.go:36-49`
(a) No `fsync` of the temp file before `os.Rename` — a crash can leave a partial
file after a "successful" rename on some filesystems. (b) Fixed `path + ".tmp"`
temp name — two concurrent writers clobber each other and can corrupt output.
(c) On `os.Rename` failure the `.tmp` file is leaked. Permissions (`0o600`/`0o700`)
are good. **Fix:** `os.CreateTemp` in the same dir (unique name), `f.Sync()` before
close, rename, and `os.Remove` the temp on any error path.

### 8. Pervasive swallowed errors hide corruption and state loss [confirmed]
`internal/app/coordinator.go:351-357` (`cards, _ := …` in `refreshData`),
`internal/session/lifecycle.go:284` (`_ = saveSession` in `persistSessionLocked`),
`:321` (`_ = deleteSession`), `:331` (`_ = os.RemoveAll` of the tmp workspace),
`internal/session/manager.go:125` (`loaded, _ := loadSessions`),
`internal/session/storage.go:31-37`
A single corrupt YAML makes `refreshData` render empty lists — and the next save
can then overwrite from the empty view. `persistSessionLocked` is fire-and-forget,
so *every* session state transition can fail invisibly (a session that shows as
running is simply gone after a restart); delete/tmp-cleanup failures leak disk; and
`loadSessions` silently `continue`s past corrupt JSON so a bad session file just
vanishes. **Fix:** give the persist path an error-returning variant used by
caller-facing mutations, surface at least the first error via `messages.ErrorMsg`,
and log the rest.

### 9. Struct triplication + discarded child exit status [confirmed]
`internal/session/manager.go:30-55` (`Session`), `:75-86` (`Snapshot`),
`internal/session/storage.go:10-21` (`Data`)
Three overlapping structs are copied field-by-field by hand in `persistSessionLocked`,
`List`, and `NewManager`. They already diverge (`OriginalDir` persisted but not in
`Snapshot`; `LastOutput`/`ExitCode` in neither). Relatedly, `watchProcess` does
`_ = cmd.Wait()` (`driver.go:273`) so `Session.ExitCode` (`manager.go:43`) is never
assigned — a crashed agent is indistinguishable from a clean exit. **Fix:** derive
`Snapshot`/`Data` via explicit mapping funcs next to the type (one place to update);
capture `cmd.Wait()`'s exit code and surface it.

### 10. `View()` side effect + synchronous blocking work on the UI goroutine [confirmed]
`internal/ui/views/dashboard.go:785`; callbacks at `:407-429`
`View()` calls `fmt.Print("\x1b[=0u")` every render — a side effect in a function
Bubble Tea may call arbitrarily often, fighting the renderer and making the view
untestable. Separately, `OnLaunchIDE` (spawns an editor) and `OnSwitch`
(`switchAgent` runs `SwitchAgent`, which spawns a process) execute synchronously
inside `Update`, freezing the whole TUI on any latency. **Fix:** emit the escape
sequence from a `tea.Cmd` / `Init` / teardown; make blocking callbacks return a
`tea.Cmd` that reports via a msg.

### 11. No repository-level serialization of concurrent writes [reported]
`internal/config/repository.go:97-99`
`FileRepository` holds no lock. The 2s refresh tick, `syncPuppetsAsync` at startup
(`coordinator.go:162-167`), and user-triggered saves all run as concurrent
`tea.Cmd` goroutines. Combined with the shared `.tmp` name (#7) this is a concrete
corruption path (e.g. startup puppet sync racing a user puppet-source save on
`puppets.yaml`). **Fix:** a `sync.Mutex` around read-modify-write sequences.

### 12. `waitForOutput` goroutine leak on a nil channel [reported]
`internal/app/coordinator.go:263-273`, `:325-331`, `:418`
`navigateTo` sets `previewSubChan = nil` but leaves `previewSessionID` unchanged; a
`SessionOutputMsg` already queued when the user navigates away re-arms
`waitForOutput` with a nil channel, and `<-ch` on nil blocks forever — one leaked
goroutine per occurrence. **Fix:** guard `if c.previewSubChan == nil { return }`, or
clear `previewSessionID` in `navigateTo`.

### 13. 2s polling re-reads and re-parses all config [reported]
`internal/app/coordinator.go:189-190,429-434`
The tick drives `refreshData`, which reads/unmarshals 6+ YAML files, sorts each, and
rebuilds built-in skills/puppets every 2 seconds regardless of change — continuous
disk/CPU churn on an idle TUI. **Fix:** refresh on demand after mutations and/or
cache with mtime checks; reserve the tick for live session output.

### 14. Dead `driver.SendInput` vs. unlocked raw PTY write on the attach path [confirmed]
`internal/driver/driver.go:61,493-502`; `internal/session/attach.go:26`;
`internal/app/attach.go:249,258`
`driver.SendInput` is part of the `AgentDriver` interface but has no callers (only a
test stub implements it). The real input path is `AttachHandle.Input`, i.e. the raw
`*os.File` PTY handed out at `attach.go:26` and written directly by
`app/attach.go:249,258` with no manager lock — it can race `Kill`/`watchProcess`
closing and nil-ing that PTY, and input is then silently lost.
**Fix:** route attach input through a locked `SendInput`, or delete the interface
method and document the raw path's ownership rules.

### 15. `manager.Kill` returns early on a benign kill error [reported]
`internal/session/lifecycle.go:295-305`, `internal/driver/driver.go:286-300`
If the process already exited, `os: process already finished` propagates and `Kill`
returns before setting `StatusTerminated`/persisting, leaving momentarily
inconsistent state and a spurious error to the caller. Also hit by `Shutdown`
(`lifecycle.go:340-348`). **Fix:** treat "already finished" as success.

### 16. The `commands` / `wiring.go` indirection layer is entirely dead [confirmed]
`internal/ui/commands/commands.go` (37), `internal/app/wiring.go` (139),
`internal/app/coordinator.go:40,53`
`WireCommands` builds a `commands.Commands` struct of ~20 closures
(`LaunchSession`, `SaveCard`, `SaveSkillSource`, `OpenIDE`, `RefreshData`, …),
`NewCoordinator` stores it (`coordinator.go:53`), and it is passed as the second
argument to **all nine** view constructors — where every view stores it in a
`cmds *commands.Commands` field that **no view ever reads**. Nothing invokes a
single one of those closures; the only reference is an assertion that one is
non-nil (`coordinator_test.go:79`). It is an abandoned first attempt at the
callback layer that `callbacks.go` now implements, kept alive by the constructor
signatures, and it costs 176 lines that duplicate live logic — including its own
copy of the #5 data-loss bug (eight times over) and of the `GetCard` not-found bug.
It also forces ~60 test call sites to pass `&commands.Commands{}`.
**Fix:** delete both files, drop the `cmds` field and constructor parameter from
the nine views, and drop `commands.Commands` from the ~60 test call sites. The
compiler finds every site. Listed as Medium rather than Low because its existence
makes the codebase read as if there are two competing command layers, and it is
where a reader looks first.

---

## Low

- **Comma-unsafe and unordered key/value handling** — `internal/ui/views/mcp_form.go:225`
  (`parseKeyValueList`), `:214` (`parseCommaList`), `:235` (`formatKeyValueMap`):
  (a) split on `,` with no escaping, so any env value / header / arg containing a
  comma is silently corrupted; (b) `formatKeyValueMap` iterates a Go map with no
  sort, so the rendered field reshuffles between renders (Go randomizes map order)
  — the MCP env field visibly reorders itself as you edit. Sort the keys, and choose
  an escaping scheme or a delimiter that cannot appear in values. [confirmed]
- **`renderAgentsMDRow` splits on the literal string `\n`** — `card_editor.go:854`
  uses a backtick raw string, so it never splits on real newlines; the multi-line
  AGENTS.md summary/ellipsis logic never fires as intended. Use `"\n"`. [confirmed]
- **Fragile parallel index systems in the card editor** — `card_editor.go:20-45,102-106,577-600,942-961`:
  `Focus*` iota, a hard-coded `FocusCount = 11`, duplicate `faceStyleIndex`/`useTmpIndex`,
  `inputs[]` indices, and a hand-listed `visibleFocusOrder()` must all stay in sync.
  Derive `FocusCount` from the iota and drive mapping/order from one table. [confirmed]
- **`DeleteIDCard` has no reference guard** — `internal/config/id_cards.go:91-102`
  deletes unconditionally, unlike `DeleteSkill`/`DeletePersona`/`DeleteMCP`; a card
  referenced by a live session's `CardID` can be deleted, breaking relaunch/switch
  ("card not found", `callbacks.go:144,160`). [confirmed]
- **`GetAppSettings` duplicates the settings schema inline** — `internal/config/repository.go:197-217`
  re-declares `AppSettings` as a local anonymous struct; a new field added to the
  real type won't load here (won't persist across restarts). Add a divergence test. [reported]
- **`wiring.go GetCard` returns `(IDCard{}, nil)` on not-found** — `internal/app/wiring.go:49-57`;
  callers can't distinguish missing from empty (inconsistent with `getCardByID`'s
  error). Both ignore the `ListIDCards` error. [confirmed]
- **Redundant config reads / O(n·m) scans per launch** — `callbacks.go:488-535`:
  each launch/relaunch/switch re-reads the full registry from disk and nested-loops
  over card IDs (~4 reads per launch). Index by ID once. [reported]
- **Byte-length truncation corrupts multibyte glyphs** — `components/card_list.go:417,424`,
  `session_preview.go:670,678`: `len(s)`/`s[:w-3]` mis-measure and can slice mid-rune;
  reuse the existing `runewidth`-based `truncateToVisibleWidth` (`session_preview.go:688`).
  ASCII-only input filters (`ch >= 33 && ch <= 126`) also drop non-ASCII typed characters
  (`card_editor.go:655`, `capability_list.go:173`, `dashboard.go:486`). [confirmed]
- **Leaky encapsulation of `CapabilityList.Selected`** — `components/capability_list.go:43`
  is a public map mutated directly by the card editor (`card_editor.go:313,734`),
  bypassing the component's own API; `IsSelectedList()` (:65) exists only to serve
  tests. Make the map private with `Select`/`Deselect` accessors. [confirmed]
- **`Card()` accessors mutate state** — `card_editor.go:298-299,967-981`:
  `selectedPersonaIDs()`/`selectedMCPIDs()` write back into receiver fields as a side
  effect of a read. Separate the sync step. [confirmed]
- **`os.ExpandEnv` blanks undefined vars in paths** — `internal/config/sync.go:17`:
  `$UNSET/skills` collapses to `/skills` silently. Consider `os.Expand` with a
  callback that errors on undefined vars. [confirmed]
- **Inconsistent workspace-source ID generation** — repository-side UUID in
  `workspaces.go:44-48` is dead because `callbacks.go:266` pre-assigns
  `fmt.Sprintf("workspace-%d", time.Now().UnixNano())` (also collision-prone on
  rapid adds). Use one scheme (repository-assigned UUID). [confirmed]
- **Synchronous disk I/O inside `Update`** — `coordinator.go:229` calls `DeleteMCP`
  directly in the render loop while most mutations defer into `tea.Cmd`s. [confirmed]
- **Dead code** — `cmds *commands.Commands` field is stored but never read across
  nine views (the tip of #16); `Modal.onConfirm`/`onCancel` + setters never invoked (`components/modal.go:63-71`);
  `Base.SetFocus` (`component/component.go:34`) unused; `SetSessionByID`
  (`session_preview.go:202`) called only from its own test; `prevStyle`
  (`puppet_selector_modal.go:21`) written at `:63` and never read; `keyCtrlE`
  (`views/constants.go:38`) unused while `card_editor.go:388` hard-codes `"ctrl+e"`,
  and `keyShiftTab` (`:37`) is used in `settings.go:152` while `dashboard.go:336`
  and `card_editor.go:349` hard-code `"shift+tab"`. Also 4 stale `//nolint:gosec` directives
  the linter already flags (see Baseline). [confirmed]
- **Duplicated modal boilerplate** — three modals re-implement
  `Open/Close/IsOpen/…` (`views/workspace_picker_modal.go`,
  `agentsmd_editor_modal.go`, `puppet_selector_modal.go`; the
  `// TODO use modal component` at `puppet_selector_modal.go:13` acknowledges it).
  Give `components.Modal` an embeddable form and embed it.
  **Still outstanding** — the *filter* half of this item was resolved in Plan 7
  (`components.Filter`, now used at all three former `filterCards` copies); the
  modal half was deliberately dropped, as Plan 7 step 5 permitted, and has no
  remaining plan that covers it. [reported, deferred out of Plan 7]
- **`config/types.go` has a standing `// TODO: Split this file`** — types for every
  domain live in one file (`internal/config/types.go:3`). Cosmetic; listed so it is
  not rediscovered as a finding. [confirmed]

---

## Noted as correct (checked, not problems)

Subscriber channel close is properly serialized under `m.mu` (no send-on-closed /
double-close); zombie reaping via `watchProcess`+`cmd.Wait()` with nil-guarded PTY
close is correct; command construction uses `exec.CommandContext` with explicit argv
(no shell — no injection hole); `CreatedAt`/`UpdatedAt` preservation on update is
correct; `attach.go` detach detection and PTY resize handling are careful
(FD range-checking, `EINTR` handling); `appendScrollback` bounds math is correct;
`upsertEnv` itself is correct (only its caller is wrong — #2).

---

## Remediation plans

The findings above are split into seven independently executable plans, ordered so
that each one's verification gate is available when it runs:

| Plan | File | Covers | Status |
| --- | --- | --- | --- |
| 1 | `debt_impl_plan_1.md` | #6, #3, #15, exit code half of #9, #2 | **done** |
| 2 | `debt_impl_plan_2.md` | #9 (structs), #8 (session half) | **done** |
| 3 | `debt_impl_plan_3.md` | #7, #11, #5, `DeleteIDCard` guard, `GetAppSettings`, ID gen, `ExpandEnv` | **done** |
| 4 | `debt_impl_plan_4.md` | #4, #12, **#17**, #10, #8 (app half), sync `DeleteMCP`, #14 | **done** |
| 5 | `debt_impl_plan_5.md` | #13, redundant per-launch reads, `wiring.go GetCard` | **done** |
| 6 | `debt_impl_plan_6.md` | #1, card-editor focus table, mutating accessors, `\n` split | **done** |
| 7 | `debt_impl_plan_7.md` | key/value parsing, unicode truncation, `CapabilityList`, **#16**, dead code, dedup | **done** (modal dedup deferred) |

Plan 1 must run first: it restores `go test -race ./internal/session/...` as a
usable gate. Plans 3–7 are otherwise independent of each other.
