package test

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/fwprovider"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

func TestAccRestrictionPolicyBasic(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	ruleName := uniqueEntityName(ctx, t)
	resourceType := "security-rule"

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogRestrictionPolicyDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogRestrictionPolicy(ruleName, resourceType),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogRestrictionPolicyExists(providers.frameworkProvider),
				),
			},
		},
	})
}

func TestAccRestrictionPolicyUpdate(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	ruleName := uniqueEntityName(ctx, t)
	resourceType := "security-rule"

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogRestrictionPolicyDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogRestrictionPolicy(ruleName, resourceType),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogRestrictionPolicyExists(providers.frameworkProvider),
				),
			},
			{
				Config: testAccCheckDatadogRestrictionPolicyUpdate(ruleName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogRestrictionPolicyExists(providers.frameworkProvider),
					resource.TestCheckTypeSetElemAttr(
						"datadog_restriction_policy.baz", "bindings.0.principals.*", "org:4dee724d-00cc-11ea-a77b-570c9d03c6c5"),
				),
			},
		},
	})
}

func TestAccRestrictionPolicyInvalidInput(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	ruleName := uniqueEntityName(ctx, t)
	resourceType := "security-rule"
	invalidResourceType := "security_rule"

	invalidResourceTypeError, _ := regexp.Compile("Invalid resource type")
	invalidPrincipalError, _ := regexp.Compile("not a valid principal")

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogRestrictionPolicyDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config:      testAccCheckDatadogRestrictionPolicy(ruleName, invalidResourceType),
				ExpectError: invalidResourceTypeError,
			},
			{
				Config:      testAccCheckDatadogRestrictionPolicyInvalidPrincipal(ruleName, resourceType),
				ExpectError: invalidPrincipalError,
			},
		},
	})

}

func testAccCheckDatadogRestrictionPolicy(ruleName string, resourceType string) string {
	return fmt.Sprintf(`
        data "datadog_role" "foo" {
          filter = "Datadog Admin Role"
        }
        resource "datadog_security_monitoring_rule" "bar" {
          name = "%s"

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
            resource_id = "%s:${datadog_security_monitoring_rule.bar.id}"
            bindings {
              principals = ["org:4dee724d-00cc-11ea-a77b-570c9d03c6c5","role:${data.datadog_role.foo.id}"]
              relation = "editor"
            }
        }`, ruleName, resourceType)
}

func testAccCheckDatadogRestrictionPolicyUpdate(ruleName string) string {
	return fmt.Sprintf(`
        resource "datadog_security_monitoring_rule" "bar" {
          name = "%s"

          message = "The rule has triggered. (updated)"
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
            resource_id = "security-rule:${datadog_security_monitoring_rule.bar.id}"
            bindings {
              principals = ["org:4dee724d-00cc-11ea-a77b-570c9d03c6c5"]
              relation = "editor"
            }
        }
        `, ruleName)
}

func testAccCheckDatadogRestrictionPolicyInvalidPrincipal(ruleName string, resourceType string) string {
	return fmt.Sprintf(`
        resource "datadog_security_monitoring_rule" "bar" {
          name = "%s"

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
            resource_id = "%s:${datadog_security_monitoring_rule.bar.id}"
            bindings {
              principals = ["org:4dee724d-00cc-11ea-a77b-570c9d03c6c5","foo:4dee724d-00cc-11ea-a77b-570c9d03c6c5"]
              relation = "editor"
            }
        }`, ruleName, resourceType)
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
