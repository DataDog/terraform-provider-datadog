package sdkbind

import (
	"fmt"

	"github.com/terraform-providers/terraform-provider-datadog/generator/internal/model"
)

// This file ports the parts of the Datadog go-sdk generator that decide what a
// oneOf wrapper's members are called. Everything here is a re-derivation from the
// pinned generator's own source — never a lookup of the generated package, and
// never reflection. The upstream sources are:
//
//	.generator/src/generator/templates/model_oneof.j2  (member + constructor shape)
//	.generator/src/generator/openapi.py     type_to_go, get_name
//	.generator/src/generator/formatter.py   simple_type
//	.generator/src/generator/utils.py       upperfirst

// simpleType ports formatter.simple_type with render_nullable=False, which is how
// model_oneof.j2 calls it (`get_type(oneOf)` passes no render_nullable, so a
// nullable alternative is NOT spelled datadog.Nullable<T> in a wrapper member —
// the nullable spelling applies to simple_type in general, not to the oneOf
// call site).
//
// The second result is false where the Python raises KeyError — an integer or
// number carrying a format the SDK does not map — because there the SDK generator
// itself cannot produce a type, so neither can a faithful derivation.
func simpleType(s *model.Schema) (string, bool) {
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
		// .get(type_format, "string"): an unmapped format falls back to string.
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

// memberBinding derives the SDK wrapper member for one alternative:
// `(get_name(oneOf) or get_type(oneOf))|upperfirst` from model_oneof.j2, plus
// whether that member is a pointer.
//
// A referenced alternative is named after its component — `get_name` wins over
// the Go type spelling for the *name* even where simple_type would also answer —
// so a variant Terraform calls aws_integration binds to the SDK's AWSIntegration
// rather than to AwsIntegration.
//
// Every alternative is a pointer except a free-form object, which the SDK emits as
// a bare map because it is already nil-able (model_oneof.j2's
// isAdditionalPropertiesContainer).
func memberBinding(v model.OneOfVariant) (name string, pointer bool, err error) {
	if v.Schema == nil {
		return "", false, fmt.Errorf("alternative has no normalized schema")
	}
	pointer = v.Schema.Kind != model.SchemaKindMap

	if v.RefName != "" {
		return model.UpperFirst(v.RefName), pointer, nil
	}

	// Anonymous alternative: the member name is its Go type spelling, so it exists
	// only for the shapes type_to_go can spell as a Go identifier.
	if len(v.Schema.Enum) > 0 {
		// type_to_go skips simple_type when a schema has an enum and then finds no
		// name, reaching `raise ValueError(f"Unknown type {type_}")`. The SDK cannot
		// generate this wrapper at all.
		return "", false, fmt.Errorf(
			"anonymous enum alternative has no SDK member name (the go-sdk generator " +
				"cannot name it either); promote it to a named schema component")
	}
	spelling, ok := simpleType(v.Schema)
	if !ok {
		return "", false, fmt.Errorf(
			"anonymous %s alternative has no SDK member name; replace the inline "+
				"alternative with a $ref to a named schema component",
			describeKind(v.Schema))
	}
	name = model.UpperFirst(spelling)
	// upperfirst of a qualified or composite spelling (time.Time, uuid.UUID,
	// []Foo, map[string]Foo) is not a Go identifier, so the SDK would emit a
	// wrapper that does not compile — meaning the specification never contains one.
	// Fail here rather than derive a plausible name for a member that cannot exist.
	if !isGoIdentifier(name) {
		return "", false, fmt.Errorf(
			"anonymous alternative of Go type %q would yield the SDK member %q, which is not a "+
				"valid Go identifier, so the go-sdk emits no such member; replace the inline "+
				"alternative with a $ref to a named schema component",
			spelling, name)
	}
	return name, pointer, nil
}

// describeKind names a schema's shape for a diagnostic, preferring the OpenAPI
// type when there is one so the reader can find the node in the specification.
func describeKind(s *model.Schema) string {
	if s.Type != "" {
		if s.Format != "" {
			return fmt.Sprintf("%s/%s", s.Type, s.Format)
		}
		return s.Type
	}
	return string(s.Kind)
}

// isGoIdentifier reports whether name is a legal, non-empty Go identifier. Go
// keywords are not checked: every name reaching this point starts with an
// upper-case rune, and no Go keyword does.
func isGoIdentifier(name string) bool {
	for i, r := range name {
		switch {
		case r == '_':
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z':
		case r >= '0' && r <= '9':
			if i == 0 {
				return false
			}
		default:
			return false
		}
	}
	return name != ""
}
