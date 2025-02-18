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

func TestAccAppsecExclusionFilterBasic(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	uniq := uniqueEntityName(ctx, t)

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogAppsecExclusionFilterDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogAppsecExclusionFilter(uniq),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogAppsecExclusionFilterExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(
						"datadog_appsec_exclusion_filter.foo", "description", "Exclude false positives on a path"),
					resource.TestCheckResourceAttr(
						"datadog_appsec_exclusion_filter.foo", "enabled", "True"),
					resource.TestCheckResourceAttr(
						"datadog_appsec_exclusion_filter.foo", "event_query", "UPDATE ME"),
					resource.TestCheckResourceAttr(
						"datadog_appsec_exclusion_filter.foo", "on_match", "UPDATE ME"),
					resource.TestCheckResourceAttr(
						"datadog_appsec_exclusion_filter.foo", "path_glob", "/accounts/*"),
				),
			},
		},
	})
}

func testAccCheckDatadogAppsecExclusionFilter(uniq string) string {
	// Update me to make use of the unique value
	return fmt.Sprintf(`resource "datadog_appsec_exclusion_filter" "foo" {
    description = "%s"
    enabled = True
    event_query = "UPDATE ME"
    ip_list = "UPDATE ME"
    on_match = "UPDATE ME"
    parameters = "UPDATE ME"
    path_glob = "/accounts/*"
    rules_target {
    rule_id = "dog-913-009"
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

func testAccCheckDatadogAppsecExclusionFilterDestroy(accProvider *fwprovider.FrameworkProvider) func(*terraform.State) error {
	return func(s *terraform.State) error {
		apiInstances := accProvider.DatadogApiInstances
		auth := accProvider.Auth

		if err := AppsecExclusionFilterDestroyHelper(auth, s, apiInstances); err != nil {
			return err
		}
		return nil
	}
}

func AppsecExclusionFilterDestroyHelper(auth context.Context, s *terraform.State, apiInstances *utils.ApiInstances) error {
	err := utils.Retry(2, 10, func() error {
		for _, r := range s.RootModule().Resources {
			if r.Type != "resource_datadog_appsec_exclusion_filter" {
				continue
			}
			id := r.Primary.ID

			_, httpResp, err := apiInstances.GetApplicationSecurityApiV2().GetApplicationSecurityExclusionFilter(auth, id)
			if err != nil {
				if httpResp != nil && httpResp.StatusCode == 404 {
					return nil
				}
				return &utils.RetryableError{Prob: fmt.Sprintf("received an error retrieving AppsecExclusionFilter %s", err)}
			}
			return &utils.RetryableError{Prob: "AppsecExclusionFilter still exists"}
		}
		return nil
	})
	return err
}

func testAccCheckDatadogAppsecExclusionFilterExists(accProvider *fwprovider.FrameworkProvider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		apiInstances := accProvider.DatadogApiInstances
		auth := accProvider.Auth

		if err := appsecExclusionFilterExistsHelper(auth, s, apiInstances); err != nil {
			return err
		}
		return nil
	}
}

func appsecExclusionFilterExistsHelper(auth context.Context, s *terraform.State, apiInstances *utils.ApiInstances) error {
	for _, r := range s.RootModule().Resources {
		if r.Type != "resource_datadog_appsec_exclusion_filter" {
			continue
		}
		id := r.Primary.ID

		_, httpResp, err := apiInstances.GetApplicationSecurityApiV2().GetApplicationSecurityExclusionFilter(auth, id)
		if err != nil {
			return utils.TranslateClientError(err, httpResp, "error retrieving AppsecExclusionFilter")
		}
	}
	return nil
}
