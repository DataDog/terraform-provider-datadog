package test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/fwprovider"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

func TestAccDerivedMetricBasic(t *testing.T) {
	t.Parallel()
	resourceName := "datadog_derived_metric.foo"
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)

	// The ID needs to be a valid metric name. uniqueEntityName() returns dash separated
	// words, which is not valid, so we replace them with underscores.
	uniq := strings.ReplaceAll(uniqueEntityName(ctx, t), "-", "_")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogDerivedMetricDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogDerivedMetric(uniq),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogDerivedMetricExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(resourceName, "id", uniq),
					resource.TestCheckResourceAttr(resourceName, "name", uniq),
					resource.TestCheckResourceAttr(resourceName, "formula", "a"),
					resource.TestCheckResourceAttr(resourceName, "queries.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "queries.0.name", "a"),
					resource.TestCheckResourceAttr(resourceName, "queries.0.data_source", "metrics"),
					resource.TestCheckResourceAttr(resourceName, "queries.0.query", "avg:system.cpu.user{*}"),
				),
			},
			{
				Config: testAccCheckDatadogDerivedMetricUpdated(uniq),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogDerivedMetricExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(resourceName, "id", uniq),
					resource.TestCheckResourceAttr(resourceName, "formula", "a + b"),
					resource.TestCheckResourceAttr(resourceName, "queries.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "queries.1.name", "b"),
					resource.TestCheckResourceAttr(resourceName, "queries.1.data_source", "metrics"),
					resource.TestCheckResourceAttr(resourceName, "queries.1.query", "avg:system.cpu.system{*}"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckDatadogDerivedMetric(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_derived_metric" "foo" {
  name    = "%s"
  formula = "a"
  queries = [
    {
      name        = "a"
      data_source = "metrics"
      query       = "avg:system.cpu.user{*}"
    },
  ]
}`, uniq)
}

func testAccCheckDatadogDerivedMetricUpdated(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_derived_metric" "foo" {
  name    = "%s"
  formula = "a + b"
  queries = [
    {
      name        = "a"
      data_source = "metrics"
      query       = "avg:system.cpu.user{*}"
    },
    {
      name        = "b"
      data_source = "metrics"
      query       = "avg:system.cpu.system{*}"
    },
  ]
}`, uniq)
}

func testAccCheckDatadogDerivedMetricDestroy(accProvider *fwprovider.FrameworkProvider) func(*terraform.State) error {
	return func(s *terraform.State) error {
		apiInstances := accProvider.DatadogApiInstances
		auth := accProvider.Auth

		if err := derivedMetricDestroyHelper(auth, s, apiInstances); err != nil {
			return err
		}
		return nil
	}
}

func derivedMetricDestroyHelper(auth context.Context, s *terraform.State, apiInstances *utils.ApiInstances) error {
	err := utils.Retry(2, 10, func() error {
		for _, r := range s.RootModule().Resources {
			if r.Type != "datadog_derived_metric" {
				continue
			}
			id := r.Primary.ID

			_, httpResp, err := apiInstances.GetMetricsApiV2().GetDerivedMetric(auth, id)
			if err != nil {
				if httpResp != nil && httpResp.StatusCode == 404 {
					return nil
				}
				return &utils.RetryableError{Prob: fmt.Sprintf("received an error retrieving derived metric %s", err)}
			}
			return &utils.RetryableError{Prob: "derived metric still exists"}
		}
		return nil
	})
	return err
}

func testAccCheckDatadogDerivedMetricExists(accProvider *fwprovider.FrameworkProvider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		apiInstances := accProvider.DatadogApiInstances
		auth := accProvider.Auth

		if err := derivedMetricExistsHelper(auth, s, apiInstances); err != nil {
			return err
		}
		return nil
	}
}

func derivedMetricExistsHelper(auth context.Context, s *terraform.State, apiInstances *utils.ApiInstances) error {
	for _, r := range s.RootModule().Resources {
		if r.Type != "datadog_derived_metric" {
			continue
		}
		id := r.Primary.ID

		_, httpResp, err := apiInstances.GetMetricsApiV2().GetDerivedMetric(auth, id)
		if err != nil {
			return utils.TranslateClientError(err, httpResp, "error retrieving derived metric")
		}
	}
	return nil
}
