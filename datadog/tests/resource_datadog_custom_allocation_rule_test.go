package test

import (
	"context"
	"fmt"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/fwprovider"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

func TestAccDatadogCustomAllocationRuleBasic(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	uniq := uniqueEntityName(ctx, t)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogCustomAllocationRuleDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogCustomAllocationRuleBasic(uniq),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogCustomAllocationRuleExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(
						"datadog_custom_allocation_rule.foo", "enabled", "true"),
					resource.TestCheckResourceAttr(
						"datadog_custom_allocation_rule.foo", "rule_name", fmt.Sprintf("tf-test-rule-%s", uniq)),
					resource.TestCheckResourceAttr(
						"datadog_custom_allocation_rule.foo", "providernames.0", "aws"),
					resource.TestCheckResourceAttr(
						"datadog_custom_allocation_rule.foo", "strategy.method", "even"),
					resource.TestCheckResourceAttrSet(
						"datadog_custom_allocation_rule.foo", "order_id"),
				),
			},
		},
	})
}

func TestAccDatadogCustomAllocationRuleUpdate(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	uniq := uniqueEntityName(ctx, t)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogCustomAllocationRuleDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogCustomAllocationRuleBasic(uniq),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogCustomAllocationRuleExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(
						"datadog_custom_allocation_rule.foo", "enabled", "true"),
					resource.TestCheckResourceAttr(
						"datadog_custom_allocation_rule.foo", "rule_name", fmt.Sprintf("tf-test-rule-%s", uniq)),
					resource.TestCheckResourceAttr(
						"datadog_custom_allocation_rule.foo", "costs_to_allocate.0.tag", "aws_product"),
					resource.TestCheckResourceAttr(
						"datadog_custom_allocation_rule.foo", "costs_to_allocate.0.value", "AmazonEC2"),
				),
			},
			{
				Config: testAccCheckDatadogCustomAllocationRuleUpdated(uniq),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogCustomAllocationRuleExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(
						"datadog_custom_allocation_rule.foo", "enabled", "false"),
					resource.TestCheckResourceAttr(
						"datadog_custom_allocation_rule.foo", "rule_name", fmt.Sprintf("tf-test-rule-updated-%s", uniq)),
					resource.TestCheckResourceAttr(
						"datadog_custom_allocation_rule.foo", "costs_to_allocate.0.tag", "aws_product"),
					resource.TestCheckResourceAttr(
						"datadog_custom_allocation_rule.foo", "costs_to_allocate.0.value", "AmazonS3"),
				),
			},
		},
	})
}

func TestAccDatadogCustomAllocationRuleImport(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	uniq := uniqueEntityName(ctx, t)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogCustomAllocationRuleDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogCustomAllocationRuleBasic(uniq),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogCustomAllocationRuleExists(providers.frameworkProvider),
				),
			},
			{
				ResourceName:            "datadog_custom_allocation_rule.foo",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"created", "updated"},
			},
		},
	})
}

func TestAccDatadogCustomAllocationRuleMultipleFilters(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	uniq := uniqueEntityName(ctx, t)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogCustomAllocationRuleDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogCustomAllocationRuleMultipleFilters(uniq),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogCustomAllocationRuleExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(
						"datadog_custom_allocation_rule.foo", "costs_to_allocate.#", "2"),
					resource.TestCheckResourceAttr(
						"datadog_custom_allocation_rule.foo", "strategy.based_on_costs.#", "2"),
				),
			},
		},
	})
}

func testAccCheckDatadogCustomAllocationRuleBasic(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_custom_allocation_rule" "foo" {
  costs_to_allocate {
    condition = "is"
    tag       = "aws_product"
    value     = "AmazonEC2"
  }
  enabled       = true
  providernames = ["aws"]
  rule_name     = "tf-test-rule-%s"
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
}`, uniq)
}

func testAccCheckDatadogCustomAllocationRuleUpdated(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_custom_allocation_rule" "foo" {
  costs_to_allocate {
    condition = "is"
    tag       = "aws_product"
    value     = "AmazonS3"
  }
  enabled       = false
  providernames = ["aws"]
  rule_name     = "tf-test-rule-updated-%s"
  strategy {
    allocated_by_tag_keys = ["team", "env"]
    based_on_costs {
      condition = "is"
      tag       = "aws_product"
      value     = "AmazonS3"
    }
    granularity = "daily"
    method      = "even"
  }
}`, uniq)
}

func testAccCheckDatadogCustomAllocationRuleMultipleFilters(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_custom_allocation_rule" "foo" {
  costs_to_allocate {
    condition = "is"
    tag       = "aws_product"
    value     = "AmazonEC2"
  }
  costs_to_allocate {
    condition = "is"
    tag       = "aws_product"
    value     = "AmazonS3"
  }
  enabled       = true
  providernames = ["aws"]
  rule_name     = "tf-test-rule-multiple-%s"
  strategy {
    allocated_by_tag_keys = ["team"]
    based_on_costs {
      condition = "is"
      tag       = "aws_product"
      value     = "AmazonEC2"
    }
    based_on_costs {
      condition = "is"
      tag       = "aws_product"
      value     = "AmazonS3"
    }
    granularity = "daily"
    method      = "even"
  }
}`, uniq)
}

func testAccCheckDatadogCustomAllocationRuleDestroy(accProvider *fwprovider.FrameworkProvider) func(*terraform.State) error {
	return func(s *terraform.State) error {
		apiInstances := accProvider.DatadogApiInstances
		auth := accProvider.Auth

		if err := DatadogCustomAllocationRuleDestroyHelper(auth, s, apiInstances); err != nil {
			return err
		}
		return nil
	}
}

func DatadogCustomAllocationRuleDestroyHelper(auth context.Context, s *terraform.State, apiInstances *utils.ApiInstances) error {
	err := utils.Retry(2, 10, func() error {
		for _, r := range s.RootModule().Resources {
			if r.Type != "datadog_custom_allocation_rule" {
				continue
			}

			ruleId, _ := strconv.ParseInt(r.Primary.ID, 10, 64)
			_, httpResp, err := apiInstances.GetCloudCostManagementApiV2().GetCustomAllocationRule(auth, ruleId)
			if err != nil {
				if httpResp != nil && httpResp.StatusCode == 404 {
					return nil
				}
				return &utils.RetryableError{Prob: fmt.Sprintf("received an error retrieving DatadogCustomAllocationRule %s", err)}
			}
			return &utils.RetryableError{Prob: "DatadogCustomAllocationRule still exists"}
		}
		return nil
	})
	return err
}

func testAccCheckDatadogCustomAllocationRuleExists(accProvider *fwprovider.FrameworkProvider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		apiInstances := accProvider.DatadogApiInstances
		auth := accProvider.Auth

		if err := datadogCustomAllocationRuleExistsHelper(auth, s, apiInstances); err != nil {
			return err
		}
		return nil
	}
}

func datadogCustomAllocationRuleExistsHelper(auth context.Context, s *terraform.State, apiInstances *utils.ApiInstances) error {
	for _, r := range s.RootModule().Resources {
		if r.Type != "datadog_custom_allocation_rule" {
			continue
		}

		ruleId, _ := strconv.ParseInt(r.Primary.ID, 10, 64)
		_, httpResp, err := apiInstances.GetCloudCostManagementApiV2().GetCustomAllocationRule(auth, ruleId)
		if err != nil {
			return utils.TranslateClientError(err, httpResp, "error retrieving DatadogCustomAllocationRule")
		}
	}
	return nil
}
