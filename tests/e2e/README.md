# Terminal E2E tests

These tests drive a real Kankuro executable with `tui-test`. The `e2e` build tag
enables model-free terminal coverage. The model-driven journey requires the
additional `token_usage` Go build tag and explicit opt-in.

## Offline Codex and workspace journeys

```sh
go test -tags=e2e ./tests/e2e -run '^Test(NativeLaunchFocusReturn|CodexOfflineWorkspaceAndReports)$' -count=1 -timeout=3m -v
```

These tests launch the real Codex CLI in disposable configuration and workspaces.
They use a dummy key and an unreachable loopback model endpoint, and never
submit an agent prompt. The launch-focus case checks that naming, Ctrl+K, opening
Settings and Enter preserve the selected live process. The workspace cases check
the directory and Git panels, live refresh after creating a file, minimum-size
handling, stop and relaunch, and reconstruction of seeded authored reports
after restarting Kankuro. Each workspace case waits for the actual refresh tick.

Screenshots and a result record go to `tests/e2e/artifacts/launch-focus/` by
default. The workspace cases use the `directory` and `git` subdirectories.
Override the artifact root with `KANKURO_E2E_ARTIFACT_DIR` for isolated runs.

## GPT-6 Luna journey with token usage

Before running, confirm the provider, model and total token budget in accordance
with [AGENTS.md](AGENTS.md). The recorded approval for a run governs that run
only. This runner selects one test and disables test retries.

```sh
KANKURO_E2E_ARTIFACT_DIR=/tmp/kankuro-luna-e2e ./tests/e2e/run token-usage
```

The runner opts in with `KANKURO_E2E_TOKEN_USAGE=1` and compiles the test with
`-tags=e2e,token_usage`. Go build constraints require underscores in tag names.
The ordinary `e2e` tag cannot run this scenario.

The test launches Codex with `gpt-6-luna` and low reasoning in a disposable
configuration, using the configured Codex account. A local fixture sender sends
one correlated task. Codex registers its real conversation and submits a progress
report. The test checks persisted attribution and the live workspace-to-report
transition. It checks the retained report, selects the saved conversation and
verifies that a new Codex process resumes the same Kankuro session and workspace.
The test then checks the budget before submitting one short recall prompt. The
total budget is 60,000 input tokens (including cached input) and 2,000 output
tokens. Recall is skipped above 44,000 input or 1,600 output tokens to reserve
16,000 input and 400 output tokens for that prompt. If recall is skipped, the
test still shuts down normally and records `budget-limited` in `result.json`
with the reason, completed assertions, screenshots and measured usage. Rollout
counts are observations, since the provider does not expose a client-enforced
hard cap for this interactive session.

Screenshot preflight, package compilation and model-free cases make no model
requests. Captures are terminal checkpoints rather than pixel-baseline comparisons.

## MCP configuration journey (no model calls)

`TestMCPManagerJourney` drives the real Kankuro executable using tui-test, with
an isolated empty configuration and disposable HOME/XDG directories. It never
launches an agent or contacts a model provider, so no provider or token-budget
confirmation is required.

```sh
go test -tags=e2e ./tests/e2e -run '^TestMCPManagerJourney$' -count=1 -timeout=3m -v
```

The scenario opens the dashboard MCP list, creates a custom server, edits it
while retaining its identity and command, opens and cancels deletion, confirms
deletion, checks the saved registry, and returns/reopens through the dashboard.
It checks visible state and help before taking screenshots: `list.png`,
`new.png`, `edit.png` at 120×40; `edited-list.png`, `confirm.png`, and
`deleted.png` at the supported minimum 80×24. These are asserted journey
checkpoints, not pixel comparisons against a baseline.

Artifacts live under `tests/e2e/artifacts/mcp-manager/` (gitignored).
`KANKURO_MCP_E2E_ARTIFACT_DIR` overrides this directory. Serialize runs sharing
the same directory. Each run removes only its known artifact names. `result.json`
records command, WORKTREE fingerprint, actual status/timestamps and reached
capture checkpoints with terminal sizes and assertions. A failed run can leave
`failure.png`; missing checkpoints are not passes or reusable old images.
Collect these captures and the result explicitly when refreshing Appediff/Appedia.
