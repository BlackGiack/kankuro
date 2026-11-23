# Roadmap: separate the terminal UI from application and backend implementations

Status: agreed architectural direction; implementation has not started.

This document is a handoff for a planning and review agent. That agent will turn
the roadmap into independently implementable `feature.md` files and perform the
final review of each feature. This roadmap does not authorize unrelated functionality changes.

## 1. Objective and scope

Make the terminal frontend independent of configuration persistence, session
management and CLI driver implementations. Application contracts should describe
what the frontend can request and observe; presentation models should describe
what screens and components need to display and edit.

The motivation is maintainability and safe evolution. Adding a backend field,
changing persistence representation, or extending session lifecycle rules should
not require views to understand the implementation change. Changes to navigation,
forms and terminal rendering should not affect application policy.

The frontend includes the current `internal/app`, not only `internal/ui/views`
and `internal/ui/components`. Moving backend dependencies from components into
the coordinator would not complete this separation.

The target is to retire `internal/app` after its responsibilities have been
redistributed. Its frontend responsibilities move into `internal/ui`; its
application responsibilities move behind application operations. Keep the
application working throughout the migration.

### Included product change: dashboard center replacement

Include `feature-dbcenter-rework.md` as a dedicated vertical slice in this roadmap
(step 6). Replace the live terminal preview with reports authored by the selected
session, newest first. If it has no retained reports, show workspace Git changes;
if the workspace is not a Git repository, show a detailed directory listing.
Display only the newest report content that fits the center panel, with no report
paging or history scrolling. Stopping a session leaves its displayed reports
visible, and relaunching it preserves report continuity. Existing durable report
history can repopulate the panel after Kankuro restarts; no separate persisted
cache of rendered panel content is required.

Do not improve or redesign the old preview as an intermediate architecture.
Keep it working only until its replacement lands, then remove its obsolete UI
component, subscriptions, messages, refresh plumbing and tests. Preserve terminal
output, buffering and subscriptions that are still required for attachment or
other supported consumers. Temporary preview dependency exceptions belong to
step 6 and must be removed there, before retiring `internal/app`.

### Non-goals

- No visual redesign or new product features beyond the dashboard center
  replacement explicitly included above.
- No remote backend, network protocol, services, event bus or plugin framework.
- No nested Go modules or additional `go.mod` files.
- No storage schema migration or change to existing MCP wire contracts merely to
  simplify the UI migration.
- No wholesale rewrite of session management, repositories, drivers or the MCP
  adapter.
- No independent frontend implementation of application rules such as session
  relaunch eligibility, card ownership/edit permissions or conversation
  compatibility. The application service owns these rules; the UI consumes their
  results. Frontend interaction rules remain in the UI, as detailed under
  "Action availability and operation outcomes" below.

## 2. Current architecture and evidence

Read `AGENTS.md` and `project.md` first and recheck the code at implementation
time. Paths below identify current responsibilities, not immutable file layouts.

| Location | Current behavior relevant to this roadmap |
|---|---|
| `main.go` | Composition root: embeds skills, initializes configuration, constructs drivers, repository, session manager, application service, MCP and Bubble Tea; owns shutdown ordering. |
| `internal/api/service.go` | Shared application policy, but public results still expose configuration/session types. |
| `internal/api/membership.go` | Dashboard-scoped human operations, card editing checks and dashboard inventory. Card updates compare a full expected card with current storage. |
| `internal/app/coordinator.go` | Owns frontend state and routing, but also retains a repository and concrete session manager. |
| `internal/app/callbacks.go` | Wires views to application operations and also performs direct configuration repository operations. |
| `internal/app/scoped_refresh.go` | Dashboard cards and sessions already pass through `ReadDashboardData`. Preserve this scoped application boundary. |
| `internal/app/refresh.go` | Reads global registries and configuration fingerprints directly from the repository. |
| `internal/app/session_actions.go` | Adapts operations into commands; also interprets session errors, deletes MCP entries and executes workspace commands using manager data. |
| `internal/app/attach.go` | Handles human terminal attachment, but consumes a concrete session handle and its PTY. |
| `internal/ui/messages/messages.go` | UI events and refresh results include backend models and an attachment handle. |
| `internal/ui/views/card_editor.go` | Keeps a complete loaded `config.IDCard` and overlays form values, preserving unexposed and hidden fields. |
| `internal/ui/components/card_list.go` | Joins full cards with skills/personas/puppets to construct display content. |
| `internal/ui/components/session_list.go` | Displays a broad session snapshot, joins card metadata and searches conversation bindings. |
| `internal/ui/views/dashboard.go` | Interprets lifecycle/deletion state to determine actions and confirmation behavior. |
| `internal/ui/views/conversation_choice.go` | Determines conversation compatibility by comparing binding driver with selected driver. |
| `internal/ui/views/session_launcher.go` | Interprets `session.ErrPersistFailed` to keep a successfully started process usable after a metadata warning. |
| `internal/ui/views/mcp_manager.go` | Checks the built-in MCP identity to restrict editing/deletion. |

