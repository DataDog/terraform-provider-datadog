package datadog

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV1"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSyntheticsGlobalConfigVariableSecure(t *testing.T) {
	for _, configured := range []interface{}{true, false, nil} {
		for _, returned := range []interface{}{true, false, nil} {
			t.Run(fmt.Sprintf("configured=%v/api=%v", configured, returned), func(t *testing.T) {
				raw := map[string]interface{}{"name": "GLOBAL_VAR", "type": "global", "id": "global-id"}
				if configured != nil {
					raw["secure"] = configured
				}
				data := schema.TestResourceDataRaw(t, map[string]*schema.Schema{"config_variable": syntheticsConfigVariable()}, map[string]interface{}{
					"config_variable": []interface{}{raw},
				})
				original := data.Get("config_variable").([]interface{})
				for cycle := 0; cycle < 3; cycle++ {
					variables := buildDatadogConfigVariables(data.Get("config_variable").([]interface{}))
					require.Len(t, variables, 1)
					assert.False(t, variables[0].HasSecure(), "global reference must not send secure")
					if returned != nil {
						variables[0].SetSecure(returned.(bool))
					}
					require.NoError(t, data.Set("config_variable", buildTerraformConfigVariables(variables, data.Get("config_variable").([]interface{}))))
					assert.Equal(t, original, data.Get("config_variable"), "cycle %d", cycle)
				}

				// Import has no prior configuration and cannot recover the reference flag.
				imported := schema.TestResourceDataRaw(t, map[string]*schema.Schema{"config_variable": syntheticsConfigVariable()}, nil)
				variables := buildDatadogConfigVariables(original)
				if returned != nil {
					variables[0].SetSecure(returned.(bool))
				}
				require.NoError(t, imported.Set("config_variable", buildTerraformConfigVariables(variables, nil)))
				assert.Equal(t, false, imported.Get("config_variable.0.secure"))
			})
		}
	}
}

func TestSyntheticsGlobalConfigVariableSecureIdentity(t *testing.T) {
	old := []interface{}{
		map[string]interface{}{"name": "FIRST_VAR", "type": "global", "id": "first-id", "secure": true},
		map[string]interface{}{"name": "SECOND_VAR", "type": "global", "id": "second-id", "secure": false},
	}
	for _, tc := range []struct {
		name, id, variableType string
		secure                 bool
	}{
		{"SECOND_VAR", "second-id", "global", false},
		{"FIRST_VAR", "first-id", "global", true},
		{"FIRST_VAR", "replacement-id", "global", false},
		{"RENAMED_VAR", "first-id", "global", false},
		{"FIRST_VAR", "first-id", "text", false},
	} {
		t.Run(tc.name+"/"+tc.id+"/"+tc.variableType, func(t *testing.T) {
			variable := datadogV1.SyntheticsConfigVariable{}
			variable.SetName(tc.name)
			variable.SetId(tc.id)
			variable.SetType(datadogV1.SyntheticsConfigVariableType(tc.variableType))
			assert.Equal(t, tc.secure, buildTerraformConfigVariables([]datadogV1.SyntheticsConfigVariable{variable}, old)[0]["secure"] == true)
		})
	}
}

func TestSyntheticsLocalConfigVariableSecure(t *testing.T) {
	for _, secure := range []bool{true, false} {
		t.Run(fmt.Sprint(secure), func(t *testing.T) {
			data := schema.TestResourceDataRaw(t, map[string]*schema.Schema{"config_variable": syntheticsConfigVariable()}, map[string]interface{}{
				"config_variable": []interface{}{map[string]interface{}{
					"name": "LOCAL_VAR", "type": "text", "secure": secure, "example": "example", "pattern": "pattern",
				}},
			})
			original := data.Get("config_variable").([]interface{})
			variables := buildDatadogConfigVariables(original)
			assert.True(t, variables[0].HasSecure())
			assert.Equal(t, secure, variables[0].GetSecure())
			if secure {
				variables[0].Example = nil
				variables[0].Pattern = nil
			}
			for cycle := 0; cycle < 3; cycle++ {
				require.NoError(t, data.Set("config_variable", buildTerraformConfigVariables(variables, data.Get("config_variable").([]interface{}))))
				assert.Equal(t, original, data.Get("config_variable"))
			}
			// Unlike global references, local variables must reflect API security changes.
			variables[0].SetSecure(!secure)
			flattened := buildTerraformConfigVariables(variables, original)
			assert.Equal(t, !secure, flattened[0]["secure"])
			imported := buildTerraformConfigVariables(variables, nil)
			assert.Equal(t, !secure, imported[0]["secure"])
		})
	}
}

func TestSyntheticsExtractedValueSecure(t *testing.T) {
	for _, secure := range []bool{true, false} {
		values := []interface{}{map[string]interface{}{
			"name": "EXTRACTED_VAR", "type": "http_body", "field": "", "secure": secure,
			"parser": []interface{}{map[string]interface{}{"type": "raw", "value": ""}},
		}}
		built := buildDatadogExtractedValues(values)
		assert.True(t, built[0].HasSecure())
		assert.Equal(t, secure, built[0].GetSecure())
		assert.Equal(t, secure, buildTerraformExtractedValues(built)[0]["secure"])
	}
}

func TestSyntheticsMobileGlobalConfigVariableSecure(t *testing.T) {
	for _, secure := range []interface{}{true, false, nil} {
		t.Run(fmt.Sprint(secure), func(t *testing.T) {
			variable := map[string]interface{}{"name": "GLOBAL_VAR", "type": "global", "id": "global-id"}
			if secure != nil {
				variable["secure"] = secure
			}
			data := schema.TestResourceDataRaw(t, resourceDatadogSyntheticsTest().SchemaFunc(), map[string]interface{}{
				"type": "mobile", "config_variable": []interface{}{variable},
			})
			original := data.Get("config_variable")
			for cycle := 0; cycle < 3; cycle++ {
				built := buildDatadogSyntheticsMobileTest(data)
				diags := updateSyntheticsMobileTestLocalState(data, built)
				require.False(t, diags.HasError(), "%v", diags)
				assert.Equal(t, original, data.Get("config_variable"))
			}
		})
	}
}

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
