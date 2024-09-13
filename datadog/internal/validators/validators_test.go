package validators

import (
	"testing"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
)

func TestResourceDatadogDowntimeRecurrenceTypeValidation(t *testing.T) {
	cases := []struct {
		Value    string
		ErrCount int
	}{
		{
			Value:    "daily",
			ErrCount: 1,
		},
		{
			Value:    "days",
			ErrCount: 0,
		},
		{
			Value:    "days,weeks",
			ErrCount: 1,
		},
		{
			Value:    "months",
			ErrCount: 0,
		},
		{
			Value:    "years",
			ErrCount: 0,
		},
		{
			Value:    "weeks",
			ErrCount: 0,
		},
	}

	for _, tc := range cases {
		_, errors := ValidateDatadogDowntimeRecurrenceType(tc.Value, "datadog_downtime_recurrence_type")

		if len(errors) != tc.ErrCount {
			t.Fatalf("Expected Datadog Downtime Recurrence Type validation to trigger %d error(s) for value %q - instead saw %d",
				tc.ErrCount, tc.Value, len(errors))
		}
	}
}

func TestResourceDatadogDowntimeRecurrenceWeekDaysValidation(t *testing.T) {
	cases := []struct {
		Value    string
		ErrCount int
	}{
		{
			Value:    "Mon",
			ErrCount: 0,
		},
		{
			Value:    "Mon,",
			ErrCount: 1,
		},
		{
			Value:    "Monday",
			ErrCount: 1,
		},
		{
			Value:    "mon",
			ErrCount: 1,
		},
		{
			Value:    "mon,",
			ErrCount: 1,
		},
		{
			Value:    "monday",
			ErrCount: 1,
		},
		{
			Value:    "mon,Tue",
			ErrCount: 1,
		},
		{
			Value:    "Mon,tue",
			ErrCount: 1,
		},
		{
			Value:    "Mon,Tue",
			ErrCount: 1,
		},
		{
			Value:    "Mon, Tue",
			ErrCount: 1,
		},
	}

	for _, tc := range cases {
		_, errors := ValidateDatadogDowntimeRecurrenceWeekDays(tc.Value, "datadog_downtime_recurrence_week_days")

		if len(errors) != tc.ErrCount {
			t.Fatalf("Expected Datadog Downtime Recurrence Week Days validation to trigger %d error(s) for value %q - instead saw %d",
				tc.ErrCount, tc.Value, len(errors))
		}
	}
}

func TestStringEnumValidation(t *testing.T) {
	cases := []struct {
		InputValue    interface{}
		ExpectedError *diag.Diagnostic
	}{
		{
			InputValue:    "log_detection",
			ExpectedError: nil,
		},
		{
			InputValue:    "signal_correlation",
			ExpectedError: nil,
		},
		{
			InputValue:    "thirdValue",
			ExpectedError: nil,
		},
		{
			InputValue: "Mon",
			ExpectedError: &diag.Diagnostic{
				Severity:      diag.Error,
				Summary:       "Invalid enum value",
				Detail:        "Invalid value 'Mon': valid values are [log_detection signal_correlation thirdValue]",
				AttributePath: cty.Path{},
			},
		},
		{
			InputValue: [2]string{"one", "two"},
			ExpectedError: &diag.Diagnostic{
				Severity:      diag.Error,
				Summary:       "Invalid value type",
				Detail:        "Field value must be of type string",
				AttributePath: cty.Path{},
			},
		},
		{
			InputValue: EnumChecker{},
			ExpectedError: &diag.Diagnostic{
				Severity:      diag.Warning,
				Summary:       "Allowed values",
				Detail:        "`log_detection`, `signal_correlation`, `thirdValue`",
				AttributePath: cty.Path{},
			},
		},
	}

	validator := ValidateStringEnumValue(datadogV2.SECURITYMONITORINGRULETYPEREAD_LOG_DETECTION,
		datadogV2.SECURITYMONITORINGSIGNALRULETYPE_SIGNAL_CORRELATION, "thirdValue")

	for _, tc := range cases {
		var diags diag.Diagnostics = validator(tc.InputValue, cty.Path{})

		if tc.ExpectedError == nil && len(diags) != 0 {
			t.Fatalf("Expected no diagnostics for input %v, found %d instead", tc.InputValue, len(diags))
		}
		if tc.ExpectedError != nil && len(diags) > 1 {
			t.Fatalf("Expected one diagnostic for input %v, found %d instead", tc.InputValue, len(diags))
		}
		if tc.ExpectedError != nil && !areEqual(diags[0], *(tc.ExpectedError)) {
			t.Fatalf("Expected %v for input %v, found %v instead", diags[0], tc.InputValue, *(tc.ExpectedError))
		}
	}
}

func areEqual(actual diag.Diagnostic, expected diag.Diagnostic) bool {
	return actual.Detail == expected.Detail && actual.Severity == expected.Severity && actual.Summary == expected.Summary
}

func TestValidateFloat64Between(t *testing.T) {
	cases := []struct {
		InputValue    string
		ExpectedError bool
	}{

		{
			InputValue:    "-0.1",
			ExpectedError: true,
		},
		{
			InputValue:    "0",
			ExpectedError: true,
		},
		{
			InputValue:    "0.1",
			ExpectedError: false,
		},
		{
			InputValue:    "0.5",
			ExpectedError: false,
		},
		{
			InputValue:    "1.1",
			ExpectedError: true,
		},
		{
			InputValue:    "Mon",
			ExpectedError: true,
		},
	}

	for _, tc := range cases {
		validationResult := validator.StringResponse{}
		Float64Between(0.1, 1).ValidateString(nil, validator.StringRequest{ConfigValue: basetypes.NewStringValue(tc.InputValue)}, &validationResult)
		if tc.ExpectedError == false && len(validationResult.Diagnostics) != 0 {
			t.Fatalf("Expected no diagnostics for input %v, found %d instead", tc.InputValue, len(validationResult.Diagnostics))
		}
	}
}

func TestValidateHttpRequestHeader(t *testing.T) {
	cases := []struct {
		Value    map[string]interface{}
		ErrCount int
	}{
		{
			Value:    map[string]interface{}{"foo": "bar"},
			ErrCount: 0,
		},
		{
			Value:    map[string]interface{}{"content-length": "1"},
			ErrCount: 0,
		},
		{
			Value:    map[string]interface{}{":authority": "example.com"},
			ErrCount: 0,
		},
		{
			Value:    map[string]interface{}{"foo": "\t"},
			ErrCount: 0,
		},
		{
			Value:    map[string]interface{}{"not?token": "any"},
			ErrCount: 1,
		},
		{
			Value:    map[string]interface{}{":unknown": "any"},
			ErrCount: 1,
		},
		{
			Value:    map[string]interface{}{"foo": "\r"},
			ErrCount: 1,
		},
		{
			Value:    map[string]interface{}{"foo": "\n"},
			ErrCount: 1,
		},
	}

	for _, tc := range cases {
		_, errors := ValidateHttpRequestHeader(tc.Value, "request_headers")

		if len(errors) != tc.ErrCount {
			t.Fatalf("Expected http request header validation to trigger %d error(s) for value %q - instead saw %d",
				tc.ErrCount, tc.Value, len(errors))
		}
	}
}
