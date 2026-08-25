package fwprovider

import (
	"context"
	"testing"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestExpandPipelineEndToEndAcknowledgements(t *testing.T) {
	tests := []struct {
		name      string
		value     types.Bool
		want      bool
		wantIsSet bool
	}{
		{name: "unset", value: types.BoolNull()},
		{name: "unknown", value: types.BoolUnknown()},
		{name: "false", value: types.BoolValue(false), wantIsSet: true},
		{name: "true", value: types.BoolValue(true), want: true, wantIsSet: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			state := observabilityPipelineModel{
				Name: types.StringValue("pipeline"),
				Config: []configModel{{
					PipelineType:             types.StringValue("logs"),
					EndToEndAcknowledgements: test.value,
				}},
			}

			pipeline, diags := expandPipeline(context.Background(), &state)
			if diags.HasError() {
				t.Fatalf("unexpected diagnostics: %v", diags)
			}

			value, isSet := pipeline.Data.Attributes.Config.GetEndToEndAcknowledgementsOk()
			if isSet != test.wantIsSet {
				t.Fatalf("expected isSet %t, got %t", test.wantIsSet, isSet)
			}
			if isSet && *value != test.want {
				t.Fatalf("expected value %t, got %t", test.want, *value)
			}
		})
	}
}

func TestFlattenPipelineEndToEndAcknowledgements(t *testing.T) {
	tests := []struct {
		name      string
		value     bool
		isSet     bool
		wantValue types.Bool
	}{
		{name: "unset", wantValue: types.BoolNull()},
		{name: "false", isSet: true, wantValue: types.BoolValue(false)},
		{name: "true", value: true, isSet: true, wantValue: types.BoolValue(true)},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			config := datadogV2.NewObservabilityPipelineConfig(nil, nil)
			if test.isSet {
				config.SetEndToEndAcknowledgements(test.value)
			}
			attributes := datadogV2.NewObservabilityPipelineDataAttributes(*config, "pipeline")
			data := datadogV2.NewObservabilityPipelineData(*attributes, "pipeline-id", "pipelines")
			pipeline := datadogV2.NewObservabilityPipeline(*data)
			state := observabilityPipelineModel{}

			flattenPipeline(context.Background(), &state, pipeline)

			if len(state.Config) != 1 {
				t.Fatalf("expected one config block, got %d", len(state.Config))
			}
			if !state.Config[0].EndToEndAcknowledgements.Equal(test.wantValue) {
				t.Fatalf("expected %s, got %s", test.wantValue, state.Config[0].EndToEndAcknowledgements)
			}
		})
	}
}
