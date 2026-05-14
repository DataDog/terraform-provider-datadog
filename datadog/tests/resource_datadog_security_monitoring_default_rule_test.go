package test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

const tfSecurityDefaultRuleName = "datadog_security_monitoring_default_rule.acceptance_test"

// runDefaultRuleAcceptanceTest is the shared runner for FW-only acceptance tests.
// It handles the datasource discovery → import → apply(applyConfig) sequence.
func runDefaultRuleAcceptanceTest(t *testing.T, applyConfig string, applyCheck resource.TestCheckFunc) {
	t.Helper()
	_, _, accProviders := testAccFrameworkMuxProviders(context.Background(), t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: accProviders,
		Steps: []resource.TestStep{
			// Define an existing default rule as one we want to import
			{
				Config: testAccDatadogSecurityMonitoringDefaultDatasource(),
			},
			// Import the rule
			{
				Config:             testAccCheckDatadogSecurityMonitoringDefaultNoop(),
				ResourceName:       tfSecurityDefaultRuleName,
				ImportState:        true,
				ImportStateIdFunc:  idFromDatasource,
				ImportStatePersist: true,
			},
			// Take the base resource
			{
				Config: testAccCheckDatadogSecurityMonitoringDefaultNoop(),
			},
			// Apply the scenario config and run checks
			{
				Config: applyConfig,
				Check:  applyCheck,
			},
			// Restore the base resource
			{
				Config: testAccCheckDatadogSecurityMonitoringDefaultNoop(),
			},
		},
	})
}

// testAccDatadogSecurityMonitoringDefaultRuleConfig is the base config builder.
// decreaseCriticality controls options.decrease_criticality_based_on_env.
// extra is appended inside the resource block after query/case/options.
//
// options is always included for LOG_DETECTION rules
func testAccDatadogSecurityMonitoringDefaultRuleConfig(decreaseCriticality bool, extra string) string {
	return fmt.Sprintf(`
resource "datadog_security_monitoring_default_rule" "acceptance_test" {
	query {
	}
	options {
		decrease_criticality_based_on_env = %t
	}
%s}
`, decreaseCriticality, extra)
}

// defaultRuleBaseCase is the minimal case block shared across scenarios.
const defaultRuleBaseCase = `	case {
		status        = "medium"
		notifications = [] 
	}
`

func TestAccDatadogSecurityMonitoringDefaultRule_DeprecationWarning(t *testing.T) {
	if !isReplaying() {
		t.Skip("this is a replay-only test")
		return
	}

	_, _, accProviders := testAccFrameworkMuxProviders(context.Background(), t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: accProviders,
		Steps: []resource.TestStep{
			// Define an existing rule
			{
				Config: testAccDatadogSecurityMonitoringDefaultDatasource(),
			},
			// Import the rule
			{
				Config:             testAccCheckDatadogSecurityMonitoringDefaultNoop(),
				ResourceName:       tfSecurityDefaultRuleName,
				ImportState:        true,
				ImportStateIdFunc:  idFromDatasource,
				ImportStatePersist: true,
			},
			// Change the "decrease criticality" flag
			// For this specific test, we manually changed the cassette recording to set a deprecation date on the rule
			// As of Jan 17, 2023, the TF testing framework does not provide a way to make assertions on warning
			// See https://github.com/hashicorp/terraform-plugin-sdk/issues/864
			// However, this test makes sure nothing breaks when the warning is returned
			{
				Config: testAccDatadogSecurityMonitoringDefaultRuleDynamicCriticality(),
				Check:  testAccCheckDatadogSecurityMonitoringDefaultDynamicCriticality(),
			},
		},
	})
}

func TestAccDatadogSecurityMonitoringDefaultRule_Basic(t *testing.T) {
	runDefaultRuleAcceptanceTest(t,
		testAccDatadogSecurityMonitoringDefaultRuleDynamicCriticality(),
		testAccCheckDatadogSecurityMonitoringDefaultDynamicCriticality(),
	)
}

func TestAccDatadogSecurityMonitoringDefaultRule_AddTag(t *testing.T) {
	runDefaultRuleAcceptanceTest(t,
		testAccDatadogSecurityMonitoringDefaultRuleAddTag(),
		testAccCheckDatadogSecurityMonitoringDefaultRuleAddTag(),
	)
}

