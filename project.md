# project.md — Agent onboarding: KANKURO

Use this file to understand the project, find the relevant code and avoid breaking
its invariants. Read `AGENTS.md` for working instructions and `README.md` for
installation. Product requirements live in `kankuro_specs.md`. The interactive
code map and execution flows are in `appedia/`.

Keep this file about the current project. Put task plans, review findings,
acceptance results, investigation logs and temporary environment details in their
own task artifacts, not here.

## 1. What this project is

KANKURO is a Go terminal application for managing AI coding-agent CLI sessions.
Users organize agent templates and sessions into dashboards, launch agents in
original or temporary workspaces, and attach to their terminals. Agents can use
the local MCP endpoint for authorized session control and communication.

| Concept | Meaning | Main type/package |
|---|---|---|
| Dashboard | Persistent owner of sessions and cards; a shared card remains owned by its creation dashboard. | `config.Dashboard` |
| ID card | Saved agent template: CLI/model, workspace, capabilities, instructions, environment and appearance. | `config.IDCard` |
| Session | Process lifecycle, saved workspace identity, output and conversation bindings, with immutable dashboard membership. | `session.Session` |
| Activation | One process generation and its deployed permissions and credential. Replacing a process does not reuse its activation identity. | `session.Snapshot` |
| Dashboard Control | Per-card opt-in to control other sessions in the same dashboard when Kankuro MCP is selected; captured per activation. | `config.IDCard.DashboardControl` |
| Workspace/capability registry | Workspaces, skills, personas and puppets discovered from configured sources; MCP servers are entered directly. Cards reference registry IDs. | `internal/config` |
| Conversation binding | Native CLI conversation ID, display name and driver, saved within one session and card. | `session.ConversationBinding` |
| Message/report | Task prompt pinned to a recipient activation and its explicit correlated reports. Qualifying reports wake the original sender. | `session.MessageRecord`, `session.MessageReply` |
| Temporary workspace | Disposable copy kept across relaunches of the session and removed during session deletion. | `internal/session` |

## 2. Tech stack

- Go **1.27.3**, single module `github.com/BlackGiack/kankuro`; `go.mod` is the
  toolchain/dependency source of truth. `GOTOOLCHAIN=auto` can select that toolchain.
- Bubble Tea v2, Lip Gloss v2 and Bubbles v2 (`charm.land/*/v2`) for the TUI.
- `creack/pty` for agent terminals; ANSI-aware display-width helpers for rendering.
- YAML configuration and JSON session, operation and message records.
- Official Go MCP SDK with an authenticated loopback Streamable HTTP endpoint.
- Standard-library testing and `.golangci.yml` for lint configuration.
- Generic CLI driver plus Codex and Claude specializations. Built-in `mcp-cli`
  and `kankuro-puppet` skill bytes are embedded by the root command.

## 3. Repository map

