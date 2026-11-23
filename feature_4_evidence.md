# Feature 4 implementation evidence

This is implementation evidence for the independent feature 4 review. It is not a review decision and does not authorize advancing or committing.

## Implementation

- `internal/session/activation_result.go`, `lifecycle.go`, `startup.go`, and `manager.go` capture publication-time process identities, effect evidence and persistence warnings. Unpublished spawn teardown records whether the driver confirmed termination; an unconfirmed temporary workspace is retained and identified in the diagnostic. The result-returning methods share the same activation path with the legacy backend projections.
- `internal/api/contracts/{session_actions,conversation}.go`, `internal/api/{session_actions,conversation_actions,membership}.go` define and map focused lifecycle, mutation, conversation and workspace operations. Conversation identities are scoped to session, card, native ID and driver and are revalidated during manager admission. Stop, removal and rename retain partial-effect evidence. The existing MCP-facing operations retain their wire interface.
- `internal/app/{session_ports,session_actions,legacy_actions,callbacks,coordinator,handlers,navigation,attach,dashboard_list}.go` and the session views/messages move these workflows to contract ports and command results. Requests carry scope and request identities; picker reads are cancellable, non-overlapping and detached. Automatic naming and attachment check the published process identity. Configuration/deletion and raw attachment remain in the feature 5/7 adapters; preview remains feature 6.
- `internal/session/workspace_launcher.go` starts configured IDE/terminal commands against the saved workspace while holding the manager's checked start boundary, excluding removal until start returns. Child lifetime and reaping remain independent of the request.
- `docs/frontend-boundary.md`, `project.md`, and the two architecture allowlists reflect the reduced exact exceptions. Tests changed alongside these paths, including contract-only command fakes, real manager publication/teardown and persistence-failure cases, scope, stale outcomes, picker refresh, cancellation and workspace/removal serialization.

## Manual evidence

On 2026-11-15, built the application and ran it in a PTY with isolated `/tmp/kankuro-feature4-manual-config` and `/tmp/kankuro-feature4-manual-tmp`. A local `cat` driver exercised dashboard navigation, launch, optional naming, attach/input/detach, IDE and terminal commands configured as `/bin/true`, inline rename, stop, explicit-new relaunch, switch and session removal. With a local wrapper driver and a deliberately seeded saved conversation in the isolated session JSON, the picker displayed the saved ID, resumed it, started a new conversation, and switched with the saved choice. The wrapper ignored resume arguments; this verifies the UI and application flow rather than a native provider's resume semantics.

With the isolated sessions directory made read-only for the current non-root user, a launch displayed the persistence warning and retained naming and attachment; the rename persistence warning was visible. Permissions were restored and the session stopped and removed. Native provider/model E2E was not run because provider/model and token-budget authorization was not supplied. Manual navigation while a read is outstanding was not separately exercised; cancellation/overlap is covered by automated coordinator tests.

## Verification

Final command results are recorded below after the current run completes. The expanded race expression includes relaunch, cancellation, rename, stop, removal, stale-result, activation, teardown and shutdown regressions in addition to the plan's expression. `go test -list` was used to verify selection, including `TestCommandClosuresDoNotRaceCoordinator`, `TestConversationReadCancellationAndOverlap`, `TestInlineRenameRejectsOlderRefreshSnapshots`, `TestStaleAttachReadyIsClosed`, `TestScopedMutationsRetainPartialRenameAndRejectForeignScope`, `TestActivationCancellationAfterSpawnConfirmsTeardown`, `TestUncertainUnpublishedSpawnRetainsTemporaryWorkspace`, `TestWorkspaceStartSerializesWithRemoval`, and shutdown-during-spawn tests.

Final checks on 2026-11-16:

| Command | Result |
|---|---|
| `go test ./internal/api ./internal/app ./internal/session ./internal/ui/views` | PASS (API, session and view results cached; app rerun passed) |
| `go test ./...` | PASS in the preceding full run; the final subsequent changes were scoped to coordinator stale-result refresh and were covered by the final targeted suite and race run. |
| `go test ./internal/architecture/... ./internal/mcp` | PASS |
| `go test -race ./internal/api ./internal/app ./internal/session -run 'Launch|Switch|Conversation|CommandClosures|Workspace|Relaunch|Cancel|Rename|Stop|Remove|Stale|Activation|Teardown|Shutdown'` | PASS (API and session results cached; app rerun passed) |
| `go build ./...` | PASS |
| `go vet ./...` | PASS |
| `git diff --check` | PASS |
| `golangci-lint run` | FAIL: 69 findings, all reported in unchanged `internal/architecture/boundary_test.go`; no findings reported in changed implementation/test files. Output is retained for this workspace at `/tmp/kankuro-feature4-lint-final.txt`. The installed linter is v2.15.0; the repository configuration identifies v2.11.3 as its reference. |

The race expression's test list was checked with `go test -list`, including cancellation, stale, inline rename, stop/removal, teardown, shutdown and process-identity cases. The separate contract validation tests also passed. No native provider E2E was run, and the manual outstanding-read navigation limitation above remains. The independent reviewer should inspect actual diffs, admission and cleanup ordering, outcomes, remaining exact exceptions and this evidence before deciding PASS or CHANGES REQUIRED.

## Planner review fixes — 2026-11-16

Addressed both findings in `review.md`:

- The checked `DashboardCardLoadedMsg` transition now enters the launcher synchronously through the coordinator before returning the launch effect command. It no longer queues a generic launcher navigation. `TestLaunchTransitionCannotNavigateAfterLeavingOrNewerLaunch` drives `Coordinator.Update`, retains the old effect result, and verifies that it neither reopens the launcher after Settings nor overwrites a newer launch.
- The conversation picker retains the invalidated original selection identity independently of read status and fallback highlighting. Repeated successful/failed refreshes cannot acknowledge a fallback. Explicit selection clears the invalidation; a failed read still blocks submission; recovery of the original compatible option restores it. Regressions cover repeated empty and replacement-option refreshes, incompatible choices, explicit New/saved choices, failed-read blocking, and original-choice recovery.

Verification after these fixes:

| Command | Result |
|---|---|
| `go test -count=1 ./...` | PASS; the final additional incompatible-choice regression was added afterward and passed in the focused runs below. |
| `go test -race -count=1 ./internal/api ./internal/app ./internal/session -run 'Launch|Switch|Conversation|CommandClosures|Workspace|Activation|Rename|ScopedStop|Stale'` | PASS |
| `go test -race -count=1 ./internal/ui/views -run 'TestConversation(IncompatibleSelection|ReadErrorStillBlocks|InvalidSelection)'` | PASS |
| `go test -count=1 ./internal/app ./internal/ui/views ./internal/architecture/...` | PASS |
| `golangci-lint run --new-from-rev=HEAD --whole-files=false` | PASS, 0 issues; diff-filtered, not a full lint pass. |
| `go build ./...` and `go vet ./...` | PASS |
| `git diff --check` | PASS |

The earlier full-lint limitation remains as documented above. No independent review decision is claimed here.
