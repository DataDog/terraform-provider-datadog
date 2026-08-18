package datadog

import (
	"testing"

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
