package fwutils

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	frameworkPath "github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// MergeAttributes combines multiple attribute maps into a single map.
// Later maps take precedence over earlier ones for duplicate keys.
func MergeAttributes(attributeMaps ...map[string]schema.Attribute) map[string]schema.Attribute {
	result := make(map[string]schema.Attribute)
	for _, attrs := range attributeMaps {
		for key, attr := range attrs {
			result[key] = attr
		}
	}
	return result
}

// WriteOnlySecretMode selects whether a schema keeps the legacy plaintext
// attribute alongside its write-only companion or exposes only the write-only
// interface. The zero value preserves all existing callers.
type WriteOnlySecretMode uint8

const (
	WriteOnlySecretModeLegacy WriteOnlySecretMode = iota
	WriteOnlySecretModeOnly
)

// WriteOnlySecretConfig configures a secret attribute. Legacy mode supports
// both a stateful plaintext attribute and its write-only companion. ModeOnly
// exposes only the write-only attribute and its stateful version trigger.
type WriteOnlySecretConfig struct {
	OriginalAttr         string // Plaintext attribute (e.g., "secret_key")
	WriteOnlyAttr        string // Write-only attribute (e.g., "secret_key_wo")
	TriggerAttr          string // Version trigger (e.g., "secret_key_wo_version")
	OriginalDescription  string
	WriteOnlyDescription string
	TriggerDescription   string
	// ParentBlocks scopes the three attributes under static nested blocks, e.g.
	// []string{"authentication", "basic"}. Empty means the resource root.
	ParentBlocks []string
	// Mode defaults to WriteOnlySecretModeLegacy for backwards compatibility.
	Mode WriteOnlySecretMode
	// Required controls both ModeOnly attributes. Legacy mode ignores it.
	Required bool
}

func (secretConfig WriteOnlySecretConfig) attrPath(attributeName string) frameworkPath.Path {
	if len(secretConfig.ParentBlocks) == 0 {
		return frameworkPath.Root(attributeName)
	}
	attributePath := frameworkPath.Root(secretConfig.ParentBlocks[0])
	for _, blockName := range secretConfig.ParentBlocks[1:] {
		attributePath = attributePath.AtName(blockName)
	}
	return attributePath.AtName(attributeName)
}

func (secretConfig WriteOnlySecretConfig) attrExpression(attributeName string) frameworkPath.Expression {
	if len(secretConfig.ParentBlocks) == 0 {
		return frameworkPath.MatchRoot(attributeName)
	}
	attributeExpression := frameworkPath.MatchRoot(secretConfig.ParentBlocks[0])
	for _, blockName := range secretConfig.ParentBlocks[1:] {
		attributeExpression = attributeExpression.AtName(blockName)
	}
	return attributeExpression.AtName(attributeName)
}

// CreateWriteOnlySecretAttributes generates three attributes for dual-mode secret support:
// 1. Original attr (plaintext) - for TF <1.11 or backwards compatibility
// 2. Write-only attr - for TF 1.11+ (not stored in state)
// 3. Version trigger - when changed, applies the write-only secret
// Users choose one mode via ExactlyOneOf validator.
func CreateWriteOnlySecretAttributes(config WriteOnlySecretConfig) map[string]schema.Attribute {
	if config.Mode == WriteOnlySecretModeOnly {
		writeOnly := schema.StringAttribute{
			Description: config.WriteOnlyDescription,
			WriteOnly:   true,
		}
		trigger := schema.StringAttribute{
			Description: config.TriggerDescription,
			Validators: []validator.String{
				stringvalidator.LengthAtLeast(1),
			},
		}
		if config.Required {
			writeOnly.Required = true
			trigger.Required = true
		} else {
			writeOnly.Optional = true
			trigger.Optional = true
			writeOnly.Validators = []validator.String{
				stringvalidator.AlsoRequires(config.attrExpression(config.TriggerAttr)),
			}
			trigger.Validators = append(trigger.Validators,
				stringvalidator.AlsoRequires(config.attrExpression(config.WriteOnlyAttr)),
			)
		}
		return map[string]schema.Attribute{
			config.WriteOnlyAttr: writeOnly,
			config.TriggerAttr:   trigger,
		}
	}

	attrs := map[string]schema.Attribute{
		config.OriginalAttr: schema.StringAttribute{
			Optional:    true,
			Description: config.OriginalDescription,
			Sensitive:   true,
			Validators: []validator.String{
				stringvalidator.ExactlyOneOf(
					config.attrExpression(config.OriginalAttr),
					config.attrExpression(config.WriteOnlyAttr),
				),
				stringvalidator.PreferWriteOnlyAttribute(
					config.attrExpression(config.WriteOnlyAttr),
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
					config.attrExpression(config.OriginalAttr),
					config.attrExpression(config.WriteOnlyAttr),
				),
				stringvalidator.AlsoRequires(
					config.attrExpression(config.TriggerAttr),
				),
			},
		},
		config.TriggerAttr: schema.StringAttribute{
			Optional:    true,
			Description: config.TriggerDescription,
			Validators: []validator.String{
				stringvalidator.LengthAtLeast(1),
				stringvalidator.AlsoRequires(frameworkPath.Expressions{
					config.attrExpression(config.WriteOnlyAttr),
				}...),
			},
		},
	}

	return attrs
}

