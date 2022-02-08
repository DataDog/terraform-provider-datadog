package test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	"github.com/terraform-providers/terraform-provider-datadog/datadog"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

// TODO: Modify CheckDestroy function to testAccCheckDatadogAuthNMappingDestroy after Delete Context is available
func TestAccDatadogAuthNMapping_Create(t *testing.T) {
	ctx, accProviders := testAccProviders(context.Background(), t)
	attrKey := strings.ToLower(uniqueEntityName(ctx, t))
	attrVal := strings.ToLower(uniqueEntityName(ctx, t))
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		CheckDestroy:      testAccCheckDatadogAuthNMappingDestroyDummy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogAuthNMappingConfig(attrKey, attrVal),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogAuthNMappingExists(accProvider, "datadog_authn_mapping.foo"),
					resource.TestCheckResourceAttr("datadog_authn_mapping.foo", "key", attrKey),
					resource.TestCheckResourceAttr("datadog_authn_mapping.foo", "value", attrVal),
					testCheckAuthNMappingHasRole("datadog_authn_mapping.foo", "data.datadog_role.ro_role"),
				),
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

// Check if AuthN Mapping Exists
func testAccCheckDatadogAuthNMappingExists(accProvider func() (*schema.Provider, error), authNMappingName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		provider, _ := accProvider()
		providerConf := provider.Meta().(*datadog.ProviderConfiguration)
		client := providerConf.DatadogClientV2
		auth := providerConf.AuthV2

		id := s.RootModule().Resources[authNMappingName].Primary.ID
		_, httpresp, err := client.AuthNMappingsApi.GetAuthNMapping(auth, id)
		if err != nil {
			return utils.TranslateClientError(err, httpresp, "error checking authn mapping existence")
		}
		return nil
	}
}

func testAccCheckDatadogAuthNMappingDestroy(accProvider func() (*schema.Provider, error)) func(*terraform.State) error {
	return func(s *terraform.State) error {
		provider, _ := accProvider()
		providerConf := provider.Meta().(*datadog.ProviderConfiguration)
		client := providerConf.DatadogClientV2
		auth := providerConf.AuthV2

		for _, r := range s.RootModule().Resources {
			if r.Type != "datadog_authn_mapping" {
				// Only care about authn mappings
				continue
			}
			_, httpresp, err := client.AuthNMappingsApi.GetAuthNMapping(auth, r.Primary.ID)
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

// Remove after DeleteContext is availoable for AuthN Mappings
func testAccCheckDatadogAuthNMappingDestroyDummy(accProvider func() (*schema.Provider, error)) func(*terraform.State) error {
	return nil
}

func testCheckAuthNMappingHasRole(authNMappingName string, roleSource string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rootModule := s.RootModule()
		roleID := rootModule.Resources[roleSource].Primary.Attributes["id"]

		return resource.TestCheckResourceAttr(authNMappingName, "role", roleID)(s)
	}
}
