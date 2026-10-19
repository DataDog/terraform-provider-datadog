package dashboardmapping

import "testing"

func TestSankeyAudienceFieldsRoundTrip(t *testing.T) {
	widget := map[string]interface{}{
		"sankey_definition": []interface{}{map[string]interface{}{
			"request": []interface{}{map[string]interface{}{
				"rum_request": []interface{}{map[string]interface{}{
					"query": []interface{}{map[string]interface{}{
						"data_source":  "product_analytics",
						"query_string": "@type:view",
						"mode":         "source",
						"audience_filters": []interface{}{map[string]interface{}{
							"user": []interface{}{map[string]interface{}{
								"name":  "buyers",
								"query": "@usr.plan:pro",
							}},
							"filter_condition": "users",
						}},
						"occurrence": []interface{}{map[string]interface{}{
							"operator": "gt",
							"value":    "2",
						}},
						"join_keys": []interface{}{map[string]interface{}{
							"primary":   "session.id",
							"secondary": []interface{}{"usr.id"},
						}},
					}},
				}},
			}},
		}},
	}

	built := BuildWidgetEngineJSONFromMap(widget)
	definition := built["definition"].(map[string]interface{})
	request := definition["requests"].([]interface{})[0].(map[string]interface{})
	query := request["query"].(map[string]interface{})
	if request["request_type"] != "sankey" {
		t.Fatalf("sankey discriminator was not serialized: %#v", request)
	}
	if query["join_keys"].(map[string]interface{})["primary"] != "session.id" {
		t.Fatalf("sankey join keys were not serialized: %#v", query)
	}
	occurrence, ok := query["occurrences"].(map[string]interface{})
	if !ok || occurrence["operator"] != "gt" || occurrence["value"] != "2" {
		t.Fatalf("sankey occurrence filter was not serialized under the plural JSON key: %#v", query)
	}
	if _, ok := query["occurrence"]; ok {
		t.Fatalf("sankey occurrence filter leaked its singular HCL key into JSON: %#v", query)
	}

	flattened, _ := FlattenWidgetEngineJSON(built)
	flatDefinition := flattened["sankey_definition"].([]interface{})[0].(map[string]interface{})
	flatRequest := flatDefinition["request"].([]interface{})[0].(map[string]interface{})
	flatRum := flatRequest["rum_request"].([]interface{})[0].(map[string]interface{})
	flatQuery := flatRum["query"].([]interface{})[0].(map[string]interface{})
	if _, ok := flatQuery["audience_filters"]; !ok {
		t.Fatalf("sankey audience filters were not restored during flatten: %#v", flatQuery)
	}
}
