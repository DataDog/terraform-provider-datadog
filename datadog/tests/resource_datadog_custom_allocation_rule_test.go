package test

import (
	"context"
	"fmt"
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
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogCustomAllocationRuleDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogCustomAllocationRule(uniq),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogCustomAllocationRuleExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(
						"datadog_custom_allocation_rule.foo", "enabled", "UPDATE ME"),
					resource.TestCheckResourceAttrSet(
						"datadog_custom_allocation_rule.foo", "order_id"),
					resource.TestCheckResourceAttrSet(
						"datadog_custom_allocation_rule.foo", "rejected"),
					resource.TestCheckResourceAttr(
						"datadog_custom_allocation_rule.foo", "rule_name", "UPDATE ME"),
					resource.TestCheckResourceAttr(
						"datadog_custom_allocation_rule.foo", "type", "UPDATE ME"),
				),
			},
		},
	})
}

func testAccCheckDatadogCustomAllocationRule(uniq string) string {
	// Update me to make use of the unique value
	return fmt.Sprintf(`resource "datadog_custom_allocation_rule" "foo" {
    costs_to_allocate {
    condition = "UPDATE ME"
    tag = "UPDATE ME"
    value = "UPDATE ME"
    values = "UPDATE ME"
    }
    enabled = "UPDATE ME"
    providernames = [""]
    rule_name = "UPDATE ME"
    strategy {
    allocated_by {
    allocated_tags {
    key = "UPDATE ME"
    value = "UPDATE ME"
    }
    percentage = "UPDATE ME"
    }
    allocated_by_filters {
    condition = "UPDATE ME"
    tag = "UPDATE ME"
    value = "UPDATE ME"
    values = "UPDATE ME"
    }
    allocated_by_tag_keys = "UPDATE ME"
    based_on_costs {
    condition = "UPDATE ME"
    tag = "UPDATE ME"
    value = "UPDATE ME"
    values = "UPDATE ME"
    }
    based_on_timeseries {
    }
    evaluate_grouped_by_filters {
    condition = "UPDATE ME"
    tag = "UPDATE ME"
    value = "UPDATE ME"
    values = "UPDATE ME"
    }
    evaluate_grouped_by_tag_keys = "UPDATE ME"
    granularity = "UPDATE ME"
    method = "UPDATE ME"
    }
    type = "UPDATE ME"
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
			if r.Type != "resource_datadog_custom_allocation_rule" {
				continue
			}
			id := r.Primary.ID

			_, httpResp, err := apiInstances.GetCloudCostManagementApiV2().GetArbitraryCostRule(auth, id)
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
		if r.Type != "resource_datadog_custom_allocation_rule" {
			continue
		}
		id := r.Primary.ID

		_, httpResp, err := apiInstances.GetCloudCostManagementApiV2().GetArbitraryCostRule(auth, id)
		if err != nil {
			return utils.TranslateClientError(err, httpResp, "error retrieving DatadogCustomAllocationRule")
		}
	}
	return nil
}
