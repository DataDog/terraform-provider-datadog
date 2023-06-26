package test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/fwprovider"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

func TestAccRestrictionPolicyBasic(t *testing.T) {
	t.Parallel()
	_, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogRestrictionPolicyDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogRestrictionPolicy(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogRestrictionPolicyExists(providers.frameworkProvider),
				),
			},
		},
	})
}

func testAccCheckDatadogRestrictionPolicy() string {
	return `
        data "datadog_role" "foo" {
          filter = "Datadog Admin Role"
        }
        resource "datadog_security_monitoring_rule" "bar" {
          name = "My rule"

          message = "The rule has triggered."
          enabled = true

          query {
            name            = "errors"
            query           = "status:error"
            aggregation     = "count"
            group_by_fields = ["host"]
          }

          query {
            name            = "warnings"
            query           = "status:warning"
            aggregation     = "count"
            group_by_fields = ["host"]
          }

          case {
            status        = "high"
            condition     = "errors > 3 && warnings > 10"
            notifications = ["@user"]
          }

          options {
            evaluation_window   = 300
            keep_alive          = 600
            max_signal_duration = 900
          }
        }
        resource "datadog_restriction_policy" "baz" {
            resource_id = "dashboard:${datadog_security_monitoring_rule.bar.id}"
            bindings {
            principals = ["org:4dee724d-00cc-11ea-a77b-570c9d03c6c5","role:${data.datadog_role.foo.id}"]
            relation = "editor"
            }
        }`
}

func testAccCheckDatadogRestrictionPolicyDestroy(accProvider *fwprovider.FrameworkProvider) func(*terraform.State) error {
	return func(s *terraform.State) error {
		apiInstances := accProvider.DatadogApiInstances
		auth := accProvider.Auth

		if err := RestrictionPolicyDestroyHelper(auth, s, apiInstances); err != nil {
			return err
		}
		return nil
	}
}

func RestrictionPolicyDestroyHelper(auth context.Context, s *terraform.State, apiInstances *utils.ApiInstances) error {
	err := utils.Retry(2, 10, func() error {
		for _, r := range s.RootModule().Resources {
			if r.Type != "resource_datadog_restriction_policy" {
				continue
			}
			id := r.Primary.ID

			_, httpResp, err := apiInstances.GetRestrictionPoliciesApiV2().GetRestrictionPolicy(auth, id)
			if err != nil {
				if httpResp != nil && httpResp.StatusCode == 404 {
					return nil
				}
				return &utils.RetryableError{Prob: fmt.Sprintf("received an error retrieving Restriction Policy %s", err)}
			}
			return &utils.RetryableError{Prob: "Restriction Policy still exists"}
		}
		return nil
	})
	return err
}

func testAccCheckDatadogRestrictionPolicyExists(accProvider *fwprovider.FrameworkProvider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		apiInstances := accProvider.DatadogApiInstances
		auth := accProvider.Auth

		if err := restrictionPolicyExistsHelper(auth, s, apiInstances); err != nil {
			return err
		}
		return nil
	}
}

func restrictionPolicyExistsHelper(auth context.Context, s *terraform.State, apiInstances *utils.ApiInstances) error {
	for _, r := range s.RootModule().Resources {
		if r.Type != "resource_datadog_restriction_policy" {
			continue
		}
		id := r.Primary.ID

		_, httpResp, err := apiInstances.GetRestrictionPoliciesApiV2().GetRestrictionPolicy(auth, id)
		if err != nil {
			return utils.TranslateClientError(err, httpResp, "error retrieving restriction policy")
		}
	}
	return nil
}
