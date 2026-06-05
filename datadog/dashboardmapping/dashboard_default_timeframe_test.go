package dashboardmapping

import (
	"testing"

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
}
