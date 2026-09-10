# Debt Plan 3 — Config durability & list-write safety

> Covers **#7** (`WriteAtomic`), **#11** (no repository locking), **#5** (ignored
> read error before save), and the config-layer Low items: `DeleteIDCard`
> reference guard, `GetAppSettings` schema duplication, workspace-source ID
> generation, `os.ExpandEnv`.
>
> Independent of Plans 1, 2, 4-7. Can run at any time, in parallel with them if
> two sessions are available (it only touches `internal/config` plus ~60 lines of
> `internal/app/callbacks.go` and `wiring.go` that no other plan edits).

## Why these ship together

Every item here is one theme: **the config layer can lose the user's data.** #7 is
the write primitive, #11 is the missing serialization around it, #5 is the caller
that feeds it a truncated list. Fixing any one alone leaves the corruption path
open.

## Context budget

`internal/config/{storage.go (69), repository.go (~230), id_cards.go (~110),
workspaces.go (~120), sync.go (~?), types.go (~120)}`, plus
and `internal/app/callbacks.go:250-420` (skip `wiring.go` — it is dead, see #16).
Around 900 lines of reading, ~10 focused edits, several new tests. Fits one
session with room to spare.

---

## Step 1 — Make `WriteAtomic` actually atomic and durable (#7)

**Problem.** `internal/config/storage.go:36-49`:
```go
tmp := path + ".tmp"
os.WriteFile(tmp, raw, 0o600)
os.Rename(tmp, path)
```
Three defects. (a) No `fsync` before the rename — on a crash the rename can be
visible while the data blocks are not, leaving a truncated or zero-length config
after an apparently successful save. (b) The `.tmp` name is **fixed per path**, so
two concurrent writers to the same file interleave into one temp file and the
rename publishes a spliced document. (c) A failed `os.Rename` leaks the `.tmp`
file next to the real config forever. Permissions (`0o600` files / `0o700` dirs)
are already correct — keep them.

**Do this.**

```go
func (r *FileRepository) WriteAtomic(path string, data any) error {
    raw, err := yaml.Marshal(data)          // unchanged
    f, err := os.CreateTemp(filepath.Dir(path), "."+filepath.Base(path)+".*.tmp")
    // defer a cleanup that os.Remove(f.Name()) unless we committed
    f.Chmod(0o600)                          // CreateTemp gives 0o600 already; be explicit
    f.Write(raw)
    f.Sync()                                // (a)
    f.Close()
    os.Rename(f.Name(), path)               // atomic within the same dir
}
```
`os.CreateTemp` in the **same directory** is required — a rename across
filesystems is not atomic. Use a `committed bool` + `defer` so every early return
removes the temp file (c).

Optionally also `fsync` the parent directory after the rename (that is what makes
the *rename itself* durable). It is one extra `os.Open(dir)` + `Sync()`; take it if
it stays clean, it is the difference between "the file is not truncated" and "the
file is definitely there".

**Test.** `internal/config`: (a) writing leaves no `*.tmp` residue in the dir —
glob and assert empty; (b) N=20 concurrent `WriteAtomic` calls to the same path
from goroutines all succeed and the final file parses as valid YAML (this fails
reliably with the shared `.tmp` name); (c) the written file's mode is `0o600`.

## Step 2 — Serialize repository read-modify-write (#11)

**Problem.** `FileRepository` (`repository.go:97-99`) holds only `dir string` — no
lock. Meanwhile *everything* in `internal/app` runs as concurrent `tea.Cmd`
goroutines: the 2s tick's `refreshData` (`coordinator.go:189`), `syncPuppetsAsync`
at startup (`:162-167`), and every user-triggered save. The concrete corruption
path: startup puppet sync racing a user's puppet-source save, both doing
load→mutate→write on `puppets.yaml`. Step 1 stops them from producing a *spliced*
file; it does not stop the classic lost update (both read v1, both write, one wins).

**Do this.** Add `mu sync.Mutex` to `FileRepository` and take it around each
**read-modify-write sequence** — not around the individual `ReadYAML`/`WriteAtomic`
primitives, which would be both insufficient (the gap between them is the race) and
deadlock-prone (the sync functions call the accessors).

Concretely: lock at the top of the exported mutators — `CreateIDCard`,
`UpdateIDCard`, `DeleteIDCard`, `Save*Sources`, `Sync*FromSources`, `SaveAppSettings`,
`Delete*` — and leave the pure `List*`/`Get*` readers taking the mutex only if they
are cheap (they are; every one is a file read). Watch for **re-entrancy**:
`SaveWorkspaceSources` (`workspaces.go:44`) calls a sync at the end, and
`Sync*FromSources` calls `List*`. Either use a non-locking internal `xxxLocked`
helper for the inner call or restructure so the lock is taken exactly once per
public entry point. Go's `sync.Mutex` is **not** reentrant — a nested `Lock()`
deadlocks the whole TUI, so verify by running the full test suite with
`-race` *and* by exercising startup (`syncPuppetsAsync`) manually.

**Test.** A concurrency test in `internal/config`: 10 goroutines each appending a
distinct skill source via `SaveSkillSources(append(List(), x))`, then assert the
final list length. Note this test documents the *primitive*'s safety; the
lost-update pattern in the callers is step 3's job.

## Step 3 — Never write a list derived from a failed read (#5)

**Problem.** The pattern appears at `internal/app/callbacks.go:357`
(`saveStringSource`), `:405` (`deleteStringSource`), `:262`
(`SaveWorkspaceSource`) — and eight more times in `internal/app/wiring.go:70-129`,
which is **dead code** (finding #16: nothing ever invokes those closures; Plan 7
deletes the file). Fix only the live `callbacks.go` sites; do not spend time on
`wiring.go`. The live pattern:
```go
sources, _ := listFn()      // error discarded
sources = append(sources, path)
saveFn(sources)             // durably persists a 1-element list
```
A transient read error (or a single malformed YAML byte) yields `nil`, and the save
then **destroys every previously configured source** — permanently, atomically,
with a success message. This is the worst finding in the document by consequence.

**Do this.** In all four places, capture the error and abort:
```go
sources, err := listFn()
if err != nil {
    return messages.ErrorMsg{Err: fmt.Errorf("read sources: %w", err)}
}
```
These closures already return `tea.Msg` and already emit `messages.ErrorMsg` for
the save failure two lines down — so this is a two-line change per site with no new
plumbing. Do the same for `deleteSource`'s workspace branch (`callbacks.go:381`).

**Test.** `internal/app`: point the coordinator at a temp config dir, write
deliberately malformed YAML into `skills_sources.yaml`, invoke the save callback,
and assert (a) an `ErrorMsg` comes back and (b) **the file is unchanged** — the
second assertion is the one that matters.

## Step 4 — Config-layer Low items

Small, independent, each with a test. Do them in this order (cheapest first):

1. **`os.ExpandEnv` blanks undefined vars** — `internal/config/sync.go:17`.
   `$UNSET/skills` silently becomes `/skills`, and the source then either fails
   mysteriously or, worse, scans the filesystem root. Replace with `os.Expand` and
   a callback that records unknown names, then return
   `fmt.Errorf("undefined variable %q in path %q", name, value)`. `expandPath`
   already returns `(string, error)`, so no signature change.
2. **Inconsistent workspace-source IDs** — `internal/app/callbacks.go:266` assigns
   `fmt.Sprintf("workspace-%d", time.Now().UnixNano())` *before* calling
   `SaveWorkspaceSources`, which makes the repository's own UUID assignment
   (`workspaces.go:44-48`) dead code. Delete the caller's ID assignment and let the
   repository own it — one scheme, no collisions on rapid adds. Check the other
   source types for the same pattern while you are in the file.
3. **`GetAppSettings` re-declares the settings schema inline** —
   `internal/config/repository.go:197-217` unmarshals into a local anonymous
   struct listing `IDECommand`, `ShowPersonasInCardEditor`, `ShowMCPsInCardEditor`.
   Any field added to the real `AppSettings` type will marshal on save and then be
   **silently dropped on load** — i.e. it won't survive a restart, which is exactly
   the bug class a settings screen produces. Unmarshal into a wrapper over the real
   `AppSettings` type. The `*bool` pointers exist to distinguish "absent" from
   "false" for defaulting — preserve that behavior (a nested struct of `*bool`
   mirroring the real type, or `yaml.Node` + a defaults pass). Add a `reflect`-based
   divergence test asserting the loader covers every field of `AppSettings`.
4. **`DeleteIDCard` has no reference guard** — `internal/config/id_cards.go:91-102`
   deletes unconditionally, unlike `DeleteSkill`/`DeletePersona`/`DeleteMCP` which
   all check references. Deleting a card that a persisted session points at
   (`Session.CardID`) breaks that session's relaunch and switch with a bare
   "card not found" (`callbacks.go:144,160`). The complication: `internal/config`
   cannot see sessions (they live in `internal/session`, which imports `config` —
   reversing that is an import cycle). So put the guard in the **caller**:
   `OnDeleteCard` (`callbacks.go:59`ff) already has the session manager in hand via
   `c.session.List()`. Refuse the delete with a clear message naming the sessions
   that reference the card. Leave `DeleteIDCard` itself as the unguarded primitive,
   and say so in its doc comment.

---

## Verification

```
go build ./...
go test ./...
go test -race ./internal/config/... ./internal/app/...
golangci-lint run ./...
```

Manual smoke (`go run ./cmd/kankuro`): add a skills source and a workspace source
(`w`, `n`), quit, restart — both must still be listed. Then hand-corrupt
`~/.config/kankuro/skills_sources.yaml`, start KANKURO, try to add a source: you
must get an error and the file must still hold its original bytes (step 3). Change
a setting in the settings view (`c`), restart, confirm it stuck (step 4.3).

## Out of scope

- Unifying `internal/session/storage.go`'s plain `os.WriteFile` (`:42-52`) with
  `WriteAtomic`. Session JSON is written on every status tick, so making it
  fsync-per-write is a real I/O change that needs its own thought. Note it as
  follow-up; do not do it here.
- `refreshData`'s discarded read errors (`coordinator.go:351-357`) → **Plan 4**.
  Step 3 above deliberately covers only the *write* paths, where the consequence is
  destruction rather than a blank screen.
- Anything about the 2s tick's frequency → **Plan 5**.
