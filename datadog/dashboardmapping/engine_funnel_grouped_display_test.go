package dashboardmapping

import "testing"

func TestFunnelGroupedDisplayRoundTrip(t *testing.T) {
	widget := map[string]interface{}{
		"funnel_definition": []interface{}{map[string]interface{}{
			"grouped_display": "side_by_side",
			"request": []interface{}{map[string]interface{}{
				"query": []interface{}{map[string]interface{}{
					"data_source":  "rum",
					"query_string": "@type:view",
					"step": []interface{}{map[string]interface{}{
						"facet": "@view.name",
						"value": "/home",
					}},
				}},
			}},
		}},
	}

	built := BuildWidgetEngineJSONFromMap(widget)
	definition := built["definition"].(map[string]interface{})
	if definition["grouped_display"] != "side_by_side" {
		t.Fatalf("funnel grouped_display was not serialized: %#v", definition)
	}
	flattened, _ := FlattenWidgetEngineJSON(built)
	flatDefinition := flattened["funnel_definition"].([]interface{})[0].(map[string]interface{})
	if flatDefinition["grouped_display"] != "side_by_side" {
		t.Fatalf("funnel grouped_display was not restored: %#v", flatDefinition)
	}
}
