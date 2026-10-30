package dashboardmapping

import (
	"reflect"
	"testing"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
)

func heatgridWidgetForTest(color map[string]interface{}, sortBy map[string]interface{}, labelColumn map[string]interface{}) map[string]interface{} {
	return map[string]interface{}{
		"heatgrid_definition": []interface{}{map[string]interface{}{
			"request": []interface{}{map[string]interface{}{
				"query": []interface{}{map[string]interface{}{
					"metric_query": []interface{}{map[string]interface{}{
						"data_source": "metrics",
						"name":        "query1",
						"query":       "avg:system.cpu.user{*} by {host}",
					}},
				}},
				"formula": []interface{}{map[string]interface{}{
					"formula_expression": "query1",
				}},
			}},
			"sort": []interface{}{map[string]interface{}{
				"nesting_display": "flat",
				"sort_by":         []interface{}{sortBy},
			}},
			"color":        []interface{}{color},
			"label_column": []interface{}{labelColumn},
		}},
	}
}

func TestHeatgridColorBinPreservesExplicitZeroLowerBound(t *testing.T) {
	widget := heatgridWidgetForTest(
		map[string]interface{}{
			"mode": "discrete", "source": "custom",
			"bin": []interface{}{
				map[string]interface{}{"color": `"#00FF00"`},
				map[string]interface{}{"color": `"#FF0000"`},
			},
		},
		map[string]interface{}{"label": []interface{}{map[string]interface{}{"order": "asc"}}},
		map[string]interface{}{},
	)
	widgetPath := cty.GetAttrPath("widget").IndexInt(0)
	lowerBoundPath := widgetPath.GetAttr("heatgrid_definition").IndexInt(0).
		GetAttr("color").IndexInt(0).GetAttr("bin").IndexInt(1).GetAttr("lower_bound")
	rawConfig := rawConfigAtFunc(func(path cty.Path) (cty.Value, diag.Diagnostics) {
		if reflect.DeepEqual(path, lowerBoundPath) {
			return cty.NumberIntVal(0), nil
		}
		return cty.NullVal(cty.Number), nil
	})
	built := BuildWidgetEngineJSONFromMapWithRawConfig(widget, rawConfig, widgetPath)
	bins := built["definition"].(map[string]interface{})["color"].(map[string]interface{})["bins"].([]interface{})
	if _, present := bins[0].(map[string]interface{})["lower_bound"]; present {
		t.Fatalf("first bin must not have a lower bound: %#v", bins[0])
	}
	if value, present := bins[1].(map[string]interface{})["lower_bound"]; !present || value != float64(0) {
		t.Fatalf("explicit zero lower bound was not serialized: %#v", bins[1])
	}
}

func TestHeatgridGradientPreservesExplicitZeroPosition(t *testing.T) {
	widget := heatgridWidgetForTest(
		map[string]interface{}{
			"mode": "gradient", "source": "custom",
			"stop": []interface{}{
				map[string]interface{}{"color": `"#FFFFFF"`},
				map[string]interface{}{"position": 100, "color": `"#000000"`},
			},
		},
		map[string]interface{}{"value": []interface{}{map[string]interface{}{"order": "desc", "aggregation": "avg"}}},
		map[string]interface{}{},
	)
	widgetPath := cty.GetAttrPath("widget").IndexInt(0)
	positionPath := widgetPath.GetAttr("heatgrid_definition").IndexInt(0).
		GetAttr("color").IndexInt(0).GetAttr("stop").IndexInt(0).GetAttr("position")
	rawConfig := rawConfigAtFunc(func(path cty.Path) (cty.Value, diag.Diagnostics) {
		if reflect.DeepEqual(path, positionPath) {
			return cty.NumberIntVal(0), nil
		}
		return cty.NullVal(cty.Number), nil
	})
	built := BuildWidgetEngineJSONFromMapWithRawConfig(widget, rawConfig, widgetPath)
	stops := built["definition"].(map[string]interface{})["color"].(map[string]interface{})["stops"].([]interface{})
	if value, present := stops[0].(map[string]interface{})["position"]; !present || value != 0 {
		t.Fatalf("explicit zero stop position was not serialized: %#v", stops[0])
	}
}

