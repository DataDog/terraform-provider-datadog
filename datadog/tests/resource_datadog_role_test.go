package test

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/terraform-providers/terraform-provider-datadog/datadog"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccDatadogRole_CreateUpdate(t *testing.T) {
	ctx, accProviders := testAccProviders(context.Background(), t)
	rolename := strings.ToLower(uniqueEntityName(ctx, t))
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		CheckDestroy:      testAccCheckDatadogRoleDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogRoleConfig(rolename),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogRoleExists(accProvider, "datadog_role.foo"),
					resource.TestCheckResourceAttr("datadog_role.foo", "name", rolename),
					testCheckRolePermission(
						"datadog_role.foo",
						"data.datadog_permissions.foo",
						"permissions.admin",
					),
					testCheckRolePermission(
						"datadog_role.foo",
						"data.datadog_permissions.foo",
						"permissions.standard",
					),
				),
			},
			{
				Config: testAccCheckDatadogRoleConfigUpdated(rolename),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogRoleExists(accProvider, "datadog_role.foo"),
					resource.TestCheckResourceAttr("datadog_role.foo", "name", rolename+"updated"),
					testCheckRolePermission(
						"datadog_role.foo",
						"data.datadog_permissions.foo",
						"permissions.logs_read_index_data",
					),
					testCheckRolePermission(
						"datadog_role.foo",
						"data.datadog_permissions.foo",
						"permissions.standard",
					),
				),
			},
			{
				Config: testAccCheckDatadogRoleConfigNoPerm(rolename),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogRoleExists(accProvider, "datadog_role.foo"),
					resource.TestCheckResourceAttr("datadog_role.foo", "permission.#", "0"),
				),
			},
		},
	})
}
func TestAccDatadogRole_InvalidPerm(t *testing.T) {
	ctx, accProviders := testAccProviders(context.Background(), t)
	rolename := strings.ToLower(uniqueEntityName(ctx, t))

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		Steps: []resource.TestStep{
			{
				Config:      testAccCheckDatadogRoleConfigRestrictedPerm(rolename),
				ExpectError: regexp.MustCompile("permission with ID .* is restricted .* or does not exist"),
			},
		},
	})
}

func testCheckRolePermission(rolename string, permissionsSource string, permissionName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rootModule := s.RootModule()
		permissionID := rootModule.Resources[permissionsSource].Primary.Attributes[permissionName]

		return resource.TestCheckTypeSetElemNestedAttrs(rolename, "permission.*", map[string]string{
			"id": permissionID,
		})(s)
	}
}

func testAccCheckDatadogRoleDestroy(accProvider func() (*schema.Provider, error)) func(*terraform.State) error {
	return func(s *terraform.State) error {
		provider, _ := accProvider()
		providerConf := provider.Meta().(*datadog.ProviderConfiguration)
		apiInstances := providerConf.DatadogApiInstances
		auth := providerConf.Auth

		for _, r := range s.RootModule().Resources {
			if r.Type != "datadog_role" {
				// Only care about roles
				continue
			}
			_, httpresp, err := apiInstances.GetRolesApiV2().GetRole(auth, r.Primary.ID)
			if err != nil {
				if !(httpresp != nil && httpresp.StatusCode == 404) {
					return utils.TranslateClientError(err, httpresp, "error getting role")
				}
				// Role was successfully deleted
				continue
			}
			return fmt.Errorf("role %s still exists", r.Primary.ID)
		}
		return nil
	}
}

func testAccCheckDatadogRoleExists(accProvider func() (*schema.Provider, error), rolename string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		provider, _ := accProvider()
		providerConf := provider.Meta().(*datadog.ProviderConfiguration)
		apiInstances := providerConf.DatadogApiInstances
		auth := providerConf.Auth

		id := s.RootModule().Resources[rolename].Primary.ID
		_, httpresp, err := apiInstances.GetRolesApiV2().GetRole(auth, id)
		if err != nil {
			return utils.TranslateClientError(err, httpresp, "error checking role existence")
		}
		return nil
	}
}

func testAccCheckDatadogRoleConfig(uniq string) string {
	return fmt.Sprintf(`
data "datadog_permissions" foo {}

resource "datadog_role" "foo" {
  name      = "%s"
  permission {
    id = "${data.datadog_permissions.foo.permissions.standard}"
  }
  permission {
    id = "${data.datadog_permissions.foo.permissions.admin}"
  }
  permission {
    id = "${data.datadog_permissions.foo.permissions.org_management}"
  }
}`, uniq)
}

func testAccCheckDatadogRoleConfigUpdated(uniq string) string {
	return fmt.Sprintf(`
data "datadog_permissions" foo {}

resource "datadog_role" "foo" {
  name      = "%supdated"
  permission {
    id = "${data.datadog_permissions.foo.permissions.logs_read_index_data}"
  }
  permission {
    id = "${data.datadog_permissions.foo.permissions.standard}"
  }
}`, uniq)
}

func testAccCheckDatadogRoleConfigNoPerm(uniq string) string {
	return fmt.Sprintf(`
data "datadog_permissions" foo {}

resource "datadog_role" "foo" {
  name      = "%snoperm"
}`, uniq)
}

func testAccCheckDatadogRoleConfigRestrictedPerm(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_role" "foo" {
  name      = "%sinvalid"
  permission {
    id = "invalid-id"
  }
}`, uniq)
}
