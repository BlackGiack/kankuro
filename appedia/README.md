# KANKURO appedia

An interactive map of this codebase: packages, class diagrams, the real code,
30 sequence-diagram flows, the rules the code relies on, and where to start for
common changes.

This snapshot covers the root executable and public module path, startup injection
of the embedded MCP-CLI skill, card editor controls (including Ctrl+X to clear
the focused Sub Workspaces field), and preserved sub workspace copies with
explicit ownership and rollback.

**Open `index.html` in a browser.** No server needed. Without network access
the page still works; it falls back to system fonts.

## What to edit

Edit `data/content.js` for package roles, flows, rules, tasks, and the glossary.
During a refresh, also review the generated impact report and record its
dispositions in `.maintenance/pending/impact.json`. Models, the content index,
renderer assets, and accepted baseline metadata are generated.

## Rebuild after the code changes

This refresh used the installed create-appedia skill at
`.agents/skills/create-appedia`, with Node 20, Go 1.26.1 and Graphviz. That skill
directory is ignored by git; another checkout needs the same skill installed,
or `appedia_skill_dir` set to its actual location. The older bundled scripts under
`kankuro_skills/create-appedia` do not include this refresh pipeline.

From the repository root, use the same extraction settings:

```bash
appedia_skill_dir=.agents/skills/create-appedia
node "$appedia_skill_dir/scripts/appedia.mjs" --action prepare --repo . --lang go \
  --exclude kankuro_skills/create-appedia,kankuro_skills/create-appediff,tmp,appedia
```

Review `.maintenance/pending/impact.md`, update `data/content.js`, and mark every
item in `impact.json` as `updated`, `unchanged`, or `removed` with a reason. Then:

```bash
node "$appedia_skill_dir/scripts/appedia.mjs" --action rebuild --repo .
npm install --prefix /tmp/appedia-smoke jsdom@24
node "$appedia_skill_dir/scripts/appedia.mjs" --action finalize --repo . \
  --jsdom /tmp/appedia-smoke
```

Rebuild validates references and regenerates diagrams. Finalize additionally
checks every route in a simulated browser before accepting a baseline. Failed
checks leave the previous baseline intact. The canonical bundled MCP-CLI asset
is included in the input history; tool fixtures and generated sites are excluded.

The first refresh established provenance for this legacy site, whose earlier
snapshot had no accepted file history. `.maintenance/baseline.json` and
`.maintenance/baseline-model.json` now support subsequent comparisons.

To check an already built site without changing its baseline:

```bash
node "$appedia_skill_dir/scripts/smoke.mjs" --site appedia --jsdom /tmp/appedia-smoke
```

How it was made and how to make one for another project:
the installed skill's `SKILL.md` and `references/content-guide.md`.
