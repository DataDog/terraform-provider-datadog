package datadog

import (
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccDatadogAPIKey_Basic(t *testing.T) {
	accProviders, cleanup := testAccProviders(t)
	defer cleanup(t)
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    accProviders,
		CheckDestroy: testAccCheckDatadogAPIKeyDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogAPIKeyConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogAPIKeyExists(accProvider, "datadog_api_key.foo"),
					resource.TestCheckResourceAttr(
						"datadog_api_key.foo", "name", "foo"),
					resource.TestCheckResourceAttrSet(
						"datadog_api_key.foo", "key"),
				),
			},
		},
	})
}

func TestAccDatadogAPIKey_Updated(t *testing.T) {
	accProviders, cleanup := testAccProviders(t)
	defer cleanup(t)
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    accProviders,
		CheckDestroy: testAccCheckDatadogAPIKeyDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogAPIKeyConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogAPIKeyExists(accProvider, "datadog_api_key.foo"),
					resource.TestCheckResourceAttr(
						"datadog_api_key.foo", "name", "foo"),
					resource.TestCheckResourceAttrSet(
						"datadog_api_key.foo", "key"),
				),
			},
			{
				Config: testAccCheckDatadogAPIKeyConfigUpdated,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogAPIKeyExists(accProvider, "datadog_api_key.foo"),
					resource.TestCheckResourceAttr(
						"datadog_api_key.foo", "name", "bar"),
					resource.TestCheckResourceAttrSet(
						"datadog_api_key.foo", "key"),
				),
			},
		},
	})
}

func testAccCheckDatadogAPIKeyDestroy(accProvider *schema.Provider) func(*terraform.State) error {
	return func(s *terraform.State) error {
		providerConf := accProvider.Meta().(*ProviderConfiguration)
		client := providerConf.CommunityClient
		for _, r := range s.RootModule().Resources {
			if _, err := client.GetAPIKey(r.Primary.ID); err != nil {
				if strings.Contains(err.Error(), "404 Not Found") {
					continue
				}
				return fmt.Errorf("received an error retrieving api key %s", err)
			}
			return fmt.Errorf("api key still exists")
		}
		return nil
	}
}

func testAccCheckDatadogAPIKeyExists(accProvider *schema.Provider, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		providerConf := accProvider.Meta().(*ProviderConfiguration)
		client := providerConf.CommunityClient

		for _, r := range s.RootModule().Resources {
			id := r.Primary.ID
			if _, err := client.GetAPIKey(id); err != nil {
				return fmt.Errorf("received an error retrieving api key %s", err)
			}
		}

		return nil
	}
}

const testAccCheckDatadogAPIKeyConfig = `
resource "datadog_api_key" "foo" {
  name = "foo"
}
`

const testAccCheckDatadogAPIKeyConfigUpdated = `
resource "datadog_api_key" "foo" {
  name = "bar"
}
`
