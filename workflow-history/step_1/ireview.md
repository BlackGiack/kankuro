# Independent review of feature 1

Decision: **CHANGES REQUIRED — not ready to advance to feature 2.**

Reviewed against `feature_1.md` and the roadmap on 2026-11-14. Did not read
`impl_plan.md` or the other review artifacts. Created this file because changes
are required. No implementation files changed and no commit was made.

## Findings

1. **[P2] Include package variable types in indirect frontend exposure checks.**
   `internal/architecture/boundary_test.go:1285` processes only TYPE declarations
   and function signatures. It skips VAR declarations entirely. An existing,
   exactly allowlisted `models.Legacy = session.Snapshot` therefore permits a new
   frontend file to declare `var Exported models.Legacy` without any finding.
   The new consumer imports only the frontend models package, so direct-import
   enforcement cannot catch it. I reproduced this through the complete `inspect`
   function with an exact exception for the existing alias; it returned no
   findings and no error. This contradicts the requirement that new backend model
   exposures fail immediately. Walk explicit package variable types through the
   same graph and add an end-to-end rejection regression. Preserve the rule that
   the legacy alias exception does not authorize new consumers.

2. **[P2] Inspect contract variants with cgo both enabled and disabled.**
   `internal/architecture/boundary_test.go:2013` treats `cgo` as a known tag,
   excluding it from custom-tag assignments, while contexts constructed at
   `:2039` retain the host's `CgoEnabled` value. On this cgo-enabled host, a contract
   file containing `//go:build !cgo` and `type Unknown MissingType` is never
   type-checked: `inspect` returns no findings and no error. The inverse is missed
   on a cgo-disabled host. Reading all file imports does not validate these type
   graphs. Cover both cgo states in compatible variant selection/cache keys and
   add regressions for inactive cgo/non-cgo contracts, including parent contracts
   whose dependency changes with that state. Invalid or unresolved inactive
   contracts must fail rather than silently pass.

3. **[P2] Exempt neutral contracts consistently during helper traversal.**
   `internal/architecture/boundary_test.go:1702` calls
   `forbiddenFrontendImport(dep)` before excluding contracts. That predicate treats
   every `internal/api` descendant, including `internal/api/contracts`, as
   forbidden; direct frontend imports correctly exempt contracts elsewhere.
   Consequently a permitted `ui/views -> internal/helper -> api/contracts`
   dependency is reported as a transitive backend implementation dependency, even
   when the helper merely aliases an independent `contracts.Name string`.
   I reproduced that exact false positive using `inspect`. Apply the contracts
   exemption before the forbidden-backend test and add a clean helper-to-contract
   fixture alongside the existing helper-to-session rejection fixture. Continue
   enforcing the contracts package's own independence checks.

## Verification

Passed independently on the reviewed checkout:

- `go test ./internal/architecture/... -count=1`
- `go build ./...`
- `go test ./... -count=1` (all packages completed successfully)
- `go vet ./...`
- `git diff --check`
- `gofmt -l` on the new Go checker and package documentation files (no output)

Additional probes ran against a temporary copy of the checker outside the
workspace. All three expectation tests failed, reproducing the findings above;
the variable exposure was also confirmed through a separate end-to-end probe
using an exact legacy exception. Temporary fixtures were removed. These failures
are absent from the checked-in regression suite; the ordinary suite remains green.

The new contracts/models packages contain package documentation only, and tracked
application/runtime source has no diff. No speculative runtime API, schema change,
preview redesign or application behavior change was introduced. The boundary
document and exact exception inventories provide the intended migration baseline,
but enforcement needs the corrections above to satisfy feature 1's review gate.

Race tests and manual/native-agent terminal checks were not run: the reviewed
change adds a synchronous checker and documentation without changing application
concurrency or terminal code. No native-model runs were performed.

## Readiness and blockers

`ireview.md` **was created**. Advancement is blocked by findings 1–3, not by an
environment or authorization problem. No question for the user is needed. Re-run
the architecture suite with regression coverage for these cases and repeat the
final review before advancing. Existing review artifacts remain unchanged.
