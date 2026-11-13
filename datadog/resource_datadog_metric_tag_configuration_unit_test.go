package datadog

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateTagConfigurationRetryError(t *testing.T) {
	apiErr := errors.New("boom")

	tests := []struct {
		name          string
		err           error
		httpResponse  *http.Response
		wantNil       bool // classifier returned nil (success)
		wantRetryable bool // only meaningful when wantNil == false
	}{
		{
			name:    "success (nil error) is terminal",
			err:     nil,
			wantNil: true,
		},
		{
			name:          "409 conflict is retryable",
			err:           apiErr,
			httpResponse:  &http.Response{StatusCode: http.StatusConflict},
			wantRetryable: true,
		},
		{
			name:          "400 bad request is not retryable",
			err:           apiErr,
			httpResponse:  &http.Response{StatusCode: http.StatusBadRequest},
			wantRetryable: false,
		},
		{
			name:          "500 is not retryable here",
			err:           apiErr,
			httpResponse:  &http.Response{StatusCode: http.StatusInternalServerError},
			wantRetryable: false,
		},
		{
			name:          "nil http response (network error) is not retryable",
			err:           apiErr,
			httpResponse:  nil,
			wantRetryable: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := createTagConfigurationRetryError("test.metric.name", tt.err, tt.httpResponse)
			if tt.wantNil {
				assert.Nil(t, got)
				return
			}
			assert.NotNil(t, got)
			assert.Equal(t, tt.wantRetryable, got.Retryable)
		})
	}
}

func TestMetricTagConfigurationNativeHistogram(t *testing.T) {
	for _, nativeType := range []string{"explicit_histogram", "exponential_histogram"} {
		t.Run(nativeType, func(t *testing.T) {
			resource := resourceDatadogMetricTagConfiguration()
			config := map[string]interface{}{
				"metric_name":         "fixture.histogram",
				"metric_type":         "distribution",
				"tags":                []interface{}{"environment"},
				"include_percentiles": true,
			}
			attributes := datadogV2.NewMetricTagConfigurationAttributesWithDefaults()
			attributes.SetMetricType(datadogV2.MetricTagConfigurationMetricTypes(nativeType))
			attributes.SetTags([]string{"environment"})
			attributes.SetIncludePercentiles(true)
			response := datadogV2.NewMetricTagConfiguration()
			response.SetId("fixture.histogram")
			response.SetAttributes(*attributes)

			written := schema.TestResourceDataRaw(t, resource.SchemaMap(), config)
			require.Empty(t, updateMetricTagConfigurationState(written, response, "distribution"))
			require.Equal(t, "distribution", written.Get("metric_type"))

			state := schema.TestResourceDataRaw(t, resource.SchemaMap(), config)
			require.NoError(t, state.Set("aggregations", []map[string]string{{"time": "avg", "space": "avg"}}))
			attributes.SetIncludePercentiles(false)
			response.SetAttributes(*attributes)
			require.Empty(t, updateMetricTagConfigurationState(state, response, ""))
			require.Equal(t, nativeType, state.Get("metric_type"))
			require.Equal(t, false, state.Get("include_percentiles"))
			require.Zero(t, state.Get("aggregations").(*schema.Set).Len())
			diff, err := resource.Diff(context.Background(), state.State(), terraform.NewResourceConfigRaw(config), nil)
			require.NoError(t, err)
			require.NotNil(t, diff)
			require.False(t, diff.RequiresNew())
			require.Contains(t, diff.Attributes, "include_percentiles")

			desired := schema.TestResourceDataRaw(t, resource.SchemaMap(), config)
			update, err := buildDatadogMetricTagConfigurationUpdate(desired, attributes.MetricType)
			require.NoError(t, err)
			require.True(t, update.Attributes.GetIncludePercentiles())

			attributes.SetIncludePercentiles(true)
			response.SetAttributes(*attributes)
			require.Empty(t, updateMetricTagConfigurationState(state, response, ""))
			diff, err = resource.Diff(context.Background(), state.State(), terraform.NewResourceConfigRaw(config), nil)
			require.NoError(t, err)
			require.True(t, diff.Empty())

			prior := state.State()
			prior.RawConfig, err = resource.CoreConfigSchema().CoerceValue(cty.ObjectVal(map[string]cty.Value{
				"metric_name": cty.StringVal("fixture.histogram"),
				"metric_type": cty.StringVal("gauge"),
				"tags":        cty.SetVal([]cty.Value{cty.StringVal("environment")}),
			}))
			require.NoError(t, err)
			diff, err = resource.Diff(context.Background(), prior, terraform.NewResourceConfigShimmed(prior.RawConfig, resource.CoreConfigSchema()), nil)
			require.NoError(t, err)
			require.True(t, diff.RequiresNew())
		})
	}
}
