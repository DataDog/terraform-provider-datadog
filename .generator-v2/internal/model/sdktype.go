package model

import "strings"

// This file ports the parts of the Datadog go-sdk generator that decide how its
// generated code spells a scalar schema. Like identifier.go, every rule here is
// a re-derivation from the pinned generator's own source — never a lookup of the
// generated package, and never reflection (FR-005a). Upstream:
//
//	.generator/src/generator/formatter.py         simple_type
//	.generator/src/generator/templates/model_enum.j2

// SDKScalarGoType ports the Datadog go-sdk generator's own scalar type rule
// (.generator/src/generator/formatter.py simple_type, with
// render_nullable=False): the Go type its generated code uses for a scalar
// schema, derived from type+format rather than looked up in the generated
// package (FR-005a).
//
// The second result is false where the Python raises KeyError — an integer or
// number carrying a format the SDK does not map — because there the SDK
// generator itself cannot produce a type, so neither can a faithful
// derivation. An unmapped *string* format falls back to string, mirroring the
// upstream .get(type_format, "string").
//
// It lives in model — the package every caller already imports — because it is
// the one answer several of them need and had begun to re-derive separately:
// the oneOf binder (internal/sdkbind), the parameter binder
// (internal/sdkbinding) and the resource request mapper all now route here
// (T137a). It sits beside the other ports of the SDK generator's naming rules
// rather than in schema.go, which speaks Terraform rather than SDK.
func SDKScalarGoType(s *Schema) (string, bool) {
	if s == nil {
		return "", false
	}
	switch s.Type {
	case "integer":
		switch s.Format {
		case "", "int32":
			return "int32", true
		case "int64":
			return "int64", true
		default:
			return "", false
		}
	case "number":
		switch s.Format {
		case "":
			return "float", true
		case "double":
			return "float64", true
		default:
			return "", false
		}
	case "string":
		switch s.Format {
		case "date", "date-time":
			return "time.Time", true
		case "binary":
			return "_io.Reader", true
		case "uuid":
			return "uuid.UUID", true
		default:
			return "string", true
		}
	case "boolean":
		return "bool", true
	default:
		return "", false
	}
}

// SDKEnumFromValueConstructor names the validating constructor the SDK declares
// for an enum component, from templates/model_enum.j2:
//
//	func New{{ name }}FromValue(v {{ model|simple_type }}) (*{{ name }}, error)
//
// where name is the component's Go type. The second result is false for a
// composite spelling, which names no component and so has no constructor.
func SDKEnumFromValueConstructor(goType string) (string, bool) {
	if !IsBareSDKTypeName(goType) {
		return "", false
	}
	return "New" + goType + "FromValue", true
}

// IsBareSDKTypeName reports whether goType names a component rather than being a
// composite spelling. The punctuation set covers every composite the binder
// produces: [] slice, * pointer, {} interface or struct literal, . qualifier.
func IsBareSDKTypeName(goType string) bool {
	return goType != "" && !strings.ContainsAny(goType, "[]*{}.")
}
