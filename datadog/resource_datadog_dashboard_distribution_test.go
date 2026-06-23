package datadog

import (
	"testing"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV1"
)

func TestBuildDatadogDistributionRequestsFormulaEventQueryLLMObservability(t *testing.T) {
	terraformRequests := []interface{}{distributionLLMObsRequest()}

	datadogRequests := buildDatadogDistributionRequests(&terraformRequests)
	if len(*datadogRequests) != 1 {
		t.Fatalf("expected 1 request, got %d", len(*datadogRequests))
	}

	request := (*datadogRequests)[0]
	responseFormat, ok := request.GetResponseFormatOk()
	if !ok {
		t.Fatal("expected response_format to be set")
	}
	if got := string(*responseFormat); got != "scalar" {
		t.Fatalf("expected response_format scalar, got %q", got)
	}

	queries, ok := request.GetQueriesOk()
	if !ok || len(*queries) != 1 {
		t.Fatalf("expected 1 query, got %d", len(*queries))
	}

	eventQuery := (*queries)[0].FormulaAndFunctionEventQueryDefinition
	if eventQuery == nil {
		t.Fatal("expected event query")
	}
	if got := string(eventQuery.GetDataSource()); got != "llm_observability" {
		t.Fatalf("expected llm_observability data source, got %q", got)
	}

	formulas, ok := request.GetFormulasOk()
	if !ok || len(*formulas) != 1 {
		t.Fatalf("expected 1 formula, got %d", len(*formulas))
	}
	if got := (*formulas)[0].GetFormula(); got != "query1" {
		t.Fatalf("expected formula query1, got %q", got)
	}
}

func TestBuildTerraformDistributionRequestsFormulaEventQueryLLMObservability(t *testing.T) {
	terraformRequests := []interface{}{distributionLLMObsRequest()}
	datadogRequests := buildDatadogDistributionRequests(&terraformRequests)

	flattenedRequests := buildTerraformDistributionRequests(datadogRequests)
	if len(*flattenedRequests) != 1 {
		t.Fatalf("expected 1 request, got %d", len(*flattenedRequests))
	}

	queries, ok := (*flattenedRequests)[0]["query"].([]map[string]interface{})
	if !ok || len(queries) != 1 {
		t.Fatalf("expected 1 query, got %#v", (*flattenedRequests)[0]["query"])
	}

	eventQueries, ok := queries[0]["event_query"].([]map[string]interface{})
	if !ok || len(eventQueries) != 1 {
		t.Fatalf("expected 1 event query, got %#v", queries[0]["event_query"])
	}
	if got := eventQueryDataSourceString(t, eventQueries[0]["data_source"]); got != "llm_observability" {
		t.Fatalf("expected llm_observability data source, got %q", got)
	}

	formulas, ok := (*flattenedRequests)[0]["formula"].([]map[string]interface{})
	if !ok || len(formulas) != 1 {
		t.Fatalf("expected 1 formula, got %#v", (*flattenedRequests)[0]["formula"])
	}
	if got := formulas[0]["formula_expression"]; got != "query1" {
		t.Fatalf("expected formula query1, got %#v", got)
	}
}

func distributionLLMObsRequest() map[string]interface{} {
	return map[string]interface{}{
		"query": []interface{}{
			map[string]interface{}{
				"event_query": []interface{}{
					map[string]interface{}{
						"data_source": "llm_observability",
						"name":        "query1",
						"indexes":     []interface{}{},
						"compute": []interface{}{
							map[string]interface{}{
								"aggregation": "count",
							},
						},
						"search": []interface{}{
							map[string]interface{}{
								"query": "@ml_app:test-app",
							},
						},
					},
				},
			},
		},
		"formula": []interface{}{
			map[string]interface{}{
				"formula_expression": "query1",
			},
		},
	}
}

func eventQueryDataSourceString(t *testing.T, dataSource interface{}) string {
	t.Helper()

	switch v := dataSource.(type) {
	case string:
		return v
	case datadogV1.FormulaAndFunctionEventsDataSource:
		return string(v)
	case *datadogV1.FormulaAndFunctionEventsDataSource:
		return string(*v)
	default:
		t.Fatalf("unexpected data_source type %T", dataSource)
		return ""
	}
}
