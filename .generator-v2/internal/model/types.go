// Package model defines the generator's internal data model: the in-memory
// types that flow from the parser, through schema conversion, into the
// emitter and the run report. These types are deliberately decoupled from
// both the OpenAPI input format and the Terraform Plugin Framework output
//
// The single exception is Spec.Components, which retains a handle to the
// libopenapi component set so that schemas can be lazily resolved without
// re-parsing the spec.
package model

import (
	"time"

	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
)

// ----------------------------------------------------------------------------
// Enumerations
//
// Internal-only enums (SchemaKind,GenerationStage) use stable lowercase
// tokens for debuggability.
//
// ----------------------------------------------------------------------------

// ArtifactKind distinguishes a read-only data source from a full-CRUD resource.
type ArtifactKind string

const (
	ArtifactKindResource   ArtifactKind = "resource"
	ArtifactKindDataSource ArtifactKind = "data_source"
)

// SchemaKind classifies a normalized Schema node by structure. Primitive, JSON,
// Object, Array and Map are directly emittable as Terraform attributes. JSON is
// an unconstrained OpenAPI value represented as normalized JSON text. OneOf
// requires a synthetic envelope described by Schema.OneOf. RefCycle,
// DepthExceeded and Unsupported are fatal.
//
// RefCycle and DepthExceeded are deliberately distinct: a cycle is a property of
// the document that no flag can fix, whereas exhausting --max-depth says only
// that the walk stopped early, and raising the limit may resolve the node.
// Reporting the latter as a cycle sends the reader hunting for a cycle that does
// not exist.
type SchemaKind string

const (
	SchemaKindPrimitive     SchemaKind = "primitive"
	SchemaKindJSON          SchemaKind = "json"
	SchemaKindObject        SchemaKind = "object"
	SchemaKindArray         SchemaKind = "array"
	SchemaKindMap           SchemaKind = "map"
	SchemaKindOneOf         SchemaKind = "one_of"
	SchemaKindRefCycle      SchemaKind = "ref_cycle"      // a $ref that re-enters a schema already being expanded
	SchemaKindDepthExceeded SchemaKind = "depth_exceeded" // $ref expansion stopped at --max-depth; not a cycle
	SchemaKindUnsupported   SchemaKind = "unsupported"    // no representable type/structure, or anyOf; always rejected

	// SchemaKindVariant is retained as a source-compatibility alias for code
	// written before the parser populated Schema.OneOf directly. New code should
	// use SchemaKindOneOf.
	SchemaKindVariant = SchemaKindOneOf
)

// Cardinality distinguishes a singular data source (resolves one item by id)
// from a plural one (returns a filtered list). It is the decoded form of the
// tracking extension's "cardinality" field; absent/empty means singular.
type Cardinality string

const (
	CardinalitySingular Cardinality = "singular"
	CardinalityPlural   Cardinality = "plural"
)

// IdStrategy describes how the Terraform resource ID is derived from the API response.
type IdStrategy string

const (
	IdStrategyDataID            IdStrategy = "data.id"
	IdStrategyDataAttributesID  IdStrategy = "data.attributes.id"
	IdStrategyDataAttributesUID IdStrategy = "data.attributes.uuid"
	IdStrategyHeaderLocation    IdStrategy = "header.location"
)

// ----------------------------------------------------------------------------
// Parser-facing types
// ----------------------------------------------------------------------------

// Spec is the root container, loaded once per tfgen run.
type Spec struct {
	// Source is the filesystem path to the OpenAPI YAML/JSON
	Source string
	// Operations holds every operation, regardless of tracking-field state,
	// sorted by (path, method) for deterministic iteration.
	Operations []*Operation
	// Components is the shared component set, retained for lazy $ref resolution.
	Components *v3.Components
	// Hash is the lowercase hex SHA-256 of the spec source
	Hash string
}

