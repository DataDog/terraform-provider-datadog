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

import "github.com/terraform-providers/terraform-provider-datadog/generator/internal/model"

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
	// "GetIncidentsApiV2". It is set when ApiInstances already exposes the API.
	APIAccessor string
	// APIConstructor is the pinned SDK constructor used when ApiInstances does
	// not expose APIStruct, e.g. "NewCaseManagementApi". Exactly one of
	// APIAccessor and APIConstructor is populated after accessor resolution.
	APIConstructor string
	// UsesUUID adds the google/uuid import and SDK-input parsing blocks.
	UsesUUID bool
	// UsesTime adds the time import and SDK-input date parsing blocks.
	UsesTime bool
	// UsesJSON adds encoding/json and normalized JSON custom-type imports.
	UsesJSON bool

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

	// UsesFmt selects the "fmt" import. The template cannot decide this: fmt is
	// reached from the by-id not-found message, from a plural filter hash, and from
	// a oneOf's ambiguous-match diagnostic, and go/format does not prune an unused
	// import — so an over- or under-estimate here is a compile error either way.
	UsesFmt bool

	// Dropped lists response members skipped from the rendered view (e.g.
	// relationships), surfaced as diagnostics in the run report. It does not
	// affect rendering.
	Dropped []DroppedMember
}

// OneOfEnvelopeView is one generated oneOf envelope: the Terraform model that
// holds one pointer per alternative, and the Datadog go-sdk wrapper it maps to.
//
// The two identities are deliberately separate: GoModel is derived from
// the envelope's Terraform name, which for an inline union is path-derived and
// names no SDK type, while SDKType is what internal/sdkbind resolved by walking
// the operation's SDK root.
type OneOfEnvelopeView struct {
	// Name is the parser's envelope identity, the key this view is deduplicated on.
	Name string
	// GoModel is the generated Go struct holding one pointer field per variant.
	GoModel string
	// SDKType is the SDK oneOf wrapper struct, e.g. "ActionConnectionIntegration".
	SDKType string
	// Path is the union's schema path, for diagnostics the mapper emits.
	Path string
	// Optional records that the whole envelope may be absent because its containing
	// field is optional or nullable, which is the one case where zero populated SDK
	// members is not an error.
	Optional bool
	// Variants are the alternatives, ordered by TFName as the projection ordered
	// them, so generated output cannot depend on OpenAPI alternative order.
	Variants []OneOfVariantView
}

// OneOfVariantView is one alternative of a OneOfEnvelopeView.
type OneOfVariantView struct {
	// TFName is the nested block name; GoField the pointer field on the envelope
	// model whose non-nil-ness selects this variant.
	TFName  string
	GoField string
	// GoModel is the generated struct for this variant's own fields.
	GoModel string
	// SDKField is the SDK wrapper member that selects this alternative and
	// SDKConstructor its `<Member>As<Union>` convenience constructor. SDKPointer is
	// false only for a free-form object, which the SDK emits as a bare map.
	SDKField       string
	SDKConstructor string
	SDKPointer     bool
	// ValueWrapped marks a non-object alternative, whose block holds a single
	// child named "value" rather than exposing fields directly.
	ValueWrapped bool
	// SDKVar and ModelVar are the locals the mapper declares for the unwrapped SDK
	// member and the Terraform variant model it populates. Scalars and Lists are
	// expressed against them, so they must be used verbatim by the mapper.
	SDKVar   string
	ModelVar string
	// Scalars and Lists are this variant's field assignments, derived by the same
	// walk that produced its model struct. Not rendered yet — the mapper adds the partial
	// that consumes them.
	Scalars []StateAssignment
	Lists   []ListAssignment
}

// DroppedMember is one response member omitted from the rendered view, carrying
// the severity its omission warrants in the run report. A member with no
// Terraform representation is informational; one dropped to resolve a name
// collision is a warning, because data the API exposes is not surfaced.
type DroppedMember struct {
	Message  string
	Severity model.DiagnosticSeverity
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
	// Arguments are required positional SDK call arguments in call order.
	Arguments []SDKArgumentView

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
	// HashInputs includes every Terraform input that identifies the returned
	// record or collection, both required positional arguments and optional filters.
	HashInputs []FilterParamView
}

