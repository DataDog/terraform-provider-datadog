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

func testAccCheckDatadogServiceLevelObjectiveConfigDuplicateTag(uniq string) string {
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

  tags = ["foo:bar", "baz", "foo:double_bar"]
}`, uniq)
}

func testAccCheckDatadogServiceLevelObjectiveConfigNoTag(uniq string) string {
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
}`, uniq)
}

func testAccCheckDatadogServiceLevelObjectiveTimeSliceConfig(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_service_level_objective" "foo" {
  name = "%s"
  type = "time_slice"
  description = "some description about foo SLO"
  sli_specification {
    time_slice {
      query {
        formula {
          formula_expression = "(query1-query2)/query1"
        }
        query {
          metric_query {
            name        = "query1"
            query       = "sum:trace.grpc.server.hits{service:monitor-history-reader}"
		  }
        }
        query {
          metric_query {
            name        = "query2"
            data_source = "metrics"
            query       = "sum:trace.grpc.server.errors{service:monitor-history-reader}"
		  }
        }
      }
      comparator = ">"
      threshold  = 0.99
    }
  }

  thresholds {
	timeframe = "7d"
	target = 99
	warning = 99.5
  }

  timeframe = "7d"

  target_threshold = 99

  warning_threshold = 99.5

  tags = ["foo:bar", "baz"]
}`, uniq)
}

func testAccCheckDatadogServiceLevelObjectiveCountSpecConfig(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_service_level_objective" "foo" {
  name = "%s"
  type = "metric"
  description = "some description about foo SLO"
  sli_specification {
    count {
      good_events_formula  = "query1"
      total_events_formula = "query2"

      queries {
        metric_query {
          name        = "query1"
          data_source = "metrics"
          query       = "sum:my.metric{status:good}.as_count()"
        }
      }

      queries {
        metric_query {
          name        = "query2"
          data_source = "metrics"
          query       = "sum:my.metric{*}.as_count()"
        }
      }
    }
  }

  thresholds {
    timeframe = "7d"
    target    = 99.5
    warning   = 99.8
  }

  thresholds {
    timeframe = "30d"
    target    = 99
  }

  timeframe = "7d"

  target_threshold = 99.5

  warning_threshold = 99.8

  tags = ["foo:bar", "baz"]
}`, uniq)
}

func testAccCheckDatadogServiceLevelObjectiveCountSpecConfigUpdated(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_service_level_objective" "foo" {
  name = "%s"
  type = "metric"
  description = "updated description about foo SLO"
  sli_specification {
    count {
      good_events_formula  = "query1"
      total_events_formula = "query2"

      queries {
        metric_query {
          name        = "query1"
          data_source = "metrics"
          query       = "sum:my.metric{status:ok}.as_count()"
        }
      }

      queries {
        metric_query {
          name        = "query2"
          data_source = "metrics"
          query       = "sum:my.metric{*}.as_count()"
        }
      }
    }
  }

  thresholds {
    timeframe = "7d"
    target    = 99.5
    warning   = 99.8
  }

  thresholds {
    timeframe = "30d"
    target    = 98
    warning   = 99
  }

  timeframe = "7d"

  target_threshold = 99.5

  warning_threshold = 99.8

  tags = ["foo:bar", "baz"]
}`, uniq)
}

func testAccCheckDatadogServiceLevelObjectiveMetricTransitionQueryConfig(uniq string) string {
	return fmt.Sprintf(`
provider "datadog" {
  api_url  = "https://dd.datad0g.com"
  validate = false
}

resource "datadog_service_level_objective" "foo" {
  name = "%s"
  type = "metric"
  description = "transition test metric SLO using query"
  query {
    numerator   = "sum:my.metric{status:good}.as_count()"
    denominator = "sum:my.metric{*}.as_count()"
  }

  thresholds {
    timeframe = "7d"
    target    = 99.5
    warning   = 99.8
  }

  thresholds {
    timeframe = "30d"
    target    = 99
  }

  timeframe = "7d"

  target_threshold = 99.5

  warning_threshold = 99.8

  tags = ["foo:bar", "baz"]
}`, uniq)
}

func testAccCheckDatadogServiceLevelObjectiveMetricTransitionQueryConfigUpdated(uniq string) string {
	return fmt.Sprintf(`
