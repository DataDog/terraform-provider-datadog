package observability_pipeline

import (
	"encoding/json"
	"testing"

	datadogV2 "github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"
)

func TestCustomProcessorSampleRoundTrip(t *testing.T) {
	model := &CustomProcessorModel{
		Remaps: []CustomProcessorRemapModel{{
			Include:     types.StringValue("*"),
			Name:        types.StringValue("Parse message"),
			Enabled:     types.BoolValue(true),
			Source:      types.StringValue(". = parse_json!(.message)"),
			Sample:      types.StringValue(`{"message":"{\"service\":\"api\"}"}`),
			DropOnError: types.BoolValue(false),
		}},
	}

	item := ExpandCustomProcessor(BaseProcessorFields{
		Id:      "custom-processor",
		Include: "*",
		Enabled: true,
	}, model)
	processor := item.ObservabilityPipelineCustomProcessor
	require.NotNil(t, processor)

	payload, err := json.Marshal(processor)
	require.NoError(t, err)
	require.JSONEq(t, `{
		"id": "custom-processor",
		"type": "custom_processor",
		"include": "*",
		"enabled": true,
		"remaps": [{
			"include": "*",
			"name": "Parse message",
			"enabled": true,
			"source": ". = parse_json!(.message)",
			"sample": "{\"message\":\"{\\\"service\\\":\\\"api\\\"}\"}",
			"drop_on_error": false
		}]
	}`, string(payload))

	var apiProcessor datadogV2.ObservabilityPipelineCustomProcessor
	require.NoError(t, json.Unmarshal(payload, &apiProcessor))
	flattened := FlattenCustomProcessor(&apiProcessor)
	require.Equal(t, model, flattened)
}
