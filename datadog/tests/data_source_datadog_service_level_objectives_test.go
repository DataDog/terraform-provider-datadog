package test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDatadogServiceLevelObjectivesDatasource(t *testing.T) {
	t.Parallel()
	ctx, accProviders := testAccProviders(context.Background(), t)
	firstSLOName := strings.ToLower(strings.ReplaceAll(uniqueEntityName(ctx, t), "-", "_"))
	secondSLOName := strings.ToLower(strings.ReplaceAll(uniqueEntityName(ctx, t), "-", "_"))
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		CheckDestroy:      testAccCheckDatadogServiceLevelObjectiveDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccDatasourceServiceLevelObjectivesIdsConfig(firstSLOName),
				Check:  checkServiceLevelObjectivesSingleResultDatasourceAttrs(accProvider, firstSLOName),
			},
			{
				Config: testAccDatasourceServiceLevelObjectivesNameFilterConfig(firstSLOName, secondSLOName),
				Check:  checkServiceLevelObjectivesSingleResultDatasourceAttrs(accProvider, firstSLOName),
			},
			{
				Config: testAccDatasourceServiceLevelObjectivesTagsFilterConfig(firstSLOName),
				Check:  checkServiceLevelObjectivesMultipleResultsDatasourceAttrs(accProvider, firstSLOName),
			},
			{
				Config: testAccDatasourceServiceLevelObjectivesMetricsFilterConfig(firstSLOName),
				Check:  checkServiceLevelObjectivesMultipleResultsDatasourceAttrs(accProvider, firstSLOName),
			},
		},
	})
}

func checkServiceLevelObjectivesSingleResultDatasourceAttrs(accProvider func() (*schema.Provider, error), uniq string) resource.TestCheckFunc {
	return resource.ComposeTestCheckFunc(
		resource.TestCheckResourceAttrSet("data.datadog_service_level_objectives.foo", "id"),
		resource.TestCheckResourceAttrSet("data.datadog_service_level_objectives.foo", "slos.0.id"),
		resource.TestCheckResourceAttr("data.datadog_service_level_objectives.foo", "slos.0.name", uniq),
		resource.TestCheckResourceAttr("data.datadog_service_level_objectives.foo", "slos.0.type", "metric"),
	)
}

func checkServiceLevelObjectivesMultipleResultsDatasourceAttrs(accProvider func() (*schema.Provider, error), uniq string) resource.TestCheckFunc {
	return resource.ComposeTestCheckFunc(
		resource.TestCheckResourceAttrSet("data.datadog_service_level_objectives.foo", "id"),
		resource.TestCheckResourceAttr("data.datadog_service_level_objectives.foo", "slos.#", "2"),
		resource.TestCheckResourceAttrSet("data.datadog_service_level_objectives.foo", "slos.0.id"),
		resource.TestCheckResourceAttr("data.datadog_service_level_objectives.foo", "slos.0.name", uniq),
		resource.TestCheckResourceAttr("data.datadog_service_level_objectives.foo", "slos.0.type", "metric"),
		resource.TestCheckResourceAttrSet("data.datadog_service_level_objectives.foo", "slos.1.id"),
		resource.TestCheckResourceAttr("data.datadog_service_level_objectives.foo", "slos.1.name", uniq),
		resource.TestCheckResourceAttr("data.datadog_service_level_objectives.foo", "slos.1.type", "metric"),
	)
}

func testAccCheckDatadogServiceLevelObjectiveUniqueTagMetricConfig(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_service_level_objective" "foo" {
  name = "%s"
  type = "metric"
  description = "some description about foo SLO"
  query {
	numerator = "sum:%s{type:good}.as_count()"
	denominator = "sum:my.metric{*}.as_count()"
  }

  target_threshold = 99.5
  timeframe = "7d"
  warning_threshold = 99.8

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

  tags = ["%s"]
}`, uniq, uniq, uniq)
}

func testAccDatasourceServiceLevelObjectivesIdsConfig(uniq string) string {
	return fmt.Sprintf(`
%s
data "datadog_service_level_objectives" "foo" {
  depends_on = [
    datadog_service_level_objective.foo,
  ]
  ids = [datadog_service_level_objective.foo.id]
}`,
		testAccCheckDatadogServiceLevelObjectiveUniqueTagMetricConfig(uniq),
	)
}

func testAccDatasourceServiceLevelObjectivesNameFilterConfig(firstSLOName string, secondSLOName string) string {
	return fmt.Sprintf(`
%s
%s
data "datadog_service_level_objectives" "foo" {
  depends_on = [
    datadog_service_level_objective.foo,
    datadog_service_level_objective.bar,
  ]
  name_query = "%s"
}
`,
		testAccCheckDatadogServiceLevelObjectiveUniqueTagMetricConfig(firstSLOName),
		strings.ReplaceAll(testAccCheckDatadogServiceLevelObjectiveUniqueTagMetricConfig(secondSLOName), "\"foo\"", "\"bar\""),
		firstSLOName,
	)
}

func testAccDatasourceServiceLevelObjectivesTagsFilterConfig(uniq string) string {
	return fmt.Sprintf(`
%s
%s
data "datadog_service_level_objectives" "foo" {
  depends_on = [
    datadog_service_level_objective.foo,
    datadog_service_level_objective.bar,
  ]
  tags_query = "%s"
}
`,
		testAccCheckDatadogServiceLevelObjectiveUniqueTagMetricConfig(uniq),
		strings.ReplaceAll(testAccCheckDatadogServiceLevelObjectiveUniqueTagMetricConfig(uniq), "\"foo\"", "\"bar\""),
		strings.ToLower(uniq),
	)
}

func testAccDatasourceServiceLevelObjectivesMetricsFilterConfig(uniq string) string {
	return fmt.Sprintf(`
%s
%s
data "datadog_service_level_objectives" "foo" {
  depends_on = [
    datadog_service_level_objective.foo,
    datadog_service_level_objective.bar,
  ]
  metrics_query = "%s"
}
`,
		testAccCheckDatadogServiceLevelObjectiveUniqueTagMetricConfig(uniq),
		strings.ReplaceAll(testAccCheckDatadogServiceLevelObjectiveUniqueTagMetricConfig(uniq), "\"foo\"", "\"bar\""),
		uniq,
	)
}
