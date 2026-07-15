# KANKURO — Technical Debt & Review Findings

Findings from a read-only review of `internal/session`, `internal/driver`,
`internal/config`, `internal/app`, and `internal/ui`. Prioritized by severity.

Confidence key: **[confirmed]** = verified by directly reading the code;
**[reported]** = surfaced by review, plausible, worth confirming (e.g. with
`go test -race`) before acting.

> Findings that the env-vars feature directly depends on are **not** listed here —
> they are folded into `session_envvar_plan.md` ("Bundled correctness fixes"):
> `mergeEnvWithPwd` upsert, the Kill→Relaunch stale-goroutine race, the swallowed
> env persist error, `formatKeyValueMap` ordering, and the `Env` half of the
> card-editor field-drop bug.

---

## Critical / High

### 1. Card editor drops `SystemInstruction` on save — silent data loss [confirmed]
`internal/ui/views/card_editor.go:284` (`Card()`), `:202` (`SetCard()`)
`Card()` rebuilds `IDCard{}` from only the fields the form exposes; `SystemInstruction`
is never loaded or written (the `Env` half of this bug is fixed in the env plan).
`UpdateIDCard` (`internal/config/id_cards.go:53`) replaces the whole record, so
editing a card that has a system instruction (e.g. hand-authored YAML) wipes it and
silently changes agent behavior. `CreatedAt`/`UpdatedAt` are correctly preserved.
**Fix:** add a UI input (or hidden round-trip) for `SystemInstruction`; better,
make `Card()` overlay edits onto the loaded card so *no* unmanaged field can be lost.

### 2. Concurrent map read/write can panic the app [reported — verify with -race]
`internal/app/coordinator.go:349-379` vs `:176`
`refreshData` returns a `tea.Cmd` that runs on a separate goroutine and reads
`c.views` (and calls `dashboard.PreviewSessionID()`), while the main `Update` loop
writes `c.views[c.currentView] = updated`. Concurrent map access is a hard runtime
panic, not a benign race, and the 2s tick keeps the window open continuously.
**Fix:** snapshot everything the closure needs on the main goroutine before
returning the `tea.Cmd`; do the preview read in `Update`, not in the async closure.

### 3. Ignored read error before a save destroys a whole source list [reported]
`internal/app/callbacks.go:357` (`saveStringSource`), `:405` (`deleteStringSource`),
`:262` (`SaveWorkspaceSource`); mirror in `internal/app/wiring.go:70-129`
Pattern `sources, _ := listFn()` then append/filter then save. A transient read
failure yields `nil`, and the subsequent atomic write durably persists a list
containing only the new entry — destroying every previously configured source.
**Fix:** propagate the list error and abort the save; never write a list derived
from a failed read.

---

## Medium

### 4. `WriteAtomic` is not crash- or concurrency-safe [reported]
`internal/config/storage.go:36-49`
(a) No `fsync` of the temp file before `os.Rename` — a crash can leave a partial
file after a "successful" rename on some filesystems. (b) Fixed `path + ".tmp"`
temp name — two concurrent writers clobber each other and can corrupt output.
(c) On `os.Rename` failure the `.tmp` file is leaked. Permissions (`0o600`/`0o700`)
are good. **Fix:** `os.CreateTemp` in the same dir (unique name), `f.Sync()` before
close, rename, and `os.Remove` the temp on any error path.

