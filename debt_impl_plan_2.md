# Debt Plan 2 — Session state modeling & persistence errors

> Covers finding **#9** (struct triplication) and the `internal/session` half of
> **#8** (swallowed errors) from `tech_debt_founding.md`.
>
> **Depends on Plan 1** (it changes `GetSession`'s signature and adds fields to
> `Snapshot`; doing this first would mean re-editing the same mapping code twice).
> Run after Plan 1 is merged.

## Context budget

`internal/session/manager.go` (240), `lifecycle.go` (349), `storage.go` (57),
plus `internal/app/callbacks.go` only where it consumes session errors
(~40 lines around `:38-56`). Small, mostly mechanical plan — the smallest of the
seven. Pair it with Plan 1 in one session if that session still has room.

---

## Step 1 — One place per field: derive `Snapshot` and `Data` from `Session` (#9)

**Problem.** Three overlapping structs — `Session` (`manager.go:30-55`),
`Snapshot` (`:75-86`), `Data` (`storage.go:10-21`) — are copied field-by-field by
hand in three separate places: `persistSessionLocked` (`lifecycle.go:270-285`),
`List` (`manager.go:178-196`), and the hydrate loop in `NewManager`
(`manager.go:128-140`). They have **already** drifted: `OriginalDir` is persisted
but absent from `Snapshot` (Plan 1 adds it); `LastOutput` and `ExitCode` are in
neither. Every future field is three edits, and forgetting one is silent data loss
that no test catches.

**Do this.** Do **not** introduce an embedded shared core struct — the three types
have genuinely different jobs (`Session` holds live handles: `Driver`, `Input`,
`Output`, `Done`, `Subscribers`; `Data` must stay a stable JSON schema; `Snapshot`
is a render-time value). Embedding would leak live handles into the JSON schema.
Instead, put explicit mapping functions next to the types and make them the *only*
copy sites:

```go
// manager.go, next to the type declarations
func (s *Session) snapshot() Snapshot   // caller holds m.mu
func (s *Session) data() *Data          // caller holds m.mu
func (d *Data) restore() *Session       // fresh Session, Status forced Terminated,
                                        // Subscribers initialized
```

