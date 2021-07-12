package test

import (
	"context"
	"fmt"
	"strconv"
	"testing"

	"github.com/terraform-providers/terraform-provider-datadog/datadog"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

const tfSecurityFiltersSource = "data.datadog_security_monitoring_filters.acceptance_test"

func TestAccDatadogSecurityMonitoringFilterDatasource(t *testing.T) {
	_, accProviders := testAccProviders(context.Background(), t)
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceSecurityMonitoringFilter(),
				Check: resource.ComposeTestCheckFunc(
					securityMonitoringCheckFilterCount(accProvider),
				),
			},
		},
	})
}

func securityMonitoringCheckFilterCount(accProvider func() (*schema.Provider, error)) func(state *terraform.State) error {
	return func(state *terraform.State) error {
		provider, _ := accProvider()
		providerConf := provider.Meta().(*datadog.ProviderConfiguration)
		authV2 := providerConf.AuthV2
		client := providerConf.DatadogClientV2

		filtersResponse, _, err := client.SecurityMonitoringApi.ListSecurityFilters(authV2)
		if err != nil {
			return err
		}
		return securityMonitoringFilterCount(state, len(*filtersResponse.Data))
	}
}



func securityMonitoringFilterCount(state *terraform.State, responseCount int) error {
	// TODO rename here
	resourceAttributes := state.RootModule().Resources[tfSecurityFiltersSource].Primary.Attributes
	filtersIdsCount, _ := strconv.Atoi(resourceAttributes["filters_ids.#"])
	filtersCount, _ := strconv.Atoi(resourceAttributes["filters.#"])

	if filtersCount != responseCount || filtersIdsCount != responseCount {
		return fmt.Errorf("expected %d filters got %d filters and %d filters ids",
			responseCount, filtersCount, filtersIdsCount)
	}
	return nil
}

func testAccDataSourceSecurityMonitoringFilter() string {
	return `
data "datadog_security_monitoring_filters" "acceptance_test" {
}
`
}