Views and components do not currently hold repositories or managers directly.
Most imports are used for data, constants and result interpretation. The problem
is broader than direct service access: backend data shape and policy are part of
the frontend's effective contract.

Some current behavior already provides important protection. In particular,
dashboard-scoped operations, stale-result filtering, editor field preservation,
serialized input and post-effect warnings must survive the migration.

## 3. Target ownership and dependency rules

### Package responsibilities

Proposed structure, to be refined in feature planning:

```text
main.go                         composition and process lifetime
internal/
  api/
    contracts/                  application inputs, read results, outcomes,
                                resource contracts and shared semantic values
    ...                         application policy and backend adapters
  config/                       configuration storage and discovery
  session/                      process/activation/workspace ownership
  driver/                       CLI integrations and serialized process input
  ui/
    model.go                    root Bubble Tea model and coordination
    navigation.go               frontend navigation and view ownership
    commands.go                 application calls adapted to tea.Cmd
    ports.go                    consumer-owned capability interfaces
    mapping.go                  application results to presentation models
    models/                     rows, form drafts and presentation data
    messages/                   frontend events and asynchronous results
    views/                      screens and interaction state
    components/                 reusable rendering and interaction
    terminal/                   raw human terminal interaction during attachment
```

Do not create a package for every file or helper. `terminal` is justified by its
terminal I/O responsibility; command/mapping files can stay in the root UI package
unless independent package boundaries provide a demonstrated benefit.

`internal/api/contracts` is the default location for stable application-facing
types. Do not rename the whole `api` package as part of this work without a
specific need. Avoid creating two competing public application APIs.

### Permitted direction

```text
main -> ui and concrete backend construction
ui -> api/contracts
ui -> its own models/messages/views/components/terminal packages
api -> api/contracts, config, session and other backend dependencies
mcp -> api and application contracts as appropriate
```

The frontend production code must not import `config`, `configstartup`, `session`,
`driver`, `mcp`, or the concrete `internal/api` implementation package. Contracts
must not import any frontend package, Bubble Tea, Lip Gloss, repositories,
managers or drivers. Backend production code must not import frontend packages
or terminal presentation frameworks.

These rules apply to indirect model leakage too: no aliases to backend models,
embedded snapshots, concrete handles, or methods returning implementation-owned
types inside otherwise neutral contracts. Inspect exported fields and signatures,
not just interface declarations.

Pure rendering helpers such as artwork utilities may remain shared if they carry
no backend implementation dependency. Verify the actual dependency graph rather
than applying a blanket prohibition to every package outside `ui`.

The composition root is intentionally exempt from frontend import restrictions.
It must still construct and wire concrete implementations. It should no longer
inject a repository or manager directly into the frontend constructor. Expose a
small frontend notification entry point for runtime failures if needed, so main
does not construct implementation-specific UI messages to report MCP failure.

Within the frontend, root coordination may import child packages. Views and
components must not import the root coordinator package. Put shared presentation
types and events in child packages to avoid cycles.

### Interfaces and implicit implementation

Define a small set of cohesive, consumer-owned interfaces where the frontend
calls application capabilities. Examples are dashboard reading, card editing,
session actions and configuration management. Avoid one enormous `Backend`
interface and avoid an interface for every trivial method.

The concrete service should satisfy these interfaces implicitly. `main` passes
the service into the UI constructor; the application service does not import the
frontend to declare conformance. Helpers can accept narrower interfaces than the
root constructor needs. Views and components normally receive models and
callbacks/intents, not services.

Illustrative signatures only; settle names and fields during planning:

```go
type cardEditor interface {
    LoadCardEditor(context.Context, string, string) (contracts.CardEditorData, error)
    SaveCard(context.Context, contracts.SaveCardRequest) (contracts.CardSaveResult, error)
}
```

Implicit interfaces do not remove data coupling. Return and argument types must
also use independent application contracts. Go requires exactly matching method
signatures: a method returning `*session.AttachHandle` does not implement a method
returning an attachment interface. Deliberately change or adapt that method.

Use packages in the existing Go module. Directory nesting does not enforce
architectural ownership; automated import checks must enforce the chosen rules.

## 4. Application contracts versus presentation models

Application contracts describe operations, facts, identity, revision and policy
outcomes. Presentation models describe the data and state useful to a screen.
Storage structs describe persistence. They are related but have different reasons
to change.

