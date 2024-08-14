package test

import (
	"context"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDatadogIpRangesDatasource_existing(t *testing.T) {
	t.Parallel()
	_, _, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	nonZeroRe := regexp.MustCompile(`[1-9]+`)
	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: accProviders,
		Steps: []resource.TestStep{
			{
				Config: `
				data "datadog_ip_ranges" "test" {}
				`,
				Check: resource.ComposeAggregateTestCheckFunc(
					//ipv4
					resource.TestMatchResourceAttr("data.datadog_ip_ranges.test", "agents_ipv4.#", nonZeroRe),
					resource.TestMatchResourceAttr("data.datadog_ip_ranges.test", "api_ipv4.#", nonZeroRe),
					resource.TestMatchResourceAttr("data.datadog_ip_ranges.test", "apm_ipv4.#", nonZeroRe),
					resource.TestMatchResourceAttr("data.datadog_ip_ranges.test", "global_ipv4.#", nonZeroRe),
					resource.TestMatchResourceAttr("data.datadog_ip_ranges.test", "logs_ipv4.#", nonZeroRe),
					resource.TestMatchResourceAttr("data.datadog_ip_ranges.test", "orchestrator_ipv4.#", nonZeroRe),
					resource.TestMatchResourceAttr("data.datadog_ip_ranges.test", "process_ipv4.#", nonZeroRe),
					resource.TestMatchResourceAttr("data.datadog_ip_ranges.test", "synthetics_ipv4.#", nonZeroRe),
					resource.TestMatchResourceAttr("data.datadog_ip_ranges.test", "synthetics_ipv4_by_location.%", nonZeroRe),
					resource.TestMatchResourceAttr("data.datadog_ip_ranges.test", "webhooks_ipv4.#", nonZeroRe),
					//ipv6
					resource.TestMatchResourceAttr("data.datadog_ip_ranges.test", "agents_ipv6.#", nonZeroRe),
					resource.TestMatchResourceAttr("data.datadog_ip_ranges.test", "api_ipv6.#", nonZeroRe),
					resource.TestMatchResourceAttr("data.datadog_ip_ranges.test", "apm_ipv6.#", nonZeroRe),
					resource.TestMatchResourceAttr("data.datadog_ip_ranges.test", "global_ipv6.#", nonZeroRe),
					resource.TestMatchResourceAttr("data.datadog_ip_ranges.test", "logs_ipv6.#", nonZeroRe),
					resource.TestMatchResourceAttr("data.datadog_ip_ranges.test", "orchestrator_ipv6.#", nonZeroRe),
					resource.TestMatchResourceAttr("data.datadog_ip_ranges.test", "process_ipv6.#", nonZeroRe),
					// These fields have no ip ranges set, this may change in the future.
					// resource.TestMatchResourceAttr("data.datadog_ip_ranges.test", "synthetics_ipv6.#", nonZeroRe),
					// resource.TestMatchResourceAttr("data.datadog_ip_ranges.test", "synthetics_ipv6_by_location.%", nonZeroRe),
					// resource.TestMatchResourceAttr("data.datadog_ip_ranges.test", "webhooks_ipv6.#", nonZeroRe),
				),
			},
		},
	})
}
