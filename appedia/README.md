# KANKURO appedia

Open [index.html](index.html) directly in a browser. No server is needed.

This refresh describes the final worktree on UX commit `474ae3a`, including the
MCP manager, launch/naming focus and stale-refresh protection, and the five
suggested dependency updates. It contains 18 packages, 63 flows, 29 rules and
33 common-change tasks. The Tests tab has 31 features, 91 scoped scenarios,
127 evidence records and 11 terminal captures: 83 scenarios have direct support
and eight retain partial evidence. These counts are not coverage percentages.

MCP management has eight supported scenarios, including cross-dashboard delays,
injected write-error recovery and a real tui-test configuration journey with six
captures. Launch focus has five supported scenarios and three startup-only
native-terminal captures: launch/name, detach and Enter reattachment retain the
created session's identity. Both current E2E journeys passed without submitting
an agent task to a model.

The Code coverage tab now shows a fresh ordinary-suite statement profile: 8,192
of 11,074 statements covered (73.98%) across 16 measured packages. It excludes
E2E-tagged tests and does not use transitive `-coverpkg` instrumentation. The older
profile is archived with its original metadata; no before/after delta is claimed.
The older model-driven registration failure remains historical. Broader existing partial
scenarios retain their limits. See [VERIFICATION.md](VERIFICATION.md) for exact
checks, the intermittent preview-test result, and source provenance.

## Editing and refreshing

Edit `data/content.js` for package roles, flows, rules, tasks and test assessments.
Models, diagrams, inventories and `data/content-index.json` are generated.
The accepted impact dispositions are recorded in
[.maintenance/mcp-launch-focus-review.json](.maintenance/mcp-launch-focus-review.json).

Requirements: Go 1.27.3, Node 18+, Python 3.9+, Graphviz and jsdom. From the root:

```sh
appedia_skill_dir=.agents/skills/create-appedia
appedia_exclusions=kankuro_skills/create-appedia,kankuro_skills/create-appediff,tmp,appedia,.agents,.kankuro
node "$appedia_skill_dir/scripts/appedia.mjs" --action prepare --repo . --lang go --exclude "$appedia_exclusions"
go test -json ./... -count=1 > tmp/final-mcp-validation/tests.jsonl
# Review pending/impact.md, update content.js and source-bound test evidence,
# then set every pending/impact.json disposition with its reasoning.
node "$appedia_skill_dir/scripts/appedia.mjs" --action rebuild --repo . --test-evidence tmp/final-mcp-validation/appedia-evidence.json
node "$appedia_skill_dir/scripts/appedia.mjs" --action finalize --repo . --jsdom /tmp/appedia-smoke --test-evidence tmp/final-mcp-validation/appedia-evidence.json
```

The pipeline parses sources, queries direct dependency versions read-only via Go,
and collects tests without executing them. It advances the baseline only after
reviewed impact dispositions, content validation and all-route smoke pass.
Use `--action refresh-renderer` for renderer-only changes.

The underlying extractor/build commands are:

```sh
appedia_repo_root="$PWD"
(cd .agents/skills/create-appedia/scripts/extract-go && go run . -root "$appedia_repo_root" -out "$appedia_repo_root/appedia/.maintenance/pending/model.json" -exclude kankuro_skills/create-appedia,kankuro_skills/create-appediff,tmp,appedia,.agents,.kankuro)
node .agents/skills/create-appedia/scripts/build.mjs --model appedia/.maintenance/pending/model.json --dependencies appedia/.maintenance/pending/dependencies.json --content appedia/data/content.js --test-evidence tmp/final-mcp-validation/appedia-evidence.json --out appedia
```

## Terminal evidence

Project-owned terminal tests use tui-test behind the `e2e` build tag. The current
checks were run explicitly:

```sh
go test -tags=e2e ./tests/e2e -run '^Test(MCPManagerJourney|NativeLaunchFocusReturn)$' -count=1 -timeout=5m -v
```

Artifacts live in the ignored `tests/e2e/artifacts/mcp-manager/` and
`tests/e2e/artifacts/launch-focus/` folders. Their `result.json` files record
source fingerprints, commands, outcomes and reached screenshot checkpoints.
Collect those files explicitly with the skill's `test-evidence.py --file` options;
this refresh retained the raw snapshot in `data/test-evidence.json`. Captures
are asserted journey checkpoints, not pixel-baseline comparisons.

`TestNativeLaunchFocusReturn` starts Codex in private state but submits no prompt.
The separate `TestNativeAgentJourney` contacts a model provider: before running
it, follow [tests/e2e/AGENTS.md](../tests/e2e/AGENTS.md) and confirm provider/model
and a total token budget. Existing saved choices are not standing approval.

## Refreshing statement coverage

Explicitly collect an instrumented profile with the same source exclusions, then
pass that evidence snapshot to rebuild/finalize. Normal documentation collection
does not run tests:

```sh
python3 .agents/skills/create-appedia/scripts/test-evidence.py --root . --revision WORKTREE --exclude kankuro_skills/create-appedia,kankuro_skills/create-appediff,tmp,appedia,.agents,.kankuro --run-go-coverage --out tmp/appedia-coverage/evidence.json
```

Retain authored evidence images/results when combining the new coverage snapshot
with `data/test-evidence.json`. The collector records the actual command, mode,
scope, toolchain, environment and source fingerprint, rejects source changes
during the run, and keeps the raw profile and Go JSON log beside its output.
