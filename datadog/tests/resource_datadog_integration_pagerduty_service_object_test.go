package test

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
)

// We're not testing for schedules because Datadog actively verifies it with Pagerduty

func TestAccDatadogIntegrationPagerdutyServiceObject_Basic(t *testing.T) {
	t.Parallel()
	ctx, accProviders := testAccProviders(context.Background(), t)
	serviceName := strings.ReplaceAll(uniqueEntityName(ctx, t), "-", "_")
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		CheckDestroy:      testAccCheckDatadogIntegrationPagerdutyDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogIntegrationPagerdutyServiceObjectConfig(serviceName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogIntegrationPagerdutyExists(accProvider),
					resource.TestCheckResourceAttr(
						"datadog_integration_pagerduty.foo", "subdomain", "testdomain"),
					resource.TestCheckResourceAttr(
						"datadog_integration_pagerduty.foo", "api_token", "********************"),
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

func TestAccDatadogIntegrationPagerdutyServiceObject_WriteOnly(t *testing.T) {
	t.Parallel()
	ctx, accProviders := testAccProviders(context.Background(), t)
	serviceName := strings.ReplaceAll(uniqueEntityName(ctx, t), "-", "_")
	accProvider := testAccProvider(t, accProviders)
	resourceName := "datadog_integration_pagerduty_service_object.testing_wo"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		CheckDestroy:      testAccCheckDatadogIntegrationPagerdutyDestroy(accProvider),
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_11_0),
		},
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogIntegrationPagerdutyServiceObjectWriteOnlyConfig(serviceName, `
  service_key            = "9876543210123456789"
  service_key_wo         = "9876543210123456789"
  service_key_wo_version = "1"`),
				ExpectError: regexp.MustCompile("only one of `service_key,service_key_wo` can be specified"),
			},
			{
				// Start from the plaintext attribute to cover the migration to the write-only one.
				Config: testAccCheckDatadogIntegrationPagerdutyServiceObjectWriteOnlyConfig(serviceName, `
  service_key = "9876543210123456789"`),
				Check: resource.TestCheckResourceAttr(resourceName, "service_key", "9876543210123456789"),
			},
			{
				Config: testAccCheckDatadogIntegrationPagerdutyServiceObjectWriteOnlyConfig(serviceName, `
  service_key_wo         = "9876543210123456789_2"
  service_key_wo_version = "1"`),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "service_name", serviceName+"_wo"),
					resource.TestCheckResourceAttr(resourceName, "service_key", ""),
					resource.TestCheckNoResourceAttr(resourceName, "service_key_wo"),
					resource.TestCheckResourceAttr(resourceName, "service_key_wo_version", "1"),
				),
			},
			{
				Config: testAccCheckDatadogIntegrationPagerdutyServiceObjectWriteOnlyConfig(serviceName, `
  service_key_wo         = "9876543210123456789_3"
  service_key_wo_version = "2"`),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "service_key", ""),
					resource.TestCheckNoResourceAttr(resourceName, "service_key_wo"),
					resource.TestCheckResourceAttr(resourceName, "service_key_wo_version", "2"),
				),
			},
			{
				// Renaming replaces the object, which covers creating it with the write-only attribute.
				Config: testAccCheckDatadogIntegrationPagerdutyServiceObjectWriteOnlyConfig(serviceName+"_new", `
  service_key_wo         = "9876543210123456789_4"
  service_key_wo_version = "1"`),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "service_name", serviceName+"_new_wo"),
					resource.TestCheckNoResourceAttr(resourceName, "service_key"),
					resource.TestCheckNoResourceAttr(resourceName, "service_key_wo"),
					resource.TestCheckResourceAttr(resourceName, "service_key_wo_version", "1"),
				),
			},
		},
	})
}

func testAccCheckDatadogIntegrationPagerdutyServiceObjectWriteOnlyConfig(uniq string, serviceKeyAttrs string) string {
	return fmt.Sprintf(`
resource "datadog_integration_pagerduty_service_object" "testing_wo" {
  service_name = "%s_wo"
%s
}`, uniq, serviceKeyAttrs)
}

func testAccCheckDatadogIntegrationPagerdutyServiceObjectConfig(uniq string) string {
	return fmt.Sprintf(`
 resource "datadog_integration_pagerduty" "foo" {

  schedules = ["https://ddog.pagerduty.com/schedules/X123VF"]
  subdomain = "testdomain"
  api_token = "********************"
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

  schedules = ["https://ddog.pagerduty.com/schedules/X123VF"]
  subdomain = "testdomain"
  api_token = "********************"
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
