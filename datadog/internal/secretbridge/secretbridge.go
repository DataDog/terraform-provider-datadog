// Package secretbridge provides AES-256-GCM encryption for computed secrets.
// Use with write-only encryption_key_wo attribute to encrypt API responses before state storage.
package secretbridge

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	resourceSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

// KeySize is the required encryption key size (32 bytes for AES-256).
const KeySize = 32

// envelope is the JSON-serialized structure stored in Terraform state.
type envelope struct {
	C []byte `json:"c"` // ciphertext
	N []byte `json:"n"` // nonce
}

// EncryptionKeyAttribute returns the encryption_key_wo write-only attribute for resource schemas.
func EncryptionKeyAttribute() resourceSchema.StringAttribute {
	return resourceSchema.StringAttribute{
		Description: "Write-only encryption key (32 bytes). Source from an ephemeral resource " +
			"like `ephemeral.random_password`. Used to encrypt computed secrets before storing in state. " +
			"The key is never persisted. Requires Terraform 1.11+.",
		Optional:  true,
		Sensitive: true,
		WriteOnly: true,
	}
}

// Encrypt encrypts plaintext using AES-256-GCM and returns JSON ciphertext.
func Encrypt(ctx context.Context, plaintext string, key []byte) (string, diag.Diagnostics) {
	gcm, diags := newGCM(key)
	if diags.HasError() {
		return "", diags
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		diags.AddError("Encryption Failed", fmt.Sprintf("failed to generate nonce: %v", err))
		return "", diags
	}

	data, err := json.Marshal(envelope{
		C: gcm.Seal(nil, nonce, []byte(plaintext), nil),
		N: nonce,
	})
	if err != nil {
		diags.AddError("Encryption Failed", fmt.Sprintf("failed to serialize: %v", err))
		return "", diags
	}

	return string(data), diags
}

// Decrypt decrypts JSON ciphertext using AES-256-GCM and returns plaintext.
func Decrypt(ctx context.Context, ciphertext string, key []byte) (string, diag.Diagnostics) {
	gcm, diags := newGCM(key)
	if diags.HasError() {
		return "", diags
	}

	var env envelope
	if err := json.Unmarshal([]byte(ciphertext), &env); err != nil {
		diags.AddError("Decryption Failed", fmt.Sprintf("invalid ciphertext: %v", err))
		return "", diags
	}

	if len(env.N) != gcm.NonceSize() {
		diags.AddError("Decryption Failed", "invalid nonce in ciphertext")
		return "", diags
	}

	plaintext, err := gcm.Open(nil, env.N, env.C, nil)
	if err != nil {
		diags.AddError("Decryption Failed", "wrong key or corrupted ciphertext")
		return "", diags
	}

	return string(plaintext), diags
}

// newGCM validates the key and creates an AES-256-GCM cipher.
func newGCM(key []byte) (cipher.AEAD, diag.Diagnostics) {
	var diags diag.Diagnostics

	if len(key) == 0 {
		diags.AddError("Key Missing", fmt.Sprintf("encryption_key_wo is required. Provide a %d-byte key from an ephemeral resource.", KeySize))
		return nil, diags
	}

	if len(key) != KeySize {
		diags.AddError("Invalid Key", fmt.Sprintf("encryption_key_wo must be %d bytes, got %d", KeySize, len(key)))
		return nil, diags
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		diags.AddError("Cipher Error", fmt.Sprintf("failed to create cipher: %v", err))
		return nil, diags
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		diags.AddError("Cipher Error", fmt.Sprintf("failed to create GCM: %v", err))
		return nil, diags
	}

	return gcm, diags
}
