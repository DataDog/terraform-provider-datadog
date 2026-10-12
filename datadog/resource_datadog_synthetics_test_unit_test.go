package datadog

import (
	"encoding/json"
	"testing"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV1"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

func TestBuildDatadogParamsElementForMobileStep_Locators(t *testing.T) {
	multiLocator := map[string]interface{}{
		"ab": `/*[local-name()="XCUIElementTypeStaticText"][1]`,
		"co": `[{"tagName":"XCUIElementTypeStaticText","text":"tap","textType":"directText"}]`,
		"ro": `//*[@name="Tap"]`,
	}
	userLocator := []interface{}{map[string]interface{}{
		"fail_test_on_cannot_locate": true,
		"values": []interface{}{map[string]interface{}{
			"type": "id", "value": "some_id",
		}},
	}}

	for _, tc := range []struct {
		name            string
		multiLocator    map[string]interface{}
		withUserLocator bool
	}{
		{name: "multi locator only", multiLocator: multiLocator},
		{name: "both locators", multiLocator: multiLocator, withUserLocator: true},
		{name: "empty multi locator", multiLocator: map[string]interface{}{}, withUserLocator: true},
		{name: "user locator only", withUserLocator: true},
		{name: "neither locator"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			element := map[string]interface{}{"context": "NATIVE_APP", "context_type": "native"}
			if tc.multiLocator != nil {
				element["multi_locator"] = tc.multiLocator
			}
			if tc.withUserLocator {
				element["user_locator"] = userLocator
			}
			paramsSchema := syntheticsMobileStepParams()
			data := schema.TestResourceDataRaw(t, map[string]*schema.Schema{"params": &paramsSchema}, map[string]interface{}{
				"params": []interface{}{map[string]interface{}{"element": []interface{}{element}}},
			})
			params := data.Get("params").([]interface{})[0].(map[string]interface{})
			var built datadogV1.SyntheticsMobileStepParams
			require.NotPanics(t, func() {
				built = buildDatadogParamsForMobileStep(datadogV1.SYNTHETICSMOBILESTEPTYPE_TAP, params)
			})
			builtElement := built.GetElement()
			assert.Equal(t, len(tc.multiLocator) > 0, builtElement.HasMultiLocator())
			assert.Equal(t, tc.withUserLocator, builtElement.HasUserLocator())
			if len(tc.multiLocator) > 0 {
				assert.Equal(t, tc.multiLocator, builtElement.GetMultiLocator())
			}
			if tc.withUserLocator {
				locator := builtElement.GetUserLocator()
				assert.True(t, locator.GetFailTestOnCannotLocate())
				require.Len(t, locator.GetValues(), 1)
				assert.Equal(t, "some_id", locator.Values[0].GetValue())
				assert.Equal(t, "id", string(locator.Values[0].GetType()))
			}

			// Exercise the SDK wire representation and the API-to-Terraform mapping.
			step := datadogV1.NewSyntheticsMobileStep("Tap", built, datadogV1.SYNTHETICSMOBILESTEPTYPE_TAP)
			encoded, err := json.Marshal(step)
			require.NoError(t, err)
			var decoded datadogV1.SyntheticsMobileStep
			require.NoError(t, json.Unmarshal(encoded, &decoded))
			flattened := buildTerraformMobileTestSteps([]datadogV1.SyntheticsMobileStep{decoded})
			require.NoError(t, data.Set("params", flattened[0]["params"]))
			assert.Equal(t, params, data.Get("params").([]interface{})[0])
		})
	}
}

func TestBuildDatadogParamsForMobileStep_TypeTextDelay(t *testing.T) {
	for _, tc := range []struct {
		name      string
		delay     interface{}
		withEnter bool
	}{
		{name: "omitted"},
		{name: "zero", delay: 0},
		{name: "configured", delay: 125, withEnter: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			rawParams := map[string]interface{}{
				"value":      "hello",
				"with_enter": tc.withEnter,
				"element": []interface{}{map[string]interface{}{
					"context": "NATIVE_APP", "context_type": "native", "text_content": "Input",
				}},
			}
			if tc.delay != nil {
				rawParams["delay"] = tc.delay
			}
			paramsSchema := syntheticsMobileStepParams()
			data := schema.TestResourceDataRaw(t, map[string]*schema.Schema{"params": &paramsSchema}, map[string]interface{}{
				"params": []interface{}{rawParams},
			})
			params := data.Get("params").([]interface{})[0].(map[string]interface{})
			delay := params["delay"].(int)
			var built datadogV1.SyntheticsMobileStepParams
			require.NotPanics(t, func() {
				built = buildDatadogParamsForMobileStep(datadogV1.SYNTHETICSMOBILESTEPTYPE_TYPETEXT, params)
			})
			assert.Equal(t, delay != 0, built.HasDelay())
			assert.Equal(t, int64(delay), built.GetDelay())

			encoded, err := json.Marshal(built)
			require.NoError(t, err)
			var wire map[string]interface{}
			require.NoError(t, json.Unmarshal(encoded, &wire))
			assert.Equal(t, "hello", wire["value"])
			assert.Equal(t, tc.withEnter, wire["withEnter"])
			if delay == 0 {
				assert.NotContains(t, wire, "delay")
			} else {
				assert.Equal(t, float64(delay), wire["delay"])
			}

			var decoded datadogV1.SyntheticsMobileStepParams
			require.NoError(t, json.Unmarshal(encoded, &decoded))
			step := datadogV1.NewSyntheticsMobileStep("Type text", decoded, datadogV1.SYNTHETICSMOBILESTEPTYPE_TYPETEXT)
			flattened := buildTerraformMobileTestSteps([]datadogV1.SyntheticsMobileStep{*step})
			require.NoError(t, data.Set("params", flattened[0]["params"]))
			assert.Equal(t, params, data.Get("params").([]interface{})[0])
		})
	}
}
