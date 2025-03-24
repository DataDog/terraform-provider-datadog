package fwutils

import (
	"fmt"
	"reflect"
	"strings"

	frameworkSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
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

	// Build description with validators
	validators := rv.Elem().FieldByName("Validators")
	if validators.IsValid() && !validators.IsNil() && validators.Len() > 0 {
		for i := 0; i < validators.Len(); i++ {
			if strings.HasPrefix(validators.Index(i).Elem().Type().Name(), "enumValidator") {
				enrichSchema := validators.Index(i).Elem().FieldByName("enrichSchema").Bool()
				if !enrichSchema {
					continue
				}
				allowedValues := validators.Index(i).Elem().FieldByName("AllowedEnumValues")
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
			if strings.HasPrefix(validators.Index(i).Elem().Type().Name(), "oneOfValidator") {
				allowedValues := validators.Index(i).Elem().FieldByName("values")
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
			if strings.HasPrefix(validators.Index(i).Elem().Type().Name(), "regexMatchesValidator") {
				validationMessage := validators.Index(i).Elem().FieldByName("message").String()
				currentDesc = fmt.Sprintf("%s %s", currentDesc, validationMessage)
				break
			}
			if strings.HasPrefix(validators.Index(i).Elem().Type().Name(), "validEntityYAMLValidator") {
				currentDesc = fmt.Sprintf("%s entity must be a valid entity YAML/JSON structure.", currentDesc)
				break
			}
			if strings.HasPrefix(validators.Index(i).Elem().Type().Name(), "cidrIpValidator") {
				currentDesc = fmt.Sprintf("%s String must be a valid CIDR block or IP address.", currentDesc)
				break
			}

			if strings.HasPrefix(validators.Index(i).Elem().Type().Name(), "lengthAtLeastValidator") {
				minLength := validators.Index(i).Elem().FieldByName("minLength").Int()
				currentDesc = fmt.Sprintf("[Length > %d] %s", minLength, currentDesc)
				break
			}
			if strings.HasPrefix(validators.Index(i).Elem().Type().Name(), "betweenValidator") {
				min := validators.Index(i).Elem().FieldByName("min").Int()
				max := validators.Index(i).Elem().FieldByName("max").Int()
				currentDesc = fmt.Sprintf("[Min %d, Max %d] %s", min, max, currentDesc)
				break
			}
			// Must have a different case for float validators
			if strings.HasPrefix(validators.Index(i).Elem().Type().Name(), "BetweenValidator") {
				min := validators.Index(i).Elem().FieldByName("min").Float()
				max := validators.Index(i).Elem().FieldByName("max").Float()
				currentDesc = fmt.Sprintf("[Min %.1f, Max %.1f] %s", min, max, currentDesc)
				break
			}
			if strings.HasPrefix(validators.Index(i).Elem().Type().Name(), "atLeastValidator") {
				min := validators.Index(i).Elem().FieldByName("min").Int()
				currentDesc = fmt.Sprintf("[Min %d] %s", min, currentDesc)
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