| Concern | Application contract | UI presentation model |
|---|---|---|
| Dashboard inventory | Dashboard identity, details, application lifecycle facts and available actions | Dashboard row, display status, help/action state |
| Session inventory | Scoped session summary, registration facts, active conversation summary and available actions | Session row with resolved card display data and presentation fields |
| Card inventory | Scoped card summary, capability references/details and ownership/action facts | Card row with resolved capability names, appearance/artwork reference and action state |
| Card editing | Editable fields, revision, edit permissions, available options; explicit changes on save | Form draft, selected options, focus, modal state, local validation feedback |
| Conversation selection | Options, identity, compatibility and reason; explicit new/resume request | Sorted/filtered choices, selection, detail scrolling and help text |
| MCP management | Summary/actions and separate editable configuration | List row and form draft; command arguments/maps can have text-entry representations |
| Settings | Small settings read/write contract | Settings form state; fields may initially closely resemble the contract |
| Launch/relaunch/switch | Structured effect, identity, warning and failure information | Navigation/attachment decision and displayed warning/error |
| Dashboard center | Recent session-authored reports or workspace Git/directory entries, read status and warnings | Report/file rows, timestamp formatting, available-space clipping, loading/stale/error indicators |
| Attachment | Authorized stream/input/resize/close capability | Terminal mode, detach key and terminal output interaction |

Do not pass terminal styles, width-dependent text, focus, keyboard shortcuts or
Bubble Tea values into application contracts. Keep escaping, display sanitization,
clipping, colors and rendering at the frontend boundary. Do not pretruncate data
in application results.

### Session models

The session manager's snapshot includes fields such as instruction ownership,
activation state, delivery readiness, workspace metadata and all card-local
conversation bindings. The session list should not receive this complete object.

Define a narrow summary sufficient for the existing non-preview screen. Preserve
the currently displayed PID, workspace identity, registration information and
active conversation. Resolve the relevant conversation rather than making a
component search the manager's map representation. Resolve card names/appearance
in mapping code rather than giving a row renderer the whole registry.

If another workflow needs more data, give it a focused read contract; do not grow
the session row back into a universal snapshot. Preserve existing searching and
sorting behavior, including identity-based selection and deleted-card display.

### Card editing and revision safety

The existing full-card overlay is necessary under the current contract. Replacing
it with a newly constructed card would lose data. Replace the contract before
removing the overlay.

Separate editor reads, submitted changes and stored cards. On save the application
must load the current record, verify scope/ownership and expected revision, apply
only intended changes, and preserve other fields. The UI must not carry
timestamps, immutable ownership or uneditable fields merely to round-trip them.

Specify presence semantics for every optional update. A hidden section means
unchanged; a submitted empty list means clear. Ordinary zero values are not enough
to distinguish these cases. Apply the same care to booleans and fields whose empty
string is a valid edit. Keep form visibility distinct from authorization.

Replace the full original card used for `reflect.DeepEqual` with an opaque
revision mechanism owned and verified by the application. Choose a mechanism
that detects the same concurrent changes, including changes to fields not exposed
by the editor. A canonical content fingerprint is one possible implementation;
do not assume timestamps alone provide equivalent protection. A revision is not
authorization. Preserve the existing atomicity/locking guarantees and rejection
of stale edits. Avoid introducing a persistence schema change for revision data.

### Action availability and operation outcomes

The application derives action availability and meaningful reasons from its
policy. The frontend uses these facts to present or disable actions and chooses
keyboard/navigation behavior. Execution must revalidate current state: an action
being enabled in a read result does not authorize a later mutation.

For example, the application determines whether a session can be relaunched,
whether the selected dashboard may edit a shared card, and whether a saved
conversation is compatible with the destination driver. It returns availability
and reasons through application contracts. The UI decides how to present those
results: disable an action, display the reason, or adjust help text. It must not
independently reconstruct eligibility from lifecycle, ownership and deletion
fields. Otherwise a backend rule change could leave the frontend applying an
outdated rule. The same application policy should serve UI and MCP operations
where their rules overlap, while preserving their different caller authorities.

Rechecking at execution time is intentional and remains in the application:
state may have changed since the UI read availability. Do not treat a UI-supplied
availability flag as authoritative or remove backend validation because the
screen already disabled an action.

Focus, navigation, keyboard shortcuts, form visibility, selection and confirmation
interaction remain frontend rules. Local form feedback may help users correct
input, but authoritative validation remains in the application. Form visibility
does not grant permission, and a confirmation dialog does not replace the
application's confirmation-token and scope checks.

Avoid duplicating the lifecycle state machine in presentation models. A semantic
status for display and a coherent set of action facts are sufficient. Unknown
statuses must have defined presentation behavior; backend extensions should not
silently enable unsafe actions.

