// Package sdkbinding derives Go SDK operation signatures from OpenAPI and can
// optionally corroborate them against the provider's pinned generated client.
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
	"unicode"

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
// pinned datadogV2 package. It is corroborating evidence only: generation does
// not require the package or a matching endpoint to be present.
type Inventory struct {
	methods map[string]methodSignature
	setters map[string]map[string]argumentSignature
}

// Load parses the generated SDK API files in dir for optional corroboration.
// Callers must treat an unavailable inventory as non-fatal because the OpenAPI
// derivation is the authoritative generation input.
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

type orderedParameter struct {
	parameter model.QueryParam
	location  string
}

// Bind derives op's SDK call binding from OpenAPI using the Datadog Go client
// generator's parameter, naming, and type rules. When inventory is present and
// contains the same endpoint, a disagreement is returned as a warning while the
// derived binding remains authoritative. A missing inventory or endpoint is an
// ordinary successful binding.
func Bind(op *model.Operation, inventory *Inventory) ([]model.Diagnostic, error) {
	if op == nil {
		return nil, nil
	}
	binding, err := derive(op)
	if err != nil {
		return nil, err
	}
	op.SDKBinding = binding
	if inventory == nil {
		return nil, nil
	}
	return inventory.corroborate(op), nil
}

func derive(op *model.Operation) (*model.SDKOperationBinding, error) {
	parameters := make([]orderedParameter, 0, len(op.PathParams)+len(op.QueryParams))
	for _, parameter := range op.PathParams {
		parameters = append(parameters, orderedParameter{parameter: parameter, location: "path"})
	}
	for _, parameter := range op.QueryParams {
		parameters = append(parameters, orderedParameter{parameter: parameter, location: "query"})
	}
	sort.SliceStable(parameters, func(i, j int) bool {
		left, right := parameters[i].parameter.DeclarationOrder, parameters[j].parameter.DeclarationOrder
		switch {
		case left > 0 && right > 0:
			return left < right
		case left > 0:
			return true
		case right > 0:
			return false
		default:
			return false
		}
	})

	binding := &model.SDKOperationBinding{}
	for _, ordered := range parameters {
		parameter := ordered.parameter
		if ordered.location != "path" && !parameter.Required {
			continue
		}
		argument, err := deriveArgument(parameter, ordered.location)
		if err != nil {
			return nil, fmt.Errorf("sdk binding: derive %s argument %s: %w", op.OperationId, parameter.Name, err)
		}
		binding.Required = append(binding.Required, argument)
	}

	for _, ordered := range parameters {
		parameter := ordered.parameter
		if ordered.location != "query" || parameter.Required {
			continue
		}
		argument, err := deriveArgument(parameter, ordered.location)
		if err != nil {
			return nil, fmt.Errorf("sdk binding: derive %s optional argument %s: %w", op.OperationId, parameter.Name, err)
		}
		argument.Setter = "With" + model.SdkName(parameter.Name)
		binding.Optional = append(binding.Optional, argument)
	}
	if len(binding.Optional) > 0 {
		binding.OptionalParamsType = op.OperationId + "OptionalParameters"
	}
	return binding, nil
}

func deriveArgument(parameter model.QueryParam, location string) (model.SDKArgument, error) {
	typeName, err := parameterGoType(parameter.Schema)
	if err != nil {
		return model.SDKArgument{}, err
	}
	return model.SDKArgument{
		Name:        parameter.Name,
		GoName:      variableName(parameter.Name),
		GoType:      typeName,
		Location:    location,
		Description: parameter.Description,
		Schema:      parameter.Schema,
	}, nil
}

// parameterGoType ports generator.openapi.type_to_go as used by
// get_type_for_parameter. The scalar and named-enum cases are the generation
// path tfgen currently emits; collection/object spellings keep corroboration
// faithful for parameters the scalar-first emitter later drops.
func parameterGoType(schema *model.Schema) (string, error) {
	if schema == nil {
		return "", fmt.Errorf("parameter has no schema")
	}
	if len(schema.Enum) == 0 {
		if simple, ok, err := parameterSimpleType(schema); ok || err != nil {
			return simple, err
		}
	}
	if schema.RefName != "" {
		switch schema.Kind {
		case model.SchemaKindPrimitive:
			if len(schema.Enum) > 0 {
				return schema.RefName, nil
			}
		case model.SchemaKindObject, model.SchemaKindOneOf:
			return schema.RefName, nil
		}
	}
	switch schema.Kind {
	case model.SchemaKindArray:
		item, err := parameterGoType(schema.Items)
		if err != nil {
			return "", err
		}
		return "[]" + item, nil
	case model.SchemaKindMap:
		item, err := parameterGoType(schema.Items)
		if err != nil {
			return "", err
		}
		return "map[string]" + item, nil
	case model.SchemaKindObject:
		return "interface{}", nil
	case model.SchemaKindPrimitive:
		if len(schema.Enum) > 0 {
			return "", fmt.Errorf("anonymous enum has no Go SDK type; promote it to a named schema component")
		}
	}
	return "", fmt.Errorf("schema kind %s has no Go SDK parameter type", schema.Kind)
}

