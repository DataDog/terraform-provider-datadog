package test

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDatadogUsersDatasourceFilter(t *testing.T) {
	t.Parallel()
	ctx, _, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	uniq := strings.ToLower(uniqueEntityName(ctx, t))
	email := uniq + "0@example.com"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: accProviders,
		Steps: []resource.TestStep{
			{
				// Step 1: Create users first to allow time for API indexing
				Config: testAccDatasourceUsersFilterConfigUsersOnly(uniq),
				Check:  resource.TestCheckResourceAttr("datadog_user.user_0", "email", email),
			},
			{
				// Step 2: Wait for API indexing, then add the data source lookup
				PreConfig: func() { time.Sleep(5 * time.Second) },
				Config:    testAccDatasourceUsersFilterConfig(uniq, email),
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
	ctx, _, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	uniq := strings.ToLower(uniqueEntityName(ctx, t))
	expectedUserName := "user 0"
	status := "Pending"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: accProviders,
		Steps: []resource.TestStep{
			{
				// Step 1: Create users first to allow time for API indexing
				Config: testAccDatasourceUsersFilterConfigUsersOnly(uniq),
				Check:  resource.TestCheckResourceAttr("datadog_user.user_0", "name", expectedUserName),
			},
			{
				// Step 2: Wait for API indexing, then add the data source lookup
				PreConfig: func() { time.Sleep(5 * time.Second) },
				Config:    testAccDatasourceUsersFilterStatusConfig(uniq, expectedUserName, status),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.datadog_users.all_users", "users.0.name", expectedUserName),
					checkRessourceAttributeRegex("data.datadog_users.all_users", "users.0.icon", "https://secure.gravatar.com/avatar/.*"),
				),
			},
		},
	})
}

func testAccDatasourceUsersFilterConfigUsersOnly(uniq string) string {
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
	return fmt.Sprintf(`
data "datadog_users" "all_users" {
	filter = "%[2]s"
	depends_on = [
		datadog_user.user_0,
		datadog_user.user_1,
		datadog_user.user_2
	]
}

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
`, uniq, filter)
}

func testAccDatasourceUsersFilterStatusConfig(uniq, filter, filterStatus string) string {
	return fmt.Sprintf(`
data "datadog_users" "all_users" {
	filter_status = "%[3]s"
	filter = "%[2]s"
	depends_on = [
		datadog_user.user_0,
		datadog_user.user_1,
		datadog_user.user_2
	]
}

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
`, uniq, filter, filterStatus)
}
