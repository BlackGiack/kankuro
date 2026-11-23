# Feature 6 final review

Decision: **PASS — ready to advance to feature 7.**

Independent final re-review: 2026-11-17, against `feature_6.md`,
`feature-dbcenter-rework.md` and the fixes requested in `ireview.md`, without
reading the implementation plan. All three prior findings are resolved; no
further required changes or blockers remain. `ireview_2.md` was not created.
The earlier `ireview.md` is historical. This handoff remains outside the commit.

## Acceptance and fixes verified

- The center projects retained reports authored by the selected stable session
  within persisted dashboard membership, across activations, in descending
  timestamp/sequence/stable-ID order. Restart reconstructs from existing durable
  history. Notifications do not become duplicate reports. Results are detached,
  bounded and carry omission metadata.
- Subscribe-before-read and rearm-before-notification preserve the first-report
  window. Coalesced events read retained records. Scope/generation/request guards,
  cancellation, stale setup release and idempotent subscription release preserve
  selection isolation. Recovery uses a non-overlapping 30-second cadence.
- Saved actual workspace inspection stays behind the application boundary. Git
  uses bounded cancellable argv execution and structured whole-repository status
  parsing. Clean repositories, non-repositories and warned directory fallback are
  distinct. Filesystem listing includes dot entries, detailed metadata and
  symlink targets, with bounded enumeration and visible truncation.
- Same-session cached reports survive navigation and remain stale on failed
  rereads. Confirmed empty reads clear them; different selections never inherit
  them; scope loss clears cached content. Resources are cancelled independently
  of the detached navigation cache. Checked-in and independent probes pass.
- Unix Git inspections have owned process groups. Context cancellation and
  output-limit termination kill helpers as well as Git. Regressions cover a real
  Git fsmonitor helper, output-limit termination, prompt return and preservation
  of unrelated processes. The original independent helper probe now passes.
- Rendering sanitizes controls and paths, wraps by display columns and clips
  newest-first to available space. Two-row headers retain time and kind when
  they fit and compact time at narrower widths. Independent wide/narrow render
  probes pass. No report paging or scrollback was added.
- Preview-exclusive components, messages, subscriptions, refresh payloads and
  allowlist entries are removed. Shared attachment replay/live output and input
  serialization remain present and their exercised regressions pass.

## Independent verification

All passed on the approved checkout:

```text
go test ./... -count=1
go test -race ./internal/api ./internal/session ./internal/app -run 'Report|Message|Center|Attach|CommandClosures|GitProcess' -count=1
go build ./...
go vet ./...
git diff --check
golangci-lint run --new-from-rev=HEAD --whole-files=false
```

The full ordinary suite includes the required focused API/session/app/component/
view packages and architecture enforcement. Changed-code lint reported zero
issues. The original navigation/history-failure, real-Git fsmonitor cancellation
and two-row metadata probes were rerun independently using temporary Go overlays;
all passed and temporary probes were removed. No manual/native-model terminal
E2E or non-Unix runtime verification was performed. Full lint was not rerun;
earlier architecture-checker lint debt is not claimed resolved.

## Commit scope and Appedia

The user authorized committing the approved step-6 implementation and appropriate
refreshed Appedia assets. Include the browsable model/content/index and README.
The README no longer links to excluded local workflow artifacts; snapshot labels
no longer claim independent review is pending. JavaScript syntax and model/content/
index parsing were checked. Browser smoke/captures were not rerun by this reviewer.
The source model's documented method-name-only interface inference caveat remains.

Exclude and preserve all roadmap/feature plans, reviews, raw evidence, verification
reports, workflow-history, Appedia maintenance records, AGENTS.md and Kankuro
configuration/credentials. In particular, Appedia baseline files, VERIFICATION.md
and raw data/test-evidence.json remain local. The generated browsable assets retain
their documented test/coverage provenance. No push is authorized.

## Handoff

Proceed to feature 7 frontend retirement and attachment/deletion/composition
boundaries. Preserve report attribution/continuity, cache scope, backend authority,
process-group cleanup and structured read-failure outcomes. Final repository-wide
verification belongs to feature 8. No question or feature 6 blocker remains.

## Approved changed files

- `appedia/README.md`
- `appedia/data/content-index.json`
- `appedia/data/content.js`
- `appedia/data/model.js`
- `docs/frontend-boundary.md`
- `internal/api/center.go`
- `internal/api/center_process_other.go`
- `internal/api/center_process_unix.go`
- `internal/api/center_process_unix_test.go`
- `internal/api/center_test.go`
- `internal/api/center_workspace.go`
- `internal/api/center_workspace_linux.go`
- `internal/api/center_workspace_other.go`
- `internal/api/center_workspace_test.go`
- `internal/api/contracts/center.go`
- `internal/api/membership.go`
- `internal/app/app.go`
- `internal/app/center.go`
- `internal/app/center_internal_test.go`
- `internal/app/center_ports.go`
- `internal/app/command_race_internal_test.go`
- `internal/app/coordinator.go`
- `internal/app/handlers.go`
- `internal/app/inline_rename_refresh_internal_test.go`
- `internal/app/integration_test.go`
- `internal/app/inventory_ordering_internal_test.go`
- `internal/app/launch_focus_internal_test.go`
- `internal/app/multi_dashboard_internal_test.go`
- `internal/app/navigation.go`
- `internal/app/preview.go`
- `internal/app/preview_broadcast_test.go`
- `internal/app/preview_integration_test.go`
- `internal/app/preview_reconciliation_internal_test.go`
- `internal/app/refresh_retention_test.go`
- `internal/app/scoped_refresh.go`
- `internal/app/scoped_refresh_error_internal_test.go`
- `internal/app/size_guard.go`
- `internal/architecture/allowlist.json`
- `internal/architecture/type_exposure_allowlist.json`
- `internal/session/authored_reports.go`
- `internal/session/authored_reports_test.go`
- `internal/ui/components/component_rows_test.go`
- `internal/ui/components/session_center.go`
- `internal/ui/components/session_center_test.go`
- `internal/ui/components/session_list.go`
- `internal/ui/components/session_list_test.go`
- `internal/ui/components/session_preview.go`
- `internal/ui/components/session_preview_integration_test.go`
- `internal/ui/components/session_preview_layout_test.go`
- `internal/ui/components/session_preview_overflow_test.go`
- `internal/ui/components/session_preview_test.go`
- `internal/ui/components/text.go`
- `internal/ui/messages/messages.go`
- `internal/ui/models/session_center.go`
- `internal/ui/views/constants.go`
- `internal/ui/views/dashboard.go`
- `internal/ui/views/dashboard_layout_test.go`
- `internal/ui/views/dashboard_test.go`
- `kankuro_specs.md`
- `main.go`
- `project.md`
