package fwprovider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadog"
	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	statusPageEmailAccessPolicyTestPageID   = "00000000-0000-4000-8000-000000000101"
	statusPageEmailAccessPolicyTestPolicyID = "00000000-0000-4000-8000-000000000102"
	statusPageEmailAccessPolicyTestOtherID  = "00000000-0000-4000-8000-000000000103"
	statusPageEmailAccessPolicyTestPath     = "/api/v2/statuspages/" + statusPageEmailAccessPolicyTestPageID + "/access_policies/email"
)

func newStatusPagesAPIForTest(server *httptest.Server) *datadogV2.StatusPagesApi {
	config := datadog.NewConfiguration()
	config.Servers = datadog.ServerConfigurations{{URL: server.URL}}
	config.OperationServers = nil
	config.HTTPClient = server.Client()
	return datadogV2.NewStatusPagesApi(datadog.NewAPIClient(config))
}

func statusPageEmailAccessPolicyTestData(id, pattern, effect string) string {
	return fmt.Sprintf(`{"id":%q,"type":"email_access_policies","attributes":{"created_at":"2026-01-01T00:00:00.000Z","effect":%q,"pattern":%q},"relationships":{"created_by_user":{"data":{"id":"00000000-0000-4000-8000-000000000104","type":"users"}},"status_page":{"data":{"id":%q,"type":"status_pages"}}}}`,
		id, effect, pattern, statusPageEmailAccessPolicyTestPageID)
}

func statusPageEmailAccessPolicyTestState(t *testing.T, r *statusPageEmailAccessPolicyResource, model statusPageEmailAccessPolicyModel) tfsdk.State {
	t.Helper()
	ctx := context.Background()
	var schemaResponse resource.SchemaResponse
	r.Schema(ctx, resource.SchemaRequest{}, &schemaResponse)
	state := tfsdk.State{Schema: schemaResponse.Schema}
	require.False(t, state.Set(ctx, model).HasError())
	return state
}

func TestStatusPageEmailAccessPolicyCreate(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		pattern string
		effect  string
	}{
		{"@example.com", "allow"},
		{"contractor@example.com", "deny"},
	} {
		t.Run(test.effect, func(t *testing.T) {
			ctx := context.Background()
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
				assert.Equal(t, http.MethodPost, request.Method)
				assert.Equal(t, statusPageEmailAccessPolicyTestPath, request.URL.Path)
				var body struct {
					Data struct {
						Type       string            `json:"type"`
						Attributes map[string]string `json:"attributes"`
					} `json:"data"`
				}
				assert.NoError(t, json.NewDecoder(request.Body).Decode(&body))
				assert.Equal(t, "email_access_policies", body.Data.Type)
				assert.Equal(t, map[string]string{"pattern": test.pattern, "effect": test.effect}, body.Data.Attributes)

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusCreated)
				fmt.Fprintf(w, `{"data":%s}`, statusPageEmailAccessPolicyTestData(statusPageEmailAccessPolicyTestPolicyID, test.pattern, test.effect))
			}))
			defer server.Close()

			r := &statusPageEmailAccessPolicyResource{Api: newStatusPagesAPIForTest(server), Auth: ctx}
			var schemaResponse resource.SchemaResponse
			r.Schema(ctx, resource.SchemaRequest{}, &schemaResponse)
			plan := tfsdk.Plan{Schema: schemaResponse.Schema}
			require.False(t, plan.Set(ctx, statusPageEmailAccessPolicyModel{
				ID:        types.StringUnknown(),
				PageID:    types.StringValue(statusPageEmailAccessPolicyTestPageID),
				Pattern:   types.StringValue(test.pattern),
				Effect:    types.StringValue(test.effect),
				CreatedAt: types.StringUnknown(),
			}).HasError())

			response := resource.CreateResponse{State: tfsdk.State{Schema: schemaResponse.Schema}}
			r.Create(ctx, resource.CreateRequest{Plan: plan}, &response)
			require.False(t, response.Diagnostics.HasError(), "%v", response.Diagnostics)

			var got statusPageEmailAccessPolicyModel
			require.False(t, response.State.Get(ctx, &got).HasError())
			require.Equal(t, statusPageEmailAccessPolicyModel{
				ID:        types.StringValue(statusPageEmailAccessPolicyTestPolicyID),
				PageID:    types.StringValue(statusPageEmailAccessPolicyTestPageID),
				Pattern:   types.StringValue(test.pattern),
				Effect:    types.StringValue(test.effect),
				CreatedAt: types.StringValue("2026-01-01T00:00:00Z"),
			}, got)
		})
	}
}

