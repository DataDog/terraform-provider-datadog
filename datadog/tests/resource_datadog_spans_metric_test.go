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

func TestAccSpansMetric_import(t *testing.T) {
	t.Parallel()
	resourceName := "datadog_spans_metric.testing_spans_metric"
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	uniq := strings.ReplaceAll(uniqueEntityName(ctx, t), "-", "_")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogSpansMetricDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogSpansMetricTesting(uniq),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccSpansMetricBasic(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	uniq := strings.ReplaceAll(uniqueEntityName(ctx, t), "-", "_")

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogSpansMetricDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogSpansMetricTesting(uniq),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogSpansMetricExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(
						"datadog_spans_metric.testing_spans_metric", "compute.aggregation_type", "distribution"),
					resource.TestCheckResourceAttr(
						"datadog_spans_metric.testing_spans_metric", "compute.include_percentiles", "false"),
					resource.TestCheckResourceAttr(
						"datadog_spans_metric.testing_spans_metric", "compute.path", "@duration"),
					resource.TestCheckResourceAttr(
						"datadog_spans_metric.testing_spans_metric", "filter.query", "@http.status_code:200 service:my-service"),
					resource.TestCheckResourceAttr(
						"datadog_spans_metric.testing_spans_metric", "group_by.#", "1"),
					resource.TestCheckResourceAttr(
						"datadog_spans_metric.testing_spans_metric", "group_by.0.path", "resource_name"),
					resource.TestCheckResourceAttr(
						"datadog_spans_metric.testing_spans_metric", "group_by.0.tag_name", "resource_name"),
				),
			},
			{
				Config: testAccCheckDatadogSpansMetricTestingUpdate(uniq),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogSpansMetricExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(
						"datadog_spans_metric.testing_spans_metric", "compute.include_percentiles", "true"),
					resource.TestCheckResourceAttr(
						"datadog_spans_metric.testing_spans_metric", "group_by.#", "0"),
					resource.TestCheckResourceAttr(
						"datadog_spans_metric.testing_spans_metric", "filter.query", "*"),
				),
			},
		},
	})
}

func TestAccSpansMetricGroupBys(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	uniq := strings.ReplaceAll(uniqueEntityName(ctx, t), "-", "_")

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogSpansMetricDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogSpansMetricTestingCountGroupBys(uniq),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogSpansMetricExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(
						"datadog_spans_metric.testing_spans_metric", "compute.aggregation_type", "count"),
					resource.TestCheckResourceAttr(
						"datadog_spans_metric.testing_spans_metric", "filter.query", "@http.status_code:200 service:my-service"),
					resource.TestCheckResourceAttr(
						"datadog_spans_metric.testing_spans_metric", "group_by.#", "2"),
					resource.TestCheckResourceAttr(
						"datadog_spans_metric.testing_spans_metric", "group_by.0.path", "resource_name1"),
					resource.TestCheckResourceAttr(
						"datadog_spans_metric.testing_spans_metric", "group_by.0.tag_name", "my_resource1"),
					resource.TestCheckResourceAttr(
						"datadog_spans_metric.testing_spans_metric", "group_by.1.path", "resource_name2"),
					resource.TestCheckResourceAttr(
						"datadog_spans_metric.testing_spans_metric", "group_by.1.tag_name", "my_resource2"),
				),
			},
		},
	})
}

func testAccCheckDatadogSpansMetricDestroy(accProvider *fwprovider.FrameworkProvider) func(*terraform.State) error {
	return func(s *terraform.State) error {
		apiInstances := accProvider.DatadogApiInstances
		auth := accProvider.Auth

		if err := SpansMetricDestroyHelper(auth, s, apiInstances); err != nil {
			return err
		}
		return nil
	}
}

func SpansMetricDestroyHelper(auth context.Context, s *terraform.State, apiInstances *utils.ApiInstances) error {
	err := utils.Retry(2, 10, func() error {
		for _, r := range s.RootModule().Resources {
			if r.Type != "resource_datadog_spans_metric" {
				continue
			}
			id := r.Primary.ID

			_, httpResp, err := apiInstances.GetSpansMetricsApiV2().GetSpansMetric(auth, id)
			if err != nil {
				if httpResp != nil && httpResp.StatusCode == 404 {
					return nil
				}
				return &utils.RetryableError{Prob: fmt.Sprintf("received an error retrieving spans metric %s", err)}
			}
			return &utils.RetryableError{Prob: "spans metric still exists"}
		}
		return nil
	})
	return err
}

func testAccCheckDatadogSpansMetricExists(accProvider *fwprovider.FrameworkProvider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		apiInstances := accProvider.DatadogApiInstances
		auth := accProvider.Auth

		if err := spansMetricExistsHelper(auth, s, apiInstances); err != nil {
			return err
		}
		return nil
	}
}

func spansMetricExistsHelper(auth context.Context, s *terraform.State, apiInstances *utils.ApiInstances) error {
	for _, r := range s.RootModule().Resources {
		if r.Type != "resource_datadog_spans_metric" {
			continue
		}
		id := r.Primary.ID

		_, httpResp, err := apiInstances.GetSpansMetricsApiV2().GetSpansMetric(auth, id)
		if err != nil {
			return utils.TranslateClientError(err, httpResp, "error retrieving spans metric")
		}
	}
	return nil
}

func testAccCheckDatadogSpansMetricTesting(uniq string) string {
	return fmt.Sprintf(`
		resource "datadog_spans_metric" "testing_spans_metric" {
			name = "%s"
			compute {
				aggregation_type    = "distribution"
				include_percentiles = false
				path                = "@duration"
			}
			filter {
				query = "@http.status_code:200 service:my-service"
			}
			group_by {
				path     = "resource_name"
				tag_name = "resource_name"
			}
		}
	`, uniq)
}

func testAccCheckDatadogSpansMetricTestingUpdate(uniq string) string {
	return fmt.Sprintf(`
		resource "datadog_spans_metric" "testing_spans_metric" {
			name = "%s"
			compute {
				aggregation_type    = "distribution"
				include_percentiles = true
				path                = "@duration"
			}
			filter {
			}
		}
	`, uniq)
}

func testAccCheckDatadogSpansMetricTestingCountGroupBys(uniq string) string {
	return fmt.Sprintf(`
		resource "datadog_spans_metric" "testing_spans_metric" {
			name = "%s"
			compute {
				aggregation_type    = "count"
			}
			filter {
				query = "@http.status_code:200 service:my-service"
			}
			group_by {
				path     = "resource_name2"
				tag_name = "my_resource2"
			}
			group_by {
				path     = "resource_name1"
				tag_name = "my_resource1"
			}
		}
	`, uniq)
}
