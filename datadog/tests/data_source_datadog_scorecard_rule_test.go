package test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDatadogScorecardRuleDataSource_Basic(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	uniq := uniqueEntityName(ctx, t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogScorecardRuleDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccDatadogScorecardRuleDataSourceConfig(uniq),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.datadog_scorecard_rule.test", "name", uniq),
					resource.TestCheckResourceAttr("data.datadog_scorecard_rule.test", "scorecard_name", uniq+"-scorecard"),
					resource.TestCheckResourceAttr("data.datadog_scorecard_rule.test", "description", "Test rule description"),
					resource.TestCheckResourceAttr("data.datadog_scorecard_rule.test", "enabled", "true"),
					resource.TestCheckResourceAttr("data.datadog_scorecard_rule.test", "level", "1"),
				),
			},
		},
	})
}

func testAccDatadogScorecardRuleDataSourceConfig(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_scorecard_rule" "foo" {
  name           = "%[1]s"
  scorecard_name = "%[1]s-scorecard"
  description    = "Test rule description"
  enabled        = true
  level          = "1"
  owner          = "test-team"
}

data "datadog_scorecard_rule" "test" {
  id = datadog_scorecard_rule.foo.id
}
`, uniq)
}
