package fwprovider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestWorkflowAutomationValidateConfigUnknownTags(t *testing.T) {
	ctx := context.Background()
	r := &workflowAutomationResource{}
	var schemaResponse resource.SchemaResponse
	r.Schema(ctx, resource.SchemaRequest{}, &schemaResponse)
	configType := schemaResponse.Schema.Type().TerraformType(ctx).(tftypes.Object)
	runAsType := configType.AttributeTypes["run_as"]
	runAs := func(identityType, id any) tftypes.Value {
		return tftypes.NewValue(runAsType, map[string]tftypes.Value{
			"type": tftypes.NewValue(tftypes.String, identityType),
			"id":   tftypes.NewValue(tftypes.String, id),
		})
	}

	tests := []struct {
		name    string
		runAs   tftypes.Value
		wantErr string
	}{
		{name: "omitted run_as", runAs: tftypes.NewValue(runAsType, nil)},
		{name: "unknown run_as", runAs: tftypes.NewValue(runAsType, tftypes.UnknownValue)},
		{name: "unknown identity type", runAs: runAs(tftypes.UnknownValue, nil)},
		{name: "owner", runAs: runAs("owner", nil)},
		{name: "initiator", runAs: runAs("initiator", nil)},
		{name: "service account", runAs: runAs("service_account", "11111111-2222-3333-4444-555555555555")},
		{name: "unknown service account ID", runAs: runAs("service_account", tftypes.UnknownValue)},
		{name: "missing identity type", runAs: runAs(nil, nil), wantErr: "Missing run_as type"},
		{name: "missing service account ID", runAs: runAs("service_account", nil), wantErr: "Missing run_as service account ID"},
		{name: "empty service account ID", runAs: runAs("service_account", ""), wantErr: "Missing run_as service account ID"},
		{name: "malformed service account ID", runAs: runAs("service_account", "invalid"), wantErr: "Invalid run_as service account ID"},
		{name: "owner with ID", runAs: runAs("owner", "11111111-2222-3333-4444-555555555555"), wantErr: "Invalid run_as ID"},
		{name: "initiator with ID", runAs: runAs("initiator", "11111111-2222-3333-4444-555555555555"), wantErr: "Invalid run_as ID"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			attributes := make(map[string]tftypes.Value, len(configType.AttributeTypes))
			for name, attributeType := range configType.AttributeTypes {
				attributes[name] = tftypes.NewValue(attributeType, nil)
			}
			// Tags passed through module variables can be unknown during validation.
			attributes["tags"] = tftypes.NewValue(configType.AttributeTypes["tags"], tftypes.UnknownValue)
			attributes["run_as"] = tt.runAs
			config := tfsdk.Config{
				Raw:    tftypes.NewValue(configType, attributes),
				Schema: schemaResponse.Schema,
			}
			var response resource.ValidateConfigResponse
			r.ValidateConfig(ctx, resource.ValidateConfigRequest{Config: config}, &response)
			errors := response.Diagnostics.Errors()
			if tt.wantErr == "" {
				if len(errors) != 0 {
					t.Fatalf("unexpected validation errors with unknown tags: %v", errors)
				}
				return
			}
			if len(errors) != 1 || errors[0].Summary() != tt.wantErr {
				t.Fatalf("expected %q, got %v", tt.wantErr, errors)
			}
		})
	}
}
