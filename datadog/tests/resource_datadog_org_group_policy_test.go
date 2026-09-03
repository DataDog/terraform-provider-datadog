package test

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/fwprovider"
)

func TestAccDatadogOrgGroupPolicy_Basic(t *testing.T) {
	if !isRecording() && !isReplaying() {
		t.Skip("org_group requires a special test org setup not available in live CI runs")
	}
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
				Config: testAccCheckDatadogOrgGroupPolicyConfig(orgGroupName, policyName, `{"org_config":false}`, "OVERRIDE_ALLOWED"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogOrgGroupPolicyExists(providers.frameworkProvider, resourceName),
					resource.TestCheckResourceAttr(resourceName, "policy_name", policyName),
					resource.TestCheckResourceAttr(resourceName, "enforcement_tier", "OVERRIDE_ALLOWED"),
					resource.TestCheckResourceAttr(resourceName, "policy_type", "org_config"),
					resource.TestCheckResourceAttr(resourceName, "content", `{"org_config":false}`),
					resource.TestCheckResourceAttrSet(resourceName, "org_group_id"),
				),
			},
			{
				Config: testAccCheckDatadogOrgGroupPolicyConfig(orgGroupName, policyName, `{"org_config":true}`, "GROUP_MANAGED"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogOrgGroupPolicyExists(providers.frameworkProvider, resourceName),
					resource.TestCheckResourceAttr(resourceName, "enforcement_tier", "GROUP_MANAGED"),
					resource.TestCheckResourceAttr(resourceName, "content", `{"org_config":true}`),
				),
			},
			{
				// org_config policies cannot be renamed in place; the API rejects the
				// update, so a policy_name change forces a replace.
				Config: testAccCheckDatadogOrgGroupPolicyConfig(orgGroupName, replacementPolicyName, `{"org_config":true}`, "GROUP_MANAGED"),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionReplace),
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

// TestAccDatadogOrgGroupPolicy_Role covers policy_type = "role". Unrecorded:
// role policies are gated behind the org_groups_shared_roles feature flag,
// which is staging-only as of writing (not yet in prod), so this cassette
// can only be recorded against a staging org.
func TestAccDatadogOrgGroupPolicy_Role(t *testing.T) {
	if !isRecording() && !isReplaying() {
		t.Skip("org_group requires a special test org setup not available in live CI runs")
	}
	skipIfNoCassette(t)
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	orgGroupName := uniqueEntityName(ctx, t)
	policyName := "finance_read_only"
	renamedPolicyName := "finance_read_write"
	permissionID := "d99415a4-dc4d-11e8-b4b1-4f9e475593a0"      // logs_read_index_data
	otherPermissionID := "d99443a8-dc4d-11e8-b4b2-9f5f4b0b2f5a" // logs_modify_indexes
	resourceName := "datadog_org_group_policy.role"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogOrgGroupPolicyDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogOrgGroupPolicyRoleConfig(orgGroupName, policyName, fmt.Sprintf(`{"permissions":[%q]}`, permissionID), "GROUP_MANAGED"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogOrgGroupPolicyExists(providers.frameworkProvider, resourceName),
					resource.TestCheckResourceAttr(resourceName, "policy_name", policyName),
					resource.TestCheckResourceAttr(resourceName, "policy_type", "role"),
					resource.TestCheckResourceAttr(resourceName, "enforcement_tier", "GROUP_MANAGED"),
					resource.TestCheckResourceAttr(resourceName, "content", fmt.Sprintf(`{"permissions":[%q]}`, permissionID)),
				),
			},
			{
				// Update the permissions set without changing enforcement_tier.
				Config: testAccCheckDatadogOrgGroupPolicyRoleConfig(orgGroupName, policyName, fmt.Sprintf(`{"permissions":[%q,%q]}`, permissionID, otherPermissionID), "GROUP_MANAGED"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogOrgGroupPolicyExists(providers.frameworkProvider, resourceName),
					resource.TestCheckResourceAttr(resourceName, "content", fmt.Sprintf(`{"permissions":[%q,%q]}`, permissionID, otherPermissionID)),
				),
			},
			{
				// policy_name is renamed in-place for role policies too.
				Config: testAccCheckDatadogOrgGroupPolicyRoleConfig(orgGroupName, renamedPolicyName, fmt.Sprintf(`{"permissions":[%q,%q]}`, permissionID, otherPermissionID), "GROUP_MANAGED"),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogOrgGroupPolicyExists(providers.frameworkProvider, resourceName),
					resource.TestCheckResourceAttr(resourceName, "policy_name", renamedPolicyName),
				),
			},
			{
				// enforcement_tier = OVERRIDE_ALLOWED is invalid for role policies; caught at plan time.
				Config:      testAccCheckDatadogOrgGroupPolicyRoleConfig(orgGroupName, renamedPolicyName, fmt.Sprintf(`{"permissions":[%q,%q]}`, permissionID, otherPermissionID), "OVERRIDE_ALLOWED"),
				PlanOnly:    true,
				ExpectError: regexp.MustCompile(`role policies only support GROUP_MANAGED and DELEGATE`),
			},
			{
				// DELEGATE disables the shared role.
				Config: testAccCheckDatadogOrgGroupPolicyRoleConfig(orgGroupName, renamedPolicyName, fmt.Sprintf(`{"permissions":[%q,%q]}`, permissionID, otherPermissionID), "DELEGATE"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogOrgGroupPolicyExists(providers.frameworkProvider, resourceName),
					resource.TestCheckResourceAttr(resourceName, "enforcement_tier", "DELEGATE"),
				),
			},
			{
				// Disabling is one-way: the provider rejects transitioning back to
				// GROUP_MANAGED before ever calling the API.
				Config:      testAccCheckDatadogOrgGroupPolicyRoleConfig(orgGroupName, renamedPolicyName, fmt.Sprintf(`{"permissions":[%q,%q]}`, permissionID, otherPermissionID), "GROUP_MANAGED"),
				ExpectError: regexp.MustCompile(`(?is)cannot be\s+transitioned back`),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckDatadogOrgGroupPolicyRoleConfig(orgGroupName, policyName, content, enforcementTier string) string {
	return fmt.Sprintf(`
resource "datadog_org_group" "role" {
  name = "%s"
}

resource "datadog_org_group_policy" "role" {
  org_group_id     = datadog_org_group.role.id
  policy_name      = "%s"
  policy_type      = "role"
  content          = jsonencode(%s)
  enforcement_tier = "%s"
}`, orgGroupName, policyName, content, enforcementTier)
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

			resp, httpResp, err := apiInstances.GetOrgGroupsApiV2().GetOrgGroupPolicy(auth, id)
			if err != nil {
				if httpResp != nil && httpResp.StatusCode == 404 {
					continue
				}
				return fmt.Errorf("received an error retrieving org group policy: %w", err)
			}

			// role policies are never hard-deleted server-side; Delete() only succeeds
			// once the policy is already disabled (enforcement_tier = DELEGATE), so
			// "destroyed" means disabled, not 404.
			if r.Primary.Attributes["policy_type"] == "role" {
				if resp.Data.Attributes.GetEnforcementTier() == "DELEGATE" {
					continue
				}
				return fmt.Errorf("role policy was not disabled on destroy")
			}

			return fmt.Errorf("org group policy still exists")
		}

		return nil
	}
}
