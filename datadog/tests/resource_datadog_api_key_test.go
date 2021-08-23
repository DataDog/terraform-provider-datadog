package test

import (
	"context"
	"fmt"
	"testing"

	"github.com/terraform-providers/terraform-provider-datadog/datadog"

	datadogV2 "github.com/DataDog/datadog-api-client-go/api/v2/datadog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccDatadogApiKey_Update(t *testing.T) {
	if isRecording() || isReplaying() {
		t.Skip("This test doesn't support recording or replaying")
	}
	t.Parallel()
	ctx, accProviders := testAccProviders(context.Background(), t)
	accProvider := testAccProvider(t, accProviders)
	apiKeyName := uniqueEntityName(ctx, t)
	apiKeyNameUpdate := apiKeyName + "-2"
	resourceName := "datadog_api_key.foo"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		CheckDestroy:      testAccCheckDatadogApiKeyDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogApiKeyConfigRequired(apiKeyName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogApiKeyExists(accProvider, resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", apiKeyName),
					testAccCheckDatadogApiKeyValueMatches(accProvider, resourceName),
				),
			},
			{
				Config: testAccCheckDatadogApiKeyConfigRequired(apiKeyNameUpdate),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogApiKeyExists(accProvider, resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", apiKeyNameUpdate),
					testAccCheckDatadogApiKeyValueMatches(accProvider, resourceName),
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
	ctx, accProviders := testAccProviders(context.Background(), t)
	apiKeyName := uniqueEntityName(ctx, t)
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		CheckDestroy:      testAccCheckDatadogApiKeyDestroy(accProvider),
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

func testAccCheckDatadogApiKeyExists(accProvider func() (*schema.Provider, error), n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		provider, _ := accProvider()
		providerConf := provider.Meta().(*datadog.ProviderConfiguration)
		datadogClientV2 := providerConf.DatadogClientV2
		authV2 := providerConf.AuthV2

		if err := datadogApiKeyExistsHelper(authV2, s, datadogClientV2, n); err != nil {
			return err
		}
		return nil
	}
}

func datadogApiKeyExistsHelper(ctx context.Context, s *terraform.State, client *datadogV2.APIClient, name string) error {
	id := s.RootModule().Resources[name].Primary.ID
	if _, _, err := client.KeyManagementApi.GetAPIKey(ctx, id); err != nil {
		return fmt.Errorf("received an error retrieving api key %s", err)
	}
	return nil
}

func testAccCheckDatadogApiKeyValueMatches(accProvider func() (*schema.Provider, error), n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		provider, _ := accProvider()
		providerConf := provider.Meta().(*datadog.ProviderConfiguration)
		datadogClientV2 := providerConf.DatadogClientV2
		authV2 := providerConf.AuthV2

		if err := datadogApiKeyValueMatches(authV2, s, datadogClientV2, n); err != nil {
			return err
		}
		return nil
	}
}

func datadogApiKeyValueMatches(ctx context.Context, s *terraform.State, client *datadogV2.APIClient, name string) error {
	primaryResource := s.RootModule().Resources[name].Primary
	id := primaryResource.ID
	expectedKey := primaryResource.Attributes["key"]
	resp, _, err := client.KeyManagementApi.GetAPIKey(ctx, id)
	if err != nil {
		return fmt.Errorf("received an error retrieving api key %s", err)
	}
	actualKey := resp.Data.Attributes.GetKey()
	if expectedKey != actualKey {
		return fmt.Errorf("api key value does not match")
	}
	return nil
}

func testAccCheckDatadogApiKeyDestroy(accProvider func() (*schema.Provider, error)) func(*terraform.State) error {
	return func(s *terraform.State) error {
		provider, _ := accProvider()
		providerConf := provider.Meta().(*datadog.ProviderConfiguration)
		datadogClientV2 := providerConf.DatadogClientV2
		authV2 := providerConf.AuthV2

		if err := datadogApiKeyDestroyHelper(authV2, s, datadogClientV2); err != nil {
			return err
		}
		return nil
	}
}

func datadogApiKeyDestroyHelper(ctx context.Context, s *terraform.State, client *datadogV2.APIClient) error {
	for _, r := range s.RootModule().Resources {
		if r.Type != "datadog_api_key" {
			continue
		}

		id := r.Primary.ID
		_, httpResponse, err := client.KeyManagementApi.GetAPIKey(ctx, id)

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
