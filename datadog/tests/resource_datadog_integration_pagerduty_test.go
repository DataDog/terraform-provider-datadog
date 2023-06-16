package test

import (
	"context"
	"fmt"
	"testing"

	"github.com/terraform-providers/terraform-provider-datadog/datadog"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	communityClient "github.com/zorkian/go-datadog-api"
)

// We're not testing for schedules because Datadog actively verifies it with Pagerduty

func TestAccDatadogIntegrationPagerduty_Basic(t *testing.T) {
	t.Parallel()
	_, accProviders := testAccProviders(context.Background(), t)
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		CheckDestroy:      testAccCheckDatadogIntegrationPagerdutyDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogIntegrationPagerdutyConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogIntegrationPagerdutyExists(accProvider),
					resource.TestCheckResourceAttr(
						"datadog_integration_pagerduty.foo", "subdomain", "testdomain"),
					resource.TestCheckResourceAttr(
						"datadog_integration_pagerduty.foo", "api_token", "abcdefghijklmnopqrst"),
					resource.TestCheckResourceAttr(
						"datadog_integration_pagerduty.foo", "schedules.0", "https://ddog.pagerduty.com/schedules/X123VF"),
				),
			},
		},
	})
}

func testAccCheckDatadogIntegrationPagerdutyExists(accProvider func() (*schema.Provider, error)) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		provider, _ := accProvider()
		providerConf := provider.Meta().(*datadog.ProviderConfiguration)
		client := providerConf.CommunityClient

		if err := datadogIntegrationPagerdutyExistsHelper(client); err != nil {
			return err
		}
		return nil
	}
}

func datadogIntegrationPagerdutyExistsHelper(client *communityClient.Client) error {
	if _, err := client.GetIntegrationPD(); err != nil {
		return fmt.Errorf("received an error retrieving integration pagerduty %s", err)
	}
	return nil
}

func testAccCheckDatadogIntegrationPagerdutyDestroy(accProvider func() (*schema.Provider, error)) func(*terraform.State) error {
	return func(s *terraform.State) error {
		// We don't destroy the integration anymore, so let's not check anything
		return nil
	}
}

// NOTE: Don't create configs with no schedules. If there's a leftover schedule from some previous test run,
// the test will fail on non-empty diff after apply, as the resource is (unfortunately) created in a way
// to not overwrite existing schedules with empty list on creation.
func testAccCheckDatadogIntegrationPagerdutyConfig() string {
	return `
 resource "datadog_integration_pagerduty" "foo" {
   schedules = ["https://ddog.pagerduty.com/schedules/X123VF"]
   subdomain = "testdomain"
   api_token = "abcdefghijklmnopqrst"
 }`
}
