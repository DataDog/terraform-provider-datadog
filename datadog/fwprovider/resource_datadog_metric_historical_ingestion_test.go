package fwprovider

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadog"
	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const historicalTestName = "test.metric_historical"
const historicalTestPath = "/api/v2/metrics/historical-metrics-configurations"
const historicalTestResponse = `{"data":{"id":"test.metric_historical","type":"historical_metrics_configurations","attributes":{"created_at":"2024-01-02T03:04:05Z"}}}`

func historicalTestResource(t *testing.T, handler http.HandlerFunc) *metricHistoricalIngestionResource {
	t.Helper()
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)
	cfg := datadog.NewConfiguration()
	cfg.Servers = datadog.ServerConfigurations{{URL: server.URL}}
	cfg.OperationServers = nil
	cfg.HTTPClient = server.Client()
	for _, name := range []string{"Create", "Get", "Delete"} {
		cfg.SetUnstableOperationEnabled("v2."+name+"HistoricalMetricsConfiguration", true)
	}
	auth := context.WithValue(context.Background(), datadog.ContextAPIKeys, map[string]datadog.APIKey{"apiKeyAuth": {Key: "example-api-key"}, "appKeyAuth": {Key: "example-app-key"}})
	return &metricHistoricalIngestionResource{api: datadogV2.NewMetricsApi(datadog.NewAPIClient(cfg)), auth: auth, timeout: 5 * time.Second}
}

func historicalTestState(t *testing.T, r *metricHistoricalIngestionResource, model metricHistoricalIngestionModel) tfsdk.State {
	t.Helper()
	var sr resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &sr)
	state := tfsdk.State{Schema: sr.Schema}
	if d := state.Set(context.Background(), &model); d.HasError() {
		t.Fatal(d)
	}
	return state
}
func historicalTestModel() metricHistoricalIngestionModel {
	return metricHistoricalIngestionModel{ID: types.StringValue(historicalTestName), MetricName: types.StringValue(historicalTestName), CreatedAt: types.StringNull()}
}
func historicalTestCreate(t *testing.T, ctx context.Context, r *metricHistoricalIngestionResource) resource.CreateResponse {
	t.Helper()
	model := historicalTestModel()
	model.ID = types.StringUnknown()
	model.CreatedAt = types.StringUnknown()
	s := historicalTestState(t, r, model)
	plan := tfsdk.Plan{Schema: s.Schema, Raw: s.Raw}
	resp := resource.CreateResponse{State: s}
	r.Create(ctx, resource.CreateRequest{Plan: plan}, &resp)
	return resp
}