### 5. No repository-level serialization of concurrent writes [reported]
`internal/config/repository.go:97-99`
`FileRepository` holds no lock. The 2s refresh tick, `syncPuppetsAsync` at startup
(`coordinator.go:162-167`), and user-triggered saves all run as concurrent
`tea.Cmd` goroutines. Combined with the shared `.tmp` name (#4) this is a concrete
corruption path (e.g. startup puppet sync racing a user puppet-source save on
`puppets.yaml`). **Fix:** a `sync.Mutex` around read-modify-write sequences.

### 6. Struct triplication + discarded child exit status [confirmed]
`internal/session/manager.go:30-55` (`Session`), `:75-86` (`Snapshot`),
`internal/session/storage.go:10-21` (`Data`)
Three overlapping structs are copied field-by-field by hand in `persistSessionLocked`,
`List`, and `NewManager`. They already diverge (`OriginalDir` persisted but not in
`Snapshot`; `LastOutput`/`ExitCode` in neither). Relatedly, `watchProcess` does
`_ = cmd.Wait()` (`driver.go:273`) so `Session.ExitCode` (`manager.go:43`) is never
assigned — a crashed agent is indistinguishable from a clean exit. **Fix:** embed a
shared core struct or derive `Snapshot`/`Data` via explicit mapping funcs next to
the type; capture `cmd.Wait()`'s exit code and surface it.

### 7. Pervasive swallowed errors hide corruption from the user [confirmed]
`internal/app/coordinator.go:351-357` (`cards, _ := …` in `refreshData`),
`internal/session/lifecycle.go:284,321,331`, `internal/session/manager.go:125`,
`internal/session/storage.go:31-37`
A single corrupt YAML makes `refreshData` render empty lists — and the next save can
then overwrite from the empty view. Session persist/delete/tmp-cleanup failures are
invisible (silent state loss, leaked disk), and `loadSessions` silently `continue`s
past corrupt JSON so a bad session file just vanishes. **Fix:** surface at least the
first error via `messages.ErrorMsg`; log the rest.

### 8. `View()` side effect + synchronous blocking work on the UI goroutine [confirmed]
`internal/ui/views/dashboard.go:785`; callbacks at `:407-429`
`View()` calls `fmt.Print("\x1b[=0u")` every render — a side effect in a function
Bubble Tea may call arbitrarily often, fighting the renderer and making the view
untestable. Separately, `OnLaunchIDE` (spawns an editor) and `OnSwitch`
(`switchAgent` runs `SwitchAgent`) execute synchronously inside `Update`, freezing
the whole TUI on any latency. **Fix:** emit the escape sequence from a `tea.Cmd` /
`Init` / teardown; make blocking callbacks return a `tea.Cmd` that reports via a msg.

### 9. `waitForOutput` goroutine leak on a nil channel [reported]
`internal/app/coordinator.go:263-273`, `:325-331`, `:418`
`navigateTo` sets `previewSubChan = nil` but leaves `previewSessionID` unchanged; a
`SessionOutputMsg` already queued when the user navigates away re-arms
`waitForOutput` with a nil channel, and `<-ch` on nil blocks forever — one leaked
goroutine per occurrence. **Fix:** guard `if c.previewSubChan == nil { return }`, or
clear `previewSessionID` in `navigateTo`.

### 10. 2s polling re-reads and re-parses all config [reported]
`internal/app/coordinator.go:189-190,429-434`
The tick drives `refreshData`, which reads/unmarshals 6+ YAML files, sorts each, and
rebuilds built-in skills/puppets every 2 seconds regardless of change — continuous
disk/CPU churn on an idle TUI. **Fix:** refresh on demand after mutations and/or
cache with mtime checks; reserve the tick for live session output.

### 11. Misleading / unsafe session API surface [confirmed]
`internal/driver/driver.go:61,493-502`; `internal/session/manager.go:147-161`
`driver.SendInput` is dead — the real input path is a raw, unlocked PTY write in
`internal/app/attach.go:249,258`, which can race `Kill`/`watchProcess` closing the
PTY (input silently lost). `GetSession`/`Sessions()` hand out shared `*Session`
pointers past the lock (latent data race; currently only tests read them, and
`Sessions()` has no callers). **Fix:** route attach input through a locked
`SendInput` or delete it; return value-copies/snapshots from `GetSession`; drop
`Sessions()`.

### 12. `manager.Kill` returns early on a benign kill error [reported]
`internal/session/lifecycle.go:295-305`, `internal/driver/driver.go:286-300`
If the process already exited, `os: process already finished` propagates and `Kill`
returns before setting `StatusTerminated`/persisting, leaving momentarily
inconsistent state and a spurious error to the caller. **Fix:** treat
"already finished" as success.

---

## Low

- **Comma-unsafe key/value parsing** — `internal/ui/views/mcp_form.go:225` (`parseKeyValueList`),
  `:214` (`parseCommaList`): split on `,` with no escaping, so any env value / header /
  arg containing a comma is silently corrupted. (Ordering half fixed in env plan.)
  Choose an escaping scheme or a delimiter that can't appear in values. [confirmed]
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
- **`GetSession` value-vs-pointer note** — see #11. [confirmed]
- **Byte-length truncation corrupts multibyte glyphs** — `components/card_list.go:417,424`,
  `session_preview.go:670,678`: `len(s)`/`s[:w-3]` mis-measure and can slice mid-rune;
  reuse the existing `runewidth`-based `truncateToVisibleWidth`. ASCII-only input
  filters (`ch >= 33 && ch <= 126`) also drop non-ASCII typed characters
  (`card_editor.go:655`, `capability_list.go:173`, `dashboard.go:486`). [confirmed]
- **Leaky encapsulation of `CapabilityList.Selected`** — `components/capability_list.go:43`
  is a public map mutated directly by the card editor (`card_editor.go:313,734`),
  bypassing the component's own API; `IsSelectedList()` (:65) is unused. Make it
  private with `Select`/`Deselect`. [confirmed]
- **`Card()` accessors mutate state** — `card_editor.go:298-299,967-981`:
  `selectedPersonaIDs()`/`selectedMCPIDs()` write back into receiver fields as a side
  effect of a read. Separate the sync step. [confirmed]
- **`os.ExpandEnv` blanks undefined vars in paths** — `internal/config/sync.go:17`:
  `$UNSET/skills` collapses to `/skills` silently. Consider `os.Expand` with a
  callback that errors on undefined vars. [confirmed]
- **Inconsistent workspace-source ID generation** — UUID in `workspaces.go:44-48` vs
  `fmt.Sprintf("workspace-%d", time.Now().UnixNano())` in `callbacks.go:266` (also
  collision-prone on rapid adds). Use one scheme (repository-assigned UUID). [confirmed]
- **Synchronous disk I/O inside `Update`** — `coordinator.go:229` calls `DeleteMCP`
  directly in the render loop while most mutations defer into `tea.Cmd`s. [confirmed]
- **Dead code** — `cmds *commands.Commands` field is stored but never read across
  five views; `Modal.onConfirm`/`onCancel` + setters never invoked (`components/modal.go:63-71`);
  `Base.SetFocus` (`component/component.go:34`), several `views/constants.go` constants,
  `SetSessionByID` (`session_preview.go:202`), `puppet_selector_modal.go:21 prevStyle`
  all unused; some code uses string literals (`"ctrl+e"`, `"shift+tab"`) instead of the
  defined `keyCtrlE`/`keyShiftTab` constants. [confirmed]
- **Duplicated filter/modal boilerplate** — `filterCards` implemented three times
  (`components/card_list.go:208`, `views/agent_switch.go:85`,
  `workspace_picker_modal.go:232`); four modals re-implement `Open/Close/IsOpen/…`
  (a `// TODO use modal component` at `puppet_selector_modal.go:13` acknowledges it).
  Extract a shared generic filter and a modal base. [reported]

---

## Noted as correct (checked, not problems)

Subscriber channel close is properly serialized under `m.mu` (no send-on-closed /
double-close); zombie reaping via `watchProcess`+`cmd.Wait()` with nil-guarded PTY
close is correct; command construction uses `exec.CommandContext` with explicit argv
(no shell — no injection hole); `CreatedAt`/`UpdatedAt` preservation on update is
correct; `attach.go` detach detection and PTY resize handling are careful
(FD range-checking, `EINTR` handling); `appendScrollback` bounds math is correct.
