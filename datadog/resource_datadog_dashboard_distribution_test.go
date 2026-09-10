package datadog

import (
	"testing"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV1"
)

func TestBuildDatadogDistributionRequestsLLMObservabilityQuery(t *testing.T) {
	terraformRequests := []interface{}{distributionLLMObservabilityRequest()}

	datadogRequests := buildDatadogDistributionRequests(&terraformRequests)
	if len(*datadogRequests) != 1 {
		t.Fatalf("expected 1 request, got %d", len(*datadogRequests))
	}

	request := (*datadogRequests)[0]
	query, ok := request.AdditionalProperties["llm_observability_query"].(*datadogV1.LogQueryDefinition)
	if !ok {
		t.Fatalf("expected llm_observability_query additional property, got %#v", request.AdditionalProperties["llm_observability_query"])
	}
	if got := query.GetIndex(); got != "*" {
		t.Fatalf("expected index *, got %q", got)
	}
}

func TestBuildTerraformDistributionRequestsLLMObservabilityQuery(t *testing.T) {
	terraformRequests := []interface{}{distributionLLMObservabilityRequest()}
	datadogRequests := buildDatadogDistributionRequests(&terraformRequests)

	flattenedRequests := buildTerraformDistributionRequests(datadogRequests)
	if len(*flattenedRequests) != 1 {
		t.Fatalf("expected 1 request, got %d", len(*flattenedRequests))
	}

	queries, ok := (*flattenedRequests)[0]["llm_observability_query"].([]map[string]interface{})
	if !ok || len(queries) != 1 {
		t.Fatalf("expected 1 llm_observability_query, got %#v", (*flattenedRequests)[0]["llm_observability_query"])
	}
	if got := queries[0]["index"]; got != "*" {
		t.Fatalf("expected index *, got %#v", got)
	}
}

func distributionLLMObservabilityRequest() map[string]interface{} {
	return map[string]interface{}{
		"llm_observability_query": []interface{}{
			map[string]interface{}{
				"index": "*",
				"compute_query": []interface{}{
					map[string]interface{}{
						"aggregation": "count",
					},
				},
				"multi_compute": []interface{}{},
				"search_query":  "@ml_app:test-app",
			},
		},
	}
}
