package test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/fwprovider"
)

const overrideTestOrgSite = "us1"

func TestAccDatadogOrgGroupPolicyOverride_Basic(t *testing.T) {
	// Not parallel: the three override tests all move the shared test org between groups;
	// parallelism causes one test's membership change to drift another's expected state.
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	orgGroupName := uniqueEntityName(ctx, t)
	resourceName := "datadog_org_group_policy_override.foo"

	orgUUID := getTestOrgUUID(providers.frameworkProvider.Auth, t, providers.frameworkProvider.DatadogApiInstances)
	originalGroupID := getOrgCurrentGroupID(providers.frameworkProvider.Auth, t, providers.frameworkProvider.DatadogApiInstances, orgUUID)

	t.Cleanup(restoreOrgMembership(t, providers.frameworkProvider, orgUUID, originalGroupID))

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: accProviders,
		CheckDestroy:             composeOrgGroupStackDestroyChecks(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogOrgGroupPolicyOverrideConfigBasic(orgGroupName, orgUUID, "datadog_org_group_policy.foo.id"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogOrgGroupPolicyOverrideExists(providers.frameworkProvider, resourceName),
					resource.TestCheckResourceAttr(resourceName, "org_uuid", orgUUID),
					resource.TestCheckResourceAttr(resourceName, "org_site", overrideTestOrgSite),
					resource.TestCheckResourceAttrSet(resourceName, "org_group_id"),
					resource.TestCheckResourceAttrSet(resourceName, "policy_id"),
					resource.TestCheckResourceAttrSet(resourceName, "content"),
				),
			},
			{
				// Swap policy_id to a second policy → RequiresReplace must fire on
				// policy_id. org_group_id has its own plancheck coverage in the
				// dedicated _OrgGroupIdRequiresReplace test (which also exercises
				// co-moving membership+policy and an unwind step). org_uuid and
				// org_site remain backed only by the resource's Update-unreachable
				// error guard — they're not applyable here (org_uuid needs a second
				// real org; org_site is server-validated against the real org's
				// site). A regression dropping RequiresReplace on any of the four
				// fields still triggers Update, which the resource errors on loudly.
				Config: testAccCheckDatadogOrgGroupPolicyOverrideConfigReplace(orgGroupName, orgUUID),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionDestroyBeforeCreate),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogOrgGroupPolicyOverrideExists(providers.frameworkProvider, resourceName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Move org back to its original group so the test org_group can be destroyed.
			{
				Config: testAccCheckDatadogOrgGroupPolicyOverrideConfigRestore(orgGroupName, orgUUID, originalGroupID),
			},
		},
	})
}

