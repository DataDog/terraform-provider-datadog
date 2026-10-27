package datadog

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	datadogClient "github.com/DataDog/datadog-api-client-go/v2/api/datadog"
	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/go-cty/cty"
	ctyjson "github.com/hashicorp/go-cty/cty/json"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

func TestSensitiveDataScannerSuppressionsPayload(t *testing.T) {
	for _, standard := range []bool{false, true} {
		for _, choice := range []interface{}{nil, false, true} {
			for _, populated := range []bool{false, true} {
				t.Run(fmt.Sprintf("standard=%v/choice=%v/populated=%v", standard, choice, populated), func(t *testing.T) {
					config := map[string]interface{}{"group_id": "group"}
					if standard {
						config["standard_pattern_id"] = "standard"
					}
					if choice != nil {
						config["use_recommended_suppressions"] = choice
					}
					if populated {
						config["suppressions"] = []interface{}{map[string]interface{}{
							"starts_with": []interface{}{"prefix"}, "ends_with": []interface{}{"suffix"}, "exact_match": []interface{}{"exact"},
						}}
					}
					data := schema.TestResourceDataRaw(t, resourceDatadogSensitiveDataScannerRule().SchemaMap(), config)
					attributes := buildSensitiveDataScannerRuleAttributes(data)
					if !standard && !populated {
						assert.False(t, attributes.HasSuppressions(), "custom-rule creation without lists preserves the existing request shape")
						return
					}
					require.True(t, attributes.HasSuppressions())
					payload, err := json.Marshal(attributes.GetSuppressions())
					require.NoError(t, err)
					var got map[string]interface{}
					require.NoError(t, json.Unmarshal(payload, &got))
					for key, value := range map[string]string{"starts_with": "prefix", "ends_with": "suffix", "exact_match": "exact"} {
						if populated {
							assert.Equal(t, []interface{}{value}, got[key])
						} else {
							assert.Equal(t, []interface{}{}, got[key], "clear must serialize an empty array")
						}
					}
					if standard {
						want := choice
						if want == nil {
							want = true
						}
						assert.Equal(t, want, got["use_recommended_suppressions"])
					} else {
						assert.NotContains(t, got, "use_recommended_suppressions")
					}
				})
			}
		}
	}
}

func TestSensitiveDataScannerSuppressionsPlan(t *testing.T) {
	const unknown = "74D93920-ED26-11E3-AC10-0800200C9A66"
	for _, tc := range []struct {
		name      string
		standard  interface{}
		choice    interface{}
		want      string
		wantError bool
	}{
		{name: "standard defaults on", standard: "standard", want: "true"},
		{name: "custom defaults off", want: "false"},
		{name: "explicit off", standard: "standard", choice: false, want: "false"},
		{name: "custom rejects on", choice: true, wantError: true},
		{name: "unknown standard", standard: unknown, choice: true, want: "true"},
		{name: "unknown standard and omitted choice", standard: unknown},
		{name: "unknown choice", standard: "standard", choice: unknown},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r := resourceDatadogSensitiveDataScannerRule()
			config := map[string]interface{}{"group_id": "group"}
			if tc.standard != nil {
				config["standard_pattern_id"] = tc.standard
			}
			if tc.choice != nil {
				config["use_recommended_suppressions"] = tc.choice
			}
			state := &terraform.InstanceState{}
			sdsRuleSetRawConfig(t, r, state, config)
			diff, err := r.Diff(context.Background(), state, terraform.NewResourceConfigRaw(config), nil)
			if tc.wantError {
				require.ErrorContains(t, err, "requires standard_pattern_id")
				return
			}
			require.NoError(t, err)
			require.NotNil(t, diff)
			attr := diff.Attributes["use_recommended_suppressions"]
			require.NotNil(t, attr)
			if tc.want == "" {
				assert.True(t, attr.NewComputed)
			} else {
				assert.Equal(t, tc.want, attr.New)
			}
		})
	}
}

func TestSensitiveDataScannerSuppressionsRuleReplacement(t *testing.T) {
	for _, choice := range []interface{}{nil, false, true} {
		t.Run(fmt.Sprint(choice), func(t *testing.T) {
			r := resourceDatadogSensitiveDataScannerRule()
			data := schema.TestResourceDataRaw(t, r.SchemaMap(), map[string]interface{}{
				"group_id": "group", "standard_pattern_id": "standard", "use_recommended_suppressions": true,
			})
			data.SetId("standard-rule")
			state := data.State()
			config := map[string]interface{}{"group_id": "group", "pattern": "custom-pattern"}
			if choice != nil {
				config["use_recommended_suppressions"] = choice
			}
			sdsRuleSetRawConfig(t, r, state, config)
			diff, err := r.Diff(context.Background(), state, terraform.NewResourceConfigRaw(config), nil)
			if choice == true {
				require.ErrorContains(t, err, "requires standard_pattern_id")
				return
			}
			require.NoError(t, err)
			require.True(t, diff.RequiresNew())
			require.NotNil(t, diff.Attributes["use_recommended_suppressions"])
			assert.Equal(t, "false", diff.Attributes["use_recommended_suppressions"].New)
		})
	}
}

