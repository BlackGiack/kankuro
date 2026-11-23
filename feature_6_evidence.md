# Feature 6 implementation evidence

Status: implementation and author verification performed; independent final review is pending.
The Appedia documentation/model refresh was subsequently completed by the planning agent;
see `appedia/VERIFICATION.md` for its checks and the documented extractor caveat.

## Implementation

- `internal/session/authored_reports.go` projects bounded retained reports by
  dashboard and author across activations; it applies a four-MiB response budget.
- `internal/api/center*.go` and `internal/api/contracts/center.go` provide scoped
  report reads, coalesced cancellable subscriptions, and saved-workspace inspection.
  Workspace inspection uses bounded Git porcelain-v2 output and an explicitly
  labelled directory fallback; directory entries and Git results are bounded.
- `internal/app/center*.go`, `internal/ui/models/session_center.go` and
  `internal/ui/components/session_center.go` replace the selected-session preview
  with report/workspace state, cancellation, recovery and presentation.
- Preview-only production wiring and component tests were removed; attachment's
  independent initial/live output and input path remains. Architecture allowlists
  and current project/spec/boundary documentation were updated.

## Verification performed

- `go test ./internal/api ./internal/session ./internal/app ./internal/ui/components ./internal/ui/views ./internal/architecture -count=1`: passed.
- `go test -race ./internal/api ./internal/session ./internal/app -run 'Report|Message|Center|Attach|CommandClosures' -count=1`: passed, including named center cases.
- `go test ./...`: passed during implementation before the final additional tests and renderer refactor; affected packages were rerun afterward.
- `go build ./...`, `go vet ./...`, `git diff --check`: passed during final verification.
- `golangci-lint run --new-from-rev=HEAD --whole-files=false`: passed with zero issues after the implementation, before the final renderer refactor; affected-code lint was subsequently checked again as part of the full run. Full `golangci-lint run ./...` reports 70 existing findings in `internal/architecture/boundary_test.go` (69) and `internal/ui/views/source_manager.go` (1); the final full-run output had no findings in the new center files.
- Inspected isolated rendered fixtures at 80×24 and 120×40 for reports, clean Git, a directory fallback and warnings. A second fixture checked resize, selection switching and clearing prior-session content. These were direct rendered fixtures, not a live native-model E2E run.

Tests include author and dashboard filtering, activation attribution and history
reload; initial subscription and rearm races; coalescing, cancellation, late
workspace results, scope removal and fake-clock retries; temporary workspace
identity; clean, staged/unstaged, rename and untracked Git, linked worktree,
non-Git and missing-Git cases; malformed, oversized and timed-out Git output;
directory truncation, hidden entries and symlinks; and safe panel clipping.
The complete acceptance matrix and final failure-policy assessment remain for the
independent reviewer; no review PASS is asserted here. The combined live terminal
resize/relaunch/recovery journey and native-model E2E were not performed.

## Changes after independent review

The three findings in `review.md` were addressed:

- Global change notices no longer bypass the 30-second retry after a failed report
  read or cancel/restart an in-flight workspace inspection. A successful first
  report still cancels obsolete workspace work immediately. Fake-clock tests
  cover repeated notices while degraded and while workspace inspection is slow.
- Git commands use a 250 ms `exec.Cmd.WaitDelay` to bound inherited-pipe draining
  after cancellation or output-limit termination. A fake Git process with a
  descendant retaining both output pipes exercises both paths.
- Git paths, rename origins, directory names and symlink targets are sanitized as
  individual single-line fields. Renderer tests cover control characters and
  verify that report bodies retain their multiline behavior.

Post-fix verification passed: focused feature packages, the filtered race suite,
`go test ./... -count=1`, `go build ./...`, `go vet ./...`, `git diff --check`,
and changed-code lint. All four independent overlay probes in `review.md` pass;
the inherited-pipe cancellation probe returned in approximately 301 ms. The
existing full-lint debt and the previously stated limits on native-model and live
terminal journeys remain. The Appedia generated source snapshot predates these
review fixes and may need a refresh by its documentation maintainer.

## Changes after step 6 re-review

The two application findings in `review_2.md` were fixed in the coordinator.
Workspace attempts are tracked independently of cached content, so a failed
initial inspection without a cache waits for the 30-second retry even when
unrelated report changes arrive. Workspace freshness is tracked separately from
report freshness; successful empty and failed report reads preserve a stale
workspace label until new workspace data or report content replaces it.

Fake-clock regressions cover repeated events before retry, recovery after the
timer, stale workspace through successful-empty and failed history reads, and
both workspace recovery and report takeover. All six original and re-review
independent overlay probes passed. Focused packages, the filtered race suite,
`go test ./... -count=1`, `go build ./...`, `go vet ./...`, `git diff --check`,
and changed-code lint passed; the final one-line lint cleanup is covered by a
targeted race rerun and overlay rerun.

The separate Appedia refresh requested by the re-review remains outstanding in
this activation: `.agents/skills/create-appedia/scripts/appedia.mjs` is absent.
The orchestrator was asked to arrange the Planner's refresh or supply the tool.
No independent review PASS is asserted here.

## Changes after independent review

The three findings in `ireview.md` were addressed. Dashboard navigation now
suspends operation resources while retaining a detached cache for the same
stable session and dashboard; successful empty reads, selection changes and
scope rejection clear or exclude the relevant content. Unix Git inspection now
runs in an owned process group and terminates the group on cancellation, timeout
or output-budget termination, retaining bounded pipe draining. The non-Unix
fallback terminates the direct process; the existing application does not build
on Windows because `internal/driver/driver.go` uses `syscall.Kill`.

Report headers retain the time and kind when they fit, including with only two
rows remaining, and use compact metadata at narrower widths. Added regressions
cover same-session navigation and failed rereads, empty clearing, other-session
selection and scope loss; real-Git fsmonitor and output-limit helper termination
with an unrelated process; and two-row report headers. Darwin and FreeBSD API
cross-builds passed. The focused packages, full `go test ./... -count=1`, filtered
race suite, `go build ./...`, `go vet ./...` and `git diff --check` passed after
the fixes; final targeted race and changed-code lint checks were also run. The
Planner owns the corresponding Appedia refresh. No independent review PASS is
asserted here.
