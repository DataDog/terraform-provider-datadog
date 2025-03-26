package test

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// test application of DD Integration Tests org (321813) in us1.prod.dog
const RumRetentionFiltersDataSourceTestAppId = "6d6b8210-efde-4a1d-bc4d-0f9773f21b0c"

func TestAccRumRetentionFiltersDatasource(t *testing.T) {
	t.Parallel()

	_, _, accProviders := testAccFrameworkMuxProviders(context.Background(), t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: accProviders,
		Steps: []resource.TestStep{
			{
				Config: dataDatadogRumRetentionFilters(),
				Check: resource.TestMatchResourceAttr("data.datadog_rum_retention_filters.testing_rum_retention_filters",
					"retention_filters.#", regexp.MustCompile(`^(0|[1-9]\d*)$`)), // check the size of list is >= 0
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
