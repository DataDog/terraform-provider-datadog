package test

import (
	"context"
	"fmt"
	"testing"

	"github.com/terraform-providers/terraform-provider-datadog/datadog"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccDatadogSloCorrection_Basic(t *testing.T) {
	t.Parallel()
	ctx, accProviders := testAccProviders(context.Background(), t)
	sloName := uniqueEntityName(ctx, t)
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		CheckDestroy:      testAccCheckDatadogSloCorrectionDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogSloCorrectionConfig(sloName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogSloCorrectionExists(accProvider, "datadog_slo_correction.testing_slo_correction"),
					resource.TestCheckResourceAttr(
						"datadog_slo_correction.testing_slo_correction", "description", "test correction on slo "+sloName),
					resource.TestCheckResourceAttr(
						"datadog_slo_correction.testing_slo_correction", "timezone", "UTC"),
					resource.TestCheckResourceAttr(
						"datadog_slo_correction.testing_slo_correction", "start", "1735707000"),
					resource.TestCheckResourceAttr(
						"datadog_slo_correction.testing_slo_correction", "end", "1735718600"),
					resource.TestCheckResourceAttr(
						"datadog_slo_correction.testing_slo_correction", "category", "Scheduled Maintenance"),
				),
			},
		},
	})
}

func TestAccDatadogSloCorrection_Updated(t *testing.T) {
	t.Parallel()
	ctx, accProviders := testAccProviders(context.Background(), t)
	sloName := uniqueEntityName(ctx, t)
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		CheckDestroy:      testAccCheckDatadogSloCorrectionDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogSloCorrectionConfig(sloName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogSloCorrectionExists(accProvider, "datadog_slo_correction.testing_slo_correction"),
					resource.TestCheckResourceAttr(
						"datadog_slo_correction.testing_slo_correction", "description", "test correction on slo "+sloName),
					resource.TestCheckResourceAttr(
						"datadog_slo_correction.testing_slo_correction", "timezone", "UTC"),
					resource.TestCheckResourceAttr(
						"datadog_slo_correction.testing_slo_correction", "start", "1735707000"),
					resource.TestCheckResourceAttr(
						"datadog_slo_correction.testing_slo_correction", "end", "1735718600"),
					resource.TestCheckResourceAttr(
						"datadog_slo_correction.testing_slo_correction", "category", "Scheduled Maintenance"),
				),
			},
			{
				Config: testAccCheckDatadogSloCorrectionConfigUpdated(sloName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogSloCorrectionExists(accProvider, "datadog_slo_correction.testing_slo_correction"),
					resource.TestCheckResourceAttr(
						"datadog_slo_correction.testing_slo_correction", "description", "updated test correction - "+sloName),
					resource.TestCheckResourceAttr(
						"datadog_slo_correction.testing_slo_correction", "timezone", "Africa/Lagos"),
					resource.TestCheckResourceAttr(
						"datadog_slo_correction.testing_slo_correction", "start", "1735707600"),
					resource.TestCheckResourceAttr(
						"datadog_slo_correction.testing_slo_correction", "end", "1735718000"),
					resource.TestCheckResourceAttr(
						"datadog_slo_correction.testing_slo_correction", "category", "Deployment"),
				),
			},
		},
	})
}

func testAccCheckDatadogSloCorrectionConfig(uniq string) string {
	return fmt.Sprintf(`
		resource "datadog_service_level_objective" "foo" {
			name = "%s"
			type = "metric"
			description = "some updated description about foo SLO"
			query {
			  numerator = "sum:my.metric{type:good}.as_count()"
			  denominator = "sum:my.metric{type:good}.as_count() + sum:my.metric{type:bad}.as_count()"
			}

			thresholds {
			  timeframe = "7d"
			  target = 99.5
			  warning = 99.8
			}

			thresholds {
			  timeframe = "30d"
			  target = 98
			  warning = 99.0
			}

			thresholds {
			  timeframe = "90d"
			  target = 99.9
			}

			tags = ["foo:bar", "baz"]
		  }
        resource "datadog_slo_correction" "testing_slo_correction" {
			category = "Scheduled Maintenance"
			description = "test correction on slo %s"
			end = 1735718600
			slo_id = datadog_service_level_objective.foo.id
			start = 1735707000
			timezone = "UTC"
        }
    `, uniq, uniq)
}

func testAccCheckDatadogSloCorrectionConfigUpdated(uniq string) string {
	return fmt.Sprintf(`
		resource "datadog_service_level_objective" "foo" {
			name = "%s"
			type = "metric"
			description = "some updated description about foo SLO"
			query {
			numerator = "sum:my.metric{type:good}.as_count()"
			denominator = "sum:my.metric{type:good}.as_count() + sum:my.metric{type:bad}.as_count()"
			}

			thresholds {
			timeframe = "7d"
			target = 99.5
			warning = 99.8
			}

			thresholds {
			timeframe = "30d"
			target = 98
			warning = 99.0
			}

			thresholds {
			timeframe = "90d"
			target = 99.9
			}

			tags = ["foo:bar", "baz"]
		}
        resource "datadog_slo_correction" "testing_slo_correction" {
			category = "Deployment"
			timezone = "Africa/Lagos"
			description = "updated test correction - %s"
			slo_id = datadog_service_level_objective.foo.id
			start = 1735707600
			end = 1735718000
        }
    `, uniq, uniq)
}

func testAccCheckDatadogSloCorrectionExists(accProvider func() (*schema.Provider, error), resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		provider, _ := accProvider()
		providerConf := provider.Meta().(*datadog.ProviderConfiguration)
		datadogClient := providerConf.DatadogClientV1
		auth := providerConf.AuthV1
		var err error

		for _, r := range s.RootModule().Resources {
			if r.Type != "datadog_slo_correction" {
				continue
			}
			id := r.Primary.ID
			if _, _, err = datadogClient.ServiceLevelObjectiveCorrectionsApi.GetSLOCorrection(auth, id).Execute(); err != nil {
				return utils.TranslateClientError(err, "error checking slo_correction existence")
			}
		}
		return nil
	}
}

func testAccCheckDatadogSloCorrectionDestroy(accProvider func() (*schema.Provider, error)) func(*terraform.State) error {
	return func(s *terraform.State) error {
		provider, _ := accProvider()
		providerConf := provider.Meta().(*datadog.ProviderConfiguration)
		datadogClient := providerConf.DatadogClientV1
		auth := providerConf.AuthV1
		for _, r := range s.RootModule().Resources {
			if r.Type != "datadog_slo_correction" {
				continue
			}

			var err error

			id := r.Primary.ID

			_, resp, err := datadogClient.ServiceLevelObjectiveCorrectionsApi.GetSLOCorrection(auth, id).Execute()

			if err != nil {
				if resp.StatusCode == 404 {
					continue // resource not found => all ok
				} else {
					return fmt.Errorf("received an error retrieving slo_correction: %s", err.Error())
				}
			} else {
				return fmt.Errorf("slo_correction %s still exists", r.Primary.ID)
			}
		}

		return nil
	}
}