Expose structured lifecycle outcomes so the UI does not identify successful
process effects by inspecting `session.ErrPersistFailed`. Represent a usable
activation with a metadata warning separately from a failure before spawn. State
which combinations are valid; do not invent several independent flags allowing
contradictory outcomes. Preserve cancellation and post-effect evidence, including
uncertain outcomes, without retrying process/input effects automatically.

### Resource ownership and copy semantics

Contracts must define whether slices/maps/byte buffers are transferred, copied or
immutable after publication. Do not accidentally share mutable backend state
with UI commands or form drafts. Clone at the appropriate ownership boundary;
do not deep-copy indiscriminately at every layer.

An attachment contract must hide the PTY file and offer input, output, resize,
initial output/header, completion and release behavior as required by the current
implementation. Closing the frontend attachment releases its subscription; it
must not accidentally terminate the agent. Input continues through the driver's
serialization gate. Specify stale activation, EOF, cancellation, resize failure,
idempotent close and cleanup of unused/stale asynchronous results.

## 5. Migration workstreams

The expected shape is seven implementation sessions and one final verification
session. This is an estimate, not a fixed count of features. The planning agent
may split large workstreams, especially card editing and attachment, provided
dependencies and acceptance criteria remain explicit.

Each feature must keep the application buildable and usable, use the new boundary
for its complete vertical slice, and remove the corresponding old path once all
callers are migrated. Temporary bridges require a named removal step. Do not
leave parallel permanent APIs or a compatibility package hiding old dependencies.

### Step 1 — Establish contracts, boundaries and enforcement

Dependencies: none.

Tasks:

1. Inventory frontend production dependencies, exported model leakage, operational
   calls and existing tests. Recheck the evidence in section 2 against the checkout.
2. Record current user-visible behavior and the relevant concurrency, scope and
   persistence guarantees for each workflow.
3. Establish `api/contracts`, the presentation model location and consumer-owned
   capability interface conventions. Define only contracts required by the first
   slices; do not speculate about remote transports.
4. Agree revision, partial-update, action availability, outcome and resource
   ownership semantics before dependent features implement them.
5. Add a dependency check that covers production imports and contract leakage.
   Initially allowlist exact existing violations, with owning removal steps. New
   violations must fail immediately. Account for build tags/platform files.
6. Document existing preview-only dependencies for removal in step 6. Do not
   create a new long-lived preview contract.

Acceptance:

- The boundary is precise enough that another agent can determine where new code
  belongs and which types it may expose.
- Contracts are independent types, not aliases or embedded backend structs.
- The application builds and existing behavior is unchanged.
- Dependency enforcement runs as a normal check; temporary exceptions are explicit.

Review focus: avoid premature public API proliferation and unused abstractions;
verify the check catches a representative forbidden import/type leak.

### Step 2 — Separate dashboard, card and session reads

Dependencies: step 1.

Tasks:

1. Expose dashboard-scoped read contracts over the existing checked read path.
   Preserve dashboard deletion recovery and deleted-card information.
2. Introduce dashboard/card/session presentation rows. Move registry joins and
   active conversation selection out of list renderers.
3. Provide application-derived action availability. Preserve execution-time checks.
4. Convert the relevant messages, setters, callbacks and selections to the new
   types; actions should generally carry identities rather than whole rows.
5. Keep request generations, session epochs and dashboard identities on asynchronous
   UI envelopes. These are frontend stale-result controls, not backend policy.
6. Preserve refresh failure semantics: failed reads must retain prior data, while
   successful empty results must clear it. If replacing nil-slice conventions,
   make the new distinction explicit and test it.
7. Move global registry reads needed by these screens behind application reads,
   or introduce their focused contracts for completion in step 5. Hide filesystem
   fingerprints behind opaque application change tokens; keep polling cadence in
   the UI and fingerprint implementation in the backend.

Acceptance:

- Dashboard and list rendering consume focused models without config/session
  imports, excluding only legacy preview plumbing scheduled for removal in step 6.
- Existing names, PID/registration/conversation details, sorting, filtering,
  selection preservation, artwork and deleted-card behavior remain available.
- Refreshes cannot update another dashboard/view or erase data after failed reads.

Review focus: accidental loss of fields currently used outside rendering, mutable
map/slice sharing, inconsistent action facts and unintended scope widening.

### Step 3 — Separate card editing and save semantics

Dependencies: steps 1–2.

Tasks:

1. Introduce editor data, form drafts, explicit changes and opaque expected revision.
2. Move overlay/preservation into application save logic. Cover `SystemInstruction`,
   `Env`, immutable identity/ownership, timestamps and future uneditable fields.
3. Preserve hidden persona/MCP selections and explicit clearing. Preserve loaded
   references missing from a failed or incomplete registry read.
4. Keep shared cards read-only for consumers and enforce owner-only mutation in
   the service. Preserve restrictions on unsharing referenced cards.
