package fwutils

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

// WriteOnlySecretConfig configures a secret attribute that supports both modes:
// - Plaintext mode: for Terraform <1.11 or users preferring state storage
// - Write-only mode: for Terraform 1.11+ with secrets not stored in state
type WriteOnlySecretConfig struct {
	OriginalAttr         string // Plaintext attribute (e.g., "secret_key")
	WriteOnlyAttr        string // Write-only attribute (e.g., "secret_key_wo")
	TriggerAttr          string // Version trigger (e.g., "secret_key_wo_version")
	OriginalDescription  string
	WriteOnlyDescription string
	TriggerDescription   string
}

// CreateWriteOnlySecretAttributes generates three attributes for dual-mode secret support:
// 1. Original attr (plaintext) - for TF <1.11 or backwards compatibility
// 2. Write-only attr - for TF 1.11+ (not stored in state)
// 3. Version trigger - when changed, applies the write-only secret
// Users choose one mode via ExactlyOneOf validator.
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

// SecretResult contains the result of secret retrieval
type SecretResult struct {
	// Value contains the secret value fetched from the write-only attribute in config.
	// It is only set when SendWriteOnlySecret is true.
	Value string

	// SendWriteOnlySecret indicates whether the caller should send the write-only secret to the API.
	//
	// Important: when this is false, it does NOT necessarily mean "use plaintext".
	// It can also mean "write-only secret exists in config, but version trigger didn't change,
	// so omit the secret for a partial update" (Pattern 1 behavior).
	SendWriteOnlySecret bool

	Diagnostics diag.Diagnostics
}

// WriteOnlySecretHandler manages secret retrieval for both plaintext and write-only modes
type WriteOnlySecretHandler struct {
	Config                 WriteOnlySecretConfig
	SecretRequiredOnUpdate bool // If true, API requires secret in every update; if false (default), secret is optional
}

// GetSecretForCreate retrieves secret for resource creation.
// Returns SendWriteOnlySecret=true if write-only secret provided in config,
// otherwise caller should fall back to the plaintext attribute (if applicable).
func (h *WriteOnlySecretHandler) GetSecretForCreate(ctx context.Context, config *tfsdk.Config) SecretResult {
	result := SecretResult{}

	// Write-only values only exist in config, never in plan/state
	var writeOnlySecret types.String
	result.Diagnostics.Append(config.GetAttribute(ctx, frameworkPath.Root(h.Config.WriteOnlyAttr), &writeOnlySecret)...)
	if result.Diagnostics.HasError() {
		return result
	}

	if !writeOnlySecret.IsNull() && !writeOnlySecret.IsUnknown() {
		result.Value = writeOnlySecret.ValueString()
		result.SendWriteOnlySecret = true
	}

	return result
}

// GetSecretForUpdate retrieves secret for resource updates, with behavior based on SecretRequiredOnUpdate.
//
// When SecretRequiredOnUpdate is false (default):
//   - Pattern 1: API supports partial updates (secret is optional)
//   - Returns SendWriteOnlySecret=true ONLY if version trigger changed AND write-only secret in config
//   - If version unchanged: returns SendWriteOnlySecret=false (caller omits field for partial update)
//   - Use for APIs where secret field is truly optional
//
// When SecretRequiredOnUpdate is true:
//   - Pattern 2: API requires secret in every update request
//   - Returns SendWriteOnlySecret=true if write-only secret exists in config (regardless of version)
//   - Version trigger only matters for forcing Terraform to detect a change
//   - Secret is always sent to API when using write-only mode
//   - Use for APIs where secret field is mandatory in update requests
func (h *WriteOnlySecretHandler) GetSecretForUpdate(ctx context.Context, config *tfsdk.Config, req *resource.UpdateRequest) SecretResult {
	result := SecretResult{}

	// First, check if write-only secret is present in config
	var writeOnlySecret types.String
	result.Diagnostics.Append(config.GetAttribute(ctx, frameworkPath.Root(h.Config.WriteOnlyAttr), &writeOnlySecret)...)
	if result.Diagnostics.HasError() {
		return result
	}

	// Not using write-only mode (no write-only secret in config).
	// Caller can fall back to plaintext mode (if the resource supports it).
	if writeOnlySecret.IsNull() || writeOnlySecret.IsUnknown() {
		return result // SendWriteOnlySecret=false
	}

	// Using write-only mode - behavior depends on SecretRequiredOnUpdate
	if h.SecretRequiredOnUpdate {
		// Pattern 2: API requires secret on every update
		// Always return the secret from config, regardless of version
		result.Value = writeOnlySecret.ValueString()
		result.SendWriteOnlySecret = true
		return result
	}

	// Pattern 1: API supports partial updates (default/preferred pattern)
	// Only return secret if version trigger changed

	// Check if version trigger changed (plan vs state)
	var planVersion, priorVersion types.String
	result.Diagnostics.Append(req.Plan.GetAttribute(ctx, frameworkPath.Root(h.Config.TriggerAttr), &planVersion)...)
	result.Diagnostics.Append(req.State.GetAttribute(ctx, frameworkPath.Root(h.Config.TriggerAttr), &priorVersion)...)
	if result.Diagnostics.HasError() {
		return result
	}

	// Version unchanged = don't send secret (partial update)
	if planVersion.Equal(priorVersion) {
		return result // SendWriteOnlySecret=false
	}

	// Version changed - return secret for rotation
	result.Value = writeOnlySecret.ValueString()
	result.SendWriteOnlySecret = true
	return result
}
