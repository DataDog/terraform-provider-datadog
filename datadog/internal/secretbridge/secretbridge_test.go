package secretbridge

import (
	"context"
	"strings"
	"testing"
)

func TestEncryptDecrypt_RoundTrip(t *testing.T) {
	ctx := context.Background()
	key := []byte("01234567890123456789012345678901") // 32 bytes
	plaintext := "my-secret-api-key-value"

	// Encrypt
	ciphertext, diags := Encrypt(ctx, plaintext, key)
	if diags.HasError() {
		t.Fatalf("Encrypt failed: %v", diags.Errors())
	}
	if ciphertext == "" {
		t.Fatal("Encrypt returned empty ciphertext")
	}
	if ciphertext == plaintext {
		t.Fatal("Ciphertext should not equal plaintext")
	}

	// Decrypt
	decrypted, diags := Decrypt(ctx, ciphertext, key)
	if diags.HasError() {
		t.Fatalf("Decrypt failed: %v", diags.Errors())
	}
	if decrypted != plaintext {
		t.Errorf("Decrypted value %q does not match original %q", decrypted, plaintext)
	}
}

func TestEncrypt_InvalidKeySize(t *testing.T) {
	ctx := context.Background()
	tests := []struct {
		name    string
		key     []byte
		wantErr string
	}{
		{
			name:    "empty key",
			key:     []byte{},
			wantErr: "encryption_key_wo is required",
		},
		{
			name:    "key too short",
			key:     []byte("short"),
			wantErr: "must be 32 bytes",
		},
		{
			name:    "key too long",
			key:     []byte("this-key-is-way-too-long-for-aes-256-encryption"),
			wantErr: "must be 32 bytes",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, diags := Encrypt(ctx, "plaintext", tt.key)
			if !diags.HasError() {
				t.Fatal("Expected error but got none")
			}
			errMsg := diags.Errors()[0].Detail()
			if !strings.Contains(errMsg, tt.wantErr) {
				t.Errorf("Error %q should contain %q", errMsg, tt.wantErr)
			}
		})
	}
}

func TestDecrypt_InvalidCiphertext(t *testing.T) {
	ctx := context.Background()
	key := []byte("01234567890123456789012345678901")

	tests := []struct {
		name       string
		ciphertext string
		wantErr    string
	}{
		{
			name:       "invalid json",
			ciphertext: "not-json",
			wantErr:    "invalid ciphertext",
		},
		{
			name:       "empty json",
			ciphertext: "{}",
			wantErr:    "invalid nonce",
		},
		{
			name:       "wrong nonce size",
			ciphertext: `{"c":"YWJj","n":"YWI="}`,
			wantErr:    "invalid nonce",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, diags := Decrypt(ctx, tt.ciphertext, key)
			if !diags.HasError() {
				t.Fatal("Expected error but got none")
			}
			errMsg := diags.Errors()[0].Detail()
			if !strings.Contains(errMsg, tt.wantErr) {
				t.Errorf("Error %q should contain %q", errMsg, tt.wantErr)
			}
		})
	}
}

func TestDecrypt_WrongKey(t *testing.T) {
	ctx := context.Background()
	encryptKey := []byte("01234567890123456789012345678901")
	decryptKey := []byte("98765432109876543210987654321098")
	plaintext := "secret-value"

	ciphertext, diags := Encrypt(ctx, plaintext, encryptKey)
	if diags.HasError() {
		t.Fatalf("Encrypt failed: %v", diags.Errors())
	}

	_, diags = Decrypt(ctx, ciphertext, decryptKey)
	if !diags.HasError() {
		t.Fatal("Expected decryption to fail with wrong key")
	}
	errMsg := diags.Errors()[0].Detail()
	if !strings.Contains(errMsg, "wrong key") {
		t.Errorf("Error should mention wrong key, got: %q", errMsg)
	}
}

func TestEncrypt_DifferentCiphertextsForSamePlaintext(t *testing.T) {
	ctx := context.Background()
	key := []byte("01234567890123456789012345678901")
	plaintext := "same-plaintext"

	ciphertext1, _ := Encrypt(ctx, plaintext, key)
	ciphertext2, _ := Encrypt(ctx, plaintext, key)

	if ciphertext1 == ciphertext2 {
		t.Error("Same plaintext should produce different ciphertexts (random nonce)")
	}

	// Both should decrypt to the same value
	decrypted1, _ := Decrypt(ctx, ciphertext1, key)
	decrypted2, _ := Decrypt(ctx, ciphertext2, key)

	if decrypted1 != plaintext || decrypted2 != plaintext {
		t.Error("Both ciphertexts should decrypt to original plaintext")
	}
}
