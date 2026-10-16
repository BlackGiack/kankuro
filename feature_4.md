# Feature 4: MCP discovery and session lifecycle control

Status: integration review accepted on 2026-10-24 for commit `7695a4c`.
Feature 5 may proceed.

Dependency: Feature 2 accepted. Feature 3 is canceled; proceed directly here.

## Integration acceptance — 2026-10-24

The card-deletion finding below is resolved. Shared manager reservations protect
create and destination-card switch preparation across separate UI/MCP services;
existence validation and deletion use the same lock, and reservations release on
failure/cancellation. Review found no blocking issue in the correction.

Independent delegated verification passed build, vet, full normal/race suites,
lint (0 issues), and diff checks. Card-reference regressions passed ten normal
and race repetitions; MCP registration/lifecycle cleanup tests passed twenty
normal and race repetitions without the previously observed shutdown timeout.
Exact commands and logs are in `.kankuro/review-feature4-current/verification.md`.
No new manual TUI run was needed for this backend/test-cleanup correction.
Earlier review and implementation entries below are historical.

## Integration review — 2026-10-24

**P2 — Card deletion can race an MCP activation.** `Service.DeleteCard` checks
only published session snapshots before deleting the card. During an MCP create's
workspace preparation, the target session is not yet in that list. A human can
therefore delete the card while preparation is in progress; creation subsequently
succeeds with a reference to the deleted card, and relaunch fails with `card not
found`. A controlled, gated workspace-copy test reproduced this sequence through
the real service and operation runner.

Coordinate card deletion with activation admission and card references across
both UI and MCP. These frontends currently have separate `api.Service` instances,
so a mutex private to one service instance is insufficient. Cover pending creates
and switches to a destination card, and ensure deletion cannot slip between the
reference check and activation admission. Add a deterministic regression proving
that concurrent deletion/activation cannot leave a session referencing a deleted
card. This was a documented limitation before MCP lifecycle control; concurrent
MCP activation now makes it relevant to this feature's acceptance.

The previously documented intermittent MCP registration-test shutdown timeout also
recurred during targeted verification. Keep it visible as a test reliability issue;
do not report that every verification invocation passed.

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
- Define explicit caller scopes for which sessions can be inspected or controlled
  on the current dashboard. Authenticate callers without recording parentage or
  introducing an interim ownership tree. Later, factory membership and its
  orchestrator association will provide the grouping and control scope.
  Caller-supplied session IDs alone must not grant authority.
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
- Unauthorized calls, retry behavior, individual-session deletion guards, and partial-success outcomes
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

## Implementation closeout — 2026-10-24

All five groups in `review_4.md` are implemented and verified:

1. **Deployed permission contract:** grant persistence across relaunch and manager
   reload, old-token revocation, explicit same-card/destination-card switch behavior,
   legacy missing-field false default, and actual credential authorization.
2. **Cancellation/lifetime:** gated temporary-copy cancellation, caller revocation,
   shutdown during preparation, cancellation after MCP config publication/before
   spawn, unpublished process teardown, reservation release, temporary-copy cleanup,
   bounded queue/disconnected wait, and cancellation-too-late publication outcome.
3. **Final journal write failure:** the real create effect returns its stable target
   ID with a durability warning, in-memory replay does not repeat it, and restart
   marks the durable nonterminal operation interrupted without executing it.
4. **MCP acceptance:** real service/manager/runner/SDK/HTTP coverage includes
   cross-project main/sub-workspace paths and retained caller-created files, A/B/A
   saved binding selection, registration-only and stale-credential lifecycle denial,
   plus delete partial-success flags and replay.
5. **Native CLI:** isolated Codex CLI 0.164.0 acceptance created and registered two
   real native IDs, selected A after A→B→A, observed Kankuro spawn `codex resume
   <A-ID>`, and retained a file created on B at the same temporary workspace path.
   Codex's sandbox blocked its in-agent `mcp-cli` network attempt; both canonical
   registrations succeeded through host `mcp-cli` with `saved` persistence.

Verification passed: `go build ./...`, `go vet ./...`, `go test ./...`,
`go test -race ./...`, `golangci-lint run` (0 issues; configuration/deprecation
warnings only), and `git diff --check`. No commit was made.

Native acceptance record and raw Codex event logs (isolated test workspace):
`/tmp/kankuro-feature4-native/artifacts/native-acceptance-result.json`,
`codex-A.jsonl`, `codex-B.jsonl`, and `codex-A-resume.jsonl`. The explicit
`codex exec resume` probe collided with the active-writer lock because Kankuro had
already resumed A; the app-spawned `codex resume <A-ID>` and selected binding were
observed directly. This does not affect lifecycle acceptance, but should be reviewed
as part of integration review. No production config was used.

Implementation complete is not integration acceptance. Hand this state to an
independent integration reviewer before starting dependent feature work.

## Targeted visual corrections — 2026-10-24

The `review_5.md` dashboard/preview findings are corrected without changing
Feature 4 scope or status. Preview metadata is fitted to a reserved text slot
beside complete art, with art omitted when it cannot fit usefully. Card rows have
balanced gutters and dashboard pane baselines remain stable across selection,
scroll/search, empty states, and resize. Regression coverage includes cell
positions and long/Unicode metadata survival. The supplied review overlay checks
pass.

