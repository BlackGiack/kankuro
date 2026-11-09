# Native-agent terminal acceptance

These are project-owned Go tests using `github.com/microsoft/tui-test/bindings/go`.
Every Go file has the `e2e` build tag. Ordinary `go test ./...` does not compile
or run the terminal journey and creates no screenshots.

Before a model-driven run, ask the user to confirm the provider/model and a
total token budget covering the intended runs and retries. Honor an existing
explicit approval for that scope; the saved model default is not standing
approval. Follow [AGENTS.md](AGENTS.md).

The initial planning estimate for one successful full journey is **20k–100k
input tokens and 1k–5k output tokens**. This is not measured usage or a guaranteed
limit. Tool calls, resumed context and retries affect the total. Share the
estimate before running, prefer actual usage when available, and distinguish
input/output, cached and reasoning tokens when reported. Free models still
consume tokens. Confirm any additional runs or changes outside the approved
provider/model and budget.

Run explicitly after that confirmation, for example when updating Appedia or
reviewing a change with Appediff:

```sh
go test -tags=e2e ./tests/e2e -run '^TestNativeAgentJourney$' -count=1 -timeout=10m -v
```

The test chooses OpenCode, then Codex, then Claude from PATH, skipping only when
none exists. The selected CLI must already have working provider/model
configuration and authentication. Login, quota, model, registration and resume
errors fail the test; they never trigger fallback to another installed agent.
A model request is part of this test. The router may select different backing
models, so a fast chat response does not guarantee tool execution or a passing
registration journey.

The test runs the actual Kankuro executable in a 120×40 terminal, launches the
native CLI from its dashboard, attaches, asks the agent to register its actual
conversation through MCP, checks durable session/card binding, verifies Ctrl+K
and dashboard keyboard restoration, stops and explicitly resumes the saved
conversation, and requires recall of a random secret omitted from the resume
prompt. OpenCode is explicitly instructed to discover its actual ID with
opencode session list --format json and send the registration arguments; the
harness does not invent or supply a conversation ID. Registration has a four-minute state wait and continuity has a
three-minute response wait within the ten-minute overall test timeout.
It does not compare screenshots against a visual baseline.
Screenshots require outline fonts covering the native CLI's symbols, including
U+2B1D. This environment uses Noto Sans Symbols2 alongside DejaVu Sans.
A screenshot preflight checks these glyphs before making any agent/model calls,
writes its probe only to a temporary directory, and fails early if fonts are
missing. Run it separately with
go test -tags=e2e ./tests/e2e -run '^TestNativeCaptureGlyphs$' -count=1.
An existing TUI_TEST_RECORDING_FONT_FAMILIES setting takes precedence.

Kankuro configuration, workspace, native conversation data and HOME/XDG paths
are disposable. Existing agent configuration/credentials and model preferences
are copied privately, without copying old conversations or server state.
For OpenCode v2, credentials are transferred with its supported auth export/import
commands and compared in memory, including active integration/value metadata.
Neither exported credentials nor their values appear in test output or artifacts.
The user's SQLite conversation database is not copied. Service configuration is
excluded from the copy; the fixture assigns its own loopback port and stops only
its own service during cleanup.
Environment-based provider credentials remain inherited. OpenCode workspace
permissions are allowed only in the disposable test project. Codex/Claude retain
their configured permissions; prompts that prevent the task completing fail it.
For OpenCode the test explicitly uses OpenRouter’s Free Models Router,
`openrouter/openrouter/free`, as requested. This selection is written
only to the disposable workspace; saved user settings remain untouched.

Outputs use stable names under `tests/e2e/artifacts/native-agent/` (gitignored):
`launched.png`, `registered.png`, `dashboard.png`, `resumed.png`,
`failure.png`, and `result.json`. Each explicit run removes only those known
files and rewrites reached checkpoints. There are no routine recordings.
`KANKURO_E2E_ARTIFACT_DIR` overrides the output directory; serialize runs that
share it. The JSON records the selected CLI, terminal dimensions, reached
assertions, actual outcome, timestamps and source fingerprint. A missing image
means its checkpoint was not reached, not that a previous image can be reused.

For Appedia, explicitly collect existing artifacts with the test-evidence
collector's `--file` option, link captures to this test, and use the recorded
result only when its fingerprint matches the collected head. Collect ordinary
Go coverage separately. Appediff can consume the same evidence snapshot via
`--test-evidence`; identify the actual base/head and do not attribute this run
to a different source snapshot.

The fingerprint uses the repository's current Appedia exclusions (see
`provenance_test.go`). If those exclusions change, update that list before the
next explicit run. Normal Appedia and Appediff collection never runs this test.
