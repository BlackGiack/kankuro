# Step 7 independent review

Decision: **CHANGES REQUIRED**. `review.md` was created because two required
behaviors remain incomplete. No implementation fixes or commit were made.

Reviewed 2026-11-17 against `feature_7.md`, `impl_plan.md` and
`feature_7_evidence.md`, including the actual contracts, callers, moved code/tests,
backend changes, architecture exceptions and main composition. Feature 6 already
has final-review PASS. The following findings block approval and feature 8.

## 1. [P1] Record output completion before accepting another attachment

Locations: `internal/session/io.go:31`–36 and `internal/session/attach.go:55`–60;
consumer: `internal/api/attachment.go:Next`.

The output tracker closes the subscribers present at EOF, deletes them from the
map, and exits without recording that this generation's stream has finished.
`attachChecked` subsequently accepts a subscription whenever session status is not
terminated. Stream completion and process/status completion are independent:
`GenericCLIDriver.streamOutput` and `watchProcess` run separately, and the manager
watcher can lag behind either. A live process can also close its terminal streams.

If attachment subscribes after tracker completion, the new channel has no producer
and nobody left to close it. `Next` now depends entirely on this channel (it does
not observe the handle's Done), so it waits until explicit cancellation/Close
instead of producing EOF. Even later process completion does not repair it because
the watcher does not close subscribers. Removal also skips closure when Output is
non-nil. This violates the attachment completion and resource-lifetime requirements.

Independent deterministic probe failed: construct a running session with its own
subscriber map and a closed output channel; run `trackOutput` to completion; call
`AttachChecked`; it succeeds, but the returned output channel remains open and
empty. The probe leaves Done open to model output EOF preceding process exit. No
sleep or scheduling assumption is needed.

Required change: persist generation-specific stream completion under the same
synchronization as subscriber admission. An attach after completion must either
reject the finished stream or return a resource whose output immediately completes;
it must never install an orphaned subscriber. Preserve buffered trailing output for
already subscribed attachments and isolate old/new activation maps. Add regressions
for attach before/during/after EOF, EOF before process completion, and the watcher-lag
window, including removal/relaunch and cancellation cleanup.

## 2. [P2] Consume deletion execution outcomes in the frontend

Locations: `internal/ui/coordinator.go:254`–269 and 862–877;
producers: `internal/ui/dashboard_list.go` confirm/retry command adapters.

The new commands populate `DashboardMutationFinishedMsg.Progress` and
`ScopedActionFinishedMsg.Progress`, but both coordinator handlers ignore Progress
entirely, set only the transient Err banner and refresh inventory. They return
handled=true, so views never receive these results either. No production frontend
consumer reads DeletionOutcome. The outcome/progress migration therefore ends at
the message envelope, contrary to the plan's requirement to retain execution
results rather than discard them.

A known durable pending deletion, a failure before intent and uncertain persistence
produce identical presentation when their error text is the same. Checkpoint facts
returned with an execution failure are lost. If the subsequent inventory/status
read fails, the frontend cannot retain the already returned progress or explain
whether retry resumes an existing intent versus confirmation needing review. An
inventory refresh cannot replace the authoritative execution outcome, especially
for uncertain effects.

Independent coordinator probe failed: deliver a current scoped deletion result with
Outcome=pending and an ordinary cleanup error; capture View(); deliver the same
current result/error with Outcome=no_intent; View() is identical. Inspection confirms
the same discard in dashboard deletion. This does not require an invalid nil-error
pending result or executing a destructive operation.

Required change: map current deletion progress into scoped presentation state,
including semantic outcome, useful checkpoints and diagnostic/recovery guidance.
Retain it across failed follow-up reads and invalidate it for the appropriate target
or visit. Handle card and dashboard results and distinguish no-intent, durable pending,
completed and uncertain effects; do not infer completion from missing inventory rows
or automatically replay uncertain effects. Add coordinator-through-view regressions
for those outcomes, post-effect failure, failed refresh/status reads and stale results.
Backend authority, confirmation and retry policy must remain in the service.

## Verified implementation and preserved boundaries

- `internal/app` is absent; coordinator, command adapters and tests live in
  `internal/ui`. Comparison with HEAD distinguishes the mechanical move from
  behavioral edits; the old legacy adapter is removed, with no forwarding package.
- Attachment and deletion contracts contain independent application/stdlib types;
  frontend ports no longer expose manager/PTY/concrete service types. Main supplies
  capabilities and startup diagnostics and forwards errors through `ui.RuntimeError`.
- Terminal raw mode, human resize observation, detach and protocol restoration live
  in `ui/terminal`; backend resize and serialized driver writes stay behind the
  attachment resource. Existing checked-in terminal/backend PTY tests pass.
- Attachment owner handles supersession, navigation, late publication and shutdown;
  deletion workers cancel/join on frontend shutdown. The EOF finding above remains
  despite these ownership improvements.
- Backend token expiry, revision/impact/scope revalidation, foreign confirmation,
  durable intent/checkpoint/retry and restart/removed-row recovery remain owned by
  existing service/session/config code. Full-suite tests exercise these paths.
- Main retains startup and runner → session → MCP teardown ownership/order.
- Production migration exceptions are removed. The import allowlist retains only
  the documented feature-8 test-support `ui/uitest/shape.go` entry; transitive type
  exposure allowlist is empty. Architecture checks pass with root UI enforcement.
- Existing feature-6 center implementation/tests moved without behavioral edits;
  full ordinary and affected race checks pass, including report/cache regressions.

## Independent verification

Passed on this checkout:

```text
go test ./... -count=1
go test -race ./internal/ui/... ./internal/api ./internal/session -count=1
go build ./...
go vet ./...
go test ./internal/architecture/... -count=1
golangci-lint run --new-from-rev=HEAD
git diff --check
```

Changed-code lint reported zero issues. Full lint was not rerun and its earlier debt
is not claimed resolved. The two independent behavioral probes used temporary Go
overlays over existing test files, leaving repository source/tests untouched:

```text
go test -overlay /tmp/kankuro-step7-review-probes/overlay.json ./internal/session ./internal/ui -run '^TestReview' -count=1
FAIL TestReviewAttachAfterOutputTrackerCompleted:
  attachment accepted after output completion but its subscription never closes
FAIL TestReviewDeletionOutcomeReachesPresentation:
  durable pending deletion and no-intent failure produce identical presentation; outcome discarded
```

The temporary probes are review experiments, not checked-in regression coverage.
The checked-in suite's success does not resolve these failed behavioral checks.
No manual/native-model E2E or non-Linux verification was performed by this reviewer;
implementation evidence's manual fixtures are not claimed independently reproduced.

## Preservation and handoff

Existing workflow documents and Appedia evidence were fingerprinted before review
and checked unchanged afterward. Only `review.md` is added by this review; prior
implementation changes remain in place. No external blocker or product question
requires user clarification. Resolve both findings and request independent re-review
before approval or feature 8. No commit is authorized.
