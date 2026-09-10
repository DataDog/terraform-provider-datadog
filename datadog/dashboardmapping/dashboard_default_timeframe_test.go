package dashboardmapping

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Engine flatten and build round-trips via DashboardDefaultTimeframeField (TypeOneOf).
// Both datadog_dashboard (v1) and datadog_dashboard_v2 share this field spec.

func TestEngineDefaultTimeframe_flattenLive(t *testing.T) {
	t.Parallel()

	resp := map[string]interface{}{
		"default_timeframe": map[string]interface{}{
			"type": "live", "unit": "week", "value": float64(1),
		},
	}
	state := FlattenEngineJSON([]FieldSpec{DashboardDefaultTimeframeField()}, resp)

	dtf, ok := state["default_timeframe"].([]interface{})
	require.True(t, ok)
	require.Len(t, dtf, 1)
	outer := dtf[0].(map[string]interface{})

	liveBlocks, ok := outer["live"].([]interface{})
	require.True(t, ok, "expected 'live' sub-block")
	require.Len(t, liveBlocks, 1)
	live := liveBlocks[0].(map[string]interface{})
	assert.Equal(t, "week", live["unit"])
	assert.Equal(t, 1, live["value"])
	assert.NotContains(t, outer, "fixed")
}

// TestEngineDefaultTimeframe_flattenJSONNumber guards the regression where the datadog
// API client decodes additionalProperties with json.Decoder.UseNumber(), so numeric
// fields arrive as json.Number rather than float64. The engine flatten must handle that
// type or the value silently flattens to 0 (see "expected 1, got 0" CI failure).
func TestEngineDefaultTimeframe_flattenJSONNumber(t *testing.T) {
	t.Parallel()

	resp := map[string]interface{}{
		"default_timeframe": map[string]interface{}{
			"type": "live", "unit": "week", "value": json.Number("1"),
		},
	}
	state := FlattenEngineJSON([]FieldSpec{DashboardDefaultTimeframeField()}, resp)

	dtf := state["default_timeframe"].([]interface{})
	require.Len(t, dtf, 1)
	live := dtf[0].(map[string]interface{})["live"].([]interface{})[0].(map[string]interface{})
	assert.Equal(t, 1, live["value"])
}

func TestEngineDefaultTimeframe_flattenFixed(t *testing.T) {
	t.Parallel()

	resp := map[string]interface{}{
		"default_timeframe": map[string]interface{}{
			"type": "fixed", "from": float64(1776000001000), "to": float64(1776003601000),
		},
	}
	state := FlattenEngineJSON([]FieldSpec{DashboardDefaultTimeframeField()}, resp)

	dtf, ok := state["default_timeframe"].([]interface{})
	require.True(t, ok)
	require.Len(t, dtf, 1)
	outer := dtf[0].(map[string]interface{})

	fixedBlocks, ok := outer["fixed"].([]interface{})
	require.True(t, ok, "expected 'fixed' sub-block")
	require.Len(t, fixedBlocks, 1)
	fixed := fixedBlocks[0].(map[string]interface{})
	assert.Equal(t, 1776000001000, fixed["from"])
	assert.Equal(t, 1776003601000, fixed["to"])
	assert.NotContains(t, outer, "live")
}

func TestEngineDefaultTimeframe_buildLive(t *testing.T) {
	t.Parallel()

	data := map[string]interface{}{
		"default_timeframe": []interface{}{
			map[string]interface{}{
				"live": []interface{}{
					map[string]interface{}{"unit": "week", "value": 1},
				},
			},
		},
	}
	result := BuildEngineJSONFromMap(data, []FieldSpec{DashboardDefaultTimeframeField()})
	dtf, ok := result["default_timeframe"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "live", dtf["type"])
	assert.Equal(t, "week", dtf["unit"])
	assert.Equal(t, 1, dtf["value"])
}

func TestEngineDefaultTimeframe_nullOnClear(t *testing.T) {
	t.Parallel()

	data := map[string]interface{}{
		"default_timeframe": nil,
	}
	result := BuildEngineJSONFromMap(data, []FieldSpec{DashboardDefaultTimeframeField()})
	val, exists := result["default_timeframe"]
	assert.True(t, exists, "key should be present with null value")
	assert.Nil(t, val)
}

func TestEngineDefaultTimeframe_omitWhenEmpty(t *testing.T) {
	t.Parallel()

	data := map[string]interface{}{
		"default_timeframe": []interface{}{},
	}
	result := BuildEngineJSONFromMap(data, []FieldSpec{DashboardDefaultTimeframeField()})
	_, exists := result["default_timeframe"]
	assert.False(t, exists, "should be omitted when empty")
}
