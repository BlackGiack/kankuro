package architecture

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"go/build"
	"go/build/constraint"
	"go/format"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"testing"
)

const module = "github.com/BlackGiack/kankuro/"

type exception struct {
	Path           string            `json:"path"`
	Dependency     string            `json:"dependency"`
	Rule           string            `json:"rule"`
	Symbols        []string          `json:"symbols"`
	Edges          []string          `json:"edges"`
	EdgeOwners     map[string]string `json:"edge_owners"`
	Reason         string            `json:"reason"`
	Owner          string            `json:"owner"`
	Classification string            `json:"classification,omitempty"`
}

type finding struct{ path, symbol, rule, dependency string }

func TestFrontendBoundary(t *testing.T) {
	root := filepath.Clean(filepath.Join("..", ".."))
	allow, err := readAllowlist(filepath.Join(root, "internal/architecture/allowlist.json"))
	if err != nil {
		t.Fatal(err)
	}
	findings, err := inspect(root, allow)
	if err != nil {
		t.Fatal(err)
	}
	if len(findings) != 0 {
		lines := make([]string, len(findings))
		for i, f := range findings {
			lines[i] = fmt.Sprintf("%s: %s: %s -> %s", f.path, f.rule, f.symbol, f.dependency)
		}
		t.Fatalf("frontend boundary violations (ordinary production Go files are included):\n%s", strings.Join(lines, "\n"))
	}
}

