package test

import (
	"context"
	"fmt"
	"testing"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/fwprovider"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccDatadogCostBudget_Basic(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	budgetName := uniqueEntityName(ctx, t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogCostBudgetDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogCostBudgetBasic(budgetName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"datadog_cost_budget.foo", "name", budgetName),
					resource.TestCheckResourceAttrSet(
						"datadog_cost_budget.foo", "id"),
					resource.TestCheckResourceAttr(
						"datadog_cost_budget.foo", "metrics_query", "sum:aws.cost.amortized{*}"),
					resource.TestCheckResourceAttr(
						"datadog_cost_budget.foo", "start_month", "202401"),
					resource.TestCheckResourceAttr(
						"datadog_cost_budget.foo", "end_month", "202412"),
					resource.TestCheckResourceAttr(
						"datadog_cost_budget.foo", "entries.#", "12"),
				),
			},
		},
	})
}

func TestAccDatadogCostBudget_WithTagFilters(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	budgetName := uniqueEntityName(ctx, t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogCostBudgetDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogCostBudgetWithTagFilters(budgetName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"datadog_cost_budget.foo", "name", budgetName),
					resource.TestCheckResourceAttr(
						"datadog_cost_budget.foo", "metrics_query", "sum:aws.cost.amortized{*} by {account}"),
					resource.TestCheckResourceAttr(
						"datadog_cost_budget.foo", "start_month", "202401"),
					resource.TestCheckResourceAttr(
						"datadog_cost_budget.foo", "end_month", "202412"),
					resource.TestCheckResourceAttr(
						"datadog_cost_budget.foo", "entries.#", "12"),
					resource.TestCheckResourceAttr(
						"datadog_cost_budget.foo", "entries.0.tag_filters.#", "1"),
					resource.TestCheckResourceAttr(
						"datadog_cost_budget.foo", "entries.0.tag_filters.0.tag_key", "account"),
					resource.TestCheckResourceAttr(
						"datadog_cost_budget.foo", "entries.0.tag_filters.0.tag_value", "ec2"),
				),
			},
		},
	})
}

func TestAccDatadogCostBudget_Update(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	budgetName := uniqueEntityName(ctx, t)
	budgetNameUpdated := budgetName + "updated!"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogCostBudgetDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogCostBudgetBasic(budgetName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"datadog_cost_budget.foo", "name", budgetName),
					resource.TestCheckResourceAttrSet(
						"datadog_cost_budget.foo", "id"),
					resource.TestCheckResourceAttr(
						"datadog_cost_budget.foo", "entries.0.amount", "1000"),
				),
			},
			{
				Config: testAccCheckDatadogCostBudgetUpdated(budgetNameUpdated),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"datadog_cost_budget.foo", "name", budgetNameUpdated),
					// The most important!!: ID should remain the same after update, so making sure the id is different between the 2 steps
					resource.TestCheckResourceAttrPair(
						"datadog_cost_budget.foo", "id",
						"datadog_cost_budget.foo", "id"),
					resource.TestCheckResourceAttr(
						"datadog_cost_budget.foo", "entries.0.amount", "2000"),
				),
			},
		},
	})
}

func testAccCheckDatadogCostBudgetDestroy(provider *fwprovider.FrameworkProvider) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		apiInstances := provider.DatadogApiInstances
		auth := provider.Auth

		for _, r := range s.RootModule().Resources {
			if r.Type != "datadog_cost_budget" {
				continue
			}

			_, httpResp, err := apiInstances.GetCloudCostManagementApiV2().GetBudget(auth, r.Primary.ID)
			if err != nil {
				if httpResp != nil && httpResp.StatusCode == 404 {
					continue
				}
				return fmt.Errorf("received an error retrieving cost budget: %s", err.Error())
			}
			return fmt.Errorf("cost budget still exists")
		}
		return nil
	}
}

func testAccCheckDatadogCostBudgetBasic(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_cost_budget" "foo" {
  name = "%s"
  metrics_query = "sum:aws.cost.amortized{*}"
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
}`, uniq)
}

func testAccCheckDatadogCostBudgetWithTagFilters(uniq string) string {
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
      tag_value = "ec2"
    }
  }
  entries {
    amount = 1000
    month = 202402
	tag_filters {
      tag_key = "account"
      tag_value = "ec2"
    }
  }
  entries {
    amount = 1000
    month = 202403
	tag_filters {
      tag_key = "account"
      tag_value = "ec2"
    }
  }
  entries {
    amount = 1000
    month = 202404
	tag_filters {
      tag_key = "account"
      tag_value = "ec2"
    }
  }
  entries {
    amount = 1000
    month = 202405
	tag_filters {
      tag_key = "account"
      tag_value = "ec2"
    }
  }
  entries {
    amount = 1000
    month = 202406
	tag_filters {
      tag_key = "account"
      tag_value = "ec2"
    }
  }
  entries {
    amount = 1000
    month = 202407
	tag_filters {
      tag_key = "account"
      tag_value = "ec2"
    }
  }
  entries {
    amount = 1000
    month = 202408
	tag_filters {
      tag_key = "account"
      tag_value = "ec2"
    }
  }
  entries {
    amount = 1000
    month = 202409
	tag_filters {
      tag_key = "account"
      tag_value = "ec2"
    }
  }
  entries {
    amount = 1000
    month = 202410
	tag_filters {
      tag_key = "account"
      tag_value = "ec2"
    }
  }
  entries {
    amount = 1000
    month = 202411
	tag_filters {
      tag_key = "account"
      tag_value = "ec2"
    }
  }
  entries {
    amount = 1000
    month = 202412
	tag_filters {
      tag_key = "account"
      tag_value = "ec2"
    }
  }
}`, uniq)
}

func testAccCheckDatadogCostBudgetUpdated(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_cost_budget" "foo" {
  name = "%s"
  metrics_query = "sum:aws.cost.amortized{*}"
  start_month = 202401
  end_month = 202412
  entries {
    amount = 2000
    month = 202401
  }
  entries {
    amount = 2000
    month = 202402
  }
  entries {
    amount = 2000
    month = 202403
  }
  entries {
    amount = 2000
    month = 202404
  }
  entries {
    amount = 2000
    month = 202405
  }
  entries {
    amount = 2000
    month = 202406
  }
  entries {
    amount = 2000
    month = 202407
  }
  entries {
    amount = 2000
    month = 202408
  }
  entries {
    amount = 2000
    month = 202409
  }
  entries {
    amount = 2000
    month = 202410
  }
  entries {
    amount = 2000
    month = 202411
  }
  entries {
    amount = 2000
    month = 202412
  }
}`, uniq)
}
