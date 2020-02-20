package datadog

import (
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/zorkian/go-datadog-api"
)

func TestAccDatadogMetricMetadata_Basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogMetricMetadataConfig,
				Check: resource.ComposeTestCheckFunc(
					checkPostEvent(),
					checkMetricMetadataExists("datadog_metric_metadata.foo"),
					resource.TestCheckResourceAttr(
						"datadog_metric_metadata.foo", "short_name", "short name for metric_metadata foo"),
					resource.TestCheckResourceAttr(
						"datadog_metric_metadata.foo", "type", "gauge"),
					resource.TestCheckResourceAttr(
						"datadog_metric_metadata.foo", "description", "some description"),
					resource.TestCheckResourceAttr(
						"datadog_metric_metadata.foo", "unit", "byte"),
					resource.TestCheckResourceAttr(
						"datadog_metric_metadata.foo", "per_unit", "second"),
					resource.TestCheckResourceAttr(
						"datadog_metric_metadata.foo", "statsd_interval", "1"),
				),
			},
		},
	})
}

func TestAccDatadogMetricMetadata_Updated(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogMetricMetadataConfig,
				Check: resource.ComposeTestCheckFunc(
					checkPostEvent(),
					checkMetricMetadataExists("datadog_metric_metadata.foo"),
					resource.TestCheckResourceAttr(
						"datadog_metric_metadata.foo", "short_name", "short name for metric_metadata foo"),
					resource.TestCheckResourceAttr(
						"datadog_metric_metadata.foo", "type", "gauge"),
					resource.TestCheckResourceAttr(
						"datadog_metric_metadata.foo", "description", "some description"),
					resource.TestCheckResourceAttr(
						"datadog_metric_metadata.foo", "unit", "byte"),
					resource.TestCheckResourceAttr(
						"datadog_metric_metadata.foo", "per_unit", "second"),
					resource.TestCheckResourceAttr(
						"datadog_metric_metadata.foo", "statsd_interval", "1"),
				),
			},
			{
				Config: testAccCheckDatadogMetricMetadataConfigUpdated,
				Check: resource.ComposeTestCheckFunc(
					checkMetricMetadataExists("datadog_metric_metadata.foo"),
					resource.TestCheckResourceAttr(
						"datadog_metric_metadata.foo", "short_name", "short name for metric_metadata foo"),
					resource.TestCheckResourceAttr(
						"datadog_metric_metadata.foo", "type", "gauge"),
					resource.TestCheckResourceAttr(
						"datadog_metric_metadata.foo", "description", "a different description"),
					resource.TestCheckResourceAttr(
						"datadog_metric_metadata.foo", "unit", "byte"),
					resource.TestCheckResourceAttr(
						"datadog_metric_metadata.foo", "per_unit", "second"),
					resource.TestCheckResourceAttr(
						"datadog_metric_metadata.foo", "statsd_interval", "1"),
				),
			},
		},
	})
}

func checkMetricMetadataExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client := testAccProvider.Meta().(*datadog.Client)
		for _, r := range s.RootModule().Resources {
			metric, ok := r.Primary.Attributes["metric"]
			if !ok {
				continue
			}
			if _, err := client.ViewMetricMetadata(metric); err != nil {
				return fmt.Errorf("Received an error retrieving metric_metadata %s", err)
			}
		}
		return nil
	}
}

func checkPostEvent() resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client := testAccProvider.Meta().(*datadog.Client)
		datapointUnixTime := float64(time.Now().Unix())
		datapointValue := float64(1)
		metric := datadog.Metric{
			Metric: datadog.String("foo"),
			Points: []datadog.DataPoint{{&datapointUnixTime, &datapointValue}},
		}
		if err := client.PostMetrics([]datadog.Metric{metric}); err != nil {
			return err
		}
		return nil
	}
}

const testAccCheckDatadogMetricMetadataConfig = `
resource "datadog_metric_metadata" "foo" {
  metric = "foo"
  short_name = "short name for metric_metadata foo"
  type = "gauge"
  description = "some description"
  unit = "byte"
  per_unit = "second"
  statsd_interval = "1"
}
`

const testAccCheckDatadogMetricMetadataConfigUpdated = `
resource "datadog_metric_metadata" "foo" {
  metric = "foo"
  short_name = "short name for metric_metadata foo"
  type = "gauge"
  description = "a different description"
  unit = "byte"
  per_unit = "second"
  statsd_interval = "1"
}
`
