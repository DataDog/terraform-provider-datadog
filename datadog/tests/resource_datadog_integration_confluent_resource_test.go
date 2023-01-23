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

func TestAccIntegrationConfluentResourceSimple(t *testing.T) {
	t.Parallel()
	ctx, accProviders := testAccProviders(context.Background(), t)
	uniq := uniqueEntityName(ctx, t)
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		CheckDestroy:      testAccCheckDatadogIntegrationConfluentResourceDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogIntegrationConfluentResource(uniq),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogIntegrationConfluentResourceExists(accProvider),

					resource.TestCheckResourceAttr(
						"datadog_body.foo", "resource_type", "kafka"),
				),
			},
		},
	})
}

func testAccCheckDatadogIntegrationConfluentResource(uniq string) string {
	// Update me to make use of the unique value
	return fmt.Sprintf(`

resource "datadog_integration_confluent_resource" "foo" {
    account_id = "UPDATE ME"
    
    resource_type = "kafka"
    tags = ["myTag", "myTag2:myValue"]
    
}
`, uniq)
}

func testAccCheckDatadogIntegrationConfluentResourceDestroy(accProvider func() (*schema.Provider, error)) func(*terraform.State) error {
	return func(s *terraform.State) error {
		provider, _ := accProvider()
		providerConf := provider.Meta().(*datadog.ProviderConfiguration)
		apiInstances := providerConf.DatadogApiInstances
		auth := providerConf.Auth

		if err := IntegrationConfluentResourceDestroyHelper(auth, s, apiInstances); err != nil {
			return err
		}
		return nil
	}
}

func IntegrationConfluentResourceDestroyHelper(auth context.Context, s *terraform.State, apiInstances *utils.ApiInstances) error {
	err := utils.Retry(2, 10, func() error {
		for _, r := range s.RootModule().Resources {
			if r.Type != "resource_datadog_integration_confluent_resource" {
				continue
			}
			accountId := r.Primary.Attributes["account_id"]
			id := r.Primary.ID

			_, httpResp, err := apiInstances.GetConfluentCloudApiV2().GetConfluentResource(auth, accountId, id)
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

func testAccCheckDatadogIntegrationConfluentResourceExists(accProvider func() (*schema.Provider, error)) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		provider, _ := accProvider()
		providerConf := provider.Meta().(*datadog.ProviderConfiguration)
		apiInstances := providerConf.DatadogApiInstances
		auth := providerConf.Auth

		if err := integrationConfluentResourceExistsHelper(auth, s, apiInstances); err != nil {
			return err
		}
		return nil
	}
}

func integrationConfluentResourceExistsHelper(auth context.Context, s *terraform.State, apiInstances *utils.ApiInstances) error {
	for _, r := range s.RootModule().Resources {
		if r.Type != "resource_datadog_integration_confluent_resource" {
			continue
		}
		accountId := r.Primary.Attributes["account_id"]
		id := r.Primary.ID

		_, httpResp, err := apiInstances.GetConfluentCloudApiV2().GetConfluentResource(auth, accountId, id)
		if err != nil {
			return utils.TranslateClientError(err, httpResp, "error retrieving monitor")
		}
	}
	return nil
}
