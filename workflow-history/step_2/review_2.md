# Feature 2 second review

**CHANGES REQUIRED.** Created because two correctness requirements still fail:
inventory refresh ordering and detached ownership of session artwork. No user
decision is needed. No implementation or commit was performed by this review.

The previous broad list pipeline, frontend policy rebuild, exported broad human
inventory read, concrete scoped read signatures and stamp-map baseline have been
removed. Focused per-registry outcomes, action fixes, Agent Switch rows, artwork
fallback and boundary-record cleanup are connected. The architecture suite now
passes. Those improvements do not resolve the findings below.

## 1. Order every publisher of dashboard inventory, including session polling

**Priority: high.** `internal/app/handlers.go:125` checks only dashboard identity,
view generation and session epoch for `SessionsRefreshedMsg`. That envelope has
no request sequence, while full/config results use `requestID` and
`lastAppliedRefreshRequest`. Both paths publish complete card/session inventories
to Dashboard and Agent Switch.

An older session poll can finish after a newer full/config refresh in the same
dashboard, view and epoch, then replace the newer rows, selection/preview metadata
and action facts with old data. Process status and external configuration changes
do not necessarily advance the UI mutation epoch. Conversely, a delayed full
refresh can replace a more recent sessions-only result because the latter never
participates in the full-refresh sequence. Two overlapping session polls have the
same problem. Existing epoch guards protect local mutation boundaries, not read
completion order.

Reproduced with a temporary Go overlay: accept full inventory request 10 containing
session `one` named `new`; deliver a previously captured sessions-only result with
the same scope/epoch containing that session named `old`. The displayed name
becomes `old`. `TestReviewOlderSessionRefreshOverwritesNewFull` fails with:

```text
older sessions result overwrote accepted newer full inventory: old
```

**Required fix:** attach a captured monotonic inventory request identity to every
inventory-producing command/envelope and reject obsolete inventory publications
across both paths before applying rows, errors or preview state. Keep registry
certification/baseline ordering distinct where necessary: accepting a newer
sessions-only inventory must not incorrectly discard valid independent registry
results or certify a token. Preserve the existing unchanged-token protections.

**Required tests:** control read completion order with fakes and drive coordinator
`Update`. Cover older session poll after newer full refresh, older full refresh
after newer session poll, and two session polls completing in reverse order.
Assert displayed names/status/action facts and selected identity/preview cannot
regress; retain dashboard/view/epoch rejection tests and token-ordering tests.

## 2. Deep-copy the new nested session artwork at public boundaries

**Priority: medium; blocks the explicit copy-isolation acceptance criterion.**
`models.SessionRow` now contains `PuppetLines []string`, but
`internal/ui/components/session_list.go:78–105` still treats rows as flat values:

- `SetRows` copies only the outer slice, retaining caller-owned artwork buffers.
- `Sessions` copies only the outer slice, exposing internal artwork buffers.
- `Selected` returns a row whose artwork slice still aliases list state.

`Dashboard.Sessions` delegates to that getter, extending the exposure outside the
component. Mutating the caller's rows after SetRows, or a row returned by Sessions,
changes future published rows. This also permits concurrent callers to race the
frontend despite the intended immutable publication boundary.

Both temporary overlay ownership regressions fail:

```text
TestReviewSessionArtworkOwnership: SetRows aliases caller artwork
TestReviewSessionGetterArtworkOwnership: Sessions exposes internal artwork
```

**Required fix:** clone nested session artwork on publication and retrieval,
including Selected and any returned list items/rows whose contract permits caller
mutation. Reuse a small cloning helper rather than repeating shallow append calls.
Audit the other newly migrated setters: `AgentSwitchView.SetRows` also shallow-
copies CardRows containing ID/name/artwork slices. Either strip fields the switch
view does not need or detach them at that boundary. Preserve the already-correct
CardList copy behavior and owned Dashboard inventory/choice caches.

**Required tests:** mutate both setter inputs and getter outputs, including
Selected, and verify subsequent rows/render inputs remain unchanged. Include
custom artwork supplied through the real mapping, not only scalar row fields.
Do not use an undocumented immutable-by-convention claim while mutable slices
remain publicly returned.

## 3. Supply the missing refresh failure/certification evidence

The updated code contains before/after token checks and independent choice
outcomes, but the added repository tests do not demonstrate several required
scenarios from `impl_plan.md` and `review.md`. The cadence tests mainly cover
normal reload/unchanged behavior and synthetic request ordering; the API choice
test verifies copied artwork, and the scoped-error tests exercise inventory
failure/empty results. Passing these does not establish all new failure branches.

Add focused deterministic regressions for token-read failure before/after loading,
different before/after tokens, and independent skill/persona/puppet failure versus
successful empty/recovery. Verify useful independent data survives, errors surface,
the baseline remains invalid and a later poll retries with unchanged metadata.
Exercise first inventory success with failed choices through coordinator Update,
not only the view setter. These tests should use narrow read-capability fakes and
must not depend on native model calls. Record the actual cases/results in the
feature review artifact instead of declaring the entire checklist complete.

## Verification performed by this reviewer

- Focused API/app/components/views/models plus architecture packages with
  `-count=1`: **PASS**; models currently reports no test files.
- `go test -race ./internal/app -run 'TestCommandClosuresDoNotRaceCoordinator|Dashboard|Refresh' -count=1`:
  **PASS**.
- `go test -race ./internal/api -count=1`: **PASS**.
- `go build ./...`, `go vet ./...`, `go test ./...`: **PASS**.
- `git diff --check`: **PASS** before authoring this artifact.
- Full `golangci-lint run`: **FAIL**, 69 findings. Independently rerunning lint in
  the executor's `/tmp/feature2-lint-baseline` also reports 69. Existing finding
  categories remain; they are not literally identical output: checker complexity
  increased from 304 to 309 and shadowing references moved. This is documented
  lint debt, not the reason for this rejection. Logs are in
  `/tmp/kankuro-review2-lint.log` and `/tmp/kankuro-review2-baseline-lint.log`.
- Three additional review-only overlay regressions: **FAIL** as described in
  findings 1 and 2. Overlay files live in `/tmp`; no repository Go files were
  edited. Reproduce against this checkout with:

```bash
go test -overlay /tmp/kankuro-review2-overlay.json ./internal/app ./internal/ui/components -run '^TestReview' -count=1
```

The overlay adds tests to the current existing test files in memory; convert the
scenarios into maintained repository tests when fixing the feature. The temporary
paths are diagnostic evidence, not durable project dependencies.

The isolated terminal smoke remains executor-reported; this reviewer did not rerun
it. Model-driven E2E remains outside the authorization for this review. After the
fixes, rerun focused regressions, architecture and relevant race checks, then the
integration build/vet/test checks and diff validation from `review.md`. Inspect
remaining exceptions and update `feature_2_review.md` with accurate evidence.
Feature 2 remains unapproved until the assigned final reviewer records PASS.
