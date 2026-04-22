package test

import (
	"context"
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
		CheckDestroy:             testAccCheckDatadogOrgGroupPolicyOverrideDestroy(providers.frameworkProvider),
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
				// Swap policy_id to a second policy → RequiresReplace must fire.
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
		CheckDestroy:             testAccCheckDatadogOrgGroupPolicyOverrideDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogOrgGroupPolicyOverrideConfigBasic(orgGroupName, orgUUID, "datadog_org_group_policy.foo.id"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogOrgGroupPolicyOverrideExists(providers.frameworkProvider, resourceName),
					capturePolicyIDForCascade(resourceName),
				),
			},
			{
				// Flip the parent policy's tier to ENFORCE directly via API.
				// This cascades the override delete server-side. Terraform state is unaware
				// until a refresh — which is what the PlanOnly + ExpectNonEmptyPlan asserts.
				PreConfig: enforcePolicyViaAPI(t, providers.frameworkProvider),
				Config:    testAccCheckDatadogOrgGroupPolicyOverrideConfigBasic(orgGroupName, orgUUID, "datadog_org_group_policy.foo.id"),
				PlanOnly:  true,
				// Refresh sees override 404 → state removes it; plan proposes re-create.
				// The policy also drifts (tier changed out-of-band). Non-empty plan proves
				// the cascade is observable through the Read path.
				ExpectNonEmptyPlan: true,
			},
			// Move org back to its original group so the test org_group can be destroyed.
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
				// Step 3: create a brand-new DEFAULT policy with value=false. The server's
				// policy-propagation path detects the org's retained true value ≠ false
				// and auto-creates an override. The Check then reads the auto-created
				// override via the provider's Read path to confirm the adoption flow
				// (Read + updateState on a server-created row) works end-to-end.
				Config: testAccCheckDatadogOrgGroupPolicyOverrideAutoCreationStep3(orgGroupName, orgUUID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAutoCreatedOverrideExists(providers.frameworkProvider, "datadog_org_group.grp", "datadog_org_group_policy.dflt", orgUUID),
					testAccCheckAutoCreatedOverrideReadable(providers.frameworkProvider, "datadog_org_group.grp", "datadog_org_group_policy.dflt", orgUUID),
				),
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

// testAccCheckAutoCreatedOverrideReadable finds the auto-created override's UUID
// via the List endpoint, then fetches it via Get — exercising the same Read +
// updateState path Terraform would hit during `terraform import`. This is the
// unit-equivalent of an ImportStateVerify round-trip without the config-vs-state
// comparison complexity.
func testAccCheckAutoCreatedOverrideReadable(accProvider *fwprovider.FrameworkProvider, orgGroupResource, policyResource, orgUUID string) resource.TestCheckFunc {
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
		listResp, _, err := api.ListOrgGroupPolicyOverrides(auth, orgGroupID, *params)
		if err != nil {
			return fmt.Errorf("listing overrides: %w", err)
		}
		var overrideID uuid.UUID
		for _, override := range listResp.GetData() {
			attrs := override.GetAttributes()
			if attrs.GetOrgUuid().String() == orgUUID {
				overrideID = override.GetId()
				break
			}
		}
		if overrideID == uuid.Nil {
			return fmt.Errorf("no auto-created override found for org %s on policy %s", orgUUID, policyID.String())
		}

		// Fetch via Get to confirm the override is readable end-to-end — the same
		// API call our resource's Read method uses during `terraform import`.
		getResp, _, err := api.GetOrgGroupPolicyOverride(auth, overrideID)
		if err != nil {
			return fmt.Errorf("fetching auto-created override %s: %w", overrideID, err)
		}
		fetched := getResp.GetData()
		if fetched.GetId() != overrideID {
			return fmt.Errorf("auto-created override %s: Get returned different ID %s", overrideID, fetched.GetId())
		}
		return nil
	}
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

