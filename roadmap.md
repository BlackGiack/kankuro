# Kankuro roadmap: agent-operated development factories

Status: agreed direction, 2026-10-16. This document captures the discussion that
refines `feature.md`; it is a roadmap, not a detailed implementation plan.
`project.md` is the onboarding guide for agents entering the repository. Preserve
that purpose and update it throughout implementation as architecture, behavior,
entry points, and invariants change. Keep planned work here and implemented facts
in `project.md`; documentation updates are part of every checkpoint.

## Big picture

Kankuro exposes its existing session-management capabilities through a local MCP
so an agent running in one session can create, inspect, and control other sessions.
Kankuro provides flexible operations; users decide how their orchestrators behave.
It does not prescribe a development workflow or introduce a ticket/session subtype,
workflow engine, or mandatory stages.

Tickets and feature requests are units of work in the user's instructions, not new
Kankuro entities. All sessions below are the existing application concept.

One motivating example is the user's personal factory workflow: a persistent
planner/orchestrator receives feature requests, creates sessions using existing
project-scoped cards, and switches those sessions between executor and reviewer
cards until the work is ready for the user. The orchestrator stays in its own
session. The role sequence, handoff files, completion conventions, and review rules
are user-defined, not built-in product behavior.

```text
Example user-defined workflow
Planner/orchestrator session (stays running)
             |
             +-- Session A: executor -> reviewer -> executor -> ready for user
             |
             +-- Session B: executor -> reviewer -> ready for user
```

This complements the native subagent mechanisms of agent CLIs. Each agent can
continue using its own subagents internally. Multiple dashboards, each usable as
a factory, come after the capabilities work well on the existing dashboard.

## Agreed foundations

### Sessions, cards, workspaces, and conversations

- Reuse the existing Kankuro session model. Its identity is distinct from a native
  CLI conversation ID; a ticket is simply one possible reason to create a session.
- Cards define agent configuration and are already scoped to workspaces. Initially
  the orchestrator chooses an existing card for the desired project; it does not
  need to create or modify workspace configuration through MCP.
- A newly created session uses the selected card's workspace configuration,
  including its temporary-workspace setting. Reuse the existing preparation of
  temporary workspaces, sub workspaces, skills, and MCP configuration. It does not
  implicitly copy the orchestrator's workspace.
- Expose the resolved session workspace paths through MCP. For additional
  preparation, the orchestrator can use its own filesystem tools to create feature
  files, plans, or other artifacts there. Kankuro need not automate that preparation.
- Switching cards on an existing session **always preserves that session's
  working directory and work**, regardless of the new card's workspace setting.
- A session retains native conversation bindings per card, including the
  driver identity. Returning to a card resumes its saved conversation when one
  exists; first use starts a new conversation.
- Conversation continuity and workspace continuity are separate. Users and
  orchestrators decide how to brief agents about changes made while another card
  was active; Kankuro supplies the input and inspection capabilities.
- Parenthood is persisted metadata, not filesystem nesting. Temporary session
  workspaces remain independently stored under their own session IDs.

### Application boundary and MCP

- Extract a narrow, UI-independent application service for shared operations.
  TUI actions and MCP tools call the same service and enforce the same rules.
- Begin with the required operations rather than wrapping every repository method.
- Host the local MCP endpoint in the running Kankuro process so it controls the
  existing session manager and live processes. Start with loopback HTTP, caller
  authentication, and appropriate Origin validation.
- Bind MCP callers to their Kankuro session and activation. Do not trust a
  caller-supplied session ID as proof of authority.
- Preserve the current application lifetime initially: quitting Kankuro stops its
  running sessions. A separate persistent daemon is outside the initial scope.

### Communication and user-defined workflows

- Terminal `waiting` is a heuristic observation, not proof that work is complete.
- Expose input delivery and output inspection with clear semantics. Users may
  arrange explicit reports or handoff files; Kankuro must not infer workflow success
  from terminal silence or impose implementation/review/completion stages.
