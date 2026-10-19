package dashboardmapping

import "testing"

func TestBuildWidgetEngineJSONFromMap_PreservesZeroLayoutCoordinatesForNestedGroupWidget(t *testing.T) {
	widget := map[string]interface{}{
		"group_definition": []interface{}{
			map[string]interface{}{
				"layout_type": "ordered",
				"widget": []interface{}{
					map[string]interface{}{
						"query_value_definition": []interface{}{
							map[string]interface{}{
								"title": "CPU",
							},
						},
						"widget_layout": []interface{}{
							map[string]interface{}{
								"x":      0,
								"y":      0,
								"width":  4,
								"height": 2,
							},
						},
					},
				},
			},
		},
	}

	result := BuildWidgetEngineJSONFromMap(widget)
	groupDefinition := result["definition"].(map[string]interface{})
	nestedWidgets := groupDefinition["widgets"].([]interface{})
	nestedWidget := nestedWidgets[0].(map[string]interface{})
	layout := nestedWidget["layout"].(map[string]interface{})

	for _, key := range []string{"x", "y", "width", "height"} {
		if _, ok := layout[key]; !ok {
			t.Fatalf("expected layout to include %q, got: %#v", key, layout)
		}
	}
	if layout["x"] != 0 {
		t.Fatalf("expected x=0, got %#v", layout["x"])
	}
	if layout["y"] != 0 {
		t.Fatalf("expected y=0, got %#v", layout["y"])
	}
}
