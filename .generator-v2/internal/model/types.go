// Package model defines the generator's internal data model: the types that
// flow from the parser, through schema conversion, into the emitter and the run
// report. They are decoupled from both the OpenAPI input and the Terraform
// Plugin Framework output, the one exception being Spec.Components, which keeps
// a libopenapi handle so schemas can be resolved lazily.
package model

import (
	"slices"
	"time"

	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
)

// ----------------------------------------------------------------------------
// Enumerations
//
// Internal-only enums use stable lowercase string tokens.
// ----------------------------------------------------------------------------

// ArtifactKind distinguishes a read-only data source from a full-CRUD resource.
type ArtifactKind string

const (
	ArtifactKindResource   ArtifactKind = "resource"
	ArtifactKindDataSource ArtifactKind = "data_source"
)

// SchemaKind classifies a normalized Schema node by structure. Primitive,
// Object, Array and Map are directly emittable as Terraform attributes; OneOf
// needs the synthetic envelope in Schema.OneOf; RefCycle, DepthExceeded and
// Unsupported are fatal. The three failures stay distinct because only
// DepthExceeded may resolve when --max-depth is raised.
type SchemaKind string

const (
	SchemaKindPrimitive     SchemaKind = "primitive"
	SchemaKindObject        SchemaKind = "object"
	SchemaKindArray         SchemaKind = "array"
	SchemaKindMap           SchemaKind = "map"
	SchemaKindOneOf         SchemaKind = "one_of"
	SchemaKindRefCycle      SchemaKind = "ref_cycle"      // a $ref that re-enters a schema already being expanded
	SchemaKindDepthExceeded SchemaKind = "depth_exceeded" // $ref expansion stopped at --max-depth; not a cycle
	SchemaKindUnsupported   SchemaKind = "unsupported"    // no representable type/structure, or anyOf; always rejected

	// SchemaKindVariant is a source-compatibility alias for SchemaKindOneOf,
	// which new code should use instead.
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

// IdStrategy describes how the Terraform resource ID is derived from the API
// response.
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
	// Unstable records that the operation declares x-unstable. Only presence is
	// kept; the extension's value is just a human-readable beta notice.
	Unstable bool
	// RequestSchema is the resolved request body schema, if any.
	RequestSchema *Schema
	// RequestRefName is the last path segment of the request body $ref, e.g.
	// "TeamCreateRequest" — the SDK Go request type. Empty when the body is
	// inline or absent.
	RequestRefName string
	// ResponseSchema is the resolved 2xx response schema, if any.
	ResponseSchema *Schema
	// ResponseRefName is the last path segment of the 2xx response body $ref,
	// e.g. "IncidentTypeResponse" — the SDK Go response type; empty when the
	// body is inline or absent.
	ResponseRefName string
	// QueryParams are the operation's in:query parameters, normalized and sorted
	// by name. DeclarationOrder retains each one's original OpenAPI position.
	QueryParams []QueryParam
	// PathParams are the operation's in:path parameters, normalized and sorted
	// by name. DeclarationOrder retains their position relative to required
	// query parameters.
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
	// Go record type. Empty for a list response (whose "data" is an array; see
	// ItemRefName) or an inline data object.
	ResponseDataRefName string
	// ResolvedGroup is Tracking.Group with every declared operationId replaced
	// by the operation it names. Nil when the operation declares no group.
	ResolvedGroup *ResolvedGroup
	// SDKBinding is the call signature derived from the operation using the Go
	// SDK generator's naming and ordering rules. Nil until it is resolved.
	SDKBinding *SDKOperationBinding
}

// GroupRole names one operation slot in a tracking group. The values are the
// annotation keys themselves, so a diagnostic can quote what the author wrote.
type GroupRole string

const (
	GroupRoleCreate GroupRole = "create"
	GroupRoleRead   GroupRole = "read"
	GroupRoleSearch GroupRole = "search"
	GroupRoleUpdate GroupRole = "update"
	GroupRoleDelete GroupRole = "delete"
)

