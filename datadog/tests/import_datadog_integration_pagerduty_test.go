package test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestDatadogIntegrationPagerduty_import(t *testing.T) {
	resourceName := "datadog_integration_pagerduty.pd"
	ctx, accProviders := testAccProviders(context.Background(), t)
	serviceName := strings.ReplaceAll(uniqueEntityName(ctx, t), "-", "_")
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		CheckDestroy:      testAccCheckDatadogIntegrationPagerdutyDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogIntegrationPagerdutyConfigImported(serviceName),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckDatadogIntegrationPagerdutyConfigImported(uniq string) string {
	return fmt.Sprintf(`
locals {
	pd_services = {
		%s = "*****"
		%s_2 = "*****"
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
  schedules = ["https://ddog.pagerduty.com/schedules/X123VF"]
  subdomain = "testdomain"
}`, uniq, uniq)
}