5. Keep create and update semantics explicit. The application allocates identity
   and timestamps and validates the target dashboard.
6. Replace full-card payloads in editor callbacks, coordinator state and save
   messages. Remove obsolete full-card concurrency plumbing.

Acceptance:

- A save request contains intended edits and concurrency identity, not an entire
  storage record.
- Unexposed/hidden fields survive saves; explicit empty selections are saved.
- Concurrent changes, including unexposed-field changes, reject stale saves.
- Ownership, sharing and deletion/admission checks remain enforced atomically.
- Editor reset/new-card flows do not inherit a previous card's state.

Review focus: this is a data preservation and concurrency change, not just type
mapping. Require targeted behavioral tests before approving the feature.

### Step 4 — Separate session interactions and outcomes

Dependencies: steps 1–2; coordinate card identities with step 3.

Tasks:

1. Migrate launch, relaunch, switch, rename, stop and remove requests/results to
   application contracts and consumer interfaces.
2. Introduce structured lifecycle outcomes; remove frontend interpretation of
   manager-specific persistence errors from views, messages and command helpers.
3. Provide conversation options with compatibility/reasons from application logic.
   Represent new/resume selection explicitly without exposing session bindings.
4. Revalidate conversation identity and compatibility during execution under the
   existing admission rules; preserve workspace identity on relaunch/switch.
5. Put IDE/terminal workspace operations behind application methods. Eliminate
   the frontend's manager lookup and its separate check-then-use implementation;
   preserve scope, actual saved workspace path and configured command behavior.
6. Keep frontend navigation, confirmation interaction and attach-after-success
   decisions in the UI. No blocking work may run in a view update callback.

Acceptance:

- The frontend uses no session/driver types or error sentinels for these workflows.
- A usable process remains usable after a metadata warning, with a visible warning.
- Fatal failures do not trigger attachment; uncertain effects are not replayed.
- Conversation selection and dashboard-scoped actions remain revalidated.
- Original/temporary workspace identity, missing-directory behavior and switch
  versus relaunch semantics are preserved.

Review focus: preserve actual effect evidence, cancellation behavior and locking;
do not reclassify every warning as either complete success or complete failure.

### Step 5 — Separate configuration management

Dependencies: steps 1–3; step 4 can proceed independently where contracts permit.

Tasks:

1. Move source listing, save, delete, refresh/sync, MCP configuration and settings
   operations behind focused application capabilities.
2. Introduce registry choice/row models and separate MCP editor inputs. Do not pass
   command credentials/maps into list rows merely because the stored object has them.
3. Replace built-in-ID policy checks in views with application-provided edit/delete
   availability and reasons. Backend enforcement remains mandatory.
4. Define source edit identity/concurrency semantics. Current callbacks use list
   indexes; do not let an outdated index edit/delete a different source after a
   refresh. Use stable identities or validated expected state as appropriate,
   without inventing a storage migration solely for this purpose.
5. Preserve repository read-modify-write atomicity and saved-source-but-sync-failed
   outcomes. Return structured saved/refresh/warning facts instead of requiring
   the UI to identify `config.SyncFailedError`.
6. Finish migration of registry reads/change detection and remove direct repository
   operations from callbacks and command helpers.
7. Preserve MCP argument/map parsing and escaping behavior. Text-field syntax is
   frontend concern; configuration validation is application/backend concern.

Acceptance:

- No frontend repository, file stamp or config error dependency remains for these
  operations.
- Built-in MCP protections, settings defaults and source synchronization behavior
  are preserved.
- Stale source edits cannot silently affect a different entry.
- Failed reads remain distinct from empty inventories, and partial success causes
  the right refresh and warning.

Review focus: error propagation, synchronous I/O inside frontend callbacks,
credential-bearing data in unnecessary models and concurrent source updates.

### Step 6 — Replace the dashboard center and remove the old preview

Dependencies: steps 1–2; coordinate shared read contracts with step 5. Must finish
before step 7 retires `internal/app`.

Source requirement: `feature-dbcenter-rework.md`. The decisions below incorporate
the user's clarifications, including approval of the failure policy and the
available-space report display. The restart reconstruction strategy below avoids
introducing another persistence mechanism for content already in report history.

Confirmed behavior:

- Show reports authored by the selected session, not reports it received or all
  reports in the dashboard. Include report text, kind and the time it was recorded.
- Keep reports across relaunches of that session. Select by stable session identity,
  not only the current activation. Preserve activation attribution internally where
  needed; a different session must not inherit the history.
- Order reports by timestamp descending, with a stable identity/sequence tie-breaker.
- Show only report content that fits the center panel's available space. Older or
  overflowing content is omitted from this display; do not provide report paging
  or history scrolling. New reports take precedence and push older content out.
  Omission from the display does not delete the underlying message-store reports
  or change existing MCP history behavior and retention.
