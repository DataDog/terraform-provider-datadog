package fwprovider

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadog"
	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestActionGeneratedConnectionSecretsAreSensitive(t *testing.T) {
	var response resource.SchemaResponse
	NewDatadogActionConnectionV2Resource().Schema(context.Background(), resource.SchemaRequest{}, &response)
	count := 0
	var walk func(map[string]schema.Attribute)
	walk = func(attributes map[string]schema.Attribute) {
		for name, attribute := range attributes {
			switch a := attribute.(type) {
			case schema.SingleNestedAttribute:
				walk(a.Attributes)
			case schema.ListNestedAttribute:
				walk(a.NestedObject.Attributes)
			case schema.StringAttribute:
				switch name {
				case "api_token", "access_token", "client_secret", "global_api_key", "api_password", "sdk_key", "api_key", "app_key", "private_key", "password":
					count++
					if !a.Sensitive {
						t.Errorf("credential %s is not sensitive", name)
					}
				}
			}
		}
	}
	walk(response.Schema.Attributes)
	if count == 0 {
		t.Fatal("no credential attributes checked")
	}
}

// Exercise the generated API bindings through framework plans and state. The
// server checks the wire envelope as well as the attributes, rather than merely
// returning a canned response regardless of what the resource sends.
func TestActionGeneratedResourceLifecycle(t *testing.T) {
	ctx := context.Background()
	stringsList := func(values ...string) types.List {
		v, d := types.ListValueFrom(ctx, types.StringType, values)
		if d.HasError() {
			t.Fatal(d)
		}
		return v
	}
	for _, kind := range []string{"connection", "policy"} {
		t.Run(kind, func(t *testing.T) {
			id := "3fa85f64-5717-4562-b3fc-2c963f66afa6"
			endpoint, apiType := "/api/v2/actions/connections", "action_connection"
			updateMethod := http.MethodPatch
			attrs := map[string]any{
				"name": "tfgen-test", "tags": []any{"env:test"},
				"integration": map[string]any{"type": "AWS", "credentials": map[string]any{
					"type": "AWSAssumeRole", "account_id": "111222333444", "role": "tfgen-role",
				}},
			}
			if kind == "policy" {
				endpoint, apiType = "/api/v2/actions/execution-policies", "execution_policy"
				updateMethod = http.MethodPut
				attrs = map[string]any{"name": "tfgen-test", "effect": "deny",
					"action_pattern": map[string]any{"integration": "INTEGRATION_SCRIPT", "action_fqns": []any{"*"}},
					"scope":          map[string]any{"scripts": map[string]any{"rules": []any{map[string]any{"target_script_names": []any{"maintenance"}}}}},
				}
			}
			var methods []string
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				methods = append(methods, req.Method)
				wantPath := endpoint + "/" + id
				if req.Method == http.MethodPost {
					wantPath = endpoint
				}
				if req.URL.Path != wantPath {
					t.Errorf("path = %s, want %s", req.URL.Path, wantPath)
				}
				if req.Method == http.MethodDelete {
					w.WriteHeader(http.StatusNoContent)
					return
				}
				if req.Method == http.MethodPost || req.Method == updateMethod {
					var body struct {
						Data struct {
							ID         string         `json:"id"`
							Type       string         `json:"type"`
							Attributes map[string]any `json:"attributes"`
						} `json:"data"`
					}
					if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
						t.Error(err)
					}
					if body.Data.Type != apiType {
						t.Errorf("type = %s, want %s", body.Data.Type, apiType)
					}
					if kind == "policy" && req.Method == updateMethod && body.Data.ID != id {
						t.Errorf("update id = %s", body.Data.ID)
					}
					if !reflect.DeepEqual(body.Data.Attributes, attrs) {
						t.Errorf("attributes = %#v, want %#v", body.Data.Attributes, attrs)
					}
				}
				w.Header().Set("Content-Type", "application/json")
				if req.Method == http.MethodPost {
					w.WriteHeader(http.StatusCreated)
				}
				responseAttrs := make(map[string]any, len(attrs))
				for k, v := range attrs {
					responseAttrs[k] = v
				}
				if kind == "policy" {
					responseAttrs["targets"] = []any{}
					responseAttrs["version"] = 1
					responseAttrs["created_at"] = "2026-11-12T10:00:00Z"
					responseAttrs["updated_at"] = "2026-11-12T10:00:00Z"
					responseAttrs["created_by"] = id
					responseAttrs["updated_by"] = id
				}
				if err := json.NewEncoder(w).Encode(map[string]any{"data": map[string]any{"id": id, "type": apiType, "attributes": responseAttrs}}); err != nil {
					t.Error(err)
				}
			}))
			defer server.Close()
			config := datadog.NewConfiguration()
			config.Servers = datadog.ServerConfigurations{{URL: server.URL}}
			config.OperationServers = nil
			config.HTTPClient = server.Client()
			for _, operation := range []string{"Create", "Get", "Update", "Delete"} {
				config.SetUnstableOperationEnabled("v2."+operation+"ExecutionPolicy", true)
			}
			client := datadog.NewAPIClient(config)
			var r resource.Resource
			var model any
			if kind == "policy" {
				r = &datadogActionExecutionPolicyV2Resource{Api: datadogV2.NewExecutionPolicyApi(client), Auth: ctx}
				model = &resourceDatadogActionExecutionPolicyV2_Model{
					ID: types.StringUnknown(), Name: types.StringValue("tfgen-test"), Effect: types.StringValue("deny"), Version: types.Int64Unknown(),
					ActionPattern: &resourceDatadogActionExecutionPolicyV2_ExecutionPolicyActionPatternModel{Integration: types.StringValue("INTEGRATION_SCRIPT"), ActionFqns: stringsList("*")},
					Scope:         &resourceDatadogActionExecutionPolicyV2_ExecutionPolicyScopeModel{Scripts: &resourceDatadogActionExecutionPolicyV2_ExecutionPolicyScriptScopeModel{Rules: []*resourceDatadogActionExecutionPolicyV2_ExecutionPolicyScriptScopeRuleModel{{TargetScriptNames: stringsList("maintenance")}}}},
				}
			} else {
				r = &datadogActionConnectionV2Resource{Api: datadogV2.NewActionConnectionApi(client), Auth: ctx}
				model = &resourceDatadogActionConnectionV2_Model{ID: types.StringUnknown(), Name: types.StringValue("tfgen-test"), Tags: stringsList("env:test"),
					Integration: &resourceDatadogActionConnectionV2_ActionConnectionIntegrationModel{AwsIntegration: &resourceDatadogActionConnectionV2_ActionConnectionIntegrationAwsIntegrationModel{Type: types.StringValue("AWS"), Credentials: &resourceDatadogActionConnectionV2_AWSCredentialsModel{AwsAssumeRole: &resourceDatadogActionConnectionV2_AWSCredentialsAwsAssumeRoleModel{Type: types.StringValue("AWSAssumeRole"), AccountId: types.StringValue("111222333444"), Role: types.StringValue("tfgen-role")}}}},
				}
			}
			var sr resource.SchemaResponse
			r.Schema(ctx, resource.SchemaRequest{}, &sr)
			plan := tfsdk.Plan{Schema: sr.Schema}
			if d := plan.Set(ctx, model); d.HasError() {
				t.Fatal(d)
			}
			created := resource.CreateResponse{State: tfsdk.State{Schema: sr.Schema}}
			r.Create(ctx, resource.CreateRequest{Plan: plan}, &created)
			if created.Diagnostics.HasError() {
				t.Fatal(created.Diagnostics)
			}
			var gotID string
			if d := created.State.GetAttribute(ctx, path.Root("id"), &gotID); d.HasError() || gotID != id {
				t.Fatalf("created id = %s: %v", gotID, d)
			}
			read := resource.ReadResponse{State: created.State}
			r.Read(ctx, resource.ReadRequest{State: created.State}, &read)
			if read.Diagnostics.HasError() {
				t.Fatal(read.Diagnostics)
			}
			attrs["name"] = "tfgen-updated"
			plan.Raw = read.State.Raw
			if d := plan.SetAttribute(ctx, path.Root("name"), "tfgen-updated"); d.HasError() {
				t.Fatal(d)
			}
			updated := resource.UpdateResponse{State: read.State}
			r.Update(ctx, resource.UpdateRequest{Plan: plan, State: read.State}, &updated)
			if updated.Diagnostics.HasError() {
				t.Fatal(updated.Diagnostics)
			}
			var gotName string
			if d := updated.State.GetAttribute(ctx, path.Root("name"), &gotName); d.HasError() || gotName != "tfgen-updated" {
				t.Fatalf("updated name = %s: %v", gotName, d)
			}
			deleted := resource.DeleteResponse{State: updated.State}
			r.Delete(ctx, resource.DeleteRequest{State: updated.State}, &deleted)
			if deleted.Diagnostics.HasError() {
				t.Fatal(deleted.Diagnostics)
			}
			if !reflect.DeepEqual(methods, []string{"POST", "GET", updateMethod, "DELETE"}) {
				t.Fatalf("methods = %v", methods)
			}
		})
	}
}
