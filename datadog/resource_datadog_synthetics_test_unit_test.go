package datadog

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
)

// TestConvertStepParamsValueForConfig_EmptyVariableOrPattern reproduces the
// panic reported by a customer: when an Optional/MaxItems(1) list-typed step
// param such as "variable" or "pattern" is left unset in the config,
// Terraform passes it through as an empty []interface{} rather than "". The
// function used to blindly index [0] into it, causing an
// "index out of range [0] with length 0" panic.
func TestConvertStepParamsValueForConfig_EmptyVariableOrPattern(t *testing.T) {
	for _, key := range []string{"variable", "pattern"} {
		t.Run(key, func(t *testing.T) {
			assert.NotPanics(t, func() {
				result, diags := convertStepParamsValueForConfig(nil, key, []interface{}{})
				assert.Nil(t, result)
				assert.Empty(t, diags)
			})
		})
	}
}

func TestConvertStepParamsValueForConfig_NonEmptyVariableOrPattern(t *testing.T) {
	value := []interface{}{map[string]interface{}{"name": "foo"}}
	result, diags := convertStepParamsValueForConfig(nil, "variable", value)
	assert.Equal(t, value[0], result)
	assert.Empty(t, diags)
}

// The synthetics API caps API-test request timeout at 60s ("Value of 'timeout'
// should be less than 60"). The provider must reject out-of-range values at
// plan time instead of letting the apply fail with the API's opaque
// "'request' value ... is invalid" 400.
func TestSyntheticsTestRequestTimeoutValidation(t *testing.T) {
	attr, ok := syntheticsTestRequest().Schema["timeout"]
	assert.True(t, ok, "timeout attribute missing from request schema")
	var vf schema.SchemaValidateFunc = attr.ValidateFunc

	for _, tc := range []struct {
		value   interface{}
		invalid bool
	}{
		{0, false},
		{60, false},
		{61, true},
		{120, true},
	} {
		_, errs := vf(tc.value, "timeout")
		if tc.invalid {
			assert.NotEmpty(t, errs, "expected %v to be rejected", tc.value)
		} else {
			assert.Empty(t, errs, "expected %v to be accepted", tc.value)
		}
	}
}
