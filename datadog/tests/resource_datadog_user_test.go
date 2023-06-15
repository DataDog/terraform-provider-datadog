package test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/terraform-providers/terraform-provider-datadog/datadog"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"

	dd "github.com/DataDog/datadog-api-client-go/v2/api/datadog"
	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccDatadogUser_Updated(t *testing.T) {
	t.Parallel()
	ctx, accProviders := testAccProviders(context.Background(), t)
	username := strings.ToLower(uniqueEntityName(ctx, t)) + "@example.com"
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		CheckDestroy:      testAccCheckDatadogUserV2Destroy(accProvider),
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

func TestAccDatadogUser_Invitation(t *testing.T) {
	t.Parallel()
	ctx, accProviders := testAccProviders(context.Background(), t)
	username := strings.ToLower(uniqueEntityName(ctx, t)) + "@example.com"
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		CheckDestroy:      testAccCheckDatadogUserV2Destroy(accProvider),
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
					resource.TestCheckResourceAttrSet(
						"datadog_user.foo", "user_invitation_id"),
				),
			},
		},
	})
}

func TestAccDatadogUser_NoInvitation(t *testing.T) {
	t.Parallel()
	ctx, accProviders := testAccProviders(context.Background(), t)
	username := strings.ToLower(uniqueEntityName(ctx, t)) + "@example.com"
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		CheckDestroy:      testAccCheckDatadogUserV2Destroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogUserConfigRequiredNoInvitation(username),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogUserV2Exists(accProvider, "datadog_user.foo"),
					resource.TestCheckResourceAttr(
						"datadog_user.foo", "email", username),
					resource.TestCheckResourceAttr(
						"datadog_user.foo", "name", "Test User"),
					resource.TestCheckResourceAttr(
						"datadog_user.foo", "verified", "false"),
					resource.TestCheckNoResourceAttr(
						"datadog_user.foo", "user_invitation_id"),
				),
			},
		},
	})
}

func TestAccDatadogUser_Existing(t *testing.T) {
	t.Parallel()
	ctx, accProviders := testAccProviders(context.Background(), t)
	username := strings.ToLower(uniqueEntityName(ctx, t)) + "@example.com"
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		CheckDestroy:      testAccCheckDatadogUserV2Destroy(accProvider),
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
	t.Parallel()
	ctx, accProviders := testAccProviders(context.Background(), t)
	username := strings.ToLower(uniqueEntityName(ctx, t)) + "@example.com"
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		CheckDestroy:      testAccCheckDatadogUserV2Destroy(accProvider),
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
	t.Parallel()
	ctx, accProviders := testAccProviders(context.Background(), t)
	username := strings.ToLower(uniqueEntityName(ctx, t)) + "@example.com"
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		CheckDestroy:      testAccCheckDatadogUserV2Destroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogUserConfigRoleUpdate1(username),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogUserV2Exists(accProvider, "datadog_user.foo"),
					resource.TestCheckResourceAttr("datadog_user.foo", "email", username),
					resource.TestCheckResourceAttr("datadog_user.foo", "name", "Test User"),
					resource.TestCheckResourceAttr("datadog_user.foo", "verified", "false"),
					resource.TestCheckResourceAttr("datadog_user.foo", "roles.#", "2"),
					testCheckUserHasRole("datadog_user.foo", "data.datadog_role.st_role"),
					testCheckUserHasRole("datadog_user.foo", "data.datadog_role.ro_role"),
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
					testCheckUserHasRole("datadog_user.foo", "data.datadog_role.st_role"),
					testCheckUserHasRole("datadog_user.foo", "data.datadog_role.adm_role"),
				),
			},
		},
	})
}

