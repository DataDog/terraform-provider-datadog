package test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/fwprovider"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccDatadogWebhookCustomVariable_Basic(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	uniqueName := strings.ToUpper(strings.ReplaceAll(uniqueEntityName(ctx, t), "-", "_"))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogWebhookDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogWebhookCustomVariableBasicConfig(uniqueName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogWebhookCustomVariableExists(providers.frameworkProvider, "datadog_webhook_custom_variable.foo"),
					testAccCheckDatadogWebhookCustomVariableExists(providers.frameworkProvider, "datadog_webhook_custom_variable.foo2"),
					resource.TestCheckResourceAttr("datadog_webhook_custom_variable.foo", "name", uniqueName),
					resource.TestCheckResourceAttr("datadog_webhook_custom_variable.foo", "value", "test-value"),
					resource.TestCheckResourceAttr("datadog_webhook_custom_variable.foo", "is_secret", "false"),
					resource.TestCheckResourceAttr("datadog_webhook_custom_variable.foo2", "name", uniqueName+"2"),
					resource.TestCheckResourceAttr("datadog_webhook_custom_variable.foo2", "value", "test-value"),
					resource.TestCheckResourceAttr("datadog_webhook_custom_variable.foo2", "is_secret", "false"),
				),
			},
			{
				Config: testAccCheckDatadogWebhookCustomVariableBasicConfigUpdated(uniqueName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogWebhookCustomVariableExists(providers.frameworkProvider, "datadog_webhook_custom_variable.foo"),
					testAccCheckDatadogWebhookCustomVariableExists(providers.frameworkProvider, "datadog_webhook_custom_variable.foo2"),
					resource.TestCheckResourceAttr("datadog_webhook_custom_variable.foo", "name", uniqueName+"UPDATED"),
					resource.TestCheckResourceAttr("datadog_webhook_custom_variable.foo", "value", "test-value-updated"),
					resource.TestCheckResourceAttr("datadog_webhook_custom_variable.foo", "is_secret", "true"),
					resource.TestCheckResourceAttr("datadog_webhook_custom_variable.foo2", "name", uniqueName+"2UPDATED"),
					resource.TestCheckResourceAttr("datadog_webhook_custom_variable.foo2", "value", "test-value-updated"),
					resource.TestCheckResourceAttr("datadog_webhook_custom_variable.foo2", "is_secret", "true"),
				),
			},
		},
	})
}

func testAccCheckDatadogWebhookCustomVariableBasicConfig(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_webhook_custom_variable" "foo" {
  name      = "%[1]s"
  value     = "test-value"
  is_secret = false
}
resource "datadog_webhook_custom_variable" "foo2" {
  name      = "%[1]s2"
  value     = "test-value"
  is_secret = false
}`, uniq)
}

func testAccCheckDatadogWebhookCustomVariableBasicConfigUpdated(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_webhook_custom_variable" "foo" {
  name      = "%[1]sUPDATED"
  value     = "test-value-updated"
  is_secret = true
}
resource "datadog_webhook_custom_variable" "foo2" {
  name      = "%[1]s2UPDATED"
  value     = "test-value-updated"
  is_secret = true
}`, uniq)
}

func testAccCheckDatadogWebhookCustomVariableExists(accProvider *fwprovider.FrameworkProvider, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		apiInstances := accProvider.DatadogApiInstances
		auth := accProvider.Auth

		id := s.RootModule().Resources[name].Primary.ID
		_, httpresp, err := apiInstances.GetWebhooksIntegrationApiV1().GetWebhooksIntegrationCustomVariable(auth, id)
		if err != nil {
			return utils.TranslateClientError(err, httpresp, "error checking webhooks custom variable existence")
		}
		return nil
	}
}
func testAccCheckDatadogWebhookCustomVariableDestroy(accProvider *fwprovider.FrameworkProvider) func(*terraform.State) error {
	return func(s *terraform.State) error {
		apiInstances := accProvider.DatadogApiInstances
		auth := accProvider.Auth

		for _, r := range s.RootModule().Resources {
			if r.Type != "datadog_webhook_custom_variable" {
				continue
			}

			var err error
			id := r.Primary.ID

			_, httpResp, err := apiInstances.GetWebhooksIntegrationApiV1().GetWebhooksIntegrationCustomVariable(auth, id)
			if err != nil {
				if httpResp != nil && httpResp.StatusCode == 404 {
					continue
				} else {
					return fmt.Errorf("received an error retrieving webhooks custom variable: %s", err.Error())
				}
			} else {
				return fmt.Errorf("webhooks_custom_variable %s still exists", r.Primary.ID)
			}
		}

		return nil
	}
}
