# Feature 2: Native conversation continuity and minimal MCP

Status: implemented; user reports successful OpenCode and Codex registration/resume; workspace-validation correction implemented, awaiting integration review before advancement (2026-10-19).

Dependency: Feature 1 accepted.

## Intent

Make native CLI conversations survive Kankuro relaunches and card switches. Deliver
this first for human use, introducing MCP with a single registration operation.

## Scope and constraints

- Evolve the driver interface for native conversation discovery/registration and
  launch arguments. Generic discovery instructs the agent to discover its ID and
  report its ID and a meaningful name through MCP. All drivers, including the
  Claude driver, use the generic discovery/registration implementation; preserve
  existing folder-name differences. Codex has one native resume-argument override.
- Session arguments are a driver method, following the model-argument pattern.
  The generic implementation returns `--session <id>` for an existing ID and no
  conversation arguments for a new conversation. Codex uses its native
  `resume <id>` subcommand; other drivers inherit the generic arguments for now.
  Record CLI compatibility limits without adding further specialized hooks.
- Add an embedded loopback HTTP MCP endpoint controlling the existing manager.
  Expose only `register_conversation(conversation_id, name)` initially. Include caller
  authentication, Origin validation, clean startup/shutdown, and session config
  injection compatible with the existing MCP access mechanism.
- Offer Kankuro MCP by default as a selectable built-in MCP, like the built-in
  `mcp-cli` skill. Resolve its live endpoint and credentials per activation without
  requiring manual server configuration or persisting runtime secrets in cards.
- Derive the Kankuro session and activation from authenticated caller identity.
  Reject stale registrations and prevent one caller from registering another
  session's conversation. Never treat registration parameters as authorization.
- Persist multiple conversation bindings per card within each session, with driver
  identity and a display name supplied at registration. Show each name alongside
  its ID in the selection view. Names are not unique identities. Preserve
  compatibility with existing records that have no bindings.
- First launch starts directly without a conversation ID. On relaunch or card
  switch, if the destination card has any conversations in the selected session,
  show a view to select one of those IDs or a new conversation. Even one saved ID
  triggers selection. If there are none, launch directly. Starting new preserves
  previous conversations. Preserve the working directory in every case.
- Make pending registration, registered identity, and failures understandable.
  Never silently replace a failed requested resume with a fresh conversation.
- Registration is asynchronous; do not block the UI waiting for an agent report.
  Ensure fresh credentials reach resumed agents and never enter logs, cards, or
  session records. Temporary workspaces isolate configs; when sessions deliberately
  share a non-temporary workspace, the latest config/credential wins and subsequent
  MCP calls act as that credential's activation rather than providing independent
  workspace identities.

## Acceptance criteria

- On a real supported CLI, capture evidence of reliable ID discovery and registration;
  do not assume the agent knows its ID merely because it was instructed to report it.
- Launch card A, register, stop and switch to B, register, stop and return to A.
  Verify A resumes its own conversation and workspace contents remain intact.
- Restart Kankuro and repeat resume from persisted state. Exercise explicit fresh
  start and loading older session records.
- Verify zero/one/many saved conversations on both relaunch and card switch,
  cancellation without activation, and multiple conversations retained per card.
  Conversations from another Kankuro session must not appear in the selection view.
  Registered names must persist and appear alongside IDs; duplicate names must not
  merge distinct conversations.
- Verify Kankuro MCP is offered as a built-in without manual server setup and
  resolves current activation credentials when selected.
- Verify stale/unauthorized reports, missing registration, invalid resume references,
  and persistence failures. Record limitations that cannot be reliably detected.

## Planning decisions

Validate the generic conversation path and the Codex resume-argument exception. Decide
how registration instructions are delivered, how missing registration affects switching,
how conversations are added without replacing older entries, and what happens if
a card's driver changes. Confirm
which card settings are actually applied when a native conversation is resumed.
If this needs multiple execution sessions, split the implementation plan at
reviewable boundaries without weakening the end-to-end acceptance criteria.

