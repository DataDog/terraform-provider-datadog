package test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func TestAccDatadogIpRangesDatasource_existing(t *testing.T) {
	ctx, accProviders, _, cleanup := testAccProviders(t, initRecorder(t))
	defer cleanup(t)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(ctx, t) },
		Providers: accProviders,
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
					resource.TestCheckResourceAttrSet("data.datadog_ip_ranges.test", "process_ipv4.#"),
					resource.TestCheckResourceAttrSet("data.datadog_ip_ranges.test", "synthetics_ipv4.#"),
					resource.TestCheckResourceAttrSet("data.datadog_ip_ranges.test", "webhooks_ipv4.#"),
					//ipv6
					resource.TestCheckResourceAttrSet("data.datadog_ip_ranges.test", "agents_ipv6.#"),
					resource.TestCheckResourceAttrSet("data.datadog_ip_ranges.test", "api_ipv6.#"),
					resource.TestCheckResourceAttrSet("data.datadog_ip_ranges.test", "apm_ipv6.#"),
					resource.TestCheckResourceAttrSet("data.datadog_ip_ranges.test", "logs_ipv6.#"),
					resource.TestCheckResourceAttrSet("data.datadog_ip_ranges.test", "process_ipv6.#"),
					resource.TestCheckResourceAttrSet("data.datadog_ip_ranges.test", "synthetics_ipv6.#"),
					resource.TestCheckResourceAttrSet("data.datadog_ip_ranges.test", "webhooks_ipv6.#"),
				),
			},
		},
	})
}
