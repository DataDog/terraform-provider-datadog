package datadog

import (
	"errors"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"testing"

	datadogV2 "github.com/DataDog/datadog-api-client-go/api/v2/datadog"
)

const tfSecurityDefaultRuleName = "datadog_security_monitoring_default_rule.acceptance_test"

func TestAccDatadogSecurityMonitoringDefaultRule_Basic(t *testing.T) {
	accProviders, clock, cleanup := testAccProviders(t, initRecorder(t))
	ruleName := uniqueEntityName(clock, t)
	defer cleanup(t)
	accProvider := testAccProvider(t, accProviders)

	var ruleId string

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    accProviders,
		CheckDestroy: testAccCheckDatadogSecurityMonitoringRuleDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				// Get an existing rule id for next checks, the config is a dummy to be able to use the check
				Config: testAccCheckDatadogSecurityMonitoringCreatedConfig(ruleName),
				Check: func(_ *terraform.State) error {
					var err error
					ruleId, err = getDefaultRuleWithHighCaseId(accProvider)
					return err
				},
			},
			{
				Config: testAccCheckDatadogSecurityMonitoringDefaultNoop(ruleId),
				Check:  resource.TestCheckResourceAttr(tfSecurityDefaultRuleName, "rule_id", ruleId),
			},
			{
				Config: testAccCheckDatadogSecurityMonitoringDefaultEnable(ruleId),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(tfSecurityDefaultRuleName, "rule_id", ruleId),
					resource.TestCheckResourceAttr(tfSecurityDefaultRuleName, "enabled", "true"),
				),
			},
			{
				Config: testAccCheckDatadogSecurityMonitoringDefaultDisable(ruleId),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(tfSecurityDefaultRuleName, "rule_id", ruleId),
					resource.TestCheckResourceAttr(tfSecurityDefaultRuleName, "disabled", "true"),
				),
			},
			{
				Config: testAccCheckDatadogSecurityMonitoringDefaultNotification(ruleId),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(tfSecurityDefaultRuleName, "rule_id", ruleId),
					resource.TestCheckResourceAttr(tfSecurityDefaultRuleName, "enabled", "true"),
					resource.TestCheckResourceAttr(tfSecurityDefaultRuleName, "case.0.status", "high"),
					resource.TestCheckResourceAttr(tfSecurityDefaultRuleName, "case.0.notifications.0", "@tf-test-notification"),
				),
			},
		},
	})
}

func testAccCheckDatadogSecurityMonitoringDefaultNoop(ruleId string) string {
	return fmt.Sprintf(`
resource "datadog_security_monitoring_default_rule" "acceptance_test" {
    rule_id = "%s"
}
`, ruleId)
}

func testAccCheckDatadogSecurityMonitoringDefaultEnable(ruleId string) string {
	return fmt.Sprintf(`
resource "datadog_security_monitoring_default_rule" "acceptance_test" {
    rule_id = "%s"
	enabled = true
}
`, ruleId)
}

func testAccCheckDatadogSecurityMonitoringDefaultDisable(ruleId string) string {
	return fmt.Sprintf(`
resource "datadog_security_monitoring_default_rule" "acceptance_test" {
    rule_id = "%s"
	disabled = true
}
`, ruleId)
}

func testAccCheckDatadogSecurityMonitoringDefaultNotification(ruleId string) string {
	return fmt.Sprintf(`
resource "datadog_security_monitoring_default_rule" "acceptance_test" {
    rule_id = "%s"
	enabled = true

	case {
		status = "high"
		notifications = ["@tf-test-notification"]
	}
}
`, ruleId)
}

func getDefaultRuleWithHighCaseId(provider *schema.Provider) (string, error) {
	providerConf := provider.Meta().(*ProviderConfiguration)
	authV2 := providerConf.AuthV2
	client := providerConf.DatadogClientV2

	response, _, err := client.SecurityMonitoringApi.ListSecurityMonitoringRules(authV2).Execute()
	if err != nil {
		return "", err
	}

	for _, rule := range response.GetData() {
		if rule.GetIsDefault() {
			for _, ruleCase := range rule.GetCases() {
				if ruleCase.GetStatus() == datadogV2.SECURITYMONITORINGRULESEVERITY_HIGH {
					return rule.GetId(), nil
				}
			}
		}
	}

	return "", errors.New("no default rules with high case found")
}
