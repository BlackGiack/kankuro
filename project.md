# project.md — Agent onboarding: KANKURO

Read this first. It covers what the project does, how it is built, where everything
lives, the rules the code relies on, and which files to open for a given kind of task.
`kankuro_specs.md` describes what the product should do; this file describes how the
code actually does it. Last updated: 2026-10-25 (Feature 5 independently accepted
at `b95b853`; fixture cleanup and repeated/full checks verified; Feature 6 may
proceed. See `feature_5.md`).

---

## 1. What this project is

**KANKURO** is a terminal UI, written in Go, that orchestrates AI coding-agent CLIs —
Claude Code today, any CLI through a generic driver. It treats each CLI as a "puppet":
you describe an agent once, launch as many sessions of it as you like, and switch
between them from a single dashboard.

The core concepts:

| Concept | What it is | Where |
|---|---|---|
| **ID card** | A saved agent template: driver (which CLI), model, workspace, optional sub workspaces, whether to use a temporary workspace, selected skills / personas / MCP servers, Dashboard Control opt-in, AGENTS.md content, env, face and puppet style. | `config.IDCard` |
| **Dashboard Control** | Per-card opt-in for discovery/control over other sessions in the linked dashboard. The grant is captured per activation, effective only when Kankuro MCP is selected, preserved on relaunch, and reapplied on explicit switch. | `config.IDCard.DashboardControl`, `session.Session.DashboardControl` |
| **Workspace registry** | Directories you can run agents in. A *source* is either one workspace or a root whose direct children each become a workspace. Cards reference a workspace by ID, never by path. | `config/workspaces.go` |
| **Capability registries** | Skills, personas and puppets are discovered from source directories; MCP servers are entered by hand. Cards reference them by ID. | `config/*_sources.go`, `config/mcps.go` |
| **Session** | A running (or stopped, relaunchable) instance of a card: a child process in a PTY, its scrollback, deployed sub-workspace descriptors, and a JSON record on disk so it survives restarts. | `internal/session` |
| **Message** | A single-line orchestrator prompt pinned to one recipient activation, with explicit replies stored separately from terminal output. | `internal/session/messaging.go`, `internal/api/messaging.go` |
| **Conversation binding** | A native CLI conversation ID, display name, and driver identity saved per card inside one Kankuro session. Names are not identities; bindings never cross Kankuro sessions. | `session.ConversationBinding` |
| **Registration state** | Per-activation `unavailable`, `pending`, `unconfirmed`, `registered`, or `failed`; metadata durability is tracked independently as `not_attempted`, `saved`, or `unsaved`. Runtime state does not survive restart. | `session.RegistrationState`, `session.RegistrationPersistence` |
| **Temporary workspace** | Optionally, a session runs in a disposable copy of the workspace rather than the original, so the agent cannot damage the source. The copy persists across relaunches of that session. | `session/workspace.go` |
| **Status** | `running`, `waiting` (the agent looks like it is waiting for input — the dashboard highlights it and shows a neutral face) or `terminated`. | `session.Status`, `driver.IsPromptLike` |

Roadmap (from the specs): Phase 0 (this MVP) is done; Phase 1 is inter-session
communication — linking agents in a chain; Phase 2 is automation (triggers, a local
LLM controller, a notification stack).

---

## 2. Tech stack

- Go **1.27.3**, single module `github.com/BlackGiack/kankuro`. No CGO. With the
  standard `GOTOOLCHAIN=auto` behavior, a compatible Go command can automatically
  download/select the required toolchain declared by `go.mod`.
- TUI: **Bubble Tea v2**, **Lip Gloss v2**, **Bubbles v2** (`charm.land/*/v2`), plus
  `charmbracelet/x/ansi` and `x/term`. `impl_plan.md` and `review.md` contain Feature 4
  planning and review history; they are not the Bubble Tea migration record.
- PTY: `creack/pty`. Terminal width of wide runes: `mattn/go-runewidth`.
- Storage: YAML (`gopkg.in/yaml.v3`) for config, JSON for session records, operation
  journals, and bounded messaging history (`<config>/messaging/history.json`).
- MCP transport: official Go MCP SDK Streamable HTTP transport on a loopback-only listener.
- IDs: `google/uuid` for cards and sources; SHA-256 hex for skills, personas,
  puppets and workspaces derived from their source path.
- No test framework beyond the standard library.
- Built-in `mcp-cli` and `kankuro-puppet` skills are embedded and injected into
  config at startup (`main.go` → `kankuro_skills/`).

---

## 3. Repository map