// Operation is a single OpenAPI operation, tagged with whether it is in scope
// for generation.
type Operation struct {
	// Path is the OpenAPI path template, e.g. /api/v2/users/{user_id}.
	Path string
	// Method is the HTTP method (GET/POST/PUT/PATCH/DELETE).
	Method string
	// OperationId is the OpenAPI operationId.
	OperationId string
	// Tag is the OpenAPI tag, driving SDK package selection.
	Tag string
	// Tracking is the decoded tracking-field extension
	Tracking *TrackingFieldMetadata
	// RequestSchema is the resolved request body schema, if any.
	RequestSchema *Schema
	// RequestRefName is the last path segment of the request body $ref, e.g.
	// "TeamCreateRequest" — the SDK Go request type, and the root the SDK oneOf
	// binding pass walks from on the request side. Empty when the body is inline
	// or absent.
	RequestRefName string
	// ResponseSchema is the resolved 2xx response schema, if any.
	ResponseSchema *Schema
	// ResponseRefName is the last path segment of the 2xx response body $ref,
	// e.g. "IncidentTypeResponse" — the SDK Go response type; empty when the
	// body is inline or absent.
	ResponseRefName string
	// QueryParams are the operation's in:query parameters, normalized and sorted
	// by name. Populated for every operation; the plural data-source path turns
	// the scalar ones into filters. DeclarationOrder retains the original OpenAPI
	// position so SDK binding can reproduce the client generator's call order.
	QueryParams []QueryParam
	// PathParams are the operation's in:path parameters, normalized and sorted
	// by name. DeclarationOrder retains their position relative to required query
	// parameters, matching the Go client generator's parameter walk.
	PathParams []QueryParam
	// Pagination is the decoded x-pagination extension, or nil when the
	// operation declares none.
	Pagination *Pagination
	// ItemRefName is the last $ref segment of the results-array element schema
	// for a list response, e.g. "Team" — the SDK Go element type. Empty when the
	// resultsPath property is absent or is not an array.
	ItemRefName string
	// ResponseDataRefName is the last $ref segment of a by-id response's "data"
	// property when it is a single object reference, e.g. "FullAPIKey" — the SDK
	// Go record type. Empty for list responses (whose "data" is an array; see
	// ItemRefName) or an inline data object. Lets a "both" data source detect when
	// its by-id record shape diverges from its list element shape.
	ResponseDataRefName string
	// SearchOp is the operation named by Tracking.Group.Search, resolved during
	// NormalizeSchemas: the list endpoint a singular data source searches to
	// resolve one record. It points at the operation itself when this op is the
	// search op (search-only), and is nil when no search is declared or the
	// declared operationId is unknown.
	SearchOp *Operation
	// SDKBinding is the call signature derived from the OpenAPI operation using
	// the Go SDK generator's naming and ordering rules. The CLI fills it before
	// artifact construction. Tests that build parser-shaped operations directly
	// may leave it nil and exercise the legacy call shape.
	SDKBinding *SDKOperationBinding
}

// QueryParam is one normalized OpenAPI path or query parameter (the historical
// name is retained to avoid broad churn). Its inner schema is normalized like a
// request/response body, and Name preserves raw spelling such as
// "filter[keyword]".
type QueryParam struct {
	Name        string
	Required    bool
	Schema      *Schema
	Description string
	// DeclarationOrder is the one-based position in operation.parameters. Zero
	// is reserved for hand-built test fixtures that do not carry source order.
	DeclarationOrder int
}

// Pagination is the decoded x-pagination extension on a list operation. It
// names the limit/page query parameters and the response property holding the
// result array.
type Pagination struct {
	// LimitParam is the page-size query parameter name, e.g. "page[size]".
	LimitParam string
	// PageParam is the page-cursor/number query parameter name, e.g. "page[number]".
	PageParam string
	// ResultsPath is the response property holding the result array, e.g. "data".
	ResultsPath string
}

// SDKOperationBinding is the Go SDK signature derived for one OpenAPI operation.
// Required arguments are positional and retain the SDK generator's order.
// Optional arguments are fields set through With* methods on OptionalParamsType.
type SDKOperationBinding struct {
	Required           []SDKArgument
	Optional           []SDKArgument
	OptionalParamsType string
}

// SDKArgument binds one SDK method argument or options setter to an OpenAPI
// parameter and, after artifact construction, its Terraform model field.
type SDKArgument struct {
	Name        string
	GoName      string
	GoType      string
	Location    string
	Description string
	Schema      *Schema
	TFName      string
	Setter      string
}

