package datadog

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func TestAccDatadogMonitorDatasource(t *testing.T) {
	accProviders, _, cleanup := testAccProviders(t, initRecorder(t))
	defer cleanup(t)
	accProvider := testAccProvider(t, accProviders)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    accProviders,
		CheckDestroy: testAccCheckDatadogMonitorDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccDatasourceMonitorNameFilterConfig,
				// Because of the `depends_on` in the datasource, the plan cannot be empty.
				// See https://www.terraform.io/docs/configuration/data-sources.html#data-resource-dependencies
				ExpectNonEmptyPlan: true,
				Check:              checkDatasourceAttrs(accProvider),
			},
			{
				Config: testAccDatasourceMonitorTagsFilterConfig,
				// Because of the `depends_on` in the datasource, the plan cannot be empty.
				// See https://www.terraform.io/docs/configuration/data-sources.html#data-resource-dependencies
				ExpectNonEmptyPlan: true,
				Check:              checkDatasourceAttrs(accProvider),
			},
			{
				Config: testAccDatasourceMonitorMonitorTagsFilterConfig,
				// Because of the `depends_on` in the datasource, the plan cannot be empty.
				// See https://www.terraform.io/docs/configuration/data-sources.html#data-resource-dependencies
				ExpectNonEmptyPlan: true,
				Check:              checkDatasourceAttrs(accProvider),
			},
		},
	})
}

func checkDatasourceAttrs(accProvider *schema.Provider) resource.TestCheckFunc {
	return resource.ComposeTestCheckFunc(
		testAccCheckDatadogMonitorExists(accProvider, "datadog_monitor.foo"),
		resource.TestCheckResourceAttr(
			"data.datadog_monitor.foo", "name", "monitor for datasource test"),
		resource.TestCheckResourceAttr(
			"data.datadog_monitor.foo", "message", "some message Notify: @hipchat-channel"),
		resource.TestCheckResourceAttr(
			"data.datadog_monitor.foo", "type", "query alert"),
		resource.TestCheckResourceAttr(
			"data.datadog_monitor.foo", "query", "avg(last_1h):avg:aws.ec2.cpu{environment:foo,host:foo,test_datasource_monitor_scope:foo} by {host} > 2"),
		resource.TestCheckResourceAttr(
			"data.datadog_monitor.foo", "notify_no_data", "false"),
		resource.TestCheckResourceAttr(
			"data.datadog_monitor.foo", "new_host_delay", "600"),
		resource.TestCheckResourceAttr(
			"data.datadog_monitor.foo", "evaluation_delay", "700"),
		resource.TestCheckResourceAttr(
			"data.datadog_monitor.foo", "renotify_interval", "60"),
		resource.TestCheckResourceAttr(
			"data.datadog_monitor.foo", "thresholds.warning", "1"),
		resource.TestCheckResourceAttr(
			"data.datadog_monitor.foo", "thresholds.warning_recovery", "0.5"),
		resource.TestCheckResourceAttr(
			"data.datadog_monitor.foo", "thresholds.critical", "2"),
		resource.TestCheckResourceAttr(
			"data.datadog_monitor.foo", "thresholds.critical_recovery", "1.5"),
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

var testAccMonitorConfig = `
resource "datadog_monitor" "foo" {
  name = "monitor for datasource test"
  type = "query alert"
  message = "some message Notify: @hipchat-channel"
  escalation_message = "the situation has escalated @pagerduty"

  query = "avg(last_1h):avg:aws.ec2.cpu{environment:foo,host:foo,test_datasource_monitor_scope:foo} by {host} > 2"

  thresholds = {
	warning = "1.0"
	critical = "2.0"
	warning_recovery = "0.5"
	critical_recovery = "1.5"
  }

  renotify_interval = 60

  notify_audit = false
  timeout_h = 60
  new_host_delay = 600
  evaluation_delay = 700
  include_tags = true
  require_full_window = true
  locked = false
  tags = ["test_datasource_monitor:foo", "baz"]
}
`

var testAccDatasourceMonitorNameFilterConfig = fmt.Sprintf(`
%s
data "datadog_monitor" "foo" {
  depends_on = [
    datadog_monitor.foo,
  ]
  name_filter = "monitor for datasource test"
}
`, testAccMonitorConfig)

var testAccDatasourceMonitorTagsFilterConfig = fmt.Sprintf(`
%s
data "datadog_monitor" "foo" {
  depends_on = [
    datadog_monitor.foo,
  ]
  tags_filter = [
    "test_datasource_monitor_scope:foo",
  ]
}
`, testAccMonitorConfig)

var testAccDatasourceMonitorMonitorTagsFilterConfig = fmt.Sprintf(`
%s
data "datadog_monitor" "foo" {
  depends_on = [
    datadog_monitor.foo,
  ]
  monitor_tags_filter = [
    "test_datasource_monitor:foo",
  ]
}
`, testAccMonitorConfig)
