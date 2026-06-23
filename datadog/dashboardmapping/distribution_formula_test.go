package dashboardmapping

import "testing"

func TestDistributionFormulaRequestBuildsLLMObservabilityQuery(t *testing.T) {
	widget := map[string]interface{}{
		"distribution_definition": []interface{}{
			map[string]interface{}{
				"request": []interface{}{
					distributionFormulaRequestState(),
				},
			},
		},
	}

	built := buildWidgetEngineJSONFromMap(widget)
	definition := built["definition"].(map[string]interface{})
	requests := definition["requests"].([]interface{})
	request := requests[0].(map[string]interface{})

	if got := request["response_format"]; got != "scalar" {
		t.Fatalf("expected scalar response_format, got %#v", got)
	}

	queries := request["queries"].([]interface{})
	eventQuery := queries[0].(map[string]interface{})
	if got := eventQuery["data_source"]; got != "llm_observability" {
		t.Fatalf("expected llm_observability data source, got %#v", got)
	}

	if _, hasLegacyQuery := request["query"]; hasLegacyQuery {
		t.Fatalf("expected plural queries JSON, got singular query in %#v", request)
	}
}

func TestDistributionFormulaRequestFlattensLLMObservabilityQuery(t *testing.T) {
	widgets := []interface{}{
		map[string]interface{}{
			"definition": map[string]interface{}{
				"type": "distribution",
				"requests": []interface{}{
					map[string]interface{}{
						"response_format": "scalar",
						"queries": []interface{}{
							map[string]interface{}{
								"data_source": "llm_observability",
								"name":        "query1",
								"compute": map[string]interface{}{
									"aggregation": "count",
								},
							},
						},
						"formulas": []interface{}{
							map[string]interface{}{
								"formula": "query1",
							},
						},
					},
				},
			},
		},
	}

	flattened, drops := FlattenWidgetsForSDKv2(widgets)
	if len(drops) != 0 {
		t.Fatalf("expected no dropped fields, got %v", drops)
	}

	widget := flattened[0].(map[string]interface{})
	definition := widget["distribution_definition"].([]interface{})[0].(map[string]interface{})
	request := definition["request"].([]interface{})[0].(map[string]interface{})
	query := request["query"].([]interface{})[0].(map[string]interface{})
	eventQuery := query["event_query"].([]interface{})[0].(map[string]interface{})

	if got := eventQuery["data_source"]; got != "llm_observability" {
		t.Fatalf("expected llm_observability data source, got %#v", got)
	}
}

func distributionFormulaRequestState() map[string]interface{} {
	return map[string]interface{}{
		"query": []interface{}{
			map[string]interface{}{
				"event_query": []interface{}{
					map[string]interface{}{
						"data_source": "llm_observability",
						"name":        "query1",
						"compute": []interface{}{
							map[string]interface{}{
								"aggregation": "count",
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
