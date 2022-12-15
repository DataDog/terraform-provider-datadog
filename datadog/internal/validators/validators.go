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

func contains(values []string, value string) bool {
	for _, v := range values {
		if value == v {
			return true
		}
	}
	return false
}

func buildMessageString(values []string) string {
	stringValues := make([]string, 0)
	for _, v := range values {
		stringValues = append(stringValues, fmt.Sprintf("`%s`", v))
	}
	return strings.Join(stringValues, ", ")
}

func ValidateStringEnumValue(allowedValues ...interface{}) schema.SchemaValidateDiagFunc {
	allowedStringValues := make([]string, 0)
	for _, v := range allowedValues {
		allowedStringValues = append(allowedStringValues, fmt.Sprint(v))
	}

	return func(val interface{}, path cty.Path) diag.Diagnostics {
		var diags diag.Diagnostics

		if _, ok := val.(EnumChecker); ok {
			return append(diags, diag.Diagnostic{
				Severity:      diag.Warning,
				Summary:       "Allowed values",
				Detail:        buildMessageString(allowedStringValues),
				AttributePath: cty.Path{},
			})
		}

		stringVal, isString := val.(string)
		if !isString {
			return append(diags, diag.Diagnostic{
				Severity:      diag.Error,
				Summary:       "Invalid value type",
				Detail:        "Field value must be of type string",
				AttributePath: path,
			})
		}

		if !contains(allowedStringValues, stringVal) {
			return append(diags, diag.Diagnostic{
				Severity:      diag.Error,
				Summary:       "Invalid enum value",
				Detail:        fmt.Sprintf("Invalid value '%v': valid values are %v", val, allowedStringValues),
				AttributePath: path,
			})
		}
		return diags
	}
}

// ValidateEnumValue returns a validate func for a collection of enum value. It takes the constructors with validation for the enum as an argument.
// Such a constructor is for instance `datadogV1.NewWidgetLineWidthFromValue`
func ValidateEnumValue(newEnumFuncs ...interface{}) schema.SchemaValidateDiagFunc {

	// Get type of arg to convert int to int32/64 for instance
	f := make([]reflect.Type, len(newEnumFuncs))
	argT := make([]reflect.Type, len(newEnumFuncs))
	for idx, newEnumFunc := range newEnumFuncs {
		f[idx] = reflect.TypeOf(newEnumFunc)
		argT[idx] = f[idx].In(0)
	}

	return func(val interface{}, path cty.Path) diag.Diagnostics {
		var diags diag.Diagnostics

		// Hack to return a specific diagnostic containing the allowed enum values and have accurate docs
		msg := ""
		sep := ""
		if _, ok := val.(EnumChecker); ok {
			for _, fi := range f {
				enum := reflect.New(fi.Out(0)).Elem()
				validValues := enum.MethodByName("GetAllowedValues").Call([]reflect.Value{})[0]

				for i := 0; i < validValues.Len(); i++ {
					if len(msg) > 0 {
						sep = ", "
					}
					msg += fmt.Sprintf("%s`%v`", sep, validValues.Index(i).Interface())
				}
			}
			return append(diags, diag.Diagnostic{
				Severity:      diag.Warning,
				Summary:       "Allowed values",
				Detail:        msg,
				AttributePath: cty.Path{},
			})
		}

		arg := reflect.ValueOf(val)
		sep = ""
		nbErrors := 0
		for idx, newEnumFunc := range newEnumFuncs {
			outs := reflect.ValueOf(newEnumFunc).Call([]reflect.Value{arg.Convert(argT[idx])})
			if err := outs[1].Interface(); err != nil {
				if len(msg) > 0 {
					sep = ", "
				}
				msg += fmt.Sprintf("%s`%v`", sep, err.(error).Error())
				nbErrors++
			}
		}
		if nbErrors == len(newEnumFuncs) {
			return append(diags, diag.Diagnostic{
				Severity:      diag.Error,
				Summary:       "Invalid enum value",
				Detail:        msg,
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

func ValidateNonEmptyStringList(v interface{}, p cty.Path) diag.Diagnostics {
	var diags diag.Diagnostics
	values, isSlice := v.([]string)
	if !isSlice {
		return append(diags, diag.Diagnostic{
			Severity:      diag.Error,
			Summary:       "Value must be a string list",
			AttributePath: p,
		})
	}
	if len(values) == 0 {
		return append(diags, diag.Diagnostic{
			Severity:      diag.Error,
			Summary:       "List must contain at least one element",
			AttributePath: p,
		})
	}
	return diags
}
