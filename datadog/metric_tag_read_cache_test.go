package datadog

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	ddclient "github.com/DataDog/datadog-api-client-go/v2/api/datadog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

func newMetricTagResourceTestProvider(t *testing.T, handler http.HandlerFunc) (*ProviderConfiguration, func()) {
	t.Helper()
	server := httptest.NewServer(handler)
	config := ddclient.NewConfiguration()
	config.Servers = ddclient.ServerConfigurations{{URL: server.URL}}
	config.OperationServers = nil
	config.HTTPClient = server.Client()
	config.RetryConfiguration.EnableRetry = false
	return &ProviderConfiguration{Auth: context.Background(), DatadogApiInstances: &utils.ApiInstances{HttpClient: ddclient.NewAPIClient(config)}}, server.Close
}

func writeMetricTagResourceJSON(t *testing.T, w http.ResponseWriter, response interface{}) {
	t.Helper()
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		t.Error(err)
	}
}

func TestMetricTagBulkReadPreservesAggregationsAndRefreshesPolicy(t *testing.T) {
	var requests atomic.Int32
	provider, closeServer := newMetricTagResourceTestProvider(t, func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)
		if r.URL.Path != "/api/v2/metrics" {
			t.Errorf("unexpected single-resource request %s", r.URL.Path)
		}
		writeMetricTagResourceJSON(t, w, map[string]interface{}{
			"data": []interface{}{
				map[string]interface{}{"id": "metric.gauge", "type": "manage_tags", "attributes": map[string]interface{}{
					"metric_type": "gauge", "exclude_tags_mode": false, "tags": []string{"env", "service"}, "aggregations": []interface{}{map[string]string{"time": "sum", "space": "sum"}},
				}},
				map[string]interface{}{"id": "metric.distribution", "type": "manage_tags", "attributes": map[string]interface{}{
					"metric_type": "distribution", "exclude_tags_mode": true, "tags": []string{"env"}, "include_percentiles": true,
				}},
			},
			"meta": map[string]interface{}{"pagination": map[string]interface{}{"next_cursor": nil, "type": "cursor_limit"}},
		})
	})
	defer closeServer()
	resource := resourceDatadogMetricTagConfiguration()
	gauge := schema.TestResourceDataRaw(t, resource.SchemaMap(), map[string]interface{}{
		"metric_name": "metric.gauge", "metric_type": "gauge", "tags": []interface{}{"original"},
		"aggregations": []interface{}{map[string]interface{}{"time": "avg", "space": "avg"}},
	})
	gauge.SetId("metric.gauge")
	if diagnostics := resourceDatadogMetricTagConfigurationRead(context.Background(), gauge, provider); diagnostics.HasError() {
		t.Fatal(diagnostics)
	}
	aggregations := gauge.Get("aggregations").(*schema.Set).List()
	if len(aggregations) != 1 || aggregations[0].(map[string]interface{})["time"] != "avg" {
		t.Fatalf("legacy aggregations overwritten: %v", aggregations)
	}
	if gauge.Get("tags").(*schema.Set).Len() != 2 {
		t.Fatal("indexed tags did not refresh")
	}
	distribution := schema.TestResourceDataRaw(t, resource.SchemaMap(), map[string]interface{}{
		"metric_name": "metric.distribution", "metric_type": "distribution", "tags": []interface{}{"original"}, "include_percentiles": false,
	})
	distribution.SetId("metric.distribution")
	if diagnostics := resourceDatadogMetricTagConfigurationRead(context.Background(), distribution, provider); diagnostics.HasError() {
		t.Fatal(diagnostics)
	}
	if !distribution.Get("include_percentiles").(bool) || !distribution.Get("exclude_tags_mode").(bool) {
		t.Fatal("percentile or exclusion policy did not refresh")
	}
	if requests.Load() != 1 {
		t.Fatalf("requests = %d, want one bulk read", requests.Load())
	}
}

func TestMetricTagImportReadBypassesSnapshot(t *testing.T) {
	provider, closeServer := newMetricTagResourceTestProvider(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v2/metrics/imported.metric/tags" {
			t.Errorf("import used %s", r.URL.Path)
		}
		writeMetricTagResourceJSON(t, w, map[string]interface{}{"data": map[string]interface{}{
			"id": "imported.metric", "type": "manage_tags", "attributes": map[string]interface{}{
				"metric_type": "gauge", "exclude_tags_mode": false, "tags": []string{"env"}, "aggregations": []interface{}{map[string]string{"time": "avg", "space": "avg"}},
			},
		}})
	})
	defer closeServer()
	data := schema.TestResourceDataRaw(t, resourceDatadogMetricTagConfiguration().SchemaMap(), nil)
	data.SetId("imported.metric")
	if diagnostics := resourceDatadogMetricTagConfigurationRead(context.Background(), data, provider); diagnostics.HasError() {
		t.Fatal(diagnostics)
	}
	if data.Get("metric_type") != "gauge" || data.Get("aggregations").(*schema.Set).Len() != 1 {
		t.Fatal("imported state was not initialized")
	}
}

