# project.md — Agent onboarding: KANKURO

Read this first. It covers what the project does, how it is built, where everything
lives, the rules the code relies on, and which files to open for a given kind of task.
`kankuro_specs.md` describes what the product should do; this file describes how the
code actually does it. Last reviewed: 2026-09-29 (commit `46580b4`).

---

## 1. What this project is

**KANKURO** is a terminal UI, written in Go, that orchestrates AI coding-agent CLIs —
Claude Code today, any CLI through a generic driver. It treats each CLI as a "puppet":
you describe an agent once, launch as many sessions of it as you like, and switch
between them from a single dashboard.

The core concepts:

| Concept | What it is | Where |
|---|---|---|
| **ID card** | A saved agent template: driver (which CLI), model, workspace, whether to use a temporary workspace, selected skills / personas / MCP servers, AGENTS.md content, env, face and puppet style. | `config.IDCard` |
| **Workspace registry** | Directories you can run agents in. A *source* is either one workspace or a root whose direct children each become a workspace. Cards reference a workspace by ID, never by path. | `config/workspaces.go` |
| **Capability registries** | Skills, personas and puppets are discovered from source directories; MCP servers are entered by hand. Cards reference them by ID. | `config/*_sources.go`, `config/mcps.go` |
| **Session** | A running (or stopped, relaunchable) instance of a card: a child process in a PTY, its scrollback, and a JSON record on disk so it survives restarts. | `internal/session` |
| **Temporary workspace** | Optionally, a session runs in a disposable copy of the workspace rather than the original, so the agent cannot damage the source. The copy persists across relaunches of that session. | `session/workspace.go` |
| **Status** | `running`, `waiting` (the agent looks like it is waiting for input — the dashboard highlights it and shows a neutral face) or `terminated`. | `session.Status`, `driver.IsPromptLike` |

Roadmap (from the specs): Phase 0 (this MVP) is done; Phase 1 is inter-session
communication — linking agents in a chain; Phase 2 is automation (triggers, a local
LLM controller, a notification stack).

---

## 2. Tech stack

- Go **1.26.1**, single module `kankuro`. No CGO.
- TUI: **Bubble Tea v2**, **Lip Gloss v2**, **Bubbles v2** (`charm.land/*/v2`), plus
  `charmbracelet/x/ansi` and `x/term`. The project was migrated from v1 to v2;
  `impl_plan.md` and `review.md` record that migration.
- PTY: `creack/pty`. Terminal width of wide runes: `mattn/go-runewidth`.
- Storage: YAML (`gopkg.in/yaml.v3`) for config, JSON for session records.
- IDs: `google/uuid` for cards and sources; SHA-256 hex for skills, personas,
  puppets and workspaces derived from their source path.
- No test framework beyond the standard library.
- One file embedded into the binary: the built-in `mcp-cli` skill
  (`embed_skills.go` → `kankuro_skills/mcp-cli/SKILL.md`).

---

## 3. Repository map

