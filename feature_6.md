# Feature 6: Validate and harden user-defined orchestration

Status: proposed; implementation and integration review pending.

Dependency: Feature 5 accepted.

## Intent

Prove that the general-purpose operations support real user-defined orchestration
on the existing dashboard, then resolve integration issues before expanding the UI.

## Scope and constraints

- Run a persistent orchestrator with separately configured project-scoped cards.
  Use the user's executor/reviewer/revision cycle as an example, not a product rule.
- Let the orchestrator retrieve paths, write handoff files with its own tools, send
  instructions, inspect responses, and control card switches. Native subagents
  remain available inside each CLI independently of Kankuro orchestration.
- Exercise multiple managed sessions, one awaiting input, human intervention,
  retries, registration/resume failures, and restart recovery.
- Verify that session records and bindings recover truthfully after restart;
  quitting still stops processes. No daemon or automatic workflow resumption.
- Improve operation errors and existing-dashboard visibility where real use exposes
  gaps. Show descendant activity independently of the parent's process status.
- Supply clear MCP tool documentation. Add a small usage skill only if needed;
  keep role selection, file conventions, review rules, and completion policy in
  user-owned instructions or an explicitly labeled example.
- Add meaningful regression coverage for issues found. Avoid speculative features
  or duplicate tests that merely repeat implementation details.

## Acceptance criteria

- Record a real end-to-end example: the orchestrator remains running while another
  session executes, reviews, revises, and reaches readiness as judged by the user's
  convention. No code transfer into the orchestrator workspace is required.
- Multiple-session use and failure recovery are understandable to a human from
  the existing dashboard and documented MCP responses.
- Required repository checks pass. `project.md` explains the implemented service,
  MCP, driver, persistence, and lifecycle contracts for the next agent.
- Handoff distinguishes automated coverage, real CLI checks, and any untested
  behavior; limitations are explicit rather than presented as successful validation.

## Planning decisions

Choose a small reproducible example and isolated configuration/workspaces. If real
CLI access or credentials are unavailable, record the blocked acceptance work;
mock tests alone do not complete the real-agent validation requirement.

## Out of scope

Built-in executor/reviewer workflows, ticket management, remote Git automation,
multiple dashboards, and agent-managed workspace configuration.

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