| Location | Responsibility and useful entry points |
|---|---|
| `main.go` | Composition root: embedded skills, config-root lock, repository, drivers, manager, shared service, operation runner, MCP and TUI lifetime. |
| `internal/configstartup/` | Exclusive root lock, direct initialization of an empty root and strict current-format validation. No configuration conversion. |
| `internal/config/` | Configuration repository and registries; `dashboards.go`, `card_tombstones.go`, `dashboard_tombstones.go` and `dashboard_deletions.go` persist membership and deletion state. |
| `internal/api/` | Shared application policy for UI and MCP. `membership.go` contains dashboard-aware human operations; `persisted_scope.go` derives MCP scope; confirmation/deletion files coordinate destructive operations. |
| `internal/api/messaging.go` | Send/report APIs and `preparedDelivery`, shared intent-publication, submission and cleanup hooks. |
| `internal/api/operations.go`, `notifications.go` | Durable bounded operation runner, task authorities and report-notification dispatcher. |
| `internal/mcp/server.go` | Transport/tool adapter over the shared service, with host/origin/auth checks; owns no sessions. |
| `internal/session/` | Sole owner of processes, activation admission, credentials, workspace preparation and session persistence. Start with `manager.go`, `lifecycle.go` and `conversation.go`. |
| `internal/session/messaging.go`, `message_notifications.go` | Durable message/report store, delivery evidence, notification attempts and destination reservations. |
| `internal/session/attach.go`, `io.go`, `watch.go` | Attachment, output subscriptions, scrollback and process/status watchers. |
| `internal/session/workspace.go`, `storage.go`, `ide_launcher.go` | Workspace copies, session records and shared IDE/terminal command execution. |
| `internal/driver/` | `AgentDriver`, CLI-specific arguments, instruction/capability deployment, workspace trust, MCP configuration and serialized PTY writes. |
| `internal/app/` | Bubble Tea adapters. `coordinator.go` routes messages; `navigation.go` owns view changes; `callbacks.go` wires views. |
| `internal/app/refresh.go`, `handlers.go`, `preview.go` | Background data reads, application of snapshots and preview subscriptions. |
| `internal/app/session_actions.go`, `attach.go`, `size_guard.go` | Command adapters, raw terminal attachment and minimum-size behavior. |
| `internal/ui/messages/` | Messages exchanged between commands and the coordinator. |
| `internal/ui/views/` | Full screens, including dashboard list, dashboard, editor, settings and conversation picker. |
| `internal/ui/components/`, `component/`, `styles/` | Reusable lists, modals, text/rendering helpers, component contract and theme. |
| `internal/ui/uitest/`, `internal/testenv/` | Layout assertions and test isolation/leak checks. |
| `internal/ascii/` | Built-in puppet artwork and canvas rendering. |
| `kankuro_skills/` | Embedded MCP CLI and agent registration/reporting instructions. |

## 4. How it fits together

### UI and application policy

Bubble Tea calls `App.Update`, which delegates to the coordinator. The coordinator
handles navigation, refreshes, errors and session events, then routes remaining
messages to the current view. Views return commands and messages; they do not
navigate directly. Commands execute asynchronously, so capture their inputs on
the coordinator goroutine before constructing them.

UI commands use `api.Service` for dashboard-aware operations such as
`LaunchInDashboard`, `RelaunchInDashboard`, `SwitchInDashboard` and
`AttachInDashboard`. The manager executes the checked process operations. Keep
business and membership rules in the shared service/manager boundary so UI and
MCP cannot disagree.

Config refresh uses repository fingerprints; session snapshots refresh on the
periodic tick. Explicit mutations request a refresh immediately. Preview output
uses a subscription. Async results carry dashboard identity and view generation
so obsolete results cannot update the newly selected dashboard.

### Session lifecycle and workspace ownership

Launch resolves a card and capabilities, prepares the actual working directory,
deploys driver files and selected MCP configuration, then spawns and publishes
the process. Drivers prepare Codex/Claude workspace trust before spawn; a failure
must preserve the original user configuration. The manager owns startup timers,
watchers, output and activation teardown.

A published process can be attached immediately. MCP message delivery has a
separate startup gate; `message_delivery_available` is not proof that the CLI is
ready to process a task. Status heuristics (`running`, `waiting`, `terminated`)
are not workflow completion signals.

Relaunch reuses saved workspace identity and deployed configuration; switching
cards reapplies the destination card configuration while preserving the session's
workspace identity. Missing saved directories cause an error rather than being
recreated. Conversation selection is explicit when bindings exist and is checked
again under activation admission. Generic, Codex and Claude drivers construct
their respective resume arguments without shell interpolation.

Termination revokes the activation credential. Workspace MCP cleanup removes the
Kankuro entry only if its credential still matches that activation, preserving a
newer shared-workspace configuration. Removing a temporary session also removes
its saved workspace trust and copy; failed cleanup stays retryable. Main shuts
down the operation runner, sessions and MCP endpoint in order on application exit.

### MCP and messaging

The MCP transport delegates to the same service used by the UI. Caller identity
and deployed permissions come from the activation credential, never caller IDs or
prompt headers. Discovery/control requires a live Dashboard Control activation;
conversation registration and a worker's report have their own authorization.
MCP scope follows persisted session membership, not the selected human dashboard.

