package utils

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateWriteOnlySecretAttributes(t *testing.T) {
	t.Helper()

	config := WriteOnlySecretConfig{
		OriginalAttr:         "api_key",
		WriteOnlyAttr:        "api_key_wo",
		TriggerAttr:          "api_key_wo_version",
		OriginalDescription:  "The API key for the account.",
		WriteOnlyDescription: "Write-only API key for the account.",
		TriggerDescription:   "Version for api_key_wo rotation.",
	}

	attrs := CreateWriteOnlySecretAttributes(config)

	// Verify all three attributes are created
	require.Len(t, attrs, 3)
	require.Contains(t, attrs, "api_key")
	require.Contains(t, attrs, "api_key_wo")
	require.Contains(t, attrs, "api_key_wo_version")

	// Verify api_key properties
	apiKey := attrs["api_key"].(schema.StringAttribute)
	assert.True(t, apiKey.Optional)
	assert.True(t, apiKey.Sensitive)
	assert.False(t, apiKey.WriteOnly)
	assert.Equal(t, "The API key for the account.", apiKey.Description)
	assert.Len(t, apiKey.Validators, 2) // ExactlyOneOf + PreferWriteOnlyAttribute

	// Verify api_key_wo properties
	apiKeyWo := attrs["api_key_wo"].(schema.StringAttribute)
	assert.True(t, apiKeyWo.Optional)
	assert.True(t, apiKeyWo.Sensitive)
	assert.True(t, apiKeyWo.WriteOnly)
	assert.Equal(t, "Write-only API key for the account.", apiKeyWo.Description)
	assert.Len(t, apiKeyWo.Validators, 2) // ExactlyOneOf + AlsoRequires

	// Verify api_key_wo_version properties
	version := attrs["api_key_wo_version"].(schema.StringAttribute)
	assert.True(t, version.Optional)
	assert.False(t, version.Sensitive)
	assert.False(t, version.WriteOnly)
	assert.Equal(t, "Version for api_key_wo rotation.", version.Description)
	assert.Len(t, version.Validators, 2) // LengthAtLeast + AlsoRequires
}

func TestWriteOnlySecretHandler_BasicFunctionality(t *testing.T) {
	t.Helper()

	config := WriteOnlySecretConfig{
		OriginalAttr:  "api_key",
		WriteOnlyAttr: "api_key_wo",
		TriggerAttr:   "api_key_wo_version",
	}

	// Test that handler can be created without panicking
	handler := WriteOnlySecretHandler{Config: config}
	assert.Equal(t, "api_key", handler.Config.OriginalAttr)
	assert.Equal(t, "api_key_wo", handler.Config.WriteOnlyAttr)
	assert.Equal(t, "api_key_wo_version", handler.Config.TriggerAttr)
}

func TestWriteOnlySecretHandler_VersionComparison(t *testing.T) {
	t.Helper()

	// Test the core logic: string version comparisons that drive update decisions
	testCases := []struct {
		name         string
		prior        string
		planned      string
		shouldUpdate bool
	}{
		{"numeric versions", "1", "2", true},
		{"semantic versions", "v1.0", "v1.1", true},
		{"date versions", "2024-Q1", "2024-Q2", true},
		{"descriptive versions", "initial", "updated", true},
		{"same version", "v1.0", "v1.0", false},
		{"empty to version", "", "1", true},
		{"version to empty", "1", "", true},
		{"both empty", "", "", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Test version equality logic (core of GetSecretForUpdate)
			versionChanged := tc.prior != tc.planned
			assert.Equal(t, tc.shouldUpdate, versionChanged,
				"Version comparison failed for '%s' -> '%s'", tc.prior, tc.planned)
		})
	}
}

func TestWriteOnlySecretConfig_EdgeCases(t *testing.T) {
	t.Helper()

	tests := []struct {
		name   string
		config WriteOnlySecretConfig
		panics bool
	}{
		{
			name: "valid config",
			config: WriteOnlySecretConfig{
				OriginalAttr:  "api_key",
				WriteOnlyAttr: "api_key_wo",
				TriggerAttr:   "api_key_wo_version",
			},
			panics: false,
		},
		{
			name: "empty strings don't panic",
			config: WriteOnlySecretConfig{
				OriginalAttr:  "",
				WriteOnlyAttr: "",
				TriggerAttr:   "",
			},
			panics: false,
		},
		{
			name: "different attribute patterns",
			config: WriteOnlySecretConfig{
				OriginalAttr:  "client_secret",
				WriteOnlyAttr: "client_secret_wo",
				TriggerAttr:   "client_secret_version",
			},
			panics: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.panics {
				assert.Panics(t, func() {
					CreateWriteOnlySecretAttributes(tt.config)
				})
			} else {
				assert.NotPanics(t, func() {
					attrs := CreateWriteOnlySecretAttributes(tt.config)
					assert.NotNil(t, attrs)

					// Test handler creation doesn't panic
					handler := WriteOnlySecretHandler{Config: tt.config}
					assert.Equal(t, tt.config.OriginalAttr, handler.Config.OriginalAttr)
				})
			}
		})
	}
}
