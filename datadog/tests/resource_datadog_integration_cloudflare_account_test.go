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

func TestAccIntegrationCloudflareAccountBasic(t *testing.T) {
	t.Parallel()
	if !isReplaying() {
		t.Skip("This test is replay only")
	}
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	uniq := uniqueEntityName(ctx, t)

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogIntegrationCloudflareAccountDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogIntegrationCloudflareAccount(uniq),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogIntegrationCloudflareAccountExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(
						"datadog_integration_cloudflare_account.foo", "api_key", "1234567891012331asdd"),
					resource.TestCheckResourceAttr(
						"datadog_integration_cloudflare_account.foo", "email", "test-email@example.com"),
					resource.TestCheckResourceAttr(
						"datadog_integration_cloudflare_account.foo", "name", uniq),
				),
			},
		},
	})
}

func testAccCheckDatadogIntegrationCloudflareAccount(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_integration_cloudflare_account" "foo" {
    api_key = "1234567891012331asdd"
    email = "test-email@example.com"
    name = "%s"
}`, uniq)
}

func testAccCheckDatadogIntegrationCloudflareAccountDestroy(accProvider *fwprovider.FrameworkProvider) func(*terraform.State) error {
	return func(s *terraform.State) error {
		apiInstances := accProvider.DatadogApiInstances
		auth := accProvider.Auth

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
				return &utils.RetryableError{Prob: fmt.Sprintf("received an error retrieving cloudflare account %s", err)}
			}
			return &utils.RetryableError{Prob: "cloudflare account still exists"}
		}
		return nil
	})
	return err
}

func testAccCheckDatadogIntegrationCloudflareAccountExists(accProvider *fwprovider.FrameworkProvider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		apiInstances := accProvider.DatadogApiInstances
		auth := accProvider.Auth

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
			return utils.TranslateClientError(err, httpResp, "error retrieving cloudflare account")
		}
	}
	return nil
}
