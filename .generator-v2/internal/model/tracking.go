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
	// lowercase snake_case, unique across the spec. Required.
	ArtifactName string `json:"artifact_name"`
	// Group declares which operations form the C/R/U/D quadruple. Required for
	// resources; for data sources only Read is meaningful.
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

// OperationGroup references, by operationId, the OpenAPI operations that form a
// resource's C/R/U/D quadruple.
type OperationGroup struct {
	// Create is the operationId of the Create endpoint.
	Create string `json:"create,omitempty"`
	// Read is the operationId of the Read endpoint. Required.
	Read string `json:"read"`
	// Update is the operationId of the Update endpoint. May be omitted; the
	// generator then marks all attributes ForceNew .
	Update string `json:"update,omitempty"`
	// Delete is the operationId of the Delete endpoint.
	Delete string `json:"delete,omitempty"`
}
