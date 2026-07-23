package dashboardmapping

import "testing"

func TestGeomapFormulaRequestDefaultsToScalar(t *testing.T) {
	widget := map[string]interface{}{
		"geomap_definition": []interface{}{map[string]interface{}{
			"request": []interface{}{map[string]interface{}{
				"query": []interface{}{map[string]interface{}{
					"metric_query": []interface{}{map[string]interface{}{
						"data_source": "metrics",
						"name":        "query1",
						"query":       "avg:system.cpu.user{*} by {country}",
					}},
				}},
			}},
		}},
	}

	built := BuildWidgetEngineJSONFromMap(widget)
	definition := built["definition"].(map[string]interface{})
	request := definition["requests"].([]interface{})[0].(map[string]interface{})
	if request["response_format"] != "scalar" {
		t.Fatalf("geomap formula request did not preserve the scalar default: %#v", request)
	}
}

func TestGeomapFormulaRequestFieldsRoundTrip(t *testing.T) {
	widget := map[string]interface{}{
		"geomap_definition": []interface{}{map[string]interface{}{
			"style": []interface{}{map[string]interface{}{
				"palette":      "hostmap_blues",
				"palette_flip": false,
			}},
			"view": []interface{}{map[string]interface{}{"focus": "WORLD"}},
			"request": []interface{}{map[string]interface{}{
				"response_format": "timeseries",
				"formula": []interface{}{map[string]interface{}{
					"formula_expression": "query1",
				}},
				"query": []interface{}{map[string]interface{}{
					"metric_query": []interface{}{map[string]interface{}{
						"data_source": "metrics",
						"name":        "query1",
						"query":       "avg:system.cpu.user{*} by {country}",
					}},
				}},
				"conditional_formats": []interface{}{map[string]interface{}{
					"comparator": ">",
					"value":      80.0,
					"palette":    "white_on_red",
					"hide_value": false,
				}},
				"style": []interface{}{map[string]interface{}{"color_by": "status"}},
				"sort": []interface{}{map[string]interface{}{
					"count": 20,
					"order_by": []interface{}{map[string]interface{}{
						"formula_sort": []interface{}{map[string]interface{}{
							"index": 1,
							"order": "desc",
						}},
					}},
				}},
			}},
		}},
	}

	built := BuildWidgetEngineJSONFromMap(widget)
	definition := built["definition"].(map[string]interface{})
	request := definition["requests"].([]interface{})[0].(map[string]interface{})
	if request["response_format"] != "timeseries" {
		t.Fatalf("geomap response format override was not serialized: %#v", request)
	}
	style := request["style"].(map[string]interface{})
	if style["color_by"] != "status" {
		t.Fatalf("geomap request style was not serialized: %#v", request)
	}
	if _, ok := request["conditional_formats"]; !ok {
		t.Fatalf("geomap conditional formats were not serialized: %#v", request)
	}
	if _, ok := request["sort"]; !ok {
		t.Fatalf("geomap sort was not serialized: %#v", request)
	}

	flattened, _ := FlattenWidgetEngineJSON(built)
	flatDefinition := flattened["geomap_definition"].([]interface{})[0].(map[string]interface{})
	flatRequest := flatDefinition["request"].([]interface{})[0].(map[string]interface{})
	if flatRequest["response_format"] != "timeseries" {
		t.Fatalf("geomap response format was not restored: %#v", flatRequest)
	}
	flatStyle := flatRequest["style"].([]interface{})[0].(map[string]interface{})
	if flatStyle["color_by"] != "status" {
		t.Fatalf("geomap request style was not restored: %#v", flatRequest)
	}
}

func TestGeomapEventListRequestRoundTrip(t *testing.T) {
	widget := map[string]interface{}{
		"geomap_definition": []interface{}{map[string]interface{}{
			"style": []interface{}{map[string]interface{}{
				"palette":      "hostmap_blues",
				"palette_flip": false,
			}},
			"view": []interface{}{map[string]interface{}{"focus": "WORLD"}},
			"request": []interface{}{map[string]interface{}{
				"response_format": "event_list",
				"columns": []interface{}{map[string]interface{}{
					"field": "timestamp",
					"width": "auto",
				}},
				"list_stream_query": []interface{}{map[string]interface{}{
					"data_source":  "rum_stream",
					"query_string": "@type:view",
				}},
				"style": []interface{}{map[string]interface{}{"color_by": "status"}},
				"text_format": []interface{}{map[string]interface{}{
					"match": []interface{}{map[string]interface{}{
						"type":  "is",
						"value": "error",
					}},
					"palette": "white_on_red",
				}},
			}},
		}},
	}

	built := BuildWidgetEngineJSONFromMap(widget)
	definition := built["definition"].(map[string]interface{})
	request := definition["requests"].([]interface{})[0].(map[string]interface{})
	if request["response_format"] != "event_list" {
		t.Fatalf("event-list response format was not serialized: %#v", request)
	}
	query := request["query"].(map[string]interface{})
	if query["data_source"] != "rum_stream" || query["query_string"] != "@type:view" {
		t.Fatalf("geomap list stream query was not serialized: %#v", request)
	}
	if len(request["columns"].([]interface{})) != 1 || len(request["text_formats"].([]interface{})) != 1 {
		t.Fatalf("geomap point-layer fields were not serialized: %#v", request)
	}

	flattened, _ := FlattenWidgetEngineJSON(built)
	flatDefinition := flattened["geomap_definition"].([]interface{})[0].(map[string]interface{})
	flatRequest := flatDefinition["request"].([]interface{})[0].(map[string]interface{})
	flatQuery := flatRequest["list_stream_query"].([]interface{})[0].(map[string]interface{})
	if flatQuery["data_source"] != "rum_stream" {
		t.Fatalf("geomap list stream query was not restored: %#v", flatRequest)
	}
	if len(flatRequest["columns"].([]interface{})) != 1 || len(flatRequest["text_format"].([]interface{})) != 1 {
		t.Fatalf("geomap point-layer fields were not restored: %#v", flatRequest)
	}
}
