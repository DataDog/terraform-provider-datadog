package test

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/fwprovider"
)

// Get the current order and inverse it
func RetentionFiltersOrderConfig() string {
	return `
	data "datadog_apm_retention_filters_order" "order" {}
	resource "datadog_apm_retention_filter_order" "filters_orders" {
	filter_ids=reverse(data.datadog_apm_retention_filters_order.order.filter_ids)
	}`
}

func GetApmFilterIds(apmRetentionFiltersOrder datadogV2.RetentionFiltersResponse) []string {
	filterIds := make([]string, len(apmRetentionFiltersOrder.Data))
	for _, rf := range apmRetentionFiltersOrder.Data {
		filterIds = append(filterIds, rf.Id)
	}
	return filterIds
}

func TestAccDatadogApmRetentionFilterOrder(t *testing.T) {
	t.Parallel()
	_, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: accProviders,
		Steps: []resource.TestStep{
			{
				Config: RetentionFiltersOrderConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRetentionFilterOrderExists(providers.frameworkProvider),
					resource.TestCheckResourceAttrSet(
						"datadog_apm_retention_filter_order.filters_orders", "filter_ids.#"),
					testAccChecRetentionFilterOrderResourceMatch(providers.frameworkProvider, "datadog_apm_retention_filter_order.filters_orders", "filter_ids.0"),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckRetentionFilterOrderExists(accProvider *fwprovider.FrameworkProvider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		apiInstances := accProvider.DatadogApiInstances
		auth := accProvider.Auth

		if err := archiveOrderExistsChecker(auth, s, apiInstances); err != nil {
			return err
		}
		return nil
	}
}

func testAccChecRetentionFilterOrderResourceMatch(accProvider *fwprovider.FrameworkProvider, name string, key string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		apiInstances := accProvider.DatadogApiInstances
		auth := accProvider.Auth

		resourceType := strings.Split(name, ".")[0]
		elemNo, _ := strconv.Atoi(strings.Split(key, ".")[1])
		for _, r := range s.RootModule().Resources {
			if r.Type == resourceType {
				currentOrder, _, err := apiInstances.GetApmRetentionFiltersApiV2().ListApmRetentionFilters(auth)
				if err != nil {
					return fmt.Errorf("received an error when retrieving APM retention filters order, (%s)", err)
				}
				filterIds := currentOrder.Data
				if elemNo >= len(filterIds) {
					println("can't match (%s), there are only %s filters", key, strconv.Itoa(len(filterIds)))
					return nil
				}
				return resource.TestCheckResourceAttr(name, key, filterIds[elemNo].Id)(s)
			}
		}
		return fmt.Errorf("didn't find %s", name)
	}
}