func TestMetricTagBulkMissRequiresIndividual404(t *testing.T) {
	var singles atomic.Int32
	provider, closeServer := newMetricTagResourceTestProvider(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v2/metrics" {
			writeMetricTagResourceJSON(t, w, map[string]interface{}{"data": []interface{}{}, "meta": map[string]interface{}{"pagination": map[string]interface{}{"next_cursor": nil, "type": "cursor_limit"}}})
			return
		}
		if singles.Add(1) > 1 {
			w.WriteHeader(404)
			writeMetricTagResourceJSON(t, w, map[string]interface{}{"errors": []string{"not found"}})
			return
		}
		writeMetricTagResourceJSON(t, w, map[string]interface{}{"data": map[string]interface{}{
			"id": "inactive.metric", "type": "manage_tags", "attributes": map[string]interface{}{"metric_type": "gauge", "exclude_tags_mode": false, "tags": []string{"env"}},
		}})
	})
	defer closeServer()
	data := schema.TestResourceDataRaw(t, resourceDatadogMetricTagConfiguration().SchemaMap(), map[string]interface{}{
		"metric_name": "inactive.metric", "metric_type": "gauge", "tags": []interface{}{"env"},
	})
	data.SetId("inactive.metric")
	for read := 0; read < 2; read++ {
		if diagnostics := resourceDatadogMetricTagConfigurationRead(context.Background(), data, provider); diagnostics.HasError() {
			t.Fatal(diagnostics)
		}
		if read == 0 && data.Id() == "" {
			t.Fatal("bulk absence removed an existing inactive metric")
		}
	}
	if data.Id() != "" || singles.Load() != 2 {
		t.Fatal("individual 404 did not remove the resource")
	}
}

func TestMetricTagUpdateInvalidatesBulkSnapshot(t *testing.T) {
	var updated atomic.Bool
	var bulkReads atomic.Int32
	provider, closeServer := newMetricTagResourceTestProvider(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPatch {
			updated.Store(true)
		}
		tags := []string{"env"}
		if updated.Load() {
			tags = []string{"service"}
		}
		config := map[string]interface{}{"id": "metric.gauge", "type": "manage_tags", "attributes": map[string]interface{}{
			"metric_type": "gauge", "exclude_tags_mode": false, "tags": tags,
		}}
		if r.URL.Path == "/api/v2/metrics" {
			bulkReads.Add(1)
			writeMetricTagResourceJSON(t, w, map[string]interface{}{"data": []interface{}{config}, "meta": map[string]interface{}{"pagination": map[string]interface{}{"next_cursor": nil, "type": "cursor_limit"}}})
		} else {
			writeMetricTagResourceJSON(t, w, map[string]interface{}{"data": config})
		}
	})
	defer closeServer()
	data := schema.TestResourceDataRaw(t, resourceDatadogMetricTagConfiguration().SchemaMap(), map[string]interface{}{
		"metric_name": "metric.gauge", "metric_type": "gauge", "tags": []interface{}{"env"},
	})
	data.SetId("metric.gauge")
	if diagnostics := resourceDatadogMetricTagConfigurationRead(context.Background(), data, provider); diagnostics.HasError() {
		t.Fatal(diagnostics)
	}
	if err := data.Set("tags", []string{"service"}); err != nil {
		t.Fatal(err)
	}
	if diagnostics := resourceDatadogMetricTagConfigurationUpdate(context.Background(), data, provider); diagnostics.HasError() {
		t.Fatal(diagnostics)
	}
	if diagnostics := resourceDatadogMetricTagConfigurationRead(context.Background(), data, provider); diagnostics.HasError() {
		t.Fatal(diagnostics)
	}
	if !data.Get("tags").(*schema.Set).Contains("service") || bulkReads.Load() != 2 {
		t.Fatalf("stale state after update: tags=%v, bulk reads=%d", data.Get("tags"), bulkReads.Load())
	}
}
