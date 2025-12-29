package test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/fwprovider"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/secretbridge"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

func TestAccDatadogApiKey_Update(t *testing.T) {
	if isRecording() || isReplaying() {
		t.Skip("This test doesn't support recording or replaying")
	}
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	apiKeyName := uniqueEntityName(ctx, t)
	apiKeyNameUpdate := apiKeyName + "-2"
	resourceName := "datadog_api_key.foo"
	var apiKey string

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogApiKeyDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogApiKeyConfigRequired(apiKeyName, nil),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogApiKeyExists(providers.frameworkProvider, resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", apiKeyName),
					resource.TestCheckResourceAttrSet(resourceName, "key"),
					// Without encryption_key_wo, encrypted_key should be empty
					resource.TestCheckResourceAttr(resourceName, "encrypted_key", ""),
					resource.TestCheckResourceAttr(resourceName, "remote_config_read_enabled", "true"),
					func(s *terraform.State) error {
						resource, ok := s.RootModule().Resources[resourceName]
						if !ok {
							return fmt.Errorf("Resource not found: %s", resourceName)
						}

						apiKey = resource.Primary.Attributes["key"]
						return nil
					},
				),
			},
			// Test will succeed only if Remote config is configured for this organisation (see /organization-settings/remote-config)
			{
				Config: testAccCheckDatadogApiKeyConfigRequired(apiKeyNameUpdate, Ptr(true)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogApiKeyExists(providers.frameworkProvider, resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", apiKeyNameUpdate),
					resource.TestCheckResourceAttr(resourceName, "remote_config_read_enabled", "true"),
					func(s *terraform.State) error {
						resource, ok := s.RootModule().Resources[resourceName]
						if !ok {
							return fmt.Errorf("Resource not found: %s", resourceName)
						}

						stateAPIKey := resource.Primary.Attributes["key"]
						if stateAPIKey != apiKey {
							return fmt.Errorf("API key (%s) does not match expected value (%s)", stateAPIKey, apiKey)
						}
						return nil
					},
				),
			},
			{
				Config: testAccCheckDatadogApiKeyConfigRequired(apiKeyNameUpdate, Ptr(false)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogApiKeyExists(providers.frameworkProvider, resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", apiKeyNameUpdate),
					resource.TestCheckResourceAttr(resourceName, "remote_config_read_enabled", "false"),
					func(s *terraform.State) error {
						resource, ok := s.RootModule().Resources[resourceName]
						if !ok {
							return fmt.Errorf("Resource not found: %s", resourceName)
						}

						stateAPIKey := resource.Primary.Attributes["key"]
						if stateAPIKey != apiKey {
							return fmt.Errorf("API key (%s) does not match expected value (%s)", stateAPIKey, apiKey)
						}
						return nil
					},
				),
			},
		},
	})
}

func TestDatadogApiKey_import(t *testing.T) {
	if isRecording() || isReplaying() {
		t.Skip("This test doesn't support recording or replaying")
	}
	t.Parallel()
	resourceName := "datadog_api_key.foo"
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	apiKeyName := uniqueEntityName(ctx, t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogApiKeyDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogApiKeyConfigRequired(apiKeyName, nil),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckDatadogApiKeyConfigRequired(uniq string, remoteConfig *bool) string {
	remoteConfigParam := ""
	if remoteConfig != nil {
		remoteConfigParam = fmt.Sprintf("remote_config_read_enabled = %t", *remoteConfig)
	}
	return fmt.Sprintf(`
resource "datadog_api_key" "foo" {
  name = "%s"
  %s
}`, uniq, remoteConfigParam)
}

func testAccCheckDatadogApiKeyExists(accProvider *fwprovider.FrameworkProvider, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		apiInstances := accProvider.DatadogApiInstances
		auth := accProvider.Auth

		if err := datadogApiKeyExistsHelper(auth, s, apiInstances, n); err != nil {
			return err
		}
		return nil
	}
}

func datadogApiKeyExistsHelper(ctx context.Context, s *terraform.State, apiInstances *utils.ApiInstances, name string) error {
	id := s.RootModule().Resources[name].Primary.ID
	if _, _, err := apiInstances.GetKeyManagementApiV2().GetAPIKey(ctx, id); err != nil {
		return fmt.Errorf("received an error retrieving api key %s", err)
	}
	return nil
}

func testAccCheckDatadogApiKeyDestroy(accProvider *fwprovider.FrameworkProvider) func(*terraform.State) error {
	return func(s *terraform.State) error {
		apiInstances := accProvider.DatadogApiInstances
		auth := accProvider.Auth

		if err := datadogApiKeyDestroyHelper(auth, s, apiInstances); err != nil {
			return err
		}
		return nil
	}
}

func datadogApiKeyDestroyHelper(ctx context.Context, s *terraform.State, apiInstances *utils.ApiInstances) error {
	for _, r := range s.RootModule().Resources {
		if r.Type != "datadog_api_key" {
			continue
		}

		id := r.Primary.ID
		_, httpResponse, err := apiInstances.GetKeyManagementApiV2().GetAPIKey(ctx, id)

		if err != nil {
			if httpResponse.StatusCode == 404 {
				continue
			}
			return fmt.Errorf("received an error retrieving api key %s", err)
		}

		return fmt.Errorf("api key still exists")
	}

	return nil
}

func Ptr[T any](v T) *T {
	return &v
}

// TestAccDatadogApiKey_WithEncryption tests that when encryption_key_wo is provided,
// the API key is encrypted and stored in encrypted_key, while key remains null.
func TestAccDatadogApiKey_WithEncryption(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	apiKeyName := uniqueEntityName(ctx, t)
	resourceName := "datadog_api_key.foo"
	// 32-byte key for AES-256
	encryptionKey := "01234567890123456789012345678901"

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogApiKeyDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogApiKeyConfigWithEncryption(apiKeyName, encryptionKey),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogApiKeyExists(providers.frameworkProvider, resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", apiKeyName),
					// When encryption is enabled, key should be null and encrypted_key should be set
					resource.TestCheckNoResourceAttr(resourceName, "key"),
					resource.TestCheckResourceAttrSet(resourceName, "encrypted_key"),
					// Verify the encrypted_key can be decrypted
					testAccCheckDatadogApiKeyEncryptedKeyValid(resourceName, encryptionKey),
				),
			},
		},
	})
}

func testAccCheckDatadogApiKeyConfigWithEncryption(name, encryptionKey string) string {
	return fmt.Sprintf(`
resource "datadog_api_key" "foo" {
  name              = "%s"
  encryption_key_wo = "%s"
}`, name, encryptionKey)
}

// testAccCheckDatadogApiKeyEncryptedKeyValid verifies that the encrypted_key attribute
// contains valid ciphertext that can be decrypted with the provided key.
func testAccCheckDatadogApiKeyEncryptedKeyValid(resourceName, encryptionKey string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}

		encryptedKey := rs.Primary.Attributes["encrypted_key"]
		if encryptedKey == "" {
			return fmt.Errorf("encrypted_key is empty")
		}

		// Verify the ciphertext is valid JSON (secretbridge format)
		if !strings.HasPrefix(encryptedKey, "{") {
			return fmt.Errorf("encrypted_key doesn't look like valid secretbridge ciphertext")
		}

		// Decrypt and verify we get a non-empty result
		plaintext, diags := secretbridge.Decrypt(context.Background(), encryptedKey, []byte(encryptionKey))
		if diags.HasError() {
			return fmt.Errorf("failed to decrypt encrypted_key: %v", diags.Errors())
		}

		if plaintext == "" {
			return fmt.Errorf("decrypted key is empty")
		}

		return nil
	}
}
