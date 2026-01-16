package test

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/terraform-providers/terraform-provider-datadog/datadog"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

// TestAccDatadogMetricTagConfiguration_GaugeBasic tests creating a gauge metric
// tag configuration without aggregations (the normal, expected usage since
// aggregations are deprecated).
func TestAccDatadogMetricTagConfiguration_GaugeBasic(t *testing.T) {
	t.Parallel()
	ctx, accProviders := testAccProviders(context.Background(), t)
	uniqueMetricTagConfig := strings.ReplaceAll(uniqueEntityName(ctx, t), "-", "_")
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		CheckDestroy:      testAccCheckDatadogMetricTagConfigurationDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogMetricTagConfigurationGaugeBasic(uniqueMetricTagConfig),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogMetricTagConfigurationExists(accProvider, "datadog_metric_tag_configuration.testing_metric_tag_config_gauge"),
					resource.TestCheckResourceAttr(
						"datadog_metric_tag_configuration.testing_metric_tag_config_gauge", "metric_name", uniqueMetricTagConfig),
					resource.TestCheckResourceAttr(
						"datadog_metric_tag_configuration.testing_metric_tag_config_gauge", "metric_type", "gauge"),
					resource.TestCheckResourceAttr(
						"datadog_metric_tag_configuration.testing_metric_tag_config_gauge", "tags.#", "2"),
					resource.TestCheckResourceAttr(
						"datadog_metric_tag_configuration.testing_metric_tag_config_gauge", "exclude_tags_mode", "false"),
					// Aggregations should be empty since the field is deprecated
					resource.TestCheckResourceAttr(
						"datadog_metric_tag_configuration.testing_metric_tag_config_gauge", "aggregations.#", "0"),
				),
			},
		},
	})
}

// TestAccDatadogMetricTagConfiguration_CountBasic tests creating a count metric
// tag configuration without aggregations.
func TestAccDatadogMetricTagConfiguration_CountBasic(t *testing.T) {
	t.Parallel()
	ctx, accProviders := testAccProviders(context.Background(), t)
	uniqueMetricTagConfig := strings.ReplaceAll(uniqueEntityName(ctx, t), "-", "_")
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		CheckDestroy:      testAccCheckDatadogMetricTagConfigurationDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogMetricTagConfigurationCountBasic(uniqueMetricTagConfig),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogMetricTagConfigurationExists(accProvider, "datadog_metric_tag_configuration.testing_metric_tag_config_count"),
					resource.TestCheckResourceAttr(
						"datadog_metric_tag_configuration.testing_metric_tag_config_count", "metric_name", uniqueMetricTagConfig),
					resource.TestCheckResourceAttr(
						"datadog_metric_tag_configuration.testing_metric_tag_config_count", "metric_type", "count"),
					resource.TestCheckResourceAttr(
						"datadog_metric_tag_configuration.testing_metric_tag_config_count", "tags.#", "1"),
					// Aggregations should be empty since the field is deprecated
					resource.TestCheckResourceAttr(
						"datadog_metric_tag_configuration.testing_metric_tag_config_count", "aggregations.#", "0"),
				),
			},
		},
	})
}

// TestAccDatadogMetricTagConfiguration_DistributionBasic tests creating a distribution
// metric tag configuration with include_percentiles.
func TestAccDatadogMetricTagConfiguration_DistributionBasic(t *testing.T) {
	t.Parallel()
	ctx, accProviders := testAccProviders(context.Background(), t)
	uniqueMetricTagConfig := strings.ReplaceAll(uniqueEntityName(ctx, t), "-", "_")
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		CheckDestroy:      testAccCheckDatadogMetricTagConfigurationDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogMetricTagConfigurationDistributionBasic(uniqueMetricTagConfig),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogMetricTagConfigurationExists(accProvider, "datadog_metric_tag_configuration.testing_metric_tag_config_distribution"),
					resource.TestCheckResourceAttr(
						"datadog_metric_tag_configuration.testing_metric_tag_config_distribution", "metric_name", uniqueMetricTagConfig),
					resource.TestCheckResourceAttr(
						"datadog_metric_tag_configuration.testing_metric_tag_config_distribution", "metric_type", "distribution"),
					resource.TestCheckResourceAttr(
						"datadog_metric_tag_configuration.testing_metric_tag_config_distribution", "include_percentiles", "true"),
				),
			},
		},
	})
}

