package datadog

import (
	"context"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestPowerpackV2EmbeddedAppIdentifierValidation(t *testing.T) {
	tests := []struct {
		name       string
		definition map[string]interface{}
		wantError  bool
	}{
		{
			name:       "app id only",
			definition: map[string]interface{}{"app_id": "7e7745f9-4343-4927-b038-80934a355915"},
		},
		{
			name:       "template id only",
			definition: map[string]interface{}{"template_id": "ec2_manager"},
		},
		{
			name:       "neither identifier",
			definition: map[string]interface{}{},
			wantError:  true,
		},
		{
			name: "both identifiers",
			definition: map[string]interface{}{
				"app_id":      "7e7745f9-4343-4927-b038-80934a355915",
				"template_id": "ec2_manager",
			},
			wantError: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			config := terraform.NewResourceConfigRaw(map[string]interface{}{
				"name": "Embedded app powerpack",
				"widget": []interface{}{
					map[string]interface{}{
						"embedded_app_definition": []interface{}{test.definition},
					},
				},
			})

			_, err := resourceDatadogPowerpackV2().Diff(context.Background(), nil, config, nil)
			if test.wantError {
				if err == nil || !strings.Contains(err.Error(), `exactly one of "app_id" or "template_id"`) {
					t.Fatalf("expected identifier validation error, got %v", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("expected valid embedded app definition, got %v", err)
			}
		})
	}
}
