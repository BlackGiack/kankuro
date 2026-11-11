# KANKURO: The AI-CLI Puppet Master

KANKURO is a terminal application for configuring, running and coordinating AI
agent CLIs. Users organize agent templates and sessions into dashboards, attach
to their terminals, and use authenticated local MCP tools for agent-to-agent
communication and session control.

This file defines product behavior. Repository structure, implementation entry
points and development practices are documented in [project.md](project.md).
Installation is documented in [README.md](README.md).

## 1. Core concepts and behavior

### 1.1 Dashboards and ownership

- A dashboard is a persistent object with a stable ID, name and description.
  The application opens on a navigable dashboard list.
- Opening another dashboard or returning to the list preserves running sessions,
  their workspaces and their conversation bindings.
- Cards belong to the dashboard where they were created. An exclusive card is
  available only there; a shared card is usable from other dashboards. Only its
  owner may edit or delete it.
- Sessions belong permanently to their creation dashboard, including sessions
  launched from shared cards. Sharing a card never shares its sessions.
- Skills, personas, puppets, MCP definitions and workspace registries are global.
- Human actions use the opened dashboard. An MCP caller's scope comes from its
  authenticated session membership; human navigation does not change that scope.

### 1.2 The ID card

An ID card is a saved agent template containing:

- CLI tool and model selection.
- Workspace selected from the workspace registry.
- Instructions and optional process environment overrides.
- Whether to create a temporary workspace.
- Optional sub workspaces, selected skills, personas and MCP servers.
- Puppet/face appearance.
- Shared-card availability and Dashboard Control permission.

A shared card cannot be made exclusive while another dashboard references it
through sessions, saved conversation bindings or pending activations. The error
must explain why the change is refused.

### 1.3 Workspace and capability registries

Workspace sources may point to one workspace or to a root whose direct child
directories become selectable workspaces. Cards reference entries by stable ID,
rather than storing their own raw working-directory path.

Skills, personas and puppets are discovered from configured source directories.
MCP servers are entered directly. Selecting the built-in Kankuro MCP deploys its
runtime endpoint/credential and the built-in MCP usage and registration/reporting
skills. Selection and Dashboard Control are separate choices.

Temporary sessions copy their source workspace into a disposable directory under
the configured temporary root. The copy persists across relaunches and card
switches. This isolates workspace changes from the source; it is not an operating
system security sandbox.

Optional sub workspaces are copied into `tmp/<workspace basename>` under the
session's actual working directory. Managed copies and ownership metadata persist
across activations; occupied or mismatched destinations cause an error. Without
temporary mode, these writes occur in the original workspace.

### 1.4 Sessions and process lifetime

A session is an active or persisted instance of a card. It retains its dashboard,
workspace identity and conversation bindings when its process terminates.

- **Launch:** create a new session from an available card and prepare its selected
  workspace, capabilities and instructions.
- **Attach/detach:** attach to a live process and use Ctrl+K to return to Kankuro
  without stopping it. Attachment includes retained output and live resizing.
- **Stop:** terminate the agent's process group, escalating from SIGTERM to
  SIGKILL after the grace period. Keep the session record for later relaunch.
- **Relaunch:** restart a terminated session in its saved workspace. Preserve
  deployed files and use current card runtime settings; refresh the deployed
  Kankuro MCP endpoint/credential when applicable.
- **Switch Agent:** apply another available card, or reapply the same card, to a
  terminated session. Preserve the session's workspace identity and dashboard.
- **Remove:** confirm removal of a terminated session, deleting its record and
  temporary workspace when applicable. A running or activating session cannot
  be removed as an ordinary individual-session deletion.

Missing or non-directory saved workspace paths cause an actionable error before
activation. Kankuro must not silently recreate or replace the saved workspace.
Temporary-workspace deletion also cleans up its saved CLI trust grants; failure
must remain visible and retryable.

Successful process publication permits immediate human attachment. MCP prompt
delivery has a separate startup gate. Delivery availability does not guarantee
that an external CLI is ready to process a task.

Quitting Kankuro stops its managed processes and keeps session records. Termination
signals and terminal closure also run cleanup. Sessions restore as terminated on
restart; Kankuro does not run them through a persistent background daemon.
Keyboard exit behavior is defined in section 4.

