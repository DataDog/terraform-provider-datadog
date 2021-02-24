package test

import (
	"context"
	"fmt"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
	"strings"
	"testing"

	"github.com/terraform-providers/terraform-provider-datadog/datadog"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccDatadogLogsMetric_import(t *testing.T) {
	t.Parallel()
	resourceName := "datadog_logs_metric.testing_logs_metric"
	ctx, accProviders := testAccProviders(context.Background(), t)
	uniqueLogsMetric := strings.ReplaceAll(uniqueEntityName(ctx, t), "-", "_")
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    accProviders,
		CheckDestroy: testAccCheckDatadogLogsMetricDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogLogsMetricConfig_Basic(uniqueLogsMetric),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccDatadogLogsMetric_Basic(t *testing.T) {
	t.Parallel()
	ctx, accProviders := testAccProviders(context.Background(), t)
	uniqueLogsMetric := strings.ReplaceAll(uniqueEntityName(ctx, t), "-", "_")
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    accProviders,
		CheckDestroy: testAccCheckDatadogLogsMetricDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogLogsMetricConfig_Basic(uniqueLogsMetric),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogLogsMetricExists(accProvider, "datadog_logs_metric.testing_logs_metric"),
					resource.TestCheckResourceAttr(
						"datadog_logs_metric.testing_logs_metric", "compute.#", "1"),
					resource.TestCheckResourceAttr(
						"datadog_logs_metric.testing_logs_metric", "compute.0.aggregation_type", "distribution"),
					resource.TestCheckResourceAttr(
						"datadog_logs_metric.testing_logs_metric", "compute.0.path", "@duration"),
					resource.TestCheckResourceAttr(
						"datadog_logs_metric.testing_logs_metric", "filter.#", "1"),
					resource.TestCheckResourceAttr(
						"datadog_logs_metric.testing_logs_metric", "filter.0.query", "service:test"),
					resource.TestCheckResourceAttr(
						"datadog_logs_metric.testing_logs_metric", "group_by.#", "2"),
					resource.TestCheckResourceAttr(
						"datadog_logs_metric.testing_logs_metric", "group_by.0.path", "@my.status"),
					resource.TestCheckResourceAttr(
						"datadog_logs_metric.testing_logs_metric", "group_by.0.tag_name", "status"),
					resource.TestCheckResourceAttr(
						"datadog_logs_metric.testing_logs_metric", "group_by.1.path", "service"),
				),
			},
		},
	})
}

func testAccCheckDatadogLogsMetricConfig_Basic(uniq string) string {
	return fmt.Sprintf(`
        resource "datadog_logs_metric" "testing_logs_metric" {
			name = "%s"
			compute {
				aggregation_type = "distribution"
				path             = "@duration"
			}
			filter {
				query = "service:test"
			}
			group_by {
				path     = "@my.status"
				tag_name = "status"
			}
			group_by {
				path     = "service"
				tag_name = "service"
			}
        }
    `, uniq)
}

func testAccCheckDatadogLogsMetricExists(accProvider *schema.Provider, resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		meta := accProvider.Meta()
		resourceId := s.RootModule().Resources[resourceName].Primary.ID
		providerConf := meta.(*datadog.ProviderConfiguration)
		datadogClient := providerConf.DatadogClientV2
		auth := providerConf.AuthV2
		var err error

		id := resourceId

		_, _, err = datadogClient.LogsMetricsApi.GetLogsMetric(auth, id).Execute()

		if err != nil {
			return utils.TranslateClientError(err, "error checking logs_metric existence")
		}

		return nil
	}
}

func testAccCheckDatadogLogsMetricDestroy(accProvider *schema.Provider) func(*terraform.State) error {
	return func(s *terraform.State) error {
		meta := accProvider.Meta()
		providerConf := meta.(*datadog.ProviderConfiguration)
		datadogClient := providerConf.DatadogClientV2
		auth := providerConf.AuthV2
		for _, r := range s.RootModule().Resources {
			if r.Type != "datadog_logs_metric" {
				continue
			}

			var err error

			id := r.Primary.ID

			_, resp, err := datadogClient.LogsMetricsApi.GetLogsMetric(auth, id).Execute()

			if err != nil {
				if resp.StatusCode == 404 {
					continue // resource not found => all ok
				} else {
					return fmt.Errorf("received an error retrieving logs_metric: %s", err.Error())
				}
			} else {
				return fmt.Errorf("logs_metric %s still exists", r.Primary.ID)
			}
		}

		return nil
	}
}
