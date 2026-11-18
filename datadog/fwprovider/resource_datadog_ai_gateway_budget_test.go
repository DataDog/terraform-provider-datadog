package fwprovider

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadog"
	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/google/uuid"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func aiGatewayBudgetTestModel() aiGatewayBudgetModel {
	return aiGatewayBudgetModel{
		ID:                   types.StringValue("11111111-1111-1111-1111-111111111111"),
		Name:                 types.StringValue("coding"),
		CapacityMicrocents:   types.Int64Value(100_000_000),
		WindowPolicy:         types.StringValue("MONTHLY"),
		Enforcement:          types.StringValue("BLOCK"),
		WarningThreshold:     types.Float64Value(0.9),
		Enabled:              types.BoolValue(true),
		ApplyToCurrentPeriod: types.BoolValue(false),
		TeamID:               types.StringNull(),
		UserID:               types.StringNull(),
		Tags:                 types.MapNull(types.StringType),
		Selector:             types.SetNull(types.ObjectType{AttrTypes: aiGatewayBudgetSelectorAttrTypes}),
		MonitorID:            types.Int64Null(),
	}
}

func marshalAiGatewayBudgetAttributes(t *testing.T, attributes interface{}) map[string]interface{} {
	t.Helper()
	raw, err := json.Marshal(attributes)
	if err != nil {
		t.Fatal(err)
	}
	got := map[string]interface{}{}
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatal(err)
	}
	return got
}

func TestBuildAiGatewayBudgetUpdateAttributes(t *testing.T) {
	tests := []struct {
		name   string
		change func(*aiGatewayBudgetModel)
		want   map[string]interface{}
	}{
		{
			name:   "an unchanged budget sends no changes",
			change: func(*aiGatewayBudgetModel) {},
			want:   map[string]interface{}{},
		},
		{
			name: "a capacity change applies from the next period by default",
			change: func(m *aiGatewayBudgetModel) {
				m.CapacityMicrocents = types.Int64Value(200_000_000)
			},
			want: map[string]interface{}{"capacity_microcents": float64(200_000_000)},
		},
		{
			name: "a capacity change with apply_to_current_period applies now",
			change: func(m *aiGatewayBudgetModel) {
				m.CapacityMicrocents = types.Int64Value(200_000_000)
				m.ApplyToCurrentPeriod = types.BoolValue(true)
			},
			want: map[string]interface{}{"capacity_microcents": float64(200_000_000), "apply_to_current_period": true},
		},
		{
			name: "apply_to_current_period alone sends no changes",
			change: func(m *aiGatewayBudgetModel) {
				m.ApplyToCurrentPeriod = types.BoolValue(true)
			},
			want: map[string]interface{}{},
		},
		{
			name: "each changed field is sent",
			change: func(m *aiGatewayBudgetModel) {
				m.Name = types.StringValue("coding-v2")
				m.Enforcement = types.StringValue("NOTIFY")
				m.WarningThreshold = types.Float64Value(0.5)
				m.Enabled = types.BoolValue(false)
			},
			want: map[string]interface{}{"name": "coding-v2", "enforcement": "NOTIFY", "warning_threshold": 0.5, "enabled": false},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prior := aiGatewayBudgetTestModel()
			state := aiGatewayBudgetTestModel()
			tt.change(&state)
			got := marshalAiGatewayBudgetAttributes(t, buildAiGatewayBudgetUpdateAttributes(&state, &prior))
			// The client always sends these two at the API's default of false.
			want := map[string]interface{}{"apply_to_current_period": false, "track_existing": false}
			for key, value := range tt.want {
				want[key] = value
			}
			tt.want = want
			if len(got) != len(tt.want) {
				t.Fatalf("got %v, want %v", got, tt.want)
			}
			for key, want := range tt.want {
				if got[key] != want {
					t.Errorf("%s = %v, want %v", key, got[key], want)
				}
			}
		})
	}
}

func TestBuildAiGatewayBudgetCreateAttributes(t *testing.T) {
	const teamID = "55555555-5555-5555-5555-555555555555"
	ctx := context.Background()
	selectorType := types.ObjectType{AttrTypes: aiGatewayBudgetSelectorAttrTypes}

	state := aiGatewayBudgetTestModel()
	state.TeamID = types.StringValue(strings.ToUpper(teamID))
	state.Tags = types.MapValueMust(types.StringType, map[string]attr.Value{"ml_app": types.StringValue("coding")})
	state.Selector = types.SetValueMust(selectorType, []attr.Value{
		types.ObjectValueMust(aiGatewayBudgetSelectorAttrTypes, map[string]attr.Value{
			"dim":       types.StringValue("model"),
			"per_value": types.BoolValue(true),
			"values":    types.SetNull(types.StringType),
		}),
		types.ObjectValueMust(aiGatewayBudgetSelectorAttrTypes, map[string]attr.Value{
			"dim":       types.StringValue("provider"),
			"per_value": types.BoolValue(false),
			"values":    types.SetValueMust(types.StringType, []attr.Value{types.StringValue("openai")}),
		}),
	})

	attributes, diags := buildAiGatewayBudgetCreateAttributes(ctx, &state)
	if diags.HasError() {
		t.Fatal(diags)
	}
	got := marshalAiGatewayBudgetAttributes(t, attributes)

	for key, want := range map[string]interface{}{
		"name":                "coding",
		"capacity_microcents": float64(100_000_000),
		"window_policy":       "MONTHLY",
		"enforcement":         "BLOCK",
		"warning_threshold":   0.9,
		"team_id":             teamID,
		"track_existing":      false,
	} {
		if got[key] != want {
			t.Errorf("%s = %v, want %v", key, got[key], want)
		}
	}
	for _, absent := range []string{"user_id", "enabled"} {
		if _, ok := got[absent]; ok {
			t.Errorf("%s is sent, want it omitted", absent)
		}
	}
	if tags := got["tags"].(map[string]interface{}); tags["ml_app"] != "coding" || len(tags) != 1 {
		t.Errorf("tags = %v", tags)
	}
	selector := got["selector"].([]interface{})
	if len(selector) != 2 {
		t.Fatalf("selector = %v, want 2 entries", selector)
	}
	for _, raw := range selector {
		entry := raw.(map[string]interface{})
		switch entry["dim"] {
		case "model":
			if entry["per_value"] != true || entry["values"] != nil {
				t.Errorf("model entry = %v, want per_value and no values", entry)
			}
		case "provider":
			if values, _ := entry["values"].([]interface{}); len(values) != 1 || values[0] != "openai" {
				t.Errorf("provider entry = %v, want values [openai]", entry)
			}
		default:
			t.Errorf("unexpected entry %v", entry)
		}
	}
}

