# Act 2 — Plan 1: Session durability and locking

Covers `tech_debt_act_2.md` **§1** (session records not written atomically) and
**§3b** (`Manager.Kill` / `Relaunch` / `SwitchAgent` act on state read outside
the lock).

**Where it fits.** This is plan 1 of 4 for Act 2:

| Plan | Covers | Packages touched |
|---|---|---|
| **1 — this one** | §1, §3b | new `internal/fsutil`, `internal/config/storage.go`, `internal/session` |
| 2 | §5 rendering-shape tests | `internal/ui/components`, `internal/ui/views`, new `internal/ui/uitest` |
| 3 | §3c MCP form indices, §4 modal boilerplate | `internal/ui/views`, `internal/ui/components/modal.go` |
| 4 | §2 coordinator split, with §3a folded in | `internal/app` |

Run them in order. Plan 1 shares no files with plans 2–4, so if you need to go
faster, plan 1 is the one that can safely run beside another. Merge-tree it first
anyway (`project.md` §9).

**Rules for this session:** don't commit unless asked. All five gates
(`go build ./...`, `go vet ./...`, `go test ./...`, `go test -race ./internal/...`,
`golangci-lint run`) must be clean at the end. For every new test, remove the fix,
watch the test fail, then put the fix back.

---

## Part A — §1: atomic session writes

### Current state

- `config.FileRepository.WriteAtomic(path, data any)` (`internal/config/storage.go`)
  marshals YAML, then: `os.CreateTemp(dir, "."+base+".*.tmp")` → `Chmod(0o600)` →
  `Write` → `Sync` → `Close` → `Rename` → best-effort fsync of the directory, and
  removes the temp file on every error path. It is a method only for historical
  reasons: it uses no repository state.
- `session.saveSession(dir, *Data)` (`internal/session/storage.go`) does
  `MkdirAll(dir, 0o700)` → `json.MarshalIndent` → plain `os.WriteFile(path, data, 0o600)`.
- Every `saveSession` call runs under `m.mu` (`persistSessionLockedErr`), so a single
  manager never has two writers on one session file at once. A unique temp name still
  matters because a second `kankuro` process could share the config directory.
- `loadSessions` globs `*.json`. A temp file named `.<id>.json.<rand>.tmp` does not
  match, so a temp file left behind by a crash is ignored rather than parsed. Keep that
  naming.

### Step A1 — factor out a byte-level helper

Create `internal/fsutil/atomic.go`:

```go
// WriteFileAtomic writes data to path so that a crash at any point leaves
// either the old file or the complete new one under path, never a truncated
// one. ...
func WriteFileAtomic(path string, data []byte, perm os.FileMode) error
```

Move the body of `config.WriteAtomic` (everything after `yaml.Marshal`) into it
unchanged: same temp-name pattern, same `committed` defer cleanup, same error wrapping
(`"create temp %s: %w"` etc., with `filepath.Base(path)`), same best-effort directory
fsync. Move the doc comment along with it. Do not add a `MkdirAll`; creating the
directory is the caller's job, as it is today.

Why a new package rather than exporting from `config`: `session` already imports
`config`, but a file-I/O helper doesn't belong to the config domain. `fsutil` has no
dependencies, so anything can import it without creating an import cycle.

### Step A2 — make `config.WriteAtomic` a thin wrapper

```go
func (r *FileRepository) WriteAtomic(path string, data any) error {
	raw, err := yaml.Marshal(data)
	if err != nil {
		return fmt.Errorf("marshal %s: %w", filepath.Base(path), err)
	}
	return fsutil.WriteFileAtomic(path, raw, configFileMode)
}
```

Keep the method and its signature: about 12 call sites in `config` use it. The existing
`config/storage_test.go` tests (`TestWriteAtomicLeavesNoTempResidue`,
`TestWriteAtomicMode`, `TestWriteAtomicConcurrentWritesNeverSplice`) must pass
unchanged. That proves the refactor kept the behaviour.

### Step A3 — use it in `saveSession`

Replace the final `os.WriteFile` with `fsutil.WriteFileAtomic(path, data, sessionFileMode)`.
Add a `sessionFileMode = 0o600` constant next to the existing `0o700` (give that one a
name too, `sessionsDirMode`). Wrap the marshal error as `fmt.Errorf("marshal session
%s: %w", s.ID, err)`; it is currently returned bare.

