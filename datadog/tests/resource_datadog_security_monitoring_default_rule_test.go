package test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

const tfSecurityDefaultRuleName = "datadog_security_monitoring_default_rule.acceptance_test"

func TestAccDatadogSecurityMonitoringDefaultRule_Basic(t *testing.T) {
	t.Parallel()
	_, accProviders := testAccProviders(context.Background(), t)

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: accProviders,
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
    name_filter = "brute"
}
`
}

func testAccCheckDatadogSecurityMonitoringDefaultNoop() string {
	return fmt.Sprintf(`
data "datadog_security_monitoring_rules" "bruteforce" {
    name_filter = "brute"
}

resource "datadog_security_monitoring_default_rule" "acceptance_test" {
}
`)
}
