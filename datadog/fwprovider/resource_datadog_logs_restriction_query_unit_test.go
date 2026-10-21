package fwprovider

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadog"
	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	frameworkPath "github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// A role bind that fails after the query was created must leave the query's ID in state, so
// Terraform tracks it instead of creating another query on the next apply.
func TestLogsRestrictionQueryCreateKeepsIDWhenRoleBindFails(t *testing.T) {
	const queryID = "79a0e60a-644a-11ea-ad29-43329f7f58b5"
	const roleID = "00000000-0000-1111-0000-000000000000"
	ctx := context.Background()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case request.Method == http.MethodPost && request.URL.Path == "/api/v2/logs/config/restriction_queries":
			_, _ = w.Write([]byte(`{"data":{"id":"` + queryID + `","type":"logs_restriction_queries",` +
				`"attributes":{"restriction_query":"service:foo"}}}`))
		case request.Method == http.MethodPost && request.URL.Path == "/api/v2/logs/config/restriction_queries/"+queryID+"/roles":
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(`{"errors":["internal error"]}`))
		default:
			t.Errorf("unexpected request: %s %s", request.Method, request.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	config := datadog.NewConfiguration()
	config.Servers = datadog.ServerConfigurations{{URL: server.URL}}
	config.OperationServers = nil
	config.HTTPClient = server.Client()
	config.RetryConfiguration.EnableRetry = false
	config.SetUnstableOperationEnabled("v2.CreateRestrictionQuery", true)
	config.SetUnstableOperationEnabled("v2.AddRoleToRestrictionQuery", true)
	r := &LogsRestrictionQueryResource{
		API:  datadogV2.NewLogsRestrictionQueriesApi(datadog.NewAPIClient(config)),
		Auth: ctx,
	}

	var schemaResponse resource.SchemaResponse
	r.Schema(ctx, resource.SchemaRequest{}, &schemaResponse)

	roleIDs, diags := types.SetValueFrom(ctx, types.StringType, []string{roleID})
	if diags.HasError() {
		t.Fatalf("building role ids: %v", diags.Errors())
	}
	plan := tfsdk.Plan{Schema: schemaResponse.Schema}
	if diags := plan.Set(ctx, &LogsRestrictionQueryModel{
		ID:               types.StringUnknown(),
		RestrictionQuery: types.StringValue("service:foo"),
		RoleIds:          roleIDs,
		CreatedAt:        timetypes.NewRFC3339Unknown(),
		ModifiedAt:       timetypes.NewRFC3339Unknown(),
	}); diags.HasError() {
		t.Fatalf("building plan: %v", diags.Errors())
	}

	// The framework server starts Create from a null state.
	nullState := tftypes.NewValue(schemaResponse.Schema.Type().TerraformType(ctx), nil)
	response := resource.CreateResponse{State: tfsdk.State{Raw: nullState, Schema: schemaResponse.Schema}}
	r.Create(ctx, resource.CreateRequest{Plan: plan}, &response)

	if !response.Diagnostics.HasError() {
		t.Fatal("Create returned no error for a failed role bind")
	}
	var id types.String
	if diags := response.State.GetAttribute(ctx, frameworkPath.Root("id"), &id); diags.HasError() {
		t.Fatalf("reading id from state: %v", diags.Errors())
	}
	if id.ValueString() != queryID {
		t.Fatalf("state id = %q, want %q: the created query would be orphaned", id.ValueString(), queryID)
	}
}
