package test

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/terraform-providers/terraform-provider-datadog/datadog"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"

	datadogV1 "github.com/DataDog/datadog-api-client-go/api/v1/datadog"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

// config
func testAccCheckDatadogServiceLevelObjectiveConfig(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_service_level_objective" "foo" {
  name = "%s"
  type = "metric"
  description = "some description about foo SLO"
  query {
	numerator = "sum:my.metric{type:good}.as_count()"
	denominator = "sum:my.metric{*}.as_count()"
  }

  thresholds {
	timeframe = "7d"
	target = 99.5
	warning = 99.8
  }

  thresholds {
	timeframe = "30d"
	target = 99
  }

  thresholds {
	timeframe = "90d"
	target = 99
  }

  tags = ["foo:bar", "baz"]
}`, uniq)
}

func testAccCheckDatadogServiceLevelObjectiveInvalidMonitorConfig(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_service_level_objective" "bar" {
	name               = "%s"
	type               = "monitor"
	description        = "My custom monitor SLO"
	monitor_ids = [1, 2, 3]
	validate = true
	thresholds {
	timeframe = "7d"
	target = 99.9
	warning = 99.99
	}
	
	thresholds {
	timeframe = "30d"
	target = 99.9
	warning = 99.99
	}
	
	tags = ["foo:bar", "baz"]
}`, uniq)
}