Update `project.md` §5 ("Session JSON is written with a plain `os.WriteFile` — not
atomic yet") and §11 (first bullet) to say session records are now atomic.

### Step A4 — tests

New file `internal/fsutil/atomic_test.go`, which takes over the generic coverage:

1. **No residue:** after a successful write, the directory contains only the target
   file.
2. **Mode:** the target ends up with exactly the requested `perm`, also when it
   overwrites an existing file that has a different mode.
3. **Error path cleans up:** make the rename fail (target path is an existing
   *directory*). Expect an error and no `*.tmp` left in the directory.
4. **Concurrent writers never splice:** N goroutines each write a distinct payload of
   identical length to the same path. The final content must equal exactly one of the
   payloads. Copy the shape of `TestWriteAtomicConcurrentWritesNeverSplice`.

In `internal/session/storage_test.go`, add:

5. **`TestSaveSessionIsAtomic`:** launch and kill a session (as in the existing test),
   then assert that `sessions/` contains only `<id>.json` (no dot-files, no `.tmp`) and
   that the file mode is `0o600`.
6. **`TestLoadSessionsIgnoresLeftoverTempFiles`:** put a
   `.<id>.json.123.tmp` file containing garbage into `sessions/`. `NewManager` must
   return `LoadErrors() == nil`, and the valid session must still load. This pins down
   the naming decision above.

Can-fail check: with `saveSession` back on `os.WriteFile`, test 5 still passes,
because `os.WriteFile` also leaves no residue. That's expected: test 5 checks
behaviour that already holds, not the fix. The fix is guarded by test 4 in `fsutil`
plus the fact that `saveSession` calls `fsutil`. To prove `saveSession` is wired in,
make `WriteFileAtomic` temporarily return an error and confirm that
`TestSetNamePropagatesPersistFailure`'s mechanism, or a new assertion in test 5, sees
it fail.

Also check how `TestSetNamePropagatesPersistFailure` and
`TestLaunchWrapsErrPersistFailedButSessionStaysUsable` force a persist failure. If they
make `sessions/<id>.json` unwritable (for example `chmod 0o400` on the file), that no
longer fails: the rename replaces the file regardless of its mode. They need to make the
**directory** unwritable instead. Adjust them if so, and confirm both still fail
without their respective fixes.

---

## Part B — §3b: act on values captured under the lock

### Current state (`internal/session/lifecycle.go`)

- **`Kill`** takes `RLock`, looks up `s`, releases the lock, reads `s.Driver` with no
  lock held, calls `Kill()`, then takes `Lock` and sets `Terminated` and persists.
  Problems:
  1. A data race on `s.Driver`, which `activateSessionLocked` writes under `m.mu`.
  2. If the session was **relaunched** between the unlock and the final `Lock`, `Kill`
     marks the *new* activation `Terminated` even though its process is running.
  3. If the session was **removed** in that window, `Kill` calls
     `persistSessionLockedErr` on a detached `*Session` and **re-creates the record
     file `Remove` just deleted**. That one is a real resurrection bug, not only a
     race.
- **`Relaunch` / `SwitchAgent`** check `s.Status != StatusTerminated` after `RUnlock`,
  so they read `s.Status` unlocked. Even with the read moved under the lock, two
  concurrent calls would both pass the guard: `startSession` / the switch body spend a
  long time unlocked (workspace sync, `Spawn`) before `activateSessionLocked` flips the
  status. `SwitchAgent` also reads `s.WorkingDir` and `s.UseTmpWorkspace` unlocked.
- **`Remove`** checks `Status == Terminated` under the lock, but a relaunch in its
  unlocked phase is still `Terminated`. `Remove` can therefore delete the temporary
  workspace while `syncDriverFiles` is writing into it.

The `generation` counter (bumped in `activateSessionLocked`) is the tool the Act 2
document points to.

### Step B1 — reserve the session during activation

Add an unexported field to `Session`, next to `generation`:

```go
// activating is set, under m.mu, by Relaunch/SwitchAgent between passing the
// "must be terminated" guard and activateSessionLocked (or failure). It is what
// makes that guard exclusive: the unlocked middle of an activation (workspace
// sync, Spawn) is long, and Status stays Terminated throughout it.
activating bool
```

Write one helper that both entry points use:

```go
// reserveForActivation checks, under m.mu, that the session can be
// (re)activated and marks it as activating. It takes the lock itself; the
// caller must call releaseActivation on every failure path.
func (m *Manager) reserveForActivation(id string, notTerminatedErr string) (*Session, activationSnapshot, error)
```

It takes `m.mu.Lock()` (not `RLock`), looks the session up, fails if
`Status != Terminated || activating`, sets `activating = true`, and returns a snapshot
of the fields the caller needs unlocked: `WorkingDir`, `UseTmpWorkspace`. Keep the two
existing error messages ("session is already running", "session must be terminated
before switching agents") so the UI text doesn't change. For the `activating` case use
"session is already being started".

- `Relaunch`: reserve, then call `startSession(..., snapshot.WorkingDir)`. On a
  non-nil error that is **not** `ErrPersistFailed`, clear `activating` under the lock.
  `ErrPersistFailed` means the activation did happen.
- `SwitchAgent`: reserve, then use `snapshot.WorkingDir` / `snapshot.UseTmpWorkspace`
  instead of reading `s.*`. Clear `activating` on every early return (`finalWD == ""`,
  stat failure, `syncDriverFiles` error, `Spawn` error). Use a deferred
  `if !activated { m.releaseActivation(id) }` so no path gets missed.
- `activateSessionLocked` sets `s.activating = false`. It is the one place an
  activation completes.
- `startSession` with `isNew == true` (`Launch`) needs no reservation: the session isn't
  in the map until it's activated.

### Step B2 — `Remove` respects the reservation

In `Remove`, after the `Status` check, add
`if s.activating { return fmt.Errorf("session is being started: %s", id) }`.
Without it, B1 closes the double-start race but leaves remove-during-relaunch open.

### Step B3 — `Kill` captures under the lock and re-checks identity

```go
func (m *Manager) Kill(id string) error {
	m.mu.RLock()
	s, ok := m.sessions[id]
	var d driver.AgentDriver
	var gen uint64
	if ok {
		d, gen = s.Driver, s.generation
	}
	m.mu.RUnlock()
	if !ok { ... }

	var killErr error
	if d != nil {
		killErr = d.Kill()
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	// Only mark terminated if this is still the same activation of the same
	// session: a Remove in the window above must not have its record
	// re-created by our persist, and a Relaunch must not be marked dead.
	if cur, still := m.sessions[id]; !still || cur != s || s.generation != gen {
		return killErr
	}
	s.Status = StatusTerminated
	s.Pinged = false
	return errors.Join(killErr, m.persistSessionLockedErr(s))
}
```

Put the *why* comment at the re-check, as above; `project.md` §9 asks for comments on
non-obvious fixes. Keep the existing doc comment's promise ("still marked Terminated
and persisted" when the driver errors). It still holds when nothing intervened.

Calling `d.Kill()` with no lock held is intentional and stays that way: it can block on
process teardown, and `watch` needs `m.mu` in the meantime.

`Shutdown` calls `Kill` per ID and needs no change.

### Step B4 — update the docs

- `project.md` §11, third bullet ("`Manager.Kill` reads `s.Driver` after releasing the
  lock…"): remove it.
- `project.md` §10, "Change session lifecycle…" row: add "mind the generation counter
  **and the `activating` reservation**".

### Step B5 — tests (`internal/session/lifecycle_test.go`)

These need deterministic interleavings, not luck. Use `WithDriverRegistry` with a fake
driver whose `Spawn` and/or `Kill` block on a channel the test controls.
`TestKillPersistsTerminatedDespiteDriverError` already uses a custom driver; reuse its
pattern.

1. **`TestConcurrentRelaunchStartsOnce`:** a terminated session, and a fake driver
   whose `Spawn` blocks until released. Start `Relaunch` in goroutine A and wait until
   its `Spawn` is entered (signal channel). Call `Relaunch` from the test goroutine: it
   must return an error at once. Release A and assert it succeeds, that `Spawn` was
   called exactly once, and that the status is `Running`.
   *Fails without B1:* the second call passes the guard and spawns again.
2. **`TestSwitchAgentDuringRelaunchRejected`:** same setup, second call is
   `SwitchAgent`.
3. **`TestFailedRelaunchReleasesReservation`:** `Spawn` returns an error, then a second
   `Relaunch` (with a working driver) succeeds. Fails if a failure path forgets to
   clear `activating`.
4. **`TestRemoveDuringRelaunchRejected`:** `Remove` while A is blocked in `Spawn`
   returns an error, and the tmp workspace still exists. Fails without B2.
5. **`TestKillAfterRemoveDoesNotResurrectRecord`:** a running session whose fake
   driver's `Kill` blocks. Goroutine: `Kill(id)`. Once `Kill` is entered, simulate the
   process exiting (set the session terminated via the watcher path, or a fake status)
   and call `Remove(id)`. Release `Kill`. Assert that `sessions/<id>.json` does not
   exist. Fails without B3's identity check.
   If driving the watcher is awkward, put the session into `Terminated` with an
   internal test helper (`export_test.go` in package `session`) rather than with sleeps.
6. **`TestKillDuringRelaunchDoesNotTerminateNewActivation`:** the same idea for the
   generation check. Kill blocks, the session is relaunched in the window, Kill is
   released, and the status must still be `Running`.

All six must pass under `go test -race ./internal/session/`. Run that package with
`-count=20` once to shake out flakiness before declaring done.

---

## Done when

- [ ] `fsutil.WriteFileAtomic` exists; `config.WriteAtomic` and `saveSession` both use it.
- [ ] The existing `config` atomic tests pass unchanged; the new `fsutil` and `session` tests pass.
- [ ] Persist-failure tests still force a failure (directory-based, if they needed changing).
- [ ] `Kill`, `Relaunch`, `SwitchAgent` and `Remove` read session state only under `m.mu`.
- [ ] Tests 1–6 of B5 pass, and each fails with its fix removed.
- [ ] `project.md` §5, §10 and §11 updated.
- [ ] All five gates clean; `go test -race -count=20 ./internal/session/` clean.
