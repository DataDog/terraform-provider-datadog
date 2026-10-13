package fwprovider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/fwprovider/observability_pipeline"
)

func TestGenerateDatadogMetricsProcessorMetricAggregationRoundTrip(t *testing.T) {
	ctx := context.Background()
	common := observability_pipeline.BaseProcessorFields{Id: "generate-metrics", Enabled: true, Include: "*"}

	for _, tc := range []struct {
		name    string
		enabled types.Bool
		want    bool
	}{
		{name: "default", enabled: types.BoolNull(), want: true},
		{name: "disabled", enabled: types.BoolValue(false), want: false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			item := expandGenerateMetricsProcessorItem(ctx, common, &generateMetricsProcessorModel{
				EnabledMetricAggregation: tc.enabled,
			})
			processor := item.ObservabilityPipelineGenerateMetricsProcessor
			require.NotNil(t, processor)
			assert.True(t, processor.HasEnabledMetricAggregation())
			assert.Equal(t, tc.want, processor.GetEnabledMetricAggregation())

			model := flattenGenerateDatadogMetricsProcessor(ctx, processor)
			require.Len(t, model.GenerateMetricsProcessor, 1)
			assert.Equal(t, tc.want, model.GenerateMetricsProcessor[0].EnabledMetricAggregation.ValueBool())
		})
	}
}

func TestGenerateMetricsV2ProcessorMetricAggregationRoundTrip(t *testing.T) {
	ctx := context.Background()
	common := observability_pipeline.BaseProcessorFields{Id: "generate-metrics-v2", Enabled: true, Include: "*"}

	for _, tc := range []struct {
		name    string
		enabled types.Bool
		want    bool
	}{
		{name: "default", enabled: types.BoolNull(), want: true},
		{name: "disabled", enabled: types.BoolValue(false), want: false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			item := expandGenerateMetricsV2ProcessorItem(ctx, common, &generateMetricsProcessorModel{
				EnabledMetricAggregation: tc.enabled,
			})
			processor := item.ObservabilityPipelineGenerateMetricsV2Processor
			require.NotNil(t, processor)
			assert.True(t, processor.HasEnabledMetricAggregation())
			assert.Equal(t, tc.want, processor.GetEnabledMetricAggregation())

			model := flattenGenerateMetricsV2Processor(ctx, processor)
			require.Len(t, model.GenerateMetricsV2Processor, 1)
			assert.Equal(t, tc.want, model.GenerateMetricsV2Processor[0].EnabledMetricAggregation.ValueBool())
		})
	}
}
