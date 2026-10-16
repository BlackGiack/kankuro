# Feature 1: Shared application operations

Status: proposed; implementation and integration review pending.

Dependency: None. Start here.

## Intent

Create the narrow UI-independent application boundary that the TUI and future MCP
endpoint will share, without changing existing user behavior.

## Scope and constraints

- Extract shared launch, relaunch, card-switch, and relevant deletion orchestration
  from UI-specific code. Capability resolution and cross-package rules belong in
  this boundary; Bubble Tea commands and messages remain UI adapters.
- Reuse the configuration repository and the existing session manager. Do not add
  another owner of process state or wrap every repository method for symmetry.
- Preserve error distinctions, including an activation that succeeded but could
  not be persisted. Preserve shutdown and activation concurrency guarantees.
- Card switching always retains the session's working directory, even when the
  destination card has a different configured workspace.

## Acceptance criteria

- Existing TUI actions go through the extracted operations and retain their
  behavior, including warnings, configuration resolution, and workspace retention.
- The operations can be called and tested without a coordinator or Bubble Tea.
- Existing race and lifecycle checks pass; tests cover the moved rules where needed.

## Planning decisions

Choose the package/API shape and extraction boundary from the current code. Keep
new interfaces limited to actual consumers.

## Out of scope

MCP transport, conversation registration, ancestry, workflow entities, and broad
configuration API refactoring.

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
