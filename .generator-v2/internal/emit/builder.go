package emit

import (
	"fmt"
	"sort"
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
	if a.Cardinality == model.CardinalityPlural {
		return buildPluralView(a)
	}

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
		Description: a.Description,
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
				RHS: wrapValue(a, getterChain(b.receiver, a.Path)),
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
// For strings it also reconciles getters whose Go return type is not a bare
// string: a date-time getter returns time.Time (rendered via .String()) and an
// enum getter returns a named string type (cast back with string(...)).
func wrapValue(a *model.Attribute, chain string) string {
	switch a.GoType {
	case "types.String":
		switch {
		case a.Format == "date-time":
			chain += ".String()"
		case a.IsEnum:
			chain = "string(" + chain + ")"
		}
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

// buildPluralView derives the plural DataSourceView from a plural Artifact: the
// scalar query params become Optional filters, and the results-array element
// (a JSON:API envelope) is flattened — "id" read off the loop variable,
// "attributes.*" off item.Attributes, "type" dropped — into one item struct
// projected per element. The walk is fail-slow: unsupported filter or
// item-element nodes are collected and returned together as an
// *UnsupportedEmitError, in which case the view is discarded.
func buildPluralView(a *model.Artifact) (DataSourceView, error) {
	var unsupported []UnsupportedNode

	var call *model.SDKCall
	if a.Lifecycle != nil {
		call = a.Lifecycle.Read
	}
	if call == nil || call.ItemType == "" {
		unsupported = append(unsupported, UnsupportedNode{Path: "response", Reason: "missing list item type"})
	}

	// Partition the top-level schema: Optional leaves are filters, the lone
	// ListNestedBlock is the items block (the model already dropped response
	// metadata siblings, keeping only the results array).
	var filterLeaves []*model.Attribute
	var itemsBlock *model.Attribute
	if a.Schema != nil {
		for _, attr := range a.Schema.Attributes {
			switch {
			case attr.TfType == "schema.ListNestedBlock":
				itemsBlock = attr
			case attr.Optional && isLeafType(attr.TfType):
				filterLeaves = append(filterLeaves, attr)
			default:
				unsupported = append(unsupported, UnsupportedNode{Path: attr.Path, Reason: unsupportedReason(attr.TfType)})
			}
		}
	}
	if itemsBlock == nil {
		unsupported = append(unsupported, UnsupportedNode{Path: "response", Reason: "missing results array block"})
	}

	// Filters: one Optional attribute + model field + param binding each.
	var filterAttrs []AttrView
	var filterFields []ModelFieldView
	var filterParams []FilterParamView
	for i, leaf := range filterLeaves {
		tfName := tfNameOf(leaf.Path)
		filterAttrs = append(filterAttrs, AttrView{
			TFName: tfName, TFType: leaf.TfType, Description: leaf.Description, Optional: true,
		})
		field := ModelFieldView{GoField: model.SdkName(tfName), GoType: leaf.GoType, TFName: tfName}
		if i == 0 {
			field.Comment = "Query Parameters"
		}
		filterFields = append(filterFields, field)
		filterParams = append(filterParams, FilterParamView{
			StateField: model.SdkName(tfName),
			ParamField: model.SdkName(tfName),
			ValueExpr:  pointerValueExpr(leaf.GoType),
		})
	}

	itemLeaves := flattenItemElement(itemsBlock, &unsupported)
	if len(unsupported) > 0 {
		return DataSourceView{}, &UnsupportedEmitError{Nodes: unsupported}
	}

	var itemAttrs []AttrView
	var itemFields []ModelFieldView
	var itemAssigns []StateAssignment
	for _, lf := range itemLeaves {
		tfName := tfNameOf(lf.attr.Path)
		itemAttrs = append(itemAttrs, AttrView{
			TFName: tfName, TFType: lf.attr.TfType, Description: lf.attr.Description, Computed: true,
		})
		itemFields = append(itemFields, ModelFieldView{
			GoField: goFieldName(tfName), GoType: lf.attr.GoType, TFName: tfName,
		})
		itemAssigns = append(itemAssigns, StateAssignment{
			LHS: goFieldName(tfName),
			RHS: wrapValue(lf.attr, lf.chain),
		})
	}

	itemStruct := model.SdkName(call.ItemType) + "Model"
	itemField := model.SdkName(a.Name)
	goName := lowerFirst(model.SdkName(a.Name))

	parentFields := append([]ModelFieldView{}, filterFields...)
	parentFields = append(parentFields,
		ModelFieldView{Comment: "Results", GoField: "ID", GoType: "types.String", TFName: "id"},
		ModelFieldView{GoField: itemField, GoType: "[]*" + itemStruct, TFName: a.Name},
	)

	return DataSourceView{
		Cardinality: Plural,
		TypeName:    a.Name,
		GoName:      goName,
		Description: a.Description,
		SDKPackage:  call.GoPackage,
		APIStruct:   call.GoApiStruct,
		APIAccessor: "Get" + call.GoApiStruct + strings.TrimPrefix(call.GoPackage, "datadog"),
		Read: SDKReadView{
			Method:             call.GoMethod,
			Paginated:          call.Paginated,
			ItemType:           call.ItemType,
			OptionalParamsType: call.OptionalParamsType,
			Filters:            filterParams,
		},
		Models: []ModelStructView{
			{Name: goName + "DataSourceModel", Fields: parentFields},
			{Name: itemStruct, Fields: itemFields},
		},
		Schema: SchemaView{
			Attributes: filterAttrs,
			Blocks: []AttrView{{
				TFName:      a.Name,
				Description: itemsBlock.Description,
				IsBlock:     true,
				ListBlock:   true,
				Attributes:  itemAttrs,
			}},
		},
		State: StateView{
			ItemStruct: itemStruct,
			ItemField:  itemField,
			ItemFields: itemAssigns,
		},
	}, nil
}

// itemElementLeaf is one flattened leaf of a list element: the source attribute
// plus the SDK getter chain that reads it off the loop variable "item".
type itemElementLeaf struct {
	attr  *model.Attribute
	chain string
}

// flattenItemElement recognizes the JSON:API element envelope on a list item
// block and flattens it: "id" is read off the loop variable, each leaf under
// "attributes" off item.Attributes, and "type" is dropped. Non-leaf or
// unexpected members append to unsupported. Leaves are returned sorted by TF name.
func flattenItemElement(block *model.Attribute, unsupported *[]UnsupportedNode) []itemElementLeaf {
	if block == nil {
		return nil
	}
	var leaves []itemElementLeaf
	for _, child := range block.Children {
		switch tfNameOf(child.Path) {
		case "type":
			// discriminator; dropped
		case "id":
			if !isLeafType(child.TfType) {
				*unsupported = append(*unsupported, UnsupportedNode{Path: child.Path, Reason: "item id must be a scalar"})
				continue
			}
			leaves = append(leaves, itemElementLeaf{attr: child, chain: itemGetter("item", tfNameOf(child.Path))})
		case "attributes":
			if child.TfType != "schema.SingleNestedBlock" {
				*unsupported = append(*unsupported, UnsupportedNode{Path: child.Path, Reason: "envelope attributes must be an object"})
				continue
			}
			for _, leaf := range child.Children {
				if !isLeafType(leaf.TfType) {
					*unsupported = append(*unsupported, UnsupportedNode{Path: leaf.Path, Reason: "nesting under item attributes is not supported"})
					continue
				}
				leaves = append(leaves, itemElementLeaf{attr: leaf, chain: itemGetter("item.Attributes", tfNameOf(leaf.Path))})
			}
		default:
			*unsupported = append(*unsupported, UnsupportedNode{
				Path:   child.Path,
				Reason: tfNameOf(child.Path) + " is not part of the recognized {id, type, attributes} envelope",
			})
		}
	}
	sort.Slice(leaves, func(i, j int) bool {
		return tfNameOf(leaves[i].attr.Path) < tfNameOf(leaves[j].attr.Path)
	})
	return leaves
}

// itemGetter builds the SDK getter reading name off receiver for one list
// element, e.g. itemGetter("item.Attributes", "link_count") →
// "item.Attributes.GetLinkCount()".
func itemGetter(receiver, name string) string {
	return receiver + ".Get" + model.SdkName(name) + "()"
}

// goFieldName is the exported Go struct field name for a TF attribute. It is
// model.SdkName except for "id", which becomes "ID" to match Go's initialism
// convention (and the hand-written models) — the SDK getter still reads GetId().
func goFieldName(tfName string) string {
	if tfName == "id" {
		return "ID"
	}
	return model.SdkName(tfName)
}

// pointerValueExpr is the model accessor producing the SDK optional-param value
// for a filter of the given model field type, e.g. "types.String" →
// "ValueStringPointer()".
func pointerValueExpr(goType string) string {
	switch goType {
	case "types.Bool":
		return "ValueBoolPointer()"
	case "types.Int64":
		return "ValueInt64Pointer()"
	case "types.Float64":
		return "ValueFloat64Pointer()"
	default:
		return "ValueStringPointer()"
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
