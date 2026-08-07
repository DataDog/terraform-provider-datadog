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

// oneOfEnvelope renders one union at one use site: it returns the nested variant
// blocks to hang under the union's own block, and the Go type of the field that
// holds the envelope model.
//
// Models and block views are cached on the envelope's parser-assigned Name, so two
// uses of one reusable oneOf component yield a single generated model and
// identical schema rather than one copy per use site. The blocks are
// cached too, not just the models: they are a pure function of the envelope's
// variants, which are shared, so recomputing them could only introduce drift.
//
// A variant's own fields go through the same walk as any nested object, which is
// what produces its model struct and its field assignments. The assignments are
// carried on the returned OneOfVariantView rather than placed here: the mapper has
// to unwrap the SDK member before they can run, and that lands separately.
func (b *dataSourceBuilder) oneOfEnvelope(a *model.Attribute) oneOfRender {
	env := a.OneOf
	if cached, ok := b.oneOfRenders[env.Name]; ok {
		return cached
	}

	// The model layer names an envelope and its variants without knowing which
	// artifact will render them — a reusable union has no single owner — so emit
	// applies the artifact scope here, at the one point that does know.
	envModel := b.namer.qualify(env.GoModel)

	// Reserve the envelope's struct slot before walking the variants so it precedes
	// their structs in Models, matching how walk orders a parent before its children.
	envIdx := len(b.models)
	b.models = append(b.models, ModelStructView{Name: envModel})

	render := oneOfRender{goModel: envModel}
	envFields := make([]ModelFieldView, 0, len(env.Variants))

	for _, v := range env.Variants {
		sdkVar := lowerFirst(v.GoField) + "Variant"
		modelVar := lowerFirst(v.GoField) + "Model"
		variantModel := b.namer.qualify(v.GoModel)

		assign := OneOfVariantAssignment{
			SDKField:   v.SDKField,
			SDKVar:     sdkVar,
			SDKPointer: v.SDKPointer,
			GoField:    v.GoField,
			GoModel:    variantModel,
			ModelVar:   modelVar,
		}

		var attrs, blocks []AttrView
		if v.ValueWrapped {
			// The SDK member of a non-object alternative *is* the value — a *string,
			// not a struct with getters — so its single "value" field is assigned by
			// dereferencing the member rather than reading through Get<Field>Ok.
			attrs, blocks = b.oneOfValueVariant(env, v, variantModel, sdkVar, modelVar, &assign)
		} else {
			var scalars []StateAssignment
			var lists []ListAssignment
			attrs, blocks, scalars, lists = b.walk(variantModel, stemOf(v.GoModel), sdkVar, modelVar, v.Attribute.Children)
			assign.Scalars, assign.Lists = scalars, lists
		}

		render.blocks = append(render.blocks, AttrView{
			TFName:      v.TFName,
			Description: v.Attribute.Description,
			Optional:    v.Attribute.Optional,
			Computed:    v.Attribute.Computed,
			Sensitive:   v.Attribute.Sensitive,
			IsBlock:     true,
			Attributes:  attrs,
			Blocks:      blocks,
		})
		envFields = append(envFields, ModelFieldView{
			GoField: v.GoField,
			GoType:  "*" + variantModel,
			TFName:  v.TFName,
		})
		render.variants = append(render.variants, assign)
	}

	b.models[envIdx].Fields = envFields
	if b.oneOfRenders == nil {
		b.oneOfRenders = make(map[string]oneOfRender)
	}
	b.oneOfRenders[env.Name] = render
	return render
}

// oneOfValueVariant handles a value-wrapped alternative: the variant model has a
// single "value" field, and the SDK member it comes from is the value itself.
//
// The walk cannot serve this case — it would emit SDKVar.GetValueOk(), and a
// *string has no methods — so the schema attribute is built here and the
// assignment dereferences the member directly. A value-wrapped alternative whose
// value is not a scalar (a list or map alternative) has no such dereference and is
// recorded unsupported rather than mis-assigned.
func (b *dataSourceBuilder) oneOfValueVariant(
	env *model.OneOfEnvelope,
	v model.OneOfEnvelopeVariant,
	variantModel string,
	sdkVar, modelVar string,
	assign *OneOfVariantAssignment,
) (attrs, blocks []AttrView) {
	if len(v.Attribute.Children) != 1 {
		b.unsupported = append(b.unsupported, UnsupportedNode{
			Path:   env.Path,
			Reason: fmt.Sprintf("oneOf envelope %q variant %q is value-wrapped but has %d children, expected exactly one", env.Name, v.TFName, len(v.Attribute.Children)),
		})
		return nil, nil
	}
	value := v.Attribute.Children[0]
	if !isLeafType(value.TfType) {
		b.unsupported = append(b.unsupported, UnsupportedNode{
			Path: env.Path,
			Reason: fmt.Sprintf(
				"oneOf envelope %q variant %q wraps a %s, and the emit path can only read a scalar "+
					"straight off an SDK oneOf member; give the alternative a named schema component "+
					"so it becomes an object variant",
				env.Name, v.TFName, value.TfType),
		})
		return nil, nil
	}

	tfName := tfNameOf(value.Path)
	goField := model.SdkName(tfName)

	attrs = []AttrView{{
		TFName:      tfName,
		TFType:      value.TfType,
		Description: value.Description,
		Required:    value.Required,
		Optional:    value.Optional,
		Computed:    value.Computed,
		Sensitive:   value.Sensitive,
	}}
	// The variant still needs its own one-field model struct; walk would normally
	// have appended it, and the envelope's field points at it either way.
	b.models = append(b.models, ModelStructView{
		Name:   variantModel,
		Fields: []ModelFieldView{{GoField: goField, GoType: value.GoType, TFName: tfName}},
	})
	assign.Value = &StateAssignment{
		LHS: modelVar + "." + goField,
		RHS: guardedValue(value, sdkVar),
	}
	return attrs, nil
}