- On session stop, keep the current report display. Relaunch must not clear it
  merely because the activation changes. On Kankuro restart, reconstruct the
  fitted display from retained durable reports rather than saving rendered lines.
  Do not promise reports beyond the existing store's retention guarantees.
- Refresh reports when a report is received. When no reports exist, refresh the
  workspace view every 30 seconds. Load immediately on selection; subscribe to new
  reports even in workspace mode so the first report changes the panel promptly.
- If no reports exist, show changed files from the session's actual workspace,
  equivalent in purpose to `git status`. If it is not a Git repository, show a
  detailed listing equivalent in purpose to `ls -la`.

Tasks:

1. Add dashboard-scoped application reads for retained reports authored by the
   selected session across activations. Reuse the existing message store without
   changing MCP authorization or impersonating an agent to call MCP tools. Define
   the human read scope and validate persisted session/dashboard membership.
2. Expose report changes through an application-owned subscription/change signal,
   with explicit unsubscribe/cancellation semantics. Deliver no Bubble Tea values
   from backend code. Close the read/subscribe race so the first new report cannot
   be missed between the initial query and subscription. Coalescing changes is fine
   if a refresh retrieves all new retained records.
3. Introduce structured contracts for reports, Git entries and directory entries;
   do not return raw storage records or terminal-formatted command output. Keep
   timestamps as time values and format them in the frontend.
4. Keep Git/filesystem inspection behind the application boundary. Inspect the
   saved actual workspace, including temporary workspaces. Use safe argv-based Git
   execution and structured status parsing; use filesystem APIs for directory
   entries where practical. No shell interpolation or blocking I/O in UI Update.
5. Bound reads and rendering to the recent content useful for the panel. The UI
   owns display-width measurement, timestamp layout, wrapping and clipping to its
   current dimensions; the backend returns structured recent reports with a
   bounded query limit, not terminal-sized or preformatted strings. Fetch enough
   recent entries to fill the available space without loading the entire history.
   Recompute on resize, allowing retained older reports to fill newly available
   space. No report pagination, history scrolling or new history deletion policy.
   Keep large workspace inventories bounded too, with visible omission indication;
   no unbounded file listing should block the UI.
6. Implement failure handling using the policy below. Distinguish confirmed empty
   results from failed reads, and mark stale/fallback content explicitly.
7. Add the UI center component and commands. Keep selected session identity,
   dashboard scope and generation on asynchronous results. Cancel obsolete reads,
   timers and subscriptions when selection changes, navigation leaves the screen,
   the session disappears or the application exits. Avoid overlapping 30-second
   workspace refreshes. New reports update the fixed newest-first display rather
   than preserving a scroll position. Session stop must not clear report content.
8. Remove the old preview component, dead selection hooks, output subscription
   wiring, preview-specific payloads and obsolete tests/documentation. Inspect
   callers before deleting shared session output/scrollback functionality used by
   attachment. Replace preview tests with center-panel behavior tests.

Agreed failure policy:

- Successful report read with reports: display reports. Workspace failure must not
  hide reports, and there is no need to inspect the workspace in this mode.
- Report read failure: show a warning and keep the last successful report view for
  that session marked stale. If there is no usable cached report view, attempt the
  workspace view and label it as fallback because reports could not be loaded.
  Never report a failed history read as proof that there are no reports.
- Git workspace with no changes: display "Working tree clean"; do not switch to a
  directory listing merely because Git returned an empty successful result.
- Confirmed non-Git workspace: display its directory listing without an error.
- Git inspection fails for another reason: show the Git error and attempt the
  directory listing as explicitly labelled fallback. Do not classify permission,
  execution, timeout or repository errors as "not a Git repository".
- Directory listing fails, or the workspace is missing/inaccessible: show the
  error. Retain suitable prior content for that same session marked stale if
  available. Do not create the workspace or show unrelated files.
- Report subscription/read failures must have bounded retry/recovery behavior;
  use the 30-second fallback refresh while degraded, then return to event-driven
  updates when recovered. Avoid error-driven busy loops.

Acceptance:

- Report author/session selection, relaunch continuity, descending order and
  immediate event-driven updates match the confirmed behavior above.
- With no reports, Git/directory information loads immediately and refreshes every
  30 seconds; a received report promptly takes precedence.
- A clean Git workspace, non-repository workspace and read failure are distinct.
- Errors remain visible alongside useful fallback information where available.
- Only the newest content fitting the available panel area is shown, with no
  report paging or history scrolling. Resize, new reports, stop, relaunch and
  Kankuro restart follow the display/continuity rules above.
- Stale results cannot display another session's reports/files. Resource cleanup
  remains correct when switching selection during reads/events.
