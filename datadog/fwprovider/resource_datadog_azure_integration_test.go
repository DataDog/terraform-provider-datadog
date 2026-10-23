package fwprovider

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	frameworkPath "github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"
)

func TestIntegrationAzureDisplayNameValidator(t *testing.T) {
	tests := []struct {
		name        string
		displayName string
		wantError   bool
	}{
		{name: "nonblank", displayName: "datadog-azure-integration"},
		{name: "blank", displayName: " \t\n", wantError: true},
		{name: "maximum length", displayName: strings.Repeat("a", 256)},
		{name: "too long", displayName: strings.Repeat("a", 257), wantError: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var response validator.StringResponse
			for _, displayNameValidator := range integrationAzureDisplayNameValidators {
				displayNameValidator.ValidateString(context.Background(), validator.StringRequest{
					Path:        frameworkPath.Root("display_name"),
					ConfigValue: types.StringValue(test.displayName),
				}, &response)
			}

			require.Equal(t, test.wantError, response.Diagnostics.HasError())
		})
	}
}

func TestBuildIntegrationAzureRequestBodyDisplayName(t *testing.T) {
	resource := integrationAzureResource{}
	state := integrationAzureModel{
		DisplayName: types.StringValue("datadog-azure-integration"),
	}

	createBody := resource.buildIntegrationAzureRequestBody(context.Background(), &state, "tenant", "client", false)
	addAzureIntegrationDisplayName(createBody, state.DisplayName)
	createPayload, err := json.Marshal(createBody)
	require.NoError(t, err)
	var createPayloadFields map[string]interface{}
	require.NoError(t, json.Unmarshal(createPayload, &createPayloadFields))
	require.Equal(t, "datadog-azure-integration", createPayloadFields["display_name"])

	updateBody := resource.buildIntegrationAzureRequestBody(context.Background(), &state, "tenant", "client", true)
	updatePayload, err := json.Marshal(updateBody)
	require.NoError(t, err)
	var updatePayloadFields map[string]interface{}
	require.NoError(t, json.Unmarshal(updatePayload, &updatePayloadFields))
	require.NotContains(t, updatePayloadFields, "display_name")
}
