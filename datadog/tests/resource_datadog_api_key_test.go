package test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/fwprovider"
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
				Config: testAccCheckDatadogApiKeyConfigRequired(apiKeyName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogApiKeyExists(providers.frameworkProvider, resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", apiKeyName),
					resource.TestCheckResourceAttrSet(resourceName, "key"),
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
			{
				Config: testAccCheckDatadogApiKeyConfigRequired(apiKeyNameUpdate),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogApiKeyExists(providers.frameworkProvider, resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", apiKeyNameUpdate),
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
				Config: testAccCheckDatadogApiKeyConfigRequired(apiKeyName),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckDatadogApiKeyConfigRequired(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_api_key" "foo" {
  name = "%s"
}`, uniq)
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