Then:
- `List()` (`manager.go:178`) becomes a loop of `s.snapshot()`.
- `GetSession()` (Plan 1's new value-returning version) returns `s.snapshot()`.
- `persistSessionLocked` (`lifecycle.go:270`) becomes `saveSession(dir, s.data())`.
- The `NewManager` hydrate loop (`manager.go:128`) becomes `d.restore()`.

Add the fields that were missing while you are here: `LastOutput` and `ExitCode`
on both `Snapshot` and `Data` (`ExitCode` is populated by Plan 1 step 4;
persisting it is what makes "the agent crashed" survive a restart). Keep the JSON
tags `omitempty` so existing session files still load.

**Test.** Add `internal/session/mapping_test.go` with a **divergence guard** —
this is the point of the exercise:
```go
// Fails when a field is added to Session/Data without updating the mappers.
func TestSnapshotCoversPersistedFields(t *testing.T)
```
Use `reflect` to walk `Data`'s fields and assert each one has a same-named field
on both `Session` and `Snapshot` (whitelist the deliberate exclusions in one
explicit slice, with a comment saying why each is excluded). Plus a round-trip
test: build a `Session`, `data()` → `saveSession` → `loadSessions` → `restore()`,
assert every persisted field survives and `Status == StatusTerminated`.

## Step 2 — Make session persistence failures reportable (#8, session half)

**Problem.** `persistSessionLocked` ends in `_ = saveSession(...)`
(`lifecycle.go:284`). Every state transition — launch, relaunch, switch, kill,
rename, status flip in `watch` — can fail to reach disk with no signal anywhere.
The user sees a running session; after restarting KANKURO it is simply gone.
`Remove` likewise discards both `deleteSession` (`:321`) and the tmp-workspace
`os.RemoveAll` (`:331`), so a failed cleanup silently leaks a full workspace copy
under `~/.config/kankuro/tmp/`.

**Do this.** Split by whether a caller is there to hear about it:

1. Add `persistSessionLockedErr(s *Session) error` — the same body, returning the
   error. Keep `persistSessionLocked` as a thin wrapper that calls it and **logs**
   rather than discarding (see step 3 for the logging decision).
2. Convert the **user-initiated** mutations to the error-returning variant and
   propagate: `SetName` (`manager.go:195`), `Kill` (`lifecycle.go:295`),
   `startSession` (`:243`), `SwitchAgent` (`:145`). These all already return
   `error` to a caller that renders `messages.ErrorMsg`, so this is pure plumbing
   with no new UI. For `startSession`/`SwitchAgent`, a persist failure after a
   successful spawn must **not** unwind the spawn — the process is running; return
   the error so it surfaces, and leave the in-memory session intact.
3. Leave the **background** caller (`watch.go`, on the periodic status flip) using
   the logging wrapper. A 2-second ticker must not spam error banners.
4. In `Remove` (`lifecycle.go:311-334`): propagate the `deleteSession` error
   (it means the session will reappear on next start — the user must know); for the
   tmp-workspace `RemoveAll`, join it into the returned error rather than dropping
   it, but still complete the in-memory `delete(m.sessions, id)` so the UI does not
   wedge on an undeletable row.
5. `loadSessions` (`storage.go:23-40`) silently `continue`s past unreadable and
   unparseable files, and `NewManager` (`manager.go:125`) discards its error
   entirely. Change `loadSessions` to return the successfully loaded sessions
   **and** a joined error (`errors.Join`) describing the skipped files; have
   `NewManager` keep starting up (never block launch on one bad file) but retain
   that error so it can be reported. Since `NewManager` has no error return, store
   it on the `Manager` and expose `func (m *Manager) LoadErrors() error` for
   `internal/app` to surface once at startup.

**Test.** `internal/session`: (a) a `saveSession` failure propagates out of
`SetName` — make the sessions dir unwritable (`os.Chmod(dir, 0o500)`, skip if
running as root); (b) `loadSessions` over a directory with one valid and one
truncated JSON file returns the valid one *and* a non-nil error; (c) `Remove`
returns a non-nil error when the session file is already gone but still drops the
session from the map.

## Step 3 — Decide the logging story (needed by step 2, and by Plan 4)

A logger **already exists, conditionally**: `cmd/kankuro/main.go:35-42` wires
`tea.LogToFile(filepath.Join(logDir(), "kankuro.log"), "debug")` behind the
`--debug` flag. When that flag is set, the standard `log` package writes to that
file and `log.Printf` is safe. When it is **not** set, `log` writes to stderr —
straight into the rendered frame, corrupting the TUI. So: never call `log.Printf`
unconditionally.

Two workable options; pick one and write the choice down:

- **Recommended:** keep using the standard `log` package, but make `main.go`
  always redirect it — when `--debug` is off, `log.SetOutput(io.Discard)`. Two
  lines in `main.go`, and every package can then log freely without knowing about
  the TUI. This is the smallest change that makes the existing facility safe.
- Alternative: a package-level `func SetLogger(func(error))` hook on
  `internal/session` (default no-op), installed by `internal/app`. More
  ceremony, but it lets the dashboard surface a recent-errors panel later.

Either way, record the decision in a comment where it is declared — Plan 4 needs
the same facility for `refreshData`'s discarded errors and must not invent a
second one.

---

## Verification

```
go build ./...
go test ./...
go test -race ./internal/session/...
golangci-lint run ./...
```

The divergence test from step 1 is the real deliverable here — confirm it actually
fails when you add a scratch field to `Data` and forget the mapper, then remove the
scratch field.

Manual smoke: launch a session, quit KANKURO, restart it — the session must be
listed as Terminated with its `WorkingDir`, `Name`, and `CardID` intact (guards the
`restore()` rewrite).

## Out of scope

- `refreshData`'s `cards, _ := …` and the other `internal/app` swallowed errors →
  **Plan 4** (they need the coordinator's message plumbing, not the session's).
- `WriteAtomic` durability → **Plan 3**. Session files use plain
  `os.WriteFile` (`storage.go:42-52`), not `WriteAtomic`; unifying the two write
  paths is tempting but belongs with Plan 3's durability work. Note it there if you
  see it, do not do it here.
- Rendering `ExitCode` in the session list/preview. Persist it, don't display it.