```
cmd/kankuro/main.go          # flags, debug logger, repository, driver registry, session manager, tea.Program
embed_skills.go              # //go:embed of kankuro_skills/mcp-cli/SKILL.md (module root, so it can reach the dir)
kankuro_skills/mcp-cli/      # the one built-in skill's source
.golangci.yml                # lint config — the suite exits 0 and must stay that way
internal/
  app/                       # glue between the UI and everything else
    app.go                   #   tea.Model wrapper around the Coordinator
    coordinator.go           #   ★ view routing, message dispatch, data refresh, preview subscription
    callbacks.go             #   every view's callbacks: launch, relaunch, switch, sources, IDE, delete card
    attach.go                #   raw-PTY attach/detach (releases Bubble Tea, forwards stdin, detects Ctrl+K)
  config/                    # YAML repository for everything the user configures
    repository.go            #   FileRepository (one mutex for all read-modify-write), interfaces, app settings
    storage.go               #   WriteAtomic (temp file + fsync + rename), Fingerprint (mtime+size per registry)
    types.go                 #   IDCard, Workspace, Skill, Persona, MCPServer, Puppet, AppSettings
    id_cards.go mcps.go skills.go personas.go workspaces.go *_sources.go
    sync.go                  #   scanning source dirs; expandPath (errors on undefined $VARS)
    builtins.go              #   built-in skills (mcp-cli); built-in puppets live in internal/ascii
    validation*.go errors.go #   card validation; SyncFailedError
  driver/                    # one AgentDriver per CLI
    driver.go                #   GenericCLIDriver (.agents/, AGENTS.md) and ClaudeCodeDriver (.claude/, CLAUDE.md);
                             #   Spawn, Kill, SendInputBytes, skill/persona/MCP file sync, IsPromptLike
    registry.go factory.go   #   name → driver constructor ("claude" built in; unknown names get the generic driver)
  fsutil/                    # WriteFileAtomic: temp file + fsync + rename, shared by config and session
  session/                   # process lifecycle and persistence
    manager.go               #   Manager, Session/Snapshot types and their mapping functions
    lifecycle.go             #   Launch, Relaunch, SwitchAgent, Kill, Remove, persistence calls
    watch.go io.go           #   per-activation goroutines: status polling, output tracking and scrollback
    attach.go                #   AttachHandle (input routed through the driver's locked writer)
    storage.go workspace.go  #   session JSON load/save; temporary-workspace creation
    ide_launcher.go          #   open a session's workspace in the configured IDE
  ui/
    messages/messages.go     #   every tea.Msg the app exchanges — read this before adding one
    views/                   #   full screens: dashboard, card_editor, mcp_form, source_manager, settings,
                             #   session_launcher, agent_switch, and three modals (each embeds components.ModalState)
    components/              #   reusable pieces: card_list, session_list, session_preview, capability_list,
                             #   source_list, modal, filter (generic search), renderer (puppets)
    component/ styles/       #   base component type; the colour palette (see crow_puppet_theme.md)
  testenv/                   # test-only: private TMPDIR root + leak guard for TestMain (session, app)
  ascii/                     # canvas and built-in puppet/face art (cat, darth, default, kankuro, strawhat, totoro)
```

Other files at the root are documents — see §12.

---

## 4. How it fits together

### The UI loop

Bubble Tea calls `App.Update`, which hands every message to the **Coordinator**. The
coordinator handles app-level messages itself (navigation, errors, data refresh,
session events, preview output) and passes everything else to the current view.
Views never call each other; they return a `tea.Cmd`, and the coordinator reacts to
the message it produces.

Commands run on their own goroutine. That is the single most important fact about this
codebase — see invariant 1 in §6.

### Data flow

- **Config** lives in `config.FileRepository`. On a 2-second tick the coordinator
  stats the registry files (`Fingerprint`) and only re-reads them if one changed; any
  in-app mutation asks for a refresh immediately with `DataRefreshRequestedMsg`.
  A full read arrives as `DataLoadedMsg` and is handed to every view through
  `DataReceiver.SetData`.
- **Sessions** live in memory in `session.Manager`. The tick refreshes them every time,
  as `SessionsRefreshedMsg`, straight into the dashboard.
- **Preview output** of the selected session streams from a subscription channel as
  `SessionOutputMsg`, forwarded as `PreviewDataMsg` to the views that render a preview.

### A session's life

1. **Launch** (`Manager.Launch` → `startSession`): resolve the workspace; create the
   temporary copy if the card asks for one; sync the driver's files into the workspace
   (AGENTS.md or CLAUDE.md, skills into `<agent folder>/skills/`, personas into
   `<agent folder>/agents/`, and the card's MCP servers into
   `.kankuro/mcp-servers.json`, which the built-in `mcp-cli` skill teaches the agent to
   use — the card editor auto-selects that skill when an MCP is added); spawn the CLI in a PTY
   with the card's env merged over the parent's; start the `watch` and `trackOutput`
   goroutines; write the session record.
2. **Attach** (`app/attach.go`): Bubble Tea releases the terminal, the scrollback is
   replayed, stdin goes to the agent. **Ctrl+K** detaches, both as the legacy byte and
   as a Kitty keyboard-protocol sequence.
