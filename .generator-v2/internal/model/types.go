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

// SchemaKind classifies a normalized Schema node after allOf flattening and
// oneOf/anyOf variant detection.
type SchemaKind string

const (
	SchemaKindPrimitive SchemaKind = "primitive"
	SchemaKindObject    SchemaKind = "object"
	SchemaKindArray     SchemaKind = "array"
	SchemaKindMap       SchemaKind = "map"
	SchemaKindVariant   SchemaKind = "variant" // oneOf / anyOf
	SchemaKindRefCycle  SchemaKind = "ref_cycle"
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
}

// Operation is a single OpenAPI operation, tagged with whether it is in scope
// for generation.
type Operation struct {
	// Path is the OpenAPI path template, e.g. /api/v2/users/{user_id}.
	Path string
	// Method is the HTTP method (GET/POST/PUT/PATCH/DELETE).
	Method string
	// OperationId is the OpenAPI operationId, used as the SDK method anchor.
	OperationId string
	// Tag is the OpenAPI tag, driving SDK package selection. Must be non-empty
	// when Tracking != nil.
	Tag string
	// Tracking is the decoded tracking-field extension; nil iff the extension
	// is absent. Defined by tracking.go.
	Tracking *TrackingFieldMetadata
	// RequestSchema is the resolved request body schema, if any.
	RequestSchema *Schema
	// ResponseSchema is the resolved 2xx response schema, if any.
	ResponseSchema *Schema
}

// Schema is a normalized, recursive view of an OpenAPI schema after allOf
// flattening and oneOf/anyOf variant detection.
type Schema struct {
	Kind SchemaKind
	// Properties is populated for objects only; iteration is always sorted.
	Properties map[string]*Schema
	// Required is populated for objects only; sorted.
	Required []string
	// Items is populated for arrays only.
	Items *Schema
	// Variants is populated for oneOf/anyOf only.
	Variants []*Schema
	// Type is the primitive type (string/integer/number/boolean).
	Type string
	// Format is the optional OpenAPI format (date-time, int64, ...).
	Format string
	// Enum holds the allowed values, if constrained.
	Enum []string
	// Sensitive is true when the schema is annotated sensitive: true.
	Sensitive bool
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
	// Schema is the Terraform schema derived from the response (and request,
	// for resources).
	Schema *AttributeTree
	// Lifecycle is set for resources only; data sources carry only a Read.
	Lifecycle *LifecycleBindings
	// SourceFile is the output path, e.g. datadog/fwprovider/<file>.go.
	SourceFile string
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

// LifecycleBindings maps each Terraform CRUD method to the SDK call that
// implements it. Resources only.
type LifecycleBindings struct {
	Create     *SDKCall
	Read       *SDKCall
	Update     *SDKCall
	Delete     *SDKCall
	IdStrategy IdStrategy
}

// SDKCall is a single datadog-api-client-go invocation plus the mappers that
// translate to and from the Terraform model.
type SDKCall struct {
	// OperationId is used to resolve the SDK method via reflection.
	OperationId string
	// RequestMappers populate the request object from the Terraform model.
	RequestMappers []Mapper
	// ResponseMappers populate the Terraform model from the SDK response.
	ResponseMappers []Mapper
}

// Mapper describes a single field-level translation between the Terraform
// model and an SDK request/response type.
type Mapper struct {
	// TfPath is the dotted attribute path in the Terraform model, e.g. spec.replicas.
	TfPath string
	// SdkPath is the corresponding field path on the SDK type.
	SdkPath string
	// GoType is the Go type used at this mapping site, e.g. types.String.
	GoType string
}
