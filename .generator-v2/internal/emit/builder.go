package emit

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/terraform-providers/terraform-provider-datadog/generator/internal/model"
)

// UnsupportedNode names one Terraform-representable attribute the singular emit
// path cannot render yet, paired with the reason it was deferred.
type UnsupportedNode struct {
	Path   string
	Reason string
}

// UnsupportedEmitError aggregates every UnsupportedNode found while walking one
// artifact's tree. It reports valid Terraform the emit path does not yet
// implement, separate from the upstream check that rejects unrepresentable schemas.
type UnsupportedEmitError struct {
	Nodes []UnsupportedNode
}

func (e *UnsupportedEmitError) Error() string {
	parts := make([]string, len(e.Nodes))
	for i, n := range e.Nodes {
		parts[i] = n.Path + ": " + n.Reason
	}
	return fmt.Sprintf("emit: %d unsupported node(s): %s", len(e.Nodes), strings.Join(parts, "; "))
}

// BuildDataSourceView derives the singular DataSourceView from a faithful,
// structure-preserving walk of a.Schema, populating Schema, Models, State,
// Read.ResponseType and Cardinality. The walk is fail-slow: every attribute it
// cannot render is collected and returned together as a *UnsupportedEmitError,
// in which case the view is discarded.
func BuildDataSourceView(a *model.Artifact, responseType string) (DataSourceView, error) {
	b := &dataSourceBuilder{}

	var topLevel []*model.Attribute
	if a.Schema != nil {
		topLevel = a.Schema.Attributes
	}
	rootStruct := lowerFirst(model.SdkName(a.Name)) + "DataSourceModel"
	attrs, blocks := b.walk(rootStruct, topLevel)

	// An inline/absent response body leaves no SDK receiver type to mirror state
	// against, so record it alongside any unrenderable attributes.
	if responseType == "" {
		b.unsupported = append(b.unsupported, UnsupportedNode{Path: "response", Reason: "missing response type name"})
	}

	if len(b.unsupported) > 0 {
		return DataSourceView{}, &UnsupportedEmitError{Nodes: b.unsupported}
	}

	return DataSourceView{
		Cardinality: Singular,
		Read:        SDKReadView{ResponseType: responseType},
		Models:      b.models,
		Schema:      SchemaView{Attributes: attrs, Blocks: blocks},
		State:       StateView{Assignments: b.assignments},
	}, nil
}

// dataSourceBuilder accumulates the cross-cutting outputs of one walk: the Go
// model structs (parent before child), the per-leaf state assignments, and the
// unsupported-node collector.
type dataSourceBuilder struct {
	models      []ModelStructView
	assignments []StateAssignment
	unsupported []UnsupportedNode
}

// walk processes one struct's worth of attributes in tree order, reserving the
// struct's slot in b.models up front so a parent precedes its children, then
// filling its fields as it goes. It returns the leaf attribute views and nested
// block views for the caller's schema map, recursing into SingleNestedBlocks.
func (b *dataSourceBuilder) walk(structName string, attrs []*model.Attribute) (attrViews, blockViews []AttrView) {
	idx := len(b.models)
	b.models = append(b.models, ModelStructView{Name: structName})
	var fields []ModelFieldView

	for _, a := range attrs {
		tfName := tfNameOf(a.Path)
		switch a.TfType {
		case "schema.StringAttribute", "schema.Int64Attribute",
			"schema.Float64Attribute", "schema.BoolAttribute":
			attrViews = append(attrViews, AttrView{
				TFName:      tfName,
				TFType:      a.TfType,
				Description: a.Description,
				Required:    a.Required,
				Optional:    a.Optional,
				Computed:    a.Computed,
				Sensitive:   a.Sensitive,
			})
			fields = append(fields, ModelFieldView{
				GoField: model.SdkName(tfName),
				GoType:  a.GoType,
				TFName:  tfName,
			})
			b.assignments = append(b.assignments, StateAssignment{
				LHS: stateLHS(a.Path),
				RHS: wrapValue(a.GoType, getterChain(a.Path)),
			})

		case "schema.SingleNestedBlock":
			item := model.SdkName(tfName)
			childStruct := item + "Model"
			fields = append(fields, ModelFieldView{
				GoField: item,
				GoType:  "*" + childStruct,
				TFName:  tfName,
			})
			childAttrs, childBlocks := b.walk(childStruct, a.Children)
			blockViews = append(blockViews, AttrView{
				TFName:      tfName,
				Description: a.Description,
				Required:    a.Required,
				Optional:    a.Optional,
				Computed:    a.Computed,
				Sensitive:   a.Sensitive,
				IsBlock:     true,
				ListBlock:   false,
				Attributes:  childAttrs,
				Blocks:      childBlocks,
			})

		default:
			b.unsupported = append(b.unsupported, UnsupportedNode{Path: a.Path, Reason: unsupportedReason(a.TfType)})
		}
	}

	b.models[idx].Fields = fields
	return attrViews, blockViews
}

