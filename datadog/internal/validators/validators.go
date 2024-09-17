package validators

import (
	"context"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-framework-validators/helpers/validatordiag"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
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

// ValidateDatadogNonEmptyStrings ensures a string isn't empty
func ValidateNonEmptyStrings(v any, p cty.Path) diag.Diagnostics {
	value, ok := v.(string)
	var diags diag.Diagnostics
	if ok && value == "" {
		return append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Invalid value",
			Detail:   "Empty strings are not supported in this field.",
		})
	}
	return diags
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

// ValidateAWSAccountID AWS Account ID must be a string exactly 12 digits long
// See https://docs.aws.amazon.com/organizations/latest/APIReference/API_Account.html
func ValidateAWSAccountID(v any, p cty.Path) diag.Diagnostics {
	value, ok := v.(string)
	var diags diag.Diagnostics
	AWSAccountIDRegex := regexp.MustCompile(`^\d{12}$`)
	AWSIAMAAccessKeyRegex := regexp.MustCompile(`^(AKIA|ASIA)[A-Z0-9]{16,20}`)
	if ok && AWSIAMAAccessKeyRegex.MatchString(value) {
		// Help the user with a deprecation warning
		// Fedramp DD previously required using an IAM access key in place of the AWS account id
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Warning,
			Summary:  "Deprecated",
			Detail:   "the provided account ID might be an IAM access key. This behavior is deprecated. Use the AWS account ID instead.",
		})
	}
	if ok && !AWSAccountIDRegex.MatchString(value) {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Invalid value",
			Detail:   "account id must be a string containing exactly 12 digits",
		})
	}
	return diags
}

var _ schema.SchemaValidateDiagFunc = ValidateBasicEmail
var basicEmailRe = regexp.MustCompile("^[^@]+@[^@]+\\.[^@.]+$")

// ValidateBasicEmail ensures a string looks like an email
func ValidateBasicEmail(val any, path cty.Path) diag.Diagnostics {
	str, ok := val.(string)
	if !ok {
		return diag.Diagnostics{{
			Severity:      diag.Error,
			Summary:       fmt.Sprintf("not a string: %s", val),
			AttributePath: path,
		}}
	}
	if !basicEmailRe.MatchString(str) {
		return diag.Diagnostics{{
			Severity:      diag.Error,
			Summary:       fmt.Sprintf("not a email: %s", str),
			AttributePath: path,
		}}
	}
	return nil
}

type BetweenValidator struct {
	min float64
	max float64
}

func (v BetweenValidator) Description(ctx context.Context) string {
	return v.MarkdownDescription(ctx)
}

func (v BetweenValidator) MarkdownDescription(_ context.Context) string {
	return fmt.Sprintf("value must be between %f and %f", v.min, v.max)
}

func (v BetweenValidator) ValidateString(ctx context.Context, request validator.StringRequest, response *validator.StringResponse) {
	if request.ConfigValue.IsNull() || request.ConfigValue.IsUnknown() {
		return
	}

	value := request.ConfigValue.ValueString()

	fValue, err := strconv.ParseFloat(value, 64)

	if err != nil {
		response.Diagnostics.AddError(
			"value must be float",
			fmt.Sprintf("was %s", value),
		)
	}

	if fValue < v.min || fValue > v.max {
		response.Diagnostics.Append(validatordiag.InvalidAttributeValueDiagnostic(
			request.Path,
			v.Description(ctx),
			fmt.Sprintf("%f", fValue),
		))
	}
}

func Float64Between(min, max float64) validator.String {
	if min > max {
		return nil
	}

	return BetweenValidator{
		min: min,
		max: max,
	}
}

func ValidateHttpRequestHeader(v interface{}, k string) (ws []string, errors []error) {
	value := v.(map[string]interface{})
	for headerField, headerValue := range value {
		if !isValidToken(headerField) && !isRequestPseudoHeader(headerField) {
			errors = append(errors, fmt.Errorf("invalid value for %s (header field must be a valid token or a http/2 request pseudo-header)", k))
			return
		}
		headerStringValue, ok := headerValue.(string)
		if !ok {
			errors = append(errors, fmt.Errorf("expected type of %s to be string", k))
			return
		} else {
			for _, r := range headerStringValue {
				if (unicode.IsControl(r) && r != '\t') || (r == '\r' || r == '\n') {
					errors = append(errors, fmt.Errorf("invalid value for %s (header value must not contain invisible characters)", k))
					return
				}
			}
		}
	}
	return
}

func isRequestPseudoHeader(header string) bool {
	// :status is a response pseudo-header, and :protocol may only be used internally in websockets
	return header == ":method" || header == ":scheme" || header == ":authority" || header == ":path"
}

func isValidToken(token string) bool {
	for _, r := range token {
		if !isTokenChar(r) {
			return false
		}
	}
	return true
}

func isTokenChar(r rune) bool {
	if r >= '!' && r <= '~' && !strings.ContainsRune("()<>@,;:\\\"/[]?={} \t", r) {
		return true
	}
	return false
}
