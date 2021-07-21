package test

import (
	"context"
	"fmt"
	"testing"

	"github.com/terraform-providers/terraform-provider-datadog/datadog"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

const tfSecurityFilterName = "datadog_security_monitoring_filter.acceptance_test"

func TestAccDatadogSecurityMonitoringFilter(t *testing.T) {
	t.Parallel()
	ctx, accProviders := testAccProviders(context.Background(), t)
	filterName := uniqueEntityName(ctx, t)
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		CheckDestroy:      testAccCheckDatadogSecurityMonitoringFilterDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogSecurityMonitoringFilterCreated(filterName),
				Check:  testAccCheckDatadogSecurityMonitorFilterCreatedCheck(accProvider, filterName),
			},
			{
				Config: testAccCheckDatadogSecurityMonitoringFilterUpdated(filterName),
				Check:  testAccCheckDatadogSecurityMonitorFilterUpdatedCheck(accProvider, filterName),
			},
		},
	})
}

func testAccCheckDatadogSecurityMonitoringFilterCreated(name string) string {
	return fmt.Sprintf(`
resource "datadog_security_monitoring_filter" "acceptance_test" {
    name = "%s"
    query = "first query"
    is_enabled = true

    exclusion_filter {
        name = "first"
        query = "does not really match much"
    }

    exclusion_filter {
        name = "second"
        query = "neither does it"
    }
}
`, name)
}

func testAccCheckDatadogSecurityMonitorFilterCreatedCheck(accProvider func() (*schema.Provider, error), filterName string) resource.TestCheckFunc {
	return resource.ComposeTestCheckFunc(
		testAccCheckDatadogSecurityMonitoringFilterExists(accProvider),
		resource.TestCheckResourceAttr(
			tfSecurityFilterName, "name", filterName),
		resource.TestCheckResourceAttr(
			tfSecurityFilterName, "query", "first query"),
		resource.TestCheckResourceAttr(
			tfSecurityFilterName, "is_enabled", "true"),
		resource.TestCheckResourceAttr(
			tfSecurityFilterName, "exclusion_filter.#", "2"),
		resource.TestCheckResourceAttr(
			tfSecurityFilterName, "exclusion_filter.0.name", "first"),
		resource.TestCheckResourceAttr(
			tfSecurityFilterName, "exclusion_filter.0.query", "does not really match much"),
		resource.TestCheckResourceAttr(
			tfSecurityFilterName, "exclusion_filter.1.name", "second"),
		resource.TestCheckResourceAttr(
			tfSecurityFilterName, "exclusion_filter.1.query", "neither does it"),
	)
}

func testAccCheckDatadogSecurityMonitoringFilterUpdated(name string) string {
	return fmt.Sprintf(`
resource "datadog_security_monitoring_filter" "acceptance_test" {
    name = "%s"
    query = "new query"
    is_enabled = false

    exclusion_filter {
        name = "first"
        query = "does not really match much"
    }

    exclusion_filter {
        name = "third"
        query = "I am new"
    }
}
`, name)
}

func testAccCheckDatadogSecurityMonitorFilterUpdatedCheck(accProvider func() (*schema.Provider, error), filterName string) resource.TestCheckFunc {
	return resource.ComposeTestCheckFunc(
		testAccCheckDatadogSecurityMonitoringFilterExists(accProvider),
		resource.TestCheckResourceAttr(
			tfSecurityFilterName, "name", filterName),
		resource.TestCheckResourceAttr(
			tfSecurityFilterName, "query", "new query"),
		resource.TestCheckResourceAttr(
			tfSecurityFilterName, "is_enabled", "false"),
		resource.TestCheckResourceAttr(
			tfSecurityFilterName, "exclusion_filter.#", "2"),
		resource.TestCheckResourceAttr(
			tfSecurityFilterName, "exclusion_filter.0.name", "first"),
		resource.TestCheckResourceAttr(
			tfSecurityFilterName, "exclusion_filter.0.query", "does not really match much"),
		resource.TestCheckResourceAttr(
			tfSecurityFilterName, "exclusion_filter.1.name", "third"),
		resource.TestCheckResourceAttr(
			tfSecurityFilterName, "exclusion_filter.1.query", "I am new"),
	)
}

func testAccCheckDatadogSecurityMonitoringFilterExists(accProvider func() (*schema.Provider, error)) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		provider, _ := accProvider()
		providerConf := provider.Meta().(*datadog.ProviderConfiguration)
		authV2 := providerConf.AuthV2
		client := providerConf.DatadogClientV2

		for _, filter := range s.RootModule().Resources {
			_, _, err := client.SecurityMonitoringApi.GetSecurityFilter(authV2, filter.Primary.ID)
			if err != nil {
				return fmt.Errorf("received an error retrieving security monitoring filter: %s", err)
			}
		}
		return nil
	}
}

func testAccCheckDatadogSecurityMonitoringFilterDestroy(accProvider func() (*schema.Provider, error)) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		provider, _ := accProvider()
		providerConf := provider.Meta().(*datadog.ProviderConfiguration)
		authV2 := providerConf.AuthV2
		client := providerConf.DatadogClientV2

		for _, resource := range s.RootModule().Resources {
			if resource.Type == "datadog_security_monitoring_filter" {
				_, httpResponse, err := client.SecurityMonitoringApi.GetSecurityFilter(authV2, resource.Primary.ID)
				if err != nil {
					if httpResponse != nil && httpResponse.StatusCode == 404 {
						continue
					}
					return fmt.Errorf("received an error deleting security monitoring filter: %s", err)
				}
				return fmt.Errorf("monitor still exists")
			}
		}
		return nil
	}

}
