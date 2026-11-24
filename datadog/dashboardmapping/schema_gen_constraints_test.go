package dashboardmapping

import (
	"reflect"
	"testing"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestFieldSpecListConstraints(t *testing.T) {
	computeSchema := FieldSpecToSDKv2(FieldSpec{
		HCLKey:   "compute",
		Type:     TypeBlockList,
		MinItems: 1,
		MaxItems: 5,
		Children: []FieldSpec{{
			HCLKey:   "aggregation",
			Type:     TypeString,
			Required: true,
		}},
	})
	if computeSchema.MinItems != 1 || computeSchema.MaxItems != 5 {
		t.Fatalf("block list constraints were not registered: %#v", computeSchema)
	}

	statesSchema := FieldSpecToSDKv2(FieldSpec{
		HCLKey:      "states",
		Type:        TypeStringList,
		MinItems:    1,
		MaxItems:    2,
		ValidValues: []string{"OPEN", "RESOLVED"},
	})
	if statesSchema.MinItems != 1 || statesSchema.MaxItems != 2 {
		t.Fatalf("string list constraints were not registered: %#v", statesSchema)
	}
	elem, ok := statesSchema.Elem.(*schema.Schema)
	if !ok || elem.ValidateDiagFunc == nil {
		t.Fatalf("string-list enum validation was not registered: %#v", statesSchema.Elem)
	}
}

func TestFieldSpecJSONValidation(t *testing.T) {
	jsonSchema := FieldSpecToSDKv2(FieldSpec{
		HCLKey:   "value",
		Type:     TypeJSON,
		Required: true,
	})
	if jsonSchema.ValidateDiagFunc == nil {
		t.Fatal("JSON validation was not registered")
	}
	if diags := jsonSchema.ValidateDiagFunc(`{"valid":true}`, cty.Path{}); diags.HasError() {
		t.Fatalf("expected valid JSON to pass validation: %v", diags)
	}
	if diags := jsonSchema.ValidateDiagFunc("not-json", cty.Path{}); !diags.HasError() {
		t.Fatal("expected invalid JSON to fail validation")
	}
}

func TestNestedGroupWidgetSchemaRoundTrip(t *testing.T) {
	group := func(title string, child map[string]interface{}) map[string]interface{} {
		return map[string]interface{}{
			"group_definition": []interface{}{map[string]interface{}{
				"title":       title,
				"layout_type": "ordered",
				"widget":      []interface{}{child},
			}},
		}
	}
	leaf := map[string]interface{}{
		"id": 123,
		"note_definition": []interface{}{map[string]interface{}{
			"content": "Nested note",
		}},
		"widget_layout": []interface{}{map[string]interface{}{
			"x": 0, "y": 0, "width": 4, "height": 2,
		}},
	}
	inner := group("Inner group", leaf)
	inner["id"] = 456
	inner["widget_layout"] = []interface{}{map[string]interface{}{
		"x": 0, "y": 0, "width": 12, "height": 4,
	}}
	config := map[string]interface{}{
		"widget": []interface{}{group("Outer group", inner)},
	}

	for _, tc := range []struct {
		name                 string
		excludePowerpackOnly bool
	}{
		{name: "dashboard"},
		{name: "powerpack", excludePowerpackOnly: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			widgetSchema := AllWidgetSDKv2Schema(tc.excludePowerpackOnly)
			children := widgetSchema
			for depth := 0; depth < 2; depth++ {
				groupSchema, ok := children["group_definition"]
				if !ok {
					t.Fatalf("missing group schema at depth %d", depth)
				}
				groupResource := groupSchema.Elem.(*schema.Resource)
				children = groupResource.Schema["widget"].Elem.(*schema.Resource).Schema
				for _, excluded := range []string{"powerpack_definition", "split_graph_definition"} {
					if _, ok := children[excluded]; ok {
						t.Fatalf("%s must remain excluded inside groups", excluded)
					}
				}
			}
			if _, ok := children["group_definition"]; ok {
				t.Fatal("group nesting must terminate after two group levels")
			}

			resourceSchema := map[string]*schema.Schema{
				"widget": {
					Type:        schema.TypeList,
					Optional:    true,
					Description: "Dashboard widgets.",
					Elem:        &schema.Resource{Schema: widgetSchema},
				},
			}
			data := schema.TestResourceDataRaw(t, resourceSchema, config)
			leafPath := "widget.0.group_definition.0.widget.0.group_definition.0.widget.0"
			if got := data.Get(leafPath + ".note_definition.0.content"); got != "Nested note" {
				t.Fatalf("nested note was lost: %#v", got)
			}
			if got := data.Get(leafPath + ".id"); got != 123 {
				t.Fatalf("nested note ID was lost: %#v", got)
			}
			widget := data.Get("widget").([]interface{})[0].(map[string]interface{})
			payload := BuildWidgetEngineJSONFromMap(widget)
			flattened, drops := FlattenWidgetsForSDKv2([]interface{}{payload})
			if len(drops) != 0 {
				t.Fatalf("nested widget fields were dropped: %v", drops)
			}
			refreshed := schema.TestResourceDataRaw(t, resourceSchema, nil)
			if err := refreshed.Set("widget", flattened); err != nil {
				t.Fatalf("cannot save nested groups to state: %v", err)
			}
			if !reflect.DeepEqual(data.Get("widget"), refreshed.Get("widget")) {
				t.Fatalf("nested groups changed after refresh:\n got: %#v\nwant: %#v", refreshed.Get("widget"), data.Get("widget"))
			}
		})
	}
}
