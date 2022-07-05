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
				Config:            testAccCheckDatadogSecurityMonitoringDefaultDynamicCriticality(),
				ResourceName:      tfSecurityDefaultRuleName,
				ImportState:       true,
				ImportStateIdFunc: idFromDatasource,
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
    name_filter = "docker"
}
`
}

func testAccCheckDatadogSecurityMonitoringDefaultNoop() string {
	return `
data "datadog_security_monitoring_rules" "bruteforce" {
    name_filter = "docker"
}

resource "datadog_security_monitoring_default_rule" "acceptance_test" {
}
`
}

func testAccCheckDatadogSecurityMonitoringDefaultDynamicCriticality() string {
	return `
resource "datadog_security_monitoring_default_rule" "acceptance_test" {
    options {
        decrease_criticality_based_on_env = true
    }
}
`
}

func testAccCheckDatadogSecurityMonitoringDefaultDynamicCriticalityCheck() resource.TestCheckFunc {
	return resource.ComposeTestCheckFunc(
		resource.TestCheckResourceAttr(
			tfSecurityDefaultRuleName, "options.0.decrease_criticality_based_on_env", "true"),
	)
}
