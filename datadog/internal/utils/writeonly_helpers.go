package utils

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	frameworkPath "github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// WriteOnlySecretConfig represents configuration for a write-only secret attribute
type WriteOnlySecretConfig struct {
	// Name of the original attribute (e.g., "api_key")
	OriginalAttr string
	// Name of the write-only attribute (e.g., "api_key_wo")
	WriteOnlyAttr string
	// Name of the version trigger attribute (e.g., "api_key_wo_version")
	TriggerAttr string
	// Description for the original attribute
	OriginalDescription string
	// Description for the write-only attribute
	WriteOnlyDescription string
	// Description for the trigger attribute
	TriggerDescription string
}

// CreateWriteOnlySecretAttributes creates schema attributes for a write-only secret pattern
func CreateWriteOnlySecretAttributes(config WriteOnlySecretConfig) map[string]schema.Attribute {
	attrs := map[string]schema.Attribute{
		config.OriginalAttr: schema.StringAttribute{
			Optional:    true,
			Description: config.OriginalDescription,
			Sensitive:   true,
			Validators: []validator.String{
				stringvalidator.ExactlyOneOf(
					frameworkPath.MatchRoot(config.OriginalAttr),
					frameworkPath.MatchRoot(config.WriteOnlyAttr),
				),
				stringvalidator.PreferWriteOnlyAttribute(
					frameworkPath.MatchRoot(config.WriteOnlyAttr),
				),
			},
		},
		config.WriteOnlyAttr: schema.StringAttribute{
			Optional:    true,
			Description: config.WriteOnlyDescription,
			Sensitive:   true,
			WriteOnly:   true,
			Validators: []validator.String{
				stringvalidator.ExactlyOneOf(
					frameworkPath.MatchRoot(config.OriginalAttr),
					frameworkPath.MatchRoot(config.WriteOnlyAttr),
				),
				stringvalidator.AlsoRequires(
					frameworkPath.MatchRoot(config.TriggerAttr),
				),
			},
		},
		config.TriggerAttr: schema.StringAttribute{
			Optional:    true,
			Description: config.TriggerDescription,
			Validators: []validator.String{
				stringvalidator.LengthAtLeast(1),
				stringvalidator.AlsoRequires(frameworkPath.Expressions{
					frameworkPath.MatchRoot(config.WriteOnlyAttr),
				}...),
			},
		},
	}

	return attrs
}

// WriteOnlySecretHandler helps handle write-only secrets in CRUD operations
type WriteOnlySecretHandler struct {
	Config WriteOnlySecretConfig
}

// GetSecretForCreate retrieves the secret value for creation, preferring write-only from config
func (h *WriteOnlySecretHandler) GetSecretForCreate(ctx context.Context, state interface{}, config *tfsdk.Config) (string, bool, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	// Try to get write-only secret from config first
	var writeOnlySecret types.String
	diags.Append(config.GetAttribute(ctx, frameworkPath.Root(h.Config.WriteOnlyAttr), &writeOnlySecret)...)
	if diags.HasError() {
		return "", false, diags
	}

	// If write-only secret is provided, use it
	if !writeOnlySecret.IsNull() && !writeOnlySecret.IsUnknown() {
		return writeOnlySecret.ValueString(), true, diags
	}

	// Otherwise, we'll use the regular attribute (handled by caller)
	return "", false, diags
}

// GetSecretForUpdate retrieves the secret value for updates, only if version changed
func (h *WriteOnlySecretHandler) GetSecretForUpdate(ctx context.Context, config *tfsdk.Config, req *resource.UpdateRequest) (string, bool, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	// Check if version changed by comparing plan vs state
	var planVersion, priorVersion types.String
	diags.Append(req.Plan.GetAttribute(ctx, frameworkPath.Root(h.Config.TriggerAttr), &planVersion)...)
	diags.Append(req.State.GetAttribute(ctx, frameworkPath.Root(h.Config.TriggerAttr), &priorVersion)...)
	if diags.HasError() {
		return "", false, diags
	}

	// Only proceed if version actually changed
	if planVersion.Equal(priorVersion) {
		return "", false, diags
	}

	// Get write-only secret from config
	var writeOnlySecret types.String
	diags.Append(config.GetAttribute(ctx, frameworkPath.Root(h.Config.WriteOnlyAttr), &writeOnlySecret)...)
	if diags.HasError() {
		return "", false, diags
	}

	if !writeOnlySecret.IsNull() && !writeOnlySecret.IsUnknown() {
		return writeOnlySecret.ValueString(), true, diags
	}

	return "", false, diags
}
