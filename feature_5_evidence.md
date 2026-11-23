# Feature 5 implementation evidence

Implementation is awaiting independent review; this is not a review or PASS.

- Source contracts and service operations validate path-list expected content and position under the repository update lock, or workspace ID and expected content. Source create is explicit; saved-plus-sync-failure, stale conflict and pre-save failure are distinct. Source UI loads, mutates and syncs asynchronously, captures edit identity, retains drafts on failure, and correlates results to a view visit and request.
- MCP rows expose identity, label and availability; the separately loaded editor owns detached connection fields. Repository operations recheck built-in and referenced-delete rules. Settings use contract defaults, asynchronous loading and captured saves. Registry refresh carries focused choice results and MCP rows, retains failed inventories and certifies change tokens only after successful stable reads.
- `internal/api/configuration_sources_test.go` covers duplicates, stale reordered and cross-kind tokens, workspace reorder/revision, concurrent creates, cancellation, malformed input, injected pre-save errors and saved-but-sync-failed outcomes across source kinds. Source manager tests cover stale drafts and failed/obsolete reads. MCP tests cover built-in protection, form parsing and copy isolation; settings tests cover failed saves and late/failed reads. The coordinator race test exercises source, settings and MCP commands.
- `go test ./... -count=1`: PASS on final full run. Earlier concurrent test runs intermittently failed `TestPreviewAutoSelection`, `TestSessionLauncherBehavior`, and `TestReplyWaitEndsOnRecipientStopAndCallerRevocation/manager_shutdown`; the first two focused reruns and the later full run passed. These involve session startup/shutdown and warrant attention if reproduced.
- `go test -race ./internal/config ./internal/api ./internal/app -run 'Source|MCP|Settings|Refresh|CommandClosures' -count=1`: PASS; additional focused source fault/concurrency race cases PASS.
- `go build ./...`, `go vet ./...`, and `git diff --check`: PASS. `golangci-lint run --new-from-rev=HEAD ./...`: PASS on changed lines. Unrestricted `golangci-lint run ./...` reports existing issues in `internal/architecture/boundary_test.go`; full lint is not clean.
- Isolated terminal checks with `KANKURO_CONFIG_DIR` and `KANKURO_TMP_DIR` under `/tmp/kankuro-feature5-manual`: at 80×24 created a workspace source, verified built-in MCP editing was denied, and saved settings; at 120×40 reopened source and settings and created an MCP. No native model-driven E2E was run; the plan requires separate provider/model and budget authorization for that.
- The feature-5 import exceptions were removed. Remaining direct and transitive composition, preview and attachment/deletion exceptions are owned by features 6–8 and listed in `docs/frontend-boundary.md` and the architecture allowlists. No commit was made.

## Review fixes (2026-11-16)

- Settings save results carry their originating settings visit and draft identity. A saved result updates applied settings globally, while navigation and error feedback require the same current visit and draft. Pending settings saves show status and block repeat submission; a newer visit can edit a draft and receives an explanation if it attempts to save before the previous operation finishes.
- MCP forms show pending status and reject keyboard and paste edits or resubmission during a save. Failure unlocks the form and permits retry; leaving remains available.
- Coordinator regressions hold real commands while navigating to Puppet Sources, reopening settings or editing a newer draft, for successful and failed settings saves. MCP regressions hold save completion, exercise blocked keyboard/paste input and duplicate submission, and verify failure retry and successful persistence.
- `go test ./... -count=1`, the `Source|MCP|Settings|Refresh|CommandClosures` race selection, `go build ./...`, `go vet ./...`, changed-line `golangci-lint`, and `git diff --check` passed. Initial verification encountered a full Go build cache; clearing that regenerable cache and using a temporary cache allowed the checks to complete.

These fixes await independent re-review; no PASS or commit is claimed here.

## Pending-feedback re-review fix

Settings now render the saving message directly from the saving state and use a boolean for blocked-save feedback. Completion clears both states without inspecting display strings. Coordinator regressions exercise Enter and `e` during a held save, then verify failed completion clears pending text while retaining the draft, error and retry. `go test ./internal/app ./internal/ui/views ./internal/architecture -count=1`, focused settings race tests, changed-line lint and `git diff --check` passed. Await independent re-review; no commit or review PASS is claimed.

## Independent review fixes

- A delayed initial settings read updates the editing baseline and untouched fields while tracking actual text and toggle edits. Cancel restores the loaded baseline. Saving is held with visible feedback until the initial read succeeds or fails, preventing a save of defaults while the result is outstanding. Regressions use the real repository to check delayed-read save/cancel with a non-default terminal command and false visibility flags, explicit edits, and save attempts before successful or failed reads.
- Both MCP deletion commands capture the consumer configuration capability. Coordinator regressions drive both entry points, verify success and failure propagation, and confirm the command retains the capability captured before execution.
- `go test ./... -count=1`, the focused configuration race selection, additional regression race tests, `go build ./...`, `go vet ./...`, changed-line lint, and `git diff --check` passed. These changes await independent re-review; no commit or PASS is claimed.
