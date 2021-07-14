package test

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/terraform-providers/terraform-provider-datadog/datadog"
	"testing"

	datadogV2 "github.com/DataDog/datadog-api-client-go/api/v2/datadog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

const datadogApiKeyResource = `
resource "datadog_api_key" "foo" {
  name = "test-api-key"
}
`

func TestAccDatadogApiKeyCreate(t *testing.T) {
	t.Parallel()
	_, accProviders := testAccProviders(context.Background(), t)
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		CheckDestroy:      testAccCheckDatadogApiKeyDestroy(accProvider, "datadog_api_key.foo"),
		Steps: []resource.TestStep{
			{
				Config: datadogApiKeyResource,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogApiKeyExists(accProvider, "datadog_api_key.foo"),
					resource.TestCheckResourceAttr("datadog_api_key.foo", "name", "test-api-key"),
					testAccCheckDatadogApiKeyValueMatches(accProvider, "datadog_api_key.foo"),
				),
			},
		},
	})
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
		return fmt.Errorf("received an error retrieving API Key %s", err)
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
		return fmt.Errorf("received an error retrieving API Key %s", err)
	}
	actualKey := resp.Data.Attributes.GetKey()
	if expectedKey != actualKey {
		return fmt.Errorf("API key value does not match")
	}
	return nil
}

func testAccCheckDatadogApiKeyDestroy(accProvider func() (*schema.Provider, error), name string) func(*terraform.State) error {
	return func(s *terraform.State) error {
		provider, _ := accProvider()
		providerConf := provider.Meta().(*datadog.ProviderConfiguration)
		datadogClientV2 := providerConf.DatadogClientV2
		authV2 := providerConf.AuthV2

		if err := datadogApiKeyDestroyHelper(authV2, s, datadogClientV2, name); err != nil {
			return err
		}
		return nil
	}
}

func datadogApiKeyDestroyHelper(ctx context.Context, s *terraform.State, client *datadogV2.APIClient, name string) error {
	id := s.RootModule().Resources[name].Primary.ID
	_, httpResponse, err := client.KeyManagementApi.GetAPIKey(ctx, id)

	if httpResponse == nil {
		return fmt.Errorf("received an error checking API key status %s", err)
	}

	if httpResponse.StatusCode != 404 {
		return fmt.Errorf("expected 404 for deleted API key, got %d", httpResponse.StatusCode)
	}

	return nil
}
