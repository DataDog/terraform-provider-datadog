package fwprovider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/defaults"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDatastoreSchemaPrimaryKeyGenerationStrategyDefault guards against a
// regression of https://github.com/DataDog/terraform-provider-datadog/issues/3610.
//
// When `primary_key_generation_strategy` is omitted, the API reads back "none",
// which previously produced a "Provider produced inconsistent result after apply"
// error (plan held null, apply yielded "none"). Marking the attribute
// Optional+Computed with a static default of "none" makes the planned value
// resolve to "none" so it matches the post-apply state.
func TestDatastoreSchemaPrimaryKeyGenerationStrategyDefault(t *testing.T) {
	ctx := context.Background()

	resp := &resource.SchemaResponse{}
	NewDatastoreResource().Schema(ctx, resource.SchemaRequest{}, resp)
	require.False(t, resp.Diagnostics.HasError(), "schema should not produce diagnostics")

	attr, ok := resp.Schema.Attributes["primary_key_generation_strategy"]
	require.True(t, ok, "primary_key_generation_strategy attribute should exist")

	strAttr, ok := attr.(schema.StringAttribute)
	require.True(t, ok, "primary_key_generation_strategy should be a StringAttribute")

	// A Default may only be set on a Computed attribute in the plugin framework,
	// and the attribute must stay Optional so users can still override it.
	assert.True(t, strAttr.Optional, "attribute should remain Optional")
	assert.True(t, strAttr.Computed, "attribute must be Computed for a Default to apply")
	require.NotNil(t, strAttr.Default, "attribute should declare a Default")

	defResp := defaults.StringResponse{}
	strAttr.Default.DefaultString(ctx, defaults.StringRequest{}, &defResp)
	require.False(t, defResp.Diagnostics.HasError(), "default should not produce diagnostics")
	assert.Equal(t, "none", defResp.PlanValue.ValueString(), "default value should be \"none\"")
}
