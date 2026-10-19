package fwprovider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestStatusPageComponentValidateConfigSkipsWhenNotFullyKnown(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	r := &statusPageComponentResource{}

	var schemaResponse resource.SchemaResponse
	r.Schema(ctx, resource.SchemaRequest{}, &schemaResponse)

	config := tfsdk.Config{
		Raw:    tftypes.NewValue(schemaResponse.Schema.Type().TerraformType(ctx), tftypes.UnknownValue),
		Schema: schemaResponse.Schema,
	}

	var response resource.ValidateConfigResponse
	r.ValidateConfig(ctx, resource.ValidateConfigRequest{Config: config}, &response)

	if response.Diagnostics.HasError() {
		t.Fatalf("ValidateConfig returned diagnostics for a not-fully-known config: %v", response.Diagnostics.Errors())
	}
}
