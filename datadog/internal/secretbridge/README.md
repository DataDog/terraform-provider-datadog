# SecretBridge

Encrypts computed secrets (API keys, app keys) for secure transfer to secret managers.

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

## Usage

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

### HCL

```hcl
ephemeral "random_password" "key" {
  length = 32
}

resource "datadog_api_key" "example" {
  name   = "my-key"
  encryption_key_wo = ephemeral.random_password.key.result
}

ephemeral "datadog_secret_decrypt" "api_key" {
  ciphertext = datadog_api_key.example.encrypted_key
  encryption_key_wo     = ephemeral.random_password.key.result
}

resource "aws_secretsmanager_secret_version" "api_key" {
  secret_id     = aws_secretsmanager_secret.api_key.id
  secret_string = ephemeral.datadog_secret_decrypt.api_key.value
}
```
