package test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/terraform-providers/terraform-provider-datadog/datadog"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccDatadogMetricMetadata_Basic(t *testing.T) {
	t.Parallel()
	ctx, accProviders := testAccProviders(context.Background(), t)
	accProvider := testAccProvider(t, accProviders)
	// A per-test metric: parallel tests and runs that share one metric overwrite each
	// other's metadata and cause drift on refresh.
	metric := strings.ReplaceAll(uniqueEntityName(ctx, t), "-", "_")

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		Steps: []resource.TestStep{
			{
				PreConfig: submitTestMetricPreConfig(t, metric, "gauge"),
				Config:    testAccCheckDatadogMetricMetadataConfig(metric, "some description"),
				Check: resource.ComposeTestCheckFunc(
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
	t.Parallel()
	ctx, accProviders := testAccProviders(context.Background(), t)
	accProvider := testAccProvider(t, accProviders)
	// A per-test metric: parallel tests and runs that share one metric overwrite each
	// other's metadata and cause drift on refresh.
	metric := strings.ReplaceAll(uniqueEntityName(ctx, t), "-", "_")

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		Steps: []resource.TestStep{
			{
				PreConfig: submitTestMetricPreConfig(t, metric, "gauge"),
				Config:    testAccCheckDatadogMetricMetadataConfig(metric, "some description"),
				Check: resource.ComposeTestCheckFunc(
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
				Config: testAccCheckDatadogMetricMetadataConfig(metric, "a different description"),
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

func metadataExistsHelper(ctx context.Context, s *terraform.State, apiInstances *utils.ApiInstances) error {
	for _, r := range s.RootModule().Resources {
		metric, ok := r.Primary.Attributes["metric"]
		if !ok {
			continue
		}

		_, httpresp, err := apiInstances.GetMetricsApiV1().GetMetricMetadata(ctx, metric)
		if err != nil {
			return utils.TranslateClientError(err, httpresp, "error retrieving metric_metadata")
		}
	}
	return nil
}

func checkMetricMetadataExists(accProvider func() (*schema.Provider, error)) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		provider, _ := accProvider()
		providerConf := provider.Meta().(*datadog.ProviderConfiguration)
		apiInstances := providerConf.DatadogApiInstances
		auth := providerConf.Auth

		if err := metadataExistsHelper(auth, s, apiInstances); err != nil {
			return err
		}
		return nil
	}
}

func testAccCheckDatadogMetricMetadataConfig(metric, description string) string {
	return fmt.Sprintf(`
resource "datadog_metric_metadata" "foo" {
  metric = "%s"
  short_name = "short name for metric_metadata foo"
  type = "gauge"
  description = "%s"
  unit = "byte"
  per_unit = "second"
  statsd_interval = "1"
}
`, metric, description)
}
