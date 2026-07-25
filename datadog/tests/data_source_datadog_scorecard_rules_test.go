package test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDatadogScorecardRulesDataSource_FilterName(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	uniq := uniqueEntityName(ctx, t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogScorecardRuleDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccDatadogScorecardRulesDataSourceConfig(uniq),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.datadog_scorecard_rules.filtered", "rules.#", "1"),
					resource.TestCheckResourceAttr("data.datadog_scorecard_rules.filtered", "rules.0.name", uniq+"-matching"),
					resource.TestCheckResourceAttr("data.datadog_scorecard_rules.filtered", "rules.0.level", "1"),
				),
			},
		},
	})
}

func testAccDatadogScorecardRulesDataSourceConfig(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_scorecard_rule" "matching" {
  name           = "%[1]s-matching"
  scorecard_name = "%[1]s-scorecard"
  level          = "1"
}

resource "datadog_scorecard_rule" "not_matching" {
  name           = "%[1]s-other"
  scorecard_name = "%[1]s-scorecard"
  level          = "2"
}

data "datadog_scorecard_rules" "filtered" {
  filter_name = "%[1]s-matching"
  depends_on = [
    datadog_scorecard_rule.matching,
    datadog_scorecard_rule.not_matching,
  ]
}
`, uniq)
}
