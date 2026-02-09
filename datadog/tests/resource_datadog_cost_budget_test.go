package test

import (
	"context"
	"fmt"
	"regexp"
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
		ProtoV6ProviderFactories: accProviders,
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
		ProtoV6ProviderFactories: accProviders,
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
		ProtoV6ProviderFactories: accProviders,
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

// TestAccDatadogCostBudget_BudgetLine tests CRUD operations with the new budget_line schema
func TestAccDatadogCostBudget_BudgetLine(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	budgetName := uniqueEntityName(ctx, t)
	budgetNameUpdated := budgetName + " Updated"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogCostBudgetDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogCostBudgetBudgetLine(budgetName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"datadog_cost_budget.foo", "name", budgetName),
					resource.TestCheckResourceAttrSet(
						"datadog_cost_budget.foo", "id"),
					resource.TestCheckResourceAttr(
						"datadog_cost_budget.foo", "budget_line.#", "2"),
				),
			},
			{
				Config: testAccCheckDatadogCostBudgetBudgetLineUpdated(budgetNameUpdated),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"datadog_cost_budget.foo", "name", budgetNameUpdated),
					resource.TestCheckResourceAttr(
						"datadog_cost_budget.foo", "budget_line.#", "2"),
				),
			},
		},
	})
}

// TestAccDatadogCostBudget_BudgetLineHierarchical tests hierarchical budgets with budget_line schema
func TestAccDatadogCostBudget_BudgetLineHierarchical(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	budgetName := uniqueEntityName(ctx, t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogCostBudgetDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogCostBudgetBudgetLineHierarchical(budgetName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"datadog_cost_budget.foo", "name", budgetName),
					resource.TestCheckResourceAttrSet(
						"datadog_cost_budget.foo", "id"),
					resource.TestCheckResourceAttr(
						"datadog_cost_budget.foo", "budget_line.#", "2"),
				),
			},
		},
	})
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

func testAccCheckDatadogCostBudgetBudgetLine(uniq string) string {
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
}`, uniq)
}

func testAccCheckDatadogCostBudgetBudgetLineUpdated(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_cost_budget" "foo" {
  name = "%s"
  metrics_query = "sum:aws.cost.amortized{*} by {environment}"
  start_month = 202601
  end_month = 202603

  budget_line {
    amounts = {
      "202601" = 2000
      "202602" = 2100
      "202603" = 2200
    }
    tag_filters {
      tag_key = "environment"
      tag_value = "production"
    }
  }

  budget_line {
    amounts = {
      "202601" = 1000
      "202602" = 1050
      "202603" = 1100
    }
    tag_filters {
      tag_key = "environment"
      tag_value = "staging"
    }
  }
}`, uniq)
}

func testAccCheckDatadogCostBudgetBudgetLineHierarchical(uniq string) string {
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
}`, uniq)
}

// TestAccDatadogCostBudget_Validation verifies client-side validation catches errors during terraform plan
func TestAccDatadogCostBudget_Validation(t *testing.T) {
	t.Parallel()
	_, _, accProviders := testAccFrameworkMuxProviders(context.Background(), t)

	// Test cases for ValidateConfig validation order
	testCases := []struct {
		name        string
		config      string
		expectError string
	}{
		// Validate tags length
		{
			name: "TooManyTags",
			config: `
resource "datadog_cost_budget" "foo" {
  name = "test-too-many-tags"
  metrics_query = "sum:aws.cost.amortized{*} by {team,account,region}"
  start_month = 202501
  end_month = 202501
  entries {
    amount = 1000
    month = 202501
    tag_filters {
      tag_key = "team"
      tag_value = "backend"
    }
    tag_filters {
      tag_key = "account"
      tag_value = "staging"
    }
    tag_filters {
      tag_key = "region"
      tag_value = "us-east-1"
    }
  }
}`,
			expectError: `tags\s+must have 0, 1 or 2 elements`,
		},
		// Validate tags are unique
		{
			name: "DuplicateTags",
			config: `
resource "datadog_cost_budget" "foo" {
  name = "test-duplicate-tags"
  metrics_query = "sum:aws.cost.amortized{*} by {team,team}"
  start_month = 202501
  end_month = 202501
  entries {
    amount = 1000
    month = 202501
    tag_filters {
      tag_key = "team"
      tag_value = "backend"
    }
    tag_filters {
      tag_key = "team"
      tag_value = "frontend"
    }
  }
}`,
			expectError: `tags\s+must be unique`,
		},
		// Validate start_month > 0
		{
			name: "InvalidStartMonth",
			config: `
