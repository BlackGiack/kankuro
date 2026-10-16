# Feature 3: Session ancestry and subtree lifecycle

Status: proposed; implementation and integration review pending.

Dependency: Feature 2 accepted.

## Intent

Allow an ordinary session to be recorded as created by another session, with clear
lifecycle and deletion behavior. Parenthood is metadata, not a new session type.

## Scope and constraints

- Persist parent relationships; older sessions load as roots. Reject invalid
  relationships and cycles, and define behavior for corrupt/orphaned persisted data.
- New sessions use the chosen card's existing workspace preparation, including
  temporary workspaces and sub workspaces. Do not copy the parent's workspace by
  default or nest a child's storage under its parent's temporary directory.
- Deleting a parent deletes its entire descendant subtree. Reject deletion if any
  member is running or launching, relaunching, or switching cards.
- Protect deletion against concurrent child creation/activation. Specify recovery
  and truthful error reporting when filesystem or record deletion partially fails.
- Keep each process's own status separate from aggregate descendant activity.
  Surface enough information on the current dashboard to explain deletion refusal;
  show the affected scope in the human deletion confirmation.
- Retain current shutdown semantics. This feature does not imply recursive stop.

## Acceptance criteria

- Parentage survives restart without changing existing session/workspace identities.
- Children can use projects different from their parent's project.
- Running and activating descendants prevent subtree deletion, including concurrent
  attempts. A stopped subtree can be removed without affecting unrelated sessions.
- Tests cover deeper descendants, temporary-workspace cleanup, legacy records, and
  failure paths rather than only a single parent/child happy path.

## Planning decisions

Choose the smallest relationship representation and deletion reservation strategy.
Document whether and how ancestry may change; creation-time parentage is sufficient
initially. Define failure behavior without promising transactional filesystem deletes.

## Out of scope

Tree navigation, factory dashboards, ticket models, and workflow scheduling.

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
