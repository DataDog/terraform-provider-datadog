package sdkbind

import (
	"fmt"

	"github.com/terraform-providers/terraform-provider-datadog/generator/internal/model"
)

// This file ports the parts of the Datadog go-sdk generator that decide what a
// oneOf wrapper's members are called — re-derived from the pinned generator's
// source, never a lookup or reflection over the generated package. Upstream:
// templates/model_oneof.j2 (member + constructor shape), openapi.py type_to_go
// and get_name, formatter.py simple_type, utils.py upperfirst.

// memberBinding derives the SDK wrapper member for one alternative —
// `(get_name(oneOf) or get_type(oneOf))|upperfirst` — and whether it is a
// pointer. get_name wins for a referenced alternative, so a variant Terraform
// calls aws_integration binds to AWSIntegration, not AwsIntegration. Every
// alternative is a pointer except a free-form object, emitted as a bare map.
func memberBinding(v model.OneOfVariant) (name string, pointer bool, err error) {
	if v.Schema == nil {
		return "", false, fmt.Errorf("alternative has no normalized schema")
	}
	pointer = v.Schema.Kind != model.SchemaKindMap

	if v.RefName != "" {
		return model.UpperFirst(v.RefName), pointer, nil
	}

	// Anonymous alternative: the member name is its Go type spelling, so it
	// exists only for shapes type_to_go can spell as a Go identifier.
	if len(v.Schema.Enum) > 0 {
		// type_to_go skips simple_type for an enum schema and then finds no name,
		// raising "Unknown type": the SDK cannot generate this wrapper at all.
		return "", false, fmt.Errorf(
			"anonymous enum alternative has no SDK member name (the go-sdk generator " +
				"cannot name it either); promote it to a named schema component")
	}
	spelling, ok := model.SDKScalarGoType(v.Schema)
	if !ok {
		return "", false, fmt.Errorf(
			"anonymous %s alternative has no SDK member name; replace the inline "+
				"alternative with a $ref to a named schema component",
			describeKind(v.Schema))
	}
	name = model.UpperFirst(spelling)
	// upperfirst of a qualified or composite spelling (time.Time, []Foo,
	// map[string]Foo) is not a Go identifier, so no such SDK member exists. Fail
	// rather than derive a plausible name for a member that cannot exist.
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
