package test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDatadogHostsDatasource(t *testing.T) {
	t.Parallel()
	_, _, accProviders := testAccFrameworkMuxProviders(context.Background(), t)

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: accProviders,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					data "datadog_hosts" "test" {
						from = %d
						include_muted_hosts_data = true
					}`, time.Now().Unix()-604800), // Get hosts for the last 7 days, instead of default 3 hours.
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.datadog_hosts.test", "total_matching"),
					resource.TestCheckResourceAttrSet("data.datadog_hosts.test", "total_returned"),
					resource.TestCheckResourceAttrSet("data.datadog_hosts.test", "host_list.0.host_name"),
					resource.TestCheckResourceAttrSet("data.datadog_hosts.test", "host_list.0.id"),
					resource.TestCheckResourceAttrSet("data.datadog_hosts.test", "host_list.0.name"),
					resource.TestCheckResourceAttrSet("data.datadog_hosts.test", "host_list.0.up"),
					resource.TestCheckResourceAttrSet("data.datadog_hosts.test", "host_list.0.metrics.cpu"),
					resource.TestCheckResourceAttrSet("data.datadog_hosts.test", "host_list.0.metrics.iowait"),
					resource.TestCheckResourceAttrSet("data.datadog_hosts.test", "host_list.0.metrics.load"),
				),
			},
		},
	})
}
