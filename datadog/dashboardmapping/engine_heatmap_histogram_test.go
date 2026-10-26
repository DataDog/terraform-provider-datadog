package dashboardmapping

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestHeatmapHistogramRequestRoundTrip(t *testing.T) {
	widget := map[string]interface{}{
		"heatmap_definition": []interface{}{map[string]interface{}{
			"request": []interface{}{map[string]interface{}{
				"histogram_request": []interface{}{map[string]interface{}{
					"histogram_query": []interface{}{map[string]interface{}{
						"metric_query": []interface{}{map[string]interface{}{
							"data_source": "metrics",
							"name":        "query1",
							"query":       "histogram:trace.servlet.request{*}",
						}},
					}},
					"style": []interface{}{map[string]interface{}{
						"palette": "dog_classic",
					}},
				}},
			}},
		}},
	}

	built := BuildWidgetEngineJSONFromMap(widget)
	definition := built["definition"].(map[string]interface{})
	request := definition["requests"].([]interface{})[0].(map[string]interface{})
	if request["request_type"] != "histogram" {
		t.Fatalf("histogram discriminator was not serialized: %#v", request)
	}
	query := request["query"].(map[string]interface{})
	if query["data_source"] != "metrics" || query["query"] != "histogram:trace.servlet.request{*}" {
		t.Fatalf("histogram query was not serialized: %#v", query)
	}
	style := request["style"].(map[string]interface{})
	if style["palette"] != "dog_classic" {
		t.Fatalf("heatmap style was not preserved: %#v", request)
	}

	flattened, _ := FlattenWidgetEngineJSON(built)
	flatDefinition := flattened["heatmap_definition"].([]interface{})[0].(map[string]interface{})
	flatRequest := flatDefinition["request"].([]interface{})[0].(map[string]interface{})
	histogramRequest := flatRequest["histogram_request"].([]interface{})[0].(map[string]interface{})
	histogramQuery := histogramRequest["histogram_query"].([]interface{})[0].(map[string]interface{})
	metricQuery := histogramQuery["metric_query"].([]interface{})[0].(map[string]interface{})
	if metricQuery["name"] != "query1" {
		t.Fatalf("histogram query was not restored: %#v", histogramQuery)
	}
}

func TestHeatmapFormulaRequestStillBuilds(t *testing.T) {
	widget := map[string]interface{}{
		"heatmap_definition": []interface{}{map[string]interface{}{
			"request": []interface{}{map[string]interface{}{
				"formula": []interface{}{map[string]interface{}{"formula_expression": "query1"}},
				"query": []interface{}{map[string]interface{}{
					"metric_query": []interface{}{map[string]interface{}{
						"data_source": "metrics",
						"name":        "query1",
						"query":       "avg:system.cpu.user{*}",
					}},
				}},
			}},
		}},
	}

	built := BuildWidgetEngineJSONFromMap(widget)
	definition := built["definition"].(map[string]interface{})
	request := definition["requests"].([]interface{})[0].(map[string]interface{})
	if request["response_format"] != "timeseries" || len(request["queries"].([]interface{})) != 1 {
		t.Fatalf("formula heatmap request changed: %#v", request)
	}
}

func TestHeatmapLegacyRequestStillFlattensFlat(t *testing.T) {
	widget := map[string]interface{}{
		"definition": map[string]interface{}{
			"type": "heatmap",
			"requests": []interface{}{map[string]interface{}{
				"q": "avg:system.cpu.user{*}",
				"style": map[string]interface{}{
					"palette": "dog_classic",
				},
			}},
		},
	}

	flattened, _ := FlattenWidgetEngineJSON(widget)
	definition := flattened["heatmap_definition"].([]interface{})[0].(map[string]interface{})
	request := definition["request"].([]interface{})[0].(map[string]interface{})
	if request["q"] != "avg:system.cpu.user{*}" {
		t.Fatalf("legacy request was not restored inline: %#v", request)
	}
	if _, ok := request["histogram_request"]; ok {
		t.Fatalf("legacy request was misclassified as histogram: %#v", request)
	}
}

