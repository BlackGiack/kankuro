# Step 1 third review — CHANGES REQUIRED

Created `review_3.md` because further changes are needed. Reviewed the latest
checker, exception files, regression tests and documentation against `feature_1.md`,
`impl_plan.md` and `review_2.md`. No implementation files were changed or committed.

## Prior fixes verified

The three independent probes from review 2 now pass: direct and imported local
aliases expose new fields/ports, methods on different receivers have distinct
direct-exception keys, and a parent contract is rechecked when its imported
contract package changes platform files. The repository's matching regression
tests also pass. The remaining findings concern the new frontend type resolver
and its exception inventory.

## Findings

1. **[P1] Indirect exposure through named types' methods is still accepted.**
   `internal/architecture/boundary_test.go:926` resolves a named type by inspecting
   only its `TypeSpec.Type`; it never follows the methods belonging to that type.
   With this existing, explicitly excepted frontend declaration:

   ```go
   type Holder struct{}
   func (Holder) Hidden() session.Snapshot { return session.Snapshot{} }
   ```

   Adding either of these new exposures produces no finding from `inspect`:

   ```go
   type New struct{ Holder }
   type Consumer interface{ Read() Holder }
   ```

   Consumers can now reach a backend snapshot via the embedded/promoted method or
   the returned Holder value. The original direct selector and method exception
   are unchanged, so they cannot catch this. Include named types' method signatures
   in frontend graph traversal, covering value and pointer receiver method sets,
   embedding and imported frontend types. Add whole-inspection rejection fixtures
   for both new exposures. This remains part of feature 1's requirement to reject
   indirect model/signature leakage rather than merely forbidden imports.

2. **[P2] Identifier matching falsely rejects independent fields and shadowed types.**
   `internal/architecture/boundary_test.go:954` walks every identifier in a
   declaration and sends it to package-level type lookup (line 879), without
   checking its syntactic role or lexical binding. Given an allowed package alias
   `type Legacy = session.Snapshot`, these independent declarations are rejected:

   ```go
   type Clean struct{ Legacy string }
   type Generic[Legacy any] struct{ Value Legacy }
   ```

   The first Legacy is a string field's name; the second is a type parameter that
   shadows the package alias. Neither exposes a backend type. The probe reports
   both as `indirect backend type exposure .../session.Snapshot`. Resolve type
   references with lexical scopes and distinguish names from type expressions;
   the plan explicitly requires correct shadowing/name resolution. Add positive
   fixtures for these cases and require zero findings without adding exceptions
   for clean declarations. Recheck the current indirect baseline after fixing
   resolution so false positives are not preserved as legacy exceptions.

3. **[P2] Every new type-exposure exception has the wrong removal owner.**
   All 84 entries in `internal/architecture/type_exposure_allowlist.json` specify
   `"owner": "1"`. Feature 1 is establishing the baseline; features 2–7 own the
   listed migrations, with feature 8 reserved for the final audit. For example,
   `agent_switch.go` surfaces belong to feature 4 and `card_editor.go` surfaces to
   feature 3, as documented in the boundary inventory. The generic instruction to
   consult another document does not assign the required per-edge removal owner.
   `readTypeExposureAllowlist` (line 799) checks only that Owner is nonempty, so
   these invalid owners pass. Assign each real exposure its migration feature,
   reconcile it with the direct exception inventory, and validate allowed owner
   values consistently with the existing exception reader. Add rejection coverage
   for invalid owners and preserve unused/duplicate-exception checks.

## Independent verification

- `go test ./internal/architecture/... -count=1` — PASS.
- `go test ./internal/app/... ./internal/ui/... ./internal/api/... -count=1` — PASS.
- `go test -race ./internal/app -run '^TestCommandClosuresDoNotRaceCoordinator$' -count=1`
  — PASS; selected test executed.
- `go test ./...` — PASS (cached package results).
- `go build ./...`, `go vet ./...`, `git diff --check` — PASS.
- Native model-driven E2E was not invoked.

Copied the current checker into `/tmp/kankuro-step1-review3-nr9h3tn9` for isolated
probes. `TestReviewIndirectAlias`, `TestReviewDistinctReceiver` and
`TestReviewImportedVariant` pass. `TestReviewMethodSetLeak` and
`TestReviewFieldNameIsNotType` fail with the acceptance/rejection defects above.
Reproduce there with `go test -run '^TestReview' -count=1 -v`. The owner inventory
was verified by parsing the JSON: 84 entries, sole owner value `1`.

Feature 1 remains at its review gate until these issues are corrected and verified.
No product questions or external blockers were identified. Only `review_3.md` was
created in the repository for this review; no commit was made.
