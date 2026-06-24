package model

// TrackingFieldMetadata is the decoded form of the x-datadog-tf-generator
// OpenAPI extension on a flagged operation. It is nil on an Operation whose
// extension is absent.
//
// Decoding is performed by `parser.DecodeTracking`. This struct only describes
// the shape.
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
	// Skip explicitly disables generation while keeping the annotation in
	// place, equivalent to removing the extension.
	Skip bool `json:"skip,omitempty"`
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
	// generator then marks all attributes ForceNew .
	Update string `json:"update,omitempty"`
	// Delete is the operationId of the Delete endpoint.
	Delete string `json:"delete,omitempty"`
}
