package test

import (
	"context"
	"fmt"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

const tfSecurityRulesSource = "data.datadog_security_monitoring_rules.acceptance_test"

func TestAccDatadogSecurityMonitoringRuleDatasource(t *testing.T) {
	t.Parallel()
	if isRecording() || isReplaying() {
		// Cassette is >100MB, too large for Github
		t.Skip("This test doesn't support recording or replaying")
	}
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	ruleName := uniqueEntityName(ctx, t)
	accProvider := providers.frameworkProvider

	securityMonitoringCreatedConfig := testAccCheckDatadogSecurityMonitoringCreatedConfig(ruleName)
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogSecurityMonitoringRuleDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: securityMonitoringCreatedConfig,
			},
			{
				Config: testAccDataSourceSecurityMonitoringRuleNameFilter(securityMonitoringCreatedConfig, ruleName),
				Check:  securityMonitoringCheckOnlyCreatedRule(ruleName),
			},
			{
				Config: testAccDataSourceSecurityMonitoringRuleTagsFilter(securityMonitoringCreatedConfig, ruleName),
				Check:  securityMonitoringCheckOnlyCreatedRule(ruleName),
			},
			{
				Config: testAccDataSourceSecurityMonitoringRuleUserFilter(securityMonitoringCreatedConfig, ruleName),
				Check:  securityMonitoringCheckOnlyCreatedRule(ruleName),
			},
		},
	})
}

// The rule list is org-wide and other tests create and delete rules in parallel.
// Each step adds name_filter on the unique rule name (filters combine with AND),
// so the only expected match is the rule this test creates.
func securityMonitoringCheckOnlyCreatedRule(ruleName string) resource.TestCheckFunc {
	return resource.ComposeTestCheckFunc(
		func(state *terraform.State) error { return securityMonitoringCheckRuleCount(state, 1) },
		resource.TestCheckResourceAttr(tfSecurityRulesSource, "rules.0.name", ruleName),
		resource.TestCheckResourceAttrPair(tfSecurityRulesSource, "rule_ids.0", "datadog_security_monitoring_rule.acceptance_test", "id"),
	)
}

func securityMonitoringCheckRuleCount(state *terraform.State, responseRuleCount int) error {
	resourceAttributes := state.RootModule().Resources[tfSecurityRulesSource].Primary.Attributes
	ruleIDCount, _ := strconv.Atoi(resourceAttributes["rule_ids.#"])
	rulesCount, _ := strconv.Atoi(resourceAttributes["rules.#"])

	if rulesCount != responseRuleCount || ruleIDCount != responseRuleCount {
		return fmt.Errorf("expected %d rules got %d rules and %d rule ids",
			responseRuleCount, rulesCount, ruleIDCount)
	}
	return nil
}

func testAccDataSourceSecurityMonitoringRuleNameFilter(existingRuleConfig, ruleName string) string {
	return existingRuleConfig + fmt.Sprintf(`
data "datadog_security_monitoring_rules" "acceptance_test" {
    name_filter = "%s"
}
`, ruleName)
}

func testAccDataSourceSecurityMonitoringRuleTagsFilter(existingRuleConfig, ruleName string) string {
	return existingRuleConfig + fmt.Sprintf(`
data "datadog_security_monitoring_rules" "acceptance_test" {
    name_filter = "%s"
    tags_filter = ["i:tomato"]
}
`, ruleName)
}

func testAccDataSourceSecurityMonitoringRuleUserFilter(existingRuleConfig, ruleName string) string {
	return existingRuleConfig + fmt.Sprintf(`
data "datadog_security_monitoring_rules" "acceptance_test" {
    name_filter      = "%s"
    user_only_filter = true
}
`, ruleName)
}