// Schema is a normalized, recursive view of an OpenAPI schema after allOf
// flattening, oneOf-envelope detection, and explicit anyOf rejection.
type Schema struct {
	Kind SchemaKind
	// Properties is populated for objects only; iteration is always sorted.
	Properties map[string]*Schema
	// Required is populated for objects only; sorted.
	Required []string
	// Items is populated for arrays only.
	Items *Schema
	// OneOf is populated when Kind is SchemaKindOneOf. It carries the stable
	// Terraform envelope identity, its non-null alternatives, and the metadata
	// required to bind those alternatives to the generated SDK wrapper.
	OneOf *OneOfSpec
	// Variants is the parser's legacy oneOf representation. It remains only as a
	// source-compatibility bridge; NormalizeSchemas leaves it empty and new model
	// and emit code must consume OneOf instead.
	//
	// Deprecated: use OneOf.Variants.
	Variants []*Schema
	// Type is the primitive type (string/integer/number/boolean).
	Type string
	// Format is the optional OpenAPI format (date-time, int64, ...).
	Format string
	// Enum holds the allowed values, if constrained.
	Enum []string
	// Sensitive is true when the schema is annotated sensitive: true.
	Sensitive bool
	// Description is the OpenAPI description, populated during NormalizeSchemas.
	Description string
	// UnsupportedReason explains why a node with Kind == SchemaKindUnsupported
	// cannot be represented. It is retained so the affected artifact can fail
	// with the parser's actionable local diagnostic without aborting spec loading
	// or preventing unrelated artifacts from being generated.
	UnsupportedReason string
	// RefName is the OpenAPI component name that supplied this node, e.g.
	// "ActionConnectionAttributes"; empty for an inline schema. It is the identity
	// the Datadog go-sdk names its generated model after, so the SDK binding pass
	// walks the normalized tree and restarts its name accumulation at every node
	// carrying one — mirroring how the SDK generator's child_models() prefers a
	// $ref name over the parent-derived alternative name.
	RefName string
}

// OneOfSpec is the normalized representation of an OpenAPI oneOf. The envelope
// exists only in generated Terraform/Go code; request and response mappers
// unwrap/wrap it when interacting with the Datadog go-sdk.
type OneOfSpec struct {
	// Name is the deterministic generated envelope type name. Reusable
	// component unions use their component name; inline unions use a
	// schema-path-derived name.
	Name string
	// Path is the canonical request/response schema path used for diagnostics
	// and as an input to inline envelope naming.
	Path string
	// RefName is the OpenAPI component name of the union node itself, empty for an
	// inline union. It is deliberately separate from Name: Name is a Terraform
	// identity that falls back to a path-derived spelling, which must never be
	// mistaken for an SDK type.
	RefName string
	// SDKType is the Datadog go-sdk oneOf wrapper struct for this union, e.g.
	// "ActionConnectionIntegration". It is resolved by the SDK binding pass
	// (internal/sdkbind) after normalization and stays empty until then.
	SDKType string
	// Optional permits the whole envelope to be absent because the containing
	// OpenAPI field is not required.
	Optional bool
	// Nullable permits OpenAPI null. Null is represented by an absent envelope,
	// never by a synthetic null variant.
	Nullable bool
	// Discriminator retains optional OpenAPI discriminator metadata for stable
	// naming and diagnostics. It is not required for branch selection.
	Discriminator *OneOfDiscriminator
	// Variants contains only non-null alternatives, sorted by TFName. Parser
	// source order and map iteration order must not affect this slice.
	Variants []OneOfVariant
}

// OneOfDiscriminator retains the OpenAPI discriminator metadata relevant to a
// normalized union. Mapping keys may participate in stable variant naming;
// consumers must sort keys before iterating over Mapping.
type OneOfDiscriminator struct {
	PropertyName string
	Mapping      map[string]string
}

