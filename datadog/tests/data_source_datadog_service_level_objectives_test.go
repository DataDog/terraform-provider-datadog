package test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func TestAccDatadogServiceLevelObjectivesDatasource(t *testing.T) {
	t.Parallel()
	ctx, accProviders := testAccProviders(context.Background(), t)
	firstSLOName := strings.ReplaceAll(uniqueEntityName(ctx, t), "-", "_")
	secondSLOName := strings.ReplaceAll(uniqueEntityName(ctx, t), "-", "_")
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    accProviders,
		CheckDestroy: testAccCheckDatadogServiceLevelObjectiveDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccDatasourceServiceLevelObjectivesIdsConfig(firstSLOName, secondSLOName),
				// Because of the `depends_on` in the datasource, the plan cannot be empty.
				// See https://www.terraform.io/docs/configuration/data-sources.html#data-resource-dependencies
				ExpectNonEmptyPlan: true,
				Check:              checkServiceLevelObjectivesMultipleResultsDatasourceAttrs(accProvider, firstSLOName, secondSLOName),
			},
			{
				Config: testAccDatasourceServiceLevelObjectivesNameFilterConfig(firstSLOName, secondSLOName),
				// Because of the `depends_on` in the datasource, the plan cannot be empty.
				// See https://www.terraform.io/docs/configuration/data-sources.html#data-resource-dependencies
				ExpectNonEmptyPlan: true,
				Check:              checkServiceLevelObjectivesSingleResultDatasourceAttrs(accProvider, firstSLOName),
			},
			{
				Config: testAccDatasourceServiceLevelObjectivesTagsFilterConfig(firstSLOName, secondSLOName),
				// Because of the `depends_on` in the datasource, the plan cannot be empty.
				// See https://www.terraform.io/docs/configuration/data-sources.html#data-resource-dependencies
				ExpectNonEmptyPlan: true,
				Check:              checkServiceLevelObjectivesMultipleResultsDatasourceAttrs(accProvider, firstSLOName, secondSLOName),
			},
			{
				Config: testAccDatasourceServiceLevelObjectivesMetricsFilterConfig(firstSLOName, secondSLOName),
				// Because of the `depends_on` in the datasource, the plan cannot be empty.
				// See https://www.terraform.io/docs/configuration/data-sources.html#data-resource-dependencies
				ExpectNonEmptyPlan: true,
				Check:              checkServiceLevelObjectivesMultipleResultsDatasourceAttrs(accProvider, firstSLOName, secondSLOName),
			},
		},
	})
}

func checkServiceLevelObjectivesSingleResultDatasourceAttrs(accProvider *schema.Provider, firstSLOName string) resource.TestCheckFunc {
	return resource.ComposeTestCheckFunc(
		resource.TestCheckResourceAttrSet("data.datadog_service_level_objectives.foo", "id"),
		resource.TestCheckResourceAttrSet("data.datadog_service_level_objectives.foo", "slos.0.id"),
		resource.TestCheckResourceAttr("data.datadog_service_level_objectives.foo", "slos.0.name", firstSLOName),
		resource.TestCheckResourceAttr("data.datadog_service_level_objectives.foo", "slos.0.type", "metric"),
	)
}

func checkServiceLevelObjectivesMultipleResultsDatasourceAttrs(accProvider *schema.Provider, firstSLOName string, secondSLOName string) resource.TestCheckFunc {
	return resource.ComposeTestCheckFunc(
		resource.TestCheckResourceAttrSet("data.datadog_service_level_objectives.foo", "id"),
		resource.TestCheckResourceAttrSet("data.datadog_service_level_objectives.foo", "slos.0.id"),
		resource.TestCheckResourceAttr("data.datadog_service_level_objectives.foo", "slos.0.name", firstSLOName),
		resource.TestCheckResourceAttr("data.datadog_service_level_objectives.foo", "slos.0.type", "metric"),
		resource.TestCheckResourceAttr("data.datadog_service_level_objectives.foo", "slos.1.name", secondSLOName),
		resource.TestCheckResourceAttr("data.datadog_service_level_objectives.foo", "slos.1.type", "metric"),
	)
}

func testAccDatasourceServiceLevelObjectivesIdsConfig(firstSLOName string, secondSLOName string) string {
	return fmt.Sprintf(`
%s
%s
data "datadog_service_level_objectives" "foo" {
  depends_on = [
    datadog_service_level_objective.foo,
    datadog_service_level_objective.bar,
  ]
  ids = [datadog_service_level_objective.foo.id, datadog_service_level_objective.bar.id]
}`,
		testAccCheckDatadogServiceLevelObjectiveConfig(firstSLOName),
		strings.ReplaceAll(testAccCheckDatadogServiceLevelObjectiveConfig(secondSLOName), "\"foo\"", "\"bar\""),
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
		testAccCheckDatadogServiceLevelObjectiveConfig(firstSLOName),
		strings.ReplaceAll(testAccCheckDatadogServiceLevelObjectiveConfig(secondSLOName), "\"foo\"", "\"bar\""),
		firstSLOName,
	)
}

func testAccDatasourceServiceLevelObjectivesTagsFilterConfig(firstSLOName string, secondSLOName string) string {
	return fmt.Sprintf(`
%s
%s
data "datadog_service_level_objectives" "foo" {
  depends_on = [
    datadog_service_level_objective.foo,
    datadog_service_level_objective.bar,
  ]
  tags_query = "foo:bar"
}
`,
		testAccCheckDatadogServiceLevelObjectiveConfig(firstSLOName),
		strings.ReplaceAll(testAccCheckDatadogServiceLevelObjectiveConfig(secondSLOName), "\"foo\"", "\"bar\""),
	)
}

func testAccDatasourceServiceLevelObjectivesMetricsFilterConfig(firstSLOName string, secondSLOName string) string {
	return fmt.Sprintf(`
%s
%s
data "datadog_service_level_objectives" "foo" {
  depends_on = [
    datadog_service_level_objective.foo,
    datadog_service_level_objective.bar,
  ]
  metrics_query = "my.metric"
}
`,
		testAccCheckDatadogServiceLevelObjectiveConfig(firstSLOName),
		strings.ReplaceAll(testAccCheckDatadogServiceLevelObjectiveConfig(secondSLOName), "\"foo\"", "\"bar\""),
	)
}
