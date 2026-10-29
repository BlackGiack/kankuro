# Feature 7: Multiple dashboards for independent factories

Status: proposed; implementation and integration review pending.

Dependency: Feature 6 accepted; deferred until the existing dashboard is useful and dependable.

## Intent

Allow users to organize associated agents and sessions into multiple dashboards,
which they can use as independent factories without imposing a factory workflow.

## Scope and constraints

- Make dashboards the top-level application objects, each with a stable ID, name,
  and description.
- Add a `shared` toggle to agent cards. An unshared card belongs exclusively to
  the dashboard where it was created; a shared card is available across dashboards.
- Apply shared/exclusive visibility and access consistently in both the human UI
  and MCP. Sharing is a property of the card, not an MCP-only permission or UI-only
  filter.
- Sessions belong strictly to the dashboard where they were created, including
  sessions launched from shared cards. Sharing a card does not share its sessions.
- Support independent dashboard views of available cards and owned sessions.
- Keep conversation continuity and workspace retention intact.
- Preserve existing installations with a clear default-dashboard migration.
- Integrate membership with MCP discovery, creation, and caller scopes so agents
  cannot accidentally mix unrelated dashboard operations.
- A caller using MCP from a session can see/control only sessions in that session's
  dashboard and use cards available there: its exclusive cards plus shared cards.
  Derive scope from the caller session's dashboard, including when its card is
  shared; changing the dashboard viewed by the human must not change caller scope.
- Identify the dashboard context of MCP operations explicitly. Resolve it from the
  authenticated caller session and expose its ID in discovery/operation context so
  the caller can tell which dashboard an operation acts in. If a tool accepts a
  dashboard ID, validate it against that authority; a supplied ID alone must not
  authorize cross-dashboard access. Human operations use the currently opened
  dashboard. Both frontends enforce the same membership rules in the shared service.
- Skills, MCP definitions, personas, and workspace registries remain global across
  dashboards. Users select those capabilities when configuring dashboard-specific
  or shared cards.
- Deleting a dashboard deletes its sessions and its unshared agent configurations.
  Shared cards and global capability/workspace registries survive dashboard deletion.
  Apply the existing process teardown, temporary workspace/trust cleanup, and
  deletion rules; do not leave running processes or delete another dashboard's
  sessions. Surface partial cleanup failures truthfully.
- The initial/top screen is a navigable list of dashboards. The user moves among
  entries and presses Enter to open the selected dashboard's agent/session view.
  Provide navigation back to the dashboard list without stopping active sessions.

## Acceptance criteria

- A user can operate two dashboards with understandable agent/session membership
  and switch between them without losing active sessions or their context.
- MCP callers see and control the intended scope according to explicit rules.
- Shared cards are available in both dashboards; unshared cards remain exclusive
  to their creation dashboard. Sessions created from a shared card remain visible
  and controllable only in their own creation dashboard.
- Human UI and MCP expose the same shared/exclusive card availability for a given
  dashboard. MCP responses make the executing dashboard identifiable; human
  navigation cannot change a running caller's scope, and arbitrary dashboard IDs
  cannot grant access to another dashboard.
- The dashboard list exposes each dashboard's identity and description and opens
  the selected dashboard with Enter. Returning to the list or opening another
  dashboard preserves running sessions and their conversation/workspace state.
- Deleting one dashboard removes its sessions and unshared cards, preserves shared
  cards and the other dashboard's data, and handles active process teardown and
  cleanup errors explicitly.
- Existing records migrate without lost cards, sessions, workspaces, or bindings. We ensure this through a migration script, not by legacy code.

## Planning decisions

The following product decisions are settled:

- Dashboard is the first layer of objects: ID, name, and description.
- Cards have a shared toggle; exclusive cards belong to their creation dashboard,
  while shared cards are usable across dashboards.
- Sessions always belong to their creation dashboard and do not become shared
  when their card is shared.
- Dashboard deletion cascades to its sessions and unshared card configurations,
  preserving shared cards and global registries.
- Application navigation starts at a dashboard list; Enter opens a dashboard.
- Existing data is migrated by an explicit migration script, preserving records
  and bindings, rather than permanent compatibility branches in runtime code.

Implementation planning must choose persistence layout, migration execution,
editing/navigation details, and the coordination and error handling for cascading
deletion. Address shared-card references from other dashboards and define what
happens when a shared card is made exclusive again; do not silently move sessions
or broaden caller permissions. Consult the user for remaining material product
choices without reopening the settled decisions above.

## Out of scope

Prescribed planner/executor/reviewer workflows, agent-managed workspace configuration,
and a persistent background daemon.

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