// OneOfVariant is one non-null oneOf alternative and its Terraform/SDK binding.
type OneOfVariant struct {
	// TFName is the stable snake_case nested-block name.
	TFName string
	// GoName is the generated Go model/field stem corresponding to TFName.
	GoName string
	// Schema is the fully normalized alternative, including constraints common
	// to the parent oneOf. It may recursively contain another oneOf.
	Schema *Schema
	// RefName is the referenced OpenAPI component name, when present.
	RefName string
	// SDKField is the Datadog go-sdk wrapper member whose presence selects this
	// alternative.
	SDKField string
	// SDKConstructor is the generated SDK convenience constructor for this
	// alternative, when the SDK exposes one.
	SDKConstructor string
	// SDKPointer is true when the wrapper member and its convenience constructor
	// take a pointer, which is every alternative except a free-form object: the SDK
	// emits that one as a bare map, already nil-able. Mappers must not take the
	// address of a member the SDK left unpointered.
	SDKPointer bool
	// ValueWrapped is true for primitive, list, and map alternatives, whose
	// Terraform variant model exposes a single field named value. Object
	// alternatives expose their generated fields directly.
	ValueWrapped bool
}

// ----------------------------------------------------------------------------
// Model / emit types
// ----------------------------------------------------------------------------

// Artifact is the internal projection of a flagged Operation, ready for
// emission. There is one Artifact per (Kind, Name) pair.
type Artifact struct {
	// Name is the Terraform-facing artifact name (without the datadog_ prefix).
	Name string
	Kind ArtifactKind
	// Cardinality selects the singular vs plural data-source shape; the emit
	// builder routes on it. Empty for resources.
	Cardinality Cardinality
	// Description is the artifact's top-level schema doc string, from the
	// tracking extension's tf_description field; empty when the author omits it.
	Description string
	// Schema is the Terraform schema derived from the response (and request,
	// for resources).
	Schema *AttributeTree
	// Lifecycle holds the SDK call bindings. For data sources only Read is set
	Lifecycle *LifecycleBindings
	// SourceFile is the output path, e.g. datadog/fwprovider/<file>.go.
	SourceFile string
	// Diagnostics carries non-fatal notes raised while building the artifact,
	// e.g. query parameters dropped from a plural data source's filter set. The
	// artifact still emits; the run report surfaces these as info.
	Diagnostics []Diagnostic
}

// AttributeTree is the root of the Terraform schema tree for one artifact.
type AttributeTree struct {
	Attributes []*Attribute
}

// Attribute mirrors a Terraform Plugin Framework attribute or nested block
// one-to-one. The emitter walks this tree to produce the Schema() method body.
type Attribute struct {
	// Path is the dot-delimited attribute path, e.g. spec.replicas. It doubles
	// as the per-attribute hook ID anchor.
	Path string
	// TfType is the framework type, e.g. schema.StringAttribute.
	TfType string
	// GoType is the corresponding model-struct type, e.g. types.String.
	GoType string
	// CustomType is the optional framework custom type expression rendered on
	// the schema attribute, e.g. jsontypes.NormalizedType{} for arbitrary JSON.
	CustomType string
	// ElementType is the framework attr.Type for a list/map element value,
	// e.g. "types.StringType" or "types.ListType{ElemType: types.StringType}".
	// Set only for ListAttribute/MapAttribute collection chains ending in a
	// primitive; empty for everything else.
	ElementType string
	// Format is the OpenAPI format (e.g. "date-time"). It distinguishes SDK
	// getters whose Go return type differs from the bare scalar: a date-time
	// string getter returns time.Time, not string.
	Format string
	// IsEnum marks a string whose SDK getter returns a named enum type rather
	// than a bare string, so the state mapper must cast it back with string(...).
	IsEnum bool

	Required  bool
	Optional  bool
	Computed  bool
	Sensitive bool

	// Default is the optional default value, encoded as a Go expression.
	Default *Literal
	// Validators is the fingerprintable validator list for this attribute.
	Validators []ValidatorSpec
	// Description is always populated from the OpenAPI description (repo convention).
	Description string
	// Children holds nested attributes for nested blocks.
	Children []*Attribute
	// ModelRefName is the OpenAPI component name that supplied the *object* schema
	// backing this attribute's generated model struct: the node's own schema for an
	// object, its element schema for an array or map of objects. Empty when that
	// schema was inline, and empty for a leaf, which has no struct.
	//
	// It exists so emit can name a nested model after its component rather than
	// after the property that happens to point at it — the same preference the SDK
	// generator's child_models() applies (get_name(schema) or alternative_name) and
	// that oneOf envelope naming already applies via OneOfSpec.Name. Without it, two
	// differently-shaped objects reachable under the same property name produce one
	// struct name and the artifact cannot compile.
	ModelRefName string
	// OneOf is non-nil when this attribute carries a synthetic oneOf envelope:
	// either the envelope itself (a union at the root or an object property) or
	// the collection whose element is a union. Children then holds the variant
	// blocks, and OneOf holds the naming and SDK-binding metadata the emit layer
	// needs to map them.
	OneOf *OneOfEnvelope
}

