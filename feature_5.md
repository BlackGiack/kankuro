# Feature 5: Session input and bounded output inspection

Status: proposed; implementation and integration review pending.

Dependency: Feature 4 accepted.

## Intent

Let an agent communicate with managed sessions and inspect their output, while
humans retain the ability to interact with those sessions safely.

## Scope and constraints

- Expose input delivery through the existing locked driver input path. Specify text
  submission versus raw keystrokes, multiline behavior, and acknowledgement semantics.
  Accepted bytes are not proof that an agent processed a request.
- Define input ownership between attached humans and MCP callers. Prevent accidental
  interleaving; make conflicts visible rather than silently overwriting input.
- Provide bounded output reads with documented cursor/activation and truncation/gap
  semantics. Do not rely on lossy preview subscriptions for complete communication.
- State whether returned output is raw terminal data, sanitized text, or a rendered
  screen. Do not claim ANSI stripping reconstructs the visible screen.
- Allow inspection while a process is running or waiting. Waiting remains heuristic;
  it is not a success signal or a mandatory gate for reading output.
- Bound any waits and handle termination, cancellation, relaunch, and card switching
  without delivering stale input or confusing output from different activations.
- Keep completion conventions in user instructions/files. No mandatory stage schema,
  ticket records, review statuses, or automatic transitions.

## Acceptance criteria

- A caller sends an instruction to a managed session and retrieves its response
  using the documented read semantics, including long/truncated output.
- Human attachment and concurrent MCP input follow the documented ownership policy.
- Tests cover stale activations, process termination, output gaps, input conflicts,
  and shutdown during waits. Terminal silence does not imply completion.

## Planning decisions

Choose the smallest input/read contract that supports real CLIs. Prefer explicit
limits over an unbounded transcript or full terminal emulation unless the latter is
necessary for the demonstrated use case.

## Out of scope

Workflow engines, autonomous scheduling, merging work products, and new dashboards.

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