Lifecycle mutations use stable UUID request identities and a durable bounded
operation runner. An operation result is historical; inspect the session for
current state. Cancellation and post-effect persistence warnings must retain
what actually happened. Interrupted effects are never blindly replayed.

`send_message` pins a target activation. `report_reply` saves explicit correlated
reports; `completed`, `blocked` and `question` also save a pending notification to
the original sender. `progress` stays silent. The dispatcher owns notification
retries and uses the existing queue and `preparedDelivery` hooks; the reporting
worker needs no general send permission. A progress report does not end the
worker's work. Notifications request `get_replies`,
not another `report_reply`. Full report content stays in history.

Prompt writes share the driver's input gate with human terminal input. Delivery
states describe bytes written, not agent processing. Only conclusive zero-byte
evidence permits another notification attempt; submitted or uncertain input must
not be replayed. See the appedia messaging flows and the implementation for wire
schemas, retention bounds and retry scheduling details.

### Main interactions

| Context | Keys and behavior |
|---|---|
| Dashboard list | `Enter` opens; `n` creates; `e` edits; Delete/Backspace starts deletion or retries incomplete deletion. `Esc` and Ctrl+C do not quit the list; `q` quits. |
| Open dashboard | `Esc` returns to the list; pane navigation and `/` search select cards or sessions. `q` quits. |
| Card/session actions | Enter launches a card or attaches/relaunches a session; `a` attaches, `x` stops, `s` switches a stopped session, Delete/Backspace removes a stopped session. |
| Session rename | Focus a selected session and press `e`; Enter saves the name and Esc cancels. |
| Workspace commands | `o` / `t` open the selected session's actual workspace in the configured IDE / terminal. |
| Configuration | `m` / `w` / `c` open MCP configuration, workspace sources and settings. |
| Attached terminal | Ctrl+K detaches and returns control to Bubble Tea. |

The supported minimum size is 80×24. Smaller terminals show a notice and ignore
normal input while background messages keep flowing. `q` remains available;
Ctrl+C is also accepted outside the dashboard list. View sizing is centralized
through `views.CoordinatorRows`. Form-specific keys belong with their views and
the product specification rather than being duplicated here.

## 5. Storage layout

The default root is `os.UserConfigDir()/kankuro` (`~/.config/kankuro` on Linux).
`configstartup` locks the root for runtime, initializes an empty root directly and
rejects incomplete/unsupported existing data. Cards and sessions require current
schema and explicit dashboard membership.

```text
layout-version                         current-format completion marker
.dashboard-layout.lock                 config-root writer lock
dashboards.yaml                        dashboard identities and state
id_cards.yaml                          owned/shared card templates
card_tombstones.yaml                   deleted-card identities and cleanup
dashboard_tombstones.yaml              retired dashboard identities
dashboard_deletions/<id>.json           checkpointed deletion manifests
skills.yaml, skills_sources.yaml       skill registry and sources
personas.yaml, persona_sources.yaml    persona registry and sources
puppets.yaml, puppet_sources.yaml      puppet registry and sources
workspaces.yaml, workspace_sources.yaml
mcps.yaml                              configured MCP servers
app_settings.yaml                      IDE/terminal commands and editor settings
sessions/<session-id>.json             durable session/workspace/binding metadata
operations/<caller>-<request>.json      safe operation intent and outcome
messaging/history.json                 bounded messages, reports and attempts
tmp/<session-id>/                      temporary workspaces, unless overridden
```

Session JSON keeps conversation bindings and workspace identity, not live
processes or activation credentials. Sessions restore terminated after restart.
Selected MCP configuration is deployed to
`<WorkingDir>/.kankuro/mcp-servers.json`; this file can contain the current bearer
credential and uses restrictive permissions. Credentials are never stored in
cards, session records or operation journals.

Writes use `fsutil.WriteFileAtomic` (sibling temporary file, sync and rename).
Repositories and the manager provide the owning synchronization. Preserve atomic
writes, permissions and explicit persistence-failure behavior when changing storage.