// OneOfEnvelope is the Terraform projection of a parser-normalized OneOfSpec: the
// synthetic block holding one nested variant block per non-null alternative,
// exactly one of which is selected whenever the envelope is present.
//
// It hangs off the Attribute standing at the union's position in the tree; that
// attribute's Children are the projected variant blocks, in the same order as
// Variants. Terraform concerns live here rather than on OneOfSpec so that
// OpenAPI normalization stays free of them.
type OneOfEnvelope struct {
	// Name is the parser-assigned envelope identity (OneOfSpec.Name): the OpenAPI
	// component name for a reusable union, a deterministic path-derived name for
	// an inline one. Two uses of the same component share a Name, which is what
	// lets emit generate one model per envelope instead of one per use site.
	Name string
	// GoModel is the generated Go struct holding one pointer field per variant.
	GoModel string
	// SDKType is the Datadog go-sdk oneOf wrapper struct this envelope maps to,
	// carried through from OneOfSpec. It is a separate identity from Name and
	// GoModel: an inline union's envelope name is path-derived and names no SDK
	// struct. Empty until the SDK binding pass has run.
	SDKType string
	// Path is the union's own schema path, used by validators and diagnostics. For
	// a collection of unions it is the element path (e.g. "response.choices[]"),
	// which is not the path of any attribute in the tree.
	Path string
	// Optional permits the whole envelope to be absent because its containing
	// OpenAPI field is optional or nullable. When false, exactly one variant must
	// be selected.
	Optional bool
	// Computed marks a response-only envelope, where selection is enforced by
	// response mapping rather than by practitioner configuration.
	Computed bool
	// Variants are the projected non-null alternatives, ordered by TFName so
	// neither OpenAPI alternative order nor map iteration can reach the output.
	Variants []OneOfEnvelopeVariant
}

// OneOfEnvelopeVariant is one projected alternative of a OneOfEnvelope.
type OneOfEnvelopeVariant struct {
	// TFName is the stable snake_case nested-block name.
	TFName string
	// GoField is the pointer field naming this variant on the envelope's model.
	// The pointer being non-nil is what selects the variant.
	GoField string
	// GoModel is the generated Go struct for this variant's own fields.
	GoModel string
	// SDKField is the Datadog go-sdk wrapper member whose presence selects this
	// alternative, and SDKConstructor the SDK convenience constructor for it. Both
	// are carried through from the parser's OneOfVariant and stay empty until the
	// SDK binding pass resolves them: the projection never derives an SDK identity
	// from a Terraform name, since the two conventions differ (a variant named
	// aws_integration binds to the SDK's AWSIntegration, not AwsIntegration).
	SDKField       string
	SDKConstructor string
	// SDKPointer is true when the SDK wrapper member and its convenience
	// constructor take a pointer — every alternative except a free-form object,
	// which the SDK emits as a bare, already-nil-able map.
	SDKPointer bool
	// ValueWrapped is true for every non-object alternative — scalar, list, map, or
	// a directly nested union — whose block holds a single child named "value".
	// Object alternatives expose their own fields directly instead.
	ValueWrapped bool
	// Attribute is this variant's projected block: the same pointer as the
	// envelope-carrying attribute's Children entry at this index. Children drives
	// schema rendering; this field lets the mapper reach a block without
	// re-deriving the ordering.
	Attribute *Attribute
}

// Literal is a default value rendered as a Go source expression
// (e.g. `true`, `"foo"`, `int64(3)`).
type Literal struct {
	GoExpr string
}