// ResolvedGroup is an OperationGroup with every declared operationId replaced by
// the *Operation it names. It is resolved only once the whole spec is
// enumerated, since a group may reference an operation appearing later. A role
// is nil both when the annotation omitted it and when the operationId it named
// matched nothing; Unresolved lists only the latter. Resolution never fails the
// run, so an unresolvable reference travels with the operation and fails only
// its own artifact.
type ResolvedGroup struct {
	Create *Operation
	Read   *Operation
	// Search is the list endpoint a singular data source resolves one record
	// through. It is the annotated operation itself for a search-only artifact.
	Search *Operation
	Update *Operation
	Delete *Operation
	// Unresolved lists the declared references whose operationId matches no
	// operation in the spec, in create/read/search/update/delete order.
	Unresolved []GroupReference
}

// GroupReference is one (role, operationId) pair exactly as the annotation
// declared it, retained so a diagnostic can name both.
type GroupReference struct {
	Role        GroupRole
	OperationId string
}

// Op returns the operation resolved for role, and nil when the annotation
// omitted that role, when the operationId it named matched nothing, or when the
// group was never resolved. Nil-safe on the receiver, so a caller holding a
// possibly-groupless operation needs no guard of its own.
func (g *ResolvedGroup) Op(role GroupRole) *Operation {
	if g == nil {
		return nil
	}
	switch role {
	case GroupRoleCreate:
		return g.Create
	case GroupRoleRead:
		return g.Read
	case GroupRoleSearch:
		return g.Search
	case GroupRoleUpdate:
		return g.Update
	case GroupRoleDelete:
		return g.Delete
	}
	return nil
}

// UnresolvedId returns the operationId the annotation declared for role when
// that reference resolved to nothing, and "" when the role was never declared
// at all. Nil-safe on the receiver, like Op.
func (g *ResolvedGroup) UnresolvedId(role GroupRole) string {
	if g == nil {
		return ""
	}
	for _, ref := range g.Unresolved {
		if ref.Role == role {
			return ref.OperationId
		}
	}
	return ""
}

// Operations returns every distinct operation the group resolved to for the
// named roles, always in create/read/search/update/delete order whatever order
// the roles are given in; passing none means all five. An operation filling two
// roles — read and search naming the same endpoint, say — appears once.
// Nil-safe, and document-independent. Filtering exists because a resource
// consumes only the CRUD quad; Search backs a data-source lookup instead.
func (g *ResolvedGroup) Operations(roles ...GroupRole) []*Operation {
	if g == nil {
		return nil
	}
	want := func(GroupRole) bool { return true }
	if len(roles) > 0 {
		want = func(r GroupRole) bool { return slices.Contains(roles, r) }
	}
	ops := make([]*Operation, 0, 5)
	for _, roled := range []struct {
		role GroupRole
		op   *Operation
	}{
		{GroupRoleCreate, g.Create}, {GroupRoleRead, g.Read}, {GroupRoleSearch, g.Search},
		{GroupRoleUpdate, g.Update}, {GroupRoleDelete, g.Delete},
	} {
		if roled.op == nil || !want(roled.role) || slices.Contains(ops, roled.op) {
			continue
		}
		ops = append(ops, roled.op)
	}
	return ops
}

// QueryParam is one normalized OpenAPI path or query parameter. Its inner
// schema is normalized like a request/response body, and Name preserves the
// raw spelling, e.g. "filter[keyword]".
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

// RequestDiscriminator describes a request body's JSON:API "data.type" member,
// which the API requires and rejects the body without. The pinned SDK's
// New<Data>WithDefaults() supplies the value only when the OpenAPI type schema
// declares a `default` (see Schema.HasDefault); otherwise the generated code
// must send it itself.
type RequestDiscriminator struct {
	// GoType is the SDK type the value converts to, e.g. "PlaylistDataType" —
	// the type property's own component name. Empty when the property is
	// inline, whose SDK name is positional and not guessed here.
	GoType string
	// Values are the allowed values. Exactly one means the discriminator is
	// determined and generated code can send it; more than one means the spec
	// does not say which; none means the type is an unconstrained string.
	Values []string
	// SDKDefaulted reports that New<Data>WithDefaults() already assigns the
	// property, so a body that cannot name the value itself is still valid.
	SDKDefaulted bool
}

