# Feature 2 completion instructions

**CHANGES REQUIRED.** The implementation is an incomplete vertical slice, not a
completed feature with only stale allowlists. Keep the feature-2 gate closed.
No product decision is needed to complete the agreed scope. This review changes
documentation only; it neither implements fixes nor authorizes a commit.

Reviewed `feature_2_review.md`, `feature_2.md`, `impl_plan.md`, the tracked diff,
new contracts/read adapters/models, their callers and the current boundary output.
The executor's review correctly lists major omissions, but additional behavioral
defects below must be fixed before final review. Preserve unrelated work.

## Evidence from this review

- `go test ./internal/architecture -run '^TestFrontendBoundary$' -count=1`:
  **FAIL**, reproduced. Full captured output is temporarily available at
  `/tmp/kankuro-feature2-boundary-review.log`.
- `go test ./internal/architecture/... -count=1`: **FAIL**, exit 1, with the same
  boundary violations. The checker still distinguishes stale records from new
  exposure edges.
- `go test ./internal/api ./internal/app ./internal/ui/components ./internal/ui/views ./internal/ui/models`:
  **PASS**, with cached results for the four tested packages; models has no tests.
- Other build, race, vet and lint results in `feature_2_review.md` are executor
  reports, not independently rerun here. Passing those suites does not resolve
  the missing behavior coverage or architecture failure.

## 1. Fix behavior and make focused data the sole list source

**Blocking: two runtime inventory pipelines still exist.**
`internal/app/scoped_refresh.go:21` and `:53` load broad inventories, then
`readFocusedDashboard` reads another inventory and all choices. `handlers.go`
broadcasts the broad data before calling `SetInventory`. Consequently
`Dashboard.SetData` calls `rebuildLegacyRows` in production, despite its comment
suggesting a compatibility/test role. That method duplicates backend policy,
enables foreign-card edits/deletes, treats unknown sessions as active, and searches
conversation maps in the view. It violates the intended boundary.

Remove `rebuildLegacyRows` and the broad list state/setter path. Convert remaining
tests to focused fixtures rather than preserving a production adapter for tests.
Apply one accepted focused update to the owned caches, then map once and publish
rows once. Preview or configuration delivery must never rebuild list rows.

Why this is behavioral: when a broad read succeeds but the focused overview fails,
the fallback can publish new rows even though the authoritative focused read
failed. If cached inventory exists, `SetInventory(nil, choices)` can instead restore
older rows after preview selection already changed. With no choice cache, a
successful inventory is currently withheld entirely by the joint cache guard at
`dashboard.go:370`. Session rows do not require choices and should still be shown;
card rows can use existing missing-choice fallbacks while surfacing the failure.
Neither a missing first choice read nor one failed registry may hide an otherwise
successful inventory.

Add coordinator-level regressions, not just direct `SetInventory` tests:

- Focused inventory fails after a valid display: rows, selected identity and
  relevant preview selection retain the prior successful state; no broad fallback
  replaces them.
- First inventory succeeds while choices fail: sessions and card identities remain
  usable, errors are visible and capability/artwork fallbacks are safe.
- Successful empty inventory clears rows and selection coherently. Choice-only
  rename/artwork updates refresh labels without replacing selection.

**Blocking: availability is inconsistent and not fully consumed.**

| Current evidence | Required correction and regression |
|---|---|
| `inventory_reads.go:140–141` disables both Relaunch and Switch when the old card is deleted/unavailable. `Dashboard.handleSwitch` now obeys that value. | Disable relaunch of the unavailable current card; preserve switching a stopped session to another available card. Verify a deleted-card stopped session opens the picker and successfully reaches the checked switch operation. The destination card determines availability at execution. |
| `Dashboard.handleSessionEnter` still checks the string `terminated`, then invokes relaunch or attach without checking those availability values. | Dispatch Enter using application availability and format the supplied reason when unavailable. Test unknown status with all actions disabled, and terminated/current-card-unavailable with Relaunch disabled. No callback should fire. |
| `Dashboard.handleDeleteCard` interprets every disabled Delete as retryable. A foreign shared card has Delete disabled but Retry false. | Retry only when Retry is explicitly enabled. Otherwise retain safe feedback/no effect. Assert Delete on a foreign shared card never invokes retry or opens an owner deletion flow. |
| `summarizeSession` disables Rename on stopped sessions, while the existing view/service permit rename; `handleEdit` ignores Rename. IDE/terminal handlers also need an availability audit. | Preserve existing supported rename behavior, including stopped sessions, in backend availability; consume it in key routing. Audit all keys, footer hints and callback paths for Rename/OpenIDE/OpenTerminal and other action facts. No new restriction merely to match the DTO. |
| Component key handlers still call callbacks without considering their rows' availability. | Either centralize action dispatch in the parent and make list components selection-only, or consistently guard component callbacks too. Test the routes actually used by Dashboard to avoid duplicate or bypassed actions. |

