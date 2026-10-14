package datadog

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	datadogClient "github.com/DataDog/datadog-api-client-go/v2/api/datadog"
	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV1"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
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

// Exercise actual Terraform apply, refresh, import and planning against a local
// API. The mock stores only API request data, so it cannot return a targetValue
// that the provider omits for isUndefined.
func TestSyntheticsUndefinedTargetValueLifecycle(t *testing.T) {
	for _, assertionType := range []string{"body", "grpcProto"} {
		for _, value := range []string{"targetvalue = false", "", "targetvalue = null"} {
			t.Run(assertionType+"/"+value, func(t *testing.T) {
				t.Parallel()
				factory, restore := syntheticsDriftProvider(t)
				config := func(operator, path, target, name string) string {
					subtype, request := "http", `method = "GET"
url = "https://example.com"`
					if assertionType == "grpcProto" {
						subtype, request = "grpc", `host = "example.com"
port = 50051`
					}
					return fmt.Sprintf(`resource "datadog_synthetics_test" "test" {
 type = "api"
 subtype = %q
 name = %q
 message = ""
 status = "paused"
 locations = ["aws:eu-central-1"]
 request_definition {
 %s
 }
 options_list { tick_every = 60 }
 assertion {
  type = %q
  operator = "validatesJSONPath"
  targetjsonpath {
   jsonpath = %q
   operator = %q
   %s
  }
 }
}`, subtype, name, request, assertionType, path, operator, target)
				}
				check := func(operator, path, target string) resource.TestCheckFunc {
					return resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("datadog_synthetics_test.test", "assertion.0.targetjsonpath.0.operator", operator),
						resource.TestCheckResourceAttr("datadog_synthetics_test.test", "assertion.0.targetjsonpath.0.jsonpath", path),
						resource.TestCheckResourceAttr("datadog_synthetics_test.test", "assertion.0.targetjsonpath.0.targetvalue", target),
						resource.TestCheckResourceAttr("datadog_synthetics_test.test", "status", "paused"),
					)
				}
				expected := ""
				var importIgnore []string
				if value == "targetvalue = false" {
					expected = "false"
					importIgnore = []string{"assertion.0.targetjsonpath.0.targetvalue"}
				}
				initial := config("isUndefined", "$.value", value, "initial")
				changed := config("isUndefined", "$.value", value, "updated")
				comparison := config("is", "$.other", `targetvalue = "expected"`, "updated")
				numeric := config("moreThan", "$.other", `targetvalue = "42"`, "updated")
				steps := []resource.TestStep{
					{Config: initial, Check: check("isUndefined", "$.value", expected)},
					{Config: initial, PlanOnly: true},
					{Config: changed, Check: check("isUndefined", "$.value", expected)},
					{Config: changed, PlanOnly: true},
					{Config: comparison, Check: check("is", "$.other", "expected")},
					{Config: comparison, PlanOnly: true},
					{Config: numeric, Check: check("moreThan", "$.other", "42")},
					{Config: numeric, PlanOnly: true},
					{Config: changed, Check: check("isUndefined", "$.value", expected)},
					{Config: changed, PlanOnly: true},
					// Verify a fresh import; the API cannot supply the ignored local value.
					{ResourceName: "datadog_synthetics_test.test", ImportState: true, ImportStateVerify: true, ImportStateVerifyIgnore: importIgnore},
					{Config: changed, Check: check("isUndefined", "$.value", expected)},
					{Config: changed, PlanOnly: true},
					{Config: config("isUndefined", "$.value", "targetvalue = null", "updated"), Check: check("isUndefined", "$.value", "")},
					{Config: config("isUndefined", "$.value", "", "updated"), PlanOnly: true},
				}
				t.Run("lifecycle", func(t *testing.T) {
					resource.UnitTest(t, resource.TestCase{ProviderFactories: map[string]func() (*schema.Provider, error){"datadog": factory}, Steps: steps})
				})
				t.Run("import", func(t *testing.T) {
					// Seed the mock with the deleted remote fixture, then import into an
					// independent empty state. A single apply restores any local-only value.
					restore()
					resource.UnitTest(t, resource.TestCase{
						ProviderFactories: map[string]func() (*schema.Provider, error){"datadog": factory},
						Steps: []resource.TestStep{
							{Config: changed, ResourceName: "datadog_synthetics_test.test", ImportStateId: "abc-def-ghi", ImportState: true, ImportStatePersist: true},
							{Config: changed, Check: check("isUndefined", "$.value", expected)},
							{Config: changed, PlanOnly: true},
							{Config: changed, PlanOnly: true},
						},
					})
				})
			})
		}
	}
}

