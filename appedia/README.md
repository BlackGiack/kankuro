# KANKURO appedia

Open **index.html** directly in a browser; no server is needed. The site falls
back to system fonts when offline.

Updated against commit `60e5808` on 2026-11-07. The site covers 18 application
packages, 55 flows, 28 rules and 31 common-change entry points.

It describes dashboard ownership and sharing, scoped UI and MCP operations,
immutable session membership, durable deletion recovery, safe navigation and
configuration startup through `internal/configstartup`. Messaging now includes
automatic orchestrator wake-ups for completed, blocked and question reports,
silent progress, shared delivery admission, exact-activation authorization,
retry scheduling, byte evidence and notification recovery.

The current create-appedia skill was reloaded from its installed `SKILL.md`;
the refresh regenerated the model, diagrams and source associations. Appediff
sites under `tmp/` remain separate change reviews.

## Dependencies

The Dependencies tab shows the 12 direct Go dependencies, selected versions,
latest known releases and comparison status. Versions were queried through
Go's package manager on 2026-11-07; the snapshot is embedded in `data/model.js`
so it works offline. Latest does not imply compatibility, and collection does
not install or update project dependencies.

Prepare now collects this snapshot automatically. To refresh only dependencies
and rebuild the existing extracted model, run from the repository root:

```bash
python3 .agents/skills/create-appedia/scripts/dependencies.py \
  --root . --lang go --out appedia/.maintenance/pending/dependencies.json \
  --exclude kankuro_skills/create-appedia,kankuro_skills/create-appediff,tmp,appedia,.agents,.kankuro
node .agents/skills/create-appedia/scripts/appedia.mjs --action rebuild --repo .
```

## Editing and refreshing

Edit `data/content.js` for explanations, flows, rules, tasks and the glossary.
Models, diagrams and `data/content-index.json` are generated. Refresh impact
reviews also require dispositions and reasons in the pending `impact.json`.

This refresh used the installed skill at `.agents/skills/create-appedia`, Node,
Go and Graphviz. The skill directory is ignored by git; install the same skill
in another checkout. From the repository root:

```bash
appedia_skill_dir=.agents/skills/create-appedia
node "$appedia_skill_dir/scripts/appedia.mjs" --action prepare --repo . --lang go \
  --exclude kankuro_skills/create-appedia,kankuro_skills/create-appediff,tmp,appedia,.agents,.kankuro
```

Review `.maintenance/pending/impact.md`, update `data/content.js`, and record
every impact item's disposition and reason. Then rebuild:

```bash
node "$appedia_skill_dir/scripts/appedia.mjs" --action rebuild --repo .
```

Rebuild validates references and regenerates the site. This update reviewed all
367 impact items. The previous accepted maintenance baseline is preserved;
the reviewed candidate is retained in `.maintenance/pending/`.

For a future update where an exhaustive route check is wanted, finalization
runs that check before accepting the candidate as the maintenance baseline:

```bash
npm install --prefix /tmp/appedia-smoke jsdom@24
node "$appedia_skill_dir/scripts/appedia.mjs" --action finalize --repo . \
  --jsdom /tmp/appedia-smoke
```

## Verification

Build and reference validation passed. All 1,491 source symbols agree with the
fresh extraction and working tree. The focused navigation check covers 143
routes, including every flow entry, the updated messaging steps, key code and
isolated type views and dependency-table navigation/data checks. Sixteen
light/dark Chrome captures were generated and
representative map, package, code, isolated and flow views were inspected.
See [VERIFICATION.md](VERIFICATION.md).

The exhaustive smoke test was not rerun, following the user's earlier request
to stop it. The accepted baseline remains intact; finalization and baseline
acceptance are not claimed. All 367 pending impact items have reviewed
reasons and dispositions. No application code or git index was changed.
