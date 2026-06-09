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

func TestAccDatadogTagIndexingRuleExemption_Basic(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)

	// The exemption API requires the metric to already exist in the org.
	// system.cpu.user is a standard Datadog agent metric present in any org with a running agent.
	const metricName = "system.cpu.user"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogTagIndexingRuleExemptionDestroy(ctx, providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogTagIndexingRuleExemptionConfigBasic(metricName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("datadog_tag_indexing_rule_exemption.foo", "metric_name", metricName),
					resource.TestCheckResourceAttr("datadog_tag_indexing_rule_exemption.foo", "reason", "Test exemption created by Terraform acceptance test"),
					resource.TestCheckResourceAttrSet("datadog_tag_indexing_rule_exemption.foo", "id"),
					resource.TestCheckResourceAttrSet("datadog_tag_indexing_rule_exemption.foo", "created_at"),
				),
			},
			{
				ResourceName:      "datadog_tag_indexing_rule_exemption.foo",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckDatadogTagIndexingRuleExemptionDestroy(ctx context.Context, frameworkProvider *fwprovider.FrameworkProvider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		apiInstances := frameworkProvider.DatadogApiInstances
		auth := frameworkProvider.Auth
		return datadogTagIndexingRuleExemptionDestroyHelper(ctx, auth, s, apiInstances)
	}
}

func datadogTagIndexingRuleExemptionDestroyHelper(_ context.Context, auth context.Context, s *terraform.State, apiInstances *utils.ApiInstances) error {
	api := apiInstances.GetMetricsApiV2()
	for _, r := range s.RootModule().Resources {
		if r.Type != "datadog_tag_indexing_rule_exemption" {
			continue
		}
		_, httpResp, err := api.GetTagIndexingRuleExemption(auth, r.Primary.ID)
		if err != nil {
			if httpResp != nil && httpResp.StatusCode == 404 {
				continue
			}
			return fmt.Errorf("received an error retrieving tag indexing rule exemption: %s", err.Error())
		}
		return fmt.Errorf("tag indexing rule exemption still exists")
	}
	return nil
}

func testAccCheckDatadogTagIndexingRuleExemptionConfigBasic(metricName string) string {
	return fmt.Sprintf(`
resource "datadog_tag_indexing_rule_exemption" "foo" {
  metric_name = %q
  reason      = "Test exemption created by Terraform acceptance test"
}`, metricName)
}
