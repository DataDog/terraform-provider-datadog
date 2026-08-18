package model

import "fmt"

// FrameworkType maps a schema node to its framework type strings: tfType is the
// schema.* symbol (e.g. schema.StringAttribute), goType the types.* value (e.g.
// types.String). Objects map to the block form (the builder rewrites to attribute
// form where needed); unrepresentable kinds return an error naming the offender.
//
// A oneOf node has no entry of its own: its Terraform shape is a synthetic
// envelope whose form depends on where the union sits, so the attribute-tree
// builder decides it. A collection *of* unions does have an entry — the list or
// map is representable regardless of what its elements are, and it nests the
// element's variant blocks the same way it would an object's properties.
func FrameworkType(s *Schema) (tfType, goType string, err error) {
	switch s.Kind {
	case SchemaKindPrimitive:
		return primitiveFrameworkType(s)

	case SchemaKindObject:
		return "schema.SingleNestedBlock", "types.Object", nil

	case SchemaKindArray:
		if s.Items == nil {
			return "", "", fmt.Errorf("model: array schema has nil items, no element type to map")
		}
		switch s.Items.Kind {
		case SchemaKindPrimitive, SchemaKindArray, SchemaKindMap:
			return "schema.ListAttribute", "types.List", nil
		case SchemaKindObject, SchemaKindOneOf:
			return "schema.ListNestedBlock", "types.List", nil
		default:
			return "", "", fmt.Errorf("model: array element kind %q is not representable", s.Items.Kind)
		}

	case SchemaKindMap:
		if s.Items == nil {
			return "", "", fmt.Errorf("model: map schema has nil value, no value type to map")
		}
		switch s.Items.Kind {
		case SchemaKindPrimitive, SchemaKindArray, SchemaKindMap:
			return "schema.MapAttribute", "types.Map", nil
		case SchemaKindObject, SchemaKindOneOf:
			return "schema.MapNestedAttribute", "types.Map", nil
		default:
			return "", "", fmt.Errorf("model: map value kind %q is not representable", s.Items.Kind)
		}

	default:
		return "", "", fmt.Errorf("model: schema kind %q is not representable", s.Kind)
	}
}

// primitiveFrameworkType maps a primitive's scalar Type to its framework types,
// ignoring Format (int32/int64 → Int64, double → Float64). An unknown or empty
// type errors, naming the type and format.
func primitiveFrameworkType(s *Schema) (tfType, goType string, err error) {
	switch s.Type {
	case "string":
		return "schema.StringAttribute", "types.String", nil
	case "integer":
		return "schema.Int64Attribute", "types.Int64", nil
	case "number":
		return "schema.Float64Attribute", "types.Float64", nil
	case "boolean":
		return "schema.BoolAttribute", "types.Bool", nil
	default:
		return "", "", fmt.Errorf("model: primitive type %q (format %q) is not representable", s.Type, s.Format)
	}
}

// ElementType recursively maps a collection element/value schema to its
// framework attr.Type (for example types.StringType or
// types.MapType{ElemType: types.ListType{ElemType: types.StringType}}). Objects
// still nest via Children and therefore have no attr.Type expression here.
func ElementType(elem *Schema) (string, error) {
	if elem == nil {
		return "", fmt.Errorf("model: collection has nil element, no element type to map")
	}
	switch elem.Kind {
	case SchemaKindPrimitive:
		switch elem.Type {
		case "string":
			return "types.StringType", nil
		case "integer":
			return "types.Int64Type", nil
		case "number":
			return "types.Float64Type", nil
		case "boolean":
			return "types.BoolType", nil
		default:
			return "", fmt.Errorf("model: primitive element type %q has no framework element type", elem.Type)
		}
	case SchemaKindArray:
		child, err := ElementType(elem.Items)
		if err != nil {
			return "", err
		}
		return "types.ListType{ElemType: " + child + "}", nil
	case SchemaKindMap:
		child, err := ElementType(elem.Items)
		if err != nil {
			return "", err
		}
		return "types.MapType{ElemType: " + child + "}", nil
	default:
		return "", fmt.Errorf("model: collection element kind %q has no framework element type", elem.Kind)
	}
}
