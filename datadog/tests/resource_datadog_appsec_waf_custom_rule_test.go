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
		ProtoV6ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogAppsecWafCustomRuleDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogAppsecWafCustomRule(uniq),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogAppsecWafCustomRuleExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(
						"datadog_appsec_waf_custom_rule.foo", "tags.category", "attack_attempt"),
					resource.TestCheckResourceAttr(
						"datadog_appsec_waf_custom_rule.foo", "path_glob", "/api/search/*"),
					resource.TestCheckResourceAttr(
						"datadog_appsec_waf_custom_rule.foo", "condition.0.parameters.input.0.address", "server.request.query"),
				),
			},
		},
	})
}

func testAccCheckDatadogAppsecWafCustomRule(uniq string) string {
	// Update me to make use of the unique value
	return fmt.Sprintf(`resource "datadog_appsec_waf_custom_rule" "foo" {
    action {
      action = "redirect_request"
      parameters {
        location = "/blocking"
        status_code = 302
      }
    }
    blocking = true
    condition {
      operator = "match_regex"
      parameters {
        input {
          address = "server.request.query"
          key_path = [ "test" ]
        }
        options {
          case_sensitive = true
        }
        regex = "test.*"
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
      category = "attack_attempt"
      type = "test"
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
