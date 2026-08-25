package fwprovider

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadog"
	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestActionExecutionPolicyWriteUsesCanonicalResponseState(t *testing.T) {
	const policyID = "7f8396a4-cbba-4a5e-8b77-eefa9e535899"
	ctx := context.Background()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		if request.URL.Path != "/api/v2/actions/execution-policies" && request.URL.Path != "/api/v2/actions/execution-policies/"+policyID {
			t.Fatalf("unexpected request path: %s", request.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{
			"data": {
				"id": %q,
				"type": "execution_policy",
				"attributes": {
					"action_pattern": {"integration": "INTEGRATION_SCRIPT", "action_fqns": ["*"]},
					"created_at": "2026-08-25T10:00:00Z",
					"created_by": "test-user",
					"effect": "allow",
					"name": "canonical-state-test",
					"targets": [{"name": "test-agents", "agent_tags": ["env:test", "service:web"]}],
					"updated_at": "2026-08-25T10:00:00Z",
					"updated_by": "test-user",
					"version": 1
				}
			}
		}`, policyID)
	}))
	defer server.Close()

	config := datadog.NewConfiguration()
	config.Servers = datadog.ServerConfigurations{{URL: server.URL}}
	config.OperationServers = nil
	config.HTTPClient = server.Client()
	config.SetUnstableOperationEnabled("v2.CreateExecutionPolicy", true)
	config.SetUnstableOperationEnabled("v2.UpdateExecutionPolicy", true)
	r := &actionExecutionPolicyResource{
		Api:  datadogV2.NewExecutionPolicyApi(datadog.NewAPIClient(config)),
		Auth: ctx,
	}

	var schemaResponse resource.SchemaResponse
	r.Schema(ctx, resource.SchemaRequest{}, &schemaResponse)

	plannedTags, diags := types.ListValueFrom(ctx, types.StringType, []string{"service:web", "env:test", "env:test"})
	if diags.HasError() {
		t.Fatalf("building planned tags: %v", diags.Errors())
	}
	actionFQNs, diags := types.ListValueFrom(ctx, types.StringType, []string{"*"})
	if diags.HasError() {
		t.Fatalf("building action FQNs: %v", diags.Errors())
	}
	planModel := actionExecutionPolicyModel{
		ID:        types.StringValue(policyID),
		Name:      types.StringValue("canonical-state-test"),
		Effect:    types.StringValue("allow"),
		Version:   types.Int32Unknown(),
		CreatedAt: types.StringUnknown(),
		CreatedBy: types.StringUnknown(),
		UpdatedAt: types.StringUnknown(),
		UpdatedBy: types.StringUnknown(),
		ActionPattern: &executionPolicyActionPatternModel{
			Integration: types.StringValue("INTEGRATION_SCRIPT"),
			ActionFqns:  actionFQNs,
		},
		Target: []*executionPolicyTargetModel{{
			Name:      types.StringValue("test-agents"),
			AgentTags: plannedTags,
		}},
	}

	plan := tfsdk.Plan{Schema: schemaResponse.Schema}
	if diags := plan.Set(ctx, &planModel); diags.HasError() {
		t.Fatalf("building plan: %v", diags.Errors())
	}
	priorState := tfsdk.State{Schema: schemaResponse.Schema}
	if diags := priorState.Set(ctx, &planModel); diags.HasError() {
		t.Fatalf("building prior state: %v", diags.Errors())
	}

	tests := []struct {
		name string
		run  func() (tfsdk.State, bool)
	}{
		{
			name: "create",
			run: func() (tfsdk.State, bool) {
				response := resource.CreateResponse{State: tfsdk.State{Raw: plan.Raw, Schema: schemaResponse.Schema}}
				r.Create(ctx, resource.CreateRequest{Plan: plan}, &response)
				return response.State, response.Diagnostics.HasError()
			},
		},
		{
			name: "update",
			run: func() (tfsdk.State, bool) {
				response := resource.UpdateResponse{State: tfsdk.State{Raw: plan.Raw, Schema: schemaResponse.Schema}}
				r.Update(ctx, resource.UpdateRequest{Plan: plan, State: priorState}, &response)
				return response.State, response.Diagnostics.HasError()
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			state, hasError := test.run()
			if hasError {
				t.Fatal("write returned diagnostics")
			}

			var got actionExecutionPolicyModel
			if diags := state.Get(ctx, &got); diags.HasError() {
				t.Fatalf("reading state: %v", diags.Errors())
			}
			var gotTags []string
			if diags := got.Target[0].AgentTags.ElementsAs(ctx, &gotTags, false); diags.HasError() {
				t.Fatalf("reading target tags: %v", diags.Errors())
			}
			if len(gotTags) != 2 || gotTags[0] != "env:test" || gotTags[1] != "service:web" {
				t.Fatalf("target tags = %v, want backend-canonical [env:test service:web]", gotTags)
			}
		})
	}
}

func TestUpdateExecutionPolicyStateFromResponseDropsEmptyScope(t *testing.T) {
	t.Parallel()

	actionPattern := datadogV2.NewExecutionPolicyActionPattern(
		[]string{"*"},
		datadogV2.EXECUTIONPOLICYINTEGRATION_INTEGRATION_SCRIPT,
	)
	attributes := datadogV2.NewExecutionPolicyAttributesWithDefaults()
	attributes.SetActionPattern(*actionPattern)
	attributes.SetScope(*datadogV2.NewExecutionPolicyScope())
	data := datadogV2.NewExecutionPolicyResponseData(
		*attributes,
		"policy-id",
		datadogV2.EXECUTIONPOLICYTYPE_EXECUTION_POLICY,
	)
	state := actionExecutionPolicyModel{
		Scope: &executionPolicyScopeModel{},
	}
	var diagnostics diag.Diagnostics

	updateExecutionPolicyStateFromResponse(context.Background(), &state, data, &diagnostics)

	if diagnostics.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diagnostics.Errors())
	}
	if state.Scope != nil {
		t.Fatalf("scope = %#v, want nil for an empty API scope", state.Scope)
	}
}
