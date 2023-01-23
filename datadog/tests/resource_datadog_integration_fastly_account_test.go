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

func TestAccIntegrationFastlyAccountSimple(t *testing.T) {
	t.Parallel()
	ctx, accProviders := testAccProviders(context.Background(), t)
	uniq := uniqueEntityName(ctx, t)
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		CheckDestroy:      testAccCheckDatadogIntegrationFastlyAccountDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogIntegrationFastlyAccount(uniq),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogIntegrationFastlyAccountExists(accProvider),
					resource.TestCheckResourceAttr(
						"datadog_body.foo", "api_key", "ABCDEFG123"),
					resource.TestCheckResourceAttr(
						"datadog_body.foo", "name", "test-name"),
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
    name = "test-name"
    services {
    id = "6abc7de6893AbcDe9fghIj"
    tags = ["myTag", "myTag2:myValue"]
    }
    
}
`, uniq)
}

func testAccCheckDatadogIntegrationFastlyAccountDestroy(accProvider func() (*schema.Provider, error)) func(*terraform.State) error {
	return func(s *terraform.State) error {
		provider, _ := accProvider()
		providerConf := provider.Meta().(*datadog.ProviderConfiguration)
		apiInstances := providerConf.DatadogApiInstances
		auth := providerConf.Auth

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
				return &utils.RetryableError{Prob: fmt.Sprintf("received an error retrieving Monitor %s", err)}
			}
			return &utils.RetryableError{Prob: "Monitor still exists"}
		}
		return nil
	})
	return err
}

func testAccCheckDatadogIntegrationFastlyAccountExists(accProvider func() (*schema.Provider, error)) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		provider, _ := accProvider()
		providerConf := provider.Meta().(*datadog.ProviderConfiguration)
		apiInstances := providerConf.DatadogApiInstances
		auth := providerConf.Auth

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
			return utils.TranslateClientError(err, httpResp, "error retrieving monitor")
		}
	}
	return nil
}
