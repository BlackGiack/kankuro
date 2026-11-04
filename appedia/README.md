# KANKURO appedia

Open **index.html** directly in a browser; no server is needed. The site falls
back to system fonts when offline.

Updated from the working tree on 2026-11-04, based on commit `32f4cec` and the
current dashboard keyboard changes. It covers 18 application packages, 48 flows,
24 rules and 28 common-change entry points.

The explanations cover dashboard ownership and sharing, scoped UI and MCP
operations, immutable session membership, durable deletion recovery, safe
navigation and configuration startup through `internal/configstartup`.
Retired dashboard migration tooling and legacy production paths are removed.
The appediff under `tmp/appediff/` remains a separate change review.

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
348 impact items. The previous accepted maintenance baseline is preserved;
the reviewed candidate is retained in `.maintenance/pending/`.

For a future update where an exhaustive route check is wanted, finalization
runs that check before accepting the candidate as the maintenance baseline:

```bash
npm install --prefix /tmp/appedia-smoke jsdom@24
node "$appedia_skill_dir/scripts/appedia.mjs" --action finalize --repo . \
  --jsdom /tmp/appedia-smoke
```

## Verification

The build, source consistency check and bounded 11-route browser-model check
passed. Ten light/dark browser captures were generated and representative
package, map and flow views were inspected. See [VERIFICATION.md](VERIFICATION.md).
The exhaustive smoke test was not rerun, following the user's request to stop
it; finalization and baseline acceptance are therefore not claimed.