func TestAccDatadogOrgGroupPolicyOverride_EnforceCascade(t *testing.T) {
	// Not parallel: shared test org's membership.
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	orgGroupName := uniqueEntityName(ctx, t)
	resourceName := "datadog_org_group_policy_override.foo"

	orgUUID := getTestOrgUUID(providers.frameworkProvider.Auth, t, providers.frameworkProvider.DatadogApiInstances)
	originalGroupID := getOrgCurrentGroupID(providers.frameworkProvider.Auth, t, providers.frameworkProvider.DatadogApiInstances, orgUUID)

	t.Cleanup(restoreOrgMembership(t, providers.frameworkProvider, orgUUID, originalGroupID))

	// Captured in step 1's Check, read from step 2's PreConfig (no state access there).
	var capturedPolicyID, capturedOverrideID string

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: accProviders,
		CheckDestroy:             composeOrgGroupStackDestroyChecks(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogOrgGroupPolicyOverrideConfigBasic(orgGroupName, orgUUID, "datadog_org_group_policy.foo.id"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogOrgGroupPolicyOverrideExists(providers.frameworkProvider, resourceName),
					func(s *terraform.State) error {
						rs, ok := s.RootModule().Resources[resourceName]
						if !ok {
							return fmt.Errorf("resource not found: %s", resourceName)
						}
						capturedPolicyID = rs.Primary.Attributes["policy_id"]
						capturedOverrideID = rs.Primary.ID
						return nil
					},
				),
			},
			{
				// Flip the parent policy's tier to ENFORCE via direct API call. The
				// server cascades the override delete; Terraform's refresh surfaces it
				// as drift. Check asserts the cascade by confirming the override is
				// 404 server-side (distinguishes a real cascade from unrelated drift).
				PreConfig: func() {
					policyID, err := uuid.Parse(capturedPolicyID)
					if err != nil {
						t.Fatalf("captured policy ID is not a UUID: %s", err)
					}
					api := providers.frameworkProvider.DatadogApiInstances.GetOrgGroupsApiV2()
					attrs := datadogV2.NewOrgGroupPolicyUpdateAttributes()
					attrs.SetEnforcementTier(datadogV2.ORGGROUPPOLICYENFORCEMENTTIER_ENFORCE)
					data := datadogV2.NewOrgGroupPolicyUpdateData(*attrs, policyID, datadogV2.ORGGROUPPOLICYTYPE_ORG_GROUP_POLICIES)
					if _, _, err := api.UpdateOrgGroupPolicy(providers.frameworkProvider.Auth, policyID, *datadogV2.NewOrgGroupPolicyUpdateRequest(*data)); err != nil {
						t.Fatalf("failed to flip policy to ENFORCE: %s", err)
					}
				},
				Config:             testAccCheckDatadogOrgGroupPolicyOverrideConfigBasic(orgGroupName, orgUUID, "datadog_org_group_policy.foo.id"),
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
				Check: func(_ *terraform.State) error {
					id, err := uuid.Parse(capturedOverrideID)
					if err != nil {
						return fmt.Errorf("captured override ID is not a UUID: %w", err)
					}
					_, httpResp, err := providers.frameworkProvider.DatadogApiInstances.GetOrgGroupsApiV2().GetOrgGroupPolicyOverride(providers.frameworkProvider.Auth, id)
					if err == nil {
						return fmt.Errorf("expected override %s to be cascaded but Get succeeded", id)
					}
					if httpResp == nil || httpResp.StatusCode != 404 {
						return fmt.Errorf("expected 404, got: %w", err)
					}
					return nil
				},
			},
			// Restore membership so the test group can be destroyed cleanly.
			{
				Config: testAccCheckDatadogOrgGroupPolicyOverrideConfigRestore(orgGroupName, orgUUID, originalGroupID),
			},
		},
	})
}

func TestAccDatadogOrgGroupPolicyOverride_AutoCreation(t *testing.T) {
	// Not parallel: the three override tests all move the shared test org between groups;
	// parallelism causes one test's membership change to drift another's expected state.
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	orgGroupName := uniqueEntityName(ctx, t)

	orgUUID := getTestOrgUUID(providers.frameworkProvider.Auth, t, providers.frameworkProvider.DatadogApiInstances)
	originalGroupID := getOrgCurrentGroupID(providers.frameworkProvider.Auth, t, providers.frameworkProvider.DatadogApiInstances, orgUUID)

	// Safety net: always restore the org to its original group, even if the test fails mid-way.
	t.Cleanup(restoreOrgMembership(t, providers.frameworkProvider, orgUUID, originalGroupID))

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: accProviders,
		CheckDestroy:             composeOrgGroupStackDestroyChecks(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				// Step 1: move org into the test group and pin its config via an ENFORCE
				// policy with value=true. ENFORCE propagation sets the org's
				// is_widget_copy_paste_enabled=true.
				Config: testAccCheckDatadogOrgGroupPolicyOverrideAutoCreationStep1(orgGroupName, orgUUID),
			},
			{
				// Step 2: delete the ENFORCE policy on its own. The org retains its true
				// value (enforce deletion does not reset org_config). This isolation is
				// required: if the DEFAULT policy were created in the same apply, Terraform
				// could issue the create before the destroy, and the server would treat
				// the new DEFAULT as an in-place update of the existing ENFORCE (since
				// policies are keyed by name+org_group), skipping the override computation.
				Config: testAccCheckDatadogOrgGroupPolicyOverrideAutoCreationStep2(orgGroupName, orgUUID),
			},
			{
				// Step 3: create a brand-new DEFAULT policy with value=false. The
				// server's policy-propagation path detects the org's retained true
				// value ≠ false and auto-creates an override. The Check just confirms
				// the override exists server-side; the fact that subsequent refreshes
				// don't fail is our "provider can read a server-created row" coverage.
				Config: testAccCheckDatadogOrgGroupPolicyOverrideAutoCreationStep3(orgGroupName, orgUUID),
				Check:  testAccCheckAutoCreatedOverrideExists(providers.frameworkProvider, "datadog_org_group.grp", "datadog_org_group_policy.dflt", orgUUID),
			},
			// Restore membership so the org_group can be destroyed cleanly.
			{
				Config: testAccCheckDatadogOrgGroupPolicyOverrideAutoCreationRestore(orgGroupName, orgUUID, originalGroupID),
			},
		},
	})
}