func TestMetricHistoricalIngestionCreate(t *testing.T) {
	for _, tc := range []struct {
		name        string
		postStatus  int
		postBody    string
		getStatuses []int
		wantError   bool
		wantID      bool
		wantGet     int
	}{
		{"created with stale read", 201, historicalTestResponse, []int{404, 200}, false, true, 2},
		{"already enabled", 200, historicalTestResponse, []int{200}, false, true, 1},
		{"unexpected success", 202, historicalTestResponse, nil, true, false, 0},
		{"forbidden read", 201, historicalTestResponse, []int{403}, true, true, 1},
		{"invalid response identity", 201, `{"data":{"id":"other","type":"historical_metrics_configurations"}}`, nil, true, true, 0},
		{"missing type", 201, `{"data":{"id":"test.metric_historical"}}`, nil, true, true, 0},
		{"invalid timestamp", 201, `{"data":{"id":"test.metric_historical","type":"historical_metrics_configurations","attributes":{"created_at":"not-a-time"}}}`, nil, true, true, 0},
		{"unparsed type", 201, `{"data":{"id":"test.metric_historical","type":"unknown"}}`, nil, true, true, 0},
		{"malformed json", 201, `{`, nil, true, true, 0},
		{"write forbidden", 403, `{"errors":[{"detail":"forbidden"}]}`, nil, true, false, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var gets atomic.Int32
			r := historicalTestResource(t, func(w http.ResponseWriter, req *http.Request) {
				if req.Header.Get("DD-API-KEY") != "example-api-key" || req.Header.Get("DD-APPLICATION-KEY") != "example-app-key" {
					t.Error("provider credentials were not forwarded")
				}
				w.Header().Set("Content-Type", "application/json")
				switch req.Method {
				case http.MethodPost:
					if req.URL.Path != historicalTestPath {
						t.Errorf("path %s", req.URL.Path)
					}
					b, _ := io.ReadAll(req.Body)
					if !strings.Contains(string(b), `"id":"`+historicalTestName+`"`) || !strings.Contains(string(b), `"type":"historical_metrics_configurations"`) {
						t.Errorf("bad request %s", b)
					}
					w.WriteHeader(tc.postStatus)
					fmt.Fprint(w, tc.postBody)
				case http.MethodGet:
					if req.URL.Path != historicalTestPath+"/"+historicalTestName {
						t.Errorf("path %s", req.URL.Path)
					}
					i := int(gets.Add(1)) - 1
					if i >= len(tc.getStatuses) {
						t.Error("unexpected GET")
						w.WriteHeader(500)
						return
					}
					w.WriteHeader(tc.getStatuses[i])
					if tc.getStatuses[i] == 200 {
						fmt.Fprint(w, historicalTestResponse)
					} else {
						fmt.Fprint(w, `{"errors":[{"detail":"not visible"}]}`)
					}
				default:
					t.Errorf("unexpected method %s", req.Method)
				}
			})
			resp := historicalTestCreate(t, context.Background(), r)
			if resp.Diagnostics.HasError() != tc.wantError {
				t.Fatalf("diagnostics: %v", resp.Diagnostics)
			}
			if int(gets.Load()) != tc.wantGet {
				t.Fatalf("GET count %d, want %d", gets.Load(), tc.wantGet)
			}
			var state metricHistoricalIngestionModel
			if d := resp.State.Get(context.Background(), &state); d.HasError() {
				t.Fatal(d)
			}
			if (state.ID.ValueString() == historicalTestName) != tc.wantID {
				t.Fatalf("state id %v", state.ID)
			}
			if (!tc.wantError || tc.name == "forbidden read") && state.CreatedAt.ValueString() != "2024-01-02T03:04:05Z" {
				t.Fatalf("created_at %v", state.CreatedAt)
			}
		})
	}
}

func TestMetricHistoricalIngestionRead(t *testing.T) {
	for _, tc := range []struct {
		name      string
		status    int
		body      string
		removed   bool
		wantError bool
	}{
		{"import id only", 200, historicalTestResponse, false, false},
		{"optional attributes absent", 200, `{"data":{"id":"test.metric_historical","type":"historical_metrics_configurations"}}`, false, false},
		{"drift absent", 404, `{"errors":[]}`, true, false},
		{"permission", 403, `{"errors":["forbidden"]}`, false, true},
		{"backend error", 500, `{"errors":["backend failure"]}`, false, true},
		{"unexpected success", 202, historicalTestResponse, false, true},
		{"invalid read", 200, `{"data":{}}`, false, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var count atomic.Int32
			r := historicalTestResource(t, func(w http.ResponseWriter, req *http.Request) {
				count.Add(1)
				if req.Method != "GET" || req.URL.Path != historicalTestPath+"/"+historicalTestName {
					t.Errorf("unexpected request %s %s", req.Method, req.URL.Path)
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tc.status)
				fmt.Fprint(w, tc.body)
			})
			m := historicalTestModel()
			m.MetricName = types.StringNull()
			m.CreatedAt = types.StringNull()
			s := historicalTestState(t, r, m)
			resp := resource.ReadResponse{State: s}
			r.Read(context.Background(), resource.ReadRequest{State: s}, &resp)
			if resp.Diagnostics.HasError() != tc.wantError || resp.State.Raw.IsNull() != tc.removed || count.Load() != 1 {
				t.Fatalf("diagnostics %v removed %v requests %d", resp.Diagnostics, resp.State.Raw.IsNull(), count.Load())
			}
			if !tc.removed && !tc.wantError {
				var got metricHistoricalIngestionModel
				if d := resp.State.Get(context.Background(), &got); d.HasError() {
					t.Fatal(d)
				}
				if got.MetricName.ValueString() != historicalTestName {
					t.Fatalf("import name %v", got.MetricName)
				}
				if tc.name == "optional attributes absent" && !got.CreatedAt.IsNull() {
					t.Fatalf("created_at %v", got.CreatedAt)
				}
			}
		})
	}
}

