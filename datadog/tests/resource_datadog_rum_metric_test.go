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

	// The ID needs to be a valid metric name, otherwise it will get rejected with a 400.
	// uniqueEntityName() returns dash separated words, which is not valid, so we replace
	// them with underscores.
	uniq := strings.ReplaceAll(uniqueEntityName(ctx, t), "-", "_")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: accProviders,
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
		ProtoV6ProviderFactories: accProviders,
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
		ProtoV6ProviderFactories: accProviders,
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
			{
				Config: distributionDatadogRumMetricUpdate(uniq),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogRumMetricExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(
						"datadog_rum_metric.testing_rum_metric", "compute.include_percentiles", "false"),
				),
			},
		},
	})

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: accProviders,
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
			{
				Config: filterDatadogRumMetricUpdate(uniq),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogRumMetricExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(
						"datadog_rum_metric.testing_rum_metric", "filter.query", "@service:another-service"),
				),
			},
			{
				Config: filterDatadogRumMetricDelete(uniq),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogRumMetricExists(providers.frameworkProvider),
					resource.TestCheckNoResourceAttr(
						"datadog_rum_metric.testing_rum_metric", "filter"),
				),
			},
		},
	})

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogRumMetricDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: groupByDatadogRumMetric(uniq),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogRumMetricExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(
						"datadog_rum_metric.testing_rum_metric", "group_by.#", "3"),
					resource.TestCheckResourceAttr(
						"datadog_rum_metric.testing_rum_metric", "group_by.0.path", "@os"),
					resource.TestCheckResourceAttr(
						"datadog_rum_metric.testing_rum_metric", "group_by.0.tag_name", "os"),
					resource.TestCheckResourceAttr(
						"datadog_rum_metric.testing_rum_metric", "group_by.1.path", "@service"),
					resource.TestCheckResourceAttr(
						"datadog_rum_metric.testing_rum_metric", "group_by.1.tag_name", "service"),
					resource.TestCheckResourceAttr(
						"datadog_rum_metric.testing_rum_metric", "group_by.2.path", "path_only"),
					resource.TestCheckResourceAttr(
						"datadog_rum_metric.testing_rum_metric", "group_by.2.tag_name", "path_only"),
				),
			},
			{
				Config: groupByDatadogRumMetricUpdate(uniq),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogRumMetricExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(
						"datadog_rum_metric.testing_rum_metric", "group_by.#", "3"),
					resource.TestCheckResourceAttr(
						"datadog_rum_metric.testing_rum_metric", "group_by.0.path", "@os"),
					resource.TestCheckResourceAttr(
						"datadog_rum_metric.testing_rum_metric", "group_by.0.tag_name", "os"),
					resource.TestCheckResourceAttr(
						"datadog_rum_metric.testing_rum_metric", "group_by.1.path", "@os.version.major"),
					resource.TestCheckResourceAttr(
						"datadog_rum_metric.testing_rum_metric", "group_by.1.tag_name", "os_version_major"),
					resource.TestCheckResourceAttr(
						"datadog_rum_metric.testing_rum_metric", "group_by.2.path", "@os.version.minor"),
					resource.TestCheckResourceAttr(
						"datadog_rum_metric.testing_rum_metric", "group_by.2.tag_name", "os_version_minor"),
				),
			},
			{
				Config: groupByDatadogRumMetricDelete(uniq),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogRumMetricExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(
						"datadog_rum_metric.testing_rum_metric", "group_by.#", "0"),
				),
			},
		},
	})

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogRumMetricDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: uniquenessDatadogRumMetric(uniq),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogRumMetricExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(
						"datadog_rum_metric.testing_rum_metric", "uniqueness.when", "match"),
				),
			},
			{
				Config: uniquenessDatadogRumMetricUpdate(uniq),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogRumMetricExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(
						"datadog_rum_metric.testing_rum_metric", "uniqueness.when", "end"),
				),
			},
			{
				// Simply verify that we don't run into an error during planning. Deleting only the uniqueness is
				// not valid, because we would need to update the event type. This makes sure the user sees the 400
				// error from the API at apply time.
				Config: uniquenessDatadogRumMetricDelete(uniq),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogRumMetricExists(providers.frameworkProvider),
				),
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
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

func distributionDatadogRumMetricUpdate(uniq string) string {
	return fmt.Sprintf(`resource "datadog_rum_metric" "testing_rum_metric" {
		name = %q
	    event_type = "action"
    	compute {
    		aggregation_type = "distribution"
    		include_percentiles = false
    		path = "@duration"
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

func filterDatadogRumMetricUpdate(uniq string) string {
	return fmt.Sprintf(`resource "datadog_rum_metric" "testing_rum_metric" {
		name = %q
	    event_type = "action"
    	compute {
    		aggregation_type = "count"
    	}
		filter {
			query = "@service:another-service"
		}
	}
	`, uniq)
}

func filterDatadogRumMetricDelete(uniq string) string {
	return fmt.Sprintf(`resource "datadog_rum_metric" "testing_rum_metric" {
		name = %q
	    event_type = "action"
    	compute {
    		aggregation_type = "count"
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
		group_by {
			path = "path_only"
		}
	}
	`, uniq)
}

func groupByDatadogRumMetricUpdate(uniq string) string {
	return fmt.Sprintf(`resource "datadog_rum_metric" "testing_rum_metric" {
		name = %q
	    event_type = "action"
    	compute {
    		aggregation_type = "count"
    	}
		group_by {
			path = "@os"
			tag_name = "os"
		}
		group_by {
			path = "@os.version.major"
			tag_name = "os_version_major"
		}
		group_by {
			path = "@os.version.minor"
			tag_name = "os_version_minor"
		}
	}
	`, uniq)
}

func groupByDatadogRumMetricDelete(uniq string) string {
	return fmt.Sprintf(`resource "datadog_rum_metric" "testing_rum_metric" {
		name = %q
	    event_type = "action"
    	compute {
    		aggregation_type = "count"
    	}
	}
	`, uniq)
}

func uniquenessDatadogRumMetric(uniq string) string {
	return fmt.Sprintf(`resource "datadog_rum_metric" "testing_rum_metric" {
		name = %q
	    event_type = "session"
    	compute {
    		aggregation_type = "count"
    	}
		uniqueness {
			when = "match"
		}
	}
	`, uniq)
}

func uniquenessDatadogRumMetricUpdate(uniq string) string {
	return fmt.Sprintf(`resource "datadog_rum_metric" "testing_rum_metric" {
		name = %q
	    event_type = "session"
    	compute {
    		aggregation_type = "count"
    	}
		uniqueness {
			when = "end"
		}
	}
	`, uniq)
}

func uniquenessDatadogRumMetricDelete(uniq string) string {
	// Note: this is not a valid configuration, but it's used to verify that the provider will
	// behave properly when the attribute is deleted.
	return fmt.Sprintf(`resource "datadog_rum_metric" "testing_rum_metric" {
		name = %q
	    event_type = "session"
    	compute {
    		aggregation_type = "count"
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
