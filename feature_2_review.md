# Feature 2 implementation review

**Current status: PASS, per the independent Reviewer approval relayed by the orchestrator.** The follow-up approval supersedes the former CHANGES REQUIRED status for preview reconciliation. Approval provenance and verification limits are recorded below. Later-feature migration exceptions and existing full-lint debt remain documented below.

## Follow-up approval and commit reconciliation

The orchestrator's message `5ace4401-ac29-5020-98e4-424b0752610d` explicitly
reports that the independent Reviewer approved step 2 in Kankuro reply
`830f1ae8-0f2d-4887-9c58-c8d0e7154760`, correlated to message
`e486ec3c-53b0-56f2-99ed-3e5c01c4ca93`, then committed the approved work as
`5403b2cf8fbe0a7044d37c448ea6337b6eb5ab78` (Separate dashboard inventory reads
from frontend models). This resolves the previously pending independent review
of preview reconciliation and satisfies feature 3's feature-2 review prerequisite.

During this documentation reconciliation, Git confirmed that exact commit exists
and is the current HEAD. Its changed files include the scoped refresh adapters
and `internal/app/preview_reconciliation_internal_test.go`. The requested
`get_replies` call was attempted but returned `forbidden`: caller activation is
not authorized for dashboard control. Therefore the approval is recorded from
the orchestrator's explicit relay, not an independently retrieved report body;
no additional reviewer findings or validation results are inferred. The sender
can retrieve the original report with an authorized activation if direct report
verification is required.

No application checks were rerun and no commit was created for this documentation
reconciliation. Statements below saying no commit was created describe the earlier
work/review stages, before the Reviewer commit identified above.

## Changes completed

- Added independent dashboard, card, session, display-choice, availability, and opaque registry-token contracts and API reads. Removed the exported broad dashboard inventory result/read; the checked broad snapshot is private to API derivation. Selected-card edit/launch uses a single-card detail adapter to preserve stored fields.
- Removed broad card/session payloads from `DataLoadedMsg` and `SessionsRefreshedMsg`, the dashboard legacy row rebuild path, and the competing broad refresh/list pipeline. Dashboard and Agent Switch consume focused rows; preview receives selected presentation metadata and retains its separate output subscription path.
- Added per-read attempted/value/error results. Skills, personas, and puppets are read independently; successful caches update or clear independently, failed caches retain their prior values, and sessions-only refresh does not read global display choices. Dashboard and coordinator own detached choice and inventory slices.
- Connected registry invalidation to opaque tokens with before/after certification, surfaced read/token errors, request ordering, dashboard/generation/session-epoch checks, immediate mutation refresh, and the existing two-second cadence.
- Fixed availability behavior for deleted current cards, stopped rename, Enter action dispatch, IDE/terminal actions, switch destination selection, and foreign-card deletion retry. Unknown session states leave unspecified actions disabled.
- Migrated Agent Switch and dashboard list/presentation mapping, preserved built-in puppet fallback, cloned published filtered rows, and added a whole-inspect contract leak regression. Reconciled removed and shifted exact architecture records without adding broad exceptions.

## Verification results

- Focused packages: `go test ./internal/api ./internal/app ./internal/ui/components ./internal/ui/views ./internal/ui/models -count=1` — PASS.
- Architecture: `go test ./internal/architecture/... -count=1` — PASS.
- Race checks: `go test -race ./internal/app -run 'TestCommandClosuresDoNotRaceCoordinator|Dashboard|Refresh' -count=1` and `go test -race ./internal/api -count=1` — PASS.
- `go build ./...` — PASS.
- `go vet ./...` — PASS.
- `go test ./...` — PASS.
- `golangci-lint run --new-from-rev=HEAD` — PASS, 0 issues.
- `golangci-lint run` — reports 69 issues. Baseline evidence: `git archive HEAD` was linted in `/tmp/feature2-lint-baseline`; it reports the same 69 findings, concentrated in the existing architecture checker tests. No new full-lint findings remain.
- `git diff --check` — PASS.
- Isolated terminal smoke with `/tmp/kankuro-feature2-manual-data` — opened the dashboard list, selected the card list, displayed built-in artwork, launched the card, showed the terminated session in the session list, and kept the preview pane alive. The temporary fixture source was removed. The smoke did not exercise dashboard deletion recovery or an external mutation during the live session; automated API/coordinator coverage is the evidence for those paths.

## Review gate and excluded verification

- The assigned final reviewer inspected the implementation and exact allowlists; the final verdict below records PASS.
- Model-driven native E2E remains outside this task because the project requires separate provider/model and total-budget authorization.

No commit was created.

## Response to `ireview.md`

Restored retained-scrollback reconciliation in the existing scoped full/config
and sessions-only refresh adapters. Each command captures the selected preview
session ID when it is created. After a successful, dashboard-scoped inventory
read, the adapter confirms that identity remains in the returned session
summaries and snapshots its retained scrollback. The coordinator applies the
snapshot only after scope, generation, session epoch and inventory ordering
checks, and only while the dashboard still displays that same preview identity.
No broad session inventory or new public preview API was added; the existing
preview adapter remains the owner of this feature-6 transition.

`TestRefreshReconcilesMissedPreviewOutput` uses a real local `cat` process and
the manager's checked attachment input. It waits until output is retained,
drains the live preview subscriber channel to model a dropped chunk, confirms
the preview is missing the marker, then verifies recovery with both the regular
sessions poll and the attach-finish full refresh while selection stays unchanged.

Verification after this fix:

