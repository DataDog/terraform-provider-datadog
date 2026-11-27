package datadog

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/dashboardmapping"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

func TestDashboardV2RefreshClearsAbsentMetadata(t *testing.T) {
	for _, mode := range []string{"omitted", "null", "empty"} {
		t.Run(mode, func(t *testing.T) {
			d := schema.TestResourceDataRaw(t, buildDashboardV2Schema(), nil)
			initial := map[string]interface{}{
				"title": "Example", "layout_type": "ordered", "reflow_type": "fixed",
				"description": "Old description", "restricted_roles": []interface{}{"old-role"},
				"notify_list": []interface{}{"user@example.com"}, "tags": []interface{}{"team:example"},
				"template_variables": []interface{}{map[string]interface{}{
					"name": "env", "prefix": "env", "defaults": []interface{}{"prod"},
				}},
				"template_variable_presets": []interface{}{map[string]interface{}{
					"name": "Production", "template_variables": []interface{}{map[string]interface{}{
						"name": "env", "values": []interface{}{"prod"},
					}},
				}},
				"tabs": []interface{}{map[string]interface{}{
					"id": "old-tab", "name": "Overview", "widget_ids": []interface{}{"@1"},
				}},
			}
			if diags := setDashboardStateSDKv2(d, initial); diags.HasError() {
				t.Fatalf("populate state: %v", diags)
			}
			fields := map[string]string{
				"description": "description", "reflow_type": "reflow_type",
				"restricted_roles": "restricted_roles", "notify_list": "notify_list", "tags": "tags",
				"template_variables": "template_variable", "template_variable_presets": "template_variable_preset",
				"tabs": "tab",
			}
			response := map[string]interface{}{"title": "Updated", "layout_type": "ordered"}
			for jsonKey, hclKey := range fields {
				if _, ok := d.GetOk(hclKey); !ok {
					t.Fatalf("test requires populated initial %s state", hclKey)
				}
				switch mode {
				case "null":
					response[jsonKey] = nil
				case "empty":
					if hclKey == "description" || hclKey == "reflow_type" {
						response[jsonKey] = ""
					} else {
						response[jsonKey] = []interface{}{}
					}
				}
			}
			client := dashboardValidationTestClient(t, func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodGet || r.URL.Path != "/api/v1/dashboard/abc-def-ghi" {
					t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
				}
				if err := json.NewEncoder(w).Encode(response); err != nil {
					t.Error(err)
				}
			})
			providerConfig := &ProviderConfiguration{
				Auth:                context.Background(),
				DatadogApiInstances: &utils.ApiInstances{HttpClient: client},
			}
			d.SetId("abc-def-ghi")
			if diags := resourceDatadogDashboardV2Read(context.Background(), d, providerConfig); diags.HasError() {
				t.Fatalf("refresh state: %v", diags)
			}
			for _, hclKey := range fields {
				if value, ok := d.GetOk(hclKey); ok {
					t.Errorf("stale %s after %s response: %#v", hclKey, mode, value)
				}
			}
			if got := d.Get("title"); got != "Updated" {
				t.Errorf("configured metadata was not refreshed: %#v", got)
			}

			// A later update must not reintroduce values cleared by refresh.
			body := dashboardmapping.BuildDashboardEngineJSONFromMap(collectDashboardData(d), "abc-def-ghi")
			for _, key := range []string{"description", "reflow_type", "restricted_roles", "tabs"} {
				if value, ok := body[key]; ok {
					t.Errorf("cleared %s reintroduced in request: %#v", key, value)
				}
			}
			// Import reads the same response into initially empty state. Refresh
			// must produce the same metadata as that clean read.
			imported := schema.TestResourceDataRaw(t, buildDashboardV2Schema(), nil)
			imported.SetId(d.Id())
			if diags := resourceDatadogDashboardV2Read(context.Background(), imported, providerConfig); diags.HasError() {
				t.Fatalf("import read: %v", diags)
			}
			for _, hclKey := range fields {
				if value, ok := imported.GetOk(hclKey); ok {
					t.Errorf("expected empty imported %s, got %#v", hclKey, value)
				}
			}
		})
	}
}

func TestDashboardV2RefreshClearsTemplateVariableOptions(t *testing.T) {
	for _, mode := range []string{"omitted", "null"} {
		t.Run(mode, func(t *testing.T) {
			d := schema.TestResourceDataRaw(t, buildDashboardV2Schema(), nil)
			initial := map[string]interface{}{
				"template_variables": []interface{}{map[string]interface{}{
					"name": "env", "prefix": "env", "default": "prod", "type": "group",
					"available_values": []interface{}{"prod", "staging"},
				}},
			}
			if diags := setDashboardStateSDKv2(d, initial); diags.HasError() {
				t.Fatal(diags)
			}
			variable := map[string]interface{}{"name": "env"}
			if mode == "null" {
				for _, key := range []string{"prefix", "default", "type", "available_values"} {
					variable[key] = nil
				}
			}
			if diags := setDashboardStateSDKv2(d, map[string]interface{}{
				"template_variables": []interface{}{variable},
			}); diags.HasError() {
				t.Fatal(diags)
			}
			body := dashboardmapping.BuildDashboardEngineJSONFromMap(collectDashboardData(d), "abc-def-ghi")
			variables := body["template_variables"].([]interface{})
			got := variables[0].(map[string]interface{})
			if len(got) != 1 || got["name"] != "env" {
				t.Fatalf("expected only name after options were cleared, got %#v", got)
			}
		})
	}
}
