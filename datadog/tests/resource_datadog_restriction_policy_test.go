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
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	uniq := uniqueEntityName(ctx, t)

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogRestrictionPolicyDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogRestrictionPolicy(uniq),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogRestrictionPolicyExists(providers.frameworkProvider),
				),
			},
		},
	})
}

func testAccCheckDatadogRestrictionPolicy(uniq string) string {
	return fmt.Sprintf(`
	resource "datadog_restriction_policy" "foo" {
	bindings {
	principals = ["role:00000000-0000-1111-0000-000000000000"]
	relation = "editor"
	}
	}`, uniq)
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
				return &utils.RetryableError{Prob: fmt.Sprintf("received an error retrieving Monitor %s", err)}
			}
			return &utils.RetryableError{Prob: "Monitor still exists"}
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
			return utils.TranslateClientError(err, httpResp, "error retrieving monitor")
		}
	}
	return nil
}