3. **Kill / termination**: SIGTERM to the agent's process group, then SIGKILL after 2 s
   if anything in the group is still alive (`GenericCLIDriver.Kill`) — the leader or a
   descendant that ignores SIGTERM. Status becomes `terminated`; the record stays on disk.
   **Quit**: every running session is terminated this way before kankuro exits —
   `main.go` calls `Manager.Shutdown` after `p.Run()` returns, whatever it returned, and
   turns SIGHUP (closing the terminal window) into a normal quit. `Shutdown` kills the
   sessions in parallel, waits for launches/relaunches/switches already in flight (each
   kills its own just-spawned process and removes a new session's tmp copy; this is the
   `Manager.activations` WaitGroup), stops the watchers, and refuses later activations
   (`ErrManagerClosed`).
4. **Relaunch / switch agent**: a terminated session restarts in the *same* workspace,
   with the same card or a different one. Each activation bumps a generation counter so
   goroutines left over from the previous run stop touching the session.
5. **Remove**: deletes the record and, for a temporary workspace, the copy.

### Dashboard keys

Three panes: ID cards (left), the preview of the selected session (centre), sessions
(right). `Tab`/`Shift+Tab`/arrows switch pane; `/` searches.

| Key | Action |
|---|---|
| `Enter` | on a card: launch a session; on a session: attach if running, relaunch if terminated |
| `n` / `e` | new card / edit the selected card |
| `a` | attach to the selected session |
| `x` | kill the selected session |
| `s` | switch agent on a terminated session |
| `Backspace` / `Delete` | remove a terminated session (asks for confirmation) |
| `o` | open the session's workspace in the IDE (command set in settings) |
| `m` / `w` / `c` | MCP form / workspace sources / settings |
| `q` | quit |
| `Ctrl+K` | detach from an attached session |

The card editor's own keys (`Tab` between fields, `Space` to toggle, `Ctrl+E` to add a
source) are listed in `kankuro_specs.md` §4.

---

## 5. Storage layout

Everything lives in the user config directory (`os.UserConfigDir()/kankuro`, so
`~/.config/kankuro` on Linux), overridable with `KANKURO_CONFIG_DIR` or `--config-dir`.

```
id_cards.yaml
skills.yaml          skills_sources.yaml       # registry + the dirs it is scanned from
personas.yaml        persona_sources.yaml
puppets.yaml         puppet_sources.yaml
workspaces.yaml      workspace_sources.yaml
mcps.yaml                                        # entered by hand, no sources
app_settings.yaml                                # IDE command, which card-editor sections are shown
sessions/<session-id>.json                       # one record per session
tmp/<session-id>/                                # temporary workspace copies (or $KANKURO_TMP_DIR)
```

- Every YAML write goes through `WriteAtomic` under the repository mutex.
- Session JSON goes through the same `fsutil.WriteFileAtomic` as config, under the
  manager mutex. A crash leaves a `.<id>.json.<rand>.tmp` at worst, which the
  `*.json` glob in `loadSessions` ignores.
- Loading a session always marks it `terminated`: a process cannot survive a restart.
- The debug log (`--debug`) goes to `os.UserCacheDir()/kankuro/kankuro.log`. Without
  `--debug`, the session package's logger discards output — **never write to stdout or
  stderr from app code**; it corrupts the TUI.

---

## 6. Invariants — rules the code relies on

These came out of real bugs during the 2026-09 tech-debt work. Each one has bitten at
least once; none of them is enforced by the compiler.

1. **No `tea.Cmd` closure touches coordinator state.** A command's closure runs on its
   own goroutine, so it must never read or write `c.views`, `c.previewSessionID`,
   `c.previewSubChan`, `c.currentView` or `c.err` — doing so is a concurrent map access
   that can crash the app. `refreshData()` snapshots what it needs *before* returning its
   closure. Code already inside a closure requests a refresh with
   `messages.DataRefreshRequestedMsg{}` (or `refreshRequestedCmd()`), never by building
   `refreshData()` there. Check: `grep -n "c.refreshData()" internal/app/callbacks.go`
   must be empty. **The tests do not catch a violation; the grep does.**
