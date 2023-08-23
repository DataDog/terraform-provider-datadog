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

func TestAccIntegrationFastlyAccountBasic(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	uniq := uniqueEntityName(ctx, t)

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogIntegrationFastlyAccountDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogIntegrationFastlyAccount(uniq),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogIntegrationFastlyAccountExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(
						"datadog_integration_fastly_account.foo", "api_key", "ABCDEFG123"),
					resource.TestCheckResourceAttr(
						"datadog_integration_fastly_account.foo", "name", uniq),
				),
			},
		},
	})
}

func testAccCheckDatadogIntegrationFastlyAccount(uniq string) string {
	// Update me to make use of the unique value
	return fmt.Sprintf(`
resource "datadog_integration_fastly_account" "foo" {
    api_key = "ABCDEFG123"
    name = "%s"
}`, uniq)
}

func testAccCheckDatadogIntegrationFastlyAccountDestroy(accProvider *fwprovider.FrameworkProvider) func(*terraform.State) error {
	return func(s *terraform.State) error {
		apiInstances := accProvider.DatadogApiInstances
		auth := accProvider.Auth

		if err := IntegrationFastlyAccountDestroyHelper(auth, s, apiInstances); err != nil {
			return err
		}
		return nil
	}
}

func IntegrationFastlyAccountDestroyHelper(auth context.Context, s *terraform.State, apiInstances *utils.ApiInstances) error {
	err := utils.Retry(2, 10, func() error {
		for _, r := range s.RootModule().Resources {
			if r.Type != "resource_datadog_integration_fastly_account" {
				continue
			}
			id := r.Primary.ID

			_, httpResp, err := apiInstances.GetFastlyIntegrationApiV2().GetFastlyAccount(auth, id)
			if err != nil {
				if httpResp != nil && httpResp.StatusCode == 404 {
					return nil
				}
				return &utils.RetryableError{Prob: fmt.Sprintf("received an error retrieving Fastly account %s", err)}
			}
			return &utils.RetryableError{Prob: "Fastly account still exists"}
		}
		return nil
	})
	return err
}

func testAccCheckDatadogIntegrationFastlyAccountExists(accProvider *fwprovider.FrameworkProvider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		apiInstances := accProvider.DatadogApiInstances
		auth := accProvider.Auth

		if err := integrationFastlyAccountExistsHelper(auth, s, apiInstances); err != nil {
			return err
		}
		return nil
	}
}

func integrationFastlyAccountExistsHelper(auth context.Context, s *terraform.State, apiInstances *utils.ApiInstances) error {
	for _, r := range s.RootModule().Resources {
		if r.Type != "resource_datadog_integration_fastly_account" {
			continue
		}
		id := r.Primary.ID

		_, httpResp, err := apiInstances.GetFastlyIntegrationApiV2().GetFastlyAccount(auth, id)
		if err != nil {
			return utils.TranslateClientError(err, httpResp, "error retrieving Fastly account")
		}
	}
	return nil
}
