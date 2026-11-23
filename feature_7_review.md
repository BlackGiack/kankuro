# Feature 7 final review

Decision: **PASS — ready to advance to feature 8.**

Independent final re-review: 2026-11-17, against `feature_7.md` and the fix
requested in `ireview.md`, without reading `impl_plan.md`. The prior P2 finding
is resolved; no further required changes or blockers remain. `ireview_2.md` was
not created. The earlier `ireview.md` is historical. This handoff stays outside
the implementation commit.

## Acceptance and fix verified

- Frontend coordination, navigation, commands and tests moved from `internal/app`
  into `internal/ui`; no forwarding package remains. Main injects focused
  capabilities and startup diagnostics. Runtime failures use a neutral frontend
  entry point. Backend packages do not import UI or presentation frameworks.
- Attachment contracts expose owned initial output, cancellable next-output/EOF,
  serialized writes, resize and idempotent close. No PTY, driver, manager or
  concrete backend handle appears in frontend signatures. The backend captures
  the subscribed generation, transfers owned output chunks and copies initial
  buffers. Writes preserve the existing driver input gate; resize remains backend.
- Generation-owned output tracking retains trailing output through replacement
  and removal before stream completion. Tests exercise subscription admission
  around EOF, old-generation closure, stopped input/resize rejection and detached
  subscription ownership. Closing an attachment does not stop the session.
- `ui/terminal` owns human raw mode, size observation, Ctrl+K handling and keyboard
  protocol restoration. Real model-free PTY tests exercise setup, resize, legacy
  and Kitty detach, EOF/output error and restoration, including under race.
- Pending attachment ownership now includes dashboard/session identity, session
  epoch and accepted inventory baseline. Full and sessions-only accepted inventory
  reads invalidate confirmed missing targets; explicit successful session removal
  cancels matching ownership. Queued resources close immediately; in-flight reads
  are cancelled; late resources close and cannot start terminal execution.
- Failed/unattempted reads, stale request/epoch/generation and other-dashboard
  results do not invalidate valid requests. New-launch attachments survive
  pre-publication inventory. Checked-in regressions cover these cases and mixed
  successful inventory/failed registry reads. The original independent removal
  probe now passes for both full and sessions-only refresh paths.
- Independent deletion preparation/progress contracts retain backend revision and
  foreign-reference confirmation, durable intent, checkpoints and retry authority.
  UI presents authoritative no-intent/pending/completed/uncertain outcomes instead
  of inferring completion from disappearing rows. Commands capture request/visit
  scope; shutdown cancels and joins deletion workers. Existing confirmation,
  foreign-scope, recovery and progress-retention regressions pass.
- Production frontend import/type exceptions are removed. The sole direct-import
  exception is checked test-only rendering support; transitive type exposure has
  an empty allowlist. Main retains config initialization and runner/session/MCP
  lifetime ownership. Relevant project and boundary references use the new paths.

## Independent verification

All passed on the approved checkout:

```text
go test ./... -count=1
go test -race ./internal/ui/... ./internal/api ./internal/session -count=1
go build ./...
go vet ./...
git diff --check
golangci-lint run --new-from-rev=HEAD --whole-files=false
```

The full suite includes required frontend/API/session/MCP packages and architecture
enforcement. Changed-code lint reported zero issues. The independent removed-target
attachment probe, adapted only to the new owner signature and extended to both
refresh paths, passed using temporary Go overlays outside the workspace; probes
were cleaned. No manual/native-model terminal E2E was run. Full lint was not rerun;
earlier architecture-checker lint debt is not claimed resolved.

## Commit scope and preservation

The user authorized committing the approved step-7 implementation after PASS.
Include source, tests, moved files and project/boundary documentation. All roadmap
and feature requests/plans, reviews, evidence, workflow-history, AGENTS.md, Kankuro
configuration and credentials remain local and excluded. Preserve pre-existing
Appedia content and evidence, including its modified baseline/verification/raw
execution files, without including them in this commit. No push is authorized.

## Handoff

Proceed to feature 8 final repository-wide architecture and behavior verification.
Preserve source/report identity semantics, structured partial outcomes, lifecycle
revalidation, scoped cache retention, process-group cleanup, attachment resource
ownership and deletion recovery authority. No feature 7 blocker or question remains.
The test-only helper exception and previously recorded checker lint debt should
remain explicit in final verification; this PASS does not claim unrun E2E results.

## Approved paths

Moves below are represented as removal and addition for exact staging verification.