func parameterSimpleType(schema *model.Schema) (typeName string, recognized bool, err error) {
	switch schema.Type {
	case "integer":
		switch schema.Format {
		case "", "int32":
			return "int32", true, nil
		case "int64":
			return "int64", true, nil
		default:
			return "", true, fmt.Errorf("integer format %q is not mapped by the Go SDK generator", schema.Format)
		}
	case "number":
		switch schema.Format {
		case "":
			return "float", true, nil
		case "double":
			return "float64", true, nil
		default:
			return "", true, fmt.Errorf("number format %q is not mapped by the Go SDK generator", schema.Format)
		}
	case "string":
		switch schema.Format {
		case "date", "date-time":
			return "time.Time", true, nil
		case "binary":
			return "_io.Reader", true, nil
		case "uuid":
			return "uuid.UUID", true, nil
		default:
			return "string", true, nil
		}
	case "boolean":
		return "bool", true, nil
	default:
		return "", false, nil
	}
}

func variableName(openAPIName string) string {
	name := model.SdkName(openAPIName)
	if name == "" {
		return ""
	}
	runes := []rune(name)
	runes[0] = unicode.ToLower(runes[0])
	return model.EscapeReservedKeyword(string(runes))
}

func (i *Inventory) corroborate(op *model.Operation) []model.Diagnostic {
	if i == nil || op == nil || op.SDKBinding == nil {
		return nil
	}
	pinned, ok := i.methods[op.OperationId]
	if !ok {
		return nil
	}
	var differences []string
	derived := op.SDKBinding
	if got, want := renderArguments(derived.Required), renderArgumentSignatures(pinned.arguments); got != want {
		differences = append(differences, fmt.Sprintf("required arguments derived [%s], pinned [%s]", got, want))
	}
	if derived.OptionalParamsType != pinned.options {
		differences = append(differences, fmt.Sprintf("optional parameters type derived %q, pinned %q", derived.OptionalParamsType, pinned.options))
	}
	derivedSetters := make(map[string]bool, len(derived.Optional))
	for _, argument := range derived.Optional {
		derivedSetters[argument.Setter] = true
		pinnedSetter, ok := i.setters[pinned.options][argument.Setter]
		if !ok {
			differences = append(differences, fmt.Sprintf("setter %s derived but absent from pinned SDK", argument.Setter))
			continue
		}
		if pinnedSetter.name != argument.GoName || pinnedSetter.typeName != argument.GoType {
			differences = append(differences, fmt.Sprintf("setter %s derived (%s %s), pinned (%s %s)", argument.Setter, argument.GoName, argument.GoType, pinnedSetter.name, pinnedSetter.typeName))
		}
	}
	var extraPinnedSetters []string
	for setter := range i.setters[pinned.options] {
		if !derivedSetters[setter] {
			extraPinnedSetters = append(extraPinnedSetters, setter)
		}
	}
	sort.Strings(extraPinnedSetters)
	for _, setter := range extraPinnedSetters {
		differences = append(differences, fmt.Sprintf("setter %s is present in the pinned SDK but absent from the OpenAPI-derived binding", setter))
	}
	if len(differences) == 0 {
		return nil
	}
	return []model.Diagnostic{{
		Severity: model.SeverityWarning,
		Message: fmt.Sprintf(
			"pinned SDK binding for %s differs from the OpenAPI-derived binding; generated with OpenAPI binding: %s",
			op.OperationId,
			strings.Join(differences, "; "),
		),
	}}
}

func renderArguments(arguments []model.SDKArgument) string {
	parts := make([]string, 0, len(arguments))
	for _, argument := range arguments {
		parts = append(parts, argument.GoName+" "+argument.GoType)
	}
	return strings.Join(parts, ", ")
}

func renderArgumentSignatures(arguments []argumentSignature) string {
	parts := make([]string, 0, len(arguments))
	for _, argument := range arguments {
		parts = append(parts, argument.name+" "+argument.typeName)
	}
	return strings.Join(parts, ", ")
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
