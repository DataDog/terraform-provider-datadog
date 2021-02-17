package datadog

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"strings"
	"testing"
)

func TestAccDatadogOrganization_CreateUpdate(t *testing.T) {
	accProviders, clock, cleanup := testAccProviders(t, initRecorder(t))
	organizationname := strings.ToLower(fmt.Sprintf("tf-%d", clock.Now().Unix()))
	defer cleanup(t)
	accProvider := testAccProvider(t, accProviders)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() {},
		Providers:    accProviders,
		CheckDestroy: testAccCheckDatadogOrganizationDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccountCheckDatadogChildOrganizationConfig(organizationname),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogChildOrganizationExists(accProvider, "datadog_child_organization.foo"),
					resource.TestCheckResourceAttr("datadog_child_organization.foo", "name", organizationname),
				),
			},
			{
				Config: testAccountCheckDatadogChildOrganizationConfig(organizationname + "updated"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogChildOrganizationExists(accProvider, "datadog_child_organization.foo"),
					resource.TestCheckResourceAttr("datadog_child_organization.foo", "name", organizationname+"updated"),
				),
			},
		},
	})
}

func testAccCheckDatadogOrganizationDestroy(accProvider *schema.Provider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		providerConf := accProvider.Meta().(*ProviderConfiguration)
		client := providerConf.DatadogClientV1
		auth := providerConf.AuthV1

		for _, r := range s.RootModule().Resources {
			if r.Type != "datadog_child_organization" {
				// Only care about roles
				continue
			}
			_, httpresp, err := client.OrganizationsApi.GetOrg(auth, r.Primary.ID).Execute()
			if err != nil {
				if !(httpresp != nil && httpresp.StatusCode == 404) {
					return translateClientError(err, "error getting child organization")
				}
				// Role was successfully deleted
				continue
			}
			return fmt.Errorf("organization %s still exists", r.Primary.ID)
		}
		return nil
	}
}

func testAccCheckDatadogChildOrganizationExists(accProvider *schema.Provider, rolename string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		providerConf := accProvider.Meta().(*ProviderConfiguration)
		client := providerConf.DatadogClientV1
		auth := providerConf.AuthV1

		id := s.RootModule().Resources[rolename].Primary.ID
		_, _, err := client.OrganizationsApi.GetOrg(auth, id).Execute()
		if err != nil {
			return translateClientError(err, "error checking organization existence")
		}
		return nil
	}
}

func testAccountCheckDatadogChildOrganizationConfig(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_child_organization" "foo" {
	name = "%s"
}
`, uniq)
}
