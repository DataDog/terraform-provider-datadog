package test

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/terraform-providers/terraform-provider-datadog/datadog"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

const tfSecurityRulesSource = "data.datadog_security_monitoring_rules.acceptance_test"

func TestAccDatadogSecurityMonitoringRuleDatasource(t *testing.T) {
	t.Parallel()
	ctx, accProviders := testAccProviders(context.Background(), t)
	ruleName := uniqueEntityName(ctx, t)
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		CheckDestroy:      testAccCheckDatadogSecurityMonitoringRuleDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				// Create a rule to make sure we have at least one non-default rule
				Config: testAccCheckDatadogSecurityMonitoringCreatedConfig(ruleName),
			},
			{
				Config: testAccDataSourceSecurityMonitoringRuleNoFilter(ruleName),
				Check: resource.ComposeTestCheckFunc(
					securityMonitoringCheckRuleCountNoFilter(accProvider),
				),
			},
			{
				Config: testAccDataSourceSecurityMonitoringRuleNameFilter(ruleName),
				Check: resource.ComposeTestCheckFunc(
					securityMonitoringCheckRuleCountNameFilter(accProvider, ruleName),
				),
			},
			{
				Config: testAccDataSourceSecurityMonitoringRuleTagsFilter(ruleName),
				Check: resource.ComposeTestCheckFunc(
					securityMonitoringCheckRuleCountTagsFilter(accProvider, "i:tomato"),
				),
			},
			{
				Config: testAccDataSourceSecurityMonitoringRuleDefaultFilter(ruleName),
				Check: resource.ComposeTestCheckFunc(
					securityMonitoringCheckRuleCountDefaultFilter(accProvider, true),
				),
			},
			{
				Config: testAccDataSourceSecurityMonitoringRuleUserFilter(ruleName),
				Check: resource.ComposeTestCheckFunc(
					securityMonitoringCheckRuleCountDefaultFilter(accProvider, false),
				),
			},
		},
	})
}

func securityMonitoringCheckRuleCountNoFilter(accProvider func() (*schema.Provider, error)) func(state *terraform.State) error {
	return func(state *terraform.State) error {
		provider, _ := accProvider()
		providerConf := provider.Meta().(datadog.ProviderConfiguration)
		authV2 := providerConf.AuthV2
		client := providerConf.DatadogClientV2

		rulesResponse, _, err := client.SecurityMonitoringApi.ListSecurityMonitoringRules(authV2).PageNumber(0).PageSize(1000).Execute()
		if err != nil {
			return err
		}
		return securityMonitoringCheckRuleCount(state, len(*rulesResponse.Data))
	}
}

func securityMonitoringCheckRuleCountNameFilter(accProvider func() (*schema.Provider, error), name string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		provider, _ := accProvider()
		providerConf := provider.Meta().(datadog.ProviderConfiguration)
		authV2 := providerConf.AuthV2
		client := providerConf.DatadogClientV2

		rulesResponse, _, err := client.SecurityMonitoringApi.ListSecurityMonitoringRules(authV2).PageSize(1000).Execute()
		if err != nil {
			return err
		}

		ruleCount := 0
		for _, rule := range *rulesResponse.Data {
			if strings.Contains(rule.GetName(), name) {
				ruleCount++
			}
		}

		return securityMonitoringCheckRuleCount(state, ruleCount)
	}
}

func securityMonitoringCheckRuleCountTagsFilter(accProvider func() (*schema.Provider, error), filterTag string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		provider, _ := accProvider()
		providerConf := provider.Meta().(datadog.ProviderConfiguration)
		authV2 := providerConf.AuthV2
		client := providerConf.DatadogClientV2
		rulesResponse, _, err := client.SecurityMonitoringApi.ListSecurityMonitoringRules(authV2).PageSize(1000).Execute()
		if err != nil {
			return err
		}

		ruleCount := 0
		for _, rule := range *rulesResponse.Data {
			for _, tag := range rule.GetTags() {
				if strings.Contains(tag, filterTag) {
					ruleCount++
				}
			}
		}
		return securityMonitoringCheckRuleCount(state, ruleCount)
	}
}

func securityMonitoringCheckRuleCountDefaultFilter(accProvider func() (*schema.Provider, error), isDefault bool) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		provider, _ := accProvider()
		providerConf := provider.Meta().(datadog.ProviderConfiguration)
		authV2 := providerConf.AuthV2
		client := providerConf.DatadogClientV2
		rulesResponse, _, err := client.SecurityMonitoringApi.ListSecurityMonitoringRules(authV2).PageSize(1000).Execute()
		if err != nil {
			return err
		}

		ruleCount := 0
		for _, rule := range *rulesResponse.Data {
			if rule.GetIsDefault() == isDefault {
				ruleCount++
			}
		}
		return securityMonitoringCheckRuleCount(state, ruleCount)
	}

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

func testAccDataSourceSecurityMonitoringRuleNoFilter(ruleName string) string {
	return testAccCheckDatadogSecurityMonitoringCreatedConfig(ruleName) + `
data "datadog_security_monitoring_rules" "acceptance_test" {
}
`
}

func testAccDataSourceSecurityMonitoringRuleNameFilter(ruleName string) string {
	return testAccCheckDatadogSecurityMonitoringCreatedConfig(ruleName) + fmt.Sprintf(`
data "datadog_security_monitoring_rules" "acceptance_test" {
    name_filter = "%s"
}
`, ruleName)
}

func testAccDataSourceSecurityMonitoringRuleTagsFilter(ruleName string) string {
	return testAccCheckDatadogSecurityMonitoringCreatedConfig(ruleName) + `
data "datadog_security_monitoring_rules" "acceptance_test" {
    tags_filter = ["i:tomato"]
}
`
}
func testAccDataSourceSecurityMonitoringRuleDefaultFilter(ruleName string) string {
	return testAccCheckDatadogSecurityMonitoringCreatedConfig(ruleName) + `
data "datadog_security_monitoring_rules" "acceptance_test" {
	default_only_filter = true
}
`
}

func testAccDataSourceSecurityMonitoringRuleUserFilter(ruleName string) string {
	return testAccCheckDatadogSecurityMonitoringCreatedConfig(ruleName) + `
data "datadog_security_monitoring_rules" "acceptance_test" {
	user_only_filter = true
}
`
}