func TestAccDatadogSecurityMonitoringDefaultRule_CustomMessage(t *testing.T) {
	runDefaultRuleAcceptanceTest(t,
		testAccDatadogSecurityMonitoringDefaultRuleCustomMessage(),
		testAccCheckDatadogSecurityMonitoringDefaultCustomMessage(),
	)
}

func TestAccDatadogSecurityMonitoringDefaultRule_CustomMessageClear(t *testing.T) {
	runDefaultRuleAcceptanceTest(t,
		testAccDatadogSecurityMonitoringDefaultRuleCustomMessageClear(),
		testAccCheckDatadogSecurityMonitoringDefaultCustomMessageClear(),
	)
}

func TestAccDatadogSecurityMonitoringDefaultRule_CustomName(t *testing.T) {
	runDefaultRuleAcceptanceTest(t,
		testAccDatadogSecurityMonitoringDefaultRuleCustomName(),
		testAccCheckDatadogSecurityMonitoringDefaultCustomName(),
	)
}

func TestAccDatadogSecurityMonitoringDefaultRule_CustomNameClear(t *testing.T) {
	runDefaultRuleAcceptanceTest(t,
		testAccDatadogSecurityMonitoringDefaultRuleCustomNameClear(),
		testAccCheckDatadogSecurityMonitoringDefaultCustomNameClear(),
	)
}

func TestAccDatadogSecurityMonitoringDefaultRule_Enabled(t *testing.T) {
	runDefaultRuleAcceptanceTest(t,
		testAccDatadogSecurityMonitoringDefaultRuleEnabled(),
		testAccCheckDatadogSecurityMonitoringDefaultEnabled(),
	)
}

func TestAccDatadogSecurityMonitoringDefaultRule_CustomStatus(t *testing.T) {
	runDefaultRuleAcceptanceTest(t,
		testAccDatadogSecurityMonitoringDefaultRuleCustomStatus(),
		testAccCheckDatadogSecurityMonitoringDefaultCustomStatus(),
	)
}

// --- helpers ---

func idFromDatasource(state *terraform.State) (string, error) {
	resources := state.RootModule().Resources
	resourceState := resources["data.datadog_security_monitoring_rules.bruteforce"]
	return resourceState.Primary.Attributes["rule_ids.0"], nil
}

func testAccDatadogSecurityMonitoringDefaultDatasource() string {
	return `
data "datadog_security_monitoring_rules" "bruteforce" {
	tags_filter = ["source:cloudtrail"]
	default_only_filter = "true"
}
`
}

func testAccCheckDatadogSecurityMonitoringDefaultNoop() string {
	return `
data "datadog_security_monitoring_rules" "bruteforce" {
	tags_filter = ["source:cloudtrail"]
	default_only_filter = "true"
}

` + testAccDatadogSecurityMonitoringDefaultRuleConfig(false, defaultRuleBaseCase)
}

func testAccDatadogSecurityMonitoringDefaultRuleDynamicCriticality() string {
	return testAccDatadogSecurityMonitoringDefaultRuleConfig(true, defaultRuleBaseCase+`	custom_tags = [
		"testtag:newtag",
	]
`)
}

func testAccCheckDatadogSecurityMonitoringDefaultDynamicCriticality() resource.TestCheckFunc {
	return resource.ComposeTestCheckFunc(
		resource.TestCheckResourceAttr(
			tfSecurityDefaultRuleName, "options.0.decrease_criticality_based_on_env", "true"),
	)
}

func testAccDatadogSecurityMonitoringDefaultRuleAddTag() string {
	return testAccDatadogSecurityMonitoringDefaultRuleConfig(false, defaultRuleBaseCase+`	custom_tags = [
		"testtag:newtag",
	]
`)
}

func testAccCheckDatadogSecurityMonitoringDefaultRuleAddTag() resource.TestCheckFunc {
	return resource.ComposeTestCheckFunc(
		resource.TestCheckResourceAttr(
			tfSecurityDefaultRuleName, "custom_tags.#", "1"),
		resource.TestCheckResourceAttr(
			tfSecurityDefaultRuleName, "custom_tags.0", "testtag:newtag"),
	)
}

