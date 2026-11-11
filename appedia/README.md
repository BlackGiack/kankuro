# KANKURO appedia

Open [index.html](index.html) directly in a browser. No server is needed.

Updated 2026-11-11 for commit `7a58943` (session rename). The site covers 18
packages, 58 flows, 28 rules and 31 common-change tasks. The Tests tab contains
29 features, 78 scenarios and 103 reviewed evidence records: 70 scenarios tested
and 8 partial. These are scoped assertion assessments, not a completeness score.

The session-rename feature has nine supported scenarios. Its three flows explain
editing/cancelling a draft, scoped asynchronous save and retry, and responsive
layout/help. Tests directly cover cancellation and scope changes, immediate
refresh without attachment or process replacement, and saved-name restoration
through a fresh repository, manager and coordinator.

Existing partial scenarios retain their limits. The native-agent captures and
failed registration result are historical evidence from 2026-11-09; this refresh
did not run a model-driven terminal journey. The prior statement profile is also
historical and is not attributed to the current head. Current ordinary test
execution is recorded separately. See [VERIFICATION.md](VERIFICATION.md).

## What to edit

Edit `data/content.js` for package roles, flows, rules, tasks and test assessments.
The model, diagrams, source inventory and content index are generated. During a
refresh, resolve the generated impact entries under `.maintenance/pending/`.
The accepted rename dispositions are in
[.maintenance/session-rename-review.json](.maintenance/session-rename-review.json).

## Reproducing a refresh

Requirements: Go 1.27.3, Node 18+, Python 3.9+, Graphviz and jsdom for smoke checks.
The refresh extracts Go sources and queries dependency versions through Go's
package manager. It does not execute project tests automatically.

From the repository root:

```sh
appedia_skill_dir=.agents/skills/create-appedia
appedia_exclusions=kankuro_skills/create-appedia,kankuro_skills/create-appediff,tmp,appedia,.agents,.kankuro
node "$appedia_skill_dir/scripts/appedia.mjs" --action prepare --repo . --lang go --exclude "$appedia_exclusions"
mkdir -p tmp
go test -json ./... > tmp/appedia-refresh-tests.jsonl
# Review pending/impact.md; update content.js, source-bound evidence/results,
# and every pending/impact.json disposition before finalization.
node "$appedia_skill_dir/scripts/appedia.mjs" --action rebuild --repo .
node "$appedia_skill_dir/scripts/appedia.mjs" --action finalize --repo . --jsdom /tmp/appedia-smoke
```

This refresh used the stored extraction settings for `prepare`. It merged only
the previously embedded native result/capture records into the fresh test
inventory, keeping their historical metadata. It reassessed unchanged test
assertions against source hashes, recorded current ordinary results from JSONL,
and added the new rename assertions. No old run or profile was relabeled.

The exact extractor and builder used by that pipeline are:

```sh
appedia_repo_root="$PWD"
(cd .agents/skills/create-appedia/scripts/extract-go && go run . -root "$appedia_repo_root" -out "$appedia_repo_root/appedia/.maintenance/pending/model.json" -exclude kankuro_skills/create-appedia,kankuro_skills/create-appediff,tmp,appedia,.agents,.kankuro)
node .agents/skills/create-appedia/scripts/build.mjs --model appedia/.maintenance/pending/model.json --dependencies appedia/.maintenance/pending/dependencies.json --content appedia/data/content.js --test-evidence appedia/.maintenance/pending/test-evidence.json --out appedia
```

Finalization accepts the new baseline only after validation and smoke checks.
Building by hand does not advance it. Use `--action refresh-renderer` for a
renderer-only update. No commits were made during this refresh.

## Explicit terminal tests

Project-owned terminal tests use the tui-test Go binding and the `e2e` build tag.
Before a model-driven run, choose a provider/model and confirm a total token
budget covering the runs and retries, per [tests/e2e/AGENTS.md](../tests/e2e/AGENTS.md).
Ordinary tests and the model-free glyph preflight do not contact a provider.

See [tests/e2e/README.md](../tests/e2e/README.md) for the explicit journey, isolation,
font requirements and artifact paths. Collect only artifacts from the actual
run and keep checkpoint/assertion limits explicit; no screenshot-baseline
comparison is implied.
