package datadog

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
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
	resource := resourceDatadogMetricTagConfiguration()
	configSchema := resource.CoreConfigSchema()

	plan := func(t *testing.T, prior *schema.ResourceData, metricType string, includePercentiles cty.Value) *terraform.InstanceDiff {
		t.Helper()
		config, err := configSchema.CoerceValue(cty.ObjectVal(map[string]cty.Value{
			"metric_name":         cty.StringVal("fixture.histogram"),
			"metric_type":         cty.StringVal(metricType),
			"tags":                cty.SetVal([]cty.Value{cty.StringVal("environment")}),
			"include_percentiles": includePercentiles,
		}))
		require.NoError(t, err)
		state := prior.State()
		state.RawConfig = config // CustomizeDiff reads the raw config.
		diff, err := resource.Diff(context.Background(), state, terraform.NewResourceConfigShimmed(config, configSchema), nil)
		require.NoError(t, err)
		return diff
	}
	requireInPlacePercentiles := func(t *testing.T, diff *terraform.InstanceDiff) {
		t.Helper()
		require.Contains(t, diff.Attributes, "include_percentiles")
		require.False(t, diff.RequiresNew())
	}

	explicit := datadogV2.METRICTAGCONFIGURATIONMETRICTYPES_EXPLICIT_HISTOGRAM
	exponential := datadogV2.METRICTAGCONFIGURATIONMETRICTYPES_EXPONENTIAL_HISTOGRAM
	for native, other := range map[datadogV2.MetricTagConfigurationMetricTypes]datadogV2.MetricTagConfigurationMetricTypes{
		explicit:    exponential,
		exponential: explicit,
	} {
		t.Run(string(native), func(t *testing.T) {
			var response datadogV2.MetricTagConfigurationResponse
			require.NoError(t, json.Unmarshal([]byte(fmt.Sprintf(
				`{"data":{"id":"fixture.histogram","type":"manage_tags","attributes":{"metric_type":%q,"tags":["environment"],"include_percentiles":true,"exclude_tags_mode":false}}}`,
				native,
			)), &response))
			require.NoError(t, utils.CheckForUnparsed(response))

			newData := func(t *testing.T, metricType string) *schema.ResourceData {
				return schema.TestResourceDataRaw(t, resource.SchemaMap(), map[string]interface{}{
					"metric_name":         "fixture.histogram",
					"metric_type":         metricType,
					"tags":                []interface{}{"environment"},
					"include_percentiles": true,
				})
			}

			t.Run("managed distribution", func(t *testing.T) {
				created := newData(t, "distribution")
				require.Empty(t, updateMetricTagConfigurationState(created, response.Data))
				require.Equal(t, "distribution", created.Get("metric_type"))

				refreshed := resource.Data(created.State())
				require.Empty(t, updateMetricTagConfigurationState(refreshed, response.Data))
				require.Equal(t, "distribution", refreshed.Get("metric_type"))

				require.True(t, plan(t, refreshed, "distribution", cty.True).Empty())
				requireInPlacePercentiles(t, plan(t, refreshed, "distribution", cty.False))
				require.True(t, plan(t, refreshed, string(native), cty.True).RequiresNew())
			})

			// Import starts with only the ID, so state records the API subtype.
			t.Run("imported subtype", func(t *testing.T) {
				imported := resource.Data(nil)
				imported.SetId("fixture.histogram")
				require.Empty(t, updateMetricTagConfigurationState(imported, response.Data))
				require.Equal(t, string(native), imported.Get("metric_type"))

				require.True(t, plan(t, imported, "distribution", cty.True).Empty())
				require.True(t, plan(t, imported, string(native), cty.True).Empty())
				requireInPlacePercentiles(t, plan(t, imported, "distribution", cty.False))
				require.True(t, plan(t, imported, string(other), cty.True).RequiresNew())
				require.True(t, plan(t, imported, "gauge", cty.NullVal(cty.Bool)).RequiresNew())
			})

			t.Run("gauge that became a native histogram", func(t *testing.T) {
				d := newData(t, "gauge")
				require.NoError(t, d.Set("aggregations", []map[string]string{{"time": "sum", "space": "sum"}}))
				require.Empty(t, updateMetricTagConfigurationState(d, response.Data))

				require.True(t, plan(t, d, "distribution", cty.True).Empty())
				_, err := buildDatadogMetricTagConfigurationUpdate(d, &native)
				require.NoError(t, err)
			})

			t.Run("requests send percentiles", func(t *testing.T) {
				create, err := buildDatadogMetricTagConfiguration(newData(t, string(native)))
				require.NoError(t, err)
				require.Equal(t, native, create.Attributes.GetMetricType())
				require.True(t, create.Attributes.GetIncludePercentiles())

				for _, enabled := range []bool{true, false} {
					d := newData(t, "distribution")
					require.NoError(t, d.Set("include_percentiles", enabled))
					update, err := buildDatadogMetricTagConfigurationUpdate(d, &native)
					require.NoError(t, err)
					require.Equal(t, &enabled, update.Attributes.IncludePercentiles)
				}
			})
		})
	}
}
