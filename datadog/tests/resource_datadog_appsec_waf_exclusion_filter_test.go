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

func TestAccAppsecWafExclusionFilterBasic(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	uniq := uniqueEntityName(ctx, t)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogAppsecWafExclusionFilterDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogAppsecWafExclusionFilter(uniq),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogAppsecWafExclusionFilterExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(
						"datadog_appsec_waf_exclusion_filter.foo", "scope.0.env", "www"),
					resource.TestCheckResourceAttr(
						"datadog_appsec_waf_exclusion_filter.foo", "rules_target.0.tags.category", "attack_attempt"),
					resource.TestCheckResourceAttr(
						"datadog_appsec_waf_exclusion_filter.foo", "path_glob", "/accounts/*"),
				),
			},
		},
	})
}

func testAccCheckDatadogAppsecWafExclusionFilter(uniq string) string {
	// Update me to make use of the unique value
	return fmt.Sprintf(`resource "datadog_appsec_waf_exclusion_filter" "foo" {
    description = "%s"
    enabled = true
    path_glob = "/accounts/*"
    rules_target {
      tags {
        category = "attack_attempt"
        type = "lfi"
      }
    }
    scope {
      env = "www"
      service = "prod"
    }
}`, uniq)
}

func testAccCheckDatadogAppsecWafExclusionFilterDestroy(accProvider *fwprovider.FrameworkProvider) func(*terraform.State) error {
	return func(s *terraform.State) error {
		apiInstances := accProvider.DatadogApiInstances
		auth := accProvider.Auth

		if err := AppsecWafExclusionFilterDestroyHelper(auth, s, apiInstances); err != nil {
			return err
		}
		return nil
	}
}

func AppsecWafExclusionFilterDestroyHelper(auth context.Context, s *terraform.State, apiInstances *utils.ApiInstances) error {
	err := utils.Retry(2, 10, func() error {
		for _, r := range s.RootModule().Resources {
			if r.Type != "resource_datadog_appsec_waf_exclusion_filter" {
				continue
			}
			id := r.Primary.ID

			_, httpResp, err := apiInstances.GetApplicationSecurityApiV2().GetApplicationSecurityWafExclusionFilter(auth, id)
			if err != nil {
				if httpResp != nil && httpResp.StatusCode == 404 {
					return nil
				}
				return &utils.RetryableError{Prob: fmt.Sprintf("received an error retrieving AppsecWafExclusionFilter %s", err)}
			}
			return &utils.RetryableError{Prob: "AppsecWafExclusionFilter still exists"}
		}
		return nil
	})
	return err
}

func testAccCheckDatadogAppsecWafExclusionFilterExists(accProvider *fwprovider.FrameworkProvider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		apiInstances := accProvider.DatadogApiInstances
		auth := accProvider.Auth

		if err := appsecWafExclusionFilterExistsHelper(auth, s, apiInstances); err != nil {
			return err
		}
		return nil
	}
}

func appsecWafExclusionFilterExistsHelper(auth context.Context, s *terraform.State, apiInstances *utils.ApiInstances) error {
	for _, r := range s.RootModule().Resources {
		if r.Type != "resource_datadog_appsec_waf_exclusion_filter" {
			continue
		}
		id := r.Primary.ID

		_, httpResp, err := apiInstances.GetApplicationSecurityApiV2().GetApplicationSecurityWafExclusionFilter(auth, id)
		if err != nil {
			return utils.TranslateClientError(err, httpResp, "error retrieving AppsecWafExclusionFilter")
		}
	}
	return nil
}
