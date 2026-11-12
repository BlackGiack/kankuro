# Verification — final MCP and launch-focus refresh

The validated source is the MCP work based on UX commit `474ae3a` (launch focus),
including MCP management and five dependency updates. The
MCP task and UX branch integrate without a reproducible functional failure in
the completed checks; one intermittent preview assertion is recorded below.

## Dependency and lint validation

The refreshed read-only Go package-manager snapshot confirms the requested
versions are selected and current at collection time:

| Module | Selected |
|---|---|
| `charm.land/bubbles/v2` | `v2.2.1` |
| `charm.land/bubbletea/v2` | `v2.1.1` |
| `charm.land/lipgloss/v2` | `v2.1.0` |
| `github.com/charmbracelet/x/ansi` | `v0.11.9` |
| `golang.org/x/sys` | `v0.50.0` |

Other direct dependencies retain their own snapshot status; the unrelated
`go-toml/v2` update is not part of the five requested changes. The tagged lint
run found and fixed shadowed error names and formatting in the startup-only
launch-focus fixture. These were test-only changes, not application behavior.

## Test execution and limits

Passed on final source:

```sh
go test -json ./... -count=1
go test -race ./internal/app/... -count=1
go test -tags=e2e ./tests/e2e -run '^Test(MCPManagerJourney|NativeLaunchFocusReturn)$' -count=1 -timeout=5m -v
go build ./...
go vet ./...
golangci-lint run
golangci-lint run --build-tags=e2e ./tests/e2e ./internal/app ./internal/ui/views
```

The initial combined app/views race run failed once at
`TestPreviewImmediateUpdate` (expected preview ID, got empty). Ten focused race
repetitions and the subsequent full app race run passed; views race tests also
passed in the original combined run. This intermittent result is retained rather
than presenting the initial broad run as clean. No unrelated production fix was
made for an unreproduced failure. Logs are in `tmp/final-mcp-validation/`.

Both E2E journeys passed after the fixture lint fixes. The MCP journey creates,
edits, cancels and confirms deletion through the actual executable and checks
persisted identity/command/state. It produced six captures, with list/modal at
80×24 and forms at 120×40. The launch-focus journey starts real Codex, confirms
naming, detaches and reattaches the same live process while retaining an older
terminated session. It produced three captures at 120×40. It submits no model
prompt; the model-driven registration/resume journey was not rerun.

The final E2E/model/evidence fingerprint is
`8995dacba1c91320604c55e44cb8736879e336aac4d1920c2bb31617e27fd09c`.
Named ordinary-test results were read from JSONL and bound to the fresh source
inventory. Current Go test-file hashes were checked against collected sources.
The Tests tab now has 31 features and 91 scenarios: 83 tested, eight partial,
127 evidence records and 11 captures. These are scoped assessments, not a
statement-coverage or functional-completeness score.

## Historical evidence

The embedded 2026-11-09 native-agent captures and failed registration result
retain their original outcome and fingerprint. They do not establish current
registration, detach/resume task execution or model behavior. A fresh atomic statement profile is now assigned to the matching current source.
The older atomic profile is archived in
`.maintenance/historical-statement-coverage-2026-11-09.json` with its original
metadata; it is not a paired baseline, and no coverage delta is claimed. Historical native result/image bytes
were preserved from the prior raw inventory when their ignored files were
absent. See `.maintenance/native-agent-verification-2026-11-09.md`.

## Appedia review

All 168 impact items were assessed with scope-specific notes in
`.maintenance/mcp-launch-focus-review.json`. MCP list/create/edit/deletion and
visit-scoping flows were added; the launch path was split into admission and
focus/naming/detach flows. Rules/tasks and existing rename/refresh explanations
were updated. Unchanged participant-associated paths retain their actual policy
and error semantics. No unresolved reference or map warnings appeared on build.

Current terminal captures/results are explicitly embedded in
`data/test-evidence.json`; original records are under the ignored E2E artifact
folders. Browser captures under `tmp/final-mcp-validation/screenshots/` check the
map, manager source, new flows, feature assessments and terminal capture pages
in light/dark themes. They document the appedia, not an application pixel-baseline
comparison. The final baseline is accepted only after the full route smoke test.

Runtime `.kankuro/mcp-servers.json` was unstaged and preserved byte-for-byte. It
is excluded from extraction/evidence, is absent from the index, and is not part
of the generated documentation. No credentials are printed in this report.

`git diff --check` passes for unstaged tracked work. The unrelated local
`feature.md` retains pre-existing trailing whitespace on lines 4–5 and is excluded
from the MCP commit alongside the other task/workflow files.

A content-only follow-up removes unused sequence participants and binds each launch-focus capture to its own checkpoint assertion, omitting unknown per-image capture times. Its single disposition is retained in `.maintenance/mcp-launch-focus-refinement.json`. Both passes use the same final application fingerprint.

Finalization accepted the refined baseline at `2026-11-12T22:20:00.887Z` after build/reference checks and the complete all-route smoke test passed. The current authored-content digest matches the accepted baseline, and no pending review items remain. Browser views were inspected in light and dark themes. E2E result files contain their standalone replay commands; the actual combined invocation used for this validation is shown above.

## Current statement coverage

The explicit coverage collector ran
`go test -json -count=1 -timeout=3m -covermode=atomic -coverprofile=<profile> ./...`
successfully without source changes. It measured **8,192 / 11,074 statements
(73.98%) across 16 packages**. Scope is the ordinary package-local suite: no E2E
tags and no transitive `-coverpkg`. Unmeasured packages are not invented as zero;
measured zero rows remain zero. Functional verdicts are unchanged.

The run metadata, raw profile and package rows are retained in
`data/test-evidence.json`; original profile/log outputs are under
`tmp/appedia-coverage/`. The profile, extracted model and evidence inventory share
source fingerprint
`8995dacba1c91320604c55e44cb8736879e336aac4d1920c2bb31617e27fd09c`.
The Code coverage page was checked for current eligibility and rendered values.

Coverage refresh finalization accepted the baseline at `2026-11-12T22:40:46.198Z` after the full route smoke check passed. Desktop and narrow browser views were inspected in light/dark themes; the current profile is eligible and displays 74.0%.
