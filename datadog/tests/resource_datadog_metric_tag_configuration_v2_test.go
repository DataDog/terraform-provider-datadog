package test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	common "github.com/DataDog/datadog-api-client-go/v2/api/datadog"
	frameworkDiag "github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/terraform-providers/terraform-provider-datadog/datadog"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/fwprovider"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

// These tests require recordings or explicit sandbox credentials. They never
// infer that a replay skip is proof of service rollout or live preservation.
func TestAccMetricTagConfigurationV2Live(t *testing.T) {
	for _, metricType := range []string{"gauge", "distribution"} {
		t.Run(metricType, func(t *testing.T) {
			skipIfNoCassette(t)
			ctx, providers, factories := testAccFrameworkMuxProviders(context.Background(), t)
			name := strings.ReplaceAll(uniqueEntityName(ctx, t), "-", "_")
			address := "datadog_metric_tag_configuration_v2.test"
			config := func(tags string) string {
				return fmt.Sprintf(`resource "datadog_metric_tag_configuration_v2" "test" {
  metric_name = %q
  metric_type = %q
  tags = %s
}`, name, metricType, tags)
			}
			resource.Test(t, resource.TestCase{
				PreCheck:                 func() { testAccPreCheck(t) },
				ProtoV6ProviderFactories: factories,
				CheckDestroy: func(state *terraform.State) error {
					_, response, err := providers.frameworkProvider.DatadogApiInstances.GetMetricsApiV2().ListTagConfigurationByName(providers.frameworkProvider.Auth, name)
					if response != nil && response.StatusCode == http.StatusNotFound {
						return nil
					}
					if err != nil {
						return utils.TranslateClientError(err, response, "verifying tag configuration deletion")
					}
					return fmt.Errorf("tag configuration %q still exists", name)
				},
				Steps: []resource.TestStep{
					{PreConfig: submitTestMetricPreConfig(t, name, metricType), Config: config(`["sport", "datacenter"]`), Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr(address, "id", name),
						resource.TestCheckResourceAttr(address, "metric_name", name),
						resource.TestCheckResourceAttr(address, "metric_type", metricType),
						resource.TestCheckResourceAttr(address, "tags.#", "2"),
					)},
					{ResourceName: address, ImportState: true, ImportStateVerify: true},
					{Config: config(`["sport"]`), Check: resource.TestCheckResourceAttr(address, "tags.#", "1")},
				},
			})
		})
	}
}

