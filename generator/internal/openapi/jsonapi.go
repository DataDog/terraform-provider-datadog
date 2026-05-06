package openapi

import (
	"fmt"

	"github.com/pb33f/libopenapi/datamodel/high/base"
)

// IsJSONAPIEnvelope checks if a schema follows the JSON:API pattern:
// a top-level object with a "data" property that has "id", "type", and "attributes".
func IsJSONAPIEnvelope(schema *base.Schema) bool {
	if schema == nil || schema.Properties == nil {
		return false
	}

	dataProxy, ok := schema.Properties.Get("data")
	if !ok || dataProxy == nil {
		return false
	}

	dataSchema, err := dataProxy.BuildSchema()
	if err != nil || dataSchema == nil || dataSchema.Properties == nil {
		return false
	}

	_, hasID := dataSchema.Properties.Get("id")
	_, hasType := dataSchema.Properties.Get("type")
	_, hasAttrs := dataSchema.Properties.Get("attributes")

	return hasID && hasType && hasAttrs
}

// UnwrapJSONAPI extracts the attributes schema proxy from a JSON:API envelope.
// Returns the attributes SchemaProxy and the data type name (from the type enum).
func UnwrapJSONAPI(schema *base.Schema) (*base.SchemaProxy, string, error) {
	if schema == nil || schema.Properties == nil {
		return nil, "", fmt.Errorf("schema has no properties")
	}

	dataProxy, ok := schema.Properties.Get("data")
	if !ok || dataProxy == nil {
		return nil, "", fmt.Errorf("schema has no 'data' property")
	}

	dataSchema, err := dataProxy.BuildSchema()
	if err != nil {
		return nil, "", fmt.Errorf("building data schema: %w", err)
	}

	if dataSchema.Properties == nil {
		return nil, "", fmt.Errorf("data schema has no properties")
	}

	attrsProxy, ok := dataSchema.Properties.Get("attributes")
	if !ok || attrsProxy == nil {
		return nil, "", fmt.Errorf("data schema has no 'attributes' property")
	}

	// Extract the type name from the type enum if available.
	var typeName string
	typeProxy, ok := dataSchema.Properties.Get("type")
	if ok && typeProxy != nil {
		typeSchema, err := typeProxy.BuildSchema()
		if err == nil && len(typeSchema.Enum) > 0 {
			typeName = typeSchema.Enum[0].Value
		}
	}

	return attrsProxy, typeName, nil
}
