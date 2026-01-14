package fwutils

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func testWriteOnlySchema() schema.Schema {
	return schema.Schema{
		Attributes: map[string]schema.Attribute{
			"api_key": schema.StringAttribute{
				Optional:  true,
				Sensitive: true,
			},
			"api_key_wo": schema.StringAttribute{
				Optional:  true,
				Sensitive: true,
				WriteOnly: true,
			},
			"api_key_wo_version": schema.StringAttribute{
				Optional: true,
			},
		},
	}
}

func ptr(s string) *string { return &s }

func makeConfigValue(apiKey, apiKeyWo *string) tftypes.Value {
	apiKeyVal := tftypes.NewValue(tftypes.String, nil)
	if apiKey != nil {
		apiKeyVal = tftypes.NewValue(tftypes.String, *apiKey)
	}
	apiKeyWoVal := tftypes.NewValue(tftypes.String, nil)
	if apiKeyWo != nil {
		apiKeyWoVal = tftypes.NewValue(tftypes.String, *apiKeyWo)
	}
	return tftypes.NewValue(tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"api_key":            tftypes.String,
			"api_key_wo":         tftypes.String,
			"api_key_wo_version": tftypes.String,
		},
	}, map[string]tftypes.Value{
		"api_key":            apiKeyVal,
		"api_key_wo":         apiKeyWoVal,
		"api_key_wo_version": tftypes.NewValue(tftypes.String, nil),
	})
}

func makeVersionValue(version *string) tftypes.Value {
	versionVal := tftypes.NewValue(tftypes.String, nil)
	if version != nil {
		versionVal = tftypes.NewValue(tftypes.String, *version)
	}
	return tftypes.NewValue(tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"api_key_wo_version": tftypes.String,
		},
	}, map[string]tftypes.Value{
		"api_key_wo_version": versionVal,
	})
}

func TestCreateWriteOnlySecretAttributes(t *testing.T) {
	config := WriteOnlySecretConfig{
		OriginalAttr:         "api_key",
		WriteOnlyAttr:        "api_key_wo",
		TriggerAttr:          "api_key_wo_version",
		OriginalDescription:  "The API key for the account.",
		WriteOnlyDescription: "Write-only API key for the account.",
		TriggerDescription:   "Version for api_key_wo rotation.",
	}

	attrs := CreateWriteOnlySecretAttributes(config)

	if len(attrs) != 3 {
		t.Errorf("expected 3 attributes, got %d", len(attrs))
	}

	// Verify api_key properties
	apiKey := attrs["api_key"].(schema.StringAttribute)
	if !apiKey.Optional || !apiKey.Sensitive || apiKey.WriteOnly {
		t.Error("api_key: expected Optional=true, Sensitive=true, WriteOnly=false")
	}
	if len(apiKey.Validators) != 2 {
		t.Errorf("api_key should have 2 validators, got %d", len(apiKey.Validators))
	}

	// Verify api_key_wo properties
	apiKeyWo := attrs["api_key_wo"].(schema.StringAttribute)
	if !apiKeyWo.Optional || !apiKeyWo.Sensitive || !apiKeyWo.WriteOnly {
		t.Error("api_key_wo: expected Optional=true, Sensitive=true, WriteOnly=true")
	}
	if len(apiKeyWo.Validators) != 2 {
		t.Errorf("api_key_wo should have 2 validators, got %d", len(apiKeyWo.Validators))
	}

	// Verify api_key_wo_version properties
	version := attrs["api_key_wo_version"].(schema.StringAttribute)
	if !version.Optional || version.Sensitive {
		t.Error("api_key_wo_version: expected Optional=true, Sensitive=false")
	}
	if len(version.Validators) != 2 {
		t.Errorf("api_key_wo_version should have 2 validators, got %d", len(version.Validators))
	}
}

func TestGetSecretForCreate(t *testing.T) {
	tests := []struct {
		name          string
		apiKey        *string
		apiKeyWo      *string
		wantShouldSet bool
		wantValue     string
	}{
		{
			name:          "write-only mode",
			apiKeyWo:      ptr("secret123"),
			wantShouldSet: true,
			wantValue:     "secret123",
		},
		{
			name:          "plaintext mode",
			apiKey:        ptr("plaintext_secret"),
			wantShouldSet: true,
			wantValue:     "plaintext_secret",
		},
		{
			name:          "no secret",
			wantShouldSet: false,
		},
		{
			name:          "empty string is valid",
			apiKey:        ptr(""),
			wantShouldSet: true,
			wantValue:     "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			handler := WriteOnlySecretHandler{
				Config: WriteOnlySecretConfig{
					OriginalAttr:  "api_key",
					WriteOnlyAttr: "api_key_wo",
					TriggerAttr:   "api_key_wo_version",
				},
			}

			tfConfig := tfsdk.Config{
				Raw:    makeConfigValue(tt.apiKey, tt.apiKeyWo),
				Schema: testWriteOnlySchema(),
			}

			result := handler.GetSecretForCreate(ctx, &tfConfig)

			if result.Diagnostics.HasError() {
				t.Errorf("unexpected error: %v", result.Diagnostics)
			}
			if result.ShouldSetValue != tt.wantShouldSet {
				t.Errorf("ShouldSetValue = %v, want %v", result.ShouldSetValue, tt.wantShouldSet)
			}
			if tt.wantShouldSet && result.Value != tt.wantValue {
				t.Errorf("Value = %q, want %q", result.Value, tt.wantValue)
			}
		})
	}
}

