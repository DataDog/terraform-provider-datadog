package test

import (
	"context"
	"fmt"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
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

func TestAccDatadogCustomAllocationRuleInPlaceUpdate(t *testing.T) {
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
						"datadog_custom_allocation_rule.foo", "costs_to_allocate.0.value", "AmazonEC2"),
					resource.TestCheckResourceAttr(
						"datadog_custom_allocation_rule.foo", "version", "1"),
				),
			},
			{
				// Change only costs_to_allocate (not rule_name) so this is an in-place
				// PATCH, not a replacement. The API bumps `updated` on every update but
				// keeps `version` at 1, so the plan action is what asserts the update
				// happened in place.
				Config: testAccCheckDatadogCustomAllocationRuleInPlaceUpdate(uniq),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("datadog_custom_allocation_rule.foo", plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogCustomAllocationRuleExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(
						"datadog_custom_allocation_rule.foo", "rule_name", fmt.Sprintf("tf-test-rule-%s", uniq)),
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

// TestAccDatadogCustomAllocationRuleTimeseries covers the simple metrics query
// shape. Requires an Enterprise or Host-Based tier org: timeseries methods are
// gated and return 403 otherwise.
func TestAccDatadogCustomAllocationRuleTimeseries(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	uniq := uniqueEntityName(ctx, t)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogCustomAllocationRuleDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogCustomAllocationRuleTimeseries(uniq),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogCustomAllocationRuleExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(
						"datadog_custom_allocation_rule.foo", "strategy.method", "proportional_timeseries"),
					resource.TestCheckResourceAttr(
						"datadog_custom_allocation_rule.foo", "strategy.granularity", "daily"),
					resource.TestCheckResourceAttrSet(
						"datadog_custom_allocation_rule.foo", "strategy.based_on_timeseries"),
				),
			},
		},
	})
}

// TestAccDatadogCustomAllocationRuleTimeseriesAggregateQuery covers the
// aggregate_augmented_query shape: a compound query carrying nested base_query
// and augment_query objects plus compute, group_by and join_condition. This is
// the majority shape in production and the reason based_on_timeseries is modeled
// as opaque JSON rather than a typed schema, so it must round-trip exactly.
func TestAccDatadogCustomAllocationRuleTimeseriesAggregateQuery(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	uniq := uniqueEntityName(ctx, t)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogCustomAllocationRuleDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogCustomAllocationRuleTimeseriesAggregate(uniq),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogCustomAllocationRuleExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(
						"datadog_custom_allocation_rule.foo", "strategy.method", "proportional_timeseries"),
					resource.TestCheckResourceAttrSet(
						"datadog_custom_allocation_rule.foo", "strategy.based_on_timeseries"),
				),
			},
			{
				// Re-applying the identical config must produce no diff. Guards
				// against the API reordering, adding or dropping keys inside the
				// opaque payload.
				Config:   testAccCheckDatadogCustomAllocationRuleTimeseriesAggregate(uniq),
				PlanOnly: true,
			},
		},
	})
}

// TestAccDatadogCustomAllocationRuleTimeseriesPreservedOnRename changes only
// rule_name and asserts based_on_timeseries survives the update untouched. This
// is the regression test for the nil-payload write path.
func TestAccDatadogCustomAllocationRuleTimeseriesPreservedOnRename(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	uniq := uniqueEntityName(ctx, t)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogCustomAllocationRuleDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogCustomAllocationRuleTimeseries(uniq),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogCustomAllocationRuleExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(
						"datadog_custom_allocation_rule.foo", "rule_name", fmt.Sprintf("tf-test-rule-ts-%s", uniq)),
				),
			},
			{
				Config: testAccCheckDatadogCustomAllocationRuleTimeseriesRenamed(uniq),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogCustomAllocationRuleExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(
						"datadog_custom_allocation_rule.foo", "rule_name", fmt.Sprintf("tf-test-rule-ts-renamed-%s", uniq)),
					resource.TestCheckResourceAttrSet(
						"datadog_custom_allocation_rule.foo", "strategy.based_on_timeseries"),
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

func testAccCheckDatadogCustomAllocationRuleInPlaceUpdate(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_custom_allocation_rule" "foo" {
  costs_to_allocate {
    condition = "is"
    tag       = "aws_product"
    value     = "AmazonS3"
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

func testAccTimeseriesRuleConfig(ruleName string, basedOnTimeseries string) string {
	return fmt.Sprintf(`
resource "datadog_custom_allocation_rule" "foo" {
  costs_to_allocate {
    condition = "is"
    tag       = "aws_product"
    value     = "AmazonEC2"
  }
  enabled       = true
  providernames = ["aws"]
  rule_name     = "%s"
  strategy {
    granularity = "daily"
    method      = "proportional_timeseries"
%s
  }
}`, ruleName, basedOnTimeseries)
}

const testAccTimeseriesMetricsQuery = `
    based_on_timeseries = jsonencode({
      response_format = "timeseries"
      queries = [{
        name        = "query1"
        data_source = "metrics"
        query       = "avg:system.cpu.user{*} by {host}"
      }]
      formulas = [{ formula = "query1" }]
    })`

func testAccCheckDatadogCustomAllocationRuleTimeseries(uniq string) string {
	return testAccTimeseriesRuleConfig(fmt.Sprintf("tf-test-rule-ts-%s", uniq), testAccTimeseriesMetricsQuery)
}

func testAccCheckDatadogCustomAllocationRuleTimeseriesRenamed(uniq string) string {
	return testAccTimeseriesRuleConfig(fmt.Sprintf("tf-test-rule-ts-renamed-%s", uniq), testAccTimeseriesMetricsQuery)
}

func testAccCheckDatadogCustomAllocationRuleTimeseriesAggregate(uniq string) string {
	return testAccTimeseriesRuleConfig(fmt.Sprintf("tf-test-rule-ts-agg-%s", uniq), `
    based_on_timeseries = jsonencode({
      response_format = "timeseries"
      formulas        = [{ formula = "query1" }]
      queries = [{
        data_source = "aggregate_augmented_query"
        name        = "query1"
        base_query = {
          data_source = "metrics"
          name        = "query1"
          query       = "sum:system.cpu.user{*} by {dd.team}"
        }
        augment_query = {
          data_source = "reference_table"
          name        = "filter_query"
          table_name  = "tf_test_team_mapping"
          columns     = [{ name = "team" }, { name = "dd_team" }]
        }
        compute  = [{ aggregation = "sum", name = "compute_result" }]
        group_by = [{ facet = "team", source = "filter_query" }]
        join_condition = {
          join_type         = "inner"
          is_negated        = false
          base_attribute    = "dd.team"
          augment_attribute = "dd_team"
        }
      }]
    })`)
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
