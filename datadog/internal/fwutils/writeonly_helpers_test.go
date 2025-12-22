package fwutils

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMergeAttributes(t *testing.T) {
	t.Skip("generic helper behavior; covered indirectly by write-only attribute schema tests")
}

func TestMergeAttributes_WithWriteOnlyAttributes(t *testing.T) {
	t.Skip("generic helper behavior; covered indirectly by write-only attribute schema tests")
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
	assert.Len(t, apiKey.Validators, 2) // ExactlyOneOf + PreferWriteOnlyAttribute

	// Verify api_key_wo properties
	apiKeyWo := attrs["api_key_wo"].(schema.StringAttribute)
	assert.True(t, apiKeyWo.Optional)
	assert.True(t, apiKeyWo.Sensitive)
	assert.True(t, apiKeyWo.WriteOnly)
	assert.Len(t, apiKeyWo.Validators, 2) // ExactlyOneOf + AlsoRequires

	// Verify api_key_wo_version properties
	version := attrs["api_key_wo_version"].(schema.StringAttribute)
	assert.True(t, version.Optional)
	assert.False(t, version.Sensitive)
	assert.Len(t, version.Validators, 2) // LengthAtLeast + AlsoRequires
}

func TestGetSecretForCreate_WriteOnlyMode(t *testing.T) {
	ctx := context.Background()
	config := WriteOnlySecretConfig{
		OriginalAttr:  "api_key",
		WriteOnlyAttr: "api_key_wo",
		TriggerAttr:   "api_key_wo_version",
	}
	handler := WriteOnlySecretHandler{Config: config}

	// Create config with write-only secret
	tfConfig := tfsdk.Config{
		Raw: tftypes.NewValue(tftypes.Object{
			AttributeTypes: map[string]tftypes.Type{
				"api_key_wo":         tftypes.String,
				"api_key_wo_version": tftypes.String,
			},
		}, map[string]tftypes.Value{
			"api_key_wo":         tftypes.NewValue(tftypes.String, "secret123"),
			"api_key_wo_version": tftypes.NewValue(tftypes.String, "1"),
		}),
		Schema: schema.Schema{
			Attributes: map[string]schema.Attribute{
				"api_key_wo": schema.StringAttribute{
					Optional:  true,
					Sensitive: true,
					WriteOnly: true,
				},
				"api_key_wo_version": schema.StringAttribute{
					Optional: true,
				},
			},
		},
	}

	result := handler.GetSecretForCreate(ctx, &tfConfig)

	assert.False(t, result.Diagnostics.HasError())
	assert.True(t, result.SendWriteOnlySecret)
	assert.Equal(t, "secret123", result.Value)
}

func TestGetSecretForCreate_PlaintextMode(t *testing.T) {
	ctx := context.Background()
	config := WriteOnlySecretConfig{
		OriginalAttr:  "api_key",
		WriteOnlyAttr: "api_key_wo",
		TriggerAttr:   "api_key_wo_version",
	}
	handler := WriteOnlySecretHandler{Config: config}

	// Create config without write-only secret (null)
	tfConfig := tfsdk.Config{
		Raw: tftypes.NewValue(tftypes.Object{
			AttributeTypes: map[string]tftypes.Type{
				"api_key_wo": tftypes.String,
			},
		}, map[string]tftypes.Value{
			"api_key_wo": tftypes.NewValue(tftypes.String, nil),
		}),
		Schema: schema.Schema{
			Attributes: map[string]schema.Attribute{
				"api_key_wo": schema.StringAttribute{
					Optional:  true,
					Sensitive: true,
					WriteOnly: true,
				},
			},
		},
	}

	result := handler.GetSecretForCreate(ctx, &tfConfig)

	assert.False(t, result.Diagnostics.HasError())
	assert.False(t, result.SendWriteOnlySecret) // Caller should fall back to plaintext attribute (if applicable)
	assert.Empty(t, result.Value)
}

