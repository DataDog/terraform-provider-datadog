package test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
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
				Config: fmt.Sprintf(`
        data "datadog_role" "ro_role" {
          filter = "Datadog Read Only Role"
        }

        resource "datadog_service_account" "some_test_service_account" {
          email = "some_linked_users_email@test.com"
          name = "Service account linked to some user"
          roles = [data.datadog_role.ro_role.id]
        }`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogUserV2Exists(accProvider, "datadog_user.foo"),
					resource.TestCheckResourceAttr(
						"datadog_service_account.some_test_service_account", "email", "some_linked_users_email@test.com"),
					resource.TestCheckResourceAttr(
						"datadog_service_account.some_test_service_account", "name", "Service account linked to some user"),
					resource.TestCheckResourceAttr(
						"datadog_service_account.some_test_service_account", "disabled", "false"),
					resource.TestCheckResourceAttr(
						"datadog_service_account.some_test_service_account", "roles", "[]"),
				),
			},
		},
	})
}
