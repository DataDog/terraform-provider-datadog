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
						"data.datadog_cost_budget.foo", "metrics_query", "sum:aws.cost.amortized{*} by {environment}"),
					resource.TestCheckResourceAttr(
						"data.datadog_cost_budget.foo", "start_month", "202601"),
					resource.TestCheckResourceAttr(
						"data.datadog_cost_budget.foo", "end_month", "202603"),
					resource.TestCheckResourceAttr(
						"data.datadog_cost_budget.foo", "entries.#", "6"),
					resource.TestCheckResourceAttr(
						"data.datadog_cost_budget.foo", "budget_line.#", "2"),
				),
			},
		},
	})
}

func TestAccDatadogCostBudgetDataSource_Hierarchical(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	budgetName := uniqueEntityName(ctx, t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogCostBudgetDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogCostBudgetDataSourceHierarchicalConfig(budgetName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.datadog_cost_budget.foo", "name", budgetName),
					resource.TestCheckResourceAttr(
						"data.datadog_cost_budget.foo", "metrics_query", "sum:aws.cost.amortized{*} by {team,environment}"),
					resource.TestCheckResourceAttr(
						"data.datadog_cost_budget.foo", "start_month", "202601"),
					resource.TestCheckResourceAttr(
						"data.datadog_cost_budget.foo", "end_month", "202602"),
					resource.TestCheckResourceAttr(
						"data.datadog_cost_budget.foo", "entries.#", "4"),
					resource.TestCheckResourceAttr(
						"data.datadog_cost_budget.foo", "budget_line.#", "2"),
				),
			},
		},
	})
}

func testAccCheckDatadogCostBudgetDataSourceConfig(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_cost_budget" "foo" {
  name = "%s"
  metrics_query = "sum:aws.cost.amortized{*} by {environment}"
  start_month = 202601
  end_month = 202603

  budget_line {
    amounts = {
      "202601" = 1000
      "202602" = 1100
      "202603" = 1200
    }
    tag_filters {
      tag_key = "environment"
      tag_value = "production"
    }
  }

  budget_line {
    amounts = {
      "202601" = 500
      "202602" = 550
      "202603" = 600
    }
    tag_filters {
      tag_key = "environment"
      tag_value = "staging"
    }
  }
}

data "datadog_cost_budget" "foo" {
  id = datadog_cost_budget.foo.id
}`, uniq)
}

func testAccCheckDatadogCostBudgetDataSourceHierarchicalConfig(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_cost_budget" "foo" {
  name = "%s"
  metrics_query = "sum:aws.cost.amortized{*} by {team,environment}"
  start_month = 202601
  end_month = 202602

  budget_line {
    amounts = {
      "202601" = 1000
      "202602" = 1100
    }
    parent_tag_filters {
      tag_key = "team"
      tag_value = "backend"
    }
    child_tag_filters {
      tag_key = "environment"
      tag_value = "production"
    }
  }

  budget_line {
    amounts = {
      "202601" = 500
      "202602" = 550
    }
    parent_tag_filters {
      tag_key = "team"
      tag_value = "frontend"
    }
    child_tag_filters {
      tag_key = "environment"
      tag_value = "staging"
    }
  }
}

data "datadog_cost_budget" "foo" {
  id = datadog_cost_budget.foo.id
}`, uniq)
}
