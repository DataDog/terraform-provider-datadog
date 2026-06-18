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

// envelopeReceiver is the local the flattened state-mapper reads hoisted leaves
// off, established by the StateView preamble "attributes := resp.Data.GetAttributes()".
const envelopeReceiver = "attributes"

// BuildDataSourceView derives the singular DataSourceView from a.Schema and
// a.Lifecycle. It resolves the SDK-call bindings onto the view, then runs a
// flattening pass that recognizes the singular JSON:API envelope
// ({data:{id,type,attributes}}) and hoists data.attributes.* to top-level
// computed attributes. The walk is fail-slow: every binding or envelope problem
// it finds is collected and returned together as a *UnsupportedEmitError, in
// which case the view is discarded.
func BuildDataSourceView(a *model.Artifact) (DataSourceView, error) {
	b := &dataSourceBuilder{receiver: envelopeReceiver}

	// SDK-call bindings. An inline/absent response body leaves no SDK receiver
	// type to mirror state against, so record it as unsupported.
	var call *model.SDKCall
	var idStrategy model.IdStrategy
	if a.Lifecycle != nil {
		call = a.Lifecycle.Read
		idStrategy = a.Lifecycle.IdStrategy
	}
	if call == nil || call.GoResponseType == "" {
		b.unsupported = append(b.unsupported, UnsupportedNode{Path: "response", Reason: "missing response type name"})
	}

	var topLevel []*model.Attribute
	if a.Schema != nil {
		topLevel = a.Schema.Attributes
	}

	rootStruct := lowerFirst(model.SdkName(a.Name)) + "DataSourceModel"
	env := b.flattenEnvelope(topLevel, idStrategy)

	if len(b.unsupported) > 0 {
		return DataSourceView{}, &UnsupportedEmitError{Nodes: b.unsupported}
	}

	// env is non-nil here: flattenEnvelope records an unsupported node (caught
	// above) on every failure path. Walk the hoisted leaves into the root struct,
	// then prepend the lookup id (field + assignment, sourced from id_strategy).
	attrs, _ := b.walk(rootStruct, env.leaves)
	b.models[0].Fields = append([]ModelFieldView{env.idField}, b.models[0].Fields...)
	assignments := append([]StateAssignment{env.idAssign}, b.assignments...)

	return DataSourceView{
		Cardinality: Singular,
		TypeName:    a.Name,
		GoName:      lowerFirst(model.SdkName(a.Name)),
		SDKPackage:  call.GoPackage,
		APIStruct:   call.GoApiStruct,
		APIAccessor: "Get" + call.GoApiStruct + strings.TrimPrefix(call.GoPackage, "datadog"),
		Read:        SDKReadView{Method: call.GoMethod, ResponseType: call.GoResponseType},
		Models:      b.models,
		Schema:      SchemaView{Attributes: attrs},
		State:       StateView{Preamble: env.preamble, Assignments: assignments},
	}, nil
}

// flattenedEnvelope is the result of recognizing a singular JSON:API envelope:
// the data.attributes.* leaves rewritten to top-level paths, plus the lookup id
// field/assignment and the "attributes := …" preamble.
type flattenedEnvelope struct {
	leaves   []*model.Attribute
	idField  ModelFieldView
	idAssign StateAssignment
	preamble []string
}