provider "datadog" {
  api_url  = "https://dd.datad0g.com"
  validate = false
}

resource "datadog_service_level_objective" "foo" {
  name = "%s"
  type = "metric"
  description = "transition test metric SLO using updated query"
  query {
    numerator   = "sum:my.metric{status:good}.as_count() - sum:my.metric{status:bad}.as_count()"
    denominator = "sum:my.metric{*}.as_count()"
  }

  thresholds {
    timeframe = "7d"
    target    = 99.5
    warning   = 99.8
  }

  thresholds {
    timeframe = "30d"
    target    = 99
  }

  timeframe = "7d"

  target_threshold = 99.5

  warning_threshold = 99.8

  tags = ["foo:bar", "baz"]
}`, uniq)
}

func testAccCheckDatadogServiceLevelObjectiveMetricTransitionCountSpecConfig(uniq string) string {
	return fmt.Sprintf(`
provider "datadog" {
  api_url  = "https://dd.datad0g.com"
  validate = false
}

resource "datadog_service_level_objective" "foo" {
  name = "%s"
  type = "metric"
  description = "transition test metric SLO using count spec"
  sli_specification {
    count {
      good_events_formula  = "query1"
      total_events_formula = "query2"

      queries {
        metric_query {
          name        = "query1"
          data_source = "metrics"
          query       = "sum:my.metric{status:good}.as_count()"
        }
      }

      queries {
        metric_query {
          name        = "query2"
          data_source = "metrics"
          query       = "sum:my.metric{*}.as_count()"
        }
      }
    }
  }

  thresholds {
    timeframe = "7d"
    target    = 99.5
    warning   = 99.8
  }

  thresholds {
    timeframe = "30d"
    target    = 99
  }

  timeframe = "7d"

  target_threshold = 99.5

  warning_threshold = 99.8

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

func TestAccDatadogServiceLevelObjective_TimeSlice(t *testing.T) {
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
				Config: testAccCheckDatadogServiceLevelObjectiveTimeSliceConfig(sloName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogServiceLevelObjectiveExists(accProvider, "datadog_service_level_objective.foo"),
					resource.TestCheckResourceAttr(
						"datadog_service_level_objective.foo", "name", sloName),
					resource.TestCheckResourceAttr(
						"datadog_service_level_objective.foo", "description", "some description about foo SLO"),
					resource.TestCheckResourceAttr(
						"datadog_service_level_objective.foo", "type", "time_slice"),
					resource.TestCheckResourceAttr(
						"datadog_service_level_objective.foo", "sli_specification.#", "1"),
					resource.TestCheckResourceAttr(
						"datadog_service_level_objective.foo", "sli_specification.0.time_slice.#", "1"),
					resource.TestCheckResourceAttr(
						"datadog_service_level_objective.foo", "sli_specification.0.time_slice.0.query.#", "1"),
					resource.TestCheckResourceAttr(
						"datadog_service_level_objective.foo", "sli_specification.0.time_slice.0.query.0.formula.#", "1"),
					resource.TestCheckResourceAttr(
						"datadog_service_level_objective.foo", "sli_specification.0.time_slice.0.query.0.formula.0.formula_expression", "(query1-query2)/query1"),
					resource.TestCheckResourceAttr(
						"datadog_service_level_objective.foo", "sli_specification.0.time_slice.0.query.0.query.#", "2"),
					resource.TestCheckResourceAttr(
						"datadog_service_level_objective.foo", "sli_specification.0.time_slice.0.query.0.query.0.metric_query.#", "1"),
					resource.TestCheckResourceAttr(
						"datadog_service_level_objective.foo", "sli_specification.0.time_slice.0.query.0.query.0.metric_query.0.name", "query1"),
					resource.TestCheckResourceAttr(
						"datadog_service_level_objective.foo", "sli_specification.0.time_slice.0.query.0.query.0.metric_query.0.data_source", "metrics"),
					resource.TestCheckResourceAttr(
						"datadog_service_level_objective.foo", "sli_specification.0.time_slice.0.query.0.query.0.metric_query.0.query", "sum:trace.grpc.server.hits{service:monitor-history-reader}"),
					resource.TestCheckResourceAttr(
						"datadog_service_level_objective.foo", "sli_specification.0.time_slice.0.query.0.query.1.metric_query.0.name", "query2"),
					resource.TestCheckResourceAttr(
						"datadog_service_level_objective.foo", "sli_specification.0.time_slice.0.query.0.query.1.metric_query.0.data_source", "metrics"),
					resource.TestCheckResourceAttr(
						"datadog_service_level_objective.foo", "sli_specification.0.time_slice.0.query.0.query.1.metric_query.0.query", "sum:trace.grpc.server.errors{service:monitor-history-reader}"),
					resource.TestCheckResourceAttr(
						"datadog_service_level_objective.foo", "sli_specification.0.time_slice.0.comparator", ">"),
					resource.TestCheckResourceAttr(
						"datadog_service_level_objective.foo", "sli_specification.0.time_slice.0.threshold", "0.99"),
					// Thresholds are a TypeList, that are sorted by timeframe alphabetically.
					resource.TestCheckResourceAttr(
						"datadog_service_level_objective.foo", "thresholds.#", "1"),
					resource.TestCheckResourceAttr(
						"datadog_service_level_objective.foo", "thresholds.0.timeframe", "7d"),
					resource.TestCheckResourceAttr(
						"datadog_service_level_objective.foo", "thresholds.0.target", "99"),
					resource.TestCheckResourceAttr(
						"datadog_service_level_objective.foo", "thresholds.0.warning", "99.5"),
					resource.TestCheckResourceAttr(
						"datadog_service_level_objective.foo", "timeframe", "7d"),
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
		},
	})
}

