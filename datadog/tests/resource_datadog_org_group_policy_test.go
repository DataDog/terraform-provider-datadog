package test

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/fwprovider"
)

func TestAccDatadogOrgGroupPolicy_Basic(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	orgGroupName := uniqueEntityName(ctx, t)
	policyName := "is_widget_copy_paste_enabled"
	replacementPolicyName := "is_dashboard_reports_enabled"
	resourceName := "datadog_org_group_policy.foo"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogOrgGroupPolicyDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogOrgGroupPolicyConfig(orgGroupName, policyName, `{"org_config":false}`, "DEFAULT"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogOrgGroupPolicyExists(providers.frameworkProvider, resourceName),
					resource.TestCheckResourceAttr(resourceName, "policy_name", policyName),
					resource.TestCheckResourceAttr(resourceName, "enforcement_tier", "DEFAULT"),
					resource.TestCheckResourceAttr(resourceName, "policy_type", "org_config"),
					resource.TestCheckResourceAttr(resourceName, "content", `{"org_config":false}`),
					resource.TestCheckResourceAttrSet(resourceName, "org_group_id"),
				),
			},
			{
				Config: testAccCheckDatadogOrgGroupPolicyConfig(orgGroupName, policyName, `{"org_config":true}`, "ENFORCE"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogOrgGroupPolicyExists(providers.frameworkProvider, resourceName),
					resource.TestCheckResourceAttr(resourceName, "enforcement_tier", "ENFORCE"),
					resource.TestCheckResourceAttr(resourceName, "content", `{"org_config":true}`),
				),
			},
			{
				// Changing policy_name must force replacement.
				Config: testAccCheckDatadogOrgGroupPolicyConfig(orgGroupName, replacementPolicyName, `{"org_config":true}`, "ENFORCE"),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionDestroyBeforeCreate),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogOrgGroupPolicyExists(providers.frameworkProvider, resourceName),
					resource.TestCheckResourceAttr(resourceName, "policy_name", replacementPolicyName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckDatadogOrgGroupPolicyConfig(orgGroupName, policyName, content, enforcementTier string) string {
	return fmt.Sprintf(`
resource "datadog_org_group" "foo" {
  name = "%s"
}

resource "datadog_org_group_policy" "foo" {
  org_group_id     = datadog_org_group.foo.id
  policy_name      = "%s"
  content          = jsonencode(%s)
  enforcement_tier = "%s"
}`, orgGroupName, policyName, content, enforcementTier)
}

func testAccCheckDatadogOrgGroupPolicyExists(accProvider *fwprovider.FrameworkProvider, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		apiInstances := accProvider.DatadogApiInstances
		auth := accProvider.Auth

		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("resource not found: %s", n)
		}

		id, err := uuid.Parse(rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("org group policy ID is not a valid UUID: %w", err)
		}

		_, _, err = apiInstances.GetOrgGroupsApiV2().GetOrgGroupPolicy(auth, id)
		if err != nil {
			return fmt.Errorf("received an error retrieving org group policy: %w", err)
		}
		return nil
	}
}

func testAccCheckDatadogOrgGroupPolicyDestroy(accProvider *fwprovider.FrameworkProvider) func(*terraform.State) error {
	return func(s *terraform.State) error {
		apiInstances := accProvider.DatadogApiInstances
		auth := accProvider.Auth

		for _, r := range s.RootModule().Resources {
			if r.Type != "datadog_org_group_policy" {
				continue
			}

			id, err := uuid.Parse(r.Primary.ID)
			if err != nil {
				return fmt.Errorf("org group policy ID is not a valid UUID: %w", err)
			}

			_, httpResp, err := apiInstances.GetOrgGroupsApiV2().GetOrgGroupPolicy(auth, id)
			if err != nil {
				if httpResp != nil && httpResp.StatusCode == 404 {
					continue
				}
				return fmt.Errorf("received an error retrieving org group policy: %w", err)
			}

			return fmt.Errorf("org group policy still exists")
		}

		return nil
	}
}
