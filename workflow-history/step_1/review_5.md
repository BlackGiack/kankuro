# Step 1 fifth review — CHANGES REQUIRED

Created `review_5.md` because one scope-resolution defect remains. Reviewed the
revised checker, fixtures and exception inventory against `feature_1.md`,
`impl_plan.md` and `review_4.md`. No implementation files were edited or committed.

## Prior fixes verified

All seven prior independent probes now pass, including variadic alias parameters
and generic constraints. The added repository regressions cover generic function
constraints, unions, approximation terms, local/imported aliases and independent
positive cases. The earlier method-set, alias, receiver-identity and contract
variant fixes remain covered and passing.

## Remaining finding

**[P2] Generic pointer receiver parameters are not recognized as lexical bindings.**

`internal/architecture/boundary_test.go:1301` collects receiver type parameters
only when the receiver expression is directly an `IndexExpr` or `IndexListExpr`.
For a pointer receiver, its outer expression is `StarExpr`, so the helper records
no parameter bindings. The frontend resolver then looks up that parameter name
as a package-level type instead.

With the existing explicitly excepted alias `type Legacy = session.Snapshot`,
this valid, independent generic declaration is rejected:

```go
type Box[T any] struct{}
func (*Box[Legacy]) Get() Legacy {
    var zero Legacy
    return zero
}
```

Here the receiver declares its own type parameter named Legacy, which shadows
the package alias. No backend snapshot appears in the method's actual signature.
Whole-repository `inspect` nevertheless reports
`func Get ... indirect backend type exposure .../session.Snapshot`.

Unwrap pointer receiver expressions before collecting type arguments as lexical
bindings, and use the same normalized receiver handling for both the method-set
resolver and top-level method inspection. Add positive whole-inspection fixtures
for pointer receivers with one and multiple type parameters shadowing backend
aliases, along with negative cases whose method signatures genuinely expose a
backend alias. Do not add exceptions for independent generic methods. This is
part of the plan's explicit name-resolution requirement and the review-3 lexical
scope finding, whose current tests cover ordinary generic type declarations.

## Independent verification

- `go test ./internal/architecture/... -count=1` — PASS.
- `go test ./internal/app/... ./internal/ui/... ./internal/api/... -count=1` — PASS.
- `go test -race ./internal/app -run '^TestCommandClosuresDoNotRaceCoordinator$' -count=1`
  — PASS; selected test executed.
- `go test ./...` — PASS (cached package results).
- `go build ./...`, `go vet ./...`, `git diff --check` — PASS.
- Native model-driven E2E was not invoked.

Copied the checker into `/tmp/kankuro-step1-review5-ezryx_vw` for isolated probes.
All seven prior `TestReview*` probes pass. The new
`TestReviewGenericPointerReceiverShadow` fails with the false positive described
above. Reproduce there with `go test -run '^TestReview' -count=1 -v`. Repository
source and checked-in tests were not changed for the probes.

Feature 1 remains at its review gate pending the correction and regression
evidence. There are no product questions or external blockers. Only `review_5.md`
was created in the repository for this review; no commit was made.