func TestSensitiveDataScannerSuppressionsRead(t *testing.T) {
	for _, choice := range []bool{false, true} {
		for _, populated := range []bool{false, true} {
			t.Run(fmt.Sprintf("choice=%v/populated=%v", choice, populated), func(t *testing.T) {
				r := resourceDatadogSensitiveDataScannerRule()
				data := r.Data(&terraform.InstanceState{ID: "imported", Attributes: map[string]string{"standard_pattern_id": "standard"}})
				suppressions := datadogV2.NewSensitiveDataScannerSuppressions()
				suppressions.SetUseRecommendedSuppressions(choice)
				if populated {
					suppressions.SetExactMatch([]string{"saved-customer-value"})
				}
				attributes := datadogV2.NewSensitiveDataScannerRuleAttributesWithDefaults()
				attributes.SetSuppressions(*suppressions)
				require.False(t, updateSensitiveDataScannerRuleState(data, attributes).HasError())
				assert.Equal(t, choice, data.Get("use_recommended_suppressions"))
				if populated {
					assert.Equal(t, "saved-customer-value", data.Get("suppressions.0.exact_match.0"))
				} else {
					assert.Empty(t, data.Get("suppressions"))
				}
				// An older API response without the toggle must not overwrite a known choice.
				suppressions.UseRecommendedSuppressions = nil
				attributes.SetSuppressions(*suppressions)
				require.False(t, updateSensitiveDataScannerRuleState(data, attributes).HasError())
				assert.Equal(t, choice, data.Get("use_recommended_suppressions"))
			})
		}
	}
	data := resourceDatadogSensitiveDataScannerRule().Data(&terraform.InstanceState{ID: "legacy", Attributes: map[string]string{"standard_pattern_id": "standard"}})
	attributes := datadogV2.NewSensitiveDataScannerRuleAttributesWithDefaults()
	require.False(t, updateSensitiveDataScannerRuleState(data, attributes).HasError())
	_, known := data.GetOkExists("use_recommended_suppressions")
	assert.False(t, known, "missing API choice must remain unset")
	payload := buildSensitiveDataScannerSuppressions(data)
	assert.False(t, payload.HasUseRecommendedSuppressions())
}

