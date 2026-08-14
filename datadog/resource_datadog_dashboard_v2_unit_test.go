package datadog

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

func TestBuildDashboardV2SchemaValidate(t *testing.T) {
	validateSchema, ok := buildDashboardV2Schema()["validate"]
	if !ok {
		t.Fatal("validate schema is missing")
	}
	if !validateSchema.Optional || validateSchema.Type.String() != "TypeBool" {
		t.Fatalf("validate schema should be an optional bool, got %#v", validateSchema)
	}
	if validateSchema.DiffSuppressFunc == nil || !validateSchema.DiffSuppressFunc("validate", "false", "true", nil) {
		t.Fatal("validate must not create a persistent diff")
	}
}

func TestFlattenDashboardWidgetDefinitions(t *testing.T) {
	widgets := []interface{}{
		map[string]interface{}{
			"group_definition": []interface{}{
				map[string]interface{}{
					"layout_type": "ordered",
					"title":       "Service overview",
					"widget": []interface{}{
						map[string]interface{}{
							"timeseries_definition": []interface{}{
								map[string]interface{}{
									"request": []interface{}{
										map[string]interface{}{
											"display_type": "line",
											"q":            "avg:system.load.1{*}",
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	definitions := flattenDashboardWidgetDefinitions(widgets)
	if len(definitions) != 2 {
		t.Fatalf("expected group and child definitions, got %d", len(definitions))
	}
	if got := definitions[0].terraformPath; got != "widget.0.group_definition.0" {
		t.Fatalf("unexpected group path %q", got)
	}
	if got := definitions[0].definition["type"]; got != "group" {
		t.Fatalf("unexpected group type %#v", got)
	}
	if children, ok := definitions[0].definition["widgets"].([]interface{}); !ok || len(children) != 0 {
		t.Fatalf("group validation payload should not contain children, got %#v", definitions[0].definition["widgets"])
	}
	if got := definitions[1].terraformPath; got != "widget.0.group_definition.0.widget.0.timeseries_definition.0" {
		t.Fatalf("unexpected child path %q", got)
	}
	if got := definitions[1].definition["type"]; got != "timeseries" {
		t.Fatalf("unexpected child type %#v", got)
	}
}

func TestDashboardV2PlanValidationCanBeDisabled(t *testing.T) {
	config := map[string]interface{}{
		"title":       "Plan-time validation test",
		"layout_type": "ordered",
		"validate":    false,
		"widget":      dashboardV2TestTimeseriesWidgets("avg:system.load.1{*}"),
	}

	// A nil provider configuration would panic if the validation HTTP call
	// were reached, so a successful diff proves the call was skipped.
	if _, err := resourceDatadogDashboardV2().Diff(
		context.Background(),
		nil,
		terraform.NewResourceConfigRaw(config),
		nil,
	); err != nil {
		t.Fatalf("unexpected diff error: %v", err)
	}
}

func TestDashboardV2PlanValidationCallsEndpointByDefault(t *testing.T) {
	called := false
	client := dashboardValidationTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		called = true
		assertDashboardValidationRequest(t, r, 1)
		writeDashboardValidationResponse(t, w, map[string]interface{}{
			"results": []interface{}{map[string]interface{}{"is_valid": true, "widget_type": "timeseries"}},
		})
	})
	providerConfig := &ProviderConfiguration{
		Auth: context.Background(),
		DatadogApiInstances: &utils.ApiInstances{
			HttpClient: client,
		},
	}
	config := terraform.NewResourceConfigRaw(map[string]interface{}{
		"title":       "Plan-time validation test",
		"layout_type": "ordered",
		"widget":      dashboardV2TestTimeseriesWidgets("avg:system.load.1{*}"),
	})

	if _, err := resourceDatadogDashboardV2().Diff(context.Background(), nil, config, providerConfig); err != nil {
		t.Fatalf("unexpected diff error: %v", err)
	}
	if !called {
		t.Fatal("expected plan-time validation endpoint to be called")
	}
}

func TestDashboardV2PlanValidationRejectsMalformedQuery(t *testing.T) {
	const malformedQuery = "sum:metric.name{env:test} by {service.ad_count()"

	client := dashboardValidationTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		definitions := assertDashboardValidationRequest(t, r, 1)
		requests, ok := definitions[0]["requests"].([]interface{})
		if !ok || len(requests) != 1 {
			t.Fatalf("expected one widget request, got %#v", definitions[0]["requests"])
		}
		request, ok := requests[0].(map[string]interface{})
		if !ok || request["q"] != malformedQuery {
			t.Fatalf("expected malformed query %q, got %#v", malformedQuery, requests[0])
		}
		writeDashboardValidationResponse(t, w, map[string]interface{}{
			"results": []interface{}{map[string]interface{}{
				"is_valid":      false,
				"widget_type":   "timeseries",
				"error_path":    "requests.0.q",
				"error_message": "invalid metric query: missing closing brace",
			}},
		})
	})
	providerConfig := &ProviderConfiguration{
		Auth: context.Background(),
		DatadogApiInstances: &utils.ApiInstances{
			HttpClient: client,
		},
	}
	config := terraform.NewResourceConfigRaw(map[string]interface{}{
		"title":       "Plan-time validation test",
		"layout_type": "ordered",
		"widget":      dashboardV2TestTimeseriesWidgets(malformedQuery),
	})

	_, err := resourceDatadogDashboardV2().Diff(context.Background(), nil, config, providerConfig)
	if err == nil {
		t.Fatal("expected malformed dashboard query to fail planning")
	}
	for _, expected := range []string{"widget.0.timeseries_definition.0.request.0.q", "timeseries", "missing closing brace"} {
		if !strings.Contains(err.Error(), expected) {
			t.Fatalf("expected error %q to contain %q", err, expected)
		}
	}
}

func TestValidateDashboardWidgetDefinitions(t *testing.T) {
	definitions := []dashboardWidgetDefinition{
		{
			terraformPath: "widget.0.group_definition.0.widget.0.timeseries_definition.0",
			definition: map[string]interface{}{
				"type": "timeseries",
				"requests": []interface{}{
					map[string]interface{}{"q": "avg:system.load.1{*}", "display_type": "line"},
				},
			},
		},
	}

	t.Run("valid", func(t *testing.T) {
		client := dashboardValidationTestClient(t, func(w http.ResponseWriter, r *http.Request) {
			assertDashboardValidationRequest(t, r, 1)
			writeDashboardValidationResponse(t, w, map[string]interface{}{
				"results": []interface{}{map[string]interface{}{"is_valid": true, "widget_type": "timeseries"}},
			})
		})

		if err := validateDashboardWidgetDefinitions(context.Background(), context.Background(), client, definitions); err != nil {
			t.Fatalf("unexpected validation error: %v", err)
		}
	})

	t.Run("invalid query", func(t *testing.T) {
		client := dashboardValidationTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
			writeDashboardValidationResponse(t, w, map[string]interface{}{
				"results": []interface{}{map[string]interface{}{
					"is_valid":      false,
					"widget_type":   "timeseries",
					"error_path":    "requests.0.q",
					"error_message": "invalid metric query: missing closing brace",
				}},
			})
		})

		err := validateDashboardWidgetDefinitions(context.Background(), context.Background(), client, definitions)
		if err == nil {
			t.Fatal("expected validation error")
		}
		for _, expected := range []string{"widget.0.group_definition.0.widget.0.timeseries_definition.0.request.0.q", "timeseries", "missing closing brace"} {
			if !strings.Contains(err.Error(), expected) {
				t.Fatalf("expected error %q to contain %q", err, expected)
			}
		}
	})

	t.Run("response cardinality mismatch", func(t *testing.T) {
		client := dashboardValidationTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
			writeDashboardValidationResponse(t, w, map[string]interface{}{"results": []interface{}{}})
		})

		err := validateDashboardWidgetDefinitions(context.Background(), context.Background(), client, definitions)
		if err == nil || !strings.Contains(err.Error(), "0 results for 1 widgets") {
			t.Fatalf("expected cardinality error, got %v", err)
		}
	})

	t.Run("endpoint error", func(t *testing.T) {
		client := dashboardValidationTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(`{"errors":["invalid request"]}`))
		})

		err := validateDashboardWidgetDefinitions(context.Background(), context.Background(), client, definitions)
		if err == nil || !strings.Contains(err.Error(), "error validating dashboard widgets") {
			t.Fatalf("expected endpoint error, got %v", err)
		}
	})
}

