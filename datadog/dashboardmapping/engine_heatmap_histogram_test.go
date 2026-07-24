package dashboardmapping

import "testing"

func TestHeatmapHistogramRequestRoundTrip(t *testing.T) {
	widget := map[string]interface{}{
		"heatmap_definition": []interface{}{map[string]interface{}{
			"request": []interface{}{map[string]interface{}{
				"request_type": "histogram",
				"histogram_query": []interface{}{map[string]interface{}{
					"metric_query": []interface{}{map[string]interface{}{
						"data_source": "metrics",
						"name":        "query1",
						"query":       "histogram:trace.servlet.request{*}",
					}},
				}},
				"style": []interface{}{map[string]interface{}{
					"palette": "dog_classic",
				}},
			}},
		}},
	}

	built := BuildWidgetEngineJSONFromMap(widget)
	definition := built["definition"].(map[string]interface{})
	request := definition["requests"].([]interface{})[0].(map[string]interface{})
	if request["request_type"] != "histogram" {
		t.Fatalf("histogram discriminator was not serialized: %#v", request)
	}
	query := request["query"].(map[string]interface{})
	if query["data_source"] != "metrics" || query["query"] != "histogram:trace.servlet.request{*}" {
		t.Fatalf("histogram query was not serialized: %#v", query)
	}
	style := request["style"].(map[string]interface{})
	if style["palette"] != "dog_classic" {
		t.Fatalf("heatmap style was not preserved: %#v", request)
	}

	flattened, _ := FlattenWidgetEngineJSON(built)
	flatDefinition := flattened["heatmap_definition"].([]interface{})[0].(map[string]interface{})
	flatRequest := flatDefinition["request"].([]interface{})[0].(map[string]interface{})
	if flatRequest["request_type"] != "histogram" {
		t.Fatalf("histogram discriminator was not restored: %#v", flatRequest)
	}
	histogramQuery := flatRequest["histogram_query"].([]interface{})[0].(map[string]interface{})
	metricQuery := histogramQuery["metric_query"].([]interface{})[0].(map[string]interface{})
	if metricQuery["name"] != "query1" {
		t.Fatalf("histogram query was not restored: %#v", histogramQuery)
	}
}

func TestHeatmapFormulaRequestStillBuilds(t *testing.T) {
	widget := map[string]interface{}{
		"heatmap_definition": []interface{}{map[string]interface{}{
			"request": []interface{}{map[string]interface{}{
				"formula": []interface{}{map[string]interface{}{"formula_expression": "query1"}},
				"query": []interface{}{map[string]interface{}{
					"metric_query": []interface{}{map[string]interface{}{
						"data_source": "metrics",
						"name":        "query1",
						"query":       "avg:system.cpu.user{*}",
					}},
				}},
			}},
		}},
	}

	built := BuildWidgetEngineJSONFromMap(widget)
	definition := built["definition"].(map[string]interface{})
	request := definition["requests"].([]interface{})[0].(map[string]interface{})
	if request["response_format"] != "timeseries" || len(request["queries"].([]interface{})) != 1 {
		t.Fatalf("formula heatmap request changed: %#v", request)
	}
}

func TestHeatmapHistogramRequestBuildsAlongsideFormulaRequest(t *testing.T) {
	widget := map[string]interface{}{
		"heatmap_definition": []interface{}{map[string]interface{}{
			"request": []interface{}{
				map[string]interface{}{
					"histogram_query": []interface{}{map[string]interface{}{
						"metric_query": []interface{}{map[string]interface{}{
							"data_source": "metrics",
							"name":        "histogram",
							"query":       "histogram:trace.servlet.request{*}",
						}},
					}},
				},
				map[string]interface{}{
					"formula": []interface{}{map[string]interface{}{"formula_expression": "query1"}},
					"query": []interface{}{map[string]interface{}{
						"metric_query": []interface{}{map[string]interface{}{
							"data_source": "metrics",
							"name":        "query1",
							"query":       "avg:system.cpu.user{*}",
						}},
					}},
				},
			},
		}},
	}

	built := BuildWidgetEngineJSONFromMap(widget)
	definition := built["definition"].(map[string]interface{})
	requests := definition["requests"].([]interface{})
	if len(requests) != 2 {
		t.Fatalf("expected both heatmap requests, got: %#v", requests)
	}
	histogramRequest := requests[0].(map[string]interface{})
	if histogramRequest["request_type"] != "histogram" {
		t.Fatalf("histogram request was not inferred from histogram_query: %#v", histogramRequest)
	}
	formulaRequest := requests[1].(map[string]interface{})
	if formulaRequest["response_format"] != "timeseries" {
		t.Fatalf("formula request was not preserved: %#v", formulaRequest)
	}
}