// Exercise real Terraform plans, the legacy SDKv2 resource, the new Framework
// resource, and the actual generated client against an isolated HTTP backend.
// This verifies the state handoff without sandbox keys or invented cassettes;
// it is deliberately not evidence of the production backend's behavior.
func TestAccMetricTagConfigurationV2MockMigration(t *testing.T) {
	for _, legacySetting := range []string{"true", "false", "omitted"} {
		t.Run(legacySetting, func(t *testing.T) {
			ctx := context.Background()
			const name = "test.tag_configuration_migration"
			var mu sync.Mutex
			configured, percentiles, newOwner := false, false, false
			tags := []string{}
			posts, patches, deletes := 0, 0, 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				mu.Lock()
				defer mu.Unlock()
				w.Header().Set("Content-Type", "application/json")
				if req.URL.Path != "/api/v2/metrics/"+name+"/tags" {
					t.Errorf("unexpected endpoint: %s %s", req.Method, req.URL.Path)
					w.WriteHeader(http.StatusNotFound)
					return
				}
				switch req.Method {
				case http.MethodPost, http.MethodPatch:
					var body struct {
						Data struct {
							Attributes struct {
								Tags        []string `json:"tags"`
								Percentiles *bool    `json:"include_percentiles"`
							} `json:"attributes"`
						} `json:"data"`
					}
					if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
						t.Errorf("invalid request: %v", err)
					}
					if newOwner && body.Data.Attributes.Percentiles != nil {
						t.Error("tags-only owner sent percentiles")
					}
					if body.Data.Attributes.Percentiles != nil {
						percentiles = *body.Data.Attributes.Percentiles
					}
					tags, configured = body.Data.Attributes.Tags, true
					if req.Method == http.MethodPost {
						posts++
					} else {
						patches++
					}
				case http.MethodDelete:
					deletes++
					configured = false
					w.WriteHeader(http.StatusNoContent)
					return
				case http.MethodGet:
					if !configured {
						w.WriteHeader(http.StatusNotFound)
						fmt.Fprint(w, `{"errors":["not configured"]}`)
						return
					}
				default:
					t.Errorf("unexpected method: %s", req.Method)
				}
				_ = json.NewEncoder(w).Encode(map[string]any{"data": map[string]any{"id": name, "type": "manage_tags", "attributes": map[string]any{
					"metric_type": "distribution", "tags": tags, "exclude_tags_mode": false, "include_percentiles": percentiles,
				}}})
			}))
			defer server.Close()
			cfg := common.NewConfiguration()
			cfg.Servers = common.ServerConfigurations{{URL: server.URL}}
			cfg.OperationServers = nil
			cfg.HTTPClient = server.Client()
			instances := &utils.ApiInstances{HttpClient: common.NewAPIClient(cfg)}
			legacy := datadog.Provider()
			legacy.ConfigureContextFunc = func(context.Context, *schema.ResourceData) (interface{}, diag.Diagnostics) {
				return &datadog.ProviderConfiguration{Auth: ctx, DatadogApiInstances: instances}, nil
			}
			framework := &fwprovider.FrameworkProvider{
				Auth: ctx, DatadogApiInstances: instances,
				ConfigureCallbackFunc: func(*fwprovider.FrameworkProvider, *provider.ConfigureRequest, *fwprovider.ProviderSchema) frameworkDiag.Diagnostics {
					return nil
				},
			}
			factories := testAccFrameworkMuxProvidersServer(ctx, legacy, framework)
			percentileLine := ""
			if legacySetting != "omitted" {
				percentileLine = "include_percentiles = " + legacySetting
			}
			oldConfig := fmt.Sprintf(`resource "datadog_metric_tag_configuration" "legacy" {
  metric_name = %q
  metric_type = "distribution"
  tags = ["env"]
  %s
}`, name, percentileLine)
			newConfig := func(tagList string) string {
				return fmt.Sprintf(`removed {
  from = datadog_metric_tag_configuration.legacy
  lifecycle { destroy = false }
}
resource "datadog_metric_tag_configuration_v2" "test" {
  metric_name = %q
  metric_type = "distribution"
  tags = %s
}
import {
  to = datadog_metric_tag_configuration_v2.test
  id = %q
}`, name, tagList, name)
			}
			expectedPercentiles := legacySetting == "true"
			resource.Test(t, resource.TestCase{
				ProtoV6ProviderFactories: factories,
				CheckDestroy: func(*terraform.State) error {
					mu.Lock()
					defer mu.Unlock()
					if configured || deletes != 1 || posts != 1 || patches != 1 {
						return fmt.Errorf("unexpected lifecycle: configured=%t POST=%d PATCH=%d DELETE=%d", configured, posts, patches, deletes)
					}
					if percentiles != !expectedPercentiles {
						return fmt.Errorf("destroy changed externally managed percentiles")
					}
					return nil
				},
				Steps: []resource.TestStep{
					{Config: oldConfig, Check: func(*terraform.State) error {
						mu.Lock()
						defer mu.Unlock()
						if percentiles != expectedPercentiles {
							return fmt.Errorf("legacy behavior changed")
						}
						newOwner = true
						return nil
					}},
					{Config: newConfig(`["env"]`), Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("datadog_metric_tag_configuration_v2.test", "id", name),
						func(*terraform.State) error {
							mu.Lock()
							defer mu.Unlock()
							if posts != 1 || patches != 0 || deletes != 0 || percentiles != expectedPercentiles {
								return fmt.Errorf("migration mutated remote state: POST=%d PATCH=%d DELETE=%d", posts, patches, deletes)
							}
							return nil
						},
					)},
					{Config: newConfig(`["env", "service"]`), Check: resource.TestCheckResourceAttr("datadog_metric_tag_configuration_v2.test", "tags.#", "2")},
					{PreConfig: func() { mu.Lock(); percentiles = !expectedPercentiles; mu.Unlock() }, Config: newConfig(`["env", "service"]`), PlanOnly: true},
				},
			})
		})
	}
}