func TestGetSecretForUpdate_VersionChanged(t *testing.T) {
	ctx := context.Background()
	config := WriteOnlySecretConfig{
		OriginalAttr:  "api_key",
		WriteOnlyAttr: "api_key_wo",
		TriggerAttr:   "api_key_wo_version",
	}
	handler := WriteOnlySecretHandler{Config: config}

	testSchema := schema.Schema{
		Attributes: map[string]schema.Attribute{
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

	// Prior state with version "1"
	priorState := tfsdk.State{
		Raw: tftypes.NewValue(tftypes.Object{
			AttributeTypes: map[string]tftypes.Type{
				"api_key_wo_version": tftypes.String,
			},
		}, map[string]tftypes.Value{
			"api_key_wo_version": tftypes.NewValue(tftypes.String, "1"),
		}),
		Schema: testSchema,
	}

	// Plan with version "2" (changed)
	plan := tfsdk.Plan{
		Raw: tftypes.NewValue(tftypes.Object{
			AttributeTypes: map[string]tftypes.Type{
				"api_key_wo_version": tftypes.String,
			},
		}, map[string]tftypes.Value{
			"api_key_wo_version": tftypes.NewValue(tftypes.String, "2"),
		}),
		Schema: testSchema,
	}

	// Config with new secret
	tfConfig := tfsdk.Config{
		Raw: tftypes.NewValue(tftypes.Object{
			AttributeTypes: map[string]tftypes.Type{
				"api_key_wo": tftypes.String,
			},
		}, map[string]tftypes.Value{
			"api_key_wo": tftypes.NewValue(tftypes.String, "newsecret"),
		}),
		Schema: testSchema,
	}

	req := resource.UpdateRequest{
		State: priorState,
		Plan:  plan,
	}

	result := handler.GetSecretForUpdate(ctx, &tfConfig, &req)

	assert.False(t, result.Diagnostics.HasError())
	assert.True(t, result.SendWriteOnlySecret) // Version changed, apply rotation
	assert.Equal(t, "newsecret", result.Value)
}

func TestGetSecretForUpdate_VersionUnchanged(t *testing.T) {
	ctx := context.Background()
	config := WriteOnlySecretConfig{
		OriginalAttr:  "api_key",
		WriteOnlyAttr: "api_key_wo",
		TriggerAttr:   "api_key_wo_version",
	}
	handler := WriteOnlySecretHandler{Config: config}

	testSchema := schema.Schema{
		Attributes: map[string]schema.Attribute{
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

	// Both state and plan have version "1"
	sameVersion := tftypes.NewValue(tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"api_key_wo_version": tftypes.String,
		},
	}, map[string]tftypes.Value{
		"api_key_wo_version": tftypes.NewValue(tftypes.String, "1"),
	})

	priorState := tfsdk.State{
		Raw:    sameVersion,
		Schema: testSchema,
	}

	plan := tfsdk.Plan{
		Raw:    sameVersion,
		Schema: testSchema,
	}

	// Config with write-only secret present
	tfConfig := tfsdk.Config{
		Raw: tftypes.NewValue(tftypes.Object{
			AttributeTypes: map[string]tftypes.Type{
				"api_key_wo": tftypes.String,
			},
		}, map[string]tftypes.Value{
			"api_key_wo": tftypes.NewValue(tftypes.String, "secret123"),
		}),
		Schema: testSchema,
	}

	req := resource.UpdateRequest{
		State: priorState,
		Plan:  plan,
	}

	result := handler.GetSecretForUpdate(ctx, &tfConfig, &req)

	assert.False(t, result.Diagnostics.HasError())
	assert.False(t, result.SendWriteOnlySecret) // Version unchanged, no rotation (Pattern 1: omit secret)
	assert.Empty(t, result.Value)
}

func TestGetSecretForUpdate_FirstUpdateAfterCreate(t *testing.T) {
	// Critical edge case: First update after resource creation
	// Prior state has null version (never set), plan has version "1"
	ctx := context.Background()
	config := WriteOnlySecretConfig{
		OriginalAttr:  "api_key",
		WriteOnlyAttr: "api_key_wo",
		TriggerAttr:   "api_key_wo_version",
	}
	handler := WriteOnlySecretHandler{Config: config}

	testSchema := schema.Schema{
		Attributes: map[string]schema.Attribute{
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

	// Prior state: version is null (resource created without write-only)
	priorState := tfsdk.State{
		Raw: tftypes.NewValue(tftypes.Object{
			AttributeTypes: map[string]tftypes.Type{
				"api_key_wo_version": tftypes.String,
			},
		}, map[string]tftypes.Value{
			"api_key_wo_version": tftypes.NewValue(tftypes.String, nil), // null
		}),
		Schema: testSchema,
	}

	// Plan: version now set to "1" (adding write-only for first time)
	plan := tfsdk.Plan{
		Raw: tftypes.NewValue(tftypes.Object{
			AttributeTypes: map[string]tftypes.Type{
				"api_key_wo_version": tftypes.String,
			},
		}, map[string]tftypes.Value{
			"api_key_wo_version": tftypes.NewValue(tftypes.String, "1"),
		}),
		Schema: testSchema,
	}

	// Config with write-only secret
	tfConfig := tfsdk.Config{
		Raw: tftypes.NewValue(tftypes.Object{
			AttributeTypes: map[string]tftypes.Type{
				"api_key_wo": tftypes.String,
			},
		}, map[string]tftypes.Value{
			"api_key_wo": tftypes.NewValue(tftypes.String, "newsecret"),
		}),
		Schema: testSchema,
	}

	req := resource.UpdateRequest{
		State: priorState,
		Plan:  plan,
	}

	result := handler.GetSecretForUpdate(ctx, &tfConfig, &req)

	// Version changed from null to "1", should trigger rotation
	assert.False(t, result.Diagnostics.HasError())
	assert.True(t, result.SendWriteOnlySecret)
	assert.Equal(t, "newsecret", result.Value)
}

func TestGetSecretForUpdate_VersionChangedButNoSecret(t *testing.T) {
	// Edge case: version changed but write-only secret not in config
	// This would happen if user changes version but forgets to provide secret
	ctx := context.Background()
	config := WriteOnlySecretConfig{
		OriginalAttr:  "api_key",
		WriteOnlyAttr: "api_key_wo",
		TriggerAttr:   "api_key_wo_version",
	}
	handler := WriteOnlySecretHandler{Config: config}

	testSchema := schema.Schema{
		Attributes: map[string]schema.Attribute{
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

	// Prior state with version "1"
	priorState := tfsdk.State{
		Raw: tftypes.NewValue(tftypes.Object{
			AttributeTypes: map[string]tftypes.Type{
				"api_key_wo_version": tftypes.String,
			},
		}, map[string]tftypes.Value{
			"api_key_wo_version": tftypes.NewValue(tftypes.String, "1"),
		}),
		Schema: testSchema,
	}

	// Plan with version "2"
	plan := tfsdk.Plan{
		Raw: tftypes.NewValue(tftypes.Object{
			AttributeTypes: map[string]tftypes.Type{
				"api_key_wo_version": tftypes.String,
			},
		}, map[string]tftypes.Value{
			"api_key_wo_version": tftypes.NewValue(tftypes.String, "2"),
		}),
		Schema: testSchema,
	}

	// Config without write-only secret (null)
	tfConfig := tfsdk.Config{
		Raw: tftypes.NewValue(tftypes.Object{
			AttributeTypes: map[string]tftypes.Type{
				"api_key_wo": tftypes.String,
			},
		}, map[string]tftypes.Value{
			"api_key_wo": tftypes.NewValue(tftypes.String, nil),
		}),
		Schema: testSchema,
	}

	req := resource.UpdateRequest{
		State: priorState,
		Plan:  plan,
	}

	result := handler.GetSecretForUpdate(ctx, &tfConfig, &req)

	// Version changed but no secret in config
	// Should return SendWriteOnlySecret=false (caller decides fallback vs error)
	assert.False(t, result.Diagnostics.HasError())
	assert.False(t, result.SendWriteOnlySecret)
	assert.Empty(t, result.Value)
}

func TestGetSecretForUpdate_WriteOnlySecretUnknown(t *testing.T) {
	// Edge case: write-only secret is unknown in config (e.g., depends on computed values).
	// Helper must not treat unknown as a usable secret.
	ctx := context.Background()
	config := WriteOnlySecretConfig{
		OriginalAttr:  "api_key",
		WriteOnlyAttr: "api_key_wo",
		TriggerAttr:   "api_key_wo_version",
	}
	handler := WriteOnlySecretHandler{Config: config}

	testSchema := schema.Schema{
		Attributes: map[string]schema.Attribute{
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

	priorState := tfsdk.State{
		Raw: tftypes.NewValue(tftypes.Object{
			AttributeTypes: map[string]tftypes.Type{
				"api_key_wo_version": tftypes.String,
			},
		}, map[string]tftypes.Value{
			"api_key_wo_version": tftypes.NewValue(tftypes.String, "1"),
		}),
		Schema: testSchema,
	}

	plan := tfsdk.Plan{
		Raw: tftypes.NewValue(tftypes.Object{
			AttributeTypes: map[string]tftypes.Type{
				"api_key_wo_version": tftypes.String,
			},
		}, map[string]tftypes.Value{
			"api_key_wo_version": tftypes.NewValue(tftypes.String, "2"),
		}),
		Schema: testSchema,
	}

	// Config with unknown write-only secret.
	tfConfig := tfsdk.Config{
		Raw: tftypes.NewValue(tftypes.Object{
			AttributeTypes: map[string]tftypes.Type{
				"api_key_wo": tftypes.String,
			},
		}, map[string]tftypes.Value{
			"api_key_wo": tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
		}),
		Schema: testSchema,
	}

	req := resource.UpdateRequest{
		State: priorState,
		Plan:  plan,
	}

	result := handler.GetSecretForUpdate(ctx, &tfConfig, &req)

	assert.False(t, result.Diagnostics.HasError())
	assert.False(t, result.SendWriteOnlySecret)
	assert.Empty(t, result.Value)
}

func TestGetSecretForUpdate_SecretRequiredOnUpdate(t *testing.T) {
	// When SecretRequiredOnUpdate=true: secret is always returned when present in config,
	// regardless of version trigger. This tests the key difference from the default behavior.
	ctx := context.Background()
	config := WriteOnlySecretConfig{
		OriginalAttr:  "api_key",
		WriteOnlyAttr: "api_key_wo",
		TriggerAttr:   "api_key_wo_version",
	}
	handler := WriteOnlySecretHandler{
		Config:                 config,
		SecretRequiredOnUpdate: true,
	}

	testSchema := schema.Schema{
		Attributes: map[string]schema.Attribute{
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

	// Both state and plan have version "1" (unchanged)
	sameVersion := tftypes.NewValue(tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"api_key_wo_version": tftypes.String,
		},
	}, map[string]tftypes.Value{
		"api_key_wo_version": tftypes.NewValue(tftypes.String, "1"),
	})

	priorState := tfsdk.State{
		Raw:    sameVersion,
		Schema: testSchema,
	}

	plan := tfsdk.Plan{
		Raw:    sameVersion,
		Schema: testSchema,
	}

	// Config with write-only secret
	tfConfig := tfsdk.Config{
		Raw: tftypes.NewValue(tftypes.Object{
			AttributeTypes: map[string]tftypes.Type{
				"api_key_wo": tftypes.String,
			},
		}, map[string]tftypes.Value{
			"api_key_wo": tftypes.NewValue(tftypes.String, "secret123"),
		}),
		Schema: testSchema,
	}

	req := resource.UpdateRequest{
		State: priorState,
		Plan:  plan,
	}

	result := handler.GetSecretForUpdate(ctx, &tfConfig, &req)

	// Secret should be returned even though version unchanged
	// This is the key difference from default behavior (SecretRequiredOnUpdate=false)
	assert.False(t, result.Diagnostics.HasError())
	assert.True(t, result.SendWriteOnlySecret)
	assert.Equal(t, "secret123", result.Value)
}