// tfNameOf returns the Terraform attribute key for an attribute path: its last
// dot-segment with array/map markers stripped.
func tfNameOf(path string) string {
	seg := path
	if i := strings.LastIndex(path, "."); i >= 0 {
		seg = path[i+1:]
	}
	return stripMarkers(seg)
}

// stripMarkers removes the "[]" and "{}" array/map markers BuildResponseTree
// embeds in collection paths, leaving a bare identifier segment.
func stripMarkers(s string) string {
	s = strings.ReplaceAll(s, "[]", "")
	return strings.ReplaceAll(s, "{}", "")
}

// responseSegments drops the "response" root from an attribute path and returns
// the remaining marker-free segments, e.g. "response.data.attributes.name" →
// ["data", "attributes", "name"].
func responseSegments(path string) []string {
	parts := strings.Split(path, ".")
	if len(parts) > 0 {
		parts = parts[1:]
	}
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		out = append(out, stripMarkers(p))
	}
	return out
}

// stateLHS mirrors an attribute path onto the Go model field path assigned in
// updateState, e.g. "response.data.attributes.name" → "state.Data.Attributes.Name".
func stateLHS(path string) string {
	segs := responseSegments(path)
	parts := make([]string, len(segs))
	for i, s := range segs {
		parts[i] = model.SdkName(s)
	}
	return "state." + strings.Join(parts, ".")
}

// getterChain builds the SDK getter chain reading an attribute off the response
// receiver, e.g. "response.data.attributes.name" →
// "resp.GetData().GetAttributes().GetName()".
func getterChain(path string) string {
	var b strings.Builder
	b.WriteString("resp")
	for _, s := range responseSegments(path) {
		b.WriteString(".Get")
		b.WriteString(model.SdkName(s))
		b.WriteString("()")
	}
	return b.String()
}

// wrapValue wraps an SDK getter chain in the types.*Value constructor matching
// the model field's GoType, casting integers to int64 as the framework expects.
func wrapValue(goType, chain string) string {
	switch goType {
	case "types.String":
		return "types.StringValue(" + chain + ")"
	case "types.Bool":
		return "types.BoolValue(" + chain + ")"
	case "types.Int64":
		return "types.Int64Value(int64(" + chain + "))"
	case "types.Float64":
		return "types.Float64Value(" + chain + ")"
	default:
		return chain
	}
}

// unsupportedReason returns the deferral message for a Terraform-representable
// TfType the singular emit path does not yet handle.
func unsupportedReason(tfType string) string {
	switch tfType {
	case "schema.ListNestedBlock":
		return "list-of-object not yet supported (plural path)"
	case "schema.ListAttribute":
		return "collection-of-primitive not yet supported"
	case "schema.MapAttribute":
		return "map not yet supported"
	case "schema.MapNestedAttribute":
		return "map-of-object not yet supported"
	case "schema.SingleNestedAttribute", "schema.ListNestedAttribute":
		return "nested-attribute form not yet supported"
	default:
		return fmt.Sprintf("attribute type %q not yet supported", tfType)
	}
}

// lowerFirst lower-cases the first rune of s, the inverse of upperFirst, turning
// an SDK PascalCase name into the lowerCamel base used for model identifiers.
func lowerFirst(s string) string {
	if s == "" {
		return s
	}
	r := []rune(s)
	r[0] = unicode.ToLower(r[0])
	return string(r)
}
