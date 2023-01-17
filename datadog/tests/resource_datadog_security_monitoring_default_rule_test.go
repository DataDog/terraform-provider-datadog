package test

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

const tfSecurityDefaultRuleName = "datadog_security_monitoring_default_rule.acceptance_test"

func TestAccDatadogSecurityMonitoringDefaultRule_Basic(t *testing.T) {
	t.Parallel()
	_, accProviders := testAccProviders(context.Background(), t)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
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
			// Change the "decrease criticality" flag
			{
				Config: testAccDatadogSecurityMonitoringDefaultRuleDynamicCriticality(),
				Check:  testAccCheckDatadogSecurityMonitoringDefaultDynamicCriticality(),
			},
		},
	})
}

func TestAccDatadogSecurityMonitoringDefaultRule_DeprecationWarning(t *testing.T) {
	t.Parallel()
	_, accProviders := testAccProviders(context.Background(), t)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
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
			{
				Config: testAccDatadogSecurityMonitoringDefaultRuleDynamicCriticality(),
				Check:  testAccCheckDatadogSecurityMonitoringDefaultDynamicCriticality(),
			},
		},
	})
}

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

resource "datadog_security_monitoring_default_rule" "acceptance_test" {
}
`
}

func testAccDatadogSecurityMonitoringDefaultRuleDynamicCriticality() string {
	return `
resource "datadog_security_monitoring_default_rule" "acceptance_test" {
    options {
        decrease_criticality_based_on_env = true
    }
}
`
}

func testAccCheckDatadogSecurityMonitoringDefaultDynamicCriticality() resource.TestCheckFunc {
	return resource.ComposeTestCheckFunc(
		resource.TestCheckResourceAttr(
			tfSecurityDefaultRuleName, "options.0.decrease_criticality_based_on_env", "true"),
	)
}
