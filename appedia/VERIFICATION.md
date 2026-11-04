# Appedia verification — 2026-11-04

The site reflects the current working tree based on `32f4cec`, including the
latest dashboard keyboard and deletion-retry changes. No application source
was modified by this documentation refresh.

- Build: passed with reference validation and regenerated diagrams.
- Source consistency: all 1,459 symbol sources match fresh extraction, the
  served browser model and the working-tree files. All 18 packages have roles.
- Content: 48 current-state flows, 24 rules and 28 common-change tasks.
- Impact review: all 348 items have dispositions and reasons.
- Focused simulated-browser navigation: 11 routes, no errors. Covered map,
  configstartup, dashboard and scoped-service types, dashboard-list Update,
  startup/navigation/deletion recovery flows, rules, tasks and reference.
- Browser captures: 10 generated with Chrome at 1600 × 1100, light and dark.
  Inspected map/light, navigation/dark, configstartup/dark and deletion
  recovery/light; these views render readable diagrams and supporting text.
- Exhaustive route smoke: not rerun, following the user's request to stop the
  long-running test. This is a bounded check, not an all-route pass.
- Previous accepted maintenance baseline: unchanged. The reviewed current
  candidate remains under `.maintenance/pending/`; finalize was not run.

Local evidence and repeatable focused-check scripts are under
`tmp/appedia-refresh/`: `build.log`, `model-check.log`, `focused-smoke.log`,
`check-model.py`, `focused-smoke.mjs`, `screenshots.log` and `screenshots/`.
These local artifacts are ignored by git. Asset hashes and verification status
are retained in `.maintenance/pending/verification.json`.