```
main.go                      # root command, embedded skill, loopback MCP lifetime, repository, drivers, session manager, tea.Program
main_test.go                 # embedded skill → config → driver deployment integration
kankuro_skills/mcp-cli/      # MCP command-line usage skill
kankuro_skills/kankuro-puppet/ # registration, message, and explicit reply protocol
.golangci.yml                # lint config — the suite exits 0 and must stay that way
internal/
  api/                       # UI-independent shared operations; owns no process or config state
    service.go               #   shared lifecycle, discovery, registration, and capability resolution
    operations.go            #   durable mutation journal, bounded workers, deduplication, polling and cancellation
    messaging.go             #   pinned send admission, explicit reports, retrieval and bounded waits
    service_test.go          #   service resolution contracts without a coordinator or Bubble Tea
  mcp/                       # authenticated Streamable HTTP transport adapter; depends on api.Service, not app/UI
    server.go                #   conversation registration, discovery, and lifecycle tools; host/origin/auth checks
    server_test.go           #   MCP handshake/tool-list/auth/origin regression checks
  app/                       # glue between the UI and everything else
    app.go                   #   tea.Model wrapper around the Coordinator
    coordinator.go           #   ★ the Coordinator struct (fields grouped by goroutine), Update, message routing
    handlers.go              #   coordinator-goroutine halves of refresh and preview: snapshot state, apply results to views
    navigation.go            #   navigateTo (invariant 2 lives in its doc comment), initViews, resize, broadcast to views
    refresh.go               #   `refresher`: config/session read commands, fingerprint check, tick. No Coordinator pointer
    preview.go               #   `previewSub`: the live-output subscription (coordinator-goroutine state)
    session_actions.go       #   `sessionActions`: Bubble Tea command adapters; shared operations delegated to api.Service
    callbacks.go             #   wiring: every view's callbacks (the one place closures can still reach `c`)
    attach.go                #   raw-PTY attach/detach (releases Bubble Tea, forwards stdin, detects Ctrl+K)
    size_guard.go            #   below 80x24: the "too small" notice, and keys dropped except quit
  config/                    # YAML repository for everything the user configures
    repository.go            #   FileRepository (one mutex for all read-modify-write), interfaces, app settings
    storage.go               #   WriteAtomic (temp file + fsync + rename), Fingerprint (mtime+size per registry)
    types.go                 #   IDCard, Workspace, Skill, Persona, MCPServer, Puppet, AppSettings
    id_cards.go mcps.go skills.go personas.go workspaces.go *_sources.go
    sync.go                  #   scanning source dirs; expandPath (errors on undefined $VARS)
    builtins.go              #   built-in skills (mcp-cli, kankuro-puppet) and selectable Kankuro MCP
    validation*.go errors.go #   card validation; SyncFailedError
  driver/                    # one AgentDriver per CLI
    driver.go                #   GenericCLIDriver (.agents/, AGENTS.md) and ClaudeCodeDriver (.claude/, CLAUDE.md);
                              #   model/session arguments, managed instructions, PTY and workspace file sync
    mcp_config.go             #   complete workspace MCP config serialization and credential-matched Kankuro cleanup
    registry.go factory.go    #   name → driver constructor ("claude"/"codex" built in; unknown names get generic driver)
  fsutil/                    # WriteFileAtomic: temp file + fsync + rename, shared by config and session
  session/                   # process lifecycle and persistence
    manager.go               #   Manager, Session/Snapshot, bindings, runtime registration states and mappings
    messaging.go             #   manager-owned durable message store, delivery receipts, reports and cursors
    conversation.go          #   per-activation credentials, pending report staging, binding persistence and revocation
    lifecycle.go             #   Launch, explicit new/resume Relaunch/SwitchAgent, Kill, Remove, persistence calls
    watch.go io.go           #   per-activation goroutines: status polling, output tracking and scrollback
    attach.go                #   AttachHandle (input routed through the driver's locked writer)
    storage.go workspace.go  #   session JSON load/save (optional per-card conversation bindings); temporary-workspace creation
    workspace_launcher.go    #   shared IDE/terminal workspace command execution
    ide_launcher.go          #   IDE compatibility wrapper
  ui/
    messages/messages.go     #   every tea.Msg the app exchanges — read this before adding one
    views/                    #   full screens: dashboard, card_editor, mcp_form, source_manager, settings,
                              #   session_launcher, agent_switch, conversation_choice, and three modals
    components/              #   reusable pieces: card_list, session_list, session_preview, capability_list,
                             #   source_list, modal, filter (generic search), renderer (puppets),
                             #   text (DisplayText, TruncateHead, TruncateTail)
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

Shared launch, relaunch, card-switch, card-deletion, and session-removal rules live
in `internal/api.Service`. It depends on a narrow config repository interface and
the existing `*session.Manager`; the manager remains the only process/lifecycle
owner. `internal/app.sessionActions` adapts these synchronous operations inside
Bubble Tea commands, mapping their errors to the existing UI messages. In
particular, `session.ErrPersistFailed` preserves a usable activation and is shown
as an attach warning. Launch and switch read capability registries once per resolution; their
errors (with list-operation context) abort activation. Relaunch reads only the
current card and does not resolve or redeploy capabilities. Card deletion is
serialized against activation admission through manager-owned pending card
references shared by UI and MCP services. The same lock protects current-card
validation, pending references, published session/conversation reference checks,
and the repository delete; it is not held across workspace preparation or spawn.
`Service.Conversations`, `RelaunchWithChoice`, `SwitchAgentWithChoice`, and
`RegisterConversation` are the conversation-aware entry points. The no-choice
relaunch/switch compatibility methods only start fresh when the destination has no
bindings; the manager rejects omission when bindings exist. `session.ErrPersistFailed`
is reserved for activation-record persistence, while registration persistence is
reported separately as `saved` or `unsaved`.
`Service.ListCards`, `ListSessions`, and `GetSession` accept the activation
credential and resolve deployed permission inside the service; a constructible
caller DTO is never authorization proof. `Service.ResolveCaller` delegates
activation identity and permission checks to `Manager.ResolveDashboardCaller`;
credentials are never returned.
The session snapshot also carries deployed sub-workspace descriptors. New launch
and explicit switch capture the manifest's source IDs/paths and actual copied
destinations; relaunch preserves the saved descriptors. Legacy inspection reads
the workspace manifest without preparing files and reports metadata unavailable
when that manifest is missing or unreadable.

The `internal/mcp` package is a transport adapter over `api.Service`; it does not
depend on the coordinator or own session/process state. It exposes
`register_conversation(conversation_id, name)`, read-only `list_cards`,
`list_sessions`, `get_session`, messaging tools `send_message`, `get_message`,
`report_reply`, `get_replies`, and the lifecycle tools `create_session`,
`stop_session`, `relaunch_session`, `switch_session_card`, `delete_session`,
`get_operation`, and `cancel_operation`. Discovery/control requires a published,
live activation with a deployed Dashboard Control grant. Listings exclude the
caller; DTOs omit secrets and terminal output. Scope currently resolves to the
single default dashboard. Lifecycle mutations require UUID request IDs and use
the durable journal/worker runner in `api/operations.go`: four workers, 32 queued
operations, a five-minute deadline including queue time, and a ten-second tool
wait before polling. Caller activation credentials are not persisted. Startup
marks nonterminal journal entries `interrupted` and never replays them. Cancellation
reaches primary and sub-workspace copies and is checked before activation
publication; stop/delete complete once irreversible work begins. Accepted calls
return an operation result; rejected calls return a structured MCP tool error with
a stable code and safe message, without creating a journal entry. `main.go` binds it to
`127.0.0.1` before the TUI can launch an agent and keeps it available until after
`Manager.Shutdown`. The manager validates the activation bearer credential and
derives session, card, driver, and generation from that credential, never from
tool arguments. Reports that arrive before spawn publication are staged and are
committed before the activation's first session-record write. If final journal
publication fails after an effect, the in-memory observed outcome and target ID are
returned with a durability warning; a restart treats the durable nonterminal intent
as interrupted instead of repeating the effect.

Messaging permissions and activation identity are resolved in `api.Service` and
`session.Manager`. A sender needs a live Dashboard Control activation; the target
must be a different live session with Kankuro MCP deployed, and sends pin the
nonsecret activation ID exposed by `get_session`. Reports require the exact live
recipient activation but do not require Dashboard Control. Message headers are a
protocol cue, never authorization. Send requests use the lifecycle operation runner
and caller UUIDs; the journal stores only the target, activation, and fingerprint,
not the prompt body. Durable intent is written before any PTY bytes. Identical
retries return the same operation/message, changed arguments conflict, and an
accepted send is never replayed after restart.

The manager-owned message store uses atomic mode-0600 JSON writes and a separate
store mutex. It retains at most 64 MiB or 10,000 whole messages per dashboard,
evicting oldest messages together with replies, with no age expiry. Send receipts
retain only compact identity/fingerprint data after history deletion or eviction.
Reports are limited to 16 KiB and 128 per message. `get_replies` defaults to 50
reports, caps a page at 100 reports and 256 KiB, and uses dashboard/filter-bound
cursors. Deletion or eviction advances a cursor generation; a gap resumes at the
oldest retained matching report and may repeat reply IDs. Optional waits default to
zero, cap at ten seconds, and allow 32 concurrent waiters per manager.

The generic driver serializes `SubmitPrompt`, attached `SendInput`, and
`SendInputBytes`; one marked single-line prompt and its final carriage return stay
under one input gate with a five-second write deadline. Admission release, intent
publication, lifecycle invalidation, and driver capture are coordinated under the
manager-owned activation boundary; publication already underway retains its reservation
until its durable transaction completes. Text is nonempty UTF-8, at
most 2 KiB, and rejects line breaks and terminal controls. Delivery states are
`accepted`, `submitting`, `submitted`, `not_submitted`, and `uncertain`; `submitted`
means bytes including Enter were written, not that the CLI processed them. Sends
are FIFO per recipient activation. Stop/replacement waits for an in-progress write
without holding the manager mutex across PTY I/O. Serialization does not clear or
reserve the human's prompt, so already-typed text can mix with an injected prompt.
Partial writes are uncertain and may leave input-buffer residue. The
`kankuro-puppet` skill is deployed alongside `mcp-cli` whenever Kankuro MCP is
selected; its first instruction is native conversation discovery and registration.

MCP creation optionally accepts a session name. Omission uses the card name; a
supplied name is trimmed and an empty result uses the session ID. Raw optional names
are part of request identity, so changing them on retry conflicts. Names persist across
relaunches and switches. Explicit `progress` replies are informational and may precede
completion; they do not advance a workflow.

### CLI startup

New PTYs start at 80×24 until attachment supplies a different size. Before
spawn, Codex and Claude prepare trust for the actual working directory. Successful
spawn is published immediately, so a human can attach and see bounded startup output.
A manager-owned ten-second timer gates MCP message delivery, not attachment or
registration. The pause is an admission policy, not evidence of CLI readiness.
Inspection reports `phase` (`starting`, `active`, or `terminated`), `process_started`,
and `message_delivery_available`. The latter also requires a live current activation,
Kankuro MCP, and no lifecycle operation. A send revalidates this state and its pinned
activation; retry a starting rejection with the same request ID after inspection.
Operation results retain their historical snapshot; `get_session` reads current state.
Restored sessions are terminated, while relaunch starts a fresh timer. Shutdown
cancels and joins timer tasks, and normal output tracking owns published output.

### A session's life

1. **Launch** (`Manager.Launch` → `startSession`): resolve the workspace; create the
   temporary copy if the card asks for one; sync the driver's files into the workspace
   (AGENTS.md or CLAUDE.md, skills into `<agent folder>/skills/`, personas into
   `<agent folder>/agents/`); resolve the complete selected MCP list and atomically
   replace `<WorkingDir>/.kankuro/mcp-servers.json` before spawn. When selected, built-in
   `kankuro` receives the loopback endpoint and a fresh activation bearer credential in
   that file. The built-in `mcp-cli` and `kankuro-puppet` skills use that
   workspace-relative path for registration guidance. Legacy path/token environment variables are stripped from
   parent/card env and never injected. Spawn the CLI in a PTY, start the `watch` and
   `trackOutput` goroutines, and write the session record.
   Instruction sync has three explicit modes: a new temporary workspace appends to
   its copied workspace instructions, a new non-temporary launch preserves an existing
   file when the card has no instructions, and an explicit switch clears managed
   content when the destination has no instructions. Relaunch preserves the deployed
   instruction file, including manual edits.
2. **Attach** (`app/attach.go`): Bubble Tea releases the terminal, the scrollback is
   replayed, stdin goes to the agent. **Ctrl+K** detaches, both as the legacy byte and
   as a Kitty keyboard-protocol sequence.
3. **Kill / termination**: SIGTERM to the agent's process group, then SIGKILL after 2 s
   if anything in the group is still alive (`GenericCLIDriver.Kill`) — the leader or a
    descendant that ignores SIGTERM. Status becomes `terminated`; the record stays on disk.
    If this activation published Kankuro MCP, the manager revokes its token and removes
    that server entry only if the current workspace config still contains that token.
    Ordinary selected servers and other workspace metadata remain. This prevents an
    older session stop from deleting a newer shared-workspace configuration.
   **Quit**: every running session is terminated this way before kankuro exits —
   `main.go` calls `Manager.Shutdown` after `p.Run()` returns, whatever it returned, and
   turns SIGHUP (closing the terminal window) into a normal quit. `Shutdown` kills the
   sessions in parallel, waits for launches/relaunches/switches already in flight (each
   kills its own just-spawned process and removes a new session's tmp copy; this is the
   `Manager.activations` WaitGroup), stops the watchers, and refuses later activations
   (`ErrManagerClosed`).
4. **Relaunch / switch agent**: a terminated session always reuses its saved
   `WorkingDir`, `OriginalDir`, and temporary-workspace mode, regardless of later card
   workspace edits. The saved working directory must exist and be a directory before
   workspace preparation or driver/MCP sync. Missing or non-directory paths produce
   an error with the saved path and restoration guidance; they are never recreated
   or replaced with the card's workspace. Failed validation releases the activation
   reservation so restoring the directory permits retry.
   Relaunch preserves instructions, skills, personas, ordinary MCP entries, and sub
   workspaces; use Switch Agent (including the same card) to reapply configuration.
   If Kankuro MCP was deployed, relaunch refreshes only its endpoint and credential
   in the existing JSON config, preserving other entries and metadata. An invalid
   config blocks that refresh rather than being replaced. A persisted
   `kankuro_mcp_enabled` flag retains that deployment choice across restarts without
   capability-registry reads or reliance on edited card selections.
   If the destination card has saved conversations, the TUI requires
   an explicit choice of a compatible saved ID or “New conversation”; first use and
    cards without bindings start new directly. Drivers pass resume arguments without
    shell interpolation (generic: `--session <id>`; Codex: `resume <id>`). Each
   activation bumps a generation counter and rotates its credential so old goroutines
   and reports cannot mutate the new activation.
5. **Remove**: deletes the record and, for a temporary workspace, the copy.

### Dashboard keys

Three panes: ID cards (left), the preview of the selected session (centre), sessions
(right). `Tab`/`Shift+Tab`/arrows switch pane; `/` searches.

| Key | Action |
|---|---|
| `Enter` | on a card: launch a fresh session; on a session: attach if running, or relaunch if terminated (with a conversation picker when that card has bindings) |
| `n` / `e` | new card / edit the selected card |
| `a` | attach to the selected session |
| `x` | kill the selected session |
| `s` | switch agent on a terminated session (with a conversation picker when the destination card has bindings) |
| `Backspace` / `Delete` | remove a terminated session (asks for confirmation) |
| `o` / `t` | open the selected session workspace in the configured IDE / terminal (Sessions pane only) |
| `m` / `w` / `c` | MCP form / workspace sources / settings |
| `q` | quit |
| `Ctrl+K` | detach from an attached session |

The conversation picker filters by name/ID/driver, always keeps “New conversation”
available, and disables bindings for a different driver. Its full selected ID/driver
details are vertically scrollable with `PgUp`/`PgDn` when they do not fit beside the
list; this keeps the footer and activation controls visible at the minimum terminal size.

The footer is context-sensitive: `[Nav]`, the focused pane's group (`[Card]` or `[Session]`)
and `[Global]`. When that is wider than the terminal it drops the navigation group, then
shrinks `[Global]` to `q=Quit`, then clips (`dashboard.go` `footerText`).

**Minimum terminal size: 80×24** (`views.MinTerminalWidth` / `MinTerminalHeight`). Below it
`Coordinator.View()` shows a centred "Terminal too small" notice in exactly the terminal's
size and `Update` drops every key except `q` / `ctrl+c` (`app/size_guard.go`). Other messages
still flow, so the real view is current when the terminal grows. A size of 0×0 means "no
`WindowSizeMsg` yet" and is not guarded. The views are still sized at small sizes
(`Dashboard.SetSize` never goes negative) but their layout there is undefined. Views are sized to
terminal rows − `views.CoordinatorRows` (4): the single definition, used by `app.resizeAll` and by
the dashboard to fill the terminal exactly.

The card editor's own keys (`Tab` between fields, `Space` to toggle/open, `Ctrl+E` to add a
source, and `Ctrl+X` to clear Sub Workspaces while focused) are listed in `kankuro_specs.md` §4. The form scrolls when it is taller than its box
(`card_editor.go` `layout` / `syncScroll`): the focused block stays in view and `↑ more` /
`↓ more` replace the first / last row of a cut edge; the help line is pinned to the last row.

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
app_settings.yaml                                # IDE/terminal commands, which card-editor sections are shown
sessions/<session-id>.json                       # one record per session
operations/<caller-id>-<request-id>.json         # safe, durable lifecycle request/outcome journal
tmp/<session-id>/                                # temporary workspace copies (or $KANKURO_TMP_DIR)
```

