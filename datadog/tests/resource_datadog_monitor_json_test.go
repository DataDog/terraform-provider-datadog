package test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDatadogMonitorJSONImport(t *testing.T) {
	t.Parallel()
	ctx, accProviders := testAccProviders(context.Background(), t)
	uniqueID := uniqueEntityName(ctx, t)
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		// Use checkMonitorDestroy() from Monitor resource
		CheckDestroy: testAccCheckDatadogMonitorDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogMonitorJSON(uniqueID),
			},
			{
				ResourceName:      "datadog_monitor_json.monitor_json",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccDatadogMonitorJSONBasic(t *testing.T) {
	t.Parallel()
	ctx, accProviders := testAccProviders(context.Background(), t)
	uniq := uniqueEntityName(ctx, t)
	uniqUpdated := fmt.Sprintf("%s-updated", uniq)
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		// Import checkDashboardDestroy() from Dashboard resource
		CheckDestroy: testAccCheckDatadogMonitorDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogMonitorJSON(uniq),
				Check: resource.ComposeTestCheckFunc(
					checkDashboardExists(accProvider),
					resource.TestCheckResourceAttr(
						"datadog_monitor_json.monitor_json", "monitor", fmt.Sprintf("{\"message\":\"Change the message triggers if any host's clock goes out of sync with the time given by NTP. The offset threshold is configured in the Agent's 'ntp.yaml' file.\\n\\nSee [Troubleshooting NTP Offset issues](https://docs.datadoghq.com/agent/troubleshooting/ntp for more details on cause and resolution.\",\"name\":\"%s\",\"options\":{\"include_tags\":true,\"locked\":false,\"new_host_delay\":150,\"notify_audit\":false,\"notify_no_data\":false,\"thresholds\":{\"critical\":1,\"ok\":1,\"warning\":1}},\"priority\":null,\"query\":\"\\\"ntp.in_sync\\\".by(\\\"*\\\").last(2).count_by_status()\",\"tags\":[],\"type\":\"service check\"}", uniq)),
				),
			},
			{
				Config: testAccCheckDatadogMonitorJSON(uniqUpdated),
				Check: resource.ComposeTestCheckFunc(
					checkDashboardExists(accProvider),
					resource.TestCheckResourceAttr(
						"datadog_monitor_json.monitor_json", "monitor", fmt.Sprintf("{\"message\":\"Change the message triggers if any host's clock goes out of sync with the time given by NTP. The offset threshold is configured in the Agent's 'ntp.yaml' file.\\n\\nSee [Troubleshooting NTP Offset issues](https://docs.datadoghq.com/agent/troubleshooting/ntp for more details on cause and resolution.\",\"name\":\"%s\",\"options\":{\"include_tags\":true,\"locked\":false,\"new_host_delay\":150,\"notify_audit\":false,\"notify_no_data\":false,\"thresholds\":{\"critical\":1,\"ok\":1,\"warning\":1}},\"priority\":null,\"query\":\"\\\"ntp.in_sync\\\".by(\\\"*\\\").last(2).count_by_status()\",\"tags\":[],\"type\":\"service check\"}", uniqUpdated)),
				),
			},
		},
	})
}

func testAccCheckDatadogMonitorJSON(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_monitor_json" "monitor_json" {
  monitor =<<-EOF
{
    "name": "%s",
    "type": "service check",
    "query": "\"ntp.in_sync\".by(\"*\").last(2).count_by_status()",
    "message": "Change the message triggers if any host's clock goes out of sync with the time given by NTP. The offset threshold is configured in the Agent's 'ntp.yaml' file.\n\nSee [Troubleshooting NTP Offset issues](https://docs.datadoghq.com/agent/troubleshooting/ntp for more details on cause and resolution.",
    "tags": [],
    "multi": true,
    "options": {
        "include_tags": true,
        "locked": false,
        "new_host_delay": 150,
        "notify_audit": false,
        "notify_no_data": false,
        "thresholds": {
            "warning": 1,
            "ok": 1,
            "critical": 1
        }
    },
    "priority": null,
    "classification": "custom"
}
EOF
}`, uniq)

}
