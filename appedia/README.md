# KANKURO appedia

Open [index.html](index.html) directly in a browser. No server is needed.

Refreshed 2026-11-09 against the current working tree. Covers 18 packages,
55 flows, 28 rules and 31 common-change tasks. The Tests tab has 28 features,
69 scenarios and 94 reviewed evidence records: 61 tested and
8 partial. These are scoped assertion assessments, not a
functional completeness percentage.

All 29 cleanup findings and additional dead helpers remain resolved; see
[the audit dispositions](../tests_to_remove.md). Temporary-workspace Space/save
has direct Go regression coverage. Repaired tests and neighboring sources are
published with current hashes; the historical audit filter is unused.

The native fixture transfers and verifies current OpenCode credentials through
supported auth export/import, excludes service configuration and uses its own
loopback port. User conversations are not copied. The selected model is
OpenRouter's Free Models Router, openrouter/openrouter/free. The run launched and attached with the requested model, but no conversation binding was saved within the four-minute registration wait. Detach/restoration and resume were not reached.

The old Big Pickle quota screen came from incomplete fixture setup and does not
describe the user's configured environment. Current captures and the actual run
record replace that snapshot. The screenshot preflight passes with the installed
Noto Sans Symbols2 fallback and writes only a temporary probe.

Ordinary-suite atomic coverage passed: 7,897 / 10,760 statements
(73.4%), excluding E2E-tagged tests. The profile and separate
native run match the current source fingerprint. Dependencies lists 12 direct
Go declarations with versions queried through Go's package manager.

## Explicit terminal tests

Every E2E Go source carries //go:build e2e. Ordinary tests and normal
Appedia/Appediff collection do not launch agents or create screenshots.

    go test -tags=e2e ./tests/e2e -run '^TestNativeAgentJourney$' -count=1 -timeout=10m -v

See [the E2E instructions](../tests/e2e/README.md) for configuration, model, timeouts,
fonts and artifact paths. Each explicit run replaces only known filenames;
missing checkpoints remain missing. No image-baseline comparison is claimed.

## Reproducing a refresh

Edit data/content.js for explanations and assessments. Models, diagrams, source
inventory and content index are generated. Requirements: Go 1.27.3, Node 18+,
Python 3.9+, Graphviz, and jsdom for smoke checks.

From the repository root:

    appedia_skill_dir=.agents/skills/create-appedia
    appedia_exclusions=kankuro_skills/create-appedia,kankuro_skills/create-appediff,tmp,appedia,.agents,.kankuro

    python3 "$appedia_skill_dir/scripts/test-evidence.py" --root . --exclude "$appedia_exclusions" --revision WORKTREE --run-go-coverage --file tests/e2e/artifacts/native-agent/result.json --file tests/e2e/artifacts/native-agent/launched.png --file tests/e2e/artifacts/native-agent/failure.png --out /tmp/kankuro-router-evidence.json
    node "$appedia_skill_dir/scripts/appedia.mjs" --action prepare --repo . --lang go --exclude "$appedia_exclusions"
    node "$appedia_skill_dir/scripts/appedia.mjs" --action rebuild --repo . --test-evidence /tmp/kankuro-router-evidence.json
    node "$appedia_skill_dir/scripts/appedia.mjs" --action finalize --repo . --jsdom /tmp/appedia-smoke --test-evidence /tmp/kankuro-router-evidence.json

Collect only artifacts produced by that explicit run. Prepare does not execute
tests. Review every pending impact item before finalize. The explicit coverage
flag executes the ordinary suite. For read-only collection, omit it and leave
coverage uncollected or import a profile with its real metadata and matching
fingerprint. This refresh imported its freshly executed ordinary profile and
then collected existing terminal artifacts, without executing E2E again.

Appediff accepts the same source-bound snapshot through --test-evidence.
See [VERIFICATION.md](VERIFICATION.md) for actual outcomes and limits.
Prior audit/cleanup review artifacts remain as history. Nothing was committed.
