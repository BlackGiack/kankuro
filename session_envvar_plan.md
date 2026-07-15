# Plan: Add environment variables to a running session

## Context

KANKURO sessions are spawned as PTY sub-processes. Today the environment for a
session comes from `IDCard.Env` (`internal/config/types.go:22`), which flows into
`driver.AgentConfig.Env` at launch (`internal/session/lifecycle.go:202`) and is
applied to the child process via `cmd.Env = mergeEnvWithPwd(...)`
(`internal/driver/driver.go:236`). Two gaps make "add env vars to a running
session" impossible right now:

1. There is **no UI** to edit env at all — the card editor silently drops
   `Env` in both `SetCard()` and `Card()` (`internal/ui/views/card_editor.go:202,284`),
   and there is no per-session env surface.
2. A process's environment is **immutable after `pty.Start`** — `cmd.Env` is read
   once. You cannot inject env into a live child; the only honest way to apply new
   env is to **re-spawn** (relaunch) the process.

**Chosen approach (per user):** a **per-session env override** stored in the
session's JSON (merged over the card's `Env` at spawn), edited from the dashboard,
and applied by **saving then confirming a relaunch**. Card `Env` stays as a
default layer.

Outcome: user selects a session, presses a key, edits `KEY=value` pairs, confirms
relaunch, and the agent restarts with the new environment. The override persists
across app restarts and future relaunches.

A code review of the touched paths surfaced three defects that this feature
directly depends on for correctness; they are folded into this plan (see
"Bundled correctness fixes"). All other review findings are tracked separately in
`tech_debt_founding.md`.

## Data model & persistence (session package)

- **`internal/session/manager.go`**
  - Add `EnvOverride map[string]string` to `Session` (struct at :30).
  - Add `Env map[string]string` to `Snapshot` (struct at :75) and populate it in
    `List()` (:178) so the dashboard can prefill the editor.
  - In the load/hydrate loop (:128), restore `EnvOverride: d.Env`.
  - Add method `SetSessionEnv(id string, env map[string]string) error`: under
    `m.mu.Lock()`, set `s.EnvOverride = env`, then persist and **return the
    persist error** (see Bundled fix C — do not swallow it).
- **`internal/session/storage.go`**
  - Add `Env map[string]string \`json:"env,omitempty"\`` to `Data` (:10).
- **`internal/session/lifecycle.go`**
  - `persistSessionLocked` (:270): set `Env: s.EnvOverride` on the `Data` literal.
  - `startSession` (:153): before `d.Spawn`, read the existing session's override
    (`m.sessions[sid].EnvOverride` under RLock — present on the relaunch path,
    nil for brand-new launches) and pass `Env: mergeEnv(card.Env, override)`
    instead of `Env: card.Env` (:202). Override keys win over card keys.
  - Apply the same merge in `SwitchAgent` (:123) so a switched agent keeps the
    session's env override.
  - Add small helper `mergeEnv(base, override map[string]string) map[string]string`
    in this package (base copied, override applied on top).

## UI: per-session env editor (views package)

- **New file `internal/ui/views/session_env_form.go`** — a small modal built on
  the existing `components.Modal` + a single `textinput`, modeled on the MCP form's
  env handling. **Reuse** `parseKeyValueList` / `formatKeyValueMap` from
  `internal/ui/views/mcp_form.go:225-241` (same package — no duplication), after
  applying Bundled fix D. Placeholder like `KEY=value, KEY2=value2`. Legend:
  `Enter: Save & relaunch • Esc: Cancel`. Exposes `SetEnv(map)`, `Env() map`, and
  the target `sessionID`.
- **`internal/ui/views/dashboard.go`**
  - Add `case "v":` to the session key switch (:341-364) → `handleEditEnv()`.
    Model it on `handleKill()` (:551): guard `focusedPane == PaneSessions`, get
    `d.sessionList.Selected()`, allow any status. Prefill the form from `sess.Env`
    and route input to it while open (mirror the `deleteModal` routing at :663-693).
  - Add callback field
    `OnEditSessionEnv func(sessionID string, env map[string]string) (tea.Cmd, error)`
    to `DashboardCallbacks` (:47).
  - On form save, invoke the callback; then show the existing-style confirm modal
    (`components.NewModal`, pattern at :102 / :663) asking *"Relaunch now to apply
    new environment?"* — only when the session is currently **Running/Waiting**
    (a terminated session just gets the env saved for its next relaunch).

## UI wiring & relaunch (app package)

