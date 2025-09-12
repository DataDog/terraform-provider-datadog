package test

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/ephemeral"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral/schema"
	"github.com/stretchr/testify/assert"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/fwprovider"
)

// Simple mock for testing wrapper functionality
type mockEphemeralResource struct{}

func (m *mockEphemeralResource) Metadata(ctx context.Context, req ephemeral.MetadataRequest, resp *ephemeral.MetadataResponse) {
	resp.TypeName = "_test_resource"
}

func (m *mockEphemeralResource) Schema(ctx context.Context, req ephemeral.SchemaRequest, resp *ephemeral.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{Required: true},
		},
	}
}

func (m *mockEphemeralResource) Open(ctx context.Context, req ephemeral.OpenRequest, resp *ephemeral.OpenResponse) {
	// Basic implementation for testing
}

func TestFrameworkEphemeralResourceWrapper_CoreMethods(t *testing.T) {
	t.Parallel()

	var mock ephemeral.EphemeralResource = &mockEphemeralResource{}
	wrapped := fwprovider.NewFrameworkEphemeralResourceWrapper(&mock)

	// Test Metadata adds provider prefix
	t.Run("Metadata", func(t *testing.T) {
		req := ephemeral.MetadataRequest{ProviderTypeName: "datadog"}
		resp := &ephemeral.MetadataResponse{}

		wrapped.Metadata(context.Background(), req, resp)

		assert.Equal(t, "datadog_test_resource", resp.TypeName)
	})

	// Test Schema calls enrichment
	t.Run("Schema", func(t *testing.T) {
		req := ephemeral.SchemaRequest{}
		resp := &ephemeral.SchemaResponse{}

		wrapped.Schema(context.Background(), req, resp)

		assert.NotNil(t, resp.Schema)
		assert.Contains(t, resp.Schema.Attributes, "id")
	})

	// Test Open delegates properly
	t.Run("Open", func(t *testing.T) {
		req := ephemeral.OpenRequest{}
		resp := &ephemeral.OpenResponse{}

		assert.NotPanics(t, func() {
			wrapped.Open(context.Background(), req, resp)
		})
	})
}

func TestFrameworkEphemeralResourceWrapper_InterfaceDetection(t *testing.T) {
	t.Parallel()

	var mock ephemeral.EphemeralResource = &mockEphemeralResource{}
	wrappedInterface := fwprovider.NewFrameworkEphemeralResourceWrapper(&mock)

	// Cast to wrapper type to access optional interface methods
	wrapped := wrappedInterface.(*fwprovider.FrameworkEphemeralResourceWrapper)

	// Test that optional methods don't panic when inner resource doesn't implement them
	t.Run("Configure_NotImplemented", func(t *testing.T) {
		req := ephemeral.ConfigureRequest{}
		resp := &ephemeral.ConfigureResponse{}

		assert.NotPanics(t, func() {
			wrapped.Configure(context.Background(), req, resp)
		})
	})

	t.Run("ValidateConfig_NotImplemented", func(t *testing.T) {
		req := ephemeral.ValidateConfigRequest{}
		resp := &ephemeral.ValidateConfigResponse{}

		assert.NotPanics(t, func() {
			wrapped.ValidateConfig(context.Background(), req, resp)
		})
	})

	t.Run("ConfigValidators_NotImplemented", func(t *testing.T) {
		validators := wrapped.ConfigValidators(context.Background())
		assert.Nil(t, validators)
	})

	t.Run("Renew_NotImplemented", func(t *testing.T) {
		req := ephemeral.RenewRequest{}
		resp := &ephemeral.RenewResponse{}

		assert.NotPanics(t, func() {
			wrapped.Renew(context.Background(), req, resp)
		})
	})

	t.Run("Close_NotImplemented", func(t *testing.T) {
		req := ephemeral.CloseRequest{}
		resp := &ephemeral.CloseResponse{}

		assert.NotPanics(t, func() {
			wrapped.Close(context.Background(), req, resp)
		})
	})
}
