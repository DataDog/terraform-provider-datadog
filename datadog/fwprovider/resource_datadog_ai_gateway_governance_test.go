package fwprovider

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestBuildAiGatewayGovernanceAttributes(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name  string
		state aiGatewayGovernanceModel
		want  map[string]interface{}
	}{
		{
			// The API keeps an omitted field, so an unset zdr_mode never changes the organization's requirement.
			name: "unset fields are omitted",
			state: aiGatewayGovernanceModel{
				DataCaptureMode: types.StringValue("metadata_only"),
				MetadataCapture: types.SetUnknown(types.StringType),
				ZdrMode:         types.StringUnknown(),
			},
			want: map[string]interface{}{"data_capture_mode": "metadata_only"},
		},
		{
			name: "an empty set is sent to clear the items",
			state: aiGatewayGovernanceModel{
				DataCaptureMode: types.StringValue("metadata_only"),
				MetadataCapture: types.SetValueMust(types.StringType, []attr.Value{}),
				ZdrMode:         types.StringNull(),
			},
			want: map[string]interface{}{"data_capture_mode": "metadata_only", "metadata_capture": []interface{}{}},
		},
		{
			name: "configured fields are sent as configured",
			state: aiGatewayGovernanceModel{
				DataCaptureMode: types.StringValue("full_capture"),
				MetadataCapture: types.SetValueMust(types.StringType, []attr.Value{types.StringValue("tool_usage")}),
				ZdrMode:         types.StringValue("required"),
			},
			want: map[string]interface{}{"data_capture_mode": "full_capture", "metadata_capture": []interface{}{"tool_usage"}, "zdr_mode": "required"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			attributes, diags := buildAiGatewayGovernanceAttributes(ctx, &tt.state)
			if diags.HasError() {
				t.Fatalf("unexpected diagnostics: %v", diags)
			}
			body, err := json.Marshal(attributes)
			if err != nil {
				t.Fatal(err)
			}
			var got map[string]interface{}
			if err := json.Unmarshal(body, &got); err != nil {
				t.Fatal(err)
			}
			gotJSON, _ := json.Marshal(got)
			wantJSON, _ := json.Marshal(tt.want)
			if string(gotJSON) != string(wantJSON) {
				t.Errorf("got %s, want %s", gotJSON, wantJSON)
			}
		})
	}
}