func TestAccDatadogServiceLevelObjective_CountSpec(t *testing.T) {
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
				Config: testAccCheckDatadogServiceLevelObjectiveCountSpecConfig(sloName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogServiceLevelObjectiveExists(accProvider, "datadog_service_level_objective.foo"),
					resource.TestCheckResourceAttr(
						"datadog_service_level_objective.foo", "name", sloName),
					resource.TestCheckResourceAttr(
						"datadog_service_level_objective.foo", "description", "some description about foo SLO"),
					resource.TestCheckResourceAttr(
						"datadog_service_level_objective.foo", "type", "metric"),
					resource.TestCheckResourceAttr(
						"datadog_service_level_objective.foo", "sli_specification.#", "1"),
					resource.TestCheckResourceAttr(
						"datadog_service_level_objective.foo", "sli_specification.0.count.#", "1"),
					resource.TestCheckResourceAttr(
						"datadog_service_level_objective.foo", "sli_specification.0.count.0.good_events_formula", "query1"),
					resource.TestCheckResourceAttr(
						"datadog_service_level_objective.foo", "sli_specification.0.count.0.total_events_formula", "query2"),
					resource.TestCheckResourceAttr(
						"datadog_service_level_objective.foo", "sli_specification.0.count.0.queries.#", "2"),
					resource.TestCheckResourceAttr(
						"datadog_service_level_objective.foo", "sli_specification.0.count.0.queries.0.metric_query.#", "1"),
					resource.TestCheckResourceAttr(
						"datadog_service_level_objective.foo", "sli_specification.0.count.0.queries.0.metric_query.0.name", "query1"),
					resource.TestCheckResourceAttr(
						"datadog_service_level_objective.foo", "sli_specification.0.count.0.queries.0.metric_query.0.data_source", "metrics"),
					resource.TestCheckResourceAttr(
						"datadog_service_level_objective.foo", "sli_specification.0.count.0.queries.0.metric_query.0.query", "sum:my.metric{status:good}.as_count()"),
					resource.TestCheckResourceAttr(
						"datadog_service_level_objective.foo", "sli_specification.0.count.0.queries.1.metric_query.#", "1"),
					resource.TestCheckResourceAttr(
						"datadog_service_level_objective.foo", "sli_specification.0.count.0.queries.1.metric_query.0.name", "query2"),
					resource.TestCheckResourceAttr(
						"datadog_service_level_objective.foo", "sli_specification.0.count.0.queries.1.metric_query.0.data_source", "metrics"),
					resource.TestCheckResourceAttr(
						"datadog_service_level_objective.foo", "sli_specification.0.count.0.queries.1.metric_query.0.query", "sum:my.metric{*}.as_count()"),
					// Thresholds
					resource.TestCheckResourceAttr(
						"datadog_service_level_objective.foo", "thresholds.#", "2"),
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
			{
				Config: testAccCheckDatadogServiceLevelObjectiveCountSpecConfigUpdated(sloNameUpdated),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogServiceLevelObjectiveExists(accProvider, "datadog_service_level_objective.foo"),
					resource.TestCheckResourceAttr(
						"datadog_service_level_objective.foo", "name", sloNameUpdated),
					resource.TestCheckResourceAttr(
						"datadog_service_level_objective.foo", "description", "updated description about foo SLO"),
					resource.TestCheckResourceAttr(
						"datadog_service_level_objective.foo", "type", "metric"),
					resource.TestCheckResourceAttr(
						"datadog_service_level_objective.foo", "sli_specification.#", "1"),
					resource.TestCheckResourceAttr(
						"datadog_service_level_objective.foo", "sli_specification.0.count.#", "1"),
					resource.TestCheckResourceAttr(
						"datadog_service_level_objective.foo", "sli_specification.0.count.0.good_events_formula", "query1"),
					resource.TestCheckResourceAttr(
						"datadog_service_level_objective.foo", "sli_specification.0.count.0.total_events_formula", "query2"),
					resource.TestCheckResourceAttr(
						"datadog_service_level_objective.foo", "sli_specification.0.count.0.queries.#", "2"),
					resource.TestCheckResourceAttr(
						"datadog_service_level_objective.foo", "sli_specification.0.count.0.queries.0.metric_query.0.name", "query1"),
					resource.TestCheckResourceAttr(
						"datadog_service_level_objective.foo", "sli_specification.0.count.0.queries.0.metric_query.0.query", "sum:my.metric{status:ok}.as_count()"),
					resource.TestCheckResourceAttr(
						"datadog_service_level_objective.foo", "sli_specification.0.count.0.queries.1.metric_query.0.name", "query2"),
					resource.TestCheckResourceAttr(
						"datadog_service_level_objective.foo", "sli_specification.0.count.0.queries.1.metric_query.0.query", "sum:my.metric{*}.as_count()"),
					// Updated thresholds
					resource.TestCheckResourceAttr(
						"datadog_service_level_objective.foo", "thresholds.#", "2"),
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

func TestAccDatadogServiceLevelObjective_MetricQueryToCountTransition(t *testing.T) {
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
				Config: testAccCheckDatadogServiceLevelObjectiveMetricTransitionQueryConfig(sloName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogServiceLevelObjectiveExists(accProvider, "datadog_service_level_objective.foo"),
					resource.TestCheckResourceAttr(
						"datadog_service_level_objective.foo", "name", sloName),
					resource.TestCheckResourceAttr(
						"datadog_service_level_objective.foo", "query.#", "1"),
					resource.TestCheckResourceAttr(
						"datadog_service_level_objective.foo", "query.0.numerator", "sum:my.metric{status:good}.as_count()"),
					resource.TestCheckResourceAttr(
						"datadog_service_level_objective.foo", "query.0.denominator", "sum:my.metric{*}.as_count()"),
					resource.TestCheckNoResourceAttr(
						"datadog_service_level_objective.foo", "sli_specification.#"),
				),
			},
			{
				Config:   testAccCheckDatadogServiceLevelObjectiveMetricTransitionQueryConfig(sloName),
				PlanOnly: true,
			},
			{
				Config: testAccCheckDatadogServiceLevelObjectiveMetricTransitionQueryConfigUpdated(sloNameUpdated),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogServiceLevelObjectiveExists(accProvider, "datadog_service_level_objective.foo"),
					resource.TestCheckResourceAttr(
						"datadog_service_level_objective.foo", "name", sloNameUpdated),
					resource.TestCheckResourceAttr(
						"datadog_service_level_objective.foo", "query.#", "1"),
					resource.TestCheckResourceAttr(
						"datadog_service_level_objective.foo", "query.0.numerator", "sum:my.metric{status:good}.as_count() - sum:my.metric{status:bad}.as_count()"),
					resource.TestCheckResourceAttr(
						"datadog_service_level_objective.foo", "query.0.denominator", "sum:my.metric{*}.as_count()"),
					resource.TestCheckNoResourceAttr(
						"datadog_service_level_objective.foo", "sli_specification.#"),
				),
			},
			{
				Config:   testAccCheckDatadogServiceLevelObjectiveMetricTransitionQueryConfigUpdated(sloNameUpdated),
				PlanOnly: true,
			},
			{
				Config: testAccCheckDatadogServiceLevelObjectiveMetricTransitionCountSpecConfig(sloNameUpdated),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogServiceLevelObjectiveExists(accProvider, "datadog_service_level_objective.foo"),
					resource.TestCheckResourceAttr(
						"datadog_service_level_objective.foo", "name", sloNameUpdated),
					resource.TestCheckResourceAttr(
						"datadog_service_level_objective.foo", "sli_specification.#", "1"),
					resource.TestCheckResourceAttr(
						"datadog_service_level_objective.foo", "sli_specification.0.count.#", "1"),
					resource.TestCheckResourceAttr(
						"datadog_service_level_objective.foo", "sli_specification.0.count.0.good_events_formula", "query1"),
					resource.TestCheckResourceAttr(
						"datadog_service_level_objective.foo", "sli_specification.0.count.0.total_events_formula", "query2"),
					resource.TestCheckResourceAttr(
						"datadog_service_level_objective.foo", "sli_specification.0.count.0.queries.#", "2"),
					resource.TestCheckNoResourceAttr(
						"datadog_service_level_objective.foo", "query.#"),
				),
			},
			{
				Config:   testAccCheckDatadogServiceLevelObjectiveMetricTransitionCountSpecConfig(sloNameUpdated),
				PlanOnly: true,
			},
		},
	})
}

