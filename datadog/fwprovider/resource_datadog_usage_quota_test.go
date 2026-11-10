package fwprovider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadog"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

func quotaTestResource(t *testing.T, handler http.HandlerFunc) *usageQuotaResource {
	t.Helper()
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)
	cfg := datadog.NewConfiguration()
	cfg.Servers = datadog.ServerConfigurations{{URL: server.URL}}
	cfg.RetryConfiguration.EnableRetry = true
	r := &usageQuotaResource{}
	var response resource.ConfigureResponse
	r.Configure(context.Background(), resource.ConfigureRequest{ProviderData: &FrameworkProvider{Auth: context.Background(), DatadogApiInstances: &utils.ApiInstances{HttpClient: datadog.NewAPIClient(cfg)}}}, &response)
	require.False(t, response.Diagnostics.HasError(), "%v", response.Diagnostics)
	require.True(t, cfg.RetryConfiguration.EnableRetry, "resource must not mutate shared retry configuration")
	return r
}
func quotaTestState(t *testing.T, m usageQuotaModel) tfsdk.State {
	t.Helper()
	state := tfsdk.State{Schema: usageQuotaSchema()}
	require.False(t, state.Set(context.Background(), &m).HasError())
	return state
}
func quotaTestModel() usageQuotaModel {
	return usageQuotaModel{ID: types.StringValue("ai_credits/quota-1"), QuotaID: types.StringValue("quota-1"), QuotaNamespace: types.StringValue("ai_credits"), Scope: types.MapNull(types.StringType), UsageLimit: types.Float64Value(100), Enforced: types.BoolValue(true), OrgPublicID: types.StringValue("org-1")}
}
func quotaTestReply(w http.ResponseWriter, bulk bool, enforced bool) {
	w.Header().Set("Content-Type", "application/json")
	data := fmt.Sprintf(`{"id":"quota-1","type":"quotas","attributes":{"usage_limit":100,"enforced":%t,"org_public_id":"org-1","scope":{}}}`, enforced)
	if bulk {
		data = "[" + data + "]"
	}
	fmt.Fprintf(w, `{"data":%s}`, data)
}

func TestUsageQuotaLifecycle(t *testing.T) {
	ctx := context.Background()
	methods := []string{}
	r := quotaTestResource(t, func(w http.ResponseWriter, req *http.Request) {
		methods = append(methods, req.Method)
		require.Equal(t, "false", req.URL.Query().Get("include_descendants"))
		switch req.Method {
		case "POST":
			require.Equal(t, "true", req.URL.Query().Get("create_only"))
			require.Equal(t, "/api/v2/usage/quotas/ai_credits", req.URL.Path)
			quotaTestReply(w, true, true)
		case "PATCH":
			var body map[string]interface{}
			require.NoError(t, json.NewDecoder(req.Body).Decode(&body))
			attrs := body["data"].(map[string]interface{})["attributes"].(map[string]interface{})
			require.Equal(t, map[string]interface{}{"enforced": false}, attrs)
			quotaTestReply(w, false, false)
		case "GET":
			quotaTestReply(w, false, true)
		case "DELETE":
			w.WriteHeader(204)
		default:
			t.Fatalf("unexpected method %s", req.Method)
		}
	})
	m := quotaTestModel()
	m.ID = types.StringUnknown()
	m.QuotaID = types.StringUnknown()
	m.OrgPublicID = types.StringUnknown()
	plan := quotaTestState(t, m)
	created := resource.CreateResponse{State: tfsdk.State{Schema: usageQuotaSchema()}}
	r.Create(ctx, resource.CreateRequest{Plan: tfsdk.Plan{Schema: usageQuotaSchema(), Raw: plan.Raw}}, &created)
	require.False(t, created.Diagnostics.HasError(), "%v", created.Diagnostics)
	var got usageQuotaModel
	require.False(t, created.State.Get(ctx, &got).HasError())
	require.Equal(t, "ai_credits/quota-1", got.ID.ValueString())
	require.True(t, got.Scope.IsNull())
	read := resource.ReadResponse{State: created.State}
	r.Read(ctx, resource.ReadRequest{State: created.State}, &read)
	require.False(t, read.Diagnostics.HasError(), "%v", read.Diagnostics)
	got.Enforced = types.BoolValue(false)
	updatedPlan := quotaTestState(t, got)
	updated := resource.UpdateResponse{State: read.State}
	r.Update(ctx, resource.UpdateRequest{State: read.State, Plan: tfsdk.Plan{Schema: usageQuotaSchema(), Raw: updatedPlan.Raw}}, &updated)
	require.False(t, updated.Diagnostics.HasError(), "%v", updated.Diagnostics)
	deleted := resource.DeleteResponse{State: updated.State}
	r.Delete(ctx, resource.DeleteRequest{State: updated.State}, &deleted)
	require.False(t, deleted.Diagnostics.HasError(), "%v", deleted.Diagnostics)
	require.Equal(t, []string{"POST", "GET", "PATCH", "DELETE"}, methods)
}

