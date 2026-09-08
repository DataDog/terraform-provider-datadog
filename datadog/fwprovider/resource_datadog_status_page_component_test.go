package fwprovider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// A for_each/count resource can be validated by Terraform core before its instances are
// expanded, at which point each.value (and anything derived from it, like components) is
// unknown. Regression test for the Value Conversion Error reported in
// https://github.com/DataDog/terraform-provider-datadog/issues/4220.
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