- `docs/frontend-boundary.md`
- `internal/api/attachment.go`
- `internal/api/attachment_test.go`
- `internal/api/card_confirmation.go`
- `internal/api/contracts/attachment.go`
- `internal/api/contracts/deletion.go`
- `internal/api/dashboard_deletion.go`
- `internal/api/deletion_frontend.go`
- `internal/api/deletion_frontend_test.go`
- `internal/api/membership.go`
- `internal/api/service_test.go`
- `internal/app/app.go`
- `internal/app/attach.go`
- `internal/app/callbacks.go`
- `internal/app/card_editor_async_internal_test.go`
- `internal/app/center.go`
- `internal/app/center_internal_test.go`
- `internal/app/center_ports.go`
- `internal/app/command_race_internal_test.go`
- `internal/app/configuration_deletion_internal_test.go`
- `internal/app/configuration_pending_internal_test.go`
- `internal/app/configuration_ports.go`
- `internal/app/conversation_choice_internal_test.go`
- `internal/app/coordinator.go`
- `internal/app/coordinator_test.go`
- `internal/app/current_fixture_external_test.go`
- `internal/app/dashboard_list.go`
- `internal/app/dashboard_list_keys_test.go`
- `internal/app/delete_card_guard_test.go`
- `internal/app/error_banner_test.go`
- `internal/app/export_test.go`
- `internal/app/handlers.go`
- `internal/app/helpers_internal_test.go`
- `internal/app/inline_rename_refresh_internal_test.go`
- `internal/app/integration_test.go`
- `internal/app/inventory_ordering_internal_test.go`
- `internal/app/launch_focus_internal_test.go`
- `internal/app/legacy_actions.go`
- `internal/app/main_test.go`
- `internal/app/mcp_manager_gaps_test.go`
- `internal/app/mcp_manager_review_test.go`
- `internal/app/multi_dashboard_internal_test.go`
- `internal/app/navigation.go`
- `internal/app/navigation_test.go`
- `internal/app/read_ports.go`
- `internal/app/refresh.go`
- `internal/app/refresh_cadence_internal_test.go`
- `internal/app/refresh_certification_internal_test.go`
- `internal/app/refresh_error_test.go`
- `internal/app/refresh_retention_test.go`
- `internal/app/scoped_refresh.go`
- `internal/app/scoped_refresh_error_internal_test.go`
- `internal/app/session_actions.go`
- `internal/app/session_actions_registry_test.go`
- `internal/app/session_ports.go`
- `internal/app/session_ports_internal_test.go`
- `internal/app/session_warning_test.go`
- `internal/app/settings_async_internal_test.go`
- `internal/app/settings_error_test.go`
- `internal/app/size_guard.go`
- `internal/app/size_guard_test.go`
- `internal/app/source_save_test.go`
- `internal/app/source_sync_warning_test.go`
- `internal/app/terminal_launch_test.go`
- `internal/architecture/allowlist.json`
- `internal/architecture/boundary_test.go`
- `internal/architecture/testdata/fixtures/child_to_root.go`
- `internal/architecture/type_exposure_allowlist.json`
- `internal/session/attach.go`
- `internal/session/attachment_generation_internal_test.go`
- `internal/session/io.go`
- `internal/session/lifecycle.go`
- `internal/session/manager.go`
- `internal/session/startup_boundary_test.go`
- `internal/testenv/testenv.go`
- `internal/ui/app.go`
- `internal/ui/attach.go`
- `internal/ui/attachment_missing_internal_test.go`
- `internal/ui/attachment_owner.go`
- `internal/ui/attachment_owner_test.go`
- `internal/ui/callbacks.go`
- `internal/ui/capabilities_external_test.go`
- `internal/ui/capabilities_internal_test.go`
- `internal/ui/card_editor_async_internal_test.go`
- `internal/ui/center.go`
- `internal/ui/center_internal_test.go`
- `internal/ui/center_ports.go`
- `internal/ui/command_race_internal_test.go`
- `internal/ui/configuration_deletion_internal_test.go`
- `internal/ui/configuration_pending_internal_test.go`
- `internal/ui/configuration_ports.go`
- `internal/ui/conversation_choice_internal_test.go`
- `internal/ui/coordinator.go`
- `internal/ui/coordinator_test.go`
- `internal/ui/current_fixture_external_test.go`
- `internal/ui/dashboard_list.go`
- `internal/ui/dashboard_list_keys_test.go`
- `internal/ui/delete_card_guard_test.go`
- `internal/ui/deletion_ports.go`
- `internal/ui/deletion_presentation.go`
- `internal/ui/deletion_presentation_internal_test.go`
- `internal/ui/deletion_workers.go`
- `internal/ui/deletion_workers_test.go`
- `internal/ui/error_banner_test.go`
- `internal/ui/export_test.go`
- `internal/ui/handlers.go`
- `internal/ui/helpers_internal_test.go`
- `internal/ui/inline_rename_refresh_internal_test.go`
- `internal/ui/integration_test.go`
- `internal/ui/inventory_ordering_internal_test.go`
- `internal/ui/launch_focus_internal_test.go`
- `internal/ui/main_test.go`
- `internal/ui/mcp_manager_gaps_test.go`
- `internal/ui/mcp_manager_review_test.go`
- `internal/ui/messages/messages.go`
- `internal/ui/multi_dashboard_internal_test.go`
- `internal/ui/navigation.go`
- `internal/ui/navigation_test.go`
- `internal/ui/read_ports.go`
- `internal/ui/refresh.go`
- `internal/ui/refresh_cadence_internal_test.go`
- `internal/ui/refresh_certification_internal_test.go`
- `internal/ui/refresh_error_test.go`
- `internal/ui/refresh_retention_test.go`
- `internal/ui/scoped_refresh.go`
- `internal/ui/scoped_refresh_error_internal_test.go`
- `internal/ui/session_actions.go`
- `internal/ui/session_actions_registry_test.go`
- `internal/ui/session_ports.go`
- `internal/ui/session_ports_internal_test.go`
- `internal/ui/session_warning_test.go`
- `internal/ui/settings_async_internal_test.go`
- `internal/ui/settings_error_test.go`
- `internal/ui/size_guard.go`
- `internal/ui/size_guard_test.go`
- `internal/ui/source_save_test.go`
- `internal/ui/source_sync_warning_test.go`
- `internal/ui/terminal/attach.go`
- `internal/ui/terminal/attach_test.go`
- `internal/ui/terminal/pty_test.go`
- `internal/ui/terminal_launch_test.go`
- `internal/ui/views/dashboard.go`
- `internal/ui/views/dashboard_list.go`
- `internal/ui/views/dashboard_list_test.go`
- `main.go`
- `project.md`