func syntheticsDriftProvider(t *testing.T) (func() (*schema.Provider, error), func()) {
	t.Helper()
	var mu sync.Mutex
	var stored, deleted map[string]interface{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		switch r.Method {
		case http.MethodPost, http.MethodPut:
			if r.URL.Path == "/api/v1/synthetics/tests/delete" {
				deleted, stored = stored, nil
				_, _ = w.Write([]byte(`{"deleted_tests":[]}`))
				return
			}
			var body map[string]interface{}
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Errorf("decode API request: %v", err)
				http.Error(w, "invalid request", http.StatusBadRequest)
				return
			}
			assertions := body["config"].(map[string]interface{})["assertions"].([]interface{})
			for _, assertion := range assertions {
				target := assertion.(map[string]interface{})["target"].(map[string]interface{})
				if target["operator"] == "isUndefined" {
					assert.NotContains(t, target, "targetValue")
				} else if target["operator"] == "moreThan" {
					assert.Equal(t, float64(42), target["targetValue"])
				} else {
					assert.Equal(t, "expected", target["targetValue"])
				}
			}
			stored = body
			stored["public_id"] = "abc-def-ghi"
			stored["monitor_id"] = 123
		case http.MethodGet:
			if stored == nil {
				http.NotFound(w, r)
				return
			}
		default:
			t.Errorf("unexpected API request: %s %s", r.Method, r.URL.Path)
		}
		assert.NoError(t, json.NewEncoder(w).Encode(stored))
	}))
	t.Cleanup(server.Close)
	cfg := datadogClient.NewConfiguration()
	cfg.Servers = datadogClient.ServerConfigurations{{URL: server.URL}}
	cfg.OperationServers = nil
	cfg.HTTPClient = server.Client()
	meta := &ProviderConfiguration{Auth: context.Background(), DatadogApiInstances: &utils.ApiInstances{HttpClient: datadogClient.NewAPIClient(cfg)}}
	return func() (*schema.Provider, error) {
			return &schema.Provider{
				ResourcesMap:         map[string]*schema.Resource{"datadog_synthetics_test": resourceDatadogSyntheticsTest()},
				ConfigureContextFunc: func(context.Context, *schema.ResourceData) (interface{}, diag.Diagnostics) { return meta, nil },
			}, nil
		}, func() {
			mu.Lock()
			defer mu.Unlock()
			stored = deleted
		}
}

