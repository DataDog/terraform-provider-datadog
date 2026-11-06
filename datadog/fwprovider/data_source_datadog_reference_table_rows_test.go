package fwprovider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReferenceTableValuesToMap(t *testing.T) {
	actual, diags := referenceTableValuesToMap(map[string]interface{}{
		"null":      nil,
		"empty":     "",
		"non_empty": "value",
	})
	require.False(t, diags.HasError())

	expected, diags := types.MapValue(types.StringType, map[string]attr.Value{
		"null":      types.StringNull(),
		"empty":     types.StringValue(""),
		"non_empty": types.StringValue("value"),
	})
	require.False(t, diags.HasError())
	assert.Equal(t, expected, actual)
}
