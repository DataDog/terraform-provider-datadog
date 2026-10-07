package fwprovider

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadog"
	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestRumTeamOwnershipResponseParsing(t *testing.T) {
	for _, operation := range []string{"create", "read"} {
		for _, matchType := range []string{"exact", "prefix", "unknown"} {
			t.Run(operation+"/"+matchType, func(t *testing.T) {
				ctx := context.Background()
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
					wantMethod, wantPath := http.MethodGet, "/api/v2/rum/config/teams-ownership/mappings/123"
					if operation == "create" {
						wantMethod, wantPath = http.MethodPost, "/api/v2/rum/config/teams-ownership/mappings"
					}
					if request.Method != wantMethod || request.URL.Path != wantPath {
						t.Errorf("request = %s %s, want %s %s", request.Method, request.URL.Path, wantMethod, wantPath)
						w.WriteHeader(http.StatusBadRequest)
						return
					}
					w.Header().Set("Content-Type", "application/json")
					if operation == "create" {
						w.WriteHeader(http.StatusCreated)
					}
					fmt.Fprintf(w, `{"data":{"id":"123","type":"teams_ownership_mappings","attributes":{
						"application_id":"00000000-0000-0000-0000-000000000000",
						"created_at":"2026-10-07T10:00:00Z","created_by":"test-user",
						"match_type":%q,"org_id":2,"service":"","team_handle":"web","view_name":"/checkout"
					}}}`, matchType)
				}))
				defer server.Close()

				config := datadog.NewConfiguration()
				config.Servers = datadog.ServerConfigurations{{URL: server.URL}}
				config.OperationServers = nil
				config.HTTPClient = server.Client()
				EnableGeneratedUnstableOperations(config)
				r := &datadogRumTeamOwnershipResource{
					Api: datadogV2.NewRumTeamsOwnershipApi(datadog.NewAPIClient(config)), Auth: ctx,
				}
				var schemaResponse resource.SchemaResponse
				r.Schema(ctx, resource.SchemaRequest{}, &schemaResponse)
				model := datadogRumTeamOwnershipResourceModel{
					ID: types.StringValue("123"), ApplicationId: types.StringValue("00000000-0000-0000-0000-000000000000"),
					MatchType: types.StringValue("exact"), Service: types.StringValue(""),
					TeamHandle: types.StringValue("web"), ViewName: types.StringValue("/checkout"),
				}
				priorState := tfsdk.State{Schema: schemaResponse.Schema}
				if diags := priorState.Set(ctx, &model); diags.HasError() {
					t.Fatalf("building state: %v", diags)
				}
				var state tfsdk.State
				var diagnostics diag.Diagnostics
				if operation == "create" {
					model.ID = types.StringUnknown()
					plan := tfsdk.Plan{Schema: schemaResponse.Schema}
					if diags := plan.Set(ctx, &model); diags.HasError() {
						t.Fatalf("building plan: %v", diags)
					}
					response := resource.CreateResponse{State: tfsdk.State{Raw: plan.Raw, Schema: schemaResponse.Schema}}
					r.Create(ctx, resource.CreateRequest{Plan: plan}, &response)
					state, diagnostics = response.State, response.Diagnostics
					if matchType == "unknown" && !state.Raw.Equal(plan.Raw) {
						t.Fatal("unparsed create response changed state")
					}
				} else {
					response := resource.ReadResponse{State: priorState}
					r.Read(ctx, resource.ReadRequest{State: priorState}, &response)
					state, diagnostics = response.State, response.Diagnostics
					if matchType == "unknown" && !state.Raw.Equal(priorState.Raw) {
						t.Fatal("unparsed read response changed prior state")
					}
				}
				if matchType == "unknown" {
					if len(diagnostics.Errors()) != 1 || diagnostics.Errors()[0].Summary() != "response contains unparsedObject" ||
						!strings.Contains(diagnostics.Errors()[0].Detail(), "unknown") {
						t.Fatalf("expected an unparsed response diagnostic containing the unknown value, got %v", diagnostics)
					}
					return
				}
				if diagnostics.HasError() {
					t.Fatalf("valid response returned errors: %v", diagnostics)
				}
				var got datadogRumTeamOwnershipResourceModel
				if diags := state.Get(ctx, &got); diags.HasError() {
					t.Fatalf("reading state: %v", diags)
				}
				if got.ID.ValueString() != "123" || got.MatchType.ValueString() != matchType {
					t.Fatalf("state ID/match_type = %s/%s, want 123/%s", got.ID.ValueString(), got.MatchType.ValueString(), matchType)
				}
			})
		}
	}
}
