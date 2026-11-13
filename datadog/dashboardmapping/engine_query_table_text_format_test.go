package dashboardmapping

import "testing"

// queryTableWidgetWithTextFormatReplace returns a query_table widget whose single
// text_format rule uses the given replace block.
func queryTableWidgetWithTextFormatReplace(replace map[string]interface{}) map[string]interface{} {
	return map[string]interface{}{
		"query_table_definition": []interface{}{map[string]interface{}{
			"request": []interface{}{map[string]interface{}{
				"q": "avg:system.cpu.user{*} by {host}",
				"text_formats": []interface{}{map[string]interface{}{
					"text_format": []interface{}{map[string]interface{}{
						"match": []interface{}{map[string]interface{}{
							"type":  "contains",
							"value": "Incentives",
						}},
						"replace": []interface{}{replace},
					}},
				}},
			}},
		}},
	}
}

func builtQueryTableTextFormatReplace(t *testing.T, widget map[string]interface{}) map[string]interface{} {
	t.Helper()
	built := BuildWidgetEngineJSONFromMap(widget)
	definition := built["definition"].(map[string]interface{})
	request := definition["requests"].([]interface{})[0].(map[string]interface{})
	rule := request["text_formats"].([]interface{})[0].([]interface{})[0].(map[string]interface{})
	replace, ok := rule["replace"].(map[string]interface{})
	if !ok {
		t.Fatalf("replace missing from text format rule: %#v", rule)
	}
	return replace
}

// An empty `with` is how a substring replace strips a prefix. The API requires the
// `with` key for both replace variants, so it must be sent even when it is "".
func TestQueryTableTextFormatReplaceSendsEmptyWith(t *testing.T) {
	for _, tc := range []struct {
		name    string
		replace map[string]interface{}
	}{
		{
			name: "substring",
			replace: map[string]interface{}{
				"type":      "substring",
				"substring": `App\Http\Controllers\`,
				"with":      "",
			},
		},
		{
			name: "all",
			replace: map[string]interface{}{
				"type": "all",
				"with": "",
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			widget := queryTableWidgetWithTextFormatReplace(tc.replace)
			replace := builtQueryTableTextFormatReplace(t, widget)
			with, ok := replace["with"]
			if !ok {
				t.Fatalf("empty `with` was dropped from replace: %#v", replace)
			}
			if with != "" {
				t.Fatalf("expected empty `with`, got %#v", with)
			}
			if replace["type"] != tc.replace["type"] {
				t.Fatalf("expected type %v, got %#v", tc.replace["type"], replace)
			}

			built := BuildWidgetEngineJSONFromMap(widget)
			flattened, droppedPaths := FlattenWidgetEngineJSON(built)
			if len(droppedPaths) != 0 {
				t.Fatalf("text format fields were dropped during flatten: %#v", droppedPaths)
			}
			flatRequest := flattened["query_table_definition"].([]interface{})[0].(map[string]interface{})["request"].([]interface{})[0].(map[string]interface{})
			flatRule := flatRequest["text_formats"].([]interface{})[0].(map[string]interface{})["text_format"].([]interface{})[0].(map[string]interface{})
			flatReplace := flatRule["replace"].([]interface{})[0].(map[string]interface{})
			if flatWith, ok := flatReplace["with"]; !ok || flatWith != "" {
				t.Fatalf("empty `with` was not restored during flatten: %#v", flatReplace)
			}
		})
	}
}

func TestQueryTableTextFormatReplaceSendsNonEmptyWith(t *testing.T) {
	replace := builtQueryTableTextFormatReplace(t, queryTableWidgetWithTextFormatReplace(map[string]interface{}{
		"type":      "substring",
		"substring": "foo",
		"with":      "bar",
	}))
	if replace["type"] != "substring" || replace["substring"] != "foo" || replace["with"] != "bar" {
		t.Fatalf("substring replace was not serialized: %#v", replace)
	}
}
