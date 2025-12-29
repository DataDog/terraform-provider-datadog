package fwprovider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/ephemeral"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/secretbridge"
)

var _ ephemeral.EphemeralResource = &secretDecryptResource{}

// NewSecretDecryptResource creates a new ephemeral resource for decrypting secrets.
func NewSecretDecryptResource() ephemeral.EphemeralResource {
	return &secretDecryptResource{}
}

type secretDecryptModel struct {
	Ciphertext    types.String `tfsdk:"ciphertext"`
	EncryptionKey types.String `tfsdk:"encryption_key_wo"`
	Value         types.String `tfsdk:"value"`
}

type secretDecryptResource struct{}

func (r *secretDecryptResource) Metadata(_ context.Context, _ ephemeral.MetadataRequest, resp *ephemeral.MetadataResponse) {
	resp.TypeName = "secret_decrypt"
}

func (r *secretDecryptResource) Schema(_ context.Context, _ ephemeral.SchemaRequest, resp *ephemeral.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Decrypts a ciphertext value that was encrypted using the secretbridge encryption. " +
			"Use this to retrieve plaintext values from encrypted computed attributes like `encrypted_key` " +
			"for passing to secret managers.",
		Attributes: map[string]schema.Attribute{
			"ciphertext": schema.StringAttribute{
				Description: "The encrypted JSON ciphertext to decrypt. Typically from an `encrypted_key` " +
					"attribute of a resource that supports encryption.",
				Required: true,
			},
			"encryption_key_wo": schema.StringAttribute{
				Description: "Decryption key (32 bytes). Must match the key used during encryption. " +
					"Source from an ephemeral resource like `ephemeral.random_password`.",
				Required:  true,
				Sensitive: true,
			},
			"value": schema.StringAttribute{
				Description: "The decrypted plaintext value. This is ephemeral and never stored in state.",
				Computed:    true,
				Sensitive:   true,
			},
		},
	}
}

func (r *secretDecryptResource) Open(ctx context.Context, req ephemeral.OpenRequest, resp *ephemeral.OpenResponse) {
	var config secretDecryptModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	plaintext, diags := secretbridge.Decrypt(
		ctx,
		config.Ciphertext.ValueString(),
		[]byte(config.EncryptionKey.ValueString()),
	)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	config.Value = types.StringValue(plaintext)
	resp.Diagnostics.Append(resp.Result.Set(ctx, &config)...)
}