func dashboardValidationTestClient(t *testing.T, handler http.HandlerFunc) *datadog.APIClient {
	t.Helper()
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)

	configuration := datadog.NewConfiguration()
	configuration.HTTPClient = server.Client()
	configuration.Servers = datadog.ServerConfigurations{{URL: server.URL}}
	return datadog.NewAPIClient(configuration)
}

func assertDashboardValidationRequest(t *testing.T, request *http.Request, definitionCount int) []map[string]interface{} {
	t.Helper()
	if request.Method != http.MethodPost {
		t.Fatalf("expected POST, got %s", request.Method)
	}
	if request.URL.Path != dashboardWidgetValidationPath {
		t.Fatalf("expected path %q, got %q", dashboardWidgetValidationPath, request.URL.Path)
	}
	var body struct {
		WidgetDefinitions []map[string]interface{} `json:"widget_definitions"`
	}
	if err := json.NewDecoder(request.Body).Decode(&body); err != nil {
		t.Fatalf("decode request: %v", err)
	}
	if len(body.WidgetDefinitions) != definitionCount {
		t.Fatalf("expected %d definitions, got %d", definitionCount, len(body.WidgetDefinitions))
	}
	return body.WidgetDefinitions
}

func writeDashboardValidationResponse(t *testing.T, writer http.ResponseWriter, response interface{}) {
	t.Helper()
	writer.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(writer).Encode(response); err != nil {
		t.Fatalf("encode response: %v", err)
	}
}

func dashboardV2TestTimeseriesWidgets(query string) []interface{} {
	return []interface{}{
		map[string]interface{}{
			"timeseries_definition": []interface{}{
				map[string]interface{}{
					"request": []interface{}{
						map[string]interface{}{
							"display_type": "line",
							"q":            query,
						},
					},
				},
			},
		},
	}
}
