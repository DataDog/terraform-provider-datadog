package datadog

import (
	"fmt"
	"reflect"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func validateAggregatorMethod(v interface{}, k string) (ws []string, errors []error) {
	value := v.(string)
	validMethods := map[string]struct{}{
		"avg":   {},
		"max":   {},
		"min":   {},
		"sum":   {},
		"last":  {},
		"count": {},
	}
	if _, ok := validMethods[value]; !ok {
		errors = append(errors, fmt.Errorf(
			`%q contains an invalid method %q. Valid methods are either "avg", "max", "min", "sum", "count", or "last"`, k, value))
	}
	return
}

// validateEnumValue returns a validate func for an enum value. It takes the constructor with validation for the enum as an argument.
// Such a constructor is for instance `datadogV1.NewWidgetLineWidthFromValue`
func validateEnumValue(newEnumFunc interface{}) schema.SchemaValidateFunc {
	return func(val interface{}, key string) (warns []string, errs []error) {
		arg := reflect.ValueOf(val)
		outs := reflect.ValueOf(newEnumFunc).Call([]reflect.Value{arg})
		err := outs[1].Interface()
		if err != nil {
			errs = append(errs, err.(error))
		}
		return
	}
}
