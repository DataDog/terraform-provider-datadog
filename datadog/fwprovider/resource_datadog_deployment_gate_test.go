package fwprovider

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func testRuleOptions() *deploymentGateRuleOptionsModel {
	return &deploymentGateRuleOptionsModel{
		AllowedResources: types.ListNull(types.StringType), ExcludedResources: types.ListNull(types.StringType),
		Duration: types.Int64Null(), FailOnNoData: types.BoolNull(),
		FailOnNoGroupsFound: types.BoolNull(), Query: types.StringNull(), Warmup: types.Int64Null(),
	}
}

func TestDeploymentGateRuleOptionsRoundTrip(t *testing.T) {
	ctx := context.Background()
	cases := []struct {
		name, ruleType string
		options        *deploymentGateRuleOptionsModel
		want           map[string]interface{}
	}{
		{
			name: "allowed resources", ruleType: "faulty_deployment_detection",
			options: func() *deploymentGateRuleOptionsModel {
				o := testRuleOptions()
				o.AllowedResources, _ = types.ListValueFrom(ctx, types.StringType, []string{"GET /health"})
				return o
			}(),
			want: map[string]interface{}{"allowed_resources": []interface{}{"GET /health"}},
		},
		{
			name: "monitor query", ruleType: "monitor",
			options: func() *deploymentGateRuleOptionsModel {
				o := testRuleOptions()
				o.Query = types.StringValue("service:web")
				o.FailOnNoData = types.BoolValue(false)
				o.FailOnNoGroupsFound = types.BoolValue(true)
				o.Warmup = types.Int64Value(30)
				return o
			}(),
			want: map[string]interface{}{"query": "service:web", "fail_on_no_data": false, "fail_on_no_groups_found": true, "warmup": float64(30)},
		},
		{
			name: "monitor ids with empty groups", ruleType: "monitor",
			options: func() *deploymentGateRuleOptionsModel {
				o := testRuleOptions()
				o.MonitorIDs = []deploymentGateMonitorIDModel{{ID: types.StringValue("123"), Groups: types.ListValueMust(types.StringType, nil)}}
				o.FailOnNoData = types.BoolValue(true)
				return o
			}(),
			want: map[string]interface{}{"monitor_ids": []interface{}{map[string]interface{}{"id": "123", "groups": []interface{}{}}}, "fail_on_no_data": true},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rule := deploymentGateRuleModel{Name: types.StringValue("rule"), Type: types.StringValue(tc.ruleType), Options: tc.options}
			if d := (&deploymentGateResource{}).validateRules(ctx, &deploymentGateModel{Rules: []deploymentGateRuleModel{rule}}); d.HasError() {
				t.Fatalf("unexpected validation diagnostics: %v", d)
			}
			for _, update := range []bool{false, true} {
				var raw []byte
				if update {
					body, d := (&deploymentGateResource{}).buildRuleUpdateRequestBody(ctx, &rule)
					if d.HasError() {
						t.Fatalf("update: %v", d)
					}
					attributes := body.Data.GetAttributes()
					raw, _ = json.Marshal(attributes.GetOptions())
				} else {
					body, d := (&deploymentGateResource{}).buildRuleRequestBody(ctx, &rule)
					if d.HasError() {
						t.Fatalf("create: %v", d)
					}
					attributes := body.Data.GetAttributes()
					raw, _ = json.Marshal(attributes.GetOptions())
				}
				var got map[string]interface{}
				if err := json.Unmarshal(raw, &got); err != nil {
					t.Fatal(err)
				}
				for key, want := range tc.want {
					gotJSON, _ := json.Marshal(got[key])
					wantJSON, _ := json.Marshal(want)
					if string(gotJSON) != string(wantJSON) {
						t.Errorf("%s: got %s, want %s", key, gotJSON, wantJSON)
					}
				}
			}
			attributes := datadogV2.DeploymentRuleResponseDataAttributes{}
			responseOptions := make(map[string]interface{}, len(tc.want))
			for key, value := range tc.want {
				responseOptions[key] = value
			}
			// Staging omits monitor options set to their API defaults.
			if responseOptions["warmup"] == float64(0) {
				delete(responseOptions, "warmup")
			}
			if responseOptions["fail_on_no_data"] == true {
				delete(responseOptions, "fail_on_no_data")
			}
			if responseOptions["fail_on_no_groups_found"] == false {
				delete(responseOptions, "fail_on_no_groups_found")
			}
			attributes.SetOptions(datadogV2.DeploymentRulesOptions{UnparsedObject: responseOptions})
			attributes.SetType(datadogV2.DeploymentRuleResponseDataAttributesType(tc.ruleType))
			(&deploymentGateResource{}).updateRuleStateFromAttributes(ctx, &rule, &attributes, false)
			options, d := buildRuleOptions(ctx, &rule)
			if d.HasError() {
				t.Fatalf("read options: %v", d)
			}
			after, _ := json.Marshal(options)
			var actual map[string]interface{}
			if err := json.Unmarshal(after, &actual); err != nil {
				t.Fatal(err)
			}
			for key, want := range tc.want {
				gotJSON, _ := json.Marshal(actual[key])
				wantJSON, _ := json.Marshal(want)
				if string(gotJSON) != string(wantJSON) {
					t.Errorf("round trip %s: got %s, want %s", key, gotJSON, wantJSON)
				}
			}
		})
	}
}

