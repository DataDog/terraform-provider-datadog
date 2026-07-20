package dashboardmapping

import "testing"

func TestHostmapInfrastructureRequestRoundTrip(t *testing.T) {
	metricQuery := map[string]interface{}{
		"metric_query": []interface{}{map[string]interface{}{
			"data_source": "metrics",
			"name":        "query1",
			"query":       "avg:system.cpu.user{*} by {host}",
		}},
	}
	enrichment := map[string]interface{}{
		"response_format": "scalar",
		"query":           []interface{}{metricQuery},
		"formula": []interface{}{map[string]interface{}{
			"formula_expression": "query1",
			"alias":              "CPU usage",
			"dimension":          "fill",
		}},
	}
	widget := map[string]interface{}{
		"hostmap_definition": []interface{}{map[string]interface{}{
			"title": "Infrastructure host map",
			"request": []interface{}{map[string]interface{}{
				"request_type": "infrastructure_hostmap",
				"node_type":    "host",
				"filter":       "env:prod",
				"group_by": []interface{}{map[string]interface{}{
					"column": "tags",
					"key":    "service",
				}},
				"enrichment": []interface{}{enrichment},
				"style": []interface{}{map[string]interface{}{
					"palette":      "hostmap_blues",
					"palette_flip": true,
					"fill_min":     0.0,
					"fill_max":     100.0,
				}},
				"conditional_formats": []interface{}{map[string]interface{}{
					"comparator": ">",
					"value":      80.0,
					"palette":    "white_on_red",
					"hide_value": false,
				}},
				"no_group_hosts":  true,
				"no_metric_hosts": true,
				"child": []interface{}{map[string]interface{}{
					"request_type": "infrastructure_hostmap",
					"node_type":    "container",
					"filter":       "kube_namespace:store",
					"enrichment":   []interface{}{enrichment},
				}},
			}},
		}},
	}

	built := BuildWidgetEngineJSONFromMap(widget)
	definition := built["definition"].(map[string]interface{})
	request := definition["requests"].(map[string]interface{})
	if request["request_type"] != "infrastructure_hostmap" || request["node_type"] != "host" {
		t.Fatalf("infrastructure discriminator was not serialized: %#v", request)
	}
	enrichments := request["enrichments"].([]interface{})
	builtEnrichment := enrichments[0].(map[string]interface{})
	builtQuery := builtEnrichment["queries"].([]interface{})[0].(map[string]interface{})
	if builtQuery["data_source"] != "metrics" {
		t.Fatalf("enrichment query was not serialized: %#v", builtEnrichment)
	}
	builtFormula := builtEnrichment["formulas"].([]interface{})[0].(map[string]interface{})
	if builtFormula["formula"] != "query1" || builtFormula["dimension"] != "fill" {
		t.Fatalf("dimension-aware formula was not serialized: %#v", builtFormula)
	}
	child := request["child"].(map[string]interface{})
	if child["node_type"] != "container" || len(child["enrichments"].([]interface{})) != 1 {
		t.Fatalf("hostmap child request was not serialized: %#v", child)
	}

	flattened, _ := FlattenWidgetEngineJSON(built)
	flatDefinition := flattened["hostmap_definition"].([]interface{})[0].(map[string]interface{})
	flatRequest := flatDefinition["request"].([]interface{})[0].(map[string]interface{})
	if flatRequest["request_type"] != "infrastructure_hostmap" || flatRequest["filter"] != "env:prod" {
		t.Fatalf("infrastructure request was not restored: %#v", flatRequest)
	}
	flatEnrichment := flatRequest["enrichment"].([]interface{})[0].(map[string]interface{})
	flatFormula := flatEnrichment["formula"].([]interface{})[0].(map[string]interface{})
	if flatFormula["formula_expression"] != "query1" || flatFormula["dimension"] != "fill" {
		t.Fatalf("infrastructure formula was not restored: %#v", flatFormula)
	}
	flatChild := flatRequest["child"].([]interface{})[0].(map[string]interface{})
	if flatChild["node_type"] != "container" {
		t.Fatalf("hostmap child request was not restored: %#v", flatChild)
	}
}

func TestHostmapLegacyRequestStillBuilds(t *testing.T) {
	widget := map[string]interface{}{
		"hostmap_definition": []interface{}{map[string]interface{}{
			"request": []interface{}{map[string]interface{}{
				"fill": []interface{}{map[string]interface{}{
					"q": "avg:system.cpu.user{*} by {host}",
				}},
			}},
		}},
	}

	built := BuildWidgetEngineJSONFromMap(widget)
	definition := built["definition"].(map[string]interface{})
	request := definition["requests"].(map[string]interface{})
	fill := request["fill"].(map[string]interface{})
	if fill["q"] != "avg:system.cpu.user{*} by {host}" {
		t.Fatalf("legacy hostmap request changed: %#v", request)
	}
}
