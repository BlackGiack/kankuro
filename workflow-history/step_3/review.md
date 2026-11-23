# Step 3 review — CHANGES REQUIRED

Reviewed the working-tree implementation against `feature_3.md` and
`impl_plan.md`, based on HEAD `5403b2cf8fbe0a7044d37c448ea6337b6eb5ab78`.
The prerequisite approval is recorded in `feature_2_review.md`. This review does
not approve advancement. No implementation changes or commit were made.

## Required changes

### 1. [P1] Reconcile a successful save when the editor is temporarily on a source screen

Location: `internal/app/coordinator.go:370` (CardSavedMsg handling), together with
`internal/app/callbacks.go:175` (save callback).

A matching save completion clears `editorSaving`, but if the current view is a
source/MCP screen it only refreshes inventory. It does not reconcile the editor's
card identity, revision, loaded baseline or completed state. Source navigation
remains available during an outstanding save and intentionally preserves the
editor for return.

Reproduction: open a new card, start its create command, go to the skills source
screen before the command finishes, deliver its successful CardSavedMsg, return
to the editor and press save again. `editorCardID` is still empty and
`editorSaving` is false, so the second save calls CreateEditorCard again and
creates a duplicate card. The same route after updating an existing card retains
the pre-save revision, causing a spurious conflict on a subsequent save.

Reconcile a matching successful result even when the same editor is temporarily
offscreen, or prevent that transition while saving. Preserve the rule that an
unrelated/newer editor must never be overwritten or closed. If preserving the
draft after completion, carry forward the actual saved identity/revision and
the appropriate baseline without losing edits made after submission. Add a
coordinator regression with delayed save completion across source-screen return,
covering both create and update. The current late-save test covers a different
editor epoch and does not exercise this case.

Reviewer reproduction `TestReviewCompletedCreationOnSourceScreenCannotCreateAgain`
failed: successful create left an empty editor identity and accepted another
create after source-screen return.

### 2. [P2] Preserve picker selection identity across registry updates while open

Locations: `internal/ui/views/workspace_picker_modal.go:106` and
`internal/ui/views/puppet_selector_modal.go:56`, called by CardEditor.SetChoices
while background refresh continues.

Both setters replace options without reconciling the active single selection by
identity. Workspace refresh only normalizes the list index; puppet refresh only
clamps its cursor. Their missing-selection flags were calculated when the modal
opened and remain false after the selected option disappears.

Reproduction for either picker: load options A/B and a card selecting A; open the
picker; deliver a successful incomplete registry read containing only B; press
Enter without navigating. The draft now contains B and submits it as an intended
change. Reordering options can cause the same substitution. With empty puppet
choices, Selected can fall back to the default puppet instead of retaining the
loaded reference. This violates the feature's preservation of references absent
from incomplete reads and the plan's prohibition on silently substituting scalar
references.

Track the active selection independently from the option index. Reconcile it
after replacement, retaining unresolved selections until the user explicitly
chooses another value. Include navigation within an open modal, removal,
reordering, empty successful results and recovery in regressions; do not reset
an intentional in-modal selection to the original card value on every refresh.

Reviewer reproduction `TestReviewOpenPickersPreserveIdentityOnChoiceRefresh`
failed for both workspace and puppet: refreshing A/B to B silently replaced A
with B upon confirmation. Current missing-selection coverage refreshes the draft
without opening these pickers, so it misses this behavior.

### 3. [P2] Return the revision of the persisted representation

Location: `internal/api/card_editor.go`, CreateEditorCard and UpdateEditorCard
success paths computing `cardRevision(saved)` after repository writes.

The revision distinguishes nil from empty, but the repository's returned record
is not necessarily identical to its YAML round trip. SubWorkspaceIDs has
`yaml:"sub_workspace_ids,omitempty"`: an explicitly present empty slice is
omitted on disk and loads back as nil. The save result hashes the pre-round-trip
empty slice, while LoadCardEditor hashes the reloaded nil value. Thus a caller
using the returned revision for its next update receives a conflict even though
no intervening change happened. This also matters when fixing finding 1 by
adopting the returned revision.

