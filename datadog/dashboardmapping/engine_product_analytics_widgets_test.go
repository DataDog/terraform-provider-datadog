package dashboardmapping

import "testing"

func productAnalyticsRetentionQueryTestMap() map[string]interface{} {
	return map[string]interface{}{
		"data_source": "product_analytics_retention",
		"name":        "retention_query",
		"search": []interface{}{map[string]interface{}{
			"cohort_criteria": []interface{}{map[string]interface{}{
				"base_query": []interface{}{map[string]interface{}{
					"data_source": "product_analytics",
					"search": []interface{}{map[string]interface{}{
						"query": "@type:view @view.name:/signup",
					}},
				}},
				"time_interval": []interface{}{map[string]interface{}{
					"type": "calendar",
					"value": []interface{}{map[string]interface{}{
						"type":     "week",
						"timezone": "UTC",
					}},
				}},
			}},
			"retention_entity": "@usr.id",
			"return_condition": "conversion_on_or_after",
		}},
		"compute": []interface{}{map[string]interface{}{
			"aggregation": "count",
			"metric":      "__dd.retention_rate",
		}},
	}
}

func TestProductAnalyticsWidgetRoundTrips(t *testing.T) {
	t.Run("funnel discriminator preserves legacy widget", func(t *testing.T) {
		legacy := map[string]interface{}{
			"funnel_definition": []interface{}{map[string]interface{}{
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
		built := BuildWidgetEngineJSONFromMap(legacy)
		request := built["definition"].(map[string]interface{})["requests"].([]interface{})[0].(map[string]interface{})
		if request["request_type"] != "funnel" {
			t.Fatalf("legacy funnel request type was not injected: %#v", built)
		}
		flattened, _ := FlattenWidgetEngineJSON(built)
		if _, ok := flattened["funnel_definition"]; !ok {
			t.Fatalf("legacy funnel was dispatched to the wrong schema: %#v", flattened)
		}
	})

	t.Run("product analytics funnel", func(t *testing.T) {
		widget := map[string]interface{}{
			"product_analytics_funnel_definition": []interface{}{map[string]interface{}{
				"grouped_display": "stacked",
				"request": []interface{}{map[string]interface{}{
					"request_type": "user_journey_funnel",
					"query": []interface{}{map[string]interface{}{
						"data_source": "product_analytics_journey",
						"search": []interface{}{map[string]interface{}{
							"node_objects": `{"step1":{"data_source":"product_analytics","search":{"query":"@type:view @view.name:/home"}},"step2":{"data_source":"product_analytics","search":{"query":"@type:action @action.name:checkout"}}}`,
							"expression":   "step1 -> step2",
						}},
						"compute": []interface{}{map[string]interface{}{
							"aggregation": "count",
							"metric":      "__dd.conversion_rate",
						}},
						"group_by": []interface{}{map[string]interface{}{
							"facet": "@usr.email",
							"sort": []interface{}{map[string]interface{}{
								"aggregation": "count",
								"order":       "desc",
							}},
						}},
					}},
					"comparison_time": []interface{}{map[string]interface{}{
						"type": "previous_week",
					}},
					"comparison_segments": []interface{}{"paid"},
				}},
			}},
		}

		built := BuildWidgetEngineJSONFromMap(widget)
		definition := built["definition"].(map[string]interface{})
		request := definition["requests"].([]interface{})[0].(map[string]interface{})
		if request["request_type"] != "user_journey_funnel" {
			t.Fatalf("Product Analytics funnel request type was overwritten: %#v", built)
		}
		flattened, _ := FlattenWidgetEngineJSON(built)
		definitionState := flattened["product_analytics_funnel_definition"].([]interface{})[0].(map[string]interface{})
		flatRequest := definitionState["request"].([]interface{})[0].(map[string]interface{})
		if flatRequest["request_type"] != "user_journey_funnel" {
			t.Fatalf("Product Analytics funnel was not restored: %#v", flattened)
		}
	})

	t.Run("cohort", func(t *testing.T) {
		widget := map[string]interface{}{
			"cohort_definition": []interface{}{map[string]interface{}{
				"request": []interface{}{map[string]interface{}{
					"request_type": "retention_grid",
					"query":        []interface{}{productAnalyticsRetentionQueryTestMap()},
				}},
			}},
		}
		built := BuildWidgetEngineJSONFromMap(widget)
		if built["definition"].(map[string]interface{})["type"] != "cohort" {
			t.Fatalf("cohort widget was not serialized: %#v", built)
		}
		flattened, _ := FlattenWidgetEngineJSON(built)
		if _, ok := flattened["cohort_definition"]; !ok {
			t.Fatalf("cohort widget was not restored: %#v", flattened)
		}
	})

	t.Run("retention curve", func(t *testing.T) {
		widget := map[string]interface{}{
			"retention_curve_definition": []interface{}{map[string]interface{}{
				"request": []interface{}{map[string]interface{}{
					"request_type": "retention_curve",
					"query":        []interface{}{productAnalyticsRetentionQueryTestMap()},
					"style": []interface{}{map[string]interface{}{
						"palette": "dog_classic",
					}},
				}},
			}},
		}
		built := BuildWidgetEngineJSONFromMap(widget)
		flattened, _ := FlattenWidgetEngineJSON(built)
		definitionState := flattened["retention_curve_definition"].([]interface{})[0].(map[string]interface{})
		request := definitionState["request"].([]interface{})[0].(map[string]interface{})
		if request["style"].([]interface{})[0].(map[string]interface{})["palette"] != "dog_classic" {
			t.Fatalf("retention curve style was not restored: %#v", flattened)
		}
	})
}
