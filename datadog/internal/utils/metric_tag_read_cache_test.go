package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadog"
)

func newMetricTagReadTestClient(t *testing.T, handler http.HandlerFunc) (*ApiInstances, func()) {
	t.Helper()
	server := httptest.NewServer(handler)
	config := datadog.NewConfiguration()
	config.Servers = datadog.ServerConfigurations{{URL: server.URL}}
	config.OperationServers = nil
	config.HTTPClient = server.Client()
	config.RetryConfiguration.EnableRetry = false
	config.SetUnstableOperationEnabled("v2.ListTagIndexingRules", true)
	config.SetUnstableOperationEnabled("v2.GetTagIndexingRule", true)
	return &ApiInstances{HttpClient: datadog.NewAPIClient(config)}, server.Close
}

func writeMetricTagJSON(t *testing.T, w http.ResponseWriter, data interface{}) {
	t.Helper()
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(data); err != nil {
		t.Errorf("encoding response: %v", err)
	}
}

func testIndexingRule(id string, order int) map[string]interface{} {
	return map[string]interface{}{"id": id, "type": "tag_indexing_rules", "attributes": map[string]interface{}{
		"name": id, "rule_order": order, "metric_name_matches": []string{id}, "ignored_metric_name_matches": []string{}, "tags": []string{"env"}, "exclude_tags_mode": false,
	}}
}

func testMetricTagConfig(name, metricType string) map[string]interface{} {
	return map[string]interface{}{"id": name, "type": "manage_tags", "attributes": map[string]interface{}{
		"metric_type": metricType, "tags": []string{"env", "service"}, "exclude_tags_mode": false, "include_percentiles": true,
	}}
}

func TestMetricTagRuleSnapshotPaginationAndConcurrency(t *testing.T) {
	var requests atomic.Int32
	client, closeServer := newMetricTagReadTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)
		if r.URL.Query().Get("page[limit]") != "1000" {
			t.Errorf("page limit = %q", r.URL.Query().Get("page[limit]"))
		}
		offset, err := strconv.Atoi(r.URL.Query().Get("page[offset]"))
		if err != nil {
			t.Error(err)
			return
		}
		page := []interface{}{}
		for idx := offset; idx < offset+2 && idx < 5; idx++ {
			page = append(page, testIndexingRule(fmt.Sprintf("rule-%d", idx), 5-idx))
		}
		writeMetricTagJSON(t, w, map[string]interface{}{"data": page, "meta": map[string]int{"total": 5}})
	})
	defer closeServer()
	var wg sync.WaitGroup
	for range 20 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			rules, err := client.ListTagIndexingRulesSnapshot(context.Background())
			if err != nil || len(rules) != 5 {
				t.Errorf("rules = %d, err = %v", len(rules), err)
				return
			}
			if rules[0].GetId() != "rule-4" {
				t.Errorf("first rule = %s", rules[0].GetId())
			}
			// Sorting a caller's returned slice must not alter another reader.
			rules[0], rules[4] = rules[4], rules[0]
		}()
	}
	wg.Wait()
	if got := requests.Load(); got != 3 {
		t.Fatalf("requests = %d, want 3 pages for all readers", got)
	}
	response, _, err := client.ReadTagIndexingRule(context.Background(), "rule-2")
	if err != nil || response.Data.GetId() != "rule-2" || requests.Load() != 3 {
		t.Fatalf("cached rule read = %v, requests = %d", err, requests.Load())
	}
}

func TestMetricTagRuleSnapshotRejectsIncompletePagesAndFallsBack(t *testing.T) {
	for _, failure := range []string{"empty", "duplicate", "total-change", "unparsed", "server-error"} {
		t.Run(failure, func(t *testing.T) {
			var bulk, single atomic.Int32
			client, closeServer := newMetricTagReadTestClient(t, func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/api/v2/metrics/tag-indexing-rules" {
					single.Add(1)
					w.WriteHeader(http.StatusNotFound)
					writeMetricTagJSON(t, w, map[string]interface{}{"errors": []string{"not found"}})
					return
				}
				call := bulk.Add(1)
				if failure == "server-error" {
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
				total := 2
				page := []interface{}{testIndexingRule("rule-1", 1)}
				if call > 1 {
					switch failure {
					case "empty":
						page = []interface{}{}
					case "duplicate":
					case "total-change":
						total = 3
					case "unparsed":
						page = []interface{}{map[string]interface{}{"id": "rule-2", "type": "future_type"}}
					}
				}
				writeMetricTagJSON(t, w, map[string]interface{}{"data": page, "meta": map[string]int{"total": total}})
			})
			defer closeServer()
			if _, err := client.ListTagIndexingRulesSnapshot(context.Background()); err == nil {
				t.Fatal("incomplete snapshot accepted")
			}
			_, response, err := client.ReadTagIndexingRule(context.Background(), "missing")
			if err == nil || response == nil || response.StatusCode != 404 || single.Load() != 1 {
				t.Fatalf("fallback response = %v, err = %v", response, err)
			}
			if bulk.Load() > 2 {
				t.Fatal("failed bulk request was repeated for each resource")
			}
		})
	}
}

