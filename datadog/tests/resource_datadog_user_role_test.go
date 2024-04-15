package test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/fwprovider"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

func TestAccUserRoleBasic(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	username := strings.ToLower(uniqueEntityName(ctx, t)) + "@example.com"

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogUserRoleDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogUserRole(username),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogUserRoleExists(providers.frameworkProvider),
				),
			},
		},
	})
}

func testAccCheckDatadogUserRole(username string) string {
	return fmt.Sprintf(`
data "datadog_role" "std_role" {
	filter = "Datadog Standard Role"
}

resource "datadog_user" "foo" {
	email = "%s"
}

# Create new user_role resource
resource "datadog_user_role" "foo" {
	role_id = data.datadog_role.std_role.id
	user_id = datadog_user.foo.id
}`, username)
}

func testAccCheckDatadogUserRoleDestroy(accProvider *fwprovider.FrameworkProvider) func(*terraform.State) error {
	return func(s *terraform.State) error {
		apiInstances := accProvider.DatadogApiInstances
		auth := accProvider.Auth

		if err := userRoleDestroyHelper(auth, s, apiInstances); err != nil {
			return err
		}
		return nil
	}
}

func userRoleDestroyHelper(auth context.Context, s *terraform.State, apiInstances *utils.ApiInstances) error {
	for _, r := range s.RootModule().Resources {
		if r.Type != "resource_datadog_user_role" {
			continue
		}
		roleId := r.Primary.Attributes["role_id"]
		userId := r.Primary.Attributes["user_id"]

		err := utils.Retry(2, 10, func() error {
			r, httpResp, err := apiInstances.GetRolesApiV2().ListRoleUsers(auth, roleId)
			if err != nil {
				if httpResp != nil && httpResp.StatusCode == 404 {
					return nil
				}
				return &utils.RetryableError{Prob: fmt.Sprintf("received an error retrieving UserRoles %s", err)}
			}
			for _, user := range r.Data {
				if user.GetId() == userId {
					return &utils.RetryableError{Prob: "UserRole still exists"}
				}
			}
			return nil
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func testAccCheckDatadogUserRoleExists(accProvider *fwprovider.FrameworkProvider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		apiInstances := accProvider.DatadogApiInstances
		auth := accProvider.Auth

		if err := userRoleExistsHelper(auth, s, apiInstances); err != nil {
			return err
		}
		return nil
	}
}

func userRoleExistsHelper(auth context.Context, s *terraform.State, apiInstances *utils.ApiInstances) error {
	for _, r := range s.RootModule().Resources {
		if r.Type != "resource_datadog_user_role" {
			continue
		}
		roleId := r.Primary.Attributes["role_id"]
		userId := r.Primary.Attributes["user_id"]
		r, httpResp, err := apiInstances.GetRolesApiV2().ListRoleUsers(auth, roleId)
		if err != nil {
			return utils.TranslateClientError(err, httpResp, "error retrieving UserRole")
		}
		for _, user := range r.Data {
			if user.GetId() == userId {
				return nil
			}
		}
		return utils.TranslateClientError(err, httpResp, "error retrieving UserRole")
	}
	return nil
}