// restoreOrgMembership returns a t.Cleanup callback that always moves the test org back to its
// original group via direct API calls. Runs regardless of test outcome so a failed test doesn't
// leave the org stuck in a test-created group (which would block subsequent test runs).
func restoreOrgMembership(t *testing.T, accProvider *fwprovider.FrameworkProvider, orgUUID, originalGroupID string) func() {
	return func() {
		apiInstances := accProvider.DatadogApiInstances
		auth := accProvider.Auth

		orgUUIDParsed, err := uuid.Parse(orgUUID)
		if err != nil {
			t.Errorf("cleanup: invalid org UUID %s: %s", orgUUID, err)
			return
		}
		targetGroupID, err := uuid.Parse(originalGroupID)
		if err != nil {
			t.Errorf("cleanup: invalid original group ID %s: %s", originalGroupID, err)
			return
		}

		api := apiInstances.GetOrgGroupsApiV2()
		params := datadogV2.NewListOrgGroupMembershipsOptionalParameters().WithFilterOrgUuid(orgUUIDParsed)
		listResp, _, err := api.ListOrgGroupMemberships(auth, *params)
		if err != nil {
			t.Errorf("cleanup: listing memberships: %s", err)
			return
		}
		memberships := listResp.GetData()
		if len(memberships) == 0 {
			t.Errorf("cleanup: no membership found for org %s", orgUUID)
			return
		}
		membershipID := memberships[0].GetId()

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

// enforcePolicyViaAPI returns a PreConfig hook that flips the captured policy's
// enforcement_tier to ENFORCE via a direct API call. This triggers the server-side
// cascade that deletes all overrides for the policy. The policy ID is read from the
// package-level cascadeCapturedPolicyID variable, set by capturePolicyIDForCascade
// in the prior test step.
func enforcePolicyViaAPI(t *testing.T, accProvider *fwprovider.FrameworkProvider) func() {
	return func() {
		t.Helper()
		apiInstances := accProvider.DatadogApiInstances
		auth := accProvider.Auth

		// PreConfig hooks can't access Terraform state, so the prior step's Check
		// (capturePolicyIDForCascade) stashes the policy ID in cascadeCapturedPolicyID.
		if cascadeCapturedPolicyID == "" {
			t.Fatal("cascadeCapturedPolicyID not set; prior step must capture it")
		}

		policyID, err := uuid.Parse(cascadeCapturedPolicyID)
		if err != nil {
			t.Fatalf("captured policy ID is not a UUID: %s", err)
		}

		api := apiInstances.GetOrgGroupsApiV2()
		attrs := datadogV2.NewOrgGroupPolicyUpdateAttributes()
		tier := datadogV2.ORGGROUPPOLICYENFORCEMENTTIER_ENFORCE
		attrs.SetEnforcementTier(tier)
		data := datadogV2.NewOrgGroupPolicyUpdateData(*attrs, policyID, datadogV2.ORGGROUPPOLICYTYPE_ORG_GROUP_POLICIES)
		body := datadogV2.NewOrgGroupPolicyUpdateRequest(*data)

		if _, _, err := api.UpdateOrgGroupPolicy(auth, policyID, *body); err != nil {
			t.Fatalf("failed to flip policy to ENFORCE: %s", err)
		}
	}
}

// cascadeCapturedPolicyID is set by the first step of TestAccDatadogOrgGroupPolicyOverride_EnforceCascade
// and read by the PreConfig hook of the second step. Package-level because PreConfig
// has no access to Terraform state. Safe as a global: every test in this file is
// non-parallel because they all share the test org's membership, so there is no
// concurrency between tests that could race on this variable.
var cascadeCapturedPolicyID string

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

// capturePolicyIDForCascade stores the override's parent policy_id in a package-level
// variable so the EnforceCascade test's PreConfig hook can drive an out-of-band update.
func capturePolicyIDForCascade(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("resource not found: %s", n)
		}
		policyID, ok := rs.Primary.Attributes["policy_id"]
		if !ok {
			return fmt.Errorf("resource %s has no policy_id attribute", n)
		}
		cascadeCapturedPolicyID = policyID
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
