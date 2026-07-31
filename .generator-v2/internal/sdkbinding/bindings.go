// Package sdkbinding resolves OpenAPI operation parameters against the pinned
// datadog-api-client-go method signatures used by generated provider code.
package sdkbinding

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/terraform-providers/terraform-provider-datadog/generator/internal/model"
)

type methodSignature struct {
	arguments []argumentSignature
	options   string
}

type argumentSignature struct {
	name     string
	typeName string
}

// Inventory is the set of SDK methods and option setters discovered in the
// pinned datadogV2 package.
type Inventory struct {
	methods map[string]methodSignature
	setters map[string]map[string]argumentSignature
}

// Load parses the generated SDK API files in dir. Parsing source instead of
// deriving signatures from OpenAPI keeps tfgen aligned with the exact client
// version the provider compiles against.
func Load(dir string) (*Inventory, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("read SDK API package %s: %w", dir, err)
	}
	inv := &Inventory{
		methods: map[string]methodSignature{},
		setters: map[string]map[string]argumentSignature{},
	}
	fset := token.NewFileSet()
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasPrefix(entry.Name(), "api_") || !strings.HasSuffix(entry.Name(), ".go") || strings.HasSuffix(entry.Name(), "_test.go") {
			continue
		}
		path := filepath.Join(dir, entry.Name())
		file, err := parser.ParseFile(fset, path, nil, parser.SkipObjectResolution)
		if err != nil {
			return nil, fmt.Errorf("parse SDK API source %s: %w", path, err)
		}
		for _, decl := range file.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Name == nil {
				continue
			}
			receiver := receiverType(fn.Recv)
			switch {
			case strings.HasSuffix(receiver, "OptionalParameters") && strings.HasPrefix(fn.Name.Name, "With"):
				args := functionArguments(fset, fn.Type.Params)
				if len(args) == 1 {
					if inv.setters[receiver] == nil {
						inv.setters[receiver] = map[string]argumentSignature{}
					}
					inv.setters[receiver][fn.Name.Name] = args[0]
				}
			case strings.HasSuffix(receiver, "Api") && !strings.HasSuffix(fn.Name.Name, "WithPagination"):
				args := functionArguments(fset, fn.Type.Params)
				if len(args) == 0 || args[0].name != "ctx" {
					continue
				}
				args = args[1:]
				sig := methodSignature{}
				if len(args) > 0 && strings.HasPrefix(args[len(args)-1].typeName, "...") && strings.HasSuffix(args[len(args)-1].typeName, "OptionalParameters") {
					sig.options = strings.TrimPrefix(args[len(args)-1].typeName, "...")
					args = args[:len(args)-1]
				}
				sig.arguments = args
				inv.methods[fn.Name.Name] = sig
			}
		}
	}
	return inv, nil
}

// Bind resolves op's SDK positional arguments and optional setters to its
// normalized OpenAPI parameters.
func (i *Inventory) Bind(op *model.Operation) error {
	if op == nil {
		return nil
	}
	sig, ok := i.methods[op.OperationId]
	if !ok {
		return fmt.Errorf("sdk binding: method %s was not found in the pinned datadogV2 SDK", op.OperationId)
	}
	parameters := map[string]model.SDKArgument{}
	for _, p := range op.PathParams {
		parameters[canonicalName(p.Name)] = sdkParameter(p, "path")
	}
	for _, p := range op.QueryParams {
		parameters[canonicalName(p.Name)] = sdkParameter(p, "query")
	}

	binding := &model.SDKOperationBinding{OptionalParamsType: sig.options}
	for _, sdkArg := range sig.arguments {
		parameter, ok := parameters[canonicalName(sdkArg.name)]
		if !ok {
			return fmt.Errorf("sdk binding: %s argument %s %s has no matching OpenAPI path or query parameter", op.OperationId, sdkArg.name, sdkArg.typeName)
		}
		parameter.GoName = sdkArg.name
		parameter.GoType = sdkArg.typeName
		binding.Required = append(binding.Required, parameter)
	}

	if sig.options != "" {
		setterNames := make([]string, 0, len(i.setters[sig.options]))
		for setter := range i.setters[sig.options] {
			setterNames = append(setterNames, setter)
		}
		sort.Strings(setterNames)
		for _, setter := range setterNames {
			sdkArg := i.setters[sig.options][setter]
			parameter, ok := parameters[canonicalName(sdkArg.name)]
			if !ok || parameter.Location != "query" {
				continue
			}
			parameter.GoName = sdkArg.name
			parameter.GoType = sdkArg.typeName
			parameter.Setter = setter
			binding.Optional = append(binding.Optional, parameter)
		}
	}
	op.SDKBinding = binding
	return nil
}

func sdkParameter(p model.QueryParam, location string) model.SDKArgument {
	return model.SDKArgument{
		Name:        p.Name,
		Location:    location,
		Description: p.Description,
		Schema:      p.Schema,
	}
}

func receiverType(fields *ast.FieldList) string {
	if fields == nil || len(fields.List) != 1 {
		return ""
	}
	t := fields.List[0].Type
	if star, ok := t.(*ast.StarExpr); ok {
		t = star.X
	}
	ident, _ := t.(*ast.Ident)
	if ident == nil {
		return ""
	}
	return ident.Name
}

func functionArguments(fset *token.FileSet, fields *ast.FieldList) []argumentSignature {
	if fields == nil {
		return nil
	}
	var out []argumentSignature
	for _, field := range fields.List {
		var rendered strings.Builder
		_ = printer.Fprint(&rendered, fset, field.Type)
		for _, name := range field.Names {
			out = append(out, argumentSignature{name: name.Name, typeName: rendered.String()})
		}
	}
	return out
}

func canonicalName(value string) string {
	var b strings.Builder
	for _, r := range value {
		if r >= 'A' && r <= 'Z' {
			r += 'a' - 'A'
		}
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
		}
	}
	return b.String()
}
