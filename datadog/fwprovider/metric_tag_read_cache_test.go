package fwprovider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadog"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

func TestMetricTagOrderReadIncludesEveryPage(t *testing.T) {
	var requests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		call := requests.Add(1)
		if request.URL.Query().Get("page[offset]") != fmt.Sprint(call-1) {
			t.Errorf("offset = %s", request.URL.Query().Get("page[offset]"))
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(map[string]interface{}{
			"data": []interface{}{map[string]interface{}{"id": fmt.Sprintf("rule-%d", call), "type": "tag_indexing_rules", "attributes": map[string]interface{}{"name": "test", "rule_order": call}}},
			"meta": map[string]int{"total": 2},
		}); err != nil {
			t.Error(err)
		}
	}))
	defer server.Close()
	config := datadog.NewConfiguration()
	config.Servers = datadog.ServerConfigurations{{URL: server.URL}}
	config.OperationServers = nil
	config.HTTPClient = server.Client()
	config.SetUnstableOperationEnabled("v2.ListTagIndexingRules", true)
	apiInstances := &utils.ApiInstances{HttpClient: datadog.NewAPIClient(config)}
	r := &tagIndexingRuleOrderResource{Api: apiInstances.GetMetricsApiV2(), Auth: context.Background(), apiInstances: apiInstances}
	schemaResponse := resource.SchemaResponse{}
	r.Schema(context.Background(), resource.SchemaRequest{}, &schemaResponse)
	state := tfsdk.State{Schema: schemaResponse.Schema}
	model := tagIndexingRuleOrderModel{ID: types.StringValue("main"), Name: types.StringValue("main"), RuleIDs: types.ListNull(types.StringType)}
	if diagnostics := state.Set(context.Background(), &model); diagnostics.HasError() {
		t.Fatal(diagnostics)
	}
	response := resource.ReadResponse{State: state}
	r.Read(context.Background(), resource.ReadRequest{State: state}, &response)
	if response.Diagnostics.HasError() {
		t.Fatal(response.Diagnostics)
	}
	if diagnostics := response.State.Get(context.Background(), &model); diagnostics.HasError() {
		t.Fatal(diagnostics)
	}
	var ids []string
	if diagnostics := model.RuleIDs.ElementsAs(context.Background(), &ids, false); diagnostics.HasError() {
		t.Fatal(diagnostics)
	}
	if len(ids) != 2 || ids[0] != "rule-1" || ids[1] != "rule-2" || requests.Load() != 2 {
		t.Fatalf("ids = %v, requests = %d", ids, requests.Load())
	}
}
