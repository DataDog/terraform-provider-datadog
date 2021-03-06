package test

import (
	"context"
	"fmt"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
	"regexp"
	"strings"
	"testing"

	"github.com/terraform-providers/terraform-provider-datadog/datadog"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccDatadogRole_CreateUpdate(t *testing.T) {
	t.Parallel()
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
					resource.TestCheckResourceAttr("datadog_role.foo", "permission.#", "2"),
					testCheckRolePermission(
						"datadog_role.foo",
						"data.datadog_permissions.foo",
						"permissions.admin",
						0,
					),
					testCheckRolePermission(
						"datadog_role.foo",
						"data.datadog_permissions.foo",
						"permissions.standard",
						1,
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
						"permissions.logs_read_index_data",
						0,
					),
					testCheckRolePermission(
						"datadog_role.foo",
						"data.datadog_permissions.foo",
						"permissions.standard",
						1,
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
	t.Parallel()
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

func testCheckRolePermission(rolename string, permissionsSource string, permissionName string, index int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rootModule := s.RootModule()
		permissionID := rootModule.Resources[permissionsSource].Primary.Attributes[permissionName]

		return resource.TestCheckResourceAttr(rolename, fmt.Sprintf("permission.%v.id", index), permissionID)(s)
	}
}

func testAccCheckDatadogRoleDestroy(accProvider func() (*schema.Provider, error)) func(*terraform.State) error {
	return func(s *terraform.State) error {
		provider, _ := accProvider()
		providerConf := provider.Meta().(*datadog.ProviderConfiguration)
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

func testAccCheckDatadogRoleExists(accProvider func() (*schema.Provider, error), rolename string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		provider, _ := accProvider()
		providerConf := provider.Meta().(*datadog.ProviderConfiguration)
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
