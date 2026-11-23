# Step 6 re-review — CHANGES REQUIRED

Reviewed the current fixes against `feature_6.md`, `impl_plan.md` and `review.md`.
Created this artifact because two workspace-failure cases still require changes.
No application code, tests or commits were changed during re-review.

## Original findings

All four original independent overlay probes now pass:

- Failed report reads coalesce events until the recovery timer fires.
- A successful cached workspace is not re-inspected on unrelated empty-report
  updates, and the new project tests retain in-flight workspace work until a
  successful report read actually finds reports.
- Git uses a 250 ms `WaitDelay`; the original 50 ms cancellation probe now
  completes in approximately **301 ms** despite inherited pipes.
- Filename newlines no longer create extra Git rows; new project tests also
  cover rename origins, directory entries, symlink targets and multiline reports.

These fixes resolve the original reproductions. The following additional cases
are required by the same failure policy and were not covered by those probes.

## Required changes

### 1. [P2] Rate-limit failed workspace reads even when no cached content exists

Location: `internal/app/center.go:273–280` and `315–316`.

After an initial workspace read fails, `data.Workspace` remains nil and a
30-second timer is scheduled. If report reads are healthy, an unrelated store
change still starts a report read immediately. When that read succeeds empty,
the `data.Workspace == nil` condition immediately starts another workspace read,
without waiting for the scheduled timer. Each subsequent unrelated message or
delivery change can repeat the failed filesystem/Git work at event frequency.

Reproduction: successful empty report read → initial workspace error without
cache → scheduled retry timer → unrelated change → successful empty report read.
`working` becomes true and `workspaceRequest` advances before the timer fires.
The new `TestCenterChangesDoNotRestartWorkspace` only covers a cached successful
workspace and an in-flight read, so it misses this state.

Required: distinguish an unattempted initial workspace load from an already failed
attempt without cache. After failure, only the scheduled refresh/recovery cycle
should make another workspace attempt due. Preserve immediate initial loading and
healthy first-report discovery. Add a fake-clock regression with repeated events
after workspace failure and no cached result; no new workspace attempt may start
until the 30-second timer fires.

### 2. [P2] Preserve workspace staleness until workspace data is refreshed

Location: `internal/app/center.go:290–295`, with workspace failure handling at
`333–339`.

A failed directory/workspace read correctly retains same-session workspace data
with `data.Stale = true`. A later successful report read unconditionally sets
`data.Stale = false`, even when it returns no reports and no new workspace data.
Because the cadence fix correctly avoids an unnecessary workspace read when a
cache exists, an unrelated store event removes `[stale]` from that failed cached
listing until the next scheduled workspace attempt. The workspace warning remains,
but the explicitly required stale label is lost. The report-error branch can
likewise overwrite the workspace stale flag based only on report count.

Reproduction: seed cached workspace → workspace refresh fails → verify stale →
inject unrelated change → successful empty report result. The cached workspace is
now marked fresh without any successful filesystem read.

Required: derive staleness from the content actually being displayed, or track
report/workspace freshness separately. A history read must not certify cached
filesystem data. Clear workspace staleness only after successful replacement
workspace data or a successful transition to report content. Add assertions for
both successful-empty and failed report reads while stale workspace data remains
displayed, followed by actual workspace recovery.

## Independent verification

Passed on the fixed worktree:

```text
go test ./internal/api ./internal/session ./internal/app ./internal/ui/components ./internal/ui/views ./internal/architecture -count=1
go test -race ./internal/api ./internal/session ./internal/app -run 'Report|Message|Center|Attach|CommandClosures' -count=1
go test ./... -count=1
go build ./...
go vet ./...
git diff --check
```

The independent probe run passes all four original assertions and fails the two
additional assertions above:

```text
go test -overlay /tmp/kankuro-step6-rereview-overlay.json ./internal/app ./internal/api ./internal/ui/components -run '^(TestReview|TestReReview)' -count=1 -v
```

Probe sources/logs are `/tmp/kankuro-step6-rereview-app_test.go` and
`/tmp/kankuro-step6-rereview-probes.log`, with the original probe files referenced
by that overlay. These are temporary review inputs; no test implementation was
added to the repository. No fresh native-model E2E, manual terminal journey or
lint run was performed. There is no environment blocker; the remaining failure
policy defects block step-6 approval.

## Appedia assessment

**Refresh required after the fixes settle.** All three changed implementation
files and their three test files differ from the accepted Appedia baseline.
The generated model does not contain `components.centerSafePath`, and its
`api.runGit` source still lacks `WaitDelay`. Center flow explanations and the
source-bound assessment hashes also predate the cadence fixes and new regressions.

Re-extract/rebuild through the existing prepare/rebuild/finalize pipeline, update
report/workspace/recovery explanations and source-bound evidence, and rerun smoke
and relevant visual checks after the remaining fixes. Keep historical execution
provenance and the documented interface-inference caveat. This re-review assesses
freshness; it does not accept a new intermediate Appedia baseline or claim a
refresh was performed.
