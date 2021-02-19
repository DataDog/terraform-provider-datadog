package test

import (
	"fmt"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
	"regexp"
	"strings"
	"testing"

	"github.com/terraform-providers/terraform-provider-datadog/datadog"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccDatadogRole_CreateUpdate(t *testing.T) {
	ctx, accProviders, clock, cleanup := testAccProviders(t, initRecorder(t))
	rolename := strings.ToLower(uniqueEntityName(clock, t))
	defer cleanup(t)
	accProvider := testAccProvider(t, accProviders)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(ctx, t) },
		Providers:    accProviders,
		CheckDestroy: testAccCheckDatadogRoleDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogRoleConfig(rolename),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogRoleExists(accProvider, "datadog_role.foo"),
					resource.TestCheckResourceAttr("datadog_role.foo", "name", rolename),
					resource.TestCheckResourceAttr("datadog_role.foo", "permission.#", "2"),
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
					resource.TestCheckResourceAttr("datadog_role.foo", "permission.#", "2"),
					testCheckRolePermission(
						"datadog_role.foo",
						"data.datadog_permissions.foo",
						"permissions.standard",
					),
					testCheckRolePermission(
						"datadog_role.foo",
						"data.datadog_permissions.foo",
						"permissions.logs_read_index_data",
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
	ctx, accProviders, clock, cleanup := testAccProviders(t, initRecorder(t))
	rolename := strings.ToLower(uniqueEntityName(clock, t))
	defer cleanup(t)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(ctx, t) },
		Providers: accProviders,
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
		perm := map[string]interface{}{
			"id": permissionID,
		}
		permissionIDHash := schema.HashResource(datadog.GetRolePermissionSchema())(perm)

		return resource.TestCheckResourceAttr(rolename, fmt.Sprintf("permission.%d.id", permissionIDHash), permissionID)(s)
	}
}

func testAccCheckDatadogRoleDestroy(accProvider *schema.Provider) func(*terraform.State) error {
	return func(s *terraform.State) error {
		providerConf := accProvider.Meta().(*datadog.ProviderConfiguration)
		client := providerConf.DatadogClientV2
		auth := providerConf.AuthV2

		for _, r := range s.RootModule().Resources {
			if r.Type != "datadog_role" {
				// Only care about roles
				continue
			}
			_, httpresp, err := client.RolesApi.GetRole(auth, r.Primary.ID).Execute()
			if err != nil {
				if !(httpresp != nil && httpresp.StatusCode == 404) {
					return utils.TranslateClientError(err, "error getting role")
				}
				// Role was successfully deleted
				continue
			}
			return fmt.Errorf("role %s still exists", r.Primary.ID)
		}
		return nil
	}
}

func testAccCheckDatadogRoleExists(accProvider *schema.Provider, rolename string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		providerConf := accProvider.Meta().(*datadog.ProviderConfiguration)
		client := providerConf.DatadogClientV2
		auth := providerConf.AuthV2

		id := s.RootModule().Resources[rolename].Primary.ID
		_, _, err := client.RolesApi.GetRole(auth, id).Execute()
		if err != nil {
			return utils.TranslateClientError(err, "error checking role existence")
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
