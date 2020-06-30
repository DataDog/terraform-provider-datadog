package datadog

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func TestDatadogIntegrationPagerduty_import(t *testing.T) {
	resourceName := "datadog_integration_pagerduty.pd"
	accProviders, cleanup := testAccProviders(t, initRecorder(t))
	defer cleanup(t)
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    accProviders,
		CheckDestroy: testAccCheckDatadogIntegrationPagerdutyDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogIntegrationPagerdutyConfigImported,
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"preserve_existing_integration",
				},
			},
		},
	})
}

const testAccCheckDatadogIntegrationPagerdutyConfigImported = `
locals {
	pd_services = {
		test_service = "*****"
		test_service_2 = "*****"
	}
}

resource "datadog_integration_pagerduty" "pd" {
  dynamic "services" {
		for_each = local.pd_services
		content {
			service_name = services.key
			service_key = services.value
		}
	}
  subdomain = "testdomain"
}
`