The live MCP config is always `<session WorkingDir>/.kankuro/mcp-servers.json`.
Kankuro credentials are present there only while that activation's entry remains
current; the session JSON and card YAML never store them.

Each session JSON may contain `conversations`, a map from card ID to a list of
`{id, name, driver}` bindings, and `instructions_managed`, which records whether
the active card installed instruction content. Older records without these fields
remain valid; absence of the ownership flag means workspace-authored instructions
were not claimed by Kankuro.
Activation tokens and process registration status are runtime-only and are never
persisted in session records or card YAML. While selected,
the current bearer credential does exist in the workspace MCP file with mode 0600;
shared non-temporary workspaces intentionally share that file.

- Every YAML write goes through `WriteAtomic` under the repository mutex.
- Session JSON goes through the same `fsutil.WriteFileAtomic` as config, under the
  manager mutex. A crash leaves a `.<id>.json.<rand>.tmp` at worst, which the
  `*.json` glob in `loadSessions` ignores.
- Loading a session always marks it `terminated`: a process cannot survive a restart.
- Loading keeps saved conversation bindings but resets live registration state to
  unavailable and revokes all old credentials. New/resumed activations get a fresh
  random bearer token in the workspace MCP config; termination, replacement, failed
  activation, and shutdown revoke it. Cleanup removes the Kankuro entry only if that
  file still contains the revoked token.