- Focused API/app/components/views/models and architecture suites — PASS.
- App race suite including preview reconciliation and shared inventory ordering — PASS.
- API race suite, build, vet, full tests and `git diff --check` — PASS.
- `golangci-lint run --new-from-rev=HEAD` — PASS, 0 issues.
- Full `golangci-lint run` remains at 69 existing findings in
  `internal/architecture/boundary_test.go`, matching the archived `HEAD`
  baseline in `/tmp/feature2-lint-baseline`.

No commit was created.

## Earlier assigned final reviewer verdict for `review_2.md`

**PASS at that review stage.** The reviewer checked the implementation then
present against `feature_2.md`,
`impl_plan.md`, `review.md` and `review_2.md`, including the checked API reads,
contracts, mappings, refresh callers, component ownership and exact boundary
records. The subsequent `ireview.md` finding supersedes that readiness decision.
**`review_3.md` was not created.**

The three second-review findings are closed:

- Full/config/session inventory publishers capture a shared monotonic inventory
  request identity. Both handler paths reject older inventory publications;
  independent registry results keep their separate ordering and token handling.
  `TestInventoryPublishersShareOrdering` gates real command completions through
  `Update` in all three requested orders and checks name/status/action facts,
  selection and preview identity.
- Session rows now detach nested artwork at SetRows, Sessions, Selected and list
  item construction. Agent Switch also detaches nested card data. Reviewed the
  maintained mapped-artwork and Agent Switch mutation regressions; both pass.
- Certification and independent-choice tests cover before/after token errors,
  token changes during load, retry with unchanged metadata, first inventory with
  failed choices, per-registry retain/clear behavior and successful recovery.

Independent verification in this review:

```text
go test ./internal/api ./internal/app ./internal/ui/components ./internal/ui/views ./internal/ui/models ./internal/architecture/... -count=1
  PASS (models has no test files)
go test -race ./internal/app -run 'TestCommandClosuresDoNotRaceCoordinator|Dashboard|Refresh|Inventory|FirstInventory' -count=1
  PASS, including new ordering and certification regressions
go test -race ./internal/api -count=1
  PASS
go build ./...
go vet ./...
go test ./...
git diff --check
  PASS
golangci-lint run --new-from-rev=HEAD
  PASS, 0 issues
golangci-lint run
  FAIL, 69 existing checker-test findings; log: /tmp/kankuro-review3-lint.log
```

The full-lint limitation remains the previously investigated baseline debt, not
a new feature blocker. Counts alone do not imply identical diagnostics: checker
complexity and source positions have changed, as noted in `review_2.md`.

Remaining migration exceptions retain their named owners: selected-card editor/
action detail and full-record writes (3/4), configuration/editor registries (5),
preview output/subscriptions (6), concrete attachment/deletion/composition (7).
The feature-2 list pipeline does not use broad card/session inventory or frontend
file stamps. Existing application operations continue to revalidate effects.

Manual terminal smoke remains executor-reported with its stated coverage limits;
this reviewer did not repeat it. No model-driven native E2E was run or authorized.
At that review stage no feature-2 blockers or user decisions were recorded. The
later preview reconciliation requirement is handled in the section above.

## Response to `review_2.md`

The executor addressed the three follow-up findings in the current worktree.
Approval was pending at that handoff; the assigned final reviewer verdict above
now records PASS.

- **Inventory ordering:** every scoped full, config and sessions-only publisher now
  receives an independent monotonic inventory request ID. The coordinator rejects
  stale rows, errors and preview metadata across publisher types while retaining
  separate registry request/token ordering. A gated `orderedInventoryReader`
  regression drives completion through `Coordinator.Update` for older session
  after newer full, older full after newer session, and reversed session polls.
  It verifies name, status, Stop availability, selected session identity and
  preview selection remain current.
- **Artwork ownership:** session row clone helpers detach artwork at setter and
  getter boundaries, including `Selected`; list items are rebuilt from owned rows.
  Agent Switch detaches nested card slices on `SetRows`; CardList continues using
  its existing row clones. Component coverage mutates caller input, `Sessions`,
  and `Selected` output using custom puppet lines produced by `models.SessionRows`.
  A package-internal Agent Switch regression mutates nested input IDs, names and
  artwork after publication.
- **Read failures and certification:** narrow fake-reader coordinator tests cover
  token failure before and after load, a token changing during load, invalidated
  baseline with an unchanged-metadata retry, and first successful inventory with
  all three choices failing. Per-registry skill/persona/puppet cases verify failed
  values retain prior caches while successful empty peers clear; a later successful
  empty read clears the recovered registry and certifies the stable token. The
  first-inventory case is driven through `Coordinator.Update` and verifies usable
  card/session identities alongside surfaced errors.
- **Architecture records:** exact shifted type-exposure entries were reconciled
  after the request-sequence and handler changes. `go test ./internal/architecture/...
  -count=1` passes; no broad exception was added.

Follow-up verification after the regressions and handler refactor:

- Focused API/app/components/views/models packages — PASS.
- Architecture suite — PASS.
- App race suite including inventory ordering — PASS.
- API race suite — PASS.
- `go build ./...`, `go vet ./...`, `go test ./...`, `git diff --check` — PASS.
- `golangci-lint run --new-from-rev=HEAD` — PASS, 0 issues.
- `golangci-lint run` — FAIL, 69 findings, all in `internal/architecture/boundary_test.go`.
  The same 69-finding lint baseline was previously reproduced from `git archive HEAD`
  in `/tmp/feature2-lint-baseline`; current changed-lines lint is clean. These are
  existing checker-test findings, not regressions introduced by the follow-up.

No commit was created.