Ensure returned revisions describe the actual persisted card under the existing
admission boundary and storage semantics, without a schema migration or weakening
meaningful nil/empty conflict detection. Apply the fix to create and update.
Avoid introducing a post-admission reread that can accidentally label a different
writer's record as this save's revision. Test save-result revision equality with
an unchanged fresh load, followed by a second update using the returned revision,
including explicit empty optional lists and timestamp serialization.

Reviewer reproduction
`TestReviewSaveReturnsPersistedRevisionAfterEmptyOptionalList` failed: an update
with present empty SubWorkspaceIDs returned a different revision from an immediate
LoadCardEditor of the unchanged persisted card.

## Inspected implementation and acceptance evidence

- Backend: independent editor contracts, focused reads/defaults/options, revision
  framing and digest, create/update paths, fresh-current overlay, ownership and
  sharing checks, manager card-reference admission, repository validation and
  persistence. The previous full-record editor save methods were removed; test
  adapters now call the replacement operations.
- Frontend: draft construction and cloning, capability selection ordering,
  missing-ID retention, explicit presence and hidden-section edit retention,
  reset, read-only controls, workspace/puppet modals, callback command capture,
  load/save envelopes, reload, epochs and navigation/source return.
- Boundaries/removals: editor reads are split from the remaining launch-only
  full-card adapter; CardSavedMsg no longer carries a stored card; the concrete
  editor config import and direct/transitive exceptions were removed. Remaining
  configuration, launch, preview and composition exceptions retain later-feature
  ownership. Architecture tests passed; no new broad editor exception was found.
- Existing behavioral coverage passed for uneditable-field preservation,
  unseen-field conflicts, concurrent saves, explicit empty/false edits, stored
  field digest coverage, shared owner policy, admission/cleanup, reset, failed
  reload retention and stale results. These passing cases do not cover the three
  demonstrated failures above.

The implementation diff spans API membership/editor/revision/contracts and tests;
app coordination/callbacks/navigation/refresh/ports and tests; UI editor/models,
capability lists, workspace/puppet modals and tests; boundary inventories/docs;
and the current project invariant update. `feature_3_evidence.md` was read as an
author handoff, not accepted as independent verification.

## Independent verification

These commands were executed in this review and passed:

```sh
go test ./internal/api ./internal/ui/views ./internal/app ./internal/ui/components ./internal/architecture/... -count=1
go test -race ./internal/api ./internal/app -run 'Card|CommandClosures|TestEditor' -count=1
go build ./...
go vet ./...
go test ./...
golangci-lint run --new-from-rev=HEAD
git diff --check
```

Changed-code lint reported zero issues. Full lint was not rerun; the author and
prior feature review document existing architecture-checker lint debt, which this
review does not independently recertify. Manual terminal verification was not
repeated; the author's smoke coverage remains author-reported. No model-driven
native E2E was run or authorized.

Review-only failing regressions were run through Go overlays; repository source
and checked-in tests were not modified. Reproduction sources and overlay mapping
are retained locally in `/tmp/kankuro-step3-review-e3_dhje0/`:

```sh
go test -overlay /tmp/kankuro-step3-review-e3_dhje0/overlay.json ./internal/ui/views ./internal/app -run '^TestReview' -count=1
go test -overlay /tmp/kankuro-step3-review-e3_dhje0/overlay.json ./internal/api -run '^TestReview' -count=1
```

Both commands failed on the required behaviors described above. The temporary
overlay path is supporting evidence, not a permanent test artifact; add durable
regressions when fixing the issues.

No environment or authorization blocker prevented this review. The three findings
block feature-3 approval; fixes and independent follow-up review are required.
Created `review.md` only; no feature approval artifact or commit was created.
