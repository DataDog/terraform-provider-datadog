package validators

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"time"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

// ValidateFloatString makes sure a string can be parsed into a float
func ValidateFloatString(v interface{}, k string) (ws []string, errors []error) {
	return validation.StringMatch(regexp.MustCompile(`\d*(\.\d*)?`), "value must be a float")(v, k)
}

// EnumChecker type to get allowed enum values from validate func
type EnumChecker struct{}

// ValidateEnumValue returns a validate func for an enum value. It takes the constructor with validation for the enum as an argument.
// Such a constructor is for instance `datadogV1.NewWidgetLineWidthFromValue`
func ValidateEnumValue(newEnumFunc interface{}) schema.SchemaValidateDiagFunc {

	// Get type of arg to convert int to int32/64 for instance
	f := reflect.TypeOf(newEnumFunc)
	argT := f.In(0)

	return func(val interface{}, path cty.Path) diag.Diagnostics {
		var diags diag.Diagnostics

		// Hack to return a specific diagnostic containing the allowed enum values and have accurate docs
		if _, ok := val.(EnumChecker); ok {
			enum := reflect.New(f.Out(0)).Elem()
			validValues := enum.MethodByName("GetAllowedValues").Call([]reflect.Value{})[0]
			msg := ""
			sep := ", "
			for i := 0; i < validValues.Len(); i++ {
				if i == validValues.Len()-1 {
					sep = ""
				}
				msg += fmt.Sprintf("`%v`%s", validValues.Index(i).Interface(), sep)
			}

			return append(diags, diag.Diagnostic{
				Severity:      diag.Warning,
				Summary:       "Allowed values",
				Detail:        msg,
				AttributePath: cty.Path{},
			})
		}

		arg := reflect.ValueOf(val)
		outs := reflect.ValueOf(newEnumFunc).Call([]reflect.Value{arg.Convert(argT)})
		if err := outs[1].Interface(); err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity:      diag.Error,
				Summary:       "Invalid enum value",
				Detail:        err.(error).Error(),
				AttributePath: path,
			})
		}
		return diags
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
