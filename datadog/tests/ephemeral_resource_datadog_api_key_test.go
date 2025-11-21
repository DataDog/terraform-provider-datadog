package test

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/ephemeral"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/fwprovider"
)

// TODO(v6-protocol): once the provider upgrades to protocol v6, add an acceptance
// test using terraform-plugin-testing's echo provider to assert the API key's
// ephemeral value never persists in Terraform state.

// TestEphemeralAPIKeyResource_Metadata tests the Metadata method
func TestEphemeralAPIKeyResource_Metadata(t *testing.T) {
	resource := fwprovider.NewEphemeralAPIKeyResource()

	req := ephemeral.MetadataRequest{
		ProviderTypeName: "datadog",
	}
	resp := &ephemeral.MetadataResponse{}

	resource.Metadata(context.Background(), req, resp)

	assert.Equal(t, "api_key", resp.TypeName)
}

// TestEphemeralAPIKeyResource_Schema tests the Schema method
func TestEphemeralAPIKeyResource_Schema(t *testing.T) {
	resource := fwprovider.NewEphemeralAPIKeyResource()

	req := ephemeral.SchemaRequest{}
	resp := &ephemeral.SchemaResponse{}

	resource.Schema(context.Background(), req, resp)

	// Verify required attributes exist
	require.NotNil(t, resp.Schema.Attributes)

	// Check that key is marked as sesnsitive
	keyAttr, exists := resp.Schema.Attributes["key"]
	require.True(t, exists)
	assert.True(t, keyAttr.IsSensitive())
}

// TestEphemeralAPIKeyResource_InterfaceAssertion tests interface compliance
func TestEphemeralAPIKeyResource_InterfaceAssertion(t *testing.T) {
	resource := fwprovider.NewEphemeralAPIKeyResource()

	// Verify the resource implements required interfaces
	_, ok := resource.(ephemeral.EphemeralResource)
	assert.True(t, ok, "Resource should implement EphemeralResource")

	_, ok = resource.(ephemeral.EphemeralResourceWithConfigure)
	assert.True(t, ok, "Resource should implement EphemeralResourceWithConfigure")
}