## 6. Invariants — rules the code relies on

1. **Commands do not access coordinator state.** Snapshot values before building a
   `tea.Cmd`; command adapters use the repository/manager and captured values.
   `refresher` and `sessionActions` hold no coordinator pointer. Keep callbacks
   safe too. Run `TestCommandClosuresDoNotRaceCoordinator` under `-race` when
   changing command construction; the test is a backstop, not a proof.
2. **Never navigate inside a view's `Update`.** The coordinator writes the returned
   view back after `Update`; changing the current view during that call corrupts
   view ownership. Navigate through coordinator messages. See `navigateTo`.
3. **`DataLoadedMsg` is a full snapshot.** A nil registry means its read failed and
   the view must retain existing data. Partial updates use their own message
   types. Losing loaded capability items can silently lose selected IDs on save.
4. **Coordinator errors clear on the next keypress or `ClearErrorMsg`.** Navigation
   does not clear them. Report asynchronous failures through messages so the
   triggering key cannot erase the error immediately.
5. **The editor overlays the loaded card.** `CardEditor.Card()` assigns edited
   fields on the card loaded by `SetCard`, preserving fields absent from the UI.
   Reset loaded state for new-card creation.
6. **Registry read-modify-write is one operation.** Use `Update*Sources(fn)` rather
   than separate list/save calls. A successful save followed by failed sync
   requires a refresh and warning (`config.SyncFailedError`).
7. **Respect repository lock ownership.** Use the repository's lock-owning
   operations; internal callers already holding the lock use the appropriate
   `*Locked` helpers. Never reenter a non-reentrant mutex through a locking method.
8. **A usable activation survives a metadata warning.** Post-spawn record failure
   uses `session.ErrPersistFailed`; callers preserve the activation and offer
   attachment while reporting the persistence problem.
9. **Rendering respects size and stable positions.** Measure display columns,
   sanitize at rendering boundaries and never truncate text by byte count. Clip
   artwork. Sized lists reserve optional rows and keep elements at the same
   coordinates across paging/search; test shape and position separately.
10. **Bindings are session/card-local; credentials are activation-local.** Deep-copy
    bindings at snapshot boundaries. Derive registration identity from the
    credential, stage registration during publication, reject stale/conflicting
    identities and report metadata durability separately. Relaunch preserves
    saved workspace identity and deployed state.
11. **Discovery/control requires deployed authority.** Relaunch preserves the
    activation's deployed Dashboard Control choice; explicit switch applies the
    destination configuration. Sharing a workspace also shares its deployed MCP
    file; credential-matched cleanup must not remove a newer entry. This is not
    isolation against processes running as the same OS user.
12. **Retries do not duplicate effects.** Reuse stable request identities and
    fingerprints. Recover uncertain process/filesystem/input effects as unknown
    rather than repeating them. Never persist bearer credentials in journals.
13. **Dashboard membership and card ownership do not move.** Sharing exposes a
    card, not its sessions or edit/delete rights. Only its creation dashboard can
    edit/delete it; confirmed deletion removes availability globally while
    preserving foreign sessions and their saved bindings. Do not reuse deleted IDs.
14. **Reports and wake-ups share durable intent.** Save a qualifying report and
    pending notification atomically. Its authority permits only the original
    sender activation and survives reporter termination, not history deletion
    or target replacement. A wake-up is not workflow approval or processing proof.
15. **Deletion closes admission before cleanup.** The manager owns reference and
    activation admission; repository writes remain short. Do not hold the
    repository lock while entering manager admission. Persist deletion intent
    before draining activations, checkpoint each cleanup effect and retain
    retryable state until complete.

## 7. Build, test, lint, run

From the repository root, use the checks appropriate to the change:

```bash
go build ./...
go vet ./...
go test ./...
go test -race ./...
golangci-lint run
git diff --check
```

Run focused tests first. Lifecycle, command and messaging concurrency changes
need race coverage. Documentation-only changes need reference/content validation,
not another application test run. Do not hide failing cleanup or test errors.