// oneOfRender is one envelope's render-ready output, cached per envelope name.
// Both halves are a function of the envelope alone: the schema blocks because the
// variants are shared, and the variant assignments because they are expressed
// against variant-derived locals and carry GoField rather than a site-specific LHS.
type oneOfRender struct {
	blocks   []AttrView
	variants []OneOfVariantAssignment
	goModel  string
}

// oneOfListAssignment builds the site-specific assignment that unwraps an
// envelope's SDK wrapper into its model, given the render shared across use sites.
func oneOfListAssignment(
	env *model.OneOfEnvelope,
	render oneOfRender,
	tfName, receiver, lhs string,
	collection bool,
) ListAssignment {
	outer := leafVar(tfName)
	assignment := &OneOfAssignment{
		Path:       env.Path,
		SDKType:    env.SDKType,
		GoModel:    render.goModel,
		LHS:        lhs,
		GetterOk:   getterOk(receiver, tfName),
		Var:        outer,
		Receiver:   outer,
		ModelVar:   outer + "Envelope",
		MatchVar:   outer + "Matches",
		Optional:   env.Optional,
		Collection: collection,
		Variants:   render.variants,
	}
	if collection {
		// The members are read off each element, not off the slice pointer.
		assignment.LoopVar = outer + "Item"
		assignment.Receiver = assignment.LoopVar
	}
	return ListAssignment{Kind: "oneof", LHS: lhs, GetterOk: assignment.GetterOk, Var: outer, OneOf: assignment}
}

// oneOfFieldType returns the Go type of the model field holding an envelope,
// given the Terraform form of the attribute the envelope hangs off: a pointer for
// a union at its own position, a slice for a collection whose element is a union.
// It reports false for a form the emit path does not represent yet, so the caller
// fails the artifact rather than dropping the union.
func oneOfFieldType(tfType, envelopeModel string) (string, bool) {
	switch tfType {
	case "schema.SingleNestedBlock", "schema.SingleNestedAttribute":
		return "*" + envelopeModel, true
	case "schema.ListNestedBlock", "schema.ListNestedAttribute":
		return "[]*" + envelopeModel, true
	default:
		return "", false
	}
}

// isCollectionForm reports whether an envelope-carrying attribute is the
// collection rather than the union itself, so each element is one envelope.
func isCollectionForm(tfType string) bool {
	return tfType == "schema.ListNestedBlock" || tfType == "schema.ListNestedAttribute"
}

