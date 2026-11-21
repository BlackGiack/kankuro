# Appedia verification for reasoning effort

The source model and assessments cover the feature worktree based on `110865d`.
The card editor flows show how the optional effort field is saved and reopened.
The activation flow shows argument ordering for launch, relaunch and switch.

- The ordinary Go suite and changed-line lint passed. The model-free E2E
  suite passed with the `e2e` tag. The existing launch-focus journey saves
  `high`, reopens the card and checks the effort row before launching Codex.
  The reopened card screenshot is embedded in the Tests tab.
- Driver fixtures check the Claude, Codex and Antigravity argument mappings
  and the generic fallback, including OpenCode. Native Antigravity behavior
  remains outside this verification.
- Appedia finalization passed its full jsdom route check. Chrome rendered
  the activation flow and terminal evidence in both themes. Historical
  captures, model runs and coverage retain their original provenance.

---

# Appedia verification at 5c80ec8

This refresh covers the two frontend boundary follow-up commits and the
split-index fix. The source model was extracted from a separate checkout
of `5c80ec8`. The existing dashboard and startup flow URLs remain valid.

- The center flows show retained-report notifications, independent scope
  checks, timestamp fitting, and isolated Git inspection with split indexes.
  They describe the labelled directory fallback after metadata or status
  errors. Source references now use the surviving lifecycle API.
- `go test -json ./... -count=1` passed on Linux. The Tests page links
  the named runs and records their limits. The refreshed regressions cover
  ordinary and linked split-index worktrees, Git helper isolation, metadata
  fallback, report creation and eviction, scope loss, and timestamp fitting.
- Prior terminal captures, native-model executions and statement coverage
  retain their earlier endpoints. This refresh ran the ordinary Go suite
  and did not rerun native-model or terminal journeys.
- The validator and full jsdom route check passed during finalization.
  The impact review is in `.maintenance/last-review.json`.
  Chrome rendered the map, both center flows and the Tests page in light
  and dark themes. These captures check the Appedia itself.

---

# Appedia verification at bd03108

This refresh covers the frontend boundary roadmap through the restored
session header and the startup puppet-sync fix. The source candidate came
from a clean checkout of `bd03108`.

- The source and diagram refresh relinks the retired coordinator package to
  `internal/ui`, shows the terminal attachment boundary and deletion outcomes,
  and preserves incoming flow URLs. The startup flow shows the puppet scan
  completing before the first registry read. The session-center flows show
  the header above reports and workspace information.
- `go test -json ./... -count=1` passed. The Tests page links named runs and
  gives every flow at least one linked scenario with scoped evidence.
- The project-owned model-free Codex directory and Git journeys, launch
  focus, and MCP manager passed through tui-test. The current dashboard
  capture shows the header and workspace body at 120×40. A separate capture
  shows the narrow header at 80×24. The previous launch-focus capture remains
  available as historical evidence.
- A temporary tui-test overlay exercised the real startup binary in two
  isolated Linux configurations. Custom artwork was available on first
  dashboard open. A missing source produced the warning and the dashboard
  still opened. Manual puppet refresh picked up an edited source. The
  project test also certifies the initial read and detects any startup puppet
  sync started by the coordinator.
- The prior native-provider runs and statement coverage retain their original
  endpoint. This refresh ran no model prompts. Live native recall, non-Linux
  terminal behavior, and combined live resize, relaunch and all-failure
  recovery remain open.

## Site checks

The refresh used the accepted extraction exclusions. The impact review is
in `.maintenance/last-review.json`. Chrome captured the map, startup and
report flows, and test evidence in both themes. The startup sync step, the
terminal package and the current launch-focus image were inspected in the
rendered site.

The full jsdom route check passed during finalization, and the pipeline accepted
the model and source baseline.

---

# Appedia step 6 refresh verification

Scope: documentation/model refresh against the uncommitted roadmap step-6 worktree,
2026-11-16. This is not the independent final feature review and grants no approval
for advancing or committing implementation.

## Content and provenance

- Re-extracted current Go sources using the accepted baseline's exclusions.
  Added package roles/map nodes for independent `api/contracts` and `ui/models`.
- Replaced the terminal-preview flow with selected-session authored reports.
  Added workspace inspection and recovery/resource-release flows, with real code
  links for every executing step. Preserved the old flow URL for incoming links.
- Removed references to deleted preview code and repaired older steps 1–5 links
  to migrated card, inventory, lifecycle and configuration operations. Backend
  policy and human presentation responsibilities remain distinct.
- Added 21 source-bound center evidence records with inspected assertions and
  explicit limits, plus one unassessed live resize/relaunch/recovery journey.
  No application tests were executed by this documentation refresh. Executor
  reports focused passes separately in `feature_6_evidence.md`.
- Two removed historical evidence records were dropped; changed historical source
  assessments were marked unassessed. Their original reviewed hashes/results are
  retained where available, so stale source/run warnings remain intentional.
  Historical images and coverage keep original provenance. No old coverage
  percentage or capture is claimed for this checkout.
- Dependency collection used Go package-manager queries, without installing or
  upgrading application dependencies. Unknown latest versions remain unknown.

## Documentation checks

The refresh pipeline validates source references, file references, flow participants,
assessment associations and generated diagrams. The initial rebuilt snapshot passed
all-route jsdom smoke: 3,459 routes, including 878 in data-flow mode, over 20 packages,
404 types, 330 functions and 65 flows, with no errors. Finalize rebuilt and passed its all-route smoke gate against the final content,
then accepted the source baseline at 2026-11-16T23:16:43.471Z. Reviewed impact
records are `.maintenance/step6-review.json` and
`.maintenance/step6-extractor-caveat-review.json`. `git diff --check` also passed.

