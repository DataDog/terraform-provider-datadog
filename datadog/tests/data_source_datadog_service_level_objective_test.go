package test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestAccDatadogServiceLevelObjectiveDatasource(t *testing.T) {
	t.Parallel()
	ctx, accProviders := testAccProviders(context.Background(), t)
	uniq := strings.ToLower(strings.ReplaceAll(uniqueEntityName(ctx, t), "-", "_"))
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		CheckDestroy:      testAccCheckDatadogServiceLevelObjectiveDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccDatasourceServiceLevelObjectiveIDConfig(uniq),
				Check:  checkServiceLevelObjectiveDatasourceAttrs(accProvider, uniq),
			},
			{
				Config: testAccDatasourceServiceLevelObjectiveNameFilterConfig(uniq),
				Check:  checkServiceLevelObjectiveDatasourceAttrs(accProvider, uniq),
			},
			{
				Config: testAccDatasourceServiceLevelObjectiveTagsFilterConfig(uniq),
				Check:  checkServiceLevelObjectiveDatasourceAttrs(accProvider, uniq),
			},
			{
				Config: testAccDatasourceServiceLevelObjectiveMetricsFilterConfig(uniq),
				Check:  checkServiceLevelObjectiveDatasourceAttrs(accProvider, uniq),
			},
		},
	})
}

func checkServiceLevelObjectiveDatasourceAttrs(accProvider func() (*schema.Provider, error), uniq string) resource.TestCheckFunc {
	return resource.ComposeTestCheckFunc(
		testAccCheckDatadogServiceLevelObjectiveExists(accProvider, ""),
		resource.TestCheckResourceAttr("data.datadog_service_level_objective.foo", "name", uniq),
		resource.TestCheckResourceAttr("data.datadog_service_level_objective.foo", "type", "metric"),
	)
}

func testAccDatasourceServiceLevelObjectiveIDConfig(uniq string) string {
	return fmt.Sprintf(`
%s
data "datadog_service_level_objective" "foo" {
  depends_on = [
    datadog_service_level_objective.foo,
  ]
  id = datadog_service_level_objective.foo.id
}`, testAccCheckDatadogServiceLevelObjectiveUniqueTagMetricConfig(uniq))
}

func testAccDatasourceServiceLevelObjectiveNameFilterConfig(uniq string) string {
	return fmt.Sprintf(`
%s
data "datadog_service_level_objective" "foo" {
  depends_on = [
    datadog_service_level_objective.foo,
  ]
  name_query = "%s"
}
`, testAccCheckDatadogServiceLevelObjectiveUniqueTagMetricConfig(uniq), uniq)
}

func testAccDatasourceServiceLevelObjectiveTagsFilterConfig(uniq string) string {
	return fmt.Sprintf(`
%s
data "datadog_service_level_objective" "foo" {
  depends_on = [
    datadog_service_level_objective.foo,
  ]
  tags_query = "%s"
}
`, testAccCheckDatadogServiceLevelObjectiveUniqueTagMetricConfig(uniq), strings.ToLower(uniq))
}

func testAccDatasourceServiceLevelObjectiveMetricsFilterConfig(uniq string) string {
	return fmt.Sprintf(`
%s
data "datadog_service_level_objective" "foo" {
  depends_on = [
    datadog_service_level_objective.foo,
  ]
  metrics_query = "%s"
}
`, testAccCheckDatadogServiceLevelObjectiveUniqueTagMetricConfig(uniq), uniq)
}
