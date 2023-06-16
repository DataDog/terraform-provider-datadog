package test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/terraform-providers/terraform-provider-datadog/datadog"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

func TestAccDatadogAuthNMapping_CreateUpdate(t *testing.T) {
	t.Parallel()
	_, accProviders := testAccProviders(context.Background(), t)
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		CheckDestroy:      testAccCheckDatadogAuthNMappingDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogAuthNMappingConfig("key_1", "value_1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogAuthNMappingExists(accProvider, "datadog_authn_mapping.foo"),
					resource.TestCheckResourceAttr("datadog_authn_mapping.foo", "key", "key_1"),
					resource.TestCheckResourceAttr("datadog_authn_mapping.foo", "value", "value_1"),
					testCheckAuthNMappingHasRole("datadog_authn_mapping.foo", "data.datadog_role.ro_role"),
				),
			},
			{
				Config: testAccCheckDatadogAuthNMappingConfigUpdated("key_2", "value_2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogAuthNMappingExists(accProvider, "datadog_authn_mapping.foo"),
					resource.TestCheckResourceAttr("datadog_authn_mapping.foo", "key", "key_2"),
					resource.TestCheckResourceAttr("datadog_authn_mapping.foo", "value", "value_2"),
					testCheckAuthNMappingHasRole("datadog_authn_mapping.foo", "data.datadog_role.standard_role"),
				),
			},
		},
	})
}

func TestAccDatadogAuthNMapping_import(t *testing.T) {
	t.Parallel()
	_, accProviders := testAccProviders(context.Background(), t)
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		CheckDestroy:      testAccCheckDatadogAuthNMappingDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogAuthNMappingConfig("key_1", "value_1"),
			},
			{
				ResourceName:      "datadog_authn_mapping.foo",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

// Create Terraform Config for AuthN Mapping
func testAccCheckDatadogAuthNMappingConfig(key string, val string) string {
	return fmt.Sprintf(`
	data "datadog_role" "ro_role" {
		filter = "Datadog Read Only Role"
	}
	  
	# Create a new AuthN mapping
	resource "datadog_authn_mapping" "foo" {
	  key   = "%s"
	  value = "%s"
	  role  = data.datadog_role.ro_role.id
	}`, key, val)
}

// Update Terraform Config for AuthN Mapping
func testAccCheckDatadogAuthNMappingConfigUpdated(key string, val string) string {
	return fmt.Sprintf(`
	data "datadog_role" "standard_role" {
		filter = "Datadog Standard Role"
	}

	resource "datadog_authn_mapping" "foo" {
	  key   = "%s"
	  value = "%s"
	  role  = data.datadog_role.standard_role.id
	}`, key, val)
}

// Check if AuthN Mapping Exists
func testAccCheckDatadogAuthNMappingExists(accProvider func() (*schema.Provider, error), authNMappingName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		provider, _ := accProvider()
		providerConf := provider.Meta().(*datadog.ProviderConfiguration)
		apiInstances := providerConf.DatadogApiInstances
		auth := providerConf.Auth

		id := s.RootModule().Resources[authNMappingName].Primary.ID
		_, httpresp, err := apiInstances.GetAuthNMappingsApiV2().GetAuthNMapping(auth, id)
		if err != nil {
			return utils.TranslateClientError(err, httpresp, "error checking authn mapping existence")
		}
		return nil
	}
}

// Verify that AuthNMapping is destroyed after test run
func testAccCheckDatadogAuthNMappingDestroy(accProvider func() (*schema.Provider, error)) func(*terraform.State) error {
	return func(s *terraform.State) error {
		provider, _ := accProvider()
		providerConf := provider.Meta().(*datadog.ProviderConfiguration)
		apiInstances := providerConf.DatadogApiInstances
		auth := providerConf.Auth

		for _, r := range s.RootModule().Resources {
			if r.Type != "datadog_authn_mapping" {
				// Only care about authn mappings
				continue
			}
			_, httpresp, err := apiInstances.GetAuthNMappingsApiV2().GetAuthNMapping(auth, r.Primary.ID)
			if err != nil {
				if !(httpresp != nil && httpresp.StatusCode == 404) {
					return utils.TranslateClientError(err, httpresp, "error getting authn mapping")
				}
				// AuthN Mapping was successfully deleted
				continue
			}
			return fmt.Errorf("authn mapping %s still exists", r.Primary.ID)
		}
		return nil
	}
}

func testCheckAuthNMappingHasRole(authNMappingName string, roleSource string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rootModule := s.RootModule()
		roleID := rootModule.Resources[roleSource].Primary.Attributes["id"]

		return resource.TestCheckResourceAttr(authNMappingName, "role", roleID)(s)
	}
}
