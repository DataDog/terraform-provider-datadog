package test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/fwprovider"
)

func TestAccDatadogCustomAllocationRuleOrder_Basic(t *testing.T) {
	// Do not run in parallel - this resource manages global rule order
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	uniq := uniqueEntityName(ctx, t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogCustomAllocationRuleOrderDestroy(ctx, providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogCustomAllocationRuleOrderConfigBasic(uniq),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogCustomAllocationRuleOrderExists("datadog_custom_allocation_rules.foo"),
					resource.TestCheckResourceAttr("datadog_custom_allocation_rules.foo", "id", "order"),
					resource.TestCheckResourceAttrSet("datadog_custom_allocation_rules.foo", "rule_ids.#"),
					// Verify the 3 rules created by this test maintain their relative order
					func(s *terraform.State) error {
						ruleIDs, err := extractRuleIDsFromResources(s, []string{
							"datadog_custom_allocation_rule.first",
							"datadog_custom_allocation_rule.second",
							"datadog_custom_allocation_rule.third",
						})
						if err != nil {
							return err
						}
						return verifyRelativeOrder(ruleIDs)(s)
					},
				),
			},
		},
	})
}

func TestAccDatadogCustomAllocationRuleOrder_Update(t *testing.T) {
	// Do not run in parallel - this resource manages global rule order
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	uniq := uniqueEntityName(ctx, t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogCustomAllocationRuleOrderDestroy(ctx, providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogCustomAllocationRuleOrderConfigBasic(uniq),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogCustomAllocationRuleOrderExists("datadog_custom_allocation_rules.foo"),
					resource.TestCheckResourceAttrSet("datadog_custom_allocation_rules.foo", "rule_ids.#"),
					// Verify initial order: first, second, third
					func(s *terraform.State) error {
						ruleIDs, err := extractRuleIDsFromResources(s, []string{
							"datadog_custom_allocation_rule.first",
							"datadog_custom_allocation_rule.second",
							"datadog_custom_allocation_rule.third",
						})
						if err != nil {
							return err
						}
						return verifyRelativeOrder(ruleIDs)(s)
					},
				),
			},
			{
				Config: testAccCheckDatadogCustomAllocationRuleOrderConfigReordered(uniq),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogCustomAllocationRuleOrderExists("datadog_custom_allocation_rules.foo"),
					resource.TestCheckResourceAttrSet("datadog_custom_allocation_rules.foo", "rule_ids.#"),
					// Verify updated order: third, first, second
					func(s *terraform.State) error {
						ruleIDs, err := extractRuleIDsFromResources(s, []string{
							"datadog_custom_allocation_rule.third",
							"datadog_custom_allocation_rule.first",
							"datadog_custom_allocation_rule.second",
						})
						if err != nil {
							return err
						}
						return verifyRelativeOrder(ruleIDs)(s)
					},
				),
			},
		},
	})
}

func TestAccDatadogCustomAllocationRuleOrder_Import(t *testing.T) {
	// Do not run in parallel - this resource manages global rule order
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	uniq := uniqueEntityName(ctx, t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogCustomAllocationRuleOrderDestroy(ctx, providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogCustomAllocationRuleOrderConfigBasic(uniq),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogCustomAllocationRuleOrderExists("datadog_custom_allocation_rules.foo"),
					func(s *terraform.State) error {
						ruleIDs, err := extractRuleIDsFromResources(s, []string{
							"datadog_custom_allocation_rule.first",
							"datadog_custom_allocation_rule.second",
							"datadog_custom_allocation_rule.third",
						})
						if err != nil {
							return err
						}
						return verifyRelativeOrder(ruleIDs)(s)
					},
				),
			},
			{
				ResourceName:      "datadog_custom_allocation_rules.foo",
				ImportState:       true,
				ImportStateId:     "order",
				ImportStateVerify: false, // Cannot verify exact match - import reads all rules from backend
			},
		},
	})
}

