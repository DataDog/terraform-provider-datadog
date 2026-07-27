package datadog

import (
	"context"
	"testing"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestResourceDiffRawConfigHelpers(t *testing.T) {
	tests := []struct {
		name              string
		config            map[string]interface{}
		rawConfig         cty.Value
		wantConfigured    bool
		wantExplicitFalse bool
	}{
		{
			name:              "omitted",
			config:            map[string]interface{}{},
			rawConfig:         cty.ObjectVal(map[string]cty.Value{"value": cty.NullVal(cty.Bool)}),
			wantConfigured:    false,
			wantExplicitFalse: false,
		},
		{
			name:              "explicit false",
			config:            map[string]interface{}{"value": false},
			rawConfig:         cty.ObjectVal(map[string]cty.Value{"value": cty.False}),
			wantConfigured:    true,
			wantExplicitFalse: true,
		},
		{
			name:              "explicit true",
			config:            map[string]interface{}{"value": true},
			rawConfig:         cty.ObjectVal(map[string]cty.Value{"value": cty.True}),
			wantConfigured:    true,
			wantExplicitFalse: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var gotConfigured, gotExplicitFalse bool
			resource := &schema.Resource{
				Schema: map[string]*schema.Schema{
					"value": {
						Type:     schema.TypeBool,
						Optional: true,
					},
				},
				CustomizeDiff: func(_ context.Context, diff *schema.ResourceDiff, _ interface{}) error {
					gotConfigured = isResourceDiffAttributeConfigured(diff, "value")
					gotExplicitFalse = isResourceDiffOptionalBoolFalse(diff, "value")
					return nil
				},
			}

			if _, err := resource.Diff(
				context.Background(),
				&terraform.InstanceState{RawConfig: test.rawConfig},
				terraform.NewResourceConfigRaw(test.config),
				nil,
			); err != nil {
				t.Fatalf("building resource diff: %v", err)
			}
			if gotConfigured != test.wantConfigured {
				t.Fatalf("configured = %t, want %t", gotConfigured, test.wantConfigured)
			}
			if gotExplicitFalse != test.wantExplicitFalse {
				t.Fatalf("explicit false = %t, want %t", gotExplicitFalse, test.wantExplicitFalse)
			}
		})
	}
}