2. **Never navigate from inside a view's `Update`.** `Coordinator.Update` writes the view
   returned by `updateCurrentView` back into `c.views[c.currentView]` *after* the view's
   `Update` returns. A callback that calls `navigateTo` during that `Update` changes
   `c.currentView` first, and the old view is then stored under the new view's key —
   this once replaced the dashboard with the agent-switch view. Navigate from a message
   handler in `handleSpecialMessages` / `handleSessionMessages`; those return early and
   never write `c.views`.
3. **`DataLoadedMsg` is always a full snapshot.** A `nil` field means "that registry could
   not be read this round — keep what you have", never "this update doesn't concern that
   field". Partial updates have their own messages (`PreviewDataMsg`,
   `SessionsRefreshedMsg`). Every `SetData` guards each field with `if msg.X != nil`.
   This matters because `CapabilityList.SelectedIDs()` only returns IDs whose items are
   currently loaded: wipe the items and saving the card silently drops its skills.
4. **The error banner is cleared by `navigateTo`.** Never batch an `ErrorMsg` beside a
   navigation; set `c.err` in the handler *after* navigating
   (`tech_debt_act_2.md` §3a proposes removing this trap).
5. **The card editor overlays, it does not rebuild.** `CardEditor.Card()` starts from the
   card loaded by `SetCard()` and assigns only the fields the form edits, so fields the
   form doesn't show (`SystemInstruction`, `Env`, timestamps, anything added later)
   survive a save. `Reset()` clears the loaded card for the "new card" path. Keep it that
   way when adding fields; `TestCardEditorPreservesUnmanagedFields` guards it.
6. **Config read-modify-write goes through `Update*Sources(fn)`**, which reads, applies
   `fn`, writes and re-syncs under one lock. Never list, mutate and save in separate
   calls — a failed read would then persist a truncated list. A save that succeeded but
   whose re-sync failed returns `*config.SyncFailedError`: refresh *and* warn.
7. **Exported repository methods take the mutex; unexported `*Locked` helpers assume it.**
   Go mutexes are not reentrant, so a locked method calling another locked method
   deadlocks the whole TUI. Internal calls use the `*Locked` variants.
8. **"Started, but the record didn't save" is not a failure.** `Launch`, `Relaunch` and
   `SwitchAgent` wrap a post-spawn persist error in `session.ErrPersistFailed`; the UI
   checks it with `errors.Is` and still offers attach, showing a warning.
9. **Rendered output keeps its shape.** Fit content to its box by measured width
   (`lipgloss.Width` / `ansi.StringWidth`), never by byte length, and never let lipgloss
   word-wrap a row that must stay on one line: wrapping leaves trailing whitespace
   hanging past the box edge, and one extra column makes the outer border wrap the whole
   row. Clip art, don't wrap it. See `components/card_list.go` (`renderPills`,
   `clipBlock`) and its sweep test.

---

## 7. Build, test, lint, run

```bash
go build ./...
go vet ./...
go test ./...
go test -race ./internal/...          # must stay clean
golangci-lint run                     # must exit 0
```

All five are clean as of the review date. Run all five before calling a change done.

**Running the app by hand** — always isolate it from your real config
(`kankuro_specs.md` §5.1):

```bash
KANKURO_TMP_DIR=/tmp/kankuro-test/tmp \
KANKURO_CONFIG_DIR=/tmp/kankuro-test/config \
  go run ./cmd/kankuro --debug
```

Both variables are read at startup. Without them, a test run writes real cards and
session records into `~/.config/kankuro`, and every temporary-workspace session leaves
a full copy of the workspace under `~/.config/kankuro/tmp/`. Reading the real config
to reproduce a user's bug is fine; writing to it is not.

**Testing notes**

- `internal/session` tests use `t.TempDir()` and real short-lived processes (`cat`) as
  agents; they include deliberate sleeps past the 2-second watch interval, which is why
  that package takes about 10 seconds.
