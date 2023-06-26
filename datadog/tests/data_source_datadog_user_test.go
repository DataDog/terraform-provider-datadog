package test

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDatadogUserDatasourceExactMatch(t *testing.T) {
	t.Parallel()
	ctx, accProviders := testAccProviders(context.Background(), t)
	username := strings.ToLower(uniqueEntityName(ctx, t)) + "@example.com"
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		CheckDestroy:      testAccCheckDatadogUserV2Destroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccDatasourceUserConfig(username),
				Check:  resource.TestCheckResourceAttr("data.datadog_user.test", "email", username),
			},
		},
	})
}

func TestAccDatadogUserDatasourceError(t *testing.T) {
	t.Parallel()
	_, accProviders := testAccProviders(context.Background(), t)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		Steps: []resource.TestStep{
			{
				Config:      testAccDatasourceUserError(),
				ExpectError: regexp.MustCompile("didn't find any user matching"),
			},
		},
	})
}

func TestAccDatadogUserDatasourceWithExactMatch(t *testing.T) {
	t.Parallel()
	ctx, accProviders := testAccProviders(context.Background(), t)
	email := strings.ToLower(uniqueEntityName(ctx, t)) + "@example.com"
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		CheckDestroy:      testAccCheckDatadogUserV2Destroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccDatasourceUserWithExactMatchConfig(email, "true"),
				Check:  resource.TestCheckResourceAttr("data.datadog_user.test", "email", email),
			},
		},
	})
}

func TestAccDatadogUserDatasourceWithExactMatchError(t *testing.T) {
	t.Parallel()
	ctx, accProviders := testAccProviders(context.Background(), t)
	email := strings.ToLower(uniqueEntityName(ctx, t)) + "@example.com"
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		CheckDestroy:      testAccCheckDatadogUserV2Destroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config:      testAccDatasourceUserWithExactMatchConfig(email, "false"),
				ExpectError: regexp.MustCompile("your query returned more than one result for filter"),
			},
		},
	})
}

func testAccDatasourceUserWithExactMatchConfig(uniq, exactMatch string) string {
	return fmt.Sprintf(`
data "datadog_user" "test" {
	filter = "%[1]s"
	exact_match = %[2]s
	depends_on = [ datadog_user.foo, datadog_user.bar ]
}

resource "datadog_user" "foo" {
	email = "%[1]s"
}
resource "datadog_user" "bar" {
	email = "other%[1]s"
}`, uniq, exactMatch)
}

func testAccDatasourceUserConfig(uniq string) string {
	return fmt.Sprintf(`
data "datadog_user" "test" {
	filter = "%s"
	depends_on = [
		datadog_user.foo
	]
}
resource "datadog_user" "foo" {
    email = "%s"
}`, uniq, uniq)
}

func testAccDatasourceUserError() string {
	return `	
	data "datadog_user" "test" {
		filter = "doesntexist01b0bb82-2000-4113-bb6b-e34b48ef37ff@example.com'"
	}
`
}