func TestAccDatadogServiceLevelObjective_CountSpecImport(t *testing.T) {
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
				Config: testAccCheckDatadogServiceLevelObjectiveMetricTransitionCountSpecConfig(sloName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogServiceLevelObjectiveExists(accProvider, "datadog_service_level_objective.foo"),
					resource.TestCheckResourceAttr(
						"datadog_service_level_objective.foo", "sli_specification.#", "1"),
					resource.TestCheckNoResourceAttr(
						"datadog_service_level_objective.foo", "query.#"),
				),
			},
			{
				ResourceName:      "datadog_service_level_objective.foo",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config:   testAccCheckDatadogServiceLevelObjectiveMetricTransitionCountSpecConfig(sloName),
				PlanOnly: true,
			},
		},
	})
}

func TestAccDatadogServiceLevelObjective_DefaultTags(t *testing.T) {
	t.Parallel()
	ctx, accProviders := testAccProviders(context.Background(), t)
	sloName := uniqueEntityName(ctx, t)
	accProvider := testAccProvider(t, accProviders)
	resource.Test(t, resource.TestCase{
		PreCheck: func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{ // New tags are correctly added and duplicates are kept
				Config: testAccCheckDatadogServiceLevelObjectiveConfigDuplicateTag(sloName),
				ProviderFactories: map[string]func() (*schema.Provider, error){
					"datadog": withDefaultTags(accProvider, map[string]interface{}{
						"default_key": "default_value",
					}),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"datadog_service_level_objective.foo", "tags.#", "4"),
					resource.TestCheckTypeSetElemAttr(
						"datadog_service_level_objective.foo", "tags.*", "baz"),
					resource.TestCheckTypeSetElemAttr(
						"datadog_service_level_objective.foo", "tags.*", "foo:bar"),
					resource.TestCheckTypeSetElemAttr(
						"datadog_service_level_objective.foo", "tags.*", "foo:double_bar"),
					resource.TestCheckTypeSetElemAttr(
						"datadog_service_level_objective.foo", "tags.*", "default_key:default_value"),
				),
			},
			{ // Resource tags take precedence over default tags and duplicates stay
				Config: testAccCheckDatadogServiceLevelObjectiveConfigDuplicateTag(sloName),
				ProviderFactories: map[string]func() (*schema.Provider, error){
					"datadog": withDefaultTags(accProvider, map[string]interface{}{
						"foo": "not_bar",
					}),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"datadog_service_level_objective.foo", "tags.#", "3"),
					resource.TestCheckTypeSetElemAttr(
						"datadog_service_level_objective.foo", "tags.*", "baz"),
					resource.TestCheckTypeSetElemAttr(
						"datadog_service_level_objective.foo", "tags.*", "foo:bar"),
					resource.TestCheckTypeSetElemAttr(
						"datadog_service_level_objective.foo", "tags.*", "foo:double_bar"),
				),
			},
			{ // Resource tags take precedence over default tags, but new tags are added
				Config: testAccCheckDatadogServiceLevelObjectiveConfig(sloName),
				ProviderFactories: map[string]func() (*schema.Provider, error){
					"datadog": withDefaultTags(accProvider, map[string]interface{}{
						"foo":     "not_bar",
						"new_tag": "new_value",
					}),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"datadog_service_level_objective.foo", "tags.#", "3"),
					resource.TestCheckTypeSetElemAttr(
						"datadog_service_level_objective.foo", "tags.*", "baz"),
					resource.TestCheckTypeSetElemAttr(
						"datadog_service_level_objective.foo", "tags.*", "foo:bar"),
					resource.TestCheckTypeSetElemAttr(
						"datadog_service_level_objective.foo", "tags.*", "new_tag:new_value"),
				),
			},
			{ // Tags without any value work correctly
				Config: testAccCheckDatadogServiceLevelObjectiveConfig(sloName),
				ProviderFactories: map[string]func() (*schema.Provider, error){
					"datadog": withDefaultTags(accProvider, map[string]interface{}{
						"no_value": "",
					}),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"datadog_service_level_objective.foo", "tags.#", "3"),
					resource.TestCheckTypeSetElemAttr(
						"datadog_service_level_objective.foo", "tags.*", "baz"),
					resource.TestCheckTypeSetElemAttr(
						"datadog_service_level_objective.foo", "tags.*", "foo:bar"),
					resource.TestCheckTypeSetElemAttr(
						"datadog_service_level_objective.foo", "tags.*", "no_value"),
				),
			},
			{ // Tags with colons in the value work correctly
				Config: testAccCheckDatadogServiceLevelObjectiveConfig(sloName),
				ProviderFactories: map[string]func() (*schema.Provider, error){
					"datadog": withDefaultTags(accProvider, map[string]interface{}{
						"repo_url": "https://github.com/repo/path",
					}),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"datadog_service_level_objective.foo", "tags.#", "3"),
					resource.TestCheckTypeSetElemAttr(
						"datadog_service_level_objective.foo", "tags.*", "baz"),
					resource.TestCheckTypeSetElemAttr(
						"datadog_service_level_objective.foo", "tags.*", "foo:bar"),
					resource.TestCheckTypeSetElemAttr(
						"datadog_service_level_objective.foo", "tags.*", "repo_url:https://github.com/repo/path"),
				),
			},
			{ // Works with monitors without a tag attribute
				Config: testAccCheckDatadogServiceLevelObjectiveConfigNoTag(sloName),
				ProviderFactories: map[string]func() (*schema.Provider, error){
					"datadog": withDefaultTags(accProvider, map[string]interface{}{
						"default_key": "default_value",
					}),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckTypeSetElemAttr(
						"datadog_service_level_objective.foo", "tags.*", "default_key:default_value"),
				),
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
	for _, r := range s.RootModule().Resources {
		if r.Type != "datadog_service_level_objective" {
			continue
		}

		err := utils.Retry(2, 5, func() error {
			if _, httpResp, err := apiInstances.GetServiceLevelObjectivesApiV1().GetSLO(ctx, r.Primary.ID); err != nil {
				if httpResp != nil && httpResp.StatusCode == 404 {
					return nil
				}
				return &utils.FatalError{Prob: fmt.Sprintf("received an error retrieving service level objective %s", err)}
			}
			return &utils.RetryableError{Prob: "service Level Objective still exists"}
		})

		if err != nil {
			return nil
		}

	}

	return nil
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
