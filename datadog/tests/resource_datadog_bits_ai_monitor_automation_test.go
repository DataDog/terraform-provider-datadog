package test

import (
	"context"
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccDatadogBitsAIMonitorAutomation(t *testing.T) {
	ctx, providers, factories := testAccFrameworkMuxProviders(context.Background(), t)
	monitor := fmt.Sprintf(`
resource "datadog_monitor" "test" {
  name = %q
  type = "metric alert"
  query = "avg(last_5m):avg:system.cpu.user{host:terraform-bits-automation-test} > 1000000"
  message = "Terraform Bits automation acceptance test"
  notify_no_data = false
}`, uniqueEntityName(ctx, t))
	config := func(enabled bool) string {
		return monitor + fmt.Sprintf(`
resource "datadog_bits_ai_monitor_automation" "test" {
  monitor_id = datadog_monitor.test.id
  enabled = %t
}`, enabled)
	}
	const name = "datadog_bits_ai_monitor_automation.test"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: factories,
		Steps: []resource.TestStep{
			{Config: config(false), Check: resource.TestCheckResourceAttr(name, "enabled", "false")},
			{Config: config(true), Check: resource.TestCheckResourceAttr(name, "enabled", "true")},
			{ResourceName: name, ImportState: true, ImportStateVerify: true},
			{Config: config(false), Check: resource.TestCheckResourceAttr(name, "enabled", "false")},
			{Config: config(true), Check: resource.TestCheckResourceAttr(name, "enabled", "true")},
			{
				// Removing the setting disables investigations and preserves the monitor.
				Config: monitor,
				Check: func(state *terraform.State) error {
					id, err := strconv.ParseInt(state.RootModule().Resources["datadog_monitor.test"].Primary.ID, 10, 64)
					if err != nil {
						return err
					}
					provider := providers.frameworkProvider
					api := datadogV2.NewBitsAIApi(provider.DatadogApiInstances.HttpClient)
					return retry.RetryContext(ctx, time.Minute, func() *retry.RetryError {
						result, _, err := api.GetMonitorAutomation(provider.Auth, id)
						if err != nil {
							return retry.NonRetryableError(err)
						}
						if result.Data.Attributes.Enabled {
							return retry.RetryableError(fmt.Errorf("monitor automation is still enabled"))
						}
						return nil
					})
				},
			},
		},
	})
}
