package dashboardmapping

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildDefaultTimeframeJSONFromMap_live(t *testing.T) {
	t.Parallel()

	result, err := BuildDefaultTimeframeJSONFromMap(map[string]interface{}{
		"type":  "live",
		"unit":  "week",
		"value": 1,
	})
	require.NoError(t, err)
	assert.Equal(t, map[string]interface{}{
		"type":  "live",
		"unit":  "week",
		"value": 1,
	}, result)
}

func TestBuildDefaultTimeframeJSONFromMap_fixed(t *testing.T) {
	t.Parallel()

	result, err := BuildDefaultTimeframeJSONFromMap(map[string]interface{}{
		"type": "fixed",
		"from": 1700000000000,
		"to":   1700086400000,
	})
	require.NoError(t, err)
	assert.Equal(t, map[string]interface{}{
		"type": "fixed",
		"from": 1700000000000,
		"to":   1700086400000,
	}, result)
}

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
}

func TestApplyDefaultTimeframeToDashboardJSON_explicitNull(t *testing.T) {
	t.Parallel()

	result := map[string]interface{}{}
	err := ApplyDefaultTimeframeToDashboardJSON(result, map[string]interface{}{
		"default_timeframe": nil,
	})
	require.NoError(t, err)
	assert.Nil(t, result["default_timeframe"])
}

func TestApplyDefaultTimeframeToDashboardJSON_omit(t *testing.T) {
	t.Parallel()

	result := map[string]interface{}{}
	err := ApplyDefaultTimeframeToDashboardJSON(result, map[string]interface{}{
		"default_timeframe": []interface{}{},
	})
	require.NoError(t, err)
	_, ok := result["default_timeframe"]
	assert.False(t, ok)
}

func TestDefaultTimeframeBuildFlattenRoundTrip(t *testing.T) {
	t.Parallel()

	block := map[string]interface{}{
		"type":  "live",
		"unit":  "week",
		"value": 1,
	}
	built, err := BuildDefaultTimeframeJSONFromMap(block)
	require.NoError(t, err)

	flattened := FlattenDefaultTimeframe(built)
	require.Len(t, flattened, 1)
	got := flattened[0].(map[string]interface{})
	assert.Equal(t, "live", got["type"])
	assert.Equal(t, "week", got["unit"])
	assert.Equal(t, 1, got["value"])
}
