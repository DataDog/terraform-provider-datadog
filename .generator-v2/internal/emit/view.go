// Package emit turns the generator's internal model into Terraform provider
// Go source. It owns the code templates and the pipeline that walks
// deterministically-sorted Artifacts, renders each through the matching
// template, canonicalizes the result with go/format, and writes it.
//
// The templates never derive anything: naming, attribute partitioning, SDK-call
// resolution, and state mapping are all computed in Go and handed to the
// template as a fully-populated *View. That split is deliberate :
// it keeps the .tmpl files flat enough to read and grep, and it keeps the
// fiddly logic in code that unit tests can pin down.
package emit

// Cardinality selects which data-source template renders an artifact. A
// singular data source resolves exactly one item whereas a plural
// data source returns a filtered list of items.
type Cardinality int

const (
	// Singular renders data_source_singular.go.tmpl.
	Singular Cardinality = iota
	// Plural renders data_source_plural.go.tmpl.
	Plural
)

// DataSourceView is the render-ready data context for a data-source template.
//
// Every field is derived from a *model.Artifact by the emit builder; the
// templates contain only iteration and layout. Keeping derivation out of the
// templates is what lets the same recursive partials serve both the singular
// and plural shapes.
type DataSourceView struct {
	// Cardinality picks the singular or plural template.
	Cardinality Cardinality

	// TypeName is the Terraform type suffix written to response.TypeName
	TypeName string
	// GoName is the lowerCamel identifier base, e.g. "incidentType". It builds
	// the Go type names <GoName>DataSource and <GoName>DataSourceModel and,
	// title-cased, the New<GoName>DataSource constructor.
	GoName string
	// Description is the schema-level data-source descriptiond
	Description string

	// SDKPackage is the versioned datadog-api-client-go package selector, e.g.
	// "datadogV2".
	SDKPackage string
	// APIStruct is the SDK API client struct embedded in the data source, e.g.
	// "IncidentsApi".
	APIStruct string
	// APIAccessor is the FrameworkProvider accessor returning that client, e.g.
	// "GetIncidentsApiV2".
	APIAccessor string

	// ByID and Searchable select how a singular data source resolves its one
	// record, driving the Read body and the "id" attribute: ByID only → by-id
	// lookup (id Required); Searchable only → search (id Computed); both → id
	// optional (id Optional+Computed, lookup when set else search).
	ByID       bool
	Searchable bool

	// Read describes the by-id SDK call. Set when ByID.
	Read SDKReadView
	// Search describes the list SDK call a singular data source searches. Set
	// when Searchable; carries the list-call fields (Paginated/ItemType/
	// OptionalParamsType) and the Filters derived from query parameters.
	Search SDKReadView

	// Models are the Go model structs to declare: the parent data-source model
	// first, then any nested item structs, in deterministic order.
	Models []ModelStructView

	// Schema holds the attributes and blocks rendered into the Schema method.
	Schema SchemaView

	// State holds what updateState assigns back into the model.
	State StateView

	// Dropped lists response members skipped from the rendered view (e.g.
	// relationships), surfaced as info diagnostics in the run report. It does
	// not affect rendering.
	Dropped []string
}

// SDKReadView describes the datadog-api-client-go call that backs Read.
type SDKReadView struct {
	// Method is the SDK method name. For a singular data source this is a
	// get-by-id (e.g. "GetIncidentType"); for a plural one it is the list call
	// (e.g. "ListTeams"), to which the template may append "WithPagination".
	Method string
	// ResponseType is the SDK response type returned by a singular Method, e.g.
	// "IncidentTypeResponse". It names the updateState receiver.
	ResponseType string

	// The fields below are plural-only.

	// Paginated selects the "<Method>WithPagination" iterator form over a
	// single-call form.
	Paginated bool
	// ItemType is the SDK element type yielded by the list call, e.g. "Team".
	ItemType string
	// OptionalParamsType is the SDK optional-parameters struct passed to the
	// list call, e.g. "ListTeamsOptionalParameters".
	OptionalParamsType string
	// Filters maps each optional query parameter from the model onto the
	// request's optional-parameters struct.
	Filters []FilterParamView
}

// FilterParamView maps one optional query parameter from the Terraform model
// onto the SDK's optional-parameters struct, e.g.
//
//	if !state.FilterKeyword.IsNull() {
//	    optionalParams.FilterKeyword = state.FilterKeyword.ValueStringPointer()
//	}
type FilterParamView struct {
	// StateField is the model field holding the filter value, e.g. "FilterKeyword".
	StateField string
	// ParamField is the SDK optional-params field set from it, e.g. "FilterKeyword".
	ParamField string
	// ValueExpr is the model accessor producing the SDK value, e.g.
	// "ValueStringPointer()".
	ValueExpr string
}

// SchemaView is the attribute/block split rendered into the Schema method. The
// "id" attribute is handled by the template itself (a required lookup key for
// singular, utils.ResourceIDAttribute() for plural), so it does not appear
// here. Attributes holds top-level leaves; Blocks holds top-level nested
// objects/lists (for a plural data source, that includes the items block).
type SchemaView struct {
	Attributes []AttrView
	Blocks     []AttrView
}

