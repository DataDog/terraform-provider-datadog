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
					"formula_sort": []interface{}{map[string]interface{}{"index": 0, "order": "desc"}},
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
	formulaSort := builtRequests[0].(map[string]interface{})["sort"].(map[string]interface{})["order_by"].([]interface{})[0].(map[string]interface{})
	if formulaSort["type"] != "formula" || formulaSort["index"] != 0 || formulaSort["order"] != "desc" {
		t.Fatalf("formula sort was not serialized: %#v", formulaSort)
	}
	groupSort := builtRequests[1].(map[string]interface{})["sort"].(map[string]interface{})["order_by"].([]interface{})[0].(map[string]interface{})
	if groupSort["type"] != "group" || groupSort["name"] != "host" || groupSort["order"] != "asc" {
		t.Fatalf("group sort was not serialized: %#v", groupSort)
	}

	flattened, droppedPaths := FlattenWidgetEngineJSON(built)
	if len(droppedPaths) != 0 {
		t.Fatalf("query table sort fields were dropped during flatten: %#v", droppedPaths)
	}
	flatDefinition := flattened["query_table_definition"].([]interface{})[0].(map[string]interface{})
	flatRequests := flatDefinition["request"].([]interface{})
	for i, request := range flatRequests {
		if _, ok := request.(map[string]interface{})["sort"]; !ok {
			t.Fatalf("sort missing after flatten for request %d: %#v", i, request)
		}
	}
	flatFormulaSort := flatRequests[0].(map[string]interface{})["sort"].([]interface{})[0].(map[string]interface{})["order_by"].([]interface{})[0].(map[string]interface{})["formula_sort"].([]interface{})[0].(map[string]interface{})
	if flatFormulaSort["index"] != 0 || flatFormulaSort["order"] != "desc" {
		t.Fatalf("formula sort was not restored during flatten: %#v", flatFormulaSort)
	}
	flatGroupSort := flatRequests[1].(map[string]interface{})["sort"].([]interface{})[0].(map[string]interface{})["order_by"].([]interface{})[0].(map[string]interface{})["group_sort"].([]interface{})[0].(map[string]interface{})
	if flatGroupSort["name"] != "host" || flatGroupSort["order"] != "asc" {
		t.Fatalf("group sort was not restored during flatten: %#v", flatGroupSort)
	}
}
