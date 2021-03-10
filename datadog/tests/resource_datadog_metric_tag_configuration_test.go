package test

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"

	"github.com/terraform-providers/terraform-provider-datadog/datadog"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccDatadogMetricTagConfiguration_import(t *testing.T) {
	resourceName := "datadog_metric_tag_configuration.testing_metric_tag_config"
	ctx, accProviders := testAccProviders(context.Background(), t)
	uniqueMetricTagConfig := strings.ReplaceAll(uniqueEntityName(ctx, t), "-", "_")
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    accProviders,
		CheckDestroy: testAccCheckDatadogMetricTagConfigurationDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogMetricTagConfigurationDistCreate(uniqueMetricTagConfig),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				ResourceName:      resourceName,
				ImportStateId:     resourceName + "123",
				ImportState:       true,
				ImportStateVerify: true,
				ExpectError:       regexp.MustCompile("Cannot import non-existent remote object*"),
			},
		},
	})
}

func TestAccDatadogMetricTagConfiguration_Basic(t *testing.T) {
	ctx, accProviders := testAccProviders(context.Background(), t)
	uniqueMetricTagConfig := strings.ReplaceAll(uniqueEntityName(ctx, t), "-", "_")
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    accProviders,
		CheckDestroy: testAccCheckDatadogMetricTagConfigurationDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogMetricTagConfigurationDistCreate(uniqueMetricTagConfig),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogMetricTagConfigurationExists(accProvider, "datadog_metric_tag_configuration.testing_metric_tag_config"),
					resource.TestCheckResourceAttr(
						"datadog_metric_tag_configuration.testing_metric_tag_config", "id", uniqueMetricTagConfig),
					resource.TestCheckResourceAttr(
						"datadog_metric_tag_configuration.testing_metric_tag_config", "metric_name", uniqueMetricTagConfig),
					resource.TestCheckResourceAttr(
						"datadog_metric_tag_configuration.testing_metric_tag_config", "metric_type", "distribution"),
					resource.TestCheckResourceAttr(
						"datadog_metric_tag_configuration.testing_metric_tag_config", "tags.#", "2"),
					resource.TestCheckResourceAttr(
						"datadog_metric_tag_configuration.testing_metric_tag_config", "include_percentiles", "false"),
				),
			},
			{
				Config: testAccCheckDatadogMetricTagConfigurationDistUpdate(uniqueMetricTagConfig),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogMetricTagConfigurationExists(accProvider, "datadog_metric_tag_configuration.testing_metric_tag_config"),
					resource.TestCheckResourceAttr(
						"datadog_metric_tag_configuration.testing_metric_tag_config", "id", uniqueMetricTagConfig),
					resource.TestCheckResourceAttr(
						"datadog_metric_tag_configuration.testing_metric_tag_config", "metric_name", uniqueMetricTagConfig),
					resource.TestCheckResourceAttr(
						"datadog_metric_tag_configuration.testing_metric_tag_config", "metric_type", "distribution"),
					resource.TestCheckResourceAttr(
						"datadog_metric_tag_configuration.testing_metric_tag_config", "tags.#", "1"),
					resource.TestCheckResourceAttr(
						"datadog_metric_tag_configuration.testing_metric_tag_config", "include_percentiles", "true"),
				),
			},
		},
	})
}

func TestAccDatadogMetricTagConfiguration_Error(t *testing.T) {
	ctx, accProviders := testAccProviders(context.Background(), t)
	uniqueMetricTagConfig := strings.ReplaceAll(uniqueEntityName(ctx, t), "-", "_")
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    accProviders,
		CheckDestroy: testAccCheckDatadogMetricTagConfigurationDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config:      testAccCheckDatadogMetricTagConfigurationIncludePercentilesError(uniqueMetricTagConfig, "count"),
				ExpectError: regexp.MustCompile("cannot use include_percentiles with a metric_type of count*"),
			},
			{
				Config:      testAccCheckDatadogMetricTagConfigurationIncludePercentilesError(uniqueMetricTagConfig, "gauge"),
				ExpectError: regexp.MustCompile("cannot use include_percentiles with a metric_type of gauge*"),
			},
		},
	})
}

func testAccCheckDatadogMetricTagConfigurationDistCreate(uniq string) string {
	return fmt.Sprintf(`
        resource "datadog_metric_tag_configuration" "testing_metric_tag_config" {
			metric_name = "%s"
			metric_type = "distribution"
			tags = ["sport","datacenter"]
			include_percentiles = false
        }
    `, uniq)
}

func testAccCheckDatadogMetricTagConfigurationDistUpdate(uniq string) string {
	return fmt.Sprintf(`
        resource "datadog_metric_tag_configuration" "testing_metric_tag_config" {
			metric_name = "%s"
			metric_type = "distribution"
			tags = ["something_new"]
			include_percentiles = true
        }
    `, uniq)
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

func testAccCheckDatadogMetricTagConfigurationExists(accProvider *schema.Provider, resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		meta := accProvider.Meta()
		resourceID := s.RootModule().Resources[resourceName].Primary.ID
		providerConf := meta.(*datadog.ProviderConfiguration)
		datadogClient := providerConf.DatadogClientV2
		auth := providerConf.AuthV2

		id := resourceID

		_, httpresp, err := datadogClient.MetricsApi.ListTagConfigurationByName(auth, id).Execute()

		if err != nil {
			return utils.TranslateClientError(err, "error checking if tag configuration exists")
		}
		if httpresp == nil {
			return fmt.Errorf("unexpeted response when checking if tag configuration exists")
		}
		if httpresp.StatusCode != 200 {
			return fmt.Errorf("got unexpected status code when checking if tag config exists %d", httpresp.StatusCode)
		}

		return nil
	}
}

func testAccCheckDatadogMetricTagConfigurationDestroy(accProvider *schema.Provider) func(*terraform.State) error {
	return func(s *terraform.State) error {
		meta := accProvider.Meta()
		providerConf := meta.(*datadog.ProviderConfiguration)
		datadogClient := providerConf.DatadogClientV2
		auth := providerConf.AuthV2
		for _, r := range s.RootModule().Resources {
			if r.Type != "datadog_metric_tag_configuration" {
				continue
			}

			var err error

			id := r.Primary.ID

			_, resp, err := datadogClient.MetricsApi.ListTagConfigurationByName(auth, id).Execute()

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