func TestMetricHistoricalIngestionDelete(t *testing.T) {
	for _, tc := range []struct {
		name      string
		gets      []int
		wantError bool
	}{
		{"eventual absence", []int{200, 404}, false},
		{"already absent", []int{404}, false},
		{"permission failure", []int{403}, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var gets, deletes atomic.Int32
			r := historicalTestResource(t, func(w http.ResponseWriter, req *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				if req.Method == "DELETE" {
					deletes.Add(1)
					w.WriteHeader(204)
					return
				}
				i := int(gets.Add(1)) - 1
				if i >= len(tc.gets) {
					t.Error("too many reads")
					w.WriteHeader(500)
					return
				}
				w.WriteHeader(tc.gets[i])
				if tc.gets[i] == 200 {
					fmt.Fprint(w, historicalTestResponse)
				} else {
					fmt.Fprint(w, `{"errors":["error"]}`)
				}
			})
			s := historicalTestState(t, r, historicalTestModel())
			resp := resource.DeleteResponse{State: s}
			r.Delete(context.Background(), resource.DeleteRequest{State: s}, &resp)
			if resp.Diagnostics.HasError() != tc.wantError || deletes.Load() != 1 || int(gets.Load()) != len(tc.gets) {
				t.Fatalf("diagnostics %v delete %d gets %d", resp.Diagnostics, deletes.Load(), gets.Load())
			}
			if tc.wantError && !resp.State.Raw.Equal(s.Raw) {
				t.Fatal("failed deletion changed prior state")
			}
		})
	}
}

func TestMetricHistoricalIngestionCancellation(t *testing.T) {
	started := make(chan struct{})
	r := historicalTestResource(t, func(w http.ResponseWriter, req *http.Request) {
		if req.Method == "POST" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(201)
			fmt.Fprint(w, historicalTestResponse)
			return
		}
		close(started)
		<-req.Context().Done()
	})
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan resource.CreateResponse, 1)
	go func() { done <- historicalTestCreate(t, ctx, r) }()
	select {
	case <-started:
		cancel()
	case <-time.After(5 * time.Second):
		t.Fatal("GET did not start")
	}
	select {
	case resp := <-done:
		if !resp.Diagnostics.HasError() {
			t.Fatal("expected cancellation diagnostic")
		}
		var state metricHistoricalIngestionModel
		if d := resp.State.Get(context.Background(), &state); d.HasError() {
			t.Fatal(d)
		}
		if state.ID.ValueString() != historicalTestName {
			t.Fatalf("lost acknowledged state: %v", state.ID)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("request did not cancel")
	}
}

func TestMetricHistoricalIngestionBoundedRetry(t *testing.T) {

	for _, phase := range []string{"write", "read"} {

		t.Run(phase, func(t *testing.T) {

			var posts, gets atomic.Int32

			r := historicalTestResource(t, func(w http.ResponseWriter, req *http.Request) {

				w.Header().Set("Content-Type", "application/json")

				if req.Method == "POST" {
					posts.Add(1)
					if phase == "read" {
						w.WriteHeader(201)
						fmt.Fprint(w, historicalTestResponse)
						return
					}
				} else {
					gets.Add(1)
				}

				w.Header().Set("X-RateLimit-Reset", "60")
				w.WriteHeader(429)
				fmt.Fprint(w, `{"errors":["rate limited"]}`)

			})

			r.api.Client.Cfg.RetryConfiguration.EnableRetry = true

			r.timeout = 200 * time.Millisecond

			start := time.Now()
			resp := historicalTestCreate(t, context.Background(), r)

			if !resp.Diagnostics.HasError() || posts.Load() != 1 || time.Since(start) > 2*time.Second {
				t.Fatalf("diagnostics %v posts %d duration %v", resp.Diagnostics, posts.Load(), time.Since(start))
			}

			if phase == "read" && gets.Load() != 1 {
				t.Fatalf("read requests %d", gets.Load())
			}

		})

	}

}

