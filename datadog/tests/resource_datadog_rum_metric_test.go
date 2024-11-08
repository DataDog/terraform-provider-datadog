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

func TestAccRumMetricImport(t *testing.T) {
	t.Parallel()
	resourceName := "datadog_rum_metric.testing_rum_metric"
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)

	// The API will currently silently remap - to _, which makes terraform unhappy. This will
	// just make the tests pass but this is a real issue.
	// Being addressed in https://datadoghq.atlassian.net/browse/RUM-7124. Note that this is
	// also an issue for other existing APIs using the same backend (spans metrics and maybe
	// logs metrics?).
	uniq := strings.ReplaceAll(uniqueEntityName(ctx, t), "-", "_")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogRumMetricDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: minimalDatadogRumMetric(uniq),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccRumMetricAttributes(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	uniq := strings.ReplaceAll(uniqueEntityName(ctx, t), "-", "_")

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogRumMetricDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: minimalDatadogRumMetric(uniq),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogRumMetricExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(
						"datadog_rum_metric.testing_rum_metric", "id", uniq),
					resource.TestCheckResourceAttr(
						"datadog_rum_metric.testing_rum_metric", "name", uniq),
					resource.TestCheckResourceAttr(
						"datadog_rum_metric.testing_rum_metric", "event_type", "action"),
					resource.TestCheckResourceAttr(
						"datadog_rum_metric.testing_rum_metric", "compute.aggregation_type", "count"),
				),
			},
		},
	})

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogRumMetricDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: distributionDatadogRumMetric(uniq),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogRumMetricExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(
						"datadog_rum_metric.testing_rum_metric", "compute.aggregation_type", "distribution"),
					resource.TestCheckResourceAttr(
						"datadog_rum_metric.testing_rum_metric", "compute.include_percentiles", "true"),
					resource.TestCheckResourceAttr(
						"datadog_rum_metric.testing_rum_metric", "compute.path", "@duration"),
				),
			},
		},
	})

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogRumMetricDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: filterDatadogRumMetric(uniq),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogRumMetricExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(
						"datadog_rum_metric.testing_rum_metric", "filter.query", "@service:web-ui"),
				),
			},
		},
	})

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogRumMetricDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: groupByDatadogRumMetric(uniq),
				Check:  testAccCheckDatadogRumMetricExists(providers.frameworkProvider),
				// resource.ComposeTestCheckFunc(

				// 	resource.TestCheckResourceAttr(
				// 		"datadog_rum_metric.testing_rum_metric", "group_by.#", "2"),
				// ),
			},
		},
	})
}

func minimalDatadogRumMetric(uniq string) string {
	return fmt.Sprintf(`resource "datadog_rum_metric" "testing_rum_metric" {
		name = %q
	    event_type = "action"
    	compute {
    		aggregation_type = "count"
    	}
	}
	`, uniq)
}

func distributionDatadogRumMetric(uniq string) string {
	return fmt.Sprintf(`resource "datadog_rum_metric" "testing_rum_metric" {
		name = %q
	    event_type = "action"
    	compute {
    		aggregation_type = "distribution"
    		include_percentiles = true
    		path = "@duration"
    	}
	}
	`, uniq)
}

func countDatadogRumMetric(uniq string) string {
	return fmt.Sprintf(`resource "datadog_rum_metric" "testing_rum_metric" {
		name = %q
	    event_type = "action"
    	compute {
    		aggregation_type = "count"
    	}
	}
	`, uniq)
}

func filterDatadogRumMetric(uniq string) string {
	return fmt.Sprintf(`resource "datadog_rum_metric" "testing_rum_metric" {
		name = %q
	    event_type = "action"
    	compute {
    		aggregation_type = "count"
    	}
		filter {
			query = "@service:web-ui"
		}
	}
	`, uniq)
}

func groupByDatadogRumMetric(uniq string) string {
	// Note: the group_bys are not defined in alphabetical order. This is on purpose to verify
	// a set behavior rather than a list behavior on the terraform attribute.
	return fmt.Sprintf(`resource "datadog_rum_metric" "testing_rum_metric" {
		name = %q
	    event_type = "action"
    	compute {
    		aggregation_type = "count"
    	}
		group_by {
			path = "@service"
			tag_name = "service"
		}
		group_by {
			path = "@os"
			tag_name = "os"
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
