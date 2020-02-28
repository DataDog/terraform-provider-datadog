package datadog

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/DataDog/datadog-api-client-go/api/v1/datadog"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccDatadogUser_Updated(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckDatadogUserDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogUserConfigRequired,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogUserExists("datadog_user.foo"),
					resource.TestCheckResourceAttr(
						"datadog_user.foo", "email", "tftestuser@example.com"),
					resource.TestCheckResourceAttr(
						"datadog_user.foo", "handle", "tftestuser@example.com"),
					resource.TestCheckResourceAttr(
						"datadog_user.foo", "name", "Test User"),
					resource.TestCheckResourceAttr(
						"datadog_user.foo", "verified", "false"),
				),
			},
			{
				Config: testAccCheckDatadogUserConfigUpdated,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogUserExists("datadog_user.foo"),
					resource.TestCheckResourceAttr(
						"datadog_user.foo", "disabled", "true"),
					// NOTE: it's not possible ATM to update email of another user
					resource.TestCheckResourceAttr(
						"datadog_user.foo", "email", "tftestuser@example.com"),
					resource.TestCheckResourceAttr(
						"datadog_user.foo", "handle", "tftestuser@example.com"),
					resource.TestCheckResourceAttr(
						"datadog_user.foo", "is_admin", "true"),
					resource.TestCheckResourceAttr(
						"datadog_user.foo", "name", "Updated User"),
					resource.TestCheckResourceAttr(
						"datadog_user.foo", "verified", "false"),
					resource.TestCheckResourceAttr(
						"datadog_user.foo", "access_role", "adm"),
				),
			},
		},
	})
}

func testAccCheckDatadogUserDestroy(s *terraform.State) error {
	providerConf := testAccProvider.Meta().(*ProviderConfiguration)
	client := providerConf.DatadogClientV1
	auth := providerConf.Auth

	if err := datadogUserDestroyHelper(auth, s, client); err != nil {
		return err
	}
	return nil
}

func testAccCheckDatadogUserExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		providerConf := testAccProvider.Meta().(*ProviderConfiguration)
		client := providerConf.DatadogClientV1
		auth := providerConf.Auth
		if err := datadogUserExistsHelper(auth, s, client); err != nil {
			return err
		}
		return nil
	}
}

const testAccCheckDatadogUserConfigRequired = `
resource "datadog_user" "foo" {
  email     = "tftestuser@example.com"
  handle    = "tftestuser@example.com"
  name      = "Test User"
}
`

const testAccCheckDatadogUserConfigUpdated = `
resource "datadog_user" "foo" {
  disabled    = true
  // NOTE: it's not possible ATM to update email of another user
  email       = "tftestuser@example.com"
  handle      = "tftestuser@example.com"
  is_admin    = true
  access_role = "adm"
  name        = "Updated User"
}
`

func datadogUserDestroyHelper(auth context.Context, s *terraform.State, client *datadog.APIClient) error {
	for _, r := range s.RootModule().Resources {
		id := r.Primary.ID
		userResponse, _, err := client.UsersApi.GetUser(auth, id).Execute()
		u := userResponse.GetUser()

		if err != nil {
			if strings.Contains(err.Error(), "404 Not Found") {
				continue
			}
			return fmt.Errorf("Received an error retrieving user %s", err)
		}

		// Datadog only disables user on DELETE
		if u.GetDisabled() {
			continue
		}
		return fmt.Errorf("User still exists")
	}
	return nil
}

func datadogUserExistsHelper(auth context.Context, s *terraform.State, client *datadog.APIClient) error {
	for _, r := range s.RootModule().Resources {
		id := r.Primary.ID
		if _, _, err := client.UsersApi.GetUser(auth, id).Execute(); err != nil {
			return fmt.Errorf("Received an error retrieving user %s", err)
		}
	}
	return nil
}
