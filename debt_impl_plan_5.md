# Debt Plan 5 — Refresh cadence & per-launch config reads

> Covers **#13** (2s polling re-reads and re-parses all config) and the Low items
> "redundant config reads / O(n·m) scans per launch" and
> "`wiring.go GetCard` returns `(IDCard{}, nil)` on not-found".
>
> **Depends on Plan 4.** Plan 4 restructures `refreshData` and its call sites for
> safety; this plan changes *when and how often* it runs. Doing them in the other
> order guarantees a merge conflict in the same function.

## Why this is its own plan

It is the only plan in the set that changes **behavior by design** rather than
fixing a defect: after it, the UI no longer reflects out-of-band file edits within
2 seconds. That is a product decision worth isolating so it can be reverted or
tuned without dragging correctness fixes with it. It also needs measurement before
and after, which the other plans do not.

## Context budget

`internal/app/coordinator.go` (~440), `callbacks.go:480-540` and its call sites,
plus a read of `internal/config/{skills.go,
puppets.go, repository.go}` to see what each `List*` actually costs. Small reading
load, but the design decision in step 1 deserves the saved context — do not bundle
extra work into this session.

---

## Step 0 — Measure first

Do not optimize blind. Before changing anything, add a temporary counter (or run
`strace -c -f -e trace=openat,read` against the binary, or just log per-refresh
timing behind an env var) and record:

- how many `openat` calls one `refreshData` makes (expect ~7 files: id_cards,
  skills, personas, mcps, workspaces, puppets, plus session JSONs);
- the wall time of one `refreshData` on a warm cache with a realistic registry;
- the same with `~50` cards / `~50` skills, to see whether the sorting and
  `MergeBuiltinSkills` rebuild (`internal/config/skills.go:17`) dominate.

Write the numbers into the commit message. If a refresh is genuinely sub-millisecond
on your machine, say so and **scope this plan down to steps 2 and 3** — 7 warm-cache
file reads every 2s is unpleasant but it is not a user-visible problem, and the
cache in step 1 has a real correctness cost (stale UI) that must be justified.

## Step 1 — Decouple the live-output tick from the config refresh (#13)

**Problem.** One 2-second tick (`coordinator.go:189-190`, `:429-434`) drives
`refreshData`, which unconditionally reads and unmarshals every registry YAML,
sorts each result, and re-merges the built-in skills
(`config/skills.go:17`) and puppets — whether or not anything changed. On an idle
TUI that is continuous disk and CPU churn forever.

Note what the tick is *actually* needed for: session status/`Pinged` transitions
and the preview scrollback. Those come from the session manager (in memory) and
are cheap. The config half is what is expensive, and config only changes when the
user or a sync changes it.

**Do this — two independent moves, take them in order:**

1. **Refresh config on demand.** Every mutation path already routes through the
   coordinator (`CardSavedMsg`, source saves, MCP delete, settings save, the sync
   functions), so each one can request a config refresh explicitly. Plan 4 step 1
   introduces `messages.DataRefreshRequestedMsg` for exactly this — reuse it.
   Split `refreshData` into `refreshConfig()` (the YAML reads) and
   `refreshSessions()` (the manager snapshot + scrollback). The tick then calls
   only `refreshSessions()`; `refreshConfig()` runs at `Init` and after mutations.
2. **Keep out-of-band edits working, cheaply.** Users hand-edit
   `~/.config/kankuro/*.yaml` (the specs treat the registries as user-editable
   files, and `kankuro_specs.md` §5 documents them as YAML on disk), so dropping
   polling entirely is a regression. Add an mtime check: one `os.Stat` per registry
   file (~7 stats, no parsing, no allocation) on the tick, and re-read only files
   whose mtime or size changed. Put the stat-and-compare in `internal/config`
   (e.g. `func (r *FileRepository) Fingerprint() (map[string]FileStamp, error)`),
   not in the coordinator — the coordinator should not know the filenames.
   Do **not** reach for `fsnotify`: it is a new dependency and a new goroutine for
   what 7 stats already solve.

Keep the tick interval at 2s. If step 0 showed the stats themselves are material
(they will not be), tune it then, not now.

