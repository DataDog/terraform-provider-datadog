package datadog

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// A nested number_format block whose attributes are all zero values is read by
// the SDK as a nil list element. Building the request must not panic on it.
func TestPowerpackNumberFormatWithEmptyNestedBlocks(t *testing.T) {
	customUnit := map[string]interface{}{"custom": []interface{}{map[string]interface{}{"label": "req"}}}
	tests := []struct {
		name         string
		numberFormat map[string]interface{}
		want         string
		notWant      []string
	}{
		{
			name:         "empty canonical unit",
			numberFormat: map[string]interface{}{"unit": []interface{}{map[string]interface{}{"canonical": []interface{}{map[string]interface{}{"unit_name": ""}}}}},
			notWant:      []string{"canonical_unit", `"unit":`},
		},
		{
			name:         "empty custom unit",
			numberFormat: map[string]interface{}{"unit": []interface{}{map[string]interface{}{"custom": []interface{}{map[string]interface{}{"label": ""}}}}},
			notWant:      []string{"custom_unit_label", `"unit":`},
		},
		{
			name:         "empty unit block",
			numberFormat: map[string]interface{}{"unit": []interface{}{map[string]interface{}{}}},
			notWant:      []string{`"unit":`},
		},
		{
			name:         "empty unit_scale",
			numberFormat: map[string]interface{}{"unit": []interface{}{customUnit}, "unit_scale": []interface{}{map[string]interface{}{"unit_name": ""}}},
			want:         `"number_format":{"unit":{"label":"req","type":"custom_unit_label"}}`,
		},
		{
			name: "canonical unit and unit_scale",
			numberFormat: map[string]interface{}{
				"unit":       []interface{}{map[string]interface{}{"canonical": []interface{}{map[string]interface{}{"unit_name": "byte", "per_unit_name": "second"}}}},
				"unit_scale": []interface{}{map[string]interface{}{"unit_name": "megabyte"}},
			},
			want: `"number_format":{"unit":{"per_unit_name":"second","type":"canonical_unit","unit_name":"byte"},"unit_scale":{"type":"canonical_unit","unit_name":"megabyte"}}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := schema.TestResourceDataRaw(t, resourceDatadogPowerpack().SchemaMap(), map[string]interface{}{
				"name": "number format",
				"widget": []interface{}{map[string]interface{}{
					"query_value_definition": []interface{}{map[string]interface{}{
						"request": []interface{}{map[string]interface{}{
							"formula": []interface{}{map[string]interface{}{
								"formula_expression": "query1",
								"number_format":      []interface{}{tt.numberFormat},
							}},
							"query": []interface{}{map[string]interface{}{
								"metric_query": []interface{}{map[string]interface{}{"name": "query1", "query": "avg:system.cpu.user{*}"}},
							}},
						}},
					}},
				}},
			})
			powerpack, diags := buildDatadogPowerpack(context.Background(), d)
			if diags.HasError() {
				t.Fatalf("unexpected diagnostics: %v", diags)
			}
			payload, err := json.Marshal(powerpack)
			if err != nil {
				t.Fatalf("marshal failed: %v", err)
			}
			body := string(payload)
			if !strings.Contains(body, `"formula":"query1"`) {
				t.Fatalf("formula missing from payload: %s", body)
			}
			if tt.want != "" && !strings.Contains(body, tt.want) {
				t.Errorf("payload does not contain %s: %s", tt.want, body)
			}
			for _, s := range tt.notWant {
				if strings.Contains(body, s) {
					t.Errorf("payload unexpectedly contains %s: %s", s, body)
				}
			}
		})
	}
}