- Artifacts can remain in the managed session's workspace. The orchestrator does
  not generally need the code copied into its own workspace.
- Git/GitLab may be used by agents when a workflow requires it. Automatic patch
  transfer, remote Git integration, and local merge automation are not prerequisites.
- Raw terminal output is useful for inspection but is not necessarily the current
  visible screen. Output-reading APIs must state what they actually return.
- Provide clear MCP tool descriptions and, only if needed, a small usage skill.
  Workflow policies and orchestrator behavior remain under user control.

### Deletion

- Deleting a parent deletes its entire descendant subtree.
- Deletion is prohibited if the parent or any descendant is running or undergoing
  launch, relaunch, or card switching. Descendant activity therefore makes the
  parent undeletable even if its own process has stopped.
- Keep the parent's own process status distinct from aggregate descendant activity.
- Enforce the subtree rule in the shared service, including protection against a
  concurrent child creation or activation during deletion.

## Main checkpoints

### 1. Native conversation continuity for human use

Deliver a useful standalone feature before introducing orchestration.

- Evolve the driver interface to supply native session launch/resume arguments.
  Follow the existing model-argument pattern: the generic implementation supplies
  `--session <id>`, and specific drivers override it where necessary. Verify actual
  resume semantics for supported CLIs; pass arguments without shell interpolation.
- Introduce the minimal local MCP endpoint with one operation:
  `register_conversation(conversation_id)`.
- Use agent-assisted registration as the generic discovery strategy: provide
  instructions for the agent to discover its native conversation ID and report it
  through MCP. Confirm that discovery is feasible for each initially supported CLI.
- Bind registration to the current activation so late reports cannot overwrite a
  newer activation's binding. Capture and persist the reference as soon as available.
- Allow driver-specific discovery implementations later, feeding the same stored
  conversation binding.
- Persist bindings per card within each Kankuro session and use them on relaunch
  and card switch, including after restarting Kankuro.
- Distinguish pending registration, registered identity, and failure. A failed
  requested resume must not silently become a fresh conversation. Provide an
  explicit way to start fresh without replacing the workspace.
- Establish the shared service boundary needed by these operations; expand it
  incrementally in subsequent checkpoints.

**Acceptance:** launch card A and register its conversation, switch to B, return
to A and recover its conversation. Restart Kankuro and verify resume still works.
The workspace remains unchanged throughout. Exercise missing registration and
invalid resume references as well as the successful path.

### 2. Expose existing session operations through the shared service and MCP

- Let an orchestrator discover existing cards and create ordinary sessions from
  them, using the existing launch and workspace-preparation behavior.
- Expose resolved workspace and sub-workspace paths so the orchestrator can inspect
  the environment and create any additional files with its own tools.
- Expose scoped session inspection, input delivery, stopping, relaunch, and card
  switching. Card switching is a core operation, not an optional extension.
- Keep stop and activation behavior explicit: a process waiting at a prompt is
  still running. Stop the current process before activating another card while
  retaining its conversation binding.
- Implement persisted ancestry and the subtree deletion rules.
- Define input ownership so human interaction and orchestrator commands cannot
  accidentally interleave into the same prompt.

**Acceptance:** an agent creates a session for a project different from its own
workspace, retrieves its paths, prepares a file using its own tools, and switches
its cards through MCP while preserving the workspace and conversation bindings.

### 3. Validate flexibility through a user-defined orchestration example

- Exercise input delivery, output inspection, and conversation resume together in
  a real workflow configured by the user.
- Use the personal executor/reviewer cycle as an integration example, not as a new
  product feature. Instructions, handoff files, reports, and completion decisions
  belong to that example's user-defined conventions.
- Fill gaps in general-purpose session operations discovered by the experiment.
  Do not add ticket records, mandatory stage schemas, or automatic transitions.
