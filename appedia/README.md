# KANKURO appedia

An interactive map of this codebase: packages, class diagrams, the real code,
39 sequence-diagram flows, the rules the code relies on, and where to start for
common changes.

This snapshot covers the current working tree on `feat/claude-resume` at
`5c6c0f7`, including the staged removal of legacy registration-block cleanup
(2026-10-29). It includes the local MCP control plane, shared application service,
native conversation bindings, explicit progress/completion replies, named sessions,
workspace trust, immediate human attachment and the separate MCP startup gate.

The existing human UI, settings, workspace selection, copy ownership, refresh,
preview and external IDE/terminal flows are retained. Package roles and new flows
reuse the reviewed appediff where applicable; change-only and removed-code steps
were converted to current-state descriptions. The appediff remains a separate
historical comparison under `tmp/appediff/`.

**Open `index.html` in a browser.** No server needed. Without network access
the page still works; it falls back to system fonts.

## What to edit

Edit `data/content.js` for package roles, flows, rules, tasks, and the glossary.
During a refresh, also review the generated impact report and record its
dispositions in `.maintenance/pending/impact.json`. Models, the content index,
renderer assets, and accepted baseline metadata are generated.

## Rebuild after the code changes

This refresh used the installed create-appedia skill at
`.agents/skills/create-appedia`, with Node 20, Go 1.27.3 and Graphviz. That skill
directory is ignored by git; another checkout needs the same skill installed,
or `appedia_skill_dir` set to its actual location. The older bundled scripts under
`kankuro_skills/create-appedia` do not include this refresh pipeline.

From the repository root, use the same extraction settings:

```bash
appedia_skill_dir=.agents/skills/create-appedia
node "$appedia_skill_dir/scripts/appedia.mjs" --action prepare --repo . --lang go \
  --exclude kankuro_skills/create-appedia,kankuro_skills/create-appediff,tmp,appedia,.agents,.kankuro
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

## Verification

The staged code cleanup already passed focused driver regressions, full tests,
race tests, build, vet and lint; evidence is in
`.kankuro/remove-legacy-cleanup-verification/`. This documentation refresh does
not repeat unchanged source tests. The finalization command above rebuilds and
checks every appedia route before recording the new working-tree baseline.
Browser captures are retained under `tmp/appedia-refresh/verification/`.
Finalization passed on 2026-10-29: the build and exhaustive route smoke test
succeeded, the accepted baseline includes the staged driver cleanup, and all
12 light/dark browser captures were inspected. Runtime credentials and local
verification artifacts are excluded from the source inventory.
Live CLI workflows were not rerun for this refresh; Claude validation is ongoing
in the user's other environment.
