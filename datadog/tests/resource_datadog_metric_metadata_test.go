package test

import (
	"context"
	"testing"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
	communityClient "github.com/zorkian/go-datadog-api"

	"github.com/terraform-providers/terraform-provider-datadog/datadog"

	"github.com/jonboulle/clockwork"

	datadogV1 "github.com/DataDog/datadog-api-client-go/api/v1/datadog"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccDatadogMetricMetadata_Basic(t *testing.T) {
	ctx, accProviders := testAccProviders(context.Background(), t)
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: accProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogMetricMetadataConfig,
				Check: resource.ComposeTestCheckFunc(
					checkPostEvent(accProvider, clockFromContext(ctx)),
					checkMetricMetadataExists(accProvider),
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
	ctx, accProviders := testAccProviders(context.Background(), t)
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: accProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogMetricMetadataConfig,
				Check: resource.ComposeTestCheckFunc(
					checkPostEvent(accProvider, clockFromContext(ctx)),
					checkMetricMetadataExists(accProvider),
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
					checkMetricMetadataExists(accProvider),
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

func metadataExistsHelper(ctx context.Context, s *terraform.State, datadogClientV1 *datadogV1.APIClient) error {
	for _, r := range s.RootModule().Resources {
		metric, ok := r.Primary.Attributes["metric"]
		if !ok {
			continue
		}

		_, _, err := datadogClientV1.MetricsApi.GetMetricMetadata(ctx, metric).Execute()
		if err != nil {
			return utils.TranslateClientError(err, "error retrieving metric_metadata")
		}
	}
	return nil
}

func checkMetricMetadataExists(accProvider *schema.Provider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		providerConf := accProvider.Meta().(*datadog.ProviderConfiguration)
		datadogClientV1 := providerConf.DatadogClientV1
		authV1 := providerConf.AuthV1

		if err := metadataExistsHelper(authV1, s, datadogClientV1); err != nil {
			return err
		}
		return nil
	}
}

func checkPostEvent(accProvider *schema.Provider, clock clockwork.FakeClock) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		providerConf := accProvider.Meta().(*datadog.ProviderConfiguration)
		client := providerConf.CommunityClient
		datapointUnixTime := float64(clock.Now().Unix())
		datapointValue := float64(1)
		metric := communityClient.Metric{
			Metric: communityClient.String("foo"),
			Points: []communityClient.DataPoint{{&datapointUnixTime, &datapointValue}},
		}
		if err := client.PostMetrics([]communityClient.Metric{metric}); err != nil {
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