func testAccCheckDatadogOrgGroupPolicyOverrideConfigBasic(orgGroupName, orgUUID, policyIDRef string) string {
	return fmt.Sprintf(`
resource "datadog_org_group" "foo" {
  name = "%s"
}

resource "datadog_org_group_membership" "foo" {
  org_group_id = datadog_org_group.foo.id
  org_uuid     = "%s"
}

resource "datadog_org_group_policy" "foo" {
  org_group_id     = datadog_org_group.foo.id
  policy_name      = "is_widget_copy_paste_enabled"
  content          = jsonencode({"org_config": false})
  enforcement_tier = "DEFAULT"
}

resource "datadog_org_group_policy_override" "foo" {
  org_group_id = datadog_org_group.foo.id
  policy_id    = %s
  org_uuid     = "%s"
  org_site     = "%s"
  depends_on   = [datadog_org_group_membership.foo]
}`, orgGroupName, orgUUID, policyIDRef, orgUUID, overrideTestOrgSite)
}

func testAccCheckDatadogOrgGroupPolicyOverrideConfigReplace(orgGroupName, orgUUID string) string {
	return fmt.Sprintf(`
resource "datadog_org_group" "foo" {
  name = "%s"
}

resource "datadog_org_group_membership" "foo" {
  org_group_id = datadog_org_group.foo.id
  org_uuid     = "%s"
}

resource "datadog_org_group_policy" "foo" {
  org_group_id     = datadog_org_group.foo.id
  policy_name      = "is_widget_copy_paste_enabled"
  content          = jsonencode({"org_config": false})
  enforcement_tier = "DEFAULT"
}

resource "datadog_org_group_policy" "bar" {
  org_group_id     = datadog_org_group.foo.id
  policy_name      = "is_dashboard_reports_enabled"
  content          = jsonencode({"org_config": false})
  enforcement_tier = "DEFAULT"
}

resource "datadog_org_group_policy_override" "foo" {
  org_group_id = datadog_org_group.foo.id
  policy_id    = datadog_org_group_policy.bar.id
  org_uuid     = "%s"
  org_site     = "%s"
  depends_on   = [datadog_org_group_membership.foo]
}`, orgGroupName, orgUUID, orgUUID, overrideTestOrgSite)
}

// testAccCheckDatadogOrgGroupPolicyOverrideConfigRestore reassigns the test org
// back to its original group so the test-created org_group can be destroyed cleanly
// (the org_group DELETE endpoint refuses to remove a group that still has members).
func testAccCheckDatadogOrgGroupPolicyOverrideConfigRestore(orgGroupName, orgUUID, originalGroupID string) string {
	return fmt.Sprintf(`
resource "datadog_org_group" "foo" {
  name = "%s"
}

resource "datadog_org_group_membership" "foo" {
  org_group_id = "%s"
  org_uuid     = "%s"
}`, orgGroupName, originalGroupID, orgUUID)
}

func testAccCheckDatadogOrgGroupPolicyOverrideAutoCreationStep1(orgGroupName, orgUUID string) string {
	return fmt.Sprintf(`
resource "datadog_org_group" "grp" {
  name = "%s"
}

resource "datadog_org_group_membership" "org" {
  org_group_id = datadog_org_group.grp.id
  org_uuid     = "%s"
}

resource "datadog_org_group_policy" "enforce" {
  org_group_id     = datadog_org_group.grp.id
  policy_name      = "is_widget_copy_paste_enabled"
  content          = jsonencode({"org_config": true})
  enforcement_tier = "ENFORCE"
  depends_on       = [datadog_org_group_membership.org]
}`, orgGroupName, orgUUID)
}

func testAccCheckDatadogOrgGroupPolicyOverrideAutoCreationStep2(orgGroupName, orgUUID string) string {
	return fmt.Sprintf(`
resource "datadog_org_group" "grp" {
  name = "%s"
}

resource "datadog_org_group_membership" "org" {
  org_group_id = datadog_org_group.grp.id
  org_uuid     = "%s"
}`, orgGroupName, orgUUID)
}

