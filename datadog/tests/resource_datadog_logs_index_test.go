package test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDatadogLogsIndex_Basic(t *testing.T) {
	_, accProviders := testAccProviders(context.Background(), t)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		Steps: []resource.TestStep{
			{
				Config:             testAccCheckDatadogCreateLogsIndexConfig(),
				ExpectNonEmptyPlan: true,
				Check:              resource.TestCheckResourceAttr("datadog_logs_index.sample_index", "name", "main"),
			},
			{
				Config: testAccCheckDatadogUpdateLogsIndexConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("datadog_logs_index.sample_index", "name", "main"),
					resource.TestCheckResourceAttr("datadog_logs_index.sample_index", "daily_limit", "20000"),
					resource.TestCheckResourceAttr("datadog_logs_index.sample_index", "disable_daily_limit", "false"),
					resource.TestCheckResourceAttr("datadog_logs_index.sample_index", "retention_days", "15"),
					resource.TestCheckResourceAttr("datadog_logs_index.sample_index", "filter.0.query", "test:query"),
					resource.TestCheckResourceAttr("datadog_logs_index.sample_index", "exclusion_filter.0.name", "Filter coredns logs"),
					resource.TestCheckResourceAttr("datadog_logs_index.sample_index", "exclusion_filter.0.is_enabled", "true"),
					resource.TestCheckResourceAttr("datadog_logs_index.sample_index", "exclusion_filter.0.filter.0.query", "app:coredns"),
					resource.TestCheckResourceAttr("datadog_logs_index.sample_index", "exclusion_filter.0.filter.0.sample_rate", "0.97"),
				),
			},
			{
				Config:                    testAccCheckDatadogUpdateLogsIndexDisableDailyLimitConfig(),
				PreventPostDestroyRefresh: true,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("datadog_logs_index.sample_index", "name", "main"),
					resource.TestCheckResourceAttr("datadog_logs_index.sample_index", "disable_daily_limit", "true"),
					resource.TestCheckResourceAttr("datadog_logs_index.sample_index", "retention_days", "15"),
					resource.TestCheckResourceAttr("datadog_logs_index.sample_index", "exclusion_filter.#", "0"),
				),
			},
		},
	})
}

func testAccCheckDatadogCreateLogsIndexConfig() string {
	return fmt.Sprintf(`
resource "datadog_logs_index" "sample_index" {
  name           = "main"
  daily_limit    = 200000
  retention_days = 15
  filter {
    query = "non-existent-query"
  }
}
`)
}

func testAccCheckDatadogUpdateLogsIndexConfig() string {
	return fmt.Sprintf(`
resource "datadog_logs_index" "sample_index" {
  name                   = "main"
  daily_limit            = 20000
  disable_daily_limit    = false
  retention_days         = 15
  filter {
    query                = "test:query"
  }
  exclusion_filter {
    name                 = "Filter coredns logs"
    is_enabled           = true
    filter {
      query              = "app:coredns"
      sample_rate        = 0.97
    }
  }
}
`)
}

func testAccCheckDatadogUpdateLogsIndexDisableDailyLimitConfig() string {
	return fmt.Sprintf(`
resource "datadog_logs_index" "sample_index" {
  name                   = "main"
  daily_limit            = 20000
  disable_daily_limit    = true
  retention_days         = 15
  filter {
    query                = "test:query"
  }
}
`)
}
