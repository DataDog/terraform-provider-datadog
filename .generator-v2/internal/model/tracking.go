package model

// TrackingFieldMetadata is the decoded form of the x-datadog-tf-generator
// OpenAPI extension on a flagged operation. It is nil on an Operation whose
// extension is absent, and describes the shape only — decoding lives elsewhere.
type TrackingFieldMetadata struct {
	// ArtifactKind selects resource (full CRUD) vs data_source (read-only).
	// Required.
	ArtifactKind ArtifactKind `json:"artifact_kind"`
	// ArtifactName is the Terraform-facing name without the datadog_ prefix,
	// lowercase snake_case, unique per artifact_kind (resources and data
	// sources are separate Terraform namespaces). Required.
	ArtifactName string `json:"artifact_name"`
	// Cardinality selects singular (one item by id) vs plural (filtered list)
	// for a data source. Optional; absent/empty decodes to singular.
	Cardinality Cardinality `json:"cardinality,omitempty"`
	// TfDescription is the author-supplied doc string for the generated
	// artifact's top-level Terraform schema. Optional; empty when omitted.
	TfDescription string `json:"tf_description,omitempty"`
	// Group names the OpenAPI operations backing this artifact: create/read/
	// update/delete for a resource, read (by-id) and/or search (list) for a data
	// source. At least one of Read/Search is required when Group is present.
	Group *OperationGroup `json:"group,omitempty"`
	// IdStrategy is how the Terraform resource ID is derived from the API
	// response. Defaults to "data.id" when omitted.
	IdStrategy IdStrategy `json:"id_strategy,omitempty"`
	// Sensitive, when attached to a Schema Object, marks the attribute as
	// Terraform-sensitive.
	Sensitive bool `json:"sensitive,omitempty"`
	// Cassette controls example-backed test generation for this artifact.
	//
	// Generation is on by default, so a resource whose description can
	// describe its own lifecycle gets a test without anyone remembering to ask.
	// Setting it false opts out and is the only way to decline.
	//
	// It is a pointer because the three states are not two: absent means
	// "defaulted in", true means "asked for by name", and the difference
	// decides what an undescribed target costs. A target that defaulted in and
	// cannot be described is skipped; one that asked explicitly is ineligible
	// and fails the run, because an explicit annotation is a request.
	Cassette *bool `json:"cassette,omitempty"`
	// Skip explicitly disables generation while keeping the annotation in
	// place, equivalent to removing the extension.
	Skip bool `json:"skip,omitempty"`
	// Overwrites names the hand-written data source constructor this generated
	// artifact supersedes, e.g. "NewDatadogTeamDataSource". When set, the
	// generated file overwrites the hand-written one in place, that constructor is
	// removed from the FrameworkProvider Datasources slice, and the generated one
	// is registered in generatedDatasources. Empty when purely additive.
	Overwrites string `json:"overwrites,omitempty"`
}

// OperationGroup references, by operationId, the OpenAPI operations backing an
// artifact: the create/read/update/delete lifecycle for a resource, or read
// (by-id) and/or search (list) for a data source.
type OperationGroup struct {
	// Create is the operationId of the Create endpoint.
	Create string `json:"create,omitempty"`
	// Read is the operationId of the Read (by-id) endpoint. At least one of
	// Read/Search must be present.
	Read string `json:"read,omitempty"`
	// Search is the operationId of the list endpoint a singular data source uses
	// to resolve a single match. Never inferred; declared explicitly.
	Search string `json:"search,omitempty"`
	// Update is the operationId of the Update endpoint. May be omitted; the
	// generator then forces replacement on every request-settable attribute
	// (RequiresReplace()), never on a Computed-only one, since there would be no
	// endpoint to reconcile a server-side change through anyway.
	Update string `json:"update,omitempty"`
	// Delete is the operationId of the Delete endpoint.
	Delete string `json:"delete,omitempty"`
}

// CassetteEnabled reports whether this artifact should get an example-backed
// test. Generation is on by default, so only an explicit false declines.
func (t *TrackingFieldMetadata) CassetteEnabled() bool {
	return t != nil && (t.Cassette == nil || *t.Cassette)
}

// CassetteRequested reports whether the description asked for a test by name
// rather than receiving one by default. An explicit request that cannot be
// satisfied fails the run; a default one is skipped.
func (t *TrackingFieldMetadata) CassetteRequested() bool {
	return t != nil && t.Cassette != nil && *t.Cassette
}
