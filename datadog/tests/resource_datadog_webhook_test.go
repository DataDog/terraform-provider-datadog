package test

import (
	"context"
	"fmt"
	"testing"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/fwprovider"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccDatadogWebhook_Basic(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	uniqueName := uniqueEntityName(ctx, t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogWebhookDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogWebhookBasicConfig(uniqueName),
				Check: resource.ComposeTestCheckFunc(
					// testAccCheckDatadogWebhookExists(accProvider, "datadog_webhook.foo"),
					testAccCheckDatadogWebhookExists(providers.frameworkProvider, "datadog_webhook.foo"),
					resource.TestCheckResourceAttr("datadog_webhook.foo", "name", uniqueName),
					resource.TestCheckResourceAttr("datadog_webhook.foo", "url", "http://example.com"),
					resource.TestCheckResourceAttr("datadog_webhook.foo", "encode_as", "json"),
					resource.TestCheckResourceAttr("datadog_webhook.foo", "custom_headers", "{\"test\":\"test\"}"),
					resource.TestCheckResourceAttr("datadog_webhook.foo", "payload", "{\"body\": \"$EVENT_MSG\", \"last_updated\": \"$LAST_UPDATED\", \"event_type\": \"$EVENT_TYPE\", \"title\": \"$EVENT_TITLE\", \"date\": \"$DATE\", \"org\": {\"id\": \"$ORG_ID\", \"name\": \"$ORG_NAME\"}, \"id\": \"$ID\"}"),
					// testAccCheckDatadogWebhookExists(accProvider, "datadog_webhook.foo2"),
					testAccCheckDatadogWebhookExists(providers.frameworkProvider, "datadog_webhook.foo2"),
					resource.TestCheckResourceAttr("datadog_webhook.foo2", "name", uniqueName+"2"),
					resource.TestCheckResourceAttr("datadog_webhook.foo2", "url", "http://example.com"),
					resource.TestCheckResourceAttr("datadog_webhook.foo2", "encode_as", "json"),
					resource.TestCheckResourceAttr("datadog_webhook.foo2", "custom_headers", "{\"test\":\"test\"}"),
					resource.TestCheckResourceAttr("datadog_webhook.foo2", "payload", "{\"body\": \"$EVENT_MSG\", \"last_updated\": \"$LAST_UPDATED\", \"event_type\": \"$EVENT_TYPE\", \"title\": \"$EVENT_TITLE\", \"date\": \"$DATE\", \"org\": {\"id\": \"$ORG_ID\", \"name\": \"$ORG_NAME\"}, \"id\": \"$ID\"}"),
				),
			},
			{
				Config: testAccCheckDatadogWebhookBasicConfigUpdated(uniqueName),
				Check: resource.ComposeTestCheckFunc(
					// testAccCheckDatadogWebhookExists(accProvider, "datadog_webhook.foo"),
					testAccCheckDatadogWebhookExists(providers.frameworkProvider, "datadog_webhook.foo"),
					resource.TestCheckResourceAttr("datadog_webhook.foo", "name", uniqueName+"UPDATED"),
					resource.TestCheckResourceAttr("datadog_webhook.foo", "url", "http://example.com/updated"),
					resource.TestCheckResourceAttr("datadog_webhook.foo", "encode_as", "json"),
					resource.TestCheckResourceAttr("datadog_webhook.foo", "custom_headers", "{\"test\":\"test-updated\"}"),
					resource.TestCheckResourceAttr("datadog_webhook.foo", "payload", "{\"custom\":\"payload\"}"),
					// testAccCheckDatadogWebhookExists(accProvider, "datadog_webhook.foo2"),
					testAccCheckDatadogWebhookExists(providers.frameworkProvider, "datadog_webhook.foo2"),
					resource.TestCheckResourceAttr("datadog_webhook.foo2", "name", uniqueName+"2UPDATED"),
					resource.TestCheckResourceAttr("datadog_webhook.foo2", "url", "http://example.com/updated"),
					resource.TestCheckResourceAttr("datadog_webhook.foo2", "encode_as", "json"),
					resource.TestCheckResourceAttr("datadog_webhook.foo2", "custom_headers", "{\"test\":\"test-updated\"}"),
					resource.TestCheckResourceAttr("datadog_webhook.foo2", "payload", "{\"custom\":\"payload\"}"),
				),
			},
		},
	})
}

func testAccCheckDatadogWebhookBasicConfig(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_webhook" "foo" {
  name           = "%[1]s"
  url            = "http://example.com"
  encode_as      = "json"
  custom_headers = jsonencode({"test": "test"})
}
resource "datadog_webhook" "foo2" {
  name           = "%[1]s2"
  url            = "http://example.com"
  encode_as      = "json"
  custom_headers = jsonencode({"test": "test"})
}`, uniq)
}

func testAccCheckDatadogWebhookBasicConfigUpdated(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_webhook" "foo" {
  name           = "%[1]sUPDATED"
  url            = "http://example.com/updated"
  encode_as      = "json"
  custom_headers = jsonencode({"test": "test-updated"})
  payload        = jsonencode({"custom": "payload"})
}
resource "datadog_webhook" "foo2" {
  name           = "%[1]s2UPDATED"
  url            = "http://example.com/updated"
  encode_as      = "json"
  custom_headers = jsonencode({"test": "test-updated"})
  payload        = jsonencode({"custom": "payload"})
}`, uniq)
}

func testAccCheckDatadogWebhookExists(accProvider *fwprovider.FrameworkProvider, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		apiInstances := accProvider.DatadogApiInstances
		auth := accProvider.Auth

		id := s.RootModule().Resources[name].Primary.ID
		_, httpresp, err := apiInstances.GetWebhooksIntegrationApiV1().GetWebhooksIntegration(auth, id)
		if err != nil {
			return utils.TranslateClientError(err, httpresp, "error checking webhooks custom variable existence")
		}
		return nil
	}
}

func testAccCheckDatadogWebhookDestroy(accProvider *fwprovider.FrameworkProvider) func(*terraform.State) error {
	return func(s *terraform.State) error {
		apiInstances := accProvider.DatadogApiInstances
		auth := accProvider.Auth
		for _, r := range s.RootModule().Resources {
			if r.Type != "datadog_webhook" {
				continue
			}

			var err error
			id := r.Primary.ID

			_, httpResp, err := apiInstances.GetWebhooksIntegrationApiV1().GetWebhooksIntegration(auth, id)
			if err != nil {
				if httpResp != nil && httpResp.StatusCode == 404 {
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
