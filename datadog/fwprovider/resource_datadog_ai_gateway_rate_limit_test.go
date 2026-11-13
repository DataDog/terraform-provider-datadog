package fwprovider

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadog"
	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/google/uuid"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestBuildAiGatewayRateLimitUpdateAttributes(t *testing.T) {
	const teamID = "55555555-5555-5555-5555-555555555555"
	ctx := context.Background()

	tests := []struct {
		name  string
		state aiGatewayRateLimitModel
		want  map[string]interface{}
	}{
		{
			name: "removed fields are sent as their clear values",
			state: aiGatewayRateLimitModel{
				Name:                  types.StringValue("coding"),
				Mode:                  types.StringValue("ENFORCE"),
				Enabled:               types.BoolValue(true),
				RequestsPerMinute:     types.Int64Value(100),
				InputTokensPerMinute:  types.Int64Null(),
				OutputTokensPerMinute: types.Int64Null(),
				TeamID:                types.StringNull(),
				UserID:                types.StringNull(),
				Tags:                  types.MapNull(types.StringType),
			},
			want: map[string]interface{}{
				"name":                     "coding",
				"mode":                     "ENFORCE",
				"enabled":                  true,
				"requests_per_minute":      float64(100),
				"input_tokens_per_minute":  float64(0),
				"output_tokens_per_minute": float64(0),
				"team_id":                  "",
				"user_id":                  "",
				"tags":                     map[string]interface{}{},
			},
		},
		{
			name: "configured fields are sent as configured",
			state: aiGatewayRateLimitModel{
				Name:                  types.StringValue("coding"),
				Mode:                  types.StringUnknown(),
				Enabled:               types.BoolValue(false),
				RequestsPerMinute:     types.Int64Null(),
				InputTokensPerMinute:  types.Int64Value(5000),
				OutputTokensPerMinute: types.Int64Value(1000),
				TeamID:                types.StringValue(teamID),
				UserID:                types.StringNull(),
				Tags:                  types.MapValueMust(types.StringType, map[string]attr.Value{"provider": types.StringValue("openai")}),
			},
			want: map[string]interface{}{
				"name":                     "coding",
				"enabled":                  false,
				"requests_per_minute":      float64(0),
				"input_tokens_per_minute":  float64(5000),
				"output_tokens_per_minute": float64(1000),
				"team_id":                  teamID,
				"user_id":                  "",
				"tags":                     map[string]interface{}{"provider": "openai"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			attributes, diags := buildAiGatewayRateLimitUpdateAttributes(ctx, &tt.state)
			if diags.HasError() {
				t.Fatalf("unexpected diagnostics: %v", diags)
			}
			body, err := json.Marshal(attributes)
			if err != nil {
				t.Fatalf("marshal: %v", err)
			}
			var got map[string]interface{}
			if err := json.Unmarshal(body, &got); err != nil {
				t.Fatalf("unmarshal: %v", err)
			}
			gotJSON, _ := json.Marshal(got)
			wantJSON, _ := json.Marshal(tt.want)
			if string(gotJSON) != string(wantJSON) {
				t.Errorf("request body\n got: %s\nwant: %s", gotJSON, wantJSON)
			}
		})
	}
}

func TestBuildAiGatewayRateLimitUpdateAttributesRejectsInvalidTeamID(t *testing.T) {
	state := aiGatewayRateLimitModel{
		Name:                  types.StringValue("coding"),
		Mode:                  types.StringNull(),
		Enabled:               types.BoolValue(true),
		RequestsPerMinute:     types.Int64Value(1),
		InputTokensPerMinute:  types.Int64Null(),
		OutputTokensPerMinute: types.Int64Null(),
		TeamID:                types.StringValue("not-a-uuid"),
		UserID:                types.StringNull(),
		Tags:                  types.MapNull(types.StringType),
	}
	if _, diags := buildAiGatewayRateLimitUpdateAttributes(context.Background(), &state); !diags.HasError() {
		t.Fatal("expected an error for a team_id that is not a UUID")
	}
}

func TestUpdateAiGatewayRateLimitStateKeepsUUIDSpelling(t *testing.T) {
	const upper = "BBBBBBBB-BBBB-4BBB-8BBB-BBBBBBBBBBBB"
	teamID := uuid.MustParse(upper)
	attributes := datadogV2.NewAiGatewayRateLimitAttributesWithDefaults()
	attributes.TeamId = *datadog.NewNullableUUID(&teamID)
	data := *datadogV2.NewAiGatewayRateLimitResource(*attributes, uuid.New(), datadogV2.AIGATEWAYRATELIMITTYPE_RATE_LIMITS)

	state := aiGatewayRateLimitModel{TeamID: types.StringValue(upper), UserID: types.StringNull(), Tags: types.MapNull(types.StringType)}
	if diags := updateAiGatewayRateLimitState(context.Background(), &state, data); diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}
	if got := state.TeamID.ValueString(); got != upper {
		t.Errorf("team_id = %q, want the configured spelling %q", got, upper)
	}

	imported := aiGatewayRateLimitModel{TeamID: types.StringNull(), UserID: types.StringNull(), Tags: types.MapNull(types.StringType)}
	updateAiGatewayRateLimitState(context.Background(), &imported, data)
	if got := imported.TeamID.ValueString(); got != teamID.String() {
		t.Errorf("imported team_id = %q, want the API's %q", got, teamID.String())
	}
}

func TestValidateAiGatewayRateLimitTagKey(t *testing.T) {
	for key, valid := range map[string]bool{
		"provider":        true,
		"ml_app":          true,
		"team-name":       true,
		"Provider":        false,
		" provider":       false,
		"provider ":       false,
		"x-dd-tag-ml_app": false,
		"":                false,
	} {
		if err := validateAiGatewayRateLimitTagKey(key); (err == nil) != valid {
			t.Errorf("%q: err = %v, want valid = %t", key, err, valid)
		}
	}
}