// unsupportedOneOfPlacement explains a union the emit path cannot place yet, named
// by the Terraform form it arrived in rather than by its schema kind, since that is
// what the emitter would have had to render.
func unsupportedOneOfPlacement(env *model.OneOfEnvelope, tfType string) UnsupportedNode {
	return UnsupportedNode{
		Path: env.Path,
		Reason: fmt.Sprintf(
			"oneOf envelope %q sits on a %s, which the emit path does not represent yet; "+
				"a union is supported at its own position and as a list element",
			env.Name, tfType,
		),
	}
}

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

	b := &dataSourceBuilder{receiver: envelopeReceiver, namer: modelNamer{base: dsGoName(a.Name)}}

	// Resolve the SDK calls. read backs the by-id lookup, search the list; the
	// presence of each selects the resolution shape (read-only / search / both).
	var read, search *model.SDKCall
	var idStrategy model.IdStrategy
	if a.Lifecycle != nil {
		read, search = a.Lifecycle.Read, a.Lifecycle.Search
		idStrategy = a.Lifecycle.IdStrategy
	}
	byID, searchable := read != nil, search != nil

	// The primary call provides the SDK package/struct the data source binds to:
	// the by-id call when present, otherwise the list call.
	primary := read
	if primary == nil {
		primary = search
	}
	if primary == nil {
		b.unsupported = append(b.unsupported, UnsupportedNode{Path: "response", Reason: "no read or search SDK call resolved"})
	}

	// The record is read off a by-id response (read-only) or a list element
	// (search/both); rootExpr is what the state mapper reads id and attributes off.
	rootExpr, paramName, paramType := "resp.Data", "resp", ""
	if searchable {
		// The record is a list element, passed by value (resp.GetData() / items[i]).
		rootExpr, paramName = "data", "data"
		if search.ItemType == "" {
			b.unsupported = append(b.unsupported, UnsupportedNode{Path: "response", Reason: "missing search item type"})
		} else {
			paramType = search.GoPackage + "." + search.ItemType
		}
	} else if read == nil || read.GoResponseType == "" {
		b.unsupported = append(b.unsupported, UnsupportedNode{Path: "response", Reason: "missing response type name"})
	} else {
		paramType = "*" + read.GoPackage + "." + read.GoResponseType
	}

	// Partition the schema: Optional leaves are the search filters, the lone
	// envelope block is the record to flatten.
	var topLevel, filterLeaves []*model.Attribute
	if a.Schema != nil {
		for _, attr := range a.Schema.Attributes {
			if attr.Optional && isLeafType(attr.TfType) {
				filterLeaves = append(filterLeaves, attr)
			} else {
				topLevel = append(topLevel, attr)
			}
		}
	}

	rootStruct := b.namer.qualify("DataSourceModel")
	env := b.flattenEnvelope(topLevel, idStrategy, rootExpr)

	if len(b.unsupported) > 0 {
		return DataSourceView{}, &UnsupportedEmitError{Nodes: b.unsupported}
	}

	// env is non-nil here: flattenEnvelope records an unsupported node (caught
	// above) on every failure path. Walk the hoisted leaves into the root struct.
	// The root's stem is empty: its inline children accumulate from the artifact base
	// alone, so a top-level "foo" object becomes <base>FooModel.
	recordAttrs, recordBlocks, recordScalars, recordLists := b.walk(rootStruct, "", b.receiver, "state", env.leaves)
	leafFields := b.models[0].Fields

	// Search filters: one Optional attribute + model field + param binding each.
	filterAttrs, filterFields, filterParams := buildSingularFilters(filterLeaves)

	// Parent model fields: the lookup id, then the search filters, then the record
	// leaves. The group comments are only emitted for the search shapes.
	idField := env.idField
	if searchable {
		idField.Comment = "Datasource ID"
		if len(leafFields) > 0 {
			leafFields[0].Comment = "Computed values"
		}
	}
	fields := append([]ModelFieldView{idField}, filterFields...)
	b.models[0].Fields = append(fields, leafFields...)

	models, conflicts := dedupeModels(b.models)
	if len(conflicts) > 0 {
		return DataSourceView{}, &UnsupportedEmitError{Nodes: conflicts}
	}

	assignments := append([]StateAssignment{env.idAssign}, recordScalars...)

	var readView, searchView SDKReadView
	if byID {
		readView = SDKReadView{Method: read.GoMethod, ResponseType: read.GoResponseType}
	}
	if searchable {
		searchView = SDKReadView{
			Method:             search.GoMethod,
			Paginated:          search.Paginated,
			ItemType:           search.ItemType,
			OptionalParamsType: search.OptionalParamsType,
			Filters:            filterParams,
		}
	}

	return DataSourceView{
		Cardinality: Singular,
		TypeName:    a.Name,
		GoName:      dsGoName(a.Name),
		Description: a.Description,
		SDKPackage:  primary.GoPackage,
		APIStruct:   primary.GoApiStruct,
		APIAccessor: "Get" + primary.GoApiStruct + strings.TrimPrefix(primary.GoPackage, "datadog"),
		ByID:        byID,
		Searchable:  searchable,
		Read:        readView,
		Search:      searchView,
		Models:      models,
		Schema:      SchemaView{Attributes: append(filterAttrs, recordAttrs...), Blocks: recordBlocks},
		State: StateView{
			ParamName:   paramName,
			ParamType:   paramType,
			Preamble:    env.preamble,
			Assignments: assignments,
			Lists:       recordLists,
		},
		UsesFmt: !searchable || len(b.oneOfRenders) > 0,
		Dropped: b.dropped,
	}, nil
}

