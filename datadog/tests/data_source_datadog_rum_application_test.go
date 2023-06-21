package test

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDatadogRUMApplicationDatasourceNameFilter(t *testing.T) {
	t.Parallel()
	ctx, accProviders := testAccProviders(context.Background(), t)
	uniq := uniqueEntityName(ctx, t)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDatasourceRUMApplicationNameFilterConfig(uniq),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.datadog_rum_application.bar", "name", uniq),
					resource.TestCheckResourceAttr(
						"data.datadog_rum_application.bar", "type", "browser"),
					resource.TestCheckResourceAttrPair(
						"data.datadog_rum_application.bar", "id", "datadog_rum_application.foo", "id"),
					resource.TestCheckResourceAttrPair(
						"data.datadog_rum_application.bar", "client_token", "datadog_rum_application.foo", "client_token"),
				),
			},
		},
	})
}

func TestAccDatadogRUMApplicationDatasourceIDFilter(t *testing.T) {
	t.Parallel()
	ctx, accProviders := testAccProviders(context.Background(), t)
	uniq := uniqueEntityName(ctx, t)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDatasourceRUMApplicationIDConfig(uniq),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.datadog_rum_application.bar", "name", uniq),
					resource.TestCheckResourceAttr(
						"data.datadog_rum_application.bar", "type", "browser"),
					resource.TestCheckResourceAttrPair(
						"data.datadog_rum_application.bar", "id", "datadog_rum_application.foo", "id"),
					resource.TestCheckResourceAttrPair(
						"data.datadog_rum_application.bar", "client_token", "datadog_rum_application.foo", "client_token"),
				),
			},
		},
	})
}

func TestAccDatadogRUMApplicationDatasourceErrorMultiple(t *testing.T) {
	t.Parallel()
	ctx, accProviders := testAccProviders(context.Background(), t)
	uniq := uniqueEntityName(ctx, t)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		Steps: []resource.TestStep{
			{
				Config:      testAccDatasourceRUMApplicationMultiple(uniq),
				ExpectError: regexp.MustCompile("more than one RUM application"),
			},
		},
	})
}

func testAccDatasourceRUMApplicationNameFilterConfig(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_rum_application" "foo" {
	name = "%[1]s"
	type = "browser"
}

resource "datadog_rum_application" "baz" {
	name = "%[1]s-extra"
	type = "android"
}

data "datadog_rum_application" "bar" {
	depends_on = [
		datadog_rum_application.foo,
		datadog_rum_application.baz,
	]
    name_filter = "%[1]s"
}`, uniq)
}

func testAccDatasourceRUMApplicationIDConfig(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_rum_application" "foo" {
	name = "%[1]s"
	type = "browser"
}

resource "datadog_rum_application" "baz" {
	name = "%[1]s"
	type = "android"
}

data "datadog_rum_application" "bar" {
	depends_on = [
		datadog_rum_application.foo,
		datadog_rum_application.baz,
	]
    id = datadog_rum_application.foo.id
}`, uniq)
}

func testAccDatasourceRUMApplicationMultiple(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_rum_application" "foo" {
	name = "%[1]s"
	type = "browser"
}

resource "datadog_rum_application" "baz" {
	name = "%[1]s"
	type = "android"
}

data "datadog_rum_application" "bar" {
	depends_on = [
		datadog_rum_application.foo,
		datadog_rum_application.baz,
	]
    name_filter = "%[1]s"
}`, uniq)
}