// ValidatorSpec is a deterministic, fingerprintable description of a framework
// validator: the constructor plus its Go-source-rendered arguments.
type ValidatorSpec struct {
	// Name is the validator constructor, e.g. stringvalidator.LengthAtLeast.
	Name string
	// Args are the constructor arguments rendered as Go source expressions.
	Args []string
}

// LifecycleBindings maps Terraform lifecycle methods to their SDK calls. For a
// singular data source: Read is the by-id call and Search the list call —
// read-only sets Read, search-only sets Search, the id-optional shape sets both.
// IdStrategy and Create/Update/Delete are zero for data sources.
type LifecycleBindings struct {
	Create *SDKCall
	Read   *SDKCall
	// Search is the list call a singular data source uses to resolve one record
	// by filter. It carries the list-call fields (ItemType/OptionalParamsType/
	// Paginated), same as a plural Read.
	Search     *SDKCall
	Update     *SDKCall
	Delete     *SDKCall
	IdStrategy IdStrategy
}

// SDKCall represents a single datadog-api-client-go invocation.
type SDKCall struct {
	// BindingResolved distinguishes an SDK method that genuinely takes no
	// positional arguments from legacy test fixtures that omit binding data.
	BindingResolved bool
	// GoPackage is the versioned SDK package, e.g. "datadogV2".
	// Rule: "datadog" + strings.ToUpper(version), where version is the path
	// segment after /api/ in Operation.Path (e.g. /api/v2/... → "datadogV2").
	GoPackage string
	// GoApiStruct is the API client struct name, e.g. "OrgGroupsApi".
	// Rule: tag_to_class_name(Operation.Tag): replaces every non-alphanumeric
	// character with a space, capitalizes each word and joins, then appends
	// "Api". Preserves original casing within each word (so "APM" → "APMApi",
	// not "ApmApi").
	GoApiStruct string
	// GoMethod is the method name on GoApiStruct, e.g. "CreateOrgGroup".
	// Rule: Operation.OperationId, no transformation applied.
	GoMethod string
	// GoRequestType is the SDK request body type, e.g. "OrgGroupCreateRequest".
	// Rule: last path component of the requestBody $ref
	// (e.g. "#/components/schemas/OrgGroupCreateRequest" → "OrgGroupCreateRequest").
	// Empty when the operation takes no request body (e.g. DELETE, GET-by-ID).
	// NOTE: Schema has no Name field; the model-builder must read this from the
	// raw libopenapi node, not from Operation.RequestSchema.
	GoRequestType string
	// GoResponseType is the SDK response type, e.g. "OrgGroupResponse".
	// Rule: last path component of the 2xx response schema $ref
	// (e.g. "#/components/schemas/OrgGroupResponse" → "OrgGroupResponse").
	// Empty when the operation returns no body (e.g. 204 No Content).
	// NOTE: Schema has no Name field; the model-builder must read this from the
	// raw libopenapi node, not from Operation.ResponseSchema.
	GoResponseType string
	// Arguments are the required positional SDK arguments in call order.
	Arguments []SDKArgument
	// OptionalArguments bind Terraform filters to OptionalParamsType setters.
	OptionalArguments []SDKArgument

	// The fields below back a plural data-source list call.

	// ItemType is the SDK element type yielded by the list call, e.g. "Team"
	// (from Operation.ItemRefName). The non-paginated read collects resp.Data
	// into []<ItemType>; the paginated read yields PaginationResult[<ItemType>].
	ItemType string
	// OptionalParamsType is the SDK optional-parameters struct, e.g.
	// "ListTeamsOptionalParameters" (<GoMethod>OptionalParameters). Empty when
	// the endpoint declares no query parameters, in which case the list call
	// takes no optional-parameters argument.
	OptionalParamsType string
	// Paginated selects the "<GoMethod>WithPagination" iterator form, set when
	// the operation declares an x-pagination extension.
	Paginated bool
}

// ----------------------------------------------------------------------------
// Run-report types
//
// Field names and JSON tags mirror contracts/run-report.schema.json so
// report.WriteJSON can marshal a RunReport straight to the structured output
// CI gates on.
// ----------------------------------------------------------------------------