func TestHeatmapHistogramRequestBuildsAlongsideFormulaRequest(t *testing.T) {
	widget := map[string]interface{}{
		"heatmap_definition": []interface{}{map[string]interface{}{
			"request": []interface{}{
				map[string]interface{}{
					"histogram_request": []interface{}{map[string]interface{}{
						"histogram_query": []interface{}{map[string]interface{}{
							"metric_query": []interface{}{map[string]interface{}{
								"data_source": "metrics",
								"name":        "histogram",
								"query":       "histogram:trace.servlet.request{*}",
							}},
						}},
					}},
				},
				map[string]interface{}{
					"formula": []interface{}{map[string]interface{}{"formula_expression": "query1"}},
					"query": []interface{}{map[string]interface{}{
						"metric_query": []interface{}{map[string]interface{}{
							"data_source": "metrics",
							"name":        "query1",
							"query":       "avg:system.cpu.user{*}",
						}},
					}},
				},
			},
		}},
	}

	built := BuildWidgetEngineJSONFromMap(widget)
	definition := built["definition"].(map[string]interface{})
	requests := definition["requests"].([]interface{})
	if len(requests) != 2 {
		t.Fatalf("expected both heatmap requests, got: %#v", requests)
	}
	histogramRequest := requests[0].(map[string]interface{})
	if histogramRequest["request_type"] != "histogram" {
		t.Fatalf("histogram request was not inferred from histogram_query: %#v", histogramRequest)
	}
	formulaRequest := requests[1].(map[string]interface{})
	if formulaRequest["response_format"] != "timeseries" {
		t.Fatalf("formula request was not preserved: %#v", formulaRequest)
	}

	flattened, _ := FlattenWidgetEngineJSON(built)
	flatDefinition := flattened["heatmap_definition"].([]interface{})[0].(map[string]interface{})
	flatRequests := flatDefinition["request"].([]interface{})
	if _, ok := flatRequests[0].(map[string]interface{})["histogram_request"]; !ok {
		t.Fatalf("histogram request variant was not restored: %#v", flatRequests[0])
	}
	if _, ok := flatRequests[1].(map[string]interface{})["query"]; !ok {
		t.Fatalf("formula request was not restored inline: %#v", flatRequests[1])
	}
}

func TestHeatmapRequestSchemaKeepsDiscriminatorInternal(t *testing.T) {
	var requestField FieldSpec
	for _, field := range HeatmapWidgetSpec.Fields {
		if field.HCLKey == "request" {
			requestField = field
			break
		}
	}
	requestSchema := FieldSpecToSDKv2(requestField)
	requestResource := requestSchema.Elem.(*schema.Resource)

	if _, ok := requestResource.Schema["request_type"]; ok {
		t.Fatal("request_type must remain an internal JSON discriminator")
	}
	if _, ok := requestResource.Schema["histogram_query"]; ok {
		t.Fatal("histogram_query must be scoped to the histogram_request variant")
	}
	for _, inlineField := range []string{"q", "query", "formula", "style"} {
		if _, ok := requestResource.Schema[inlineField]; !ok {
			t.Fatalf("existing inline request field %q was removed", inlineField)
		}
	}
	histogramSchema, ok := requestResource.Schema["histogram_request"]
	if !ok {
		t.Fatal("histogram_request variant is missing from the request schema")
	}
	histogramResource := histogramSchema.Elem.(*schema.Resource)
	histogramQuerySchema, ok := histogramResource.Schema["histogram_query"]
	if !ok {
		t.Fatal("histogram_query is missing from the histogram_request variant")
	}
	histogramQueryResource := histogramQuerySchema.Elem.(*schema.Resource)
	if !histogramQueryResource.Schema["metric_query"].Required {
		t.Fatal("the sole histogram query variant must remain required")
	}
}
