package test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestAccDatadogMonitorsDatasource(t *testing.T) {
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
				Config: testAccDatasourceMonitorsNameFilterConfig(uniq),
				Check:  checkDatasourceMonitorsAttrs(accProvider, uniq, fmt.Sprintf("%s||", uniq)),
			},
			{
				Config: testAccDatasourceMonitorsTagsFilterConfig(uniq),
				Check:  checkDatasourceMonitorsAttrs(accProvider, uniq, fmt.Sprintf("|test_datasource_monitor_scope:%s|", uniq)),
			},
			{
				Config: testAccDatasourceMonitorsMonitorTagsFilterConfig(uniq),
				Check:  checkDatasourceMonitorsAttrs(accProvider, uniq, fmt.Sprintf("||test_datasource_monitor:%s", uniq)),
			},
		},
	})
}

func checkDatasourceMonitorsAttrs(accProvider func() (*schema.Provider, error), uniq, id string) resource.TestCheckFunc {
	return resource.ComposeTestCheckFunc(
		testAccCheckDatadogMonitorExists(accProvider),
		resource.TestCheckResourceAttr("data.datadog_monitors.foo", "monitors.#", "2"),
		resource.TestCheckResourceAttr("data.datadog_monitors.foo", "monitors.0.name", uniq),
		resource.TestCheckResourceAttr("data.datadog_monitors.foo", "monitors.1.name", uniq),
		resource.TestCheckResourceAttr("data.datadog_monitors.foo", "monitors.0.type", "query alert"),
		resource.TestCheckResourceAttr("data.datadog_monitors.foo", "monitors.1.type", "query alert"),
		resource.TestCheckResourceAttr("data.datadog_monitors.foo", "id", id),
	)
}

func testAccMonitorsConfig(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_monitor" "foo" {
  name = "%s"
  type = "query alert"
  message = "some message Notify: @hipchat-channel"
  escalation_message = "the situation has escalated @pagerduty"

  query = "avg(last_4h):anomalies(avg:aws.ec2.cpu{environment:foo,host:foo,test_datasource_monitor_scope:%s} by {host}, 'basic', 2, direction='both', alert_window='last_15m', interval=60, count_default_zero='true') >= 1"

  monitor_thresholds {
	warning = "0.5"
	critical = "1.0"
	warning_recovery = "0.3"
	critical_recovery = "0.7"
  }

	monitor_threshold_windows {
		trigger_window = "last_15m"
		recovery_window = "last_15m"
	}

  renotify_interval = 60

  notify_audit = false
  timeout_h = 60
  new_group_delay = 500
  evaluation_delay = 700
  include_tags = true
  require_full_window = true
  locked = false
  tags = ["test_datasource_monitor:%s", "baz"]
}
resource "datadog_monitor" "bar" {
  name = "%s"
  type = "query alert"
  message = "some message Notify: @hipchat-channel"
  escalation_message = "the situation has escalated @pagerduty"

  query = "avg(last_4h):anomalies(avg:aws.ec2.cpu{environment:foo,host:foo,test_datasource_monitor_scope:%s} by {host}, 'basic', 2, direction='both', alert_window='last_15m', interval=60, count_default_zero='true') >= 1"

  monitor_thresholds {
	warning = 0.5
	critical = 1.0
	warning_recovery = 0.3
	critical_recovery = 0.7
  }

	monitor_threshold_windows {
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
}`, uniq, uniq, uniq, uniq, uniq, uniq)
}

func testAccDatasourceMonitorsNameFilterConfig(uniq string) string {
	return fmt.Sprintf(`
%s
data "datadog_monitors" "foo" {
  depends_on = [
    datadog_monitor.foo,
  ]
  name_filter = "%s"
}`, testAccMonitorsConfig(uniq), uniq)
}

func testAccDatasourceMonitorsTagsFilterConfig(uniq string) string {
	return fmt.Sprintf(`
%s
data "datadog_monitors" "foo" {
  depends_on = [
    datadog_monitor.foo,
  ]
  tags_filter = [
    "test_datasource_monitor_scope:%s",
  ]
}
`, testAccMonitorsConfig(uniq), uniq)
}

func testAccDatasourceMonitorsMonitorTagsFilterConfig(uniq string) string {
	return fmt.Sprintf(`
%s
data "datadog_monitors" "foo" {
  depends_on = [
    datadog_monitor.foo,
  ]
  monitor_tags_filter = [
    "test_datasource_monitor:%s",
  ]
}
`, testAccMonitorsConfig(uniq), uniq)
}
