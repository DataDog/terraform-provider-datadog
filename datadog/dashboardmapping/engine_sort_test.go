package dashboardmapping

import "testing"

// TestBuildWidgetSortByJSONFromMap_PreservesZeroFormulaIndex guards against a
// regression where formula_sort.index == 0 (the first/only formula, the most
// common case) was dropped from the built JSON because it was mistaken for
// "unset" rather than a valid required value.
func TestBuildWidgetSortByJSONFromMap_PreservesZeroFormulaIndex(t *testing.T) {
	sortMap := map[string]interface{}{
		"count": 10,
		"order_by": []interface{}{
			map[string]interface{}{
				"formula_sort": []interface{}{
					map[string]interface{}{
						"index": 0,
						"order": "desc",
					},
				},
			},
		},
	}

	result := buildWidgetSortByJSONFromMap(sortMap)

	orderBy, ok := result["order_by"].([]interface{})
	if !ok || len(orderBy) != 1 {
		t.Fatalf("expected one order_by entry, got: %#v", result["order_by"])
	}
	entry, ok := orderBy[0].(map[string]interface{})
	if !ok {
		t.Fatalf("expected order_by entry to be a map, got: %T", orderBy[0])
	}
	if entry["type"] != "formula" {
		t.Errorf("expected type=formula, got %v", entry["type"])
	}
	if idx, ok := entry["index"]; !ok || idx != 0 {
		t.Errorf("expected index=0 to be present, got %#v (present: %v)", idx, ok)
	}
	if entry["order"] != "desc" {
		t.Errorf("expected order=desc, got %v", entry["order"])
	}
}

// TestBuildWidgetSortByJSONFromMap_PreservesNonZeroFormulaIndex is the
// non-zero counterpart, verifying the fix didn't change behavior for indexes
// that were already serialized correctly.
func TestBuildWidgetSortByJSONFromMap_PreservesNonZeroFormulaIndex(t *testing.T) {
	sortMap := map[string]interface{}{
		"order_by": []interface{}{
			map[string]interface{}{
				"formula_sort": []interface{}{
					map[string]interface{}{
						"index": 2,
						"order": "asc",
					},
				},
			},
		},
	}

	result := buildWidgetSortByJSONFromMap(sortMap)

	orderBy := result["order_by"].([]interface{})
	entry := orderBy[0].(map[string]interface{})
	if entry["index"] != 2 {
		t.Errorf("expected index=2, got %#v", entry["index"])
	}
}
