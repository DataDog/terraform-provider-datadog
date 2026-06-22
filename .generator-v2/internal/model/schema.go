package model

import (
	"fmt"
	"sort"
	"strconv"
)

// nestingMode tracks whether the current subtree nests as blocks or as attributes.
// Inside a map<object> value blocks are forbidden, so the builder switches to
// nestAttribute and rewrites block forms to their attribute counterparts.
type nestingMode int

const (
	nestBlock nestingMode = iota
	nestAttribute
)

// UnsupportedKindError reports a schema kind that cannot become a Terraform
// attribute. It should be unreachable — CheckSchemaRepresentability rejects such
// kinds first — and exists to fail loudly if that invariant is violated.
type UnsupportedKindError struct {
	Path string
	Kind SchemaKind
}

func (e *UnsupportedKindError) Error() string {
	return fmt.Sprintf("model: cannot build attribute at %q: schema kind %q is "+
		"not representable (should have been rejected by CheckSchemaRepresentability)", e.Path, e.Kind)
}

// BuildResponseTree converts a response-body schema into an AttributeTree,
// rooting every attribute path at "response.".
func BuildResponseTree(s *Schema) (*AttributeTree, error) {
	return build(s, "response")
}

// BuildRequestTree converts a request-body schema into an AttributeTree, rooting
// every attribute path at "request.".
func BuildRequestTree(s *Schema) (*AttributeTree, error) {
	return build(s, "request")
}

// build is the shared recursion behind both entry points, differing only in root.
// A root object explodes its properties into top-level attributes; any other kind
// becomes one attribute at Path == root, and a nil schema yields an empty tree.
func build(s *Schema, root string) (*AttributeTree, error) {
	tree := &AttributeTree{}
	if s == nil {
		return tree, nil
	}
	if s.Kind == SchemaKindObject {
		attrs, err := buildChildren(s.Properties, root+".", nestBlock)
		if err != nil {
			return nil, err
		}
		tree.Attributes = attrs
		return tree, nil
	}
	attr, err := buildAttribute(s, root, nestBlock)
	if err != nil {
		return nil, err
	}
	tree.Attributes = []*Attribute{attr}
	return tree, nil
}

// buildAttribute converts one schema node at path into an Attribute, recursing
// into its properties, element, or value schema. mode threads the nesting world down.
func buildAttribute(s *Schema, path string, mode nestingMode) (*Attribute, error) {
	// Defensive guard: these kinds are rejected upstream, so reaching one here is a
	// broken invariant — fail loudly rather than emit garbage.
	switch s.Kind {
	case SchemaKindPrimitive, SchemaKindObject, SchemaKindArray, SchemaKindMap:
		// representable — continue
	default:
		return nil, &UnsupportedKindError{Path: path, Kind: s.Kind}
	}

	tfType, goType, err := FrameworkType(s)
	if err != nil {
		return nil, err
	}
	// Inside a map value the framework forbids blocks, so rewrite block forms to
	// attribute forms (leaf types are unaffected).
	if mode == nestAttribute {
		switch tfType {
		case "schema.SingleNestedBlock":
			tfType = "schema.SingleNestedAttribute"
		case "schema.ListNestedBlock":
			tfType = "schema.ListNestedAttribute"
		}
	}

	attr := &Attribute{
		Path:        path,
		TfType:      tfType,
		GoType:      goType,
		Format:      s.Format,
		Computed:    true,
		Sensitive:   s.Sensitive,
		Description: s.Description,
	}

	// A string enum becomes a OneOf validator; non-string enums produce none for now.
	if s.Kind == SchemaKindPrimitive && s.Type == "string" && len(s.Enum) > 0 {
		attr.IsEnum = true
		args := make([]string, len(s.Enum))
		for i, v := range s.Enum {
			args[i] = strconv.Quote(v)
		}
		attr.Validators = []ValidatorSpec{{Name: "stringvalidator.OneOf", Args: args}}
	}

	// Recurse into nested shapes, or record a primitive collection's element type.
	// FrameworkType already validated array/map elements, so the else branch is primitive.
	switch s.Kind {
	case SchemaKindObject:
		children, err := buildChildren(s.Properties, path+".", mode)
		if err != nil {
			return nil, err
		}
		attr.Children = children

	case SchemaKindArray:
		if s.Items.Kind == SchemaKindObject {
			children, err := buildChildren(s.Items.Properties, path+"[].", mode)
			if err != nil {
				return nil, err
			}
			attr.Children = children
		} else {
			elem, err := ElementType(s.Items)
			if err != nil {
				return nil, err
			}
			attr.ElementType = elem
		}

	case SchemaKindMap:
		if s.Items.Kind == SchemaKindObject {
			// A map<object> is a NestedAttributeObject; force everything beneath it
			// into attribute form regardless of the incoming mode.
			children, err := buildChildren(s.Items.Properties, path+"{}.", nestAttribute)
			if err != nil {
				return nil, err
			}
			attr.Children = children
		} else {
			elem, err := ElementType(s.Items)
			if err != nil {
				return nil, err
			}
			attr.ElementType = elem
		}
	}

	return attr, nil
}

// buildChildren builds one child attribute per property, each pathed prefix+key.
// Keys are visited sorted, making recursion deterministic and the result Path-sorted.
func buildChildren(props map[string]*Schema, prefix string, mode nestingMode) ([]*Attribute, error) {
	keys := make([]string, 0, len(props))
	for k := range props {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	children := make([]*Attribute, 0, len(props))
	for _, key := range keys {
		child, err := buildAttribute(props[key], prefix+key, mode)
		if err != nil {
			return nil, err
		}
		children = append(children, child)
	}
	return children, nil
}
