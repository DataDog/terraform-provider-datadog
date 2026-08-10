//go:build integration

// This test corroborates the derivation in bind.go/gotype.go against the pinned
// datadog-api-client-go, and is the one place in this package that reads generated
// SDK source. It exists because a derivation fails differently from a lookup: a
// lookup that misses returns nothing, while a wrong derivation returns a plausible
// name for a symbol that does not exist,
//
// It is corroboration, never the source of a name and it reads declarations with
// go/ast only.
//
// Behind -tags=integration because it shells out to `go list -m` and parses a few
// thousand generated files, neither of which belongs in the default unit run.
package sdkbind

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.yaml.in/yaml/v4"

	"github.com/terraform-providers/terraform-provider-datadog/generator/internal/model"
	tfparser "github.com/terraform-providers/terraform-provider-datadog/generator/internal/parser"
)

// corpusDir holds the checked-in mini-OAS slices: real Datadog response shapes,
// unannotated, which this test annotates in memory to opt each read operation in.
const corpusDir = "../testdata/mini-oas"

// providerRoot is the provider module, whose go.mod pins the SDK this generator's
// output must compile against.
const providerRoot = "../../.."

// packageDecls is what one generated SDK package declares, read from source.
type packageDecls struct {
	// structFields maps a struct type name to its field names.
	structFields map[string][]string
	// funcs holds every top-level function name.
	funcs map[string]bool
}

func (d packageDecls) hasField(structName, field string) bool {
	for _, f := range d.structFields[structName] {
		if f == field {
			return true
		}
	}
	return false
}

// sdkPackageDecls parses <sdk>/api/<pkg> and returns its declarations. It reads
// only exported declarations' names, never types, so nothing here depends on the
// SDK being linkable.
func sdkPackageDecls(pkg string) (packageDecls, error) {
	sdkDir, err := sdkModuleDir()
	if err != nil {
		return packageDecls{}, err
	}
	dir := filepath.Join(sdkDir, "api", pkg)
	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, dir, nil, parser.SkipObjectResolution)
	if err != nil {
		return packageDecls{}, err
	}

	decls := packageDecls{
		structFields: make(map[string][]string),
		funcs:        make(map[string]bool),
	}
	for _, p := range pkgs {
		for _, file := range p.Files {
			for _, d := range file.Decls {
				switch decl := d.(type) {
				case *ast.FuncDecl:
					if decl.Recv == nil {
						decls.funcs[decl.Name.Name] = true
					}
				case *ast.GenDecl:
					for _, spec := range decl.Specs {
						ts, ok := spec.(*ast.TypeSpec)
						if !ok {
							continue
						}
						st, ok := ts.Type.(*ast.StructType)
						if !ok {
							continue
						}
						var fields []string
						for _, f := range st.Fields.List {
							for _, name := range f.Names {
								fields = append(fields, name.Name)
							}
						}
						decls.structFields[ts.Name.Name] = fields
					}
				}
			}
		}
	}
	return decls, nil
}