// SecretResult contains the result of secret retrieval from either write-only or plaintext attributes.
type SecretResult struct {
	// Value contains the secret value (from write-only OR plaintext attribute).
	// Only valid when ShouldSetValue is true.
	Value string

	// ShouldSetValue indicates whether the caller should set the secret field.
	// True when a value was found (handles empty string as valid value).
	// False when no value found or when omitting for partial update.
	ShouldSetValue bool

	Diagnostics diag.Diagnostics
}

// WriteOnlySecretHandler manages secret retrieval for both plaintext and write-only modes
type WriteOnlySecretHandler struct {
	Config                 WriteOnlySecretConfig
	SecretRequiredOnUpdate bool // If true, API requires secret in every update; if false (default), secret is optional
}

func (h *WriteOnlySecretHandler) requireUpdateSecret(result SecretResult) SecretResult {
	if !h.SecretRequiredOnUpdate || result.ShouldSetValue || result.Diagnostics.HasError() {
		return result
	}
	result.Diagnostics.AddError(
		"Missing write-only secret required for update",
		fmt.Sprintf(
			"The API requires the write-only attribute %q for every update, but configuration did not provide a known value.",
			h.Config.attrPath(h.Config.WriteOnlyAttr).String(),
		),
	)
	return result
}

// GetSecretForCreate retrieves secret for resource creation.
// Checks write-only attribute first, then falls back to plaintext attribute.
// Returns ShouldSetValue=true if either attribute has a value.
func (h *WriteOnlySecretHandler) GetSecretForCreate(ctx context.Context, config *tfsdk.Config) SecretResult {
	result := SecretResult{}

	// Check write-only attribute first (only exists in config, never in plan/state)
	var writeOnlySecret types.String
	result.Diagnostics.Append(config.GetAttribute(ctx, h.Config.attrPath(h.Config.WriteOnlyAttr), &writeOnlySecret)...)
	if result.Diagnostics.HasError() {
		return result
	}

	if !writeOnlySecret.IsNull() && !writeOnlySecret.IsUnknown() {
		result.Value = writeOnlySecret.ValueString()
		result.ShouldSetValue = true
		return result
	}
	if h.Config.Mode == WriteOnlySecretModeOnly {
		return result
	}

	// Fall back to plaintext attribute
	var plaintextSecret types.String
	result.Diagnostics.Append(config.GetAttribute(ctx, h.Config.attrPath(h.Config.OriginalAttr), &plaintextSecret)...)
	if result.Diagnostics.HasError() {
		return result
	}

	if !plaintextSecret.IsNull() && !plaintextSecret.IsUnknown() {
		result.Value = plaintextSecret.ValueString()
		result.ShouldSetValue = true
	}

	return result
}

// GetSecretForUpdate retrieves secret for resource updates, with behavior based on SecretRequiredOnUpdate.
//
// When SecretRequiredOnUpdate is false (default):
//   - Pattern 1: API supports partial updates (secret is optional)
//   - For write-only: returns ShouldSetValue=true ONLY if version trigger changed
//   - If version unchanged: returns ShouldSetValue=false (omit for partial update)
//   - For plaintext: returns ShouldSetValue=true if plaintext attr is set
//
// When SecretRequiredOnUpdate is true:
//   - Pattern 2: API requires secret in every update request
//   - Returns ShouldSetValue=true if write-only or plaintext attr exists in config
//   - Returns an error diagnostic if configuration has no known secret value
//   - Version trigger only matters for forcing Terraform to detect a change
func (h *WriteOnlySecretHandler) GetSecretForUpdate(ctx context.Context, config *tfsdk.Config, req *resource.UpdateRequest) SecretResult {
	result := SecretResult{}

	// Check if write-only secret is present in config
	var writeOnlySecret types.String
	result.Diagnostics.Append(config.GetAttribute(ctx, h.Config.attrPath(h.Config.WriteOnlyAttr), &writeOnlySecret)...)
	if result.Diagnostics.HasError() {
		return result
	}

	// If write-only is set, handle based on SecretRequiredOnUpdate and version trigger
	if !writeOnlySecret.IsNull() && !writeOnlySecret.IsUnknown() {
		if h.SecretRequiredOnUpdate {
			// Pattern 2: API requires secret on every update
			result.Value = writeOnlySecret.ValueString()
			result.ShouldSetValue = true
			return result
		}

		// Pattern 1: API supports partial updates
		// Only return secret if version trigger changed
		var planVersion, priorVersion types.String
		result.Diagnostics.Append(req.Plan.GetAttribute(ctx, h.Config.attrPath(h.Config.TriggerAttr), &planVersion)...)
		result.Diagnostics.Append(req.State.GetAttribute(ctx, h.Config.attrPath(h.Config.TriggerAttr), &priorVersion)...)
		if result.Diagnostics.HasError() {
			return result
		}

		// Version unchanged = omit secret for partial update
		if planVersion.Equal(priorVersion) {
			return result // ShouldSetValue=false
		}

		// Version changed - return secret for rotation
		result.Value = writeOnlySecret.ValueString()
		result.ShouldSetValue = true
		return result
	}
	if h.Config.Mode == WriteOnlySecretModeOnly {
		return h.requireUpdateSecret(result)
	}

	// Fall back to plaintext attribute
	var plaintextSecret types.String
	result.Diagnostics.Append(config.GetAttribute(ctx, h.Config.attrPath(h.Config.OriginalAttr), &plaintextSecret)...)
	if result.Diagnostics.HasError() {
		return result
	}

	if !plaintextSecret.IsNull() && !plaintextSecret.IsUnknown() {
		result.Value = plaintextSecret.ValueString()
		result.ShouldSetValue = true
	}

	return h.requireUpdateSecret(result)
}
