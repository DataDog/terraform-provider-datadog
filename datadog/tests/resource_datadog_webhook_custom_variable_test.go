package test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/terraform-providers/terraform-provider-datadog/datadog"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccDatadogWebhookCustomVariable_Basic(t *testing.T) {
	t.Parallel()
	ctx, accProviders := testAccProviders(context.Background(), t)
	uniqueName := strings.ToUpper(strings.ReplaceAll(uniqueEntityName(ctx, t), "-", "_"))
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		CheckDestroy:      testAccCheckDatadogWebhookCustomVariableDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogWebhookCustomVariableBasicConfig(uniqueName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogWebhookCustomVariableExists(accProvider, "datadog_webhook_custom_variable.foo"),
					testAccCheckDatadogWebhookCustomVariableExists(accProvider, "datadog_webhook_custom_variable.foo2"),
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
					testAccCheckDatadogWebhookCustomVariableExists(accProvider, "datadog_webhook_custom_variable.foo"),
					testAccCheckDatadogWebhookCustomVariableExists(accProvider, "datadog_webhook_custom_variable.foo2"),
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

func testAccCheckDatadogWebhookCustomVariableExists(accProvider func() (*schema.Provider, error), name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		provider, _ := accProvider()
		providerConf := provider.Meta().(*datadog.ProviderConfiguration)
		apiInstances := providerConf.DatadogApiInstances
		auth := providerConf.Auth

		id := s.RootModule().Resources[name].Primary.ID
		_, httpresp, err := apiInstances.GetWebhooksIntegrationApiV1().GetWebhooksIntegrationCustomVariable(auth, id)
		if err != nil {
			return utils.TranslateClientError(err, httpresp, "error checking webhooks custom variable existence")
		}
		return nil
	}
}

func testAccCheckDatadogWebhookCustomVariableDestroy(accProvider func() (*schema.Provider, error)) func(*terraform.State) error {
	return func(s *terraform.State) error {
		provider, _ := accProvider()
		meta := provider.Meta()
		providerConf := meta.(*datadog.ProviderConfiguration)
		apiInstances := providerConf.DatadogApiInstances
		auth := providerConf.Auth
		for _, r := range s.RootModule().Resources {
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
