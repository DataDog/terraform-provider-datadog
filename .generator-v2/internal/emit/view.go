// Package emit turns the generator's internal model into Terraform provider
// Go source: it owns the code templates, renders each Artifact through the
// matching one, canonicalizes the result with go/format, and writes it.
// Templates derive nothing — naming, attribute partitioning, SDK-call
// resolution and state mapping all arrive precomputed on a *View.
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
// Every field is derived from a *model.Artifact, which is what lets one set of
// recursive partials serve both the singular and plural shapes.
type DataSourceView struct {
	// Cardinality picks the singular or plural template.
	Cardinality Cardinality

	// TypeName is the Terraform type suffix written to response.TypeName
	TypeName string
	// GoName is the lowerCamel identifier base, e.g. "incidentType". It builds
	// the Go type names <GoName>DataSource and <GoName>DataSourceModel and,
	// title-cased, the New<GoName>DataSource constructor.
	GoName string
	// Description is the schema-level data-source description.
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
	// UsesStrconv adds the strconv import for id-bound SDK-input parsing blocks
	// (e.g. an integer path parameter aliased to the string "id" attribute).
	UsesStrconv bool

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

// OneOfEnvelopeView is one generated oneOf envelope: the Terraform model
// holding one pointer per alternative, and the go-sdk wrapper it maps to. The
// two identities stay separate because GoModel comes from the envelope's
// Terraform name — path-derived for an inline union, naming no SDK type —
// while SDKType is resolved by walking the operation's SDK root.
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
	// TFName is the nested attribute name; GoField the pointer field on the envelope
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
	// Scalars and Lists are this variant's field assignments, derived by the
	// same walk that produced its model struct.
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

// SDKArgumentView is one rendered positional SDK call argument. An argument
// that must be recovered by parsing a string (a uuid-typed argument, or any
// non-string argument aliased from the always-string "id" attribute) carries
// a preparation variable and the call that parses it; all other scalar
// arguments render Expression directly at the call site.
type SDKArgumentView struct {
	Expression string
	ParsedVar  string
	ParseCall  string
	TFName     string
	// GoType is the SDK parameter's Go type, used to decide whether an update
	// body's data.id can reuse this argument's parsed local or must parse its
	// own.
	GoType string
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
	// Setter is the SDK With* method. Empty renders a direct field assignment
	// instead, the form used when no SDK bindings were resolved.
	Setter string
	// ParsedVar and ParseCall request a parse preparation inside the filter's
	// non-null guard before ValueExpr is passed to Setter. They are populated only
	// when the pinned SDK setter accepts a type (uuid.UUID) that must be recovered
	// from the filter's string-typed model field.
	ParsedVar string
	ParseCall string
	// TFName is the Terraform attribute name used in parse diagnostics.
	TFName string
}

// SchemaView is the attribute/block split rendered into the Schema method. The
// "id" attribute is handled by the template itself (a required lookup key for
// singular, utils.ResourceIDAttribute() for plural), so it does not appear
// here. Attributes holds top-level leaves; Blocks holds top-level nested
// objects/lists (for a plural data source, that includes the items block).
type SchemaView struct {
	Attributes          []AttrView
	Blocks              []AttrView
	HasWriteOnlySecrets bool
	IncludeResourceID   bool
}

// WriteOnlySecretView is one generated write-only-only Terraform pair. The
// Terraform names and descriptions drive schema expansion, while SDKField
// deliberately retains the original OpenAPI setter identity for request
// routing. ParentBlocks contains only statically addressed single nested
// attributes and is empty for a resource-root secret.
type WriteOnlySecretView struct {
	OriginalAttr         string
	WriteOnlyAttr        string
	TriggerAttr          string
	SDKField             string
	ParentBlocks         []string
	RequiredOnCreate     bool
	RequiredOnUpdate     bool
	WriteOnlyDescription string
	TriggerDescription   string
	// ConfigVar is the package-scoped configuration shared by schema expansion
	// and lifecycle handlers. HandlerVar and ResultVar are method-local names;
	// GetterCall selects Create or Update retrieval on per-call copies.
	ConfigVar  string
	HandlerVar string
	ResultVar  string
	GetterCall string
}

// AttrView is one node of the Terraform schema tree. A leaf renders a typed
// schema.*Attribute; a nested container (IsBlock) renders a
// schema.*NestedAttribute and recurses through its own Attributes and Blocks.
type AttrView struct {
	// TFName is the Terraform attribute key, snake_case, e.g. "link_count".
	TFName string
	// TFType is the framework attribute type token for a leaf, e.g.
	// "schema.StringAttribute". Ignored for nested containers (ListBlock picks
	// the type).
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

	// IsBlock marks a nested object/list; templates render it in an Attributes map.
	IsBlock bool
	// ListBlock renders schema.ListNestedAttribute when true and
	// schema.SingleNestedAttribute when false. Ignored unless IsBlock.
	ListBlock bool

	// Validators renders the "Validators: []validator.<T>{...}" field, one
	// rendered constructor call per entry (e.g. `stringvalidator.OneOf("a", "b")`).
	// An attribute is either a leaf or a nested object, never both, so one list
	// serves both with ValidatorType naming the element type.
	Validators []string
	// ValidatorType is the validator.<T> slice element type for Validators
	// ("String" for a leaf, "Object" for a nested object attribute, whose
	// variants enforce exactly-one selection at configuration-validation time).
	// Empty unless Validators is non-empty.
	ValidatorType string
	// PlanModifiers renders the "PlanModifiers: []planmodifier.<T>{...}" field,
	// one rendered constructor call per entry (e.g. `stringplanmodifier.UseStateForUnknown`,
	// with "()" appended by the template). Empty unless the underlying
	// attribute carries plan modifiers.
	PlanModifiers []string
	// PlanModifierType is the planmodifier.<T> slice element type matching this
	// attribute's GoType (e.g. "String", "Object", "List"). Empty unless
	// PlanModifiers is non-empty.
	PlanModifierType string

	// WriteOnlySecret replaces an original write-only leaf. It is rendered by
	// the containing attribute map through fwutils rather than as an ordinary
	// schema attribute. HasWriteOnlySecrets marks a nested container whose
	// immediate Attributes slice contains at least one such replacement.
	WriteOnlySecret     *WriteOnlySecretView
	HasWriteOnlySecrets bool
	IncludeResourceID   bool

	// Attributes and Blocks are the leaf and nested children of a container; both
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
// model. It only carries the assignment expressions so the template can lay
// them out. Singular data sources use Preamble + Assignments; plural ones use
// the Item* / IDHashExpr fields.
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

// StateAssignment is a single assignment rendered in updateState. LHS is the
// full target ("state.Name") for a singular assignment, or just the field name
// ("Handle") for a plural item field. Var is bound from the SDK optional getter
// GetterOk and RHS reads through it ("types.StringValue(*name)"); both are
// empty for plural item fields, which render unguarded.
type StateAssignment struct {
	LHS      string
	RHS      string
	Var      string
	GetterOk string
}

// ListAssignment is one non-scalar state assignment rendered by updateState's
// "renderList" partial: a primitive list/map converted via types.ListValueFrom,
// an object list looped into a generated nested-model slice (recursing through
// Scalars and Lists), an object_single assigned once, or a oneof unwrapped
// through OneOf. All forms are guarded by an Ok-getter so an absent field stays
// null. Envelopes ride this type to reuse its nesting/placement plumbing.
type ListAssignment struct {
	// Kind is "primitive", "object", "object_single" (a single nested object,
	// assigned once rather than appended in a loop), or "oneof" (see OneOf).
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
	// PreserveExisting is enabled for resource response mapping so fields the
	// request owns but the API never returns survive a nested-model rebuild.
	PreserveExisting bool
	// PreserveConfiguredPresence guards response mapping with an existing model
	// check so an optional configuration-owned container remains absent when the
	// API returns server defaults for it.
	PreserveConfiguredPresence bool

	// The fields below back an object list (Kind == "object").

	// LoopVar is the per-element loop variable, e.g. "entriesItem".
	LoopVar string
	// LoopIndex and ExistingVar let object collections retain request-only
	// fields from the corresponding configured element while rebuilding the
	// response-backed list.
	LoopIndex   string
	ExistingVar string
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

// OneOfAssignment maps one go-sdk oneOf wrapper into one generated Terraform
// envelope model. The generated code inspects every wrapper member rather than
// taking the first non-nil one, so zero, multiple or unparsed members are
// reported at the union's schema path instead of silently resolving to one
// branch. Zero populated members is an error unless Optional.
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
	// PreserveExisting retains the previously selected variant model when the
	// response selects that same variant, preserving request-only leaves.
	PreserveExisting bool
	// PreserveConfiguredPresence has the same container-presence semantics as
	// ListAssignment.PreserveConfiguredPresence for a oneOf envelope.
	PreserveConfiguredPresence bool
	// Collection marks a list whose element is an envelope; LoopVar is then the
	// per-element local.
	Collection bool
	LoopVar    string
	// LoopIndex and ExistingVar serve the same positional preservation purpose
	// as their ListAssignment counterparts for collection-shaped unions.
	LoopIndex   string
	ExistingVar string
	// Variants are the alternatives, ordered by Terraform variant name.
	Variants []OneOfVariantAssignment
}

// OneOfVariantAssignment unwraps one alternative of a OneOfAssignment. Every
// field is a function of the envelope alone, never of the use site, so one
// reusable variant body is computed once and shared: the envelope-model target
// is carried as GoField and composed with the enclosing
// OneOfAssignment.ModelVar at render time.
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

// ResourceView is the render-ready data context for the resource template.
type ResourceView struct {
	TypeName    string
	GoName      string
	Description string

	SDKPackage     string
	APIStruct      string
	APIAccessor    string
	APIConstructor string

	// Create, Read, Update and Delete describe the four lifecycle SDK calls.
	// Update is the zero value when UpdateUnsupported.
	Create CRUDCallView
	Read   CRUDCallView
	Update CRUDCallView
	Delete CRUDCallView

	// Import describes how ImportState recovers the resource's identity: the
	// part count, the expected format string and the id's own position are
	// decided once here rather than in the template.
	Import ImportView
	// UpdateUnsupported means the group resolves no Update role: the generated
	// Update method is a stub that errors rather than building a request the
	// SDK has no endpoint for.
	UpdateUnsupported bool

	Models []ModelStructView
	Schema SchemaView
	State  StateView

	UsesFmt              bool
	UsesValidators       bool
	UsesStringValidators bool
	UsesObjectValidators bool
	UsesPlanModifiers    bool
	PlanModifierPackages []string
	// UsesUUID and UsesStrconv add the google/uuid and strconv imports for a
	// path argument that must be recovered by parsing (see SDKArgumentView).
	// UsesUUID is also true when a request field needs a uuid.Parse (see
	// RequestFieldView.ParseCall).
	UsesUUID    bool
	UsesStrconv bool
	// UsesTime adds the "time" import for a date-time request field's
	// time.Parse call (see RequestFieldView.ParseCall).
	UsesTime bool

	// WriteOnlySecrets is the deterministic resource-wide list used by request
	// routing. UsesWriteOnly controls the fwutils import; schema placement is
	// carried independently by the placeholder AttrViews at each nesting level.
	WriteOnlySecrets []WriteOnlySecretView
	UsesWriteOnly    bool

	// Dropped lists response members skipped from the rendered view (e.g.
	// relationships), surfaced as diagnostics in the run report.
	Dropped []DroppedMember
}

// CRUDCallView describes one lifecycle SDK call. GoRequestType is empty for
// Read and Delete, which send no body; GoResponseType is empty for Delete,
// whose 204 response carries none.
type CRUDCallView struct {
	Method         string
	GoRequestType  string
	GoResponseType string
	// Arguments are the positional SDK call arguments in call order (e.g. the
	// terminal path id, aliased to "id").
	Arguments []SDKArgumentView
	// Envelope describes how this role builds its JSON:API request body. Nil
	// for a call that sends none (Read, Delete).
	Envelope *RequestEnvelopeView
	// WriteOnlySecrets contains only secrets declared by this call's own
	// request body, in deterministic resource-tree order.
	WriteOnlySecrets []WriteOnlySecretView
	// NormalizeUnknowns contains the computed Framework values that may remain
	// unknown after a sparse Create response. It is populated only for Create;
	// Read begins from state and Update preserves its planned values.
	NormalizeUnknowns []UnknownNormalizationView
}

// UnknownNormalizationView describes one recursively-addressable generated
// model value whose computed unknown state must not reach post-Create state.
// Kind is value, object, or list_object. Value nodes carry NullExpr; container
// nodes carry only descendants that can themselves remain unknown.
type UnknownNormalizationView struct {
	Kind     string
	Expr     string
	NullExpr string
	ItemVar  string
	Children []UnknownNormalizationView
}

// ImportView describes how a resource recovers its identity from a terraform
// import id. A top-level resource passes the id straight through; a
// sub-resource's identity is its parent path parameters plus its own id, so the
// import id is composite and has to be split. The part count, the per-segment
// SetAttribute calls and the quoted format string all derive from one place.
type ImportView struct {
	// Parts are the attributes the import id carries, in order: each parent
	// path parameter, then "id". A single element means passthrough.
	Parts []string
	// Format is the id shape a failure message quotes, e.g. "<team_id>:<id>".
	Format string
}

// Composite reports whether the import id carries more than the resource's own
// id, and so has to be split.
func (v ImportView) Composite() bool { return len(v.Parts) > 1 }

// RequestEnvelopeView is the recipe for constructing one role's JSON:API
// request body, level by level, each from its own New<Type>WithDefaults(). The
// data level is built rather than reached through: a request wrapper's
// constructor returns the zero struct, so body.Data.Attributes.Set<F>(...)
// leaves the JSON:API "type" discriminator empty and panics where Data is a
// pointer — only the data component's own constructor assigns the type.
type RequestEnvelopeView struct {
	// SDKPackage qualifies every constructor this envelope renders. It is
	// carried here rather than read off the root view because the partial that
	// renders an envelope is handed one CRUDCallView, not the whole view.
	SDKPackage string
	// Fields are the Set<Field>(...) calls that populate the attributes local,
	// narrowed to what this role's own body declares.
	Fields []RequestFieldView
	// DataVar is the local holding the constructed data member.
	DataVar string
	// DataType is the SDK component behind DataVar, e.g.
	// "IncidentTypeCreateData".
	DataType string
	// TypeExpr is the JSON:API discriminator the body sends, as a Go expression,
	// e.g. `datadogV2.PlaylistDataType("rum_replay_playlist")`. It is emitted
	// whenever the spec determines the value, so a spec that drops a `default`
	// cannot silently post an empty type. Empty when the data component
	// declares no type property, or when the SDK supplies an ambiguous one.
	TypeExpr string
	// IDExpr is the expression assigned to data.id, and IDPrep the parse
	// declaration it depends on — nil when the expression needs no parse, or
	// when it reuses a path argument's local that argPrep already declared.
	// Both are empty unless the body declares an id: a create has no record
	// yet, and the SDK generates SetId only for a declared property.
	IDExpr string
	IDPrep *SDKArgumentView
	// AttributesType is the SDK component of the attributes member, and doubles
	// as the test for whether this body has an attributes level at all: it is
	// set exactly when there are fields to put there. The local it is built
	// into is named by the requestAttributesVar constant, read by the template
	// through a func of the same name so it cannot diverge from Target.
	AttributesType string
}

// RequestFieldView is one field of one role's request body, set via the SDK's
// Set<GoField>(v) setter — one view shape serves a required (non-pointer) and
// an optional (pointer) SDK field alike. Exactly one of ValueExpr, Nested and
// Collection is populated: a leaf sets the parent's field directly, an object
// field builds Nested's own value first, a list/map field Collection's.
type RequestFieldView struct {
	// GoField is the SDK setter suffix, e.g. "Name" for SetName.
	GoField string
	// Target is the expression Set<GoField> is called on: the constructed
	// attributes local ("bodyAttributes") at the top level, or an ancestor's
	// own RequestNestedView.Var one or more levels down. Precomputed because a
	// partial invoked on one nested field cannot see its ancestors' state.
	Target string
	// Required renders the call unconditionally; false guards it behind
	// NullCheck (a leaf) or a nil check on Nested's ModelExpr (an object).
	Required bool
	// NullCheck is the guard expression for a non-Required leaf, e.g.
	// "!state.Description.IsNull() && !state.Description.IsUnknown()". Empty
	// for a Required leaf or for any Nested field (object fields guard on
	// Nested.ModelExpr instead).
	NullCheck string

	// ValueExpr reads the unwrapped Go value a leaf's Set<GoField> takes, e.g.
	// "state.Name.ValueString()" or, after a ParsedVar parse, the parsed
	// local's own name. Empty when Nested is set.
	ValueExpr string
	// ParsedVar and ParseCall request a parse step before ValueExpr can be
	// used, the same shape SDKArgumentView uses for a path argument: declare
	// "<ParsedVar>, err := <ParseCall>", check err, then pass ParsedVar (which
	// equals ValueExpr) to the setter. Empty when the model's own accessor
	// already produces the setter's expected type.
	ParsedVar string
	ParseCall string
	// TFName names the field in a parse-failure diagnostic. Set only when
	// ParsedVar is.
	TFName string
	// WriteOnlyResult names the SecretResult retrieved from configuration before
	// request construction. When set, the original SDK setter receives its Value
	// only when ShouldSetValue is true; no Terraform model field is read.
	WriteOnlyResult string

	// Nested is set for an object field: the request-side SDK type to build
	// via New<SDKType>WithDefaults(), and its own Set<GoField> calls
	// (recursing through this same view). Nil for a leaf field.
	Nested *RequestNestedView

	// Collection is set for a list/map field: either a primitive-terminal
	// collection, decoded via ElementsAs, or a list-of-objects one, built one
	// element at a time (see RequestCollectionView). Nil for a leaf or a
	// single-object field.
	Collection *RequestCollectionView

	// OneOf is set for a oneOf field: the selected variant is expanded into
	// this role's own SDK wrapper before the setter takes it. Nil for every
	// other shape.
	OneOf *RequestOneOfView
}

// RequestOneOfView is the recipe for turning a Terraform oneOf envelope back
// into the go-sdk wrapper one role's request body declares. Mirroring
// OneOfAssignment, it counts the variant blocks configured and refuses anything
// but exactly one, since the SDK's MarshalJSON would silently send the first
// non-nil member. Every SDK identity here is this role's own: the three bodies
// declare three distinct wrappers for one logical union.
type RequestOneOfView struct {
	// TFName is the envelope attribute's Terraform name, the diagnostic's summary.
	TFName string
	// SDKType is the package-qualified wrapper this role's setter takes, e.g.
	// "datadogV2.ElasticCloudIntegrationAccountAuthenticationRequest".
	SDKType string
	// Var is the local holding the constructed wrapper, and MatchVar the
	// selection counter guarding it. Both are declared inside the enclosing
	// RequestFieldView.NullCheck block, so they are scoped to one envelope.
	Var      string
	MatchVar string
	// SelectionMessage is the detail a failed selection reports, with a single
	// %d for the count, e.g. `data.attributes.auth: exactly one of "basic" or
	// "token" must be set, got 2`. It is assembled here and rendered as a
	// quoted Go literal, so no variant name or schema path can escape it.
	SelectionMessage string
	// Variants are the alternatives in envelope order.
	Variants []RequestOneOfVariantView
}

// RequestOneOfVariantView is one alternative's expansion: how to tell it was
// selected, how to build its SDK member, and how to wrap that member in the
// union.
type RequestOneOfVariantView struct {
	// TFName is the variant block's Terraform name.
	TFName string
	// ModelExpr is the envelope model's pointer to this variant's block, e.g.
	// "state.Authentication.BasicAuth". Non-nil selects the variant — the same
	// test a nested object field uses, for the same reason (see
	// RequestNestedView.ModelExpr).
	ModelExpr string
	// ElemVar is the local holding the SDK member before it is wrapped.
	ElemVar string
	// Constructor builds an object alternative's member, e.g.
	// "datadogV2.NewIntegrationAccountBasicAuthRequestWithDefaults()". Empty
	// for a value-wrapped alternative, whose member is the value itself.
	Constructor string
	// Fields populate an object alternative's member, narrowed to what this
	// role's own alternative declares. Empty for a value-wrapped one.
	Fields []RequestFieldView
	// Value is set for a value-wrapped alternative: the SDK member *is* the
	// scalar, so it is converted straight out of the block's single "value"
	// child rather than constructed.
	Value *RequestOneOfValueView
	// WrapCall is the SDK convenience constructor applied to ElemVar, e.g.
	// "datadogV2.IntegrationAccountBasicAuthRequestAsElasticCloudIntegrationAccountAuthenticationRequest(basicAuthValue)".
	WrapCall string
}

// RequestOneOfValueView converts a value-wrapped alternative's single "value"
// child into the scalar the SDK member is. It carries the same
// ParsedVar/ParseCall pair a leaf RequestFieldView does, because a date-time or
// uuid value has to survive a fallible parse before the SDK can take it.
type RequestOneOfValueView struct {
	ValueExpr string
	ParsedVar string
	ParseCall string
	// TFName names the value in a parse-failure diagnostic.
	TFName string
}

// RequestNestedView is one nested object a resource's request body constructs
// via New<SDKType>WithDefaults() before setting it on its parent — the same
// idiom as the request root, applied at every nesting depth.
type RequestNestedView struct {
	// Constructor is the fully package-qualified call building this node's
	// zero value, e.g. "datadogV2.NewSettingsRequestWithDefaults()".
	// Precomputed because the package name lives on ResourceView, out of reach
	// once template recursion has descended past the top-level fields.
	Constructor string
	// Var is the local variable holding the constructed value.
	Var string
	// ModelExpr reads the model's own pointer to this nested value, through the
	// parent path, so a field nested two levels deep reads
	// "state.Settings.Retry" rather than "state.Retry". Fields' own
	// ValueExpr/NullCheck are expressed relative to *ModelExpr, not to Var:
	// Var only ever holds the SDK value being built, never the source model.
	ModelExpr string
	// Fields are this nested object's own Set<GoField>(...) calls.
	Fields []RequestFieldView
}

// RequestCollectionView is a list/map field a request body builds before
// handing it to the parent's Set<GoField>(v) setter. "primitive" holds a raw
// types.List/types.Map the setter cannot take, so it is decoded into a native
// slice/map via ElementsAs, which reports diagnostics. "object" is already
// decoded into []*<ElemModel> by the framework, so each element becomes its
// own request value, appended one at a time.
type RequestCollectionView struct {
	// Kind is "primitive" or "object", selecting which of the two strategies
	// above the requestField template renders.
	Kind string

	// ConvertVar is the local variable that ends up holding the value passed
	// to Set<GoField>: the ElementsAs target for a primitive collection, or
	// the accumulator slice appended to on each loop iteration for an object
	// one.
	ConvertVar string
	// ConvertType declares ConvertVar's Go type. For Kind == "primitive" it
	// is the ElementsAs target type, e.g. "[]string" or "map[string]int64".
	// Empty for Kind == "object", which instead declares "[]" + ElementGoType.
	ConvertType string
	// ConvertCall is the ElementsAs call itself for Kind == "primitive", e.g.
	// "state.Tags.ElementsAs(ctx, &tagsElements, false)". Empty for Kind ==
	// "object".
	ConvertCall string

	// RangeExpr is the already-decoded state slice a Kind == "object"
	// collection ranges over, e.g. "state.Tags". Empty for Kind ==
	// "primitive".
	RangeExpr string
	// LoopVar is the per-element loop variable (a *<ElemModel> pointer).
	// Empty for Kind == "primitive".
	LoopVar string
	// ElemVar is the local holding one constructed request element inside
	// the loop. Empty for Kind == "primitive".
	ElemVar string
	// Constructor builds ElemVar's zero value, e.g.
	// "datadogV2.NewTagItemWithDefaults()". Empty for Kind == "primitive".
	Constructor string
	// ElementGoType is the constructed element's SDK type, package-qualified,
	// used to declare ConvertVar's slice ("[]" + ElementGoType). Empty for
	// Kind == "primitive".
	ElementGoType string
	// Fields are ElemVar's own Set<GoField>(...) calls, recursing through
	// RequestFieldView the same way RequestNestedView.Fields does. Empty for
	// Kind == "primitive".
	Fields []RequestFieldView
}
