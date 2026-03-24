package fwutils

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	datasourceSchema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	ephemeralSchema "github.com/hashicorp/terraform-plugin-framework/ephemeral/schema"
	resourceSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

var stringType = reflect.TypeOf("")

// =============================================================================
// RESOURCE SCHEMA ENRICHMENT FUNCTIONS
// =============================================================================

func EnrichFrameworkResourceSchema(s *resourceSchema.Schema) {
	for i, attr := range s.Attributes {
		s.Attributes[i] = enrichResourceDescription(attr)
	}
	enrichResourceMapBlocks(s.Blocks)
}

func enrichResourceMapBlocks(blocks map[string]resourceSchema.Block) {
	for _, block := range blocks {
		switch v := block.(type) {
		case resourceSchema.ListNestedBlock:
			for i, attr := range v.NestedObject.Attributes {
				v.NestedObject.Attributes[i] = enrichResourceDescription(attr)
			}
			enrichResourceMapBlocks(v.NestedObject.Blocks)
		case resourceSchema.SingleNestedBlock:
			for i, attr := range v.Attributes {
				v.Attributes[i] = enrichResourceDescription(attr)
			}
			enrichResourceMapBlocks(v.Blocks)
		case resourceSchema.SetNestedBlock:
			for i, attr := range v.NestedObject.Attributes {
				v.NestedObject.Attributes[i] = enrichResourceDescription(attr)
			}
			enrichResourceMapBlocks(v.NestedObject.Blocks)
		}
	}
}

func enrichResourceDescription(r any) resourceSchema.Attribute {
	switch v := r.(type) {
	case resourceSchema.StringAttribute:
		buildEnrichedSchemaDescription(reflect.ValueOf(&v))
		return v
	case resourceSchema.Int64Attribute:
		buildEnrichedSchemaDescription(reflect.ValueOf(&v))
		return v
	case resourceSchema.Float64Attribute:
		buildEnrichedSchemaDescription(reflect.ValueOf(&v))
		return v
	case resourceSchema.BoolAttribute:
		buildEnrichedSchemaDescription(reflect.ValueOf(&v))
		return v
	default:
		return r.(resourceSchema.Attribute)
	}
}

// =============================================================================
// DATASOURCE SCHEMA ENRICHMENT FUNCTIONS
// =============================================================================

func EnrichFrameworkDatasourceSchema(s *datasourceSchema.Schema) {
	for i, attr := range s.Attributes {
		s.Attributes[i] = enrichDatasourceDescription(attr)
	}
	enrichDatasourceMapBlocks(s.Blocks)
}

func enrichDatasourceMapBlocks(blocks map[string]datasourceSchema.Block) {
	for _, block := range blocks {
		switch v := block.(type) {
		case datasourceSchema.ListNestedBlock:
			for i, attr := range v.NestedObject.Attributes {
				v.NestedObject.Attributes[i] = enrichDatasourceDescription(attr)
			}
			enrichDatasourceMapBlocks(v.NestedObject.Blocks)
		case datasourceSchema.SingleNestedBlock:
			for i, attr := range v.Attributes {
				v.Attributes[i] = enrichDatasourceDescription(attr)
			}
			enrichDatasourceMapBlocks(v.Blocks)
		case datasourceSchema.SetNestedBlock:
			for i, attr := range v.NestedObject.Attributes {
				v.NestedObject.Attributes[i] = enrichDatasourceDescription(attr)
			}
			enrichDatasourceMapBlocks(v.NestedObject.Blocks)
		}
	}
}

func enrichDatasourceDescription(r any) datasourceSchema.Attribute {
	switch v := r.(type) {
	case datasourceSchema.StringAttribute:
		buildEnrichedSchemaDescription(reflect.ValueOf(&v))
		return v
	case datasourceSchema.Int64Attribute:
		buildEnrichedSchemaDescription(reflect.ValueOf(&v))
		return v
	case datasourceSchema.Float64Attribute:
		buildEnrichedSchemaDescription(reflect.ValueOf(&v))
		return v
	case datasourceSchema.BoolAttribute:
		buildEnrichedSchemaDescription(reflect.ValueOf(&v))
		return v
	default:
		return r.(datasourceSchema.Attribute)
	}
}

// =============================================================================
// EPHEMERAL SCHEMA ENRICHMENT FUNCTIONS
// =============================================================================

func EnrichFrameworkEphemeralResourceSchema(s *ephemeralSchema.Schema) {
	for i, attr := range s.Attributes {
		s.Attributes[i] = enrichEphemeralDescription(attr)
	}
	enrichEphemeralMapBlocks(s.Blocks)
}

func enrichEphemeralMapBlocks(blocks map[string]ephemeralSchema.Block) {
	for _, block := range blocks {
		switch v := block.(type) {
		case ephemeralSchema.ListNestedBlock:
			for i, attr := range v.NestedObject.Attributes {
				v.NestedObject.Attributes[i] = enrichEphemeralDescription(attr)
			}
			enrichEphemeralMapBlocks(v.NestedObject.Blocks)
		case ephemeralSchema.SingleNestedBlock:
			for i, attr := range v.Attributes {
				v.Attributes[i] = enrichEphemeralDescription(attr)
			}
			enrichEphemeralMapBlocks(v.Blocks)
		case ephemeralSchema.SetNestedBlock:
			for i, attr := range v.NestedObject.Attributes {
				v.NestedObject.Attributes[i] = enrichEphemeralDescription(attr)
			}
			enrichEphemeralMapBlocks(v.NestedObject.Blocks)
		}
	}
}

