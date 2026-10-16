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

	flattened := flattenAmazonOpenSearchDestination(&apiDest)
	require.Equal(t, src, flattened)
}
