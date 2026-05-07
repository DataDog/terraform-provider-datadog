package dashboardmapping

import "testing"

// TestBuildWidgetEngineJSONFromMap_LayoutZeroValuesPreserved verifies that
// widget_layout x/y are emitted in the API JSON even when the user-supplied
// values are zero. Stripping zero-valued x/y previously produced 400 errors
// of the form "MSL widget out of grid. 'x' is a required property" for any
// widget anchored at the dashboard origin (or row/column 0). See #3750.
func TestBuildWidgetEngineJSONFromMap_LayoutZeroValuesPreserved(t *testing.T) {
	cases := []struct {
		name   string
		layout map[string]interface{}
		wantX  int
		wantY  int
	}{
		{"origin", map[string]interface{}{"x": 0, "y": 0, "width": 4, "height": 5}, 0, 0},
		{"left edge", map[string]interface{}{"x": 0, "y": 5, "width": 4, "height": 5}, 0, 5},
		{"top edge", map[string]interface{}{"x": 4, "y": 0, "width": 4, "height": 5}, 4, 0},
		{"interior", map[string]interface{}{"x": 4, "y": 5, "width": 4, "height": 5}, 4, 5},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			widget := map[string]interface{}{
				"note_definition": []interface{}{
					map[string]interface{}{"content": "x"},
				},
				"widget_layout": []interface{}{tc.layout},
			}
			out := buildWidgetEngineJSONFromMap(widget)
			if out == nil {
				t.Fatalf("buildWidgetEngineJSONFromMap returned nil")
			}
			layoutOut, ok := out["layout"].(map[string]interface{})
			if !ok {
				t.Fatalf("expected layout map in widget JSON, got: %#v", out["layout"])
			}
			for _, key := range []string{"x", "y", "width", "height"} {
				if _, ok := layoutOut[key]; !ok {
					t.Errorf("layout missing %q (full layout: %#v)", key, layoutOut)
				}
			}
			if got := layoutOut["x"]; got != tc.wantX {
				t.Errorf("x = %v, want %v", got, tc.wantX)
			}
			if got := layoutOut["y"]; got != tc.wantY {
				t.Errorf("y = %v, want %v", got, tc.wantY)
			}
		})
	}
}
