package test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/fwprovider"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

func TestAccDatadogDomainAllowlist_CreateUpdate(t *testing.T) {
	if !isReplaying() {
		// Skip in non-replay mode due to backend replication delay issues.
		t.Skip("This test only runs in replay mode")
	}
	_, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)

	// When generating the casette, it may be necessary to add sleep functions before the check
	// The endpoint has a zonal cache with a ttl of 2 second
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogDomainAllowlistDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogDomainAllowlistConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("datadog_domain_allowlist.foo", "enabled", "true"),
					resource.TestCheckResourceAttr("datadog_domain_allowlist.foo", "domains.0", "@test.com"),
					resource.TestCheckResourceAttr("datadog_domain_allowlist.foo", "domains.1", "@datadoghq.com"),
				),
			},
			{
				Config: testAccCheckDatadogDomainAllowlistConfigUpdated(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("datadog_domain_allowlist.foo", "domains.0", "@gmail.com"),
					resource.TestCheckResourceAttr("datadog_domain_allowlist.foo", "domains.1", "@datadoghq.com"),
					resource.TestCheckResourceAttr("datadog_domain_allowlist.foo", "enabled", "false"),
				),
			},
		},
	})
}

func testAccCheckDatadogDomainAllowlistDestroy(accProvider *fwprovider.FrameworkProvider) func(*terraform.State) error {
	return func(s *terraform.State) error {
		apiInstances := accProvider.DatadogApiInstances
		auth := accProvider.Auth

		for _, r := range s.RootModule().Resources {
			if r.Type != "datadog_domain_allowlist" {
				// Only care about domain allowlist
				continue
			}
			resp, httpresp, err := apiInstances.GetDomainAllowlistApiV2().GetDomainAllowlist(auth)
			if err != nil {
				return utils.TranslateClientError(err, httpresp, "error getting Domain allowlist")
			}
			domainAllowlistAttributes := resp.GetData().Attributes
			if domainAllowlistAttributes.GetEnabled() != false || len(domainAllowlistAttributes.GetDomains()) != 0 {
				return fmt.Errorf("Domain allowlist not disabled or empty")
			}
		}
		return nil
	}
}

func testAccCheckDatadogDomainAllowlistConfig() string {
	return `
resource "datadog_domain_allowlist" "foo" {
	enabled = true
	domains = ["@test.com", "@datadoghq.com"]
}`
}

func testAccCheckDatadogDomainAllowlistConfigUpdated() string {
	return `
resource "datadog_domain_allowlist" "foo" {
	enabled = false
	domains = ["@gmail.com", "@datadoghq.com"]
}`
}