Keep execution-time checks, revision-bound confirmation, compatibility checks and
post-effect warnings intact. Denial reasons must describe the actual disabled
action; the hardcoded deleting-card error is unsuitable for every disabled Launch.
Unknown/degraded states must not acquire enabled actions in a frontend mapper.

## 2. Finish explicit independent read outcomes

`Service.ReadDisplayChoices` returns early on skill/persona/puppet failure. This
turns three independent registries into one failed bundle: a failed skill read
prevents valid persona/artwork updates. `DataLoadedMsg` still uses nil fields for
other global registries. Pointer-valued Inventory/Choices plus aggregate Err does
not meet the promised explicit per-read success/error semantics.

Introduce a clear result for each attempted read and an explicit unattempted state
on partial envelopes. Read skills, personas and puppets independently through the
application capability; retain each cache on failure, clear it on successful empty,
apply other successful caches and surface contextual errors. Legacy config/editor
messages may remain for feature 5, isolated from lists, but their delivery must
preserve the same independent retain/clear semantics. Avoid leaking config records
through the new display capability.

Test every registry failing with unchanged file metadata, successful empty after
populated, later recovery, and a sessions-only update that does not attempt global
reads. Preserve error clearing on the next user keypress and stale-result filtering.

## 3. Connect the opaque token and preserve cadence and ordering

`ReadRegistryChangeToken` exists but frontend refresh still uses `repo.Fingerprint`,
`config.FileStamp`, `fingerprintsEqual` and a map baseline. The new method alone is
not completion. Fingerprint failures are currently dropped in `scopedFull`; in
`scopedConfigIfChanged` the fingerprint error is overwritten by the focused read.
No before/after certification or new request sequencing is wired.

Replace the frontend stamp/map path with the opaque token capability. Have the
backend certify a registry load with token reads before and after it, or provide
an equally safe application read result: failures or differing tokens return no
valid new baseline while preserving independently successful data. Surface token
errors. Any required failed registry read keeps the baseline invalid so polling
retries even when metadata did not change. Do not read fresh tokens after a load
and certify data that may precede a concurrent write.

Capture request sequence, dashboard, view generation, session epoch and old token
before constructing each async command. Reject obsolete results before updating
rows, cached choices, errors, preview state or baseline. An unchanged result must
not overwrite a forced refresh baseline; an older changed read must not overwrite
a newer one. Update `hasStaleSessionSnapshot`: it currently detects session data
only through the legacy `Sessions != nil` field. A focused Inventory result also
contains sessions and needs epoch protection after the old field is removed.

Keep the two-second UI cadence and immediate mutation refresh. Session polling
must not reload global display choices every tick as the new `scopedSessions`
currently does. Cache/reload choices on the config cadence and remap as needed.
Add deterministic delayed-fake tests for overlapping reads, mutation/rename epochs,
dashboard/view switches, unchanged replies, token errors, and writes during load.

## 4. Complete caller migration and isolate permitted legacy detail

- Introduce consumer-owned narrow read interfaces returning contracts. Inject them
  from `main`/composition into refresh/list commands; remove `*api.Service` from
  `scopedFull`, `scopedSessions`, `scopedConfigIfChanged`, `readFocusedDashboard`
  and `dashboardListCommand`. Existing effect adapters may keep their separately
  documented later-feature exceptions. An interface returning `DashboardData` or
  `config.IDCard` does not solve a read boundary leak.
- Migrate `AgentSwitchView` inventory/render joins to focused rows. Its lifecycle
  and conversation effects remain feature 4; its list inventory does not. Preserve
  its search, selection and destination-driver behavior.
- Remove Cards/Sessions broad inventory delivery from `DataLoadedMsg` and
  `SessionsRefreshedMsg` once callers migrate. Delete the broad exported human
  inventory result/read, retaining private backend checked-inventory helpers used
  by the new overview and admission policy.
- Replace `dashboardCardCommand`'s dashboard-wide broad read with an explicitly
  scoped selected-card detail loader for the legacy editor/action adapter. Retain
  complete loaded cards only there to preserve hidden fields and shared read-only
  inspection (features 3/4); never round-trip list rows into a storage record.
