package emit

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)

// ResolveAPIAccessors parses the provider's ApiInstances helper at path and maps
// each V2 SDK API struct to the accessor method that returns it, e.g. "RUMApi" ->
// "GetRumApiV2". It is the source of truth for accessor names, which diverge from
// the struct name for a few APIs (RUM, APM, Observability Pipelines).
func ResolveAPIAccessors(path string) (map[string]string, error) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, path, nil, 0)
	if err != nil {
		return nil, fmt.Errorf("parse api instances helper %s: %w", path, err)
	}

	accessors := map[string]string{}
	for _, decl := range file.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || fn.Name == nil || !isApiInstancesReceiver(fn.Recv) {
			continue
		}
		name := fn.Name.Name
		if !strings.HasPrefix(name, "Get") || !strings.HasSuffix(name, "V2") {
			continue
		}
		if t := singleV2ResultType(fn.Type); t != "" {
			accessors[t] = name
		}
	}
	return accessors, nil
}

// ResolveAPIConstructors parses the pinned SDK's datadogV2 package at dir and
// maps each API struct to the constructor that accepts the shared API client,
// e.g. "CaseManagementApi" -> "NewCaseManagementApi". Discovering constructors
// from source preserves SDK acronym spelling and prevents the generator from
// emitting a guessed function name.
func ResolveAPIConstructors(dir string) (map[string]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("read SDK API package %s: %w", dir, err)
	}

	fset := token.NewFileSet()
	constructors := map[string]string{}
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
			if !ok || fn.Recv != nil || fn.Name == nil || !strings.HasPrefix(fn.Name.Name, "New") {
				continue
			}
			apiStruct := apiConstructorResultType(fn.Type)
			if apiStruct == "" {
				continue
			}
			if previous, ok := constructors[apiStruct]; ok && previous != fn.Name.Name {
				return nil, fmt.Errorf("SDK API struct %s has multiple constructors: %s and %s", apiStruct, previous, fn.Name.Name)
			}
			constructors[apiStruct] = fn.Name.Name
		}
	}
	return constructors, nil
}

// isApiInstancesReceiver reports whether recv is the pointer receiver
// (i *ApiInstances).
func isApiInstancesReceiver(recv *ast.FieldList) bool {
	if recv == nil || len(recv.List) != 1 {
		return false
	}
	star, ok := recv.List[0].Type.(*ast.StarExpr)
	if !ok {
		return false
	}
	ident, ok := star.X.(*ast.Ident)
	return ok && ident.Name == "ApiInstances"
}

// singleV2ResultType returns the struct name X of a method returning exactly one
// *datadogV2.X value, or "" if the results do not match that shape.
func singleV2ResultType(ft *ast.FuncType) string {
	if ft.Results == nil || len(ft.Results.List) != 1 {
		return ""
	}
	star, ok := ft.Results.List[0].Type.(*ast.StarExpr)
	if !ok {
		return ""
	}
	sel, ok := star.X.(*ast.SelectorExpr)
	if !ok {
		return ""
	}
	pkg, ok := sel.X.(*ast.Ident)
	if !ok || pkg.Name != "datadogV2" {
		return ""
	}
	return sel.Sel.Name
}

// apiConstructorResultType returns X for a function with the SDK constructor
// shape func(*datadog.APIClient) *X, or "" if the signature does not match.
func apiConstructorResultType(ft *ast.FuncType) string {
	if ft.Params == nil || len(ft.Params.List) != 1 {
		return ""
	}
	param, ok := ft.Params.List[0].Type.(*ast.StarExpr)
	if !ok {
		return ""
	}
	paramType, ok := param.X.(*ast.SelectorExpr)
	if !ok || paramType.Sel.Name != "APIClient" {
		return ""
	}
	paramPkg, ok := paramType.X.(*ast.Ident)
	if !ok || paramPkg.Name != "datadog" {
		return ""
	}

	if ft.Results == nil || len(ft.Results.List) != 1 {
		return ""
	}
	result, ok := ft.Results.List[0].Type.(*ast.StarExpr)
	if !ok {
		return ""
	}
	apiStruct, ok := result.X.(*ast.Ident)
	if !ok || !strings.HasSuffix(apiStruct.Name, "Api") {
		return ""
	}
	return apiStruct.Name
}

// ApplyAPIAccessor configures view to use the provider's existing ApiInstances
// accessor when one returns view.APIStruct. If no accessor exists, it selects
// the constructor discovered from the pinned SDK. The builder's derived
// accessor is never retained: an unresolved client initializer is an error
// instead of latent uncompilable Go.
func ApplyAPIAccessor(view *DataSourceView, accessors, constructors map[string]string) error {
	view.APIConstructor = ""
	if acc, ok := accessors[view.APIStruct]; ok {
		view.APIAccessor = acc
		return nil
	}
	view.APIAccessor = ""
	if constructor, ok := constructors[view.APIStruct]; ok {
		view.APIConstructor = constructor
		return nil
	}
	return fmt.Errorf(
		"resolve SDK API client %s.%s: no ApiInstances accessor or SDK constructor was found; verify the pinned datadog-api-client-go package exports func(*datadog.APIClient) *%s, or add an ApiInstances getter",
		view.SDKPackage,
		view.APIStruct,
		view.APIStruct,
	)
}
