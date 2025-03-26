package fwutils

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	frameworkSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

var stringType = reflect.TypeOf("")

func EnrichFrameworkResourceSchema(s *frameworkSchema.Schema) {
	for i, attr := range s.Attributes {
		s.Attributes[i] = enrichDescription(attr)
	}
	enrichMapBlocks(s.Blocks)
}

func enrichMapBlocks(blocks map[string]frameworkSchema.Block) {
	for _, block := range blocks {
		switch v := block.(type) {
		case frameworkSchema.ListNestedBlock:
			for i, attr := range v.NestedObject.Attributes {
				v.NestedObject.Attributes[i] = enrichDescription(attr)
			}
			enrichMapBlocks(v.NestedObject.Blocks)
		case frameworkSchema.SingleNestedBlock:
			for i, attr := range v.Attributes {
				v.Attributes[i] = enrichDescription(attr)
			}
			enrichMapBlocks(v.Blocks)
		case frameworkSchema.SetNestedBlock:
			for i, attr := range v.NestedObject.Attributes {
				v.NestedObject.Attributes[i] = enrichDescription(attr)
			}
			enrichMapBlocks(v.NestedObject.Blocks)
		}
	}
}

func enrichDescription(r any) frameworkSchema.Attribute {
	switch v := r.(type) {
	case frameworkSchema.StringAttribute:
		buildEnrichedSchemaDescription(reflect.ValueOf(&v))
		return v
	case frameworkSchema.Int64Attribute:
		buildEnrichedSchemaDescription(reflect.ValueOf(&v))
		return v
	case frameworkSchema.Float64Attribute:
		buildEnrichedSchemaDescription(reflect.ValueOf(&v))
		return v
	case frameworkSchema.BoolAttribute:
		buildEnrichedSchemaDescription(reflect.ValueOf(&v))
		return v
	default:
		return r.(frameworkSchema.Attribute)
	}
}

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
				currentDesc = fmt.Sprintf("%s. Valid values are %s.", currentDesc, validValuesMsg)
				break
			}

			// String validators
			if strings.HasPrefix(rv_validators.Index(i).Elem().Type().Name(), "regexMatchesValidator") ||
				strings.HasPrefix(rv_validators.Index(i).Elem().Type().Name(), "validEntityYAMLValidator") ||
				strings.HasPrefix(rv_validators.Index(i).Elem().Type().Name(), "cidrIpValidator") ||
				strings.HasPrefix(rv_validators.Index(i).Elem().Type().Name(), "lengthAtLeastValidator") ||
				// BetweenValidator is a "homemade" validator and does not come from Hashicorp, it lives in out validators package as Float64Between
				// It validates a string representation of a float64
				strings.HasPrefix(rv_validators.Index(i).Elem().Type().Name(), "BetweenValidator") {
				validationMessage := rv_validators.Index(i).Elem().Interface().(validator.String).Description(context.Background())
				currentDesc = fmt.Sprintf("%s %s.", currentDesc, validationMessage)
				break
			}

			// Int64 validators
			if strings.HasPrefix(rv_validators.Index(i).Elem().Type().Name(), "betweenValidator") ||
				strings.HasPrefix(rv_validators.Index(i).Elem().Type().Name(), "atLeastValidator") {
				validationMessage := rv_validators.Index(i).Elem().Interface().(validator.Int64).Description(context.Background())
				currentDesc = fmt.Sprintf("%s %s.", currentDesc, validationMessage)
				break
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
