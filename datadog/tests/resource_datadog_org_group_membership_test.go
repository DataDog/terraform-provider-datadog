package test

import (
	"context"
	"fmt"
	"testing"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/fwprovider"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

// getTestOrgUUID uses the V1 Organizations API (through the cassette recorder) to get the test org's public ID.
func getTestOrgUUID(t *testing.T, auth context.Context, apiInstances *utils.ApiInstances) string {
	t.Helper()

	resp, _, err := apiInstances.GetOrganizationsApiV1().ListOrgs(auth)
	if err != nil {
		t.Fatalf("error listing orgs: %s", err)
	}

	orgs := resp.GetOrgs()
	if len(orgs) == 0 {
		t.Fatal("no orgs found")
	}

	return orgs[0].GetPublicId()
}

// getOrgCurrentGroupID looks up the org's current org group membership and returns the group ID.
func getOrgCurrentGroupID(t *testing.T, auth context.Context, apiInstances *utils.ApiInstances, orgUUID string) string {
	t.Helper()

	id, err := uuid.Parse(orgUUID)
	if err != nil {
		t.Fatalf("invalid org UUID: %s", err)
	}

	api := apiInstances.GetOrgGroupsApiV2()
	params := datadogV2.NewListOrgGroupMembershipsOptionalParameters().WithFilterOrgUuid(id)
	resp, _, err := api.ListOrgGroupMemberships(auth, *params)
	if err != nil {
		t.Fatalf("error listing org group memberships: %s", err)
	}

	memberships := resp.GetData()
	if len(memberships) == 0 {
		t.Fatalf("no membership found for org %s", orgUUID)
	}

	membership := memberships[0]
	if rels, ok := membership.GetRelationshipsOk(); ok && rels != nil {
		if orgGroup, ok := rels.GetOrgGroupOk(); ok && orgGroup != nil {
			if orgGroupData, ok := orgGroup.GetDataOk(); ok {
				return orgGroupData.GetId().String()
			}
		}
	}

	t.Fatalf("membership for org %s has no org group relationship", orgUUID)
	return ""
}

func TestAccDatadogOrgGroupMembership_Basic(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	orgGroupName := uniqueEntityName(ctx, t)
	orgGroupName2 := orgGroupName + "-2"
	resourceName := "datadog_org_group_membership.foo"

	orgUUID := getTestOrgUUID(t, providers.frameworkProvider.Auth, providers.frameworkProvider.DatadogApiInstances)
	originalGroupID := getOrgCurrentGroupID(t, providers.frameworkProvider.Auth, providers.frameworkProvider.DatadogApiInstances, orgUUID)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: accProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogOrgGroupMembershipConfig(orgGroupName, orgUUID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogOrgGroupMembershipExists(providers.frameworkProvider, resourceName),
					resource.TestCheckResourceAttr(resourceName, "org_uuid", orgUUID),
					resource.TestCheckResourceAttrSet(resourceName, "org_site"),
					resource.TestCheckResourceAttrSet(resourceName, "org_name"),
					resource.TestCheckResourceAttrSet(resourceName, "org_group_id"),
				),
			},
			{
				Config: testAccCheckDatadogOrgGroupMembershipConfigUpdated(orgGroupName, orgGroupName2, orgUUID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogOrgGroupMembershipExists(providers.frameworkProvider, resourceName),
					resource.TestCheckResourceAttr(resourceName, "org_uuid", orgUUID),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Move org back to its original group so the test org groups can be destroyed
			{
				Config: testAccCheckDatadogOrgGroupMembershipConfigRestore(orgGroupName, orgGroupName2, orgUUID, originalGroupID),
			},
		},
	})
}

func testAccCheckDatadogOrgGroupMembershipConfig(orgGroupName, orgUUID string) string {
	return fmt.Sprintf(`
resource "datadog_org_group" "test" {
  name = "%s"
}

resource "datadog_org_group_membership" "foo" {
  org_group_id = datadog_org_group.test.id
  org_uuid     = "%s"
}`, orgGroupName, orgUUID)
}

func testAccCheckDatadogOrgGroupMembershipConfigUpdated(orgGroupName, orgGroupName2, orgUUID string) string {
	return fmt.Sprintf(`
resource "datadog_org_group" "test" {
  name = "%s"
}

resource "datadog_org_group" "test2" {
  name = "%s"
}

resource "datadog_org_group_membership" "foo" {
  org_group_id = datadog_org_group.test2.id
  org_uuid     = "%s"
}`, orgGroupName, orgGroupName2, orgUUID)
}

// testAccCheckDatadogOrgGroupMembershipConfigRestore moves the org back to its
// original group and removes the membership resource so the test org groups
// can be cleanly destroyed.
func testAccCheckDatadogOrgGroupMembershipConfigRestore(orgGroupName, orgGroupName2, orgUUID, originalGroupID string) string {
	return fmt.Sprintf(`
resource "datadog_org_group" "test" {
  name = "%s"
}

resource "datadog_org_group" "test2" {
  name = "%s"
}

resource "datadog_org_group_membership" "foo" {
  org_group_id = "%s"
  org_uuid     = "%s"
}`, orgGroupName, orgGroupName2, originalGroupID, orgUUID)
}

func testAccCheckDatadogOrgGroupMembershipExists(accProvider *fwprovider.FrameworkProvider, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		apiInstances := accProvider.DatadogApiInstances
		auth := accProvider.Auth

		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("resource not found: %s", n)
		}

		id, err := uuid.Parse(rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("membership ID is not a valid UUID: %s", err)
		}

		_, _, err = apiInstances.GetOrgGroupsApiV2().GetOrgGroupMembership(auth, id)
		if err != nil {
			return fmt.Errorf("received an error retrieving org group membership: %s", err)
		}
		return nil
	}
}
