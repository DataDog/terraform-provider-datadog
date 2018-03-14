package datadog

import (
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	datadog "gopkg.in/zorkian/go-datadog-api.v2"
)

// We're not testing for schedules because Datadog actively verifies it with Pagerduty

func TestAccDatadogIntegrationPagerduty_Basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckDatadogIntegrationPagerdutyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogIntegrationPagerdutyConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogIntegrationPagerdutyExists("datadog_integration_pagerduty.foo"),
					resource.TestCheckResourceAttr(
						"datadog_integration_pagerduty.foo", "subdomain", "testdomain"),
					resource.TestCheckResourceAttr(
						"datadog_integration_pagerduty.foo", "api_token", "*****"),
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
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckDatadogIntegrationPagerdutyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogIntegrationPagerdutyConfig_TwoServices,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogIntegrationPagerdutyExists("datadog_integration_pagerduty.foo"),
					resource.TestCheckResourceAttr(
						"datadog_integration_pagerduty.foo", "subdomain", "testdomain"),
					resource.TestCheckResourceAttr(
						"datadog_integration_pagerduty.foo", "api_token", "*****"),
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

func testAccCheckDatadogIntegrationPagerdutyExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client := testAccProvider.Meta().(*datadog.Client)
		if err := datadogIntegrationPagerdutyExistsHelper(s, client); err != nil {
			return err
		}
		return nil
	}
}

func datadogIntegrationPagerdutyExistsHelper(s *terraform.State, client *datadog.Client) error {
	if _, err := client.GetIntegrationPD(); err != nil {
		return fmt.Errorf("Received an error retrieving integration pagerduty %s", err)
	}
	return nil
}

func testAccCheckDatadogIntegrationPagerdutyDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*datadog.Client)

	_, err := client.GetIntegrationPD()
	if err != nil {
		if strings.Contains(err.Error(), "pagerduty not found") {
			return nil
		}

		return fmt.Errorf("Received an error retrieving integration pagerduty %s", err)
	}

	return fmt.Errorf("Integration pagerduty is not properly destroyed")
}

const testAccCheckDatadogIntegrationPagerdutyConfig = `
 resource "datadog_integration_pagerduty" "foo" {
   services
     {
         service_name = "test_service",
         service_key  = "*****",
     }

   subdomain = "testdomain"
   api_token = "*****"
 }
 `

const testAccCheckDatadogIntegrationPagerdutyConfig_TwoServices = `
 resource "datadog_integration_pagerduty" "foo" {
   services
     {
         service_name = "test_service",
         service_key  = "*****",
     }

   services
     {
		service_name = "test_service_2",
		service_key  = "*****",
	 }

   subdomain = "testdomain"
   api_token = "*****"
 }
`
