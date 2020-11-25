package datadog

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"regexp"
	"testing"
)

const tfSecurityDefaultRuleName = "datadog_security_monitoring_default_rule.acceptance_test"

func TestAccDatadogSecurityMonitoringDefaultRule_Basic(t *testing.T) {
	accProviders, _, cleanup := testAccProviders(t, initRecorder(t))
	defer cleanup(t)

	matchAny := regexp.MustCompile(".*")
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: accProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogSecurityMonitoringDefaultNoop(),
				Check:  resource.TestMatchResourceAttr(tfSecurityDefaultRuleName, "rule_id", matchAny),
			},
			{
				Config: testAccCheckDatadogSecurityMonitoringDefaultEnable(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr(tfSecurityDefaultRuleName, "rule_id", matchAny),
					resource.TestCheckResourceAttr(tfSecurityDefaultRuleName, "enabled", "true"),
				),
			},
			{
				Config: testAccCheckDatadogSecurityMonitoringDefaultDisable(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr(tfSecurityDefaultRuleName, "rule_id", matchAny),
					resource.TestCheckResourceAttr(tfSecurityDefaultRuleName, "disabled", "true"),
				),
			},
			{
				Config: testAccCheckDatadogSecurityMonitoringDefaultNotification(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr(tfSecurityDefaultRuleName, "rule_id", matchAny),
					resource.TestCheckResourceAttr(tfSecurityDefaultRuleName, "enabled", "true"),
					resource.TestCheckResourceAttr(tfSecurityDefaultRuleName, "case.0.status", "high"),
					resource.TestCheckResourceAttr(tfSecurityDefaultRuleName, "case.0.notifications.0", "@tf-test-notification"),
				),
			},
		},
	})
}

func testAccCheckDatadogSecurityMonitoringDefaultNoop() string {
	return fmt.Sprintf(`
data "datadog_security_monitoring_rules" "bruteforce" {
    name_filter = "brute"
}

resource "datadog_security_monitoring_default_rule" "acceptance_test" {
    rule_id = "${data.datadog_security_monitoring_rules.bruteforce.rule_ids.0}"
}
`)
}

func testAccCheckDatadogSecurityMonitoringDefaultEnable() string {
	return fmt.Sprintf(`
data "datadog_security_monitoring_rules" "bruteforce" {
	name_filter = "brute"
}

resource "datadog_security_monitoring_default_rule" "acceptance_test" {
    rule_id = "${data.datadog_security_monitoring_rules.bruteforce.rule_ids.0}"
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
    rule_id = "${data.datadog_security_monitoring_rules.bruteforce.rule_ids.0}"
	disabled = true
}
`)
}

func testAccCheckDatadogSecurityMonitoringDefaultNotification() string {
	return fmt.Sprintf(`
data "datadog_security_monitoring_rules" "bruteforce" {
	name_filter = "brute"
}

resource "datadog_security_monitoring_default_rule" "acceptance_test" {
    rule_id = "${data.datadog_security_monitoring_rules.bruteforce.rule_ids.0}"
	enabled = true

	case {
		status = "high"
		notifications = ["@tf-test-notification"]
	}
}
`)
}