func TestDeploymentGateRefreshDetectsOmittedOptionDrift(t *testing.T) {
	ctx := context.Background()
	tests := []struct {
		name, ruleType string
		response       map[string]interface{}
		check          func(*deploymentGateRuleOptionsModel) bool
	}{
		{
			name: "monitor non-defaults", ruleType: "monitor",
			response: map[string]interface{}{"query": "service:web", "fail_on_no_data": false, "fail_on_no_groups_found": true, "warmup": 15},
			check: func(o *deploymentGateRuleOptionsModel) bool {
				return !o.FailOnNoData.ValueBool() && o.FailOnNoGroupsFound.ValueBool() && o.Warmup.ValueInt64() == 15 &&
					!o.FailOnNoData.IsNull() && !o.FailOnNoGroupsFound.IsNull() && !o.Warmup.IsNull()
			},
		},
		{
			name: "allowed resources", ruleType: "faulty_deployment_detection",
			response: map[string]interface{}{"allowed_resources": []string{"GET /api/status"}},
			check: func(o *deploymentGateRuleOptionsModel) bool {
				return !o.AllowedResources.IsNull() && len(o.AllowedResources.Elements()) == 1
			},
		},
		{
			name: "monitor ids", ruleType: "monitor",
			response: map[string]interface{}{"monitor_ids": []interface{}{map[string]interface{}{"id": "123", "groups": []string{}}}},
			check: func(o *deploymentGateRuleOptionsModel) bool {
				return len(o.MonitorIDs) == 1 && o.MonitorIDs[0].ID.ValueString() == "123"
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			attributes := datadogV2.DeploymentRuleResponseDataAttributes{}
			attributes.SetType(datadogV2.DeploymentRuleResponseDataAttributesType(tc.ruleType))
			attributes.SetOptions(datadogV2.DeploymentRulesOptions{UnparsedObject: tc.response})
			for _, refresh := range []bool{false, true} {
				rule := deploymentGateRuleModel{Type: types.StringValue(tc.ruleType), Options: testRuleOptions()}
				(&deploymentGateResource{}).updateRuleStateFromAttributes(ctx, &rule, &attributes, refresh)
				if got := tc.check(rule.Options); got != refresh {
					t.Errorf("refresh=%t: observed out-of-band options=%t, want %t", refresh, got, refresh)
				}
			}
		})
	}
	attributes := datadogV2.DeploymentRuleResponseDataAttributes{}
	attributes.SetType("monitor")
	attributes.SetOptions(datadogV2.DeploymentRulesOptions{UnparsedObject: map[string]interface{}{
		"query": "service:web", "fail_on_no_data": true, "fail_on_no_groups_found": false, "warmup": 0,
	}})
	rule := deploymentGateRuleModel{Type: types.StringValue("monitor"), Options: testRuleOptions()}
	(&deploymentGateResource{}).updateRuleStateFromAttributes(ctx, &rule, &attributes, true)
	if !rule.Options.FailOnNoData.IsNull() || !rule.Options.FailOnNoGroupsFound.IsNull() || !rule.Options.Warmup.IsNull() {
		t.Fatalf("default values should not populate omitted options: %#v", rule.Options)
	}
}

func TestDeploymentGateRuleOptionsValidation(t *testing.T) {
	ctx := context.Background()
	options := testRuleOptions()
	options.Query = types.StringValue("service:web")
	options.MonitorIDs = []deploymentGateMonitorIDModel{{ID: types.StringValue("123"), Groups: types.ListValueMust(types.StringType, nil)}}
	options.AllowedResources = types.ListValueMust(types.StringType, nil)
	options.ExcludedResources = types.ListValueMust(types.StringType, nil)
	rule := deploymentGateRuleModel{Name: types.StringValue("rule"), Type: types.StringValue("monitor"), Options: options}
	if diags := (&deploymentGateResource{}).validateRules(ctx, &deploymentGateModel{Rules: []deploymentGateRuleModel{rule}}); !diags.HasError() {
		t.Fatal("expected mutually exclusive options to fail validation")
	}
}