- Isolate selected-session preview metadata/output behind a narrowly documented
  feature-6 adapter/message. Dashboard must not hold complete card/session maps
  merely to feed the preview. Keep preview subscription and initial/live output
  working; preserve attachment's shared output machinery. Do not implement the
  center replacement in this feature.
- `DashboardRow`/`DashboardRows` are currently unused: wire the presentation mapping
  into the dashboard list, or remove dead helpers and explicitly meet the planned
  presentation boundary with a reviewed equivalent. Do not keep competing unused
  models.

## 5. Preserve artwork and published data ownership

`models.CardRows` leaves `PuppetBuiltin` false when a named puppet is unresolved.
`RenderPuppetArtwork(name, false, nil)` renders empty custom art; the old
`RenderAgentPuppet` falls back to the named built-in puppet. Preserve that fallback
for stale/missing references, along with empty-name default and custom artwork.
Add rendering coverage for unknown/missing names, empty registries, choice-read
failure, built-ins and oversized/custom lines, not only registered artwork.

`CardList.FilteredCards` returns the generic Filter result with nested row slices
still aliasing internal data. Clone published rows here as for Cards/Selected, or
provide an enforceable immutable API. Test mutation of returned IDs, pill names and
artwork cannot modify subsequent rendering or unrelated caches. Retain the existing
copy-in tests and extend them through the real read-to-render path.

## 6. Repair boundary records after fixing the actual paths

The reproduced failure has three distinct causes:

1. **Removed dependencies still allowed.** Delete unused import/type records for
   card_list.go and session_list.go and retired message/callback exposures.
2. **Surviving legitimate later-feature declarations moved.** Exact transitive
   keys include line/column positions: for example settingsCallbacks moved from
   `@375:43` to `@379:43`, and conversationChoiceCallbacks from `@482:53` to
   `@486:53`. Reconcile retained card-editor/config/conversation/attachment records
   against their actual declarations and existing owners. Direct edge lists also
   include complete declaration/member signatures; prune removed edges and add
   only narrowly reviewed surviving bridges with removal ownership.
3. **Real new/unremoved feature-2 exposure.** Concrete scoped read signatures,
   broad inventory message fields, `rebuildLegacyRows`, and dashboard-wide backend
   state must be removed, not granted new exceptions or reassigned to feature 6.

The five-line contract-package exemption added to `frontendTypeGraphFindings`
matches the independent contract boundary in principle. Keep it only with a
regression proving clean contracts are accepted and a contract that indirectly
leaks a backend type still fails whole `inspect`. Never weaken the checker,
disable the test, wildcard edges or blindly regenerate the exposure baseline.
Update `docs/frontend-boundary.md` and both allowlists to describe the final exact
exceptions; resolve fingerprint ownership now, since feature 2 owns this seam.

## Completion order and verification gate

Work in buildable increments: remove dual row publication and fix action facts;
finish independent outcomes/caches; connect token/order/cadence; migrate remaining
inventory callers and narrow legacy detail; finish ownership/artwork; reconcile
the exact exception graph. Run focused regressions as each slice changes.

At integration, record actual output/results for:

```bash
go test ./internal/api ./internal/app ./internal/ui/components ./internal/ui/views ./internal/ui/models -count=1
go test ./internal/architecture/... -count=1
go test -race ./internal/app -run 'TestCommandClosuresDoNotRaceCoordinator|Dashboard|Refresh' -count=1
go test -race ./internal/api -count=1
go build ./...
go vet ./...
go test ./...
golangci-lint run
git diff --check
```

Ensure race selections include new out-of-order/epoch tests even if their names
do not match the regex. Verify list row positions and clipping, names/PID/registration/
conversation/workspace details, search/sort/selection, deleting dashboard recovery
and foreign-card behavior with behavioral tests. Perform the isolated manual
terminal checks from `impl_plan.md`; model-driven native E2E still requires the
documented provider/model/budget authorization.

The executor reports 72 full-lint findings while changed-code lint passes.
This review has not verified attribution of that backlog. Include untracked new
Go files in lint/check coverage, fix findings introduced by this work, and record
any verified pre-existing remainder with baseline evidence. Do not silently turn
off full lint or use a clean changed-line filter as proof of a clean full run.
If resolving established old findings would expand scope materially, report the
specific remainder for a separate decision; that does not block finishing the
feature-2 fixes above.

Update `feature_2_review.md` with actual changed files, evidence for each feature
task, exact surviving exceptions and checks/results. Keep **CHANGES REQUIRED**
until the assigned final reviewer inspects the completed code and records PASS.
No implementation, feature advancement or commit follows from this review alone.