// Determined reports whether generated code can send the discriminator itself:
// the value is unambiguous and its SDK type is nameable.
func (d *RequestDiscriminator) Determined() bool {
	return d != nil && d.GoType != "" && len(d.Values) == 1
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
	// OneOf is populated when Kind is SchemaKindOneOf, carrying the envelope
	// identity, its non-null alternatives and their SDK binding metadata.
	OneOf *OneOfSpec
	// Variants is the parser's legacy oneOf representation; NormalizeSchemas
	// leaves it empty.
	//
	// Deprecated: use OneOf.Variants.
	Variants []*Schema
	// Type is the primitive type (string/integer/number/boolean).
	Type string
	// Format is the optional OpenAPI format (date-time, int64, ...).
	Format string
	// Enum holds the allowed values, if constrained.
	Enum []string
	// HasDefault records that the schema declares a `default`, ReadOnly that it
	// declares `readOnly: true`. Neither reaches the Terraform schema; together
	// they reproduce the pinned SDK generator's predicate for whether
	// New<Model>WithDefaults() pre-assigns a property (`default` defined, type
	// not object/array, not readOnly).
	HasDefault bool
	ReadOnly   bool
	// WriteOnlySecret retains the OpenAPI writeOnly marker independently from
	// Terraform display sensitivity. On a normalized schema it is the raw
	// OpenAPI annotation; on a merged resource schema it is selected only from
	// the Create or Update request role.
	WriteOnlySecret bool
	// SecretRequiredOnCreate and SecretRequiredOnUpdate record the containing
	// request object's requiredness independently. They are meaningful only
	// when WriteOnlySecret is true and are consumed by generated request routing.
	SecretRequiredOnCreate bool
	SecretRequiredOnUpdate bool
	// WriteOnlyDescription retains the request-side description for generated
	// write-only configuration. The ordinary Description may independently
	// prefer the Read response during cosmetic resource merging.
	WriteOnlyDescription string
	// Sensitive is true when explicitly annotated sensitive or inferred from an
	// unoverridden OpenAPI writeOnly / Datadog x-secret marker.
	Sensitive bool
	// Description is the OpenAPI description, populated during NormalizeSchemas.
	Description string
	// UnsupportedReason explains why a node with Kind == SchemaKindUnsupported
	// cannot be represented, so the affected artifact can fail with a local
	// diagnostic instead of aborting spec loading.
	UnsupportedReason string
	// RefName is the OpenAPI component name that supplied this node, e.g.
	// "ActionConnectionAttributes"; empty for an inline schema. The Datadog
	// go-sdk names its generated models after it, so SDK name accumulation
	// restarts at every node carrying one, preferring it over a parent-derived
	// alternative name.
	RefName string
	// RequestRefName is RefName's request-side counterpart: the Create body's
	// own component name at this node, falling back to Update's when Create
	// does not reach it. Set only by a resource schema merge, which reconciles
	// RefName itself toward the Read response and so cannot leave it naming a
	// Create-only component. Empty when every body left this node inline.
	RequestRefName string
	// Provenance is non-nil only on a node produced by unioning the Create
	// request, Update request and Read response bodies.
	Provenance *SchemaProvenance
}