func TestBoundaryRuleFixtures(t *testing.T) {
	tests := []struct {
		name, path, source, want string
	}{
		{"clean graph", "internal/ui/views/view.go", "package views\ntype Row struct{ Label string }", ""},
		{"forbidden direct import", "internal/ui/views/view.go", "package views\nimport \"github.com/BlackGiack/kankuro/internal/session\"", "frontend imports backend implementation"},
		{"reverse backend import", "internal/api/service.go", "package api\nimport \"github.com/BlackGiack/kankuro/internal/ui/models\"", "backend imports frontend"},
		{"child imports root", "internal/ui/views/view.go", "package views\nimport \"github.com/BlackGiack/kankuro/internal/app\"", "child frontend imports root coordinator"},
		{"framework contract import", "internal/api/contracts/types.go", "package contracts\nimport tea \"charm.land/bubbletea/v2\"\ntype Event tea.Msg", "contracts import implementation or presentation"},
		{"inactive build-tag import", "internal/ui/views/other_windows.go", "//go:build windows\npackage views\nimport \"github.com/BlackGiack/kankuro/internal/driver\"", "frontend imports backend implementation"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f, err := parser.ParseFile(token.NewFileSet(), tt.path, tt.source, parser.ParseComments)
			if err != nil {
				t.Fatal(err)
			}
			var got string
			for _, im := range f.Imports {
				dep, err := strconv.Unquote(im.Path.Value)
				if err != nil {
					t.Fatal(err)
				}
				if hit := directFinding(tt.path, dep); hit != nil {
					got = hit.rule
				}
			}
			if got != tt.want {
				t.Fatalf("rule = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestCheckedInImportFixtures(t *testing.T) {
	tests := []struct{ name, path, want string }{
		{"clean", "internal/ui/views/clean.go", ""},
		{"direct", "internal/ui/views/direct_import.go", "frontend imports backend implementation"},
		{"inactive platform", "internal/ui/views/inactive_windows.go", "frontend imports backend implementation"},
		{"reverse backend", "internal/api/backend_reverse.go", "backend imports frontend"},
		{"child to root", "internal/ui/views/child_to_root.go", "child frontend imports root coordinator"},
		{"contracts graph", "internal/api/contracts/contracts_graph.go", "contracts import implementation or presentation"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join("testdata", "fixtures", filepath.Base(tt.path))
			source, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			f, err := parser.ParseFile(token.NewFileSet(), tt.path, source, parser.ParseComments)
			if err != nil {
				t.Fatal(err)
			}
			got := ""
			for _, im := range f.Imports {
				dep, _ := strconv.Unquote(im.Path.Value)
				if hit := directFinding(tt.path, dep); hit != nil {
					got = hit.rule
				}
			}
			if got != tt.want {
				t.Fatalf("rule %q, want %q", got, tt.want)
			}
		})
	}
}

func TestExactLegacyExceptionSymbols(t *testing.T) {
	parse := func(src string) (*ast.File, *ast.ImportSpec) {
		t.Helper()
		f, err := parser.ParseFile(token.NewFileSet(), "internal/ui/views/legacy.go", src, 0)
		if err != nil {
			t.Fatal(err)
		}
		return f, f.Imports[0]
	}
	dep := module + "internal/session"
	f, imp := parse("package views\nimport \"" + dep + "\"\ntype Row struct{ Choice session.ConversationChoice }")
	if got, ok := symbolsMatch(f, imp, dep, []string{"ConversationChoice"}); !ok {
		t.Fatalf("baseline symbols %v did not match", got)
	}
	f, imp = parse("package views\nimport \"" + dep + "\"\ntype Row struct{ Choice session.ConversationChoice; Snapshot session.Snapshot }")
	if got, ok := symbolsMatch(f, imp, dep, []string{"ConversationChoice"}); ok {
		t.Fatalf("new symbol accepted: %v", got)
	}
	f, imp = parse("package views\nimport \"" + dep + "\"\nfunc local(){ session := struct{ Snapshot int }{}; _ = session.Snapshot }\ntype Row struct{ Choice session.ConversationChoice }")
	if got, ok := symbolsMatch(f, imp, dep, []string{"ConversationChoice"}); !ok {
		t.Fatalf("shadowed local selector polluted package symbols: %v", got)
	}
}

func TestExistingSelectorCannotAuthorizeNewDeclarationEdges(t *testing.T) {
	dep := module + "internal/session"
	baseline := "package views\nimport session \"" + dep + "\"\ntype Existing struct{ Choice session.ConversationChoice }"
	baseFile, err := parser.ParseFile(token.NewFileSet(), "legacy.go", baseline, 0)
	if err != nil {
		t.Fatal(err)
	}
	var baseImport *ast.ImportSpec
	for _, im := range baseFile.Imports {
		d, _ := strconv.Unquote(im.Path.Value)
		if d == dep {
			baseImport = im
		}
	}
	symbols, _ := symbolsMatch(baseFile, baseImport, dep, nil)
	edges, _ := declarationEdges(baseFile, baseImport, dep, nil)
	for _, addition := range []string{
		"type Existing struct{ Choice session.ConversationChoice; Added session.ConversationChoice }",
		"type consumer interface{ Choose() session.ConversationChoice }",
	} {
		t.Run(addition, func(t *testing.T) {
			root := newFixtureRoot(t)
			source := strings.Replace(baseline, "type Existing struct{ Choice session.ConversationChoice }", addition, 1)
			writeFixture(t, filepath.Join(root, "internal/ui/views/legacy.go"), source)
			entry := exception{Path: "internal/ui/views/legacy.go", Dependency: dep, Rule: "frontend imports backend implementation", Symbols: symbols, Edges: edges, EdgeOwners: map[string]string{}, Reason: "fixture legacy edge", Owner: "4"}
			for _, edge := range edges {
				entry.EdgeOwners[edge] = "4"
			}
			findings, err := inspect(root, map[string]exception{entry.Path + "|" + dep: entry})
			if err != nil {
				t.Fatal(err)
			}
			for _, f := range findings {
				if f.rule == "new backend exposure on excepted edge" {
					return
				}
			}
			t.Fatalf("new declaration edge with existing selector was accepted: %#v", findings)
		})
	}
}

func TestSameNamedMethodUsesDistinctReceiverEdge(t *testing.T) {
	dep := module + "internal/session"
	base := "package views\nimport session \"" + dep + "\"\ntype One struct{}\ntype Two struct{}\nfunc (One) Snapshot() session.Snapshot { return session.Snapshot{} }"
	parsed, err := parser.ParseFile(token.NewFileSet(), "legacy.go", base, 0)
	if err != nil {
		t.Fatal(err)
	}
	imp := parsed.Imports[0]
	symbols, _ := symbolsMatch(parsed, imp, dep, nil)
	edges, _ := declarationEdges(parsed, imp, dep, nil)
	entry := exception{Path: "internal/ui/views/legacy.go", Dependency: dep, Rule: "frontend imports backend implementation", Symbols: symbols, Edges: edges, EdgeOwners: map[string]string{}, Reason: "fixture method surface", Owner: "4"}
	for _, edge := range edges {
		entry.EdgeOwners[edge] = "4"
	}
	for _, source := range []string{base, base + "\nfunc (Two) Snapshot() session.Snapshot { return session.Snapshot{} }"} {
		root := newFixtureRoot(t)
		writeFixture(t, filepath.Join(root, "internal/ui/views/legacy.go"), source)
		findings, err := inspect(root, map[string]exception{entry.Path + "|" + dep: entry})
		if err != nil {
			t.Fatal(err)
		}
		if source == base {
			if len(findings) != 0 {
				t.Fatalf("baseline method rejected: %#v", findings)
			}
			continue
		}
		found := false
		for _, f := range findings {
			if f.rule == "new backend exposure on excepted edge" {
				found = true
			}
		}
		if !found {
			t.Fatalf("same-named method on a new receiver was accepted: %#v", findings)
		}
	}
}

func TestInspectRejectsIndirectFrontendTypeExposure(t *testing.T) {
	root := newFixtureRoot(t)
	dep := module + "internal/session"
	legacy := "package models\nimport session \"" + dep + "\"\ntype Legacy = session.Snapshot"
	f, err := parser.ParseFile(token.NewFileSet(), "legacy.go", legacy, 0)
	if err != nil {
		t.Fatal(err)
	}
	symbols, _ := symbolsMatch(f, f.Imports[0], dep, nil)
	edges, _ := declarationEdges(f, f.Imports[0], dep, nil)
	entry := exception{Path: "internal/ui/models/legacy.go", Dependency: dep, Rule: "frontend imports backend implementation", Symbols: symbols, Edges: edges, EdgeOwners: map[string]string{}, Reason: "fixture", Owner: "4"}
	for _, edge := range edges {
		entry.EdgeOwners[edge] = "4"
	}
	writeFixture(t, filepath.Join(root, entry.Path), legacy)
	writeFixture(t, filepath.Join(root, "internal/ui/views/new.go"), "package views\nimport models \""+module+"internal/ui/models\"\ntype New struct{ Value models.Legacy }; type Port interface{ Read() models.Legacy }")
	findings, err := inspect(root, map[string]exception{entry.Path + "|" + dep: entry})
	if err != nil {
		t.Fatal(err)
	}
	count := 0
	for _, finding := range findings {
		if finding.rule == "indirect backend type exposure" {
			count++
		}
	}
	if count < 2 {
		t.Fatalf("indirect model and port exposures were not both found: %#v", findings)
	}
}

func TestInspectRejectsIndirectPackageVariableExposure(t *testing.T) {
	root := newFixtureRoot(t)
	dep := module + "internal/session"
	legacy := "package models\nimport session \"" + dep + "\"\ntype Legacy = session.Snapshot"
	f, err := parser.ParseFile(token.NewFileSet(), "legacy.go", legacy, 0)
	if err != nil {
		t.Fatal(err)
	}
	symbols, _ := symbolsMatch(f, f.Imports[0], dep, nil)
	edges, _ := declarationEdges(f, f.Imports[0], dep, nil)
	entry := exception{Path: "internal/ui/models/legacy.go", Dependency: dep, Rule: "frontend imports backend implementation", Symbols: symbols, Edges: edges, EdgeOwners: map[string]string{}, Reason: "fixture alias", Owner: "4"}
	for _, e := range edges {
		entry.EdgeOwners[e] = "4"
	}
	writeFixture(t, filepath.Join(root, entry.Path), legacy)
	writeFixture(t, filepath.Join(root, "internal/ui/views/exported.go"), "package views\nimport models \""+module+"internal/ui/models\"\nvar Exported models.Legacy")
	findings, err := inspect(root, map[string]exception{entry.Path + "|" + dep: entry})
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range findings {
		if f.rule == "indirect backend type exposure" && strings.HasPrefix(f.symbol, "var Exported") {
			return
		}
	}
	t.Fatalf("package variable backend exposure accepted: %#v", findings)
}

func TestInspectTracksNamedTypeMethodSets(t *testing.T) {
	root := newFixtureRoot(t)
	dep := module + "internal/session"
	base := "package models\nimport session \"" + dep + "\"\ntype Holder struct{}\nfunc (Holder) Hidden() session.Snapshot { return session.Snapshot{} }"
	f, err := parser.ParseFile(token.NewFileSet(), "legacy.go", base, 0)
	if err != nil {
		t.Fatal(err)
	}
	symbols, _ := symbolsMatch(f, f.Imports[0], dep, nil)
	edges, _ := declarationEdges(f, f.Imports[0], dep, nil)
	entry := exception{Path: "internal/ui/models/legacy.go", Dependency: dep, Rule: "frontend imports backend implementation", Symbols: symbols, Edges: edges, EdgeOwners: map[string]string{}, Reason: "fixture method-set surface", Owner: "4"}
	for _, edge := range edges {
		entry.EdgeOwners[edge] = "4"
	}
	for _, addition := range []string{
		"type New struct{ Holder }",
		"type Consumer interface{ Read() Holder }",
		"type New struct{ models.Holder }",
	} {
		t.Run(addition, func(t *testing.T) {
			writeFixture(t, filepath.Join(root, entry.Path), base)
			if strings.Contains(addition, "models.Holder") {
				writeFixture(t, filepath.Join(root, "internal/ui/views/use.go"), "package views\nimport models \""+module+"internal/ui/models\"\n"+addition)
			} else {
				writeFixture(t, filepath.Join(root, entry.Path), base+"\n"+addition)
			}
			findings, err := inspect(root, map[string]exception{entry.Path + "|" + dep: entry})
			if err != nil {
				t.Fatal(err)
			}
			for _, f := range findings {
				if f.rule == "indirect backend type exposure" {
					return
				}
			}
			t.Fatalf("method-set exposure accepted: %s; findings=%#v", addition, findings)
		})
	}
}

func TestInspectIgnoresFieldNamesAndTypeParameterShadowing(t *testing.T) {
	root := newFixtureRoot(t)
	dep := module + "internal/session"
	source := "package models\nimport session \"" + dep + "\"\ntype Legacy = session.Snapshot\ntype Clean struct{ Legacy string }\ntype Generic[Legacy any] struct{ Value Legacy }"
	f, err := parser.ParseFile(token.NewFileSet(), "legacy.go", source, 0)
	if err != nil {
		t.Fatal(err)
	}
	symbols, _ := symbolsMatch(f, f.Imports[0], dep, nil)
	edges, _ := declarationEdges(f, f.Imports[0], dep, nil)
	entry := exception{Path: "internal/ui/models/legacy.go", Dependency: dep, Rule: "frontend imports backend implementation", Symbols: symbols, Edges: edges, EdgeOwners: map[string]string{}, Reason: "fixture alias", Owner: "4"}
	for _, e := range edges {
		entry.EdgeOwners[e] = "4"
	}
	writeFixture(t, filepath.Join(root, entry.Path), source)
	findings, err := inspect(root, map[string]exception{entry.Path + "|" + dep: entry})
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range findings {
		if f.rule == "indirect backend type exposure" {
			t.Fatalf("non-type identifier was treated as an exposure: %#v", findings)
		}
	}
}

func TestInspectFindsVariadicAndGenericConstraintLeaks(t *testing.T) {
	dep := module + "internal/session"
	legacy := "package models\nimport session \"" + dep + "\"\ntype Legacy = session.Snapshot"
	cases := []struct{ name, files string }{
		{"variadic interface method", "internal/ui/views/use.go\npackage views\nimport models \"" + module + "internal/ui/models\"\ntype Consumer interface{ Read(...models.Legacy) }"},
		{"generic type constraint", "internal/ui/views/use.go\npackage views\nimport models \"" + module + "internal/ui/models\"\ntype Constrained[T models.Legacy] struct{}"},
		{"generic function constraint", "internal/ui/views/use.go\npackage views\nimport models \"" + module + "internal/ui/models\"\nfunc Consume[T models.Legacy]() {}"},
		{"union and approximation", "internal/ui/views/use.go\npackage views\nimport models \"" + module + "internal/ui/models\"\ntype Constrained[T interface{ models.Legacy | ~string }] struct{}"},
		{"consumer local alias", "internal/ui/views/use.go\npackage views\nimport models \"" + module + "internal/ui/models\"\ntype Local = models.Legacy\ntype Constrained[T Local] struct{}"},
		{"local constraint alias", "internal/ui/models/constraint.go\npackage models\ntype Local = Legacy\n---\ninternal/ui/views/use.go\npackage views\nimport models \"" + module + "internal/ui/models\"\ntype Constrained[T models.Local] struct{}"},
		{"imported constraint alias", "internal/ui/models/constraint.go\npackage models\ntype Local = Legacy\n---\ninternal/ui/views/use.go\npackage views\nimport models \"" + module + "internal/ui/models\"\ntype Constrained[T models.Local] struct{}"},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			root := newFixtureRoot(t)
			f, err := parser.ParseFile(token.NewFileSet(), "legacy.go", legacy, 0)
			if err != nil {
				t.Fatal(err)
			}
			symbols, _ := symbolsMatch(f, f.Imports[0], dep, nil)
			edges, _ := declarationEdges(f, f.Imports[0], dep, nil)
			entry := exception{Path: "internal/ui/models/legacy.go", Dependency: dep, Rule: "frontend imports backend implementation", Symbols: symbols, Edges: edges, EdgeOwners: map[string]string{}, Reason: "fixture backend alias", Owner: "4"}
			for _, e := range edges {
				entry.EdgeOwners[e] = "4"
			}
			writeFixture(t, filepath.Join(root, entry.Path), legacy)
			for _, part := range strings.Split(tt.files, "\n---\n") {
				path, content, ok := strings.Cut(part, "\n")
				if !ok {
					t.Fatalf("bad fixture encoding: %q", part)
				}
				writeFixture(t, filepath.Join(root, path), content)
			}
			findings, err := inspect(root, map[string]exception{entry.Path + "|" + dep: entry})
			if err != nil {
				t.Fatal(err)
			}
			for _, f := range findings {
				if f.rule == "indirect backend type exposure" {
					return
				}
			}
			t.Fatalf("generic or variadic exposure accepted: %#v", findings)
		})
	}
	t.Run("independent variadic and unconstrained generic", func(t *testing.T) {
		root := newFixtureRoot(t)
		f, err := parser.ParseFile(token.NewFileSet(), "legacy.go", legacy, 0)
		if err != nil {
			t.Fatal(err)
		}
		symbols, _ := symbolsMatch(f, f.Imports[0], dep, nil)
		edges, _ := declarationEdges(f, f.Imports[0], dep, nil)
		entry := exception{Path: "internal/ui/models/legacy.go", Dependency: dep, Rule: "frontend imports backend implementation", Symbols: symbols, Edges: edges, EdgeOwners: map[string]string{}, Reason: "fixture backend alias", Owner: "4"}
		for _, e := range edges {
			entry.EdgeOwners[e] = "4"
		}
		writeFixture(t, filepath.Join(root, entry.Path), legacy)
		writeFixture(t, filepath.Join(root, "internal/ui/views/clean.go"), "package views\nfunc Consume(...string) {}\ntype Generic[T any] struct{ Value T }")
		findings, err := inspect(root, map[string]exception{entry.Path + "|" + dep: entry})
		if err != nil {
			t.Fatal(err)
		}
		for _, f := range findings {
			if f.rule == "indirect backend type exposure" {
				t.Fatalf("independent type was rejected: %#v", findings)
			}
		}
	})
}

func TestPointerGenericReceiverTypeParamsRespectShadowing(t *testing.T) {
	dep := module + "internal/session"
	fixtures := []struct {
		name, source string
		wantFinding  bool
	}{
		{"one receiver parameter shadows alias", "package models\nimport session \"" + dep + "\"\ntype Legacy = session.Snapshot\ntype Box[T any] struct{}\nfunc (*Box[Legacy]) Get() Legacy { var zero Legacy; return zero }", false},
		{"multiple receiver parameters shadow aliases", "package models\nimport session \"" + dep + "\"\ntype Legacy = session.Snapshot\ntype Other = session.ConversationChoice\ntype Pair[T, U any] struct{}\nfunc (*Pair[Legacy, Other]) Get() (Legacy, Other) { var zero Legacy; var other Other; return zero, other }", false},
		{"method genuinely exposes backend alias", "package models\nimport session \"" + dep + "\"\ntype Legacy = session.Snapshot\ntype Box[T any] struct{}\nfunc (*Box[T]) Get() Legacy { return Legacy{} }", true},
	}
	for _, tt := range fixtures {
		t.Run(tt.name, func(t *testing.T) {
			root := newFixtureRoot(t)
			f, err := parser.ParseFile(token.NewFileSet(), "legacy.go", tt.source, 0)
			if err != nil {
				t.Fatal(err)
			}
			var im *ast.ImportSpec
			for _, candidate := range f.Imports {
				d, _ := strconv.Unquote(candidate.Path.Value)
				if d == dep {
					im = candidate
				}
			}
			symbols, _ := symbolsMatch(f, im, dep, nil)
			edges, _ := declarationEdges(f, im, dep, nil)
			entry := exception{Path: "internal/ui/models/legacy.go", Dependency: dep, Rule: "frontend imports backend implementation", Symbols: symbols, Edges: edges, EdgeOwners: map[string]string{}, Reason: "fixture generic alias", Owner: "4"}
			for _, edge := range edges {
				entry.EdgeOwners[edge] = "4"
			}
			writeFixture(t, filepath.Join(root, entry.Path), tt.source)
			findings, err := inspect(root, map[string]exception{entry.Path + "|" + dep: entry})
			if err != nil {
				t.Fatal(err)
			}
			got := false
			for _, f := range findings {
				if f.rule == "indirect backend type exposure" {
					got = true
				}
			}
			if got != tt.wantFinding {
				t.Fatalf("indirect exposure=%v want %v: %#v", got, tt.wantFinding, findings)
			}
		})
	}
}

func TestTypeExposureAllowlistRejectsInvalidOwnerAndDuplicate(t *testing.T) {
	for name, data := range map[string]string{
		"invalid owner": `[{"path":"internal/ui/views/example.go","symbol":"type Example @1:2","dependency":"` + module + `internal/session.Snapshot","owner":"1","reason":"fixture"}]`,
		"duplicate":     `[{"path":"internal/ui/views/example.go","symbol":"type Example @1:2","dependency":"` + module + `internal/session.Snapshot","owner":"2","reason":"fixture"},{"path":"internal/ui/views/example.go","symbol":"type Example @1:2","dependency":"` + module + `internal/session.Snapshot","owner":"2","reason":"fixture"}]`,
	} {
		t.Run(name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "allowlist.json")
			if err := os.WriteFile(path, []byte(data), 0600); err != nil {
				t.Fatal(err)
			}
			if _, err := readTypeExposureAllowlist(path); err == nil {
				t.Fatal("invalid type exposure allowlist accepted")
			}
		})
	}
}

