package test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestServiceAccountCreate(t *testing.T) {
	t.Parallel()
	_, accProviders := testAccProviders(context.Background(), t)
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		CheckDestroy:      testAccCheckDatadogUserV2Destroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: generateServiceAccountConfig("some_linked_users_email@test.com", "Service account linked to some user", []string{"data.datadog_role.ro_role.id"}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"datadog_service_account.automated_test_service_account", "email", "some_linked_users_email@test.com"),
					resource.TestCheckResourceAttr(
						"datadog_service_account.automated_test_service_account", "name", "Service account linked to some user"),
					resource.TestCheckResourceAttr(
						"datadog_service_account.automated_test_service_account", "disabled", "false"),
					resource.TestCheckResourceAttr(
						"datadog_service_account.automated_test_service_account", "roles.#", "1"),
				),
			},
		},
	})
}

func TestServiceAccountUpdate(t *testing.T) {
	t.Parallel()
	_, accProviders := testAccProviders(context.Background(), t)
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		CheckDestroy:      testAccCheckDatadogUserV2Destroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: generateServiceAccountConfig("some_linked_users_email@test.com", "Service account linked to some user", []string{"data.datadog_role.ro_role.id"}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"datadog_service_account.automated_test_service_account", "email", "some_linked_users_email@test.com"),
					resource.TestCheckResourceAttr(
						"datadog_service_account.automated_test_service_account", "name", "Service account linked to some user"),
					resource.TestCheckResourceAttr(
						"datadog_service_account.automated_test_service_account", "disabled", "false"),
					resource.TestCheckResourceAttr(
						"datadog_service_account.automated_test_service_account", "roles.#", "1"),
				),
			},
			{
				Config: generateServiceAccountConfig("some_linked_users_email@test.com", "New name for the service account", []string{}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"datadog_service_account.automated_test_service_account", "email", "some_linked_users_email@test.com"),
					resource.TestCheckResourceAttr(
						"datadog_service_account.automated_test_service_account", "name", "New name for the service account"),
					resource.TestCheckResourceAttr(
						"datadog_service_account.automated_test_service_account", "disabled", "false"),
					resource.TestCheckResourceAttr(
						"datadog_service_account.automated_test_service_account", "roles.#", "0"),
				),
			},
		},
	})
}

// Generates a terraform config with a read only role and single service account
func generateServiceAccountConfig(email, name string, roles []string) string {
	return fmt.Sprintf(`
        data "datadog_role" "ro_role" {
          filter = "Datadog Read Only Role"
        }

        resource "datadog_service_account" "automated_test_service_account" {
          email = "%v"
          name = "%v"
          roles = [%v]
        }
  `, email, name, strings.Join(roles, ","))
}