func TestMetricHistoricalIngestionTimeout(t *testing.T) {

	r := historicalTestResource(t, func(w http.ResponseWriter, req *http.Request) {

		if req.Method == "POST" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(201)
			fmt.Fprint(w, historicalTestResponse)
			return
		}
		<-req.Context().Done()

	})

	r.timeout = 100 * time.Millisecond
	start := time.Now()
	resp := historicalTestCreate(t, context.Background(), r)

	if !resp.Diagnostics.HasError() || time.Since(start) > 2*time.Second {
		t.Fatalf("diagnostics %v elapsed %v", resp.Diagnostics, time.Since(start))
	}

	var state metricHistoricalIngestionModel
	if d := resp.State.Get(context.Background(), &state); d.HasError() {
		t.Fatal(d)
	}
	if state.ID.ValueString() != historicalTestName {
		t.Fatal("lost acknowledged ID")
	}

}

func TestMetricHistoricalIngestionDeleteVerificationTimeout(t *testing.T) {

	r := historicalTestResource(t, func(w http.ResponseWriter, req *http.Request) {

		if req.Method == http.MethodDelete {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, historicalTestResponse)

	})

	r.timeout = 100 * time.Millisecond
	s := historicalTestState(t, r, historicalTestModel())
	resp := resource.DeleteResponse{State: s}

	r.Delete(context.Background(), resource.DeleteRequest{State: s}, &resp)

	if !resp.Diagnostics.HasError() || !resp.State.Raw.Equal(s.Raw) {
		t.Fatalf("diagnostics=%v prior state preserved=%v", resp.Diagnostics, resp.State.Raw.Equal(s.Raw))
	}

}

func TestMetricHistoricalIngestionUnexpectedDeleteStatus(t *testing.T) {

	var gets atomic.Int32

	r := historicalTestResource(t, func(w http.ResponseWriter, req *http.Request) {
		if req.Method == http.MethodDelete {
			w.WriteHeader(202)
			return
		}
		gets.Add(1)
	})

	s := historicalTestState(t, r, historicalTestModel())
	resp := resource.DeleteResponse{State: s}
	r.Delete(context.Background(), resource.DeleteRequest{State: s}, &resp)

	if !resp.Diagnostics.HasError() || !resp.State.Raw.Equal(s.Raw) || gets.Load() != 0 {
		t.Fatalf("diagnostics=%v prior preserved=%v reads=%d", resp.Diagnostics, resp.State.Raw.Equal(s.Raw), gets.Load())
	}

}

func TestMetricHistoricalIngestionWriteAndReadCancellation(t *testing.T) {

	for _, operation := range []string{"create", "read", "delete"} {
		t.Run(operation, func(t *testing.T) {

			started := make(chan struct{})
			release := make(chan struct{})
			defer close(release)
			r := historicalTestResource(t, func(w http.ResponseWriter, req *http.Request) {
				_, _ = io.Copy(io.Discard, req.Body)
				close(started)
				select {
				case <-req.Context().Done():
				case <-release:
				}
			})

			state := historicalTestState(t, r, historicalTestModel())
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			done := make(chan bool, 1)

			switch operation {

			case "create":

				model := historicalTestModel()
				model.ID = types.StringUnknown()
				model.CreatedAt = types.StringUnknown()
				planState := historicalTestState(t, r, model)
				plan := tfsdk.Plan{Schema: planState.Schema, Raw: planState.Raw}

				go func() {
					resp := resource.CreateResponse{State: planState}
					r.Create(ctx, resource.CreateRequest{Plan: plan}, &resp)
					done <- resp.Diagnostics.HasError()
				}()

			case "read":

				go func() {
					resp := resource.ReadResponse{State: state}
					r.Read(ctx, resource.ReadRequest{State: state}, &resp)
					done <- resp.Diagnostics.HasError() && resp.State.Raw.Equal(state.Raw)
				}()

			case "delete":

				go func() {
					resp := resource.DeleteResponse{State: state}
					r.Delete(ctx, resource.DeleteRequest{State: state}, &resp)
					done <- resp.Diagnostics.HasError() && resp.State.Raw.Equal(state.Raw)
				}()

			}

			select {
			case <-started:
				cancel()
			case <-time.After(5 * time.Second):
				t.Fatal("request did not start")
			}

			select {
			case ok := <-done:
				if !ok {
					t.Fatal("cancellation did not preserve state or report an error")
				}
			case <-time.After(5 * time.Second):
				t.Fatal("request did not cancel")
			}

		})
	}

}
