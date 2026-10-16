# Feature 5: MCP message delivery and explicit replies

Status: independent integration review accepted on 2026-10-25 for commit
`b95b853`. Feature 6 may proceed.
The implementation handoff reports that automated checks and the real
`tui-test` CLI acceptance matrix passed with Codex CLI 0.164.0 and OpenCode 2.1.4.
Codex's two-phase text/500 ms pause/Enter submission passed attached and detached
sends. Both recipients registered their actual native IDs and reported through
MCP themselves. The user's corrected Codex permission defaults worked in managed
sessions without CLI overrides. See `review_6.md` and
`/tmp/kankuro-feature5-e2e6/verification.md` for lifecycle, deletion, restart,
interruption, screenshots, recordings, and sanitized MCP evidence.

Dependency: Feature 4 accepted.

## Integration acceptance — 2026-10-25

The test-isolation finding below is resolved: cleanup restores the previous
contents of both built-in skills instead of clearing shared state. The exact
previously failing sequence passed five repetitions under the race detector;
the full API package passed two repetitions normally and under the race detector.
Build, vet, full normal/race suites, lint (0 issues), and diff checks also passed.
Delegated verification logs survived the usage-limit interruption and were checked
before acceptance; see `.kankuro/review-feature5-fix/verification.md`.

This completes the independent review of the messaging implementation and its
previously inspected Codex/OpenCode evidence. No new native CLI run was needed for
this test-only correction. The earlier finding below is retained as review history.

## Independent integration review — 2026-10-24

Reviewed commit: `aa5d6b8`. Code review and audited native CLI evidence support
the agreed explicit-message protocol, including attached delivery and recipient
MCP reports. No production defect was established in this review.

**P2 — Restore shared built-in skill state in test cleanup.**
`TestMessagingOperationsPinActivationsAndPersistExplicitReports` clears the
`kankuro-puppet` content in cleanup (`internal/api/service_test.go:1740`), although
`TestMain` initializes it only once. Other messaging fixtures initialize only
`mcp-cli` and rely on that shared puppet content. Repeated runs then fail during
setup rather than exercising the messaging behavior. Deterministic reproduction:

```sh
go test -race -count=2 ./internal/api -run 'TestMessagingOperationsPinActivationsAndPersistExplicitReports|TestReplyWaitReturnsWhenRecipientHistoryWasDeleted'
```

The second iteration fails with `built-in skill content not found:
kankuro-puppet`. Restore the previous global contents on cleanup, or make the
fixtures consistently own initialization/restoration of both skills. Verify the
reproducer and repeated messaging checks before advancing. This is a test-fixture
issue, not an observed production messaging failure. Independent evidence and logs
are retained in `.kankuro/review-feature5-current/`.

## Intent

Let an orchestrator send messages to managed sessions and receive explicit replies
through MCP, without reading verbose CLI output. Terminal input delivers the
message; MCP carries the recipient's reports. Humans can remain attached to observe
and intervene without blocking orchestration.

This brief supersedes earlier roadmap references to terminal-output inspection as
the communication mechanism for Feature 5.

## Scope and constraints

- Expose an MCP operation for sending a message to a managed session, reusing the
  shared service and authenticated dashboard scope from Feature 4. Kankuro assigns
  a message ID and injects the message with a recognizable header identifying it as
  orchestrator input and explaining how to reply through MCP. The header is a
  protocol cue, not proof of authority; authorization remains credential-based.
- Add a built-in skill explaining the messaging protocol. A recipient replies
  through MCP, referencing the message ID, rather than relying only on its TUI chat.
  Support completion reports, questions, and reports that work cannot proceed,
  without prescribing executor/reviewer stages or a workflow-specific schema.
- Put prompt submission in the driver interface. The generic implementation writes
  the marked message and submits it with Enter through the existing locked input
  path; driver-specific implementations may adapt multiline input and submission
  behavior. Treat each submission as one coordinated driver operation and document
  the supported behavior with real CLIs.
- Allow delivery while a human is attached. Attachment does not reserve input or
  pause orchestration; humans retain their normal ability to interrupt the agent.
  Coordinate writes to prevent byte interleaving, but explicitly document that this
  does not clear text already in the CLI input buffer. Automated prompts may arrive
  during human typing; do not promise conflict-free simultaneous prompt composition
  without CLI-specific support.
- Separate acceptance, process submission, and explicit agent replies. Writing the
  prompt and Enter is not proof that the CLI processed it or that work completed.
  A completion report means the agent declares that it finished responding to that
  message, not that Kankuro approved the work or advanced a workflow stage. The skill
  is an instruction, not a guarantee: absent a report, a message stays unanswered.
- Store explicit replies in Kankuro and expose bounded MCP retrieval, with message
  correlation, a cursor, and optional bounded waiting. Define size/retention limits,
  pagination, and any expiry/gap behavior. Do not consume terminal output, scrape
  the visible screen, or use preview subscriptions to obtain agent responses.
- Give sends stable caller request IDs and define retry/deduplication behavior so
  retries cannot silently submit a prompt twice. Bind delivery and replies to the
  appropriate session and activation; a delayed send must not reach a replacement
  card after relaunch or switch. Derive reply sender identity from its authenticated
  activation, and define how late replies are handled without misattribution.
- Bound waits and handle process termination, cancellation, shutdown, relaunch, and
  card switching explicitly. Terminal silence and heuristic `waiting` status are
  neither completion signals nor evidence that a message was processed.
- Kankuro handles message delivery and storage, not creation of feature files,
  plans, or folders beyond existing workspace preparation. Continue exposing actual
  session workspace/sub-workspace paths so the orchestrator can create or inspect
  artifacts directly using its own filesystem tools.
- Keep workflow conventions in user instructions/files. No mandatory stage schema,
  ticket records, review statuses, or automatic transitions.

## Acceptance criteria

- An orchestrator sends an instruction; the recipient CLI receives its marked
  prompt and submission, then follows the built-in skill to report through MCP.
  The orchestrator retrieves the correlated reply without reading CLI output.
- A recipient can ask a question or report inability to proceed through the same
  protocol. Missing replies never become inferred completion.
- Messages are delivered while a human is attached; observation does not block
  orchestration and normal human interruption remains available. Coordinated
  submission and the limitation around already-typed text are demonstrated.
- Tests cover duplicate requests, stale activations, reply identity/correlation,
  process termination, concurrent submissions, bounded reply retrieval, and
  cancellation/shutdown during waits. Document partial/uncertain submission outcomes
  rather than claiming processing or safe replay without evidence.
- The orchestrator can reference files it creates directly in the exposed session
  workspace; Kankuro needs no new artifact-generation or filesystem-writing tool.

## Planning decisions

Choose tool names/schemas, header format, built-in skill deployment, and driver
submission behavior. Define reply access permissions (including recipients that
do not have Dashboard Control), correlation, delivery ordering, retry semantics,
retention/durability across restart, cursor behavior, and bounded waits. Reuse
existing activation identity and lifecycle infrastructure. Keep the protocol small
and general-purpose; explicit reports carry user-defined meaning rather than
driving a built-in workflow engine.

## Out of scope

Terminal-output reading through MCP, terminal scraping/emulation, inferred task
completion, new artifact/file/folder creation tools, workflow engines, autonomous
scheduling, merging work products, and new dashboards. Existing human-facing
terminal output and preview behavior remain available.

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
