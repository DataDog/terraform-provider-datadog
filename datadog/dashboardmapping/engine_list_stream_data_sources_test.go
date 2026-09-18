package dashboardmapping

import (
	"testing"

	"github.com/hashicorp/go-cty/cty"
)

func TestListStreamDataSourcesValidate(t *testing.T) {
	schema := FieldSpecsToSDKv2Schema(listStreamQueryFields)["data_source"]
	if schema.ValidateDiagFunc == nil {
		t.Fatal("list stream data source enum validation was not registered")
	}

	for _, dataSource := range []string{
		"security_runtime_stream",
		"security_signals_stream",
		"incidents_stream",
	} {
		if diags := schema.ValidateDiagFunc(dataSource, cty.Path{}); diags.HasError() {
			t.Fatalf("list stream data source %q was rejected: %#v", dataSource, diags)
		}
	}

	if diags := schema.ValidateDiagFunc("unknown_stream", cty.Path{}); !diags.HasError() {
		t.Fatal("unknown list stream data source was accepted")
	}
}

func TestListStreamDataSourcesRoundTrip(t *testing.T) {
	for _, dataSource := range []string{
		"security_runtime_stream",
		"security_signals_stream",
		"incidents_stream",
	} {
		t.Run(dataSource, func(t *testing.T) {
			widget := map[string]interface{}{
				"list_stream_definition": []interface{}{map[string]interface{}{
					"request": []interface{}{map[string]interface{}{
						"response_format": "event_list",
						"query": []interface{}{map[string]interface{}{
							"data_source":  dataSource,
							"query_string": "env:prod",
						}},
					}},
				}},
			}

			built := BuildWidgetEngineJSONFromMap(widget)
			definition := built["definition"].(map[string]interface{})
			request := definition["requests"].([]interface{})[0].(map[string]interface{})
			query := request["query"].(map[string]interface{})
			if query["data_source"] != dataSource {
				t.Fatalf("list stream data source was not serialized: %#v", query)
			}

			flattened, _ := FlattenWidgetEngineJSON(built)
			flatDefinition := flattened["list_stream_definition"].([]interface{})[0].(map[string]interface{})
			flatRequest := flatDefinition["request"].([]interface{})[0].(map[string]interface{})
			flatQuery := flatRequest["query"].([]interface{})[0].(map[string]interface{})
			if flatQuery["data_source"] != dataSource {
				t.Fatalf("list stream data source was not restored during flatten: %#v", flatQuery)
			}
		})
	}
}