func testAccCheckDatadogServiceLevelObjectiveConfigUpdated(uniq string) string {
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
}`, uniq)
}

// tests

func TestAccDatadogServiceLevelObjective_Basic(t *testing.T) {
	t.Parallel()
	ctx, accProviders := testAccProviders(context.Background(), t)
	sloName := uniqueEntityName(ctx, t)
	sloNameUpdated := sloName + "-updated"
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    accProviders,
		CheckDestroy: testAccCheckDatadogServiceLevelObjectiveDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogServiceLevelObjectiveConfig(sloName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogServiceLevelObjectiveExists(accProvider, "datadog_service_level_objective.foo"),
					resource.TestCheckResourceAttr(
						"datadog_service_level_objective.foo", "name", sloName),
					resource.TestCheckResourceAttr(
						"datadog_service_level_objective.foo", "description", "some description about foo SLO"),
					resource.TestCheckResourceAttr(
						"datadog_service_level_objective.foo", "type", "metric"),
					resource.TestCheckResourceAttr(
						"datadog_service_level_objective.foo", "query.0.numerator", "sum:my.metric{type:good}.as_count()"),
					resource.TestCheckResourceAttr(
						"datadog_service_level_objective.foo", "query.0.denominator", "sum:my.metric{*}.as_count()"),
					// Thresholds are a TypeList, that are sorted by timeframe alphabetically.
					resource.TestCheckResourceAttr(
						"datadog_service_level_objective.foo", "thresholds.#", "3"),
					resource.TestCheckResourceAttr(
						"datadog_service_level_objective.foo", "thresholds.0.timeframe", "7d"),
					resource.TestCheckResourceAttr(
						"datadog_service_level_objective.foo", "thresholds.0.target", "99.5"),
					resource.TestCheckResourceAttr(
						"datadog_service_level_objective.foo", "thresholds.0.warning", "99.8"),
					resource.TestCheckResourceAttr(
						"datadog_service_level_objective.foo", "thresholds.1.timeframe", "30d"),
					resource.TestCheckResourceAttr(
						"datadog_service_level_objective.foo", "thresholds.1.target", "99"),
					resource.TestCheckResourceAttr(
						"datadog_service_level_objective.foo", "thresholds.2.timeframe", "90d"),
					resource.TestCheckResourceAttr(
						"datadog_service_level_objective.foo", "thresholds.2.target", "99"),
					// Tags are a TypeSet => use a weird way to access members by their hash
					// TF TypeSet is internally represented as a map that maps computed hashes
					// to actual values. Since the hashes are always the same for one value,
					// this is the way to get them.
					resource.TestCheckResourceAttr(
						"datadog_service_level_objective.foo", "tags.#", "2"),
					resource.TestCheckResourceAttr(
						"datadog_service_level_objective.foo", "tags.2644851163", "baz"),
					resource.TestCheckResourceAttr(
						"datadog_service_level_objective.foo", "tags.1750285118", "foo:bar"),
				),
			},
			{
				Config: testAccCheckDatadogServiceLevelObjectiveConfigUpdated(sloNameUpdated),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogServiceLevelObjectiveExists(accProvider, "datadog_service_level_objective.foo"),
					resource.TestCheckResourceAttr(
						"datadog_service_level_objective.foo", "name", sloNameUpdated),
					resource.TestCheckResourceAttr(
						"datadog_service_level_objective.foo", "description", "some updated description about foo SLO"),
					resource.TestCheckResourceAttr(
						"datadog_service_level_objective.foo", "type", "metric"),
					resource.TestCheckResourceAttr(
						"datadog_service_level_objective.foo", "query.0.numerator", "sum:my.metric{type:good}.as_count()"),
					resource.TestCheckResourceAttr(
						"datadog_service_level_objective.foo", "query.0.denominator", "sum:my.metric{type:good}.as_count() + sum:my.metric{type:bad}.as_count()"),
					// Thresholds are a TypeList, that are sorted by timeframe alphabetically.
					resource.TestCheckResourceAttr(
						"datadog_service_level_objective.foo", "thresholds.#", "3"),
					resource.TestCheckResourceAttr(
						"datadog_service_level_objective.foo", "thresholds.0.timeframe", "7d"),
					resource.TestCheckResourceAttr(
						"datadog_service_level_objective.foo", "thresholds.0.target", "99.5"),
					resource.TestCheckResourceAttr(
						"datadog_service_level_objective.foo", "thresholds.0.warning", "99.8"),
					resource.TestCheckResourceAttr(
						"datadog_service_level_objective.foo", "thresholds.1.timeframe", "30d"),
					resource.TestCheckResourceAttr(
						"datadog_service_level_objective.foo", "thresholds.1.target", "98"),
					resource.TestCheckResourceAttr(
						"datadog_service_level_objective.foo", "thresholds.1.warning", "99"),
					resource.TestCheckResourceAttr(
						"datadog_service_level_objective.foo", "thresholds.2.timeframe", "90d"),
					resource.TestCheckResourceAttr(
						"datadog_service_level_objective.foo", "thresholds.2.target", "99.9"),
					// Tags are a TypeSet => use a weird way to access members by their hash
					// TF TypeSet is internally represented as a map that maps computed hashes
					// to actual values. Since the hashes are always the same for one value,
					// this is the way to get them.
					resource.TestCheckResourceAttr(
						"datadog_service_level_objective.foo", "tags.#", "2"),
					resource.TestCheckResourceAttr(
						"datadog_service_level_objective.foo", "tags.2644851163", "baz"),
					resource.TestCheckResourceAttr(
						"datadog_service_level_objective.foo", "tags.1750285118", "foo:bar"),
				),
			},
		},
	})
}

func TestAccDatadogServiceLevelObjective_InvalidMonitor(t *testing.T) {
	t.Parallel()
	ctx, accProviders := testAccProviders(context.Background(), t)
	sloName := uniqueEntityName(ctx, t)
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    accProviders,
		CheckDestroy: testAccCheckDatadogServiceLevelObjectiveDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config:      testAccCheckDatadogServiceLevelObjectiveInvalidMonitorConfig(sloName),
				ExpectError: regexp.MustCompile("error finding monitor to add to SLO"),
			},
		},
	})
}

// helpers

func testAccCheckDatadogServiceLevelObjectiveDestroy(accProvider *schema.Provider) func(*terraform.State) error {
	return func(s *terraform.State) error {
		providerConf := accProvider.Meta().(*datadog.ProviderConfiguration)
		datadogClientV1 := providerConf.DatadogClientV1
		authV1 := providerConf.AuthV1
		if err := destroyServiceLevelObjectiveHelper(authV1, s, datadogClientV1); err != nil {
			return err
		}
		return nil
	}
}

func destroyServiceLevelObjectiveHelper(ctx context.Context, s *terraform.State, datadogClientV1 *datadogV1.APIClient) error {
	err := utils.Retry(2, 5, func() error {
		for _, r := range s.RootModule().Resources {
			if r.Primary.ID != "" {
				if _, httpResp, err := datadogClientV1.ServiceLevelObjectivesApi.GetSLO(ctx, r.Primary.ID).Execute(); err != nil {
					if httpResp != nil && httpResp.StatusCode == 404 {
						return nil
					}
					return &utils.FatalError{Prob: fmt.Sprintf("received an error retrieving service level objective %s", err)}
				}
				return &utils.RetryableError{Prob: fmt.Sprintf("service Level Objective still exists")}
			}
		}
		return nil
	})
	return err
}

func existsServiceLevelObjectiveHelper(ctx context.Context, s *terraform.State, datadogClientV1 *datadogV1.APIClient) error {
	for _, r := range s.RootModule().Resources {
		if _, _, err := datadogClientV1.ServiceLevelObjectivesApi.GetSLO(ctx, r.Primary.ID).Execute(); err != nil {
			return fmt.Errorf("received an error retrieving service level objective %s", err)
		}
	}
	return nil
}

func testAccCheckDatadogServiceLevelObjectiveExists(accProvider *schema.Provider, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		providerConf := accProvider.Meta().(*datadog.ProviderConfiguration)
		datadogClientV1 := providerConf.DatadogClientV1
		authV1 := providerConf.AuthV1

		if err := existsServiceLevelObjectiveHelper(authV1, s, datadogClientV1); err != nil {
			return err
		}
		return nil
	}
}
