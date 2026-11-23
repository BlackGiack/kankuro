# Step 3 follow-up review — CHANGES REQUIRED

Reviewed the current working tree against `feature_3.md`, `impl_plan.md` and
`review.md`, including the remediation and its new tests. The original three
findings are resolved, but one editor navigation path still lacks epoch checking.
No implementation changes or commit were made during this follow-up.

## Required change

### [P2] Correlate editor cancellation navigation with its originating editor

Location: `internal/app/callbacks.go:227`, especially the return of
`c.navigationCommand(messages.ViewDashboard)` at line 236; destination handler:
`internal/app/coordinator.go:302`.

OnCancel increments `editorEpoch` but publishes an ordinary ScopedNavigateMsg.
That envelope includes only dashboard identity and view generation. Neither
changes when another card editor is opened in the same dashboard, so an older
cancel navigation remains acceptable after the old editor has been replaced.
The new EditorNavigateMsg epoch protection is used for source navigation but
does not cover cancellation.

Reproduction using the coordinator's real callbacks and Update routing:

1. Open card A's editor and press Escape twice before the returned navigation
   commands are delivered. Each callback creates a separate cancel command.
2. Deliver the first cancel command; it returns to the dashboard.
3. Open card B's editor in that same dashboard and deliver its load result.
4. Deliver the second, delayed cancel command from A.

The coordinator navigates away from B's editor even though the command originated
from A. A delayed cancel can also interfere with a later reload. This violates
the feature/plan requirement to reject obsolete editor asynchronous results and
preserve editor epochs across cancellation/reopening. The save-pending guard
does not solve this because B need not be saving.

Give cancellation an editor-correlated envelope (or equivalent checked intent)
and accept it only for the originating editor epoch/state. Retain ordinary
Escape behavior, including closing a completed editor, and preserve pending-save
protection. Merely incrementing the epoch without transporting and checking it
does not protect this navigation. Add a durable regression for the sequence
above, plus normal cancel and stale cancel after reload; do not route completed
editor cancellation through a handler that rejects it solely because the save
completed.

Reviewer regression `TestReview2DelayedCancelCannotCloseNewEditor` failed:
`old cancel navigated newer editor away: view=0 card="second"`.
The reproduction overlays the current app regression test file, leaves all
production code unchanged, and is retained locally at:

```sh
go test -overlay /tmp/kankuro-step3-rereview-u1c2j1x6/overlay.json ./internal/app -run '^TestReview2' -count=1
```

## Original findings resolved

- **Save completion across source screens:** pending saves now freeze editing
  and normal source/cancel navigation. Matching offscreen completion records the
  saved identity/revision and requires an explicit reload before another save;
  newer editor epochs remain protected. Inspected the callbacks, result handler,
  editor pending/completed state and create/update regressions.
- **Picker refresh identity:** workspace replacements reconcile the retained
  selection; puppet replacements capture/reselect the active identity. Tests
  cover removal, reorder, empty results, recovery and explicit navigation, plus
  editor-level confirmation. Original workspace/puppet reproductions pass.
- **Persisted revision:** repository create/update now return the decoded
  representation of the exact YAML bytes prepared for the atomic write. The
  conversion happens before the write under the existing repository/admission
  ownership, with no post-admission reread or schema change. The application
  hashes that returned representation. Create/update fresh-load equality and
  subsequent saves using the returned revision pass, including empty optional
  lists. Inspected all changed repository paths and their callers.

## Independent verification

The following checks were executed in this follow-up and passed:

```sh
go test ./internal/api ./internal/config ./internal/ui/views ./internal/app ./internal/ui/components ./internal/architecture/... -count=1
go test -race ./internal/api ./internal/app -run 'Card|CommandClosures|TestEditor' -count=1
go build ./...
go vet ./...
go test ./...
golangci-lint run --new-from-rev=HEAD
git diff --check
go test -overlay /tmp/kankuro-step3-review-e3_dhje0/overlay.json ./internal/ui/views ./internal/app ./internal/api -run '^TestReview' -count=1
```

Changed-code lint reported zero issues. The original review overlay command now
passes all original reproductions; the separate new cancel regression fails as
documented above. Temporary overlays are supporting evidence; add the regression
to the repository when fixing it.

Full lint and manual terminal smoke were not repeated; full-lint baseline debt
and prior manual smoke remain documented in the author handoff and earlier
reviews. No native model-driven E2E was run. Remaining boundary exceptions retain
their later-feature ownership; the architecture suite passed.

No environment or authorization blocker prevented review. The cancellation
finding blocks approval pending correction and independent follow-up. Created
`review_2.md` only; no feature approval artifact or commit was created.
