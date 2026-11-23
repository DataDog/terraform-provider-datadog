package fwprovider

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"
)

func TestAmazonOpenSearchDestinationExpandFlattenRoundTrip(t *testing.T) {
	ctx := context.Background()
	inputs, _ := types.ListValueFrom(ctx, types.StringType, []string{"source-1"})
	dest := &destinationModel{
		Id:     types.StringValue("aos-dest-1"),
		Inputs: inputs,
	}
	src := &amazonOpenSearchDestinationModel{
		BulkIndex:      types.StringNull(),
		EndpointUrlKey: types.StringValue("AMAZON_OPENSEARCH_ENDPOINT_URL"),
		Auth: []amazonOpenSearchAuthModel{{
			Strategy:    types.StringValue("basic"),
			AwsRegion:   types.StringNull(),
			AssumeRole:  types.StringNull(),
			ExternalId:  types.StringNull(),
			SessionName: types.StringNull(),
			UsernameKey: types.StringValue("AMAZON_OPENSEARCH_USERNAME"),
			PasswordKey: types.StringValue("AMAZON_OPENSEARCH_PASSWORD"),
		}},
		DataStream: []opensearchDestinationDataStreamModel{{
			Dtype:     types.StringValue("logs"),
			Dataset:   types.StringValue("{{log_types}}"),
			Namespace: types.StringValue("default"),
		}},
	}

	item := expandAmazonOpenSearchDestination(ctx, dest, src)
	body, err := json.Marshal(item.ObservabilityPipelineAmazonOpenSearchDestination)
	require.NoError(t, err)

	var payload map[string]interface{}
	require.NoError(t, json.Unmarshal(body, &payload))
	require.Equal(t, "AMAZON_OPENSEARCH_ENDPOINT_URL", payload["endpoint_url_key"])
	require.Equal(t, map[string]interface{}{
		"dtype":     "logs",
		"dataset":   "{{log_types}}",
		"namespace": "default",
	}, payload["data_stream"])
	require.Equal(t, map[string]interface{}{
		"strategy":     "basic",
		"username_key": "AMAZON_OPENSEARCH_USERNAME",
		"password_key": "AMAZON_OPENSEARCH_PASSWORD",
	}, payload["auth"])

	var apiDest datadogV2.ObservabilityPipelineAmazonOpenSearchDestination
	require.NoError(t, json.Unmarshal(body, &apiDest))
	require.Nil(t, apiDest.UnparsedObject)
	require.Empty(t, apiDest.AdditionalProperties)
	require.Empty(t, apiDest.Auth.AdditionalProperties)

	flattened := flattenAmazonOpenSearchDestination(&apiDest)
	require.Equal(t, src, flattened)
}

func TestAmazonOpenSearchDestinationFlattenWithoutAuth(t *testing.T) {
	var apiDest datadogV2.ObservabilityPipelineAmazonOpenSearchDestination
	require.NoError(t, json.Unmarshal([]byte(`{
		"id": "aos-dest-1",
		"inputs": ["source-1"],
		"type": "amazon_opensearch",
		"endpoint_url_key": "AMAZON_OPENSEARCH_ENDPOINT_URL",
		"bulk_index": "logs"
	}`), &apiDest))

	flattened := flattenAmazonOpenSearchDestination(&apiDest)
	require.Nil(t, flattened.Auth)
	require.Equal(t, types.StringValue("AMAZON_OPENSEARCH_ENDPOINT_URL"), flattened.EndpointUrlKey)
	require.Equal(t, types.StringValue("logs"), flattened.BulkIndex)
}

func TestDatadogLogsDestinationRouteNameRoundTrip(t *testing.T) {
	ctx := context.Background()
	inputs, diags := types.ListValueFrom(ctx, types.StringType, []string{"source-1"})
	require.False(t, diags.HasError())
	dest := &destinationModel{
		Id:     types.StringValue("datadog-logs-destination"),
		Inputs: inputs,
	}
	src := &datadogLogsDestinationModel{
		Routes: []datadogLogsDestinationRouteModel{{
			RouteId:   types.StringValue("route-us1"),
			Name:      types.StringValue("Production logs"),
			Include:   types.StringValue("service:api"),
			Site:      types.StringValue("us1"),
			ApiKeyKey: types.StringValue("API_KEY_IDENT"),
		}},
	}

	item := expandDatadogLogsDestination(ctx, dest, src)
	apiDestination := item.ObservabilityPipelineDatadogLogsDestination
	require.NotNil(t, apiDestination)
	payload, err := json.Marshal(apiDestination)
	require.NoError(t, err)
	require.JSONEq(t, `{
		"id": "datadog-logs-destination",
		"type": "datadog_logs",
		"inputs": ["source-1"],
		"routes": [{
			"route_id": "route-us1",
			"name": "Production logs",
			"include": "service:api",
			"site": "us1",
			"api_key_key": "API_KEY_IDENT"
		}]
	}`, string(payload))

	var destination datadogV2.ObservabilityPipelineDatadogLogsDestination
	require.NoError(t, json.Unmarshal(payload, &destination))
	flattened := flattenDatadogLogsDestination(ctx, &destination)
	require.Equal(t, src, flattened)
}