// ArtifactStatus is the terminal state of an artifact in a generate run.
type ArtifactStatus string

const (
	ArtifactStatusCreated   ArtifactStatus = "created"
	ArtifactStatusUpdated   ArtifactStatus = "updated"
	ArtifactStatusUnchanged ArtifactStatus = "unchanged"
	ArtifactStatusSkipped   ArtifactStatus = "skipped"
	ArtifactStatusFailed    ArtifactStatus = "failed"
	// ArtifactStatusRetired marks an artifact whose files and registration were
	// deleted because its annotation is gone and no recorded cassette adopted it.
	ArtifactStatusRetired ArtifactStatus = "retired"
	// ArtifactStatusRetireBlocked marks an orphaned artifact left in place because
	// a recorded cassette (or a missing generated marker) makes deletion unsafe.
	ArtifactStatusRetireBlocked ArtifactStatus = "retire_blocked"
	// ArtifactStatusRegistrationRetired marks a stale registration dropped on its
	// own: the constructor was still listed in datasources_generated.go but its
	// generated files were already gone, so only the registration line changed.
	// The entry's Constructor carries the removed identifier, since no file
	// remains to recover the artifact name from.
	ArtifactStatusRegistrationRetired ArtifactStatus = "registration_retired"
)

// DiagnosticSeverity classifies a Diagnostic.
type DiagnosticSeverity string

const (
	SeverityError   DiagnosticSeverity = "error"
	SeverityWarning DiagnosticSeverity = "warning"
	SeverityInfo    DiagnosticSeverity = "info"
)

// SkipReason explains why an operation produced no artifact.
type SkipReason string

const (
	SkipReasonTrackingFieldAbsent SkipReason = "tracking_field_absent"
	SkipReasonTrackingFieldSkip   SkipReason = "tracking_field_skip_true"
)

// RunReport is the structured output of a tfgen generate run.
type RunReport struct {
	RunId             string                `json:"run_id"`
	GeneratorVersion  string                `json:"generator_version"`
	SpecHash          string                `json:"spec_hash"`
	StartedAt         time.Time             `json:"started_at"`
	FinishedAt        time.Time             `json:"finished_at"`
	Artifacts         []ArtifactReportEntry `json:"artifacts"`
	SkippedOperations []SkippedOperation    `json:"skipped_operations,omitempty"`
	Summary           *RunSummary           `json:"summary,omitempty"`
}

// RunSummary holds convenience counts for CI assertions, one per ArtifactStatus.
type RunSummary struct {
	Created             int `json:"created"`
	Updated             int `json:"updated"`
	Unchanged           int `json:"unchanged"`
	Skipped             int `json:"skipped"`
	Failed              int `json:"failed"`
	Retired             int `json:"retired"`
	RetireBlocked       int `json:"retire_blocked"`
	RegistrationRetired int `json:"registration_retired"`
}

// ArtifactReportEntry is the per-artifact section of a RunReport.
type ArtifactReportEntry struct {
	Name        string         `json:"name"`
	Kind        ArtifactKind   `json:"kind"`
	Status      ArtifactStatus `json:"status"`
	Path        string         `json:"path"`
	Diagnostics []Diagnostic   `json:"diagnostics,omitempty"`
	// OrphanedHooks lists hook functions declared but no longer referenced.
	OrphanedHooks []string `json:"orphaned_hooks,omitempty"`
	// Constructor is set only on registration_retired entries: the removed
	// registration identifier. It is authoritative because the artifact name
	// cannot be recovered once the generated files are gone.
	Constructor string `json:"constructor,omitempty"`
}

// Diagnostic is a single error/warning/info collected during generation.
type Diagnostic struct {
	Severity DiagnosticSeverity `json:"severity"`
	Message  string             `json:"message"`
	// Location is an optional source-side anchor,
	// e.g. spec:components.schemas.Pet.properties.tags.
	Location string `json:"location,omitempty"`
}

// SkippedOperation records an operation that produced no artifact, listed for
// visibility rather than as a failure.
type SkippedOperation struct {
	OperationId string     `json:"operation_id"`
	Path        string     `json:"path"`
	Method      string     `json:"method"`
	Reason      SkipReason `json:"reason"`
}