func enrichEphemeralDescription(r any) ephemeralSchema.Attribute {
	switch v := r.(type) {
	case ephemeralSchema.StringAttribute:
		buildEnrichedSchemaDescription(reflect.ValueOf(&v))
		return v
	case ephemeralSchema.Int64Attribute:
		buildEnrichedSchemaDescription(reflect.ValueOf(&v))
		return v
	case ephemeralSchema.Float64Attribute:
		buildEnrichedSchemaDescription(reflect.ValueOf(&v))
		return v
	case ephemeralSchema.BoolAttribute:
		buildEnrichedSchemaDescription(reflect.ValueOf(&v))
		return v
	default:
		return r.(ephemeralSchema.Attribute)
	}
}

// =============================================================================
// REUSABLE CORE FUNCTIONS (TYPE-AGNOSTIC VIA REFLECTION)
// =============================================================================

func buildEnrichedSchemaDescription(rv reflect.Value) {
	descField := rv.Elem().FieldByName("Description")
	currentDesc := descField.String()

	// Build description with rv_validators
	rv_validators := rv.Elem().FieldByName("Validators")
	if rv_validators.IsValid() && !rv_validators.IsNil() && rv_validators.Len() > 0 {
		for i := 0; i < rv_validators.Len(); i++ {
			if strings.HasPrefix(rv_validators.Index(i).Elem().Type().Name(), "enumValidator") {
				enrichSchema := rv_validators.Index(i).Elem().FieldByName("enrichSchema").Bool()
				if !enrichSchema {
					continue
				}
				allowedValues := rv_validators.Index(i).Elem().FieldByName("AllowedEnumValues")
				v := reflect.ValueOf(allowedValues.Interface())
				validValuesMsg := ""
				sep := ""
				for i := 0; i < v.Len(); i++ {
					if len(validValuesMsg) > 0 {
						sep = ", "
					}
					validValuesMsg += fmt.Sprintf("%s`%v`", sep, v.Index(i).Interface())
				}
				currentDesc = fmt.Sprintf("%s Valid values are %s.", currentDesc, validValuesMsg)
				break
			}
			if strings.HasPrefix(rv_validators.Index(i).Elem().Type().Name(), "oneOfValidator") {
				allowedValues := rv_validators.Index(i).Elem().FieldByName("values")
				validValuesMsg := ""
				sep := ""
				for i := 0; i < allowedValues.Len(); i++ {
					if len(validValuesMsg) > 0 {
						sep = ", "
					}
					// Index(i).Field(1) is the value of the types.String
					// If we would use "only" Index(i) we would have { 2 <VALUE> }
					validValuesMsg += fmt.Sprintf("%s`%v`", sep, allowedValues.Index(i).Field(1))
				}
				currentDesc = fmt.Sprintf("%s Valid values are %s.", currentDesc, validValuesMsg)
				break
			}

			// String validators
			if strings.HasPrefix(rv_validators.Index(i).Elem().Type().Name(), "regexMatchesValidator") ||
				strings.HasPrefix(rv_validators.Index(i).Elem().Type().Name(), "validEntityYAMLValidator") ||
				strings.HasPrefix(rv_validators.Index(i).Elem().Type().Name(), "cidrIpValidator") ||
				strings.HasPrefix(rv_validators.Index(i).Elem().Type().Name(), "lengthAtLeastValidator") ||
				// BetweenValidator is a custom validator and does not come from Hashicorp, it lives in out validators package as Float64Between
				// It validates a string representation of a float64
				strings.HasPrefix(rv_validators.Index(i).Elem().Type().Name(), "BetweenValidator") {
				validationMessage := rv_validators.Index(i).Elem().Interface().(validator.String).Description(context.Background())
				currentDesc = fmt.Sprintf("%s %s", ensureTrailingPoint(currentDesc), formatDescription(validationMessage))
				break
			}

			// Float64 and Int64 validators (both use "betweenValidator" and "atLeastValidator" names)
			// Try Float64 first, then fall back to Int64
			if strings.HasPrefix(rv_validators.Index(i).Elem().Type().Name(), "betweenValidator") ||
				strings.HasPrefix(rv_validators.Index(i).Elem().Type().Name(), "atLeastValidator") {

				// Try Float64 validator first
				if float64Val, ok := rv_validators.Index(i).Elem().Interface().(validator.Float64); ok {
					validationMessage := float64Val.Description(context.Background())
					currentDesc = fmt.Sprintf("%s %s", ensureTrailingPoint(currentDesc), formatDescription(validationMessage))
					break
				}

				// Fall back to Int64 validator
				if int64Val, ok := rv_validators.Index(i).Elem().Interface().(validator.Int64); ok {
					validationMessage := int64Val.Description(context.Background())
					currentDesc = fmt.Sprintf("%s %s", ensureTrailingPoint(currentDesc), formatDescription(validationMessage))
					break
				}
			}

		}
	}

	// Build description with Defaults
	defaultField := rv.Elem().FieldByName("Default")
	if defaultField.IsValid() && !defaultField.IsNil() {
		defaultVal := defaultField.Elem().FieldByName("defaultVal")
		if defaultVal.IsValid() {
			switch defaultVal.Type() {
			case stringType:
				currentDesc = fmt.Sprintf("%s Defaults to `\"%v\"`.", currentDesc, defaultVal)
			default:
				currentDesc = fmt.Sprintf("%s Defaults to `%v`.", currentDesc, defaultVal)
			}
		}
	}

	descField.SetString(currentDesc)
}

func formatDescription(s string) string {
	return ensureTrailingPoint(ensureCapitalize(s))
}

func ensureCapitalize(s string) string {
	if len(s) == 0 {
		return s
	}
	return strings.ToUpper(s[0:1]) + s[1:]
}

func ensureTrailingPoint(s string) string {
	if len(s) == 0 {
		return s
	}
	if s[len(s)-1:] == "." {
		return s
	}
	return s + "."
}
