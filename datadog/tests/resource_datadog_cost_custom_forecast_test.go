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

func TestAccDatadogCostCustomForecast_Basic(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	budgetName := uniqueEntityName(ctx, t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogCostCustomForecastDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogCostCustomForecastBasic(budgetName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						"datadog_cost_custom_forecast.foo", "id"),
					resource.TestCheckResourceAttrPair(
						"datadog_cost_custom_forecast.foo", "budget_uid",
						"datadog_cost_budget.foo", "id"),
					resource.TestCheckResourceAttr(
						"datadog_cost_custom_forecast.foo", "entries.#", "1"),
				),
			},
			{
				ResourceName: "datadog_cost_custom_forecast.foo",
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					rs := s.RootModule().Resources["datadog_cost_custom_forecast.foo"]
					return rs.Primary.Attributes["budget_uid"], nil
				},
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccDatadogCostCustomForecast_Update(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	budgetName := uniqueEntityName(ctx, t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogCostCustomForecastDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogCostCustomForecastBasic(budgetName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"datadog_cost_custom_forecast.foo", "entries.0.amount", "400"),
				),
			},
			{
				Config: testAccCheckDatadogCostCustomForecastUpdated(budgetName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"datadog_cost_custom_forecast.foo", "entries.0.amount", "500"),
				),
			},
		},
	})
}

func TestAccDatadogCostCustomForecast_NegativeAmountRejected(t *testing.T) {
	t.Parallel()
	_, _, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	budgetName := uniqueEntityName(context.Background(), t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: accProviders,
		Steps: []resource.TestStep{
			{
				Config:      testAccCheckDatadogCostCustomForecastNegativeAmount(budgetName),
				ExpectError: regexp.MustCompile(`value must be at least`),
			},
		},
	})
}

func testAccCheckDatadogCostCustomForecastDestroy(provider *fwprovider.FrameworkProvider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		apiInstances := provider.DatadogApiInstances
		auth := provider.Auth

		for _, r := range s.RootModule().Resources {
			if r.Type != "datadog_cost_custom_forecast" {
				continue
			}

			_, httpResp, err := apiInstances.GetCloudCostManagementApiV2().GetCustomForecast(auth, r.Primary.Attributes["budget_uid"])
			if err != nil {
				if httpResp != nil && httpResp.StatusCode == 404 {
					continue
				}
				return fmt.Errorf("received an error retrieving custom forecast: %s", err.Error())
			}
			return fmt.Errorf("cost custom forecast still exists")
		}
		return nil
	}
}

func testAccCheckDatadogCostCustomForecastBasic(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_cost_budget" "foo" {
  name          = "%s"
  metrics_query = "sum:aws.cost.amortized{service:ec2} by {service}"
  start_month   = 202501
  end_month     = 202502
  entries {
    amount = 1000
    month  = 202501
    tag_filters {
      tag_key   = "service"
      tag_value = "ec2"
    }
  }
  entries {
    amount = 1000
    month  = 202502
    tag_filters {
      tag_key   = "service"
      tag_value = "ec2"
    }
  }
}

resource "datadog_cost_custom_forecast" "foo" {
  budget_uid = datadog_cost_budget.foo.id
  entries {
    amount = 400
    month  = 202501
    tag_filters {
      tag_key   = "service"
      tag_value = "ec2"
    }
  }
}
`, uniq)
}

func testAccCheckDatadogCostCustomForecastUpdated(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_cost_budget" "foo" {
  name          = "%s"
  metrics_query = "sum:aws.cost.amortized{service:ec2} by {service}"
  start_month   = 202501
  end_month     = 202502
  entries {
    amount = 1000
    month  = 202501
    tag_filters {
      tag_key   = "service"
      tag_value = "ec2"
    }
  }
  entries {
    amount = 1000
    month  = 202502
    tag_filters {
      tag_key   = "service"
      tag_value = "ec2"
    }
  }
}

resource "datadog_cost_custom_forecast" "foo" {
  budget_uid = datadog_cost_budget.foo.id
  entries {
    amount = 500
    month  = 202501
    tag_filters {
      tag_key   = "service"
      tag_value = "ec2"
    }
  }
}
`, uniq)
}

func testAccCheckDatadogCostCustomForecastNegativeAmount(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_cost_budget" "foo" {
  name          = "%s"
  metrics_query = "sum:aws.cost.amortized{service:ec2} by {service}"
  start_month   = 202501
  end_month     = 202501
  entries {
    amount = 1000
    month  = 202501
    tag_filters {
      tag_key   = "service"
      tag_value = "ec2"
    }
  }
}

resource "datadog_cost_custom_forecast" "foo" {
  budget_uid = datadog_cost_budget.foo.id
  entries {
    amount = -1
    month  = 202501
    tag_filters {
      tag_key   = "service"
      tag_value = "ec2"
    }
  }
}
`, uniq)
}