// SchemaProvenance records which of the three bodies — Create request,
// Update request, Read response — contributed a given node during a resource
// schema merge.
type SchemaProvenance struct {
	// InRequest is true when the node is present in the Create or Update
	// request body.
	InRequest bool
	// RequestRequired is true when the node is named in the Create body's
	// Required list. The Update body is never consulted for requiredness — a
	// PATCH body marks everything optional.
	RequestRequired bool
	// InResponse is true when the node is present in the Read response body.
	InResponse bool
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
	// RefName is the OpenAPI component name of the union node itself, empty for
	// an inline union. Separate from Name, which falls back to a path-derived
	// spelling and so must never be mistaken for an SDK type.
	RefName string
	// SDKType is the Datadog go-sdk oneOf wrapper struct for this union, e.g.
	// "ActionConnectionIntegration". Empty until the SDK binding pass runs.
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
	// SDKPointer is true when the wrapper member and its convenience
	// constructor take a pointer — every alternative except a free-form object,
	// which the SDK emits as a bare, already-nil-able map.
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
	// Cardinality selects the singular vs plural data-source shape. Empty for
	// resources.
	Cardinality Cardinality
	// Description is the artifact's top-level schema doc string, from the
	// tracking extension's tf_description field; empty when the author omits it.
	Description string
	// Schema is the Terraform schema derived from the response, or for a
	// resource from the union of the Create request, Update request and Read
	// response bodies.
	Schema *AttributeTree
	// Lifecycle holds the SDK call bindings. For data sources only Read is set
	Lifecycle *LifecycleBindings
	// SourceFile is the output path, e.g. datadog/fwprovider/<file>.go.
	SourceFile string
	// Diagnostics carries non-fatal notes raised while building the artifact,
	// e.g. query parameters dropped from a plural data source's filter set. The
	// artifact still emits.
	Diagnostics []Diagnostic
	// UnstableOperations are the SDK keys ("v2.GetTwilioIntegrationAccount") of
	// every x-unstable operation this artifact calls, sorted and deduplicated.
	// The pinned SDK defaults each to disabled, so every such call fails at
	// runtime until the provider enables them. Empty when all are stable.
	UnstableOperations []string
}

// AttributeTree is the root of the Terraform schema tree for one artifact.
type AttributeTree struct {
	Attributes []*Attribute
}

// Attribute mirrors a Terraform Plugin Framework attribute or nested container
// one-to-one.
type Attribute struct {
	// Path is the dot-delimited attribute path, e.g. spec.replicas. It doubles
	// as the per-attribute hook ID anchor.
	Path string
	// TfType is the framework type, e.g. schema.StringAttribute.
	TfType string
	// GoType is the corresponding model-struct type, e.g. types.String.
	GoType string
	// ElementType is the framework attr.Type for a list/map element value, e.g.
	// "types.StringType" or "types.ListType{ElemType: types.StringType}". Set
	// only for a collection chain ending in a primitive; empty otherwise.
	ElementType string
	// ElementFormat and ElementIsEnum carry a collection element's own OpenAPI
	// format and enum-ness, which ElementType's generic attr.Type mapping
	// discards (a list of date-time strings and a plain list of strings both
	// map to "types.StringType"). Empty/false for anything else.
	ElementFormat string
	ElementIsEnum bool
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
	// WriteOnlySecret and its role-specific requiredness are copied only while
	// building a managed-resource tree. Response/data-source trees deliberately
	// leave them false even if their OpenAPI response schema carries writeOnly.
	WriteOnlySecret        bool
	SecretRequiredOnCreate bool
	SecretRequiredOnUpdate bool
	WriteOnlyDescription   string

	// InResponse mirrors Schema.Provenance.InResponse, set only for a resource
	// tree. Required alone cannot answer it: a required write-only field and a
	// required field that is also read back both come out Required.
	InResponse bool

	// OpenAPIName is the property name this attribute was built from, before
	// SnakeCase normalized it for Terraform (e.g. "hostTagsLists" behind the
	// "host_tags_lists" Path). Empty for a node with no property name of its
	// own: a root, or an array/map element. It is kept because SnakeCase
	// collapses several spellings onto one, so Path cannot be inverted.
	OpenAPIName string

	// FromPathParameter marks an attribute that came from an operation's path
	// rather than from any request or response body — a sub-resource's parent
	// id. Like InResponse it is not derivable from Required, since a required
	// body field and a required path parameter both come out Required.
	FromPathParameter bool

	// Default is the optional default value, encoded as a Go expression.
	Default *Literal
	// Validators is the fingerprintable validator list for this attribute.
	Validators []ValidatorSpec
	// PlanModifiers holds this attribute's plan modifiers: UseStateForUnknown()
	// on an Optional+Computed attribute, and RequiresReplace() on a
	// request-settable one when no Update role exists or unconditionally on a
	// path parameter. Never set on a Computed-only attribute, where the server
	// may change the value during apply and either modifier would misfire.
	PlanModifiers []PlanModifierSpec
	// Description is always populated from the OpenAPI description.
	Description string
	// Children holds the child attributes of a nested container.
	Children []*Attribute
	// ModelRefName is the OpenAPI component name that supplied the *object*
	// schema backing this attribute's model struct: the node's own schema for an
	// object, its element schema for an array or map of objects. Empty when that
	// schema was inline, and for a leaf, which has no struct. Naming the struct
	// after the component rather than the property that points at it keeps two
	// differently-shaped objects under one property name from colliding.
	ModelRefName string
	// RequestModelRefName mirrors ModelRefName from Schema.RequestRefName
	// rather than Schema.RefName: the request-side (Create/Update) component
	// name, which a resource merge does not overwrite with the Read response's.
	// Non-empty only for a request-settable object node, or a string enum leaf
	// naming its SDK enum type, reached through a resource schema merge.
	RequestModelRefName string
	// OneOf is non-nil when this attribute carries a synthetic oneOf envelope:
	// either the envelope itself (a union at the root or an object property) or
	// the collection whose element is a union. Children then holds the variant
	// attributes, and OneOf their naming and SDK-binding metadata.
	OneOf *OneOfEnvelope
}

