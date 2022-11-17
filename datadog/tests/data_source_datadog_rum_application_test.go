package test

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestAccDatadogRUMApplicationDatasourceNameFilter(t *testing.T) {
	ctx, accProviders := testAccProviders(context.Background(), t)
	uniq := uniqueEntityName(ctx, t)
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDatasourceRUMApplicationNameFilterConfig(uniq),
				Check:  checkDatasourceRUMApplicationAttrs(accProvider, uniq),
			},
		},
	})
}

func TestAccDatadogRUMApplicationDatasourceIDFilter(t *testing.T) {
	ctx, accProviders := testAccProviders(context.Background(), t)
	uniq := uniqueEntityName(ctx, t)
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDatasourceRUMApplicationIDConfig(uniq),
				Check:  checkDatasourceRUMApplicationAttrs(accProvider, uniq),
			},
		},
	})
}

func TestAccDatadogRUMApplicationDatasourceErrorMultiple(t *testing.T) {
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

func checkDatasourceRUMApplicationAttrs(accProvider func() (*schema.Provider, error), uniq string) resource.TestCheckFunc {
	return resource.ComposeTestCheckFunc(
		resource.TestCheckResourceAttr(
			"data.datadog_rum_application.bar", "name", uniq),
		resource.TestCheckResourceAttr(
			"data.datadog_rum_application.bar", "type", "browser"),
		resource.TestCheckResourceAttrSet(
			"data.datadog_rum_application.bar", "id"),
		resource.TestCheckResourceAttrSet(
			"data.datadog_rum_application.bar", "client_token"),
	)
}

func testAccDatasourceRUMApplicationNameFilterConfig(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_rum_application" "foo" {
	name = "%[1]s"
	type = "browser"
}
data "datadog_rum_application" "bar" {
	depends_on = [
		datadog_rum_application.foo,
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
data "datadog_rum_application" "bar" {
	depends_on = [
		datadog_rum_application.foo,
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
resource "datadog_rum_application" "foobar" {
	name = "%[1]s"
	type = "react-native"
}
data "datadog_rum_application" "bar" {
	depends_on = [
		datadog_rum_application.foo,
		datadog_rum_application.foobar,
	]
    name_filter = "%[1]s"
}`, uniq, uniq)
}
