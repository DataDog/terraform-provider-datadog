package dashboardmapping

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// FlattenDefaultTimeframe is used by the v1 datadog_dashboard resource (flat schema).

func TestFlattenDefaultTimeframe_live(t *testing.T) {
	t.Parallel()

	result := FlattenDefaultTimeframe(map[string]interface{}{
		"type":  "live",
		"unit":  "week",
		"value": float64(1),
	})
	require.Len(t, result, 1)
	block := result[0].(map[string]interface{})
	assert.Equal(t, "live", block["type"])
	assert.Equal(t, "week", block["unit"])
	assert.Equal(t, 1, block["value"])
	assert.NotContains(t, block, "from")
	assert.NotContains(t, block, "to")
}

func TestFlattenDefaultTimeframe_fixed(t *testing.T) {
	t.Parallel()

	result := FlattenDefaultTimeframe(map[string]interface{}{
		"type": "fixed",
		"from": float64(1776000001000),
		"to":   float64(1776003601000),
	})
	require.Len(t, result, 1)
	block := result[0].(map[string]interface{})
	assert.Equal(t, "fixed", block["type"])
	assert.Equal(t, 1776000001000, block["from"])
	assert.Equal(t, 1776003601000, block["to"])
	assert.NotContains(t, block, "unit")
	assert.NotContains(t, block, "value")
}

// TestFlattenDefaultTimeframeSDKv2State_live verifies that d.Set with a partial map
// (missing from/to) still stores correct zero values via SDKv2 schema processing.
func TestFlattenDefaultTimeframeSDKv2State_live(t *testing.T) {
	t.Parallel()

	d := schema.TestResourceDataRaw(t, map[string]*schema.Schema{
		"default_timeframe": DashboardDefaultTimeframeSchema(),
	}, map[string]interface{}{})

	err := d.Set("default_timeframe", FlattenDefaultTimeframe(map[string]interface{}{
		"type":  "live",
		"unit":  "week",
		"value": float64(1),
	}))
	require.NoError(t, err)

	blocks := d.Get("default_timeframe").([]interface{})
	require.Len(t, blocks, 1)
	block := blocks[0].(map[string]interface{})
	assert.Equal(t, "live", block["type"])
	assert.Equal(t, "week", block["unit"])
	assert.Equal(t, 1, block["value"])
	assert.Equal(t, 0, block["from"])
	assert.Equal(t, 0, block["to"])
}

// Engine flatten and build round-trips via DashboardDefaultTimeframeField (TypeOneOf).

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
