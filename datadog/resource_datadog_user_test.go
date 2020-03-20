package datadog

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/DataDog/datadog-api-client-go/api/v1/datadog"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccDatadogUser_Updated(t *testing.T) {
	accProviders, cleanup := testAccProviders(t)
	defer cleanup(t)
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    accProviders,
		CheckDestroy: testAccCheckDatadogUserDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogUserConfigRequired,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogUserExists(accProvider),
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
					testAccCheckDatadogUserExists(accProvider),
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

func testAccCheckDatadogUserDestroy(accProvider *schema.Provider) func(*terraform.State) error {
	return func(s *terraform.State) error {
		providerConf := accProvider.Meta().(*ProviderConfiguration)
		client := providerConf.DatadogClientV1
		auth := providerConf.Auth

		if err := datadogUserDestroyHelper(auth, s, client); err != nil {
			return err
		}
		return nil
	}
}

func testAccCheckDatadogUserExists(accProvider *schema.Provider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		providerConf := accProvider.Meta().(*ProviderConfiguration)
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
		resp, _, err := client.UsersApi.GetUser(auth, id).Execute()
		u := resp.GetUser()

		if err != nil {
			if strings.Contains(err.Error(), "404 Not Found") {
				continue
			}
			return fmt.Errorf("received an error retrieving user %s", err)
		}

		// Datadog only disables user on DELETE
		if u.GetDisabled() {
			continue
		}
		return fmt.Errorf("user still exists")
	}
	return nil
}

func datadogUserExistsHelper(auth context.Context, s *terraform.State, client *datadog.APIClient) error {
	for _, r := range s.RootModule().Resources {
		id := r.Primary.ID
		if _, _, err := client.UsersApi.GetUser(auth, id).Execute(); err != nil {
			return fmt.Errorf("received an error retrieving user %s", err)
		}
	}
	return nil
}
