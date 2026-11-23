# Independent review — feature 7

Decision: **CHANGES REQUIRED. Not ready to advance to feature 8.**

Reviewed actual changes against `feature_7.md` and relevant project boundary,
asynchronous-result, attachment and deletion requirements. Did not read
`impl_plan.md`. This file was created because one required change remains.
No implementation changes or commits were made. Workflow and Appedia evidence
were preserved.

## Finding

1. **[P2] Dispose of pending attachments whose session has disappeared.**
   `internal/ui/coordinator.go:726–757` checks dashboard/generation, lifecycle
   request and attachment-owner sequence before running an attachment. Neither
   that path nor accepted inventory refreshes invalidate the resource when its
   target session disappears. `attachmentOwner` retains only a sequence/resource,
   so an otherwise current queued result remains usable after a newer successful
   inventory has confirmed removal. The coordinator returns a terminal execution
   command for the removed session rather than closing the unused handle.

   An independent temporary-overlay probe constructed a dashboard containing
   session `gone`, registered a pending attachment and published its resource,
   then delivered a successful scoped SessionsRefreshedMsg with a newer accepted
   inventory request and an empty inventory. Delivering the queued
   ScopedAttachReadyMsg for `gone` afterward left the resource open and returned
   a non-nil terminal command (`closed=false command=true`). This violates the
   explicit feature-7 requirement to close obsolete asynchronous attachment
   results immediately, including missing sessions. It can also display obsolete
   initial output or enter terminal setup for a session already removed from the
   dashboard.

   Bind pending attachment ownership to its dashboard/session and invalidate it
   when an authoritative accepted inventory/removal result makes that target
   obsolete; reject and close any corresponding queued result. Preserve the
   distinction between a failed/unattempted inventory and confirmed absence,
   and do not reject a newly launched session merely because an older snapshot
   predates launch publication. Add regressions for removal while acquisition or
   its result is pending, alongside launch-before-refresh and failed-read cases.

## Verification

Passed independently:

```text
go test ./... -count=1
go test -race ./internal/ui/... ./internal/api ./internal/session -count=1
go build ./...
go vet ./...
git diff --check
golangci-lint run --new-from-rev=HEAD --whole-files=false
```

The full suite includes required focused frontend/API/session/MCP checks and the
architecture checker. Changed-code lint reported zero issues. The additional
removed-target attachment probe failed as described above; its Go overlay lived
outside the workspace and was cleaned. Full lint was not rerun, and previously
recorded architecture-checker lint debt is not claimed resolved. No manual or
native-model terminal E2E was performed. Checked-in model-free PTY tests actually
exercise terminal setup, resize, legacy/Kitty detach, EOF/output failure and raw
state restoration, and passed, including under the race detector.

## Reviewed scope and acceptance evidence

- `internal/app` production files/tests are removed and migrated under
  `internal/ui`; main injects focused capabilities and startup diagnostics and
  uses the neutral runtime-error entry point. No forwarding package remains.
  Production import exceptions are gone; the sole direct exception is verified
  test-only rendering support, and the transitive type-exposure allowlist is empty.
- Inspected attachment contracts, API adapter, session admission/generation checks,
  output tracking and subscription ownership, frontend attachment owner and raw
  terminal command. Public frontend signatures contain no PTY/driver/manager or
  concrete session handle. Initial bytes are copied; broadcast chunks are owned
  per subscriber. Input uses the existing driver gate and resize is backend-owned.
- Generation-owned subscriber maps and tracker-owned closure preserve queued tail
  output through replacement/removal and EOF. Checked tests cover output admission
  around EOF, cancelled subscriptions, stale generation closure, resize and stopped
  input rejection. Owner tests cover supersession/navigation/shutdown but miss
  the known-removed target in the finding above.
- Inspected independent card/dashboard deletion preparation/progress contracts,
  human adapters, backend confirmation changes, command workers, views and result
  presentation. Revision/foreign confirmation and durable recovery remain backend
  policy. Pre-intent failure, uncertain storage results, pending cleanup and
  completion remain distinct. UI presents execution progress independently of
  inventory errors, with request/visit guards and shutdown cancellation/join.
- Main still owns configuration initialization, manager/runner/MCP composition and
  shutdown ordering. Startup load diagnostics and runtime endpoint errors are
  forwarded through frontend values. Relevant project and boundary documentation
  points to the moved packages. Appedia/evidence files were not modified by review.

The positive checks do not resolve the missing-session disposal finding. No user
question or environment blocker remains; this code issue blocks advancement.

## Readiness

- `ireview.md` **was created** and is the only review-authored workspace file.
- **Not ready to advance to feature 8.** Fix and verify the finding first.
- No commit or push.
