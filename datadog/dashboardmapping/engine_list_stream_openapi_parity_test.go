package dashboardmapping

import "testing"

func TestListStreamOpenAPIParityRoundTrip(t *testing.T) {
	widget := map[string]interface{}{
		"list_stream_definition": []interface{}{map[string]interface{}{
			"request": []interface{}{map[string]interface{}{
				"response_format": "event_list",
				"columns": []interface{}{map[string]interface{}{
					"field": "timestamp",
					"width": "auto",
				}},
				"query": []interface{}{map[string]interface{}{
					"data_source":      "issue_stream",
					"query_string":     "service:web-store",
					"states":           []interface{}{"OPEN", "ACKNOWLEDGED"},
					"assignee_uuids":   []interface{}{"user-uuid"},
					"suspected_causes": []interface{}{"deployment"},
					"team_handles":     []interface{}{"team-platform"},
					"persona":          "backend",
					"version":          "sequential_query",
					"compute": []interface{}{map[string]interface{}{
						"aggregation": "count",
						"facet":       "resource_name",
					}},
				}},
			}},
		}},
	}

	built := BuildWidgetEngineJSONFromMap(widget)
	definition := built["definition"].(map[string]interface{})
	request := definition["requests"].([]interface{})[0].(map[string]interface{})
	query := request["query"].(map[string]interface{})
	if query["persona"] != "backend" {
		t.Fatalf("list stream persona was not serialized: %#v", query)
	}
	if query["version"] != "sequential_query" {
		t.Fatalf("list stream version was not serialized: %#v", query)
	}
	if len(query["compute"].([]interface{})) != 1 {
		t.Fatalf("list stream compute was not serialized: %#v", query)
	}

	flattened, _ := FlattenWidgetEngineJSON(built)
	flatDefinition := flattened["list_stream_definition"].([]interface{})[0].(map[string]interface{})
	flatRequest := flatDefinition["request"].([]interface{})[0].(map[string]interface{})
	flatQuery := flatRequest["query"].([]interface{})[0].(map[string]interface{})
	if flatQuery["persona"] != "backend" {
		t.Fatalf("list stream fields were not restored during flatten: %#v", flatQuery)
	}
	if flatQuery["version"] != "sequential_query" {
		t.Fatalf("list stream version was not restored during flatten: %#v", flatQuery)
	}
}
