package datadog

import (
	"context"
	"fmt"
	"strings"
	"testing"

	datadogV2 "github.com/DataDog/datadog-api-client-go/api/v2/datadog"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/zorkian/go-datadog-api"
)

func TestAccDatadogUser_Updated(t *testing.T) {
	accProviders, clock, cleanup := testAccProviders(t, initRecorder(t))
	username := strings.ToLower(uniqueEntityName(clock, t)) + "@example.com"
	defer cleanup(t)
	accProvider := testAccProvider(t, accProviders)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    accProviders,
		CheckDestroy: testAccCheckDatadogUserV2Destroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogUserConfigRequired(username),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogUserV2Exists(accProvider, "datadog_user.foo"),
					resource.TestCheckResourceAttr(
						"datadog_user.foo", "email", username),
					resource.TestCheckResourceAttr(
						"datadog_user.foo", "name", "Test User"),
					resource.TestCheckResourceAttr(
						"datadog_user.foo", "verified", "false"),
				),
			},
			{
				Config: testAccCheckDatadogUserConfigUpdated(username),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogUserV2Exists(accProvider, "datadog_user.foo"),
					resource.TestCheckResourceAttr(
						"datadog_user.foo", "disabled", "true"),
					// NOTE: it's not possible ATM to update email of another user
					resource.TestCheckResourceAttr(
						"datadog_user.foo", "email", username),
					resource.TestCheckResourceAttr(
						"datadog_user.foo", "name", "Updated User"),
					resource.TestCheckResourceAttr(
						"datadog_user.foo", "verified", "false"),
				),
			},
		},
	})
}

func TestAccDatadogUser_Existing(t *testing.T) {
	accProviders, clock, cleanup := testAccProviders(t, initRecorder(t))
	username := strings.ToLower(uniqueEntityName(clock, t)) + "@example.com"
	defer cleanup(t)
	accProvider := testAccProvider(t, accProviders)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    accProviders,
		CheckDestroy: testAccCheckDatadogUserV2Destroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogUserConfigRequired(username),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogUserV2Exists(accProvider, "datadog_user.foo"),
					resource.TestCheckResourceAttr(
						"datadog_user.foo", "email", username),
					resource.TestCheckResourceAttr(
						"datadog_user.foo", "name", "Test User"),
					resource.TestCheckResourceAttr(
						"datadog_user.foo", "verified", "false"),
				),
			},
			{
				Config: testAccCheckDatadogUserConfigOtherUser(username),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogUserV2Exists(accProvider, "datadog_user.bar"),
					resource.TestCheckResourceAttr(
						"datadog_user.bar", "email", username),
					resource.TestCheckResourceAttr(
						"datadog_user.bar", "name", "Other User"),
					resource.TestCheckResourceAttr(
						"datadog_user.bar", "verified", "false"),
				),
			},
		},
	})
}

func TestAccDatadogUser_RoleDatasource(t *testing.T) {
	accProviders, clock, cleanup := testAccProviders(t, initRecorder(t))
	username := strings.ToLower(uniqueEntityName(clock, t)) + "@example.com"
	defer cleanup(t)
	accProvider := testAccProvider(t, accProviders)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    accProviders,
		CheckDestroy: testAccCheckDatadogUserDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogUserConfigReadOnlyRole(username),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogUserV2Exists(accProvider, "datadog_user.foo"),
					resource.TestCheckResourceAttr("datadog_user.foo", "email", username),
					resource.TestCheckResourceAttr("datadog_user.foo", "name", "Test User"),
					resource.TestCheckResourceAttr("datadog_user.foo", "verified", "false"),
					resource.TestCheckResourceAttr("datadog_user.foo", "roles.#", "1"),
					testCheckUserHasRole("datadog_user.foo", "data.datadog_role.ro_role"),
				),
			},
		},
	})
}

func TestAccDatadogUser_UpdateRole(t *testing.T) {
	accProviders, clock, cleanup := testAccProviders(t, initRecorder(t))
	username := strings.ToLower(uniqueEntityName(clock, t)) + "@example.com"
	defer cleanup(t)
	accProvider := testAccProvider(t, accProviders)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    accProviders,
		CheckDestroy: testAccCheckDatadogUserDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogUserConfigRoleUpdate1(username),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogUserV2Exists(accProvider, "datadog_user.foo"),
					resource.TestCheckResourceAttr("datadog_user.foo", "email", username),
					resource.TestCheckResourceAttr("datadog_user.foo", "name", "Test User"),
					resource.TestCheckResourceAttr("datadog_user.foo", "verified", "false"),
					resource.TestCheckResourceAttr("datadog_user.foo", "roles.#", "2"),
					testCheckUserHasRole("datadog_user.foo", "data.datadog_role.ro_role"),
					testCheckUserHasRole("datadog_user.foo", "data.datadog_role.st_role"),
				),
			},
			{
				Config: testAccCheckDatadogUserConfigRoleUpdate2(username),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogUserV2Exists(accProvider, "datadog_user.foo"),
					resource.TestCheckResourceAttr("datadog_user.foo", "email", username),
					resource.TestCheckResourceAttr("datadog_user.foo", "name", "Test User"),
					resource.TestCheckResourceAttr("datadog_user.foo", "verified", "false"),
					resource.TestCheckResourceAttr("datadog_user.foo", "roles.#", "2"),
					testCheckUserHasRole("datadog_user.foo", "data.datadog_role.adm_role"),
					testCheckUserHasRole("datadog_user.foo", "data.datadog_role.st_role"),
				),
			},
		},
	})
}

