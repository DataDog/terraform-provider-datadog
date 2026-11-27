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
	t.Cleanup(datadog.ResetPermissionsCache())
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
			// {
			// 	Config: testAccCheckDatadogRoleConfigUpdated(rolename),
			// 	Check: resource.ComposeTestCheckFunc(
			// 		testAccCheckDatadogRoleExists(accProvider, "datadog_role.foo"),
			// 		resource.TestCheckResourceAttr("datadog_role.foo", "name", rolename),
			// 		testCheckRolePermission(
			// 			"datadog_role.foo",
			// 			"data.datadog_permissions.foo",
			// 			"permissions.logs_read_index_data",
			// 		),
			// 		testCheckRolePermission(
			// 			"datadog_role.foo",
			// 			"data.datadog_permissions.foo",
			// 			"permissions.standard",
			// 		),
			// 	),
			// },
			{
				Config: testAccCheckDatadogRoleConfigNoUnrestrictedPerm(rolename),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogRoleExists(accProvider, "datadog_role.foo"),
					resource.TestCheckResourceAttr("datadog_role.foo", "permission.#", "0"),
				),
			},
		},
	})

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		CheckDestroy:      testAccCheckDatadogRoleDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogRoleWithRestrictedPermsConfig(rolename),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogRoleExists(accProvider, "datadog_role.foo"),
					resource.TestCheckResourceAttr("datadog_role.foo", "name", rolename),
					testCheckRolePermission(
						"datadog_role.foo",
						"data.datadog_permissions.foo",
						"permissions.dashboards_read",
					),
					testCheckRolePermission(
						"datadog_role.foo",
						"data.datadog_permissions.foo",
						"permissions.monitors_read",
					),
					resource.TestCheckResourceAttr("datadog_role.foo", "permission.#", fmt.Sprint(2+len(implicitRolePermissions))),
				),
			},
			{
				Config: testAccCheckDatadogRoleConfigNoRestrictedPerm(rolename),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogRoleExists(accProvider, "datadog_role.foo"),
					resource.TestCheckResourceAttr("datadog_role.foo", "permission.#", fmt.Sprint(len(implicitRolePermissions))),
					testCheckRolePermissions("datadog_role.foo", "data.datadog_permissions.foo", implicitRolePermissions),
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
				Config:      testAccCheckDatadogRoleConfigInvalidNonexistentPerm(rolename),
				ExpectError: regexp.MustCompile("permission with ID .* does not exist"),
			},
			{
				Config:      testAccCheckDatadogRoleConfigInvalidRestrictedPerm(rolename),
				ExpectError: regexp.MustCompile("permission with ID .* is a restricted"),
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

func testCheckRolePermissions(rolename string, permissionsSource string, permissionNames []string) resource.TestCheckFunc {
	checks := make([]resource.TestCheckFunc, 0, len(permissionNames))
	for _, name := range permissionNames {
		checks = append(checks, testCheckRolePermission(rolename, permissionsSource, "permissions."+name))
	}
	return resource.ComposeTestCheckFunc(checks...)
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

// The API keeps these restricted permissions on every role, also with
// default_permissions_opt_out = true, and returns them in the role relationships.
// They are not removable until Minimal Access Roles is enabled on the org, so an
// opt-out config must list them to get an empty plan.
var implicitRolePermissions = []string{
	"appsec_vm_library_read",
	"built_in_features",
	"events_read",
	"hosts_read",
	"metrics_read",
	"static_analysis_settings_read",
	"timeseries_query",
	"user_self_profile_read",
	"user_self_profile_write",
}

func implicitRolePermissionsHCL() string {
	quoted := make([]string, len(implicitRolePermissions))
	for i, p := range implicitRolePermissions {
		quoted[i] = fmt.Sprintf("%q", p)
	}
	return "[" + strings.Join(quoted, ", ") + "]"
}

func testAccCheckDatadogRoleWithRestrictedPermsConfig(uniq string) string {
	return fmt.Sprintf(`
data "datadog_permissions" foo {
  include_restricted = true
}

resource "datadog_role" "foo" {
  name      = "%s"
  permission {
    id = "${data.datadog_permissions.foo.permissions.dashboards_read}"
  }
  permission {
    id = "${data.datadog_permissions.foo.permissions.monitors_read}"
  }
  dynamic "permission" {
    for_each = %s
    content {
      id = data.datadog_permissions.foo.permissions[permission.value]
    }
  }
  default_permissions_opt_out = true
}`, uniq, implicitRolePermissionsHCL())
}

func testAccCheckDatadogRoleConfigUpdated(uniq string) string {
	return fmt.Sprintf(`
data "datadog_permissions" foo {}

resource "datadog_role" "foo" {
  name      = "%s"
  permission {
    id = "${data.datadog_permissions.foo.permissions.logs_read_index_data}"
  }
  permission {
    id = "${data.datadog_permissions.foo.permissions.standard}"
  }
}`, uniq)
}

func testAccCheckDatadogRoleConfigNoUnrestrictedPerm(uniq string) string {
	return fmt.Sprintf(`
data "datadog_permissions" foo {}

resource "datadog_role" "foo" {
  name      = "%s"
}`, uniq)
}

func testAccCheckDatadogRoleConfigNoRestrictedPerm(uniq string) string {
	return fmt.Sprintf(`
data "datadog_permissions" foo {
  include_restricted = true
}

resource "datadog_role" "foo" {
  name      = "%s"
  dynamic "permission" {
    for_each = %s
    content {
      id = data.datadog_permissions.foo.permissions[permission.value]
    }
  }
  default_permissions_opt_out = true
}`, uniq, implicitRolePermissionsHCL())
}

func testAccCheckDatadogRoleConfigInvalidNonexistentPerm(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_role" "foo" {
  name      = "%sinvalid"
  permission {
    id = "invalid-id"
  }
}`, uniq)
}

func testAccCheckDatadogRoleConfigInvalidRestrictedPerm(uniq string) string {
	return fmt.Sprintf(`
data "datadog_permissions" foo {
  include_restricted = true
}

resource "datadog_role" "foo" {
  name      = "%s invalid restricted"
  permission {
    id = "${data.datadog_permissions.foo.permissions.dashboards_read}"
  }
}`, uniq)
}
