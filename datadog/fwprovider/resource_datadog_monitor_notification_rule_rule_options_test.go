package fwprovider

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestMonitorNotificationRuleRuleOptionsRequest(t *testing.T) {
	ctx := context.Background()
	r := &MonitorNotificationRuleResource{}
	newState := func(ruleOptions types.Object) *MonitorNotificationRuleModel {
		return &MonitorNotificationRuleModel{
			ID:         types.StringValue("00000000-0000-0000-0000-000000000000"),
			Name:       types.StringValue("rule"),
			Recipients: types.SetValueMust(types.StringType, []attr.Value{types.StringValue("slack-foo")}),
			MonitorNotificationRuleFilter: &MonitorNotificationRuleFilter{
				Scope: types.StringNull(),
				Tags:  types.SetValueMust(types.StringType, []attr.Value{types.StringValue("team:foo")}),
			},
			MonitorNotificationRuleBundleConfig: &MonitorNotificationRuleBundleConfig{Duration: types.Int32Value(3600)},
			RuleOptions:                         ruleOptions,
		}
	}
	threaded := func(v bool) types.Object {
		return types.ObjectValueMust(ruleOptionsAttrTypes, map[string]attr.Value{"is_threaded": types.BoolValue(v)})
	}

	for name, tc := range map[string]struct {
		ruleOptions types.Object
		want        string
	}{
		"threaded": {threaded(true), `"rule_options":{"is_threaded":true}`},
		"separate": {threaded(false), `"rule_options":{"is_threaded":false}`},
		"null":     {types.ObjectNull(ruleOptionsAttrTypes), ""},
		"empty":    {types.ObjectValueMust(ruleOptionsAttrTypes, map[string]attr.Value{"is_threaded": types.BoolNull()}), ""},
		"unknown":  {types.ObjectUnknown(ruleOptionsAttrTypes), ""},
	} {
		t.Run(name, func(t *testing.T) {
			req, diags := r.buildMonitorNotificationRuleUpdateRequest(ctx, newState(tc.ruleOptions))
			if diags.HasError() {
				t.Fatalf("unexpected diags: %v", diags)
			}
			raw, err := json.Marshal(req)
			if err != nil {
				t.Fatal(err)
			}
			body := string(raw)
			for _, s := range []string{`"name":"rule"`, `"recipients":["slack-foo"]`, `"tags":["team:foo"]`, `"bundle_config":{"duration":3600}`, `"type":"monitor-notification-rule"`} {
				if !strings.Contains(body, s) {
					t.Errorf("request body %s is missing %s", body, s)
				}
			}
			if tc.want == "" && strings.Contains(body, "rule_options") {
				t.Errorf("request body %s must not contain rule_options", body)
			}
			if tc.want != "" && !strings.Contains(body, tc.want) {
				t.Errorf("request body %s is missing %s", body, tc.want)
			}
		})
	}
}

func TestMonitorNotificationRuleRuleOptionsState(t *testing.T) {
	r := &MonitorNotificationRuleResource{}
	for name, tc := range map[string]struct {
		attributes string
		wantNull   bool
		want       bool
	}{
		"threaded": {`{"name":"rule","rule_options":{"is_threaded":true}}`, false, true},
		"separate": {`{"name":"rule","rule_options":{"is_threaded":false}}`, false, false},
		"absent":   {`{"name":"rule"}`, true, false},
	} {
		t.Run(name, func(t *testing.T) {
			var resp datadogV2.MonitorNotificationRuleResponse
			if err := json.Unmarshal([]byte(`{"data":{"id":"x","type":"monitor-notification-rule","attributes":`+tc.attributes+`}}`), &resp); err != nil {
				t.Fatal(err)
			}
			state := &MonitorNotificationRuleModel{}
			r.updateState(context.Background(), state, &resp)
			if state.RuleOptions.IsNull() != tc.wantNull {
				t.Fatalf("rule_options null = %v, want %v", state.RuleOptions.IsNull(), tc.wantNull)
			}
			if tc.wantNull {
				return
			}
			if got := state.RuleOptions.Attributes()["is_threaded"].(types.Bool).ValueBool(); got != tc.want {
				t.Errorf("is_threaded = %v, want %v", got, tc.want)
			}
		})
	}
}
