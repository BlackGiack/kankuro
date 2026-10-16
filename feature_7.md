# Feature 7: Multiple dashboards for independent factories

Status: proposed; implementation and integration review pending.

Dependency: Feature 6 accepted; deferred until the existing dashboard is useful and dependable.

## Intent

Allow users to organize associated agents and sessions into multiple dashboards,
which they can use as independent factories without imposing a factory workflow.

## Scope and constraints

- Introduce dashboard membership independently of session ancestry. A parent/child
  relationship does not determine navigation or act as a substitute for membership.
- Support independent dashboard views of associated agents/cards and sessions.
  Keep conversation continuity and workspace retention intact.
- Design navigation after reviewing the working single-dashboard experience. The
  original left/right parent-child drill-down idea is not a requirement.
- One orchestrator per dashboard is an expected user convention, not an enforced
  limit. Workflows remain user-defined.
- Preserve existing installations with a clear default-dashboard migration.
- Integrate membership with MCP discovery, creation, and caller scopes so agents
  cannot accidentally mix unrelated dashboard operations.

## Acceptance criteria

- A user can operate two dashboards with understandable agent/session membership
  and switch between them without losing active sessions or their context.
- MCP callers see and control the intended scope according to explicit rules.
- Existing records migrate without lost cards, sessions, workspaces, or bindings.
- Parenthood, descendant deletion restrictions, and dashboard membership behave
  consistently in the cases supported by the chosen design.

## Planning decisions

Settle the dashboard data model, shared-versus-exclusive card membership, session
moves, cross-dashboard ancestry, dashboard deletion, and navigation at this stage.
Do not assume dashboard deletion means deleting its sessions. Consult the user for
material UX/product choices once the preceding features provide concrete evidence.

## Out of scope

Mandatory planner roles, workflow engines, agent-managed workspace configuration,
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
