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

func TestAccRumMetric_import(t *testing.T) {
	t.Parallel()
	resourceName := "datadog_rum_metric.testing_rum_metric"
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	uniq := strings.ReplaceAll(uniqueEntityName(ctx, t), "-", "_")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogRumMetricDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogRumMetric(uniq),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccRumMetricBasic(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	uniq := strings.ReplaceAll(uniqueEntityName(ctx, t), "-", "_")

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogRumMetricDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogRumMetric(uniq),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogRumMetricExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(
						"datadog_rum_metric.testing_rum_metric", "id", uniq),
					resource.TestCheckResourceAttr(
						"datadog_rum_metric.testing_rum_metric", "name", uniq),
					resource.TestCheckResourceAttr(
						"datadog_rum_metric.testing_rum_metric", "event_type", "session"),
					resource.TestCheckResourceAttr(
						"datadog_rum_metric.testing_rum_metric", "compute.aggregation_type", "distribution"),
					resource.TestCheckResourceAttr(
						"datadog_rum_metric.testing_rum_metric", "compute.include_percentiles", "true"),
					resource.TestCheckResourceAttr(
						"datadog_rum_metric.testing_rum_metric", "compute.path", "@duration"),
					resource.TestCheckResourceAttr(
						"datadog_rum_metric.testing_rum_metric", "filter.query", "@service:web-ui"),
					resource.TestCheckResourceAttr(
						"datadog_rum_metric.testing_rum_metric", "group_by.#", "1"),
					resource.TestCheckResourceAttr(
						"datadog_rum_metric.testing_rum_metric", "group_by.0.path", "@browser.name"),
					resource.TestCheckResourceAttr(
						"datadog_rum_metric.testing_rum_metric", "group_by.0.tag_name", "browser_name"),
					resource.TestCheckResourceAttr(
						"datadog_rum_metric.testing_rum_metric", "uniqueness.when", "match"),
				),
			},
		},
	})
}

func testAccCheckDatadogRumMetric(uniq string) string {
	return fmt.Sprintf(`resource "datadog_rum_metric" "testing_rum_metric" {
		name = %q
	    event_type = "session"
    	compute {
    		aggregation_type = "distribution"
    		include_percentiles = true
    		path = "@duration"
    	}
    	filter {
    		query = "@service:web-ui"
    	}
    	group_by {
			path = "@browser.name"
			tag_name = "browser_name"
    	}
		uniqueness {
			when = "match"
    	}
	}
	`, uniq)
}

func testAccCheckDatadogRumMetricDestroy(accProvider *fwprovider.FrameworkProvider) func(*terraform.State) error {
	return func(s *terraform.State) error {
		apiInstances := accProvider.DatadogApiInstances
		auth := accProvider.Auth

		if err := RumMetricDestroyHelper(auth, s, apiInstances); err != nil {
			return err
		}
		return nil
	}
}

func RumMetricDestroyHelper(auth context.Context, s *terraform.State, apiInstances *utils.ApiInstances) error {
	err := utils.Retry(2, 10, func() error {
		for _, r := range s.RootModule().Resources {
			if r.Type != "resource_datadog_rum_metric" {
				continue
			}
			id := r.Primary.ID

			_, httpResp, err := apiInstances.GetRumMetricsApiV2().GetRumMetric(auth, id)
			if err != nil {
				if httpResp != nil && httpResp.StatusCode == 404 {
					return nil
				}
				return &utils.RetryableError{Prob: fmt.Sprintf("received an error retrieving RumMetric %s", err)}
			}
			return &utils.RetryableError{Prob: "RumMetric still exists"}
		}
		return nil
	})
	return err
}

func testAccCheckDatadogRumMetricExists(accProvider *fwprovider.FrameworkProvider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		apiInstances := accProvider.DatadogApiInstances
		auth := accProvider.Auth

		if err := rumMetricExistsHelper(auth, s, apiInstances); err != nil {
			return err
		}
		return nil
	}
}

func rumMetricExistsHelper(auth context.Context, s *terraform.State, apiInstances *utils.ApiInstances) error {
	for _, r := range s.RootModule().Resources {
		if r.Type != "resource_datadog_rum_metric" {
			continue
		}
		id := r.Primary.ID

		_, httpResp, err := apiInstances.GetRumMetricsApiV2().GetRumMetric(auth, id)
		if err != nil {
			return utils.TranslateClientError(err, httpResp, "error retrieving RumMetric")
		}
	}
	return nil
}