func TestAccDatadogUser_ReEnableRoleUpdate(t *testing.T) {
	t.Parallel()
	ctx, accProviders := testAccProviders(context.Background(), t)
	username := strings.ToLower(uniqueEntityName(ctx, t)) + "@example.com"
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		CheckDestroy:      testAccCheckDatadogUserV2Destroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogUserConfigRoleUpdate1(username),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogUserV2Exists(accProvider, "datadog_user.foo"),
					resource.TestCheckResourceAttr("datadog_user.foo", "email", username),
					resource.TestCheckResourceAttr("datadog_user.foo", "name", "Test User"),
					resource.TestCheckResourceAttr("datadog_user.foo", "verified", "false"),
					resource.TestCheckResourceAttr("datadog_user.foo", "roles.#", "2"),
					testCheckUserHasRole("datadog_user.foo", "data.datadog_role.st_role"),
					testCheckUserHasRole("datadog_user.foo", "data.datadog_role.ro_role"),
				),
			},
			{
				// Destroy the user resource by passing data source resource only
				Config: roleDatasources,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserIsDisabled(accProvider, username),
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
					testCheckUserHasRole("datadog_user.foo", "data.datadog_role.st_role"),
					testCheckUserHasRole("datadog_user.foo", "data.datadog_role.adm_role"),
				),
			},
		},
	})
}

func testAccCheckUserIsDisabled(accProvider func() (*schema.Provider, error), username string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		provider, _ := accProvider()
		providerConf := provider.Meta().(*datadog.ProviderConfiguration)
		apiInstances := providerConf.DatadogApiInstances
		auth := providerConf.Auth

		resp, _, err := apiInstances.GetUsersApiV2().ListUsers(auth, datadogV2.ListUsersOptionalParameters{Filter: &username, FilterStatus: dd.PtrString("Disabled")})
		if err != nil {
			return fmt.Errorf("received an error listing users %s", err)
		}
		if len(resp.GetData()) == 0 {
			return fmt.Errorf("user is not disabled")
		}
		return nil
	}
}

func testCheckUserHasRole(username string, roleSource string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rootModule := s.RootModule()
		roleID := rootModule.Resources[roleSource].Primary.Attributes["id"]

		return resource.TestCheckTypeSetElemAttr(username, "roles.*", roleID)(s)
	}
}

func testAccCheckDatadogUserV2Destroy(accProvider func() (*schema.Provider, error)) func(*terraform.State) error {
	return func(s *terraform.State) error {
		provider, _ := accProvider()
		providerConf := provider.Meta().(*datadog.ProviderConfiguration)
		apiInstances := providerConf.DatadogApiInstances
		auth := providerConf.Auth

		if err := datadogUserV2DestroyHelper(auth, s, apiInstances); err != nil {
			return err
		}
		return nil
	}
}

func testAccCheckDatadogUserV2Exists(accProvider func() (*schema.Provider, error), n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		provider, _ := accProvider()
		providerConf := provider.Meta().(*datadog.ProviderConfiguration)
		apiInstances := providerConf.DatadogApiInstances
		auth := providerConf.Auth

		if err := datadogUserV2ExistsHelper(auth, s, apiInstances, n); err != nil {
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
  send_user_invitation = true
}`, uniq)
}

func testAccCheckDatadogUserConfigRequiredNoInvitation(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_user" "foo" {
  email     = "%s"
  name      = "Test User"
  send_user_invitation = false
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

func datadogUserV2DestroyHelper(ctx context.Context, s *terraform.State, apiInstances *utils.ApiInstances) error {
	for _, r := range s.RootModule().Resources {
		id := r.Primary.ID
		userResponse, httpResponse, err := apiInstances.GetUsersApiV2().GetUser(ctx, id)

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

func datadogUserV2ExistsHelper(ctx context.Context, s *terraform.State, apiInstances *utils.ApiInstances, name string) error {
	id := s.RootModule().Resources[name].Primary.ID
	if _, _, err := apiInstances.GetUsersApiV2().GetUser(ctx, id); err != nil {
		return fmt.Errorf("received an error retrieving user %s", err)
	}
	return nil
}
