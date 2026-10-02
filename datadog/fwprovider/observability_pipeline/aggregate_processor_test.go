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

	payload, err := json.Marshal(item.ObservabilityPipelineAggregateProcessor)
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

	model := FlattenAggregateProcessor(&processor)
	require.NotNil(t, model)
	require.Len(t, model.AggregationTiming, 1)
	assert.Equal(t, "event_time", model.AggregationTiming[0].Type.ValueString())
	assert.Equal(t, int64(15), model.AggregationTiming[0].AllowedLatenessSecs.ValueInt64())
}
