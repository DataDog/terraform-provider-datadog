package dashboardmapping

import (
	"encoding/json"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/stretchr/testify/require"
)

func TestQueryValueInlineDDSQLRoundTrip(t *testing.T) {
	const apiWidget = `{"definition": {
		"type": "query_value", "title": "Inline DDSQL",
		"requests": [{
			"request_type": "advanced_query",
			"datasets": [{"data_source": "analysis_dataset", "name": "ddsql_query",
				"query": {"type": "sql_analysis", "sql_query": "SELECT 42 AS value"}}],
			"dataset_to_query": "ddsql_query", "response_format": "tabular",
			"projection": {"type": "query_value", "dimensions": [{"column": "value", "dimension": "value"}]}
		}]
	}}`
	widgetSchema := &schema.Resource{Schema: AllWidgetSDKv2Schema(false)}
	for _, tc := range []struct{ name, numberFormat string }{
		{"exported request", ""},
		{"precision only", `{"precision":0}`},
		{"explicit null scale", `{"precision":"*","unit":{"type":"custom_unit_label","label":"%"},"unit_scale":null}`},
		{"canonical scale", `{"precision":"integer","unit_scale":{"type":"canonical_unit","unit_name":"second"}}`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var widget map[string]interface{}
			require.NoError(t, json.Unmarshal([]byte(apiWidget), &widget))
			request := widget["definition"].(map[string]interface{})["requests"].([]interface{})[0].(map[string]interface{})
			if tc.numberFormat != "" {
				var format map[string]interface{}
				require.NoError(t, json.Unmarshal([]byte(tc.numberFormat), &format))
				dim := request["projection"].(map[string]interface{})["dimensions"].([]interface{})[0].(map[string]interface{})
				dim["number_format"] = format
				request["datasets"].([]interface{})[0].(map[string]interface{})["name"] = "local_result"
				request["dataset_to_query"] = "local_result"
			}

			widgets, dropped := FlattenWidgetsForSDKv2([]interface{}{widget})
			require.Empty(t, dropped)
			state := widgets[0].(map[string]interface{})
			diags := widgetSchema.Validate(terraform.NewResourceConfigRaw(state))
			require.False(t, diags.HasError(), "%v", diags)
			data := widgetSchema.TestResourceData()
			const key = "query_value_definition"
			require.NoError(t, data.Set(key, state[key]))
			const req = key + ".0.request.0."
			require.Equal(t, request["dataset_to_query"], data.Get(req+"dataset.0.name"))
			require.Equal(t, "SELECT 42 AS value", data.Get(req+"dataset.0.query.0.sql_query"))
			require.Equal(t, "value", data.Get(req+"projection.0.dimension.0.column"))

			rebuilt := BuildWidgetEngineJSONFromMap(map[string]interface{}{key: data.Get(key)})
			requests := rebuilt["definition"].(map[string]interface{})["requests"]
			want, err := json.Marshal([]interface{}{request})
			require.NoError(t, err)
			got, err := json.Marshal(requests)
			require.NoError(t, err)
			require.JSONEq(t, string(want), string(got))
		})
	}
}