func TestMetricTagConfigurationsSnapshotPaginationAndUnsupportedItems(t *testing.T) {
	var requests atomic.Int32
	client, closeServer := newMetricTagReadTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)
		if r.URL.Query().Get("filter[configured]") != "true" || r.URL.Query().Get("window[seconds]") != "604800" || r.URL.Query().Get("page[size]") != "10000" {
			t.Error("bulk metric filters missing")
		}
		page := []interface{}{testMetricTagConfig("metric.gauge", "gauge"), testMetricTagConfig("metric.future", "future_type")}
		var next interface{} = "next-page"
		if r.URL.Query().Get("page[cursor]") != "" {
			page = []interface{}{testMetricTagConfig("metric.distribution", "distribution")}
			next = nil
		}
		writeMetricTagJSON(t, w, map[string]interface{}{"data": page, "meta": map[string]interface{}{"pagination": map[string]interface{}{"next_cursor": next, "type": "cursor_limit"}}})
	})
	defer closeServer()
	for _, name := range []string{"metric.gauge", "metric.distribution"} {
		config, ok := client.CachedMetricTagConfiguration(context.Background(), name)
		if !ok || config.GetId() != name {
			t.Fatalf("missing config %s", name)
		}
	}
	if _, ok := client.CachedMetricTagConfiguration(context.Background(), "metric.future"); ok {
		t.Fatal("unsupported metric returned from cache")
	}
	if _, ok := client.CachedMetricTagConfiguration(context.Background(), "inactive.metric"); ok {
		t.Fatal("inactive metric returned from cache")
	}
	if got := requests.Load(); got != 2 {
		t.Fatalf("requests = %d, want 2 pages", got)
	}
}

func TestMetricTagConfigurationsSnapshotRejectsMalformedPagination(t *testing.T) {
	for _, failure := range []string{"missing-meta", "cursor-loop", "server-error", "duplicate"} {
		t.Run(failure, func(t *testing.T) {
			var requests atomic.Int32
			client, closeServer := newMetricTagReadTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
				requests.Add(1)
				if failure == "server-error" {
					w.WriteHeader(500)
					return
				}
				page := []interface{}{testMetricTagConfig("metric.gauge", "gauge")}
				if failure == "cursor-loop" && requests.Load() > 1 {
					page = []interface{}{}
				}
				response := map[string]interface{}{"data": page}
				if failure != "missing-meta" {
					response["meta"] = map[string]interface{}{"pagination": map[string]interface{}{"next_cursor": "repeated", "type": "cursor_limit"}}
				}
				writeMetricTagJSON(t, w, response)
			})
			defer closeServer()
			for range 2 {
				if _, ok := client.CachedMetricTagConfiguration(context.Background(), "metric.gauge"); ok {
					t.Fatal("partial snapshot was published")
				}
			}
			if got := requests.Load(); got > 2 {
				t.Fatalf("repeated failed bulk calls: %d", got)
			}
		})
	}
}

func TestMetricTagSnapshotsInvalidationAcrossClients(t *testing.T) {
	var ruleRequests, metricRequests atomic.Int32
	rules, closeRules := newMetricTagReadTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		ruleRequests.Add(1)
		writeMetricTagJSON(t, w, map[string]interface{}{"data": []interface{}{testIndexingRule("rule-1", 1)}, "meta": map[string]int{"total": 1}})
	})
	defer closeRules()
	metrics, closeMetrics := newMetricTagReadTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		metricRequests.Add(1)
		writeMetricTagJSON(t, w, map[string]interface{}{"data": []interface{}{testMetricTagConfig("metric.gauge", "gauge")}, "meta": map[string]interface{}{"pagination": map[string]interface{}{"next_cursor": nil, "type": "cursor_limit"}}})
	})
	defer closeMetrics()
	for range 2 {
		if _, err := rules.ListTagIndexingRulesSnapshot(context.Background()); err != nil {
			t.Fatal(err)
		}
		if _, ok := metrics.CachedMetricTagConfiguration(context.Background(), "metric.gauge"); !ok {
			t.Fatal("missing gauge snapshot")
		}
	}
	rules.InvalidateMetricTagReadCaches()
	if _, err := rules.ListTagIndexingRulesSnapshot(context.Background()); err != nil {
		t.Fatal(err)
	}
	if _, ok := metrics.CachedMetricTagConfiguration(context.Background(), "metric.gauge"); !ok {
		t.Fatal("missing refreshed gauge snapshot")
	}
	if ruleRequests.Load() != 2 || metricRequests.Load() != 2 {
		t.Fatalf("rule requests %d, metric requests %d", ruleRequests.Load(), metricRequests.Load())
	}
}

func TestMetricTagSnapshotInvalidationDuringFill(t *testing.T) {
	started, release := make(chan struct{}), make(chan struct{})
	var requests atomic.Int32
	client, closeServer := newMetricTagReadTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		if requests.Add(1) == 1 {
			close(started)
			<-release
		}
		writeMetricTagJSON(t, w, map[string]interface{}{"data": []interface{}{testIndexingRule("rule-1", 1)}, "meta": map[string]int{"total": 1}})
	})
	defer closeServer()
	done := make(chan error, 1)
	go func() {
		_, err := client.ListTagIndexingRulesSnapshot(context.Background())
		done <- err
	}()
	<-started
	client.InvalidateMetricTagReadCaches()
	close(release)
	if err := <-done; err == nil {
		t.Fatal("stale in-flight snapshot was published")
	}
	if _, err := client.ListTagIndexingRulesSnapshot(context.Background()); err != nil {
		t.Fatal(err)
	}
	if requests.Load() != 2 {
		t.Fatal("snapshot was not refreshed after invalidation")
	}
}