func TestHeatgridWidgetRoundTrips(t *testing.T) {
	cases := []struct {
		name        string
		color       map[string]interface{}
		sortBy      map[string]interface{}
		labelColumn map[string]interface{}
	}{
		{
			name: "preset gradient and value sort",
			color: map[string]interface{}{
				"mode": "gradient", "source": "preset", "preset_name": "classic",
			},
			sortBy: map[string]interface{}{
				"value": []interface{}{map[string]interface{}{"order": "desc", "aggregation": "avg"}},
			},
			labelColumn: map[string]interface{}{"width": "m"},
		},
		{
			name: "custom discrete bins and replacement rules",
			color: map[string]interface{}{
				"mode": "discrete", "source": "custom",
				"bin": []interface{}{
					map[string]interface{}{"color": `"#00FF00"`, "label": "Normal"},
					map[string]interface{}{"color": `"#FF0000"`, "lower_bound": float64(0), "label": "High"},
				},
			},
			sortBy: map[string]interface{}{
				"label": []interface{}{map[string]interface{}{"order": "asc"}},
			},
			labelColumn: map[string]interface{}{
				"text_format": []interface{}{
					map[string]interface{}{
						"match": []interface{}{map[string]interface{}{"type": "starts_with", "value": "api-"}},
						"replace": []interface{}{map[string]interface{}{
							"substring": []interface{}{map[string]interface{}{"substring": "api-", "with": ""}},
						}},
					},
					map[string]interface{}{
						"match": []interface{}{map[string]interface{}{"type": "is", "value": "unknown"}},
						"replace": []interface{}{map[string]interface{}{
							"all": []interface{}{map[string]interface{}{"with": "Other"}},
						}},
					},
				},
			},
		},
		{
			name: "preset discrete colors",
			color: map[string]interface{}{
				"mode": "discrete", "source": "preset", "preset_name": "classic",
			},
			sortBy: map[string]interface{}{
				"label": []interface{}{map[string]interface{}{"order": "desc"}},
			},
			labelColumn: map[string]interface{}{},
		},
		{
			name: "custom gradient theme colors",
			color: map[string]interface{}{
				"mode": "gradient", "source": "custom",
				"stop": []interface{}{
					map[string]interface{}{"position": 0, "color": `["#FFFFFF","#000000"]`},
					map[string]interface{}{"position": 100, "color": `"#FF0000"`},
				},
			},
			sortBy: map[string]interface{}{
				"value": []interface{}{map[string]interface{}{"order": "desc", "aggregation": "sum"}},
			},
			labelColumn: map[string]interface{}{},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			built := BuildWidgetEngineJSONFromMap(heatgridWidgetForTest(tc.color, tc.sortBy, tc.labelColumn))
			definition := built["definition"].(map[string]interface{})
			request := definition["requests"].([]interface{})[0].(map[string]interface{})
			if request["response_format"] != "timeseries" {
				t.Fatalf("Heatgrid request must be timeseries: %#v", request)
			}
			if _, configured := tc.labelColumn["width"]; !configured {
				if labelColumn, ok := definition["label_column"].(map[string]interface{}); ok {
					if _, emitted := labelColumn["width"]; emitted {
						t.Fatalf("Omitted label width was emitted: %#v", labelColumn)
					}
				}
			}
			flattened, drops := FlattenWidgetEngineJSON(built)
			if len(drops) != 0 {
				t.Fatalf("Heatgrid fields were dropped during flatten: %v", drops)
			}
			rebuilt := BuildWidgetEngineJSONFromMap(flattened)
			if !reflect.DeepEqual(rebuilt, built) {
				t.Fatalf("Heatgrid did not round trip:\nfirst:  %#v\nsecond: %#v", built, rebuilt)
			}
		})
	}
}
