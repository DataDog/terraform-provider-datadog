package datadog

import (
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccDatadogAppKey_Basic(t *testing.T) {
	accProviders, cleanup := testAccProviders(t)
	defer cleanup(t)
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    accProviders,
		CheckDestroy: testAccCheckDatadogAppKeyDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogAppKeyConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogAppKeyExists(accProvider, "datadog_app_key.foo"),
					resource.TestCheckResourceAttr(
						"datadog_app_key.foo", "name", "foo"),
					resource.TestCheckResourceAttrSet(
						"datadog_app_key.foo", "hash"),
				),
			},
		},
	})
}

func TestAccDatadogAppKey_Updated(t *testing.T) {
	accProviders, cleanup := testAccProviders(t)
	defer cleanup(t)
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    accProviders,
		CheckDestroy: testAccCheckDatadogAppKeyDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogAppKeyConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogAppKeyExists(accProvider, "datadog_app_key.foo"),
					resource.TestCheckResourceAttr(
						"datadog_app_key.foo", "name", "foo"),
					resource.TestCheckResourceAttrSet(
						"datadog_app_key.foo", "hash"),
				),
			},
			{
				Config: testAccCheckDatadogAppKeyConfigUpdated,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogAppKeyExists(accProvider, "datadog_app_key.foo"),
					resource.TestCheckResourceAttr(
						"datadog_app_key.foo", "name", "bar"),
					resource.TestCheckResourceAttrSet(
						"datadog_app_key.foo", "hash"),
				),
			},
		},
	})
}

func testAccCheckDatadogAppKeyDestroy(accProvider *schema.Provider) func(*terraform.State) error {
	return func(s *terraform.State) error {
		providerConf := accProvider.Meta().(*ProviderConfiguration)
		client := providerConf.CommunityClient
		for _, r := range s.RootModule().Resources {
			if _, err := client.GetAPPKey(r.Primary.ID); err != nil {
				if strings.Contains(err.Error(), "404 Not Found") {
					continue
				}
				return fmt.Errorf("received an error retrieving app key %s", err)
			}
			return fmt.Errorf("app key still exists")
		}
		return nil
	}
}

func testAccCheckDatadogAppKeyExists(accProvider *schema.Provider, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		providerConf := accProvider.Meta().(*ProviderConfiguration)
		client := providerConf.CommunityClient

		for _, r := range s.RootModule().Resources {
			id := r.Primary.ID
			if _, err := client.GetAPPKey(id); err != nil {
				return fmt.Errorf("received an error retrieving app key %s", err)
			}
		}

		return nil
	}
}

const testAccCheckDatadogAppKeyConfig = `
resource "datadog_app_key" "foo" {
  name = "foo"
}
`

const testAccCheckDatadogAppKeyConfigUpdated = `
resource "datadog_app_key" "foo" {
  name = "bar"
}
`