Final verification passed: `gofmt`, `golangci-lint run` (0 issues), `go test ./...`,
`go test -race ./...`, `go vet ./...`, `go build ./...`, and `git diff --check`.
Real TUI captures at 80×24, 120×40, and 160×45 were opened and inspected; see
`/tmp/opencode/kankuro-review5-visual/artifacts/`. The narrow capture omits preview
art rather than clipping it; the two wider captures show the intact art and aligned
pane content. The TUI exited normally.

Race-test observation: an earlier full race run intermittently failed
`TestStreamableMCPRegistrationAndAuthentication` at `internal/mcp/server_test.go:131`
with `shutdown MCP: context deadline exceeded`. Five isolated repetitions and a full
rerun passed; the final full race run also passed. Cause is undiagnosed, and cleanup
errors remain unsuppressed. Status remains implementation complete / ready for
integration review; integration acceptance is not claimed, and Feature 5 has not
started.

## Empty-preview follow-up — 2026-10-24

The `review_6.md` empty-state defect is corrected: `SessionPreview` now uses its
provided width directly rather than deducting border columns already handled by
the dashboard. Placeholder text has no structural indentation, wraps at word
boundaries, and remains centered. A maintained 8×19 regression verifies complete
`session` and `selected` words and bounded output; empty dashboard tests verify
readability and vertical centering at 80×24, 120×40, and 160×45. The supplied
overlay regression passes.

Required checks passed again: `go test ./...`, `go test -race ./...`, `go vet ./...`,
`go build ./...`, `golangci-lint run` (0 issues; configuration warnings only), and
`git diff --check`. Fresh populated and empty captures at all three dimensions were
opened and inspected in `/tmp/opencode/kankuro-review6-final/artifacts/` and
`/tmp/opencode/kankuro-review6-empty/artifacts/`; both isolated applications exited
normally. The intermittent MCP shutdown timeout did not recur. Feature 4 remains
implementation complete / ready for integration review; acceptance is not claimed,
and no dependent feature work was started.

## Integration correction — 2026-10-24

The `review_7.md` card-reference race is corrected. Pending card references are
owned by the shared `session.Manager`, not an individual `api.Service`, so UI and
MCP services coordinate through the same lock. Admission checks current card
existence under that lock and records a pending reference before capability
resolution/workspace preparation. Delete checks pending activations plus published
session card/conversation references and performs the repository deletion under
the same lock. Reservations are released on failure/cancellation and held through
manager publication on launch or switch. This covers UI launch/switch and MCP
create/switch without holding the manager lock through workspace copying or spawn.
Refusal errors distinguish activations in progress; terminated-session and saved
conversation references remain protected.

Deterministic regressions use gates rather than sleeps and two `api.Service`
instances sharing one manager/repository: MCP create during temporary copy blocks
UI deletion and then permits stop/relaunch; MCP switch before destination
publication blocks deletion for a card with no saved binding; deletion-first
rejects a stale loaded card snapshot; and canceled preparation releases the
pending reference for later deletion. Published references remain protected after
termination. The targeted tests passed normally and under `-race`.

MCP shutdown reliability was investigated separately. Before client cleanup
changes, the focused test failed twice in 20 normal repetitions with the original
5-second shutdown deadline, and the full race suite also failed
`TestRealServiceLifecycleResultsOverMCP` at shutdown. Tests now own per-client
HTTP transports, close client sessions before runtime shutdown, and report close
errors instead of discarding them. The focused test passed 20/20 normally and
20/20 with `-race`; the whole MCP package passed five repeated race runs and the
final full race suite passed. Client/test cleanup ownership is the observed
condition; no production root cause is inferred. Failure and rerun output is
retained under `/tmp/opencode/kankuro-review7/`.

Final required checks passed: `go test ./...`, `go test -race ./...`, `go vet ./...`,
`go build ./...`, `golangci-lint run` (0 issues; configuration warnings only), and
`git diff --check`. Feature 4 is ready for another integration review; acceptance
is not claimed, and Feature 5 has not started.

## Verification correction — 2026-10-24

The `review_8.md` test-ordering issue is corrected in
`TestOperationQueueBoundAndDisconnectedWaitReturnPromptly`. After preserving the
queue-full rejection and waiting for the disconnected operations' terminal
outcomes, the test now shuts down and joins the capacity-phase runner before
creating a fresh runner for the publication/late-cancellation phase. This ensures
all queued tasks have been consumed; it does not infer worker readiness from
terminal-state publication.

The exact reproducer passed: `GOMAXPROCS=1 go test -race ./internal/api -run
'^TestOperationQueueBoundAndDisconnectedWaitReturnPromptly$' -count=50` (50/50).
`go test -race ./...` and `go test ./...` passed, as did `go vet ./...`,
`go build ./...`, `golangci-lint run` (0 issues; configuration warnings only), and
`git diff --check`. Feature 4 remains implementation complete / ready for another
integration review; acceptance is not claimed and Feature 5 remains blocked.
