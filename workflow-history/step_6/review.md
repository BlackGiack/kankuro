# Step 6 review — CHANGES REQUIRED

Reviewed the uncommitted step-6 worktree against `feature_6.md`,
`feature-dbcenter-rework.md` and `impl_plan.md`. This artifact is created because
the independent probes below reproduce required changes. No implementation was
modified or committed during this review.

## Required changes

### 1. [P2] Enforce the recovery/workspace cadence on change signals

Location: `internal/app/center.go:253–306`.

Every accepted `centerChangeMsg` immediately cancels workspace work and starts a
report read, including while a failed report read's 30-second recovery timer is
pending. The message store signals changes for other sessions, delivery updates
and lifecycle events as well as authored reports. If the read fails again, pending
events also trigger an immediate follow-up at lines 301–303. Thus ongoing unrelated
messaging bypasses the specified minimum recovery cadence.

The same path starts filesystem inspection after every successful empty report
read, even when the panel already has a successful workspace snapshot and its
30-second timer has not fired. Repeated unrelated store signals can repeatedly
cancel/restart Git inspection and prevent it from completing.

Reproduced with two deterministic coordinator probes: after a failed read with
cached reports and a scheduled recovery timer, inject one change without firing
the fake clock; a new read starts immediately. Separately, give the selected
session a successful workspace snapshot and scheduled timer, inject a store
change, then an empty successful report result; another workspace read starts
before the timer fires.

Required: distinguish event-driven report discovery from scheduled workspace work
and degraded retries. Keep the healthy first-report transition prompt, but gate
failed-read retries by the 30-second cadence, coalesce pending notices while
degraded, and avoid restarting a successful/in-flight workspace inspection merely
because an unrelated event still yields no reports. Add fake-clock regressions
for both paths, including repeated events while reads fail or Git runs slowly.

### 2. [P2] Bound subprocess pipe draining after Git cancellation

Location: `internal/api/center_workspace.go:141–168` (`runGit`).

`exec.CommandContext` kills the direct process, but `cmd.Wait()` also waits for
the stdout/stderr copying goroutines. A helper inheriting either pipe can keep
them open after Git is killed. No `WaitDelay` or equivalent bounded pipe-drain /
descendant cleanup exists, so the 10-second command deadline and selection
cancellation do not bound completion or release of the read's resources.

Reproduced with an isolated fake Git executable containing
`sleep 2 &` followed by `wait`. A 50 ms context deadline returned only after
approximately **2.005 seconds**, when the inherited pipes finally closed. A
long-lived helper can extend that wait indefinitely. Existing cancellation tests
only exercise a direct busy process without a pipe-holding descendant.

Required: bound `Wait` and inherited-pipe cleanup on deadline, selection cancellation
and output-budget termination, using an appropriate `exec.Cmd.WaitDelay` and/or
platform-specific process cleanup strategy. Preserve bounded output and reaping.
Add a subprocess regression with a descendant holding stdout/stderr open; assert
prompt cancellation without waiting for the descendant's normal exit.

### 3. [P2] Sanitize filenames independently of multiline report text

Location: `internal/ui/components/session_center.go:139–175`; shared text handling
at lines 50–73.

Git paths, rename origins, directory names and symlink targets are interpolated
into rows and passed through `centerSafeText`, which deliberately preserves
newlines for report bodies. Consequently a filename containing a newline creates
additional apparent file rows. The backend correctly preserves such filenames;
the missing boundary is their frontend representation.

Reproduced with one Git entry whose path is `actual\n?? forged`. The center renders:

```text
Git: /workspace
?? actual
?? forged
```

Required: escape or flatten control whitespace in each path/name/target before
row formatting, preserving multiline wrapping for report bodies. Apply this to
current/original Git paths and directory/symlink fields. Add renderer assertions
for newline/tab/control-bearing paths so one entry cannot impersonate another
status or warning row, while still fitting the panel's display-column bounds.

## Verification performed

Passed independently on the reviewed worktree:

```text
go test ./internal/api ./internal/session ./internal/app ./internal/ui/components ./internal/ui/views ./internal/architecture -count=1
go test -race ./internal/api ./internal/session ./internal/app -run 'Report|Message|Center|Attach|CommandClosures' -count=1
go test ./... -count=1
go build ./...
go vet ./...
git diff --check
```

Four additional review probes fail as described above (two cadence probes, one
pipe-drain probe, one filename-rendering probe). They were injected using a Go
overlay; no probe source was added to the repository:

```text
go test -overlay /tmp/kankuro-step6-review-overlay.json ./internal/app ./internal/api -run '^TestReview' -count=1 -v
go test -overlay /tmp/kankuro-step6-review-overlay.json ./internal/ui/components -run '^TestReview' -count=1 -v
```

Temporary source and output artifacts are `/tmp/kankuro-step6-review-*`, including
`probes.log` and `render.log`. These are actual failing assertions, not failures
of the existing supplied suite. No native-model E2E or new manual terminal journey
was run for this review. Lint was not rerun; build/vet/test success does not claim
resolution of previously documented lint debt.

## Other inspected acceptance boundaries

Authored report projection filters by dashboard and reply author, sorts by time /
sequence / ID, spans activations, bounds results and excludes notification payloads.
The subscription captures the store channel before initial reads and rearms before
notifying; service reads revalidate persisted membership. Workspace inspection uses
the saved actual directory, structured porcelain-v2 parsing and bounded inventories.
Contracts and presentation copies are independent of storage and terminal types.

Preview-exclusive production symbols and feature-6 dependency exceptions are
removed. Attachment's separate initial/live output and serialized input paths
remain; the full suite and focused race checks include their existing regressions.
Remaining concrete attachment/composition dependencies are explicitly feature 7.
Appedia regeneration and its extractor-interface caveat are recorded in
`appedia/VERIFICATION.md`; that documentation check does not approve this feature.

No environment blocker prevented review. The three findings above block step-6
approval; fix them, add regressions and return to the review gate before advancing.
