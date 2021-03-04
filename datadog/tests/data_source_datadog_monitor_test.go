package test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestAccDatadogMonitorDatasource(t *testing.T) {
	t.Parallel()
	ctx, accProviders := testAccProviders(context.Background(), t)
	uniq := strings.ReplaceAll(uniqueEntityName(ctx, t), "-", "_")
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		CheckDestroy:      testAccCheckDatadogMonitorDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccDatasourceMonitorNameFilterConfig(uniq),
				// Because of the `depends_on` in the datasource, the plan cannot be empty.
				// See https://www.terraform.io/docs/configuration/data-sources.html#data-resource-dependencies
				ExpectNonEmptyPlan: true,
				Check:              checkDatasourceAttrs(accProvider, uniq),
			},
			{
				Config: testAccDatasourceMonitorTagsFilterConfig(uniq),
				// Because of the `depends_on` in the datasource, the plan cannot be empty.
				// See https://www.terraform.io/docs/configuration/data-sources.html#data-resource-dependencies
				ExpectNonEmptyPlan: true,
				Check:              checkDatasourceAttrs(accProvider, uniq),
			},
			{
				Config: testAccDatasourceMonitorMonitorTagsFilterConfig(uniq),
				// Because of the `depends_on` in the datasource, the plan cannot be empty.
				// See https://www.terraform.io/docs/configuration/data-sources.html#data-resource-dependencies
				ExpectNonEmptyPlan: true,
				Check:              checkDatasourceAttrs(accProvider, uniq),
			},
		},
	})
}

func checkDatasourceAttrs(accProvider func() (*schema.Provider, error), uniq string) resource.TestCheckFunc {
	return resource.ComposeTestCheckFunc(
		testAccCheckDatadogMonitorExists(accProvider),
		resource.TestCheckResourceAttr(
			"data.datadog_monitor.foo", "name", uniq),
		resource.TestCheckResourceAttr(
			"data.datadog_monitor.foo", "message", "some message Notify: @hipchat-channel"),
		resource.TestCheckResourceAttr(
			"data.datadog_monitor.foo", "type", "query alert"),
		resource.TestCheckResourceAttr(
			"data.datadog_monitor.foo", "query", fmt.Sprintf("avg(last_4h):anomalies(avg:aws.ec2.cpu{environment:foo,host:foo,test_datasource_monitor_scope:%s} by {host}, 'basic', 2, direction='both', alert_window='last_15m', interval=60, count_default_zero='true') >= 1", uniq)),
		resource.TestCheckResourceAttr(
			"data.datadog_monitor.foo", "notify_no_data", "false"),
		resource.TestCheckResourceAttr(
			"data.datadog_monitor.foo", "new_host_delay", "600"),
		resource.TestCheckResourceAttr(
			"data.datadog_monitor.foo", "evaluation_delay", "700"),
		resource.TestCheckResourceAttr(
			"data.datadog_monitor.foo", "renotify_interval", "60"),
		resource.TestCheckResourceAttr(
			"data.datadog_monitor.foo", "thresholds.warning", "0.5"),
		resource.TestCheckResourceAttr(
			"data.datadog_monitor.foo", "monitor_thresholds.0.warning", "0.5"),
		resource.TestCheckResourceAttr(
			"data.datadog_monitor.foo", "thresholds.warning_recovery", "0.3"),
		resource.TestCheckResourceAttr(
			"data.datadog_monitor.foo", "monitor_thresholds.0.warning_recovery", "0.3"),
		resource.TestCheckResourceAttr(
			"data.datadog_monitor.foo", "thresholds.critical", "1"),
		resource.TestCheckResourceAttr(
			"data.datadog_monitor.foo", "monitor_thresholds.0.critical", "1"),
		resource.TestCheckResourceAttr(
			"data.datadog_monitor.foo", "thresholds.critical_recovery", "0.7"),
		resource.TestCheckResourceAttr(
			"data.datadog_monitor.foo", "monitor_thresholds.0.critical_recovery", "0.7"),
		resource.TestCheckResourceAttr(
			"data.datadog_monitor.foo", "require_full_window", "true"),
		resource.TestCheckResourceAttr(
			"data.datadog_monitor.foo", "locked", "false"),
		// Tags are a TypeSet => use a weird way to access members by their hash
		// TF TypeSet is internally represented as a map that maps computed hashes
		// to actual values. Since the hashes are always the same for one value,
		// this is the way to get them.
		resource.TestCheckResourceAttr(
			"data.datadog_monitor.foo", "tags.#", "2"),
		resource.TestCheckResourceAttr(
			"data.datadog_monitor.foo", "tags.2644851163", "baz"),
	)
}

func testAccMonitorConfig(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_monitor" "foo" {
  name = "%s"
  type = "metric alert"
  message = "some message Notify: @hipchat-channel"
  escalation_message = "the situation has escalated @pagerduty"

  query = "avg(last_4h):anomalies(avg:aws.ec2.cpu{environment:foo,host:foo,test_datasource_monitor_scope:%s} by {host}, 'basic', 2, direction='both', alert_window='last_15m', interval=60, count_default_zero='true') >= 1"

  thresholds = {
	warning = "0.5"
	critical = "1.0"
	warning_recovery = "0.3"
	critical_recovery = "0.7"
  }

	threshold_windows = {
		trigger_window = "last_15m"
		recovery_window = "last_15m"
	}

  renotify_interval = 60

  notify_audit = false
  timeout_h = 60
  new_host_delay = 600
  evaluation_delay = 700
  include_tags = true
  require_full_window = true
  locked = false
  tags = ["test_datasource_monitor:%s", "baz"]
}`, uniq, uniq, uniq)
}

func testAccDatasourceMonitorNameFilterConfig(uniq string) string {
	return fmt.Sprintf(`
%s
data "datadog_monitor" "foo" {
  depends_on = [
    datadog_monitor.foo,
  ]
  name_filter = "%s"
}`, testAccMonitorConfig(uniq), uniq)
}

func testAccDatasourceMonitorTagsFilterConfig(uniq string) string {
	return fmt.Sprintf(`
%s
data "datadog_monitor" "foo" {
  depends_on = [
    datadog_monitor.foo,
  ]
  tags_filter = [
    "test_datasource_monitor_scope:%s",
  ]
}
`, testAccMonitorConfig(uniq), uniq)
}

func testAccDatasourceMonitorMonitorTagsFilterConfig(uniq string) string {
	return fmt.Sprintf(`
%s
data "datadog_monitor" "foo" {
  depends_on = [
    datadog_monitor.foo,
  ]
  monitor_tags_filter = [
    "test_datasource_monitor:%s",
  ]
}
`, testAccMonitorConfig(uniq), uniq)
}