func testCheckUserHasRole(username string, roleSource string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rootModule := s.RootModule()
		roleID := rootModule.Resources[roleSource].Primary.Attributes["id"]
		roleIDHash := schema.HashSchema(&schema.Schema{Type: schema.TypeString})(roleID)

		return resource.TestCheckResourceAttr(username, fmt.Sprintf("roles.%d", roleIDHash), roleID)(s)
	}
}

func testAccCheckDatadogUserDestroy(accProvider *schema.Provider) func(*terraform.State) error {
	return func(s *terraform.State) error {
		providerConf := accProvider.Meta().(*ProviderConfiguration)
		client := providerConf.CommunityClient

		if err := datadogUserDestroyHelper(s, client); err != nil {
			return err
		}
		return nil
	}
}

func testAccCheckDatadogUserExists(accProvider *schema.Provider, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		providerConf := accProvider.Meta().(*ProviderConfiguration)
		client := providerConf.CommunityClient

		if err := datadogUserExistsHelper(s, client); err != nil {
			return err
		}
		return nil
	}
}

func testAccCheckDatadogUserV2Destroy(accProvider *schema.Provider) func(*terraform.State) error {
	return func(s *terraform.State) error {
		providerConf := accProvider.Meta().(*ProviderConfiguration)
		datadogClientV2 := providerConf.DatadogClientV2
		authV2 := providerConf.AuthV2

		if err := datadogUserV2DestroyHelper(s, datadogClientV2, authV2); err != nil {
			return err
		}
		return nil
	}
}

func testAccCheckDatadogUserV2Exists(accProvider *schema.Provider, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		providerConf := accProvider.Meta().(*ProviderConfiguration)
		datadogClientV2 := providerConf.DatadogClientV2
		authV2 := providerConf.AuthV2

		if err := datadogUserV2ExistsHelper(s, datadogClientV2, authV2, n); err != nil {
			return err
		}
		return nil
	}
}

func testAccCheckDatadogUserConfigRequired(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_user" "foo" {
  email     = "%s"
  name      = "Test User"
}`, uniq)
}

func testAccCheckDatadogUserConfigReadOnlyRole(uniq string) string {
	return fmt.Sprintf(`
data "datadog_role" "ro_role" {
  filter = "Datadog Read Only Role"
}

resource "datadog_user" "foo" {
  email     = "%s"
  name      = "Test User"
  roles     = [data.datadog_role.ro_role.id]
}`, uniq)
}

var roleDatasources = `
data "datadog_role" "ro_role" {
  filter = "Datadog Read Only Role"
}
data "datadog_role" "st_role" {
  filter = "Datadog Standard Role"
}
data "datadog_role" "adm_role" {
  filter = "Datadog Admin Role"
}`

func testAccCheckDatadogUserConfigRoleUpdate1(uniq string) string {
	return fmt.Sprintf(`%s

resource "datadog_user" "foo" {
  email     = "%s"
  name      = "Test User"
  roles     = [data.datadog_role.ro_role.id, data.datadog_role.st_role.id]
}`, roleDatasources, uniq)
}

func testAccCheckDatadogUserConfigRoleUpdate2(uniq string) string {
	return fmt.Sprintf(`%s

resource "datadog_user" "foo" {
  email     = "%s"
  name      = "Test User"
  roles     = [data.datadog_role.st_role.id, data.datadog_role.adm_role.id]
}`, roleDatasources, uniq)
}

func testAccCheckDatadogUserConfigUpdated(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_user" "foo" {
  disabled    = true
  // NOTE: it's not possible ATM to update email of another user
  email       = "%s"
  name        = "Updated User"
}`, uniq)
}

func testAccCheckDatadogUserConfigOtherUser(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_user" "bar" {
  email     = "%s"
  name      = "Other User"
}`, uniq)
}

func datadogUserDestroyHelper(s *terraform.State, client *datadog.Client) error {
	for _, r := range s.RootModule().Resources {
		id := r.Primary.ID
		u, err := client.GetUser(id)

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
		return fmt.Errorf("user still enabled")
	}
	return nil
}

func datadogUserExistsHelper(s *terraform.State, client *datadog.Client) error {
	for _, r := range s.RootModule().Resources {
		id := r.Primary.ID
		if _, err := client.GetUser(id); err != nil {
			return fmt.Errorf("received an error retrieving user %s", err)
		}
	}
	return nil
}

func datadogUserV2DestroyHelper(s *terraform.State, client *datadogV2.APIClient, auth context.Context) error {
	for _, r := range s.RootModule().Resources {
		id := r.Primary.ID
		userResponse, httpResponse, err := client.UsersApi.GetUser(auth, id).Execute()

		if err != nil {
			if httpResponse.StatusCode == 404 {
				continue
			}
			return fmt.Errorf("received an error retrieving user %s", err)
		}

		userData := userResponse.GetData()
		userAttributes := userData.GetAttributes()
		// Datadog only disables user on DELETE
		if userAttributes.GetDisabled() {
			continue
		}
		return fmt.Errorf("user still exists")
	}
	return nil
}

func datadogUserV2ExistsHelper(s *terraform.State, client *datadogV2.APIClient, auth context.Context, name string) error {
	id := s.RootModule().Resources[name].Primary.ID
	if _, _, err := client.UsersApi.GetUser(auth, id).Execute(); err != nil {
		return fmt.Errorf("received an error retrieving user %s", err)
	}
	return nil
}