// sdkModuleDir resolves the pinned SDK's module directory through the provider
// module, so the version read is exactly the one the provider builds against.
func sdkModuleDir() (string, error) {
	cmd := exec.Command("go", "list", "-m", "-f", "{{.Dir}}", "github.com/DataDog/datadog-api-client-go/v2")
	cmd.Dir = providerRoot
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

// annotatedCorpus loads every mini-OAS slice with its by-id (else first) GET
// annotated as a singular data source, and binds it. It returns the operations
// that carry at least one union, keyed by "<artifact>:<operationId>".
func annotatedCorpus() map[string]*model.Operation {
	GinkgoHelper()
	entries, err := os.ReadDir(corpusDir)
	Expect(err).To(Succeed())

	tmp := GinkgoT().TempDir()
	withUnions := make(map[string]*model.Operation)

	for _, e := range entries {
		name, ok := strings.CutPrefix(e.Name(), "mini-datadog_")
		if !ok || !strings.HasSuffix(name, ".yaml") {
			continue
		}
		artifact := strings.TrimSuffix(name, ".yaml")

		annotated, ok := annotateReadOp(filepath.Join(corpusDir, e.Name()), artifact, tmp)
		if !ok {
			continue
		}
		spec, err := tfparser.LoadSpec(annotated)
		if err != nil {
			// A slice the parser rejects outright (e.g. a $ref cycle) has no bindings
			// to corroborate; the corpus-level failures are tracked as their own tasks.
			continue
		}
		for _, op := range spec.Operations {
			if op.Tracking == nil {
				continue
			}
			_ = BindOperation(op) // failures are asserted separately; bound unions still count
			if len(collectUnions(op)) > 0 {
				withUnions[artifact+":"+op.OperationId] = op
			}
		}
	}
	return withUnions
}

// annotateReadOp writes a copy of the slice at src with its read operation opted
// in, and returns the copy's path. It mirrors the README's documented annotation.
func annotateReadOp(src, artifact, tmpDir string) (string, bool) {
	GinkgoHelper()
	raw, err := os.ReadFile(src)
	Expect(err).To(Succeed())

	var doc yaml.Node
	Expect(yaml.Unmarshal(raw, &doc)).To(Succeed())

	var spec map[string]any
	Expect(yaml.Unmarshal(raw, &spec)).To(Succeed())

	paths, ok := spec["paths"].(map[string]any)
	if !ok {
		return "", false
	}
	// Prefer a by-id GET (a path with a {param}); that is the singular read.
	var chosenPath string
	var chosenOp map[string]any
	for _, p := range sortedMapKeys(paths) {
		item, ok := paths[p].(map[string]any)
		if !ok {
			continue
		}
		get, ok := item["get"].(map[string]any)
		if !ok || get["operationId"] == nil {
			continue
		}
		if chosenOp == nil || (!strings.Contains(chosenPath, "{") && strings.Contains(p, "{")) {
			chosenPath, chosenOp = p, get
		}
	}
	if chosenOp == nil {
		return "", false
	}
	chosenOp["x-datadog-tf-generator"] = map[string]any{
		"artifact_kind": "data_source",
		"artifact_name": artifact,
		"group":         map[string]any{"read": chosenOp["operationId"]},
	}

	out, err := yaml.Marshal(spec)
	Expect(err).To(Succeed())
	dst := filepath.Join(tmpDir, artifact+".yaml")
	Expect(os.WriteFile(dst, out, 0o600)).To(Succeed())
	return dst, true
}

func sortedMapKeys(m map[string]any) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// collectUnions returns every union reachable from op's request and response
// schemas, keyed by schema path.
func collectUnions(op *model.Operation) map[string]*model.OneOfSpec {
	found := make(map[string]*model.OneOfSpec)
	seen := make(map[*model.Schema]bool)
	var walk func(s *model.Schema)
	walk = func(s *model.Schema) {
		if s == nil || seen[s] {
			return
		}
		seen[s] = true
		if s.Kind == model.SchemaKindOneOf && s.OneOf != nil {
			found[s.OneOf.Path] = s.OneOf
			for _, v := range s.OneOf.Variants {
				walk(v.Schema)
			}
			return
		}
		for _, child := range s.Properties {
			walk(child)
		}
		walk(s.Items)
	}
	walk(op.RequestSchema)
	walk(op.ResponseSchema)
	return found
}

var _ = Describe("SDK oneOf binding, corroborated against the pinned SDK", func() {

	It("resolves every union in the mini-OAS corpus to a wrapper, member and constructor that exist", func() {
		corpus := annotatedCorpus()
		Expect(corpus).NotTo(BeEmpty(), "no corpus slice produced a union to corroborate")

		declsByPackage := make(map[string]packageDecls)
		var problems []string
		var checked int

		for key, op := range corpus {
			pkg := model.SDKPackageForPath(op.Path)
			decls, ok := declsByPackage[pkg]
			if !ok {
				var err error
				decls, err = sdkPackageDecls(pkg)
				Expect(err).To(Succeed(), "reading SDK package %s", pkg)
				declsByPackage[pkg] = decls
			}

			for path, union := range collectUnions(op) {
				if union.SDKType == "" {
					// An unresolved union is reported by BindOperation as a diagnostic, not
					// a wrong name; it is out of scope for a name-existence check.
					continue
				}
				checked++
				if _, exists := decls.structFields[union.SDKType]; !exists {
					problems = append(problems, key+" "+path+
						": derived wrapper "+pkg+"."+union.SDKType+" is not declared in the SDK")
					continue
				}
				if !decls.hasField(union.SDKType, "UnparsedObject") {
					problems = append(problems, key+" "+path+
						": derived wrapper "+union.SDKType+" is not a oneOf wrapper (no UnparsedObject member)")
				}
				for _, v := range union.Variants {
					if v.SDKField == "" {
						continue
					}
					if !decls.hasField(union.SDKType, v.SDKField) {
						problems = append(problems, key+" "+path+" variant "+v.TFName+
							": derived member "+union.SDKType+"."+v.SDKField+" is not declared")
					}
					if !decls.funcs[v.SDKConstructor] {
						problems = append(problems, key+" "+path+" variant "+v.TFName+
							": derived constructor "+pkg+"."+v.SDKConstructor+" is not declared")
					}
				}
			}
		}

		// Reported so a shrinking corpus is visible rather than silently vacuous: a
		// corroboration that checks nothing passes just as loudly as one that checks
		// everything.
		AddReportEntry("unions corroborated", checked)
		AddReportEntry("corpus operations with unions", len(corpus))

		Expect(checked).To(BeNumerically(">", 0), "no bound union was corroborated")
		Expect(problems).To(BeEmpty(), "derived SDK bindings that do not exist in the pinned SDK")
	})

	It("binds every alternative of action_connection's integration union", func() {
		// The acceptance bar: 25 unions, and every outer alternative
		// resolving both a member and a constructor.
		corpus := annotatedCorpus()
		var op *model.Operation
		for key, candidate := range corpus {
			if strings.HasPrefix(key, "action_connection:") {
				op = candidate
				break
			}
		}
		Expect(op).NotTo(BeNil(), "action_connection produced no union")

		unions := collectUnions(op)
		Expect(len(unions)).To(BeNumerically(">=", 25),
			"expected the integration union plus one credentials union per integration variant")

		integration := unions["response.data.attributes.integration"]
		Expect(integration).NotTo(BeNil())
		Expect(integration.SDKType).To(Equal("ActionConnectionIntegration"))
		Expect(integration.Variants).To(HaveLen(24))
		for _, v := range integration.Variants {
			Expect(v.SDKField).NotTo(BeEmpty(), "variant %s bound no SDK member", v.TFName)
			Expect(v.SDKConstructor).To(Equal(v.SDKField+"AsActionConnectionIntegration"),
				"variant %s", v.TFName)
		}
	})
})
