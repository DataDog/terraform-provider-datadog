package test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

const tfSecurityDefaultRuleName = "datadog_security_monitoring_default_rule.acceptance_test"

func TestAccDatadogSecurityMonitoringDefaultRule_Basic(t *testing.T) {
	accProviders, _, cleanup := testAccProviders(t, initRecorder(t))
	defer cleanup(t)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: accProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogSecurityMonitoringDefaultDatasource(),
			},
			{
				Config:            testAccCheckDatadogSecurityMonitoringDefaultNoop(),
				ResourceName:      tfSecurityDefaultRuleName,
				ImportState:       true,
				ImportStateIdFunc: idFromDatasource,
			},
			{
				Config:            testAccCheckDatadogSecurityMonitoringDefaultEnable(),
				ResourceName:      tfSecurityDefaultRuleName,
				ImportState:       true,
				ImportStateIdFunc: idFromDatasource,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(tfSecurityDefaultRuleName, "enabled", "true"),
				),
			},
			{
				Config:            testAccCheckDatadogSecurityMonitoringDefaultDisable(),
				ResourceName:      tfSecurityDefaultRuleName,
				ImportState:       true,
				ImportStateIdFunc: idFromDatasource,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(tfSecurityDefaultRuleName, "enabled", "false"),
				),
			},
			{
				Config:            testAccCheckDatadogSecurityMonitoringDefaultNotification(),
				ResourceName:      tfSecurityDefaultRuleName,
				ImportState:       true,
				ImportStateIdFunc: idFromDatasource,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(tfSecurityDefaultRuleName, "enabled", "true"),
					resource.TestCheckResourceAttr(tfSecurityDefaultRuleName, "case.0.status", "high"),
					resource.TestCheckResourceAttr(tfSecurityDefaultRuleName, "case.0.notifications.0", "@tf-test-notification"),
				),
			},
		},
	})
}

func idFromDatasource(state *terraform.State) (string, error) {
	resources := state.RootModule().Resources
	resourceState := resources["data.datadog_security_monitoring_rules.bruteforce"]
	return resourceState.Primary.Attributes["rule_ids.0"], nil
}

func testAccCheckDatadogSecurityMonitoringDefaultDatasource() string {
	return `
data "datadog_security_monitoring_rules" "bruteforce" {
    name_filter = "brute"
}
`
}

func testAccCheckDatadogSecurityMonitoringDefaultNoop() string {
	return fmt.Sprintf(`
data "datadog_security_monitoring_rules" "bruteforce" {
    name_filter = "brute"
}

resource "datadog_security_monitoring_default_rule" "acceptance_test" {
}
`)
}

func testAccCheckDatadogSecurityMonitoringDefaultEnable() string {
	return fmt.Sprintf(`
data "datadog_security_monitoring_rules" "bruteforce" {
	name_filter = "brute"
}

resource "datadog_security_monitoring_default_rule" "acceptance_test" {
	enabled = true
}
`)
}

func testAccCheckDatadogSecurityMonitoringDefaultDisable() string {
	return fmt.Sprintf(`
data "datadog_security_monitoring_rules" "bruteforce" {
	name_filter = "brute"
}

resource "datadog_security_monitoring_default_rule" "acceptance_test" {
	enabled = false
}
`)
}

func testAccCheckDatadogSecurityMonitoringDefaultNotification() string {
	return fmt.Sprintf(`
data "datadog_security_monitoring_rules" "bruteforce" {
	name_filter = "brute"
}

resource "datadog_security_monitoring_default_rule" "acceptance_test" {
	enabled = true

	case {
		status = "high"
		notifications = ["@tf-test-notification"]
	}
}
`)
}