For a manual run, isolate both configuration and temporary workspaces:

```bash
KANKURO_TMP_DIR=/tmp/kankuro-test/tmp \
KANKURO_CONFIG_DIR=/tmp/kankuro-test/config \
  go run . --debug
```

Use a dedicated config directory; an empty one initializes automatically. Avoid
writing to real user configuration when testing. Workspace trust preparation
writes user-level CLI configuration, so native-agent experiments also need an
isolated HOME. Local development uses `go run .`, `go build .` or `go install .`;
installation and PATH setup are documented in `README.md`.

Testing practices:

- Native-agent E2E tests live in `tests/e2e` behind the `e2e` build tag. Before
  model-driven runs, ask the user which provider/model to use and confirm a
  total token budget covering the intended runs and retries. Explain estimated
  input/output usage first, report measured usage when available, and label
  estimates as estimates. The initial full-journey estimate is 20k–100k input
  and 1k–5k output tokens per successful run; tool context, resume and retries
  can change it. A saved default is not standing approval, while an existing
  explicit approval covering the current scope should be honored. See
  `tests/e2e/AGENTS.md` and `tests/e2e/README.md`. Ordinary tests and the
  model-free screenshot preflight do not require model/budget confirmation.
- Reuse the package's lifecycle fixtures. Session/app tests provide
  `newTestManager`; API tests provide `newServiceFixture` and messaging fixtures.
  If constructing a manager or runner directly, register shutdown cleanup and
  ensure it finishes before temporary directories disappear. `internal/testenv`
  supplies isolation/leak guards where wired by the package's `TestMain`.
- PTY/process tests must clean up process groups even when assertions fail.
  Native CLI configuration and built-in skill fixtures need isolated state and
  cleanup; tests that mutate shared built-in content must not run in parallel.
- Drive coordinator behavior through `Update`, then inspect views/state. The
  constructor is `app.NewCoordinator(repo, manager, service)`, with the shared
  `api.Service`. Calling callbacks alone bypasses routing and view write-back.
- Test shared policy in `internal/api` without constructing a TUI. MCP tests
  exercise the transport over the real service; manager tests own process and
  persistence behavior. Close MCP clients/transports before their server.
- Use `internal/ui/uitest` for exact or bounded rectangles and region assertions.
  Test component output before a parent's clamp, and include adversarial text.
  A list-like component needs position checks as well as fit checks.
- Make regressions demonstrate the failure they prevent. Keep detailed harness
  limitations and fault-injection recipes beside the relevant tests.

## 8. Configuration

| Setting | Meaning |
|---|---|
| `KANKURO_CONFIG_DIR` / `--config-dir` | Configuration root. |
| `KANKURO_TMP_DIR` | Temporary workspace root; defaults to `<config>/tmp`. |
| `--debug` | Log to `<user cache>/kankuro/kankuro.log`. App/session output must not corrupt the live TUI. |
| `app_settings.yaml` | Independent IDE/terminal commands and visible editor sections; edited through Settings. |
| Per-card `env` | Overrides inherited process environment; configured in card YAML. |

The built-in Kankuro MCP is selectable, protected from registry shadowing and not
automatically selected for existing cards. Selecting it deploys the current
endpoint/credential and the built-in MCP CLI and puppet skills. Read those skills
for the agent-facing registration and reporting protocol.

`KANKURO_MCP_CONFIG_PATH` and `KANKURO_MCP_TOKEN` are reserved names stripped from
inherited/card environment, not supported configuration inputs. The deployed
workspace MCP file is the source of endpoint/authentication configuration.

## 9. Conventions & workflow

- Never commit unless explicitly asked; follow the current `AGENTS.md` and task
  instructions. Do not infer a branch or merge policy from old task artifacts.
- Keep business rules in the shared service/manager boundary and UI adaptation
  in the coordinator/command layer. Reuse existing components and delivery paths.
- Match surrounding error handling: contextual wrapping, `errors.Is` / `errors.As`
  and safe user-facing errors. Do not print from live app/session code; use the
  injected logger or UI messages. Main owns startup/shutdown diagnostics.
