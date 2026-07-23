package dashboardmapping

import "testing"

func TestApmMetricsFormulaAndDistributionHistogram(t *testing.T) {
	apmMetricsQuery := map[string]interface{}{
		"data_source":    "apm_metrics",
		"name":           "query1",
		"stat":           "hits",
		"service":        "web-store",
		"operation_name": "web.request",
		"span_kind":      "server",
	}
	formulaWidget := map[string]interface{}{
		"query_value_definition": []interface{}{map[string]interface{}{
			"request": []interface{}{map[string]interface{}{
				"formula": []interface{}{map[string]interface{}{"formula_expression": "query1"}},
				"query": []interface{}{map[string]interface{}{
					"apm_metrics_query": []interface{}{apmMetricsQuery},
				}},
			}},
		}},
	}
	builtFormula := BuildWidgetEngineJSONFromMap(formulaWidget)
	formulaDefinition := builtFormula["definition"].(map[string]interface{})
	formulaRequest := formulaDefinition["requests"].([]interface{})[0].(map[string]interface{})
	builtQuery := formulaRequest["queries"].([]interface{})[0].(map[string]interface{})
	if builtQuery["data_source"] != "apm_metrics" {
		t.Fatalf("APM metrics query was not serialized: %#v", builtQuery)
	}

	distributionWidget := map[string]interface{}{
		"distribution_definition": []interface{}{map[string]interface{}{
			"request": []interface{}{map[string]interface{}{
				"request_type": "histogram",
				"histogram_query": []interface{}{map[string]interface{}{
					"apm_metrics_query": []interface{}{apmMetricsQuery},
				}},
			}},
		}},
	}
	builtDistribution := BuildWidgetEngineJSONFromMap(distributionWidget)
	distributionDefinition := builtDistribution["definition"].(map[string]interface{})
	distributionRequest := distributionDefinition["requests"].([]interface{})[0].(map[string]interface{})
	histogramQuery := distributionRequest["query"].(map[string]interface{})
	if distributionRequest["request_type"] != "histogram" || histogramQuery["data_source"] != "apm_metrics" {
		t.Fatalf("APM metrics histogram query was not serialized: %#v", distributionRequest)
	}
}
