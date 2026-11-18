package test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDatadogUsersDatasourceFilter(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	uniq := strings.ToLower(uniqueEntityName(ctx, t))
	email := uniq + "0@example.com"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: accProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDatasourceUsersResourcesConfig(uniq),
				Check:  testAccCheckUsersListed(frameworkAPIClient(providers.frameworkProvider), email, "", 1),
			},
			{
				Config: testAccDatasourceUsersFilterConfig(uniq, email),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.datadog_users.all_users", "users.0.email", email),
					checkRessourceAttributeRegex("data.datadog_users.all_users", "users.0.icon", "https://secure.gravatar.com/avatar/.*"),
					resource.TestCheckResourceAttr("data.datadog_users.all_users", "users.#", "1"),
				),
			},
			{
				Config: testAccDatasourceUsersFilterConfig(uniq, email),
				Check:  resource.TestCheckNoResourceAttr("data.datadog_users.all_users", "users.1.email"),
			},
		},
	})
}

func TestAccDatadogUsersDatasourceFilterStatus(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	uniq := strings.ToLower(uniqueEntityName(ctx, t))
	expectedUserName := "user 0"
	status := "Pending"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: accProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDatasourceUsersResourcesConfig(uniq),
				Check:  testAccCheckUsersListed(frameworkAPIClient(providers.frameworkProvider), uniq+"0@example.com", status, 1),
			},
			{
				Config: testAccDatasourceUsersFilterStatusConfig(uniq, expectedUserName, status),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.datadog_users.all_users", "users.0.name", expectedUserName),
					checkRessourceAttributeRegex("data.datadog_users.all_users", "users.0.icon", "https://secure.gravatar.com/avatar/.*"),
				),
			},
		},
	})
}

func testAccDatasourceUsersResourcesConfig(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_user" "user_0" {
	name = "user 0"
	email = "%[1]s0@example.com"
	send_user_invitation = false
}
resource "datadog_user" "user_1" {
	name = "user 1"
	email = "%[1]s1@example.com"
	send_user_invitation = false
}
resource "datadog_user" "user_2" {
	name = "user 2"
	email = "%[1]s2@example.com"
	send_user_invitation = false
}
`, uniq)
}

func testAccDatasourceUsersFilterConfig(uniq, filter string) string {
	return testAccDatasourceUsersResourcesConfig(uniq) + fmt.Sprintf(`
data "datadog_users" "all_users" {
	filter = "%[1]s"
	depends_on = [
		datadog_user.user_0,
		datadog_user.user_1,
		datadog_user.user_2
	]
}
`, filter)
}

func testAccDatasourceUsersFilterStatusConfig(uniq, filter, filterStatus string) string {
	return testAccDatasourceUsersResourcesConfig(uniq) + fmt.Sprintf(`
data "datadog_users" "all_users" {
	filter_status = "%[2]s"
	filter = "%[1]s"
	depends_on = [
		datadog_user.user_0,
		datadog_user.user_1,
		datadog_user.user_2
	]
}
`, filter, filterStatus)
}
