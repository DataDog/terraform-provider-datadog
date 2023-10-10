package fwutils

import (
	"fmt"
	"reflect"
	"strings"

	frameworkSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

func EnrichFrameworkResourceSchema(s *frameworkSchema.Schema) {
	for i, attr := range s.Attributes {
		s.Attributes[i] = enrichDescription(attr)
	}

	for _, block := range s.Blocks {
		switch v := block.(type) {
		case frameworkSchema.ListNestedBlock:
			for i, attr := range v.NestedObject.Attributes {
				v.NestedObject.Attributes[i] = enrichDescription(attr)
			}
		case frameworkSchema.SingleNestedBlock:
			for i, attr := range v.Attributes {
				v.Attributes[i] = enrichDescription(attr)
			}
		case frameworkSchema.SetNestedBlock:
			for i, attr := range v.NestedObject.Attributes {
				v.NestedObject.Attributes[i] = enrichDescription(attr)
			}
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
	curentDesc := descField.String()

	// Build description with validators
	validators := rv.Elem().FieldByName("Validators")
	if validators.IsValid() && !validators.IsNil() && validators.Len() > 0 {
		for i := 0; i < validators.Len(); i++ {
			if strings.HasPrefix(validators.Index(i).Elem().Type().Name(), "enumValidator") {
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
				curentDesc = fmt.Sprintf("%s Valid values are %s.", curentDesc, validValuesMsg)
				break
			}
		}
	}

	// Build description with Defaults
	_default := rv.Elem().FieldByName("Default")
	if _default.IsValid() && !_default.IsNil() {
		defaultField := _default.Elem().FieldByName("defaultVal")
		if defaultField.IsValid() {
			curentDesc = fmt.Sprintf("%s Defaults to `%v`.", curentDesc, defaultField)
		}
	}

	descField.SetString(curentDesc)
}