- Add a small MCP usage skill only if tool descriptions are insufficient. Explain
  the operations and their limitations without prescribing an orchestrator policy.

**Acceptance:** a persistent orchestrator uses the available operations to manage
implementation, review, and revision in another session, according to user
instructions. No transfer into its workspace or built-in workflow engine is needed.

### 4. Reliability and usability on the existing dashboard

- Handle duplicate requests, bounded waits, cancellation, stale activation reports,
  and concurrent lifecycle operations without duplicate work or lost bindings.
- Define recovery of session metadata and conversation bindings after restart, respecting the current rule
  that processes stop when Kankuro exits.
- Provide bounded output inspection with explicit truncation/gap behavior; do not
  treat lossy preview subscriptions as a reliable communication channel.
- Show descendant activity even when the parent process is stopped, and make
  session state and operation failures understandable on the existing dashboard.
- Verify deletion, manual intervention, registration failure, and resume failure
  through real agent runs and appropriate automated checks.
- Document the supported CLI discovery/resume behavior and any limitations.

**Acceptance:** one orchestrator can manage multiple sessions on the current
dashboard, including an agent awaiting input and human intervention, with understandable
status and recoverable failures.

### 5. Multiple factory dashboards

Only begin after the single-dashboard workflow is useful and dependable.

- Give each factory its own dashboard and associated agents/sessions.
- Separate factory membership from session ancestry; parenthood does not dictate
  navigation or layout.
- Design factory navigation at this stage. The original left/right parent-child
  drill-down proposal in `feature.md` is superseded by this direction.
- Expect one orchestrator per factory in the usual workflow without making that a
  hard architectural constraint.

**Acceptance:** independent factories can be operated and inspected without mixing
their associated agents and sessions. Workflow conventions remain user-defined.

## Later extensions

- Agent-managed workspace and card configuration through MCP.
- Specialized native conversation discovery where generic registration is unreliable.
- Local collaboration and work-product integration when needed.
- Optional GitLab automation and a CLI frontend over the shared operations.
- Background operation independent of the TUI, if required by actual workflows.

## Details to settle during implementation planning

- Reliable native conversation-ID discovery for the first supported CLI, and the
  behavior when registration never arrives.
- Exact resume flags and whether resuming reloads the selected card's instructions,
  model, skills, and MCP configuration as intended.
- Caller credentials across relaunches, copied workspaces, and native conversation
  resume; credentials must identify the current activation.
- Input/output semantics sufficient for user-defined communication and handoffs,
  without introducing a workflow-specific data model.
- Treatment of saved conversation references when a card's driver changes.

## Feature briefs and review sequence

Proceed sequentially with a dedicated planner/reviewer and executor for each
feature. After each feature's implementation cycle, return the work for integration
review against this roadmap before starting dependent work. A feature may need
more than one execution session; these briefs do not prescribe implementation plans.

| Order | Brief | Roadmap coverage |
|-------|-------|------------------|
| 1 | [Shared application operations](feature_1.md) | Foundation for checkpoint 1 and later tools |
| 2 | [Native conversation continuity and minimal MCP](feature_2.md) | Checkpoint 1 |
| 3 | [Session ancestry and subtree lifecycle](feature_3.md) | Checkpoint 2 lifecycle foundation |
| 4 | [MCP discovery and session lifecycle control](feature_4.md) | Checkpoint 2 lifecycle tools |
| 5 | [Session input and bounded output inspection](feature_5.md) | Checkpoint 2 communication tools |
| 6 | [Validate and harden user-defined orchestration](feature_6.md) | Checkpoints 3 and 4 |
| 7 | [Multiple dashboards for independent factories](feature_7.md) | Checkpoint 5; deferred until feature 6 is accepted |

Each feature includes acceptance criteria and a review handoff. Keep `project.md`
current throughout, and carry unresolved decisions and validation limitations into
the handoff so later sessions do not need to reconstruct this conversation.
