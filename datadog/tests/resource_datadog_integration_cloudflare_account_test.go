package test

import (
	"context"
	"fmt"
	"testing"

	"github.com/terraform-providers/terraform-provider-datadog/datadog"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccIntegrationCloudflareAccountSimple(t *testing.T) {
	t.Parallel()
	ctx, accProviders := testAccProviders(context.Background(), t)
	uniq := uniqueEntityName(ctx, t)
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		CheckDestroy:      testAccCheckDatadogIntegrationCloudflareAccountDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogIntegrationCloudflareAccount(uniq),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogIntegrationCloudflareAccountExists(accProvider),
					resource.TestCheckResourceAttr(
						"datadog_body.foo", "api_key", "a94a8fe5ccb19ba61c4c0873d391e987982fbbd3"),
					resource.TestCheckResourceAttr(
						"datadog_body.foo", "email", "test-email@example.com"),
					resource.TestCheckResourceAttr(
						"datadog_body.foo", "name", "test-name"),
				),
			},
		},
	})
}

func testAccCheckDatadogIntegrationCloudflareAccount(uniq string) string {
	// Update me to make use of the unique value
	return fmt.Sprintf(`

resource "datadog_integration_cloudflare_account" "foo" {
    api_key = "a94a8fe5ccb19ba61c4c0873d391e987982fbbd3"
    email = "test-email@example.com"
    name = "test-name"
    
}
`, uniq)
}

func testAccCheckDatadogIntegrationCloudflareAccountDestroy(accProvider func() (*schema.Provider, error)) func(*terraform.State) error {
	return func(s *terraform.State) error {
		provider, _ := accProvider()
		providerConf := provider.Meta().(*datadog.ProviderConfiguration)
		apiInstances := providerConf.DatadogApiInstances
		auth := providerConf.Auth

		if err := IntegrationCloudflareAccountDestroyHelper(auth, s, apiInstances); err != nil {
			return err
		}
		return nil
	}
}

func IntegrationCloudflareAccountDestroyHelper(auth context.Context, s *terraform.State, apiInstances *utils.ApiInstances) error {
	err := utils.Retry(2, 10, func() error {
		for _, r := range s.RootModule().Resources {
			if r.Type != "resource_datadog_integration_cloudflare_account" {
				continue
			}
			id := r.Primary.ID

			_, httpResp, err := apiInstances.GetCloudflareIntegrationApiV2().GetCloudflareAccount(auth, id)
			if err != nil {
				if httpResp != nil && httpResp.StatusCode == 404 {
					return nil
				}
				return &utils.RetryableError{Prob: fmt.Sprintf("received an error retrieving Monitor %s", err)}
			}
			return &utils.RetryableError{Prob: "Monitor still exists"}
		}
		return nil
	})
	return err
}

func testAccCheckDatadogIntegrationCloudflareAccountExists(accProvider func() (*schema.Provider, error)) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		provider, _ := accProvider()
		providerConf := provider.Meta().(*datadog.ProviderConfiguration)
		apiInstances := providerConf.DatadogApiInstances
		auth := providerConf.Auth

		if err := integrationCloudflareAccountExistsHelper(auth, s, apiInstances); err != nil {
			return err
		}
		return nil
	}
}

func integrationCloudflareAccountExistsHelper(auth context.Context, s *terraform.State, apiInstances *utils.ApiInstances) error {
	for _, r := range s.RootModule().Resources {
		if r.Type != "resource_datadog_integration_cloudflare_account" {
			continue
		}
		id := r.Primary.ID

		_, httpResp, err := apiInstances.GetCloudflareIntegrationApiV2().GetCloudflareAccount(auth, id)
		if err != nil {
			return utils.TranslateClientError(err, httpResp, "error retrieving monitor")
		}
	}
	return nil
}