- The Kankuro MCP is a selectable built-in (`builtin:kankuro`), not auto-enabled on
  existing cards. Its endpoint exposes conversation registration, discovery,
  lifecycle, and messaging tools; each call enforces activation authorization.
  The manager binds reports to the credential's session/card/activation and rejects
  stale, mismatched, or conflicting IDs. A token authenticates the caller but cannot
  prove ownership of an arbitrary external CLI conversation ID.
- Generic drivers resume with `--session <id>`; `CodexCLIDriver` overrides that to
  `resume <id>` because Codex CLI rejects `--session`. On 2026-10-24, Codex CLI
  0.164.0 created two actual native thread IDs in the isolated Kankuro target's
  temporary workspace. Both were registered with `mcp-cli` (`saved`), and Kankuro's
  A→B→A switch selected the exact first ID and spawned `codex resume <A-ID>` while
  retaining a file created on B. Feature 5 end-to-end acceptance subsequently
  passed with Codex CLI 0.164.0 and OpenCode 2.1.4. Codex prompt submission writes
  text, waits 500 ms cancellably under the input gate, then writes Enter separately
  after checking cancellation. OpenCode uses generic prompt-plus-Enter submission.
  Repeated attached and detached sends produced real assistant responses and
  recipient-originated `completed`, `question`, and `blocked` MCP reports, retrieved
  through MCP. Both CLIs discovered and registered their actual native IDs with
  persistence `saved`; both received the deployed `kankuro-puppet` skill.
  The user's root `sandbox_mode = "danger-full-access"` and
  `approval_policy = "never"` defaults worked in managed Codex sessions without
  CLI permission overrides. Earlier sandbox-denied and host-assisted runs are
  historical diagnostics, superseded by these recipient-originated results.
  Both CLIs passed typed-input interaction, interruption/detach, stop/resume,
  card-switch/stale-activation rejection, and original-send deduplication checks.
  UI/MCP deletion removed history and workspaces without replaying old sends.
  Codex and OpenCode replies survived normal restart; a forced crash during
  submission recovered uncertain delivery and an interrupted operation without
  replay. Existing human drafts join injected prompts; serialization does not
  isolate prompt composition. Missing reports remain unanswered.
  See `review_6.md` and `/tmp/kankuro-feature5-e2e6/verification.md` for the accepted
  matrix, inspected 80×24/120×40 PNGs, recordings, sanitized MCP results, and test-tool
  limitations. Final shutdown exited 0 and left no isolated test processes.
  Feature 4 command outcomes are in `/tmp/kankuro-feature4-native/artifacts/`.
