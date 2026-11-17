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

	securityMonitoringCreatedConfig := testAccCheckDatadogSecurityMonitoringCreatedConfig(ruleName) +
		testAccSecurityMonitoringRuleUntaggedConfig(ruleName)
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogSecurityMonitoringRuleDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: securityMonitoringCreatedConfig,
			},
			{
				Config: testAccDataSourceSecurityMonitoringRuleFilter(securityMonitoringCreatedConfig, ruleName, ""),
				Check:  securityMonitoringCheckRules(ruleName, ruleName+untaggedRuleSuffix),
			},
			{
				Config: testAccDataSourceSecurityMonitoringRuleFilter(securityMonitoringCreatedConfig, ruleName, `tags_filter = ["i:tomato"]`),
				Check:  securityMonitoringCheckRules(ruleName),
			},
			{
				Config: testAccDataSourceSecurityMonitoringRuleFilter(securityMonitoringCreatedConfig, ruleName, "user_only_filter = true"),
				Check:  securityMonitoringCheckRules(ruleName, ruleName+untaggedRuleSuffix),
			},
			{
				Config: testAccDataSourceSecurityMonitoringRuleFilter(securityMonitoringCreatedConfig, ruleName, "default_only_filter = true"),
				Check:  securityMonitoringCheckRules(),
			},
		},
	})
}

// The rule list is org-wide and other tests create and delete rules in parallel.
// Each step adds name_filter on the unique rule name (filters combine with AND), so
// only the two rules of this test can match. The untagged rule makes the tags filter
// observable. Both rules are user rules, so default_only_filter must return none.
func securityMonitoringCheckRules(names ...string) resource.TestCheckFunc {
	checks := []resource.TestCheckFunc{
		func(state *terraform.State) error { return securityMonitoringCheckRuleCount(state, len(names)) },
	}
	for _, name := range names {
		checks = append(checks, resource.TestCheckTypeSetElemNestedAttrs(tfSecurityRulesSource, "rules.*", map[string]string{"name": name}))
	}
	return resource.ComposeTestCheckFunc(checks...)
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

const untaggedRuleSuffix = "-untagged"

func testAccSecurityMonitoringRuleUntaggedConfig(ruleName string) string {
	return fmt.Sprintf(`
resource "datadog_security_monitoring_rule" "acceptance_test_untagged" {
    name    = "%s%s"
    message = "acceptance rule triggered"
    enabled = false

    query {
        name            = "first"
        query           = "does not really match much"
        aggregation     = "count"
        data_source     = "logs"
        group_by_fields = ["host"]
    }

    case {
        name      = "high case"
        status    = "high"
        condition = "first > 3"
    }

    options {
        evaluation_window   = 300
        keep_alive          = 600
        max_signal_duration = 900
    }

    tags = ["u:tomato"]
}
`, ruleName, untaggedRuleSuffix)
}

func testAccDataSourceSecurityMonitoringRuleFilter(existingRuleConfig, ruleName, extraFilter string) string {
	return existingRuleConfig + fmt.Sprintf(`
data "datadog_security_monitoring_rules" "acceptance_test" {
    name_filter = "%s"
    %s
}
`, ruleName, extraFilter)
}
