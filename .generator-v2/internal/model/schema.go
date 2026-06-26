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
// attribute — anyOf (classified unsupported), a ref_cycle, or any other
// unsupported node. The attribute-tree builder fails the artifact when it reaches
// one rather than emitting a types.Dynamic escape hatch. (oneOf variants are
// dropped, not errored.)
type UnsupportedKindError struct {
	Path string
	Kind SchemaKind
}

func (e *UnsupportedKindError) Error() string {
	return fmt.Sprintf("model: cannot build attribute at %q: schema kind %q is not representable", e.Path, e.Kind)
}

// BuildResponseTree converts a response-body schema into an AttributeTree,
// rooting every attribute path at "response.". It also returns the info
// diagnostics raised while dropping oneOf-variant nodes from the tree.
func BuildResponseTree(s *Schema) (*AttributeTree, []Diagnostic, error) {
	return build(s, "response")
}

// BuildRequestTree converts a request-body schema into an AttributeTree, rooting
// every attribute path at "request.". Like BuildResponseTree it returns the
// drop diagnostics raised during the walk.
func BuildRequestTree(s *Schema) (*AttributeTree, []Diagnostic, error) {
	return build(s, "request")
}

// build is the shared recursion behind both entry points, differing only in root.
// A root object explodes its properties into top-level attributes; any other kind
// becomes one attribute at Path == root, and a nil schema yields an empty tree.
// A root that is itself a dropped oneOf variant has nothing to render and fails.
func build(s *Schema, root string) (*AttributeTree, []Diagnostic, error) {
	tree := &AttributeTree{}
	var diags []Diagnostic
	if s == nil {
		return tree, diags, nil
	}
	if s.Kind == SchemaKindObject {
		attrs, err := buildChildren(s.Properties, root+".", nestBlock, &diags)
		if err != nil {
			return nil, nil, err
		}
		tree.Attributes = attrs
		return tree, diags, nil
	}
	attr, err := buildAttribute(s, root, nestBlock, &diags)
	if err != nil {
		return nil, nil, err
	}
	if attr == nil {
		return nil, nil, fmt.Errorf("model: schema at %q is a dropped oneOf variant with no representable attributes", root)
	}
	tree.Attributes = []*Attribute{attr}
	return tree, diags, nil
}

// buildAttribute converts one schema node at path into an Attribute, recursing
// into its properties, element, or value schema. mode threads the nesting world
// down. A oneOf variant — at this node, or as a collection's element — has no
// Terraform representation, so buildAttribute drops it: it returns a nil
// Attribute (for the caller to skip) and records an info diagnostic on diags.
func buildAttribute(s *Schema, path string, mode nestingMode, diags *[]Diagnostic) (*Attribute, error) {
	// A oneOf variant node: drop it from the tree.
	if s.Kind == SchemaKindVariant {
		*diags = append(*diags, droppedDiag(path, "oneOf variant has no Terraform representation"))
		return nil, nil
	}

	// The remaining non-representable kinds (anyOf and other unsupported nodes,
	// ref_cycle) have no Terraform representation: fail the artifact here rather
	// than emit garbage.
	switch s.Kind {
	case SchemaKindPrimitive, SchemaKindObject, SchemaKindArray, SchemaKindMap:
		// representable — continue
	default:
		return nil, &UnsupportedKindError{Path: path, Kind: s.Kind}
	}

	// A collection whose element is a oneOf variant has no representable element
	// type, so drop the whole collection attribute.
	if (s.Kind == SchemaKindArray || s.Kind == SchemaKindMap) && s.Items != nil && s.Items.Kind == SchemaKindVariant {
		*diags = append(*diags, droppedDiag(path, "collection element is a oneOf variant, which has no Terraform representation"))
		return nil, nil
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
		children, err := buildChildren(s.Properties, path+".", mode, diags)
		if err != nil {
			return nil, err
		}
		attr.Children = children

	case SchemaKindArray:
		if s.Items.Kind == SchemaKindObject {
			children, err := buildChildren(s.Items.Properties, path+"[].", mode, diags)
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
			children, err := buildChildren(s.Items.Properties, path+"{}.", nestAttribute, diags)
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
// Keys are visited sorted, making recursion deterministic and the result
// Path-sorted. A property that builds to a nil Attribute (a dropped oneOf
// variant) is skipped rather than appended.
func buildChildren(props map[string]*Schema, prefix string, mode nestingMode, diags *[]Diagnostic) ([]*Attribute, error) {
	keys := make([]string, 0, len(props))
	for k := range props {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	children := make([]*Attribute, 0, len(props))
	for _, key := range keys {
		// Terraform attribute names must be snake_case; snakeCase normalizes camelCase
		// OAS names and is idempotent on already-snake names (SdkName recovers the getter).
		child, err := buildAttribute(props[key], prefix+snakeCase(key), mode, diags)
		if err != nil {
			return nil, err
		}
		if child == nil {
			continue // dropped (e.g. a oneOf variant)
		}
		children = append(children, child)
	}
	return children, nil
}

// droppedDiag is an info diagnostic recording one node skipped from the attribute
// tree, keeping the run report explicit about what was not rendered.
func droppedDiag(path, reason string) Diagnostic {
	return Diagnostic{
		Severity: SeverityInfo,
		Message:  fmt.Sprintf("dropped %q: %s", path, reason),
	}
}
