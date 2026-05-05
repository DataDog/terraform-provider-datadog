package test

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/fwprovider"
)

const RumRetentionFiltersOrderResourceTestAppId = "fdb76419-eff5-4344-867a-4b72bad613cb"

func validDatadogRumRetentionFiltersOrder() string {
	return fmt.Sprintf(`data "datadog_rum_retention_filters" "my_retention_filters" {
		application_id = %q
	}

	resource "datadog_rum_retention_filters_order" "testing_rum_retention_filters_order" {
		application_id = %q
	    retention_filter_ids = concat([
		    data.datadog_rum_retention_filters.my_retention_filters.retention_filters[0].id
	    ],[
		    for rf in slice(data.datadog_rum_retention_filters.my_retention_filters.retention_filters, 1, length(data.datadog_rum_retention_filters.my_retention_filters.retention_filters)) : rf.id
	    ])
	}
	`, RumRetentionFiltersOrderResourceTestAppId, RumRetentionFiltersOrderResourceTestAppId)
}

func unmatchedErrorDatadogRumRetentionFiltersOrder() string {
	return fmt.Sprintf(`resource "datadog_rum_retention_filters_order" "testing_rum_retention_filters_order_" {
		application_id = %q
	    retention_filter_ids = ["not_found_uuid"]
	}
	`, RumRetentionFiltersOrderResourceTestAppId)
}

func TestAccRumRetentionFilterOrder(t *testing.T) {
	t.Parallel()
	_, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: accProviders,
		Steps: []resource.TestStep{
			{
				Config: validDatadogRumRetentionFiltersOrder(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr("datadog_rum_retention_filters_order.testing_rum_retention_filters_order", "retention_filter_ids.#", regexp.MustCompile(`3`)),
					testAccCheckDatadogRumRetentionFiltersOrderResourceMatch(providers.frameworkProvider,
						"datadog_rum_retention_filters_order.testing_rum_retention_filters_order", "retention_filter_ids.0"),
					testAccCheckDatadogRumRetentionFiltersOrderResourceMatch(providers.frameworkProvider,
						"datadog_rum_retention_filters_order.testing_rum_retention_filters_order", "retention_filter_ids.1"),
					testAccCheckDatadogRumRetentionFiltersOrderResourceMatch(providers.frameworkProvider,
						"datadog_rum_retention_filters_order.testing_rum_retention_filters_order", "retention_filter_ids.2"),
				),
			},
		},
	})
}

func TestAccRumRetentionFilterOrderError(t *testing.T) {
	t.Parallel()
	_, _, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: accProviders,
		Steps: []resource.TestStep{
			{
				Config:      unmatchedErrorDatadogRumRetentionFiltersOrder(),
				ExpectError: regexp.MustCompile("error re-ordering retention filters"),
			},
		},
	})
}

func testAccCheckDatadogRumRetentionFiltersOrderResourceMatch(accProvider *fwprovider.FrameworkProvider, name string, key string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		apiInstances := accProvider.DatadogApiInstances
		auth := accProvider.Auth

		resp, _, err := apiInstances.GetRumRetentionFiltersApiV2().ListRetentionFilters(auth, RumRetentionFiltersOrderResourceTestAppId)
		if err != nil {
			return fmt.Errorf("received an error when retrieving RUM retention filters, (%s)", err)
		}
		retentionFilters := resp.Data

		elemNo, _ := strconv.Atoi(strings.Split(key, ".")[1])

		if elemNo >= len(retentionFilters) {
			println("can't match (%s), there are only %s filters", key, strconv.Itoa(len(retentionFilters)))
			return nil
		}
		return resource.TestCheckResourceAttr(name, key, *retentionFilters[elemNo].Id)(s)
	}
}
