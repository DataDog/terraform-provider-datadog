package dashboardmapping

import (
	"reflect"
	"sort"
	"testing"
)

// TestPrune_TreemapStyleInGroup verifies that a treemap widget nested inside
// a group preserves its `style` block end-to-end. This was the originally
// reported bug — `style` was emitted by the formula request flattener but the
// schema did not declare it on treemap requests, so the prune pass dropped it.
// `treemap_definition.request.style` is now declared, so flattened state must
// retain the field and emit zero drops for this dashboard shape.
func TestPrune_TreemapStyleInGroup(t *testing.T) {
	apiWidgets := []interface{}{
		map[string]interface{}{
			"id": float64(6005583415655279),
			"definition": map[string]interface{}{
				"title":       "New group",
				"type":        "group",
				"layout_type": "ordered",
				"widgets": []interface{}{
					map[string]interface{}{
						"id": float64(3095397288867546),
						"definition": map[string]interface{}{
							"type":  "treemap",
							"title": "",
							"requests": []interface{}{
								map[string]interface{}{
									"response_format": "scalar",
									"queries": []interface{}{
										map[string]interface{}{
											"data_source": "metrics",
											"name":        "query1",
											"query":       "sum:system.mem.total{*} by {service}",
											"aggregator":  "sum",
										},
									},
									"style": map[string]interface{}{
										"palette": "datadog16",
									},
								},
							},
						},
					},
				},
			},
		},
	}

	result, drops := FlattenWidgetsForSDKv2(apiWidgets)
	if len(result) != 1 {
		t.Fatalf("expected 1 flattened widget, got %d", len(result))
	}
	if len(drops) != 0 {
		t.Errorf("expected zero drops now that treemap.style is declared, got: %v", drops)
	}

	// Walk into the result and verify both legitimate keys survived flattening.
	groupWrapper := result[0].(map[string]interface{})
	groupDef := groupWrapper["group_definition"].([]interface{})[0].(map[string]interface{})
	nestedWidgets := groupDef["widget"].([]interface{})
	treemapWrapper := nestedWidgets[0].(map[string]interface{})
	treemapDef := treemapWrapper["treemap_definition"].([]interface{})[0].(map[string]interface{})
	reqs := treemapDef["request"].([]interface{})
	req0 := reqs[0].(map[string]interface{})
	if _, hasStyle := req0["style"]; !hasStyle {
		t.Errorf("expected 'style' to survive flatten, got: %v", req0)
	}
	if _, hasQuery := req0["query"]; !hasQuery {
		t.Error("expected 'query' to remain")
	}
}

// TestPrune_KeepsKnownFields verifies that timeseries requests (which legitimately
// have a 'style' block in their schema) are not affected by pruning.
func TestPrune_KeepsKnownFields(t *testing.T) {
	apiWidgets := []interface{}{
		map[string]interface{}{
			"definition": map[string]interface{}{
				"type": "timeseries",
				"requests": []interface{}{
					map[string]interface{}{
						"response_format": "timeseries",
						"queries": []interface{}{
							map[string]interface{}{
								"data_source": "metrics",
								"name":        "query1",
								"query":       "avg:system.cpu.user{*}",
							},
						},
						"style": map[string]interface{}{
							"palette":    "dog_classic",
							"line_type":  "solid",
							"line_width": "normal",
						},
					},
				},
			},
		},
	}
	_, drops := FlattenWidgetsForSDKv2(apiWidgets)
	if len(drops) != 0 {
		t.Errorf("expected no drops on a well-formed timeseries widget, got: %v", drops)
	}
}

// TestPrune_DirectHelper exercises pruneUnknownFields against a synthetic state map.
// FlattenEngineJSON itself filters API JSON to declared fields, so the pruner exists
// to clean up keys that POST-PROCESSORS inject (the treemap.style case is the
// motivating example). This test simulates that situation directly.
func TestPrune_DirectHelper(t *testing.T) {
	state := map[string]interface{}{
		"title":            "ok",
		"injected_unknown": "from a buggy post-processor",
		"request": []interface{}{
			map[string]interface{}{
				"query":          []interface{}{},
				"injected_style": map[string]interface{}{"palette": "x"},
			},
		},
	}
	fields := []FieldSpec{
		{HCLKey: "title", Type: TypeString},
		{HCLKey: "request", Type: TypeBlockList, Children: []FieldSpec{
			{HCLKey: "query", Type: TypeBlockList},
		}},
	}
	drops := pruneUnknownFields(state, fields, nil, "widget_def")

	gotPaths := make([]string, len(drops))
	copy(gotPaths, drops)
	sort.Strings(gotPaths)
	wantPaths := []string{
		"widget_def.injected_unknown",
		"widget_def.request[0].injected_style",
	}
	if !reflect.DeepEqual(gotPaths, wantPaths) {
		t.Errorf("dropped paths mismatch.\n got: %v\nwant: %v", gotPaths, wantPaths)
	}
	if _, present := state["injected_unknown"]; present {
		t.Error("expected injected_unknown to be removed from state")
	}
	req0 := state["request"].([]interface{})[0].(map[string]interface{})
	if _, present := req0["injected_style"]; present {
		t.Error("expected injected_style to be removed from nested state")
	}
	if _, present := req0["query"]; !present {
		t.Error("expected legitimate 'query' to remain")
	}
}

// TestPrune_RespectsExtraKnown verifies that fields listed in extraKnown are not dropped,
// even though they don't appear in the FieldSpec list.
func TestPrune_RespectsExtraKnown(t *testing.T) {
	state := map[string]interface{}{
		"title":  "ok",
		"widget": []interface{}{map[string]interface{}{}}, // post-process injection on group
	}
	fields := []FieldSpec{
		{HCLKey: "title", Type: TypeString},
	}
	extra := map[string][]FieldSpec{"widget": nil}
	drops := pruneUnknownFields(state, fields, extra, "group_def")
	if len(drops) != 0 {
		t.Errorf("expected no drops, got %v", drops)
	}
	if _, present := state["widget"]; !present {
		t.Error("expected 'widget' to be preserved via extraKnown")
	}
}