func TestContractLeakFixtures(t *testing.T) {
	fixtures := []struct{ name, source string }{
		{"alias", "package contracts\nimport \"github.com/BlackGiack/kankuro/internal/session\"\ntype Alias = session.Snapshot"},
		{"embed and nested container", "package contracts\nimport backend \"github.com/BlackGiack/kankuro/internal/config\"\ntype Inner struct{ Values map[string][]*backend.IDCard }; type Outer struct{ Inner }"},
		{"method and generic", "package contracts\nimport \"github.com/BlackGiack/kankuro/internal/session\"\ntype Box[T any] struct{ Value T }; type Result struct{}; func (Result) Get() Box[session.Snapshot] { return Box[session.Snapshot]{} }"},
		{"cycle", "package contracts\nimport \"github.com/BlackGiack/kankuro/internal/session\"\ntype A struct{ B *B }; type B struct{ A *A; Value session.Snapshot }"},
	}
	for _, tt := range fixtures {
		t.Run(tt.name, func(t *testing.T) {
			f, err := parser.ParseFile(token.NewFileSet(), "types.go", tt.source, 0)
			if err != nil {
				t.Fatal(err)
			}
			aliases := map[string]string{}
			for _, im := range f.Imports {
				d, _ := strconv.Unquote(im.Path.Value)
				name := filepath.Base(d)
				if im.Name != nil {
					name = im.Name.Name
				}
				aliases[name] = d
			}
			found := false
			ast.Inspect(f, func(n ast.Node) bool {
				s, ok := n.(*ast.SelectorExpr)
				if !ok {
					return true
				}
				id, ok := s.X.(*ast.Ident)
				if ok && strings.HasPrefix(aliases[id.Name], module+"internal/") {
					found = true
				}
				return true
			})
			if !found {
				t.Fatal("backend type edge was not found")
			}
		})
	}
}

func TestInspectRejectsBackendLeakThroughContractGraph(t *testing.T) {
	root := newFixtureRoot(t)
	writeFixture(t, filepath.Join(root, "internal/ui/views/view.go"), "package views\nimport \""+module+"internal/api/contracts\"")
	contractSource := "package contracts\nimport \"" + module + "internal/session\"\ntype Row struct{ Snapshot session.Snapshot }"
	writeFixture(t, filepath.Join(root, "internal/api/contracts/name.go"), contractSource)
	findings, err := inspect(root, map[string]exception{})
	if err != nil {
		t.Fatal(err)
	}
	for _, finding := range findings {
		if finding.path == "internal/api/contracts/name.go" && finding.rule == "contracts import implementation or presentation" {
			return
		}
	}
	t.Fatalf("whole inspect accepted a backend type leak through the contract package: %#v", findings)
}

func TestTypedContractGraphFixtures(t *testing.T) {
	backend := types.NewPackage(module+"internal/session", "session")
	obj := types.NewTypeName(token.NoPos, backend, "Snapshot", nil)
	_ = types.NewNamed(obj, types.NewStruct(nil, nil), nil)
	backend.Scope().Insert(obj)
	backend.MarkComplete()
	fixtures := []struct{ name, source string }{
		{"alias", "package contracts\nimport s \"" + backend.Path() + "\"\ntype Alias = s.Snapshot"},
		{"embed and nested containers", "package contracts\nimport s \"" + backend.Path() + "\"\ntype Inner struct{ Values map[string][]*s.Snapshot }; type Outer struct{ Inner }"},
		{"generic method signature", "package contracts\nimport s \"" + backend.Path() + "\"\ntype Box[T any] struct{ Value T }; func (b Box[T]) Get() s.Snapshot { return s.Snapshot{} }"},
		{"cycle", "package contracts\nimport s \"" + backend.Path() + "\"\ntype A struct{ B *B }; type B struct{ A *A; Value s.Snapshot }"},
	}
	for _, tt := range fixtures {
		t.Run(tt.name, func(t *testing.T) {
			fset := token.NewFileSet()
			file, err := parser.ParseFile(fset, "fixture.go", tt.source, 0)
			if err != nil {
				t.Fatal(err)
			}
			imp := mapImporter{packages: map[string]*types.Package{backend.Path(): backend}}
			pkg, err := (&types.Config{Importer: imp}).Check("fixture/contracts", fset, []*ast.File{file}, nil)
			if err != nil {
				t.Fatalf("fixture type-check: %v", err)
			}
			leaked := false
			for _, name := range pkg.Scope().Names() {
				if !token.IsExported(name) {
					continue
				}
				if leak := walkContractType(pkg.Scope().Lookup(name).Type(), name, map[types.Type]bool{}); leak != "" {
					leaked = true
					break
				}
			}
			if !leaked {
				t.Fatal("typed backend edge was not found")
			}
		})
	}
}

type mapImporter struct{ packages map[string]*types.Package }

func (m mapImporter) Import(path string) (*types.Package, error) {
	if p := m.packages[path]; p != nil {
		return p, nil
	}
	return importer.Default().Import(path)
}

func TestHelperAndTestSupportFixtures(t *testing.T) {
	t.Run("neutral helper may depend on independent contract", func(t *testing.T) {
		root := newFixtureRoot(t)
		writeFixture(t, filepath.Join(root, "internal/ui/views/view.go"), "package views\nimport \""+module+"internal/helper\"")
		writeFixture(t, filepath.Join(root, "internal/helper/helper.go"), "package helper\nimport \""+module+"internal/api/contracts\"\ntype Name = contracts.Name")
		writeFixture(t, filepath.Join(root, "internal/api/contracts/name.go"), "package contracts\ntype Name string")
		findings, err := inspect(root, map[string]exception{})
		if err != nil {
			t.Fatal(err)
		}
		for _, f := range findings {
			if f.rule == "transitive frontend dependency" {
				t.Fatalf("independent contract was treated as implementation: %#v", findings)
			}
		}
	})
	t.Run("arbitrary helper and ASCII concealment", func(t *testing.T) {
		for _, start := range []string{"internal/newhelper", "internal/ascii"} {
			root := newFixtureRoot(t)
			writeFixture(t, filepath.Join(root, "internal/ui/views/view.go"), "package views\nimport \""+module+start+"\"")
			if start == "internal/ascii" {
				writeFixture(t, filepath.Join(root, "internal/ascii/ascii.go"), "package ascii\nimport \""+module+"internal/newhelper\"")
			}
			writeFixture(t, filepath.Join(root, "internal/newhelper/helper.go"), "package newhelper\nimport \""+module+"internal/session\"\ntype Hidden struct{ Value session.Snapshot }")
			findings, err := inspect(root, map[string]exception{})
			if err != nil {
				t.Fatal(err)
			}
			found := false
			for _, f := range findings {
				if f.rule == "transitive frontend dependency" {
					found = true
				}
			}
			if !found {
				t.Fatalf("%s helper bypass was accepted: %#v", start, findings)
			}
		}
	})
	t.Run("cyclic shared helpers terminate and report every forbidden edge", func(t *testing.T) {
		root := newFixtureRoot(t)
		writeFixture(t, filepath.Join(root, "internal/ui/views/view.go"), "package views\nimport \""+module+"internal/helpera\"")
		writeFixture(t, filepath.Join(root, "internal/helpera/a.go"), "package helpera\nimport \""+module+"internal/helperb\"")
		writeFixture(t, filepath.Join(root, "internal/helperb/b.go"), "package helperb\nimport (\""+module+"internal/helpera\"; \""+module+"internal/session\"; \""+module+"internal/config\")")
		findings, err := inspect(root, map[string]exception{})
		if err != nil {
			t.Fatal(err)
		}
		count := 0
		for _, f := range findings {
			if f.rule == "transitive frontend dependency" {
				count++
			}
		}
		if count != 2 {
			t.Fatalf("found %d transitive edges, want both: %#v", count, findings)
		}
	})
	t.Run("backend cannot import coordinator", func(t *testing.T) {
		root := newFixtureRoot(t)
		writeFixture(t, filepath.Join(root, "internal/config/backend.go"), "package config\nimport \""+module+"internal/app\"")
		findings, err := inspect(root, map[string]exception{})
		if err != nil {
			t.Fatal(err)
		}
		for _, f := range findings {
			if f.rule == "backend imports frontend" {
				return
			}
		}
		t.Fatalf("backend coordinator import accepted: %#v", findings)
	})
	t.Run("helper parse failures are surfaced", func(t *testing.T) {
		root := newFixtureRoot(t)
		writeFixture(t, filepath.Join(root, "internal/ui/views/view.go"), "package views\nimport \""+module+"internal/newhelper\"")
		writeFixture(t, filepath.Join(root, "internal/newhelper/helper.go"), "package newhelper\nfunc broken(")
		if _, err := inspect(root, map[string]exception{}); err == nil {
			t.Fatal("helper parse error was hidden")
		}
	})
	t.Run("production code is not test support", func(t *testing.T) {
		root := t.TempDir()
		prod := filepath.Join(root, "internal/ui/views/view.go")
		test := filepath.Join(root, "internal/ui/views/view_test.go")
		for _, p := range []string{prod, test} {
			if err := os.MkdirAll(filepath.Dir(p), 0755); err != nil {
				t.Fatal(err)
			}
		}
		if err := os.WriteFile(prod, []byte("package views\nimport \""+module+"internal/ui/uitest\""), 0600); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(test, []byte("package views\nimport \""+module+"internal/ui/uitest\""), 0600); err != nil {
			t.Fatal(err)
		}
		files := []string{prod, test}
		imports := map[string][]string{"internal/ui/views/view.go": {module + "internal/ui/uitest"}, "internal/ui/views/view_test.go": {module + "internal/ui/uitest"}}
		if testSupportIsSafe("internal/ui/uitest/shape.go", files, imports, root) {
			t.Fatal("production importer was accepted as test support")
		}
	})
}

func newFixtureRoot(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	writeFixture(t, filepath.Join(root, "internal/api/contracts/doc.go"), "package contracts")
	return root
}
func writeFixture(t *testing.T, path, source string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
}

func TestUnusedExceptionFixture(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "internal"), 0755); err != nil {
		t.Fatal(err)
	}
	contractsDir := filepath.Join(root, "internal/api/contracts")
	if err := os.MkdirAll(contractsDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(contractsDir, "doc.go"), []byte("package contracts"), 0600); err != nil {
		t.Fatal(err)
	}
	entry := exception{Path: "internal/ui/views/removed.go", Dependency: module + "internal/config", Reason: "fixture", Owner: "2"}
	findings, err := inspect(root, map[string]exception{entry.Path + "|" + entry.Dependency: entry})
	if err != nil {
		t.Fatal(err)
	}
	if len(findings) != 1 || findings[0].rule != "unused exception" {
		t.Fatalf("findings = %#v, want one unused-exception finding", findings)
	}
}

