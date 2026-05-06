package types

// AttributeMode represents how a Terraform attribute is used.
type AttributeMode int

const (
	// Computed means the attribute is set by the provider (read-only).
	Computed AttributeMode = iota
	// Required means the attribute must be set by the user.
	Required
	// Optional means the attribute may be set by the user.
	Optional
)

// ResolveAttributeMode determines the Terraform attribute mode for a field.
// Rules:
//   - Path parameters are Required (used for lookup)
//   - Query parameters are Optional (used for filtering)
//   - Response fields are Computed (read-only)
func ResolveAttributeMode(isPathParam, isQueryParam bool) AttributeMode {
	if isPathParam {
		return Required
	}
	if isQueryParam {
		return Optional
	}
	return Computed
}
