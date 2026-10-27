package observability_pipeline

import (
	"context"
	"encoding/json"
	"testing"

	datadogV2 "github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOpentelemetryMetricsDestinationClientRoundTrip(t *testing.T) {
	inputs := types.ListValueMust(types.StringType, []attr.Value{types.StringValue("metrics-source")})
	item := ExpandOpentelemetryMetricsDestination(context.Background(), "otel", inputs, &OpentelemetryMetricsDestinationModel{
		HttpClientUriKey: types.StringValue("OTLP_ENDPOINT"),
	})
	payload, err := json.Marshal(item)
	require.NoError(t, err)
	assert.JSONEq(t, `{"id":"otel","inputs":["metrics-source"],"type":"opentelemetry","http_client_uri_key":"OTLP_ENDPOINT"}`, string(payload))

	var response datadogV2.ObservabilityPipelineConfigDestinationItem
	require.NoError(t, json.Unmarshal(payload, &response))
	model := FlattenOpentelemetryMetricsDestination(response.ObservabilityPipelineOpentelemetryDestination)
	require.NotNil(t, model)
	assert.Equal(t, "OTLP_ENDPOINT", model.HttpClientUriKey.ValueString())
}