**Test.** `internal/app`: (a) two consecutive ticks with no file change perform
exactly one parse — inject a counting `config` fake or wrap the repo interface;
(b) touching a registry file (write new content, bump mtime) causes the next tick
to pick up the change; (c) a mutation callback triggers an immediate refresh
without waiting for a tick.

**Watch out:** mtime granularity is 1s on some filesystems, so a write within the
same second as the last stat can be missed. Compare **size and mtime** together and
accept the residual (the mutation path in move 1 covers every in-app change; only
hand-edits rely on the stat, and a 2s-late pickup of a hand-edit is fine). Do not
try to solve this with content hashing — that reintroduces the reads you removed.

## Step 2 — Index registries once per launch instead of re-reading them (Low)

**Problem.** `internal/app/callbacks.go:488-535`: `getCardByID`, `skillsForCard`,
`personasForCard`, `mcpsForCard` each call a `List*` (a full file read + parse +
sort) and then nested-loop the result against the card's ID slice. A single
launch/relaunch/switch does **~4 full registry reads** plus O(n·m) scans, and
`relaunchSession` (`:131-157`) calls three of them back to back.

**Do this.** Add one resolver that reads each registry once and builds
`map[string]T` indexes:
```go
type cardBundle struct {
    Card     config.IDCard
    Skills   []config.Skill
    Personas []config.Persona
    MCPs     []config.MCPServer
}
func (c *Coordinator) resolveCard(cardID string) (cardBundle, error)
```
Use it from `launchCard` (`callbacks.go:~110`), `relaunchSession` (`:131`), and
`switchAgent` (`:157`). Preserve the current **ordering semantics** — the existing
loops iterate the registry outer / card IDs inner, so results come out in registry
order (name-sorted), not card-ID order. A map-based rewrite that iterates
`card.SkillIDs` instead will silently change the order in which skills are written
to disk and listed in AGENTS.md. Keep registry order: iterate the registry, test
membership against a `map[string]struct{}` built from the card's IDs.

Also propagate the discarded `ListSkills`/`ListPersonas`/`ListMCPs` errors here —
today a failed read means the agent launches with **no skills** and no warning,
which is a silent behavior change mid-session.

**Test.** `internal/app/integration_test.go` already exercises launch/relaunch;
extend it to assert the resolved skill order matches registry order, and add a case
where a registry read fails and the launch is aborted with an error rather than
proceeding capability-less.

## Step 3 — `getCardByID`'s discarded read error (Low)

**Problem, revised.** The debt doc lists `wiring.go GetCard`
(`internal/app/wiring.go:49-57`) returning `(config.IDCard{}, nil)` on not-found.
That function is **dead code** — finding #16: nothing ever calls it, and Plan 7
deletes the whole file. Do not fix it here.

What *is* live is its twin, `getCardByID` (`callbacks.go:488-497`), which returns a
proper `fmt.Errorf("card not found: %s", cardID)` but discards the `ListIDCards`
error (`cards, _ := c.repo.ListIDCards()`). A failed registry read therefore
surfaces as the misleading "card not found" — sending the user to look for a card
that is right there in the file they cannot read.

**Do this.** Propagate the read error with distinct wording. Step 2's `resolveCard`
subsumes this function, so fold it in there rather than fixing both.

---

## Verification

```
go build ./...
go test ./...
go test -race ./internal/app/... ./internal/config/...
golangci-lint run ./...
```

Manual smoke (`go run ./cmd/kankuro`):
1. Idle the app for 60s and confirm the I/O drop against step 0's baseline
   (re-run the same measurement — this is the deliverable).
2. Edit `~/.config/kankuro/id_cards.yaml` in another terminal (rename a card); the
   dashboard must pick it up within a few seconds (step 1 move 2).
3. Create a card in the UI — it must appear immediately, not after a tick
   (step 1 move 1).
4. Launch a card with several skills and confirm the skills land in the workspace in
   the same order as before the change (step 2).

## Out of scope

- `internal/session`'s own persistence cadence (a JSON write per status flip). It is
  adjacent and arguably worse than the config polling, but it belongs with the
  session package — note it as follow-up.
- Any change to the 2s interval constant itself, unless step 0's numbers justify it.
- `fsnotify` or any new dependency.
