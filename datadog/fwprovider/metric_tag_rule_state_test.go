package fwprovider

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadog"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

type metricTagRuleRoundTripper func(*http.Request) (*http.Response, error)

func (f metricTagRuleRoundTripper) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func metricTagRuleResponse(status int, v interface{}) (*http.Response, error) {
	body, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	return &http.Response{StatusCode: status, Header: http.Header{"Content-Type": []string{"application/json"}}, Body: io.NopCloser(bytes.NewReader(body))}, nil
}

func metricTagRuleRule(id string, order int) map[string]interface{} {
	return map[string]interface{}{"id": id, "type": "tag_indexing_rules", "attributes": map[string]interface{}{
		"name": "n-" + id, "rule_order": order, "metric_name_matches": []string{id + ".*"}, "ignored_metric_name_matches": []string{id + ".skip"},
		"tags": []string{"env", "service"}, "exclude_tags_mode": true,
		"created_at": "2026-01-02T03:04:05Z", "modified_at": "2026-01-03T03:04:05Z", "created_by_handle": "a@example.com", "modified_by_handle": "b@example.com",
		"options": map[string]interface{}{"version": 1, "data": map[string]interface{}{"override_previous_rules": true, "manage_preexisting_metrics": false,
			"dynamic_tags": map[string]interface{}{"exclude_not_queried_window_seconds": 3600, "exclude_not_used_in_assets": true}}},
	}}
}

func metricTagRuleClient(transport http.RoundTripper) *utils.ApiInstances {
	config := datadog.NewConfiguration()
	config.Servers = datadog.ServerConfigurations{{URL: "https://example.invalid"}}
	config.OperationServers = nil
	config.HTTPClient = &http.Client{Transport: transport}
	config.RetryConfiguration.EnableRetry = false
	config.SetUnstableOperationEnabled("v2.ListTagIndexingRules", true)
	config.SetUnstableOperationEnabled("v2.GetTagIndexingRule", true)
	return &utils.ApiInstances{HttpClient: datadog.NewAPIClient(config)}
}

func metricTagRuleReadRule(t *testing.T, api *utils.ApiInstances, prior tagIndexingRuleModel) resource.ReadResponse {
	t.Helper()
	r := &tagIndexingRuleResource{Api: api.GetMetricsApiV2(), Auth: context.Background(), apiInstances: api}
	sr := resource.SchemaResponse{}
	r.Schema(context.Background(), resource.SchemaRequest{}, &sr)
	state := tfsdk.State{Schema: sr.Schema}
	if d := state.Set(context.Background(), &prior); d.HasError() {
		t.Fatal(d)
	}
	resp := resource.ReadResponse{State: state}
	r.Read(context.Background(), resource.ReadRequest{State: state}, &resp)
	return resp
}

func metricTagRulePrior(id string, imported bool) tagIndexingRuleModel {
	m := tagIndexingRuleModel{
		ID: types.StringValue(id), Name: types.StringNull(), RuleOrder: types.Int64Null(), ExcludeTagsMode: types.BoolNull(),
		MetricNameMatches: types.ListNull(types.StringType), IgnoredMetricNameMatches: types.ListNull(types.StringType), Tags: types.ListNull(types.StringType),
		CreatedAt: types.StringNull(), ModifiedAt: types.StringNull(), CreatedByHandle: types.StringNull(), ModifiedByHandle: types.StringNull(),
	}
	if !imported {
		m.CreatedAt = types.StringValue("old")
	}
	return m
}

// Same record served by list and by GET must yield identical framework state
// whether the read came from the snapshot or from the single-resource fallback.
func TestMetricTagRuleReadSnapshotMatchesSingleGet(t *testing.T) {
	for _, imported := range []bool{false, true} {
		var listOK atomic.Bool
		var gets atomic.Int32
		api := metricTagRuleClient(metricTagRuleRoundTripper(func(r *http.Request) (*http.Response, error) {
			if strings.HasSuffix(r.URL.Path, "/tag-indexing-rules") {
				if !listOK.Load() {
					return metricTagRuleResponse(500, map[string]interface{}{"errors": []string{"boom"}})
				}
				return metricTagRuleResponse(200, map[string]interface{}{"data": []interface{}{metricTagRuleRule("r1", 1)}, "meta": map[string]int{"total": 1}})
			}
			gets.Add(1)
			return metricTagRuleResponse(200, map[string]interface{}{"data": metricTagRuleRule("r1", 1)})
		}))
		fallback := metricTagRuleReadRule(t, api, metricTagRulePrior("r1", imported))
		if fallback.Diagnostics.HasError() || gets.Load() != 1 {
			t.Fatalf("fallback: diags=%v gets=%d", fallback.Diagnostics, gets.Load())
		}
		listOK.Store(true)
		api.InvalidateMetricTagReadCaches()
		cached := metricTagRuleReadRule(t, api, metricTagRulePrior("r1", imported))
		if cached.Diagnostics.HasError() || gets.Load() != 1 {
			t.Fatalf("cached: diags=%v gets=%d", cached.Diagnostics, gets.Load())
		}
		if !cached.State.Raw.Equal(fallback.State.Raw) {
			t.Fatalf("imported=%v state differs\ncached=%s\nfallback=%s", imported, cached.State.Raw.String(), fallback.State.Raw.String())
		}
		var m tagIndexingRuleModel
		if d := cached.State.Get(context.Background(), &m); d.HasError() {
			t.Fatal(d)
		}
		if imported && (m.Options == nil || m.Options.Data == nil || m.Options.Data.DynamicTags == nil) {
			t.Fatalf("imported read via snapshot dropped options: %+v", m.Options)
		}
		if m.RuleOrder.ValueInt64() != 1 || !m.ExcludeTagsMode.ValueBool() || m.CreatedByHandle.ValueString() != "a@example.com" {
			t.Fatalf("unexpected state %+v", m)
		}
		t.Logf("imported=%v identical state via snapshot and GET: %s", imported, m.CreatedAt.ValueString())
	}
}

// Snapshot miss must consult the single endpoint, and its 404 must remove state.
func TestMetricTagRuleReadMissFallsBackAnd404Removes(t *testing.T) {
	var gets atomic.Int32
	api := metricTagRuleClient(metricTagRuleRoundTripper(func(r *http.Request) (*http.Response, error) {
		if strings.HasSuffix(r.URL.Path, "/tag-indexing-rules") {
			return metricTagRuleResponse(200, map[string]interface{}{"data": []interface{}{metricTagRuleRule("other", 1)}, "meta": map[string]int{"total": 1}})
		}
		gets.Add(1)
		return metricTagRuleResponse(404, map[string]interface{}{"errors": []string{"not found"}})
	}))
	resp := metricTagRuleReadRule(t, api, metricTagRulePrior("gone", false))
	if resp.Diagnostics.HasError() || gets.Load() != 1 || !resp.State.Raw.IsNull() {
		t.Fatalf("diags=%v gets=%d null=%v", resp.Diagnostics, gets.Load(), resp.State.Raw.IsNull())
	}
}
