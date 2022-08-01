package test

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/terraform-providers/terraform-provider-datadog/datadog"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccDatadogMetricTagConfiguration_Error(t *testing.T) {
	ctx, accProviders := testAccProviders(context.Background(), t)
	uniqueMetricTagConfig := strings.ReplaceAll(uniqueEntityName(ctx, t), "-", "_")
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		CheckDestroy:      testAccCheckDatadogMetricTagConfigurationDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config:      testAccCheckDatadogMetricTagConfigurationIncludePercentilesError(uniqueMetricTagConfig, "count"),
				ExpectError: regexp.MustCompile("cannot use include_percentiles with a metric_type of count*"),
			},
			{
				Config:      testAccCheckDatadogMetricTagConfigurationIncludePercentilesError(uniqueMetricTagConfig, "gauge"),
				ExpectError: regexp.MustCompile("cannot use include_percentiles with a metric_type of gauge*"),
			},
			{
				Config:      testAccCheckDatadogMetricTagConfigurationAggregationsError(uniqueMetricTagConfig, "distribution"),
				ExpectError: regexp.MustCompile("cannot use aggregations with a metric_type of distribution*"),
			},
		},
	})
}

func testAccCheckDatadogMetricTagConfigurationIncludePercentilesError(uniq string, metricType string) string {
	return fmt.Sprintf(`
        resource "datadog_metric_tag_configuration" "testing_metric_tag_config_icl_count" {
			metric_name = "%s"
			metric_type = "%s"
			tags = ["sport"]
			include_percentiles = false
        }
    `, uniq, metricType)
}

func testAccCheckDatadogMetricTagConfigurationAggregationsError(uniq string, metricType string) string {
	return fmt.Sprintf(`
        resource "datadog_metric_tag_configuration" "testing_metric_tag_config_aggregations" {
			metric_name = "%s"
			metric_type = "%s"
			tags = ["sport"]
			aggregations {
				time = "sum"
				space = "sum"
			}
			aggregations {
				time = "avg"
				space = "avg"
			}
        }
    `, uniq, metricType)
}

func testAccCheckDatadogMetricTagConfigurationDestroy(accProvider func() (*schema.Provider, error)) func(*terraform.State) error {
	return func(s *terraform.State) error {
		provider, _ := accProvider()
		meta := provider.Meta()
		providerConf := meta.(*datadog.ProviderConfiguration)
		datadogClient := providerConf.DatadogClient
		auth := providerConf.Auth
		for _, r := range s.RootModule().Resources {
			if r.Type != "datadog_metric_tag_configuration" {
				continue
			}

			var err error

			id := r.Primary.ID

			_, resp, err := utils.GetMetricsApiV2(datadogClient).ListTagConfigurationByName(auth, id)

			if err != nil {
				if resp.StatusCode == 404 {
					continue // resource not found => all ok
				} else {
					return fmt.Errorf("received an error retrieving metric_tag_configuration: %s", err.Error())
				}
			} else {
				return fmt.Errorf("metric_tag_configuration %s still exists", r.Primary.ID)
			}
		}

		return nil
	}
}