- Comments explain constraints and non-obvious decisions. Update or remove them
  when behavior changes; do not retain task-status or review-history comments.
- Lint suppressions require a reason. Keep changes focused and preserve unrelated
  user work, including uncommitted changes.

## 10. Where to start, by task

| I want to… | Start here |
|---|---|
| Add an agent CLI or change resume arguments | `internal/driver/driver.go`, `registry.go`, `factory.go`; specialize the generic driver only where needed. |
| Change workspace deployment or trust | `internal/driver/`, `internal/session/lifecycle.go`, `workspace.go`; preserve relaunch versus switch behavior. |
| Change lifecycle or conversation selection | `internal/api/membership.go`, `internal/session/lifecycle.go`, `conversation.go`; the picker is `internal/ui/views/conversation_choice.go`. |
| Change dashboard ownership or deletion | `internal/api/membership.go`, confirmation/deletion files, `internal/session/dashboard_admission.go`, `internal/config/dashboard_deletions.go`. |
| Change configuration startup | `internal/configstartup/`; initialization and validation run before the repository and runtime. |
| Add a card field or registry | `internal/config/types.go`, validation and repository files; `internal/ui/views/card_editor.go` for editable card fields. |
| Add a view or UI message | `internal/ui/views/`, `messages/messages.go`, `internal/app/navigation.go`, `callbacks.go` and coordinator handlers. |
| Change dashboard keys or layout | `internal/ui/views/dashboard_list.go`, `dashboard.go`, `internal/app/size_guard.go`; keep product key documentation current. |
| Change text, lists or puppet rendering | `internal/ui/components/text.go`, list/rendering components, `internal/ui/styles/`, `internal/ascii/`; preserve layout contracts. |
| Change refresh or preview | `internal/app/refresh.go`, `handlers.go`, `preview.go`; capture scope/generation and keep commands independent of coordinator state. |
| Add a session-level UI action | `internal/app/session_actions.go` and `callbacks.go`; delegate policy to `api.Service` and return messages. |
| Change MCP tools or security | `internal/mcp/server.go`, shared API methods and activation authorization; `main.go` owns endpoint lifetime. |
| Change message admission or report kinds | `internal/api/messaging.go`, `internal/session/messaging.go`, `internal/driver/driver.go`, deployed puppet skill. |
| Change notification retries or authorization | `internal/api/notifications.go`, `operations.go`, `internal/session/message_notifications.go`; reuse `preparedDelivery`. |
| Change attachment or workspace commands | `internal/app/attach.go`, `internal/session/attach.go`, `ide_launcher.go`. |

## 11. Known quirks / pitfalls

- Process shutdown reaches the agent's process group. A descendant that creates
  its own session/process group can escape it; do not assume all descendants died.
- Several search fields remove a byte on Backspace rather than a rune. Display
  sanitization does not make that editing Unicode-correct.
- Serialized prompt injection does not clear a human draft already in the CLI
  input buffer. Partial writes can leave residue; respect uncertain delivery.
- Non-temporary sessions sharing a workspace intentionally share its deployed
  MCP configuration. Temporary workspaces isolate those files.
- MCP form key/value lists use commas with backslash escaping. Keep parser and
  serialization expectations aligned when changing that format.
- The minimum-size notice is the only supported UI below 80×24; individual view
  layouts there are not a second supported mode.

## 12. Other documents

| Document | Purpose |
|---|---|
| `AGENTS.md` | Current working instructions. |
| `README.md` | Installation and local development entry points. |
| `kankuro_specs.md` | Product behavior and interactions. |
| `appedia/index.html` | Interactive package/type map, execution flows, rules and task entry points; open directly in a browser. |
| `crow_puppet_theme.md` | Terminal theme specification. |
| `kankuro_skills/` | Embedded agent-facing MCP usage and registration/reporting instructions. |

Feature plans, review files and handoffs describe their own tasks. They are not
current project-wide policy or proof that a later checkout passes validation.