- **`internal/app/callbacks.go`** — add `OnEditSessionEnv` to the
  `DashboardCallbacks{}` literal (:18). Handler:
  1. `c.session.SetSessionEnv(sessionID, env)` — surface its error via
     `messages.ErrorMsg` (so a failed persist is not silently reported as saved).
  2. If confirmed-relaunch and the session is Running/Waiting: `c.session.Kill(sessionID)`
     then `c.session.Relaunch(sessionID, card, skills, personas, mcps)` — reuse the
     exact card/skills/personas/mcps resolution already in `relaunchSession`
     (:131-157). If terminated, skip the relaunch.
  3. Return `c.refreshData()` (or `SessionAttachRequestedMsg` to re-attach).

## Bundled correctness fixes (dependencies of this feature)

These are prerequisites — the feature is incorrect without them.

**A. `mergeEnvWithPwd` must upsert, not append** — `internal/driver/driver.go:537-543`.
Today custom env is `append`-ed onto `os.Environ()`, producing duplicate keys;
`getenv` returns the *inherited* value, so an override that collides with an
existing variable is silently ignored. Route each `config.Env` entry through the
existing `upsertEnv` helper (same treatment `PWD` already gets two lines down).
Without this, the entire feature can silently no-op for common vars.

**B. Relaunch must not be clobbered by the previous life's goroutines** —
`internal/session/watch.go:9-38` and `internal/session/io.go:10-34`. The apply
path is Kill→Relaunch, which is exactly the window where the *old* per-launch
`watch` / `trackOutput` goroutines (capturing the old driver / output channel) can
wake up and flip the freshly relaunched session back to `Terminated` and close its
new subscribers. Add an ownership guard: give each activation a generation counter
(field on `Session`, incremented in `activateSessionLocked` at `lifecycle.go:247`),
capture it in the goroutine closures, and make `watch`/`trackOutput` no-op/return
when `s.generation` no longer matches. Verify with `go test -race`.

**C. Do not swallow the env persist error** — `internal/session/lifecycle.go:284`
(`_ = saveSession`). `SetSessionEnv` needs to report failure so the UI does not
tell the user env was saved when the write failed. Add a persisting variant that
returns the error (or have `SetSessionEnv` call `saveSession` directly and return
its error) rather than reusing the fire-and-forget `persistSessionLocked`.

**D. Stabilize the reused key/value helpers** — `internal/ui/views/mcp_form.go:235`
(`formatKeyValueMap`). It iterates a map with no sort, so the env text field will
reshuffle across renders. Sort keys before joining. (The comma-delimited,
no-escaping limitation is documented as a known constraint in `tech_debt_founding.md`,
not fixed here.)

## Docs

- Update `kankuro_specs.md` §4 "Dashboard Interactions" to document the new key:
  ``* `v`: Edit environment variables for the selected session (prompts to relaunch
  to apply).``

## Out of scope (note for user)

- The card editor also drops `SystemInstruction` on save (same root cause as the
  `Env` drop). This plan fixes the `Env` half via a round-trip in `SetCard()`/`Card()`
  (store the loaded card and copy `Env` back through in `Card()`); the broader
  card-editor field-preservation problem and adding a UI input for
  `SystemInstruction` are tracked in `tech_debt_founding.md`.

## Verification

1. `go build ./...` and `go test ./internal/...`, plus `go test -race ./internal/session/...`
   to confirm Bundled fix B (relaunch race) under the Kill→Relaunch path.
2. Add/extend tests:
   - `internal/session`: `SetSessionEnv` persists to JSON (`Data.Env`) and reloads
     into `Session.EnvOverride`; `mergeEnv` override-wins semantics; a Kill→Relaunch
     sequence keeps the session Running (guards fix B).
   - `internal/driver`: `mergeEnvWithPwd` override of an already-present env var
     actually wins (guards fix A).
   - `internal/app/integration_test.go` (mirrors `OnRelaunch` at :251): set env,
     relaunch, assert the merged env reaches `driver.AgentConfig.Env`.
   - `internal/ui/views`: env-form round-trip; `formatKeyValueMap` output is sorted
     (guards fix D).
3. Manual end-to-end (`go run ./cmd/kankuro`): launch a session; press `v`; enter
   `FOO=bar`; confirm relaunch; attach and run `printenv FOO` in the agent terminal.
   Then set an override for a var that already exists in the parent env (e.g.
   `PATH`-adjacent) and confirm the override wins. Restart KANKURO, relaunch the
   same session, confirm `FOO` still applies (persistence).
