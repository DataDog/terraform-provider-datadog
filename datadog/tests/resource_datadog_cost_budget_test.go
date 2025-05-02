package test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/terraform-providers/terraform-provider-datadog/datadog"
)

func TestAccDatadogCostBudget_Basic(t *testing.T) {
	t.Parallel()
	ctx, accProviders := testAccProviders(context.Background(), t)
	budgetName := uniqueEntityName(ctx, t)
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		CheckDestroy:      testAccCheckDatadogCostBudgetDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogCostBudgetBasic(budgetName),
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
				),
			},
		},
	})
}

func TestAccDatadogCostBudget_WithTagFilters(t *testing.T) {
	t.Parallel()
	ctx, accProviders := testAccProviders(context.Background(), t)
	budgetName := uniqueEntityName(ctx, t)
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		CheckDestroy:      testAccCheckDatadogCostBudgetDestroy(accProvider),
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
						"datadog_cost_budget.foo", "entries.0.tag_filters.0.tag_key", "service"),
					resource.TestCheckResourceAttr(
						"datadog_cost_budget.foo", "entries.0.tag_filters.0.tag_value", "ec2"),
				),
			},
		},
	})
}

func testAccCheckDatadogCostBudgetDestroy(accProvider func() (*schema.Provider, error)) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		provider, _ := accProvider()
		providerConf := provider.Meta().(*datadog.ProviderConfiguration)
		apiInstances := providerConf.DatadogApiInstances
		auth := providerConf.Auth

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
      tag_key = "service"
      tag_value = "ec2"
    }
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
