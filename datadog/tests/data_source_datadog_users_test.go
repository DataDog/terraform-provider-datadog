package test

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDatadogUsersDatasourceBasic(t *testing.T) {
	t.Parallel()
	ctx, _, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	uniq := strings.ToLower(uniqueEntityName(ctx, t))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: accProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDatasourceUsersConfig(uniq),
				Check:  resource.TestCheckResourceAttr("data.datadog_users.all_users", "users.1.name", "user 1"),
			},
		},
	})
}

func TestAccDatadogUsersDatasourceExactMatch(t *testing.T) {
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

func TestAccDatadogUsersDatasourceError(t *testing.T) {
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

func TestAccDatadogUsersDatasourceWithExactMatch(t *testing.T) {
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

func TestAccDatadogUsersDatasourceWithExactMatchError(t *testing.T) {
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

func testAccDatasourceUsersConfig(uniq string) string {
	return fmt.Sprintf(`
data "datadog_users" "all_users" {
	filter = "%[1]s"
	depends_on = [
		datadog_user.user_0,
		datadog_user.user_1,
		datadog_user.user_2
	]
}

resource "datadog_user" "user_0" {
	name = "user 0"
	email = "%[1]s0@example.com"
}
resource "datadog_user" "user_1" {
	name = "user 1"
	email = "%[1]s1@example.com"
}
resource "datadog_user" "user_2" {
	name = "user 2"
	email = "%[1]s2@example.com"
}
`, uniq)
}

func testAccDatasourceUsersWithExactMatchConfig(uniq, exactMatch string) string {
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
