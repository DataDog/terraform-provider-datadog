package emit

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"strings"

	"github.com/terraform-providers/terraform-provider-datadog/generator/internal/model"
)

// ResolveAPIAccessors parses the provider's ApiInstances helper at path and maps
// each V2 SDK API struct to the accessor method returning it, e.g. "RUMApi" ->
// "GetRumApiV2". Names are read rather than derived because they diverge from
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

// ApplyAPIAccessor points view at the provider's existing ApiInstances accessor
// when one returns view.APIStruct, and otherwise derives the client constructor
// from the SDK's deterministic New<APIStruct> rule — so provider aliases win
// without the pinned SDK source being needed.
func ApplyAPIAccessor(view *DataSourceView, accessors map[string]string) error {
	accessor, constructor, err := resolveAPIAccessor(view.SDKPackage, view.APIStruct, accessors)
	view.APIAccessor, view.APIConstructor = accessor, constructor
	return err
}

// ApplyResourceAPIAccessor is ApplyAPIAccessor for a ResourceView.
func ApplyResourceAPIAccessor(view *ResourceView, accessors map[string]string) error {
	accessor, constructor, err := resolveAPIAccessor(view.SDKPackage, view.APIStruct, accessors)
	view.APIAccessor, view.APIConstructor = accessor, constructor
	return err
}

// defaultAPIAccessor is the Get<Struct><V1|V2> accessor a call resolves to by
// convention, e.g. "GetTeamsApiV2". It seeds a freshly built view so rendering
// is valid before accessor resolution overwrites it from the provider's real
// ApiInstances helper.
func defaultAPIAccessor(call *model.SDKCall) string {
	return "Get" + call.GoApiStruct + strings.TrimPrefix(call.GoPackage, "datadog")
}

// resolveAPIAccessor returns exactly one of accessor and constructor non-empty
// on success; an operation with no usable API tag is an error.
func resolveAPIAccessor(sdkPackage, apiStruct string, accessors map[string]string) (accessor, constructor string, err error) {
	if acc, ok := accessors[apiStruct]; ok {
		return acc, "", nil
	}
	if apiStruct == "" || apiStruct == "Api" {
		return "", "", fmt.Errorf("resolve SDK API client %s.%s: OpenAPI operation has no usable API tag", sdkPackage, apiStruct)
	}
	return "", "New" + apiStruct, nil
}