// buildSingularFilters turns the Optional filter leaves of a search/both data
// source into Terraform attributes, model fields, and the request-param bindings
// that set the list call's optional parameters — mirroring the plural filter set.
func buildSingularFilters(leaves []*model.Attribute) (attrs []AttrView, fields []ModelFieldView, params []FilterParamView) {
	for i, leaf := range leaves {
		tfName := tfNameOf(leaf.Path)
		attrs = append(attrs, AttrView{TFName: tfName, TFType: leaf.TfType, Description: leaf.Description, Optional: true})
		field := ModelFieldView{GoField: model.SdkName(tfName), GoType: leaf.GoType, TFName: tfName}
		if i == 0 {
			field.Comment = "Query Parameters"
		}
		fields = append(fields, field)
		params = append(params, FilterParamView{
			StateField: model.SdkName(tfName),
			ParamField: model.SdkName(tfName),
			ValueExpr:  pointerValueExpr(leaf.GoType),
		})
	}
	return attrs, fields, params
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
// and reshapes it for the walk. It expects a top-level "data" object whose members
// are a subset of {id, type, attributes}, with "attributes" an object of leaves
// only. It hoists each attribute leaf to a top-level path ("response.<leaf>"),
// surfaces "id" from id_strategy (data.id only), and drops "type". Anything outside
// the recognized envelope is appended to b.unsupported and the result is nil.
func (b *dataSourceBuilder) flattenEnvelope(topLevel []*model.Attribute, idStrategy model.IdStrategy, rootExpr string) *flattenedEnvelope {
	// A JSON:API response carries sideloading and request metadata beside the
	// primary resource: "included" (the hydrated targets of data.relationships),
	// "meta", "links". None of it is the resource's own state — no hand-written
	// artifact in the provider surfaces any of it — so drop those siblings the same
	// way data.relationships is dropped below, rather than failing on a shape that
	// is in fact recognized. Dropping "included" also keeps its element union out
	// of the view; that union is the reason these responses need a decision at all.
	var data *model.Attribute
	for _, attr := range topLevel {
		if tfNameOf(attr.Path) == "data" {
			data = attr
			continue
		}
		b.dropped = append(b.dropped, droppedEnvelopeMember(attr.Path))
	}
	if data == nil || data.TfType != "schema.SingleNestedBlock" {
		b.unsupported = append(b.unsupported, UnsupportedNode{
			Path:   "response",
			Reason: "expected a JSON:API envelope with a data object {data:{...}}",
		})
		return nil
	}

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
			// Members outside {id, type, attributes} (e.g. relationships) have no
			// place in the attributes-only view; drop them rather than failing.
			b.dropped = append(b.dropped, droppedEnvelopeMember(child.Path))
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

	// Hoist the attribute children to top-level paths: scalar leaves, array nodes
	// (list-of-primitive / list-of-object), and bare nested objects are in scope;
	// a map is not.
	leaves := make([]*model.Attribute, 0, len(attributes.Children))
	for _, child := range attributes.Children {
		if isAuditField(tfNameOf(child.Path)) {
			b.dropped = append(b.dropped, droppedAuditField(child.Path))
			continue
		}
		// The envelope id is surfaced unconditionally below, so an id under
		// attributes always collides with it.
		if tfNameOf(child.Path) == "id" {
			b.dropped = append(b.dropped, droppedIDCollision(child.Path))
			continue
		}
		if !isLeafType(child.TfType) && !isArrayType(child.TfType) && !isObjectType(child.TfType) {
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
		leaves:  leaves,
		idField: ModelFieldView{GoField: "ID", GoType: "types.String", TFName: "id"},
		idAssign: StateAssignment{
			Var:      "id",
			GetterOk: rootExpr + ".GetIdOk()",
			LHS:      "state.ID",
			RHS:      "types.StringValue(*id)",
		},
		preamble: []string{"attributes := " + rootExpr + ".GetAttributes()"},
	}
}

// dataSourceBuilder accumulates the cross-cutting outputs of one walk: the Go
// model structs (parent before child), the per-leaf state assignments, and the
// unsupported-node collector. receiver is the getter root the state-mapper reads
// leaves off (e.g. "attributes" for the flattened envelope).
type dataSourceBuilder struct {
	receiver string
	// namer scopes every model struct name to this artifact and derives nested
	// names from their OpenAPI component. See modelname.go.
	namer       modelNamer
	models      []ModelStructView
	unsupported []UnsupportedNode
	// dropped notes envelope members skipped from the attributes-only view
	// (e.g. relationships), surfaced as diagnostics rather than failures.
	dropped []DroppedMember
	// oneOfBlocks caches the variant block views of each envelope already rendered,
	// keyed on the parser's envelope Name, and doubles as the "already emitted its
	// models" marker. envelopes accumulates one view per distinct envelope in first-use
	// order, for the state mapper to walk.
	oneOfRenders map[string]oneOfRender
}

// droppedEnvelopeMember is the info-diagnostic note for a JSON:API response
// member skipped from the attributes-only view, e.g. relationships.
func droppedEnvelopeMember(path string) DroppedMember {
	return DroppedMember{
		Message:  fmt.Sprintf("dropped %q: not part of the surfaced {id, type, attributes} envelope", path),
		Severity: model.SeverityInfo,
	}
}

// droppedAuditField is the info-diagnostic note for a top-level audit attribute
// omitted from a generated data source.
func droppedAuditField(path string) DroppedMember {
	return DroppedMember{
		Message:  fmt.Sprintf("dropped %q: server-managed audit field", path),
		Severity: model.SeverityInfo,
	}
}

// droppedIDCollision is the warning-diagnostic note for a member promoted out of
// attributes whose name is already claimed by the envelope id. Flattening would
// emit two attributes named "id"; the envelope id wins because it carries the
// data source's Terraform identity.
func droppedIDCollision(path string) DroppedMember {
	return DroppedMember{
		Message:  fmt.Sprintf("dropped %q: collides with the envelope id surfaced as \"id\"", path),
		Severity: model.SeverityWarning,
	}
}

// walk processes one struct's worth of attributes in tree order, reserving the
// struct's slot in b.models up front so a parent precedes its children, then
// filling its fields as it goes. receiver is the SDK getter root the state mapper
// reads off ("attributes" at the record root, the loop variable inside a list
// element); lhsPrefix is the model target the assignments write into ("state", or
// the per-element accumulator). It returns the schema attr/block views plus the
// scalar and list state assignments for the caller to place, recursing through
// nested blocks.
//
// stem is this struct's own naming stem, which its children accumulate from when
// their schema is inline (modelname.go). It is separate from structName because
// structName is already qualified and suffixed, and a stem must be neither.
//
// The same struct may be reserved twice when one OpenAPI component is reached from
// two properties: each site needs its own assignments, which are expressed against
// site-specific receivers and LHS prefixes, so both walks run. dedupeModels
// collapses the duplicate declarations afterwards.
func (b *dataSourceBuilder) walk(structName, stem, receiver, lhsPrefix string, attrs []*model.Attribute) (attrViews, blockViews []AttrView, scalars []StateAssignment, lists []ListAssignment) {
	idx := len(b.models)
	b.models = append(b.models, ModelStructView{Name: structName})
	var fields []ModelFieldView

	for _, a := range attrs {
		tfName := tfNameOf(a.Path)
		field := model.SdkName(tfName)

		// A union is keyed on OneOf, not on TfType: an envelope arrives wearing the
		// same schema.SingleNestedBlock as an ordinary nested object, and walking it
		// as one would emit Get<Variant>Ok getters the SDK oneOf wrapper does not have.
		if a.OneOf != nil {
			render := b.oneOfEnvelope(a)
			goType, ok := oneOfFieldType(a.TfType, render.goModel)
			if !ok {
				b.unsupported = append(b.unsupported, unsupportedOneOfPlacement(a.OneOf, a.TfType))
				continue
			}
			collection := isCollectionForm(a.TfType)
			fields = append(fields, ModelFieldView{GoField: field, GoType: goType, TFName: tfName})
			blockViews = append(blockViews, AttrView{
				TFName:      tfName,
				Description: a.Description,
				Required:    a.Required,
				Optional:    a.Optional,
				Computed:    a.Computed,
				Sensitive:   a.Sensitive,
				IsBlock:     true,
				ListBlock:   collection,
				Blocks:      render.blocks,
			})
			lists = append(lists, oneOfListAssignment(
				a.OneOf, render, tfName, receiver, lhsPrefix+"."+field, collection))
			continue
		}

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
			fields = append(fields, ModelFieldView{GoField: field, GoType: a.GoType, TFName: tfName})
			varName := leafVar(tfName)
			scalars = append(scalars, StateAssignment{
				Var:      varName,
				GetterOk: getterOk(receiver, tfName),
				LHS:      lhsPrefix + "." + field,
				RHS:      guardedValue(a, varName),
			})

		case "schema.ListAttribute":
			attrViews = append(attrViews, AttrView{
				TFName:      tfName,
				TFType:      a.TfType,
				ElementType: a.ElementType,
				Description: a.Description,
				Required:    a.Required,
				Optional:    a.Optional,
				Computed:    a.Computed,
				Sensitive:   a.Sensitive,
			})
			fields = append(fields, ModelFieldView{GoField: field, GoType: a.GoType, TFName: tfName}) // types.List
			lists = append(lists, ListAssignment{
				Kind:        "primitive",
				LHS:         lhsPrefix + "." + field,
				GetterOk:    getterOk(receiver, tfName),
				Var:         leafVar(tfName),
				ElementType: a.ElementType,
			})

		case "schema.ListNestedBlock":
			elemStruct, childStem := b.namer.nested(stem, a)
			base := lowerFirst(field)
			loopVar, elemVar := base+"Item", base+"Model"
			fields = append(fields, ModelFieldView{GoField: field, GoType: "[]*" + elemStruct, TFName: tfName})
			childAttrs, childBlocks, childScalars, childLists := b.walk(elemStruct, childStem, loopVar, elemVar, a.Children)
			blockViews = append(blockViews, AttrView{
				TFName:      tfName,
				Description: a.Description,
				Required:    a.Required,
				Optional:    a.Optional,
				Computed:    a.Computed,
				Sensitive:   a.Sensitive,
				IsBlock:     true,
				ListBlock:   true,
				Attributes:  childAttrs,
				Blocks:      childBlocks,
			})
			lists = append(lists, ListAssignment{
				Kind:       "object",
				LHS:        lhsPrefix + "." + field,
				GetterOk:   getterOk(receiver, tfName),
				Var:        leafVar(tfName),
				LoopVar:    loopVar,
				ElemVar:    elemVar,
				ElemStruct: elemStruct,
				Scalars:    childScalars,
				Lists:      childLists,
			})

		case "schema.SingleNestedBlock":
			childStruct, childStem := b.namer.nested(stem, a)
			objVar := leafVar(tfName)
			elemVar := lowerFirst(field) + "Model"
			fields = append(fields, ModelFieldView{GoField: field, GoType: "*" + childStruct, TFName: tfName})
			childAttrs, childBlocks, childScalars, childLists := b.walk(childStruct, childStem, objVar, elemVar, a.Children)
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
			lists = append(lists, ListAssignment{
				Kind:       "object_single",
				LHS:        lhsPrefix + "." + field,
				GetterOk:   getterOk(receiver, tfName),
				Var:        objVar,
				ElemVar:    elemVar,
				ElemStruct: childStruct,
				Scalars:    childScalars,
				Lists:      childLists,
			})

		default:
			b.unsupported = append(b.unsupported, UnsupportedNode{Path: a.Path, Reason: unsupportedReason(a.TfType)})
		}
	}

	b.models[idx].Fields = fields
	return attrViews, blockViews, scalars, lists
}

// getterOk builds the SDK optional getter reading name off receiver, e.g.
// getterOk("attributes", "visible_modules") → "attributes.GetVisibleModulesOk()".
func getterOk(receiver, name string) string {
	return receiver + ".Get" + model.SdkName(name) + "Ok()"
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

// isArrayType reports whether tfType is one of the array attribute forms the
// envelope hoists into list machinery: a collection-of-primitive (ListAttribute)
// or an array-of-object (ListNestedBlock).
func isArrayType(tfType string) bool {
	switch tfType {
	case "schema.ListAttribute", "schema.ListNestedBlock":
		return true
	default:
		return false
	}
}

// isObjectType reports whether tfType is a bare nested object the envelope
// hoists into single-object machinery (schema.SingleNestedBlock).
func isObjectType(tfType string) bool { return tfType == "schema.SingleNestedBlock" }

// auditFields are server-managed audit attributes dropped from the top level of a
// generated data source: their timestamps and actor handles add schema noise
// without configuration value.
var auditFields = map[string]bool{
	"created_at":  true,
	"updated_at":  true,
	"created_by":  true,
	"updated_by":  true,
	"modified_at": true,
}

// isAuditField reports whether a top-level record attribute is server-managed
// audit metadata the emit path drops rather than surfacing.
func isAuditField(tfName string) bool { return auditFields[tfName] }

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

// leafVar is the local variable a guarded assignment binds the optional getter's
// value to: the attribute's lowerCamel name, suffixed with "Value" when it would
// shadow an identifier in updateState's scope (state, attributes, the receiver).
func leafVar(tfName string) string {
	v := lowerFirst(model.SdkName(tfName))
	switch v {
	case "state", "attributes", "ok", "d", "data", "resp", "items", "ctx":
		return v + "Value"
	}
	// A property named "type" — common in JSON:API payloads and in every
	// action_connection oneOf variant — would otherwise emit `if type, ok := ...`,
	// which is not Go. The SDK generator solves this by suffixing "Var"; reuse its
	// rule rather than a second scheme. The shadow suffix above stays "Value"
	// because it answers a different question (colliding with a local the template
	// declares, not with the language).
	return model.EscapeReservedKeyword(v)
}

// guardedValue wraps a guarded assignment's local (bound from an Ok-getter, so a
// pointer) in the types.*Value constructor matching the model field's GoType. A
// date-time pointer renders via .String(); a named enum pointer is dereferenced
// and cast back to string; integers are cast to int64 as the framework expects.
func guardedValue(a *model.Attribute, varName string) string {
	switch a.GoType {
	case "types.String":
		switch {
		case a.Format == "date-time":
			return "types.StringValue(" + varName + ".String())"
		case a.IsEnum:
			return "types.StringValue(string(*" + varName + "))"
		default:
			return "types.StringValue(*" + varName + ")"
		}
	case "types.Bool":
		return "types.BoolValue(*" + varName + ")"
	case "types.Int64":
		return "types.Int64Value(int64(*" + varName + "))"
	case "types.Float64":
		return "types.Float64Value(*" + varName + ")"
	default:
		return "*" + varName
	}
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
	var dropped []DroppedMember

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

	// b hosts walk so list-of-object item fields generate their element structs.
	b := &dataSourceBuilder{namer: modelNamer{base: dsGoName(a.Name)}}
	scalarLeaves, nonScalars := flattenItemElement(itemsBlock, &unsupported, &dropped)

	// The item element's own stem. call.ItemType is the response element's component
	// name, so it plays the same role here that ModelRefName plays under walk: the
	// item struct is named after its component, and the item's inline children
	// accumulate from it.
	// Verbatim, like any other component name: ItemType is already a Go type name
	// (it is spelled as call.GoPackage + "." + call.ItemType elsewhere), so SdkName
	// would only mangle its acronyms.
	var itemStem string
	if call != nil {
		itemStem = call.ItemType
	}
	itemStruct := b.namer.qualify(itemStem + "Model")

	// Scalar leaves project into the item struct literal, unguarded, off the loop
	// variable "item".
	var itemAttrs []AttrView
	var itemFields []ModelFieldView
	var itemAssigns []StateAssignment
	for _, lf := range scalarLeaves {
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

	// Non-scalar attributes append after the scalars and map after the literal via
	// ItemLists, read off item.Attributes: a list-of-primitive as a ListAttribute,
	// a list-of-object as a ListNestedBlock, and a bare object as a
	// SingleNestedBlock — each with its element struct walked.
	var itemBlocks []AttrView
	var itemLists []ListAssignment
	for _, n := range nonScalars {
		tfName := tfNameOf(n.Path)
		field := goFieldName(tfName)
		getter := getterOk("item.Attributes", tfName)

		// Same rule as walk: a union is keyed on OneOf, never on TfType.
		if n.OneOf != nil {
			render := b.oneOfEnvelope(n)
			goType, ok := oneOfFieldType(n.TfType, render.goModel)
			if !ok {
				unsupported = append(unsupported, unsupportedOneOfPlacement(n.OneOf, n.TfType))
				continue
			}
			collection := isCollectionForm(n.TfType)
			itemFields = append(itemFields, ModelFieldView{GoField: field, GoType: goType, TFName: tfName})
			itemBlocks = append(itemBlocks, AttrView{
				TFName:      tfName,
				Description: n.Description,
				Computed:    true,
				IsBlock:     true,
				ListBlock:   collection,
				Blocks:      render.blocks,
			})
			itemLists = append(itemLists, oneOfListAssignment(
				n.OneOf, render, tfName, "item.Attributes", "r."+field, collection))
			continue
		}

		switch n.TfType {
		case "schema.ListAttribute":
			itemAttrs = append(itemAttrs, AttrView{
				TFName: tfName, TFType: n.TfType, ElementType: n.ElementType, Description: n.Description, Computed: true,
			})
			itemFields = append(itemFields, ModelFieldView{GoField: field, GoType: n.GoType, TFName: tfName})
			itemLists = append(itemLists, ListAssignment{
				Kind: "primitive", LHS: "r." + field, GetterOk: getter, Var: leafVar(tfName), ElementType: n.ElementType,
			})
		case "schema.ListNestedBlock":
			// itemStem, not "": these hang off the item element, so an inline child
			// accumulates from the item struct the same way it would under walk.
			elemStruct, childStem := b.namer.nested(itemStem, n)
			base := lowerFirst(model.SdkName(tfName))
			loopVar, elemVar := base+"Item", base+"Model"
			childAttrs, childBlocks, childScalars, childLists := b.walk(elemStruct, childStem, loopVar, elemVar, n.Children)
			itemBlocks = append(itemBlocks, AttrView{
				TFName: tfName, Description: n.Description,
				IsBlock: true, ListBlock: true, Attributes: childAttrs, Blocks: childBlocks,
			})
			itemFields = append(itemFields, ModelFieldView{GoField: field, GoType: "[]*" + elemStruct, TFName: tfName})
			itemLists = append(itemLists, ListAssignment{
				Kind: "object", LHS: "r." + field, GetterOk: getter, Var: leafVar(tfName),
				LoopVar: loopVar, ElemVar: elemVar, ElemStruct: elemStruct,
				Scalars: childScalars, Lists: childLists,
			})
		case "schema.SingleNestedBlock":
			elemStruct, childStem := b.namer.nested(itemStem, n)
			objVar := leafVar(tfName)
			elemVar := lowerFirst(model.SdkName(tfName)) + "Model"
			childAttrs, childBlocks, childScalars, childLists := b.walk(elemStruct, childStem, objVar, elemVar, n.Children)
			itemBlocks = append(itemBlocks, AttrView{
				TFName: tfName, Description: n.Description,
				IsBlock: true, ListBlock: false, Attributes: childAttrs, Blocks: childBlocks,
			})
			itemFields = append(itemFields, ModelFieldView{GoField: field, GoType: "*" + elemStruct, TFName: tfName})
			itemLists = append(itemLists, ListAssignment{
				Kind: "object_single", LHS: "r." + field, GetterOk: getter, Var: objVar,
				ElemVar: elemVar, ElemStruct: elemStruct,
				Scalars: childScalars, Lists: childLists,
			})

		default:
			// This switch had no default, so an item attribute in any other form was
			// dropped without a field, a block or a diagnostic. Fail instead: silently
			// omitting a representable attribute is not an acceptable outcome.
			unsupported = append(unsupported, UnsupportedNode{Path: n.Path, Reason: unsupportedReason(n.TfType)})
		}
	}

	unsupported = append(unsupported, b.unsupported...)
	if len(unsupported) > 0 {
		return DataSourceView{}, &UnsupportedEmitError{Nodes: unsupported}
	}

	itemField := model.SdkName(a.Name)
	goName := dsGoName(a.Name)

	parentFields := append([]ModelFieldView{}, filterFields...)
	parentFields = append(parentFields,
		ModelFieldView{Comment: "Results", GoField: "ID", GoType: "types.String", TFName: "id"},
		ModelFieldView{GoField: itemField, GoType: "[]*" + itemStruct, TFName: a.Name},
	)

	// Models: parent, the item struct, then any element structs walked for
	// list-of-object item fields.
	models := []ModelStructView{
		{Name: b.namer.qualify("DataSourceModel"), Fields: parentFields},
		{Name: itemStruct, Fields: itemFields},
	}
	models = append(models, b.models...)
	models, conflicts := dedupeModels(models)
	if len(conflicts) > 0 {
		return DataSourceView{}, &UnsupportedEmitError{Nodes: conflicts}
	}

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
		Models: models,
		Schema: SchemaView{
			Attributes: filterAttrs,
			Blocks: []AttrView{{
				TFName:      a.Name,
				Description: itemsBlock.Description,
				IsBlock:     true,
				ListBlock:   true,
				Attributes:  itemAttrs,
				Blocks:      itemBlocks,
			}},
		},
		State: StateView{
			ItemStruct: itemStruct,
			ItemField:  itemField,
			ItemFields: itemAssigns,
			ItemLists:  itemLists,
		},
		UsesFmt: len(filterParams) > 0 || len(b.oneOfRenders) > 0,
		Dropped: dropped,
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
// "attributes" off item.Attributes, and "type" is dropped. Members outside
// {id, type, attributes} (e.g. relationships) are dropped with a note on dropped;
// non-leaf id/attributes still append to unsupported. Leaves are sorted by TF name.
func flattenItemElement(block *model.Attribute, unsupported *[]UnsupportedNode, dropped *[]DroppedMember) (scalars []itemElementLeaf, nonScalars []*model.Attribute) {
	if block == nil {
		return nil, nil
	}
	// Children are name-sorted, so "attributes" is walked before "id". Look the
	// envelope id up first so a colliding id under attributes can be dropped.
	envelopeHasID := false
	for _, child := range block.Children {
		if tfNameOf(child.Path) == "id" {
			envelopeHasID = true
		}
	}
	for _, child := range block.Children {
		switch tfNameOf(child.Path) {
		case "type":
			// discriminator; dropped
		case "id":
			if !isLeafType(child.TfType) {
				*unsupported = append(*unsupported, UnsupportedNode{Path: child.Path, Reason: "item id must be a scalar"})
				continue
			}
			scalars = append(scalars, itemElementLeaf{attr: child, chain: itemGetter("item", tfNameOf(child.Path))})
		case "attributes":
			if child.TfType != "schema.SingleNestedBlock" {
				*unsupported = append(*unsupported, UnsupportedNode{Path: child.Path, Reason: "envelope attributes must be an object"})
				continue
			}
			for _, leaf := range child.Children {
				if isAuditField(tfNameOf(leaf.Path)) {
					*dropped = append(*dropped, droppedAuditField(leaf.Path))
					continue
				}
				if envelopeHasID && tfNameOf(leaf.Path) == "id" {
					*dropped = append(*dropped, droppedIDCollision(leaf.Path))
					continue
				}
				switch {
				case isLeafType(leaf.TfType):
					scalars = append(scalars, itemElementLeaf{attr: leaf, chain: itemGetter("item.Attributes", tfNameOf(leaf.Path))})
				case isArrayType(leaf.TfType), isObjectType(leaf.TfType):
					nonScalars = append(nonScalars, leaf)
				default:
					*unsupported = append(*unsupported, UnsupportedNode{Path: leaf.Path, Reason: "nesting under item attributes is not supported"})
				}
			}
		default:
			// Members outside {id, type, attributes} (e.g. relationships) have no
			// place in the attributes-only view; drop them rather than failing.
			*dropped = append(*dropped, droppedEnvelopeMember(child.Path))
		}
	}
	sort.Slice(scalars, func(i, j int) bool {
		return tfNameOf(scalars[i].attr.Path) < tfNameOf(scalars[j].attr.Path)
	})
	// nonScalars keep the sorted order of attributes.Children (buildChildren sorts keys).
	return scalars, nonScalars
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

// dsGoName is the lowerCamel identifier base for a generated data source, e.g.
// "datadogTeam". Every generated data source is uniformly Datadog-prefixed so the
// struct <dsGoName>DataSource, model <dsGoName>DataSourceModel, and constructor
// New<title dsGoName>DataSource match the provider's convention.
func dsGoName(name string) string {
	return lowerFirst("Datadog" + model.SdkName(name))
}
