package fwprovider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadog"
	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const metricTagV2TestName = "test.tags_only"

func metricTagV2TestModel(t *testing.T, tags []string, exclude bool) metricTagConfigurationV2Model {
	t.Helper()
	set, diags := types.SetValueFrom(context.Background(), types.StringType, tags)
	metricTagV2NoErrors(t, diags)
	return metricTagConfigurationV2Model{
		ID: types.StringValue(metricTagV2TestName), MetricName: types.StringValue(metricTagV2TestName),
		MetricType: types.StringValue("distribution"), Tags: set, ExcludeTagsMode: types.BoolValue(exclude),
	}
}

func metricTagV2NoErrors(t *testing.T, diags diag.Diagnostics) {
	t.Helper()
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags.Errors())
	}
}

func metricTagV2TestResource(t *testing.T, handler http.HandlerFunc) *metricTagConfigurationV2Resource {
	t.Helper()
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)
	config := datadog.NewConfiguration()
	config.Servers = datadog.ServerConfigurations{{URL: server.URL}}
	config.OperationServers = nil
	config.HTTPClient = server.Client()
	auth := context.WithValue(context.Background(), datadog.ContextAPIKeys, map[string]datadog.APIKey{
		"apiKeyAuth": {Key: "test-api-key"}, "appKeyAuth": {Key: "test-app-key"},
	})
	return &metricTagConfigurationV2Resource{API: datadogV2.NewMetricsApi(datadog.NewAPIClient(config)), Auth: auth, visibilityTimeout: time.Second * 5}
}

func metricTagV2TestPlan(t *testing.T, r *metricTagConfigurationV2Resource, model metricTagConfigurationV2Model) tfsdk.Plan {
	t.Helper()
	var schemaResp resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &schemaResp)
	plan := tfsdk.Plan{Schema: schemaResp.Schema}
	metricTagV2NoErrors(t, plan.Set(context.Background(), &model))
	return plan
}

func metricTagV2TestState(plan tfsdk.Plan) tfsdk.State {
	return tfsdk.State{Schema: plan.Schema, Raw: plan.Raw}
}

func metricTagV2WriteResponse(w http.ResponseWriter, name string, tags []string, exclude, percentiles bool) {
	w.Header().Set("Content-Type", "application/json")
	// A response may contain percentiles; v2 must never adopt their ownership.
	_ = json.NewEncoder(w).Encode(map[string]any{"data": map[string]any{
		"id": name, "type": "manage_tags", "attributes": map[string]any{
			"metric_type": "distribution", "tags": tags, "exclude_tags_mode": exclude, "include_percentiles": percentiles,
		},
	}})
}

