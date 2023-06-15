package test

import (
	"context"
	"fmt"
	"testing"

	"github.com/terraform-providers/terraform-provider-datadog/datadog"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccDatadogMonitorConfigPolicy_Basic(t *testing.T) {
	t.Parallel()
	_, accProviders := testAccProviders(context.Background(), t)
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		CheckDestroy:      testAccCheckDatadogMonitorConfigPolicyDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogMonitorConfigPolicyConfig("test", "tagKey"),
				Check:  createTestCheckFunc(accProvider, "tagKey"),
			},
			{
				Config: testAccCheckDatadogMonitorConfigPolicyConfig("test", "tagKeyUpdated"),
				Check:  createTestCheckFunc(accProvider, "tagKeyUpdated"),
			},
		},
	})
}

func createTestCheckFunc(accProvider func() (*schema.Provider, error), tagKey string) resource.TestCheckFunc {
	return resource.ComposeTestCheckFunc(
		testAccCheckDatadogMonitorConfigPolicyExists(accProvider, "datadog_monitor_config_policy.test"),
		resource.TestCheckResourceAttr("datadog_monitor_config_policy.test", "policy_type", "tag"),
		resource.TestCheckResourceAttr("datadog_monitor_config_policy.test", "tag_policy.0.tag_key", tagKey),
		resource.TestCheckResourceAttr("datadog_monitor_config_policy.test", "tag_policy.0.tag_key_required", "false"),
		resource.TestCheckResourceAttr("datadog_monitor_config_policy.test", "tag_policy.0.valid_tag_values.#", "1"),
		resource.TestCheckResourceAttr("datadog_monitor_config_policy.test", "tag_policy.0.valid_tag_values.0", "value"),
	)
}

func testAccCheckDatadogMonitorConfigPolicyConfig(name string, tagKey string) string {
	return fmt.Sprintf(`
  resource "datadog_monitor_config_policy" "%s" {
    policy_type = "tag"
    tag_policy {
		tag_key          = "%s"
		tag_key_required = false
		valid_tag_values = ["value"]
    }
  }
    `, name, tagKey)
}

func testAccCheckDatadogMonitorConfigPolicyExists(accProvider func() (*schema.Provider, error), resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		provider, _ := accProvider()
		providerConf := provider.Meta().(*datadog.ProviderConfiguration)
		apiInstances := providerConf.DatadogApiInstances
		auth := providerConf.Auth

		for _, r := range s.RootModule().Resources {
			id := r.Primary.ID
			if _, httpresp, err := apiInstances.GetMonitorsApiV2().GetMonitorConfigPolicy(auth, id); err != nil {
				return utils.TranslateClientError(err, httpresp, "error checking monitor config policy existence")
			}
		}
		return nil
	}
}

func testAccCheckDatadogMonitorConfigPolicyDestroy(accProvider func() (*schema.Provider, error)) func(*terraform.State) error {
	return func(s *terraform.State) error {
		provider, _ := accProvider()
		providerConf := provider.Meta().(*datadog.ProviderConfiguration)
		apiInstances := providerConf.DatadogApiInstances
		auth := providerConf.Auth

		apps, _, err := apiInstances.GetMonitorsApiV2().ListMonitorConfigPolicies(auth)
		if err != nil {
			return fmt.Errorf("failed to get monitor config policies")
		}

		for _, r := range s.RootModule().Resources {
			id := r.Primary.ID

			for _, m := range apps.Data {
				if m.GetId() == id {
					return fmt.Errorf("monitor config policy with id %s still exists", id)
				}
			}
		}

		return nil
	}
}
