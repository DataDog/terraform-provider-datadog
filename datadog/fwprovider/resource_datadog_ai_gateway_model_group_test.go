package fwprovider

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func aiGatewayModelGroupTestJSON(t *testing.T, value interface{}) string {
	t.Helper()
	body, err := json.Marshal(value)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	// Round-trip through a map so keys are compared in a stable order.
	var decoded map[string]interface{}
	if err := json.Unmarshal(body, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	normalized, _ := json.Marshal(decoded)
	return string(normalized)
}

func TestBuildAiGatewayModelGroupUpdateAttributes(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name  string
		state aiGatewayModelGroupModel
		want  map[string]interface{}
	}{
		{
			name: "removed fields are sent as their clear values",
			state: aiGatewayModelGroupModel{
				ID:                types.StringValue("7"),
				Name:              types.StringValue("coding"),
				ModelIDs:          types.SetNull(types.Int64Type),
				ProviderIDs:       types.SetNull(types.Int64Type),
				RoutingRuleSlugs:  types.SetNull(types.StringType),
				AllowFastMode:     types.BoolValue(true),
				MaxReasoningLevel: types.StringNull(),
			},
			want: map[string]interface{}{
				"name":                "coding",
				"allow_fast_mode":     true,
				"model_ids":           []interface{}{},
				"provider_ids":        []interface{}{},
				"routing_rule_slugs":  []interface{}{},
				"max_reasoning_level": nil,
			},
		},
		{
			name: "fields configured as empty sets are sent as []",
			state: aiGatewayModelGroupModel{
				ID:                types.StringValue("7"),
				Name:              types.StringValue("coding"),
				ModelIDs:          types.SetValueMust(types.Int64Type, []attr.Value{}),
				ProviderIDs:       types.SetValueMust(types.Int64Type, []attr.Value{}),
				RoutingRuleSlugs:  types.SetValueMust(types.StringType, []attr.Value{}),
				AllowFastMode:     types.BoolValue(false),
				MaxReasoningLevel: types.StringNull(),
			},
			want: map[string]interface{}{
				"name":                "coding",
				"allow_fast_mode":     false,
				"model_ids":           []interface{}{},
				"provider_ids":        []interface{}{},
				"routing_rule_slugs":  []interface{}{},
				"max_reasoning_level": nil,
			},
		},
		{
			name: "configured fields are sent as configured",
			state: aiGatewayModelGroupModel{
				ID:                types.StringValue("7"),
				Name:              types.StringValue("coding"),
				ModelIDs:          types.SetValueMust(types.Int64Type, []attr.Value{types.Int64Value(101)}),
				ProviderIDs:       types.SetValueMust(types.Int64Type, []attr.Value{types.Int64Value(3)}),
				RoutingRuleSlugs:  types.SetValueMust(types.StringType, []attr.Value{types.StringValue("coding-route")}),
				AllowFastMode:     types.BoolValue(false),
				MaxReasoningLevel: types.StringValue("high"),
			},
			want: map[string]interface{}{
				"name":                "coding",
				"allow_fast_mode":     false,
				"model_ids":           []interface{}{float64(101)},
				"provider_ids":        []interface{}{float64(3)},
				"routing_rule_slugs":  []interface{}{"coding-route"},
				"max_reasoning_level": "high",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			attributes, diags := buildAiGatewayModelGroupUpdateAttributes(ctx, &tt.state)
			if diags.HasError() {
				t.Fatalf("unexpected diagnostics: %v", diags)
			}
			got := aiGatewayModelGroupTestJSON(t, attributes)
			wantJSON, _ := json.Marshal(tt.want)
			if got != string(wantJSON) {
				t.Errorf("request body\n got: %s\nwant: %s", got, wantJSON)
			}
		})
	}
}

func TestBuildAiGatewayModelGroupCreateAttributes(t *testing.T) {
	state := aiGatewayModelGroupModel{
		Name:              types.StringValue("coding"),
		ModelIDs:          types.SetNull(types.Int64Type),
		ProviderIDs:       types.SetValueMust(types.Int64Type, []attr.Value{types.Int64Value(3)}),
		RoutingRuleSlugs:  types.SetNull(types.StringType),
		AllowFastMode:     types.BoolValue(false),
		MaxReasoningLevel: types.StringNull(),
	}
	attributes, diags := buildAiGatewayModelGroupCreateAttributes(context.Background(), &state)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}
	got := aiGatewayModelGroupTestJSON(t, attributes)
	want := `{"allow_fast_mode":false,"name":"coding","provider_ids":[3]}`
	if got != want {
		t.Errorf("request body\n got: %s\nwant: %s", got, want)
	}
}

