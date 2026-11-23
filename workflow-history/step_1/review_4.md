# Step 1 fourth review — CHANGES REQUIRED

Created `review_4.md` because one enforcement gap remains. Reviewed the revised
checker, exceptions, fixtures and documentation against `feature_1.md`,
`impl_plan.md` and `review_3.md`. No implementation files were changed or committed.

## Prior findings verified

The review-3 probes now pass: method-set exposure is rejected, clean field names
and shadowing type parameters are accepted, and all 80 current indirect exposure
exceptions have valid removal owners in features 2–7. Invalid-owner and duplicate
entry regression tests are present. The earlier alias, receiver and imported
variant probes continue to pass.

## Remaining finding

**[P1] Frontend type traversal skips variadic parameter types and generic constraints.**

`internal/architecture/boundary_test.go:916` implements its own AST type resolver
and reporter. Neither traversal handles `*ast.Ellipsis`; the declaration loop
reports `TypeSpec.Type` but only collects parameter names from `TypeSpec.TypeParams`,
without inspecting their constraints. The function-type traversal similarly
collects type-parameter names without traversing constraints.

Given an existing, explicitly excepted alias:

```go
type Legacy = session.Snapshot
```

Either new declaration passes whole-repository `inspect` with zero findings and
without changing the baseline exception:

```go
type Consumer interface{ Read(...Legacy) }
type Constrained[T Legacy] struct{}
```

The first exposes a backend snapshot as a consumer method's variadic parameter.
The second introduces backend coupling in an exported type's constraint. These
are valid Go type/signature forms covered by the feature's indirect leakage
requirement and the plan's explicit generic-constraint requirement. No new
package-qualified selector is added, so direct exception tracking cannot help.

Handle variadic element types and traverse every type-parameter constraint in
both type and function declarations. Cover constraint unions and approximation
terms when completing the traversal, and ensure unsupported type forms cannot
silently return a clean result. A semantic type graph is preferable where it
avoids maintaining incomplete parallel AST walks; any solution must retain the
required build-variant and lexical-scope behavior.

Add whole-inspection rejection fixtures for both examples, a generic function
constraint, and constraints expressed through local/imported aliases. Include
positive fixtures for variadic independent values and unconstrained generics.
Preserve the passing prior probes and exact, owner-assigned exception behavior.

## Independent verification

- `go test ./internal/architecture/... -count=1` — PASS.
- `go test ./internal/app/... ./internal/ui/... ./internal/api/... -count=1` — PASS.
- `go test -race ./internal/app -run '^TestCommandClosuresDoNotRaceCoordinator$' -count=1`
  — PASS; selected test executed.
- `go test ./...` — PASS (cached package results).
- `go build ./...`, `go vet ./...`, `git diff --check` — PASS.
- Native model-driven E2E was not invoked.

Copied the checker into `/tmp/kankuro-step1-review4-etst5052` for isolated probes.
The five prior probes pass. `TestReviewVariadicAlias` and
`TestReviewGenericConstraint` fail because the new backend exposures are accepted.
Reproduce there with `go test -run '^TestReview' -count=1 -v`. No repository source
or checked-in tests were changed for these probes.

Feature 1 remains at its review gate pending this correction and regression
evidence. No product questions or external blockers were identified. Only
`review_4.md` was created in the repository for this review; no commit was made.
