package test

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/terraform-providers/terraform-provider-datadog/datadog"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
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
	warning = 99.5
  }

  thresholds {
	timeframe = "90d"
	target = 99
  }

  timeframe = "30d"

  target_threshold = 99

  warning_threshold = 99.5

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

  timeframe = "7d"

  target_threshold = 99.5

  warning_threshold = 99.8

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
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		CheckDestroy:      testAccCheckDatadogServiceLevelObjectiveDestroy(accProvider),
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
						"datadog_service_level_objective.foo", "thresholds.1.warning", "99.5"),
					resource.TestCheckResourceAttr(
						"datadog_service_level_objective.foo", "thresholds.2.timeframe", "90d"),
					resource.TestCheckResourceAttr(
						"datadog_service_level_objective.foo", "thresholds.2.target", "99"),
					resource.TestCheckResourceAttr(
						"datadog_service_level_objective.foo", "timeframe", "30d"),
					resource.TestCheckResourceAttr(
						"datadog_service_level_objective.foo", "target_threshold", "99"),
					resource.TestCheckResourceAttr(
						"datadog_service_level_objective.foo", "warning_threshold", "99.5"),
					resource.TestCheckResourceAttr(
						"datadog_service_level_objective.foo", "tags.#", "2"),
					resource.TestCheckTypeSetElemAttr(
						"datadog_service_level_objective.foo", "tags.*", "baz"),
					resource.TestCheckTypeSetElemAttr(
						"datadog_service_level_objective.foo", "tags.*", "foo:bar"),
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
					resource.TestCheckResourceAttr(
						"datadog_service_level_objective.foo", "timeframe", "7d"),
					resource.TestCheckResourceAttr(
						"datadog_service_level_objective.foo", "target_threshold", "99.5"),
					resource.TestCheckResourceAttr(
						"datadog_service_level_objective.foo", "warning_threshold", "99.8"),
					resource.TestCheckResourceAttr(
						"datadog_service_level_objective.foo", "tags.#", "2"),
					resource.TestCheckTypeSetElemAttr(
						"datadog_service_level_objective.foo", "tags.*", "baz"),
					resource.TestCheckTypeSetElemAttr(
						"datadog_service_level_objective.foo", "tags.*", "foo:bar"),
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
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		CheckDestroy:      testAccCheckDatadogServiceLevelObjectiveDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config:      testAccCheckDatadogServiceLevelObjectiveInvalidMonitorConfig(sloName),
				ExpectError: regexp.MustCompile("error finding monitor to add to SLO"),
			},
		},
	})
}

// helpers

func testAccCheckDatadogServiceLevelObjectiveDestroy(accProvider func() (*schema.Provider, error)) func(*terraform.State) error {
	return func(s *terraform.State) error {
		provider, _ := accProvider()
		providerConf := provider.Meta().(*datadog.ProviderConfiguration)
		apiInstances := providerConf.DatadogApiInstances
		auth := providerConf.Auth
		if err := destroyServiceLevelObjectiveHelper(auth, s, apiInstances); err != nil {
			return err
		}
		return nil
	}
}

func destroyServiceLevelObjectiveHelper(ctx context.Context, s *terraform.State, apiInstances *utils.ApiInstances) error {
	err := utils.Retry(2, 5, func() error {
		for _, r := range s.RootModule().Resources {
			if r.Primary.ID != "" {
				if _, httpResp, err := apiInstances.GetServiceLevelObjectivesApiV1().GetSLO(ctx, r.Primary.ID); err != nil {
					if httpResp != nil && httpResp.StatusCode == 404 {
						return nil
					}
					return &utils.FatalError{Prob: fmt.Sprintf("received an error retrieving service level objective %s", err)}
				}
				return &utils.RetryableError{Prob: "service Level Objective still exists"}
			}
		}
		return nil
	})
	return err
}

func existsServiceLevelObjectiveHelper(ctx context.Context, s *terraform.State, apiInstances *utils.ApiInstances) error {
	for _, r := range s.RootModule().Resources {
		if _, _, err := apiInstances.GetServiceLevelObjectivesApiV1().GetSLO(ctx, r.Primary.ID); err != nil {
			return fmt.Errorf("received an error retrieving service level objective %s", err)
		}
	}
	return nil
}

func testAccCheckDatadogServiceLevelObjectiveExists(accProvider func() (*schema.Provider, error), n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		provider, _ := accProvider()
		providerConf := provider.Meta().(*datadog.ProviderConfiguration)
		apiInstances := providerConf.DatadogApiInstances
		auth := providerConf.Auth

		if err := existsServiceLevelObjectiveHelper(auth, s, apiInstances); err != nil {
			return err
		}
		return nil
	}
}