- The debug log (`--debug`) goes to `os.UserCacheDir()/kankuro/kankuro.log`. Without
  `--debug`, the session package's logger discards output — **never write to stdout or
  stderr from app code**; it corrupts the TUI.

---

## 6. Invariants — rules the code relies on

These came out of real bugs during the 2026-09 tech-debt work. Each one has bitten at
least once; none of them is enforced by the compiler.

1. **No `tea.Cmd` closure touches coordinator state.** A command's closure runs on its
   own goroutine, so it must never read or write `c.views`, `c.preview`, `c.currentView`,
   `c.err` or `c.settings` — doing so is a concurrent map access that can crash the app.
   The `Coordinator` struct groups its fields by which goroutine may touch them. Two
   helper types make this hold *by construction*: `refresher` (`refresh.go`) and
   `sessionActions` (`session_actions.go`) hold no `*Coordinator`, so a closure they build
   can only reach the repository, the session manager and the values passed in. The
   coordinator's thin wrappers (`refreshDataSeeded`, `launchIDE`, …) snapshot what a
   command needs — the preview session ID, the fingerprint baseline, the IDE command —
   *before* calling them. The risk now lives only in `callbacks.go` (closures there
   still capture `c`) and in the few closures that capture a channel, like
   `previewSub.next`. Code already inside a closure requests a refresh with
   `messages.DataRefreshRequestedMsg{}` (or `refreshRequestedCmd()`), never by building
   `refreshData()` there. Two checks, run both:
   - `grep -n "c.refreshData()" internal/app/callbacks.go` must be empty, and
     `grep -n "\bc\." internal/app/refresh.go internal/app/session_actions.go internal/app/preview.go`
     must be empty;
   - `go test -race -run TestCommandClosuresDoNotRaceCoordinator ./internal/app` is a
     runtime backstop on top of the greps and the structure, not a proof. It runs the
     commands each message returns on their own goroutines while the test goroutine writes
     coordinator state through `Update`. It reports a read only if a write lands before the
     test receives that closure's result, so it catches reads made early in a closure
     (fast ones included) but can miss one made late, after other work or a delay. Its
     can-fail check is documented in the test's header: plant
     `if c.settings.IDECommand == "planted" {…}` in a closure wrapping `launchIDE`'s
     command (the bug plan 4 fixed) and the test must fail. Use the value in a condition;
     a bare `_ = c.settings.X` is a dead load the compiler can drop. It only means
     something under `-race`; redo the check if you change the harness.
