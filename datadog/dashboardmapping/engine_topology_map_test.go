package dashboardmapping

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestTopologyMapQueryRoundTrip(t *testing.T) {
	for _, tc := range []struct {
		name    string
		query   map[string]interface{}
		filters []string
	}{
		{
			name: "service_map",
			query: map[string]interface{}{
				"data_source": "service_map",
				"service":     "master-db",
				"filters":     []interface{}{"env:prod", "datacenter:dc1"},
			},
			filters: []string{"env:prod", "datacenter:dc1"},
		},
		{
			name: "data_streams with query_string",
			query: map[string]interface{}{
				"data_source":  "data_streams",
				"service":      "",
				"filters":      []interface{}{"env:prod"},
				"query_string": "service:web-store",
			},
			filters: []string{"env:prod"},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			widget := map[string]interface{}{
				"topology_map_definition": []interface{}{map[string]interface{}{
					"request": []interface{}{map[string]interface{}{
						"request_type": "topology",
						"query":        []interface{}{tc.query},
					}},
				}},
			}

			built := BuildWidgetEngineJSONFromMap(widget)
			definition := built["definition"].(map[string]interface{})
			request := definition["requests"].([]interface{})[0].(map[string]interface{})
			builtQuery := request["query"].(map[string]interface{})
			if builtQuery["data_source"] != tc.query["data_source"] {
				t.Fatalf("topology query was not serialized: %#v", builtQuery)
			}
			if _, ok := tc.query["query_string"]; ok {
				if builtQuery["query_string"] != tc.query["query_string"] {
					t.Fatalf("query_string was not serialized: %#v", builtQuery)
				}
			} else if _, ok := builtQuery["query_string"]; ok {
				t.Fatalf("query_string should be omitted when unset: %#v", builtQuery)
			}

			flattened, _ := FlattenWidgetEngineJSON(apiRoundTrip(t, built))
			flatDefinition := flattened["topology_map_definition"].([]interface{})[0].(map[string]interface{})
			flatRequest := flatDefinition["request"].([]interface{})[0].(map[string]interface{})
			flatQuery := flatRequest["query"].([]interface{})[0].(map[string]interface{})
			for key, want := range tc.query {
				if key == "filters" {
					continue
				}
				if got := flatQuery[key]; !reflect.DeepEqual(got, want) {
					t.Fatalf("flattened %s = %#v, want %#v", key, got, want)
				}
			}
			if got := flatQuery["filters"]; !reflect.DeepEqual(got, tc.filters) {
				t.Fatalf("flattened filters = %#v, want %#v", got, tc.filters)
			}
		})
	}
}

func apiRoundTrip(t *testing.T, data map[string]interface{}) map[string]interface{} {
	t.Helper()
	encoded, err := json.Marshal(data)
	if err != nil {
		t.Fatalf("marshal: %s", err)
	}
	var decoded map[string]interface{}
	if err := json.Unmarshal(encoded, &decoded); err != nil {
		t.Fatalf("unmarshal: %s", err)
	}
	return decoded
}