Headless Chrome captured map, contracts package, isolated center type, center source,
three center flows and Tests page at 1600×1100 in both themes (16 images). The map,
report flow and Tests overview were inspected directly: text and nodes are readable,
no overlapping node boxes were observed, and the assessment scope is visible.
These are Appedia browser captures, not application terminal E2E evidence. Eight additional screenshots at 900×1000 cover center feature detail, source
evidence, historical capture and coverage views in both themes. Direct inspection
confirmed readable assessment limits, explicit not-run execution status and old
coverage marked as unbound to the current endpoint. The isolated center view and
workspace flow were also inspected. These checks do not prove every viewport.

Build/refresh commands are in `README.md`. Temporary build/smoke logs and screenshots
are under `/tmp/kankuro-appedia-*`; generated browser assets and raw source evidence
remain in the site. The previous MCP/launch-focus execution report is retained as
`.maintenance/historical-mcp-launch-verification.md` and does not describe this head.

## Limits

Source/model freshness and browser routing do not prove the center implementation
satisfies every feature-6 acceptance criterion. No current project-owned step-6
terminal journey, fresh statement profile or model-driven run was collected here.
The full combined failure-policy/resize/relaunch journey remains unassessed in the
Tests tab. Feature approval belongs to the separate final review gate.

## Extractor limitation observed

The Go extractor's interface inference compares method names, not signatures
(`scripts/extract-go/main.go`, pass 3). In the isolated center view it infers
`SessionCenter -> messages.DataReceiver`, but `SetData(SessionCenterData)` differs
from `SetData(DataLoadedMsg)`. That arrow is a false positive, not an application
relationship. The components package notes flag this caveat. No application or
skill-tooling code was changed to hide it; the real source signatures are linked.
A signature-aware extractor regression/fix is a separate tooling follow-up.

## Refresh after both application review fix rounds

The prior sections describe the initial step-6 snapshot. This refresh supersedes
its source/model freshness and center test-execution status. Both `review.md` and
`review_2.md` remain historical findings; no `review_3.md` is needed unless a new
required change is identified.

Re-extracted the current coordinator, Git inspection, path rendering and their
regressions. Flows now explain attempt/due state independently of cache presence,
stale workspace retention across report-only reads, first-report takeover after a
successful read, 30-second failure recovery, 250 ms inherited-pipe drain bounds,
and separate single-line path sanitization. The existing extractor inference caveat
remains documented; no skill-tooling or application code was changed here.

All six original independent review-overlay assertions now pass. Independently
passed `go test -json ./... -count=1`, the required filtered report/message/center/
attach/command-closure race suite, `go build ./...`, `go vet ./...`, and
`git diff --check`. Current test results are bound to the prepared source fingerprint
in 28 center evidence records and `.maintenance/step6-third-verification.json`.
The ordinary suite does not run E2E-tagged tests. No new native-model run, complete
live-terminal journey or statement-coverage profile is claimed. Existing broader
historical test limitations are retained. No additional application changes were
identified on this re-review.

The accepted post-fix impact dispositions are recorded in
`.maintenance/step6-postfix-review.json`. The post-fix rebuild and all-route smoke passed; baseline accepted at
2026-11-16T23:39:13.192Z. Headless Chrome produced 12 screenshots at 1600×1100
and four at 900×1000 in both themes. Direct inspection of the recovery flow,
inherited-pipe evidence and narrow workspace-freshness evidence confirmed readable
labels, current source excerpts and recorded execution provenance. Temporary
logs/screenshots are `/tmp/kankuro-step6-final-*`. Final application re-review: PASS;
no further required changes, no environment blockers and no `review_3.md` created.
No commit was made.

## Refresh for independent-review fixes

This snapshot follows the additional findings in `ireview.md` and their Executor
fixes. It supersedes earlier source freshness and center execution records, not
that review's approval status. **Independent implementation re-review is pending.**
The earlier planner PASS above describes a prior worktree and does not approve
these new changes or close the independent review.

Re-extracted the current cache/process/header code. Explanations now distinguish
navigation suspension (detached same-dashboard/session content cache plus operation
cancellation) from scope/selection invalidation, describe Unix-owned Git process
groups for cancellation/output limits, state the non-Unix direct-process fallback,
and describe width-aware time/kind headers in the last two content rows.

Read six new regression tests and refreshed all current center source hashes and
excerpt ranges. Independently passed:

```text
go test -json ./internal/api ./internal/session ./internal/app ./internal/ui/components -count=1
```

The 34 center evidence records have actual per-test pass events, commands, timestamps
and the prepared source fingerprint, retained in
`.maintenance/step6-ireview-verification.json`. These checks support the stated
assertion limits; they do not constitute the independent final code review.
No new native-model/live-terminal E2E run or statement profile was collected.
Historical broader execution/coverage/capture provenance and the extractor's
method-name-only interface caveat remain explicit.

Reference validation, rebuild and all-route smoke passed; baseline accepted at
2026-11-16T23:56:20.811Z. After activation restart, the accepted source and content
hashes still matched the checkout and all 34 center evidence hashes matched.
The all-route smoke was rerun successfully. Headless Chrome recreated 14
screenshots at 1600×1100 and four at 900×1000 in both themes. Direct inspection
of the cache diagram, recovery flow, process-cleanup evidence and narrow
header/navigation evidence found readable content and current source/results.
Reviewed impacts are in `.maintenance/step6-ireview-refresh.json`; current
logs/screenshots use `/tmp/kankuro-step6-ireview-*`. No application code or skill
tooling was changed, and no commit was made. Independent code approval remains
with the Reviewer.
