package datadog

import (
	"testing"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV1"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestUpdateMonitorState_OnMissingData_MutualExclusion verifies that when the API
// returns on_missing_data, the Read function does NOT set notify_no_data or
// no_data_timeframe in state. These fields are declared ConflictsWith in the schema.
//
// Strategy: seed ResourceData with sentinel values for the conflicting fields.
// After updateMonitorState, check whether those sentinels were overwritten.
func TestUpdateMonitorState_OnMissingData_MutualExclusion(t *testing.T) {
	tests := []struct {
		name            string
		onMissingData   string
		notifyNoData    bool
		noDataTimeframe *int64
		// Expected behavior of Read for conflicting fields
		expectReadSetsOnMissingData   bool
		expectReadSetsNotifyNoData    bool
		expectReadSetsNoDataTimeframe bool
	}{
		{
			name:                          "on_missing_data set — Read must NOT touch notify_no_data or no_data_timeframe",
			onMissingData:                 "show_no_data",
			notifyNoData:                  false,
			noDataTimeframe:               nil,
			expectReadSetsOnMissingData:   true,
			expectReadSetsNotifyNoData:    false,
			expectReadSetsNoDataTimeframe: false,
		},
		{
			name:                          "on_missing_data empty — Read must set notify_no_data and no_data_timeframe",
			onMissingData:                 "",
			notifyNoData:                  true,
			noDataTimeframe:               int64Ptr(15),
			expectReadSetsOnMissingData:   false,
			expectReadSetsNotifyNoData:    true,
			expectReadSetsNoDataTimeframe: true,
		},
		{
			name:                          "on_missing_data resolve — Read must NOT touch notify_no_data or no_data_timeframe",
			onMissingData:                 "resolve",
			notifyNoData:                  false,
			noDataTimeframe:               nil,
			expectReadSetsOnMissingData:   true,
			expectReadSetsNotifyNoData:    false,
			expectReadSetsNoDataTimeframe: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			monitor := buildMockMonitor(tc.onMissingData, tc.notifyNoData, tc.noDataTimeframe)

			resourceSchema := resourceDatadogMonitor()

			// Seed with sentinel values that differ from the mock monitor's values.
			// If updateMonitorState incorrectly calls d.Set() for these fields,
			// the sentinels will be overwritten — proving state pollution.
			sentinelNotifyNoData := true  // mock returns false → overwrite detectable
			sentinelNoDataTimeframe := 99 // mock returns nil (→0) → overwrite detectable
			d := schema.TestResourceDataRaw(t, resourceSchema.SchemaFunc(), map[string]interface{}{
				"name":              "test-monitor",
				"message":           "test",
				"query":             "avg(last_5m):avg:system.cpu.user{*} > 90",
				"type":              "query alert",
				"notify_no_data":    sentinelNotifyNoData,
				"no_data_timeframe": sentinelNoDataTimeframe,
			})
			d.SetId("12345")

			diags := updateMonitorState(d, nil, monitor)
			require.False(t, diags.HasError(), "updateMonitorState returned errors: %v", diags)

			// Check on_missing_data
			onMissingVal := d.Get("on_missing_data").(string)
			if tc.expectReadSetsOnMissingData {
				assert.Equal(t, tc.onMissingData, onMissingVal,
					"on_missing_data should be set by Read")
			}

			// Check notify_no_data — sentinel detection
			notifyVal := d.Get("notify_no_data").(bool)
			if tc.expectReadSetsNotifyNoData {
				assert.Equal(t, tc.notifyNoData, notifyVal,
					"notify_no_data should be set by Read from API value")
			} else {
				assert.Equal(t, sentinelNotifyNoData, notifyVal,
					"notify_no_data must NOT be touched by Read when on_missing_data is set")
			}

			// Check no_data_timeframe — sentinel detection
			noDataVal := d.Get("no_data_timeframe").(int)
			if tc.expectReadSetsNoDataTimeframe {
				assert.Equal(t, int(*tc.noDataTimeframe), noDataVal,
					"no_data_timeframe should be set by Read from API value")
			} else {
				assert.Equal(t, sentinelNoDataTimeframe, noDataVal,
					"no_data_timeframe must NOT be touched by Read when on_missing_data is set")
			}
		})
	}
}

// buildMockMonitor creates a datadogV1.Monitor with minimal options
// needed to test the Read mutual exclusion logic.
func buildMockMonitor(onMissingData string, notifyNoData bool, noDataTimeframe *int64) *datadogV1.Monitor {
	monitorType := datadogV1.MONITORTYPE_QUERY_ALERT
	m := datadogV1.NewMonitor("avg(last_5m):avg:system.cpu.user{*} > 90", monitorType)
	m.SetName("test-monitor")
	m.SetMessage("test")
	m.SetId(12345)

	opts := datadogV1.MonitorOptions{}
	opts.SetNotifyNoData(notifyNoData)
	opts.SetIncludeTags(true)
	opts.SetNewHostDelay(300)

	if onMissingData != "" {
		opts.SetOnMissingData(datadogV1.OnMissingDataOption(onMissingData))
	}

	if noDataTimeframe != nil {
		opts.SetNoDataTimeframe(*noDataTimeframe)
	}

	// Required: thresholds (updateMonitorState reads these)
	thresholds := datadogV1.MonitorThresholds{}
	thresholds.SetCritical(90.0)
	opts.SetThresholds(thresholds)

	// Required: threshold_windows (updateMonitorState reads these)
	thresholdWindows := datadogV1.MonitorThresholdWindowOptions{}
	opts.SetThresholdWindows(thresholdWindows)

	m.SetOptions(opts)
	return m
}

func int64Ptr(v int64) *int64 {
	return &v
}

// TestUpdateMonitorState_OnMissingData_SchemaDefaultsLeak is a regression test
// verifying that notify_no_data and no_data_timeframe do not leak into
// InstanceState when on_missing_data is set. These fields use Computed: true
// (not Default) so they only appear in state when explicitly set via d.Set().
// Leaking defaults into state causes generate-config-out to emit conflicting
// attributes in the generated HCL.
func TestUpdateMonitorState_OnMissingData_SchemaDefaultsLeak(t *testing.T) {
	monitor := buildMockMonitor("show_no_data", false, nil)
	resourceSchema := resourceDatadogMonitor()

	// Simulate import: only set the required fields, let schema handle the rest.
	d := schema.TestResourceDataRaw(t, resourceSchema.SchemaFunc(), map[string]interface{}{
		"name":    "test-monitor",
		"message": "test",
		"query":   "avg(last_5m):avg:system.cpu.user{*} > 90",
		"type":    "query alert",
	})
	d.SetId("12345")

	diags := updateMonitorState(d, nil, monitor)
	require.False(t, diags.HasError(), "updateMonitorState returned errors: %v", diags)

	assert.Equal(t, "show_no_data", d.Get("on_missing_data").(string))

	// Verify that conflicting fields do NOT appear in InstanceState.
	// If they do, generate-config-out will emit them alongside on_missing_data,
	// triggering ConflictsWith validation errors.
	instanceState := d.State()
	_, notifyInState := instanceState.Attributes["notify_no_data"]
	_, noDataTFInState := instanceState.Attributes["no_data_timeframe"]

	assert.False(t, notifyInState,
		"notify_no_data must not appear in InstanceState when on_missing_data is set")
	assert.False(t, noDataTFInState,
		"no_data_timeframe must not appear in InstanceState when on_missing_data is set")
}