// OneOfEnvelope is the Terraform projection of a normalized OneOfSpec: the
// synthetic attribute holding one nested variant attribute per non-null
// alternative, exactly one of which is selected whenever the envelope is
// present. It hangs off the Attribute standing at the union's position, whose
// Children are the projected variant attributes in the same order as Variants.
type OneOfEnvelope struct {
	// Name is the parser-assigned envelope identity (OneOfSpec.Name): the
	// component name for a reusable union, a deterministic path-derived name for
	// an inline one. Two uses of one component share a Name, so one model is
	// generated per envelope rather than per use site.
	Name string
	// GoModel is the generated Go struct holding one pointer field per variant.
	GoModel string
	// SDKType is the Datadog go-sdk oneOf wrapper struct this envelope maps to,
	// carried through from OneOfSpec. A separate identity from Name and GoModel:
	// an inline union's path-derived name names no SDK struct. Empty until the
	// SDK binding pass has run.
	SDKType string
	// Path is the union's own schema path. For a collection of unions it is the
	// element path (e.g. "response.choices[]"), which is the path of no
	// attribute in the tree.
	Path string
	// Optional permits the whole envelope to be absent because its containing
	// OpenAPI field is optional or nullable. When false, exactly one variant must
	// be selected.
	Optional bool
	// Computed marks a response-only envelope, where selection is enforced by
	// response mapping rather than by practitioner configuration.
	Computed bool
	// Variants are the projected non-null alternatives, ordered by TFName so
	// neither OpenAPI order nor map iteration can reach the output.
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
	// alternative, and SDKConstructor the SDK convenience constructor for it.
	// Both are carried through from the parser, never derived from a Terraform
	// name: a variant named aws_integration binds to AWSIntegration, not
	// AwsIntegration. Empty until the SDK binding pass resolves them.
	SDKField       string
	SDKConstructor string
	// SDKPointer is true when the SDK wrapper member and its convenience
	// constructor take a pointer — every alternative except a free-form object,
	// which the SDK emits as a bare, already-nil-able map.
	SDKPointer bool
	// ValueWrapped is true for every non-object alternative — scalar, list, map,
	// or a directly nested union — whose block holds a single child named
	// "value". Object alternatives expose their own fields directly instead.
	ValueWrapped bool
	// Attribute is this variant's projected block — the same pointer as the
	// envelope-carrying attribute's Children entry at this index. Children
	// drives schema rendering; this reaches a block without re-deriving order.
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

// PlanModifierSpec is one plan modifier attached to a resource attribute,
// rendered as a Go constructor call and typed per framework value kind
// (planmodifier.Bool, planmodifier.Object, ...) derived from Attribute.GoType.
type PlanModifierSpec struct {
	// Name is the plan modifier constructor, e.g.
	// stringplanmodifier.UseStateForUnknown. No modifier the resource tree emits
	// takes arguments, and the template renders every one as Name().
	Name string
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
	Search *SDKCall
	Update *SDKCall
	Delete *SDKCall
	// UpdateUnsupported records that the tracking group declares no update role,
	// so no endpoint can modify this resource in place. Not derivable from
	// Update == nil, which is also true of every data source, where this flag is
	// false.
	UpdateUnsupported bool
	IdStrategy        IdStrategy
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
	// Rule: SdkClassName(Operation.Tag) — strip every non-alphanumeric
	// character, then append "Api", with no re-capitalization.
	GoApiStruct string
	// GoMethod is the method name on GoApiStruct, e.g. "CreateOrgGroup".
	// Rule: Operation.OperationId, no transformation applied.
	GoMethod string
	// GoRequestType is the SDK request body type, e.g. "OrgGroupCreateRequest".
	// Rule: last path component of the requestBody $ref. Empty when the
	// operation takes no request body (e.g. DELETE, GET-by-ID).
	GoRequestType string
	// GoResponseType is the SDK response type, e.g. "OrgGroupResponse".
	// Rule: last path component of the 2xx response schema $ref. Empty when the
	// operation returns no body (e.g. 204 No Content).
	GoResponseType string
	// GoRequestDataType is the SDK type of the request body's JSON:API "data"
	// member, e.g. "IncidentTypeCreateData": the RefName of RequestSchema's
	// "data" property, read per role since Create's and Update's routinely
	// differ. Only its New<...>WithDefaults() sets the "type" discriminator —
	// the wrapper's returns a zero struct without building Data. Empty when the
	// operation sends no body, or "data" is inline and names no SDK component.
	GoRequestDataType string
	// RequestDeclaresID records that the body's data member declares an id at
	// all. A body that does not must not call SetId: the SDK generates that
	// setter only for a declared property, so the call would not compile.
	RequestDeclaresID bool
	// RequestIDGoType is the Go type the SDK's SetId takes on this body's data
	// member, derived from the data.id schema via SDKScalarGoType, e.g.
	// "string" or "uuid.UUID". Empty when the body declares no id, or its
	// format is one the SDK generator cannot type. Read from the body, not
	// borrowed from the path parameter: an int64 path id can carry a string id.
	RequestIDGoType string
	// RequestDiscriminator describes the request body's JSON:API data.type
	// member. Nil when the operation sends no body or the data component
	// declares no type property.
	RequestDiscriminator *RequestDiscriminator
	// GoRequestAttributesType is the SDK type of the request body's
	// data.attributes member, e.g. "IncidentTypeAttributes", derived the same
	// way. Empty when the envelope carries no attributes object.
	GoRequestAttributesType string
	// RequestAttributesSchema is this role's own data.attributes node, which
	// narrows the merged tree to the fields *this* body declares and carries
	// the per-role SDK names (nested and element components, enum types, oneOf
	// wrappers and their `<Member>As<Union>` constructors). The merged request
	// side is the union of the Create and Update bodies, which is right for the
	// Terraform schema and wrong here: the SDK generates Set<Field> only on the
	// body that declares the field, and spells one merged name where the SDK
	// declares two or three (…SettingsRequest vs …SettingsUpdate).
	RequestAttributesSchema *Schema
	// Arguments are the required positional SDK arguments in call order.
	Arguments []SDKArgument
	// OptionalArguments bind Terraform filters to OptionalParamsType setters.
	OptionalArguments []SDKArgument

	// The fields below back a plural data-source list call.

	// ItemType is the SDK element type yielded by the list call, e.g. "Team".
	// A non-paginated read collects resp.Data into []<ItemType>; a paginated
	// one yields PaginationResult[<ItemType>].
	ItemType string
	// OptionalParamsType is the SDK optional-parameters struct, e.g.
	// "ListTeamsOptionalParameters" (<GoMethod>OptionalParameters). Empty when
	// the endpoint declares no query parameters, and the call then takes no
	// optional-parameters argument.
	OptionalParamsType string
	// Paginated selects the "<GoMethod>WithPagination" iterator form, set when
	// the operation declares an x-pagination extension.
	Paginated bool
}

// ----------------------------------------------------------------------------
// Run-report types
//
// Field names and JSON tags mirror contracts/run-report.schema.json.
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
	// ArtifactStatusRegistrationRetired marks a stale registration dropped on
	// its own: the constructor was still listed in datasources_generated.go but
	// its generated files were already gone, so only the registration line
	// changed. Constructor carries the removed identifier.
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
	// registration identifier, authoritative because the artifact name cannot
	// be recovered once the generated files are gone.
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
