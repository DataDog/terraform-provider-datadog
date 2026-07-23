package dashboardmapping

import "testing"

func TestQueryTableSortBuildsForFormulaAndLegacyRequests(t *testing.T) {
	requests := []interface{}{
		map[string]interface{}{
			"formula": []interface{}{map[string]interface{}{"formula_expression": "query1"}},
			"query": []interface{}{map[string]interface{}{
				"metric_query": []interface{}{map[string]interface{}{
					"data_source": "metrics",
					"name":        "query1",
					"query":       "avg:system.cpu.user{*}",
				}},
			}},
			"sort": []interface{}{map[string]interface{}{
				"order_by": []interface{}{map[string]interface{}{
					"formula_sort": []interface{}{map[string]interface{}{"index": 1, "order": "desc"}},
				}},
			}},
		},
		map[string]interface{}{
			"q": "avg:system.mem.used{*}",
			"sort": []interface{}{map[string]interface{}{
				"order_by": []interface{}{map[string]interface{}{
					"group_sort": []interface{}{map[string]interface{}{"name": "host", "order": "asc"}},
				}},
			}},
		},
	}
	widget := map[string]interface{}{
		"query_table_definition": []interface{}{map[string]interface{}{"request": requests}},
	}

	built := BuildWidgetEngineJSONFromMap(widget)
	definition := built["definition"].(map[string]interface{})
	builtRequests := definition["requests"].([]interface{})
	for i, request := range builtRequests {
		if _, ok := request.(map[string]interface{})["sort"]; !ok {
			t.Fatalf("sort missing from request %d: %#v", i, request)
		}
	}

	flattened, _ := FlattenWidgetEngineJSON(built)
	flatDefinition := flattened["query_table_definition"].([]interface{})[0].(map[string]interface{})
	flatRequests := flatDefinition["request"].([]interface{})
	for i, request := range flatRequests {
		if _, ok := request.(map[string]interface{})["sort"]; !ok {
			t.Fatalf("sort missing after flatten for request %d: %#v", i, request)
		}
	}
}