func TestMetricTagConfigurationV2CRUDPreservesPercentiles(t *testing.T) {
	for _, enabled := range []bool{false, true} {
		t.Run(fmt.Sprintf("percentiles=%t", enabled), func(t *testing.T) {
			ctx := context.Background()
			configured := false
			percentiles := enabled
			var tags []string
			var exclude bool
			writes := 0
			var backendMu sync.Mutex
			r := metricTagV2TestResource(t, func(w http.ResponseWriter, req *http.Request) {
				backendMu.Lock()
				defer backendMu.Unlock()
				if req.URL.Path != "/api/v2/metrics/"+metricTagV2TestName+"/tags" {
					t.Errorf("unexpected path: %s", req.URL.Path)
				}
				if req.Header.Get("DD-API-KEY") != "test-api-key" || req.Header.Get("DD-APPLICATION-KEY") != "test-app-key" {
					t.Error("provider authentication was lost")
				}
				switch req.Method {
				case http.MethodPost, http.MethodPatch:
					var body struct {
						Data struct {
							ID         string                     `json:"id"`
							Attributes map[string]json.RawMessage `json:"attributes"`
						} `json:"data"`
					}
					if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
						t.Errorf("decode write: %v", err)
					}
					if body.Data.ID != metricTagV2TestName {
						t.Errorf("incorrect payload ID: %q", body.Data.ID)
					}
					if _, present := body.Data.Attributes["include_percentiles"]; present {
						t.Error("tag write took ownership of percentiles")
						percentiles = !enabled
					}
					if _, present := body.Data.Attributes["aggregations"]; present {
						t.Error("obsolete aggregations sent")
					}
					if err := json.Unmarshal(body.Data.Attributes["tags"], &tags); err != nil {
						t.Errorf("tags: %v", err)
					}
					if string(body.Data.Attributes["tags"]) == "null" {
						t.Error("explicit empty tags became JSON null")
					}
					if err := json.Unmarshal(body.Data.Attributes["exclude_tags_mode"], &exclude); err != nil {
						t.Errorf("exclude mode: %v", err)
					}
					configured = true
					writes++
					metricTagV2WriteResponse(w, metricTagV2TestName, tags, exclude, percentiles)
				case http.MethodGet:
					if !configured {
						w.WriteHeader(http.StatusNotFound)
						fmt.Fprint(w, `{"errors":["not configured"]}`)
						return
					}
					metricTagV2WriteResponse(w, metricTagV2TestName, tags, exclude, percentiles)
				case http.MethodDelete:
					if req.ContentLength > 0 {
						t.Error("DELETE unexpectedly has a body")
					}
					configured = false
					w.WriteHeader(http.StatusNoContent)
				default:
					t.Errorf("unexpected method: %s", req.Method)
				}
			})
			model := metricTagV2TestModel(t, []string{"service", "env"}, false)
			model.ID = types.StringUnknown()
			plan := metricTagV2TestPlan(t, r, model)
			create := resource.CreateResponse{State: metricTagV2TestState(plan)}
			r.Create(ctx, resource.CreateRequest{Plan: plan}, &create)
			metricTagV2NoErrors(t, create.Diagnostics)
			var state metricTagConfigurationV2Model
			metricTagV2NoErrors(t, create.State.Get(ctx, &state))
			if state.ID.ValueString() != metricTagV2TestName {
				t.Fatal("create did not retain metric ID")
			}

			// Empty tags are meaningful, not an absent/null field.
			updated := metricTagV2TestModel(t, []string{}, false)
			updatePlan := metricTagV2TestPlan(t, r, updated)
			update := resource.UpdateResponse{State: create.State}
			r.Update(ctx, resource.UpdateRequest{Plan: updatePlan, State: create.State}, &update)
			metricTagV2NoErrors(t, update.Diagnostics)
			metricTagV2NoErrors(t, update.State.Get(ctx, &state))
			if len(state.Tags.Elements()) != 0 {
				t.Error("empty tag set was not retained")
			}

			read := resource.ReadResponse{State: update.State}
			r.Read(ctx, resource.ReadRequest{State: update.State}, &read)
			metricTagV2NoErrors(t, read.Diagnostics)
			if !read.State.Raw.Equal(update.State.Raw) {
				t.Error("refresh changed owned state")
			}
			remove := resource.DeleteResponse{State: read.State}
			r.Delete(ctx, resource.DeleteRequest{State: read.State}, &remove)
			metricTagV2NoErrors(t, remove.Diagnostics)
			backendMu.Lock()
			defer backendMu.Unlock()
			if percentiles != enabled {
				t.Error("percentiles changed during tag lifecycle")
			}
			if writes != 2 {
				t.Errorf("write count = %d, want create+update only", writes)
			}
		})
	}
}

func TestMetricTagConfigurationV2ImportAndDrift(t *testing.T) {
	ctx := context.Background()
	var missing atomic.Bool
	var gets atomic.Int32
	r := metricTagV2TestResource(t, func(w http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodGet || req.URL.Path != "/api/v2/metrics/"+metricTagV2TestName+"/tags" {
			t.Errorf("import/refresh mutated or used empty name: %s %s", req.Method, req.URL.Path)
		}
		gets.Add(1)
		if missing.Load() {
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprint(w, `{"errors":["absent"]}`)
			return
		}
		metricTagV2WriteResponse(w, metricTagV2TestName, []string{"service"}, false, true)
	})
	model := metricTagConfigurationV2Model{ID: types.StringNull(), MetricName: types.StringNull(), MetricType: types.StringNull(), Tags: types.SetNull(types.StringType), ExcludeTagsMode: types.BoolNull()}
	plan := metricTagV2TestPlan(t, r, model)
	imported := resource.ImportStateResponse{State: metricTagV2TestState(plan)}
	r.ImportState(ctx, resource.ImportStateRequest{ID: metricTagV2TestName}, &imported)
	metricTagV2NoErrors(t, imported.Diagnostics)
	read := resource.ReadResponse{State: imported.State}
	r.Read(ctx, resource.ReadRequest{State: imported.State}, &read)
	metricTagV2NoErrors(t, read.Diagnostics)
	metricTagV2NoErrors(t, read.State.Get(ctx, &model))
	if model.MetricName.ValueString() != metricTagV2TestName || model.MetricType.ValueString() != "distribution" {
		t.Errorf("import not hydrated: %+v", model)
	}
	missing.Store(true)
	drift := resource.ReadResponse{State: read.State}
	r.Read(ctx, resource.ReadRequest{State: read.State}, &drift)
	metricTagV2NoErrors(t, drift.Diagnostics)
	if !drift.State.Raw.IsNull() {
		t.Error("missing tag configuration not removed from state")
	}
	if gets.Load() != 2 {
		t.Error("ordinary refresh retried to hide drift")
	}
}

