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

// SchemaKind classifies a normalized Schema node by structure. Primitive,
// Object, Array and Map are emittable as Terraform attributes. Variant
// (oneOf/anyOf), RefCycle ($ref cycle or beyond --max-depth) and Unsupported
// (no representable type or structure) are not — the representability check
// rejects them rather than emitting a types.Dynamic escape hatch.
type SchemaKind string

const (
	SchemaKindPrimitive   SchemaKind = "primitive"
	SchemaKindObject      SchemaKind = "object"
	SchemaKindArray       SchemaKind = "array"
	SchemaKindMap         SchemaKind = "map"
	SchemaKindVariant     SchemaKind = "variant"     // oneOf / anyOf
	SchemaKindRefCycle    SchemaKind = "ref_cycle"   // $ref cycle or beyond --max-depth
	SchemaKindUnsupported SchemaKind = "unsupported" // no representable type/structure; always rejected
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
	// ResponseSchema is the resolved 2xx response schema, if any.
	ResponseSchema *Schema
	// ResponseRefName is the last path segment of the 2xx response body $ref,
	// e.g. "IncidentTypeResponse" — the SDK Go response type; empty when the
	// body is inline or absent.
	ResponseRefName string
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
	// Description is the OpenAPI description, populated during NormalizeSchemas.
	Description string
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
	// Lifecycle holds the SDK call bindings. For data sources only Read is set
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
	// ElementType is the framework attr.Type for a list/map element value,
	// e.g. "types.StringType". Set ONLY for ListAttribute/MapAttribute
	// (collection-of-primitive); empty for everything else.
	ElementType string

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

// LifecycleBindings maps Terraform lifecycle methods to their SDK calls.
// For data sources only Read is populated; IdStrategy and Create/Update/Delete are zero.
type LifecycleBindings struct {
	Create     *SDKCall
	Read       *SDKCall
	Update     *SDKCall
	Delete     *SDKCall
	IdStrategy IdStrategy
}

// SDKCall represents a single datadog-api-client-go invocation.
type SDKCall struct {
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
	Created   int `json:"created"`
	Updated   int `json:"updated"`
	Unchanged int `json:"unchanged"`
	Skipped   int `json:"skipped"`
	Failed    int `json:"failed"`
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
