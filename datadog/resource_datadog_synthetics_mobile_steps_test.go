package datadog

import (
	"context"
	"encoding/json"
	"reflect"
	"testing"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV1"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/stretchr/testify/require"
)

func mobileStepTestData(t *testing.T, raw map[string]interface{}) *schema.ResourceData {
	t.Helper()
	params := syntheticsMobileStepParams()
	return schema.TestResourceDataRaw(t, map[string]*schema.Schema{"params": &params}, map[string]interface{}{"params": []interface{}{raw}})
}

func buildMobileTestForStep(t *testing.T, kind string, params map[string]interface{}) (*schema.ResourceData, *datadogV1.SyntheticsMobileTest) {
	t.Helper()
	resource := resourceDatadogSyntheticsTest()
	raw := map[string]interface{}{
		"name": "Audit", "type": "mobile", "status": "paused", "locations": []interface{}{},
		"mobile_options_list": []interface{}{map[string]interface{}{
			"tick_every": 43200, "device_ids": []interface{}{"synthetics:mobile:device:apple_iphone_14_plus_ios_16"},
			"mobile_application": []interface{}{map[string]interface{}{"application_id": "app", "reference_id": "version", "reference_type": "version"}},
		}},
		"mobile_step": []interface{}{map[string]interface{}{"name": "Audit", "type": kind, "params": []interface{}{params}}},
	}
	config := terraform.NewResourceConfigRaw(raw)
	provider := &schema.Provider{ResourcesMap: map[string]*schema.Resource{"datadog_synthetics_test": resource}}
	if diags := provider.ValidateResource("datadog_synthetics_test", config); diags.HasError() {
		t.Fatalf("invalid input: %v", diags)
	}
	if _, err := resource.Diff(context.Background(), nil, config, nil); err != nil {
		t.Fatalf("plan error: %v", err)
	}
	data := schema.TestResourceDataRaw(t, resource.SchemaMap(), raw)
	built, diags := buildDatadogSyntheticsMobileTest(data)
	if diags.HasError() {
		t.Fatal(diags)
	}
	return data, built
}

func TestSyntheticsMobileSteps_Build(t *testing.T) {
	element := []interface{}{map[string]interface{}{"context": "NATIVE_APP", "context_type": "native", "text_content": "Input"}}
	cases := []struct {
		kind string
		raw  map[string]interface{}
	}{
		{"assertElementContent", map[string]interface{}{"element": element, "value": "hello", "check": "contains"}},
		{"assertScreenContains", map[string]interface{}{"value": "hello"}},
		{"assertScreenLacks", map[string]interface{}{"value": "hello"}},
		{"doubleTap", map[string]interface{}{"element": element}},
		{"extractVariable", map[string]interface{}{"element": element, "variable": []interface{}{map[string]interface{}{"name": "VALUE", "example": "example"}}}},
		{"flick", map[string]interface{}{"positions": []interface{}{map[string]interface{}{"x": 0.1, "y": 0.2}, map[string]interface{}{"x": 0.3, "y": 0.4}}}},
		{"openDeeplink", map[string]interface{}{"value": "example://hello"}},
		{"pressBack", map[string]interface{}{}},
		{"restartApplication", map[string]interface{}{}},
		{"rotate", map[string]interface{}{"value": "landscape"}},
		{"scroll", map[string]interface{}{"element": element, "x": 0.1, "y": 0.2}},
		{"scrollToElement", map[string]interface{}{"element": element, "direction": "down", "max_scrolls": 3}},
		{"tap", map[string]interface{}{"element": element}},
		{"toggleWiFi", map[string]interface{}{"enable": true}},
		{"toggleWiFi", map[string]interface{}{"enable": false}},
		{"typeText", map[string]interface{}{"element": element, "value": "hello", "delay": 125}},
		{"wait", map[string]interface{}{"value": "5"}},
	}
	for _, tc := range cases {
		t.Run(tc.kind, func(t *testing.T) {
			defer func() {
				if p := recover(); p != nil {
					t.Errorf("builder panic: %v", p)
				}
			}()
			data, built := buildMobileTestForStep(t, tc.kind, tc.raw)
			original := data.Get("mobile_step")
			params := built.GetSteps()[0].GetParams()
			if tc.kind == "toggleWiFi" {
				require.True(t, params.HasEnabled())
				require.Equal(t, tc.raw["enable"], params.GetEnabled())
			}
			if tc.kind == "wait" {
				require.NotNil(t, params.Value.SyntheticsMobileStepParamsValueNumber)
			}
			encoded, err := json.Marshal(built)
			require.NoError(t, err)
			var decoded datadogV1.SyntheticsMobileTest
			require.NoError(t, json.Unmarshal(encoded, &decoded))
			require.NoError(t, data.Set("mobile_step", buildTerraformMobileTestSteps(decoded.GetSteps())))
			if len(tc.raw) > 0 {
				require.Equal(t, original, data.Get("mobile_step"))
			}
			rebuilt, diags := buildDatadogSyntheticsMobileTest(data)
			require.False(t, diags.HasError())
			rebuiltJSON, err := json.Marshal(rebuilt)
			require.NoError(t, err)
			require.JSONEq(t, string(encoded), string(rebuiltJSON))
		})
	}
}

func TestSyntheticsMobileSteps_MissingElement(t *testing.T) {
	for _, kind := range []string{"assertElementContent", "doubleTap", "scroll", "scrollToElement", "tap", "typeText"} {
		t.Run(kind, func(t *testing.T) {
			defer func() {
				if p := recover(); p != nil {
					t.Errorf("omitted optional element panic: %v", p)
				}
			}()
			buildMobileTestForStep(t, kind, map[string]interface{}{"value": "hello"})
		})
	}
}