func testAccCheckDatadogOrgGroupPolicyOverrideAutoCreationStep3(orgGroupName, orgUUID string) string {
	return fmt.Sprintf(`
resource "datadog_org_group" "grp" {
  name = "%s"
}

resource "datadog_org_group_membership" "org" {
  org_group_id = datadog_org_group.grp.id
  org_uuid     = "%s"
}

resource "datadog_org_group_policy" "dflt" {
  org_group_id     = datadog_org_group.grp.id
  policy_name      = "is_widget_copy_paste_enabled"
  content          = jsonencode({"org_config": false})
  enforcement_tier = "DEFAULT"
  depends_on       = [datadog_org_group_membership.org]
}`, orgGroupName, orgUUID)
}

func testAccCheckDatadogOrgGroupPolicyOverrideAutoCreationRestore(orgGroupName, orgUUID, originalGroupID string) string {
	return fmt.Sprintf(`
resource "datadog_org_group" "grp" {
  name = "%s"
}

resource "datadog_org_group_membership" "org" {
  org_group_id = "%s"
  org_uuid     = "%s"
}`, orgGroupName, originalGroupID, orgUUID)
}

// restoreOrgMembership returns a t.Cleanup callback that moves the test org back
// to its original group via direct API calls when a test fails mid-way. We gate
// on t.Failed() because a passing test has already run its Terraform Restore step;
// running the API fallback too would just make noise. When cleanup *must* fire
// (the test did fail), only the final Update error is promoted to t.Errorf —
// other branches log because the test is already failing and more errors just
// crowd out the real one.
func restoreOrgMembership(t *testing.T, accProvider *fwprovider.FrameworkProvider, orgUUID, originalGroupID string) func() {
	return func() {
		if !t.Failed() {
			return
		}

		apiInstances := accProvider.DatadogApiInstances
		auth := accProvider.Auth

		orgUUIDParsed, err := uuid.Parse(orgUUID)
		if err != nil {
			t.Logf("cleanup: invalid org UUID %s: %s", orgUUID, err)
			return
		}
		targetGroupID, err := uuid.Parse(originalGroupID)
		if err != nil {
			t.Logf("cleanup: invalid original group ID %s: %s", originalGroupID, err)
			return
		}

		api := apiInstances.GetOrgGroupsApiV2()
		params := datadogV2.NewListOrgGroupMembershipsOptionalParameters().WithFilterOrgUuid(orgUUIDParsed)
		listResp, _, err := api.ListOrgGroupMemberships(auth, *params)
		if err != nil {
			t.Logf("cleanup: listing memberships: %s", err)
			return
		}
		memberships := listResp.GetData()
		if len(memberships) == 0 {
			t.Logf("cleanup: no membership found for org %s", orgUUID)
			return
		}
		membership := memberships[0]
		// If the org is already in the target group, the PATCH would be a no-op.
		if rels, ok := membership.GetRelationshipsOk(); ok && rels != nil {
			if orgGroup, ok := rels.GetOrgGroupOk(); ok && orgGroup != nil {
				if ogData, ok := orgGroup.GetDataOk(); ok && ogData.GetId() == targetGroupID {
					t.Logf("cleanup: org %s already in target group %s, nothing to do", orgUUID, originalGroupID)
					return
				}
			}
		}
		membershipID := membership.GetId()

		orgGroupRef := datadogV2.NewOrgGroupRelationshipToOneData(targetGroupID, datadogV2.ORGGROUPTYPE_ORG_GROUPS)
		rel := datadogV2.NewOrgGroupRelationshipToOne(*orgGroupRef)
		rels := datadogV2.NewOrgGroupMembershipUpdateRelationships(*rel)
		data := datadogV2.NewOrgGroupMembershipUpdateData(membershipID, *rels, datadogV2.ORGGROUPMEMBERSHIPTYPE_ORG_GROUP_MEMBERSHIPS)
		body := datadogV2.NewOrgGroupMembershipUpdateRequest(*data)

		if _, _, err := api.UpdateOrgGroupMembership(auth, membershipID, *body); err != nil {
			t.Errorf("cleanup: restoring org %s to group %s: %s", orgUUID, originalGroupID, err)
		}
	}
}

