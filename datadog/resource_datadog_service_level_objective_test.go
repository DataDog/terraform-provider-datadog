package datadog

import (
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/zorkian/go-datadog-api"
)

// config
const testAccCheckDatadogServiceLevelObjectiveConfig = `
resource "datadog_service_level_objective" "foo" {
  name = "name for metric SLO foo"
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
}
`

const testAccCheckDatadogServiceLevelObjectiveConfigUpdated = `
resource "datadog_service_level_objective" "foo" {
  name = "updated name for metric SLO foo"
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
`

// tests

func TestAccDatadogServiceLevelObjective_Basic(t *testing.T) {
	accProviders, cleanup := testAccProviders(t)
	defer cleanup(t)
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    accProviders,
		CheckDestroy: testAccCheckDatadogServiceLevelObjectiveDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogServiceLevelObjectiveConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogServiceLevelObjectiveExists(accProvider, "datadog_service_level_objective.foo"),
					resource.TestCheckResourceAttr(
						"datadog_service_level_objective.foo", "name", "name for metric SLO foo"),
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
				Config: testAccCheckDatadogServiceLevelObjectiveConfigUpdated,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogServiceLevelObjectiveExists(accProvider, "datadog_service_level_objective.foo"),
					resource.TestCheckResourceAttr(
						"datadog_service_level_objective.foo", "name", "updated name for metric SLO foo"),
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

// helpers

func destroyServiceLevelObjectiveHelper(s *terraform.State, client *datadog.Client) error {
	for _, r := range s.RootModule().Resources {
		if r.Primary.ID != "" {
			if _, err := client.GetServiceLevelObjective(r.Primary.ID); err != nil {
				if strings.Contains(strings.ToLower(err.Error()), "not found") {
					continue
				}
				return fmt.Errorf("Received an error retrieving service level objective %s", err)
			}
			return fmt.Errorf("Service Level Objective still exists")
		}
	}
	return nil
}

func existsServiceLevelObjectiveHelper(s *terraform.State, client *datadog.Client) error {
	for _, r := range s.RootModule().Resources {
		if _, err := client.GetServiceLevelObjective(r.Primary.ID); err != nil {
			return fmt.Errorf("Received an error retrieving service level objective %s", err)
		}
	}
	return nil
}

func testAccCheckDatadogServiceLevelObjectiveDestroy(accProvider *schema.Provider) func(*terraform.State) error {
	return func(s *terraform.State) error {
		client := accProvider.Meta().(*datadog.Client)

		if err := destroyServiceLevelObjectiveHelper(s, client); err != nil {
			return err
		}
		return nil
	}
}

func testAccCheckDatadogServiceLevelObjectiveExists(accProvider *schema.Provider, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client := accProvider.Meta().(*datadog.Client)
		if err := existsServiceLevelObjectiveHelper(s, client); err != nil {
			return err
		}
		return nil
	}
}
