package datadog

import (
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/zorkian/go-datadog-api"
)

// We're not testing for schedules because Datadog actively verifies it with Pagerduty

func TestAccDatadogIntegrationPagerduty_Basic(t *testing.T) {
	accProviders, cleanup := testAccProviders(t)
	defer cleanup(t)
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    accProviders,
		CheckDestroy: testAccCheckDatadogIntegrationPagerdutyDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogIntegrationPagerdutyConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogIntegrationPagerdutyExists(accProvider),
					resource.TestCheckResourceAttr(
						"datadog_integration_pagerduty.foo", "subdomain", "testdomain"),
					resource.TestCheckResourceAttr(
						"datadog_integration_pagerduty.foo", "api_token", "secret"),
					resource.TestCheckResourceAttr(
						"datadog_integration_pagerduty.foo", "services.0.service_name", "test_service"),
					resource.TestCheckResourceAttr(
						"datadog_integration_pagerduty.foo", "services.0.service_key", "*****"),
				),
			},
		},
	})
}

func TestAccDatadogIntegrationPagerduty_TwoServices(t *testing.T) {
	accProviders, cleanup := testAccProviders(t)
	defer cleanup(t)
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    accProviders,
		CheckDestroy: testAccCheckDatadogIntegrationPagerdutyDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogIntegrationPagerdutyConfigTwoServices,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogIntegrationPagerdutyExists(accProvider),
					resource.TestCheckResourceAttr(
						"datadog_integration_pagerduty.foo", "subdomain", "testdomain"),
					resource.TestCheckResourceAttr(
						"datadog_integration_pagerduty.foo", "api_token", "secret"),
					resource.TestCheckResourceAttr(
						"datadog_integration_pagerduty.foo", "services.0.service_name", "test_service"),
					resource.TestCheckResourceAttr(
						"datadog_integration_pagerduty.foo", "services.0.service_key", "*****"),
					resource.TestCheckResourceAttr(
						"datadog_integration_pagerduty.foo", "services.1.service_name", "test_service_2"),
					resource.TestCheckResourceAttr(
						"datadog_integration_pagerduty.foo", "services.1.service_key", "*****"),
				),
			},
		},
	})
}

func TestAccDatadogIntegrationPagerduty_Migrate2ServiceObjects(t *testing.T) {
	accProviders, cleanup := testAccProviders(t)
	defer cleanup(t)
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    accProviders,
		CheckDestroy: testAccCheckDatadogIntegrationPagerdutyDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogIntegrationPagerdutyConfigBeforeMigration,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogIntegrationPagerdutyExists(accProvider),
					resource.TestCheckResourceAttr(
						"datadog_integration_pagerduty.pd", "subdomain", "ddog"),
					resource.TestCheckResourceAttr(
						"datadog_integration_pagerduty.pd", "api_token", "secret"),
					resource.TestCheckResourceAttr(
						"datadog_integration_pagerduty.pd", "services.0.service_name", "testing_bar"),
					resource.TestCheckResourceAttr(
						"datadog_integration_pagerduty.pd", "services.0.service_key", "*****"),
					resource.TestCheckResourceAttr(
						"datadog_integration_pagerduty.pd", "services.1.service_name", "testing_foo"),
					resource.TestCheckResourceAttr(
						"datadog_integration_pagerduty.pd", "services.1.service_key", "*****"),
				),
			},
			{
				// this represents the intermediary step which will ensure the old
				// inline-defined service objects get removed
				Config: testAccCheckDatadogIntegrationPagerdutyConfigDuringMigration,
			},
			{
				Config: testAccCheckDatadogIntegrationPagerdutyConfigAfterMigration,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogIntegrationPagerdutyExists(accProvider),
					resource.TestCheckResourceAttr(
						"datadog_integration_pagerduty.pd", "subdomain", "ddog"),
					resource.TestCheckResourceAttr(
						"datadog_integration_pagerduty.pd", "api_token", "secret"),
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
		},
	})
}

func testAccCheckDatadogIntegrationPagerdutyExists(accProvider *schema.Provider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		providerConf := accProvider.Meta().(*ProviderConfiguration)
		client := providerConf.CommunityClient

		if err := datadogIntegrationPagerdutyExistsHelper(client); err != nil {
			return err
		}
		return nil
	}
}

func datadogIntegrationPagerdutyExistsHelper(client *datadog.Client) error {
	if _, err := client.GetIntegrationPD(); err != nil {
		return fmt.Errorf("received an error retrieving integration pagerduty %s", err)
	}
	return nil
}

func testAccCheckDatadogIntegrationPagerdutyDestroy(accProvider *schema.Provider) func(*terraform.State) error {
	return func(s *terraform.State) error {
		providerConf := accProvider.Meta().(*ProviderConfiguration)
		client := providerConf.CommunityClient

		_, err := client.GetIntegrationPD()
		if err != nil {
			if strings.Contains(err.Error(), "pagerduty not found") {
				return nil
			}

			return fmt.Errorf("received an error retrieving integration pagerduty %s", err)
		}

		return fmt.Errorf("integration pagerduty is not properly destroyed")
	}
}

const testAccCheckDatadogIntegrationPagerdutyConfig = `
 resource "datadog_integration_pagerduty" "foo" {
   services {
        service_name = "test_service"
        service_key  = "*****"
    }

   subdomain = "testdomain"
   api_token = "secret"
 }
 `

const testAccCheckDatadogIntegrationPagerdutyConfigTwoServices = `
 locals {
	 pd_services = {
		 test_service = "*****"
		 test_service_2 = "*****"
	 }
 }
 resource "datadog_integration_pagerduty" "foo" {
  dynamic "services" {
		for_each = local.pd_services
		content {
			service_name = services.key
			service_key = services.value
		}
	}

   subdomain = "testdomain"
   api_token = "secret"
}
`

const testAccCheckDatadogIntegrationPagerdutyConfigBeforeMigration = `
locals {
  pd_services = {
	  testing_foo = "*****"
	  testing_bar = "*****"
	}
}

# Create a new Datadog - PagerDuty integration
resource "datadog_integration_pagerduty" "pd" {
  dynamic "services" {
	  for_each = local.pd_services
	  content {
		  service_name = services.key
		  service_key = services.value
	  }
  }
  schedules = [
	  "https://ddog.pagerduty.com/schedules/X123VF",
	  "https://ddog.pagerduty.com/schedules/X321XX"
	]
  subdomain = "ddog"
  api_token = "secret"
}`

const testAccCheckDatadogIntegrationPagerdutyConfigDuringMigration = `
resource "datadog_integration_pagerduty" "pd" {
  schedules = [
	  "https://ddog.pagerduty.com/schedules/X123VF",
	  "https://ddog.pagerduty.com/schedules/X321XX"
	]
  subdomain = "ddog"
  api_token = "secret"
}`

const testAccCheckDatadogIntegrationPagerdutyConfigAfterMigration = `
resource "datadog_integration_pagerduty" "pd" {
  individual_services = true
  schedules = [
	  "https://ddog.pagerduty.com/schedules/X123VF",
	  "https://ddog.pagerduty.com/schedules/X321XX"
	]
  subdomain = "ddog"
  api_token = "secret"
}

resource "datadog_integration_pagerduty_service_object" "testing_foo" {
  depends_on = ["datadog_integration_pagerduty.pd"]
  service_name = "testing_foo"
  service_key  = "9876543210123456789"
}

resource "datadog_integration_pagerduty_service_object" "testing_bar" {
  depends_on = ["datadog_integration_pagerduty.pd"]
  service_name = "testing_bar"
  service_key  = "54321098765432109876"
}`
