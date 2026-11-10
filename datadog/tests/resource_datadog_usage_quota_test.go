package test

import (
	"context"
	"fmt"
	"testing"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// Requires a sandbox organization with AI credits quotas enabled and role management permissions.
func TestAccDatadogUsageQuota(t *testing.T) {
	ctx, providers, factories := testAccFrameworkMuxProviders(context.Background(), t)
	name := uniqueEntityName(ctx, t)
	const address = "datadog_usage_quota.test"
	config := func(limit int, enforced bool) string {
		return fmt.Sprintf(`
resource "datadog_role" "quota" { name = %q }
resource "datadog_usage_quota" "test" {
 quota_namespace = "ai_credits"
 scope = { role_uuid = datadog_role.quota.id }
 usage_limit = %d
 enforced = %t
}`, name, limit, enforced)
	}
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: factories,
		CheckDestroy: func(state *terraform.State) error {
			p := providers.frameworkProvider
			cfg := p.DatadogApiInstances.HttpClient.GetConfig()
			cfg.SetUnstableOperationEnabled("v2.GetQuota", true)
			api := p.DatadogApiInstances.GetUsageMeteringApiV2()
			for _, rs := range state.RootModule().Resources {
				if rs.Type != "datadog_usage_quota" {
					continue
				}
				_, h, err := api.GetQuota(p.Auth, rs.Primary.Attributes["quota_namespace"], rs.Primary.Attributes["quota_id"], *datadogV2.NewGetQuotaOptionalParameters().WithIncludeDescendants(false))
				if h != nil && h.StatusCode == 404 {
					continue
				}
				if err != nil {
					return err
				}
				return fmt.Errorf("usage quota still exists")
			}
			return nil
		},
		Steps: []resource.TestStep{
			{Config: config(100, true), Check: resource.ComposeTestCheckFunc(resource.TestCheckResourceAttr(address, "usage_limit", "100"), resource.TestCheckResourceAttr(address, "enforced", "true"), resource.TestCheckResourceAttrSet(address, "quota_id"), resource.TestCheckResourceAttrSet(address, "org_public_id"))},
			{ResourceName: address, ImportState: true, ImportStateVerify: true},
			{Config: config(200, true), Check: resource.TestCheckResourceAttr(address, "usage_limit", "200")},
			{Config: config(200, false), Check: resource.TestCheckResourceAttr(address, "enforced", "false")},
		},
	})
}
