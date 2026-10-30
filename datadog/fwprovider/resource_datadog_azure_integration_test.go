package fwprovider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestIntegrationAzureValidateConfig(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	r := &integrationAzureResource{}

	var schemaResponse resource.SchemaResponse
	r.Schema(ctx, resource.SchemaRequest{}, &schemaResponse)
	objectType := schemaResponse.Schema.Type().TerraformType(ctx).(tftypes.Object)

	newConfig := func(values map[string]tftypes.Value) tfsdk.Config {
		attrs := make(map[string]tftypes.Value, len(objectType.AttributeTypes))
		for name, attrType := range objectType.AttributeTypes {
			attrs[name] = tftypes.NewValue(attrType, nil)
		}
		for name, value := range values {
			attrs[name] = value
		}
		return tfsdk.Config{
			Raw:    tftypes.NewValue(objectType, attrs),
			Schema: schemaResponse.Schema,
		}
	}
	unknown := func(name string) tftypes.Value {
		return tftypes.NewValue(objectType.AttributeTypes[name], tftypes.UnknownValue)
	}

	testCases := map[string]struct {
		values    map[string]tftypes.Value
		expectErr bool
	}{
		"unknown resource_provider_configs": {
			values: map[string]tftypes.Value{
				"client_secret":             tftypes.NewValue(tftypes.String, "secret"),
				"resource_provider_configs": unknown("resource_provider_configs"),
			},
		},
		"unknown secretless_auth_enabled without client_secret": {
			values: map[string]tftypes.Value{
				"secretless_auth_enabled": unknown("secretless_auth_enabled"),
			},
		},
		"unknown client_secret": {
			values: map[string]tftypes.Value{
				"client_secret": unknown("client_secret"),
			},
		},
		"secretless_auth_enabled true without client_secret": {
			values: map[string]tftypes.Value{
				"secretless_auth_enabled": tftypes.NewValue(tftypes.Bool, true),
			},
		},
		"missing client_secret": {
			values:    map[string]tftypes.Value{},
			expectErr: true,
		},
		"secretless_auth_enabled false without client_secret": {
			values: map[string]tftypes.Value{
				"secretless_auth_enabled": tftypes.NewValue(tftypes.Bool, false),
			},
			expectErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var response resource.ValidateConfigResponse
			r.ValidateConfig(ctx, resource.ValidateConfigRequest{Config: newConfig(tc.values)}, &response)

			if response.Diagnostics.HasError() != tc.expectErr {
				t.Fatalf("expected error: %v, got diagnostics: %v", tc.expectErr, response.Diagnostics)
			}
		})
	}
}