func TestUsageQuotaCreateFailuresNeverAdoptOrRetry(t *testing.T) {
	for _, status := range []int{409, 500} {
		t.Run(fmt.Sprint(status), func(t *testing.T) {
			calls := 0
			r := quotaTestResource(t, func(w http.ResponseWriter, req *http.Request) {
				calls++
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(status)
				fmt.Fprint(w, `{"errors":[{"status":"409","title":"Conflict","detail":"exists","meta":{"quota_id":"other"}}]}`)
			})
			m := quotaTestModel()
			m.ID = types.StringUnknown()
			m.QuotaID = types.StringUnknown()
			p := quotaTestState(t, m)
			resp := resource.CreateResponse{State: tfsdk.State{Schema: usageQuotaSchema()}}
			r.Create(context.Background(), resource.CreateRequest{Plan: tfsdk.Plan{Schema: usageQuotaSchema(), Raw: p.Raw}}, &resp)
			require.True(t, resp.Diagnostics.HasError())
			require.Equal(t, 1, calls)
			require.True(t, resp.State.Raw.IsNull())
		})
	}
}

func TestUsageQuotaReadDeleteMissingAndForbidden(t *testing.T) {
	for _, status := range []int{403, 404} {
		t.Run(fmt.Sprint(status), func(t *testing.T) {
			r := quotaTestResource(t, func(w http.ResponseWriter, req *http.Request) {
				require.Equal(t, "false", req.URL.Query().Get("include_descendants"))
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(status)
				fmt.Fprint(w, `{"errors":["unavailable"]}`)
			})
			state := quotaTestState(t, quotaTestModel())
			read := resource.ReadResponse{State: state}
			r.Read(context.Background(), resource.ReadRequest{State: state}, &read)
			require.Equal(t, status == 403, read.Diagnostics.HasError())
			require.Equal(t, status == 404, read.State.Raw.IsNull())
			del := resource.DeleteResponse{State: state}
			r.Delete(context.Background(), resource.DeleteRequest{State: state}, &del)
			require.Equal(t, status == 403, del.Diagnostics.HasError())
		})
	}
}

func TestUsageQuotaImport(t *testing.T) {
	r := &usageQuotaResource{}
	resp := resource.ImportStateResponse{State: quotaTestState(t, usageQuotaModel{Scope: types.MapNull(types.StringType)})}
	r.ImportState(context.Background(), resource.ImportStateRequest{ID: "ai_credits/opaque-id"}, &resp)
	require.False(t, resp.Diagnostics.HasError(), "%v", resp.Diagnostics)
	var m usageQuotaModel
	require.False(t, resp.State.Get(context.Background(), &m).HasError())
	require.Equal(t, "opaque-id", m.QuotaID.ValueString())
	require.Equal(t, "ai_credits", m.QuotaNamespace.ValueString())
}

func TestUsageQuotaPendingLimitWarning(t *testing.T) {
	r := quotaTestResource(t, func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"data":{"id":"quota-1","type":"quotas","attributes":{"usage_limit":100,"enforced":true,"org_public_id":"org-1","pending_usage_limit":50,"scope":{"role_uuid":"role-1"}}}}`)
	})
	state := quotaTestState(t, quotaTestModel())
	resp := resource.ReadResponse{State: state}
	r.Read(context.Background(), resource.ReadRequest{State: state}, &resp)
	require.False(t, resp.Diagnostics.HasError())
	require.Equal(t, 1, resp.Diagnostics.WarningsCount())
	var m usageQuotaModel
	require.False(t, resp.State.Get(context.Background(), &m).HasError())
	require.Equal(t, 100.0, m.UsageLimit.ValueFloat64())
	require.Equal(t, 1, len(m.Scope.Elements()))
}

func TestUsageQuotaCreateItemError(t *testing.T) {
	r := quotaTestResource(t, func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"data":[{"id":"quota-1","type":"quotas","attributes":{"error":"write failed"}}]}`)
	})
	p := quotaTestState(t, quotaTestModel())
	resp := resource.CreateResponse{State: tfsdk.State{Schema: usageQuotaSchema()}}
	r.Create(context.Background(), resource.CreateRequest{Plan: tfsdk.Plan{Schema: usageQuotaSchema(), Raw: p.Raw}}, &resp)
	require.True(t, resp.Diagnostics.HasError())
	require.True(t, resp.State.Raw.IsNull())
}
