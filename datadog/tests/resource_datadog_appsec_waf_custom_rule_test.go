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

func TestAccAppsecWafCustomRuleBasic(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	uniq := uniqueEntityName(ctx, t)

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogAppsecWafCustomRuleDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogAppsecWafCustomRule(uniq),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogAppsecWafCustomRuleExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(
						"datadog_appsec_waf_custom_rule.foo", "name", "Block request from a bad useragent"),
					resource.TestCheckResourceAttr(
						"datadog_appsec_waf_custom_rule.foo", "path_glob", "/api/search/*"),
				),
			},
		},
	})
}

func testAccCheckDatadogAppsecWafCustomRule(uniq string) string {
	// Update me to make use of the unique value
	return fmt.Sprintf(`resource "datadog_appsec_waf_custom_rule" "foo" {
    action {
    action = "block_request"
    parameters {
    location = "/blocking"
    status_code = 403
    }
    }
    blocking = false
    conditions {
    operator = "match_regex"
    parameters {
    data = "blocked_users"
    inputs {
    address = "server.db.statement"
    }
    options {
    case_sensitive = true
    }
    regex = "path.*"
    value = "custom_tag"
    }
    }
    enabled = true
    name = "%s"
    path_glob = "/api/search/*"
    scope {
    env = "prod"
    service = "billing-service"
    }
    tags = {
    category = "business_logic"
    type = "users.login.success"
    }
}`, uniq)
}

func testAccCheckDatadogAppsecWafCustomRuleDestroy(accProvider *fwprovider.FrameworkProvider) func(*terraform.State) error {
	return func(s *terraform.State) error {
		apiInstances := accProvider.DatadogApiInstances
		auth := accProvider.Auth

		if err := AppsecWafCustomRuleDestroyHelper(auth, s, apiInstances); err != nil {
			return err
		}
		return nil
	}
}

func AppsecWafCustomRuleDestroyHelper(auth context.Context, s *terraform.State, apiInstances *utils.ApiInstances) error {
	err := utils.Retry(2, 10, func() error {
		for _, r := range s.RootModule().Resources {
			if r.Type != "resource_datadog_appsec_waf_custom_rule" {
				continue
			}
			id := r.Primary.ID

			_, httpResp, err := apiInstances.GetApplicationSecurityApiV2().GetApplicationSecurityWafCustomRule(auth, id)
			if err != nil {
				if httpResp != nil && httpResp.StatusCode == 404 {
					return nil
				}
				return &utils.RetryableError{Prob: fmt.Sprintf("received an error retrieving AppsecWafCustomRule %s", err)}
			}
			return &utils.RetryableError{Prob: "AppsecWafCustomRule still exists"}
		}
		return nil
	})
	return err
}

func testAccCheckDatadogAppsecWafCustomRuleExists(accProvider *fwprovider.FrameworkProvider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		apiInstances := accProvider.DatadogApiInstances
		auth := accProvider.Auth

		if err := appsecWafCustomRuleExistsHelper(auth, s, apiInstances); err != nil {
			return err
		}
		return nil
	}
}

func appsecWafCustomRuleExistsHelper(auth context.Context, s *terraform.State, apiInstances *utils.ApiInstances) error {
	for _, r := range s.RootModule().Resources {
		if r.Type != "resource_datadog_appsec_waf_custom_rule" {
			continue
		}
		id := r.Primary.ID

		_, httpResp, err := apiInstances.GetApplicationSecurityApiV2().GetApplicationSecurityWafCustomRule(auth, id)
		if err != nil {
			return utils.TranslateClientError(err, httpResp, "error retrieving AppsecWafCustomRule")
		}
	}
	return nil
}
