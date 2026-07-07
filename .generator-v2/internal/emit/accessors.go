package emit

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
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

// ApplyAPIAccessor overrides view.APIAccessor with the accessor the provider
// actually exposes for view.APIStruct when accessors resolved one; absent a match
// it leaves the builder's derived name in place.
func ApplyAPIAccessor(view *DataSourceView, accessors map[string]string) {
	if acc, ok := accessors[view.APIStruct]; ok {
		view.APIAccessor = acc
	}
}