func TestGetSecretForUpdate(t *testing.T) {
	tests := []struct {
		name                   string
		stateVersion           *string
		planVersion            *string
		apiKey                 *string
		apiKeyWo               *string
		apiKeyWoUnknown        bool
		secretRequiredOnUpdate bool
		wantShouldSet          bool
		wantValue              string
	}{
		{
			name:          "version changed triggers update",
			stateVersion:  ptr("1"),
			planVersion:   ptr("2"),
			apiKeyWo:      ptr("newsecret"),
			wantShouldSet: true,
			wantValue:     "newsecret",
		},
		{
			name:          "version unchanged skips update (partial update)",
			stateVersion:  ptr("1"),
			planVersion:   ptr("1"),
			apiKeyWo:      ptr("secret123"),
			wantShouldSet: false,
		},
		{
			name:          "null to value transition triggers update",
			stateVersion:  nil,
			planVersion:   ptr("1"),
			apiKeyWo:      ptr("newsecret"),
			wantShouldSet: true,
			wantValue:     "newsecret",
		},
		{
			name:          "plaintext fallback",
			stateVersion:  nil,
			planVersion:   nil,
			apiKey:        ptr("plaintext_value"),
			wantShouldSet: true,
			wantValue:     "plaintext_value",
		},
		{
			name:            "unknown write-only with no plaintext",
			stateVersion:    ptr("1"),
			planVersion:     ptr("2"),
			apiKeyWoUnknown: true,
			wantShouldSet:   false,
		},
		{
			name:                   "secret required on update ignores version",
			stateVersion:           ptr("1"),
			planVersion:            ptr("1"),
			apiKeyWo:               ptr("secret123"),
			secretRequiredOnUpdate: true,
			wantShouldSet:          true,
			wantValue:              "secret123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			testSchema := testWriteOnlySchema()
			handler := WriteOnlySecretHandler{
				Config: WriteOnlySecretConfig{
					OriginalAttr:  "api_key",
					WriteOnlyAttr: "api_key_wo",
					TriggerAttr:   "api_key_wo_version",
				},
				SecretRequiredOnUpdate: tt.secretRequiredOnUpdate,
			}

			priorState := tfsdk.State{
				Raw:    makeVersionValue(tt.stateVersion),
				Schema: testSchema,
			}
			plan := tfsdk.Plan{
				Raw:    makeVersionValue(tt.planVersion),
				Schema: testSchema,
			}

			// Build config value, handling unknown case
			var configRaw tftypes.Value
			if tt.apiKeyWoUnknown {
				apiKeyVal := tftypes.NewValue(tftypes.String, nil)
				configRaw = tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"api_key":            tftypes.String,
						"api_key_wo":         tftypes.String,
						"api_key_wo_version": tftypes.String,
					},
				}, map[string]tftypes.Value{
					"api_key":            apiKeyVal,
					"api_key_wo":         tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
					"api_key_wo_version": tftypes.NewValue(tftypes.String, nil),
				})
			} else {
				configRaw = makeConfigValue(tt.apiKey, tt.apiKeyWo)
			}

			tfConfig := tfsdk.Config{
				Raw:    configRaw,
				Schema: testSchema,
			}

			req := resource.UpdateRequest{
				State: priorState,
				Plan:  plan,
			}

			result := handler.GetSecretForUpdate(ctx, &tfConfig, &req)

			if result.Diagnostics.HasError() {
				t.Errorf("unexpected error: %v", result.Diagnostics)
			}
			if result.ShouldSetValue != tt.wantShouldSet {
				t.Errorf("ShouldSetValue = %v, want %v", result.ShouldSetValue, tt.wantShouldSet)
			}
			if tt.wantShouldSet && result.Value != tt.wantValue {
				t.Errorf("Value = %q, want %q", result.Value, tt.wantValue)
			}
		})
	}
}
