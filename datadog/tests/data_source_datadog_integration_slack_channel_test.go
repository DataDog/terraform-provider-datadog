package test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestAccDatadogIntegrationSlackChannelDatasource(t *testing.T) {
	t.Parallel()
	ctx, accProviders := testAccProviders(context.Background(), t)
	uniq := uniqueEntityName(ctx, t)
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		CheckDestroy:      testAccCheckDatadogIntegrationSlackChannelDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccDatasourceSlackIntegrationChannelConfig(uniq),
				Check:  checkDatasourceSlackIntegrationChannelAttrs(accProvider, uniq),
			},
		},
	})
}

func checkDatasourceSlackIntegrationChannelAttrs(accProvider func() (*schema.Provider, error), uniq string) resource.TestCheckFunc {
	return resource.ComposeTestCheckFunc(
		testAccCheckDatadogIntegrationSlackChannelExists(accProvider, "datadog_integration_slack_channel.foo"),
		resource.TestCheckResourceAttr("data.datadog_integration_slack_channel.foo", "channel_name", uniq),
	)
}

func testAccSlackIntegrationChannelConfig(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_integration_slack_channel" "foo" {
	account_name = "test_account"
	channel_name = "#%s"
	display {
		message = false
		notified = false
		snapshot = false
		tags = false
	}
}
`, uniq)
}

func testAccDatasourceSlackIntegrationChannelConfig(uniq string) string {
	return fmt.Sprintf(`
%s
data "datadog_integration_slack_channel" "foo" {
	depends_on = [
		datadog_integration_slack_channel.foo
	]
	account_name = "test_account"
	channel_name = "%s"
}`, testAccSlackIntegrationChannelConfig(uniq), uniq)
}