func TestSyntheticsMobileSteps_Read(t *testing.T) {
	cases := []struct {
		kind, wire, path string
		want             interface{}
	}{
		{"flick", `{"positions":[{"x":0.1,"y":0.2}]}`, "positions.0.x", 0.1},
		{"extractVariable", `{"variable":{"name":"VALUE","example":"example"}}`, "variable.0.name", "VALUE"},
		{"toggleWiFi", `{"enabled":true}`, "enable", true},
		{"toggleWiFi", `{"enabled":false}`, "enable", false},
		{"scrollToElement", `{"maxScrolls":3}`, "max_scrolls", 3},
		{"wait", `{"value":5}`, "value", "5"},
		{"typeText", `{"value":"hello","withEnter":true,"delay":125}`, "with_enter", true},
	}
	for _, tc := range cases {
		t.Run(tc.kind, func(t *testing.T) {
			defer func() {
				if p := recover(); p != nil {
					t.Errorf("read panic: %v", p)
				}
			}()
			var params datadogV1.SyntheticsMobileStepParams
			if err := json.Unmarshal([]byte(tc.wire), &params); err != nil {
				t.Fatal(err)
			}
			step := datadogV1.NewSyntheticsMobileStep("Audit", params, datadogV1.SyntheticsMobileStepType(tc.kind))
			flattened := buildTerraformMobileTestSteps([]datadogV1.SyntheticsMobileStep{*step})
			data := mobileStepTestData(t, map[string]interface{}{})
			if err := data.Set("params", flattened[0]["params"]); err != nil {
				t.Fatalf("state write error: %v", err)
			}
			actual := data.Get("params.0." + tc.path)
			if !reflect.DeepEqual(actual, tc.want) {
				t.Errorf("state %s = %#v; want %#v", tc.path, actual, tc.want)
			}
		})
	}
}

func TestSyntheticsMobileSteps_EmptyParams(t *testing.T) {
	for _, kind := range []string{"assertElementContent", "assertScreenContains", "assertScreenLacks", "doubleTap", "extractVariable", "flick", "openDeeplink", "playSubTest", "pressBack", "restartApplication", "rotate", "scroll", "scrollToElement", "tap", "toggleWiFi", "typeText", "wait"} {
		t.Run(kind, func(t *testing.T) {
			require.NotPanics(t, func() {
				_, built := buildMobileTestForStep(t, kind, map[string]interface{}{})
				encoded, err := json.Marshal(built.GetSteps()[0].GetParams())
				require.NoError(t, err)
				require.JSONEq(t, `{}`, string(encoded))
			})
		})
	}
}

func TestSyntheticsMobileSteps_OptionalBlocks(t *testing.T) {
	for _, tc := range []struct {
		name, kind string
		params     map[string]interface{}
	}{
		{"empty element", "tap", map[string]interface{}{"element": []interface{}{map[string]interface{}{}}}},
		{"empty user locator", "tap", map[string]interface{}{"element": []interface{}{map[string]interface{}{"user_locator": []interface{}{map[string]interface{}{}}}}}},
		{"empty locator value", "tap", map[string]interface{}{"element": []interface{}{map[string]interface{}{"user_locator": []interface{}{map[string]interface{}{"values": []interface{}{map[string]interface{}{}}}}}}}},
		{"empty relative position", "tap", map[string]interface{}{"element": []interface{}{map[string]interface{}{"relative_position": []interface{}{map[string]interface{}{}}}}}},
		{"empty position", "flick", map[string]interface{}{"positions": []interface{}{map[string]interface{}{}}}},
		{"zero position", "flick", map[string]interface{}{"positions": []interface{}{map[string]interface{}{"x": 0.0, "y": 0.0}}}},
		{"omitted variable", "extractVariable", map[string]interface{}{"element": []interface{}{map[string]interface{}{"text_content": "Input"}}}},
		{"omitted max scrolls", "scrollToElement", map[string]interface{}{"direction": "down"}},
		{"zero max scrolls", "scrollToElement", map[string]interface{}{"max_scrolls": 0}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			require.NotPanics(t, func() { buildMobileTestForStep(t, tc.kind, tc.params) })
		})
	}
}

func TestSyntheticsMobileSteps_InvalidWait(t *testing.T) {
	for _, value := range []string{"not-a-number", "1.5", "9223372036854775808"} {
		t.Run(value, func(t *testing.T) {
			data := schema.TestResourceDataRaw(t, resourceDatadogSyntheticsTest().SchemaMap(), map[string]interface{}{
				"mobile_step": []interface{}{map[string]interface{}{"name": "Wait", "type": "wait", "params": []interface{}{map[string]interface{}{"value": value}}}},
			})
			var built *datadogV1.SyntheticsMobileTest
			require.NotPanics(t, func() {
				var diags diag.Diagnostics
				built, diags = buildDatadogSyntheticsMobileTest(data)
				require.True(t, diags.HasError())
				require.Contains(t, diags[0].Summary, "mobile_step.0.params: value must be an integer")
				require.NotContains(t, diags[0].Summary, value)
			})
			require.Nil(t, built)
		})
	}
}

func TestSyntheticsMobileSteps_EmptyOptionBlocks(t *testing.T) {
	for _, field := range []string{"retry", "monitor_options", "bindings"} {
		t.Run(field, func(t *testing.T) {
			data := schema.TestResourceDataRaw(t, map[string]*schema.Schema{"mobile_options_list": syntheticsMobileTestOptionsList()}, map[string]interface{}{
				"mobile_options_list": []interface{}{map[string]interface{}{
					"tick_every": 43200,
					field:        []interface{}{map[string]interface{}{}},
				}},
			})
			require.NotPanics(t, func() { buildDatadogMobileTestOptions(data) })
		})
	}
}
