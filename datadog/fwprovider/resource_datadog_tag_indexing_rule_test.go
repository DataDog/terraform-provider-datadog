package fwprovider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestTagIndexingRuleValidateConfig(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	r := &tagIndexingRuleResource{}
	var schemaResponse resource.SchemaResponse
	r.Schema(ctx, resource.SchemaRequest{}, &schemaResponse)

	configType := schemaResponse.Schema.Type().TerraformType(ctx).(tftypes.Object)
	optionsType := configType.AttributeTypes["options"].(tftypes.Object)
	dataType := optionsType.AttributeTypes["data"].(tftypes.Object)
	dynamicTagsType := dataType.AttributeTypes["dynamic_tags"].(tftypes.Object)

	options := func(dynamicTags tftypes.Value) tftypes.Value {
		return tftypes.NewValue(optionsType, map[string]tftypes.Value{
			"version": tftypes.NewValue(tftypes.Number, 1),
			"data": tftypes.NewValue(dataType, map[string]tftypes.Value{
				"override_previous_rules":    tftypes.NewValue(tftypes.Bool, nil),
				"manage_preexisting_metrics": tftypes.NewValue(tftypes.Bool, nil),
				"dynamic_tags":               dynamicTags,
			}),
		})
	}
	dynamicTags := func(window any) tftypes.Value {
		return tftypes.NewValue(dynamicTagsType, map[string]tftypes.Value{
			"exclude_not_queried_window_seconds": tftypes.NewValue(tftypes.Number, window),
			"exclude_not_used_in_assets":         tftypes.NewValue(tftypes.Bool, nil),
		})
	}

	tests := []struct {
		name            string
		options         tftypes.Value
		excludeTagsMode any
		wantErr         string
	}{
		{name: "no options", options: tftypes.NewValue(optionsType, nil), excludeTagsMode: false},
		{name: "unknown options", options: tftypes.NewValue(optionsType, tftypes.UnknownValue), excludeTagsMode: false},
		{name: "unknown dynamic_tags", options: options(tftypes.NewValue(dynamicTagsType, tftypes.UnknownValue)), excludeTagsMode: false},
		{name: "usage field with exclude mode", options: options(dynamicTags(3600)), excludeTagsMode: true},
		{name: "usage field with unknown exclude mode", options: options(dynamicTags(3600)), excludeTagsMode: tftypes.UnknownValue},
		{name: "usage field without exclude mode", options: options(dynamicTags(3600)), excludeTagsMode: false, wantErr: "Invalid dynamic_tags usage configuration"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			attributes := make(map[string]tftypes.Value, len(configType.AttributeTypes))
			for name, attributeType := range configType.AttributeTypes {
				attributes[name] = tftypes.NewValue(attributeType, nil)
			}
			attributes["options"] = tt.options
			attributes["exclude_tags_mode"] = tftypes.NewValue(tftypes.Bool, tt.excludeTagsMode)

			config := tfsdk.Config{
				Raw:    tftypes.NewValue(configType, attributes),
				Schema: schemaResponse.Schema,
			}
			var response resource.ValidateConfigResponse
			r.ValidateConfig(ctx, resource.ValidateConfigRequest{Config: config}, &response)

			errors := response.Diagnostics.Errors()
			if tt.wantErr == "" {
				if len(errors) != 0 {
					t.Fatalf("unexpected validation errors: %v", errors)
				}
				return
			}
			if len(errors) != 1 || errors[0].Summary() != tt.wantErr {
				t.Fatalf("expected %q, got %v", tt.wantErr, errors)
			}
		})
	}
}
