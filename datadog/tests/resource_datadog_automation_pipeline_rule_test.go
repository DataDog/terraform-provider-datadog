package test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// config
func testAccCheckDatadogAutomationPipelineRuleConfig(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_automation_pipeline_rule" "foo" {
  name = "%s"
}`, uniq)
}

// tests

func TestAccDatadogAutomationPipelineRule(t *testing.T) {
	t.Parallel()
	ctx, _, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	sloName := uniqueEntityName(ctx, t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: accProviders,
		// CheckDestroy:      testAccCheckDatadogServiceLevelObjectiveDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogAutomationPipelineRuleConfig(sloName),
				Check: resource.ComposeTestCheckFunc(
					// testAccCheckDatadogServiceLevelObjectiveExists(accProvider, "datadog_service_level_objective.foo"),
					resource.TestCheckResourceAttr(
						"datadog_automation_pipeline_rule.foo", "name", sloName),
				),
			},
		},
	})
}