func TestTestImportsCannotConsumeProductionExceptions(t *testing.T) {
	root := newFixtureRoot(t)
	writeFixture(t, filepath.Join(root, "internal/ui/views/legacy_test.go"), "package views_test\nimport \""+module+"internal/config\"")
	entry := exception{Path: "internal/ui/views/legacy.go", Dependency: module + "internal/config", Rule: "frontend imports backend implementation", Reason: "fixture", Owner: "2"}
	findings, err := inspect(root, map[string]exception{entry.Path + "|" + entry.Dependency: entry})
	if err != nil {
		t.Fatal(err)
	}
	if len(findings) != 1 || findings[0].rule != "unused exception" {
		t.Fatalf("test import consumed a production exception: %#v", findings)
	}
	data := `[{"path":"internal/ui/views/legacy_test.go","dependency":"` + module + `internal/config","rule":"frontend imports backend implementation","reason":"bad","owner":"2"}]`
	file := filepath.Join(root, "test-exception.json")
	writeFixture(t, file, data)
	if _, err := readAllowlist(file); err == nil {
		t.Fatal("test file exception accepted")
	}
}

func TestContractUnknownTypeAndExceptionFixtures(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, "internal/api/contracts")
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "types.go")
	if err := os.WriteFile(path, []byte("package contracts\ntype Public struct{ Value MissingType }"), 0600); err != nil {
		t.Fatal(err)
	}
	var files []string
	err := filepath.WalkDir(filepath.Join(root, "internal"), func(p string, d fs.DirEntry, e error) error {
		if e != nil {
			return e
		}
		if !d.IsDir() && strings.HasSuffix(p, ".go") {
			files = append(files, p)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := contractFindings(root, files); err == nil {
		t.Fatal("unresolved contract type was accepted")
	}
	badImportRoot := newFixtureRoot(t)
	writeFixture(t, filepath.Join(badImportRoot, "internal/api/contracts/bad.go"), "package contracts\nimport \""+module+"internal/session\"\ntype Leaked struct{ Snapshot session.Snapshot }")
	badFindings, err := inspect(badImportRoot, map[string]exception{})
	if err != nil {
		t.Fatalf("forbidden imports should produce findings, not mask the import edge: %v", err)
	}
	if len(badFindings) == 0 || badFindings[0].rule != "contracts import implementation or presentation" {
		t.Fatalf("forbidden contract import missing diagnostic: %#v", badFindings)
	}
	inactiveRoot := newFixtureRoot(t)
	writeFixture(t, filepath.Join(inactiveRoot, "internal/api/contracts/inactive_windows.go"), "//go:build windows && architecture_fixture\npackage contracts\ntype Inactive struct{ Value MissingOnWindows }")
	if _, err := contractFindings(inactiveRoot, allFixtureGoFiles(t, inactiveRoot)); err == nil {
		t.Fatal("unresolved inactive-variant contract type was accepted")
	}
	data := `[{"path":"internal/api/contracts/types.go","dependency":"` + module + `internal/session","reason":"bad","owner":"2"}]`
	file := filepath.Join(root, "exceptions.json")
	if err := os.WriteFile(file, []byte(data), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := readAllowlist(file); err == nil {
		t.Fatal("contract exception was accepted")
	}
}

func TestContractInspectionAcceptsLocalCyclesStandardImportsAndTests(t *testing.T) {
	root := newFixtureRoot(t)
	writeFixture(t, filepath.Join(root, "internal/api/contracts/base.go"), "package contracts\nimport \"time\"\ntype Base struct{ Created time.Time; Parent *Base }\ntype Row struct{ Base Base; Next *Row }\ntype Timestamp = time.Time")
	writeFixture(t, filepath.Join(root, "internal/api/contracts/variant_linux.go"), "//go:build linux\npackage contracts\ntype Variant struct{ Linux string }")
	writeFixture(t, filepath.Join(root, "internal/api/contracts/variant_windows.go"), "//go:build windows\npackage contracts\ntype Variant struct{ Windows int }")
	writeFixture(t, filepath.Join(root, "internal/api/contracts/types_test.go"), "package contracts\nimport \"github.com/BlackGiack/kankuro/internal/session\"\nvar _ session.Snapshot")
	writeFixture(t, filepath.Join(root, "internal/api/contracts/external_test.go"), "package contracts_test\nimport \"github.com/BlackGiack/kankuro/internal/config\"\nvar _ config.Dashboard")
	writeFixture(t, filepath.Join(root, "internal/api/contracts/cards/card.go"), "package cards\nimport (\"time\"; base \""+module+"internal/api/contracts\")\ntype Summary struct{ Base base.Base; Created time.Time }")
	findings, err := contractFindings(root, allFixtureGoFiles(t, root))
	if err != nil {
		t.Fatalf("valid contract declarations rejected: %v", err)
	}
	if len(findings) != 0 {
		t.Fatalf("valid contract declarations reported leaks: %#v", findings)
	}
}

func TestContractParentRecheckedWhenChildVariantChanges(t *testing.T) {
	root := newFixtureRoot(t)
	writeFixture(t, filepath.Join(root, "internal/api/contracts/base.go"), "package contracts\nimport detail \""+module+"internal/api/contracts/detail\"\ntype Result = detail.LinuxOnly")
	writeFixture(t, filepath.Join(root, "internal/api/contracts/detail/detail_linux.go"), "//go:build linux\npackage detail\ntype LinuxOnly struct{ Name string }")
	writeFixture(t, filepath.Join(root, "internal/api/contracts/detail/detail_windows.go"), "//go:build windows\npackage detail\ntype WindowsOnly struct{ Name string }")
	if _, err := contractFindings(root, allFixtureGoFiles(t, root)); err == nil {
		t.Fatal("parent contract accepted a type missing from its Windows dependency variant")
	}
}

func TestContractVariantsCheckBothCgoStates(t *testing.T) {
	for _, tag := range []string{"cgo", "!cgo"} {
		t.Run(tag, func(t *testing.T) {
			root := newFixtureRoot(t)
			writeFixture(t, filepath.Join(root, "internal/api/contracts/cgo.go"), "//go:build "+tag+"\npackage contracts\ntype Invalid struct{ MissingType }")
			if _, err := contractFindings(root, allFixtureGoFiles(t, root)); err == nil {
				t.Fatalf("unresolved contract guarded by %s was not checked", tag)
			}
		})
	}
}

func TestContractParentRecheckedWhenChildCgoVariantChanges(t *testing.T) {
	root := newFixtureRoot(t)
	writeFixture(t, filepath.Join(root, "internal/api/contracts/base.go"), "package contracts\nimport detail \""+module+"internal/api/contracts/detail\"\ntype Result = detail.CgoOnly")
	writeFixture(t, filepath.Join(root, "internal/api/contracts/detail/detail_cgo.go"), "//go:build cgo\npackage detail\ntype CgoOnly struct{ Name string }")
	writeFixture(t, filepath.Join(root, "internal/api/contracts/detail/detail_nocgo.go"), "//go:build !cgo\npackage detail\ntype NoCgoOnly struct{ Name string }")
	if _, err := contractFindings(root, allFixtureGoFiles(t, root)); err == nil {
		t.Fatal("parent accepted a type missing from the non-cgo dependency variant")
	}
}

func allFixtureGoFiles(t *testing.T, root string) []string {
	t.Helper()
	var out []string
	err := filepath.WalkDir(filepath.Join(root, "internal"), func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && strings.HasSuffix(path, ".go") {
			out = append(out, path)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	sort.Strings(out)
	return out
}

func readAllowlist(path string) (map[string]exception, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var rows []exception
	if err := json.Unmarshal(data, &rows); err != nil {
		return nil, err
	}
	out := make(map[string]exception, len(rows))
	for _, e := range rows {
		key := e.Path + "|" + e.Dependency
		if e.Path == "" || e.Dependency == "" || e.Rule == "" || e.Reason == "" || e.Owner == "" {
			return nil, fmt.Errorf("incomplete exception %q", key)
		}
		if strings.HasPrefix(e.Path, "internal/api/contracts/") {
			return nil, fmt.Errorf("contracts cannot have exceptions: %s", key)
		}
		for _, owner := range strings.Split(e.Owner, "/") {
			if owner < "2" || owner > "8" || len(owner) != 1 {
				return nil, fmt.Errorf("invalid removal owner %q on %s", owner, key)
			}
		}
		if e.Classification == "test-support" && (e.Path != "internal/ui/uitest/shape.go" || e.Dependency != module+"internal/config") {
			return nil, fmt.Errorf("test-support is restricted to the reviewed helper edge: %s", key)
		}
		seenSymbols := map[string]bool{}
		for _, symbol := range e.Symbols {
			if symbol == "" || seenSymbols[symbol] {
				return nil, fmt.Errorf("invalid/duplicate selector %q on %s", symbol, key)
			}
			seenSymbols[symbol] = true
		}
		seenEdges := map[string]bool{}
		for _, edge := range e.Edges {
			if edge == "" || seenEdges[edge] {
				return nil, fmt.Errorf("invalid/duplicate declaration edge on %s", key)
			}
			seenEdges[edge] = true
			owner, ok := e.EdgeOwners[edge]
			if !ok {
				return nil, fmt.Errorf("missing symbol-level removal owner for %q on %s", edge, key)
			}
			if len(owner) != 1 || owner < "2" || owner > "8" {
				return nil, fmt.Errorf("invalid symbol-level removal owner %q for %q", owner, edge)
			}
		}
		if len(e.EdgeOwners) != len(e.Edges) {
			return nil, fmt.Errorf("orphan symbol-level removal owner on %s", key)
		}
		if strings.HasSuffix(e.Path, "_test.go") {
			return nil, fmt.Errorf("test files cannot have production exceptions: %s", key)
		}
		if len(e.Symbols) == 0 || len(e.Edges) == 0 {
			return nil, fmt.Errorf("exception lacks exact imported symbol/declaration inventory: %s", key)
		}
		if _, ok := out[key]; ok {
			return nil, fmt.Errorf("duplicate exception %q", key)
		}
		if e.Classification != "" && e.Classification != "test-support" {
			return nil, fmt.Errorf("invalid classification on %q", key)
		}
		out[key] = e
	}
	return out, nil
}

func inspect(root string, allow map[string]exception) ([]finding, error) {
	used := map[string]bool{}
	var files []string
	err := filepath.WalkDir(filepath.Join(root, "internal"), func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() && (d.Name() == "testdata" || d.Name() == "vendor") {
			return filepath.SkipDir
		}
		if !d.IsDir() && strings.HasSuffix(path, ".go") {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Strings(files)
	var findings []finding
	importsByFile := map[string][]string{}
	for _, path := range files {
		rel, _ := filepath.Rel(root, path)
		rel = filepath.ToSlash(rel)
		file, err := parser.ParseFile(token.NewFileSet(), path, nil, parser.ParseComments)
		if err != nil {
			return nil, fmt.Errorf("parse %s: %w", rel, err)
		}
		if strings.HasSuffix(rel, "_test.go") {
			continue
		}
		for _, imp := range file.Imports {
			dep, err := strconvUnquote(imp.Path.Value)
			if err != nil {
				return nil, fmt.Errorf("%s: %w", rel, err)
			}
			importsByFile[rel] = append(importsByFile[rel], dep)
			f := directFinding(rel, dep)
			if f == nil {
				continue
			}
			key := rel + "|" + dep
			if allowed, ok := allow[key]; ok {
				if strings.HasSuffix(rel, "_test.go") {
					return nil, fmt.Errorf("test file cannot consume production exception %s", key)
				}
				used[key] = true
				if allowed.Rule != f.rule {
					return nil, fmt.Errorf("exception %s declares rule %q; current finding is %q", key, allowed.Rule, f.rule)
				}
				if got, ok := symbolsMatch(file, imp, dep, allow[key].Symbols); !ok {
					findings = append(findings, finding{path: rel, symbol: "imported selector", rule: "new backend symbol on excepted edge", dependency: dep + " [" + strings.Join(got, ",") + "]"})
				}
				if got, ok := declarationEdges(file, imp, dep, allow[key].Edges); !ok {
					findings = append(findings, finding{path: rel, symbol: "declaration/member/signature edge", rule: "new backend exposure on excepted edge", dependency: dep + " [" + strings.Join(got, "; ") + "]"})
				}
			} else {
				findings = append(findings, *f)
			}
		}
	}
	// Repeat test-support validation after all production imports are known.
	for key, e := range allow {
		if e.Classification == "test-support" && !testSupportIsSafe(e.Path, files, importsByFile, root) {
			return nil, fmt.Errorf("test-support exception %s has a production importer", key)
		}
	}
	for key := range allow {
		if !used[key] {
			findings = append(findings, finding{path: key, rule: "unused exception", symbol: "allowlist entry", dependency: "remove stale exception"})
		}
	}
	// Follow every permitted shared internal package reachable from a frontend.
	for rel, deps := range importsByFile {
		if !isFrontend(rel) {
			continue
		}
		for _, dep := range deps {
			if strings.HasPrefix(dep, module+"internal/") && !isFrontendImport(dep) && !forbiddenFrontendImport(dep) && !isContractPackagePath(dep) {
				graph, e := dependencyGraphFindings(root, dep, rel+" -> "+dep, map[string]bool{})
				if e != nil {
					return nil, e
				}
				findings = append(findings, graph...)
			}
		}
	}
	contractImportFailed := false
	for _, f := range findings {
		if strings.HasPrefix(f.path, "internal/api/contracts/") && f.rule == "contracts import implementation or presentation" {
			contractImportFailed = true
			break
		}
	}
	if !contractImportFailed {
		contractEdges, err := contractFindings(root, files)
		if err != nil {
			return nil, err
		}
		findings = append(findings, contractEdges...)
	}
	typeFindings, err := frontendTypeGraphFindings(root, files)
	if err != nil {
		return nil, err
	}
	typeAllowPath := filepath.Join(root, "internal/architecture/type_exposure_allowlist.json")
	typeAllow, err := readTypeExposureAllowlist(typeAllowPath)
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	typeUsed := map[string]bool{}
	for _, f := range typeFindings {
		key := typeExposureKey(f)
		if typeAllow[key] {
			typeUsed[key] = true
		} else {
			findings = append(findings, f)
		}
	}
	for key := range typeAllow {
		if !typeUsed[key] {
			findings = append(findings, finding{path: key, rule: "unused indirect type exposure", symbol: "allowlist entry", dependency: "remove stale entry"})
		}
	}
	sort.Slice(findings, func(i, j int) bool {
		a, b := findings[i], findings[j]
		if a.path != b.path {
			return a.path < b.path
		}
		if a.rule != b.rule {
			return a.rule < b.rule
		}
		return a.dependency < b.dependency
	})
	return findings, nil
}

func typeExposureKey(f finding) string { return f.path + "|" + f.symbol + "|" + f.dependency }
func readTypeExposureAllowlist(path string) (map[string]bool, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var entries []struct {
		Path       string `json:"path"`
		Symbol     string `json:"symbol"`
		Dependency string `json:"dependency"`
		Owner      string `json:"owner"`
		Reason     string `json:"reason"`
	}
	if err := json.Unmarshal(data, &entries); err != nil {
		return nil, fmt.Errorf("decode type exposure allowlist: %w", err)
	}
	out := map[string]bool{}
	for _, e := range entries {
		if e.Path == "" || e.Symbol == "" || e.Dependency == "" || e.Reason == "" {
			return nil, fmt.Errorf("invalid type exposure allowlist entry")
		}
		if len(e.Owner) != 1 || e.Owner < "2" || e.Owner > "8" {
			return nil, fmt.Errorf("invalid type exposure removal owner %q for %s", e.Owner, e.Path)
		}
		key := e.Path + "|" + e.Symbol + "|" + e.Dependency
		if out[key] {
			return nil, fmt.Errorf("duplicate type exposure allowlist entry %s", key)
		}
		out[key] = true
	}
	return out, nil
}

// frontendTypeGraphFindings resolves only type-expression names. It follows
// aliases and named-type methods, including methods promoted by embedding.
func frontendTypeGraphFindings(root string, files []string) ([]finding, error) {
	type source struct {
		rel, importPath string
		file            *ast.File
		fset            *token.FileSet
		imports         map[string]string
	}
	packages := map[string][]*source{}
	for _, path := range files {
		rel, _ := filepath.Rel(root, path)
		rel = filepath.ToSlash(rel)
		if !isFrontend(rel) || strings.HasSuffix(rel, "_test.go") {
			continue
		}
		fset := token.NewFileSet()
		f, err := parser.ParseFile(fset, path, nil, 0)
		if err != nil {
			return nil, fmt.Errorf("parse frontend type graph %s: %w", rel, err)
		}
		pkgPath := module + filepath.ToSlash(filepath.Dir(rel))
		s := &source{rel: rel, importPath: pkgPath, file: f, fset: fset, imports: map[string]string{}}
		for _, im := range f.Imports {
			dep, e := strconvUnquote(im.Path.Value)
			if e != nil {
				continue
			}
			alias := filepath.Base(dep)
			if im.Name != nil {
				alias = im.Name.Name
			}
			s.imports[alias] = dep
		}
		packages[pkgPath] = append(packages[pkgPath], s)
	}
	for pkgPath := range packages {
		sort.Slice(packages[pkgPath], func(i, j int) bool {
			return packages[pkgPath][i].rel < packages[pkgPath][j].rel
		})
	}
	var resolve func(pkg, name string, visiting map[string]bool) string
	var tainted func(pkg string, f *ast.File, imports map[string]string, expr ast.Expr, shadow map[string]bool, visiting map[string]bool) string
	tainted = func(pkg string, f *ast.File, imports map[string]string, expr ast.Expr, shadow map[string]bool, visiting map[string]bool) string {
		switch x := expr.(type) {
		case *ast.SelectorExpr:
			if id, ok := x.X.(*ast.Ident); ok {
				dep := imports[id.Name]
				// Contract package graphs are checked independently below. Their
				// declarations are intentionally safe frontend-facing types.
				if isContractPackagePath(dep) {
					return ""
				}
				if forbiddenFrontendImport(dep) {
					return dep + "." + x.Sel.Name
				}
				if isFrontendImport(dep) {
					return resolve(dep, x.Sel.Name, visiting)
				}
			}
		case *ast.Ident:
			if shadow[x.Name] {
				return ""
			}
			return resolve(pkg, x.Name, visiting)
		case *ast.StarExpr:
			return tainted(pkg, f, imports, x.X, shadow, visiting)
		case *ast.ArrayType:
			return tainted(pkg, f, imports, x.Elt, shadow, visiting)
		case *ast.MapType:
			if d := tainted(pkg, f, imports, x.Key, shadow, visiting); d != "" {
				return d
			}
			return tainted(pkg, f, imports, x.Value, shadow, visiting)
		case *ast.ChanType:
			return tainted(pkg, f, imports, x.Value, shadow, visiting)
		case *ast.Ellipsis:
			return tainted(pkg, f, imports, x.Elt, shadow, visiting)
		case *ast.UnaryExpr:
			return tainted(pkg, f, imports, x.X, shadow, visiting)
		case *ast.BinaryExpr:
			if d := tainted(pkg, f, imports, x.X, shadow, visiting); d != "" {
				return d
			}
			return tainted(pkg, f, imports, x.Y, shadow, visiting)
		case *ast.FuncType:
			local := copyNames(shadow)
			addFieldNames(local, x.TypeParams)
			if d := taintedTypeParamConstraints(pkg, f, imports, x.TypeParams, local, visiting, tainted); d != "" {
				return d
			}
			if x.Params != nil {
				for _, fl := range x.Params.List {
					if d := tainted(pkg, f, imports, fl.Type, local, visiting); d != "" {
						return d
					}
				}
			}
			if x.Results != nil {
				for _, fl := range x.Results.List {
					if d := tainted(pkg, f, imports, fl.Type, local, visiting); d != "" {
						return d
					}
				}
			}
		case *ast.StructType:
			for _, fl := range x.Fields.List {
				if d := tainted(pkg, f, imports, fl.Type, shadow, visiting); d != "" {
					return d
				}
			}
		case *ast.InterfaceType:
			for _, fl := range x.Methods.List {
				if d := tainted(pkg, f, imports, fl.Type, shadow, visiting); d != "" {
					return d
				}
			}
		case *ast.IndexExpr:
			if d := tainted(pkg, f, imports, x.X, shadow, visiting); d != "" {
				return d
			}
			return tainted(pkg, f, imports, x.Index, shadow, visiting)
		case *ast.IndexListExpr:
			if d := tainted(pkg, f, imports, x.X, shadow, visiting); d != "" {
				return d
			}
			for _, arg := range x.Indices {
				if d := tainted(pkg, f, imports, arg, shadow, visiting); d != "" {
					return d
				}
			}
		case *ast.ParenExpr:
			return tainted(pkg, f, imports, x.X, shadow, visiting)
		}
		return ""
	}
	resolve = func(pkg, name string, visiting map[string]bool) string {
		key := pkg + "." + name
		if visiting[key] {
			return ""
		}
		visiting[key] = true
		defer delete(visiting, key)
		for _, s := range packages[pkg] {
			for _, decl := range s.file.Decls {
				if g, ok := decl.(*ast.GenDecl); ok && g.Tok == token.TYPE {
					for _, spec := range g.Specs {
						ts := spec.(*ast.TypeSpec)
						if ts.Name.Name != name {
							continue
						}
						shadow := map[string]bool{}
						addFieldNames(shadow, ts.TypeParams)
						if d := taintedTypeParamConstraints(pkg, s.file, s.imports, ts.TypeParams, shadow, visiting, tainted); d != "" {
							return d
						}
						if d := tainted(pkg, s.file, s.imports, ts.Type, shadow, visiting); d != "" {
							return d
						}
						// Both value and pointer receiver methods are part of the type's exposed method sets.
						for _, methodDecl := range packages[pkg] {
							for _, md := range methodDecl.file.Decls {
								fn, ok := md.(*ast.FuncDecl)
								if !ok || fn.Recv == nil || !receiverBaseMatches(fn.Recv, name) {
									continue
								}
								methodShadow := copyNames(shadow)
								addReceiverTypeParams(methodShadow, fn.Recv)
								if d := tainted(pkg, methodDecl.file, methodDecl.imports, fn.Type, methodShadow, visiting); d != "" {
									return d
								}
							}
						}
					}
				}
			}
		}
		return ""
	}
	var out []finding
	packagePaths := make([]string, 0, len(packages))
	for packagePath := range packages {
		packagePaths = append(packagePaths, packagePath)
	}
	sort.Strings(packagePaths)
	for _, packagePath := range packagePaths {
		for _, s := range packages[packagePath] {
			var walkErr error
			report := func(expr ast.Expr, shadow map[string]bool, edge string) {
				var walk func(ast.Expr, map[string]bool)
				walk = func(e ast.Expr, scope map[string]bool) {
					if e == nil {
						return
					}
					switch x := e.(type) {
					case *ast.Ident:
						if !scope[x.Name] {
							if dep := tainted(s.importPath, s.file, s.imports, x, scope, map[string]bool{}); dep != "" {
								pos := s.fset.Position(x.Pos())
								out = append(out, finding{path: s.rel, symbol: fmt.Sprintf("%s @%d:%d", edge, pos.Line, pos.Column), rule: "indirect backend type exposure", dependency: dep})
							}
						}
					case *ast.SelectorExpr:
						if id, ok := x.X.(*ast.Ident); ok && isFrontendImport(s.imports[id.Name]) {
							if dep := tainted(s.importPath, s.file, s.imports, x, scope, map[string]bool{}); dep != "" {
								pos := s.fset.Position(x.Pos())
								out = append(out, finding{path: s.rel, symbol: fmt.Sprintf("%s @%d:%d", edge, pos.Line, pos.Column), rule: "indirect backend type exposure", dependency: dep})
							}
						}
					case *ast.StarExpr:
						walk(x.X, scope)
					case *ast.ArrayType:
						walk(x.Elt, scope)
					case *ast.MapType:
						walk(x.Key, scope)
						walk(x.Value, scope)
					case *ast.ChanType:
						walk(x.Value, scope)
					case *ast.Ellipsis:
						walk(x.Elt, scope)
					case *ast.UnaryExpr:
						walk(x.X, scope)
					case *ast.BinaryExpr:
						walk(x.X, scope)
						walk(x.Y, scope)
					case *ast.FuncType:
						local := copyNames(scope)
						addFieldNames(local, x.TypeParams)
						for _, tp := range typeParams(x.TypeParams) {
							walk(tp.Type, local)
						}
						if x.Params != nil {
							for _, f := range x.Params.List {
								walk(f.Type, local)
							}
						}
						if x.Results != nil {
							for _, f := range x.Results.List {
								walk(f.Type, local)
							}
						}
					case *ast.StructType:
						for _, f := range x.Fields.List {
							walk(f.Type, scope)
						}
					case *ast.InterfaceType:
						for _, f := range x.Methods.List {
							walk(f.Type, scope)
						}
					case *ast.IndexExpr:
						walk(x.X, scope)
						walk(x.Index, scope)
					case *ast.IndexListExpr:
						walk(x.X, scope)
						for _, arg := range x.Indices {
							walk(arg, scope)
						}
					case *ast.ParenExpr:
						walk(x.X, scope)
					default:
						walkErr = fmt.Errorf("unsupported frontend type expression %T in %s", e, s.rel)
					}
				}
				walk(expr, shadow)
			}
			for _, decl := range s.file.Decls {
				switch x := decl.(type) {
				case *ast.GenDecl:
					if x.Tok == token.TYPE {
						for _, sp := range x.Specs {
							ts := sp.(*ast.TypeSpec)
							shadow := map[string]bool{}
							addFieldNames(shadow, ts.TypeParams)
							for _, tp := range typeParams(ts.TypeParams) {
								report(tp.Type, shadow, "type "+ts.Name.Name+" constraint")
							}
							report(ts.Type, shadow, "type "+ts.Name.Name)
						}
					} else if x.Tok == token.VAR {
						for _, sp := range x.Specs {
							vs := sp.(*ast.ValueSpec)
							if vs.Type == nil {
								continue
							}
							for _, name := range vs.Names {
								report(vs.Type, map[string]bool{}, "var "+name.Name)
							}
						}
					}
				case *ast.FuncDecl:
					shadow := map[string]bool{}
					addFieldNames(shadow, x.Type.TypeParams)
					for _, tp := range typeParams(x.Type.TypeParams) {
						report(tp.Type, shadow, "func "+x.Name.Name+" constraint")
					}
					if x.Recv != nil {
						addReceiverTypeParams(shadow, x.Recv)
					}
					report(x.Type, shadow, "func "+x.Name.Name)
				}
			}
			if walkErr != nil {
				return nil, walkErr
			}
		}
	}
	return out, nil
}

func typeParams(list *ast.FieldList) []*ast.Field {
	if list == nil {
		return nil
	}
	return list.List
}
func taintedTypeParamConstraints(pkg string, f *ast.File, imports map[string]string, list *ast.FieldList, shadow, visiting map[string]bool, tainted func(string, *ast.File, map[string]string, ast.Expr, map[string]bool, map[string]bool) string) string {
	for _, param := range typeParams(list) {
		if dep := tainted(pkg, f, imports, param.Type, shadow, visiting); dep != "" {
			return dep
		}
	}
	return ""
}

func copyNames(in map[string]bool) map[string]bool {
	out := map[string]bool{}
	for n := range in {
		out[n] = true
	}
	return out
}
func addFieldNames(dst map[string]bool, fields *ast.FieldList) {
	if fields == nil {
		return
	}
	for _, f := range fields.List {
		for _, n := range f.Names {
			dst[n.Name] = true
		}
	}
}
func addReceiverTypeParams(dst map[string]bool, recv *ast.FieldList) {
	if recv == nil {
		return
	}
	for _, f := range recv.List {
		e := f.Type
		for {
			switch x := e.(type) {
			case *ast.StarExpr:
				e = x.X
			case *ast.IndexExpr:
				if id, ok := x.Index.(*ast.Ident); ok {
					dst[id.Name] = true
				}
				e = x.X
			case *ast.IndexListExpr:
				for _, arg := range x.Indices {
					if id, ok := arg.(*ast.Ident); ok {
						dst[id.Name] = true
					}
				}
				e = x.X
			default:
				goto done
			}
		}
	done:
	}
}
func receiverBaseMatches(recv *ast.FieldList, name string) bool {
	if recv == nil {
		return false
	}
	for _, f := range recv.List {
		e := f.Type
		for {
			switch x := e.(type) {
			case *ast.StarExpr:
				e = x.X
			case *ast.IndexExpr:
				e = x.X
			case *ast.IndexListExpr:
				e = x.X
			default:
				goto base
			}
		}
	base:
		if id, ok := e.(*ast.Ident); ok && id.Name == name {
			return true
		}
	}
	return false
}

func symbolsMatch(file *ast.File, imp *ast.ImportSpec, dependency string, expected []string) ([]string, bool) {
	name := ""
	if imp.Name != nil {
		name = imp.Name.Name
	}
	if name == "" {
		name = filepath.Base(dependency)
	}
	if name == "." || name == "_" {
		return nil, false
	}
	gotSet := map[string]bool{}
	ast.Inspect(file, func(n ast.Node) bool {
		sel, ok := n.(*ast.SelectorExpr)
		if !ok {
			return true
		}
		id, ok := sel.X.(*ast.Ident)
		if ok && id.Name == name && (id.Obj == nil || id.Obj.Kind == ast.Pkg) {
			gotSet[sel.Sel.Name] = true
		}
		return true
	})
	got := make([]string, 0, len(gotSet))
	for s := range gotSet {
		got = append(got, s)
	}
	sort.Strings(got)
	want := append([]string(nil), expected...)
	sort.Strings(want)
	return got, strings.Join(got, "\x00") == strings.Join(want, "\x00")
}

func declarationEdges(file *ast.File, imp *ast.ImportSpec, dependency string, expected []string) ([]string, bool) {
	name := ""
	if imp.Name != nil {
		name = imp.Name.Name
	}
	if name == "" {
		name = filepath.Base(dependency)
	}
	if name == "." || name == "_" {
		return nil, false
	}
	set := map[string]bool{}
	var ancestors []ast.Node
	fset := token.NewFileSet()
	ast.Inspect(file, func(n ast.Node) bool {
		if n == nil {
			ancestors = ancestors[:len(ancestors)-1]
			return false
		}
		ancestors = append(ancestors, n)
		sel, ok := n.(*ast.SelectorExpr)
		if !ok {
			return true
		}
		id, ok := sel.X.(*ast.Ident)
		if !ok || id.Name != name || (id.Obj != nil && id.Obj.Kind != ast.Pkg) {
			return true
		}
		set[declarationEdgeKey(fset, ancestors, sel)] = true
		return true
	})
	got := make([]string, 0, len(set))
	for edge := range set {
		got = append(got, edge)
	}
	sort.Strings(got)
	want := append([]string(nil), expected...)
	sort.Strings(want)
	return got, strings.Join(got, "\x00") == strings.Join(want, "\x00")
}

func declarationEdgeKey(fset *token.FileSet, ancestors []ast.Node, selector *ast.SelectorExpr) string {
	var typeSpec *ast.TypeSpec
	var funcDecl *ast.FuncDecl
	var field *ast.Field
	var valueSpec *ast.ValueSpec
	var fieldIndex, typeIndex, funcIndex int
	for i := len(ancestors) - 1; i >= 0; i-- {
		switch n := ancestors[i].(type) {
		case *ast.Field:
			if field == nil {
				field = n
				fieldIndex = i
			}
		case *ast.TypeSpec:
			if typeSpec == nil {
				typeSpec = n
				typeIndex = i
			}
		case *ast.FuncDecl:
			if funcDecl == nil {
				funcDecl = n
				funcIndex = i
			}
		case *ast.ValueSpec:
			if valueSpec == nil {
				valueSpec = n
			}
		}
	}
	selectorText := nodeText(fset, selector)
	if typeSpec != nil {
		owner := "type " + typeSpec.Name.Name
		if field != nil && fieldIndex > typeIndex {
			methodContext := ""
			for i := typeIndex + 1; i < fieldIndex; i++ {
				if parent, ok := ancestors[i].(*ast.Field); ok && parent.Type != nil {
					if ft, ok := parent.Type.(*ast.FuncType); ok && len(parent.Names) > 0 {
						methodContext = " method " + parent.Names[0].Name
						_ = ft
						break
					}
				}
			}
			if field.Type != nil {
				fieldNames := make([]string, len(field.Names))
				for i, n := range field.Names {
					fieldNames[i] = n.Name
				}
				if len(fieldNames) == 0 {
					fieldNames = []string{"<embedded>"}
				}
				role := "field " + strings.Join(fieldNames, ",")
				for i := fieldIndex - 1; i > typeIndex; i-- {
					if list, ok := ancestors[i].(*ast.FieldList); ok {
						if i > 0 {
							if _, isFunc := ancestors[i-1].(*ast.FuncType); isFunc {
								if methodContext != "" {
									role = "signature " + role
								} else if list.List == nil {
									role = "signature " + role
								}
							}
						}
						break
					}
				}
				return owner + methodContext + " " + role + " " + nodeText(fset, field.Type) + " uses " + selectorText
			}
		}
		return owner + " declaration " + nodeText(fset, typeSpec.Type) + " uses " + selectorText
	}
	if funcDecl != nil {
		funcName := "func "
		if funcDecl.Recv != nil && len(funcDecl.Recv.List) > 0 {
			funcName += "receiver " + nodeText(fset, funcDecl.Recv.List[0].Type) + " "
		}
		funcName += funcDecl.Name.Name
		if field != nil && fieldIndex > funcIndex {
			role := "parameter"
			for i := fieldIndex - 1; i > funcIndex; i-- {
				if fl, ok := ancestors[i].(*ast.FieldList); ok {
					if funcDecl.Type.Results == fl {
						role = "result"
					}
					break
				}
			}
			fieldName := "<unnamed>"
			if len(field.Names) > 0 {
				fieldName = field.Names[0].Name
			}
			return funcName + " " + role + " " + fieldName + " " + nodeText(fset, field.Type) + " uses " + selectorText
		}
		return funcName + " body uses " + selectorText
	}
	if valueSpec != nil {
		kind := "var"
		if valueSpec.Type == nil {
			kind = "value"
		}
		names := make([]string, len(valueSpec.Names))
		for i, id := range valueSpec.Names {
			names[i] = id.Name
		}
		if len(names) == 0 {
			names = []string{"<unnamed>"}
		}
		return kind + " " + strings.Join(names, ",") + " " + nodeText(fset, valueSpec.Type) + " uses " + selectorText
	}
	return "file declaration uses " + selectorText
}

func nodeText(fset *token.FileSet, node ast.Node) string {
	var b strings.Builder
	if err := format.Node(&b, fset, node); err != nil {
		return "<unformattable>"
	}
	return b.String()
}

func strconvUnquote(s string) (string, error) { return strconv.Unquote(s) }

func directFinding(rel, dep string) *finding {
	if strings.HasSuffix(rel, "_test.go") {
		return nil
	}
	if isFrontend(rel) {
		if strings.HasPrefix(rel, "internal/app/") && importWithin(dep, "internal/ui") {
			return nil
		}
		if isContractPackagePath(dep) {
			return nil
		}
		if forbiddenFrontendImport(dep) {
			return &finding{path: rel, symbol: "import", rule: "frontend imports backend implementation", dependency: dep}
		}
		if strings.HasPrefix(rel, "internal/ui/") && importWithin(dep, "internal/app") {
			return &finding{path: rel, symbol: "import", rule: "child frontend imports root coordinator", dependency: dep}
		}
	}
	if strings.HasPrefix(rel, "internal/api/contracts/") {
		if (strings.HasPrefix(dep, module+"internal/") && !isContractPackagePath(dep)) || forbiddenFramework(dep) {
			return &finding{path: rel, symbol: "import", rule: "contracts import implementation or presentation", dependency: dep}
		}
	}
	if !isFrontend(rel) && !strings.HasPrefix(rel, "internal/api/contracts/") {
		if isFrontendImport(dep) || forbiddenFramework(dep) {
			return &finding{path: rel, symbol: "import", rule: "backend imports frontend", dependency: dep}
		}
	}
	return nil
}

func forbiddenFrontendImport(dep string) bool {
	if !strings.HasPrefix(dep, module+"internal/") {
		return false
	}
	part := strings.TrimPrefix(dep, module+"internal/")
	for _, name := range []string{"api", "config", "configstartup", "session", "driver", "mcp"} {
		if part == name || strings.HasPrefix(part, name+"/") {
			return true
		}
	}
	return false
}

func forbiddenFramework(dep string) bool {
	return strings.HasPrefix(dep, "charm.land/bubbletea") || strings.HasPrefix(dep, "charm.land/lipgloss") || strings.HasPrefix(dep, "charm.land/bubbles") || strings.HasPrefix(dep, "github.com/charmbracelet/")
}
func isFrontend(path string) bool {
	return strings.HasPrefix(path, "internal/app/") || strings.HasPrefix(path, "internal/ui/")
}

func testSupportIsSafe(rel string, files []string, imports map[string][]string, root string) bool {
	if !strings.HasPrefix(rel, "internal/ui/uitest/") {
		return false
	}
	for _, path := range files {
		r, _ := filepath.Rel(root, path)
		r = filepath.ToSlash(r)
		if strings.HasSuffix(r, "_test.go") {
			continue
		}
		for _, d := range imports[r] {
			if d == module+"internal/ui/uitest" || strings.HasPrefix(d, module+"internal/ui/uitest/") {
				return false
			}
		}
	}
	return true
}

func dependencyGraphFindings(root, importPath, chain string, seen map[string]bool) ([]finding, error) {
	if !strings.HasPrefix(importPath, module+"internal/") {
		return nil, nil
	}
	if isContractPackagePath(importPath) || isFrontendImport(importPath) || forbiddenFrontendImport(importPath) {
		return nil, nil
	}
	if seen[importPath] {
		return nil, nil
	}
	seen[importPath] = true
	rel := strings.TrimPrefix(importPath, module)
	dir := filepath.Join(root, filepath.FromSlash(rel))
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("read shared dependency %s: %w", rel, err)
	}
	var out []finding
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") || strings.HasSuffix(entry.Name(), "_test.go") {
			continue
		}
		path := filepath.Join(dir, entry.Name())
		file, err := parser.ParseFile(token.NewFileSet(), path, nil, parser.ParseComments)
		if err != nil {
			return nil, fmt.Errorf("parse shared dependency %s: %w", filepath.ToSlash(path), err)
		}
		for _, im := range file.Imports {
			dep, err := strconv.Unquote(im.Path.Value)
			if err != nil {
				return nil, fmt.Errorf("%s: %w", path, err)
			}
			if isContractPackagePath(dep) {
				continue
			}
			if forbiddenFrontendImport(dep) {
				out = append(out, finding{path: filepath.ToSlash(filepath.Join(rel, entry.Name())), symbol: "import", rule: "transitive frontend dependency", dependency: chain + " -> " + dep})
				continue
			}
			if isFrontendImport(dep) {
				out = append(out, finding{path: filepath.ToSlash(filepath.Join(rel, entry.Name())), symbol: "import", rule: "shared dependency imports frontend", dependency: chain + " -> " + dep})
				continue
			}
			if forbiddenFramework(dep) {
				out = append(out, finding{path: filepath.ToSlash(filepath.Join(rel, entry.Name())), symbol: "import", rule: "transitive frontend presentation dependency", dependency: chain + " -> " + dep})
				continue
			}
			if strings.HasPrefix(dep, module+"internal/") && !isContractPackagePath(dep) {
				more, e := dependencyGraphFindings(root, dep, chain+" -> "+dep, seen)
				if e != nil {
					return nil, e
				}
				out = append(out, more...)
			}
		}
	}
	return out, nil
}

func isFrontendImport(path string) bool {
	return importWithin(path, "internal/app") || importWithin(path, "internal/ui")
}

func importWithin(path, packagePath string) bool {
	base := module + packagePath
	return path == base || strings.HasPrefix(path, base+"/")
}

func contractFindings(root string, files []string) ([]finding, error) {
	dirSet := map[string]bool{}
	for _, path := range files {
		rel, _ := filepath.Rel(root, path)
		rel = filepath.ToSlash(rel)
		if strings.HasSuffix(rel, "_test.go") {
			continue
		}
		if rel == "internal/api/contracts" || strings.HasPrefix(rel, "internal/api/contracts/") {
			dirSet[filepath.Dir(path)] = true
		}
	}
	if len(dirSet) == 0 {
		return nil, fmt.Errorf("contracts have no production Go files")
	}
	dirs := make([]string, 0, len(dirSet))
	for dir := range dirSet {
		dirs = append(dirs, dir)
	}
	sort.Strings(dirs)
	contexts, err := platformContexts(dirs...)
	if err != nil {
		return nil, fmt.Errorf("discover contract contexts: %w", err)
	}
	resolver := &contractImporter{root: root, cache: map[string]*types.Package{}, loading: map[string]bool{}, context: build.Default}
	var out []finding
	for _, dir := range dirs {
		rel, _ := filepath.Rel(root, dir)
		rel = filepath.ToSlash(rel)
		importPath := module + rel
		seen := map[string]bool{}
		for _, ctx := range contexts {
			files, err := selectContractVariant(dir, ctx)
			if err != nil {
				return nil, err
			}
			if len(files) == 0 {
				continue
			}
			resolver.context = ctx
			key, err := resolver.effectiveVariantKey(importPath, map[string]bool{})
			if err != nil {
				return nil, fmt.Errorf("resolve contract variant %s: %w", rel, err)
			}
			if seen[key] {
				continue
			}
			seen[key] = true
			pkg, err := resolver.check(importPath, files)
			if err != nil {
				return nil, fmt.Errorf("type-check contract package %s (%s/%s tags=%v): %w", rel, ctx.GOOS, ctx.GOARCH, ctx.BuildTags, err)
			}
			for _, name := range pkg.Scope().Names() {
				if !token.IsExported(name) {
					continue
				}
				obj := pkg.Scope().Lookup(name)
				if obj == nil {
					continue
				}
				if leak := walkContractType(obj.Type(), name, map[types.Type]bool{}); leak != "" {
					out = append(out, finding{path: rel, symbol: name, rule: "contract type leaks implementation", dependency: leak})
				}
			}
		}
	}
	return out, nil
}

type contractFile struct{ name, path string }
type contractVariant struct {
	context build.Context
	files   []contractFile
}

type contractImporter struct {
	root    string
	cache   map[string]*types.Package
	loading map[string]bool
	context build.Context
}

func (i *contractImporter) Import(path string) (*types.Package, error) {
	if !isContractPackagePath(path) {
		return importer.Default().Import(path)
	}
	rel := strings.TrimPrefix(path, module)
	return i.load(filepath.Join(i.root, filepath.FromSlash(rel)), path)
}

func (i *contractImporter) load(dir, importPath string) (*types.Package, error) {
	files, err := selectContractVariant(dir, i.context)
	if err != nil {
		return nil, err
	}
	return i.check(importPath, files)
}

// effectiveVariantKey includes every local contract package selected through
// imports, so a parent is rechecked when a dependency's platform files change.
func (i *contractImporter) effectiveVariantKey(importPath string, visiting map[string]bool) (string, error) {
	if visiting[importPath] {
		return "cycle:" + importPath, nil
	}
	visiting[importPath] = true
	defer delete(visiting, importPath)
	rel := strings.TrimPrefix(importPath, module)
	dir := filepath.Join(i.root, filepath.FromSlash(rel))
	files, err := selectContractVariant(dir, i.context)
	if err != nil {
		return "", err
	}
	parts := []string{fmt.Sprintf("cgo=%t", i.context.CgoEnabled), importPath + "=" + contractVariantKey(files)}
	for _, file := range files {
		f, err := parser.ParseFile(token.NewFileSet(), file.path, nil, 0)
		if err != nil {
			return "", err
		}
		for _, im := range f.Imports {
			dep, err := strconv.Unquote(im.Path.Value)
			if err != nil {
				return "", err
			}
			if !isContractPackagePath(dep) {
				continue
			}
			child, err := i.effectiveVariantKey(dep, visiting)
			if err != nil {
				return "", err
			}
			parts = append(parts, child)
		}
	}
	sort.Strings(parts[1:])
	return strings.Join(parts, "\x00"), nil
}

func (i *contractImporter) check(importPath string, files []contractFile) (*types.Package, error) {
	variant, err := i.effectiveVariantKey(importPath, map[string]bool{})
	if err != nil {
		return nil, err
	}
	key := importPath + "|" + variant
	if pkg := i.cache[key]; pkg != nil {
		return pkg, nil
	}
	if i.loading[key] {
		return nil, fmt.Errorf("contract package import cycle at %s", importPath)
	}
	i.loading[key] = true
	defer delete(i.loading, key)
	if len(files) == 0 {
		return nil, fmt.Errorf("no compatible production files for %s", importPath)
	}
	fset := token.NewFileSet()
	asts := make([]*ast.File, 0, len(files))
	pkgName := ""
	for _, source := range files {
		file, err := parser.ParseFile(fset, source.path, nil, parser.ParseComments)
		if err != nil {
			return nil, err
		}
		if pkgName == "" {
			pkgName = file.Name.Name
		} else if pkgName != file.Name.Name {
			return nil, fmt.Errorf("mutually compatible production package declarations conflict: %s and %s", pkgName, file.Name.Name)
		}
		asts = append(asts, file)
	}
	pkg, err := (&types.Config{Importer: i}).Check(importPath, fset, asts, nil)
	if err != nil {
		return nil, err
	}
	i.cache[key] = pkg
	return pkg, nil
}

func contractVariants(dir string) ([]contractVariant, error) {
	contexts, err := platformContexts(dir)
	if err != nil {
		return nil, err
	}
	seen := map[string]bool{}
	var out []contractVariant
	for _, ctx := range contexts {
		files, err := selectContractVariant(dir, ctx)
		if err != nil {
			return nil, err
		}
		if len(files) == 0 {
			continue
		}
		key := fmt.Sprintf("cgo=%t|%s", ctx.CgoEnabled, contractVariantKey(files))
		if !seen[key] {
			seen[key] = true
			out = append(out, contractVariant{context: ctx, files: files})
		}
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("no compatible production Go files")
	}
	return out, nil
}

func selectContractVariant(dir string, ctx build.Context) ([]contractFile, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", dir, err)
	}
	var files []contractFile
	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}
		match, err := ctx.MatchFile(dir, name)
		if err != nil {
			return nil, fmt.Errorf("match %s: %w", name, err)
		}
		if match {
			files = append(files, contractFile{name: name, path: filepath.Join(dir, name)})
		}
	}
	sort.Slice(files, func(a, b int) bool { return files[a].name < files[b].name })
	return files, nil
}

func contractVariantKey(files []contractFile) string {
	names := make([]string, len(files))
	for i, file := range files {
		names[i] = file.name
	}
	return strings.Join(names, "\x00")
}

func platformContexts(dirs ...string) ([]build.Context, error) {
	// Supported pairs mirror `go tool dist list` for the Go toolchain in go.mod;
	// keep this list in sync when that toolchain changes.
	pairs := strings.Fields(`aix/ppc64 android/386 android/amd64 android/arm android/arm64 darwin/amd64 darwin/arm64 dragonfly/amd64 freebsd/386 freebsd/amd64 freebsd/arm freebsd/arm64 illumos/amd64 ios/amd64 ios/arm64 js/wasm linux/386 linux/amd64 linux/arm linux/arm64 linux/loong64 linux/mips linux/mips64 linux/mips64le linux/mipsle linux/ppc64 linux/ppc64le linux/riscv64 linux/s390x netbsd/386 netbsd/amd64 netbsd/arm netbsd/arm64 openbsd/386 openbsd/amd64 openbsd/arm openbsd/arm64 openbsd/ppc64 openbsd/riscv64 plan9/386 plan9/amd64 plan9/arm solaris/amd64 wasip1/wasm windows/386 windows/amd64 windows/arm64`)
	base := build.Default
	out := []build.Context{}
	gooses, arches := map[string]bool{}, map[string]bool{}
	for _, pair := range pairs {
		parts := strings.Split(pair, "/")
		gooses[parts[0]] = true
		arches[parts[1]] = true
	}
	cgoStates := []bool{false, true}
	for _, enabled := range cgoStates {
		ctx := base
		ctx.CgoEnabled = enabled
		out = append(out, ctx)
	}
	tagSet := map[string]bool{}
	for _, dir := range dirs {
		entries, err := os.ReadDir(dir)
		if err != nil {
			return nil, err
		}
		for _, entry := range entries {
			if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") || strings.HasSuffix(entry.Name(), "_test.go") {
				continue
			}
			data, err := os.ReadFile(filepath.Join(dir, entry.Name()))
			if err != nil {
				return nil, err
			}
			for _, line := range strings.Split(string(data), "\n") {
				line = strings.TrimSpace(line)
				if !strings.HasPrefix(line, "//go:build ") && !strings.HasPrefix(line, "// +build ") {
					continue
				}
				expr, err := constraint.Parse(line)
				if err != nil {
					return nil, fmt.Errorf("parse build constraint %s: %w", entry.Name(), err)
				}
				collectConstraintTags(expr, tagSet)
			}
		}
	}
	var custom []string
	known := map[string]bool{"unix": true, "gc": true, "gccgo": true, "cgo": true}
	for goos := range gooses {
		known[goos] = true
	}
	for arch := range arches {
		known[arch] = true
	}
	for tag := range tagSet {
		if !known[tag] && !strings.HasPrefix(tag, "go1.") {
			custom = append(custom, tag)
		}
	}
	sort.Strings(custom)
	if len(custom) > 8 {
		return nil, fmt.Errorf("cannot exhaustively resolve %d custom build tags across contract packages", len(custom))
	}
	assignments := 1 << len(custom)
	for bits := 0; bits < assignments; bits++ {
		tags := []string{}
		for i, tag := range custom {
			if bits&(1<<i) != 0 {
				tags = append(tags, tag)
			}
		}
		for _, pair := range pairs {
			parts := strings.Split(pair, "/")
			for _, enabled := range cgoStates {
				ctx := base
				ctx.GOOS = parts[0]
				ctx.GOARCH = parts[1]
				ctx.CgoEnabled = enabled
				ctx.BuildTags = append(append([]string(nil), base.BuildTags...), tags...)
				out = append(out, ctx)
			}
		}
	}
	return out, nil
}

