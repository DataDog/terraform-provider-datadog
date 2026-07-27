package datadog

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestCollectDashboardDataPreservesHostmapExplicitZeroStyleBounds(t *testing.T) {
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

	collected := collectDashboardData(data)
	widget := collected["widget"].([]interface{})[0].(map[string]interface{})
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
}
