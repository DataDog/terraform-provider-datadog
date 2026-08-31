package dashboardmapping

import "testing"

func TestHostmapDDSQLRequestRoundTrip(t *testing.T) {
	widget := map[string]interface{}{
		"hostmap_definition": []interface{}{map[string]interface{}{
			"title": "DDSQL host map",
			"request": []interface{}{map[string]interface{}{
				"request_type": "data_projection",
				"query": []interface{}{map[string]interface{}{
					"data_source":      "dataset",
					"dataset_provider": "ddsql_query",
					"dataset_id":       "dataset-123",
					"filter":           "service:web-store",
					"limit":            100,
					"sort": []interface{}{map[string]interface{}{
						"field": []interface{}{map[string]interface{}{
							"name":  "cpu_usage",
							"order": "desc",
						}},
					}},
				}},
				"projection": []interface{}{map[string]interface{}{
					"type": "hostmap",
					"dimension": []interface{}{
						map[string]interface{}{"column": "entity_id", "dimension": "node"},
						map[string]interface{}{"column": "service", "dimension": "group"},
						map[string]interface{}{
							"column": "cpu_usage", "dimension": "fill", "alias": "CPU",
							"number_format": []interface{}{map[string]interface{}{
								"unit": []interface{}{map[string]interface{}{
									"custom": []interface{}{map[string]interface{}{"label": "%"}},
								}},
							}},
						},
					},
				}},
				"limit": 250,
			}},
		}},
	}

	built := BuildWidgetEngineJSONFromMap(widget)
	definition := built["definition"].(map[string]interface{})
	request := definition["requests"].(map[string]interface{})
	if request["request_type"] != "data_projection" || request["limit"] != 250 {
		t.Fatalf("DDSQL request was not serialized: %#v", request)
	}
	query := request["query"].(map[string]interface{})
	if query["dataset_provider"] != "ddsql_query" || query["dataset_id"] != "dataset-123" {
		t.Fatalf("dataset query was not serialized: %#v", query)
	}
	sort := query["sort"].(map[string]interface{})
	fields := sort["fields"].([]interface{})
	if fields[0].(map[string]interface{})["order"] != "desc" {
		t.Fatalf("dataset sort was not serialized: %#v", sort)
	}
	projection := request["projection"].(map[string]interface{})
	dimensions := projection["dimensions"].([]interface{})
	if dimensions[1].(map[string]interface{})["dimension"] != "group" {
		t.Fatalf("projection dimensions were not serialized: %#v", projection)
	}

	flattened, _ := FlattenWidgetEngineJSON(built)
	flatDefinition := flattened["hostmap_definition"].([]interface{})[0].(map[string]interface{})
	flatRequest := flatDefinition["request"].([]interface{})[0].(map[string]interface{})
	if flatRequest["request_type"] != "data_projection" || flatRequest["limit"] != 250 {
		t.Fatalf("DDSQL request was not restored: %#v", flatRequest)
	}
	flatQuery := flatRequest["query"].([]interface{})[0].(map[string]interface{})
	if flatQuery["dataset_id"] != "dataset-123" {
		t.Fatalf("dataset query was not restored: %#v", flatQuery)
	}
	flatProjection := flatRequest["projection"].([]interface{})[0].(map[string]interface{})
	flatDimensions := flatProjection["dimension"].([]interface{})
	if flatDimensions[2].(map[string]interface{})["alias"] != "CPU" {
		t.Fatalf("projection dimensions were not restored: %#v", flatProjection)
	}
}