2. **Never navigate from inside a view's `Update`.** `Coordinator.Update` writes the view
   returned by `updateCurrentView` back into `c.views[c.currentView]` *after* the view's
   `Update` returns. A callback that calls `navigateTo` during that `Update` changes
   `c.currentView` first, and the old view is then stored under the new view's key —
   this once replaced the dashboard with the agent-switch view. Navigate from a message
   handler in `handleSpecialMessages` / `handleSessionMessages`; those return early and
   never write `c.views`. The full explanation is the doc comment on `navigateTo`
   (`navigation.go`).
3. **`DataLoadedMsg` is always a full snapshot.** A `nil` field means "that registry could
   not be read this round — keep what you have", never "this update doesn't concern that
   field". Partial updates have their own messages (`PreviewDataMsg`,
   `SessionsRefreshedMsg`). Every `SetData` guards each field with `if msg.X != nil`.
   This matters because `CapabilityList.SelectedIDs()` only returns IDs whose items are
   currently loaded: wipe the items and saving the card silently drops its skills.
4. **The error banner is cleared by the next keypress or `ClearErrorMsg`; navigation
   leaves it alone.** `Coordinator.Update` clears `c.err` on any `tea.KeyPressMsg` before
   routing it. That is safe because an error always arrives as a *later* message than the
   keypress that caused it. So an `ErrorMsg` batched beside a `NavigateMsg` survives in
   either delivery order (`TestErrorSurvivesNavigationInEitherOrder`).
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
   **A component that is given a size keeps its elements in place.** Where an element
   sits (its row, and its column from the box edge) depends on the size the component was
   given, never on what it happens to show: which page, how far it scrolled, what the
   search matched. In practice: pad a list to the width it was given (Bubbles pads only to
   its widest line) and join pieces `lipgloss.Left`, not `Center`; reserve the rows of
   optional elements (the picker's search row, a capability list's `...` rows) and fill
   them with plain empty lines. The pattern for any list-like component is
   `TestWorkspacePickerKeepsColumnsAcrossPages` (column offsets across pages) and
   `TestCapabilityListItemsKeepTheirRows` / `TestCardEditorListsKeepTheirPlace` (rows
    across scroll and search). A sized `CapabilityList` renders exactly its `Height` lines
    (no trailing newline); an unsized one keeps the old content-driven layout.
10. **Conversation bindings are session-local, while runtime registration is activation-local.**
    Persist bindings only in that session's JSON under the associated card ID; deep-copy
    the map and slices at snapshot boundaries. The manager authenticates reports from a
    rotating per-activation bearer token and derives session/card/driver/generation from
    that credential. Never trust caller-supplied Kankuro IDs or persist the credential.
    Stage reports that race spawn publication, commit them before the initial session
    write, and reject stale or conflicting native IDs without replacing saved bindings.
    A failed metadata write leaves the live binding in memory as `unsaved`; a later
    successful write repairs that state. Temporary workspaces isolate MCP config and
    credentials. Sessions deliberately sharing a non-temporary workspace accept a
    shared config and identity: the last successful write determines selected servers
    and the credential used by subsequent client calls. Older cleanup must not remove
    a newer credential. This is not a security boundary against arbitrary processes
    running as the same OS user. Conversation selection
    happens before reservation; the manager revalidates the explicit new/resume choice
    under reservation before spawn. All existing activations keep saved workspace
    identity, not edited destination-card workspace flags.
11. **Dashboard discovery is activation-authorized.** A card's
    `dashboard_control` setting is deployed only when that activation selected
    Kankuro MCP; relaunch preserves the deployed grant, while an explicit switch
    applies the destination card's current setting. Discovery requires a published,
    live activation. Caller-supplied IDs are never proof of identity.
12. **MCP lifecycle retries do not repeat effects.** Intent is atomically persisted
    before queueing; identical caller/request/fingerprint calls replay the same
    operation, while changed arguments conflict. A crash can occur between process
    or filesystem effects and the final journal write, so recovery reports
    `interrupted` with unknown effects and never blindly retries. Credentials are
    runtime-only and are never written into operation records.

---

## 7. Build, test, lint, run

Feature 4 integration acceptance at `7695a4c`: independent delegated build, vet,
normal/race suites, lint, and focused repeated regressions passed. The prior card
deletion blocker is resolved; repeated MCP cleanup tests did not reproduce the
shutdown timeout. Evidence: `.kankuro/review-feature4-current/verification.md`.
The dated implementation/verification entries below preserve earlier history.

```bash
go build ./...
go vet ./...
go test ./...
go test -race ./...                   # must stay clean, including root integration tests
golangci-lint run                     # must exit 0
```

The Feature 4 closeout checks passed on 2026-10-24: `go build ./...`, `go vet ./...`,
`go test ./...`, `go test -race ./...`, and `golangci-lint run` (0 issues; existing
configuration/deprecation warnings only). `git diff --check` passed too. Run all five
before calling a later change done.

The 2026-10-24 `review_5.md` visual corrections also passed those checks after the
final source edits. The temporary review overlay regressions passed, and real
`tui-test` PNGs were captured and inspected at 80×24, 120×40, and 160×45 in
`/tmp/opencode/kankuro-review5-visual/artifacts/`. Preview art now uses a reserved
text/art split and is omitted when it cannot fit intact; card rows use balanced
gutters, and dashboard pane titles/content share stable cell baselines. An earlier
race run intermittently failed `TestStreamableMCPRegistrationAndAuthentication`
(`server_test.go:131`, shutdown MCP context deadline exceeded); five isolated
repetitions and complete reruns passed, including the final complete race run.
Cause remains undiagnosed; do not suppress cleanup errors.

