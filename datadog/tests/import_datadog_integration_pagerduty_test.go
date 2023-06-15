package test

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestDatadogIntegrationPagerduty_import(t *testing.T) {
	t.Parallel()
	resourceName := "datadog_integration_pagerduty.pd"
	_, accProviders := testAccProviders(context.Background(), t)
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		CheckDestroy:      testAccCheckDatadogIntegrationPagerdutyDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogIntegrationPagerdutyConfigImported(),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"api_token"},
			},
		},
	})
}

func testAccCheckDatadogIntegrationPagerdutyConfigImported() string {
	return `
resource "datadog_integration_pagerduty" "pd" {
  schedules = ["https://ddog.pagerduty.com/schedules/X123VF"]
  subdomain = "testdomain"
  api_token = "********************"
}`
}
