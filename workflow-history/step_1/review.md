# Step 1 review — CHANGES REQUIRED

Reviewed the added architecture checker, allowlist, fixtures, boundary document,
package documentation and existing callers against `feature_1.md` and
`impl_plan.md`. `review.md` was created because changes are required. Do not advance
to feature 2 on the existing `feature_1_review.md` PASS claim.

## Findings

1. **[P1] Valid contract declarations are classified as implementation leaks.**
   `internal/architecture/boundary_test.go:770` forbids every module-internal
   package, including `internal/api/contracts` itself. The named-type walker
   therefore rejects `package contracts; type Row struct{ Name string }` as
   `Row -> .../internal/api/contracts.Row` without inspecting its fields. The
   current doc-only package hides this failure; the typed fixtures use the
   unrelated package path `fixture/contracts`, so they also miss it. Exempt the
   legitimate contract package(s) from implementation classification, traverse
   their definitions, and add end-to-end acceptance fixtures using the real
   contract package path, including nested independent local types and cycles.

2. **[P1] The contract checker cannot import ordinary contract dependencies.**
   `internal/architecture/boundary_test.go:647` constructs `types.Config` without
   an Importer. A valid `import "time"; type Timestamp = time.Time` fails with
   `could not import time (Config.Importer not installed)`. This prevents the
   time/context-bearing contracts contemplated by the specification. Install
   suitable import resolution and test valid standard-library dependencies
   through `contractFindings`, alongside genuinely forbidden imports. Positive
   fixtures must use the same inspection path as the repository gate.

3. **[P1] Existing exceptions allow new declarations and leaks using the same type.**
   `internal/architecture/boundary_test.go:473` compares a set of imported selector
   names per file. Once `session.Snapshot` is allowed, adding another exported
   struct field, alias or port method returning `session.Snapshot` in that file
   produces the same set and passes. There is no declaration/member/signature
   inventory check for frontend leakage or consumer ports. This violates the
   requirement that new violations fail immediately and exceptions identify exact
   file/symbol/import surfaces. Track the declaring symbol and member/signature
   edge, with its removal owner, rather than just the dependency's type name.
   Add rejection fixtures that reuse an already permitted selector in a new field
   and consumer interface. Review and remove stale entries at the same granularity.

4. **[P1] Direction enforcement misses backend-to-coordinator and arbitrary helper paths.**
   `internal/architecture/boundary_test.go:531` rejects backend imports of
   `internal/ui`, but permits `internal/app`, which is explicitly frontend during
   migration. `inspect` only starts helper traversal for imports beginning with
   `internal/ascii` (line 448), and `helperLeaks` only recurses within ASCII child
   paths (line 603). A frontend importing `internal/newhelper` which imports
   `internal/session` passes with no findings. Existing ASCII helpers can also
   conceal coupling through a helper outside that subtree. Enforce backend-to-
   frontend direction for both frontend roots, and traverse the actual permitted
   shared-package dependency graph with cycle detection and explicit exceptions.
   Add whole-inspection fixtures for each bypass. Propagate traversal read/parse
   errors rather than silently skipping unknown edges.

5. **[P2] External contract tests are included in production contract type checking.**
   `internal/architecture/boundary_test.go:619` does not exclude `_test.go` files
   when collecting contract files. Adding a normal `package contracts_test` file
   makes enforcement fail with `contracts package declarations conflict:
   contracts and contracts_test`. Same-package tests can also pull test-only
   imports into production inspection. Filter test files consistently before
   type-checking production contracts; add positive fixtures for both internal
   and external tests and verify test imports cannot consume production exceptions.

6. **[P2] The promised exact migration/test inventory is incomplete.**
   `docs/frontend-boundary.md:82` introduces a test-integration table with no data
   rows, despite existing frontend test imports of repositories/managers/API.
   The production migration table describes groups and selected examples rather
   than the complete declaring-symbol/field/signature and operational-call
   inventory required by the plan. For example, `DashboardOpenRequestedMsg.Dashboard`
   carries `config.Dashboard` but is omitted from the claimed field inventory;
   `CardSavedMsg.Card` belongs to feature 3 but its file's config exception owners
   omit 3. Complete the inventory from the actual declarations/calls, assign exact
   removal owners, populate test classifications, and align it with the finer
   exception keys requested above. Replace the current PASS artifact with a review
   result supported by actual acceptance evidence after fixes.

## Additional enforcement scope to retain during the fixes

The implementation documents host-only contract type checking as a limitation,
whereas the plan calls for declaration resolution for inactive variants too.
Direct import scanning across variants is useful but does not replace the planned
type/signature checks. Add fixtures for inactive variant type graphs and ensure
unknown edges fail explicitly without merging mutually exclusive declarations.
Keep the checker usable for independent contract subpackages rather than merging
all descendant files into a single package check. Do not solve these issues by
relaxing forbidden imports or accepting unused exceptions.

## Independent verification

- `go test ./internal/architecture/...` — PASS, including an uncached run.
- `go test ./internal/app/... ./internal/ui/... ./internal/api/...` — PASS.
- `go test -race ./internal/app -run '^TestCommandClosuresDoNotRaceCoordinator$' -count=1`
  — PASS; selected test executed.
- `go test ./...` — PASS (cached package results); model-driven E2E not invoked.
- `go vet ./...` — PASS.
- `go build ./...` — PASS.
- `git diff --check` — PASS. Added untracked artifacts were inspected separately.

Six isolated regression probes against a copy of the checker all reproduced the
defects: clean contract rejected; standard import rejected; external test package
rejected; new same-selector leak accepted; arbitrary helper bypass accepted;
backend-to-current-coordinator import accepted. The probes assert intended valid
or invalid behavior and therefore fail against this implementation. Their scratch
location is `/tmp/kankuro-step1-review-p7i3mn51`; reproduce with
`go test -run '^TestReview' -count=1 -v` there. No production files or checked-in
tests were changed to run them.

No product questions or external blockers were identified. These are implementation
and verification gaps. Only this review artifact was created for this review;
no implementation edits or commits were made.