// SDKArgumentView is one rendered positional SDK call argument. UUID arguments
// carry a preparation variable; all other scalar arguments render Expression
// directly at the call site.
type SDKArgumentView struct {
	Expression string
	UUIDVar    string
	UUIDSource string
	TimeVar    string
	TimeSource string
	TimeLayout string
	TFName     string
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
	// Setter is the SDK With* method. Empty retains the legacy direct-field form
	// used by parser-shaped unit fixtures without resolved SDK bindings.
	Setter string
	// UUIDVar and UUIDSource request a uuid.Parse preparation inside the filter's
	// non-null guard before ValueExpr is passed to Setter. They are populated only
	// when the pinned SDK setter accepts uuid.UUID.
	UUIDVar    string
	UUIDSource string
	// TimeVar, TimeSource, and TimeLayout request a time.Parse preparation for
	// SDK setters accepting time.Time.
	TimeVar    string
	TimeSource string
	TimeLayout string
	// TFName is the Terraform attribute name used in parse diagnostics.
	TFName string
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
	// CustomType is the optional framework custom type expression for a leaf.
	CustomType string
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
	// ItemLists are the item's collection-valued assignments, rendered by
	// "renderList" after the struct literal (they cannot sit in the literal: a
	// primitive-terminal collection uses a two-value ValueFrom helper, while an
	// object list uses a loop).
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

// ListAssignment is one non-scalar state assignment rendered by the updateState
// "renderList" partial. Despite the name it covers every shape that is not a bare
// leaf: a primitive list maps the SDK slice into a types.List via
// types.ListValueFrom; an object list loops the SDK elements into a generated
// nested model slice, recursing through Scalars (the element's leaf fields) and
// Lists (its nested list fields); an object_single maps one nested object into a
// generated model pointer, assigned once instead of looped; and a oneof unwraps an
// SDK oneOf wrapper through OneOf. JSON stores an unconstrained SDK value as
// normalized JSON text. All forms are guarded by an Ok-getter so an absent field
// stays null. Primitive-terminal collections retain whether they
// are lists or maps so the matching framework conversion helper is rendered.
//
// A oneOf envelope rides this type rather than a parallel one because it needs
// exactly the same placement plumbing — it can appear at the top level, inside a
// nested object, inside a list element, or inside another envelope's variant — and
// duplicating that composition for one extra shape would be the larger cost.
type ListAssignment struct {
	// Kind is "primitive", "json", "object", "object_single" (a single nested
	// object, assigned once rather than appended in a loop), or "oneof" (see OneOf).
	Kind string
	// ContainerKind is "list" or "map" for Kind == "primitive". It selects
	// the framework ValueFrom and Null constructors used by the template.
	ContainerKind string
	// LHS is the assignment target, e.g. "state.VisibleModules" (top level) or
	// "entriesModel.TagFilters" (nested element field).
	LHS string
	// GetterOk is the guarded optional getter returning (slice pointer, bool),
	// e.g. "attributes.GetVisibleModulesOk()".
	GetterOk string
	// Var is the local bound from GetterOk (a pointer to the slice).
	Var string
	// ElementType is the framework element type for a primitive-terminal
	// collection, e.g. "types.ListType{ElemType: types.StringType}". Empty for
	// an object list.
	ElementType string
	// Path identifies JSON conversion diagnostics.
	Path string

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

	// OneOf backs Kind == "oneof" and is nil otherwise.
	OneOf *OneOfAssignment
}

// OneOfAssignment maps one Datadog go-sdk oneOf wrapper into one generated
// Terraform envelope model.
//
// The generated code inspects *every* wrapper member rather than taking the first
// non-nil one: the SDK's own MarshalJSON and GetActualInstance are first-match, and
// the contract requires zero, multiple, or unparsed to be reported at the union's schema
// path instead of silently resolving to one branch. Exactly one populated member
// assigns the envelope; anything else either leaves it absent (permitted only when
// Optional) or raises a diagnostic.
type OneOfAssignment struct {
	// Path is the union's schema path, named in every diagnostic this emits.
	Path string
	// SDKType is the wrapper struct, named in diagnostics so a reader can find it
	// in the SDK.
	SDKType string
	// GoModel is the generated envelope struct, and LHS where it is stored, e.g.
	// "state.Integration". For Collection, LHS is appended to.
	GoModel string
	LHS     string
	// GetterOk is the ordinary optional getter that yields the wrapper off its
	// parent, e.g. "attributes.GetIntegrationOk()". Only the wrapper's *members*
	// lack getters; the wrapper itself is a normal field on its parent.
	GetterOk string
	// Var is the local bound from GetterOk.
	Var string
	// Receiver is the expression the members are read off: Var for a single
	// envelope, LoopVar inside a collection.
	Receiver string
	// ModelVar is the local accumulating the envelope model, MatchVar the local
	// counting populated members.
	ModelVar string
	MatchVar string
	// Optional permits zero populated members, which is how an absent nullable
	// union arrives. When false, zero members is an error.
	Optional bool
	// Collection marks a list whose element is an envelope; LoopVar is then the
	// per-element local.
	Collection bool
	LoopVar    string
	// Variants are the alternatives, ordered by Terraform variant name.
	Variants []OneOfVariantAssignment
}

// OneOfVariantAssignment unwraps one alternative of a OneOfAssignment.
//
// Every field here is a function of the envelope alone, never of the site using it,
// so one reusable oneOf component's variant bodies are computed once and shared.
// That is why the envelope-model target is carried as GoField and composed with the
// enclosing OneOfAssignment.ModelVar at render time rather than being a
// precomputed LHS: the model var differs per use site, the field does not.
type OneOfVariantAssignment struct {
	// SDKField is the wrapper member whose non-nil-ness selects this alternative,
	// and SDKVar the local bound to it. SDKPointer is false only for a free-form
	// object, which the SDK emits as a bare, already-nil-able map.
	SDKField   string
	SDKVar     string
	SDKPointer bool
	// GoField is the envelope-model field this variant assigns, e.g.
	// "AwsIntegration"; GoModel is its generated struct and ModelVar the local
	// accumulating it.
	GoField  string
	GoModel  string
	ModelVar string
	// Value is set for a value-wrapped (non-object) alternative, whose SDK member is
	// the scalar itself and therefore has no getters to read fields through: the
	// single "value" field is assigned by dereferencing SDKVar directly.
	Value *StateAssignment
	// Scalars are this alternative's leaf fields and Lists its non-scalar ones,
	// expressed against SDKVar and ModelVar. A union nested inside this alternative
	// arrives in Lists with Kind "oneof", so recursion needs no extra channel.
	Scalars []StateAssignment
	Lists   []ListAssignment
}
