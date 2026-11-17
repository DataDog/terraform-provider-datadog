package fwprovider

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func marshalAiGatewayProviderBody(t *testing.T, body interface{}) map[string]interface{} {
	t.Helper()
	raw, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var got map[string]interface{}
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	return got
}

func assertAiGatewayProviderJSON(t *testing.T, got, want map[string]interface{}) {
	t.Helper()
	gotJSON, _ := json.Marshal(got)
	wantJSON, _ := json.Marshal(want)
	if string(gotJSON) != string(wantJSON) {
		t.Errorf("request body\n got: %s\nwant: %s", gotJSON, wantJSON)
	}
}

func TestBuildAiGatewayProviderCreateAttributes(t *testing.T) {
	ctx := context.Background()
	integrationConfigNull := types.ObjectNull(aiGatewayProviderIntegrationConfigAttrTypes)

	tests := []struct {
		name  string
		state aiGatewayProviderModel
		want  map[string]interface{}
	}{
		{
			name: "server defaults are left out",
			state: aiGatewayProviderModel{
				Name:                 types.StringValue("production"),
				Slug:                 types.StringValue("prod-us"),
				KnownProvider:        types.StringValue("openai"),
				IntegrationAccountID: types.StringValue("account-123"),
				IntegrationConfig:    types.ObjectUnknown(aiGatewayProviderIntegrationConfigAttrTypes),
				CostMultiplier:       types.Float64Unknown(),
			},
			want: map[string]interface{}{
				"integration_account_id": "account-123",
				"known_provider":         "openai",
				"name":                   "production",
				"slug":                   "prod-us",
			},
		},
		{
			name: "configured fields are sent as configured",
			state: aiGatewayProviderModel{
				Name:                 types.StringValue("internal"),
				Slug:                 types.StringValue("internal"),
				KnownProvider:        types.StringValue("custom"),
				IntegrationAccountID: types.StringNull(),
				IntegrationConfig: types.ObjectValueMust(aiGatewayProviderIntegrationConfigAttrTypes, map[string]attr.Value{
					"project_id":          types.StringNull(),
					"location":            types.StringNull(),
					"bedrock_region":      types.StringValue("us-east-1"),
					"foundry_resource_id": types.StringNull(),
					"foundry_project":     types.StringNull(),
					"supported_operations": types.ListValueMust(types.StringType, []attr.Value{
						types.StringValue("OPENAI_RESPONSES"),
						types.StringValue("ANTHROPIC_MESSAGES"),
					}),
				}),
				CostMultiplier: types.Float64Value(1.5),
			},
			want: map[string]interface{}{
				"cost_multiplier": 1.5,
				"integration_config": map[string]interface{}{
					"bedrock_region":       "us-east-1",
					"supported_operations": []interface{}{"OPENAI_RESPONSES", "ANTHROPIC_MESSAGES"},
				},
				"known_provider": "custom",
				"name":           "internal",
				"slug":           "internal",
			},
		},
		{
			name: "computed supported_operations is left out",
			state: aiGatewayProviderModel{
				Name:                 types.StringValue("internal"),
				Slug:                 types.StringValue("internal"),
				KnownProvider:        types.StringValue("custom"),
				IntegrationAccountID: types.StringValue("account-123"),
				IntegrationConfig: types.ObjectValueMust(aiGatewayProviderIntegrationConfigAttrTypes, map[string]attr.Value{
					"project_id":           types.StringNull(),
					"location":             types.StringNull(),
					"bedrock_region":       types.StringNull(),
					"foundry_resource_id":  types.StringNull(),
					"foundry_project":      types.StringNull(),
					"supported_operations": types.ListUnknown(types.StringType),
				}),
				CostMultiplier: types.Float64Null(),
			},
			want: map[string]interface{}{
				"integration_account_id": "account-123",
				"integration_config":     map[string]interface{}{},
				"known_provider":         "custom",
				"name":                   "internal",
				"slug":                   "internal",
			},
		},
		{
			name: "null integration_config is left out",
			state: aiGatewayProviderModel{
				Name:                 types.StringValue("anthropic"),
				Slug:                 types.StringValue("anthropic"),
				KnownProvider:        types.StringValue("anthropic"),
				IntegrationAccountID: types.StringValue("account-123"),
				IntegrationConfig:    integrationConfigNull,
				CostMultiplier:       types.Float64Value(1),
			},
			want: map[string]interface{}{
				"cost_multiplier":        float64(1),
				"integration_account_id": "account-123",
				"known_provider":         "anthropic",
				"name":                   "anthropic",
				"slug":                   "anthropic",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			attributes, diags := buildAiGatewayProviderCreateAttributes(ctx, &tt.state)
			if diags.HasError() {
				t.Fatalf("unexpected diagnostics: %v", diags)
			}
			// Marshal the whole request so the check covers what goes on the wire.
			request := datadogV2.NewAiGatewayProviderCreateRequest(*datadogV2.NewAiGatewayProviderCreateData(*attributes, datadogV2.AIGATEWAYPROVIDERTYPE_PROVIDERS))
			body := marshalAiGatewayProviderBody(t, request)
			data := body["data"].(map[string]interface{})
			got := data["attributes"].(map[string]interface{})

			// initial_bindings returns HTTP 503 even when empty or null, and models are not managed here.
			for _, forbidden := range []string{"initial_bindings", "models"} {
				if _, ok := got[forbidden]; ok {
					t.Errorf("create body must not contain %q: %v", forbidden, got)
				}
			}
			assertAiGatewayProviderJSON(t, got, tt.want)
		})
	}
}

