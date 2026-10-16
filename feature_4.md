# Feature 4: MCP discovery and session lifecycle control

Status: proposed; implementation and integration review pending.

Dependency: Feature 3 accepted.

## Intent

Let an agent discover existing cards and manage other ordinary sessions through the
same operations available to the human UI.

## Scope and constraints

- Expose card discovery sufficient to choose an existing project-scoped agent.
  Expose session listing/inspection, creation, stop, relaunch, card switch, and
  deletion with the shared lifecycle rules.
- Return stable session IDs and resolved workspace/sub-workspace paths. Reuse
  existing workspace, skill, persona, and MCP preparation. Additional files are
  created by the orchestrator using its own filesystem tools.
- Creation from a session records authenticated parentage. Define explicit caller
  scopes for which sessions can be inspected or controlled; ancestry alone and
  caller-supplied IDs must not grant arbitrary authority.
- Make stopped-versus-running requirements explicit. A waiting process is still
  running. If switching requires stopping first, document the sequence and errors.
- Apply conversation continuity from feature 2 to remote lifecycle operations.
- Return unambiguous outcomes, including IDs when an operation succeeded with a
  persistence warning. Establish duplicate-request behavior now so retries do not
  accidentally create multiple sessions or repeat destructive operations.
- Preserve responsiveness during workspace copying and process activation; define
  cancellation and shutdown behavior for operations in flight.

## Acceptance criteria

- A caller creates a session using a card for another project, retrieves usable
  paths, and creates a file there using its own tools.
- It stops, switches, and relaunches that session without changing its workspace,
  and obtains the expected native conversation binding when returning to a card.
- Unauthorized calls, retry behavior, subtree deletion, and partial-success outcomes
  are exercised through the MCP boundary as well as the shared service.

## Planning decisions

Choose tool names/schemas, caller scopes, request identity, and timeout semantics.
Keep schemas useful to agents without exposing secrets or internal mutable objects.
Do not defer basic correctness of retries or permissions to the final hardening pass.

## Out of scope

Workspace/card mutation tools, server-side feature-file generation, Git automation,
input/output communication (feature 5), and prescribed orchestration policies.

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
