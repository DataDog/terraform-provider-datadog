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

func TestAccIntegrationConfluentAccountSimple(t *testing.T) {
	t.Parallel()
	ctx, accProviders := testAccProviders(context.Background(), t)
	uniq := uniqueEntityName(ctx, t)
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		CheckDestroy:      testAccCheckDatadogIntegrationConfluentAccountDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogIntegrationConfluentAccount(uniq),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogIntegrationConfluentAccountExists(accProvider),
					resource.TestCheckResourceAttr(
						"datadog_body.foo", "api_key", "TESTAPIKEY123"),
					resource.TestCheckResourceAttr(
						"datadog_body.foo", "api_secret", "test-api-secret-123"),
				),
			},
		},
	})
}

func testAccCheckDatadogIntegrationConfluentAccount(uniq string) string {
	// Update me to make use of the unique value
	return fmt.Sprintf(`

resource "datadog_integration_confluent_account" "foo" {
    api_key = "TESTAPIKEY123"
    api_secret = "test-api-secret-123"
    resources {
    id = "resource-id-123"
    resource_type = "kafka"
    tags = ["myTag", "myTag2:myValue"]
    }
    tags = ["myTag", "myTag2:myValue"]
    
}
`, uniq)
}

func testAccCheckDatadogIntegrationConfluentAccountDestroy(accProvider func() (*schema.Provider, error)) func(*terraform.State) error {
	return func(s *terraform.State) error {
		provider, _ := accProvider()
		providerConf := provider.Meta().(*datadog.ProviderConfiguration)
		apiInstances := providerConf.DatadogApiInstances
		auth := providerConf.Auth

		if err := IntegrationConfluentAccountDestroyHelper(auth, s, apiInstances); err != nil {
			return err
		}
		return nil
	}
}

func IntegrationConfluentAccountDestroyHelper(auth context.Context, s *terraform.State, apiInstances *utils.ApiInstances) error {
	err := utils.Retry(2, 10, func() error {
		for _, r := range s.RootModule().Resources {
			if r.Type != "resource_datadog_integration_confluent_account" {
				continue
			}
			id := r.Primary.ID

			_, httpResp, err := apiInstances.GetConfluentCloudApiV2().GetConfluentAccount(auth, id)
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

func testAccCheckDatadogIntegrationConfluentAccountExists(accProvider func() (*schema.Provider, error)) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		provider, _ := accProvider()
		providerConf := provider.Meta().(*datadog.ProviderConfiguration)
		apiInstances := providerConf.DatadogApiInstances
		auth := providerConf.Auth

		if err := integrationConfluentAccountExistsHelper(auth, s, apiInstances); err != nil {
			return err
		}
		return nil
	}
}

func integrationConfluentAccountExistsHelper(auth context.Context, s *terraform.State, apiInstances *utils.ApiInstances) error {
	for _, r := range s.RootModule().Resources {
		if r.Type != "resource_datadog_integration_confluent_account" {
			continue
		}
		id := r.Primary.ID

		_, httpResp, err := apiInstances.GetConfluentCloudApiV2().GetConfluentAccount(auth, id)
		if err != nil {
			return utils.TranslateClientError(err, httpResp, "error retrieving monitor")
		}
	}
	return nil
}
