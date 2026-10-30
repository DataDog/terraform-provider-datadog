package dashboardmapping

import "testing"

// API-created widgets and widgets edited through the JSON editor can omit
// these fields. Their state must match the visible default without overwriting
// an explicitly configured false value.
func TestFlattenWidgetEngineJSON_BoolDefaults(t *testing.T) {
	fields := []struct {
		widgetType string
		block      string
		field      string
	}{
		{"note", "note_definition", "has_padding"},
		{"image", "image_definition", "has_background"},
		{"image", "image_definition", "has_border"},
		{"group", "group_definition", "show_title"},
	}
	cases := []struct {
		name    string
		present bool
		value   interface{}
		want    bool
	}{
		{"missing", false, nil, true},
		{"null", true, nil, true},
		{"explicit_false", true, false, false},
		{"explicit_true", true, true, true},
	}

	for _, field := range fields {
		for _, tc := range cases {
			t.Run(field.field+"/"+tc.name, func(t *testing.T) {
				definition := map[string]interface{}{
					"type": field.widgetType,
				}
				switch field.widgetType {
				case "note":
					definition["content"] = "Existing note"
				case "image":
					definition["url"] = "https://example.com/image.png"
				case "group":
					definition["layout_type"] = "ordered"
				}
				if tc.present {
					definition[field.field] = tc.value
				}

				got, drops := FlattenWidgetEngineJSON(map[string]interface{}{
					"definition": definition,
				})
				if len(drops) != 0 {
					t.Fatalf("unexpected dropped fields: %v", drops)
				}
				blocks, ok := got[field.block].([]interface{})
				if !ok || len(blocks) != 1 {
					t.Fatalf("expected one %s block, got %#v", field.block, got)
				}
				state := blocks[0].(map[string]interface{})
				if state[field.field] != tc.want {
					t.Errorf("%s = %#v, want %v", field.field, state[field.field], tc.want)
				}

				// An unrelated later apply must preserve the same visual setting,
				// including an explicit false, in the outgoing API payload.
				rebuilt := BuildWidgetEngineJSONFromMap(got)
				rebuiltDefinition := rebuilt["definition"].(map[string]interface{})
				if rebuiltDefinition[field.field] != tc.want {
					t.Errorf("rebuilt %s = %#v, want %v", field.field, rebuiltDefinition[field.field], tc.want)
				}
			})
		}
	}
}

func TestFlattenWidgetsForSDKv2_BoolDefaultsInGroup(t *testing.T) {
	widgets := []interface{}{
		map[string]interface{}{
			"definition": map[string]interface{}{
				"type":        "group",
				"layout_type": "ordered",
				"widgets": []interface{}{
					map[string]interface{}{
						"definition": map[string]interface{}{
							"type":    "note",
							"content": "Missing padding setting",
						},
					},
					map[string]interface{}{
						"definition": map[string]interface{}{
							"type":        "note",
							"content":     "Explicitly disabled padding",
							"has_padding": false,
						},
					},
				},
			},
		},
	}

	got, drops := FlattenWidgetsForSDKv2(widgets)
	if len(drops) != 0 {
		t.Fatalf("unexpected dropped fields: %v", drops)
	}
	if len(got) != 1 {
		t.Fatalf("expected one group widget, got %#v", got)
	}
	groupWrapper := got[0].(map[string]interface{})
	group := groupWrapper["group_definition"].([]interface{})[0].(map[string]interface{})
	if group["show_title"] != true {
		t.Errorf("group show_title = %#v, want true", group["show_title"])
	}
	children := group["widget"].([]interface{})
	if len(children) != 2 {
		t.Fatalf("expected two nested notes, got %#v", children)
	}
	for i, want := range []bool{true, false} {
		child := children[i].(map[string]interface{})
		note := child["note_definition"].([]interface{})[0].(map[string]interface{})
		if note["has_padding"] != want {
			t.Errorf("nested note %d has_padding = %#v, want %v", i, note["has_padding"], want)
		}
	}
}