// TestAccDatadogMetricTagConfiguration_GaugeUpdate tests updating a gauge metric
// tag configuration (e.g., changing tags).
func TestAccDatadogMetricTagConfiguration_GaugeUpdate(t *testing.T) {
	t.Parallel()
	ctx, accProviders := testAccProviders(context.Background(), t)
	uniqueMetricTagConfig := strings.ReplaceAll(uniqueEntityName(ctx, t), "-", "_")
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		CheckDestroy:      testAccCheckDatadogMetricTagConfigurationDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogMetricTagConfigurationGaugeBasic(uniqueMetricTagConfig),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogMetricTagConfigurationExists(accProvider, "datadog_metric_tag_configuration.testing_metric_tag_config_gauge"),
					resource.TestCheckResourceAttr(
						"datadog_metric_tag_configuration.testing_metric_tag_config_gauge", "tags.#", "2"),
				),
			},
			{
				Config: testAccCheckDatadogMetricTagConfigurationGaugeUpdated(uniqueMetricTagConfig),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogMetricTagConfigurationExists(accProvider, "datadog_metric_tag_configuration.testing_metric_tag_config_gauge"),
					resource.TestCheckResourceAttr(
						"datadog_metric_tag_configuration.testing_metric_tag_config_gauge", "tags.#", "3"),
				),
			},
		},
	})
}

// TestAccDatadogMetricTagConfiguration_Import tests importing an existing
// metric tag configuration resource.
func TestAccDatadogMetricTagConfiguration_Import(t *testing.T) {
	t.Parallel()
	resourceName := "datadog_metric_tag_configuration.testing_metric_tag_config_gauge"
	ctx, accProviders := testAccProviders(context.Background(), t)
	uniqueMetricTagConfig := strings.ReplaceAll(uniqueEntityName(ctx, t), "-", "_")
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		CheckDestroy:      testAccCheckDatadogMetricTagConfigurationDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogMetricTagConfigurationGaugeBasic(uniqueMetricTagConfig),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

// TestAccDatadogMetricTagConfiguration_Error tests various error conditions.
func TestAccDatadogMetricTagConfiguration_Error(t *testing.T) {
	t.Parallel()
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
				Config:      testAccCheckDatadogMetricTagConfigurationExcludeTagsModeError(uniqueMetricTagConfig),
				ExpectError: regexp.MustCompile("cannot use exclude_tags_mode without configuring any tags"),
			},
		},
	})
}

// TestAccDatadogMetricTagConfiguration_AggregationsDeprecatedError tests that
// using the deprecated aggregations field returns an error.
func TestAccDatadogMetricTagConfiguration_AggregationsDeprecatedError(t *testing.T) {
	t.Parallel()
	ctx, accProviders := testAccProviders(context.Background(), t)
	uniqueMetricTagConfig := strings.ReplaceAll(uniqueEntityName(ctx, t), "-", "_")
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		CheckDestroy:      testAccCheckDatadogMetricTagConfigurationDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config:      testAccCheckDatadogMetricTagConfigurationAggregationsDeprecated(uniqueMetricTagConfig, "gauge"),
				ExpectError: regexp.MustCompile("the 'aggregations' field is deprecated and no longer supported by the Datadog API"),
			},
			{
				Config:      testAccCheckDatadogMetricTagConfigurationAggregationsDeprecated(uniqueMetricTagConfig, "count"),
				ExpectError: regexp.MustCompile("the 'aggregations' field is deprecated and no longer supported by the Datadog API"),
			},
			{
				Config:      testAccCheckDatadogMetricTagConfigurationAggregationsDeprecated(uniqueMetricTagConfig, "rate"),
				ExpectError: regexp.MustCompile("the 'aggregations' field is deprecated and no longer supported by the Datadog API"),
			},
		},
	})
}

// TestAccDatadogMetricTagConfiguration_ExcludeTagsMode tests the exclude_tags_mode feature.
func TestAccDatadogMetricTagConfiguration_ExcludeTagsMode(t *testing.T) {
	t.Parallel()
	ctx, accProviders := testAccProviders(context.Background(), t)
	uniqueMetricTagConfig := strings.ReplaceAll(uniqueEntityName(ctx, t), "-", "_")
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		CheckDestroy:      testAccCheckDatadogMetricTagConfigurationDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogMetricTagConfigurationExcludeTagsMode(uniqueMetricTagConfig),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogMetricTagConfigurationExists(accProvider, "datadog_metric_tag_configuration.testing_metric_tag_config_exclude"),
					resource.TestCheckResourceAttr(
						"datadog_metric_tag_configuration.testing_metric_tag_config_exclude", "exclude_tags_mode", "true"),
					resource.TestCheckResourceAttr(
						"datadog_metric_tag_configuration.testing_metric_tag_config_exclude", "tags.#", "1"),
				),
			},
		},
	})
}

