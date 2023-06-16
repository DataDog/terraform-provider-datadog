package test

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/terraform-providers/terraform-provider-datadog/datadog"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

func TestAccDatadogIPAllowlist_CreateUpdate(t *testing.T) {
	_, accProviders := testAccProviders(context.Background(), t)
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		CheckDestroy:      testAccCheckDatadogIPAllowlistDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogIPAllowlistConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("datadog_ip_allowlist.foo", "enabled", "false"),
					resource.TestCheckTypeSetElemNestedAttrs("datadog_ip_allowlist.foo", "entry.*", map[string]string{
						"cidr_block": "127.0.0.1/32",
						"note":       "entry 1",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("datadog_ip_allowlist.foo", "entry.*", map[string]string{
						"cidr_block": "1.2.3.4/32",
						"note":       "entry 2",
					}),
				),
			},
			{
				Config: testAccCheckDatadogIPAllowlistConfigUpdated(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("datadog_ip_allowlist.foo", "enabled", "true"),
					resource.TestCheckTypeSetElemNestedAttrs("datadog_ip_allowlist.foo", "entry.*", map[string]string{
						"cidr_block": "0.0.0.0/0",
						"note":       "all",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("datadog_ip_allowlist.foo", "entry.*", map[string]string{
						"cidr_block": "1.2.3.4/32",
						"note":       "fake entry",
					}),
				),
			},
			{
				Config:      testAccCheckDatadogIPAllowlistConfigInvalid(),
				ExpectError: regexp.MustCompile("Cannot enable or keep enabled an IP Allowlist without the current IP address in it"),
			},
		},
	})
}

func testAccCheckDatadogIPAllowlistDestroy(accProvider func() (*schema.Provider, error)) func(*terraform.State) error {
	return func(s *terraform.State) error {
		provider, _ := accProvider()
		providerConf := provider.Meta().(*datadog.ProviderConfiguration)
		apiInstances := providerConf.DatadogApiInstances
		auth := providerConf.Auth

		for _, r := range s.RootModule().Resources {
			if r.Type != "datadog_ip_allowlist" {
				// Only care about ip allowlist
				continue
			}
			resp, httpresp, err := apiInstances.GetIPAllowlistApiV2().GetIPAllowlist(auth)
			if err != nil {
				return utils.TranslateClientError(err, httpresp, "error getting IP allowlist")
			}
			ipAllowlistAttributes := resp.GetData().Attributes
			if ipAllowlistAttributes.GetEnabled() != false || len(ipAllowlistAttributes.GetEntries()) != 0 {
				return fmt.Errorf("IP allowlist not disabled or empty")
			}
		}
		return nil
	}
}

func testAccCheckDatadogIPAllowlistConfig() string {
	return `
resource "datadog_ip_allowlist" "foo" {
	enabled = false
	entry {
		cidr_block = "127.0.0.1/32"
		note = "entry 1"
	}
	entry {
		cidr_block = "1.2.3.4/32"
		note = "entry 2"
	}
}`
}

func testAccCheckDatadogIPAllowlistConfigUpdated() string {
	return `
resource "datadog_ip_allowlist" "foo" {
	enabled = true
	entry {
		cidr_block = "1.2.3.4/32"
		note = "fake entry"
	}
	entry {
		cidr_block = "0.0.0.0/0"
		note = "all"
	}
}`
}

func testAccCheckDatadogIPAllowlistConfigInvalid() string {
	return `
resource "datadog_ip_allowlist" "foo" {
	enabled = true
	entry {
		cidr_block = "1.2.3.4/32"
		note = "fake entry"
	}
}`
}
