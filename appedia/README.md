# KANKURO appedia

An interactive map of this codebase: packages, class diagrams, the real code,
20 sequence-diagram flows, the rules the code relies on, and where to start for
common changes.

**Open `index.html` in a browser.** No server needed. Without network access
the page still works; it falls back to system fonts.

## What to edit

Only `data/content.js`: package roles, flows, rules, tasks, glossary. All other
files are generated or copied by the build.

## Rebuild after the code changes

From the repository root:

```bash
(cd kankuro_skills/create-appedia/scripts/extract-go && go run . -root ../../../.. -out /tmp/kankuro-model.json)
node kankuro_skills/create-appedia/scripts/build.mjs --model /tmp/kankuro-model.json --content appedia/data/content.js --out appedia
```

The build fails if `content.js` names a type, method or function that no
longer exists, which is the signal to update a flow. Graphviz (`dot`,
`unflatten`) and Node 18+ must be installed.

Optional check, which loads every page in a simulated browser:

```bash
npm install --prefix /tmp/appedia-smoke jsdom@24
node kankuro_skills/create-appedia/scripts/smoke.mjs --site appedia --jsdom /tmp/appedia-smoke
```

How it was made and how to make one for another project:
`kankuro_skills/create-appedia/SKILL.md`.