## Out of scope

Session-control MCP tools beyond registration, task/stage records, automatic
workflow progression, and multiple dashboards.

## Working agreement

Read `project.md` for onboarding and `roadmap.md` for the agreed direction, then
inspect the current code and prerequisite handoffs before planning. This is a
feature brief, not an implementation plan. Reuse existing session concepts and
workspace preparation; user-defined workflows must remain outside the core model.

Update `project.md` during implementation with accurate entry points, behavior,
and invariants, preserving its role as an agent onboarding guide. Run the required
checks documented there and add meaningful tests for new behavior. Do not commit
unless explicitly asked.

Before dependent work starts, hand back the implementation for integration review.
Provide a concise handoff with changed contracts, verification results, known
limitations, and unresolved decisions. Update this brief's status to reflect the
actual outcome; implementation complete and integration review accepted are
separate states. Resolve review findings before starting dependent features.

## Integration review — 2026-10-19

Reviewed commit `109698d` against `58dc8dd`. The user reports successful native
conversation registration and resume with both OpenCode and Codex. This is
user-performed real-CLI validation, not a new real-agent run by the reviewer.

Independent baseline checks passed: `go build ./...`, `go vet ./...`,
`go test ./...`, `go test -race ./...`, and `golangci-lint run` (zero issues).
The installed `mcp-cli` was available for interoperability coverage. The focused
regression reproduction below fails despite the existing suite passing.

### Required correction: reject missing workspaces on relaunch

In `internal/session/lifecycle.go`, the existing-session branch of `startSession`
now takes `workspaceOverride` directly and checks only whether it is empty. It
bypasses the previous temporary-workspace existence check. Subsequent driver/MCP
setup can recreate the missing directory and start an agent in a workspace that
no longer contains the user's work. Card switching already checks existence.

Reproduced with an isolated temporary workspace and a `cat` card with no managed
instructions: launch, stop, rename the working directory aside, then relaunch.
Relaunch returned success and recreated the original path. A focused review test
asserting refusal failed; the temporary test file was removed after reproduction.

Validate that the saved working directory exists and is a directory before any
relaunch preparation mutates it. Return an actionable error rather than recreating
it or falling back to the card's workspace. Add regression coverage proving that
missing workspaces are not recreated, no agent starts, and the activation
reservation is released so retry works after the workspace is restored. Cover
both original and temporary workspace sessions.

Advancement to feature 3 awaits this correction and its review.

### Correction handoff — 2026-10-19

Relaunch and card switching now share saved-workspace validation before workspace
preparation or driver/MCP synchronization. The saved path must exist and be a
directory. Errors include the path and instructions to restore it before retrying;
no fallback workspace or replacement directory is created. Failure continues to
release the activation reservation.

Regression coverage exercises missing paths and paths replaced by regular files,
for both original and temporary workspaces, on relaunch and card switching. It
checks that no process spawn is attempted, the invalid path is untouched, the
session stays terminated, and restoring the directory allows retry with saved
work intact. The missing-workspace relaunch cases failed against the prior code
by returning success in both workspace modes.

Verification passed: `go build ./...`, `go vet ./...`, `go test ./...`,
`go test -race ./...`, `golangci-lint run` (zero issues), and `git diff --check`.
This correction is ready for integration review. No new real-agent end-to-end run
is claimed by the implementer.

### Rebase compatibility — 2026-10-21

Main's relaunch regression fix is retained: relaunch preserves deployed instructions,
skills, personas, ordinary MCP configuration, and sub workspaces, and does not read
capability registries. Switch Agent, including selecting the same card, explicitly
reapplies configuration. Feature 2's credential rotation is preserved by refreshing
only the built-in Kankuro entry on relaunch when it was deployed. Its deployment
selection is persisted without credentials, so manager restart and subsequent
card edits do not lose or silently alter that choice. Invalid deployed MCP JSON
blocks credential refresh and allows retry after repair without overwriting it.
