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
		ProtoV6ProviderFactories: accProviders,
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
	tag_filters {
	  tag_key = "account"
	  tag_value = "foo"
	}
  }
  entries {
    amount = 1000
    month = 202402
	tag_filters {
	  tag_key = "account"
	  tag_value = "foo"
	}
  }
  entries {
    amount = 1000
    month = 202403
	tag_filters {
	  tag_key = "account"
	  tag_value = "foo"
	}
  }
  entries {
    amount = 1000
    month = 202404
	tag_filters {
	  tag_key = "account"
	  tag_value = "foo"
	}
  }
  entries {
    amount = 1000
    month = 202405
	tag_filters {
	  tag_key = "account"
	  tag_value = "foo"
	}
  }
  entries {
    amount = 1000
    month = 202406
	tag_filters {
	  tag_key = "account"
	  tag_value = "foo"
	}
  }
  entries {
    amount = 1000
    month = 202407
	tag_filters {
	  tag_key = "account"
	  tag_value = "foo"
	}
  }
  entries {
    amount = 1000
    month = 202408
	tag_filters {
	  tag_key = "account"
	  tag_value = "foo"
	}
  }
  entries {
    amount = 1000
    month = 202409
	tag_filters {
	  tag_key = "account"
	  tag_value = "foo"
	}
  }
  entries {
    amount = 1000
    month = 202410
	tag_filters {
	  tag_key = "account"
	  tag_value = "foo"
	}
  }
  entries {
    amount = 1000
    month = 202411
	tag_filters {
	  tag_key = "account"
	  tag_value = "foo"
	}
  }
  entries {
    amount = 1000
    month = 202412
	tag_filters {
	  tag_key = "account"
	  tag_value = "foo"
	}
  }
}

data "datadog_cost_budget" "foo" {
  id = datadog_cost_budget.foo.id
}`, uniq)
}