func TestSyntheticsUndefinedTargetValueRefreshIdentity(t *testing.T) {
	for _, tc := range []struct {
		name, assertionType, operator, path, property, target, expected string
		importing                                                       bool
	}{
		{name: "unchanged", assertionType: "body", operator: "isUndefined", path: "$.value", expected: "false"},
		{name: "import", assertionType: "body", operator: "isUndefined", path: "$.value", importing: true},
		{name: "remote path change", assertionType: "body", operator: "isUndefined", path: "$.other"},
		{name: "remote type change", assertionType: "grpcProto", operator: "isUndefined", path: "$.value"},
		{name: "remote property change", assertionType: "body", operator: "isUndefined", path: "$.value", property: "other"},
		{name: "remote string comparison", assertionType: "body", operator: "is", path: "$.value", target: "remote", expected: "remote"},
		{name: "remote numeric comparison", assertionType: "body", operator: "moreThan", path: "$.value", target: "42", expected: "42"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			assertionSchema := map[string]*schema.Schema{"assertion": syntheticsAPIAssertion()}
			raw := func(assertionType, operator, path, property, value string) map[string]interface{} {
				return map[string]interface{}{"assertion": []interface{}{map[string]interface{}{
					"type": assertionType, "operator": "validatesJSONPath", "property": property,
					"targetjsonpath": []interface{}{map[string]interface{}{"jsonpath": path, "operator": operator, "targetvalue": value}},
				}}}
			}
			prior := schema.TestResourceDataRaw(t, assertionSchema, raw("body", "isUndefined", "$.value", "", "false"))
			remote := schema.TestResourceDataRaw(t, assertionSchema, raw(tc.assertionType, tc.operator, tc.path, tc.property, tc.target))
			assertions, diags := buildDatadogAssertions(remote.Get("assertion").([]interface{}))
			require.False(t, diags.HasError(), "%v", diags)
			flattened, diags := buildTerraformAssertions(assertions, nil)
			require.False(t, diags.HasError(), "%v", diags)
			expected, err := json.Marshal(flattened)
			require.NoError(t, err)
			var expectedFields []map[string]interface{}
			require.NoError(t, json.Unmarshal(expected, &expectedFields))
			previous := prior.Get("assertion").([]interface{})
			if tc.importing {
				previous = nil
			}
			flattened, diags = buildTerraformAssertions(assertions, previous)
			require.False(t, diags.HasError(), "%v", diags)
			target := flattened[0]["targetjsonpath"].([]map[string]string)[0]
			require.Equal(t, tc.expected, target["targetvalue"])
			require.Equal(t, tc.operator, target["operator"])
			require.Equal(t, tc.path, target["jsonpath"])
			require.Equal(t, tc.assertionType, flattened[0]["type"])
			property, _ := flattened[0]["property"].(string)
			require.Equal(t, tc.property, property)
			// Preservation must change only the API-ignored value, never other fields.
			delete(flattened[0]["targetjsonpath"].([]map[string]string)[0], "targetvalue")
			delete(expectedFields[0]["targetjsonpath"].([]interface{})[0].(map[string]interface{}), "targetvalue")
			actual, err := json.Marshal(flattened)
			require.NoError(t, err)
			expected, err = json.Marshal(expectedFields)
			require.NoError(t, err)
			require.JSONEq(t, string(expected), string(actual))
		})
	}
}

func TestSyntheticsUndefinedTargetValueAPIStep(t *testing.T) {
	data := schema.TestResourceDataRaw(t, resourceDatadogSyntheticsTest().SchemaFunc(), map[string]interface{}{
		"type": "api", "subtype": "multi", "name": "steps", "status": "paused", "locations": []interface{}{"aws:eu-central-1"},
		"api_step": []interface{}{map[string]interface{}{
			"name": "step", "subtype": "grpc",
			"request_definition": []interface{}{map[string]interface{}{"host": "example.com", "port": "50051"}},
			"assertion": []interface{}{map[string]interface{}{
				"type": "grpcProto", "operator": "validatesJSONPath",
				"targetjsonpath": []interface{}{map[string]interface{}{"jsonpath": "$.value", "operator": "isUndefined", "targetvalue": "false"}},
			}},
		}},
	})
	for _, reading := range []bool{false, true, false, true} {
		apiTest, diags := buildDatadogSyntheticsAPITest(data)
		require.False(t, diags.HasError(), "%v", diags)
		encoded, err := json.Marshal(apiTest)
		require.NoError(t, err)
		require.NotContains(t, string(encoded), "targetValue")
		var response datadogV1.SyntheticsAPITest
		require.NoError(t, json.Unmarshal(encoded, &response))
		diags = updateSyntheticsAPITestLocalState(data, &response, reading)
		require.False(t, diags.HasError(), "%v", diags)
		require.Equal(t, "false", data.Get("api_step.0.assertion.0.targetjsonpath.0.targetvalue"))
		imported := schema.TestResourceDataRaw(t, resourceDatadogSyntheticsTest().SchemaFunc(), nil)
		diags = updateSyntheticsAPITestLocalState(imported, &response, true)
		require.False(t, diags.HasError(), "%v", diags)
		require.Equal(t, "", imported.Get("api_step.0.assertion.0.targetjsonpath.0.targetvalue"))
	}
}