### 1.5 Native conversation continuity

- A session can retain multiple native conversation bindings per card, including
  native ID, display name and driver identity.
- With Kankuro MCP selected, an agent discovers its actual native conversation
  ID and registers it through `register_conversation`. Registration is bound to
  the authenticated activation; stale or conflicting reports are rejected.
- When relaunching or switching to a card with saved bindings, the human chooses
  a compatible saved conversation or **New conversation**. Even one saved binding
  requires a choice. A card without bindings starts new directly.
- Starting new does not erase older bindings. Bindings persist across application
  restarts and remain local to their session/card.
- A requested resume must not silently fall back to a fresh conversation.
  Registration confirmation and metadata persistence are reported separately.
- Conversation continuity and workspace continuity are independent: choosing a
  conversation never moves the session to a different workspace.

### 1.6 Status and completion

`running`, `waiting` and `terminated` describe process observations. Waiting is a
heuristic that the live agent may be awaiting input; the UI highlights it and
uses the neutral face. Waiting, terminal silence and preview text do not prove
that a task is complete.

Task completion, questions and blockage are communicated through explicit reports.
Users and their orchestrators decide what those reports mean for their workflow;
Kankuro does not impose implementation/review stages or automatic approval.

### 1.7 Card and dashboard deletion

Only a card's owner may delete it. When other dashboards reference the card,
require explicit foreign-impact confirmation before deleting it globally.

- Preserve existing foreign sessions, workspace paths and saved bindings.
- Once deletion intent is accepted, refuse new launches or switches to that card.
  Existing live sessions may continue and be attached or stopped. A stopped
  session cannot relaunch the deleted card but can switch to an available one.
- Retain deleted identity and retryable cleanup state; deleted IDs are not reused.

Deleting a dashboard stops/removes its sessions and deletes all cards it owns,
including shared cards globally. Foreign references require explicit confirmation.
Sessions in other dashboards, their workspaces/bindings, cards owned by other
dashboards and global registries survive.

Deletion is recoverable work. Surface partial failures and retain completed
progress so the same action resumes remaining cleanup, including after restart.
A deleting dashboard exposes its status and cannot accept ordinary new work.

## 2. Local MCP control and communication

### 2.1 Authorization and session operations

A card may opt into Dashboard Control. Discovery/control requires both that grant
and Kankuro MCP on the live activation. Relaunch preserves the deployed grant;
explicit card switch applies the destination card's setting.

The local MCP exposes:

- Registration: `register_conversation`.
- Discovery: `list_cards`, `list_sessions`, `get_session`.
- Lifecycle: `create_session`, `stop_session`, `relaunch_session`,
  `switch_session_card`, `delete_session`.
- Operation inspection/cancellation: `get_operation`, `cancel_operation`.
- Messaging: `send_message`, `get_message`, `report_reply`, `get_replies`.

MCP and human operations enforce the same card availability and session membership.
A supplied dashboard/session ID or prompt header never grants authority. Shared
cards do not broaden a caller's scope. Credentials identify the current activation
and are not exposed in discovery results.

Lifecycle mutations require caller-generated UUID request IDs. Identical retries
return the original operation; changed arguments conflict. Calls may return an
operation ID after a bounded wait. Persisted operation results describe the effect
at that time; current inspection describes the session now. Cancellation is
explicit, and a successful effect may carry a persistence warning. Interrupted
operations after restart have unknown effects and are not automatically replayed.

Waiting processes are still live: stop and await termination before relaunching
or switching. User-defined workflows stay outside Kankuro's lifecycle model.

### 2.2 Task messages and explicit reports

A controller sends a short, single-line prompt with `send_message`, pinning the
target's current activation. Delivery is ordered per recipient activation and
coordinated with terminal input. Existing human drafts are not cleared, so prompt
serialization does not guarantee separate composition.

A recipient uses `report_reply` with the exact message ID and a stable request
UUID. It may report `progress`, `completed`, `question` or `blocked` without having
Dashboard Control. Reports are authenticated, correlated, durable and retrievable
through bounded `get_replies` pages. Identical report retries retain the same reply;
changed content conflicts. Missing reports remain unanswered.

