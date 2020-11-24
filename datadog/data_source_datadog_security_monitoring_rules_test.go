package datadog

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"regexp"
	"testing"
)

const tfSecurityRulesSource = "data.datadog_security_monitoring_rules.acceptance_test"

func TestAccDatadogSecurityMonitoringRuleDatasource(t *testing.T) {
	accProviders, clock, cleanup := testAccProviders(t, initRecorder(t))
	ruleName := uniqueEntityName(clock, t)
	defer cleanup(t)
	accProvider := testAccProvider(t, accProviders)

	matchAny := regexp.MustCompile(".*")
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    accProviders,
		CheckDestroy: testAccCheckDatadogSecurityMonitoringRuleDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				// Create a rule to make sure we have at least one non-default rule
				Config: testAccCheckDatadogSecurityMonitoringCreatedConfig(ruleName),
			},
			{
				// Ideally we would like to check the size and content of the lists, but the terraform
				// testing framework makes this more difficult so for now we only test that we have at least
				// one element
				Config: testAccDataSourceSecurityMonitoringRuleNoFilter(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(tfSecurityRulesSource, "rule_ids.0", matchAny),
					resource.TestMatchResourceAttr(tfSecurityRulesSource, "rules.0", matchAny),
				),
			},
			{
				Config: testAccDataSourceSecurityMonitoringRuleNameFilter(ruleName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(tfSecurityRulesSource, "rule_ids.0", matchAny),
					resource.TestMatchResourceAttr(tfSecurityRulesSource, "rules.0", matchAny),
				),
			},
			{
				Config: testAccDataSourceSecurityMonitoringRuleTagsFilter(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(tfSecurityRulesSource, "rule_ids.0", matchAny),
					resource.TestMatchResourceAttr(tfSecurityRulesSource, "rules.0", matchAny),
				),
			},
			{
				Config: testAccDataSourceSecurityMonitoringRuleDefaultFilter(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(tfSecurityRulesSource, "rule_ids.0", matchAny),
					resource.TestMatchResourceAttr(tfSecurityRulesSource, "rules.0", matchAny),
				),
			},
			{
				Config: testAccDataSourceSecurityMonitoringRuleUserFilter(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(tfSecurityRulesSource, "rule_ids.0", matchAny),
					resource.TestMatchResourceAttr(tfSecurityRulesSource, "rules.0", matchAny),
				),
			},
		},
	})
}

func testAccDataSourceSecurityMonitoringRuleNoFilter() string {
	return `
data "datadog_security_monitoring_rules" "acceptance_test" {
}
`
}

func testAccDataSourceSecurityMonitoringRuleNameFilter(name string) string {
	return fmt.Sprintf(`
data "datadog_security_monitoring_rules" "acceptance_test" {
    name_filter = "%s"
}
`, name)
}

func testAccDataSourceSecurityMonitoringRuleTagsFilter() string {
	return `
data "datadog_security_monitoring_rules" "acceptance_test" {
    tags_filter = ["i:tomato"]
}
`
}
func testAccDataSourceSecurityMonitoringRuleDefaultFilter() string {
	return `
data "datadog_security_monitoring_rules" "acceptance_test" {
	default_only_filter = true
}
`
}

func testAccDataSourceSecurityMonitoringRuleUserFilter() string {
	return `
data "datadog_security_monitoring_rules" "acceptance_test" {
	user_only_filter = true
}
`
}
