package test

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/terraform-providers/terraform-provider-datadog/datadog"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/jonboulle/clockwork"
	communityClient "github.com/zorkian/go-datadog-api"
)

func TestAccDatadogMetricMetadataDatasource(t *testing.T) {
	t.Parallel()
	ctx, accProviders := testAccProviders(context.Background(), t)
	accProvider := testAccProvider(t, accProviders)
	_, _, fwAccProviders := testAccFrameworkMuxProviders(context.Background(), t)
	metricName := strings.ReplaceAll(uniqueEntityName(ctx, t), "-", "_")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: fwAccProviders,
		Steps: []resource.TestStep{
			{
				// Step 1: Create the metric resource first
				Config: testAccDatasourceMetricMetadataResourceConfig(metricName),
				Check: resource.ComposeTestCheckFunc(
					checkPostEventDatasource(accProvider, clockFromContext(ctx), metricName),
					resource.TestCheckResourceAttr("datadog_metric_metadata.test", "metric", metricName),
				),
			},
			{
				// Step 2: Wait for indexing, then read with datasource
				PreConfig: func() { time.Sleep(3 * time.Second) },
				Config:    testAccDatasourceMetricMetadataFullConfig(metricName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.datadog_metric_metadata.foo", "metric_name", metricName),
					resource.TestCheckResourceAttr("data.datadog_metric_metadata.foo", "short_name", "short name for metric_metadata foo"),
					resource.TestCheckResourceAttr("data.datadog_metric_metadata.foo", "description", "some description"),
					resource.TestCheckResourceAttr("data.datadog_metric_metadata.foo", "unit", "byte"),
					resource.TestCheckResourceAttr("data.datadog_metric_metadata.foo", "per_unit", "second"),
					resource.TestCheckResourceAttr("data.datadog_metric_metadata.foo", "statsd_interval", "1"),
					resource.TestCheckResourceAttr("data.datadog_metric_metadata.foo", "type", "gauge"),
				),
			},
		},
	})
}

func checkPostEventDatasource(accProvider func() (*schema.Provider, error), clock clockwork.FakeClock, metricName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		provider, _ := accProvider()
		providerConf := provider.Meta().(*datadog.ProviderConfiguration)
		client := providerConf.CommunityClient
		datapointUnixTime := float64(clock.Now().Unix())
		datapointValue := float64(1)
		metric := communityClient.Metric{
			Metric: communityClient.String(metricName),
			Points: []communityClient.DataPoint{{&datapointUnixTime, &datapointValue}},
		}
		if err := client.PostMetrics([]communityClient.Metric{metric}); err != nil {
			return err
		}
		return nil
	}
}

func testAccDatasourceMetricMetadataResourceConfig(metricName string) string {
	return fmt.Sprintf(`
resource "datadog_metric_metadata" "test" {
  metric = "%s"
  short_name = "short name for metric_metadata foo"
  type = "gauge"
  description = "some description"
  unit = "byte"
  per_unit = "second"
  statsd_interval = "1"
}
`, metricName)
}

func testAccDatasourceMetricMetadataFullConfig(metricName string) string {
	return fmt.Sprintf(`
resource "datadog_metric_metadata" "test" {
  metric = "%s"
  short_name = "short name for metric_metadata foo"
  type = "gauge"
  description = "some description"
  unit = "byte"
  per_unit = "second"
  statsd_interval = "1"
}

data "datadog_metric_metadata" "foo" {
  metric_name = "%s"
  depends_on = [datadog_metric_metadata.test]
}
`, metricName, metricName)
}