Progress is informational and silent; reporting progress does not end the
worker's work. The full report may contain multiline text or artifact paths even
though injected task prompts are single-line.

Delivery states distinguish accepted, submitting, submitted, not submitted and
uncertain input. Submitted means prompt bytes and Enter were written, not that
the CLI processed the message. Partial writes or unknown outcomes must not cause
blind resubmission. Retained history is bounded; retrieval exposes cursor gaps
when older content has been removed.

### 2.3 Automatic report wake-ups

A `completed`, `blocked` or `question` report atomically saves a pending wake-up
alongside the report. Kankuro owns delivery after that save; the worker can finish
its turn without polling or retrying notification delivery. Progress creates no
wake-up obligation.

- Notify only the original task sender's exact session and activation. Never
  redirect a notification to a replacement or resumed activation.
- Stopping the reporting process does not invalidate its saved notification.
  Deleting the underlying history removes the obligation.
- Retry temporary admission failures automatically while the destination remains
  valid, with bounded backoff. Reuse the same pre-admission attempt identity.
- Once admitted, reconcile existing delivery evidence. A new attempt is permitted
  only after conclusive zero-byte delivery; submitted or uncertain input is never
  automatically replayed.
- If the original target stops or is replaced, retain the report and expose
  notification unavailability. Report persistence and notification delivery are
  separate outcomes.
- The wake-up carries the original message ID, reply ID, reporter and kind, and
  tells the orchestrator to read the full report with `get_replies`. It requires
  **no `report_reply` in response**, avoiding a report loop.
- Expose notification state through retained reports: `pending`, `accepted`,
  `submitted`, `uncertain`, `unavailable` or `not_required`. Older reports without
  notification metadata remain readable and generate no retrospective wake-up.

These wake-ups use the same bounded delivery machinery as ordinary messages.
They neither grant the reporting worker arbitrary messaging permissions nor
advance a prescribed workflow.

## 3. Supported capabilities and future ideas

Current capabilities include dashboard/card management, workspace and capability
registries, PTY attachment, saved-session relaunch, native conversation continuity,
scoped MCP operations, durable reports and automatic report wake-ups.

Output/time-based triggers, local-LLM controllers and a persistent UI notification
stack are future ideas, not currently provided behavior. They must not be inferred
from report wake-ups, waiting status or the generic session-control tools.

The product remains CLI-agnostic. A driver must support process lifecycle,
workspace deployment, native conversation arguments, terminal input/output and
prompt-submission evidence. A generic implementation supplies common behavior;
CLI specializations adapt only what differs. The authoritative interface and
extension entry points are linked from [project.md](project.md), rather than
copied into this product specification.

## 4. UI/UX design

### 4.1 Dashboard list and navigation

- Show dashboard identity, name and description. Use the established dashboard
  title styling with the `KANKURO - ` prefix and contextual keyboard help.
- `Enter` opens an active dashboard. On a deleting dashboard it opens deletion
  status instead.
- `n` creates a dashboard; `e` edits its name and description.
- Delete/Backspace starts deletion confirmation for an active dashboard or retries
  an incomplete deletion. Deletion status uses the same retry keys.
- `Esc` inside an open dashboard returns to the list without stopping sessions.
  In forms/dialogs it cancels or goes back as appropriate. Repeated Esc on the
  dashboard list does not quit.
- In list navigation, `q` quits. Ctrl+C is ignored on the dashboard list; this
  protection also applies when the terminal is too small.

### 4.2 Open dashboard

Show cards on the left, the selected session's preview in the center, and a
scrollable session list on the right. Session status and puppet faces distinguish
running, waiting and terminated processes.

