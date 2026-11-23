# Feature 8 final review

Decision **PASS** for the step 8 candidate based on `dc3b4689eb01b527ab9c0bb3e9ecd8790cb181e4`.
The tracked delta excluding Appedia has SHA-256
`4198f162c4ab5d633d49239b53356ecf1d8ea7f13f24798fb9e9efe26304c084`.
The two untracked source hashes match `feature_8_evidence.md`. Both earlier
review findings are resolved. No `review_3.md` was created.

| Criterion | Evidence |
| --- | --- |
| 1. Frontend ownership | `main.go` injects focused capabilities into `internal/ui`. Navigation, commands and coordinator state stay in the UI. `TestCommandClosuresDoNotRaceCoordinator` passed under race. |
| 2. Independent dependencies and types | The production import inventory and `internal/architecture` check passed. Exported contracts expose independent values and resource methods. The negative tests still reject direct, transitive and inactive-variant leaks. |
| 3. Backend independence | Backend source and active dependency graphs show no UI or presentation framework dependencies. Main performs the concrete wiring. |
| 4. Retired package and bridges | `internal/app` is absent. Both allowlists are empty. The last fixture that needed a backend type now lives in a views test file. |
| 5. Focused presentation data and intents | UI callers use focused inventory, editor, action and deletion contracts. Focus-only tests inject small capability fakes. Persistence and real-service seam tests remain identifiable. |
| 6. Policy and revalidation | API, session and MCP coverage includes `TestSessionContractsRevalidateScopeAndChoice`, stale card revalidation, source outcomes, deletion admission and `TestDashboardScopeOverHTTP`. |
| 7. Card preservation | `TestEditorSavePreservesUneditableFieldsAndRejectsStaleRevision`, `TestEditorConcurrentSavesOnlyOneSucceeds` and `TestEditorExplicitClearAndHiddenFieldConflict` passed. |
| 8. Attachment ownership and input | Backend resources remain private. Generation cleanup and PTY restore regressions passed. The decoder batches ordinary bytes before the shared input gate, handles split and legacy Ctrl+K, and bounds the wait for ambiguous input. |
| 9. Preserved behavior and center | Refresh, session warnings and recovery tests passed. Center tests cover failure retention, delayed reads after a selection switch, workspace fallback and authored report takeover. `TestCenterReportSurvivesBackendReconstruction` covers the persisted history through the coordinator. Git helper cancellation is covered at the backend. |
| 10. Enforcement and preview removal | The checker and negative fixtures pass with both allowlists empty. The retired preview is absent, and attachment retains shared scrollback, stream, and serialized input. |
| 11. Composition and onboarding | The documented constructor and ownership match main and the frontend boundary. The live endpoint and process checks confirm cleanup after quit. Appedia's stale generated ownership references are recorded for a separately authorized refresh. |

## Independent verification

- `go build ./...`, `go vet ./...`, `go test -race ./... -count=1` and
  `git diff --check` passed on the final candidate. The implementation
  evidence also records the passing ordinary full suite.
- Changed-code lint reported zero issues. Sorting the final full-lint and
  clean-baseline logs produced identical content, including the 70 existing
  findings.
- A Linux PTY probe observed standalone Escape delivered in 30–33 ms on
  three runs, compared with roughly 802 ms on the prior candidate. A whole
  Up Arrow reached the attachment in one write on all three runs. The
  deterministic decoder tests cover individual ordinary writes and
  explicit split Ctrl+K inputs. The PTY restore tests exercise the command
  lifecycle.
- The current binary completed 18 split-key reattach/detach cycles and 36
  ordinary cycles across the ordinary and endpoint runs, with a new child
  acknowledgment after each reattach. The endpoint returned HTTP 401
  before quit and was unreachable afterward. The child process was gone.
  Independent logs are in `/tmp/kankuro-step8-final-review.i685tx/`.
- The earlier intermittent timeouts remain unattributed. The split-key
  defect and the two ordinary-input regressions are reproducibly resolved.
  Runtime checks were Linux and model-free. Delayed dashboard/session
  switching and report restart reconstruction were verified at coordinator
  and integration layers. Existing Appedia evidence was preserved.

## Reviewed implementation paths

- `docs/frontend-boundary.md`, `project.md`, `internal/architecture/allowlist.json`
- `internal/ui/center_internal_test.go`, `internal/ui/center_integration_internal_test.go`, `internal/ui/configuration_pending_internal_test.go`, `internal/ui/launch_focus_internal_test.go`
- `internal/ui/terminal/attach.go`, `internal/ui/terminal/attach_test.go`, `internal/ui/terminal/pty_test.go`
- `internal/ui/uitest/shape.go`, `internal/ui/views/dashboard_layout_test.go`, `internal/ui/views/modals_layout_test.go`, `internal/ui/views/puppet_fixture_test.go`
