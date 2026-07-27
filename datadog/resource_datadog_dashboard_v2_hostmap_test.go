package datadog

import (
	"reflect"
	"testing"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/dashboardmapping"
)

type rawConfigAtFunc func(cty.Path) (cty.Value, diag.Diagnostics)

func (f rawConfigAtFunc) GetRawConfigAt(path cty.Path) (cty.Value, diag.Diagnostics) {
	return f(path)
}

func TestRestoreHostmapExplicitZeroStyleBounds(t *testing.T) {
	resource := resourceDatadogDashboardV2()
	data := schema.TestResourceDataRaw(t, resource.SchemaFunc(), map[string]interface{}{
		"title":       "HostMap zero bounds",
		"layout_type": "ordered",
		"widget": []interface{}{map[string]interface{}{
			"hostmap_definition": []interface{}{map[string]interface{}{
				"request": []interface{}{map[string]interface{}{
					"request_type": "infrastructure_hostmap",
					"style": []interface{}{map[string]interface{}{
						"fill_min": 0.0,
						"fill_max": 100.0,
					}},
					"child": []interface{}{map[string]interface{}{
						"request_type": "infrastructure_hostmap",
						"style": []interface{}{map[string]interface{}{
							"fill_min": -100.0,
							"fill_max": 0.0,
						}},
					}},
				}},
			}},
		}},
	})

	widgets := data.Get("widget")
	rawWidget := widgets.([]interface{})[0].(map[string]interface{})
	rawDefinition := rawWidget["hostmap_definition"].([]interface{})[0].(map[string]interface{})
	rawRequest := rawDefinition["request"].([]interface{})[0].(map[string]interface{})
	rawStyle := rawRequest["style"].([]interface{})[0].(map[string]interface{})
	rawChild := rawRequest["child"].([]interface{})[0].(map[string]interface{})
	rawChildStyle := rawChild["style"].([]interface{})[0].(map[string]interface{})
	fillMinPath := cty.GetAttrPath("widget").
		IndexInt(0).
		GetAttr("hostmap_definition").
		IndexInt(0).
		GetAttr("request").
		IndexInt(0).
		GetAttr("style").
		IndexInt(0).
		GetAttr("fill_min")
	childFillMaxPath := cty.GetAttrPath("widget").
		IndexInt(0).
		GetAttr("hostmap_definition").
		IndexInt(0).
		GetAttr("request").
		IndexInt(0).
		GetAttr("child").
		IndexInt(0).
		GetAttr("style").
		IndexInt(0).
		GetAttr("fill_max")

	// ResourceData's decoded nested maps omit optional zero-valued attributes in
	// real Terraform operations. TestResourceDataRaw retains them, so remove
	// the keys and provide the raw configuration through the same interface used
	// by ResourceData in production.
	delete(rawStyle, "fill_min")
	delete(rawChildStyle, "fill_max")
	rawConfig := rawConfigAtFunc(func(path cty.Path) (cty.Value, diag.Diagnostics) {
		if reflect.DeepEqual(path, fillMinPath) || reflect.DeepEqual(path, childFillMaxPath) {
			return cty.NumberFloatVal(0), nil
		}
		return cty.NullVal(cty.Number), nil
	})

	restoreHostmapExplicitZeroStyleBounds(rawConfig, widgets)
	widget := widgets.([]interface{})[0].(map[string]interface{})
	definition := widget["hostmap_definition"].([]interface{})[0].(map[string]interface{})
	request := definition["request"].([]interface{})[0].(map[string]interface{})
	style := request["style"].([]interface{})[0].(map[string]interface{})
	if value, present := style["fill_min"]; !present || value != 0.0 {
		t.Fatalf("explicit fill_min = 0 was not preserved: %#v", style)
	}

	child := request["child"].([]interface{})[0].(map[string]interface{})
	childStyle := child["style"].([]interface{})[0].(map[string]interface{})
	if value, present := childStyle["fill_max"]; !present || value != 0.0 {
		t.Fatalf("explicit child fill_max = 0 was not preserved: %#v", childStyle)
	}

	built := dashboardmapping.BuildWidgetEngineJSONFromMap(widget)
	builtDefinition := built["definition"].(map[string]interface{})
	builtRequest := builtDefinition["requests"].(map[string]interface{})
	builtStyle := builtRequest["style"].(map[string]interface{})
	if value, present := builtStyle["fill_min"]; !present || value != 0.0 {
		t.Fatalf("explicit fill_min = 0 was not serialized: %#v", builtStyle)
	}
	builtChild := builtRequest["child"].(map[string]interface{})
	builtChildStyle := builtChild["style"].(map[string]interface{})
	if value, present := builtChildStyle["fill_max"]; !present || value != 0.0 {
		t.Fatalf("explicit child fill_max = 0 was not serialized: %#v", builtChildStyle)
	}
}

