package test

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDatadogAgentlessScanningGcpScanOptionsDataSource_Basic(t *testing.T) {
	t.Parallel()
	_, _, accProviders := testAccFrameworkMuxProviders(context.Background(), t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: accProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDatadogAgentlessScanningGcpScanOptionsDataSourceConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.datadog_agentless_scanning_gcp_scan_options.test", "id"),
				),
			},
		},
	})
}

func testAccDatadogAgentlessScanningGcpScanOptionsDataSourceConfig() string {
	return `
data "datadog_agentless_scanning_gcp_scan_options" "test" {}
`
}
