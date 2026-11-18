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
	t.Parallel()

	_, _, accProviders := testAccFrameworkMuxProviders(context.Background(), t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: accProviders,
		Steps: []resource.TestStep{
			{
				Config: dataDatadogRumRetentionFilters(),
				// Default filters differ between orgs and API versions (the replay filter
				// was renamed and is absent in some orgs), and their order is not
				// guaranteed. Only assert filters that are stable across orgs.
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr("data.datadog_rum_retention_filters.testing_rum_retention_filters", "retention_filters.#", regexp.MustCompile(`^[1-9][0-9]*$`)),
					resource.TestCheckTypeSetElemNestedAttrs("data.datadog_rum_retention_filters.testing_rum_retention_filters", "retention_filters.*", map[string]string{
						"id":   "default_errors",
						"name": "Sessions with errors",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("data.datadog_rum_retention_filters.testing_rum_retention_filters", "retention_filters.*", map[string]string{
						"id":   "default_sessions",
						"name": "Sessions",
					}),
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
