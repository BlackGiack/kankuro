# Revised step 1 review — CHANGES REQUIRED

`review_2.md` was created because three enforcement gaps remain. Reviewed the
revised checker, allowlist, fixtures and documentation against `feature_1.md`,
`impl_plan.md` and `review.md`. No implementation files were edited or committed.

## Improvements verified

The revised checker accepts independent contract types at the real module path,
standard-library imports, local cycles, contract subpackages and internal/external
test files. Direct new fields/interfaces reusing an allowed backend selector are
rejected. Backend-to-coordinator and arbitrary/cyclic shared-helper dependencies
are checked, with read/parse failures propagated. The migration/test inventory is
populated and the previously identified message fields have explicit owners.
The new regression tests for these corrections pass.

## Remaining findings

1. **[P1] New indirect frontend type leaks still bypass exact exceptions.**
   `internal/architecture/boundary_test.go:728` records package-qualified selector
   occurrences, but never resolves the frontend type graph. Given an existing
   excepted declaration `type Legacy = session.Snapshot`, adding these declarations
   leaves both selector and declaration-edge sets unchanged:

   ```go
   type New struct{ Value Legacy }
   type Port interface{ Read() Legacy }
   ```

   Whole-repository `inspect` returns no findings. The new exported model and
   consumer port therefore expose a backend type without a new exception. This
   is the indirect alias/signature leakage explicitly covered by feature 1 and
   the plan, and is the remaining part of prior finding 3. Resolve types reachable
   from frontend declarations/consumer signatures, including local aliases and
   intermediate models, and compare exact exposure edges with the baseline.
   Add whole-inspection regression fixtures for these examples and an imported
   frontend alias; do not fix this by adding exceptions for every future alias use.

2. **[P2] Method exception keys omit receiver identity and merge distinct methods.**
   `internal/architecture/boundary_test.go:769` finds the enclosing `FuncDecl`, but
   its keys at lines 853–855 use only `funcDecl.Name.Name`. For an excepted
   `func (One) Snapshot() session.Snapshot`, adding
   `func (Two) Snapshot() session.Snapshot` with the same body does not add any
   selector or edge key. `inspect` accepts the new backend exposure because the
   two methods collapse to the same result/body strings. Include normalized
   receiver identity in method declaration/signature/body keys and in the reviewed
   inventory; preserve distinct member/signature positions where necessary.
   Add a regression that adds a same-named method on another receiver in the same
   excepted file and requires a new violation.

3. **[P2] Variant deduplication and caching ignore imported contract variants.**
   `internal/architecture/boundary_test.go:1154` deduplicates contexts using only
   the current package's source filenames. The importer cache at line 1104 uses
   the same incomplete identity. A package with platform-neutral source can
   resolve differently when an imported contract package selects different
   platform/tag files. For example:

   ```go
   // contracts/base.go
   package contracts
   import detail "github.com/BlackGiack/kankuro/internal/api/contracts/detail"
   type Result = detail.LinuxOnly

   // contracts/detail/detail_linux.go
   package detail
   type LinuxOnly struct{ Name string }

   // contracts/detail/detail_windows.go
   package detail
   type WindowsOnly struct{ Name string }
   ```

   On the current Linux host, `contractFindings` accepts the complete tree,
   although the parent contract cannot resolve `detail.LinuxOnly` on Windows.
   Checking the child independently does not validate its consumers under that
   context. Keep contexts distinct when the dependency graph can change and key
   importer caches by the effective context/dependency variants. Discover custom
   tags across dependent contract packages as well. Add a parent/child variant
   regression, including an unchanged parent file set and a changed child export.
   This is required to support the documented all-variant type-resolution claim.

## Independent verification

- `go test ./internal/architecture/... -count=1` — PASS.
- `go test ./internal/app/... ./internal/ui/... ./internal/api/... -count=1` — PASS.
- `go test -race ./internal/app -run '^TestCommandClosuresDoNotRaceCoordinator$' -count=1`
  — PASS; the selected test executed.
- `go test ./...` — PASS (cached results); native model-driven E2E not invoked.
- `go build ./...`, `go vet ./...`, `git diff --check` — PASS.

Isolated probes copied the current checker to
`/tmp/kankuro-step1-rereview-7wo04jv7`, without modifying repository source. Its
existing correction tests passed. Three additional assertions reproduced the
remaining gaps and failed because invalid changes were accepted:
`TestReviewIndirectAlias`, `TestReviewDistinctReceiver`, and
`TestReviewImportedVariant`. Reproduce there with
`go test -run '^TestReview' -count=1 -v`.

The existing repository suite passing does not establish the missing rejection
behavior. Feature 1 remains at its review gate pending these fixes and regression
evidence; the revised `feature_1_review.md` PASS overstates enforcement completeness.
There are no product questions or external blockers. Only `review_2.md` was created
in the repository during this review; no commit was made.