func TestMetricTagConfigurationV2WriteErrors(t *testing.T) {
	for _, status := range []int{400, 403, 409, 429, 500} {
		for _, body := range []string{`{"errors":["write rejected"]}`, `{"errors":[{"status":"400","detail":"write rejected"}]}`} {
			t.Run(fmt.Sprintf("status=%d/%s", status, body), func(t *testing.T) {
				var calls atomic.Int32
				r := metricTagV2TestResource(t, func(w http.ResponseWriter, req *http.Request) {
					calls.Add(1)
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(status)
					fmt.Fprint(w, body)
				})
				model := metricTagV2TestModel(t, []string{"env"}, false)
				model.ID = types.StringUnknown()
				plan := metricTagV2TestPlan(t, r, model)
				response := resource.CreateResponse{State: metricTagV2TestState(plan)}
				r.Create(context.Background(), resource.CreateRequest{Plan: plan}, &response)
				if !response.Diagnostics.HasError() {
					t.Fatal("write error was ignored")
				}
				if calls.Load() != 1 {
					t.Error("mutation was retried")
				}
				if !response.State.Raw.Equal(plan.Raw) {
					t.Error("failed write changed state")
				}
				if !strings.Contains(fmt.Sprint(response.Diagnostics.Errors()), "write rejected") {
					t.Error("API detail lost")
				}
			})
		}
	}
}

func TestMetricTagConfigurationV2AcknowledgedWriteRetainsState(t *testing.T) {
	for _, scenario := range []string{"forbidden", "timeout", "unparsed", "wrong-id", "malformed"} {
		t.Run(scenario, func(t *testing.T) {
			var writes atomic.Int32
			r := metricTagV2TestResource(t, func(w http.ResponseWriter, req *http.Request) {
				if req.Method == http.MethodPost {
					writes.Add(1)
					if scenario == "malformed" {
						fmt.Fprint(w, `{`)
						return
					}
					if scenario == "unparsed" {
						fmt.Fprint(w, `{"data":{"id":"test.tags_only","type":"future_type","attributes":{}}}`)
						return
					}
					name := metricTagV2TestName
					if scenario == "wrong-id" {
						name = "other.metric"
					}
					metricTagV2WriteResponse(w, name, []string{"env"}, false, true)
					return
				}
				if scenario == "forbidden" {
					w.WriteHeader(http.StatusForbidden)
				} else {
					w.WriteHeader(http.StatusNotFound)
				}
				fmt.Fprint(w, `{"errors":["not visible"]}`)
			})
			r.visibilityTimeout = 25 * time.Millisecond
			model := metricTagV2TestModel(t, []string{"env"}, false)
			model.ID = types.StringUnknown()
			plan := metricTagV2TestPlan(t, r, model)
			response := resource.CreateResponse{State: metricTagV2TestState(plan)}
			r.Create(context.Background(), resource.CreateRequest{Plan: plan}, &response)
			if !response.Diagnostics.HasError() {
				t.Fatal("expected verification/response error")
			}
			var state metricTagConfigurationV2Model
			metricTagV2NoErrors(t, response.State.Get(context.Background(), &state))
			if state.ID.ValueString() != metricTagV2TestName {
				t.Error("acknowledged resource was orphaned")
			}
			if writes.Load() != 1 {
				t.Error("write was retried during readback")
			}
		})
	}
}

func TestMetricTagConfigurationV2DelayedVisibility(t *testing.T) {
	var reads atomic.Int32
	r := metricTagV2TestResource(t, func(w http.ResponseWriter, req *http.Request) {
		switch reads.Add(1) {
		case 1:
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprint(w, `{"errors":["not visible"]}`)
		case 2:
			metricTagV2WriteResponse(w, metricTagV2TestName, []string{"stale"}, false, true)
		default:
			// Tag order must not affect convergence; percentiles are not compared.
			metricTagV2WriteResponse(w, metricTagV2TestName, []string{"service", "env"}, true, false)
		}
	})
	model := metricTagV2TestModel(t, []string{"env", "service"}, true)
	if err := r.waitForVisibility(context.Background(), &model, false); err != nil {
		t.Fatal(err)
	}
	if reads.Load() < 3 {
		t.Error("stale state was accepted")
	}
}