func TestAccDatadogCustomAllocationRuleOrder_SingleRule(t *testing.T) {
	// Do not run in parallel - this resource manages global rule order
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	uniq := uniqueEntityName(ctx, t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogCustomAllocationRuleOrderDestroy(ctx, providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogCustomAllocationRuleOrderConfigSingle(uniq),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogCustomAllocationRuleOrderExists("datadog_custom_allocation_rules.foo"),
					resource.TestCheckResourceAttrSet("datadog_custom_allocation_rules.foo", "rule_ids.#"),
					// Verify the single rule is present in the order
					func(s *terraform.State) error {
						ruleIDs, err := extractRuleIDsFromResources(s, []string{
							"datadog_custom_allocation_rule.single",
						})
						if err != nil {
							return err
						}
						return verifyRelativeOrder(ruleIDs)(s)
					},
				),
			},
		},
	})
}

func testAccCheckDatadogCustomAllocationRuleOrderExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no ID is set for resource: %s", resourceName)
		}

		// The order resource should always have ID "order"
		if rs.Primary.ID != "order" {
			return fmt.Errorf("expected ID to be 'order', got: %s", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckDatadogCustomAllocationRuleOrderChanged(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not found: %s", resourceName)
		}

		// Just verify the resource exists - relative order is checked by other assertions
		return nil
	}
}

// verifyRelativeOrder checks that the specified rule IDs maintain their relative order in the state
func verifyRelativeOrder(expectedOrder []string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources["datadog_custom_allocation_rules.foo"]
		if !ok {
			return fmt.Errorf("resource not found")
		}

		// Get all rule IDs from state
		countStr := rs.Primary.Attributes["rule_ids.#"]
		count := 0
		fmt.Sscanf(countStr, "%d", &count)

		// Extract the actual rule IDs from state
		actualRuleIDs := make([]string, count)
		for i := 0; i < count; i++ {
			actualRuleIDs[i] = rs.Primary.Attributes[fmt.Sprintf("rule_ids.%d", i)]
		}

		// Find positions of expected rules in actual order
		positions := make(map[string]int)
		for i, ruleID := range actualRuleIDs {
			positions[ruleID] = i
		}

		// Verify all expected rules are present
		for _, expectedID := range expectedOrder {
			if _, found := positions[expectedID]; !found {
				return fmt.Errorf("expected rule ID %s not found in state", expectedID)
			}
		}

		// Verify relative order is maintained
		for i := 0; i < len(expectedOrder)-1; i++ {
			currentPos := positions[expectedOrder[i]]
			nextPos := positions[expectedOrder[i+1]]

			if currentPos >= nextPos {
				return fmt.Errorf("relative order violation: rule %s (pos %d) should come before rule %s (pos %d)",
					expectedOrder[i], currentPos, expectedOrder[i+1], nextPos)
			}
		}

		return nil
	}
}

// extractRuleIDsFromResources extracts rule IDs from the terraform state for the given resource names
func extractRuleIDsFromResources(s *terraform.State, resourceNames []string) ([]string, error) {
	ruleIDs := make([]string, len(resourceNames))
	for i, name := range resourceNames {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return nil, fmt.Errorf("resource %s not found", name)
		}
		ruleIDs[i] = rs.Primary.ID
	}
	return ruleIDs, nil
}

func testAccCheckDatadogCustomAllocationRuleOrderDestroy(ctx context.Context, frameworkProvider *fwprovider.FrameworkProvider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		// Since the order resource is a management resource that doesn't actually
		// delete anything when destroyed (it's a no-op), we don't need to check
		// for destruction. The rules themselves should still exist.
		// This follows the same pattern as other order resources in the provider.
		return nil
	}
}

// Test configurations

func testAccCheckDatadogCustomAllocationRuleOrderConfigBasic(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_custom_allocation_rule" "first" {
  costs_to_allocate {
    condition = "is"
    tag       = "aws_product"
    value     = "AmazonEC2"
  }
  enabled       = true
  providernames = ["aws"]
  rule_name     = "tf-test-order-first-%s"
  strategy {
    allocated_by_tag_keys = ["team"]
    based_on_costs {
      condition = "is"
      tag       = "aws_product"
      value     = "AmazonEC2"
    }
    granularity = "daily"
    method      = "even"
  }
}

