package dashboardmapping

import (
	"encoding/json"
	"testing"

	ctyjson "github.com/hashicorp/go-cty/cty/json"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/stretchr/testify/require"
)

func TestInlineDDSQLFieldsDoNotChangeLegacyRequests(t *testing.T) {
	const q = "avg:system.load.1{*}"
	for _, viz := range []string{"query_value", "query_table"} {
		t.Run(viz, func(t *testing.T) {
			req := map[string]interface{}{"q": q, "response_format": "scalar"}
			widget := map[string]interface{}{"definition": map[string]interface{}{"type": viz, "requests": []interface{}{req}}}
			flat, _ := FlattenWidgetsForSDKv2([]interface{}{widget})
			def := flat[0].(map[string]interface{})[WidgetDefinitionHCLKey(viz)].([]interface{})[0].(map[string]interface{})
			request := def["request"].([]interface{})[0].(map[string]interface{})
			require.Equal(t, q, request["q"])
			require.NotContains(t, request, "response_format")
		})
	}
}

// Start with API requests, as refresh/import does. Passing through ResourceData
// catches both unknown schema fields and defaults introduced by SDKv2.
func TestInlineDDSQLWidgetRequestsRoundTrip(t *testing.T) {
	const tree = `[{"column":"team","dimension":"group"},{"column":"service","dimension":"group"},
		{"column":"n","dimension":"value","number_format":{"precision":0,"unit_scale":null}}]`
	const timeDims = `[{"column":"ts","dimension":"time"},{"column":"service","dimension":"group"},
		{"column":"n","dimension":"value","number_format":{"unit":{"type":"canonical_unit","unit_name":"second"}}}]`
	const tableDims = `[{"column":"service","dimension":"column","alias":"Service","text_align":"left",
		"link":{"url":"/services/{{service}}","label":"Open service"},"text_formats":[{"match":{"type":"is_null"},"palette":"red_on_white","highlight_row":false}]},
		{"column":"n","dimension":"column","cell_display_mode":"bar","number_format":{"precision":0},
		"conditional_formats":[{"comparator":">","value":0,"palette":"red_on_white","highlight_row":true}]},
		{"column":"duration","dimension":"column","is_hidden":false,"text_formats":[],"conditional_formats":[{"type":"range","palette":"transparent_to_green","min":0,"reverse":false}]},
		{"column":"ts","dimension":"column","is_hidden":true,"date_format":"MMM d"}]`
	const scatter = `[{"column":"service","dimension":"group"},{"column":"n","dimension":"x","number_format":{"unit_scale":null}},
		{"column":"duration","dimension":"y"},{"column":"size","dimension":"radius"},{"column":"team","dimension":"color"}]`
	const overlay = `{"display_type":"overlay","on_right_yaxis":false,"response_format":"timeseries",
		"queries":[{"data_source":"metrics","name":"query1","query":"avg:system.cpu.user{*}"}],"formulas":[{"formula":"query1"}]}`
	const thresholds = `{"conditional_formats":[{"comparator":">=","palette":"white_on_green","value":0}]}`
	const sql = "SELECT CURRENT_TIMESTAMP AS ts, 'core' AS team, 'api' AS service,\n       42 AS n, 7 AS duration, 2 AS size"
	widgetSchema := &schema.Resource{Schema: AllWidgetSDKv2Schema(false)}
	for _, tc := range []struct {
		name, viz, dimensions, extras string
		withOverlay                   bool
	}{
		{"timeseries", "timeseries", timeDims, ``, false},
		{"timeseries with Formula overlay first", "timeseries", timeDims, `{"display_type":"adjacent_bars","style":{"palette":"cool","line_width":"thin","color_order":"shuffled","order_reverse":false,"has_value_labels":true,"value_labels_density":"maximize"},"on_right_yaxis":true}`, true},
		{"query table without projection", "query_table", ``, ``, false},
		{"query table with formatted columns", "query_table", tableDims, ``, false},
		{"toplist", "toplist", tree, thresholds, false},
		{"bar chart", "bar_chart", tree, thresholds, false},
		{"treemap", "treemap", tree, `{"style":{"palette":"warm"}}`, false},
		{"sunburst", "sunburst", tree, `{"style":{"palette":"cool"}}`, false},
		{"scatterplot table", "scatterplot", scatter, ``, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			req := map[string]interface{}{
				"request_type": "advanced_query", "response_format": "tabular", "dataset_to_query": "result_rows",
				"datasets": []interface{}{map[string]interface{}{"data_source": "analysis_dataset", "name": "result_rows",
					"query": map[string]interface{}{"type": "sql_analysis", "sql_query": sql}}},
			}
			if tc.viz == "timeseries" {
				req["response_format"] = "timeseries"
			}
			if tc.dimensions != "" {
				var dimensions []interface{}
				require.NoError(t, json.Unmarshal([]byte(tc.dimensions), &dimensions))
				projection := map[string]interface{}{"type": tc.viz, "dimensions": dimensions}
				if tc.viz == "query_table" {
					projection["column_order"], projection["unlisted_columns"] = "dimensions", "hide"
				}
				req["projection"] = projection
			}
			if tc.extras != "" {
				require.NoError(t, json.Unmarshal([]byte(tc.extras), &req))
			}
			var requests interface{} = []interface{}{req}
			path := ".0.request.0."
			if tc.withOverlay {
				var formula map[string]interface{}
				require.NoError(t, json.Unmarshal([]byte(overlay), &formula))
				requests = []interface{}{formula, req}
				path = ".0.request.1."
			} else if tc.viz == "scatterplot" {
				requests = map[string]interface{}{"table": req}
				path += "scatterplot_table.0."
			}
			widget := map[string]interface{}{"definition": map[string]interface{}{"type": tc.viz, "requests": requests}}
			widgets, drops := FlattenWidgetsForSDKv2([]interface{}{widget})
			require.Empty(t, drops)
			flat := widgets[0].(map[string]interface{})
			diags := widgetSchema.Validate(terraform.NewResourceConfigRaw(flat))
			require.False(t, diags.HasError(), "%v", diags)
			// Reapply the exported configuration: ResourceData.Get materializes
			// default false even where the exported HCL omitted the attribute.
			hcl, err := json.Marshal(flat)
			require.NoError(t, err)
			var raw ctyjson.SimpleJSONValue
			require.NoError(t, json.Unmarshal(hcl, &raw))
			data := widgetSchema.Data(&terraform.InstanceState{RawConfig: raw.Value})
			key := WidgetDefinitionHCLKey(tc.viz)
			require.NoError(t, data.Set(key, flat[key]))
			require.Equal(t, sql, data.Get(key+path+"dataset.0.query.0.sql_query"))
			rebuilt := BuildWidgetEngineJSONFromMapWithRawConfig(map[string]interface{}{key: data.Get(key)}, data, nil)
			got, err := json.Marshal(rebuilt["definition"].(map[string]interface{})["requests"])
			require.NoError(t, err)
			want, err := json.Marshal(requests)
			require.NoError(t, err)
			require.JSONEq(t, string(want), string(got))
		})
	}
}
