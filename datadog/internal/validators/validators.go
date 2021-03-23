package validators

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

// ValidateFloatString makes sure a string can be parsed into a float
func ValidateFloatString(v interface{}, k string) (ws []string, errors []error) {
	return validation.StringMatch(regexp.MustCompile("\\d*(\\.\\d*)?"), "value must be a float")(v, k)
}

// ValidateAggregatorMethod ensures a string is a valid aggregator method
func ValidateAggregatorMethod(v interface{}, k string) (ws []string, errors []error) {
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

// ValidateEnumValue returns a validate func for an enum value. It takes the constructor with validation for the enum as an argument.
// Such a constructor is for instance `datadogV1.NewWidgetLineWidthFromValue`
func ValidateEnumValue(newEnumFunc interface{}) schema.SchemaValidateFunc {

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

// ValidateDatadogDowntimeRecurrenceType ensures a string is a valid recurrence type
func ValidateDatadogDowntimeRecurrenceType(v interface{}, k string) (ws []string, errors []error) {
	value := v.(string)
	switch value {
	case "days", "months", "weeks", "years", "rrule":
		break
	default:
		errors = append(errors, fmt.Errorf(
			"%q contains an invalid recurrence type parameter %q. Valid parameters are days, months, weeks, years, or rrule", k, value))
	}
	return
}

// ValidateDatadogDowntimeRecurrenceWeekDays ensures a string is a valid recurrence week day
func ValidateDatadogDowntimeRecurrenceWeekDays(v interface{}, k string) (ws []string, errors []error) {
	value := v.(string)
	switch value {
	case "Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun":
		break
	default:
		errors = append(errors, fmt.Errorf(
			"%q contains an invalid recurrence week day parameter %q. Valid parameters are Mon, Tue, Wed, Thu, Fri, Sat, or Sun", k, value))
	}
	return
}

// ValidateDatadogDowntimeTimezone ensures a string is a valid timezone
func ValidateDatadogDowntimeTimezone(v interface{}, k string) (ws []string, errors []error) {
	value := v.(string)
	switch strings.ToLower(value) {
	case "utc", "":
		break
	case "local", "localtime":
		// get current zone from machine
		zone, _ := time.Now().Local().Zone()
		return ValidateDatadogDowntimeRecurrenceType(zone, k)
	default:
		_, err := time.LoadLocation(value)
		if err != nil {
			errors = append(errors, fmt.Errorf(
				"%q contains an invalid timezone parameter: %q, Valid parameters are IANA Time Zone names",
				k, value))
		}
	}
	return
}