func TestUpdateAiGatewayModelGroupState(t *testing.T) {
	ctx := context.Background()
	var response datadogV2.AiGatewayModelGroupResponse
	body := `{"data":{"id":"7","type":"model-groups","attributes":{
		"name":"coding","allow_fast_mode":false,"max_reasoning_level":null,
		"model_ids":[101,102],"models":[],"provider_ids":[],"providers":[],
		"routing_rules":[{"slug":"coding-route"}],
		"created_at":"2026-11-01T00:00:00Z","updated_at":"2026-11-01T00:00:00Z"}}}`
	if err := json.Unmarshal([]byte(body), &response); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	tests := []struct {
		name          string
		priorProvider types.Set
		wantProvider  types.Set
	}{
		{"an omitted selection that the API returns empty is null", types.SetNull(types.Int64Type), types.SetNull(types.Int64Type)},
		{"a selection configured as empty stays empty", types.SetValueMust(types.Int64Type, []attr.Value{}), types.SetValueMust(types.Int64Type, []attr.Value{})},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state := aiGatewayModelGroupModel{
				ModelIDs:          types.SetNull(types.Int64Type),
				ProviderIDs:       tt.priorProvider,
				RoutingRuleSlugs:  types.SetNull(types.StringType),
				MaxReasoningLevel: types.StringValue("high"),
			}
			if diags := updateAiGatewayModelGroupState(ctx, &state, response.GetData()); diags.HasError() {
				t.Fatalf("unexpected diagnostics: %v", diags)
			}
			want := aiGatewayModelGroupModel{
				ID:                types.StringValue("7"),
				Name:              types.StringValue("coding"),
				ModelIDs:          types.SetValueMust(types.Int64Type, []attr.Value{types.Int64Value(101), types.Int64Value(102)}),
				ProviderIDs:       tt.wantProvider,
				RoutingRuleSlugs:  types.SetValueMust(types.StringType, []attr.Value{types.StringValue("coding-route")}),
				AllowFastMode:     types.BoolValue(false),
				MaxReasoningLevel: types.StringNull(),
			}
			for _, field := range []struct {
				name      string
				got, want attr.Value
			}{
				{"id", state.ID, want.ID},
				{"name", state.Name, want.Name},
				{"model_ids", state.ModelIDs, want.ModelIDs},
				{"provider_ids", state.ProviderIDs, want.ProviderIDs},
				{"routing_rule_slugs", state.RoutingRuleSlugs, want.RoutingRuleSlugs},
				{"allow_fast_mode", state.AllowFastMode, want.AllowFastMode},
				{"max_reasoning_level", state.MaxReasoningLevel, want.MaxReasoningLevel},
			} {
				if !field.got.Equal(field.want) {
					t.Errorf("%s: got %s, want %s", field.name, field.got, field.want)
				}
			}
		})
	}
}

func TestUpdateAiGatewayModelGroupStateRejectsDefaultGroup(t *testing.T) {
	var response datadogV2.AiGatewayModelGroupResponse
	body := `{"data":{"id":"default","type":"model-groups","attributes":{
		"name":"default","allow_fast_mode":true,"max_reasoning_level":null,"revision":3,
		"model_ids":[],"models":[],"provider_ids":[],"providers":[],"routing_rules":[],
		"created_at":"2026-11-01T00:00:00Z","updated_at":"2026-11-01T00:00:00Z"}}}`
	if err := json.Unmarshal([]byte(body), &response); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	state := aiGatewayModelGroupModel{
		ModelIDs:         types.SetNull(types.Int64Type),
		ProviderIDs:      types.SetNull(types.Int64Type),
		RoutingRuleSlugs: types.SetNull(types.StringType),
	}
	if diags := updateAiGatewayModelGroupState(context.Background(), &state, response.GetData()); !diags.HasError() {
		t.Fatal("expected an error for the default model group")
	}
}
