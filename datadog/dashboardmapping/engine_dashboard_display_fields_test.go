package dashboardmapping

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestDashboardDisplayFieldsRoundTrip(t *testing.T) {
	for name, definition := range map[string]string{
		"monitor_summary":              `{"type":"manage_status","query":"tag:service:example","count":50,"start":0,"last_triggered_format":"relative","show_status":true,"show_investigation":false}`,
		"monitor_summary_false_status": `{"type":"manage_status","query":"tag:service:example","show_status":false,"show_investigation":true}`,
		"slo_monthly_rollup":           `{"type":"slo_list","requests":[{"request_type":"slo_list","query":{"query_string":"team:example","rollup":{"type":"month"}}}]}`,
		"event_grouping_and_palette":   `{"type":"timeseries","requests":[{"response_format":"timeseries","formulas":[{"formula":"events"}],"queries":[{"data_source":"events","name":"events","compute":{"aggregation":"count"},"group_by":[{"facet":"host","should_exclude_missing":true}]}],"style":{"palette":"dog_classic","color_order":"monotonic"}}]}`,
		"include_missing_events":       `{"type":"timeseries","requests":[{"response_format":"timeseries","formulas":[{"formula":"events"}],"queries":[{"data_source":"events","name":"events","compute":{"aggregation":"count"},"group_by":[{"facet":"host","should_exclude_missing":false}]}]}]}`,
	} {
		t.Run(name, func(t *testing.T) {
			var expected map[string]interface{}
			if err := json.Unmarshal([]byte(definition), &expected); err != nil {
				t.Fatal(err)
			}
			apiWidget := map[string]interface{}{"definition": expected}
			// Exercise the state-pruning path used by resource reads and imports.
			flat, dropped := FlattenWidgetsForSDKv2([]interface{}{apiWidget})
			if len(dropped) != 0 {
				t.Fatalf("fields lost during import: %v", dropped)
			}
			rebuilt := BuildWidgetEngineJSONFromMap(flat[0].(map[string]interface{}))
			encoded, err := json.Marshal(rebuilt)
			if err != nil {
				t.Fatal(err)
			}
			var actual map[string]interface{}
			if err := json.Unmarshal(encoded, &actual); err != nil {
				t.Fatal(err)
			}
			assertDashboardDisplaySubset(t, apiWidget, actual)
		})
	}
}

// Existing fields may acquire provider defaults; every supplied API field must survive.
func assertDashboardDisplaySubset(t *testing.T, expected, actual interface{}) {
	t.Helper()
	switch want := expected.(type) {
	case map[string]interface{}:
		got, ok := actual.(map[string]interface{})
		if !ok {
			t.Fatalf("expected object %v, got %v", want, actual)
		}
		for key, value := range want {
			if _, ok := got[key]; !ok {
				t.Fatalf("missing field %s in %v", key, got)
			}
			assertDashboardDisplaySubset(t, value, got[key])
		}
	case []interface{}:
		got, ok := actual.([]interface{})
		if !ok || len(want) != len(got) {
			t.Fatalf("expected list %v, got %v", want, actual)
		}
		for i := range want {
			assertDashboardDisplaySubset(t, want[i], got[i])
		}
	default:
		if !reflect.DeepEqual(expected, actual) {
			t.Fatalf("expected %v, got %v", expected, actual)
		}
	}
}

func TestDashboardOptionalFalseRawConfig(t *testing.T) {
	for _, configured := range []bool{false, true} {
		t.Run(map[bool]string{false: "unset", true: "explicit_false"}[configured], func(t *testing.T) {
			fields := []FieldSpec{{HCLKey: "show_status", Type: TypeBool, OmitEmpty: true, PreserveZero: true}}
			ctx := mapBuildContext{rawConfig: rawConfigAtFunc(func(cty.Path) (cty.Value, diag.Diagnostics) {
				if configured {
					return cty.False, nil
				}
				return cty.NullVal(cty.Bool), nil
			})}
			got := buildEngineJSONFromMap(map[string]interface{}{}, fields, ctx)
			value, present := got["show_status"]
			if present != configured || (present && value != false) {
				t.Fatalf("configured=%v: unexpected result %v", configured, got)
			}
		})
	}
}

func TestDashboardUnsetDisplayFieldsSDKData(t *testing.T) {
	d := schema.TestResourceDataRaw(t, map[string]*schema.Schema{
		"widget": {Type: schema.TypeList, Optional: true, Elem: &schema.Resource{Schema: AllWidgetSDKv2Schema(false)}},
	}, map[string]interface{}{
		"widget": []interface{}{map[string]interface{}{
			"manage_status_definition": []interface{}{map[string]interface{}{"query": "tag:service:example"}},
		}},
	})
	widget := d.Get("widget").([]interface{})[0].(map[string]interface{})
	rawConfig := rawConfigAtFunc(func(cty.Path) (cty.Value, diag.Diagnostics) {
		return cty.NullVal(cty.Bool), nil
	})
	got := BuildWidgetEngineJSONFromMapWithRawConfig(widget, rawConfig, cty.GetAttrPath("widget").IndexInt(0))
	definition := got["definition"].(map[string]interface{})
	for _, field := range []string{"show_status", "show_investigation", "count", "start"} {
		if _, present := definition[field]; present {
			t.Errorf("unset %s must remain omitted; got %v", field, definition[field])
		}
	}
}