func TestMetricTagConfigurationV2Cancellation(t *testing.T) {
	started := make(chan struct{}, 1)
	finished := make(chan struct{}, 1)
	r := metricTagV2TestResource(t, func(w http.ResponseWriter, req *http.Request) {
		started <- struct{}{}
		<-req.Context().Done()
		finished <- struct{}{}
	})
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	model := metricTagV2TestModel(t, []string{"env"}, false)
	done := make(chan error, 1)
	go func() { done <- r.waitForVisibility(ctx, &model, false) }()
	select {
	case <-started:
	case <-time.After(time.Second):
		t.Fatal("request did not start")
	}
	cancel()
	select {
	case err := <-done:
		if err == nil {
			t.Error("cancellation returned success")
		}
	case <-time.After(time.Second):
		t.Fatal("waiter did not cancel")
	}
	select {
	case <-finished:
	case <-time.After(time.Second):
		t.Fatal("actual HTTP request did not cancel")
	}
}

func TestMetricTagConfigurationV2DeleteFailures(t *testing.T) {
	for _, scenario := range []string{"missing", "forbidden", "stale"} {
		t.Run(scenario, func(t *testing.T) {
			r := metricTagV2TestResource(t, func(w http.ResponseWriter, req *http.Request) {
				if req.Method == http.MethodDelete {
					if scenario == "missing" {
						w.WriteHeader(http.StatusNotFound)
						fmt.Fprint(w, `{"errors":["absent"]}`)
						return
					}
					if scenario == "forbidden" {
						w.WriteHeader(http.StatusForbidden)
						fmt.Fprint(w, `{"errors":["denied"]}`)
						return
					}
					w.WriteHeader(http.StatusNoContent)
					return
				}
				metricTagV2WriteResponse(w, metricTagV2TestName, []string{"env"}, false, true)
			})
			r.visibilityTimeout = 25 * time.Millisecond
			plan := metricTagV2TestPlan(t, r, metricTagV2TestModel(t, []string{"env"}, false))
			state := metricTagV2TestState(plan)
			response := resource.DeleteResponse{State: state}
			r.Delete(context.Background(), resource.DeleteRequest{State: state}, &response)
			if response.Diagnostics.HasError() != (scenario != "missing") {
				t.Fatalf("unexpected diagnostics: %v", response.Diagnostics)
			}
			if scenario != "missing" && !response.State.Raw.Equal(state.Raw) {
				t.Error("failed destroy lost ownership")
			}
		})
	}
}

func TestMetricTagConfigurationV2SchemaAndValidation(t *testing.T) {
	r := NewMetricTagConfigurationV2Resource().(*metricTagConfigurationV2Resource)
	var schemaResp resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &schemaResp)
	for _, name := range []string{"include_percentiles", "aggregations", "manage_percentiles"} {
		if _, exists := schemaResp.Schema.Attributes[name]; exists {
			t.Errorf("unexpected ownership attribute %s", name)
		}
	}
	for name, attribute := range schemaResp.Schema.Attributes {
		if attribute.GetDescription() == "" {
			t.Errorf("missing description for %s", name)
		}
	}
	for _, exclude := range []bool{true, false} {
		plan := metricTagV2TestPlan(t, r, metricTagV2TestModel(t, []string{}, exclude))
		resp := resource.ValidateConfigResponse{}
		r.ValidateConfig(context.Background(), resource.ValidateConfigRequest{Config: tfsdk.Config{Schema: plan.Schema, Raw: plan.Raw}}, &resp)
		if resp.Diagnostics.HasError() != exclude {
			t.Errorf("exclude=%t, diagnostics=%v", exclude, resp.Diagnostics)
		}
	}
	var configure resource.ConfigureResponse
	r.Configure(context.Background(), resource.ConfigureRequest{}, &configure)
	metricTagV2NoErrors(t, configure.Diagnostics)
	r.Configure(context.Background(), resource.ConfigureRequest{ProviderData: "incorrect"}, &configure)
	if !configure.Diagnostics.HasError() {
		t.Error("bad provider type was accepted")
	}
	var metadata resource.MetadataResponse
	r.Metadata(context.Background(), resource.MetadataRequest{}, &metadata)
	if metadata.TypeName != "metric_tag_configuration_v2" {
		t.Fatal(metadata.TypeName)
	}
}

func TestMetricTagConfigurationV2ContextValues(t *testing.T) {
	type key string
	auth := context.WithValue(context.Background(), key("auth"), "credentials")
	operation := context.WithValue(context.Background(), key("trace"), "operation")
	ctx := metricTagConfigurationContext{Context: operation, auth: auth}
	if !reflect.DeepEqual([]any{ctx.Value(key("auth")), ctx.Value(key("trace"))}, []any{"credentials", "operation"}) {
		t.Error("context values lost")
	}
}
