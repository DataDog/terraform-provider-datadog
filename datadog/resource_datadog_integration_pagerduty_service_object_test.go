package datadog

import (
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

// We're not testing for schedules because Datadog actively verifies it with Pagerduty

func TestAccDatadogIntegrationPagerdutyServiceObject_Basic(t *testing.T) {
	accProviders, clock, cleanup := testAccProviders(t, initRecorder(t))
	serviceName := strings.ReplaceAll(uniqueEntityName(clock, t), "-", "_")
	defer cleanup(t)
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    accProviders,
		CheckDestroy: testAccCheckDatadogIntegrationPagerdutyDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogIntegrationPagerdutyServiceObjectConfig(serviceName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogIntegrationPagerdutyExists(accProvider),
					resource.TestCheckResourceAttr(
						"datadog_integration_pagerduty.foo", "subdomain", "testdomain"),
					resource.TestCheckResourceAttr(
						"datadog_integration_pagerduty.foo", "api_token", "*****"),
					resource.TestCheckResourceAttr(
						"datadog_integration_pagerduty.foo", "individual_services", "true"),
					resource.TestCheckResourceAttr(
						"datadog_integration_pagerduty.foo", "schedules.0", "https://ddog.pagerduty.com/schedules/X123VF"),
					resource.TestCheckResourceAttr(
						"datadog_integration_pagerduty_service_object.testing_foo", "service_name", serviceName+"_foo"),
					resource.TestCheckResourceAttr(
						"datadog_integration_pagerduty_service_object.testing_foo", "service_key", "9876543210123456789"),
					resource.TestCheckResourceAttr(
						"datadog_integration_pagerduty_service_object.testing_bar", "service_name", serviceName+"_bar"),
					resource.TestCheckResourceAttr(
						"datadog_integration_pagerduty_service_object.testing_bar", "service_key", "54321098765432109876"),
				),
			},
			{
				Config: testAccCheckDatadogIntegrationPagerdutyServiceObjectUpdatedConfig(serviceName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"datadog_integration_pagerduty_service_object.testing_foo", "service_name", serviceName+"_foo_2"),
					resource.TestCheckResourceAttr(
						"datadog_integration_pagerduty_service_object.testing_foo", "service_key", "9876543210123456789_2"),
					resource.TestCheckResourceAttr(
						"datadog_integration_pagerduty_service_object.testing_bar", "service_name", serviceName+"_bar_2"),
					resource.TestCheckResourceAttr(
						"datadog_integration_pagerduty_service_object.testing_bar", "service_key", "54321098765432109876_2"),
				),
			},
			{
				// make sure that updating the PD resource itself doesn't delete the individual service objects
				Config: strings.Replace(testAccCheckDatadogIntegrationPagerdutyServiceObjectUpdatedConfig(serviceName), "testdomain", "testdomain2", -1),
			},
		},
	})
}

func testAccCheckDatadogIntegrationPagerdutyServiceObjectConfig(uniq string) string {
	return fmt.Sprintf(`
 resource "datadog_integration_pagerduty" "foo" {
  individual_services = true

  schedules = ["https://ddog.pagerduty.com/schedules/X123VF"]
  subdomain = "testdomain"
  api_token = "*****"
 }

 resource "datadog_integration_pagerduty_service_object" "testing_foo" {
  # when creating the integration object for the first time, the service
  # objects have to be created *after* the integration
  depends_on = ["datadog_integration_pagerduty.foo"]
  service_name = "%s_foo"
  service_key  = "9876543210123456789"
}

resource "datadog_integration_pagerduty_service_object" "testing_bar" {
  depends_on = ["datadog_integration_pagerduty.foo"]
  service_name = "%s_bar"
  service_key  = "54321098765432109876"
}`, uniq, uniq)
}

func testAccCheckDatadogIntegrationPagerdutyServiceObjectUpdatedConfig(uniq string) string {
	return fmt.Sprintf(`
 resource "datadog_integration_pagerduty" "foo" {
  individual_services = true

  schedules = ["https://ddog.pagerduty.com/schedules/X123VF"]
  subdomain = "testdomain"
  api_token = "*****"
 }

 resource "datadog_integration_pagerduty_service_object" "testing_foo" {
  # when creating the integration object for the first time, the service
  # objects have to be created *after* the integration
  depends_on = ["datadog_integration_pagerduty.foo"]
  service_name = "%s_foo_2"
  service_key  = "9876543210123456789_2"
}

resource "datadog_integration_pagerduty_service_object" "testing_bar" {
  depends_on = ["datadog_integration_pagerduty.foo"]
  service_name = "%s_bar_2"
  service_key  = "54321098765432109876_2"
}`, uniq, uniq)
}
