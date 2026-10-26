package dashboardmapping

import (
	"reflect"
	"testing"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
)

type rawConfigAtFunc func(cty.Path) (cty.Value, diag.Diagnostics)

func (f rawConfigAtFunc) GetRawConfigAt(path cty.Path) (cty.Value, diag.Diagnostics) {
	return f(path)
}

func TestBuildWidgetEngineJSONFromMapWithRawConfigPreservesExplicitZero(t *testing.T) {
	widget := hostmapInfrastructureWidgetWithStyleBounds()
	request := widget["hostmap_definition"].([]interface{})[0].(map[string]interface{})["request"].([]interface{})[0].(map[string]interface{})
	style := request["style"].([]interface{})[0].(map[string]interface{})
	childStyle := request["child"].([]interface{})[0].(map[string]interface{})["style"].([]interface{})[0].(map[string]interface{})

	// Match SDKv2's decoded nested maps, where explicitly configured zeroes are absent.
	delete(style, "fill_min")
	delete(childStyle, "fill_max")

	widgetPath := cty.GetAttrPath("widget").IndexInt(0)
	fillMinPath := widgetPath.GetAttr("hostmap_definition").IndexInt(0).
		GetAttr("request").IndexInt(0).GetAttr("style").IndexInt(0).GetAttr("fill_min")
	childFillMaxPath := widgetPath.GetAttr("hostmap_definition").IndexInt(0).
		GetAttr("request").IndexInt(0).GetAttr("child").IndexInt(0).
		GetAttr("style").IndexInt(0).GetAttr("fill_max")
	rawConfig := rawConfigAtFunc(func(path cty.Path) (cty.Value, diag.Diagnostics) {
		if reflect.DeepEqual(path, fillMinPath) || reflect.DeepEqual(path, childFillMaxPath) {
			return cty.NumberIntVal(0), nil
		}
		return cty.NullVal(cty.Number), nil
	})

	built := BuildWidgetEngineJSONFromMapWithRawConfig(widget, rawConfig, widgetPath)
	builtRequest := built["definition"].(map[string]interface{})["requests"].(map[string]interface{})
	builtStyle := builtRequest["style"].(map[string]interface{})
	if value, present := builtStyle["fill_min"]; !present || value != 0.0 {
		t.Fatalf("explicit fill_min = 0 was not serialized: %#v", builtStyle)
	}
	builtChildStyle := builtRequest["child"].(map[string]interface{})["style"].(map[string]interface{})
	if value, present := builtChildStyle["fill_max"]; !present || value != 0.0 {
		t.Fatalf("explicit child fill_max = 0 was not serialized: %#v", builtChildStyle)
	}
}

func TestBuildWidgetEngineJSONFromMapWithRawConfigOmitsUnsetZero(t *testing.T) {
	widget := hostmapInfrastructureWidgetWithStyleBounds()
	request := widget["hostmap_definition"].([]interface{})[0].(map[string]interface{})["request"].([]interface{})[0].(map[string]interface{})
	style := request["style"].([]interface{})[0].(map[string]interface{})
	childStyle := request["child"].([]interface{})[0].(map[string]interface{})["style"].([]interface{})[0].(map[string]interface{})
	delete(style, "fill_min")
	delete(childStyle, "fill_max")

	rawConfig := rawConfigAtFunc(func(cty.Path) (cty.Value, diag.Diagnostics) {
		return cty.NullVal(cty.Number), nil
	})
	built := BuildWidgetEngineJSONFromMapWithRawConfig(widget, rawConfig, cty.GetAttrPath("widget").IndexInt(0))
	builtRequest := built["definition"].(map[string]interface{})["requests"].(map[string]interface{})
	builtStyle := builtRequest["style"].(map[string]interface{})
	if _, present := builtStyle["fill_min"]; present {
		t.Fatalf("unset fill_min was serialized: %#v", builtStyle)
	}
	builtChildStyle := builtRequest["child"].(map[string]interface{})["style"].(map[string]interface{})
	if _, present := builtChildStyle["fill_max"]; present {
		t.Fatalf("unset child fill_max was serialized: %#v", builtChildStyle)
	}
}

func hostmapInfrastructureWidgetWithStyleBounds() map[string]interface{} {
	return map[string]interface{}{
		"hostmap_definition": []interface{}{map[string]interface{}{
			"request": []interface{}{map[string]interface{}{
				"request_type": "infrastructure_hostmap",
				"node_type":    "host",
				"style": []interface{}{map[string]interface{}{
					"fill_min": 0.0,
					"fill_max": 100.0,
				}},
				"child": []interface{}{map[string]interface{}{
					"request_type": "infrastructure_hostmap",
					"node_type":    "container",
					"style": []interface{}{map[string]interface{}{
						"fill_min": -100.0,
						"fill_max": 0.0,
					}},
				}},
			}},
		}},
	}
}