// flattenEnvelope recognizes the singular JSON:API envelope at the response root
// and reshapes it for the walk. It expects a single top-level "data" object whose
// members are a subset of {id, type, attributes}, with "attributes" an object of
// leaves only. It hoists each attribute leaf to a top-level path ("response.<leaf>"),
// surfaces "id" from id_strategy (data.id only), and drops "type". Anything outside
// the recognized envelope is appended to b.unsupported and the result is nil.
func (b *dataSourceBuilder) flattenEnvelope(topLevel []*model.Attribute, idStrategy model.IdStrategy) *flattenedEnvelope {
	if len(topLevel) != 1 || tfNameOf(topLevel[0].Path) != "data" || topLevel[0].TfType != "schema.SingleNestedBlock" {
		b.unsupported = append(b.unsupported, UnsupportedNode{
			Path:   "response",
			Reason: "expected a single-member JSON:API envelope {data:{...}}",
		})
		return nil
	}
	data := topLevel[0]

	// data members must be a subset of {id, type, attributes}.
	var attributes *model.Attribute
	ok := true
	for _, child := range data.Children {
		switch tfNameOf(child.Path) {
		case "id", "type":
			// id is surfaced from id_strategy below; type is the discriminator, dropped.
		case "attributes":
			attributes = child
		default:
			b.unsupported = append(b.unsupported, UnsupportedNode{
				Path:   child.Path,
				Reason: tfNameOf(child.Path) + " is not part of the recognized {id, type, attributes} envelope",
			})
			ok = false
		}
	}

	if attributes == nil {
		b.unsupported = append(b.unsupported, UnsupportedNode{
			Path:   data.Path,
			Reason: "envelope data is missing an attributes object",
		})
		return nil
	}
	if attributes.TfType != "schema.SingleNestedBlock" {
		b.unsupported = append(b.unsupported, UnsupportedNode{
			Path:   attributes.Path,
			Reason: "envelope attributes must be an object",
		})
		return nil
	}

	// Hoist the attribute leaves to top-level paths; anything non-leaf is out of scope.
	leaves := make([]*model.Attribute, 0, len(attributes.Children))
	for _, child := range attributes.Children {
		if !isLeafType(child.TfType) {
			b.unsupported = append(b.unsupported, UnsupportedNode{
				Path:   child.Path,
				Reason: "nesting under attributes is not supported",
			})
			ok = false
			continue
		}
		hoisted := *child
		hoisted.Path = "response." + tfNameOf(child.Path)
		leaves = append(leaves, &hoisted)
	}

	// Only the data.id lookup strategy is supported.
	if idStrategy != model.IdStrategyDataID {
		b.unsupported = append(b.unsupported, UnsupportedNode{
			Path:   data.Path,
			Reason: fmt.Sprintf("id_strategy %q is not yet supported (only data.id)", string(idStrategy)),
		})
		ok = false
	}

	if !ok {
		return nil
	}

	return &flattenedEnvelope{
		leaves:   leaves,
		idField:  ModelFieldView{GoField: "ID", GoType: "types.String", TFName: "id"},
		idAssign: StateAssignment{LHS: "state.ID", RHS: "types.StringValue(resp.Data.GetId())"},
		preamble: []string{"attributes := resp.Data.GetAttributes()"},
	}
}

// dataSourceBuilder accumulates the cross-cutting outputs of one walk: the Go
// model structs (parent before child), the per-leaf state assignments, and the
// unsupported-node collector. receiver is the getter root the state-mapper reads
// leaves off (e.g. "attributes" for the flattened envelope).
type dataSourceBuilder struct {
	receiver    string
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
				RHS: wrapValue(a.GoType, getterChain(b.receiver, a.Path)),
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

// isLeafType reports whether tfType is one of the four scalar attribute forms the
// envelope may hoist directly to a top-level leaf.
func isLeafType(tfType string) bool {
	switch tfType {
	case "schema.StringAttribute", "schema.Int64Attribute",
		"schema.Float64Attribute", "schema.BoolAttribute":
		return true
	default:
		return false
	}
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
// updateState, e.g. "response.name" → "state.Name".
func stateLHS(path string) string {
	segs := responseSegments(path)
	parts := make([]string, len(segs))
	for i, s := range segs {
		parts[i] = model.SdkName(s)
	}
	return "state." + strings.Join(parts, ".")
}

// getterChain builds the SDK getter chain reading an attribute off receiver,
// e.g. getterChain("attributes", "response.name") → "attributes.GetName()".
func getterChain(receiver, path string) string {
	var b strings.Builder
	b.WriteString(receiver)
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
