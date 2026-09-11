package emit

import (
	"fmt"
	"go/token"
	"slices"
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
// artifact's tree: valid Terraform the emit path does not implement yet.
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

// envelopeReceiver is the local the flattened state mapper reads hoisted leaves
// off, declared by the StateView preamble.
const envelopeReceiver = "attributes"

// oneOfEnvelope renders one union: the variant blocks to hang under the union's
// attribute, plus the Go model name of the envelope struct. The whole render is
// cached on the envelope's Name, so two uses of one reusable oneOf component
// yield a single model and identical schema. A variant's own fields go through
// walk like any nested object; its assignments ride on the returned
// OneOfVariantAssignment because the mapper must unwrap the SDK member first.
func (b *dataSourceBuilder) oneOfEnvelope(a *model.Attribute) oneOfRender {
	env := a.OneOf
	// Schema views containing write-only configuration carry an absolute
	// ParentBlocks path, so the same reusable union reached at two resource
	// locations needs one render per location. Model declarations remain
	// deduplicated by dedupeModels after the walk.
	cacheKey := env.Name + "\x00" + a.Path
	if cached, ok := b.oneOfRenders[cacheKey]; ok {
		return cached
	}

	// The model layer names an envelope without knowing which artifact renders
	// it — a reusable union has no single owner — so scope the name here.
	envModel := b.namer.qualify(env.GoModel)

	// Reserve the envelope's struct slot before walking the variants so it
	// precedes theirs in Models, as walk orders a parent before its children.
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
			// dereferencing the member, not through Get<Field>Ok.
			attrs, blocks = b.oneOfValueVariant(env, v, variantModel, sdkVar, modelVar, &assign)
		} else {
			var scalars []StateAssignment
			var lists []ListAssignment
			attrs, blocks, scalars, lists = b.walk(variantModel, stemOf(v.GoModel), sdkVar, modelVar, v.Attribute.Children)
			assign.Scalars, assign.Lists = scalars, lists
		}

		variantValidators, variantValidatorType := b.oneOfVariantValidators(env, v)
		render.blocks = append(render.blocks, AttrView{
			TFName:              v.TFName,
			Description:         v.Attribute.Description,
			Optional:            v.Attribute.Optional,
			Computed:            v.Attribute.Computed,
			Sensitive:           v.Attribute.Sensitive,
			IsBlock:             true,
			Attributes:          attrs,
			Blocks:              blocks,
			HasWriteOnlySecrets: hasWriteOnlySecrets(attrs),
			Validators:          variantValidators,
			ValidatorType:       variantValidatorType,
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
	b.oneOfRenders[cacheKey] = render
	return render
}

// oneOfValueVariant handles a value-wrapped alternative, whose variant model
// has a single "value" field read straight off the SDK member. walk cannot
// serve it — it would emit SDKVar.GetValueOk() on a *string — so the attribute
// is built here and the assignment dereferences the member. A non-scalar value
// (a list or map alternative) has nothing to dereference: unsupported.
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
	// The variant still needs the one-field model struct walk would have
	// appended; the envelope's field points at it either way.
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
// Both halves depend on the envelope alone: the variants are shared, and their
// assignments name variant-derived locals and a GoField, never a site LHS.
type oneOfRender struct {
	blocks   []AttrView
	variants []OneOfVariantAssignment
	goModel  string
}

// oneOfListAssignment builds the site-specific assignment that unwraps an
// envelope's SDK wrapper into its model, from the render shared across sites.
func oneOfListAssignment(
	env *model.OneOfEnvelope,
	render oneOfRender,
	tfName, receiver, lhs string,
	collection bool,
	preserveExisting bool,
) ListAssignment {
	outer := leafVar(tfName)
	assignment := &OneOfAssignment{
		Path:             env.Path,
		SDKType:          env.SDKType,
		GoModel:          render.goModel,
		LHS:              lhs,
		GetterOk:         getterOk(receiver, tfName),
		Var:              outer,
		Receiver:         outer,
		ModelVar:         outer + "Envelope",
		MatchVar:         outer + "Matches",
		Optional:         env.Optional,
		PreserveExisting: preserveExisting,
		Collection:       collection,
		Variants:         render.variants,
	}
	if collection {
		// The members are read off each element, not off the slice pointer.
		assignment.LoopVar = outer + "Item"
		assignment.LoopIndex = outer + "Index"
		assignment.ExistingVar = outer + "Existing"
		assignment.Receiver = assignment.LoopVar
	}
	return ListAssignment{Kind: "oneof", LHS: lhs, GetterOk: assignment.GetterOk, Var: outer, OneOf: assignment}
}

// oneOfFieldType returns the Go type of the model field holding an envelope,
// from the Terraform form the envelope hangs off: a pointer for a union at its
// own position, a slice for a collection of unions. False for any other form.
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

// unsupportedOneOfPlacement explains a union the emit path cannot place yet,
// naming the Terraform form it arrived in rather than its schema kind.
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
// a.Lifecycle: it resolves the SDK-call bindings, then flattens the singular
// JSON:API envelope ({data:{id,type,attributes}}), hoisting data.attributes.*
// to top-level computed attributes. Fail-slow — every binding or envelope
// problem collects into one *UnsupportedEmitError and the view is discarded.
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
	hasRead, searchable := read != nil, search != nil
	byID := hasRead && (!read.BindingResolved || callHasArgument(read, "id"))

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

	// Partition the schema: required/optional leaves are SDK inputs, while the
	// lone computed envelope block is the record to flatten.
	var topLevel, inputLeaves, filterLeaves []*model.Attribute
	if a.Schema != nil {
		for _, attr := range a.Schema.Attributes {
			if (attr.Required || attr.Optional) && isLeafType(attr.TfType) {
				inputLeaves = append(inputLeaves, attr)
				if attr.Optional {
					filterLeaves = append(filterLeaves, attr)
				}
			} else {
				topLevel = append(topLevel, attr)
			}
		}
	}

	goName := dsGoName(a.Name)
	rootStruct := b.namer.qualify("DataSourceModel")
	env := b.flattenEnvelope(topLevel, idStrategy, rootExpr)

	if len(b.unsupported) > 0 {
		return DataSourceView{}, &UnsupportedEmitError{Nodes: b.unsupported}
	}

	// env is non-nil here: flattenEnvelope records an unsupported node (caught
	// above) on every failure path. The root's stem is empty, so its inline
	// children accumulate from the artifact base alone: "foo" -> <base>FooModel.
	recordAttrs, recordBlocks, recordScalars, recordLists := b.walk(rootStruct, "", b.receiver, "state", env.leaves)
	leafFields := b.models[0].Fields

	inputAttrs, inputFields := b.buildInputViews(inputLeaves)
	filterParams, filterUUID, filterStrconv := buildFilterParams(search, filterLeaves, &b.unsupported)
	readArgs, readUUID, readStrconv := buildArgumentViews(read, &b.unsupported)
	searchArgs, searchUUID, searchStrconv := buildArgumentViews(search, &b.unsupported)
	if len(b.unsupported) > 0 {
		return DataSourceView{}, &UnsupportedEmitError{Nodes: b.unsupported}
	}

	// Parent model fields: the lookup id, then the search filters, then the record
	// leaves. The group comments are only emitted for the search shapes.
	idField := env.idField
	if searchable {
		idField.Comment = "Datasource ID"
		if len(leafFields) > 0 {
			leafFields[0].Comment = "Computed values"
		}
	}
	fields := append([]ModelFieldView{idField}, inputFields...)
	b.models[0].Fields = append(fields, leafFields...)

	models, conflicts := dedupeModels(b.models)
	if len(conflicts) > 0 {
		return DataSourceView{}, &UnsupportedEmitError{Nodes: conflicts}
	}

	assignments := recordScalars
	if env.idAssign != nil {
		assignments = append([]StateAssignment{*env.idAssign}, assignments...)
	}

	var readView, searchView SDKReadView
	if hasRead {
		readView = SDKReadView{Method: read.GoMethod, ResponseType: read.GoResponseType, Arguments: readArgs}
	}
	if searchable {
		searchView = SDKReadView{
			Method:             search.GoMethod,
			Paginated:          search.Paginated,
			ItemType:           search.ItemType,
			OptionalParamsType: search.OptionalParamsType,
			Filters:            filterParams,
			Arguments:          searchArgs,
			HashInputs:         buildHashInputs(inputLeaves),
		}
	}

	return DataSourceView{
		Cardinality: Singular,
		TypeName:    a.Name,
		GoName:      goName,
		Description: a.Description,
		SDKPackage:  primary.GoPackage,
		APIStruct:   primary.GoApiStruct,
		APIAccessor: defaultAPIAccessor(primary),
		UsesUUID:    readUUID || searchUUID || filterUUID,
		UsesStrconv: readStrconv || searchStrconv || filterStrconv,
		ByID:        byID,
		Searchable:  searchable,
		Read:        readView,
		Search:      searchView,
		Models:      models,
		Schema:      SchemaView{Attributes: append(inputAttrs, recordAttrs...), Blocks: recordBlocks},
		State: StateView{
			ParamName:   paramName,
			ParamType:   paramType,
			Preamble:    env.preamble,
			Assignments: assignments,
			Lists:       recordLists,
		},
		UsesFmt: !searchable || len(searchView.HashInputs) > 0 || len(b.oneOfRenders) > 0,
		Dropped: b.dropped,
	}, nil
}

// buildInputViews turns Terraform input leaves into schema attributes and model
// fields shared by singular and plural data sources.
func (b *dataSourceBuilder) buildInputViews(leaves []*model.Attribute) (attrs []AttrView, fields []ModelFieldView) {
	comment := "Query Parameters"
	for _, leaf := range leaves {
		if leaf.Required {
			comment = "SDK call parameters"
			break
		}
	}
	for i, leaf := range leaves {
		tfName := tfNameOf(leaf.Path)
		planModifiers, planModifierType := b.planModifierViews(leaf)
		attrs = append(attrs, AttrView{
			TFName: tfName, TFType: leaf.TfType, Description: leaf.Description,
			Required: leaf.Required, Optional: leaf.Optional,
			PlanModifiers: planModifiers, PlanModifierType: planModifierType,
		})
		field := ModelFieldView{GoField: model.SdkName(tfName), GoType: leaf.GoType, TFName: tfName}
		if i == 0 {
			field.Comment = comment
		}
		fields = append(fields, field)
	}
	return attrs, fields
}

func buildHashInputs(leaves []*model.Attribute) []FilterParamView {
	inputs := make([]FilterParamView, 0, len(leaves))
	for _, leaf := range leaves {
		tfName := tfNameOf(leaf.Path)
		inputs = append(inputs, FilterParamView{
			StateField: model.SdkName(tfName),
			ValueExpr:  pointerValueExpr(leaf.GoType),
		})
	}
	return inputs
}

func callHasArgument(call *model.SDKCall, tfName string) bool {
	if call == nil {
		return false
	}
	for _, arg := range call.Arguments {
		if arg.TFName == tfName {
			return true
		}
	}
	return false
}

func buildArgumentViews(call *model.SDKCall, unsupported *[]UnsupportedNode) ([]SDKArgumentView, bool, bool) {
	if call == nil || !call.BindingResolved {
		return nil, false, false
	}
	var views []SDKArgumentView
	usesUUID, usesStrconv := false, false
	for _, arg := range call.Arguments {
		expr, parsedVar, parseCall, reason := sdkArgumentExpression(call.GoPackage, arg)
		if reason != "" {
			*unsupported = append(*unsupported, UnsupportedNode{
				Path: "sdk." + call.GoMethod + "." + arg.Name, Reason: reason,
			})
			continue
		}
		views = append(views, SDKArgumentView{
			Expression: expr, ParsedVar: parsedVar, ParseCall: parseCall,
			TFName: arg.TFName, GoType: arg.GoType,
		})
		u, s, _ := parseCallImports(parseCall)
		usesUUID, usesStrconv = usesUUID || u, usesStrconv || s
	}
	return views, usesUUID, usesStrconv
}

func buildFilterParams(call *model.SDKCall, leaves []*model.Attribute, unsupported *[]UnsupportedNode) ([]FilterParamView, bool, bool) {
	if call == nil {
		return nil, false, false
	}
	byName := map[string]model.SDKArgument{}
	for _, arg := range call.OptionalArguments {
		byName[arg.TFName] = arg
	}
	var params []FilterParamView
	usesUUID, usesStrconv := false, false
	for _, leaf := range leaves {
		tfName := tfNameOf(leaf.Path)
		if !call.BindingResolved {
			params = append(params, FilterParamView{
				StateField: model.SdkName(tfName), ParamField: model.SdkName(tfName), ValueExpr: pointerValueExpr(leaf.GoType),
			})
			continue
		}
		arg, ok := byName[tfName]
		if !ok {
			*unsupported = append(*unsupported, UnsupportedNode{
				Path:   "sdk." + call.GoMethod + "." + tfName,
				Reason: "no matching optional-parameter setter in the pinned SDK",
			})
			continue
		}
		expr, parsedVar, parseCall, reason := sdkArgumentExpression(call.GoPackage, arg)
		if reason != "" {
			*unsupported = append(*unsupported, UnsupportedNode{Path: "sdk." + call.GoMethod + "." + arg.Name, Reason: reason})
			continue
		}
		param := FilterParamView{
			StateField: model.SdkName(tfName), ParamField: model.SdkName(tfName), ValueExpr: expr, Setter: arg.Setter,
		}
		if parsedVar != "" {
			param.ParsedVar, param.ParseCall, param.TFName = parsedVar, parseCall, tfName
			u, s, _ := parseCallImports(parseCall)
			usesUUID, usesStrconv = usesUUID || u, usesStrconv || s
		}
		params = append(params, param)
	}
	return params, usesUUID, usesStrconv
}

// parseCallImports reports which extra import a ParseCall expression needs,
// switching on the package prefix of the parse function it calls.
func parseCallImports(parseCall string) (usesUUID, usesStrconv, usesTime bool) {
	switch {
	case strings.HasPrefix(parseCall, "uuid."):
		return true, false, false
	case strings.HasPrefix(parseCall, "strconv."):
		return false, true, false
	case strings.HasPrefix(parseCall, "time."):
		return false, false, true
	default:
		return false, false, false
	}
}

func sdkArgumentExpression(sdkPackage string, arg model.SDKArgument) (expr, parsedVar, parseCall, reason string) {
	if arg.Schema == nil || arg.Schema.Kind != model.SchemaKindPrimitive {
		return "", "", "", fmt.Sprintf("SDK argument type %s is outside scalar-first support", arg.GoType)
	}
	field := goFieldName(arg.TFName)
	source := "state." + field

	// The id-aliased argument's model field is always types.String, whatever the
	// path parameter's own type, so a non-string Go type has to be recovered by
	// parsing that string rather than by the schema-type switch below.
	if arg.TFName == "id" {
		return idArgumentExpression(sdkPackage, source, parsedIDVar, arg.GoType, len(arg.Schema.Enum) > 0)
	}

	var value string
	switch arg.Schema.Type {
	case "string":
		value = source + ".ValueString()"
	case "boolean":
		value = source + ".ValueBool()"
	case "integer":
		value = source + ".ValueInt64()"
	case "number":
		value = source + ".ValueFloat64()"
	default:
		return "", "", "", fmt.Sprintf("OpenAPI scalar type %q is not supported", arg.Schema.Type)
	}

	switch arg.GoType {
	case "string", "bool", "int64", "float64":
		return value, "", "", ""
	case "int", "int32", "float32":
		return arg.GoType + "(" + value + ")", "", "", ""
	case "uuid.UUID":
		name := "parsed" + model.SdkName(arg.TFName)
		return name, name, "uuid.Parse(" + value + ")", ""
	}
	// A named component type is cast directly; the attribute's own validator
	// rejected a non-member before the cast runs.
	if !model.IsBareSDKTypeName(arg.GoType) {
		return "", "", "", fmt.Sprintf("SDK argument type %s is outside scalar-first support", arg.GoType)
	}
	return sdkPackage + "." + arg.GoType + "(" + value + ")", "", "", ""
}

// idArgumentExpression binds an SDK argument aliased from the "id" attribute,
// whose model field is always types.String. A string-typed argument reads it
// directly; every other type is recovered by parsing that string.
func idArgumentExpression(sdkPackage, source, name, goType string, hasEnum bool) (expr, parsedVar, parseCall, reason string) {
	value := source + ".ValueString()"
	switch goType {
	case "string":
		return value, "", "", ""
	case "uuid.UUID":
		return name, name, "uuid.Parse(" + value + ")", ""
	case "int64":
		return name, name, "strconv.ParseInt(" + value + ", 10, 64)", ""
	case "int", "int32":
		return goType + "(" + name + ")", name, "strconv.ParseInt(" + value + ", 10, 64)", ""
	case "float64":
		return name, name, "strconv.ParseFloat(" + value + ", 64)", ""
	case "float32":
		return goType + "(" + name + ")", name, "strconv.ParseFloat(" + value + ", 64)", ""
	case "bool":
		return name, name, "strconv.ParseBool(" + value + ")", ""
	}

	// An enum-typed parameter is spelled as its component's Go type, so it misses
	// the "string" case above. The validating constructor recovers it — the id
	// carries no validator of its own — and returns a pointer, hence the deref.
	if constructor, ok := model.SDKEnumFromValueConstructor(goType); ok && hasEnum {
		return "*" + name, name, sdkPackage + "." + constructor + "(" + value + ")", ""
	}
	return "", "", "", fmt.Sprintf("id-bound SDK argument type %s is outside scalar-first support", goType)
}

// flattenedEnvelope is the result of recognizing a singular JSON:API envelope:
// the data.attributes.* leaves rewritten to top-level paths, plus the lookup id
// field/assignment and the "attributes := …" preamble.
type flattenedEnvelope struct {
	leaves   []*model.Attribute
	idField  ModelFieldView
	idAssign *StateAssignment
	preamble []string
}

// flattenEnvelope recognizes the singular JSON:API envelope at the response
// root and reshapes it for the walk: a top-level "data" object whose members
// are a subset of {id, type, attributes}. Each attributes child is hoisted to
// "response.<leaf>", "id" is surfaced (data.id strategy only), "type" dropped.
// Anything unrecognized appends to b.unsupported and the result is nil.
func (b *dataSourceBuilder) flattenEnvelope(topLevel []*model.Attribute, idStrategy model.IdStrategy, rootExpr string, reserved ...*model.Attribute) *flattenedEnvelope {
	// Siblings of "data" ("included", "meta", "links") are sideloading and request
	// metadata, not the resource's own state, so they are dropped the same way
	// data.relationships is below rather than failing a recognized shape.
	reservedNames := map[string]func(string) DroppedMember{"id": droppedIDCollision}
	for _, attr := range reserved {
		reservedNames[tfNameOf(attr.Path)] = droppedPathParameterCollision
	}

	var data *model.Attribute
	for _, attr := range topLevel {
		if tfNameOf(attr.Path) == "data" {
			data = attr
			continue
		}
		b.dropped = append(b.dropped, droppedEnvelopeMember(attr.Path))
	}
	if data == nil || data.TfType != "schema.SingleNestedAttribute" {
		b.unsupported = append(b.unsupported, UnsupportedNode{
			Path:   "response",
			Reason: "expected a JSON:API envelope with a data object {data:{...}}",
		})
		return nil
	}

	// data members must be a subset of {id, type, attributes}.
	var id, attributes *model.Attribute
	ok := true
	for _, child := range data.Children {
		switch tfNameOf(child.Path) {
		case "id":
			id = child
		case "type":
			// type is the discriminator and is not surfaced.
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
	if attributes.TfType != "schema.SingleNestedAttribute" {
		b.unsupported = append(b.unsupported, UnsupportedNode{
			Path:   attributes.Path,
			Reason: "envelope attributes must be an object",
		})
		return nil
	}

	// Hoist the attribute children to top-level paths: scalar leaves, typed list/map
	// attributes, object-list attributes, and bare nested objects are in scope.
	leaves := make([]*model.Attribute, 0, len(attributes.Children))
	for _, child := range attributes.Children {
		if isAuditField(tfNameOf(child.Path)) {
			b.dropped = append(b.dropped, droppedAuditField(child.Path))
			continue
		}
		// The root model claims "id" for Terraform identity (even when the response
		// has no data.id getter) and each path parameter the caller passed, which
		// the API merely echoes back here. One gate for every claimed name.
		if note, claimed := reservedNames[tfNameOf(child.Path)]; claimed {
			b.dropped = append(b.dropped, note(child.Path))
			continue
		}
		if !isLeafType(child.TfType) && !isArrayType(child.TfType) && !isMapType(child.TfType) && !isObjectType(child.TfType) {
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

	var idAssign *StateAssignment
	if id != nil {
		idAssign = &StateAssignment{
			Var:      "id",
			GetterOk: rootExpr + ".GetIdOk()",
			LHS:      "state.ID",
			RHS:      guardedValue(id, "id"),
		}
	}
	var preamble []string
	if len(leaves) > 0 {
		preamble = []string{"attributes := " + rootExpr + ".GetAttributes()"}
	}
	return &flattenedEnvelope{
		leaves:   leaves,
		idField:  ModelFieldView{GoField: "ID", GoType: "types.String", TFName: "id"},
		idAssign: idAssign,
		preamble: preamble,
	}
}

// dataSourceBuilder accumulates the cross-cutting outputs of one walk: model
// structs (parent before child), unsupported nodes, dropped members and import
// flags. receiver is the getter root leaves are read off, e.g. "attributes".
type dataSourceBuilder struct {
	receiver string
	// namer scopes every model struct name to this artifact and derives nested
	// names from their OpenAPI component.
	namer       modelNamer
	models      []ModelStructView
	unsupported []UnsupportedNode
	// dropped notes envelope members skipped from the attributes-only view
	// (e.g. relationships), surfaced as diagnostics rather than failures.
	dropped []DroppedMember
	// oneOfRenders caches each envelope's render, keyed on the parser's envelope
	// Name, and doubles as the "already emitted its models" marker.
	oneOfRenders map[string]oneOfRender
	// filterByResponse, when true, makes walk emit a response-mapping assignment
	// only for a node with InResponse, since an attribute absent from the response
	// type has no Get<Field>Ok to call. Its schema attribute and model field are
	// emitted either way. False assigns every attribute unconditionally.
	filterByResponse bool
	// planModifierPkgs collects the distinct planmodifier subpackages
	// (e.g. "stringplanmodifier") any attribute's PlanModifiers reference.
	planModifierPkgs map[string]struct{}
	// usesStringValidators records that some scalar attribute rendered a string
	// validator, for the template's import block.
	usesStringValidators bool
	// usesObjectValidators records oneOf selection validators, kept separate from
	// the string ones so only the referenced constructor packages are imported.
	usesObjectValidators bool
	// writeOnlySecrets is populated by the resource schema walk in stable tree
	// order. Data-source walks never append because resource-only model metadata
	// is the sole selector.
	writeOnlySecrets []WriteOnlySecretView
	writeOnlySeen    map[string]struct{}
}

// responds reports whether walk should build a's response-mapping assignment.
func (b *dataSourceBuilder) responds(a *model.Attribute) bool {
	return !b.filterByResponse || a.InResponse
}

// planModifierViews renders a's plan modifiers for its AttrView and records
// their subpackages on the builder ("stringplanmodifier.UseStateForUnknown"
// records "stringplanmodifier"). sliceType is the planmodifier.<T> element type
// matching a.GoType ("types.String" -> "String"). Zero values if a has none.
func (b *dataSourceBuilder) planModifierViews(a *model.Attribute) (names []string, sliceType string) {
	if len(a.PlanModifiers) == 0 {
		return nil, ""
	}
	names = make([]string, len(a.PlanModifiers))
	for i, pm := range a.PlanModifiers {
		names[i] = pm.Name
	}
	if b.planModifierPkgs == nil {
		b.planModifierPkgs = make(map[string]struct{})
	}
	b.planModifierPkgs[model.PlanModifierPackage(a.GoType)] = struct{}{}
	return names, strings.TrimPrefix(a.GoType, "types.")
}

// validatorViews renders a's validators for its AttrView, e.g.
// `stringvalidator.OneOf("low", "high")`. Empty unless a is configurable
// (Required or Optional), since a Computed-only validator would never run.
func (b *dataSourceBuilder) validatorViews(a *model.Attribute) (names []string, elemType string) {
	if len(a.Validators) == 0 || !(a.Required || a.Optional) {
		return nil, ""
	}
	names = make([]string, len(a.Validators))
	for i, v := range a.Validators {
		names[i] = v.Name + "(" + strings.Join(v.Args, ", ") + ")"
	}
	b.usesStringValidators = true
	return names, "String"
}

// oneOfVariantValidators returns an ExactlyOneOf validator for a configurable
// variant. Only sibling paths are passed: the framework validator already
// includes the attribute it is attached to. It defers while any candidate is
// unknown, so planning is not blocked, and otherwise rejects zero or several.
func (b *dataSourceBuilder) oneOfVariantValidators(env *model.OneOfEnvelope, current model.OneOfEnvelopeVariant) (names []string, elemType string) {
	if !current.Attribute.Optional {
		return nil, ""
	}

	expressions := make([]string, 0, len(env.Variants)-1)
	for _, candidate := range env.Variants {
		if candidate.TFName == current.TFName {
			continue
		}
		expressions = append(expressions,
			fmt.Sprintf("path.MatchRelative().AtParent().AtName(%q)", candidate.TFName))
	}

	b.usesObjectValidators = true
	return []string{"objectvalidator.ExactlyOneOf(" + strings.Join(expressions, ", ") + ")"}, "Object"
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

// droppedIDCollision is the warning-diagnostic note for a member promoted out
// of attributes whose name the envelope id already claims: flattening would
// emit two attributes named "id", and the identity-carrying envelope id wins.
func droppedIDCollision(path string) DroppedMember {
	return DroppedMember{
		Message:  fmt.Sprintf("dropped %q: collides with the envelope id surfaced as \"id\"", path),
		Severity: model.SeverityWarning,
	}
}

// droppedPathParameterCollision is the info-diagnostic note for a member
// promoted out of attributes whose name a path parameter already claims.
// Surfacing both would declare one tfsdk tag twice and, where the echo is
// readOnly, ask for an attribute that is at once Required and Computed.
func droppedPathParameterCollision(path string) DroppedMember {
	return DroppedMember{
		Message: fmt.Sprintf(
			"dropped %q: the same name is a path parameter, which the practitioner supplies and the API echoes back here",
			path),
		Severity: model.SeverityInfo,
	}
}

// walk processes one struct's worth of attributes in tree order, reserving the
// struct's slot in b.models up front so a parent precedes its children, and
// recursing into nested attributes. receiver is the SDK getter root, lhsPrefix
// the model target assignments write into, and stem the naming stem inline
// children accumulate from — unqualified and unsuffixed, unlike structName. One
// component reached from two properties is walked twice, since each site needs
// its own receiver and LHS; dedupeModels collapses the duplicates.
func (b *dataSourceBuilder) walk(structName, stem, receiver, lhsPrefix string, attrs []*model.Attribute) (attrViews, blockViews []AttrView, scalars []StateAssignment, lists []ListAssignment) {
	idx := len(b.models)
	b.models = append(b.models, ModelStructView{Name: structName})
	var fields []ModelFieldView

	for _, a := range attrs {
		tfName := tfNameOf(a.Path)
		field := model.SdkName(tfName)
		if a.WriteOnlySecret {
			secret := buildWriteOnlySecretView(a, b.namer.base)
			attrViews = append(attrViews, AttrView{TFName: tfName, WriteOnlySecret: &secret})
			fields = append(fields,
				ModelFieldView{GoField: model.SdkName(secret.WriteOnlyAttr), GoType: "types.String", TFName: secret.WriteOnlyAttr},
				ModelFieldView{GoField: model.SdkName(secret.TriggerAttr), GoType: "types.String", TFName: secret.TriggerAttr},
			)
			b.collectWriteOnlySecret(secret)
			continue
		}
		// A function of a alone, so computed before the oneOf branch, which
		// returns without reaching the switch.
		pmNames, pmType := b.planModifierViews(a)

		// A union is keyed on OneOf, not TfType: an envelope wears the same
		// schema.SingleNestedAttribute as a nested object, and walking it as one
		// would emit Get<Variant>Ok getters the SDK oneOf wrapper lacks.
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
				TFName:              tfName,
				Description:         a.Description,
				Required:            a.Required,
				Optional:            a.Optional,
				Computed:            a.Computed,
				Sensitive:           a.Sensitive,
				IsBlock:             true,
				ListBlock:           collection,
				Blocks:              render.blocks,
				HasWriteOnlySecrets: false,
				PlanModifiers:       pmNames,
				PlanModifierType:    pmType,
			})
			if b.responds(a) {
				lists = append(lists, oneOfListAssignment(
					a.OneOf, render, tfName, receiver, lhsPrefix+"."+field, collection, b.filterByResponse))
			}
			continue
		}

		switch a.TfType {
		case "schema.StringAttribute", "schema.Int64Attribute",
			"schema.Float64Attribute", "schema.BoolAttribute":
			leafValidators, leafValidatorType := b.validatorViews(a)
			attrViews = append(attrViews, AttrView{
				TFName:           tfName,
				TFType:           a.TfType,
				Description:      a.Description,
				Required:         a.Required,
				Optional:         a.Optional,
				Computed:         a.Computed,
				Sensitive:        a.Sensitive,
				Validators:       leafValidators,
				ValidatorType:    leafValidatorType,
				PlanModifiers:    pmNames,
				PlanModifierType: pmType,
			})
			fields = append(fields, ModelFieldView{GoField: field, GoType: a.GoType, TFName: tfName})
			if b.responds(a) {
				varName := leafVar(tfName)
				scalars = append(scalars, StateAssignment{
					Var:      varName,
					GetterOk: getterOk(receiver, tfName),
					LHS:      lhsPrefix + "." + field,
					RHS:      guardedValue(a, varName),
				})
			}

		case "schema.ListAttribute":
			attrViews = append(attrViews, AttrView{
				TFName:           tfName,
				TFType:           a.TfType,
				ElementType:      a.ElementType,
				Description:      a.Description,
				Required:         a.Required,
				Optional:         a.Optional,
				Computed:         a.Computed,
				Sensitive:        a.Sensitive,
				PlanModifiers:    pmNames,
				PlanModifierType: pmType,
			})
			fields = append(fields, ModelFieldView{GoField: field, GoType: a.GoType, TFName: tfName}) // types.List
			if b.responds(a) {
				lists = append(lists, ListAssignment{
					Kind:          "primitive",
					ContainerKind: "list",
					LHS:           lhsPrefix + "." + field,
					GetterOk:      getterOk(receiver, tfName),
					Var:           leafVar(tfName),
					ElementType:   a.ElementType,
				})
			}

		case "schema.MapAttribute":
			attrViews = append(attrViews, AttrView{
				TFName:           tfName,
				TFType:           a.TfType,
				ElementType:      a.ElementType,
				Description:      a.Description,
				Required:         a.Required,
				Optional:         a.Optional,
				Computed:         a.Computed,
				Sensitive:        a.Sensitive,
				PlanModifiers:    pmNames,
				PlanModifierType: pmType,
			})
			fields = append(fields, ModelFieldView{GoField: field, GoType: a.GoType, TFName: tfName})
			if b.responds(a) {
				lists = append(lists, ListAssignment{
					Kind:          "primitive",
					ContainerKind: "map",
					LHS:           lhsPrefix + "." + field,
					GetterOk:      getterOk(receiver, tfName),
					Var:           leafVar(tfName),
					ElementType:   a.ElementType,
				})
			}

		case "schema.ListNestedAttribute":
			elemStruct, childStem := b.namer.nested(stem, a)
			base := lowerFirst(field)
			loopVar, elemVar := base+"Item", base+"Model"
			fields = append(fields, ModelFieldView{GoField: field, GoType: "[]*" + elemStruct, TFName: tfName})
			childAttrs, childBlocks, childScalars, childLists := b.walk(elemStruct, childStem, loopVar, elemVar, a.Children)
			blockViews = append(blockViews, AttrView{
				TFName:              tfName,
				Description:         a.Description,
				Required:            a.Required,
				Optional:            a.Optional,
				Computed:            a.Computed,
				Sensitive:           a.Sensitive,
				IsBlock:             true,
				ListBlock:           true,
				Attributes:          childAttrs,
				Blocks:              childBlocks,
				HasWriteOnlySecrets: hasWriteOnlySecrets(childAttrs),
				PlanModifiers:       pmNames,
				PlanModifierType:    pmType,
			})
			if b.responds(a) {
				lists = append(lists, ListAssignment{
					Kind:             "object",
					LHS:              lhsPrefix + "." + field,
					GetterOk:         getterOk(receiver, tfName),
					Var:              leafVar(tfName),
					LoopVar:          loopVar,
					LoopIndex:        base + "Index",
					ExistingVar:      base + "Existing",
					ElemVar:          elemVar,
					ElemStruct:       elemStruct,
					Scalars:          childScalars,
					Lists:            childLists,
					PreserveExisting: b.filterByResponse,
				})
			}

		case "schema.SingleNestedAttribute":
			childStruct, childStem := b.namer.nested(stem, a)
			objVar := leafVar(tfName)
			elemVar := lowerFirst(field) + "Model"
			fields = append(fields, ModelFieldView{GoField: field, GoType: "*" + childStruct, TFName: tfName})
			childAttrs, childBlocks, childScalars, childLists := b.walk(childStruct, childStem, objVar, elemVar, a.Children)
			blockViews = append(blockViews, AttrView{
				TFName:              tfName,
				Description:         a.Description,
				Required:            a.Required,
				Optional:            a.Optional,
				Computed:            a.Computed,
				Sensitive:           a.Sensitive,
				IsBlock:             true,
				ListBlock:           false,
				Attributes:          childAttrs,
				Blocks:              childBlocks,
				HasWriteOnlySecrets: hasWriteOnlySecrets(childAttrs),
				PlanModifiers:       pmNames,
				PlanModifierType:    pmType,
			})
			if b.responds(a) {
				lists = append(lists, ListAssignment{
					Kind:             "object_single",
					LHS:              lhsPrefix + "." + field,
					GetterOk:         getterOk(receiver, tfName),
					Var:              objVar,
					ElemVar:          elemVar,
					ElemStruct:       childStruct,
					Scalars:          childScalars,
					Lists:            childLists,
					PreserveExisting: b.filterByResponse,
				})
			}

		default:
			b.unsupported = append(b.unsupported, UnsupportedNode{Path: a.Path, Reason: unsupportedReason(a.TfType)})
		}
	}

	b.models[idx].Fields = fields
	return attrViews, blockViews, scalars, lists
}

func writeOnlyContainment(inherited, tfType string) string {
	if inherited != "" {
		return inherited
	}
	switch tfType {
	case "schema.ListAttribute", "schema.ListNestedAttribute", "schema.ListNestedBlock":
		return "list"
	case "schema.MapAttribute", "schema.MapNestedAttribute":
		return "map"
	case "schema.SetAttribute", "schema.SetNestedAttribute", "schema.SetNestedBlock":
		return "set"
	default:
		return ""
	}
}

func unsupportedWriteOnlyContainment(path, category string) UnsupportedNode {
	return UnsupportedNode{
		Path:   path,
		Reason: fmt.Sprintf("write-only secret nested beneath a %s is not supported", category),
	}
}

func unsupportedWriteOnlyCollision(path, name string) UnsupportedNode {
	return UnsupportedNode{
		Path:   path,
		Reason: fmt.Sprintf("write-only companion name collision with %q", name),
	}
}

// validateWriteOnlySecrets checks the complete, unflattened resource tree so
// diagnostics retain their canonical OpenAPI-derived paths. Expansion later in
// walk can stay a single-purpose transformation instead of duplicating these
// support-boundary checks after paths have been rewritten for rendering.
func validateWriteOnlySecrets(attributes []*model.Attribute) []UnsupportedNode {
	var unsupported []UnsupportedNode
	var walk func([]*model.Attribute, string)
	walk = func(nodes []*model.Attribute, containment string) {
		siblingNames := make(map[string]struct{}, len(nodes))
		for _, sibling := range nodes {
			siblingNames[tfNameOf(sibling.Path)] = struct{}{}
		}

		for _, attribute := range nodes {
			if attribute.WriteOnlySecret {
				switch {
				case containment != "":
					unsupported = append(unsupported, unsupportedWriteOnlyContainment(attribute.Path, containment))
				case attribute.TfType != "schema.StringAttribute":
					unsupported = append(unsupported, UnsupportedNode{
						Path:   attribute.Path,
						Reason: "write-only secret support is limited to string attributes",
					})
				default:
					name := tfNameOf(attribute.Path)
					for _, companion := range []string{name + "_wo", name + "_wo_version"} {
						if _, exists := siblingNames[companion]; exists {
							unsupported = append(unsupported, unsupportedWriteOnlyCollision(attribute.Path, companion))
							break
						}
					}
				}
			}

			childContainment := writeOnlyContainment(containment, attribute.TfType)
			walk(attribute.Children, childContainment)
			if attribute.OneOf != nil {
				for _, variant := range attribute.OneOf.Variants {
					walk(variant.Attribute.Children, childContainment)
				}
			}
		}
	}
	walk(attributes, "")
	return unsupported
}

func hasWriteOnlySecrets(attributes []AttrView) bool {
	for _, attribute := range attributes {
		if attribute.WriteOnlySecret != nil {
			return true
		}
	}
	return false
}

func buildWriteOnlySecretView(attribute *model.Attribute, artifactBase string) WriteOnlySecretView {
	tfName := tfNameOf(attribute.Path)
	sdkField := model.SdkName(attribute.OpenAPIName)
	if sdkField == "" {
		sdkField = model.SdkName(tfName)
	}
	description := strings.TrimSpace(attribute.WriteOnlyDescription)
	if description == "" {
		description = strings.TrimSpace(attribute.Description)
	}
	if description != "" {
		description += " "
	}
	writeOnlyAttr := tfName + "_wo"
	parentBlocks := writeOnlyParentBlocks(attribute.Path)
	localStem := lowerFirst(model.SdkName(strings.Join(append(append([]string(nil), parentBlocks...), tfName), "_")))
	return WriteOnlySecretView{
		OriginalAttr:         tfName,
		WriteOnlyAttr:        writeOnlyAttr,
		TriggerAttr:          tfName + "_wo_version",
		SDKField:             sdkField,
		ParentBlocks:         parentBlocks,
		RequiredOnCreate:     attribute.SecretRequiredOnCreate,
		RequiredOnUpdate:     attribute.SecretRequiredOnUpdate,
		WriteOnlyDescription: description + "This write-only value is not stored in Terraform state.",
		TriggerDescription:   "Version trigger for " + writeOnlyAttr + " rotation.",
		ConfigVar:            artifactBase + upperFirst(localStem) + "WriteOnlySecretConfig",
		HandlerVar:           localStem + "WriteOnlySecretHandler",
		ResultVar:            localStem + "SecretResult",
	}
}

func (b *dataSourceBuilder) collectWriteOnlySecret(secret WriteOnlySecretView) {
	if b.writeOnlySeen == nil {
		b.writeOnlySeen = make(map[string]struct{})
	}
	if _, exists := b.writeOnlySeen[secret.ConfigVar]; exists {
		return
	}
	b.writeOnlySeen[secret.ConfigVar] = struct{}{}
	b.writeOnlySecrets = append(b.writeOnlySecrets, secret)
}

func writeOnlyParentBlocks(attributePath string) []string {
	const attributesMarker = ".data.attributes."
	if index := strings.Index(attributePath, attributesMarker); index >= 0 {
		attributePath = attributePath[index+len(attributesMarker):]
	} else {
		attributePath = strings.TrimPrefix(attributePath, "resource.")
		attributePath = strings.TrimPrefix(attributePath, "response.")
	}
	segments := strings.Split(attributePath, ".")
	if len(segments) <= 1 {
		return nil
	}
	parents := make([]string, 0, len(segments)-1)
	for _, segment := range segments[:len(segments)-1] {
		parents = append(parents, stripMarkers(segment))
	}
	return parents
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

// isArrayType reports whether tfType is a collection-of-primitive
// (ListAttribute) or an array-of-object (ListNestedAttribute).
func isArrayType(tfType string) bool {
	switch tfType {
	case "schema.ListAttribute", "schema.ListNestedAttribute":
		return true
	default:
		return false
	}
}

// isMapType reports whether tfType is a typed dynamic-key map attribute.
func isMapType(tfType string) bool { return tfType == "schema.MapAttribute" }

// isObjectType reports whether tfType is a bare nested object the envelope
// hoists into single-object machinery (schema.SingleNestedAttribute).
func isObjectType(tfType string) bool { return tfType == "schema.SingleNestedAttribute" }

// auditFields are server-managed timestamps and actor handles, dropped from the
// top level of a generated data source as schema noise.
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
// value to: the attribute's lowerCamel name, suffixed with "Var" for Go keywords
// or "Value" when it would shadow an identifier in updateState's scope (state,
// attributes, the receiver).
func leafVar(tfName string) string {
	v := lowerFirst(model.SdkName(tfName))
	if token.Lookup(v).IsKeyword() {
		return v + "Var"
	}
	switch v {
	case "state", "attributes", "ok", "d", "data", "resp", "items", "ctx":
		return v + "Value"
	}
	// A property named "type" would otherwise emit `if type, ok := ...`, which is
	// not Go; the SDK generator's own rule suffixes "Var". The shadow suffix above
	// stays "Value" because it answers a different question: colliding with a
	// local the template declares rather than with the language.
	return model.EscapeReservedKeyword(v)
}

// guardedValue wraps a guarded assignment's local — bound from an Ok-getter,
// so a pointer — in the types.*Value constructor matching the model field's
// GoType. A date-time or UUID renders via .String(), an enum is dereferenced
// and cast back to string, integers are cast to int64 as the framework wants.
func guardedValue(a *model.Attribute, varName string) string {
	switch a.GoType {
	case "types.String":
		switch {
		case a.Format == "date-time" || a.Format == "uuid":
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
// the model field's GoType, casting integers to int64. Strings whose getter
// does not return a bare string are reconciled first: date-time and UUID
// getters return stringers (.String()), enum getters named types (string(…)).
func wrapValue(a *model.Attribute, chain string) string {
	switch a.GoType {
	case "types.String":
		switch {
		case a.Format == "date-time" || a.Format == "uuid":
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
	case "schema.MapNestedAttribute":
		return "map-of-object not yet supported"
	default:
		return fmt.Sprintf("attribute type %q not yet supported", tfType)
	}
}

// buildPluralView derives the plural DataSourceView: scalar query params become
// Optional filters, and the results-array element (a JSON:API envelope) is
// flattened into one item struct projected per element — "id" read off the loop
// variable, "attributes.*" off item.Attributes, "type" dropped. Fail-slow: bad
// filter or item nodes collect into one *UnsupportedEmitError, discarding view.
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

	// Partition the top-level schema: required/optional leaves are SDK inputs, the lone
	// ListNestedAttribute is the items container (the model already dropped response
	// metadata siblings, keeping only the results array).
	var inputLeaves, filterLeaves []*model.Attribute
	var itemsBlock *model.Attribute
	if a.Schema != nil {
		for _, attr := range a.Schema.Attributes {
			switch {
			case attr.TfType == "schema.ListNestedAttribute":
				itemsBlock = attr
			case (attr.Required || attr.Optional) && isLeafType(attr.TfType):
				inputLeaves = append(inputLeaves, attr)
				if attr.Optional {
					filterLeaves = append(filterLeaves, attr)
				}
			default:
				unsupported = append(unsupported, UnsupportedNode{Path: attr.Path, Reason: unsupportedReason(attr.TfType)})
			}
		}
	}
	if itemsBlock == nil {
		unsupported = append(unsupported, UnsupportedNode{Path: "response", Reason: "missing results array block"})
	}

	goName := dsGoName(a.Name)

	// b hosts walk, so list-of-object item fields generate their element structs,
	// and buildInputViews, so an input's plan-modifier package is registered for
	// the import block (a plural data source's filters carry none).
	b := &dataSourceBuilder{namer: modelNamer{base: goName}}

	inputAttrs, inputFields := b.buildInputViews(inputLeaves)
	filterParams, filterUUID, filterStrconv := buildFilterParams(call, filterLeaves, &unsupported)
	callArgs, usesUUID, usesStrconv := buildArgumentViews(call, &unsupported)
	scalarLeaves, nonScalars := flattenItemElement(itemsBlock, &unsupported, &dropped)

	// The item element's own stem: call.ItemType is the response element's
	// component name, so the item struct is named after it and the item's inline
	// children accumulate from it. Taken verbatim — ItemType is already a Go type
	// name, so SdkName would only mangle its acronyms.
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
	// ItemLists, read off item.Attributes: primitive lists/maps as leaf
	// attributes, list-of-object and bare object as nested ones, element walked.
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
				n.OneOf, render, tfName, "item.Attributes", "r."+field, collection, false))
			continue
		}

		switch n.TfType {
		case "schema.ListAttribute", "schema.MapAttribute":
			containerKind := "list"
			if n.TfType == "schema.MapAttribute" {
				containerKind = "map"
			}
			itemAttrs = append(itemAttrs, AttrView{
				TFName: tfName, TFType: n.TfType, ElementType: n.ElementType, Description: n.Description, Computed: true,
			})
			itemFields = append(itemFields, ModelFieldView{GoField: field, GoType: n.GoType, TFName: tfName})
			itemLists = append(itemLists, ListAssignment{
				Kind: "primitive", ContainerKind: containerKind, LHS: "r." + field, GetterOk: getter, Var: leafVar(tfName), ElementType: n.ElementType,
			})
		case "schema.ListNestedAttribute":
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
		case "schema.SingleNestedAttribute":
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
			// Fail rather than omit a representable attribute with no field, block
			// or diagnostic.
			unsupported = append(unsupported, UnsupportedNode{Path: n.Path, Reason: unsupportedReason(n.TfType)})
		}
	}

	unsupported = append(unsupported, b.unsupported...)
	if len(unsupported) > 0 {
		return DataSourceView{}, &UnsupportedEmitError{Nodes: unsupported}
	}

	itemField := model.SdkName(a.Name)

	parentFields := append([]ModelFieldView{}, inputFields...)
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
		APIAccessor: defaultAPIAccessor(call),
		UsesUUID:    usesUUID || filterUUID,
		UsesStrconv: usesStrconv || filterStrconv,
		Read: SDKReadView{
			Method:             call.GoMethod,
			Paginated:          call.Paginated,
			ItemType:           call.ItemType,
			OptionalParamsType: call.OptionalParamsType,
			Filters:            filterParams,
			HashInputs:         buildHashInputs(inputLeaves),
			Arguments:          callArgs,
		},
		Models: models,
		Schema: SchemaView{
			Attributes: inputAttrs,
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
		UsesFmt: len(callArgs) > 0 || len(filterParams) > 0 || len(b.oneOfRenders) > 0,
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
			if child.TfType != "schema.SingleNestedAttribute" {
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
				case isArrayType(leaf.TfType), isMapType(leaf.TfType), isObjectType(leaf.TfType):
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

// goFieldName is the exported Go struct field name for a TF attribute:
// model.SdkName, except "id" becomes "ID" for Go's initialism convention. The
// SDK getter is unaffected and still reads GetId().
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
// "datadogTeam". The Datadog prefix is unconditional, so <dsGoName>DataSource,
// <dsGoName>DataSourceModel and New<title dsGoName>DataSource all line up.
func dsGoName(name string) string {
	return lowerFirst("Datadog" + model.SdkName(name))
}

// checkResponseTypeMatches fails a resource whose role response type disagrees
// with Read's: Create, Read and Update share one generated updateState method,
// which cannot map two different SDK response types.
func checkResponseTypeMatches(artifact, role, roleType, readType string) error {
	if roleType == readType {
		return nil
	}
	return fmt.Errorf(
		"emit: resource %q: %s response type %q differs from Read's %q — a single updateState cannot map both",
		artifact, role, roleType, readType)
}

// BuildResourceView assembles the render-ready view for a resource artifact.
// a.Schema must already hold the merged request/response tree; a nil Schema
// fails outright rather than rendering an empty schema and a CRUD that
// populates nothing. A request body is built one JSON:API level at a time, each
// from its own New<Type>WithDefaults(), then populated with Set<Field>(...)
// calls against the constructed local. Anything the request path cannot map
// fails the artifact rather than generating a body that never sends it.
func BuildResourceView(a *model.Artifact) (ResourceView, error) {
	if a.Schema == nil {
		return ResourceView{}, fmt.Errorf(
			"emit: resource %q has no schema — buildResourceArtifact must run the request/response "+
				"merge (MergeResourceSchema, BuildResourceTree) before emit", a.Name)
	}
	lc := a.Lifecycle
	if lc == nil || lc.Create == nil || lc.Read == nil || lc.Delete == nil {
		return ResourceView{}, fmt.Errorf("emit: resource %q is missing a required lifecycle binding (Create/Read/Delete)", a.Name)
	}
	if err := checkResponseTypeMatches(a.Name, "Create", lc.Create.GoResponseType, lc.Read.GoResponseType); err != nil {
		return ResourceView{}, err
	}
	if lc.Update != nil {
		if err := checkResponseTypeMatches(a.Name, "Update", lc.Update.GoResponseType, lc.Read.GoResponseType); err != nil {
			return ResourceView{}, err
		}
	}
	if unsupported := validateWriteOnlySecrets(a.Schema.Attributes); len(unsupported) > 0 {
		return ResourceView{}, &UnsupportedEmitError{Nodes: unsupported}
	}

	primary := lc.Read
	goName := dsGoName(a.Name)
	b := &dataSourceBuilder{receiver: "attributes", namer: modelNamer{base: goName}}

	// Split the path parameters back out: they are not part of the JSON:API
	// envelope, so flattenEnvelope would drop them as stray siblings. Keyed on
	// the attribute's origin, not on Required, which a body leaf also sets.
	var topLevel, pathInputs []*model.Attribute
	for _, attr := range a.Schema.Attributes {
		if attr.FromPathParameter {
			pathInputs = append(pathInputs, attr)
			continue
		}
		topLevel = append(topLevel, attr)
	}

	env := b.flattenEnvelope(topLevel, lc.IdStrategy, "resp.Data", pathInputs...)
	if len(b.unsupported) > 0 {
		return ResourceView{}, &UnsupportedEmitError{Nodes: b.unsupported}
	}

	rootStruct := b.namer.qualify("ResourceModel")
	b.filterByResponse = true
	recordAttrs, recordBlocks, recordScalars, recordLists := b.walk(rootStruct, "", b.receiver, "state", env.leaves)

	// Each body-sending role gets its own field list, walked against its own
	// data.attributes schema. A shared list cannot work: the merged tree's
	// request side is the union of both bodies, and the SDK generates Set<Field>
	// only on the request type declaring it, so a create-only field set on an
	// update body would not compile.
	createFields, requestImps := buildRequestFields(
		env.leaves, lc.Create.RequestAttributesSchema, "state", requestAttributesVar, primary.GoPackage, &b.unsupported)
	createWriteOnlySecrets := writeOnlySecretsForFields(
		b.writeOnlySecrets, createFields, "GetSecretForCreate(ctx, &request.Config)")
	createArgs, createUUID, createStrconv := buildArgumentViews(lc.Create, &b.unsupported)
	readArgs, readUUID, readStrconv := buildArgumentViews(lc.Read, &b.unsupported)
	deleteArgs, deleteUUID, deleteStrconv := buildArgumentViews(lc.Delete, &b.unsupported)
	// Resolved before the gate below, so a level with no SDK component or an
	// untypable body id fails rather than rendering an empty or ill-typed body.
	createEnvelope := buildRequestEnvelope(a.Name, "Create", primary.GoPackage, lc.Create, createFields, &b.unsupported)

	updateView := CRUDCallView{}
	var updateUUID, updateStrconv bool
	if lc.Update != nil {
		updateArgs, uuidArg, strconvArg := buildArgumentViews(lc.Update, &b.unsupported)
		updateUUID, updateStrconv = uuidArg, strconvArg
		// Both walks cover the same merged tree, so a node neither body can map
		// is found twice; dedupe so the count counts problems, not walks.
		var updateUnsupported []UnsupportedNode
		updateFields, updateImps := buildRequestFields(
			env.leaves, lc.Update.RequestAttributesSchema, "state", requestAttributesVar, primary.GoPackage, &updateUnsupported)
		for _, n := range updateUnsupported {
			if !slices.Contains(b.unsupported, n) {
				b.unsupported = append(b.unsupported, n)
			}
		}
		requestImps = requestImps.or(updateImps)
		updateView = CRUDCallView{
			Method: lc.Update.GoMethod, GoRequestType: lc.Update.GoRequestType,
			GoResponseType: lc.Update.GoResponseType, Arguments: updateArgs,
			Envelope: buildRequestEnvelope(a.Name, "Update", primary.GoPackage, lc.Update, updateFields, &b.unsupported),
			WriteOnlySecrets: writeOnlySecretsForFields(
				b.writeOnlySecrets, updateFields, "GetSecretForUpdate(ctx, &request.Config, &request)"),
		}
		// Keyed on whether the body itself declares an id, and of what type —
		// not on env.idAssign, which describes only the response side.
		if envelope := updateView.Envelope; envelope != nil {
			if reason := setUpdateBodyID(envelope, lc.Update, updateArgs); reason != "" {
				b.unsupported = append(b.unsupported, UnsupportedNode{
					Path:   "sdk." + lc.Update.GoMethod + ".data.id",
					Reason: reason,
				})
			}
			if envelope.IDPrep != nil {
				u, str, _ := parseCallImports(envelope.IDPrep.ParseCall)
				updateUUID, updateStrconv = updateUUID || u, updateStrconv || str
			}
		}
	}

	if len(b.unsupported) > 0 {
		return ResourceView{}, &UnsupportedEmitError{Nodes: b.unsupported}
	}

	// Path parameters sit beside id in the schema and the model but never in the
	// request or response mapping: the body has neither a setter nor a getter
	// for them, so they are not part of env.leaves.
	pathAttrs, pathFields := b.buildInputViews(pathInputs)
	recordAttrs = append(pathAttrs, recordAttrs...)
	b.models[0].Fields = append(
		append([]ModelFieldView{env.idField}, pathFields...),
		b.models[0].Fields...)

	models, conflicts := dedupeModels(b.models)
	if len(conflicts) > 0 {
		return ResourceView{}, &UnsupportedEmitError{Nodes: conflicts}
	}

	assignments := recordScalars
	if env.idAssign != nil {
		assignments = append([]StateAssignment{*env.idAssign}, assignments...)
	}

	planModifierPkgs := sortedKeys(b.planModifierPkgs)

	return ResourceView{
		TypeName:    a.Name,
		GoName:      goName,
		Description: a.Description,
		SDKPackage:  primary.GoPackage,
		APIStruct:   primary.GoApiStruct,
		APIAccessor: defaultAPIAccessor(primary),
		Create: CRUDCallView{
			Method: lc.Create.GoMethod, GoRequestType: lc.Create.GoRequestType,
			GoResponseType: lc.Create.GoResponseType, Arguments: createArgs,
			Envelope: createEnvelope, WriteOnlySecrets: createWriteOnlySecrets,
			NormalizeUnknowns: buildUnknownNormalizations(env.leaves, "state"),
		},
		Read: CRUDCallView{
			Method: lc.Read.GoMethod, GoResponseType: lc.Read.GoResponseType, Arguments: readArgs,
		},
		Update:            updateView,
		UpdateUnsupported: lc.UpdateUnsupported,
		Delete: CRUDCallView{
			Method: lc.Delete.GoMethod, Arguments: deleteArgs,
		},
		Models: models,
		Schema: SchemaView{
			Attributes:          recordAttrs,
			Blocks:              recordBlocks,
			HasWriteOnlySecrets: hasWriteOnlySecrets(recordAttrs),
			IncludeResourceID:   true,
		},
		State: StateView{
			ParamName:   "resp",
			ParamType:   "*" + primary.GoPackage + "." + primary.GoResponseType,
			Preamble:    env.preamble,
			Assignments: assignments,
			Lists:       recordLists,
		},
		// fmt comes from a oneOf's ambiguous-match diagnostic on the response
		// side and its exactly-one selection check on the request side. A
		// request-only union never reaches oneOfRenders, since the state walk
		// filters by response, hence the second term.
		UsesFmt:              len(b.oneOfRenders) > 0 || requestImps.fmt,
		UsesValidators:       b.usesStringValidators || b.usesObjectValidators,
		UsesStringValidators: b.usesStringValidators,
		UsesObjectValidators: b.usesObjectValidators,
		UsesPlanModifiers:    len(planModifierPkgs) > 0,
		PlanModifierPackages: planModifierPkgs,
		Import:               buildImportView(pathAttrs),
		UsesUUID:             createUUID || readUUID || updateUUID || deleteUUID || requestImps.uuid,
		UsesStrconv:          createStrconv || readStrconv || updateStrconv || deleteStrconv,
		UsesTime:             requestImps.time,
		WriteOnlySecrets:     b.writeOnlySecrets,
		UsesWriteOnly:        len(b.writeOnlySecrets) > 0,
		Dropped:              b.dropped,
	}, nil
}

// setUpdateBodyID fills in how a JSON:API update body sets its data.id: the
// expression, plus the argument view declaring a parsed local when the value
// must be recovered from the string Terraform holds. The type comes from the
// body's own data.id, never from the path parameter naming the same record
// (rum_replay_playlist pairs an int64 path id with a string data.id); the path
// argument's local is reused only when the types agree, found by name, since a
// sub-resource's path names a parent first and index 0 is not the record id.
func setUpdateBodyID(envelope *RequestEnvelopeView, call *model.SDKCall, args []SDKArgumentView) (reason string) {
	// A body declaring no id has nothing to set, and no generated setter.
	if !call.RequestDeclaresID {
		return ""
	}
	goType := call.RequestIDGoType
	if goType == "" {
		return "the update body's data.id carries a format the SDK generator itself cannot type, so there is no expression to set it from"
	}

	// Reuse the path argument's local when the type matches, so the string is
	// parsed once. By name, never by position.
	for _, arg := range args {
		if arg.TFName == "id" && arg.GoType == goType {
			envelope.IDExpr = arg.Expression
			return ""
		}
	}

	// Otherwise the body parses under a name of its own, since a path argument
	// of another type may already hold parsedIDVar here. hasEnum is always
	// false: RequestIDGoType reads type and format only, so an enum-typed
	// data.id arrives spelled "string".
	expr, parsedVar, parseCall, unsupportedReason := idArgumentExpression(
		call.GoPackage, "state.ID", requestIDVar, goType, false)
	if unsupportedReason != "" {
		return unsupportedReason
	}
	envelope.IDExpr = expr
	if parsedVar != "" {
		envelope.IDPrep = &SDKArgumentView{ParsedVar: parsedVar, ParseCall: parseCall, TFName: "id"}
	}
	return ""
}

// buildImportView derives how ImportState recovers identity: the path
// parameters in order, then "id", plus the "<a>:<b>" format a failure message
// quotes.
func buildImportView(pathAttrs []AttrView) ImportView {
	parts := make([]string, 0, len(pathAttrs)+1)
	for _, attr := range pathAttrs {
		parts = append(parts, attr.TFName)
	}
	parts = append(parts, "id")

	quoted := make([]string, len(parts))
	for i, part := range parts {
		quoted[i] = "<" + part + ">"
	}
	return ImportView{Parts: parts, Format: strings.Join(quoted, ":")}
}

// requestDataVar and requestAttributesVar name the locals a resource's Create
// and Update bodies build their JSON:API envelope into. Fixed, not derived:
// each is local to one lifecycle method, and the "body" prefix keeps them clear
// of the locals buildRequestFields allocates from leaf names.
const (
	requestDataVar       = "bodyData"
	requestAttributesVar = "bodyAttributes"
	// requestIDVar holds the update body's data.id when it must be parsed
	// independently of the path argument, whose local is parsedIDVar.
	requestIDVar = "parsedBodyId"
	// parsedIDVar is the local idArgumentExpression declares for an id-aliased
	// path argument that has to be parsed.
	parsedIDVar = "parsedId"
)

// buildRequestEnvelope derives how one role constructs its JSON:API request
// body. Every level is built from its own New<Type>WithDefaults(): the
// wrapper's constructor returns the zero struct and never builds Data, so the
// "type" discriminator only the data component assigns would stay empty, and
// reaching through a pointer Data dereferences nil at apply time. A level with
// no SDK component therefore fails the artifact instead of falling back —
// except the attributes level, which is not constructed when fields is empty.
func buildRequestEnvelope(artifact, role, sdkPackage string, call *model.SDKCall, fields []RequestFieldView, unsupported *[]UnsupportedNode) *RequestEnvelopeView {
	if call == nil {
		return nil
	}
	if call.GoRequestDataType == "" {
		*unsupported = append(*unsupported, UnsupportedNode{
			Path: "sdk." + call.GoMethod + ".data",
			Reason: fmt.Sprintf(
				"%s request body for resource %q leaves its JSON:API \"data\" member inline, so there is no SDK component to construct it from and the \"type\" discriminator cannot be set",
				role, artifact),
		})
		return nil
	}
	envelope := &RequestEnvelopeView{
		SDKPackage: sdkPackage, Fields: fields,
		DataVar: requestDataVar, DataType: call.GoRequestDataType,
	}

	// Send the discriminator explicitly wherever the spec determines it; where it
	// does not, the SDK's own constructor must supply it, or the request would go
	// out with "type":"" and be rejected.
	switch d := call.RequestDiscriminator; {
	case d == nil:
		// No type property on the data component: not a discriminated envelope.
	case d.Determined():
		envelope.TypeExpr = fmt.Sprintf("%s.%s(%q)", sdkPackage, d.GoType, d.Values[0])
	case d.SDKDefaulted:
		// The pinned SDK assigns it in New<Data>WithDefaults().
	default:
		*unsupported = append(*unsupported, UnsupportedNode{
			Path:   "sdk." + call.GoMethod + ".data.type",
			Reason: discriminatorReason(role, artifact, d),
		})
		return nil
	}

	if len(fields) == 0 {
		return envelope
	}
	if call.GoRequestAttributesType == "" {
		*unsupported = append(*unsupported, UnsupportedNode{
			Path: "sdk." + call.GoMethod + ".data.attributes",
			Reason: fmt.Sprintf(
				"%s request body for resource %q has settable attributes but leaves data.attributes inline, so there is no SDK component to construct them on",
				role, artifact),
		})
		return nil
	}
	envelope.AttributesType = call.GoRequestAttributesType
	return envelope
}

// discriminatorReason explains why a request body cannot name its JSON:API
// type, distinguishing three causes: the spec allows several values and names
// no default, it constrains none at all, or the property is inline and so has
// no component name to convert through.
func discriminatorReason(role, artifact string, d *model.RequestDiscriminator) string {
	switch {
	case len(d.Values) > 1:
		return fmt.Sprintf(
			"%s request body for resource %q cannot set the JSON:API \"type\" discriminator: its schema allows %s and declares no default, so the spec does not say which value a request carries",
			role, artifact, strings.Join(quoteAll(d.Values), " or "))
	case len(d.Values) == 0:
		return fmt.Sprintf(
			"%s request body for resource %q cannot set the JSON:API \"type\" discriminator: its schema constrains no values and declares no default",
			role, artifact)
	default:
		return fmt.Sprintf(
			"%s request body for resource %q cannot set the JSON:API \"type\" discriminator to %q: the property is inline, so it has no SDK component name to convert the value through",
			role, artifact, d.Values[0])
	}
}

// quoteAll %q-quotes values for an error message, so a reader can tell an
// empty string or a stray space from a real value.
func quoteAll(values []string) []string {
	out := make([]string, 0, len(values))
	for _, v := range values {
		out = append(out, fmt.Sprintf("%q", v))
	}
	return out
}

// requestImports records the imports one request subtree's generated code
// needs. It is accumulated on the way out of the walk that builds the subtree,
// never re-derived from the finished view: only the branch emitting a
// time.Parse or fmt.Sprintf knows that it did.
type requestImports struct{ uuid, time, fmt bool }

func (i requestImports) or(other requestImports) requestImports {
	return requestImports{i.uuid || other.uuid, i.time || other.time, i.fmt || other.fmt}
}

// buildRequestFields derives one role's RequestFieldView list, one per
// practitioner-settable (Required or Optional) attribute of attrs that role's
// own body declares, recursing into an object's Children at every depth. role
// is that body's own data.attributes schema, which roleChild narrows attrs —
// the union of both request sides — against: an attribute the role omits is
// skipped silently, one it declares but this cannot map fails the artifact.
// stateExpr is what attrs' fields are read off, target what each Set<GoField>
// is called on, precomputed because a recursing partial cannot see ancestors.
func buildRequestFields(attrs []*model.Attribute, role *model.Schema, stateExpr, target, sdkPackage string, unsupported *[]UnsupportedNode) (fields []RequestFieldView, imports requestImports) {
	for _, a := range attrs {
		if !a.Required && !a.Optional {
			continue // Computed-only: not request-settable.
		}
		tfName := tfNameOf(a.Path)
		roleChildSchema, declared := roleChild(role, a)
		if !declared {
			continue // Present in the other request body only; this one has no setter for it.
		}
		field := model.SdkName(tfName)
		childState := stateExpr + "." + field
		if a.WriteOnlySecret {
			secret := buildWriteOnlySecretView(a, "")
			fields = append(fields, RequestFieldView{
				GoField:         secret.SDKField,
				Target:          target,
				WriteOnlyResult: secret.ResultVar,
			})
			continue
		}

		if a.OneOf != nil {
			rf, oneOfImports, ok := buildOneOfRequestField(a, roleChildSchema, tfName, field, childState, target, sdkPackage, unsupported)
			if !ok {
				continue
			}
			fields = append(fields, rf)
			imports = imports.or(oneOfImports)
			continue
		}

		switch a.TfType {
		case "schema.SingleNestedBlock", "schema.SingleNestedAttribute":
			rf, nested, ok := buildNestedRequestField(a, roleChildSchema, field, childState, target, sdkPackage, unsupported)
			if !ok {
				continue
			}
			fields = append(fields, rf)
			imports = imports.or(nested)
			continue

		case "schema.ListAttribute", "schema.MapAttribute":
			rf, reason := buildPrimitiveCollectionField(a, tfName, field, childState, target)
			if reason != "" {
				*unsupported = append(*unsupported, UnsupportedNode{Path: elementPath(a), Reason: reason})
				continue
			}
			fields = append(fields, rf)
			continue

		case "schema.ListNestedBlock", "schema.ListNestedAttribute":
			rf, element, ok := buildObjectCollectionField(a, roleChildSchema, tfName, field, childState, target, sdkPackage, unsupported)
			if !ok {
				continue
			}
			fields = append(fields, rf)
			imports = imports.or(element)
			continue

			// schema.MapNestedAttribute (a map of objects) has no case here: walk
			// over the same tree already fails the artifact with its own
			// "map-of-object not yet supported", so this could only duplicate it.
		}

		expr, parsedVar, parseCall, reason := requestValueExpr(a, roleChildSchema, childState, field, sdkPackage)
		if reason != "" {
			*unsupported = append(*unsupported, UnsupportedNode{Path: a.Path, Reason: reason})
			continue
		}
		rf := RequestFieldView{GoField: field, Target: target, ValueExpr: expr, ParsedVar: parsedVar, ParseCall: parseCall, Required: a.Required}
		if parsedVar != "" {
			rf.TFName = tfName
		}
		if !a.Required {
			rf.NullCheck = notNullOrUnknown(childState)
		}
		fields = append(fields, rf)
		u, _, t := parseCallImports(parseCall)
		imports.uuid = imports.uuid || u
		imports.time = imports.time || t
	}
	return fields, imports
}

func writeOnlySecretsForFields(all []WriteOnlySecretView, fields []RequestFieldView, getterCall string) []WriteOnlySecretView {
	used := make(map[string]struct{})
	var collect func([]RequestFieldView)
	collect = func(nodes []RequestFieldView) {
		for _, field := range nodes {
			if field.WriteOnlyResult != "" {
				used[field.WriteOnlyResult] = struct{}{}
			}
			if field.Nested != nil {
				collect(field.Nested.Fields)
			}
			if field.Collection != nil {
				collect(field.Collection.Fields)
			}
			if field.OneOf != nil {
				for _, variant := range field.OneOf.Variants {
					collect(variant.Fields)
				}
			}
		}
	}
	collect(fields)
	if len(used) == 0 {
		return nil
	}

	result := make([]WriteOnlySecretView, 0, len(used))
	for _, secret := range all {
		if _, exists := used[secret.ResultVar]; !exists {
			continue
		}
		secret.GetterCall = getterCall
		result = append(result, secret)
	}
	return result
}

func buildUnknownNormalizations(attributes []*model.Attribute, stateExpr string) []UnknownNormalizationView {
	var result []UnknownNormalizationView
	for _, attribute := range attributes {
		if attribute.WriteOnlySecret {
			continue
		}
		field := model.SdkName(tfNameOf(attribute.Path))
		expr := stateExpr + "." + field

		switch attribute.TfType {
		case "schema.SingleNestedBlock", "schema.SingleNestedAttribute":
			children := buildUnknownNormalizations(attribute.Children, expr)
			if len(children) != 0 {
				result = append(result, UnknownNormalizationView{Kind: "object", Expr: expr, Children: children})
			}
			continue
		case "schema.ListNestedBlock", "schema.ListNestedAttribute":
			itemVar := leafVar(field + "Item")
			children := buildUnknownNormalizations(attribute.Children, itemVar)
			if len(children) != 0 {
				result = append(result, UnknownNormalizationView{
					Kind: "list_object", Expr: expr, ItemVar: itemVar, Children: children,
				})
			}
			continue
		}

		if !attribute.Computed {
			continue
		}
		nullExpr := computedNullExpr(attribute)
		if nullExpr != "" {
			result = append(result, UnknownNormalizationView{Kind: "value", Expr: expr, NullExpr: nullExpr})
		}
	}
	return result
}

func computedNullExpr(attribute *model.Attribute) string {
	switch attribute.GoType {
	case "types.String":
		return "types.StringNull()"
	case "types.Bool":
		return "types.BoolNull()"
	case "types.Int64":
		return "types.Int64Null()"
	case "types.Float64":
		return "types.Float64Null()"
	case "types.List":
		if attribute.ElementType != "" {
			return "types.ListNull(" + attribute.ElementType + ")"
		}
	case "types.Map":
		if attribute.ElementType != "" {
			return "types.MapNull(" + attribute.ElementType + ")"
		}
	}
	return ""
}

// roleChild resolves one role body's schema node for attribute a at the current
// level, and reports whether that body declares it at all. The lookup is by
// a.OpenAPIName, since SnakeCase cannot be inverted from the Terraform path;
// the node returned is the level whose properties line up with a.Children, so a
// collection is reached through Items. declared is false only when the role
// demonstrably describes an object here and omits a. Every other shape — a nil
// role, a level unreadable as an object, an attribute with no OpenAPI name —
// means "unknown" and keeps the attribute rather than dropping it silently.
func roleChild(role *model.Schema, a *model.Attribute) (child *model.Schema, declared bool) {
	if role == nil || len(role.Properties) == 0 || a.OpenAPIName == "" {
		return nil, true
	}
	child, declared = role.Properties[a.OpenAPIName]
	if child != nil && child.Items != nil {
		child = child.Items
	}
	return child, declared
}

// buildNestedRequestField derives one object field's RequestFieldView: the
// request-side SDK type to build via New<SDKType>WithDefaults(), plus its child
// fields read off modelExpr — a's own model pointer, so a doubly-nested field
// resolves against its immediate parent, not the top-level state. role is this
// object's node in the calling role's request schema, which both names the
// component and narrows its children; imports are the subtree's own needs.
func buildNestedRequestField(a *model.Attribute, role *model.Schema, field, modelExpr, target, sdkPackage string, unsupported *[]UnsupportedNode) (rf RequestFieldView, imports requestImports, ok bool) {
	refName := roleRequestRefName(a, role)
	if refName == "" {
		*unsupported = append(*unsupported, UnsupportedNode{
			Path: a.Path,
			Reason: "nested object has no named request-side SDK type (its schema was never reached " +
				"through a $ref component in the Create or Update body)",
		})
		return RequestFieldView{}, requestImports{}, false
	}
	nestedVar := lowerFirst(field) + "Value"
	childFields, imports := buildRequestFields(a.Children, role, modelExpr, nestedVar, sdkPackage, unsupported)
	rf = RequestFieldView{
		GoField:  field,
		Target:   target,
		Required: a.Required,
		Nested: &RequestNestedView{
			Constructor: sdkPackage + ".New" + refName + "WithDefaults()",
			Var:         nestedVar,
			ModelExpr:   modelExpr,
			Fields:      childFields,
		},
	}
	if !a.Required {
		rf.NullCheck = modelExpr + " != nil"
	}
	return rf, imports, true
}

// buildPrimitiveCollectionField derives a primitive-terminal list or map
// field's RequestFieldView. State holds the framework's raw types.List /
// types.Map, which the SDK's Set<Field>([]<T>) / Set<Field>(map[string]<T>)
// cannot take, so it is first decoded into a native Go slice or map via
// ElementsAs — a ctx-taking, diag.Diagnostics-returning conversion. reason is
// non-empty when the element itself is unsupported, naming why.
func buildPrimitiveCollectionField(a *model.Attribute, tfName, field, childState, target string) (rf RequestFieldView, reason string) {
	goElem, reason := primitiveElementGoType(a)
	if reason != "" {
		return RequestFieldView{}, reason
	}

	convertType := "[]" + goElem
	if isMapType(a.TfType) {
		convertType = "map[string]" + goElem
	}
	convertVar := leafVar(tfName) + "Elements"
	rf = RequestFieldView{
		GoField:  field,
		Target:   target,
		Required: a.Required,
		Collection: &RequestCollectionView{
			Kind:        "primitive",
			ConvertVar:  convertVar,
			ConvertType: convertType,
			ConvertCall: childState + ".ElementsAs(ctx, &" + convertVar + ", false)",
		},
	}
	if !a.Required {
		rf.NullCheck = notNullOrUnknown(childState)
	}
	return rf, ""
}

// primitiveElementGoType maps a primitive collection's element to the native Go
// type ElementsAs decodes into, rejecting whatever the request side cannot
// recover a typed SDK value for: a format or enum on the element — ElementType
// collapses those to plain "types.StringType", hence the separate
// ElementFormat/ElementIsEnum — or an element outside the four scalar forms.
func primitiveElementGoType(a *model.Attribute) (goType, reason string) {
	if a.ElementFormat != "" {
		return "", fmt.Sprintf("collection element format %q is not yet supported on a resource request", a.ElementFormat)
	}
	if a.ElementIsEnum {
		return "", "a collection of enum strings is not yet supported on a resource request"
	}
	switch a.ElementType {
	case "types.StringType":
		return "string", ""
	case "types.Int64Type":
		return "int64", ""
	case "types.Float64Type":
		return "float64", ""
	case "types.BoolType":
		return "bool", ""
	default:
		return "", fmt.Sprintf("collection element type %q is not yet supported on a resource request", a.ElementType)
	}
}

// elementPath names a collection attribute's element for a diagnostic, using
// the same "[]"/"{}" markers model.ChildPath appends to child paths.
func elementPath(a *model.Attribute) string {
	if isMapType(a.TfType) {
		return model.ChildPath(a.Path, "{}")
	}
	return model.ChildPath(a.Path, "[]")
}

// notNullOrUnknown guards a non-Required framework value field (expr, e.g.
// "state.Tags") against being sent to the SDK unset or still unresolved.
func notNullOrUnknown(expr string) string {
	return "!" + expr + ".IsNull() && !" + expr + ".IsUnknown()"
}

// buildObjectCollectionField derives a list-of-objects field's
// RequestFieldView. Get(ctx, &state) has already decoded a ListNestedAttribute
// into a native []*<ElemModel>, so no ElementsAs applies here: each element is
// instead built via New<Element>WithDefaults() and appended to a Go slice
// before the parent's setter runs. role is the *element's* own node in the
// calling role's request schema — roleChild reaches through the array first —
// so it both names the element component and narrows the element's fields.
func buildObjectCollectionField(a *model.Attribute, role *model.Schema, tfName, field, childState, target, sdkPackage string, unsupported *[]UnsupportedNode) (rf RequestFieldView, imports requestImports, ok bool) {
	refName := roleRequestRefName(a, role)
	if refName == "" {
		*unsupported = append(*unsupported, UnsupportedNode{
			Path: elementPath(a),
			Reason: "list element has no named request-side SDK type (its schema was never reached " +
				"through a $ref component in the Create or Update body)",
		})
		return RequestFieldView{}, requestImports{}, false
	}

	base := leafVar(tfName)
	loopVar, elemVar := base+"Item", base+"Element"
	childFields, imports := buildRequestFields(a.Children, role, loopVar, elemVar, sdkPackage, unsupported)
	rf = RequestFieldView{
		GoField:  field,
		Target:   target,
		Required: a.Required,
		Collection: &RequestCollectionView{
			Kind:          "object",
			ConvertVar:    base + "Elements",
			RangeExpr:     childState,
			LoopVar:       loopVar,
			ElemVar:       elemVar,
			Constructor:   sdkPackage + ".New" + refName + "WithDefaults()",
			ElementGoType: sdkPackage + "." + refName,
			Fields:        childFields,
		},
	}
	if !a.Required {
		rf.NullCheck = childState + " != nil"
	}
	return rf, imports, true
}

// requestValueExpr returns the unwrapped Go value expression a leaf's
// Set<Field> setter takes, read off stateExpr: directly for a plain scalar, or
// via a ParsedVar/ParseCall pair when the type needs a fallible parse
// (date-time, uuid), the local named from field so siblings cannot collide. An
// enum leaf casts to its named SDK type, qualified with sdkPackage. reason is
// non-empty for anything unsupported on the request side, naming why.
func requestValueExpr(a *model.Attribute, role *model.Schema, stateExpr, field, sdkPackage string) (expr, parsedVar, parseCall, reason string) {
	if a.IsEnum {
		refName := roleRequestRefName(a, role)
		if refName == "" {
			return "", "", "", "enum leaf has no named request-side SDK type (its schema was never reached through a $ref component)"
		}
		return sdkPackage + "." + refName + "(" + stateExpr + ".ValueString())", "", "", ""
	}

	switch a.GoType {
	case "types.String":
		switch a.Format {
		case "":
			return stateExpr + ".ValueString()", "", "", ""
		case "date-time":
			name := lowerFirst(field) + "Parsed"
			return name, name, "time.Parse(time.RFC3339, " + stateExpr + ".ValueString())", ""
		case "uuid":
			name := lowerFirst(field) + "Parsed"
			return name, name, "uuid.Parse(" + stateExpr + ".ValueString())", ""
		default:
			return "", "", "", fmt.Sprintf("string format %q is not yet supported on a resource request", a.Format)
		}
	case "types.Bool":
		if a.Format != "" {
			return "", "", "", fmt.Sprintf("bool format %q is not yet supported on a resource request", a.Format)
		}
		return stateExpr + ".ValueBool()", "", "", ""
	case "types.Int64":
		switch a.Format {
		case "", "int64":
			return stateExpr + ".ValueInt64()", "", "", ""
		case "int32":
			return "int32(" + stateExpr + ".ValueInt64())", "", "", ""
		default:
			return "", "", "", fmt.Sprintf("integer format %q is not yet supported on a resource request", a.Format)
		}
	case "types.Float64":
		switch a.Format {
		case "", "double":
			return stateExpr + ".ValueFloat64()", "", "", ""
		case "float":
			return "float32(" + stateExpr + ".ValueFloat64())", "", "", ""
		default:
			return "", "", "", fmt.Sprintf("number format %q is not yet supported on a resource request", a.Format)
		}
	default:
		return "", "", "", fmt.Sprintf("request-settable field of GoType %q is not yet supported on a resource request", a.GoType)
	}
}

// buildOneOfRequestField derives one union field's RequestFieldView: how the
// configured variant block becomes the SDK oneOf wrapper this role's setter
// takes. role, this union's node in the *calling role's* request schema, is the
// source of every SDK identity; the merged tree supplies the Terraform side
// (which blocks exist, their names and fields). The two are correlated on the
// name mergeOneOf published, never on position.
func buildOneOfRequestField(
	a *model.Attribute,
	role *model.Schema,
	tfName, field, modelExpr, target, sdkPackage string,
	unsupported *[]UnsupportedNode,
) (rf RequestFieldView, imports requestImports, ok bool) {
	env := a.OneOf
	fail := func(reason string) (RequestFieldView, requestImports, bool) {
		*unsupported = append(*unsupported, UnsupportedNode{Path: env.Path, Reason: reason})
		return RequestFieldView{}, requestImports{}, false
	}

	if isCollectionForm(a.TfType) {
		return fail("a collection whose element is a oneOf union is not yet settable on a resource request; " +
			"the union itself is supported at its own position")
	}
	if role == nil || role.OneOf == nil {
		return fail("this role's own request body does not describe a oneOf union here, so the SDK oneOf " +
			"wrapper to build cannot be resolved")
	}
	if role.OneOf.SDKType == "" {
		return fail("this role's oneOf union has no resolved SDK wrapper type (internal/sdkbind left it empty)")
	}

	roleVariants := make(map[string]model.OneOfVariant, len(role.OneOf.Variants))
	for _, v := range role.OneOf.Variants {
		roleVariants[model.StripOneOfRoleSuffix(v.TFName)] = v
	}

	view := &RequestOneOfView{
		TFName:   tfName,
		SDKType:  sdkPackage + "." + role.OneOf.SDKType,
		Var:      lowerFirst(field) + "Union",
		MatchVar: lowerFirst(field) + "Matches",
		Variants: make([]RequestOneOfVariantView, 0, len(env.Variants)),
	}
	names := make([]string, 0, len(env.Variants))

	for _, v := range env.Variants {
		roleVariant, declared := roleVariants[v.TFName]
		if !declared {
			// mergeOneOf refuses to correlate bodies whose alternative sets
			// differ, so this means the merged envelope and the role schema
			// disagree; dropping the branch would discard a configurable one.
			return fail(fmt.Sprintf("variant %q is in the merged schema but not in this role's own union", v.TFName))
		}
		if roleVariant.SDKConstructor == "" {
			return fail(fmt.Sprintf("variant %q has no SDK convenience constructor for this role's wrapper", v.TFName))
		}
		variant, variantImports, reason := oneOfRequestVariant(v, roleVariant, modelExpr, sdkPackage, unsupported)
		if reason != "" {
			return fail(reason)
		}
		view.Variants = append(view.Variants, variant)
		imports = imports.or(variantImports)
		names = append(names, v.TFName)
	}
	view.SelectionMessage = fmt.Sprintf(
		"%s: exactly one of %s must be set, got %%d", env.Path, strings.Join(quoteAll(names), " or "))

	// The envelope pointer is guarded even when the union is required, so an
	// absent block is reported rather than dereferenced. Same guard as a nested
	// object, so it travels on the same field.
	rf = RequestFieldView{
		GoField:   field,
		Target:    target,
		Required:  a.Required,
		NullCheck: modelExpr + " != nil",
		OneOf:     view,
	}
	// The selection diagnostic is a fmt.Sprintf.
	return rf, imports.or(requestImports{fmt: true}), true
}

// oneOfRequestVariant derives one alternative's expansion. v is the merged
// tree's projection (the Terraform blocks and their fields), roleVariant this
// role's binding for the same alternative (its SDK member and constructor).
// reason is non-empty for a shape the request path cannot build.
func oneOfRequestVariant(
	v model.OneOfEnvelopeVariant,
	roleVariant model.OneOfVariant,
	modelExpr, sdkPackage string,
	unsupported *[]UnsupportedNode,
) (RequestOneOfVariantView, requestImports, string) {
	blockExpr := modelExpr + "." + v.GoField
	elemVar := lowerFirst(v.GoField) + "Variant"
	variant := RequestOneOfVariantView{TFName: v.TFName, ModelExpr: blockExpr, ElemVar: elemVar}
	// The SDK takes an address for every member except a free-form object,
	// which it emits as an already-nil-able bare map.
	argument := elemVar
	if !roleVariant.SDKPointer {
		argument = "*" + elemVar
	}

	if !v.ValueWrapped {
		variant.Constructor = sdkPackage + ".New" + roleVariant.SDKField + "WithDefaults()"
		fields, imports := buildRequestFields(
			v.Attribute.Children, roleVariant.Schema, blockExpr, elemVar, sdkPackage, unsupported)
		variant.Fields = fields
		variant.WrapCall = sdkPackage + "." + roleVariant.SDKConstructor + "(" + argument + ")"
		return variant, imports, ""
	}

	// A value-wrapped alternative's SDK member *is* the value: nothing to
	// construct and nothing to set on, so the single "value" child is converted
	// and its address taken. A list or map alternative has no scalar to take.
	if len(v.Attribute.Children) != 1 {
		return variant, requestImports{}, fmt.Sprintf(
			"variant %q is value-wrapped but has %d children, expected exactly one",
			v.TFName, len(v.Attribute.Children))
	}
	value := v.Attribute.Children[0]
	if !isLeafType(value.TfType) {
		return variant, requestImports{}, fmt.Sprintf(
			"variant %q wraps a %s, and the request path can only convert a scalar into an SDK oneOf member; "+
				"give the alternative a named schema component so it becomes an object variant",
			v.TFName, value.TfType)
	}
	valueField := model.SdkName(tfNameOf(value.Path))
	expr, parsedVar, parseCall, reason := requestValueExpr(value, roleVariant.Schema, blockExpr+"."+valueField, valueField, sdkPackage)
	if reason != "" {
		return variant, requestImports{}, fmt.Sprintf("variant %q: %s", v.TFName, reason)
	}
	if !roleVariant.SDKPointer {
		return variant, requestImports{}, fmt.Sprintf(
			"variant %q is a scalar the SDK declares unpointered, which only a free-form object member is; "+
				"the generator has no address to take", v.TFName)
	}
	// The member is the scalar itself, so the local holds a value and the
	// constructor takes its address, never the "*elemVar" of the object case.
	variant.WrapCall = sdkPackage + "." + roleVariant.SDKConstructor + "(&" + elemVar + ")"
	variant.Value = &RequestOneOfValueView{
		ValueExpr: expr, ParsedVar: parsedVar, ParseCall: parseCall, TFName: tfNameOf(value.Path),
	}
	u, _, t := parseCallImports(parseCall)
	return variant, requestImports{uuid: u, time: t}, ""
}

// roleRequestRefName is the SDK component to construct at this node *for the
// role being rendered*, answering roleChild's two outcomes differently. When
// roleChild resolved a node, that node's own RefName is the answer even when
// empty — empty means this body declared the schema inline, so the SDK
// generated no component and the caller's diagnostic is correct. The merged
// a.RequestModelRefName is Create-first with Update fallback, so substituting
// it would spell one name where the SDK declares two. Only for a nil node,
// where roleChild means "unknown", is the merged name the only name there is.
func roleRequestRefName(a *model.Attribute, role *model.Schema) string {
	if role != nil {
		return role.RefName
	}
	return a.RequestModelRefName
}
