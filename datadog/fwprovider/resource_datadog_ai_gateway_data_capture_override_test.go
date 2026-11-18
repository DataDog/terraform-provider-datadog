package fwprovider

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestBuildAiGatewayDataCaptureMetadata(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name       string
		configured types.Set
		want       string
	}{
		// The API keeps omitted items, so an unset metadata_capture must not be sent.
		{"unset items are omitted", types.SetNull(types.StringType), "null"},
		{"an empty set is sent to clear the items", types.SetValueMust(types.StringType, []attr.Value{}), "[]"},
		{"configured items are sent", types.SetValueMust(types.StringType, []attr.Value{types.StringValue("tool_usage")}), `["tool_usage"]`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metadata, diags := buildAiGatewayDataCaptureMetadata(ctx, tt.configured)
			if diags.HasError() {
				t.Fatalf("unexpected diagnostics: %v", diags)
			}
			got, err := json.Marshal(metadata)
			if err != nil {
				t.Fatal(err)
			}
			if string(got) != tt.want {
				t.Errorf("got %s, want %s", got, tt.want)
			}
		})
	}
}

func TestAiGatewayDataCaptureScopeEqual(t *testing.T) {
	const teamID = "55555555-5555-5555-5555-555555555555"
	tags := types.MapValueMust(types.StringType, map[string]attr.Value{"ml_app": types.StringValue("coding")})
	model := func(teamID types.String, tags types.Map) *aiGatewayDataCaptureOverrideModel {
		return &aiGatewayDataCaptureOverrideModel{TeamID: teamID, UserID: types.StringNull(), Tags: tags}
	}

	tests := []struct {
		name string
		a, b *aiGatewayDataCaptureOverrideModel
		want bool
	}{
		{"same scope", model(types.StringValue(teamID), tags), model(types.StringValue(teamID), tags), true},
		{"UUID spelling doesn't count as a change", model(types.StringValue("55555555-5555-5555-5555-55555555555A"), tags), model(types.StringValue("55555555-5555-5555-5555-55555555555a"), tags), true},
		{"removed team", model(types.StringNull(), tags), model(types.StringValue(teamID), tags), false},
		{"changed tags", model(types.StringValue(teamID), types.MapNull(types.StringType)), model(types.StringValue(teamID), tags), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := aiGatewayDataCaptureScopeEqual(tt.a, tt.b); got != tt.want {
				t.Errorf("got %t, want %t", got, tt.want)
			}
		})
	}
}