func TestUpdateAiGatewayBudgetState(t *testing.T) {
	ctx := context.Background()
	id := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	teamID := uuid.MustParse("55555555-5555-5555-5555-555555555555")

	attributes := datadogV2.AiGatewayBudgetAttributes{
		Name:               "coding",
		CapacityMicrocents: 100_000_000,
		WindowPolicy:       datadogV2.AIGATEWAYBUDGETWINDOWPOLICY_MONTHLY,
		Enforcement:        datadogV2.AIGATEWAYBUDGETENFORCEMENT_BLOCK,
		WarningThreshold:   0.9,
		Enabled:            true,
		TeamId:             &teamID,
		Tags:               map[string]string{},
		MonitorId:          *datadog.NewNullableInt64(datadog.PtrInt64(42)),
		Selector: []datadogV2.AiGatewayBudgetSelectorEntry{
			{Dim: "model", PerValue: true},
			{Dim: "provider", Values: []string{"openai"}},
		},
	}
	data := datadogV2.AiGatewayBudgetResource{Id: id, Attributes: attributes, Type: datadogV2.AIGATEWAYBUDGETTYPE_BUDGETS}

	t.Run("an import fills config-only and computed fields", func(t *testing.T) {
		state := aiGatewayBudgetModel{ID: types.StringValue(id.String())}
		if diags := updateAiGatewayBudgetState(ctx, &state, data); diags.HasError() {
			t.Fatal(diags)
		}
		if !state.ApplyToCurrentPeriod.Equal(types.BoolValue(false)) {
			t.Errorf("apply_to_current_period = %v, want false", state.ApplyToCurrentPeriod)
		}
		if !state.MonitorID.Equal(types.Int64Value(42)) {
			t.Errorf("monitor_id = %v, want 42", state.MonitorID)
		}
		if !state.Tags.IsNull() {
			t.Errorf("tags = %v, want null for {}", state.Tags)
		}
		if !state.TeamID.Equal(types.StringValue(teamID.String())) {
			t.Errorf("team_id = %v", state.TeamID)
		}

		var entries []aiGatewayBudgetSelectorModel
		if diags := state.Selector.ElementsAs(ctx, &entries, false); diags.HasError() {
			t.Fatal(diags)
		}
		if len(entries) != 2 {
			t.Fatalf("selector = %v, want 2 entries", entries)
		}
		for _, entry := range entries {
			if entry.Dim.ValueString() == "model" && !entry.Values.IsNull() {
				t.Errorf("model values = %v, want null", entry.Values)
			}
		}
	})

	t.Run("an earlier choice and UUID spelling are kept", func(t *testing.T) {
		upper := strings.ToUpper(teamID.String())
		state := aiGatewayBudgetModel{
			ApplyToCurrentPeriod: types.BoolValue(true),
			TeamID:               types.StringValue(upper),
		}
		if diags := updateAiGatewayBudgetState(ctx, &state, data); diags.HasError() {
			t.Fatal(diags)
		}
		if !state.ApplyToCurrentPeriod.Equal(types.BoolValue(true)) {
			t.Errorf("apply_to_current_period = %v, want true", state.ApplyToCurrentPeriod)
		}
		if state.TeamID.ValueString() != upper {
			t.Errorf("team_id = %v, want %s", state.TeamID, upper)
		}
	})
}

func TestValidateAiGatewayBudgetSelectorDim(t *testing.T) {
	for dim, valid := range map[string]bool{
		"model":               true,
		"user_id":             true,
		"tag:ml_app":          true,
		"tag:":                false,
		"tag:ML_APP":          false,
		"tag:x-dd-tag-ml_app": false,
		"tag: ml_app":         false,
		"org":                 false,
		"Model":               false,
	} {
		if err := validateAiGatewayBudgetSelectorDim(dim); (err == nil) != valid {
			t.Errorf("%q: err = %v, want valid = %t", dim, err, valid)
		}
	}
}