- No obsolete preview UI plumbing or migration exception remains. Attachment still
  receives its initial output and live stream and retains the input gate.
- Frontend code imports no Git/process/storage implementation to provide the panel.

Review focus: report attribution across activations, fitting/clipping and restart continuity,
read/subscribe races, scope, cancellation, large inventories, fallback labels and
shared attachment functionality accidentally removed during preview cleanup.

### Step 7 — Close resource boundaries and retire internal/app

Dependencies: steps 2–6.

Tasks:

1. Introduce and implement the application attachment contract from section 4.
   Replace concrete handles in frontend messages and terminal code.
2. Move PTY resize mechanics behind the backend contract; keep human terminal raw
   mode, keyboard protocol restoration and detach handling in `ui/terminal`.
3. Finish deletion/confirmation boundaries for cards and dashboards. UI receives
   summaries, tokens and operation outcomes rather than persistence manifests.
   Preserve revision-bound confirmation, foreign-reference confirmation and retry.
4. Audit all remaining frontend imports, signatures and operations.
   Remove legacy bridges and unused repository/manager fields and constructors.
5. Move the now-separated coordinator, navigation, commands and handlers from
   `internal/app` into `internal/ui`. Keep root/child dependencies acyclic.
6. Update `main` to inject application capabilities and forward runtime failures
   through a frontend entry point. Preserve initialization and shutdown ordering.
7. Move/update tests and documentation references; remove the obsolete package.
   Do not leave a permanent forwarding `internal/app` package.

Acceptance:

- `internal/app` is retired, and its frontend behavior lives under `internal/ui`.
- Production frontend constructor dependencies are application capabilities only.
- No raw backend PTY/manager/driver crosses the application boundary.
- Attach/detach, terminal restoration, resize, output completion and stale-result
  cleanup work without resource leaks or terminating the session on detach.
- Card/dashboard deletion remains recoverable, scoped and confirmation-bound.
- `main` retains composition and lifetime ownership; no backend imports the UI.

Review focus: input serialization, attachment ownership, goroutine/subscription
leaks, terminal restoration on errors and accidental business logic moved into UI.

### Step 8 — Final architectural and behavioral verification

Dependencies: steps 1–7.

Tasks:

1. Remove migration allowlists and compatibility shims. Check production imports,
   exported contracts, aliases and platform-specific/build-tagged files.
2. Ensure unit tests for frontend orchestration use small fakes implementing
   application capabilities. Keep deliberate real-backend integration tests
   separately identifiable; their imports must not mask production dependencies.
3. Review public contracts for backend fields added merely for convenience,
   presentation concerns in backend results and shared mutable data ownership.
4. Verify behavior across dashboard scope, refresh failure, stale results, card
   preservation/concurrency, lifecycle warnings, conversations, deletion and attach.
5. Run final appropriate build/static/test/race checks and isolated terminal checks.
6. Update `project.md` and relevant architecture/onboarding references to describe
   the final implementation. Keep task logs and review evidence out of project.md.
7. Produce a final review report with checks run, results, known limitations and
   explicit evidence for each completion criterion.

Acceptance:

- Every final criterion in section 7 is satisfied with evidence.
- No legacy exception is silently carried into a claim of full separation.
- Any remaining unrelated issue is documented without expanding this migration.

## 6. Validation and review strategy

Validate each vertical slice when it changes; do not postpone all testing until
step 8. Use the current repository's required checks and skills for implementation.
Typical commands are:

```bash
go build ./...
go vet ./...
go test ./...
go test -race ./...
golangci-lint run
git diff --check
```

Start with focused tests. Lifecycle, attachment, shared-state or command changes
need relevant race coverage, including the existing coordinator command race
backstop after it is moved. Broaden checks at integration points and final review;
avoid repeatedly running unchanged expensive suites without a reason.

Tests should verify behavior and contracts, not merely mirror mapping assignments.
Priority scenarios:

- Another dashboard/view is selected while a read/save/action is in flight.
- A stale attachment result is discarded and its subscription is released.
- A registry read fails versus successfully returns no items.
- A card save preserves unexposed fields and hidden selections; explicit clear
  works; a concurrent backend edit rejects the stale save.
- Shared-card ownership/unsharing constraints and deletion admission remain intact.
- A process starts but metadata persistence fails; the UI still offers usable flow.
- A conversation is replaced or becomes incompatible between selection and action.
- Source ordering changes between form load and mutation.
- Source configuration saves while derived synchronization fails.
- Built-in MCP edit/delete is attempted through both UI and application entry points.
- Attach resizes, detaches, handles process exit/error, and restores terminal state.
- The center panel shows reports authored across session activations, switches from
  files to reports on receipt, refreshes workspace data every 30 seconds without
  reports, clips content to available space, preserves it through stop/relaunch
  and exposes read errors alongside labelled fallback.