func testAccCheckDatadogOrgGroupPolicyOverrideExists(accProvider *fwprovider.FrameworkProvider, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		apiInstances := accProvider.DatadogApiInstances
		auth := accProvider.Auth

		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("resource not found: %s", n)
		}

		id, err := uuid.Parse(rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("override ID is not a valid UUID: %w", err)
		}

		_, _, err = apiInstances.GetOrgGroupsApiV2().GetOrgGroupPolicyOverride(auth, id)
		if err != nil {
			return fmt.Errorf("received an error retrieving org group policy override: %w", err)
		}
		return nil
	}
}

func testAccCheckDatadogOrgGroupPolicyOverrideDestroy(accProvider *fwprovider.FrameworkProvider) func(*terraform.State) error {
	return func(s *terraform.State) error {
		apiInstances := accProvider.DatadogApiInstances
		auth := accProvider.Auth

		for _, r := range s.RootModule().Resources {
			if r.Type != "datadog_org_group_policy_override" {
				continue
			}

			id, err := uuid.Parse(r.Primary.ID)
			if err != nil {
				return fmt.Errorf("override ID is not a valid UUID: %w", err)
			}

			_, httpResp, err := apiInstances.GetOrgGroupsApiV2().GetOrgGroupPolicyOverride(auth, id)
			if err != nil {
				if httpResp != nil && httpResp.StatusCode == 404 {
					continue
				}
				return fmt.Errorf("received an error retrieving org group policy override: %w", err)
			}

			return fmt.Errorf("org group policy override still exists")
		}

		return nil
	}
}

// composeOrgGroupStackDestroyChecks runs all three sub-destroy checks via
// errors.Join so one test run surfaces every leaked resource type at once.
// Membership is intentionally excluded — its Delete is a state-only no-op.
func composeOrgGroupStackDestroyChecks(accProvider *fwprovider.FrameworkProvider) func(*terraform.State) error {
	checks := []func(*terraform.State) error{
		testAccCheckDatadogOrgGroupPolicyOverrideDestroy(accProvider),
		testAccCheckDatadogOrgGroupPolicyDestroy(accProvider),
		testAccCheckDatadogOrgGroupDestroy(accProvider),
	}
	return func(s *terraform.State) error {
		var errs []error
		for _, check := range checks {
			if err := check(s); err != nil {
				errs = append(errs, err)
			}
		}
		return errors.Join(errs...)
	}
}

// testAccCheckAutoCreatedOverrideExists lists overrides for the given org_group and
// verifies an override exists for our test org on the given policy. This proves the
// server's membership-propagation path created the override when the org was moved.
func testAccCheckAutoCreatedOverrideExists(accProvider *fwprovider.FrameworkProvider, orgGroupResource, policyResource, orgUUID string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		apiInstances := accProvider.DatadogApiInstances
		auth := accProvider.Auth

		orgGroupRS, ok := s.RootModule().Resources[orgGroupResource]
		if !ok {
			return fmt.Errorf("resource not found: %s", orgGroupResource)
		}
		policyRS, ok := s.RootModule().Resources[policyResource]
		if !ok {
			return fmt.Errorf("resource not found: %s", policyResource)
		}

		orgGroupID, err := uuid.Parse(orgGroupRS.Primary.ID)
		if err != nil {
			return fmt.Errorf("org_group ID is not a valid UUID: %w", err)
		}
		policyID, err := uuid.Parse(policyRS.Primary.ID)
		if err != nil {
			return fmt.Errorf("policy ID is not a valid UUID: %w", err)
		}

		api := apiInstances.GetOrgGroupsApiV2()
		params := datadogV2.NewListOrgGroupPolicyOverridesOptionalParameters().WithFilterPolicyId(policyID)
		resp, _, err := api.ListOrgGroupPolicyOverrides(auth, orgGroupID, *params)
		if err != nil {
			return fmt.Errorf("listing overrides: %w", err)
		}

		for _, override := range resp.GetData() {
			attrs := override.GetAttributes()
			if attrs.GetOrgUuid().String() == orgUUID {
				return nil
			}
		}

		return fmt.Errorf("expected auto-created override for org %s on policy %s, none found", orgUUID, policyID.String())
	}
}
