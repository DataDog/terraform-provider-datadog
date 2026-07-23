package dashboardmapping

import "testing"

func TestQueryValueComparisonRoundTrip(t *testing.T) {
	widget := map[string]interface{}{
		"query_value_definition": []interface{}{map[string]interface{}{
			"request": []interface{}{map[string]interface{}{
				"q": "avg:system.cpu.user{*}",
				"comparison": []interface{}{map[string]interface{}{
					"type":           "absolute",
					"directionality": "neutral",
					"duration": []interface{}{map[string]interface{}{
						"type": "custom_timeframe",
						"custom_timeframe": []interface{}{map[string]interface{}{
							"from": 1779290190000,
							"to":   1779894990000,
						}},
					}},
				}},
			}},
		}},
	}

	built := BuildWidgetEngineJSONFromMap(widget)
	definition := built["definition"].(map[string]interface{})
	request := definition["requests"].([]interface{})[0].(map[string]interface{})
	comparison := request["comparison"].(map[string]interface{})
	if comparison["type"] != "absolute" || comparison["directionality"] != "neutral" {
		t.Fatalf("comparison defaults were not serialized: %#v", comparison)
	}
	duration := comparison["duration"].(map[string]interface{})
	custom := duration["custom_timeframe"].(map[string]interface{})
	if custom["from"] != 1779290190000 || custom["to"] != 1779894990000 {
		t.Fatalf("custom comparison timeframe was not serialized: %#v", custom)
	}

	flattened, _ := FlattenWidgetEngineJSON(built)
	flatDefinition := flattened["query_value_definition"].([]interface{})[0].(map[string]interface{})
	flatRequest := flatDefinition["request"].([]interface{})[0].(map[string]interface{})
	if _, ok := flatRequest["comparison"]; !ok {
		t.Fatalf("comparison was not restored during flatten: %#v", flatRequest)
	}
}

func TestQueryValueFormulaComparisonRoundTrip(t *testing.T) {
	widget := map[string]interface{}{
		"query_value_definition": []interface{}{map[string]interface{}{
			"request": []interface{}{map[string]interface{}{
				"formula": []interface{}{map[string]interface{}{
					"formula_expression": "query1",
				}},
				"query": []interface{}{map[string]interface{}{
					"metric_query": []interface{}{map[string]interface{}{
						"data_source": "metrics",
						"name":        "query1",
						"query":       "avg:system.cpu.user{*}",
					}},
				}},
				"comparison": []interface{}{map[string]interface{}{
					"type":           "relative",
					"directionality": "decrease_better",
					"duration": []interface{}{map[string]interface{}{
						"type": "previous_week",
					}},
				}},
			}},
		}},
	}

	built := BuildWidgetEngineJSONFromMap(widget)
	definition := built["definition"].(map[string]interface{})
	request := definition["requests"].([]interface{})[0].(map[string]interface{})
	comparison, ok := request["comparison"].(map[string]interface{})
	if !ok || comparison["type"] != "relative" || comparison["directionality"] != "decrease_better" {
		t.Fatalf("formula request comparison was not serialized: %#v", request)
	}
	duration, ok := comparison["duration"].(map[string]interface{})
	if !ok || duration["type"] != "previous_week" {
		t.Fatalf("formula request comparison duration was not serialized: %#v", request)
	}

	flattened, _ := FlattenWidgetEngineJSON(built)
	flatDefinition := flattened["query_value_definition"].([]interface{})[0].(map[string]interface{})
	flatRequest := flatDefinition["request"].([]interface{})[0].(map[string]interface{})
	flatComparison := getBlockFromMap(flatRequest, "comparison")
	if flatComparison == nil || flatComparison["type"] != "relative" || flatComparison["directionality"] != "decrease_better" {
		t.Fatalf("formula request comparison was not restored: %#v", flatRequest)
	}
	flatDuration := getBlockFromMap(flatComparison, "duration")
	if flatDuration == nil || flatDuration["type"] != "previous_week" {
		t.Fatalf("formula request comparison duration was not restored: %#v", flatRequest)
	}
}
