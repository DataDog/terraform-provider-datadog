package test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDatadogMonitorValidationDatasource_ValidMonitor(t *testing.T) {
	t.Parallel()
	_, _, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	query := "sum(last_10m):sum:datadog.agent.running{!region:ca-central-1} by {region}.as_count() < 10"
	monitorType := "query alert"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: accProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDatasourceMonitorValidationConfig(query, monitorType),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.datadog_monitor_validation.test", "query", query),
					resource.TestCheckResourceAttr("data.datadog_monitor_validation.test", "type", monitorType),
					resource.TestCheckResourceAttr("data.datadog_monitor_validation.test", "valid", "true"),
				),
			},
		},
	})
}

func TestAccDatadogMonitorValidationDatasource_InvalidMonitor(t *testing.T) {
	t.Parallel()
	_, _, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	query := "sumlast_24h):default_zero(sum:datadog.agent.running{}.as_count()) < 1"
	monitorType := "query alert"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: accProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDatasourceMonitorValidationConfig(query, monitorType),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.datadog_monitor_validation.test", "query", query),
					resource.TestCheckResourceAttr("data.datadog_monitor_validation.test", "type", monitorType),
					resource.TestCheckResourceAttr("data.datadog_monitor_validation.test", "valid", "false"),
					resource.TestCheckResourceAttr("data.datadog_monitor_validation.test", "validation_errors.0", "The value provided for parameter 'query' is invalid"),
				),
			},
		},
	})
}

func testAccDatasourceMonitorValidationConfig(query, monitorType string) string {
	return fmt.Sprintf(`
data "datadog_monitor_validation" "test" {
	query = "%s"
	type  = "%s"
}
`, query, monitorType)
}