- Application shutdown still drains operations, sessions and MCP in the right order.

Use fake application capabilities for frontend behavior and backend tests for
policy/ownership/persistence. Add integration coverage where the seam itself is
material, especially attachment and post-effect outcomes. Preserve MCP wire and
authorization coverage when shared service implementation changes.

Manual terminal runs must isolate configuration and temporary workspaces as
specified by `project.md`. Native-agent runs additionally need isolated user
configuration. Native model-driven E2E runs are not automatically authorized by
this roadmap: follow `tests/e2e/AGENTS.md` and the documented provider/model and
total token-budget approval requirements before running them.

No commits unless explicitly requested. Preserve unrelated working-tree changes.

## 7. Final definition of done

1. All frontend production code, including the former `internal/app`, lives within
   the frontend boundary and depends on independent application contracts.
2. Frontend code imports no configuration/session/driver/MCP implementation or
   concrete application service. No public contract leaks those types indirectly.
3. Backend production code imports no UI package or presentation framework.
4. `internal/app` and temporary migration bridges are removed.
5. Views/components consume focused presentation data and emit identities/intents;
   they do not round-trip full storage objects or interpret backend error sentinels.
6. Application logic owns validation, authorization, compatibility, mutations and
   action availability; operations revalidate rather than trusting stale UI facts.
7. Card preservation and concurrent-edit safety are at least as strong as before.
8. Attachment hides backend resources while preserving serialization and cleanup.
9. Existing product behavior and failure guarantees are preserved except for the
   explicitly included preview replacement. The new center meets step 6 criteria;
   any further intentional behavior change is separately identified and approved.
10. Automated dependency enforcement and meaningful behavioral coverage protect the
    boundary. Old preview code and exceptions are removed without removing shared
    attachment functionality.
11. Main remains the concrete composition root, and documentation matches the final
    ownership and startup/shutdown design.

## 8. Instructions for the planning and review agent

Read this document as a whole before splitting it. Produce as many feature files
as needed for reviewable changes; eight workstreams do not require exactly eight
feature files. Do not implement the entire migration as one feature.

Each feature file should contain:

- Concrete problem, intended result and bounded scope.
- Prerequisite features and sequencing with the dashboard center replacement.
- Current entry points and affected callers/tests, revalidated from the checkout.
- Target ownership, contract/data changes and dependency rules.
- Validity rules for revisions, optional updates, outcomes and resource lifetime
  relevant to that feature.
- Ordered implementation/migration steps, including removal of the old path.
- Behavioral invariants, acceptance criteria and focused validation commands.
- Temporary exceptions, their exact scope and which feature removes them.
- A final-review checklist with evidence required for acceptance.

Carry details into the step that implements them; do not reduce card editing,
attachment or partial-success handling to generic instructions to “add DTOs.”
Preserve the distinction between application contracts and UI models, and between
presentation action availability and authoritative execution checks.

During final review of each feature, inspect the implementation and callers,
not only the feature author's summary. Confirm checks were run and inspect their
actual results. Record remaining work accurately. A compiling package rename is
not evidence of separation, and passing frontend tests alone is not evidence of
preserved backend policy.

## 9. Trade-offs and decisions to settle during planning

The agreed direction costs explicit mapping, extra contract types and deliberate
API evolution. In return it limits the reach of backend representation changes,
enables frontend testing without concrete managers/repositories and makes data
preservation and resource ownership explicit.

Keeping current imports would be less work but would retain the coupling that
motivated this roadmap. Moving shared structs into a neutral package alone would
clarify imports without separating broad snapshots and whole-record writes.
Duplicating every persistence field into UI types would retain conceptual coupling
and add mapping cost. Prefer focused operation/read contracts and purpose-built
presentation models.

Settle these details in the dependent feature specifications:

- Exact capability grouping and constructor shape; keep interfaces cohesive.
- Revision generation/verification and atomic save integration.
- Presence-aware update representation and hidden-field behavior.
- Lifecycle outcome variants and compatibility with existing MCP responses.
- Attachment stream, close, cancellation and resize semantics.
- Registry change-token semantics and source edit identity.
- Dependency-check location and precise test-only exceptions.
- Center-panel bounded recent-read limits and report wrapping/clipping details,
  Git scope within a repository, dotfile/symlink representation and degraded-read
  retry details; preserve step 6's agreed behavior and failure policy.

These choices refine the agreed boundary; they should not delay independent work
that does not depend on them. Revisit the design if implementing it requires
duplicating policy, losing current failure evidence, inventing persistence/network
infrastructure, or growing a universal model/interface. Those are signs to narrow
the contract rather than expand the migration.
