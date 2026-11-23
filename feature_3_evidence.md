# Feature 3 implementation evidence

Implementation author handoff; this is not an independent review or approval.
The required assigned reviewer must inspect the actual working tree and record a
verdict in `feature_3_review.md`.

The card editor now uses editable-only contracts, explicit changes, detached
options and a versioned full-record revision. Owner and sharing policy and the
revision are revalidated in the existing card-admission boundary. The UI retains
unresolved selections, tracks hidden edits, handles reload/conflict and filters
asynchronous results by dashboard, request and editor epoch. The old full-card
editor save operations and editor dependency exceptions were removed; the
selected full-card launch adapter remains for feature 4. See the tracked diff
and untracked `internal/api/card_*`, `internal/api/contracts/card_editor.go`,
`internal/ui/models/card_editor.go`, and editor regression tests for the actual
implementation.

Checks run in this workspace:

- `go test ./internal/api ./internal/ui/views ./internal/app -count=1` passed
  during integration; a subsequent full suite passed after implementation cleanup.
- `go test ./internal/config ./internal/ui/models ./internal/ui/components -count=1` passed.
- `go test ./internal/architecture/... -count=1` passed; the full suite also passed it.
- `go test ./...` passed. The final subsequent test-only addition was run with
  the API race selection below.
- `go test -race ./internal/api ./internal/app -run 'Card|CommandClosures' -count=1` passed.
- `go test -race ./internal/api -run 'TestEditor|TestCardRevision' -count=1` passed,
  explicitly covering the new save, revision and concurrent-save tests omitted
  by the first regex.
- `go build ./...`, `go vet ./...`, and `git diff --check` passed.
- `golangci-lint run --new-from-rev=HEAD` passed with zero issues. Full
  `golangci-lint run` reports 69 findings, all in the pre-existing
  `internal/architecture/boundary_test.go` debt documented in feature 2; the
  complete current log is `/tmp/kankuro-final-lint.log`.
- In an isolated configuration, temporary workspace and HOME, a terminal run
  exercised create, owner edit, foreign read-only, a stale save and explicit
  reload, return from workspace sources, and clearing a skill while a hidden
  persona remained selected. A wrong foreign owner label found during that run
  was corrected and covered by `TestCardEditorShowsActualOwnerForReadOnlyCard`.
  The terminal run preceded the final encoder/formatting cleanup; automated
  checks above ran afterward.

No native model-driven E2E was run. No commit was made. The Kankuro activation
cannot discover the assigned reviewer: `list_sessions` returned “caller
activation is not authorized for dashboard control”. The orchestrator must
arrange the independent review; no reviewer verdict has been inferred.

## Planner review remediation

The three findings in `review.md` have been addressed; this section is author
verification and is not an independent review verdict.

1. While a save is pending, the editor prevents further edits, cancellation and
   source-screen navigation. Editor navigation carries the editor epoch, so a
   delayed navigation cannot open a stale source screen. If an otherwise matching
   completion arrives while the editor is offscreen, it records the returned card
   identity/revision and enters a completed state that preserves the draft and
   requires an explicit reload before another save. Regression tests cover both
   creation and update, a queued navigation, and a newer editor epoch.
2. The workspace and puppet pickers retain their active option identity through
   successful registry replacement, including reorder, removal, empty results,
   recovery and explicit navigation. Editor-level tests exercise confirmation
   after an incomplete refresh with each picker open.
3. Card persistence now decodes the exact YAML bytes prepared for its atomic
   write before writing them and returns the decoded stored representation.
   Consequently, create/update results derive their revisions from the actual
   serialized representation without a post-admission read. Regressions compare
   create/update revisions to fresh loads and perform another update using the
   returned revision, including an explicitly empty optional list.

Post-remediation verification:

- `go test ./...` — passed.
- `go test -race ./internal/api ./internal/app -run 'Card|CommandClosures|TestEditor' -count=1` — passed.
- `go test ./internal/architecture/... -count=1` — passed after updating exact
  line-position entries for existing, unchanged dependency exposures.
- `go test -overlay /tmp/kankuro-step3-review-e3_dhje0/overlay.json ./internal/ui/views ./internal/app ./internal/api -run '^TestReview' -count=1` — all three original reviewer reproductions passed.
- `go build ./...`, `go vet ./...`, `git diff --check`, and
  `golangci-lint run --new-from-rev=HEAD` — passed.
- Full `golangci-lint run` reports the same 69 pre-existing architecture-checker
  findings, with no diagnostics outside `internal/architecture/boundary_test.go`.

No additional manual terminal pass or model-driven E2E was run for these fixes.
No commit was made. Independent follow-up review remains the orchestrator's gate.

## Planner follow-up cancellation remediation

The `review_2.md` finding is addressed. Cancellation now publishes an intent
capturing dashboard, view generation, editor epoch and request identity. The
coordinator accepts it only for the active, non-saving editor and advances the
epoch upon acceptance. Duplicate delayed cancellations cannot close a newer
editor; a reload or intervening save invalidates an older cancellation. Completed
editors remain cancellable, and in-flight save protection remains in place.

Durable coordinator regressions exercise normal cancellation, duplicate Escape
followed by opening a different editor, stale cancellation after reload or a
failed save, and cancellation of a completed editor. After this change:

- `go test ./...` and `go test -race ./internal/api ./internal/app -run 'Card|CommandClosures|TestEditor' -count=1` passed.
- `go test -overlay /tmp/kankuro-step3-rereview-u1c2j1x6/overlay.json ./internal/app -run '^TestReview2' -count=1` passed the Planner's original reproduction.
- `go test ./internal/architecture/... -count=1`, `go build ./...`, `go vet ./...`,
  `git diff --check`, and `golangci-lint run --new-from-rev=HEAD` passed.
- Full `golangci-lint run` still reports the 69 previously documented findings
  in `internal/architecture/boundary_test.go`; no other diagnostics appeared.

No new manual terminal pass or model-driven E2E was run. Independent follow-up
review is pending, and no commit was made.

## Independent review remediation

The two findings in `ireview.md` are addressed. Before workspace options are
replaced, the picker records the currently highlighted selection; explicit
search changes update that identity, while unresolved selections survive empty
search results. Modal and card-editor regressions cover search, refresh and
confirmation, plus unresolved-reference recovery. Existing arrow-navigation,
removal and reorder tests remain in place.

The pending temporary-workspace cleanup test now updates `WorkspaceID` directly
with a presence-aware application change and checks the persisted workspace ID
before deletion. Its foreign-session, admission and retry assertions remain in
place. The specific permission-based test was run as non-root and passed.

Post-remediation: `go test ./...`, `go test -race ./internal/api ./internal/app
-run 'Card|CommandClosures|TestEditor' -count=1`, `go build ./...`, `go vet ./...`,
`go test ./internal/architecture/... -count=1`, `git diff --check`, and
`golangci-lint run --new-from-rev=HEAD --whole-files=false` passed. Full lint
reports the same 69 pre-existing findings confined to the architecture checker.
No additional manual terminal or model-driven E2E run was performed. No commit
was made; independent follow-up review remains pending.
