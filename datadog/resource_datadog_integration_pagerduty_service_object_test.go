package datadog

import (
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

// We're not testing for schedules because Datadog actively verifies it with Pagerduty

func TestAccDatadogIntegrationPagerdutyServiceObject_Basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckDatadogIntegrationPagerdutyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogIntegrationPagerdutyServiceObjectConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogIntegrationPagerdutyExists("datadog_integration_pagerduty.foo"),
					resource.TestCheckResourceAttr(
						"datadog_integration_pagerduty.foo", "subdomain", "testdomain"),
					resource.TestCheckResourceAttr(
						"datadog_integration_pagerduty.foo", "api_token", "*****"),
					resource.TestCheckResourceAttr(
						"datadog_integration_pagerduty.foo", "individual_services", "true"),
					resource.TestCheckResourceAttr(
						"datadog_integration_pagerduty_service_object.testing_foo", "service_name", "testing_foo"),
					resource.TestCheckResourceAttr(
						"datadog_integration_pagerduty_service_object.testing_foo", "service_key", "9876543210123456789"),
					resource.TestCheckResourceAttr(
						"datadog_integration_pagerduty_service_object.testing_bar", "service_name", "testing_bar"),
					resource.TestCheckResourceAttr(
						"datadog_integration_pagerduty_service_object.testing_bar", "service_key", "54321098765432109876"),
				),
			},
			{
				Config: testAccCheckDatadogIntegrationPagerdutyServiceObjectUpdatedConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"datadog_integration_pagerduty_service_object.testing_foo", "service_name", "testing_foo_2"),
					resource.TestCheckResourceAttr(
						"datadog_integration_pagerduty_service_object.testing_foo", "service_key", "9876543210123456789_2"),
					resource.TestCheckResourceAttr(
						"datadog_integration_pagerduty_service_object.testing_bar", "service_name", "testing_bar_2"),
					resource.TestCheckResourceAttr(
						"datadog_integration_pagerduty_service_object.testing_bar", "service_key", "54321098765432109876_2"),
				),
			},
			{
				// make sure that updating the PD resource itself doesn't delete the individual service objects
				Config: strings.Replace(testAccCheckDatadogIntegrationPagerdutyServiceObjectUpdatedConfig, "testdomain", "testdomain2", -1),
			},
		},
	})
}

const testAccCheckDatadogIntegrationPagerdutyServiceObjectConfig = `
 resource "datadog_integration_pagerduty" "foo" {
	individual_services = true

  subdomain = "testdomain"
  api_token = "*****"
 }

 resource "datadog_integration_pagerduty_service_object" "testing_foo" {
  # when creating the integration object for the first time, the service
  # objects have to be created *after* the integration
  depends_on = ["datadog_integration_pagerduty.foo"]
  service_name = "testing_foo"
  service_key  = "9876543210123456789"
}

resource "datadog_integration_pagerduty_service_object" "testing_bar" {
  depends_on = ["datadog_integration_pagerduty.foo"]
  service_name = "testing_bar"
  service_key  = "54321098765432109876"
}
`

const testAccCheckDatadogIntegrationPagerdutyServiceObjectUpdatedConfig = `
 resource "datadog_integration_pagerduty" "foo" {
	individual_services = true

  subdomain = "testdomain"
  api_token = "*****"
 }

 resource "datadog_integration_pagerduty_service_object" "testing_foo" {
  # when creating the integration object for the first time, the service
  # objects have to be created *after* the integration
  depends_on = ["datadog_integration_pagerduty.foo"]
  service_name = "testing_foo_2"
  service_key  = "9876543210123456789_2"
}

resource "datadog_integration_pagerduty_service_object" "testing_bar" {
  depends_on = ["datadog_integration_pagerduty.foo"]
  service_name = "testing_bar_2"
  service_key  = "54321098765432109876_2"
}
`
