package test

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDatadogStandardPatternDatasourceNameFilter(t *testing.T) {
	t.Parallel()
	_, accProviders := testAccProviders(context.Background(), t)
	name := "AWS Access Key ID Scanner"

	datasource_name := "data.datadog_sensitive_data_scanner_standard_pattern.sample_sp"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDatasourceStandardPatternConfig(name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						datasource_name, "name", name),
				),
			},
		},
	})
}

func TestAccDatadogStandardPatternDatasourceErrorMultiple(t *testing.T) {
	t.Parallel()
	_, accProviders := testAccProviders(context.Background(), t)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		Steps: []resource.TestStep{
			{
				Config:      testAccDatasourceStandardPatternConfig("aws"),
				ExpectError: regexp.MustCompile("Your query returned more than one result, please try a more specific search criteria"),
			},
		},
	})
}

func TestAccDatadogStandardPatternDatasourceErrorNotFound(t *testing.T) {
	t.Parallel()
	_, accProviders := testAccProviders(context.Background(), t)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		Steps: []resource.TestStep{
			{
				Config:      testAccDatasourceStandardPatternConfig("foobarbaz"),
				ExpectError: regexp.MustCompile("Couldn't find the standard pattern with name foobarbaz"),
			},
		},
	})
}

func testAccDatasourceStandardPatternConfig(name string) string {
	return fmt.Sprintf(`
data "datadog_sensitive_data_scanner_standard_pattern" "sample_sp" {
  filter = "%s"
}`, name)
}
