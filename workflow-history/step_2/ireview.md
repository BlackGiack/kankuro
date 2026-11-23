# Independent review — feature 2

Decision: **CHANGES REQUIRED. Not ready to advance to feature 3.**

Reviewed the step 2 working-tree changes against `feature_2.md`, the retained
preview behavior and affected callers/tests. Did not read `impl_plan.md` or other
implementation plans/review summaries. This file was created because a required
behavior regressed. No implementation files were changed or committed.

## Required change

**[P2] Preserve preview scrollback reconciliation until feature 6 removes it.**

`internal/app/scoped_refresh.go:46`–68 returns only the focused inventory from the
session poll. The full refresh at `:20`–42 likewise no longer populates
`PreviewSessionID`/`PreviewScrollback`. Previously both producers read the selected
session's retained scrollback after validating scoped membership, and the
coordinator forwarded that snapshot to the existing preview. Those reconciliation
reads have been removed without an equivalent replacement.

The preview live subscription is intentionally lossy:
`internal/session/io.go:broadcast` drops chunks when a subscriber channel is full.
It can lag during output bursts or terminal attachment. Now, when the selected
session remains the same, `maybeSubscribeAutoSelectedPreview` does not resubscribe
and no poll or detach/full refresh repairs the missing content. The preview can
stay incomplete indefinitely until the user changes selection or navigates away
and back. Replacing backend inventory records with contracts must not remove this
existing output-recovery behavior early; feature 2 explicitly retains preview
plumbing until its replacement in feature 6.

Restore a scoped, stale-result-safe reconciliation of the selected preview with
retained output during the existing refresh paths. It may remain in the temporary
feature-6-owned preview adapter; do not reintroduce broad session snapshots into
list contracts or create a permanent preview API. Preserve dashboard/session
identity, generations and inventory ordering when applying the snapshot. Add a
regression that models missed subscriber output and verifies recovery on both a
session poll and a full/detach refresh while selection stays unchanged.

## Independent reproduction

A temporary Go test was supplied with `go test -overlay`, outside the workspace:

1. Use the existing isolated coordinator/manager fixtures and launch a local `cat`
   session in the default dashboard.
2. Apply a full refresh so the preview selects and subscribes to that session.
3. Write a marker through the checked attachment's serialized input and wait until
   the real manager scrollback contains it.
4. Discard its live subscriber delivery to model a dropped chunk from the lossy
   broadcast, leaving the preview behind the retained scrollback.
5. Apply the normal session poll, then the full refresh returned after
   `AttachFinishedMsg`, without changing the selected session.

Both recovery assertions failed. The manager retained
`INDEPENDENT_PREVIEW_RECOVERY_MARKER\r\n`, while the preview scrollback remained
empty. The temporary overlay and test source were removed; no probe test was
added to the implementation tree. The first probe attempt checked an LF-only
marker against PTY CRLF output and could not establish its setup; the corrected
probe used a newline-independent marker and reproduced the two recovery failures.

## Verification and scope

Passed independently:

- `go test ./internal/api ./internal/app ./internal/ui/components ./internal/ui/views ./internal/architecture -count=1`
- `go test -race ./internal/app -run 'TestCommandClosuresDoNotRaceCoordinator|Dashboard|Refresh' -count=1`
- `go build ./...`
- `go vet ./...`
- `git diff --check`
- `golangci-lint run --new-from-rev=HEAD --whole-files=false` — zero issues

`golangci-lint run` exits 1 with 69 issues in the architecture checker, including
complexity, modernization, test-package and unused-helper diagnostics. The
diff-filtered check reports no new-line issues. This lint debt is recorded
separately from the behavioral blocker; it does not justify expanding this fix.

The focused contracts, presentation mapping, scoped inventory reads, explicit
success/error distinction, change-token certification and shared publication
ordering have meaningful tests and passed the checks above. The existing test
suite did not cover reconciliation after missed preview output. The architectural
allowlists and remaining editor/action/configuration/attachment adapters remain
temporary migration work, not a claim of final separation.

No native-model or manual terminal E2E run was performed. The reproduction used a
local process and isolated test fixtures, not a model provider.

## Readiness and blockers

- `ireview.md` **was created**.
- **Not ready to advance:** fix preview reconciliation and verify the new
  regression before final approval.
- No user question or environment blocker. No commit was made.