| Key | Action |
|---|---|
| `Tab` | Move focus between cards and sessions. |
| `/` | Search the focused list. |
| `Enter` on a card | Launch a new session. |
| `Enter` on a session | Attach if live; relaunch if terminated, with conversation selection when needed. |
| `a` | Attach to the selected live session. |
| `n` / `e` | Create / edit a card, subject to ownership. |
| `e` on a selected session | Edit its display name; Enter saves and Esc cancels. |
| `x` | Stop the selected live session. |
| `s` | Switch the selected terminated session to an available card. |
| Delete/Backspace on a session | Confirm removal of a terminated session and its temporary workspace. |
| `o` / `t` | Open the selected session's actual workspace in its configured IDE / terminal; available for live or terminated sessions. |
| `m` / `w` / `c` | MCP configuration / workspace sources / settings. |
| `Esc` | Return to the dashboard list. |
| `q` | Quit Kankuro. |
| Ctrl+K while attached | Detach from the agent terminal. |

The footer shows applicable navigation, focused-pane and global keys. At narrower
widths, shorten or omit groups while preserving dashboard navigation and quit
where space permits. User text must fit its display slot without corrupting layout.

The minimum supported terminal size is **80×24**. Below it, replace the current
screen with a size notice and ignore ordinary action keys while continuing
background updates. `q` still quits; Ctrl+C quits outside the dashboard list.
Restore the current screen when sufficient space returns.

### 4.3 Conversation selection

When the destination card has saved bindings, show compatible conversations and
an always-available **New conversation** choice. Search by name, ID or driver;
incompatible-driver bindings cannot be selected. Long selected-conversation
details must remain inspectable without hiding activation controls or the footer.

### 4.4 Settings

- Configure independent IDE and terminal commands. The terminal command must
  open a separate terminal window; its default is empty. Kankuro does not infer
  an emulator or embed a shell.
- Both commands substitute `$PATH` with the selected session's actual working
  directory and use that directory as the process working directory. Quote
  `"$PATH"` where paths may contain spaces. Execute directly without implicit
  shell expansion.
- Toggle visibility of Personas and MCP lists in the card editor.
- Enter saves, Space changes a focused toggle, and Esc cancels/returns.

### 4.5 Card editor

- Tab cycles through Name, Driver, Model, Workspace, Sub Workspaces, Face, Tmp,
  Shared, Dashboard Control, AgentsMD and capability lists. Hidden optional
  sections are removed from focus order.
- Space on Workspace opens its searchable picker; on toggles it changes their
  value; on capability lists it toggles the highlighted selection.
- Shared controls cross-dashboard availability subject to the reference rules
  in section 1. Dashboard Control takes effect for a deployed activation only
  when Kankuro MCP is selected.
- Ctrl+X on Sub Workspaces clears selections; Enter saves the change.
- Ctrl+E on supported source/list fields opens the relevant source/server form.
- Delete on the MCP list removes a global MCP definition only when no card
  references it; built-in definitions remain protected.
- Enter saves and returns; Esc cancels and returns. A tall form scrolls to keep
  its focused field visible, with `↑ more` / `↓ more` indicators.

### 4.6 Workspace source manager

- `n` adds a source; `e` / Space edits it; `d` deletes it.
- `r` refreshes discovery from the configured sources.
- Esc returns to the dashboard.

## 5. Persistence and configuration

Configuration and application settings persist under the user config root.
Session metadata, conversation bindings, operation outcomes, message/report
history and deletion progress survive application restart. Live processes,
activation credentials and runtime registration state do not survive restart.

An empty config root initializes automatically in the current supported format.
Existing unsupported or incomplete layouts are rejected with an actionable error;
runtime does not convert them. One running application holds the configuration
root's writer lock, preventing another writer using the same root.

`KANKURO_CONFIG_DIR` / `--config-dir` select the config root;
`KANKURO_TMP_DIR` selects the temporary-workspace root, defaulting to `<config>/tmp`.
The default config root uses the platform's user configuration directory
(`~/.config/kankuro` on Linux).

Selected MCP endpoint/credential configuration lives in the session's actual
workspace. Temporary workspaces isolate that file; deliberately shared original
workspaces share it. Cleanup must not remove a newer activation's credential.
Do not store live bearer credentials in card, session or operation records.

### 5.1 Development and manual testing

Installation and local development commands are in [README.md](README.md).
Isolated manual runs, test cleanup and verification practices are in
[project.md](project.md#7-build-test-lint-run). Product specifications do not serve
as acceptance logs or a source of task-specific development instructions.