func TestStatusPageEmailAccessPolicyRead(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		name        string
		status      int
		body        string
		wantRemoved bool
		wantError   bool
	}{
		{
			name:   "found in list",
			status: http.StatusOK,
			body: fmt.Sprintf(`{"data":[%s,%s]}`,
				statusPageEmailAccessPolicyTestData(statusPageEmailAccessPolicyTestOtherID, "other@example.com", "allow"),
				statusPageEmailAccessPolicyTestData(statusPageEmailAccessPolicyTestPolicyID, "@example.com", "deny")),
		},
		{
			name:        "missing from list",
			status:      http.StatusOK,
			body:        fmt.Sprintf(`{"data":[%s]}`, statusPageEmailAccessPolicyTestData(statusPageEmailAccessPolicyTestOtherID, "other@example.com", "allow")),
			wantRemoved: true,
		},
		{
			name:        "status page not found",
			status:      http.StatusNotFound,
			body:        `{"errors":["Not Found"]}`,
			wantRemoved: true,
		},
		{
			name:      "server error",
			status:    http.StatusForbidden,
			body:      `{"errors":["Forbidden"]}`,
			wantError: true,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			ctx := context.Background()
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
				assert.Equal(t, http.MethodGet, request.Method)
				assert.Equal(t, statusPageEmailAccessPolicyTestPath, request.URL.Path)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(test.status)
				fmt.Fprint(w, test.body)
			}))
			defer server.Close()

			r := &statusPageEmailAccessPolicyResource{Api: newStatusPagesAPIForTest(server), Auth: ctx}
			state := statusPageEmailAccessPolicyTestState(t, r, statusPageEmailAccessPolicyModel{
				ID:        types.StringValue(statusPageEmailAccessPolicyTestPolicyID),
				PageID:    types.StringValue(statusPageEmailAccessPolicyTestPageID),
				Pattern:   types.StringValue("@example.com"),
				Effect:    types.StringValue("allow"),
				CreatedAt: types.StringValue("2026-01-01T00:00:00Z"),
			})

			response := resource.ReadResponse{State: state}
			r.Read(ctx, resource.ReadRequest{State: state}, &response)
			require.Equal(t, test.wantError, response.Diagnostics.HasError(), "%v", response.Diagnostics)
			if test.wantError {
				return
			}
			require.Equal(t, test.wantRemoved, response.State.Raw.IsNull())
			if test.wantRemoved {
				return
			}

			var got statusPageEmailAccessPolicyModel
			require.False(t, response.State.Get(ctx, &got).HasError())
			require.Equal(t, types.StringValue(statusPageEmailAccessPolicyTestPolicyID), got.ID)
			require.Equal(t, types.StringValue("@example.com"), got.Pattern)
			require.Equal(t, types.StringValue("deny"), got.Effect)
		})
	}
}

func TestStatusPageEmailAccessPolicyDelete(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		name      string
		status    int
		wantError bool
	}{
		{"deleted", http.StatusNoContent, false},
		{"already deleted", http.StatusNotFound, false},
		{"forbidden", http.StatusForbidden, true},
	} {
		t.Run(test.name, func(t *testing.T) {
			ctx := context.Background()
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
				assert.Equal(t, http.MethodDelete, request.Method)
				assert.Equal(t, statusPageEmailAccessPolicyTestPath+"/"+statusPageEmailAccessPolicyTestPolicyID, request.URL.Path)
				if test.status != http.StatusNoContent {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(test.status)
					fmt.Fprint(w, `{"errors":["error"]}`)
					return
				}
				w.WriteHeader(test.status)
			}))
			defer server.Close()

			r := &statusPageEmailAccessPolicyResource{Api: newStatusPagesAPIForTest(server), Auth: ctx}
			state := statusPageEmailAccessPolicyTestState(t, r, statusPageEmailAccessPolicyModel{
				ID:        types.StringValue(statusPageEmailAccessPolicyTestPolicyID),
				PageID:    types.StringValue(statusPageEmailAccessPolicyTestPageID),
				Pattern:   types.StringValue("@example.com"),
				Effect:    types.StringValue("allow"),
				CreatedAt: types.StringValue("2026-01-01T00:00:00Z"),
			})

			response := resource.DeleteResponse{State: state}
			r.Delete(ctx, resource.DeleteRequest{State: state}, &response)
			require.Equal(t, test.wantError, response.Diagnostics.HasError(), "%v", response.Diagnostics)
		})
	}
}

func TestStatusPageEmailAccessPolicyImportState(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	r := &statusPageEmailAccessPolicyResource{}
	var schemaResponse resource.SchemaResponse
	r.Schema(ctx, resource.SchemaRequest{}, &schemaResponse)
	emptyState := func() tfsdk.State {
		return tfsdk.State{
			Raw:    tftypes.NewValue(schemaResponse.Schema.Type().TerraformType(ctx), nil),
			Schema: schemaResponse.Schema,
		}
	}

	response := resource.ImportStateResponse{State: emptyState()}
	r.ImportState(ctx, resource.ImportStateRequest{ID: statusPageEmailAccessPolicyTestPageID + ":" + statusPageEmailAccessPolicyTestPolicyID}, &response)
	require.False(t, response.Diagnostics.HasError(), "%v", response.Diagnostics)
	var pageID, id types.String
	require.False(t, response.State.GetAttribute(ctx, path.Root("page_id"), &pageID).HasError())
	require.False(t, response.State.GetAttribute(ctx, path.Root("id"), &id).HasError())
	require.Equal(t, statusPageEmailAccessPolicyTestPageID, pageID.ValueString())
	require.Equal(t, statusPageEmailAccessPolicyTestPolicyID, id.ValueString())

	response = resource.ImportStateResponse{State: emptyState()}
	r.ImportState(ctx, resource.ImportStateRequest{ID: statusPageEmailAccessPolicyTestPolicyID}, &response)
	require.True(t, response.Diagnostics.HasError())
}
