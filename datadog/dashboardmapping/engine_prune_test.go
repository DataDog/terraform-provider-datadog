package dashboardmapping

import (
	"reflect"
	"sort"
	"testing"
)

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
