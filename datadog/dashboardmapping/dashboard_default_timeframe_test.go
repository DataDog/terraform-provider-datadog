package dashboardmapping

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