resource "datadog_custom_allocation_rule" "second" {
  costs_to_allocate {
    condition = "is"
    tag       = "aws_product"
    value     = "AmazonS3"
  }
  enabled       = true
  providernames = ["aws"]
  rule_name     = "tf-test-order-second-%s"
  strategy {
    allocated_by_tag_keys = ["team"]
    based_on_costs {
      condition = "is"
      tag       = "aws_product"
      value     = "AmazonS3"
    }
    granularity = "daily"
    method      = "even"
  }
}

resource "datadog_custom_allocation_rule" "third" {
  costs_to_allocate {
    condition = "is"
    tag       = "aws_product"
    value     = "AmazonRDS"
  }
  enabled       = true
  providernames = ["aws"]
  rule_name     = "tf-test-order-third-%s"
  strategy {
    allocated_by_tag_keys = ["team"]
    based_on_costs {
      condition = "is"
      tag       = "aws_product"
      value     = "AmazonRDS"
    }
    granularity = "daily"
    method      = "even"
  }
}

resource "datadog_custom_allocation_rules" "foo" {
  rule_ids = [
    datadog_custom_allocation_rule.first.id,
    datadog_custom_allocation_rule.second.id,
    datadog_custom_allocation_rule.third.id
  ]
}`, uniq, uniq, uniq)
}

func testAccCheckDatadogCustomAllocationRuleOrderConfigReordered(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_custom_allocation_rule" "first" {
  costs_to_allocate {
    condition = "is"
    tag       = "aws_product"
    value     = "AmazonEC2"
  }
  enabled       = true
  providernames = ["aws"]
  rule_name     = "tf-test-order-first-%s"
  strategy {
    allocated_by_tag_keys = ["team"]
    based_on_costs {
      condition = "is"
      tag       = "aws_product"
      value     = "AmazonEC2"
    }
    granularity = "daily"
    method      = "even"
  }
}

resource "datadog_custom_allocation_rule" "second" {
  costs_to_allocate {
    condition = "is"
    tag       = "aws_product"
    value     = "AmazonS3"
  }
  enabled       = true
  providernames = ["aws"]
  rule_name     = "tf-test-order-second-%s"
  strategy {
    allocated_by_tag_keys = ["team"]
    based_on_costs {
      condition = "is"
      tag       = "aws_product"
      value     = "AmazonS3"
    }
    granularity = "daily"
    method      = "even"
  }
}

resource "datadog_custom_allocation_rule" "third" {
  costs_to_allocate {
    condition = "is"
    tag       = "aws_product"
    value     = "AmazonRDS"
  }
  enabled       = true
  providernames = ["aws"]
  rule_name     = "tf-test-order-third-%s"
  strategy {
    allocated_by_tag_keys = ["team"]
    based_on_costs {
      condition = "is"
      tag       = "aws_product"
      value     = "AmazonRDS"
    }
    granularity = "daily"
    method      = "even"
  }
}

resource "datadog_custom_allocation_rules" "foo" {
  rule_ids = [
    datadog_custom_allocation_rule.third.id,
    datadog_custom_allocation_rule.first.id,
    datadog_custom_allocation_rule.second.id
  ]
}`, uniq, uniq, uniq)
}

func testAccCheckDatadogCustomAllocationRuleOrderConfigSingle(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_custom_allocation_rule" "single" {
  costs_to_allocate {
    condition = "is"
    tag       = "aws_product"
    value     = "AmazonEC2"
  }
  enabled       = true
  providernames = ["aws"]
  rule_name     = "tf-test-order-single-%s"
  strategy {
    allocated_by_tag_keys = ["team"]
    based_on_costs {
      condition = "is"
      tag       = "aws_product"
      value     = "AmazonEC2"
    }
    granularity = "daily"
    method      = "even"
  }
}

resource "datadog_custom_allocation_rules" "foo" {
  rule_ids = [
    datadog_custom_allocation_rule.single.id
  ]
}`, uniq)
}
