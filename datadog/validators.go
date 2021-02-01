package datadog

import (
	"fmt"
	"reflect"
	"regexp"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func validateFloatString(v interface{}, k string) (ws []string, errors []error) {
	return validation.StringMatch(regexp.MustCompile("\\d*(\\.\\d*)?"), "value must be a float")(v, k)
}

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

	// Get type of arg to convert int to int32/64 for instance
	f := reflect.TypeOf(newEnumFunc)
	argT := f.In(0)

	return func(val interface{}, key string) (warns []string, errs []error) {
		arg := reflect.ValueOf(val)
		outs := reflect.ValueOf(newEnumFunc).Call([]reflect.Value{arg.Convert(argT)})
		if err := outs[1].Interface(); err != nil {
			errs = append(errs, err.(error))
		}
		return
	}
}