func TestSensitiveDataScannerSuppressionsLifecycle(t *testing.T) {
	for _, standardID := range []string{"regex-standard", "ml-standard"} {
		t.Run(standardID, func(t *testing.T) {
			ctx := context.Background()
			r := resourceDatadogSensitiveDataScannerRule()
			var saved map[string]interface{}
			var relationships map[string]interface{}
			var lastPayload map[string]interface{}
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				if req.Method != http.MethodGet {
					var body struct {
						Data struct {
							Attributes    map[string]interface{} `json:"attributes"`
							Relationships map[string]interface{} `json:"relationships"`
						} `json:"data"`
					}
					if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
						t.Errorf("decode request: %v", err)
						w.WriteHeader(400)
						return
					}
					saved = body.Data.Attributes
					lastPayload = saved["suppressions"].(map[string]interface{})
					if body.Data.Relationships != nil {
						relationships = body.Data.Relationships
					}
				}
				item := map[string]interface{}{"id": "rule", "type": "sensitive_data_scanner_rule", "attributes": saved, "relationships": relationships}
				response := map[string]interface{}{"data": item}
				if req.Method == http.MethodGet {
					response = map[string]interface{}{"data": map[string]interface{}{"id": "config", "type": "sensitive_data_scanner_configuration"}, "included": []interface{}{item}}
				}
				if err := json.NewEncoder(w).Encode(response); err != nil {
					t.Errorf("encode response: %v", err)
				}
			}))
			defer server.Close()
			apiConfig := datadogClient.NewConfiguration()
			apiConfig.Servers = datadogClient.ServerConfigurations{{URL: server.URL}}
			apiConfig.OperationServers = nil
			apiConfig.HTTPClient = server.Client()
			provider := &ProviderConfiguration{DatadogApiInstances: &utils.ApiInstances{HttpClient: datadogClient.NewAPIClient(apiConfig)}, Auth: ctx}
			config := map[string]interface{}{"name": "test", "description": "test", "group_id": "group", "standard_pattern_id": standardID, "is_enabled": true, "priority": 1, "included_keyword_configuration": []interface{}{map[string]interface{}{"keywords": []interface{}{}, "character_count": 30}}}
			var state *terraform.InstanceState
			// Apply through the SDK, including CustomizeDiff and actual HTTP client serialization.
			for step, tc := range []struct {
				choice     interface{}
				block      interface{}
				wantChoice bool
			}{
				{wantChoice: true},
				{block: sdsRuleSuppressionBlock("original"), wantChoice: true},
				{choice: false, block: sdsRuleSuppressionBlock("original")},
				{choice: true, block: sdsRuleSuppressionBlock("original"), wantChoice: true},
				{block: sdsRuleSuppressionBlock("replacement"), wantChoice: true},
				{block: []interface{}{map[string]interface{}{"exact_match": []interface{}{"replacement"}}}, wantChoice: true},
				{wantChoice: true},
				{choice: false},
				{block: sdsRuleSuppressionBlock("custom-only")},
				{block: []interface{}{map[string]interface{}{"exact_match": []interface{}{"custom-only"}}}},
				{},
				{block: []interface{}{map[string]interface{}{}}},
				{},
			} {
				t.Logf("step %d", step)
				config["name"] = fmt.Sprintf("test-step-%d", step)
				delete(config, "use_recommended_suppressions")
				delete(config, "suppressions")
				if tc.choice != nil {
					config["use_recommended_suppressions"] = tc.choice
				}
				if tc.block != nil {
					config["suppressions"] = tc.block
				}
				if state == nil {
					state = &terraform.InstanceState{}
				}
				sdsRuleSetRawConfig(t, r, state, config)
				diff, err := r.Diff(ctx, state, terraform.NewResourceConfigRaw(config), provider)
				require.NoError(t, err)
				if diff != nil && !diff.Empty() {
					diff.RawConfig = state.RawConfig
					updated, diags := r.Apply(ctx, state, diff, provider)
					require.False(t, diags.HasError(), "%v", diags)
					state = updated
				}
				assert.Equal(t, tc.wantChoice, lastPayload["use_recommended_suppressions"])
				for _, key := range []string{"starts_with", "ends_with", "exact_match"} {
					want := []interface{}{}
					if blocks, ok := tc.block.([]interface{}); ok {
						if values, ok := blocks[0].(map[string]interface{})[key]; ok {
							want = values.([]interface{})
						}
					}
					assert.Equal(t, want, lastPayload[key], key)
				}
				sdsRuleSetRawConfig(t, r, state, config)
				converged, err := r.Diff(ctx, state, terraform.NewResourceConfigRaw(config), provider)
				require.NoError(t, err)
				assert.True(t, converged == nil || converged.Empty(), "perpetual diff: %#v", converged)
				imported := &terraform.InstanceState{ID: "rule"}
				refreshed, diags := r.RefreshWithoutUpgrade(ctx, imported, provider)
				require.False(t, diags.HasError(), "%v", diags)
				assert.Equal(t, fmt.Sprint(tc.wantChoice), refreshed.Attributes["use_recommended_suppressions"])
				if tc.block == nil || len(refreshed.Attributes["suppressions.0.exact_match.0"]) > 0 {
					sdsRuleSetRawConfig(t, r, refreshed, config)
					importDiff, err := r.Diff(ctx, refreshed, terraform.NewResourceConfigRaw(config), provider)
					require.NoError(t, err)
					assert.True(t, importDiff == nil || importDiff.Empty(), "import diff: %#v", importDiff)
				}
			}
		})
	}
}

func sdsRuleSuppressionBlock(value string) []interface{} {
	return []interface{}{map[string]interface{}{"starts_with": []interface{}{value}, "ends_with": []interface{}{value}, "exact_match": []interface{}{value}}}
}

func sdsRuleSetRawConfig(t *testing.T, r *schema.Resource, state *terraform.InstanceState, config map[string]interface{}) {
	t.Helper()
	encoded, err := json.Marshal(config)
	require.NoError(t, err)
	typ, err := ctyjson.ImpliedType(encoded)
	require.NoError(t, err)
	value, err := ctyjson.Unmarshal(encoded, typ)
	require.NoError(t, err)
	value, err = cty.Transform(value, func(_ cty.Path, v cty.Value) (cty.Value, error) {
		if v.Type() == cty.String && v.AsString() == "74D93920-ED26-11E3-AC10-0800200C9A66" {
			return cty.UnknownVal(cty.String), nil
		}
		return v, nil
	})
	require.NoError(t, err)
	state.RawConfig, err = r.CoreConfigSchema().CoerceValue(value)
	require.NoError(t, err)
}
