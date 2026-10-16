# Feature 2: Native conversation continuity and minimal MCP

Status: proposed; implementation and integration review pending.

Dependency: Feature 1 accepted.

## Intent

Make native CLI conversations survive Kankuro relaunches and card switches. Deliver
this first for human use, introducing MCP with a single registration operation.

## Scope and constraints

- Evolve the driver interface for native conversation discovery/registration and
  launch arguments. Generic discovery instructs the agent to discover its ID and
  report it through MCP; specialized drivers may override the mechanism.
- Session arguments are a driver method, following the model-argument pattern.
  The generic implementation returns `--session <id>` as arguments; specialized
  drivers override it. Verify resume semantics for each supported CLI.
- Add an embedded loopback HTTP MCP endpoint controlling the existing manager.
  Expose only `register_conversation(conversation_id)` initially. Include caller
  authentication, Origin validation, clean startup/shutdown, and session config
  injection compatible with the existing MCP access mechanism.
- Derive the Kankuro session and activation from authenticated caller identity.
  Reject stale registrations and prevent one caller from registering another
  session's conversation. Never treat registration parameters as authorization.
- Persist conversation bindings per card within each session, with driver identity.
  Preserve compatibility with existing session records that have no binding.
- Reuse saved bindings on relaunch and when returning to a card. First use starts
  fresh. Preserve the working directory in all cases, including explicit fresh start.
- Make pending registration, registered identity, and failures understandable.
  Never silently replace a failed requested resume with a fresh conversation.
- Registration is asynchronous; do not block the UI waiting for an agent report.
  Ensure current credentials reach resumed agents and do not leak into logs or
  unrelated sessions, including sessions sharing a source workspace.

## Acceptance criteria

- On a real supported CLI, capture evidence of reliable ID discovery and registration;
  do not assume the agent knows its ID merely because it was instructed to report it.
- Launch card A, register, stop and switch to B, register, stop and return to A.
  Verify A resumes its own conversation and workspace contents remain intact.
- Restart Kankuro and repeat resume from persisted state. Exercise explicit fresh
  start and loading older session records.
- Verify stale/unauthorized reports, missing registration, invalid resume references,
  and persistence failures. Record limitations that cannot be reliably detected.

## Planning decisions

Investigate the first supported CLI before finalizing the interface. Decide how
registration instructions are delivered, how missing registration affects switching,
when a binding is replaced, and what happens if a card's driver changes. Confirm
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