// AttrView is one node of the Terraform schema tree. A leaf renders a typed
// schema.*Attribute; a block (IsBlock) renders a schema.*NestedBlock and
// recurses through its own Attributes and Blocks.
type AttrView struct {
	// TFName is the Terraform attribute key, snake_case, e.g. "link_count".
	TFName string
	// TFType is the framework attribute type token for a leaf, e.g.
	// "schema.StringAttribute". Ignored for blocks (ListBlock picks the type).
	TFType string
	// ElementType is the framework attr.Type rendered on a schema.ListAttribute,
	// e.g. "types.StringType". Non-empty only for a collection-of-primitive leaf.
	ElementType string
	// Description is the attribute description (repo convention: always set).
	Description string

	Required  bool
	Optional  bool
	Computed  bool
	Sensitive bool

	// IsBlock marks a nested object/list, rendered under a Blocks map.
	IsBlock bool
	// ListBlock renders schema.ListNestedBlock when true and
	// schema.SingleNestedBlock when false. Ignored unless IsBlock.
	ListBlock bool

	// Attributes and Blocks are the leaf and nested children of a block; both
	// are empty for a leaf attribute.
	Attributes []AttrView
	Blocks     []AttrView
}

// ModelStructView is one Go struct in the generated file's data model.
type ModelStructView struct {
	// Name is the Go struct type name, e.g. "incidentTypeDataSourceModel" or
	// "TeamModel".
	Name string
	// Fields are emitted in declaration order; a field's leading Comment groups
	// it in the output (e.g. "Query Parameters", "Results").
	Fields []ModelFieldView
}

// ModelFieldView is one field of a generated model struct.
type ModelFieldView struct {
	// Comment, when non-empty, is emitted as a // line above the field and is
	// preceded by a blank line so successive groups read clearly.
	Comment string
	// GoField is the exported field name, e.g. "LinkCount".
	GoField string
	// GoType is the field type, e.g. "types.String", "types.Int64",
	// "[]*TeamModel".
	GoType string
	// TFName is the tfsdk struct-tag value, e.g. "link_count".
	TFName string
}

// StateView is what the generated updateState method writes back into the
// model. The assignment expressions themselves are produced by the
// response-mapper builder; this view only carries them so the template
// can lay them out. Singular data sources use Preamble + Assignments; plural
// data sources use the Item* / IDHashExpr fields.
type StateView struct {
	// ParamName / ParamType are the updateState record parameter for a singular
	// data source: ("resp", "*pkg.XResponse") when the record is a by-id response,
	// ("data", "*pkg.XItem") when it is a list element (search/both).
	ParamName string
	ParamType string
	// Preamble holds raw statements emitted before the assignments, e.g.
	// "attributes := resp.Data.GetAttributes()". Singular only.
	Preamble []string
	// Assignments are the singular record assignments, each rendered as a guarded
	// block: "if <Var>, ok := <GetterOk>; ok && <Var> != nil { <LHS> = <RHS> }",
	// so an absent field stays null rather than a zero value.
	Assignments []StateAssignment
	// Lists are the singular record's list-valued assignments (collection-of-primitive
	// and list-of-object), rendered by the "renderList" partial after Assignments.
	Lists []ListAssignment

	// The fields below are plural-only.

	// ItemStruct is the Go item struct built per element, e.g. "TeamModel".
	ItemStruct string
	// ItemField is the parent-model slice field assigned the result, e.g.
	// "Teams".
	ItemField string
	// ItemFields are the item struct's literal fields ("<GoField>: <RHS>"),
	// evaluated against the loop variable "item".
	ItemFields []StateAssignment
	// ItemLists are the item's list-valued assignments, rendered by "renderList"
	// after the struct literal (they cannot sit in the literal: a primitive list
	// is a two-value ListValueFrom, an object list is a loop).
	ItemLists []ListAssignment
}

// StateAssignment is a single assignment rendered in updateState. For a
// singular assignment LHS is the full target ("state.Name") and RHS the value
// expression; for a plural item field LHS is the struct field name ("Handle").
//
// Var and GetterOk back the guarded singular form: Var is the local bound from
// the SDK's optional getter GetterOk (e.g. "name" from "attributes.GetNameOk()"),
// and RHS reads through it (e.g. "types.StringValue(*name)"). They are empty for
// plural item fields, which render unguarded.
type StateAssignment struct {
	LHS      string
	RHS      string
	Var      string
	GetterOk string
}

// ListAssignment is a nested-state assignment rendered by the updateState
// "renderList" partial. A primitive list maps the SDK slice into a types.List via
// types.ListValueFrom; an object list loops the SDK elements into a generated
// nested model slice, recursing through Scalars (the element's leaf fields) and
// Lists (its nested list fields); an object_single maps one nested object into a
// generated model pointer, assigned once instead of looped. All forms are guarded
// by an Ok-getter so an absent field stays null.
type ListAssignment struct {
	// Kind is "primitive", "object", or "object_single" (a single nested object,
	// assigned once rather than appended in a loop).
	Kind string
	// LHS is the assignment target, e.g. "state.VisibleModules" (top level) or
	// "entriesModel.TagFilters" (nested element field).
	LHS string
	// GetterOk is the guarded optional getter returning (slice pointer, bool),
	// e.g. "attributes.GetVisibleModulesOk()".
	GetterOk string
	// Var is the local bound from GetterOk (a pointer to the slice).
	Var string
	// ElementType is the framework element type for a primitive list, e.g.
	// "types.StringType". Empty for an object list.
	ElementType string

	// The fields below back an object list (Kind == "object").

	// LoopVar is the per-element loop variable, e.g. "entriesItem".
	LoopVar string
	// ElemVar is the per-element model accumulator, e.g. "entriesModel".
	ElemVar string
	// ElemStruct is the generated nested model struct, e.g. "EntriesModel".
	ElemStruct string
	// Scalars are the element's leaf fields, assigned off LoopVar into ElemVar.
	Scalars []StateAssignment
	// Lists are the element's nested list fields (recursion).
	Lists []ListAssignment
}