func collectConstraintTags(expr constraint.Expr, set map[string]bool) {
	switch x := expr.(type) {
	case *constraint.TagExpr:
		set[x.Tag] = true
	case *constraint.NotExpr:
		collectConstraintTags(x.X, set)
	case *constraint.AndExpr:
		collectConstraintTags(x.X, set)
		collectConstraintTags(x.Y, set)
	case *constraint.OrExpr:
		collectConstraintTags(x.X, set)
		collectConstraintTags(x.Y, set)
	}
}

func walkContractType(t types.Type, chain string, seen map[types.Type]bool) string {
	if t == nil || seen[t] {
		return ""
	}
	seen[t] = true
	switch x := t.(type) {
	case *types.Alias:
		if x.Obj().Pkg() != nil && forbiddenPackagePath(x.Obj().Pkg().Path()) {
			return chain + " -> " + x.Obj().Pkg().Path() + "." + x.Obj().Name()
		}
		return walkContractType(types.Unalias(x), chain+".target", seen)
	case *types.Named:
		if x.Obj().Pkg() != nil && forbiddenPackagePath(x.Obj().Pkg().Path()) {
			return chain + " -> " + x.Obj().Pkg().Path() + "." + x.Obj().Name()
		}
		for i := 0; i < x.TypeArgs().Len(); i++ {
			if leak := walkContractType(x.TypeArgs().At(i), chain+".type-arg", seen); leak != "" {
				return leak
			}
		}
		if x.TypeParams() != nil {
			for i := 0; i < x.TypeParams().Len(); i++ {
				if leak := walkContractType(x.TypeParams().At(i), chain+".type-param", seen); leak != "" {
					return leak
				}
			}
		}
		if leak := walkContractType(x.Underlying(), chain+".underlying", seen); leak != "" {
			return leak
		}
		for i := 0; i < x.NumMethods(); i++ {
			if leak := walkContractType(x.Method(i).Type(), chain+".method "+x.Method(i).Name(), seen); leak != "" {
				return leak
			}
		}
	case *types.Pointer:
		return walkContractType(x.Elem(), chain+".*", seen)
	case *types.Slice:
		return walkContractType(x.Elem(), chain+".[]", seen)
	case *types.Array:
		return walkContractType(x.Elem(), chain+".[]", seen)
	case *types.Chan:
		return walkContractType(x.Elem(), chain+".chan", seen)
	case *types.Map:
		if leak := walkContractType(x.Key(), chain+".key", seen); leak != "" {
			return leak
		}
		return walkContractType(x.Elem(), chain+".value", seen)
	case *types.Struct:
		for i := 0; i < x.NumFields(); i++ {
			if leak := walkContractType(x.Field(i).Type(), chain+"."+x.Field(i).Name(), seen); leak != "" {
				return leak
			}
		}
	case *types.Signature:
		if x.TypeParams() != nil {
			for i := 0; i < x.TypeParams().Len(); i++ {
				if leak := walkContractType(x.TypeParams().At(i), chain+".type-param", seen); leak != "" {
					return leak
				}
			}
		}
		if x.RecvTypeParams() != nil {
			for i := 0; i < x.RecvTypeParams().Len(); i++ {
				if leak := walkContractType(x.RecvTypeParams().At(i), chain+".receiver-type-param", seen); leak != "" {
					return leak
				}
			}
		}
		if leak := walkContractType(x.Params(), chain+".params", seen); leak != "" {
			return leak
		}
		return walkContractType(x.Results(), chain+".results", seen)
	case *types.Tuple:
		for i := 0; i < x.Len(); i++ {
			if leak := walkContractType(x.At(i).Type(), chain+".result "+x.At(i).Name(), seen); leak != "" {
				return leak
			}
		}
	case *types.Interface:
		x.Complete()
		for i := 0; i < x.NumEmbeddeds(); i++ {
			if leak := walkContractType(x.EmbeddedType(i), chain+".embed", seen); leak != "" {
				return leak
			}
		}
		for i := 0; i < x.NumMethods(); i++ {
			if leak := walkContractType(x.Method(i).Type(), chain+".method "+x.Method(i).Name(), seen); leak != "" {
				return leak
			}
		}
	case *types.TypeParam:
		return walkContractType(x.Constraint(), chain+".constraint", seen)
	case *types.Union:
		for i := 0; i < x.Len(); i++ {
			if leak := walkContractType(x.Term(i).Type(), chain+".term", seen); leak != "" {
				return leak
			}
		}
	}
	return ""
}

func forbiddenPackagePath(path string) bool {
	if isContractPackagePath(path) {
		return false
	}
	if strings.HasPrefix(path, module+"internal/") {
		return true
	}
	return forbiddenFramework(path)
}

func isContractPackagePath(path string) bool {
	base := module + "internal/api/contracts"
	return path == base || strings.HasPrefix(path, base+"/")
}