The targeted `review_6.md` follow-up removed the empty preview's obsolete border
width deduction: the dashboard already sizes `SessionPreview` to its content slot.
The placeholder now word-wraps without embedded indentation and is centered within
that supplied size. Its 8×19 regression preserves complete `session` and `selected`
words; empty-dashboard readability/centering is checked at 80×24, 120×40, and
160×45. Final checks passed again: `go test ./...`, `go test -race ./...`,
`go vet ./...`, `go build ./...`, `golangci-lint run` (0 issues), and
`git diff --check`. Fresh populated and empty `tui-test` captures were opened and
inspected under `/tmp/opencode/kankuro-review6-final/artifacts/` and
`/tmp/opencode/kankuro-review6-empty/artifacts/`; both isolated TUI runs exited
normally. The intermittent MCP shutdown timeout did not recur in this verification;
its previously recorded undiagnosed status remains.

The `review_7.md` integration correction adds manager-owned card-reference
reservations shared by separate UI and MCP service instances. Deterministic gated
tests cover MCP create and switch against UI deletion, deletion-first with a stale
card snapshot, cancellation/release, relaunch after create, and continued reference
protection after termination. The focused lifecycle tests pass normally and under
`-race`.

MCP shutdown investigation: before test cleanup changes, the focused registration
test failed twice in 20 normal repetitions at its 5-second shutdown deadline; the
full race suite also failed `TestRealServiceLifecycleResultsOverMCP` during
shutdown. MCP integration tests now own dedicated HTTP transports, close client
sessions before the server, and check close errors. Afterwards, the focused test
passed 20/20 normally and 20/20 with `-race`, the MCP package passed five repeated
race runs, and final full normal/race suites passed. This addresses observed client
cleanup ownership; no production root cause is claimed. Logs are retained in
`/tmp/opencode/kankuro-review7/`. The implementation is ready for another Feature 4
integration review; Feature 5 remains blocked and acceptance is not claimed.

The `review_8.md` verification correction makes the operation-queue test shut down
and join its capacity-phase runner before starting the separately gated publication
phase. The queue-full rejection and late-cancellation/publication assertions remain.
`GOMAXPROCS=1 go test -race ./internal/api -run
'^TestOperationQueueBoundAndDisconnectedWaitReturnPromptly$' -count=50` passed
50/50; full `go test ./...` and `go test -race ./...` passed, along with vet, build,
lint (0 issues), and `git diff --check`. Feature 4 is ready for another integration
review, not accepted; Feature 5 remains blocked.

**Running the app by hand** — always isolate it from your real config
(`kankuro_specs.md` §5.1):

```bash
KANKURO_TMP_DIR=/tmp/kankuro-test/tmp \
KANKURO_CONFIG_DIR=/tmp/kankuro-test/config \
  go run . --debug
```

For local development, use `go run .`, `go build .`, or `go install .`. A Go
installation places the executable in `GOBIN`, or the default Go bin directory
when `GOBIN` is unset; that directory must be on `PATH` to invoke `kankuro`.
Public installation is `go install github.com/BlackGiack/kankuro@latest`.

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
- Shared operations are tested in `internal/api` without constructing a coordinator;
  lifecycle tests should continue to use `newTestManager` so shutdown and leak
  guards are registered.
- Relaunch preservation across restart is covered in `internal/session/workspace_test.go`;
  `relaunch_mcp_test.go` additionally verifies fresh Kankuro credentials, preserved
  ordinary MCP configuration and manual instructions, and retry after invalid JSON.
- `internal/session/workspace_validation_test.go` covers missing and non-directory
  saved paths on relaunch and card switch in original and temporary workspaces:
  refusal before spawn, no path recreation or replacement, and successful retry
  with saved work intact after restoration.
- Conversation binding persistence, early registration staging, revocation, and retry are
  exercised in `internal/session/conversation_test.go`; it also calls the installed
  `mcp-cli` client from the agent's child shell against `.kankuro/mcp-servers.json`.
  Workspace MCP serialization/conditional cleanup is tested in
  `internal/driver/mcp_config_test.go`, with shared-workspace interleavings and natural
  exit/write-failure cases in `internal/session/workspace_mcp_*_test.go`. Transport
  handshake/auth/origin tests are in `internal/mcp/server_test.go`; picker and registration layout tests are in
  `internal/ui/views/conversation_choice_test.go` and
  `internal/ui/components/session_list_layout_test.go`.
- `command_race_internal_test.go` is the runtime check for invariant 1: it executes the
  commands a message returns concurrently with `Update` writes, so it needs `-race`.
- `collectMsgs` (two copies: `session_warning_test.go` and
  `refresh_cadence_internal_test.go`) unwraps `tea.BatchMsg` only; commands built with
  `tea.Sequence` come back opaque.
- **User text goes into a fixed slot through `components/text.go`.** `DisplayText(s)` makes it
  one-line safe (tabs and line breaks become a space; other control bytes and ANSI escapes are
  dropped) and `TruncateHead(s, w)` / `TruncateTail(s, w)` fit it to `w` display columns, keeping
  the start / the end and marking the cut with `...`. Use `TruncateTail` for paths and for text
  being typed. Never slice a string by byte for display, and sanitise at display time: the stored
  text stays what the user typed. A tab left in is expanded by lipgloss after the text was fitted,
  which is how a name overflowed its row.
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
  `card_editor_layout_test.go`, `modals_layout_test.go`. The 12 layout bugs
  these sweeps found are all fixed (`impl_plan_5.md`); a new one found this way is a
  `t.Skip("bug: ...")` test listed under "Found, not fixed" in the plan that found it. The sweeps
  use `uitest.AdversarialNames` in full (tabs included) and no longer exclude any known cause.