func TestBuildAiGatewayProviderUpdateAttributes(t *testing.T) {
	fullConfig := types.ObjectValueMust(aiGatewayProviderIntegrationConfigAttrTypes, map[string]attr.Value{
		"project_id":           types.StringValue("my-gcp-project"),
		"location":             types.StringValue("us-central1"),
		"bedrock_region":       types.StringNull(),
		"foundry_resource_id":  types.StringNull(),
		"foundry_project":      types.StringNull(),
		"supported_operations": types.ListNull(types.StringType),
	})

	tests := []struct {
		name  string
		state aiGatewayProviderModel
		want  map[string]interface{}
	}{
		{
			name: "only name and cost_multiplier are sent",
			state: aiGatewayProviderModel{
				ID:                   types.StringValue("42"),
				Name:                 types.StringValue("renamed"),
				Slug:                 types.StringValue("prod-us"),
				KnownProvider:        types.StringValue("gemini"),
				IntegrationAccountID: types.StringValue("account-123"),
				IntegrationConfig:    fullConfig,
				CostMultiplier:       types.Float64Value(2.5),
			},
			want: map[string]interface{}{
				"cost_multiplier": 2.5,
				"name":            "renamed",
			},
		},
		{
			name: "unknown cost_multiplier is left out",
			state: aiGatewayProviderModel{
				ID:                   types.StringValue("42"),
				Name:                 types.StringValue("renamed"),
				Slug:                 types.StringValue("prod-us"),
				KnownProvider:        types.StringValue("gemini"),
				IntegrationAccountID: types.StringValue("account-123"),
				IntegrationConfig:    fullConfig,
				CostMultiplier:       types.Float64Unknown(),
			},
			want: map[string]interface{}{
				"name": "renamed",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			attributes := buildAiGatewayProviderUpdateAttributes(&tt.state)
			request := datadogV2.NewAiGatewayProviderUpdateRequest(*datadogV2.NewAiGatewayProviderUpdateData(*attributes, tt.state.ID.ValueString(), datadogV2.AIGATEWAYPROVIDERTYPE_PROVIDERS))
			body := marshalAiGatewayProviderBody(t, request)
			data := body["data"].(map[string]interface{})
			assertAiGatewayProviderJSON(t, data["attributes"].(map[string]interface{}), tt.want)
		})
	}
}

func TestParseAiGatewayProviderID(t *testing.T) {
	if id, diags := parseAiGatewayProviderID(types.StringValue("42")); diags.HasError() || id != 42 {
		t.Fatalf("got %d, %v; want 42", id, diags)
	}
	for _, invalid := range []string{"", "abc", "0", "-1", "4.2"} {
		if _, diags := parseAiGatewayProviderID(types.StringValue(invalid)); !diags.HasError() {
			t.Errorf("expected an error for id %q", invalid)
		}
	}
}

func TestUpdateAiGatewayProviderState(t *testing.T) {
	ctx := context.Background()
	attributes := datadogV2.NewAiGatewayProviderAttributesWithDefaults()
	attributes.SetName("internal")
	attributes.SetSlug("internal")
	attributes.SetKnownProvider(datadogV2.AIGATEWAYKNOWNPROVIDER_CUSTOM)
	attributes.SetIntegrationAccountId("")
	attributes.SetCostMultiplier(1)
	config := datadogV2.NewAiGatewayProviderIntegrationConfig()
	config.SetSupportedOperations([]datadogV2.AiGatewayCustomProviderSupportedOperation{datadogV2.AIGATEWAYCUSTOMPROVIDERSUPPORTEDOPERATION_OPENAI_CHAT_COMPLETIONS})
	attributes.SetIntegrationConfig(*config)
	data := datadogV2.NewAiGatewayProviderResource(*attributes, "42", datadogV2.AIGATEWAYPROVIDERTYPE_PROVIDERS)

	var state aiGatewayProviderModel
	if diags := updateAiGatewayProviderState(ctx, &state, *data); diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}
	if state.ID.ValueString() != "42" {
		t.Errorf("id = %q, want 42", state.ID.ValueString())
	}
	if !state.IntegrationAccountID.IsNull() {
		t.Errorf("an empty integration_account_id must be null, got %v", state.IntegrationAccountID)
	}
	want := types.ObjectValueMust(aiGatewayProviderIntegrationConfigAttrTypes, map[string]attr.Value{
		"project_id":           types.StringNull(),
		"location":             types.StringNull(),
		"bedrock_region":       types.StringNull(),
		"foundry_resource_id":  types.StringNull(),
		"foundry_project":      types.StringNull(),
		"supported_operations": types.ListValueMust(types.StringType, []attr.Value{types.StringValue("OPENAI_CHAT_COMPLETIONS")}),
	})
	if !state.IntegrationConfig.Equal(want) {
		t.Errorf("integration_config = %v, want %v", state.IntegrationConfig, want)
	}

	attributes.IntegrationConfig = nil
	data = datadogV2.NewAiGatewayProviderResource(*attributes, "42", datadogV2.AIGATEWAYPROVIDERTYPE_PROVIDERS)
	if diags := updateAiGatewayProviderState(ctx, &state, *data); diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}
	if !state.IntegrationConfig.IsNull() {
		t.Errorf("a missing integration_config must be null, got %v", state.IntegrationConfig)
	}
}