resource "datadog_cost_budget" "foo" {
  name = "test-invalid-start-month"
  metrics_query = "sum:aws.cost.amortized{*}"
  start_month = 0
  end_month = 202501
  entries {
    amount = 1000
    month = 202501
  }
}`,
			expectError: `start_month\s+must be greater than 0`,
		},
		// Validate end_month > 0
		{
			name: "InvalidEndMonth",
			config: `
resource "datadog_cost_budget" "foo" {
  name = "test-invalid-end-month"
  metrics_query = "sum:aws.cost.amortized{*}"
  start_month = 202501
  end_month = 0
  entries {
    amount = 1000
    month = 202501
  }
}`,
			expectError: `end_month\s+must be greater than 0`,
		},
		// Validate end_month >= start_month
		{
			name: "EndMonthBeforeStartMonth",
			config: `
resource "datadog_cost_budget" "foo" {
  name = "test-end-before-start"
  metrics_query = "sum:aws.cost.amortized{*}"
  start_month = 202512
  end_month = 202501
  entries {
    amount = 1000
    month = 202501
  }
}`,
			expectError: `end_month\s+must be greater than or equal to start_month`,
		},
		// Validate entry month in range
		{
			name: "MonthOutOfRange",
			config: `
resource "datadog_cost_budget" "foo" {
  name = "test-month-out-of-range"
  metrics_query = "sum:aws.cost.amortized{*}"
  start_month = 202501
  end_month = 202503
  entries {
    amount = 1000
    month = 202506
  }
}`,
			expectError: `entry\s+month must be between start_month and end_month`,
		},
		// Validate entry amount >= 0
		{
			name: "NegativeAmount",
			config: `
resource "datadog_cost_budget" "foo" {
  name = "test-negative-amount"
  metrics_query = "sum:aws.cost.amortized{*}"
  start_month = 202501
  end_month = 202501
  entries {
    amount = -100
    month = 202501
  }
}`,
			expectError: `entry\s+amount must be greater than or equal to 0`,
		},
		// Validate tag_filters count matches tags
		{
			name: "WrongTagCount",
			config: `
resource "datadog_cost_budget" "foo" {
  name = "test-wrong-tag-count"
  metrics_query = "sum:aws.cost.amortized{*} by {team,account}"
  start_month = 202501
  end_month = 202501
  entries {
    amount = 1000
    month = 202501
    tag_filters {
      tag_key = "team"
      tag_value = "backend"
    }
  }
}`,
			expectError: `entry\s+tag_filters must include all group by tags`,
		},
		// Validate tag_key is in metrics_query tags
		{
			name: "InvalidTagKey",
			config: `
resource "datadog_cost_budget" "foo" {
  name = "test-invalid-tag-key"
  metrics_query = "sum:aws.cost.amortized{*} by {account}"
  start_month = 202501
  end_month = 202501
  entries {
    amount = 1000
    month = 202501
    tag_filters {
      tag_key = "wrong_tag"
      tag_value = "value"
    }
  }
}`,
			expectError: `tag_key\s+must be one of the values inside the tags array`,
		},
		// Validate entries or budget_line exist
		{
			name: "NoEntries",
			config: `
resource "datadog_cost_budget" "foo" {
  name = "test-no-entries"
  metrics_query = "sum:aws.cost.amortized{*}"
  start_month = 202501
  end_month = 202501
}`,
			expectError: `Either 'entries' or 'budget_line' must be specified`,
		},
		// Validate all tag combinations have entries for all months
		{
			name: "MissingMonths",
			config: `
resource "datadog_cost_budget" "foo" {
  name = "test-missing-months"
  metrics_query = "sum:aws.cost.amortized{*}"
  start_month = 202501
  end_month = 202503
  entries {
    amount = 1000
    month = 202501
  }
}`,
			expectError: `missing\s+entries for one or more tag combinations`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resource.Test(t, resource.TestCase{
				PreCheck:                 func() { testAccPreCheck(t) },
				ProtoV6ProviderFactories: accProviders,
				Steps: []resource.TestStep{
					{
						Config:      tc.config,
						ExpectError: regexp.MustCompile(tc.expectError),
					},
				},
			})
		})
	}
}
