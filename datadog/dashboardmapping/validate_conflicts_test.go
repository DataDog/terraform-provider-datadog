package dashboardmapping

import (
	"strings"
	"testing"
)

func TestValidateWidgetConflicts_QWithQueryAndFormula(t *testing.T) {
	// Simulate a timeseries widget request with both "q" and "query"/"formula" set
	data := map[string]interface{}{
		"widget": []interface{}{
			map[string]interface{}{
				"timeseries_definition": []interface{}{
					map[string]interface{}{
						"request": []interface{}{
							map[string]interface{}{
								"q": "avg:system.cpu.user{*}",
								"query": []interface{}{
									map[string]interface{}{
										"metric_query": []interface{}{
											map[string]interface{}{
												"name":  "query1",
												"query": "avg:system.cpu.user{*}",
											},
										},
									},
								},
								"formula": []interface{}{
									map[string]interface{}{
										"formula_expression": "query1",
									},
								},
							},
						},
					},
				},
			},
		},
	}

	errs := ValidateWidgetConflicts(data)
	if len(errs) == 0 {
		t.Fatal("expected validation errors for q + query/formula conflict, got none")
	}

	// Should mention both conflicts: q vs query, q vs formula
	combined := strings.Join(errs, "\n")
	if !strings.Contains(combined, `"q" conflicts with "query"`) {
		t.Errorf("expected error about q conflicts with query, got: %s", combined)
	}
	if !strings.Contains(combined, `"q" conflicts with "formula"`) {
		t.Errorf("expected error about q conflicts with formula, got: %s", combined)
	}
}

func TestValidateWidgetConflicts_QOnly(t *testing.T) {
	// Legacy q-only request — should pass
	data := map[string]interface{}{
		"widget": []interface{}{
			map[string]interface{}{
				"query_value_definition": []interface{}{
					map[string]interface{}{
						"request": []interface{}{
							map[string]interface{}{
								"q": "avg:system.cpu.user{*}",
							},
						},
					},
				},
			},
		},
	}

	errs := ValidateWidgetConflicts(data)
	if len(errs) != 0 {
		t.Errorf("expected no errors for q-only request, got: %v", errs)
	}
}

func TestValidateWidgetConflicts_FormulaOnly(t *testing.T) {
	// Formula-only request — should pass
	data := map[string]interface{}{
		"widget": []interface{}{
			map[string]interface{}{
				"query_value_definition": []interface{}{
					map[string]interface{}{
						"request": []interface{}{
							map[string]interface{}{
								"q": "",
								"query": []interface{}{
									map[string]interface{}{
										"metric_query": []interface{}{
											map[string]interface{}{
												"name":  "query1",
												"query": "avg:system.cpu.user{*}",
											},
										},
									},
								},
								"formula": []interface{}{
									map[string]interface{}{
										"formula_expression": "query1",
									},
								},
							},
						},
					},
				},
			},
		},
	}

	errs := ValidateWidgetConflicts(data)
	if len(errs) != 0 {
		t.Errorf("expected no errors for formula-only request, got: %v", errs)
	}
}

func TestValidateWidgetConflicts_GroupWidgetNested(t *testing.T) {
	// Group widget with nested widget that has a conflict
	data := map[string]interface{}{
		"widget": []interface{}{
			map[string]interface{}{
				"group_definition": []interface{}{
					map[string]interface{}{
						"widget": []interface{}{
							map[string]interface{}{
								"toplist_definition": []interface{}{
									map[string]interface{}{
										"request": []interface{}{
											map[string]interface{}{
												"q": "avg:system.cpu.user{*}",
												"query": []interface{}{
													map[string]interface{}{
														"metric_query": []interface{}{
															map[string]interface{}{
																"name":  "q1",
																"query": "avg:system.cpu.user{*}",
															},
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	errs := ValidateWidgetConflicts(data)
	if len(errs) == 0 {
		t.Fatal("expected validation errors for nested group widget conflict, got none")
	}
	combined := strings.Join(errs, "\n")
	if !strings.Contains(combined, "group_definition") {
		t.Errorf("expected error path to include group_definition, got: %s", combined)
	}
}