// Config helper functions

func testAccCheckDatadogMetricTagConfigurationGaugeBasic(uniq string) string {
	return fmt.Sprintf(`
		resource "datadog_metric_tag_configuration" "testing_metric_tag_config_gauge" {
			metric_name       = "%s"
			metric_type       = "gauge"
			tags              = ["sport", "datacenter"]
			exclude_tags_mode = false
		}
	`, uniq)
}

func testAccCheckDatadogMetricTagConfigurationGaugeUpdated(uniq string) string {
	return fmt.Sprintf(`
		resource "datadog_metric_tag_configuration" "testing_metric_tag_config_gauge" {
			metric_name       = "%s"
			metric_type       = "gauge"
			tags              = ["sport", "datacenter", "region"]
			exclude_tags_mode = false
		}
	`, uniq)
}

func testAccCheckDatadogMetricTagConfigurationCountBasic(uniq string) string {
	return fmt.Sprintf(`
		resource "datadog_metric_tag_configuration" "testing_metric_tag_config_count" {
			metric_name       = "%s"
			metric_type       = "count"
			tags              = ["env"]
			exclude_tags_mode = false
		}
	`, uniq)
}

func testAccCheckDatadogMetricTagConfigurationDistributionBasic(uniq string) string {
	return fmt.Sprintf(`
		resource "datadog_metric_tag_configuration" "testing_metric_tag_config_distribution" {
			metric_name         = "%s"
			metric_type         = "distribution"
			tags                = ["sport"]
			include_percentiles = true
		}
	`, uniq)
}

func testAccCheckDatadogMetricTagConfigurationIncludePercentilesError(uniq string, metricType string) string {
	return fmt.Sprintf(`
		resource "datadog_metric_tag_configuration" "testing_metric_tag_config_icl_count" {
			metric_name         = "%s"
			metric_type         = "%s"
			tags                = ["sport"]
			exclude_tags_mode   = false
			include_percentiles = false
		}
	`, uniq, metricType)
}

func testAccCheckDatadogMetricTagConfigurationAggregationsDeprecated(uniq string, metricType string) string {
	return fmt.Sprintf(`
		resource "datadog_metric_tag_configuration" "testing_metric_tag_config_aggregations" {
			metric_name = "%s"
			metric_type = "%s"
			tags        = ["sport"]
			aggregations {
				time  = "sum"
				space = "sum"
			}
		}
	`, uniq, metricType)
}

func testAccCheckDatadogMetricTagConfigurationExcludeTagsModeError(uniq string) string {
	return fmt.Sprintf(`
		resource "datadog_metric_tag_configuration" "testing_metric_tag_config_exclude_error" {
			metric_name       = "%s"
			metric_type       = "gauge"
			tags              = []
			exclude_tags_mode = true
		}
	`, uniq)
}

func testAccCheckDatadogMetricTagConfigurationExcludeTagsMode(uniq string) string {
	return fmt.Sprintf(`
		resource "datadog_metric_tag_configuration" "testing_metric_tag_config_exclude" {
			metric_name       = "%s"
			metric_type       = "gauge"
			tags              = ["env"]
			exclude_tags_mode = true
		}
	`, uniq)
}

// Check helper functions

func testAccCheckDatadogMetricTagConfigurationExists(accProvider func() (*schema.Provider, error), resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		provider, _ := accProvider()
		providerConf := provider.Meta().(*datadog.ProviderConfiguration)
		resourceID := s.RootModule().Resources[resourceName].Primary.ID
		apiInstances := providerConf.DatadogApiInstances
		auth := providerConf.Auth

		_, httpresp, err := apiInstances.GetMetricsApiV2().ListTagConfigurationByName(auth, resourceID)
		if err != nil {
			return utils.TranslateClientError(err, httpresp, "error checking metric_tag_configuration existence")
		}

		return nil
	}
}

func testAccCheckDatadogMetricTagConfigurationDestroy(accProvider func() (*schema.Provider, error)) func(*terraform.State) error {
	return func(s *terraform.State) error {
		provider, _ := accProvider()
		meta := provider.Meta()
		providerConf := meta.(*datadog.ProviderConfiguration)
		apiInstances := providerConf.DatadogApiInstances
		auth := providerConf.Auth
		for _, r := range s.RootModule().Resources {
			if r.Type != "datadog_metric_tag_configuration" {
				continue
			}

			var err error

			id := r.Primary.ID

			_, resp, err := apiInstances.GetMetricsApiV2().ListTagConfigurationByName(auth, id)

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