- Shape tests check *fit* (the render stays in its box); the position tests above
  (`components/capability_list_layout_test.go`, `TestWorkspacePicker…` in
  `modals_layout_test.go`, `TestCardEditorListsKeepTheirPlace`) check *position* (elements
  stay put as the content changes). A new list-like component needs both.
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
| IDE/terminal commands, card-editor sections | Settings view (`c`) → `app_settings.yaml` | what `o` / `t` run; whether personas / MCPs are shown in the card editor. `terminal_command` defaults empty. |
| per-card `env:` | `id_cards.yaml` | merged over the parent environment when the agent is spawned; a card key overrides an inherited one. There is no UI for it yet — edit the YAML. |
| `KANKURO_MCP_CONFIG_PATH` | legacy reserved variable | stripped from parent/card env; never set. Use `.kankuro/mcp-servers.json` from the workspace root. |
| `KANKURO_MCP_TOKEN` | legacy reserved variable | stripped from parent/card env; never set. Authentication is read from the selected workspace MCP entry. |

`config.ListMCPs` merges the built-in `kankuro` server into the selectable registry;
it is not written to `mcps.yaml`, cannot be edited/deleted or shadowed by name/ID,
and is not silently selected for existing cards. Selecting it auto-adds the `mcp-cli`
skill and places its current endpoint/credential in the workspace MCP config.

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
| Add a screen | `ui/views/`, register it in `initViews` (`navigation.go`), add a `ViewType`, wire callbacks in `app/callbacks.go` |
| Add a message | `ui/messages/messages.go`, handle it in `coordinator.go` (`handleSpecialMessages` / `handleSessionMessages`); if it carries partial data, give it its own type (invariant 3) |
| Change how cards look | `components/card_list.go` (`RenderCardItem`), `components/renderer.go` (puppets), `styles/styles.go`; invariant 9 and the sweep test |
| Add or change a puppet | built-in art in `internal/ascii/characters`; user puppets are `.puppet` text files found in puppet source dirs (`config/puppet_sources.go`, which also truncates oversized art on read) |
| Change dashboard layout or keys | `views/dashboard.go` (`footerText` for the footer), then update the key table in §4 and in `kankuro_specs.md` |
| Put user text into a fixed-width slot | `components/text.go` (`DisplayText`, `TruncateHead`, `TruncateTail`) |
| Change the minimum terminal size or the "too small" screen | `views/constants.go` (`MinTerminalWidth/Height`), `app/size_guard.go` |
| Change refresh behaviour | `refresh.go` (the commands) and `handlers.go` (applying results), `config.Fingerprint` |
| Add a session-level action (Phase 1 linking) | a method on `sessionActions` in `session_actions.go` that returns a `tea.Cmd` and takes any coordinator state *by value*, plus a message and a handler in `coordinator.go`; wire the trigger in `callbacks.go` |
| Add a shared launch/relaunch/switch/delete/remove rule | `internal/api/service.go`; keep UI message mapping in `app/session_actions.go` and callbacks in `app/callbacks.go` |
| Change native conversation state or registration | `internal/session/conversation.go` (manager-owned bindings, credentials, staging, persistence) and `internal/api/service.go` (transport-facing registration methods) |
| Change MCP protocol/security/lifetime | `internal/mcp/server.go`; `main.go` owns listener startup and shutdown order |
| Change the conversation picker | `ui/views/conversation_choice.go`, `app/coordinator.go` message routing, `app/callbacks.go` command construction |
| Change the preview subscription | `preview.go` (`previewSub`) and `handlers.go` (`handlePreviewRequest`, `handleSessionOutput`) |
| Pick up known debt | `tech_debt_act_2.md` |

---

## 11. Known quirks / pitfalls (check before "fixing")

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
- **Below 80×24 there is only the "too small" notice.** There is no degraded layout (for
  example hiding the preview pane under 80 columns and keeping cards and sessions usable);
  that would be a second layout to build and test. Possible follow-up.
- **Backspace in a search box deletes a byte, not a rune** (dashboard, card-editor lists,
  workspace picker): `query[:len(query)-1]`. Deleting the last rune of a multi-byte character
  leaves invalid UTF-8 in the query. Display is safe (`DisplayText` / `TruncateTail`); the
  editing is not fixed.
- **The `commands` package and `app/wiring.go` are gone.** Older documents still mention
  them; they were dead code.

---

## 12. Other documents in the repository

| File | What it is |
|---|---|
| `kankuro_specs.md` | Product specification: concepts, roadmap, UI and keys. §5.1 is the manual-testing setup. |
| `crow_puppet_theme.md` | The colour palette used in `ui/styles`. |
| `.leftover/` | Hand-off notes from earlier sessions. Historical. |

### Session startup trust and publication

Before spawning a new, relaunched, or switched session, the selected driver prepares
trust for the session's actual working directory, including a temporary copy. Codex
updates `~/.codex/config.toml` under `projects.<absolute path>.trust_level` to
`trusted`; Claude updates the `trustedDirectories` array in
`~/.claude/settings.json`. Generic drivers make no trust change. These are
persistent user-level trust grants for exactly that directory. Configuration is
parsed and preserved structurally; TOML comments and formatting may change. Invalid
configuration and write failures abort startup without truncating the original.
Updates use a restrictive sibling temporary file and atomic rename, retaining an
existing file's permissions. A process-local lock serializes Kankuro updates, but
other processes are not covered by that lock.

After successful spawn, publication and human attachment are immediate. The fixed
ten-second admission delay for MCP delivery is not a readiness guarantee. Registration
is independent of the timer. An exit after publication leaves a terminated session for
inspection and relaunch. Registration guidance is supplied by the deployed
`kankuro-puppet` skill when Kankuro MCP is selected. Normal instruction synchronization
removes legacy Kankuro registration blocks.

Deleting a temporary session first removes its exact saved working-directory trust
from both Codex and Claude, including grants from earlier card switches. If either
configuration cannot be updated, deletion stops and the session remains available
to retry after fixing the configuration; one CLI's cleanup may already have
succeeded. Failed new temporary activations also attempt trust cleanup and report
the exact path for manual removal if it fails. Non-temporary deletion does not
remove trust.
