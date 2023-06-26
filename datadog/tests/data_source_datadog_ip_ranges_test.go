package test

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDatadogIpRangesDatasource_existing(t *testing.T) {
	t.Parallel()
	_, _, accProviders := testAccFrameworkMuxProviders(context.Background(), t)

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: accProviders,
		Steps: []resource.TestStep{
			{
				Config: `
				data "datadog_ip_ranges" "test" {}
				`,
				Check: resource.ComposeAggregateTestCheckFunc(
					//ipv4
					resource.TestCheckResourceAttrSet("data.datadog_ip_ranges.test", "agents_ipv4.#"),
					resource.TestCheckResourceAttrSet("data.datadog_ip_ranges.test", "api_ipv4.#"),
					resource.TestCheckResourceAttrSet("data.datadog_ip_ranges.test", "apm_ipv4.#"),
					resource.TestCheckResourceAttrSet("data.datadog_ip_ranges.test", "logs_ipv4.#"),
					resource.TestCheckResourceAttrSet("data.datadog_ip_ranges.test", "orchestrator_ipv4.#"),
					resource.TestCheckResourceAttrSet("data.datadog_ip_ranges.test", "process_ipv4.#"),
					resource.TestCheckResourceAttrSet("data.datadog_ip_ranges.test", "synthetics_ipv4.#"),
					resource.TestCheckResourceAttrSet("data.datadog_ip_ranges.test", "synthetics_ipv4_by_location.%"),
					resource.TestCheckResourceAttrSet("data.datadog_ip_ranges.test", "webhooks_ipv4.#"),
					//ipv6
					resource.TestCheckResourceAttrSet("data.datadog_ip_ranges.test", "agents_ipv6.#"),
					resource.TestCheckResourceAttrSet("data.datadog_ip_ranges.test", "api_ipv6.#"),
					resource.TestCheckResourceAttrSet("data.datadog_ip_ranges.test", "apm_ipv6.#"),
					resource.TestCheckResourceAttrSet("data.datadog_ip_ranges.test", "logs_ipv6.#"),
					resource.TestCheckResourceAttrSet("data.datadog_ip_ranges.test", "orchestrator_ipv6.#"),
					resource.TestCheckResourceAttrSet("data.datadog_ip_ranges.test", "process_ipv6.#"),
					resource.TestCheckResourceAttrSet("data.datadog_ip_ranges.test", "synthetics_ipv6.#"),
					resource.TestCheckResourceAttrSet("data.datadog_ip_ranges.test", "synthetics_ipv6_by_location.%"),
					resource.TestCheckResourceAttrSet("data.datadog_ip_ranges.test", "webhooks_ipv6.#"),
				),
			},
		},
	})
}
