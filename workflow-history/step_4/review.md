# Step 4 review

Decision: **CHANGES REQUIRED** — 2026-11-16.

Reviewed against `feature_4.md` and `impl_plan.md`. Created this file because the
two behavioral findings below need fixes before approval. No implementation files
were edited and no commit was made. No user decision or environment blocker is
needed; these findings block advancement.

## Required changes

### 1. [P2] Guard delayed launch navigation with its request identity

Locations: `internal/app/callbacks.go:145–148`,
`internal/app/coordinator.go:327–334`.

`launchCard` batches the effect command with an ordinary `ScopedNavigateMsg` to
the launcher. That navigation carries only dashboard/generation. Leaving the
dashboard invalidates `launchRequest`, but dashboard generation stays unchanged
for ordinary modal navigation. A queued launch navigation therefore still passes
the generic navigation handler after the user opens another screen.

Reproduced through `Coordinator.Update`: accept `DashboardCardLoadedMsg` for the
current launch request, retain its batched launcher-navigation message, navigate
to Settings, then deliver the retained message. Actual: current view becomes
`ViewSessionLauncher` (9). Expected: remain in Settings. The obsolete launch's
completion is separately rejected by the request check, so the reopened launcher
can also remain on “Launching session...” with no accepted completion.

Make the launch navigation part of the same checked request transition as launch
admission. For example, navigate in the already checked card-loaded handler before
returning its effect command, or use a launch-specific navigation envelope checked
against request/card identity and the expected source view. Preserve the rule
against navigation inside a view's Update. Add a regression through Update that
delivers old navigation after leaving the dashboard and after a newer launch.
Do not weaken completion filtering to make the obsolete launcher usable.

This violates feature 4's stale-result/navigation requirements and the plan's
requirement that earlier requests cannot navigate over later user interactions.

### 2. [P2] Keep an invalidated conversation choice blocked across refreshes

Locations: `internal/ui/views/conversation_choice.go:118–143,159–162,259–264`,
`internal/app/coordinator.go:854`.

When the selected saved option disappears or becomes incompatible, `RefreshContext`
first replaces it with the first compatible option (or New), then sets
`selectionInvalid` and blocks Enter. The next ordinary refresh sees that fallback
as the prior selection and clears the error. `SetContextError("Refreshing
conversations")` also resets `selectionInvalid` before the read. No explicit user
choice is required to clear the invalidation.

Reproduced: open the picker with one compatible saved option selected; refresh
successfully with no saved options; Enter is correctly blocked. Start/complete a
second periodic refresh with the same empty result; press Enter without moving
selection. Actual: OnSelect is called with New. Expected: the vanished saved
selection remains invalid until the user explicitly chooses an available option
or cancels/reopens. The same pattern can silently substitute another saved
conversation when alternatives exist.

Preserve invalidated selection identity/acknowledgment independently from the
temporary read/error status and the list's fallback highlight. Periodic reads
must not acknowledge the replacement. Clear the invalidation on an explicit user
selection (including a deliberate New choice), cancellation/reopen, or recovery
of the actual original choice as appropriate. Add regressions for two or more
refreshes, including empty options and a different compatible saved option.

This violates the plan's explicit missing/incompatible-selection blocking rule
and the feature's explicit new-versus-resume interaction requirement. It does not
bypass backend scope checks, but it can execute a different conversation intent.

## Review coverage and positive evidence

- Inspected manager publication and teardown in `internal/session/lifecycle.go`,
  `activation_result.go`, `startup.go`, `manager.go`, `conversation.go`,
  `attach.go`, and `workspace_launcher.go`, including legacy callers. Process
  identities are allocated independently of MCP authority, copied under the
  publication lock before persistence, and excluded from persisted SessionData.
  Metadata failures preserve published evidence. Unconfirmed unpublished teardown
  retains uncertainty and its temporary workspace. Reservation/release and
  rollback guards distinguish publication from non-nil errors.
- Inspected independent contracts and application mapping in
  `internal/api/contracts/{session_actions,conversation}.go`,
  `internal/api/{session_actions,conversation_actions,membership}.go`, including
  existing `operations.go` callers. Conversation selection is scoped by
  session/card/native ID/driver and validated under activation admission.
  Rename/stop/remove preserve scoped checks and partial-effect facts. Workspace
  commands resolve saved paths/configuration behind a checked start boundary.
  Existing MCP operation results remain behind backend projections.
- Inspected migrated command ports, callbacks, coordinator/message handlers,
  navigation, launcher, switch and conversation views, and expected-process
  naming/attachment guards. The old full-card launcher read and frontend
  sentinel/driver-comparison paths were removed. Exact dependency and exposure
  checks pass; configuration, preview and raw attachment exceptions retain their
  feature 5/6/7 ownership. These remaining exceptions do not excuse the findings.
- Inspected backend publication/teardown and API contract regressions, command
  fakes, picker/stale-result tests, lifecycle warnings, workspace command and
  rename refresh tests. The existing suites miss the two reproductions above.
  Actual relaunch/switch persistence-warning subtests passed without skips.

## Independent verification

All of the following passed on the reviewed checkout:

```text
go test -count=1 ./internal/api ./internal/app ./internal/session ./internal/ui/views ./internal/architecture/... ./internal/mcp
go test -count=1 ./...
go test -race -count=1 ./internal/api ./internal/app ./internal/session -run 'Launch|Switch|Conversation|CommandClosures|Workspace|Activation|Rename|ScopedStop|Stale'
go test ./internal/api/contracts -count=1
go test ./internal/app -run '^TestPersistFailureWarningReachesViewOnBothPaths$' -count=1 -v
go build ./...
go vet ./...
golangci-lint run --new-from-rev=HEAD --whole-files=false
git diff --check
```

The lint command reported zero issues; it is diff-filtered, not evidence of a
clean full lint run. The implementation's separate evidence reports existing
full-lint debt in unchanged architecture-checker code; full lint was not rerun
for this review. Race selection was inspected with `go test -list`; the selected
cases include command closures, activation cancellation, workspace/removal
serialization, conversation refresh and stale results. The full ordinary suite
also covers tests outside that race expression.

Two temporary Go-overlay regression probes failed as expected on current code:

```text
TestReviewDelayedLaunchNavigationAfterLeavingDashboard
  stale launch navigation changed current view to 9; expected settings
TestReviewMissingSelectionRemainsBlockedUntilExplicitChoice
  periodic refresh silently replaced removed saved choice with new conversation; calls=1
```

Reproduction command used:
`go test -overlay /tmp/kankuro-step4-review-vst0fqv0/overlay.json ./internal/app ./internal/ui/views -run '^TestReview' -count=1 -v`.
Probe files are outside the repository; production and checked-in test files were
not changed. The scenarios and required permanent regressions are specified above
so the findings do not depend on retaining those temporary files.

No manual terminal or native model-driven E2E was performed independently. The
implementation's recorded isolated manual run is supporting evidence, not claimed
as this review's own verification. Re-review the fixes and their regressions
before recording PASS or advancing.