func testAccDatadogSecurityMonitoringDefaultRuleCustomMessage() string {
	return testAccDatadogSecurityMonitoringDefaultRuleConfig(false, defaultRuleBaseCase+`	custom_message = "overridden by test"
`)
}

func testAccCheckDatadogSecurityMonitoringDefaultCustomMessage() resource.TestCheckFunc {
	return resource.ComposeTestCheckFunc(
		resource.TestCheckResourceAttr(
			tfSecurityDefaultRuleName, "custom_message", "overridden by test"),
	)
}

func testAccDatadogSecurityMonitoringDefaultRuleCustomMessageClear() string {
	return testAccDatadogSecurityMonitoringDefaultRuleConfig(false, defaultRuleBaseCase+`	custom_message = ""
`)
}

func testAccCheckDatadogSecurityMonitoringDefaultCustomMessageClear() resource.TestCheckFunc {
	return resource.ComposeTestCheckFunc(
		resource.TestCheckResourceAttr(
			tfSecurityDefaultRuleName, "custom_message", ""),
	)
}

func testAccDatadogSecurityMonitoringDefaultRuleCustomName() string {
	return testAccDatadogSecurityMonitoringDefaultRuleConfig(false, defaultRuleBaseCase+`	custom_name = "Test override name"
`)
}

func testAccCheckDatadogSecurityMonitoringDefaultCustomName() resource.TestCheckFunc {
	return resource.ComposeTestCheckFunc(
		resource.TestCheckResourceAttr(
			tfSecurityDefaultRuleName, "custom_name", "Test override name"),
	)
}

func testAccDatadogSecurityMonitoringDefaultRuleCustomNameClear() string {
	return testAccDatadogSecurityMonitoringDefaultRuleConfig(false, defaultRuleBaseCase+`	custom_name = ""
`)
}

func testAccCheckDatadogSecurityMonitoringDefaultCustomNameClear() resource.TestCheckFunc {
	return resource.ComposeTestCheckFunc(
		resource.TestCheckResourceAttr(
			tfSecurityDefaultRuleName, "custom_name", ""),
	)
}

func testAccDatadogSecurityMonitoringDefaultRuleEnabled() string {
	return testAccDatadogSecurityMonitoringDefaultRuleConfig(false, defaultRuleBaseCase+`	enabled = false
`)
}

func testAccCheckDatadogSecurityMonitoringDefaultEnabled() resource.TestCheckFunc {
	return resource.ComposeTestCheckFunc(
		resource.TestCheckResourceAttr(
			tfSecurityDefaultRuleName, "enabled", "false"),
	)
}

func testAccDatadogSecurityMonitoringDefaultRuleCustomStatus() string {
	return testAccDatadogSecurityMonitoringDefaultRuleConfig(false, `	case {
		status        = "medium"
		custom_status = "high"
		notifications = []
	}
`)
}

func testAccCheckDatadogSecurityMonitoringDefaultCustomStatus() resource.TestCheckFunc {
	return resource.ComposeTestCheckFunc(
		resource.TestCheckResourceAttr(
			tfSecurityDefaultRuleName, "case.0.custom_status", "high"),
	)
}

func TestAccDatadogSecurityMonitoringDefaultRule_CustomQueryExtension(t *testing.T) {
	runDefaultRuleAcceptanceTest(t,
		testAccDatadogSecurityMonitoringDefaultRuleCustomQueryExtension(),
		testAccCheckDatadogSecurityMonitoringDefaultCustomQueryExtension(),
	)
}

func testAccDatadogSecurityMonitoringDefaultRuleCustomQueryExtension() string {
	return fmt.Sprintf(`
resource "datadog_security_monitoring_default_rule" "acceptance_test" {
	query {
		custom_query_extension = "env:test-acceptance"
	}
	options {
		decrease_criticality_based_on_env = false
	}
%s}
`, defaultRuleBaseCase)
}

func testAccCheckDatadogSecurityMonitoringDefaultCustomQueryExtension() resource.TestCheckFunc {
	return resource.ComposeTestCheckFunc(
		resource.TestCheckResourceAttr(
			tfSecurityDefaultRuleName, "query.0.custom_query_extension", "env:test-acceptance"),
	)
}
