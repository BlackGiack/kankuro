# Step 5 re-review

Decision: **CHANGES REQUIRED** — 2026-11-16.

Created `review_2.md` because one failure-feedback defect remains in the pending
save fix. No implementation fixes or commits were made. Reviewed against
`feature_5.md`, `impl_plan.md` and the findings in `review.md`.

## Previous findings

Both original defects are addressed:

- Settings save messages now capture visit and draft identity separately from
  global completion. Success updates applied settings globally, while obsolete
  results do not navigate or replace current feedback. The held-command tests
  cover navigation to Puppet Sources, reopened settings and newer drafts, with
  both successful and failed saves.
- MCP forms now expose saving status and block keyboard/paste edits and repeat
  submission during a pending save. Failure unlocks the form for retry. Settings
  similarly block a repeat edit in the submitting visit and explain a blocked
  submission from a newer visit.

## Required change

### [P2] Clear all pending-save feedback when a settings save finishes

Location: `internal/ui/views/settings.go:110` (`SaveFinished`), with the stale
message assigned at lines 144 and 154.

While saving, pressing Enter or `e` in the settings display sets
`s.feedback = "Saving settings…"`. When the operation fails, the coordinator calls
`SaveFinished`, which clears `s.saving` but clears feedback only when it equals
the other pending message, `"A settings save is still in progress"`. The form
therefore continues to render **Saving settings…** after the save has finished
and failed. This falsely indicates an ongoing operation; the error banner can
also clear on the next keypress while that stale pending text remains. Retrying
is possible, but the failure/retry state is misleading.

Clear the pending status for both branches on completion, preferably without
making state cleanup depend on exact display strings. Preserve the existing
draft and error, and keep retry available. Extend the pending-save regression to
press Enter (and cover `e`) during the held save, deliver failure, and assert
that the resulting view no longer claims a save is in progress before retry.

Independent reproduction: temporary
`TestReReviewSettingsFailureClearsPendingFeedback` used the existing coordinator
fixture and `pendingSaveOverride`, entered settings editing, submitted a held
save, pressed Enter a second time, then delivered the failed save result through
`Update`. The operation was released and the expected error was present, but the
render assertion failed with:

```text
completed failed save still displays Saving settings
```

The temporary test was removed after verification. The production code and
implementation tests were not changed by this review.

## Boundary and scope

The acceptance mapping and reviewed implementation inventory in `review.md`
remain applicable. Rechecked the new message fields, captures, pending-state
handlers, view guards, navigation and failure-retry paths, and the new
`internal/app/configuration_pending_internal_test.go` coverage. No additional
backend contract, source concurrency, credential exposure, storage or MCP wire
issue was found. Preview remains feature 6; composition/deletion/attachment
exceptions remain feature 7; test-support audit remains feature 8.

## Verification

Passed independently:

```text
go test ./internal/config ./internal/api ./internal/app ./internal/ui/views ./internal/ui/components ./internal/api/contracts ./internal/architecture -count=1
go test -race ./internal/config ./internal/api ./internal/app -run 'Source|MCP|Settings|Refresh|CommandClosures' -count=1
go test ./... -count=1
go build ./...
go vet ./...
golangci-lint run --new-from-rev=HEAD ./...
git diff --check
```

The final full suite passed, with no reported intermittent failure reproduced.
Changed-line lint reported zero issues; unrestricted lint was not rerun and the
previously recorded architecture-checker debt is not claimed resolved.
The temporary regression probe intentionally failed as described above. An
initial lint run overlapped its removal and failed on that disappearing test
file; it is not a production-code failure and requires the clean rerun below.

No independent manual terminal or native-model E2E run was performed. The earlier
intermittent-test investigation remains recorded in `review.md`; no root cause
or universal absence of those intermittent failures is claimed. No external or
environment blocker was found. The remaining failure-feedback fix is required
before approval and advancement. Do not commit as part of this review.

The follow-up changes inspected are in `internal/app/callbacks.go`,
`coordinator.go`, `navigation.go`, `internal/ui/messages/messages.go`,
`internal/ui/views/mcp_form.go`, `settings.go`, the new
`internal/app/configuration_pending_internal_test.go`, and the appended
`feature_5_evidence.md` review-fix record. The rest of the feature remains covered
by the prior changed-file inventory and the repeated integration checks above.
