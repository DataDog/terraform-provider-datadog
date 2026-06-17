package test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDatadogSecurityFindingsMuteRulesOrder(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	uniq := uniqueEntityName(ctx, t)
	resourceName := "datadog_security_findings_mute_rules_order.order"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: accProviders,
		// The order resource has no server-side delete; the rules it references are checked instead.
		CheckDestroy: testAccCheckDatadogSecurityFindingsMuteRuleDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogSecurityFindingsMuteRulesOrderConfig(uniq, "first", "second"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogSecurityFindingsMuteRuleExists(providers.frameworkProvider, "datadog_security_findings_mute_rule.first"),
					testAccCheckDatadogSecurityFindingsMuteRuleExists(providers.frameworkProvider, "datadog_security_findings_mute_rule.second"),
					resource.TestCheckResourceAttr(resourceName, "rule_ids.#", "2"),
					resource.TestCheckResourceAttrPair(resourceName, "rule_ids.0", "datadog_security_findings_mute_rule.first", "id"),
					resource.TestCheckResourceAttrPair(resourceName, "rule_ids.1", "datadog_security_findings_mute_rule.second", "id"),
				),
			},
			{
				// Swap the order of the two rules.
				Config: testAccCheckDatadogSecurityFindingsMuteRulesOrderConfig(uniq, "second", "first"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "rule_ids.#", "2"),
					resource.TestCheckResourceAttrPair(resourceName, "rule_ids.0", "datadog_security_findings_mute_rule.second", "id"),
					resource.TestCheckResourceAttrPair(resourceName, "rule_ids.1", "datadog_security_findings_mute_rule.first", "id"),
				),
			},
		},
	})
}

// testAccCheckDatadogSecurityFindingsMuteRulesOrderConfig builds two mute rules and an order
// resource that lists them in the order given by first/second (each being "first" or "second").
func testAccCheckDatadogSecurityFindingsMuteRulesOrderConfig(uniq, first, second string) string {
	return fmt.Sprintf(`
resource "datadog_security_findings_mute_rule" "first" {
  name    = "%[1]s-first"
  enabled = true
  rule {
    finding_types = ["misconfiguration"]
    query         = "env:dev"
  }
  action {
    reason = "risk_accepted"
  }
}

resource "datadog_security_findings_mute_rule" "second" {
  name    = "%[1]s-second"
  enabled = true
  rule {
    finding_types = ["secret"]
    query         = "env:staging"
  }
  action {
    reason = "no_fix"
  }
}

resource "datadog_security_findings_mute_rules_order" "order" {
  name = "%[1]s-order"
  rule_ids = [
    datadog_security_findings_mute_rule.%[2]s.id,
    datadog_security_findings_mute_rule.%[3]s.id,
  ]
}
`, uniq, first, second)
}
