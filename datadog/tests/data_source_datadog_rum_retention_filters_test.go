package test

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// test application of DD Integration Tests org (321813) in us1.prod.dog
const RumRetentionFiltersDataSourceTestAppId = "5c2f3e68-adae-4f4e-9e91-7c31bb70329b"

func TestAccRumRetentionFiltersDatasource(t *testing.T) {
	if !isReplaying() {
		t.Skip("This test only supports replaying - requires specific RUM retention filter names in test environment")
	}
	t.Parallel()

	_, _, accProviders := testAccFrameworkMuxProviders(context.Background(), t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: accProviders,
		Steps: []resource.TestStep{
			{
				Config: dataDatadogRumRetentionFilters(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr("data.datadog_rum_retention_filters.testing_rum_retention_filters", "retention_filters.#", regexp.MustCompile(`3`)),
					resource.TestMatchResourceAttr("data.datadog_rum_retention_filters.testing_rum_retention_filters", "retention_filters.0.id", regexp.MustCompile(`default_replays`)),
					resource.TestMatchResourceAttr("data.datadog_rum_retention_filters.testing_rum_retention_filters", "retention_filters.0.name", regexp.MustCompile(`Sessions with replays`)),
					resource.TestMatchResourceAttr("data.datadog_rum_retention_filters.testing_rum_retention_filters", "retention_filters.1.id", regexp.MustCompile(`default_errors`)),
					resource.TestMatchResourceAttr("data.datadog_rum_retention_filters.testing_rum_retention_filters", "retention_filters.1.name", regexp.MustCompile(`Sessions with errors`)),
					resource.TestMatchResourceAttr("data.datadog_rum_retention_filters.testing_rum_retention_filters", "retention_filters.2.id", regexp.MustCompile(`default_sessions`)),
					resource.TestMatchResourceAttr("data.datadog_rum_retention_filters.testing_rum_retention_filters", "retention_filters.2.name", regexp.MustCompile(`Sessions`)),
				),
			},
		},
	})
}

func dataDatadogRumRetentionFilters() string {
	return fmt.Sprintf(`data "datadog_rum_retention_filters" "testing_rum_retention_filters" {
		application_id = %q
	}
	`, RumRetentionFiltersDataSourceTestAppId)
}
