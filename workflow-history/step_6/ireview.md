# Independent review — feature 6

Decision: **CHANGES REQUIRED. Not ready to advance to feature 7.**

Reviewed actual changes against `feature_6.md` and
`feature-dbcenter-rework.md`, without reading `impl_plan.md`. This review was
created because the findings below require changes. No implementation files
were edited; no commit or push was made.

## Findings

1. **[P2] Preserve successful same-session content when returning to the dashboard.**
   `internal/app/center.go:56–72` erases successful report/workspace data together
   with operation resources. Navigation away calls this method. On return,
   `reconcileCenter` at lines 95–102 publishes a fresh loading-only snapshot,
   overwriting the dashboard component's retained data even when the selected
   dashboard and stable session are unchanged. A subsequent history read failure
   therefore cannot keep the last successful report view marked stale: it falls
   back to workspace/loading despite already having displayed usable reports for
   that same session. This violates the specified failure table and same-session
   successful-content retention.

   An independent coordinator probe loaded a successful report for session `s`,
   navigated to Settings and back without changing the dashboard/selection, then
   injected a history-read failure. The prior report disappeared and the panel
   showed warnings and Loading instead of the cached report marked stale.
   Separate cancellation/resource teardown from the successful content cache.
   Preserve or restore a detached cache keyed by dashboard and stable session
   when returning, while continuing to reject stale results and clearing content
   when scope is revoked or another session is selected. Add regressions through
   navigation and a failed reread, including successful empty reads clearing the
   cache and a different selection never receiving it.

2. **[P2] Terminate owned Git helper processes on cancellation and output limits.**
   `internal/api/center_workspace.go:143–170` uses CommandContext cancellation and
   `cmd.Process.Kill`, both targeting only the direct Git process. WaitDelay bounds
   pipe draining but does not terminate Git's descendants. Git status can run
   a repository-configured fsmonitor helper (and subprocesses). Selection changes,
   timeout or application exit can therefore return from the read while helpers
   continue running; degraded periodic inspection can accumulate them.

   An independent probe used real Git in an isolated repository configured with
   a local fsmonitor script that started `/bin/sleep 60`. After observing the
   helper start, it cancelled the read. runGit returned promptly, but its
   sleep child remained alive; the probe explicitly killed both the child and shell
   during cleanup. Existing inherited-pipe tests establish bounded return time,
   not descendant cleanup. Give inspection processes an owned process group/tree
   and terminate that owned group on cancellation, timeout and output-limit
   termination, with platform-appropriate behavior and no signals to unrelated
   processes. Add a regression that checks helper termination as well as prompt
   command completion.

3. **[P3] Keep the report kind when the header fits in one row.**
   `internal/ui/components/session_center.go:113–116` unconditionally replaces
   the timestamp-and-kind header with the timestamp whenever two rows remain.
   This drops a required field even when both fit horizontally, including the
   final report displayed in an otherwise normal-height panel. A width-80,
   height-2 independent render probe with kind `blocked` and text `Need input`
   produced `2026-11-16 12:34\nNeed input`; the full timestamp/kind header easily
   fits in the first row. Keep kind and time together when they fit; use actual
   display width to compact metadata when needed. Test the last two available
   rows and a narrow header, not just output bounds.

## Verification

Passed independently on this checkout:

```text
go test ./internal/api ./internal/session ./internal/app ./internal/ui/components ./internal/ui/views ./internal/architecture -count=1
go test -race ./internal/api ./internal/session ./internal/app -run 'Report|Message|Center|Attach|CommandClosures' -count=1
go test ./... -count=1
go build ./...
go vet ./...
git diff --check
golangci-lint run --new-from-rev=HEAD --whole-files=false
```

Changed-code lint reported zero issues. The three additional independent probes
failed with the results described above. They used temporary Go overlays outside
the workspace and were removed; helper processes started by the cancellation
probe were explicitly killed. No native/model-driven or manual terminal E2E was
run. Full lint was not rerun, and earlier architecture-checker lint debt is not
claimed resolved. Component rendering was inspected through its actual output.

## Reviewed scope and acceptance evidence

Inspected the new application contracts, backend scope checks and reports/
workspace reads, retained-report projection, subscription lifecycle, coordinator
state machine, presentation models/component and affected dashboard/list callers.
Also inspected removal of app/preview.go, SessionPreview, preview messages,
scrollback refresh payloads and allowlist exceptions. Shared attach replay,
output subscriptions and input serialization remain present and their exercised
regressions pass. Updated project/product/boundary documentation was inspected;
implementation evidence and other review summaries were not used as approval.

- Authored reports are selected by SenderSessionID within persisted dashboard
  membership, across activations, ordered by time/sequence/stable ID. Projection
  and reload tests cover attribution, limits, omission and detached ownership.
- Subscribe-before-read and rearm-before-notification close the first-report
  window. Context cancellation, stale setup release, scope rejection, timer
  coalescing, report takeover and failed-read cadence have passing tests.
- Workspace inspection uses saved actual directories, argv-based bounded Git
  porcelain parsing and filesystem APIs. Tests cover whole-repository changes,
  renames/untracked paths, linked worktrees, temporary workspaces, clean versus
  non-Git versus warned fallback, detailed symlinks and bounded inventories.
- Existing same-visit failure tests preserve successful content and distinguish
  confirmed empty results. They miss the navigation round trip in finding 1.
- Rendering sanitizes report/path content and clamps by display columns; fitting
  tests miss the lost kind in finding 3. Process/pipe bounds miss finding 2.

These positives do not resolve the findings above. Remaining attachment/deletion/
composition retirement belongs to feature 7; no new preview bridge was found.

## Readiness and blockers

- `ireview.md` **was created**; it is the only review-authored workspace file.
- **Not ready to advance.** Resolve the findings and verify behavioral regressions
  before final approval.
- No user question or environment blocker. The blockers are the code issues above.
- No commit or push.
