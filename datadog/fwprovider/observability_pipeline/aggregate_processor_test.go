package observability_pipeline

import (
	"encoding/json"
	"testing"

	datadogV2 "github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExpandAggregateProcessorWithEventTime(t *testing.T) {
	item := ExpandAggregateProcessor(BaseProcessorFields{
		Id:      "aggregate-processor",
		Enabled: true,
		Include: "*",
	}, &AggregateProcessorModel{
		IntervalSecs: types.Int64Value(10),
		Mode:         types.StringValue("sum"),
		AggregationTiming: []*AggregateAggregationTimingModel{{
			Type:                types.StringValue("event_time"),
			AllowedLatenessSecs: types.Int64Value(15),
		}},
	})

	processor := item.ObservabilityPipelineAggregateProcessor
	require.NotNil(t, processor.AggregationTiming)
	assert.Empty(t, processor.AdditionalProperties)
	assert.Equal(t, "event_time", string(processor.AggregationTiming.GetType()))
	assert.Equal(t, int64(15), processor.AggregationTiming.GetAllowedLatenessSecs())

	payload, err := json.Marshal(processor)
	require.NoError(t, err)
	assert.JSONEq(t, `{
		"id": "aggregate-processor",
		"type": "aggregate",
		"include": "*",
		"enabled": true,
		"interval_secs": 10,
		"mode": "sum",
		"aggregation_timing": {
			"type": "event_time",
			"allowed_lateness_secs": 15
		}
	}`, string(payload))
}

func TestFlattenAggregateProcessorWithEventTime(t *testing.T) {
	var processor datadogV2.ObservabilityPipelineAggregateProcessor
	err := json.Unmarshal([]byte(`{
		"id": "aggregate-processor",
		"type": "aggregate",
		"include": "*",
		"enabled": true,
		"interval_secs": 10,
		"mode": "sum",
		"aggregation_timing": {
			"type": "event_time",
			"allowed_lateness_secs": 15
		}
	}`), &processor)
	require.NoError(t, err)

	require.NotNil(t, processor.AggregationTiming)
	assert.Empty(t, processor.AdditionalProperties)

	model := FlattenAggregateProcessor(&processor)
	require.NotNil(t, model)
	require.Len(t, model.AggregationTiming, 1)
	assert.Equal(t, "event_time", model.AggregationTiming[0].Type.ValueString())
	assert.Equal(t, int64(15), model.AggregationTiming[0].AllowedLatenessSecs.ValueInt64())
}

func TestAggregateProcessorTimingRoundTrip(t *testing.T) {
	for _, tc := range []struct {
		name     string
		timing   string
		lateness types.Int64
	}{
		{name: "omitted"},
		{name: "system_time", timing: "system_time", lateness: types.Int64Null()},
		{name: "event_time_default_lateness", timing: "event_time", lateness: types.Int64Null()},
		{name: "event_time_unknown_lateness", timing: "event_time", lateness: types.Int64Unknown()},
		{name: "event_time_zero_lateness", timing: "event_time", lateness: types.Int64Value(0)},
		{name: "event_time_max_lateness", timing: "event_time", lateness: types.Int64Value(3600)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			model := &AggregateProcessorModel{
				IntervalSecs: types.Int64Value(10),
				Mode:         types.StringValue("sum"),
			}
			if tc.timing != "" {
				model.AggregationTiming = []*AggregateAggregationTimingModel{{
					Type:                types.StringValue(tc.timing),
					AllowedLatenessSecs: tc.lateness,
				}}
			}

			item := ExpandAggregateProcessor(BaseProcessorFields{Id: "aggregate", Include: "*", Enabled: true}, model)
			processor := item.ObservabilityPipelineAggregateProcessor
			assert.Empty(t, processor.AdditionalProperties)
			payload, err := json.Marshal(processor)
			require.NoError(t, err)
			var decoded datadogV2.ObservabilityPipelineAggregateProcessor
			require.NoError(t, json.Unmarshal(payload, &decoded))
			assert.Empty(t, decoded.UnparsedObject)
			assert.Empty(t, decoded.AdditionalProperties)

			flattened := FlattenAggregateProcessor(&decoded)
			if tc.timing == "" {
				assert.False(t, processor.HasAggregationTiming())
				assert.Nil(t, flattened.AggregationTiming)
				return
			}
			require.True(t, processor.HasAggregationTiming())
			require.Len(t, flattened.AggregationTiming, 1)
			assert.Equal(t, model.AggregationTiming[0].Type, flattened.AggregationTiming[0].Type)
			if tc.lateness.IsNull() || tc.lateness.IsUnknown() {
				assert.False(t, processor.AggregationTiming.HasAllowedLatenessSecs())
				assert.True(t, flattened.AggregationTiming[0].AllowedLatenessSecs.IsNull())
			} else {
				assert.True(t, processor.AggregationTiming.HasAllowedLatenessSecs())
				assert.Equal(t, tc.lateness, flattened.AggregationTiming[0].AllowedLatenessSecs)
			}
		})
	}
}
