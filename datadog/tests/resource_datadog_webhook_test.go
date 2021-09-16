package test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/terraform-providers/terraform-provider-datadog/datadog"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccDatadogWebhook_Basic(t *testing.T) {
	ctx, accProviders := testAccProviders(context.Background(), t)
	uniqueName := uniqueEntityName(ctx, t)
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		CheckDestroy:      testAccCheckDatadogWebhookDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogWebhookBasicConfig(uniqueName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogWebhookExists(accProvider, "datadog_webhook.foo"),
					resource.TestCheckResourceAttr("datadog_webhook.foo", "name", uniqueName),
					resource.TestCheckResourceAttr("datadog_webhook.foo", "url", "example.com"),
					resource.TestCheckResourceAttr("datadog_webhook.foo", "encode_as", "json"),
					resource.TestCheckResourceAttr("datadog_webhook.foo", "custom_headers", "{\"test\":\"test\"}"),
					resource.TestCheckResourceAttr("datadog_webhook.foo", "payload", "{\"body\": \"$EVENT_MSG\", \"last_updated\": \"$LAST_UPDATED\", \"event_type\": \"$EVENT_TYPE\", \"title\": \"$EVENT_TITLE\", \"date\": \"$DATE\", \"org\": {\"id\": \"$ORG_ID\", \"name\": \"$ORG_NAME\"}, \"id\": \"$ID\"}"),
				),
			},
			{
				Config: testAccCheckDatadogWebhookBasicConfigUpdated(uniqueName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogWebhookExists(accProvider, "datadog_webhook.foo"),
					resource.TestCheckResourceAttr("datadog_webhook.foo", "name", uniqueName+"UPDATED"),
					resource.TestCheckResourceAttr("datadog_webhook.foo", "url", "example.com/updated"),
					resource.TestCheckResourceAttr("datadog_webhook.foo", "encode_as", "json"),
					resource.TestCheckResourceAttr("datadog_webhook.foo", "custom_headers", "{\"test\":\"test-updated\"}"),
					resource.TestCheckResourceAttr("datadog_webhook.foo", "payload", "{\"custom\":\"payload\"}"),
				),
			},
		},
	})
}

func testAccCheckDatadogWebhookBasicConfig(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_webhook" "foo" {
  name           = "%s"
  url            = "example.com"
  encode_as      = "json"
  custom_headers = jsonencode({"test": "test"})
}`, uniq)
}

func testAccCheckDatadogWebhookBasicConfigUpdated(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_webhook" "foo" {
  name           = "%sUPDATED"
  url            = "example.com/updated"
  encode_as      = "json"
  custom_headers = jsonencode({"test": "test-updated"})
  payload        = jsonencode({"custom": "payload"})
}`, uniq)
}

func testAccCheckDatadogWebhookExists(accProvider func() (*schema.Provider, error), name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		provider, _ := accProvider()
		providerConf := provider.Meta().(*datadog.ProviderConfiguration)
		client := providerConf.DatadogClientV1
		auth := providerConf.AuthV1

		id := s.RootModule().Resources[name].Primary.ID
		_, httpresp, err := client.WebhooksIntegrationApi.GetWebhooksIntegration(auth, id)
		if err != nil {
			return utils.TranslateClientError(err, httpresp, "error checking webhooks custom variable existence")
		}
		return nil
	}
}

func testAccCheckDatadogWebhookDestroy(accProvider func() (*schema.Provider, error)) func(*terraform.State) error {
	return func(s *terraform.State) error {
		provider, _ := accProvider()
		meta := provider.Meta()
		providerConf := meta.(*datadog.ProviderConfiguration)
		datadogClient := providerConf.DatadogClientV1
		auth := providerConf.AuthV1
		for _, r := range s.RootModule().Resources {
			var err error

			id := r.Primary.ID

			_, _, err = datadogClient.WebhooksIntegrationApi.GetWebhooksIntegration(auth, id)
			if err != nil {
				// Api returns 400 when the webhook custom variable does not exist
				if strings.Contains(err.Error(), "400 Bad Request") {
					continue
				} else {
					return fmt.Errorf("received an error retrieving webhook: %s", err.Error())
				}
			} else {
				return fmt.Errorf("webhook %s still exists", r.Primary.ID)
			}
		}

		return nil
	}
}
