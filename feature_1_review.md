# Feature 1 review

**Status: PASS after addressing `review.md`, `review_2.md`, `review_3.md`, `review_4.md`, `review_5.md` and `ireview.md`.** The boundary checker preserves exact import and declaration exceptions, resolves indirect frontend type and method-set exposure, distinguishes receiver-specific methods, handles lexical type-name shadowing, and checks contract consumers across imported build variants. Later features still own removal of the explicitly reviewed legacy surfaces.

## Planner findings addressed

The six findings from `review.md` and three findings from `review_2.md` remain addressed as documented below. The three findings from `review_3.md`, the findings from `review_4.md` and `review_5.md`, and the findings from `ireview.md` are also closed:

1. **Indirect frontend type leakage:** `frontendTypeGraphFindings` resolves local aliases, intermediate types, and imported frontend aliases while scanning model declarations and consumer function signatures. A separate `type_exposure_allowlist.json` records each reviewed legacy use by file, declaration, line and column, backend type, owner and reason. New uses fail unless their exact exposure is reviewed. `TestInspectRejectsIndirectFrontendTypeExposure` runs `inspect` on a fixture with the allowed `Legacy` alias and rejects both a new model field and an interface consumer signature through an imported frontend package.
2. **Receiver-specific method edges:** method keys include the normalized receiver for declaration, signature and body edges. `TestSameNamedMethodUsesDistinctReceiverEdge` confirms a second `Snapshot` method on another receiver adds a finding.
3. **Imported contract variants:** candidate custom build tags are collected across all contract packages. Effective variant identities and importer caches include the selected source files of transitive contract dependencies. Parent packages are checked again when a child selects different files. `TestContractParentRecheckedWhenChildVariantChanges` keeps the parent source set unchanged and confirms its Linux-only child type reference fails for Windows.
4. **Named type method sets:** frontend alias resolution follows value and pointer receiver method signatures for named types, so embedded/promoted methods and imported frontend types retain their backend exposure. `TestInspectTracksNamedTypeMethodSets` verifies that embedding `Holder`, returning it from an interface, and embedding it through an imported package are rejected.
5. **Lexical type references:** the graph walk now visits type expressions rather than every identifier in a declaration, tracks generic and receiver type parameters, and ignores field names and shadowed type parameters. `TestInspectIgnoresFieldNamesAndTypeParameterShadowing` verifies a `Legacy` field name and `Generic[Legacy]` remain clean without exceptions. The transitive baseline was regenerated after removing false-positive entries.
6. **Removal ownership:** every transitive type exposure now has an owner in features 2–7, and the checker rejects missing/invalid owners and duplicate records. Owners align with the feature inventory in `docs/frontend-boundary.md`; regression coverage checks invalid ownership and duplicate entries.
7. **Variadics and generic constraints:** the frontend type resolver follows variadic element types and type-parameter constraints on named declarations, function declarations and nested function types. It traverses union and approximation expressions and fails closed on unhandled type-expression forms. `TestInspectFindsVariadicAndGenericConstraintLeaks` checks interface variadics, generic type and function constraints, unions, approximation terms, and local/imported aliases through whole `inspect`; its positive case accepts independent variadics and unconstrained generics.
8. **Pointer generic receiver scopes:** receiver type parameter collection unwraps pointer and generic receiver syntax and is shared by top-level signature inspection and named-type method-set traversal. `TestPointerGenericReceiverTypeParamsRespectShadowing` accepts one and multiple receiver parameters that shadow backend aliases, while rejecting a method that returns the package-level backend alias.
9. **Package variables, cgo variants and neutral contracts:** package-level variable types now pass through the indirect exposure graph. Contract contexts check and cache cgo-enabled and cgo-disabled graphs separately, including parent packages against cgo-dependent child variants. Shared-helper traversal exempts neutral contract packages before backend classification, while `contractFindings` continues enforcing their independence. Regressions cover an exported variable using an allowed alias, invalid inactive cgo/non-cgo declarations, a parent referencing a type present in only one child cgo variant, and a clean helper-to-contract dependency.

Prior findings remain addressed:

- Contract packages and descendants are treated as independent contract types. Their exported declarations are traversed through local types, private fields, aliases, embeds, methods, generic arguments and cycles.
- Standard-library imports use `importer.Default`; sibling contract imports use source loading. Unknown referenced types and import errors fail.
- Production exceptions bind rule, path, import, package selectors, declaration/member/signature edges and a single owner. New fields or interface methods cannot reuse an already allowed selector to bypass review.
- Backend-to-frontend checks cover `internal/app` and `internal/ui`; reachable shared internal helpers are traversed transitively with cycle detection and propagated file errors.
- `_test.go` files do not consume production exceptions or enter contract build sets. Test-support imports are explicitly classified and cannot have production importers.
- `docs/frontend-boundary.md` contains the import, migration-surface, backend-typed message, and test-import inventories.

## Verification

- `go test ./internal/architecture/... -count=1` — PASS, including the `review_2.md` through `review_5.md` regressions and the `ireview.md` cases.
- `go test ./internal/app/... ./internal/ui/... ./internal/api/... -count=1` — PASS.
- `go test -race ./internal/app -run '^TestCommandClosuresDoNotRaceCoordinator$' -count=1` — PASS.
- `go test ./...` — PASS.
- `go vet ./...` — PASS.
- `go build ./...` — PASS.
- `git diff --check` — PASS.

No runtime, storage, MCP wire, process lifecycle or user-facing behavior changed. No commit was made.
