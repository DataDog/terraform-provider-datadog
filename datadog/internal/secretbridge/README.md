# SecretBridge

Encrypts computed secrets for secure transfer to secret managers.

## Problem

Resources like `datadog_api_key` receive sensitive values from the API (computed attributes). Write-only attributes only work for user inputs. We need to encrypt the API response before storing in state.

## Solution

```
API → Encrypt(encryption_key_wo) → State (ciphertext) → Decrypt(encryption_key_wo) → Secret Manager
```

## Requirements

- **Terraform 1.11+** (write-only attributes)
- **Ephemeral key source** (e.g., `ephemeral.random_password`)

## API

```go
func Encrypt(ctx context.Context, plaintext string, key []byte) (string, diag.Diagnostics)
func Decrypt(ctx context.Context, ciphertext string, key []byte) (string, diag.Diagnostics)
func EncryptionKeyAttribute() resourceSchema.StringAttribute
```

## Usage in Resources

### Resource Schema

```go
Attributes: map[string]schema.Attribute{
    "encryption_key_wo":        secretbridge.EncryptionKeyAttribute(),
    "key":           schema.StringAttribute{Computed: true, Sensitive: true},
    "encrypted_key": schema.StringAttribute{Computed: true},
}
```

### Resource Create

```go
plaintextKey := apiResp.GetData().GetAttributes().GetKey()

if !state.EncryptionKey.IsNull() {
    encrypted, diags := secretbridge.Encrypt(ctx, plaintextKey, []byte(state.EncryptionKey.ValueString()))
    resp.Diagnostics.Append(diags...)
    state.EncryptedKey = types.StringValue(encrypted)
    state.Key = types.StringNull()
} else {
    state.Key = types.StringValue(plaintextKey)
}
```

### Ephemeral Decrypter

```go
plaintext, diags := secretbridge.Decrypt(ctx, config.Ciphertext.ValueString(), []byte(config.EncryptionKey.ValueString()))
resp.Diagnostics.Append(diags...)
```

## See Also

- [User examples](../../../examples/resources/datadog_api_key/)
- [Ephemeral decrypter docs](../../../docs/ephemeral-resources/secret_decrypt.md)
- [Development guide](../../../DEVELOPMENT.md#ephemeral-resources-and-secrets)
