package test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/terraform-providers/terraform-provider-datadog/datadog"
	communityClient "github.com/zorkian/go-datadog-api"
)

// We're not testing for schedules because Datadog actively verifies it with Pagerduty

func TestAccDatadogIntegrationPagerduty_Basic(t *testing.T) {
	ctx, accProviders, clock, cleanup := testAccProviders(t, initRecorder(t))
	serviceName := strings.ReplaceAll(uniqueEntityName(clock, t), "-", "_")
	defer cleanup(t)
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(ctx, t) },
		Providers:    accProviders,
		CheckDestroy: testAccCheckDatadogIntegrationPagerdutyDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogIntegrationPagerdutyConfig(serviceName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogIntegrationPagerdutyExists(accProvider),
					resource.TestCheckResourceAttr(
						"datadog_integration_pagerduty.foo", "subdomain", "testdomain"),
					resource.TestCheckResourceAttr(
						"datadog_integration_pagerduty.foo", "api_token", "secret"),
					resource.TestCheckResourceAttr(
						"datadog_integration_pagerduty.foo", "services.0.service_name", serviceName),
					resource.TestCheckResourceAttr(
						"datadog_integration_pagerduty.foo", "services.0.service_key", "*****"),
					resource.TestCheckResourceAttr(
						"datadog_integration_pagerduty.foo", "schedules.0", "https://ddog.pagerduty.com/schedules/X123VF"),
				),
			},
		},
	})
}

func TestAccDatadogIntegrationPagerduty_TwoServices(t *testing.T) {
	ctx, accProviders, clock, cleanup := testAccProviders(t, initRecorder(t))
	serviceName := strings.ReplaceAll(uniqueEntityName(clock, t), "-", "_")
	defer cleanup(t)
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(ctx, t) },
		Providers:    accProviders,
		CheckDestroy: testAccCheckDatadogIntegrationPagerdutyDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogIntegrationPagerdutyConfigTwoServices(serviceName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogIntegrationPagerdutyExists(accProvider),
					resource.TestCheckResourceAttr(
						"datadog_integration_pagerduty.foo", "subdomain", "testdomain"),
					resource.TestCheckResourceAttr(
						"datadog_integration_pagerduty.foo", "api_token", "secret"),
					resource.TestCheckResourceAttr(
						"datadog_integration_pagerduty.foo", "services.0.service_name", serviceName),
					resource.TestCheckResourceAttr(
						"datadog_integration_pagerduty.foo", "services.0.service_key", "*****"),
					resource.TestCheckResourceAttr(
						"datadog_integration_pagerduty.foo", "services.1.service_name", serviceName+"_2"),
					resource.TestCheckResourceAttr(
						"datadog_integration_pagerduty.foo", "services.1.service_key", "*****"),
					resource.TestCheckResourceAttr(
						"datadog_integration_pagerduty.foo", "schedules.0", "https://ddog.pagerduty.com/schedules/X123VF"),
				),
			},
		},
	})
}

func TestAccDatadogIntegrationPagerduty_Migrate2ServiceObjects(t *testing.T) {
	ctx, accProviders, clock, cleanup := testAccProviders(t, initRecorder(t))
	serviceName := strings.ReplaceAll(uniqueEntityName(clock, t), "-", "_")
	defer cleanup(t)
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(ctx, t) },
		Providers:    accProviders,
		CheckDestroy: testAccCheckDatadogIntegrationPagerdutyDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogIntegrationPagerdutyConfigBeforeMigration(serviceName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogIntegrationPagerdutyExists(accProvider),
					resource.TestCheckResourceAttr(
						"datadog_integration_pagerduty.pd", "subdomain", "ddog"),
					resource.TestCheckResourceAttr(
						"datadog_integration_pagerduty.pd", "api_token", "secret"),
					resource.TestCheckResourceAttr(
						"datadog_integration_pagerduty.pd", "services.0.service_name", serviceName+"_bar"),
					resource.TestCheckResourceAttr(
						"datadog_integration_pagerduty.pd", "services.0.service_key", "*****"),
					resource.TestCheckResourceAttr(
						"datadog_integration_pagerduty.pd", "services.1.service_name", serviceName+"_foo"),
					resource.TestCheckResourceAttr(
						"datadog_integration_pagerduty.pd", "services.1.service_key", "*****"),
				),
			},
			{
				// this represents the intermediary step which will ensure the old
				// inline-defined service objects get removed
				Config: testAccCheckDatadogIntegrationPagerdutyConfigDuringMigration(serviceName),
			},
			{
				Config: testAccCheckDatadogIntegrationPagerdutyConfigAfterMigration(serviceName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogIntegrationPagerdutyExists(accProvider),
					resource.TestCheckResourceAttr(
						"datadog_integration_pagerduty.pd", "subdomain", "ddog"),
					resource.TestCheckResourceAttr(
						"datadog_integration_pagerduty.pd", "api_token", "secret"),
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
		},
	})
}

func testAccCheckDatadogIntegrationPagerdutyExists(accProvider *schema.Provider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		providerConf := accProvider.Meta().(*datadog.ProviderConfiguration)
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

func testAccCheckDatadogIntegrationPagerdutyDestroy(accProvider *schema.Provider) func(*terraform.State) error {
	return func(s *terraform.State) error {
		// We don't destroy the integration anymore, so let's not check anything
		return nil
	}
}

// NOTE: Don't create configs with no schedules. If there's a leftover schedule from some previous test run,
// the test will fail on non-empty diff after apply, as the resource is (unfortunately) created in a way
// to not overwrite existing schedules with empty list on creation.
func testAccCheckDatadogIntegrationPagerdutyConfig(uniq string) string {
	return fmt.Sprintf(`
 resource "datadog_integration_pagerduty" "foo" {
   services {
        service_name = "%s"
        service_key  = "*****"
    }

   schedules = ["https://ddog.pagerduty.com/schedules/X123VF"]
   subdomain = "testdomain"
   api_token = "secret"
 }`, uniq)
}

func testAccCheckDatadogIntegrationPagerdutyConfigTwoServices(uniq string) string {
	return fmt.Sprintf(`
 locals {
	 pd_services = {
		 %s = "*****"
		 %s_2 = "*****"
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

   schedules = ["https://ddog.pagerduty.com/schedules/X123VF"]
   subdomain = "testdomain"
   api_token = "secret"
}`, uniq, uniq)
}

func testAccCheckDatadogIntegrationPagerdutyConfigBeforeMigration(uniq string) string {
	return fmt.Sprintf(`
locals {
  pd_services = {
	  %s_foo = "*****"
	  %s_bar = "*****"
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
}`, uniq, uniq)
}

func testAccCheckDatadogIntegrationPagerdutyConfigDuringMigration(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_integration_pagerduty" "pd" {
  schedules = [
	  "https://ddog.pagerduty.com/schedules/X123VF",
	  "https://ddog.pagerduty.com/schedules/X321XX"
	]
  subdomain = "ddog"
  api_token = "secret"
}`)
}

func testAccCheckDatadogIntegrationPagerdutyConfigAfterMigration(uniq string) string {
	return fmt.Sprintf(`
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
  service_name = "%s_foo"
  service_key  = "9876543210123456789"
}

resource "datadog_integration_pagerduty_service_object" "testing_bar" {
  depends_on = ["datadog_integration_pagerduty.pd"]
  service_name = "%s_bar"
  service_key  = "54321098765432109876"
}`, uniq, uniq)
}
