package fwutils

import (
	"fmt"
	"reflect"

	frameworkSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

func EnrichFrameworkResourceSchema(s frameworkSchema.Schema) frameworkSchema.Schema {
	for i, attr := range s.Attributes {
		s.Attributes[i] = updateDescription(attr)
	}

	return s
}

func updateDescription(r any) frameworkSchema.Attribute {
	switch v := r.(type) {
	case frameworkSchema.StringAttribute:
		v.Description = getUpdatedDescriptionWithValidators(v.Description, reflect.ValueOf(v.Validators))
		return v
	case frameworkSchema.Int64Attribute:
		v.Description = getUpdatedDescriptionWithValidators(v.Description, reflect.ValueOf(v.Validators))
		return v
	default:
		return r.(frameworkSchema.Attribute)
	}
}

func getUpdatedDescriptionWithValidators(description string, validators reflect.Value) string {
	if validators.Len() == 0 {
		return description
	}

	for i := 0; i < validators.Len(); i++ {
		allowedValues := validators.Index(i).Elem().FieldByName("AllowedEnumValues")
		if allowedValues.IsValid() {
			v := reflect.ValueOf(allowedValues.Interface())
			validValuesMsg := ""
			sep := ""
			for i := 0; i < v.Len(); i++ {
				if len(validValuesMsg) > 0 {
					sep = ", "
				}
				validValuesMsg += fmt.Sprintf("%s`%v`", sep, v.Index(i).Interface())
			}
			description = fmt.Sprintf("%s Valid values are %s.", description, validValuesMsg)
			break
		}
	}

	return description
}