- **Tests that launch sessions must build the manager with `newTestManager`** (in each
  package's `main_test.go`; `app` has a second copy in `helpers_internal_test.go`), never
  `session.NewManager` directly. It registers a `Shutdown` cleanup, so the manager's watcher
  goroutines stop before `t.TempDir()` is removed — a watcher that outlives its test
  recreates `<tempdir>/sessions/` to write a record, and nobody cleans that up.
  `internal/testenv` (called from both `TestMain`s) redirects `TMPDIR` and
  `KANKURO_TMP_DIR` into a private root, and after the run waits 3 s (longer than the
  watch interval) and fails the package if anything is left in that root. That costs
  about 3 s per package. The root is always removed, so even a failure never leaks into the
  real `/tmp`.
- `internal/driver` kill tests run real `sh` scripts in a PTY and register a cleanup that
  SIGKILLs the process group, so a failing test does not leave `sleep 1000` running.
- The coordinator is tested by driving it: `app.NewCoordinator(repo, sm)`, then
  `coord.Update(msg)`, then inspect `coord.View()`, `coord.CurrentView()`,
  `coord.Views()`. Go through `Update` rather than calling a callback directly —
  calling `coord.AgentSwitchCallbacks().OnSwitch(...)` skips the write-back where
  invariant 2 breaks.
- `collectMsgs` (two copies: `session_warning_test.go` and
  `refresh_cadence_internal_test.go`) unwraps `tea.BatchMsg` only; commands built with
  `tea.Sequence` come back opaque.
- Rendering tests: `internal/ui/uitest` holds the shape assertions and adversarial
  fixtures (a normal package, imported only from `_test` files). Every view or
  component has a layout test with its contract written in a header comment, and the
  contract is one of two kinds. **Exact**: every line is exactly `W` columns and there
  are exactly `H` lines (`AssertExactRect`) — anything bordered or placed into a fixed
  box, such as the dashboard body, `lipgloss.Place` overlays, the card. **Bounded**: no
  line wider than `W`, at most `H` lines (`AssertWithin`) — unpadded components such as
  `SessionList`, `SessionPreview` and `CardEditor`, whose parent pads them. Add a
  region check (`uitest.Region`) wherever the layout has fixed regions, since a
  rectangle can survive a spill. **Test a component's own output, not only after its
  parent's clamp**: `Dashboard.renderPreviewPane` clamps the preview, so a preview
  that only fits because of that clamp would pass a dashboard test. Files:
  `components/card_render_layout_test.go`, `session_list_layout_test.go`,
  `session_preview_layout_test.go`, `views/dashboard_layout_test.go`,
  `card_editor_layout_test.go`, `modals_layout_test.go`. Known layout bugs found this
  way are `t.Skip("bug: ...")` tests listed in `impl_plan_2.md` "Found, not fixed".
- **Make sure each new test can fail.** Remove the fix, run the test, watch it fail, put
  the fix back. The recurring problem in past work was tests that passed with the bug
  still present.

---

## 8. Configuration

| Setting | Where | Effect |
|---|---|---|
| `KANKURO_CONFIG_DIR` / `--config-dir` | env / flag | config root (default `~/.config/kankuro`) |
| `KANKURO_TMP_DIR` | env | root for temporary workspaces (default `<config>/tmp`) |
| `--debug` | flag | log to `<user cache dir>/kankuro/kankuro.log` |
| IDE command, card-editor sections | Settings view (`c`) → `app_settings.yaml` | what `o` runs; whether personas / MCPs are shown in the card editor |
| per-card `env:` | `id_cards.yaml` | merged over the parent environment when the agent is spawned; a card key overrides an inherited one. There is no UI for it yet — edit the YAML. |

---

## 9. Conventions & workflow

- **Never commit unless explicitly asked** (`AGENTS.md`).
- Main working branch: `feat/envvar`; the upstream default is `master`. The history is
  linear — when a branch has to join another, prefer cherry-picking or rebasing a single
  commit over merging, and never stash/pop across branches that delete files.
- If you work in parallel sessions, `git merge-tree` the branches before merging.
  `internal/app` is where parallel work collides.
- Match the surrounding code: explicit error wrapping (`fmt.Errorf("...: %w", err)`),
  `errors.Is` / `errors.As` for sentinels, no `log.Printf` to stdout.
- When a change fixes a non-obvious bug, leave a comment saying *why* the code is the way
  it is — most of the invariants in §6 are documented that way at their call sites.
  Keep those comments true: if you change the behaviour, update the comment.
- Lint suppressions need a reason in the comment (`//nolint:mnd // why`). The one
  project-wide exclusion (the `driver` package name) is in `.golangci.yml` with its
  reason.

---

## 10. Where to start, by task

| I want to… | Look at |
|---|---|
| Add a new agent CLI | `internal/driver/driver.go` (embed `GenericCLIDriver`, override only what differs — agent folder, instruction file name, model flag), register it in `driver/registry.go` |
| Change what gets written into a session's workspace | `driver.go` (`SetAgentsMD`, `SetSkills`, `SetPersonas`, `SelectMCPs` → `.kankuro/mcp-servers.json`), `session/lifecycle.go` `syncDriverFiles` (which also clears generated skills/personas from temporary workspaces on each launch) |
| Change session lifecycle or status detection | `session/lifecycle.go`, `session/watch.go`, `session/io.go`, `driver.IsPromptLike`; mind the generation counter **and the `activating` reservation**; goroutines that can write to disk must be stoppable by `Shutdown` (see `Manager.stop` / `startWatchLocked`) |
| Change attach / detach / keyboard handling during attach | `app/attach.go` |
| Add a field to ID cards | `config/types.go`, `config/validation.go`; if the form edits it, `views/card_editor.go` (`SetCard`, `Card`, `focusTargetTable`); invariant 5 |
| Add a new config registry or source type | `config/repository.go` (interfaces, filename), a new `config/*.go`, `Update*Sources`, `Fingerprint`'s file list in `config/storage.go` |
| Add a screen | `ui/views/`, register it in `coordinator.initViews`, add a `ViewType`, wire callbacks in `app/callbacks.go` |
| Add a message | `ui/messages/messages.go`, handle it in `coordinator.go`; if it carries partial data, give it its own type (invariant 3) |
| Change how cards look | `components/card_list.go` (`RenderCardItem`), `components/renderer.go` (puppets), `styles/styles.go`; invariant 9 and the sweep test |
| Add or change a puppet | built-in art in `internal/ascii/characters`; user puppets are `.puppet` text files found in puppet source dirs (`config/puppet_sources.go`, which also truncates oversized art on read) |
| Change dashboard layout or keys | `views/dashboard.go`, then update the key table in §4 and in `kankuro_specs.md` |
| Change refresh behaviour | `coordinator.go` (`refreshData`, `refreshSessions`, `maybeRefreshConfig`), `config.Fingerprint` |
| Pick up known debt | `tech_debt_act_2.md` |

---

## 11. Known quirks / pitfalls (check before "fixing")

- **`navigateTo` clears the error banner** — see invariant 4.
- **Key/value fields in the MCP form** use `,` as the separator with `\` escaping
  (`\,`, `\\`); keys containing `=` and surrounding whitespace in values do not
  round-trip (neither can occur in env or header names).
- **The card's driver line reads `claude |`** with nothing after it when the card has no
  model set.
- **Capability pills show at most `MaxPillsShown` (4)**, then a `+N` pill, and fewer when
  the card is narrow — deliberate, so the row never wraps.
- **The agent-switch search matches the concatenation `Name Driver Model`**, so a query
  spanning two fields matches too.
- **`Kill` (and so quit) only reaches the agent's process group.** A descendant that starts
  its own session or process group (a daemonising tool, a job-control shell) survives it.
  `Pdeathsig` is no fix: it fires when the spawning OS *thread* exits, which in Go can
  happen at any time.
- **The `commands` package and `app/wiring.go` are gone.** Older documents still mention
  them; they were dead code.

---

## 12. Other documents in the repository

| File | What it is |
|---|---|
| `kankuro_specs.md` | Product specification: concepts, roadmap, UI and keys. §5.1 is the manual-testing setup. |
| `crow_puppet_theme.md` | The colour palette used in `ui/styles`. |
| `.leftover/` | Hand-off notes from earlier sessions. Historical. |
