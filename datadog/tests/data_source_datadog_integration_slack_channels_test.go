package test

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDatadogIntegrationSlackChannelsDatasource(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	reg := regexp.MustCompile("[^a-zA-Z0-9]+")
	uniq := reg.ReplaceAllString(uniqueEntityName(ctx, t), "")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogIntegrationSlackChannelDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				// Workaround to ensure we create the slack integration before running the tests.
				Config: emptyLogsArchiveConfig(),
				Check: resource.ComposeTestCheckFunc(
					createSlackIntegration(providers.frameworkProvider),
				),
			},
			{
				Config: testAccDatasourceIntegrationSlackChannelsConfig(uniq),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.datadog_integration_slack_channels.foo", "id", "test_account"),
					resource.TestCheckResourceAttr("data.datadog_integration_slack_channels.foo", "account_name", "test_account"),
					resource.TestCheckTypeSetElemNestedAttrs("data.datadog_integration_slack_channels.foo", "channels.*", map[string]string{
						"name":                 "#" + uniq,
						"display.message":      "true",
						"display.notified":     "false",
						"display.snapshot":     "true",
						"display.tags":         "false",
						"display.mute_buttons": "true",
					}),
				),
			},
		},
	})
}

func testAccDatasourceIntegrationSlackChannelsConfig(uniq string) string {
	return fmt.Sprintf(`
		resource "datadog_integration_slack_channel" "foo" {
			display {
				message      = true
				notified     = false
				snapshot     = true
				tags         = false
				mute_buttons = true
			}
			channel_name = "#%s"
			account_name = "test_account"
		}

		data "datadog_integration_slack_channels" "foo" {
			account_name = datadog_integration_slack_channel.foo.account_name
		}
	`, uniq)
}