func TestRestoreHostmapExplicitZeroStyleBoundsDoesNotInventUnsetBounds(t *testing.T) {
	resource := resourceDatadogDashboardV2()
	data := schema.TestResourceDataRaw(t, resource.SchemaFunc(), map[string]interface{}{
		"title":       "HostMap automatic bounds",
		"layout_type": "ordered",
		"widget": []interface{}{map[string]interface{}{
			"hostmap_definition": []interface{}{map[string]interface{}{
				"request": []interface{}{map[string]interface{}{
					"request_type": "infrastructure_hostmap",
					"style": []interface{}{map[string]interface{}{
						"palette": "hostmap_blues",
					}},
					"child": []interface{}{map[string]interface{}{
						"request_type": "infrastructure_hostmap",
						"style": []interface{}{map[string]interface{}{
							"palette": "hostmap_blues",
						}},
					}},
				}},
			}},
		}},
	})

	widgets := data.Get("widget")
	rawWidget := widgets.([]interface{})[0].(map[string]interface{})
	rawDefinition := rawWidget["hostmap_definition"].([]interface{})[0].(map[string]interface{})
	rawRequest := rawDefinition["request"].([]interface{})[0].(map[string]interface{})
	rawStyle := rawRequest["style"].([]interface{})[0].(map[string]interface{})
	rawChild := rawRequest["child"].([]interface{})[0].(map[string]interface{})
	rawChildStyle := rawChild["style"].([]interface{})[0].(map[string]interface{})
	delete(rawStyle, "fill_min")
	delete(rawStyle, "fill_max")
	delete(rawChildStyle, "fill_min")
	delete(rawChildStyle, "fill_max")

	rawConfig := rawConfigAtFunc(func(cty.Path) (cty.Value, diag.Diagnostics) {
		return cty.NullVal(cty.Number), nil
	})
	restoreHostmapExplicitZeroStyleBounds(rawConfig, widgets)
	widget := widgets.([]interface{})[0].(map[string]interface{})
	definition := widget["hostmap_definition"].([]interface{})[0].(map[string]interface{})
	request := definition["request"].([]interface{})[0].(map[string]interface{})
	style := request["style"].([]interface{})[0].(map[string]interface{})
	if _, present := style["fill_min"]; present {
		t.Fatalf("unset fill_min was restored as an explicit zero: %#v", style)
	}
	if _, present := style["fill_max"]; present {
		t.Fatalf("unset fill_max was restored as an explicit zero: %#v", style)
	}

	child := request["child"].([]interface{})[0].(map[string]interface{})
	childStyle := child["style"].([]interface{})[0].(map[string]interface{})
	if _, present := childStyle["fill_min"]; present {
		t.Fatalf("unset child fill_min was restored as an explicit zero: %#v", childStyle)
	}
	if _, present := childStyle["fill_max"]; present {
		t.Fatalf("unset child fill_max was restored as an explicit zero: %#v", childStyle)
	}

	built := dashboardmapping.BuildWidgetEngineJSONFromMap(widget)
	builtDefinition := built["definition"].(map[string]interface{})
	builtRequest := builtDefinition["requests"].(map[string]interface{})
	builtStyle := builtRequest["style"].(map[string]interface{})
	if _, present := builtStyle["fill_min"]; present {
		t.Fatalf("unset fill_min was serialized as an explicit zero: %#v", builtStyle)
	}
	if _, present := builtStyle["fill_max"]; present {
		t.Fatalf("unset fill_max was serialized as an explicit zero: %#v", builtStyle)
	}
	builtChild := builtRequest["child"].(map[string]interface{})
	builtChildStyle := builtChild["style"].(map[string]interface{})
	if _, present := builtChildStyle["fill_min"]; present {
		t.Fatalf("unset child fill_min was serialized as an explicit zero: %#v", builtChildStyle)
	}
	if _, present := builtChildStyle["fill_max"]; present {
		t.Fatalf("unset child fill_max was serialized as an explicit zero: %#v", builtChildStyle)
	}
}
