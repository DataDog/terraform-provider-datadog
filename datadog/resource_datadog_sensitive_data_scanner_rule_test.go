package datadog

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	datadogClient "github.com/DataDog/datadog-api-client-go/v2/api/datadog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

func TestSensitiveDataScannerRuleReadsShareConfigSnapshot(t *testing.T) {
	var requestCount atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		requestCount.Add(1)
		if request.Method != http.MethodGet {
			t.Errorf("request method = %s, want GET", request.Method)
		}
		if request.URL.Path != "/api/v2/sensitive-data-scanner/config" {
			t.Errorf("request path = %s, want /api/v2/sensitive-data-scanner/config", request.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		_, err := fmt.Fprint(w, `{
			"data": {
				"id": "config",
				"type": "sensitive_data_scanner_configuration"
			},
			"included": [
				{
					"id": "rule-1",
					"type": "sensitive_data_scanner_rule",
					"attributes": {
						"description": "first description",
						"excluded_namespaces": [],
						"included_keyword_configuration": {"keywords": [], "character_count": 30},
						"is_enabled": true,
						"name": "first rule",
						"namespaces": [],
						"pattern": "first-pattern",
						"priority": 1,
						"suppressions": {},
						"tags": ["first"],
						"text_replacement": {"type": "none"}
					},
					"relationships": {
						"group": {
							"data": {"id": "group-1", "type": "sensitive_data_scanner_group"}
						}
					}
				},
				{
					"id": "rule-2",
					"type": "sensitive_data_scanner_rule",
					"attributes": {
						"description": "second description",
						"excluded_namespaces": [],
						"included_keyword_configuration": {"keywords": [], "character_count": 30},
						"is_enabled": false,
						"name": "second rule",
						"namespaces": [],
						"pattern": "second-pattern",
						"priority": 2,
						"suppressions": {},
						"tags": ["second"],
						"text_replacement": {"type": "none"}
					},
					"relationships": {
						"group": {
							"data": {"id": "group-2", "type": "sensitive_data_scanner_group"}
						}
					}
				}
			]
		}`)
		if err != nil {
			t.Errorf("writing response: %v", err)
		}
	}))
	defer server.Close()

	config := datadogClient.NewConfiguration()
	config.Servers = datadogClient.ServerConfigurations{{URL: server.URL}}
	config.OperationServers = nil
	config.HTTPClient = server.Client()
	apiInstances := &utils.ApiInstances{HttpClient: datadogClient.NewAPIClient(config)}
	providerConfig := &ProviderConfiguration{
		DatadogApiInstances: apiInstances,
		Auth:                context.Background(),
	}
	resourceSchema := resourceDatadogSensitiveDataScannerRule().SchemaFunc()

	tests := []struct {
		id      string
		name    string
		groupID string
	}{
		{id: "rule-1", name: "first rule", groupID: "group-1"},
		{id: "rule-2", name: "second rule", groupID: "group-2"},
	}
	for _, test := range tests {
		resourceData := schema.TestResourceDataRaw(t, resourceSchema, map[string]interface{}{
			"group_id": test.groupID,
		})
		resourceData.SetId(test.id)

		diags := resourceDatadogSensitiveDataScannerRuleRead(context.Background(), resourceData, providerConfig)
		if diags.HasError() {
			t.Fatalf("reading %s returned diagnostics: %v", test.id, diags)
		}
		if got := resourceData.Get("name"); got != test.name {
			t.Fatalf("%s name = %q, want %q", test.id, got, test.name)
		}
		if got := resourceData.Get("group_id"); got != test.groupID {
			t.Fatalf("%s group_id = %q, want %q", test.id, got, test.groupID)
		}
	}

	if got := requestCount.Load(); got != 1 {
		t.Fatalf("request count = %d, want 1", got)
	}
}
