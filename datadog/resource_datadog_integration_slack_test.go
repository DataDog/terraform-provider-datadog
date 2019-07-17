package datadog

import (
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	datadog "github.com/zorkian/go-datadog-api"
)

func TestAccDatadogIntegrationSlack_Basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckDatadogIntegrationSlackDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogIntegrationSlackConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogIntegrationSlackExists("datadog_integration_slack.foo"),
					resource.TestCheckResourceAttr(
						"datadog_integration_slack.foo", "service_hook.0.account", "test_service"),
					resource.TestCheckResourceAttr(
						"datadog_integration_slack.foo", "service_hook.0.url", "*****"),
					resource.TestCheckResourceAttr(
						"datadog_integration_slack.foo", "channel.0.channel_name", "#test_channel"),
					resource.TestCheckResourceAttr(
						"datadog_integration_slack.foo", "channel.0.account", "test_service"),
					resource.TestCheckResourceAttr(
						"datadog_integration_slack.foo", "channel.0.transfer_all_user_comments", "true"),
				),
			},
		},
	})
}

func testAccCheckDatadogIntegrationSlackExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client := testAccProvider.Meta().(*datadog.Client)
		if err := datadogIntegrationSlackExistsHelper(s, client); err != nil {
			return err
		}
		return nil
	}
}

func datadogIntegrationSlackExistsHelper(s *terraform.State, client *datadog.Client) error {
	if _, err := client.GetIntegrationPD(); err != nil {
		return fmt.Errorf("Received an error retrieving integration slack %s", err)
	}
	return nil
}

func testAccCheckDatadogIntegrationSlackDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*datadog.Client)

	_, err := client.GetIntegrationPD()
	if err != nil {
		if strings.Contains(err.Error(), "slack not found") {
			return nil
		}

		return fmt.Errorf("Received an error retrieving integration slack %s", err)
	}

	return fmt.Errorf("Integration slack is not properly destroyed")
}

const testAccCheckDatadogIntegrationSlackConfig = `
 resource "datadog_integration_slack" "foo" {
   service_hook {
      account = "test_service"
      url = "*****"
   }

   channel {
      channel_name = "#test_channel"
      account = "test_service"
      transfer_all_user_comments = true
   }
 }
 `
