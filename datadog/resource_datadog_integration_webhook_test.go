package datadog

import (
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	datadog "github.com/zorkian/go-datadog-api"
)

const testAccCheckDatadogIntegrationWebhookConfig = `
resource "datadog_integration_webhook" "my_webhook" {
	name           = "my_webhook"
	url            = "http://example.com"
	payload        = <<EOF
{    
    "body": "$EVENT_MSG",
    "last_updated": "$LAST_UPDATED",
    "event_type": "$EVENT_TYPE",
	"title": "$EVENT_TITLE",
	"id": "$ID"
}
EOF
	custom_headers = "{\"x-api-key\":\"foo\"}"
	encode_as_form = false
}
`

func TestAccDatadogIntegrationWebhook(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckDatadogIntegrationWebhookDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogIntegrationWebhookConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogIntegrationWebhookExists("datadog_integration_webhook.my_webhook"),
					resource.TestCheckResourceAttr(
						"datadog_integration_webhook.my_webhook", "name", "my_webhook"),
					resource.TestCheckResourceAttr(
						"datadog_integration_webhook.my_webhook", "url", "http://example.com"),
					resource.TestCheckResourceAttr(
						"datadog_integration_webhook.my_webhook", "encode_as_form", "false"),
					resource.TestCheckResourceAttr(
						"datadog_integration_webhook.my_webhook", "custom_headers", "{\"x-api-key\":\"foo\"}"),
					resource.TestCheckResourceAttr(
						"datadog_integration_webhook.my_webhook", "payload", "{    \n    \"body\": \"$EVENT_MSG\",\n    \"last_updated\": \"$LAST_UPDATED\",\n    \"event_type\": \"$EVENT_TYPE\",\n\t\"title\": \"$EVENT_TITLE\",\n\t\"id\": \"$ID\"\n}\n"),
				),
			},
		},
	})
}

func testAccCheckDatadogIntegrationWebhookExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client := testAccProvider.Meta().(*datadog.Client)
		if err := datadogIntegrationWebhookExistsHelper(s, client); err != nil {
			return err
		}
		return nil
	}
}

func datadogIntegrationWebhookExistsHelper(s *terraform.State, client *datadog.Client) error {
	if _, err := client.GetIntegrationWebhook(); err != nil {
		return fmt.Errorf("Received an error retrieving webhook integration %s", err)
	}
	return nil
}

func testAccCheckDatadogIntegrationWebhookDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*datadog.Client)

	_, err := client.GetIntegrationWebhook()
	if err != nil {
		if strings.Contains(err.Error(), "webhooks not found") {
			return nil
		}
		return fmt.Errorf("Received an error on retrieving integration webhook %s", err)
	}
	return fmt.Errorf("Integration webhook is not properly destroyed")
}
