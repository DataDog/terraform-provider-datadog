package dashboardmapping

import "testing"

func TestProductAnalyticsFormulaQueriesRoundTrip(t *testing.T) {
	tests := []struct {
		name       string
		hclKey     string
		dataSource string
		query      map[string]interface{}
		assert     func(*testing.T, map[string]interface{}, map[string]interface{})
	}{
		{
			name:       "extended",
			hclKey:     "product_analytics_extended_query",
			dataSource: "product_analytics_extended",
			query: map[string]interface{}{
				"product_analytics_extended_query": []interface{}{map[string]interface{}{
					"data_source": "product_analytics_extended",
					"name":        "extended_query",
					"query": []interface{}{map[string]interface{}{
						"data_source": "product_analytics",
						"search": []interface{}{map[string]interface{}{
							"query": "@type:view @view.name:/checkout",
						}},
					}},
					"compute": []interface{}{map[string]interface{}{
						"aggregation": "count",
						"name":        "views",
						"rollup": []interface{}{map[string]interface{}{
							"type":      "day",
							"timezone":  "UTC",
							"quantity":  1,
							"alignment": "1am",
						}},
					}},
					"group_by": []interface{}{map[string]interface{}{
						"facet":                  "@geo.country",
						"limit":                  10,
						"should_exclude_missing": true,
						"sort": []interface{}{map[string]interface{}{
							"aggregation": "count",
							"order":       "desc",
						}},
					}},
					"audience_filters": []interface{}{map[string]interface{}{
						"user": []interface{}{map[string]interface{}{
							"name":  "buyers",
							"query": "@usr.plan:paid",
						}},
					}},
					"indexes": []interface{}{"*"},
				}},
			},
			assert: func(t *testing.T, built, flat map[string]interface{}) {
				compute := built["compute"].(map[string]interface{})
				if compute["rollup"].(map[string]interface{})["type"] != "day" {
					t.Fatalf("extended rollup was not serialized: %#v", built)
				}
				if _, ok := flat["audience_filters"]; !ok {
					t.Fatalf("extended audience filters were not restored: %#v", flat)
				}
			},
		},
		{
			name:       "user journey",
			hclKey:     "user_journey_query",
			dataSource: "product_analytics_journey",
			query: map[string]interface{}{
				"user_journey_query": []interface{}{map[string]interface{}{
					"data_source": "product_analytics_journey",
					"name":        "journey_query",
					"search": []interface{}{map[string]interface{}{
						"node_objects": `{"node_0":{"data_source":"product_analytics","search":{"query":"@type:view @view.name:/home"}},"node_1":{"data_source":"product_analytics","search":{"query":"@type:action @action.name:checkout"}}}`,
						"expression":   "node_0 -> node_1",
						"step_aliases": `{"node_0":"Home","node_1":"Checkout"}`,
						"join_keys": []interface{}{map[string]interface{}{
							"primary":   "@session.id",
							"secondary": []interface{}{"@usr.id"},
						}},
						"filters": []interface{}{map[string]interface{}{
							"string_filter": "@usr.plan:paid",
							"graph_filter": []interface{}{map[string]interface{}{
								"name":     "count",
								"operator": "gt",
								"value":    1,
								"target": []interface{}{map[string]interface{}{
									"type":  "step",
									"value": "node_1",
								}},
							}},
						}},
					}},
					"compute": []interface{}{map[string]interface{}{
						"aggregation": "count",
						"metric":      "__dd.conversion_rate",
						"interval":    60000.0,
						"target": []interface{}{map[string]interface{}{
							"type":  "step",
							"value": "node_1",
						}},
					}},
					"group_by": []interface{}{map[string]interface{}{
						"facet": "@usr.email",
						"target": []interface{}{map[string]interface{}{
							"type":  "step",
							"value": "node_0",
						}},
					}},
				}},
			},
			assert: func(t *testing.T, built, flat map[string]interface{}) {
				nodes := built["search"].(map[string]interface{})["node_objects"].(map[string]interface{})
				if _, ok := nodes["node_1"]; !ok {
					t.Fatalf("journey node objects were not serialized: %#v", built)
				}
				if flatSearch := flat["search"].([]interface{})[0].(map[string]interface{}); flatSearch["node_objects"] == "" {
					t.Fatalf("journey node objects were not restored: %#v", flat)
				}
			},
		},
		{
			name:       "retention",
			hclKey:     "retention_query",
			dataSource: "product_analytics_retention",
			query: map[string]interface{}{
				"retention_query": []interface{}{map[string]interface{}{
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
						"return_criteria": []interface{}{map[string]interface{}{
							"base_query": []interface{}{map[string]interface{}{
								"data_source": "product_analytics",
								"search": []interface{}{map[string]interface{}{
									"query": "@type:action @action.name:purchase",
								}},
							}},
							"time_interval": []interface{}{map[string]interface{}{
								"type":  "fixed",
								"value": 1.0,
								"unit":  "week",
							}},
						}},
					}},
					"compute": []interface{}{map[string]interface{}{
						"aggregation": "count",
						"metric":      "__dd.retention_rate",
					}},
					"group_by": []interface{}{map[string]interface{}{
						"facet":  "@geo.country",
						"target": "cohort",
						"sort": []interface{}{map[string]interface{}{
							"order": "desc",
						}},
					}},
				}},
			},
			assert: func(t *testing.T, built, flat map[string]interface{}) {
				search := built["search"].(map[string]interface{})
				if search["retention_entity"] != "@usr.id" {
					t.Fatalf("retention search was not serialized: %#v", built)
				}
				if flat["compute"].([]interface{})[0].(map[string]interface{})["metric"] != "__dd.retention_rate" {
					t.Fatalf("retention compute was not restored: %#v", flat)
				}
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			built := buildQueryFromMapAttrs(test.query)
			if built["data_source"] != test.dataSource {
				t.Fatalf("query was not serialized: %#v", built)
			}
			flattened := flattenFormulaQueryJSON(built)
			block := flattened[test.hclKey].([]interface{})[0].(map[string]interface{})
			if block["data_source"] != test.dataSource {
				t.Fatalf("query was not restored: %#v", flattened)
			}
			test.assert(t, built, block)
		})
	}
}
