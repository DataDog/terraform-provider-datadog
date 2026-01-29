package test

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDatadogAgentlessScanningAwsScanOptionsDataSource_Basic(t *testing.T) {
	t.Parallel()
	_, _, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	accountID := "123456789012"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: accProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDatadogAgentlessScanningAwsScanOptionsDataSourceConfig(accountID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.datadog_agentless_scanning_aws_scan_options.test", "id"),
				),
			},
		},
	})
}

func testAccDatadogAgentlessScanningAwsScanOptionsDataSourceConfig(accountID string) string {
	return `
data "datadog_agentless_scanning_aws_scan_options" "test" {}
`
}
