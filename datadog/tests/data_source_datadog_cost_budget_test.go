package test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDatadogCostBudgetDataSource_Basic(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	budgetName := uniqueEntityName(ctx, t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogCostBudgetDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogCostBudgetDataSourceConfig(budgetName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.datadog_cost_budget.foo", "name", budgetName),
					resource.TestCheckResourceAttr(
						"data.datadog_cost_budget.foo", "metrics_query", "sum:aws.cost.amortized{*} by {account}"),
					resource.TestCheckResourceAttr(
						"data.datadog_cost_budget.foo", "start_month", "202401"),
					resource.TestCheckResourceAttr(
						"data.datadog_cost_budget.foo", "end_month", "202412"),
					resource.TestCheckResourceAttr(
						"data.datadog_cost_budget.foo", "entries.#", "12"),
				),
			},
		},
	})
}

func testAccCheckDatadogCostBudgetDataSourceConfig(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_cost_budget" "foo" {
  name = "%s"
  metrics_query = "sum:aws.cost.amortized{*} by {account}"
  start_month = 202401
  end_month = 202412
  entries {
    amount = 1000
    month = 202401
  }
  entries {
    amount = 1000
    month = 202402
  }
  entries {
    amount = 1000
    month = 202403
  }
  entries {
    amount = 1000
    month = 202404
  }
  entries {
    amount = 1000
    month = 202405
  }
  entries {
    amount = 1000
    month = 202406
  }
  entries {
    amount = 1000
    month = 202407
  }
  entries {
    amount = 1000
    month = 202408
  }
  entries {
    amount = 1000
    month = 202409
  }
  entries {
    amount = 1000
    month = 202410
  }
  entries {
    amount = 1000
    month = 202411
  }
  entries {
    amount = 1000
    month = 202412
  }
}

data "datadog_cost_budget" "foo" {
  id = datadog_cost_budget.foo.id
}`, uniq)
}
