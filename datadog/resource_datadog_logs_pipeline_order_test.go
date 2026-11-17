package datadog

import (
	"context"
	"encoding/json"
	"maps"
	"net/http"
	"net/http/httptest"
	"slices"
	"testing"

	dd "github.com/DataDog/datadog-api-client-go/v2/api/datadog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

func TestLogsPipelineOrderHandlesUnlistedPipelines(t *testing.T) {
	for _, tc := range []struct {
		name, operation string
		relative        bool
		desired, state  []string
		put             []string
		wantError       bool
	}{
		{"reorder", "update", true, []string{"b", "a"}, []string{"b", "a"}, []string{"b", "outside", "a"}, false},
		{"already ordered", "create", true, []string{"a", "b"}, []string{"a", "b"}, nil, false},
		{"refresh after reorder and deletion", "read", true, []string{"b", "gone", "a"}, []string{"a", "b"}, nil, false},
		{"missing", "update", true, []string{"b", "missing"}, nil, nil, true},
		{"duplicate", "update", true, []string{"b", "b"}, nil, nil, true},
		{"cancellation after a rate-limited write", "cancel", true, []string{"b", "a"}, nil, []string{"b", "outside", "a"}, true},
		{"disabling relative order rejects an incomplete list", "disable", false, []string{"b", "a"}, nil, []string{"b", "a"}, true},
		{"default refresh", "read", false, []string{"a", "b"}, []string{"a", "outside", "b"}, nil, false},
		{"default update", "update", false, []string{"b", "a", "outside"}, []string{"b", "a", "outside"}, []string{"b", "a", "outside"}, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(t.Context())
			defer cancel()
			remote := []string{"a", "outside", "b"}
			var put []string
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "/api/v1/logs/config/pipeline-order", r.URL.Path)
				w.Header().Set("Content-Type", "application/json")
				switch r.Method {
				case http.MethodGet:
				case http.MethodPut:
					var order struct {
						IDs []string `json:"pipeline_ids"`
					}
					if err := json.NewDecoder(r.Body).Decode(&order); !assert.NoError(t, err) {
						w.WriteHeader(http.StatusBadRequest)
						return
					}
					put = order.IDs
					if tc.operation == "cancel" && ctx.Err() == nil {
						cancel()
						w.WriteHeader(http.StatusTooManyRequests)
						return
					}
					for _, id := range remote {
						if !slices.Contains(put, id) {
							w.WriteHeader(http.StatusUnprocessableEntity)
							_, _ = w.Write([]byte(`{"errors":["Missing pipeline ID"]}`))
							return
						}
					}
					remote = put
				default:
					t.Errorf("unexpected method %s", r.Method)
					w.WriteHeader(http.StatusMethodNotAllowed)
					return
				}
				_ = json.NewEncoder(w).Encode(map[string][]string{"pipeline_ids": remote})
			}))
			defer server.Close()
			config := dd.NewConfiguration()
			config.Servers = dd.ServerConfigurations{{URL: server.URL}}
			config.OperationServers = nil
			config.HTTPClient = server.Client()
			config.RetryConfiguration.EnableRetry = tc.operation == "cancel"
			meta := &ProviderConfiguration{Auth: t.Context(), DatadogApiInstances: &utils.ApiInstances{HttpClient: dd.NewAPIClient(config)}}
			resource := resourceDatadogLogsPipelineOrder()
			pipelines := make([]interface{}, len(tc.desired))
			for i, id := range tc.desired {
				pipelines[i] = id
			}
			values := map[string]interface{}{"name": "order", "pipelines": pipelines}
			if tc.relative {
				values["ignore_unlisted_pipelines"] = true
			}
			var prior, state *terraform.InstanceState
			if tc.operation != "create" {
				priorValues := maps.Clone(values)
				if tc.operation != "read" {
					priorValues["pipelines"] = []interface{}{"a", "b"}
				}
				if tc.operation == "disable" {
					priorValues["ignore_unlisted_pipelines"] = true
				}
				data := schema.TestResourceDataRaw(t, resource.SchemaMap(), priorValues)
				data.SetId("order")
				prior = data.State()
			}
			var diags diag.Diagnostics
			if tc.operation == "read" {
				state, diags = resource.RefreshWithoutUpgrade(ctx, prior, meta)
			} else {
				diff, err := resource.Diff(ctx, prior, terraform.NewResourceConfigRaw(values), meta)
				require.NoError(t, err)
				state, diags = resource.Apply(ctx, prior, diff, meta)
			}
			server.Close()
			require.Equal(t, tc.wantError, diags.HasError(), "%v", diags)
			require.Equal(t, tc.put, put)
			if tc.wantError {
				require.Equal(t, prior.Attributes, state.Attributes)
				return
			}
			require.Equal(t, "order", state.ID)
			require.Equal(t, tc.state, utils.GetStringSlice(resource.Data(state), "pipelines"))
		})
	}
}
