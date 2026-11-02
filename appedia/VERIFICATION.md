# Appedia verification — 2026-11-07

The site describes commit `60e5808` and the current working tree. This refresh
changes documentation and generated appedia assets only.

## Results

- Reloaded the installed create-appedia `SKILL.md` and content guide, then used
  its prepare/rebuild pipeline with the existing extraction settings.
- Build and reference validation passed without warnings.
- All 18 application packages have current roles; all 1,491 source symbols match
  fresh extraction, the served browser model and the working-tree source files.
- Content covers 55 current-state flows, 28 rules and 31 common-change tasks.
  Every flow has 4–13 steps with valid code references.
- All 367 impact items have reviewed dispositions and reasons. The prior
  dashboard review was retained where the wake-up feature does not change the
  documented behavior; new messaging and startup/shutdown associations were
  updated. Retired declarations remain excluded from current explanations.
- Focused simulated-browser navigation passed: 142 routes, no errors. Covers
  every flow entry, detailed messaging/retry/recovery steps, five package pages,
  key code, one- and two-hop preparedDelivery isolation, rules, tasks and reference.
- Twelve Chrome captures were generated at 1600 × 1100 in light/dark themes.
  Inspected representative map, API package, preparedDelivery.publish code,
  preparedDelivery isolation, report persistence and notification-retry views.

## Limits and maintenance state

The exhaustive every-route smoke was not rerun, following the user's earlier
request to stop that long-running test. Finalize was therefore not run and no
new accepted baseline is claimed. The reviewed candidate and impact report
remain in `.maintenance/pending/`; `.maintenance/baseline.json` and the accepted
baseline model are unchanged.

No application tests were rerun for this documentation-only refresh. The wake-up
implementation passed the previously completed code review, including full Go
tests/vet/lint, API race checks and real MCP/PTY notification integration tests.
No new external agent CLI acceptance result is claimed.

The installation README now agrees with the Go version declared by `go.mod`.

## Dependencies tab addition

The updated skill's collector queried all 12 direct module declarations through
`go list -m -u -json`, with read-only module settings. All used/latest values were
resolved; the snapshot reports 7 newer releases and 5 current declarations. No
project dependencies were installed or upgraded. Confirmed source import paths
were checked against each module prefix, including test usage.

The snapshot is retained at `.maintenance/pending/dependencies.json` and embedded
in `data/model.js`. The rebuild regenerated the shell, artifact entry point,
renderer and styles with the Dependencies tab. Authored content, module files,
accepted baseline and the git index were unchanged by this addition.

The focused check was rerun against the updated renderer: 143 routes, no errors.
It additionally verifies all 12 dependency table rows against the snapshot and
checks that clicking the tab navigates to `#dependencies`. Four new light/dark
Chrome captures show Dependencies and the map; both dependency-table themes were
inspected. The exhaustive route smoke remains unrun.

Dependency-specific evidence is in `tmp/appedia-dependencies-refresh/`:
`collection.log`, `build.log`, `focused-smoke.mjs`, `focused-smoke.log`,
`screenshots.log` and `screenshots/`. The dependency snapshot was collected
separately rather than resetting the existing reviewed impact report.

## Evidence

Repeatable local scripts and detailed logs are in `tmp/appedia-awake-refresh/`:
`update-content.py`, `review-impact.py`, `check-model.py`, `focused-smoke.mjs`,
`prepare.log`, `build.log`, `model-check.log`, `focused-smoke.log`,
`source-hashes.json`, `screenshots.log`, `screenshots-isolated.log`,
`screenshots/` and `screenshots-isolated/`. This local evidence directory is
ignored by git. Asset hashes and verification status are retained in
`.maintenance/pending/verification.json`.

```bash
python3 tmp/appedia-awake-refresh/check-model.py
timeout 55s node tmp/appedia-awake-refresh/focused-smoke.mjs \
  --site appedia --jsdom /tmp/appedia-smoke
```

No commit was made and the git index was not changed.
